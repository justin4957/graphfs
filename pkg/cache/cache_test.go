package cache

import (
	"testing"
	"time"
)

func TestCacheBasicOperations(t *testing.T) {
	c := NewCache(10, 5*time.Minute)

	// Test Set and Get
	c.Set("key1", "value1", 10)
	val, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test Get non-existent key
	_, found = c.Get("nonexistent")
	if found {
		t.Error("Should not find nonexistent key")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	c := NewCache(3, 5*time.Minute)

	// Fill cache
	c.Set("key1", "value1", 10)
	c.Set("key2", "value2", 10)
	c.Set("key3", "value3", 10)

	// Add one more - should evict key1
	c.Set("key4", "value4", 10)

	// key1 should be evicted
	_, found := c.Get("key1")
	if found {
		t.Error("key1 should have been evicted")
	}

	// key2, key3, key4 should exist
	if _, found := c.Get("key2"); !found {
		t.Error("key2 should exist")
	}
	if _, found := c.Get("key3"); !found {
		t.Error("key3 should exist")
	}
	if _, found := c.Get("key4"); !found {
		t.Error("key4 should exist")
	}
}

func TestCacheTTL(t *testing.T) {
	c := NewCache(10, 100*time.Millisecond)

	c.Set("key1", "value1", 10)

	// Should exist immediately
	_, found := c.Get("key1")
	if !found {
		t.Error("key1 should exist")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = c.Get("key1")
	if found {
		t.Error("key1 should have expired")
	}
}

func TestCacheUpdate(t *testing.T) {
	c := NewCache(10, 5*time.Minute)

	c.Set("key1", "value1", 10)
	c.Set("key1", "value2", 10)

	val, found := c.Get("key1")
	if !found {
		t.Error("key1 should exist")
	}
	if val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}

	// Cache size should still be 1
	stats := c.Stats()
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
}

func TestCacheDelete(t *testing.T) {
	c := NewCache(10, 5*time.Minute)

	c.Set("key1", "value1", 10)
	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("key1 should have been deleted")
	}
}

func TestCacheClear(t *testing.T) {
	c := NewCache(10, 5*time.Minute)

	c.Set("key1", "value1", 10)
	c.Set("key2", "value2", 10)
	c.Set("key3", "value3", 10)

	c.Clear()

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0, got %d", stats.Size)
	}
}

func TestCacheStats(t *testing.T) {
	c := NewCache(10, 5*time.Minute)

	c.Set("key1", "value1", 100)
	c.Set("key2", "value2", 200)

	// Hit
	c.Get("key1")
	c.Get("key1")

	// Miss
	c.Get("nonexistent")

	stats := c.Stats()

	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
	if stats.TotalBytes != 300 {
		t.Errorf("Expected 300 bytes, got %d", stats.TotalBytes)
	}
	if stats.HitRate < 0.66 || stats.HitRate > 0.67 {
		t.Errorf("Expected hit rate ~0.67, got %f", stats.HitRate)
	}
}

func TestCacheCleanExpired(t *testing.T) {
	c := NewCache(10, 100*time.Millisecond)

	c.Set("key1", "value1", 10)
	c.Set("key2", "value2", 10)
	c.Set("key3", "value3", 10)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	removed := c.CleanExpired()

	if removed != 3 {
		t.Errorf("Expected 3 removed, got %d", removed)
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0, got %d", stats.Size)
	}
}

func TestCacheHitUpdatesLRU(t *testing.T) {
	c := NewCache(3, 5*time.Minute)

	c.Set("key1", "value1", 10)
	c.Set("key2", "value2", 10)
	c.Set("key3", "value3", 10)

	// Access key1 to move it to front
	c.Get("key1")

	// Add key4 - should evict key2 (least recently used)
	c.Set("key4", "value4", 10)

	// key2 should be evicted
	if _, found := c.Get("key2"); found {
		t.Error("key2 should have been evicted")
	}

	// key1, key3, key4 should exist
	if _, found := c.Get("key1"); !found {
		t.Error("key1 should exist")
	}
	if _, found := c.Get("key3"); !found {
		t.Error("key3 should exist")
	}
	if _, found := c.Get("key4"); !found {
		t.Error("key4 should exist")
	}
}

func TestGenerateKey(t *testing.T) {
	key1 := GenerateKey("query", "param1", "param2")
	key2 := GenerateKey("query", "param1", "param2")
	key3 := GenerateKey("query", "param1", "different")

	if key1 != key2 {
		t.Error("Same inputs should generate same key")
	}

	if key1 == key3 {
		t.Error("Different inputs should generate different keys")
	}

	if len(key1) != 64 {
		t.Errorf("Expected 64 character key, got %d", len(key1))
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := NewCache(100, 5*time.Minute)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			c.Set("key", i, 10)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			c.Get("key")
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic due to race conditions
}

func BenchmarkCacheSet(b *testing.B) {
	c := NewCache(10000, 5*time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Set("key", "value", 10)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := NewCache(10000, 5*time.Minute)
	c.Set("key", "value", 10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}
