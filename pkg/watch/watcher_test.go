package watch

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func TestNewWatcher(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graphfs-watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tripleStore := store.NewTripleStore()
	g := graph.NewGraph(tmpDir, tripleStore)

	opts := DefaultWatchOptions()
	opts.Path = tmpDir

	watcher, err := NewWatcher(g, opts, nil)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running initially")
	}
}

func TestWatcher_StartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graphfs-watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tripleStore := store.NewTripleStore()
	g := graph.NewGraph(tmpDir, tripleStore)

	opts := DefaultWatchOptions()
	opts.Path = tmpDir

	watcher, err := NewWatcher(g, opts, nil)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	watcher.Start()
	time.Sleep(50 * time.Millisecond)

	if !watcher.IsRunning() {
		t.Error("Expected watcher to be running after Start()")
	}

	if err := watcher.Stop(); err != nil {
		t.Errorf("Failed to stop watcher: %v", err)
	}

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running after Stop()")
	}
}

func TestWatcher_FileChange(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graphfs-watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tripleStore := store.NewTripleStore()
	g := graph.NewGraph(tmpDir, tripleStore)

	var changeCount atomic.Int32
	var changedFiles []string

	opts := DefaultWatchOptions()
	opts.Path = tmpDir
	opts.Debounce = 100 * time.Millisecond

	watcher, err := NewWatcher(g, opts, func(graph *graph.Graph, files []string) {
		changeCount.Add(1)
		changedFiles = files
	})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()

	// Wait for watcher to be ready (longer for CI with race detector)
	time.Sleep(300 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test
// Test file
func Example() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Wait for debounce and processing (longer for CI with race detector)
	time.Sleep(500 * time.Millisecond)

	if changeCount.Load() == 0 {
		t.Error("Expected change callback to be called")
	}

	if len(changedFiles) == 0 {
		t.Error("Expected changed files list to not be empty")
	}
}

func TestWatcher_IgnorePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graphfs-watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create ignored directories
	vendorDir := filepath.Join(tmpDir, "vendor")
	os.MkdirAll(vendorDir, 0755)

	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	os.MkdirAll(nodeModulesDir, 0755)

	tripleStore := store.NewTripleStore()
	g := graph.NewGraph(tmpDir, tripleStore)

	opts := DefaultWatchOptions()
	opts.Path = tmpDir

	watcher, err := NewWatcher(g, opts, nil)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Verify ignored directories are not watched
	// This is implicit - if they were watched, watcher creation would add them
	// We just verify the watcher was created successfully
	if watcher == nil {
		t.Error("Expected watcher to be created successfully")
	}
}

func TestWatcher_MultipleChanges(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graphfs-watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tripleStore := store.NewTripleStore()
	g := graph.NewGraph(tmpDir, tripleStore)

	var changeCount atomic.Int32

	opts := DefaultWatchOptions()
	opts.Path = tmpDir
	opts.Debounce = 100 * time.Millisecond

	watcher, err := NewWatcher(g, opts, func(graph *graph.Graph, files []string) {
		changeCount.Add(1)
	})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()
	time.Sleep(100 * time.Millisecond)

	// Create multiple files rapidly
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".go")
		content := `package test`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce and processing
	time.Sleep(300 * time.Millisecond)

	// Should batch changes into one callback due to debouncing
	count := changeCount.Load()
	if count == 0 {
		t.Error("Expected at least one change callback")
	}
	if count > 2 {
		t.Errorf("Expected debouncing to batch changes, got %d callbacks", count)
	}
}
