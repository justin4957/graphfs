/*
# Module: pkg/watch/watcher.go
File system watcher for live monitoring.

Monitors file system changes and triggers incremental graph updates with
debouncing to batch rapid changes.

## Linked Modules
- [debouncer](./debouncer.go) - Change debouncing
- [../graph](../graph/graph.go) - Graph updates
- [../scanner](../scanner/scanner.go) - File scanning
- [../parser](../parser/parser.go) - File parsing

## Tags
watch, filesystem, monitoring

## Exports
Watcher, WatchOptions, NewWatcher

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#watcher.go> a code:Module ;
    code:name "pkg/watch/watcher.go" ;
    code:description "File system watcher for live monitoring" ;
    code:language "go" ;
    code:layer "watch" ;
    code:linksTo <./debouncer.go>, <../graph/graph.go>, <../scanner/scanner.go>, <../parser/parser.go> ;
    code:exports <#Watcher>, <#WatchOptions>, <#NewWatcher> ;
    code:tags "watch", "filesystem", "monitoring" .
<!-- End LinkedDoc RDF -->
*/

package watch

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
)

// WatchOptions configures watch behavior
type WatchOptions struct {
	Path           string        // Root path to watch
	Debounce       time.Duration // Debounce duration for batching changes
	IgnorePatterns []string      // Patterns to ignore
	Verbose        bool          // Enable verbose logging
}

// DefaultWatchOptions returns default watch options
func DefaultWatchOptions() WatchOptions {
	return WatchOptions{
		Path:     ".",
		Debounce: 300 * time.Millisecond,
		IgnorePatterns: []string{
			".git",
			".graphfs",
			"node_modules",
			"vendor",
			".idea",
			".vscode",
		},
		Verbose: false,
	}
}

// Watcher monitors file system changes and updates the graph
type Watcher struct {
	watcher   *fsnotify.Watcher
	graph     *graph.Graph
	scanner   *scanner.Scanner
	debouncer *Debouncer
	onChange  func(*graph.Graph, []string) // Callback with changed files
	opts      WatchOptions
	mu        sync.Mutex
	running   bool
	changes   map[string]bool // Track pending changes
}

// NewWatcher creates a new file system watcher
func NewWatcher(g *graph.Graph, opts WatchOptions, onChange func(*graph.Graph, []string)) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		watcher:   watcher,
		graph:     g,
		scanner:   scanner.NewScanner(),
		debouncer: NewDebouncer(opts.Debounce),
		onChange:  onChange,
		opts:      opts,
		changes:   make(map[string]bool),
	}

	// Watch directory recursively
	if err := w.watchRecursive(opts.Path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to setup watches: %w", err)
	}

	return w, nil
}

// watchRecursive adds watches to all directories recursively
func (w *Watcher) watchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip ignored directories
		baseName := filepath.Base(path)
		for _, pattern := range w.opts.IgnorePatterns {
			if strings.Contains(path, pattern) || baseName == pattern {
				if w.opts.Verbose {
					log.Printf("Skipping ignored directory: %s", path)
				}
				return filepath.SkipDir
			}
		}

		// Skip hidden directories
		if strings.HasPrefix(baseName, ".") && baseName != "." {
			if w.opts.Verbose {
				log.Printf("Skipping hidden directory: %s", path)
			}
			return filepath.SkipDir
		}

		if err := w.watcher.Add(path); err != nil {
			return fmt.Errorf("failed to watch %s: %w", path, err)
		}

		if w.opts.Verbose {
			log.Printf("Watching: %s", path)
		}

		return nil
	})
}

// Start begins monitoring for file changes
func (w *Watcher) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				if w.shouldProcess(event) {
					w.trackChange(event.Name)
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watch error: %v", err)
			}
		}
	}()
}

// shouldProcess determines if an event should trigger processing
func (w *Watcher) shouldProcess(event fsnotify.Event) bool {
	// Only process write and create events
	if event.Op&fsnotify.Write != fsnotify.Write &&
		event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}

	// Check if it's a supported file type
	lang := scanner.DetectLanguage(event.Name)
	if lang == "unknown" {
		return false
	}

	// Check if file should be ignored
	for _, pattern := range w.opts.IgnorePatterns {
		if strings.Contains(event.Name, pattern) {
			return false
		}
	}

	return true
}

// trackChange records a file change and triggers debounced processing
func (w *Watcher) trackChange(path string) {
	w.mu.Lock()
	w.changes[path] = true
	w.mu.Unlock()

	w.debouncer.Trigger(func() {
		w.processChanges()
	})
}

// processChanges handles all pending file changes
func (w *Watcher) processChanges() {
	w.mu.Lock()
	changedFiles := make([]string, 0, len(w.changes))
	for path := range w.changes {
		changedFiles = append(changedFiles, path)
	}
	w.changes = make(map[string]bool) // Clear changes
	w.mu.Unlock()

	if len(changedFiles) == 0 {
		return
	}

	if w.opts.Verbose {
		log.Printf("Processing %d changed file(s)", len(changedFiles))
	}

	// Re-scan changed files and update graph
	for _, path := range changedFiles {
		if err := w.updateFile(path); err != nil {
			log.Printf("Failed to update %s: %v", path, err)
		}
	}

	// Notify callback
	if w.onChange != nil {
		w.onChange(w.graph, changedFiles)
	}
}

// updateFile re-parses a file and updates the graph
func (w *Watcher) updateFile(path string) error {
	// Check if file still exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File was deleted - remove from graph
		relPath, _ := filepath.Rel(w.opts.Path, path)
		w.graph.RemoveModule(relPath)
		if w.opts.Verbose {
			log.Printf("Removed deleted file: %s", relPath)
		}
		return nil
	}

	// Re-scan the file
	fileInfo, err := w.scanner.ScanFile(path)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Only update if file has LinkedDoc
	if !fileInfo.HasLinkedDoc {
		if w.opts.Verbose {
			log.Printf("Skipping file without LinkedDoc: %s", path)
		}
		return nil
	}

	// Get relative path
	relPath, err := filepath.Rel(w.opts.Path, path)
	if err != nil {
		relPath = path
	}

	if w.opts.Verbose {
		log.Printf("Updated: %s", relPath)
	}

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	w.debouncer.Stop()

	return w.watcher.Close()
}

// IsRunning returns true if the watcher is running
func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}
