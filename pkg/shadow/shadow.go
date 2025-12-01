/*
# Module: pkg/shadow/shadow.go
Shadow file system for storing graph metadata separately from source code.

Provides a parallel directory structure that mirrors the codebase, storing
semantic/conceptual graph metadata without altering actual source files.

## Linked Modules
- [entry](./entry.go) - Shadow entry data structure
- [manager](./manager.go) - Shadow file system manager
- [index](./index.go) - Shadow index for fast lookups

## Tags
shadow, metadata, filesystem, non-invasive

## Exports
ShadowFS, NewShadowFS, Config

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#shadow.go> a code:Module ;
    code:name "pkg/shadow/shadow.go" ;
    code:description "Shadow file system for storing graph metadata separately from source code" ;
    code:language "go" ;
    code:layer "shadow" ;
    code:linksTo <./entry.go>, <./manager.go>, <./index.go> ;
    code:exports <#ShadowFS>, <#NewShadowFS>, <#Config> ;
    code:tags "shadow", "metadata", "filesystem", "non-invasive" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultShadowDir is the default directory name for shadow files
	DefaultShadowDir = ".graphfs/shadow"

	// ShadowVersion is the current shadow file format version
	ShadowVersion = "1.0"

	// ShadowExtension is the file extension for shadow files
	ShadowExtension = ".shadow.json"
)

// Config configures the shadow file system behavior
type Config struct {
	// ShadowDir is the root directory for shadow files (default: .graphfs/shadow)
	ShadowDir string

	// AutoSync enables automatic sync when source files change
	AutoSync bool

	// PreserveManual prevents overwriting manual annotations during auto-sync
	PreserveManual bool

	// CompactJSON uses compact JSON encoding (no indentation)
	CompactJSON bool

	// ValidateOnWrite validates entries before writing
	ValidateOnWrite bool
}

// DefaultConfig returns the default shadow configuration
func DefaultConfig() Config {
	return Config{
		ShadowDir:       DefaultShadowDir,
		AutoSync:        false,
		PreserveManual:  true,
		CompactJSON:     false,
		ValidateOnWrite: true,
	}
}

// ShadowFS manages the shadow file system for storing graph metadata
// It provides a non-invasive way to annotate codebases with semantic metadata
type ShadowFS struct {
	// Root path of the project
	rootPath string

	// Shadow directory path (absolute)
	shadowPath string

	// Configuration
	config Config

	// Index for fast lookups
	index *Index

	// Statistics
	stats Statistics

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Statistics tracks shadow file system usage
type Statistics struct {
	TotalEntries   int
	ManualEntries  int
	AutoGenEntries int
	TotalTriples   int
	LastSync       time.Time
	SyncCount      int
	LastBuildTime  time.Duration
}

// NewShadowFS creates a new shadow file system manager
func NewShadowFS(rootPath string, config Config) (*ShadowFS, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Determine shadow directory
	shadowDir := config.ShadowDir
	if shadowDir == "" {
		shadowDir = DefaultShadowDir
	}

	// Make shadow path absolute
	var shadowPath string
	if filepath.IsAbs(shadowDir) {
		shadowPath = shadowDir
	} else {
		shadowPath = filepath.Join(absRoot, shadowDir)
	}

	shadowFS := &ShadowFS{
		rootPath:   absRoot,
		shadowPath: shadowPath,
		config:     config,
		index:      NewIndex(),
	}

	return shadowFS, nil
}

// Initialize creates the shadow directory structure
func (s *ShadowFS) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create shadow root directory
	if err := os.MkdirAll(s.shadowPath, 0755); err != nil {
		return fmt.Errorf("failed to create shadow directory: %w", err)
	}

	// Create index file if it doesn't exist
	indexPath := filepath.Join(s.shadowPath, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		if err := s.index.Save(indexPath); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// RootPath returns the project root path
func (s *ShadowFS) RootPath() string {
	return s.rootPath
}

// ShadowPath returns the shadow directory path
func (s *ShadowFS) ShadowPath() string {
	return s.shadowPath
}

// Config returns the current configuration
func (s *ShadowFS) Config() Config {
	return s.config
}

// GetShadowPath converts a source file path to its shadow file path
func (s *ShadowFS) GetShadowPath(sourcePath string) (string, error) {
	// Get relative path from root
	relPath, err := s.getRelativePath(sourcePath)
	if err != nil {
		return "", err
	}

	// Construct shadow file path
	shadowFile := filepath.Join(s.shadowPath, relPath+ShadowExtension)
	return shadowFile, nil
}

// GetSourcePath converts a shadow file path back to its source file path
func (s *ShadowFS) GetSourcePath(shadowPath string) (string, error) {
	// Get relative path from shadow root
	relPath, err := filepath.Rel(s.shadowPath, shadowPath)
	if err != nil {
		return "", fmt.Errorf("path is not under shadow directory: %w", err)
	}

	// Remove shadow extension
	if filepath.Ext(relPath) == ".json" {
		baseName := relPath[:len(relPath)-len(".json")]
		if filepath.Ext(baseName) == ".shadow" {
			relPath = baseName[:len(baseName)-len(".shadow")]
		}
	}

	// Construct source file path
	sourcePath := filepath.Join(s.rootPath, relPath)
	return sourcePath, nil
}

// Get retrieves a shadow entry for a source file
func (s *ShadowFS) Get(sourcePath string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return nil, err
	}

	return LoadEntry(shadowPath)
}

// Set stores a shadow entry for a source file
func (s *ShadowFS) Set(sourcePath string, entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return err
	}

	// Ensure directory exists
	shadowDir := filepath.Dir(shadowPath)
	if err := os.MkdirAll(shadowDir, 0755); err != nil {
		return fmt.Errorf("failed to create shadow directory: %w", err)
	}

	// Validate if enabled
	if s.config.ValidateOnWrite {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Save entry
	if err := entry.Save(shadowPath, !s.config.CompactJSON); err != nil {
		return err
	}

	// Update index
	relPath, _ := s.getRelativePath(sourcePath)
	s.index.Add(relPath, entry)

	return nil
}

// Delete removes a shadow entry for a source file
func (s *ShadowFS) Delete(sourcePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return err
	}

	// Remove from index
	relPath, _ := s.getRelativePath(sourcePath)
	s.index.Remove(relPath)

	// Delete shadow file
	if err := os.Remove(shadowPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete shadow file: %w", err)
	}

	return nil
}

// Exists checks if a shadow entry exists for a source file
func (s *ShadowFS) Exists(sourcePath string) bool {
	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return false
	}

	_, err = os.Stat(shadowPath)
	return err == nil
}

// List returns all shadow entries
func (s *ShadowFS) List() ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entries []*Entry

	err := filepath.Walk(s.shadowPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-shadow files
		if info.IsDir() || !isShadowFile(path) {
			return nil
		}

		entry, err := LoadEntry(path)
		if err != nil {
			// Log warning but continue
			return nil
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list shadow entries: %w", err)
	}

	return entries, nil
}

// Merge combines an existing shadow entry with new data
// Manual annotations are preserved if PreserveManual is enabled
func (s *ShadowFS) Merge(sourcePath string, newEntry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return err
	}

	// Try to load existing entry
	existing, err := LoadEntry(shadowPath)
	if err != nil {
		// No existing entry, just save the new one
		return s.setUnlocked(sourcePath, newEntry)
	}

	// Merge entries
	merged := existing.Merge(newEntry, s.config.PreserveManual)

	// Save merged entry
	return s.setUnlocked(sourcePath, merged)
}

// setUnlocked saves an entry without acquiring the lock (caller must hold lock)
func (s *ShadowFS) setUnlocked(sourcePath string, entry *Entry) error {
	shadowPath, err := s.GetShadowPath(sourcePath)
	if err != nil {
		return err
	}

	// Ensure directory exists
	shadowDir := filepath.Dir(shadowPath)
	if err := os.MkdirAll(shadowDir, 0755); err != nil {
		return fmt.Errorf("failed to create shadow directory: %w", err)
	}

	// Validate if enabled
	if s.config.ValidateOnWrite {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Save entry
	if err := entry.Save(shadowPath, !s.config.CompactJSON); err != nil {
		return err
	}

	// Update index
	relPath, _ := s.getRelativePath(sourcePath)
	s.index.Add(relPath, entry)

	return nil
}

// Statistics returns current shadow file system statistics
func (s *ShadowFS) Statistics() Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.stats
}

// RefreshStatistics recalculates statistics by scanning shadow files
func (s *ShadowFS) RefreshStatistics() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.listUnlocked()
	if err != nil {
		return err
	}

	s.stats = Statistics{}
	s.stats.TotalEntries = len(entries)

	for _, entry := range entries {
		if entry.Source == SourceManual {
			s.stats.ManualEntries++
		} else {
			s.stats.AutoGenEntries++
		}
		s.stats.TotalTriples += len(entry.Triples)
	}

	return nil
}

// listUnlocked returns all entries without acquiring lock (caller must hold lock)
func (s *ShadowFS) listUnlocked() ([]*Entry, error) {
	var entries []*Entry

	err := filepath.Walk(s.shadowPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-shadow files
		if info.IsDir() || !isShadowFile(path) {
			return nil
		}

		entry, err := LoadEntry(path)
		if err != nil {
			return nil
		}

		entries = append(entries, entry)
		return nil
	})

	return entries, err
}

// Index returns the shadow file system index
func (s *ShadowFS) Index() *Index {
	return s.index
}

// LoadIndex loads the index from disk
func (s *ShadowFS) LoadIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	indexPath := filepath.Join(s.shadowPath, "index.json")
	return s.index.Load(indexPath)
}

// SaveIndex saves the index to disk
func (s *ShadowFS) SaveIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	indexPath := filepath.Join(s.shadowPath, "index.json")
	return s.index.Save(indexPath)
}

// RebuildIndex scans all shadow files and rebuilds the index
func (s *ShadowFS) RebuildIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing index
	s.index = NewIndex()

	// Walk shadow directory
	err := filepath.Walk(s.shadowPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-shadow files
		if info.IsDir() || !isShadowFile(path) {
			return nil
		}

		// Load entry
		entry, err := LoadEntry(path)
		if err != nil {
			return nil // Skip invalid entries
		}

		// Add to index
		s.index.Add(entry.SourcePath, entry)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	// Save rebuilt index
	indexPath := filepath.Join(s.shadowPath, "index.json")
	return s.index.Save(indexPath)
}

// getRelativePath gets the relative path from root
func (s *ShadowFS) getRelativePath(sourcePath string) (string, error) {
	// Make source path absolute if needed
	absSource := sourcePath
	if !filepath.IsAbs(sourcePath) {
		absSource = filepath.Join(s.rootPath, sourcePath)
	}

	// Get relative path
	relPath, err := filepath.Rel(s.rootPath, absSource)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return relPath, nil
}

// isShadowFile checks if a file is a shadow file
func isShadowFile(path string) bool {
	return filepath.Ext(path) == ".json" &&
		len(path) > len(ShadowExtension) &&
		path[len(path)-len(ShadowExtension):] == ShadowExtension
}

// Close performs cleanup operations
func (s *ShadowFS) Close() error {
	return s.SaveIndex()
}
