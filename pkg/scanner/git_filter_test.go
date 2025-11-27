/*
# Module: pkg/scanner/git_filter_test.go
Tests for git-based file filtering.

## Linked Modules
- [git_filter](./git_filter.go) - Git filter implementation

## Tags
scanner, git, filter, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#git_filter_test.go> a code:Module ;
    code:name "pkg/scanner/git_filter_test.go" ;
    code:description "Tests for git-based file filtering" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./git_filter.go> ;
    code:tags "scanner", "git", "filter", "test" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitFilterIsGitRepository(t *testing.T) {
	// Test with current directory (should be a git repo)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1]) // Remove trailing newline
	filter := NewGitFilter(gitRoot)

	if !filter.IsGitRepository() {
		t.Error("IsGitRepository() returned false for a git repo")
	}
}

func TestGitFilterIsGitRepositoryFalse(t *testing.T) {
	tempDir := t.TempDir()
	filter := NewGitFilter(tempDir)

	if filter.IsGitRepository() {
		t.Error("IsGitRepository() returned true for non-git directory")
	}
}

func TestGitFilterGetCurrentBranch(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1])
	filter := NewGitFilter(gitRoot)

	branch, err := filter.GetCurrentBranch()
	if err != nil {
		t.Errorf("GetCurrentBranch() error: %v", err)
	}

	if branch == "" {
		t.Error("GetCurrentBranch() returned empty string")
	}
}

func TestGitFilterGetDefaultBranch(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1])
	filter := NewGitFilter(gitRoot)

	defaultBranch := filter.GetDefaultBranch()
	if defaultBranch != "main" && defaultBranch != "master" {
		t.Errorf("GetDefaultBranch() returned %q, want 'main' or 'master'", defaultBranch)
	}
}

func TestGitFilterFilterToExisting(t *testing.T) {
	tempDir := t.TempDir()

	// Create some files
	existingFile := filepath.Join(tempDir, "exists.go")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	files := []string{
		existingFile,
		filepath.Join(tempDir, "does_not_exist.go"),
	}

	filter := NewGitFilter(tempDir)
	result := filter.FilterToExisting(files)

	if len(result) != 1 {
		t.Errorf("FilterToExisting() returned %d files, want 1", len(result))
	}

	if result[0] != existingFile {
		t.Errorf("FilterToExisting() returned wrong file: %s", result[0])
	}
}

func TestGitFilterFilterSupported(t *testing.T) {
	files := []string{
		"/path/to/file.go",
		"/path/to/file.ts",
		"/path/to/file.py",
		"/path/to/file.unknown",
		"/path/to/file.xyz",
	}

	filter := NewGitFilter("/path")
	result := filter.FilterSupported(files)

	// go, ts, and py should be supported
	if len(result) < 3 {
		t.Errorf("FilterSupported() returned %d files, want at least 3", len(result))
	}
}

func TestGitFilterParseFileList(t *testing.T) {
	filter := NewGitFilter("/project")

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple list",
			input:    "file1.go\nfile2.go\nfile3.go",
			expected: []string{"/project/file1.go", "/project/file2.go", "/project/file3.go"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "  \n\n  ",
			expected: []string{},
		},
		{
			name:     "with trailing newline",
			input:    "file1.go\nfile2.go\n",
			expected: []string{"/project/file1.go", "/project/file2.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.parseFileList(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseFileList() returned %d files, want %d", len(result), len(tt.expected))
				return
			}

			for i, f := range result {
				if f != tt.expected[i] {
					t.Errorf("parseFileList()[%d] = %q, want %q", i, f, tt.expected[i])
				}
			}
		})
	}
}

func TestGitFilterChangedSinceInvalidRef(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1])
	filter := NewGitFilter(gitRoot)

	_, err = filter.ChangedSince("nonexistent-ref-12345")
	if err == nil {
		t.Error("ChangedSince() should error with invalid ref")
	}
}

func TestGitFilterUncommittedChanges(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1])
	filter := NewGitFilter(gitRoot)

	// Just test that it doesn't error - actual changes depend on repo state
	_, err = filter.UncommittedChanges()
	if err != nil {
		t.Errorf("UncommittedChanges() error: %v", err)
	}
}

func TestGitFilterStagedChanges(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Find git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Not running in a git repository")
	}

	gitRoot := string(output[:len(output)-1])
	filter := NewGitFilter(gitRoot)

	// Just test that it doesn't error - actual changes depend on repo state
	_, err = filter.StagedChanges()
	if err != nil {
		t.Errorf("StagedChanges() error: %v", err)
	}
}
