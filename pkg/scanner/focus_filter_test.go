/*
# Module: pkg/scanner/focus_filter_test.go
Tests for focus pattern filtering.

## Linked Modules
- [focus_filter](./focus_filter.go) - Focus filter implementation

## Tags
scanner, filter, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#focus_filter_test.go> a code:Module ;
    code:name "pkg/scanner/focus_filter_test.go" ;
    code:description "Tests for focus pattern filtering" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./focus_filter.go> ;
    code:tags "scanner", "filter", "test" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"testing"
)

func TestFocusFilterMatch(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		files    []string
		expected []string
	}{
		{
			name:     "no patterns returns all files",
			patterns: []string{},
			files:    []string{"/path/to/file1.go", "/path/to/file2.go"},
			expected: []string{"/path/to/file1.go", "/path/to/file2.go"},
		},
		{
			name:     "simple extension pattern",
			patterns: []string{"*.go"},
			files:    []string{"/path/to/file.go", "/path/to/file.ts", "/path/to/file.js"},
			expected: []string{"/path/to/file.go"},
		},
		{
			name:     "double star recursive pattern",
			patterns: []string{"api/**/*.go"},
			files: []string{
				"/root/api/handler.go",
				"/root/api/v1/users.go",
				"/root/api/v1/posts.go",
				"/root/pkg/utils.go",
			},
			expected: []string{
				"/root/api/handler.go",
				"/root/api/v1/users.go",
				"/root/api/v1/posts.go",
			},
		},
		{
			name:     "multiple patterns OR logic",
			patterns: []string{"*.go", "*.ts"},
			files: []string{
				"/path/file.go",
				"/path/file.ts",
				"/path/file.js",
			},
			expected: []string{"/path/file.go", "/path/file.ts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFocusFilter(tt.patterns, "/root")
			result := filter.Match(tt.files)

			if len(result) != len(tt.expected) {
				t.Errorf("Match() returned %d files, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Want: %v", tt.expected)
				return
			}

			expectedSet := make(map[string]bool)
			for _, f := range tt.expected {
				expectedSet[f] = true
			}

			for _, f := range result {
				if !expectedSet[f] {
					t.Errorf("Unexpected file in result: %s", f)
				}
			}
		})
	}
}

func TestFocusFilterMatchWithReason(t *testing.T) {
	filter := NewFocusFilter([]string{"*.go", "api/**"}, "/root")
	files := []string{
		"/root/main.go",
		"/root/api/handler.go",
		"/root/pkg/utils.js",
	}

	result := filter.MatchWithReason(files)

	if len(result) != 2 {
		t.Errorf("MatchWithReason() returned %d files, want 2", len(result))
	}

	if _, ok := result["/root/main.go"]; !ok {
		t.Error("main.go should be in result")
	}

	if _, ok := result["/root/api/handler.go"]; !ok {
		t.Error("api/handler.go should be in result")
	}
}

func TestFocusFilterMatchWithStats(t *testing.T) {
	filter := NewFocusFilter([]string{"*.go"}, "/root")
	files := []string{
		"/root/file1.go",
		"/root/file2.go",
		"/root/file3.ts",
	}

	result, stats := filter.MatchWithStats(files)

	if stats.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", stats.TotalFiles)
	}

	if stats.MatchedFiles != 2 {
		t.Errorf("MatchedFiles = %d, want 2", stats.MatchedFiles)
	}

	if len(result) != 2 {
		t.Errorf("Match returned %d files, want 2", len(result))
	}

	if stats.ByPattern["*.go"] != 2 {
		t.Errorf("ByPattern[*.go] = %d, want 2", stats.ByPattern["*.go"])
	}
}

func TestFocusFilterHasPatterns(t *testing.T) {
	filterEmpty := NewFocusFilter([]string{}, "/root")
	if filterEmpty.HasPatterns() {
		t.Error("HasPatterns() should return false for empty patterns")
	}

	filterWithPatterns := NewFocusFilter([]string{"*.go"}, "/root")
	if !filterWithPatterns.HasPatterns() {
		t.Error("HasPatterns() should return true for non-empty patterns")
	}
}

func TestFocusFilterPatterns(t *testing.T) {
	patterns := []string{"*.go", "*.ts"}
	filter := NewFocusFilter(patterns, "/root")

	result := filter.Patterns()
	if len(result) != len(patterns) {
		t.Errorf("Patterns() returned %d patterns, want %d", len(result), len(patterns))
	}
}

func TestFocusFilterAddPattern(t *testing.T) {
	filter := NewFocusFilter([]string{"*.go"}, "/root")
	filter.AddPattern("*.ts")

	if len(filter.Patterns()) != 2 {
		t.Errorf("After AddPattern, Patterns() returned %d, want 2", len(filter.Patterns()))
	}
}

func TestFocusFilterDoubleStarPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		basePath string
		files    []string
		expected []string
	}{
		{
			name:     "prefix double star",
			pattern:  "api/**",
			basePath: "/project",
			files: []string{
				"/project/api/handler.go",
				"/project/api/v1/users.go",
				"/project/pkg/utils.go",
			},
			expected: []string{
				"/project/api/handler.go",
				"/project/api/v1/users.go",
			},
		},
		{
			name:     "suffix double star",
			pattern:  "**/*.go",
			basePath: "/project",
			files: []string{
				"/project/main.go",
				"/project/api/handler.go",
				"/project/pkg/utils.ts",
			},
			expected: []string{
				"/project/main.go",
				"/project/api/handler.go",
			},
		},
		{
			name:     "middle double star",
			pattern:  "cmd/**/main.go",
			basePath: "/project",
			files: []string{
				"/project/cmd/app/main.go",
				"/project/cmd/tool/main.go",
				"/project/pkg/main.go",
			},
			expected: []string{
				"/project/cmd/app/main.go",
				"/project/cmd/tool/main.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFocusFilter([]string{tt.pattern}, tt.basePath)
			result := filter.Match(tt.files)

			if len(result) != len(tt.expected) {
				t.Errorf("Match() returned %d files, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Want: %v", tt.expected)
			}
		})
	}
}

func TestFocusFilterEmptyFiles(t *testing.T) {
	filter := NewFocusFilter([]string{"*.go"}, "/root")
	result := filter.Match([]string{})

	if len(result) != 0 {
		t.Errorf("Match() on empty files returned %d, want 0", len(result))
	}
}
