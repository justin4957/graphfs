/*
# Module: pkg/shadow/git_watcher.go
Git commit watcher for automatic shadow file system updates.

Monitors git commits and automatically updates the shadow file system
when commits are made, running in the background.

## Linked Modules
- [shadow](./shadow.go) - Shadow file system manager
- [builder](./builder.go) - Shadow builder
- [../scanner/git_filter](../scanner/git_filter.go) - Git filter

## Tags
shadow, git, watch, background, sync

## Exports
GitWatcher, NewGitWatcher, GitWatchOptions

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#git_watcher.go> a code:Module ;
    code:name "pkg/shadow/git_watcher.go" ;
    code:description "Git commit watcher for automatic shadow file system updates" ;
    code:language "go" ;
    code:layer "shadow" ;
    code:linksTo <./shadow.go>, <./builder.go>, <../scanner/git_filter.go> ;
    code:exports <#GitWatcher>, <#NewGitWatcher>, <#GitWatchOptions> ;
    code:tags "shadow", "git", "watch", "background", "sync" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/justin4957/graphfs/pkg/scanner"
)

// GitWatchOptions configures git watching behavior
type GitWatchOptions struct {
	// PollInterval is how often to check for new commits (default: 2s)
	PollInterval time.Duration

	// Debounce is the delay before processing after detecting changes (default: 500ms)
	Debounce time.Duration

	// Verbose enables detailed logging
	Verbose bool

	// OnCommit is called when a commit is detected
	OnCommit func(commit string, files []string)

	// OnSync is called after shadow sync completes
	OnSync func(result *SyncResult)

	// OnError is called when an error occurs
	OnError func(err error)

	// BuildOptions for shadow building
	BuildOptions BuildOptions

	// WatchHEAD watches for HEAD changes (commits, checkouts, merges)
	WatchHEAD bool

	// WatchUncommitted also watches for uncommitted changes
	WatchUncommitted bool

	// SyncOnStart performs an initial sync when starting
	SyncOnStart bool
}

// DefaultGitWatchOptions returns default git watch options
func DefaultGitWatchOptions() GitWatchOptions {
	return GitWatchOptions{
		PollInterval:     2 * time.Second,
		Debounce:         500 * time.Millisecond,
		Verbose:          false,
		WatchHEAD:        true,
		WatchUncommitted: false,
		SyncOnStart:      true,
		BuildOptions:     DefaultBuildOptions(),
	}
}

// GitWatcher monitors git commits and updates the shadow file system
type GitWatcher struct {
	shadowFS  *ShadowFS
	builder   *Builder
	gitFilter *scanner.GitFilter
	opts      GitWatchOptions

	// State tracking
	lastCommit    string
	lastHeadRef   string
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	debounceTimer *time.Timer

	// Statistics
	stats GitWatchStats
}

// GitWatchStats tracks watcher statistics
type GitWatchStats struct {
	StartTime       time.Time
	CommitsDetected int
	SyncsPerformed  int
	FilesUpdated    int
	Errors          int
	LastSyncTime    time.Time
	LastCommit      string
}

// NewGitWatcher creates a new git commit watcher
func NewGitWatcher(shadowFS *ShadowFS, opts GitWatchOptions) (*GitWatcher, error) {
	gitFilter := scanner.NewGitFilter(shadowFS.RootPath())

	// Verify it's a git repository
	if !gitFilter.IsGitRepository() {
		return nil, fmt.Errorf("not a git repository: %s", shadowFS.RootPath())
	}

	// Get current HEAD commit
	currentCommit, err := getHEADCommit(shadowFS.RootPath())
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}

	// Get current HEAD ref
	headRef, _ := getHEADRef(shadowFS.RootPath())

	return &GitWatcher{
		shadowFS:    shadowFS,
		builder:     NewBuilder(shadowFS),
		gitFilter:   gitFilter,
		opts:        opts,
		lastCommit:  currentCommit,
		lastHeadRef: headRef,
	}, nil
}

// Start begins monitoring for git commits
func (w *GitWatcher) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.running = true
	w.stats.StartTime = time.Now()
	w.mu.Unlock()

	// Perform initial sync if enabled
	if w.opts.SyncOnStart {
		if w.opts.Verbose {
			log.Println("Performing initial shadow sync...")
		}
		if err := w.performSync(nil); err != nil {
			if w.opts.OnError != nil {
				w.opts.OnError(err)
			}
			if w.opts.Verbose {
				log.Printf("Initial sync error: %v", err)
			}
		}
	}

	// Start the polling loop
	go w.pollLoop()

	return nil
}

// Stop stops the git watcher
func (w *GitWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	if w.cancel != nil {
		w.cancel()
	}

	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}

	return nil
}

// IsRunning returns true if the watcher is running
func (w *GitWatcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// Stats returns current watcher statistics
func (w *GitWatcher) Stats() GitWatchStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stats
}

// TriggerSync manually triggers a shadow sync
func (w *GitWatcher) TriggerSync() error {
	return w.performSync(nil)
}

// pollLoop continuously checks for git changes
func (w *GitWatcher) pollLoop() {
	ticker := time.NewTicker(w.opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.checkForChanges()
		}
	}
}

// checkForChanges checks for new commits or HEAD changes
func (w *GitWatcher) checkForChanges() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	// Check for HEAD changes (commits, checkouts, merges)
	if w.opts.WatchHEAD {
		currentCommit, err := getHEADCommit(w.shadowFS.RootPath())
		if err != nil {
			if w.opts.OnError != nil {
				w.opts.OnError(err)
			}
			return
		}

		currentRef, _ := getHEADRef(w.shadowFS.RootPath())

		w.mu.Lock()
		commitChanged := currentCommit != w.lastCommit
		refChanged := currentRef != w.lastHeadRef
		w.mu.Unlock()

		if commitChanged || refChanged {
			if w.opts.Verbose {
				if commitChanged {
					log.Printf("New commit detected: %s", currentCommit[:8])
				}
				if refChanged {
					log.Printf("HEAD ref changed: %s -> %s", w.lastHeadRef, currentRef)
				}
			}

			w.mu.Lock()
			oldCommit := w.lastCommit
			w.lastCommit = currentCommit
			w.lastHeadRef = currentRef
			w.stats.CommitsDetected++
			w.stats.LastCommit = currentCommit
			w.mu.Unlock()

			// Get changed files
			var changedFiles []string
			if commitChanged && oldCommit != "" {
				changedFiles, _ = w.getChangedFilesSince(oldCommit)
			}

			// Notify callback
			if w.opts.OnCommit != nil {
				w.opts.OnCommit(currentCommit, changedFiles)
			}

			// Debounce the sync
			w.scheduleSync(changedFiles)
		}
	}

	// Check for uncommitted changes if enabled
	if w.opts.WatchUncommitted {
		uncommitted, err := w.gitFilter.UncommittedChanges()
		if err == nil && len(uncommitted) > 0 {
			// Only process if there are supported files
			supported := w.gitFilter.FilterSupported(uncommitted)
			if len(supported) > 0 {
				w.scheduleSync(supported)
			}
		}
	}
}

// scheduleSync schedules a debounced sync operation
func (w *GitWatcher) scheduleSync(changedFiles []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Cancel existing timer
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}

	// Schedule new sync
	w.debounceTimer = time.AfterFunc(w.opts.Debounce, func() {
		if err := w.performSync(changedFiles); err != nil {
			if w.opts.OnError != nil {
				w.opts.OnError(err)
			}
			if w.opts.Verbose {
				log.Printf("Sync error: %v", err)
			}
		}
	})
}

// performSync executes the shadow sync
func (w *GitWatcher) performSync(changedFiles []string) error {
	if w.opts.Verbose {
		if len(changedFiles) > 0 {
			log.Printf("Syncing %d changed files...", len(changedFiles))
		} else {
			log.Println("Performing full shadow sync...")
		}
	}

	var result *SyncResult
	var err error

	if len(changedFiles) > 0 {
		// Incremental sync for specific files
		result, err = w.incrementalSync(changedFiles)
	} else {
		// Full sync
		result, err = w.builder.Sync(w.opts.BuildOptions)
	}

	if err != nil {
		w.mu.Lock()
		w.stats.Errors++
		w.mu.Unlock()
		return err
	}

	// Update statistics
	w.mu.Lock()
	w.stats.SyncsPerformed++
	w.stats.LastSyncTime = time.Now()
	if result != nil && result.BuildResult != nil {
		w.stats.FilesUpdated += result.BuildResult.NewEntries + result.BuildResult.UpdatedEntries + result.BuildResult.MergedEntries
	}
	w.mu.Unlock()

	// Notify callback
	if w.opts.OnSync != nil && result != nil {
		w.opts.OnSync(result)
	}

	if w.opts.Verbose && result != nil && result.BuildResult != nil {
		log.Printf("Sync complete: %d new, %d updated, %d merged",
			result.BuildResult.NewEntries,
			result.BuildResult.UpdatedEntries,
			result.BuildResult.MergedEntries)
	}

	return nil
}

// incrementalSync syncs only the specified files
func (w *GitWatcher) incrementalSync(files []string) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		BuildResult: &BuildResult{},
		CleanResult: &CleanResult{},
	}

	for _, filePath := range files {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File was deleted, remove shadow entry
			if err := w.shadowFS.Delete(filePath); err == nil {
				result.CleanResult.RemovedEntries++
			}
			continue
		}

		// Build shadow entry for file
		err := w.builder.BuildFile(filePath, w.opts.BuildOptions)
		if err != nil {
			result.BuildResult.Errors = append(result.BuildResult.Errors, BuildError{
				Path:    filePath,
				Message: err.Error(),
				Err:     err,
			})
			continue
		}

		result.BuildResult.ProcessedFiles++
		result.BuildResult.UpdatedEntries++
	}

	result.Duration = time.Since(startTime)

	// Save index
	if err := w.shadowFS.SaveIndex(); err != nil {
		return result, fmt.Errorf("failed to save index: %w", err)
	}

	return result, nil
}

// getChangedFilesSince returns files changed since the given commit
func (w *GitWatcher) getChangedFilesSince(commit string) ([]string, error) {
	files, err := w.gitFilter.ChangedSince(commit)
	if err != nil {
		return nil, err
	}

	// Filter to existing and supported files
	existing := w.gitFilter.FilterToExisting(files)
	supported := w.gitFilter.FilterSupported(existing)

	return supported, nil
}

// SyncOnCommit performs a one-time sync based on a specific commit
func (w *GitWatcher) SyncOnCommit(commit string) (*SyncResult, error) {
	files, err := w.gitFilter.ChangedInCommit(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to get files in commit: %w", err)
	}

	// Filter files
	existing := w.gitFilter.FilterToExisting(files)
	supported := w.gitFilter.FilterSupported(existing)

	if len(supported) == 0 {
		return &SyncResult{
			BuildResult: &BuildResult{},
			CleanResult: &CleanResult{},
		}, nil
	}

	return w.incrementalSync(supported)
}

// Helper functions

// getHEADCommit returns the current HEAD commit hash
func getHEADCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse failed: %s: %w", stderr.String(), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// getHEADRef returns the current HEAD reference (branch name or "HEAD" if detached)
func getHEADRef(repoPath string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Might be in detached HEAD state
		return "HEAD", nil
	}

	return strings.TrimSpace(stdout.String()), nil
}

// getRecentCommits returns the N most recent commit hashes
func getRecentCommits(repoPath string, count int) ([]string, error) {
	cmd := exec.Command("git", "log", "--oneline", "-n", fmt.Sprintf("%d", count), "--format=%H")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git log failed: %s: %w", stderr.String(), err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	commits := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, nil
}

// InstallGitHook installs a post-commit hook to trigger shadow sync
func InstallGitHook(repoPath string, graphfsPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "post-commit")

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Read existing hook
		content, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("failed to read existing hook: %w", err)
		}

		// Check if our hook is already installed
		if strings.Contains(string(content), "graphfs shadow") {
			return nil // Already installed
		}

		// Append to existing hook
		f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			return fmt.Errorf("failed to open hook file: %w", err)
		}
		defer f.Close()

		hookScript := fmt.Sprintf("\n# GraphFS shadow sync\n%s shadow sync --quiet &\n", graphfsPath)
		if _, err := f.WriteString(hookScript); err != nil {
			return fmt.Errorf("failed to write hook: %w", err)
		}
	} else {
		// Create new hook
		hookScript := fmt.Sprintf(`#!/bin/sh
# GraphFS post-commit hook
# Automatically sync shadow file system after commits

%s shadow sync --quiet &
`, graphfsPath)

		if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
			return fmt.Errorf("failed to create hook: %w", err)
		}
	}

	return nil
}

// UninstallGitHook removes the graphfs post-commit hook
func UninstallGitHook(repoPath string) error {
	hookPath := filepath.Join(repoPath, ".git", "hooks", "post-commit")

	content, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No hook to remove
		}
		return fmt.Errorf("failed to read hook: %w", err)
	}

	// Remove graphfs lines and associated comments
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inGraphFSBlock := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip GraphFS comment lines
		if strings.Contains(trimmedLine, "# GraphFS") ||
			strings.Contains(trimmedLine, "graphfs shadow") ||
			strings.Contains(trimmedLine, "# Automatically sync shadow") {
			inGraphFSBlock = true
			continue
		}

		// Skip empty lines within the GraphFS block
		if inGraphFSBlock && trimmedLine == "" {
			continue
		}

		// Exit GraphFS block when we see non-empty, non-graphfs content
		inGraphFSBlock = false
		newLines = append(newLines, line)
	}

	newContent := strings.TrimSpace(strings.Join(newLines, "\n"))

	// If only shebang remains or empty, remove the file
	if newContent == "#!/bin/sh" || newContent == "" {
		return os.Remove(hookPath)
	}

	return os.WriteFile(hookPath, []byte(newContent+"\n"), 0755)
}
