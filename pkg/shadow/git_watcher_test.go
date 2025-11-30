/*
# Module: pkg/shadow/git_watcher_test.go
Tests for git commit watcher functionality.

## Linked Modules
- [git_watcher](./git_watcher.go) - Git watcher implementation

## Tags
shadow, git, watch, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#git_watcher_test.go> a code:Module ;
    code:name "pkg/shadow/git_watcher_test.go" ;
    code:description "Tests for git commit watcher functionality" ;
    code:language "go" ;
    code:layer "test" ;
    code:linksTo <./git_watcher.go> ;
    code:tags "shadow", "git", "watch", "test" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDefaultGitWatchOptions tests that default options are reasonable
func TestDefaultGitWatchOptions(t *testing.T) {
	opts := DefaultGitWatchOptions()

	if opts.PollInterval != 2*time.Second {
		t.Errorf("Expected poll interval 2s, got %v", opts.PollInterval)
	}

	if opts.Debounce != 500*time.Millisecond {
		t.Errorf("Expected debounce 500ms, got %v", opts.Debounce)
	}

	if !opts.WatchHEAD {
		t.Error("Expected WatchHEAD to be true by default")
	}

	if opts.WatchUncommitted {
		t.Error("Expected WatchUncommitted to be false by default")
	}

	if !opts.SyncOnStart {
		t.Error("Expected SyncOnStart to be true by default")
	}
}

// TestNewGitWatcher tests creating a new git watcher
func TestNewGitWatcher(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create shadow file system
	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	// Create watcher
	watcher, err := NewGitWatcher(shadowFS, DefaultGitWatchOptions())
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	if watcher == nil {
		t.Fatal("Expected watcher to not be nil")
	}

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running initially")
	}
}

// TestNewGitWatcherNotGitRepo tests error when not a git repository
func TestNewGitWatcherNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create shadow file system (without git repo)
	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	// Create watcher - should fail
	_, err = NewGitWatcher(shadowFS, DefaultGitWatchOptions())
	if err == nil {
		t.Error("Expected error when not a git repository")
	}

	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("Expected 'not a git repository' error, got: %v", err)
	}
}

// TestGitWatcherStartStop tests starting and stopping the watcher
func TestGitWatcherStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	opts := DefaultGitWatchOptions()
	opts.SyncOnStart = false // Skip initial sync for faster tests

	watcher, err := NewGitWatcher(shadowFS, opts)
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	// Start watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	if !watcher.IsRunning() {
		t.Error("Expected watcher to be running after Start()")
	}

	// Try to start again - should fail
	if err := watcher.Start(); err == nil {
		t.Error("Expected error when starting already running watcher")
	}

	// Stop watcher
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}

	if watcher.IsRunning() {
		t.Error("Expected watcher to not be running after Stop()")
	}

	// Stop again should be no-op
	if err := watcher.Stop(); err != nil {
		t.Fatalf("Stop on stopped watcher should not error: %v", err)
	}
}

// TestGitWatcherStats tests watcher statistics
func TestGitWatcherStats(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	opts := DefaultGitWatchOptions()
	opts.SyncOnStart = false

	watcher, err := NewGitWatcher(shadowFS, opts)
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	// Start watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	// Get stats
	stats := watcher.Stats()

	if stats.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}

	if stats.CommitsDetected != 0 {
		t.Errorf("Expected 0 commits detected initially, got %d", stats.CommitsDetected)
	}
}

// TestGitWatcherTriggerSync tests manual sync trigger
func TestGitWatcherTriggerSync(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create a source file with LinkedDoc
	sourceFile := filepath.Join(tmpDir, "test.go")
	sourceContent := `/*
# Module: test.go
Test file for sync testing.

## Tags
test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> a code:Module ;
    code:name "test.go" .
<!-- End LinkedDoc RDF -->
*/
package main
`
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Commit the file
	runGitCmd(t, tmpDir, "add", "test.go")
	runGitCmd(t, tmpDir, "commit", "-m", "Add test file")

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	opts := DefaultGitWatchOptions()
	opts.SyncOnStart = false

	watcher, err := NewGitWatcher(shadowFS, opts)
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	// Trigger sync
	if err := watcher.TriggerSync(); err != nil {
		t.Fatalf("TriggerSync failed: %v", err)
	}

	// Verify shadow entry was created
	if !shadowFS.Exists(sourceFile) {
		t.Error("Expected shadow entry to be created after sync")
	}
}

// TestGitWatcherCallbacks tests callback functionality
func TestGitWatcherCallbacks(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	syncCalled := false
	errorCalled := false

	opts := DefaultGitWatchOptions()
	opts.SyncOnStart = false
	opts.OnSync = func(result *SyncResult) {
		syncCalled = true
	}
	opts.OnError = func(err error) {
		errorCalled = true
	}

	watcher, err := NewGitWatcher(shadowFS, opts)
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	// Trigger sync to invoke callback
	if err := watcher.TriggerSync(); err != nil {
		t.Fatalf("TriggerSync failed: %v", err)
	}

	if !syncCalled {
		t.Error("Expected OnSync callback to be called")
	}

	if errorCalled {
		t.Error("Did not expect OnError callback to be called")
	}
}

// TestGetHEADCommit tests getting the current HEAD commit
func TestGetHEADCommit(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	commit, err := getHEADCommit(tmpDir)
	if err != nil {
		t.Fatalf("getHEADCommit failed: %v", err)
	}

	if commit == "" {
		t.Error("Expected non-empty commit hash")
	}

	// Git commit hashes are 40 characters
	if len(commit) != 40 {
		t.Errorf("Expected 40 character commit hash, got %d characters", len(commit))
	}
}

// TestGetHEADRef tests getting the current HEAD reference
func TestGetHEADRef(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	ref, err := getHEADRef(tmpDir)
	if err != nil {
		t.Fatalf("getHEADRef failed: %v", err)
	}

	// Should be on main/master branch
	if ref != "main" && ref != "master" {
		t.Errorf("Expected ref to be main or master, got %s", ref)
	}
}

// TestGetRecentCommits tests getting recent commits
func TestGetRecentCommits(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create additional commits
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, "test"+string(rune('a'+i))+".txt")
		os.WriteFile(testFile, []byte("content"), 0644)
		runGitCmd(t, tmpDir, "add", ".")
		runGitCmd(t, tmpDir, "commit", "-m", "Commit "+string(rune('1'+i)))
	}

	commits, err := getRecentCommits(tmpDir, 5)
	if err != nil {
		t.Fatalf("getRecentCommits failed: %v", err)
	}

	// Should have at least 4 commits (initial + 3 new)
	if len(commits) < 4 {
		t.Errorf("Expected at least 4 commits, got %d", len(commits))
	}

	// Each commit should be 40 characters
	for _, commit := range commits {
		if len(commit) != 40 {
			t.Errorf("Expected 40 character commit hash, got %d characters", len(commit))
		}
	}
}

// TestInstallGitHook tests git hook installation
func TestInstallGitHook(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Install hook
	if err := InstallGitHook(tmpDir, "/usr/local/bin/graphfs"); err != nil {
		t.Fatalf("InstallGitHook failed: %v", err)
	}

	// Verify hook exists
	hookPath := filepath.Join(tmpDir, ".git", "hooks", "post-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("Expected post-commit hook to exist")
	}

	// Read hook content
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Failed to read hook: %v", err)
	}

	if !strings.Contains(string(content), "graphfs shadow") {
		t.Error("Expected hook to contain 'graphfs shadow'")
	}

	// Install again - should not duplicate
	if err := InstallGitHook(tmpDir, "/usr/local/bin/graphfs"); err != nil {
		t.Fatalf("Second InstallGitHook failed: %v", err)
	}

	content2, _ := os.ReadFile(hookPath)
	count := strings.Count(string(content2), "graphfs shadow")
	if count != 1 {
		t.Errorf("Expected exactly 1 graphfs command in hook, got %d", count)
	}
}

// TestUninstallGitHook tests git hook removal
func TestUninstallGitHook(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Install hook first
	if err := InstallGitHook(tmpDir, "/usr/local/bin/graphfs"); err != nil {
		t.Fatalf("InstallGitHook failed: %v", err)
	}

	// Uninstall hook
	if err := UninstallGitHook(tmpDir); err != nil {
		t.Fatalf("UninstallGitHook failed: %v", err)
	}

	// Hook file should be removed (since it only contained graphfs)
	hookPath := filepath.Join(tmpDir, ".git", "hooks", "post-commit")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("Expected post-commit hook to be removed")
	}
}

// TestUninstallGitHookWithOtherContent tests hook removal preserves other content
func TestUninstallGitHookWithOtherContent(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create hook with existing content
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	hookPath := filepath.Join(hooksDir, "post-commit")
	existingContent := "#!/bin/sh\necho 'Existing hook'\n"
	os.WriteFile(hookPath, []byte(existingContent), 0755)

	// Install graphfs hook
	if err := InstallGitHook(tmpDir, "/usr/local/bin/graphfs"); err != nil {
		t.Fatalf("InstallGitHook failed: %v", err)
	}

	// Uninstall graphfs hook
	if err := UninstallGitHook(tmpDir); err != nil {
		t.Fatalf("UninstallGitHook failed: %v", err)
	}

	// Hook should still exist with original content
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("Expected post-commit hook to still exist")
	}

	content, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(content), "Existing hook") {
		t.Error("Expected original hook content to be preserved")
	}

	if strings.Contains(string(content), "graphfs") {
		t.Error("Expected graphfs content to be removed")
	}
}

// TestInstallGitHookNotGitRepo tests error when not a git repository
func TestInstallGitHookNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	err := InstallGitHook(tmpDir, "/usr/local/bin/graphfs")
	if err == nil {
		t.Error("Expected error when not a git repository")
	}

	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("Expected 'not a git repository' error, got: %v", err)
	}
}

// TestSyncOnCommit tests syncing based on a specific commit
func TestSyncOnCommit(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create a source file with LinkedDoc
	sourceFile := filepath.Join(tmpDir, "module.go")
	sourceContent := `/*
# Module: module.go
Test module.

## Tags
test, module

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#module.go> a code:Module ;
    code:name "module.go" ;
    code:tags "test", "module" .
<!-- End LinkedDoc RDF -->
*/
package main
`
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Commit the file
	runGitCmd(t, tmpDir, "add", "module.go")
	runGitCmd(t, tmpDir, "commit", "-m", "Add module file")

	// Get the commit hash
	commitHash, err := getHEADCommit(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get HEAD commit: %v", err)
	}

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow FS: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow FS: %v", err)
	}

	opts := DefaultGitWatchOptions()
	opts.SyncOnStart = false

	watcher, err := NewGitWatcher(shadowFS, opts)
	if err != nil {
		t.Fatalf("Failed to create git watcher: %v", err)
	}

	// Sync based on the commit
	result, err := watcher.SyncOnCommit(commitHash)
	if err != nil {
		t.Fatalf("SyncOnCommit failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify shadow entry was created
	if !shadowFS.Exists(sourceFile) {
		t.Error("Expected shadow entry to be created after sync")
	}
}

// Helper functions

// initGitRepo initializes a git repository in the given directory
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	runGitCmd(t, dir, "init")
	runGitCmd(t, dir, "config", "user.email", "test@test.com")
	runGitCmd(t, dir, "config", "user.name", "Test User")

	// Create initial commit
	initialFile := filepath.Join(dir, "README.md")
	os.WriteFile(initialFile, []byte("# Test Repo"), 0644)
	runGitCmd(t, dir, "add", ".")
	runGitCmd(t, dir, "commit", "-m", "Initial commit")
}

// runGitCmd runs a git command in the given directory
func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}
}
