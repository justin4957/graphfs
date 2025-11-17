/*
# Module: pkg/cache/manager.go
Persistent cache manager for knowledge graph modules.

Provides disk-based caching with content-based invalidation to avoid
re-parsing unchanged files on subsequent scans.

## Linked Modules
- [cache](./cache.go) - In-memory caching
- [graph/module](../graph/module.go) - Module data structure

## Tags
cache, persistence, performance

## Exports
Manager, NewManager, CacheStats

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#manager.go> a code:Module ;
    code:name "pkg/cache/manager.go" ;
    code:description "Persistent cache manager for knowledge graph modules" ;
    code:language "go" ;
    code:layer "cache" ;
    code:linksTo <./cache.go>, <../graph/module.go> ;
    code:exports <#Manager>, <#NewManager>, <#CacheStats> ;
    code:tags "cache", "persistence", "performance" .
<!-- End LinkedDoc RDF -->
*/

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	cacheVersion     = "v1"
	metadataBucket   = "metadata"
	modulesBucket    = "modules"
	fileHashesBucket = "file_hashes"
	defaultCacheDir  = ".graphfs/cache"
)

// CacheStats represents cache statistics
type CacheStats struct {
	ModuleCount int
	CacheHits   int64
	CacheMisses int64
	CacheSize   int64 // Total bytes
	HitRate     float64
	LastUpdated time.Time
}

// CachedModule wraps a module with cache metadata
type CachedModule struct {
	Module      interface{} // Stored as JSON, can be any module type
	FileHash    string
	CachedAt    time.Time
	FileModTime time.Time
}

// Manager handles persistent caching of parsed modules
type Manager struct {
	db       *bolt.DB
	root     string
	cacheDir string
	hits     int64
	misses   int64
}

// NewManager creates a new cache manager
func NewManager(root string) (*Manager, error) {
	cacheDir := filepath.Join(root, defaultCacheDir)

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Open BoltDB database
	dbPath := filepath.Join(cacheDir, "modules.db")
	db, err := bolt.Open(dbPath, 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	// Initialize buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(metadataBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(modulesBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(fileHashesBucket)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize cache buckets: %w", err)
	}

	manager := &Manager{
		db:       db,
		root:     root,
		cacheDir: cacheDir,
	}

	// Store cache version
	if err := manager.setMetadata("version", cacheVersion); err != nil {
		db.Close()
		return nil, err
	}

	return manager, nil
}

// Get retrieves a cached module if it's still valid
// Returns the module as JSON bytes that need to be unmarshaled by the caller
func (m *Manager) Get(filePath string) ([]byte, bool) {
	// Calculate current file hash
	currentHash, modTime, err := m.calculateFileHash(filePath)
	if err != nil {
		m.misses++
		return nil, false
	}

	// Check cache
	var cached CachedModule
	err = m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(modulesBucket))
		data := bucket.Get([]byte(filePath))
		if data == nil {
			return fmt.Errorf("not found")
		}

		return json.Unmarshal(data, &cached)
	})

	if err != nil {
		m.misses++
		return nil, false
	}

	// Validate cache: check if file hash matches and mtime hasn't changed
	if cached.FileHash != currentHash || !cached.FileModTime.Equal(modTime) {
		m.misses++
		return nil, false
	}

	m.hits++

	// Return the module as JSON bytes
	moduleBytes, err := json.Marshal(cached.Module)
	if err != nil {
		m.misses++
		m.hits-- // Undo the hit count
		return nil, false
	}

	return moduleBytes, true
}

// Set stores a module in the cache
// The module parameter must be JSON-serializable
func (m *Manager) Set(filePath string, module interface{}) error {
	// Calculate file hash
	fileHash, modTime, err := m.calculateFileHash(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}

	cached := CachedModule{
		Module:      module,
		FileHash:    fileHash,
		CachedAt:    time.Now(),
		FileModTime: modTime,
	}

	// Serialize module
	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("failed to serialize module: %w", err)
	}

	// Store in database
	return m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(modulesBucket))
		if err := bucket.Put([]byte(filePath), data); err != nil {
			return err
		}

		// Store file hash separately for quick lookups
		hashBucket := tx.Bucket([]byte(fileHashesBucket))
		return hashBucket.Put([]byte(filePath), []byte(fileHash))
	})
}

// Invalidate removes a module from the cache
func (m *Manager) Invalidate(filePath string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(modulesBucket))
		if err := bucket.Delete([]byte(filePath)); err != nil {
			return err
		}

		hashBucket := tx.Bucket([]byte(fileHashesBucket))
		return hashBucket.Delete([]byte(filePath))
	})
}

// Clear removes all cached modules
func (m *Manager) Clear() error {
	return m.db.Update(func(tx *bolt.Tx) error {
		// Delete and recreate buckets
		if err := tx.DeleteBucket([]byte(modulesBucket)); err != nil {
			// Ignore if bucket doesn't exist
			if err.Error() != "bucket not found" {
				return err
			}
		}
		if err := tx.DeleteBucket([]byte(fileHashesBucket)); err != nil {
			// Ignore if bucket doesn't exist
			if err.Error() != "bucket not found" {
				return err
			}
		}

		if _, err := tx.CreateBucket([]byte(modulesBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucket([]byte(fileHashesBucket)); err != nil {
			return err
		}

		return nil
	})
}

// Stats returns cache statistics
func (m *Manager) Stats() (CacheStats, error) {
	stats := CacheStats{
		CacheHits:   m.hits,
		CacheMisses: m.misses,
	}

	// Calculate hit rate
	total := m.hits + m.misses
	if total > 0 {
		stats.HitRate = float64(m.hits) / float64(total)
	}

	// Count modules and calculate size
	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(modulesBucket))

		stats.ModuleCount = bucket.Stats().KeyN

		// Calculate total size
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			stats.CacheSize += int64(len(v))
		}

		return nil
	})

	if err != nil {
		return stats, err
	}

	stats.LastUpdated = time.Now()
	return stats, nil
}

// Close closes the cache database
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// calculateFileHash computes SHA256 hash of file contents
func (m *Manager) calculateFileHash(filePath string) (string, time.Time, error) {
	// Get file info for modification time
	info, err := os.Stat(filePath)
	if err != nil {
		return "", time.Time{}, err
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", time.Time{}, err
	}
	defer file.Close()

	// Calculate hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", time.Time{}, err
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	return hashStr, info.ModTime(), nil
}

// setMetadata stores metadata in the cache
func (m *Manager) setMetadata(key, value string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(metadataBucket))
		return bucket.Put([]byte(key), []byte(value))
	})
}

// IsEnabled checks if caching is enabled (database is open)
func (m *Manager) IsEnabled() bool {
	return m.db != nil
}
