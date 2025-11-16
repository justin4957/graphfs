/*
# Module: pkg/cache/cache.go
Query result caching with LRU eviction.

Provides in-memory caching for SPARQL, GraphQL, and REST API queries
with configurable TTL and LRU eviction policy.

## Linked Modules
None (standalone package)

## Tags
cache, performance, lru

## Exports
Cache, NewCache, CacheEntry, Stats

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cache.go> a code:Module ;
    code:name "pkg/cache/cache.go" ;
    code:description "Query result caching with LRU eviction" ;
    code:language "go" ;
    code:layer "cache" ;
    code:exports <#Cache>, <#NewCache>, <#CacheEntry>, <#Stats> ;
    code:tags "cache", "performance", "lru" .
<!-- End LinkedDoc RDF -->
*/

package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CacheEntry represents a cached value with metadata
type CacheEntry struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
	Hits      int64
	Size      int64
}

// Stats represents cache statistics
type Stats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	Size       int
	MaxSize    int
	TotalBytes int64
	HitRate    float64
}

// Cache is an LRU cache with TTL support
type Cache struct {
	mu         sync.RWMutex
	entries    map[string]*list.Element
	lruList    *list.List
	maxEntries int
	ttl        time.Duration
	hits       int64
	misses     int64
	evictions  int64
	totalBytes int64
}

type cacheItem struct {
	key   string
	entry *CacheEntry
}

// NewCache creates a new cache with the given max entries and TTL
func NewCache(maxEntries int, ttl time.Duration) *Cache {
	return &Cache{
		entries:    make(map[string]*list.Element),
		lruList:    list.New(),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, found := c.entries[key]
	if !found {
		c.misses++
		return nil, false
	}

	item := elem.Value.(*cacheItem)

	// Check if expired
	if time.Now().After(item.entry.ExpiresAt) {
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(elem)
	item.entry.Hits++
	c.hits++

	return item.entry.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update it
	if elem, found := c.entries[key]; found {
		item := elem.Value.(*cacheItem)
		c.totalBytes -= item.entry.Size
		item.entry.Value = value
		item.entry.Size = size
		item.entry.ExpiresAt = time.Now().Add(c.ttl)
		c.totalBytes += size
		c.lruList.MoveToFront(elem)
		return
	}

	// Evict if at capacity
	if c.lruList.Len() >= c.maxEntries {
		c.evictOldest()
	}

	// Add new entry
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
		Hits:      0,
		Size:      size,
	}

	item := &cacheItem{
		key:   key,
		entry: entry,
	}

	elem := c.lruList.PushFront(item)
	c.entries[key] = elem
	c.totalBytes += size
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.entries[key]; found {
		c.removeElement(elem)
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lruList = list.New()
	c.totalBytes = 0
	c.evictions += int64(len(c.entries))
}

// Stats returns cache statistics
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return Stats{
		Hits:       c.hits,
		Misses:     c.misses,
		Evictions:  c.evictions,
		Size:       c.lruList.Len(),
		MaxSize:    c.maxEntries,
		TotalBytes: c.totalBytes,
		HitRate:    hitRate,
	}
}

// evictOldest removes the least recently used entry
func (c *Cache) evictOldest() {
	elem := c.lruList.Back()
	if elem != nil {
		c.removeElement(elem)
		c.evictions++
	}
}

// removeElement removes an element from the cache
func (c *Cache) removeElement(elem *list.Element) {
	item := elem.Value.(*cacheItem)
	delete(c.entries, item.key)
	c.lruList.Remove(elem)
	c.totalBytes -= item.entry.Size
}

// CleanExpired removes all expired entries
func (c *Cache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	// Iterate through all entries
	var toRemove []*list.Element
	for elem := c.lruList.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*cacheItem)
		if now.After(item.entry.ExpiresAt) {
			toRemove = append(toRemove, elem)
		}
	}

	// Remove expired entries
	for _, elem := range toRemove {
		c.removeElement(elem)
		removed++
	}

	c.evictions += int64(removed)
	return removed
}

// GenerateKey generates a cache key from query and parameters
func GenerateKey(prefix string, params ...string) string {
	h := sha256.New()
	h.Write([]byte(prefix))
	for _, p := range params {
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}
