/*
# Module: pkg/scanner/git_filter.go
Git-based file filtering for changed files.

Filters files based on git history to enable incremental analysis of only
changed files since a specific commit or branch.

## Linked Modules
- [scanner](./scanner.go) - Main scanner

## Tags
scanner, git, filter, incremental

## Exports
GitFilter, NewGitFilter

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#git_filter.go> a code:Module ;
    code:name "pkg/scanner/git_filter.go" ;
    code:description "Git-based file filtering for changed files" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./scanner.go> ;
    code:exports <#GitFilter>, <#NewGitFilter> ;
    code:tags "scanner", "git", "filter", "incremental" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitFilter filters files based on git history
type GitFilter struct {
	repoPath string
}

// NewGitFilter creates a new git filter for the given repository path
func NewGitFilter(repoPath string) *GitFilter {
	return &GitFilter{
		repoPath: repoPath,
	}
}

// IsGitRepository checks if the path is inside a git repository
func (g *GitFilter) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.repoPath
	return cmd.Run() == nil
}

// ChangedSince returns files changed since the given git reference
// The reference can be a branch name, tag, or commit hash
func (g *GitFilter) ChangedSince(ref string) ([]string, error) {
	// Get files changed between ref and HEAD
	cmd := exec.Command("git", "diff", "--name-only", ref+"...HEAD")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Try alternative: diff directly (for when ref is an ancestor)
		cmd = exec.Command("git", "diff", "--name-only", ref)
		cmd.Dir = g.repoPath
		stdout.Reset()
		stderr.Reset()
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("git diff failed: %s: %w", stderr.String(), err)
		}
	}

	return g.parseFileList(stdout.String()), nil
}

// ChangedInCommit returns files changed in a specific commit
func (g *GitFilter) ChangedInCommit(commit string) ([]string, error) {
	cmd := exec.Command("git", "show", "--name-only", "--format=", commit)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git show failed: %s: %w", stderr.String(), err)
	}

	return g.parseFileList(stdout.String()), nil
}

// UncommittedChanges returns files with uncommitted changes (staged and unstaged)
func (g *GitFilter) UncommittedChanges() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %s: %w", stderr.String(), err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		// Status output format: "XY filename" where XY is status
		file := strings.TrimSpace(line[3:])
		// Handle renamed files (old -> new format)
		if idx := strings.Index(file, " -> "); idx != -1 {
			file = file[idx+4:]
		}
		if file != "" {
			files = append(files, filepath.Join(g.repoPath, file))
		}
	}

	return files, nil
}

// StagedChanges returns only staged (added to index) files
func (g *GitFilter) StagedChanges() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %s: %w", stderr.String(), err)
	}

	return g.parseFileList(stdout.String()), nil
}

// parseFileList parses newline-separated file list from git output
func (g *GitFilter) parseFileList(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Convert to absolute path
			absPath := filepath.Join(g.repoPath, line)
			files = append(files, absPath)
		}
	}

	return files
}

// FilterToExisting filters the file list to only include files that exist
func (g *GitFilter) FilterToExisting(files []string) []string {
	existing := make([]string, 0, len(files))
	for _, file := range files {
		if fileExists(file) {
			existing = append(existing, file)
		}
	}
	return existing
}

// FilterSupported filters files to only supported source file types
func (g *GitFilter) FilterSupported(files []string) []string {
	supported := make([]string, 0, len(files))
	for _, file := range files {
		lang := DetectLanguage(file)
		if lang != "unknown" {
			supported = append(supported, file)
		}
	}
	return supported
}

// GetCurrentBranch returns the name of the current branch
func (g *GitFilter) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %s: %w", stderr.String(), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetDefaultBranch attempts to determine the default branch (main or master)
func (g *GitFilter) GetDefaultBranch() string {
	// Try 'main' first
	cmd := exec.Command("git", "rev-parse", "--verify", "main")
	cmd.Dir = g.repoPath
	if cmd.Run() == nil {
		return "main"
	}

	// Try 'master'
	cmd = exec.Command("git", "rev-parse", "--verify", "master")
	cmd.Dir = g.repoPath
	if cmd.Run() == nil {
		return "master"
	}

	// Default to 'main'
	return "main"
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	// Use ls for cross-platform compatibility
	cmd := exec.Command("ls", path)
	return cmd.Run() == nil
}
