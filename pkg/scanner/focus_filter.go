/*
# Module: pkg/scanner/focus_filter.go
Focus pattern filtering for targeted analysis.

Filters files based on glob patterns to enable focused analysis of specific
subsystems or file types.

## Linked Modules
- [scanner](./scanner.go) - Main scanner

## Tags
scanner, filter, pattern, focus

## Exports
FocusFilter, NewFocusFilter

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#focus_filter.go> a code:Module ;
    code:name "pkg/scanner/focus_filter.go" ;
    code:description "Focus pattern filtering for targeted analysis" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./scanner.go> ;
    code:exports <#FocusFilter>, <#NewFocusFilter> ;
    code:tags "scanner", "filter", "pattern", "focus" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"path/filepath"
	"strings"
)

// FocusFilter filters files based on glob patterns
type FocusFilter struct {
	patterns []string
	basePath string
}

// NewFocusFilter creates a new focus filter with the given patterns
// Patterns support standard glob syntax including:
// - * matches any sequence of non-separator characters
// - ** matches any sequence including separators (recursive)
// - ? matches any single non-separator character
func NewFocusFilter(patterns []string, basePath string) *FocusFilter {
	return &FocusFilter{
		patterns: patterns,
		basePath: basePath,
	}
}

// Match filters files to only include those matching any of the patterns
// Uses OR logic - a file is included if it matches ANY pattern
func (f *FocusFilter) Match(files []string) []string {
	if len(f.patterns) == 0 {
		return files
	}

	matched := make([]string, 0, len(files))

	for _, file := range files {
		if f.matchesAny(file) {
			matched = append(matched, file)
		}
	}

	return matched
}

// matchesAny checks if a file matches any of the patterns
func (f *FocusFilter) matchesAny(file string) bool {
	// Get relative path for matching
	relPath := file
	if f.basePath != "" {
		if rel, err := filepath.Rel(f.basePath, file); err == nil {
			relPath = rel
		}
	}

	for _, pattern := range f.patterns {
		if f.matchPattern(pattern, relPath) || f.matchPattern(pattern, file) {
			return true
		}
	}

	return false
}

// matchPattern matches a single pattern against a path
func (f *FocusFilter) matchPattern(pattern, path string) bool {
	// Handle ** (double star) for recursive matching
	if strings.Contains(pattern, "**") {
		return f.matchDoubleStarPattern(pattern, path)
	}

	// Use standard filepath.Match for simple patterns
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	if matched {
		return true
	}

	// Also try matching against just the filename
	matched, err = filepath.Match(pattern, filepath.Base(path))
	if err != nil {
		return false
	}
	return matched
}

// matchDoubleStarPattern handles patterns with ** (recursive glob)
func (f *FocusFilter) matchDoubleStarPattern(pattern, path string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 1 {
		// No ** found, use regular match
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Handle patterns like "**/*.go" or "api/**/*.go"
	prefix := parts[0]
	suffix := parts[1]

	// Remove leading/trailing slashes
	prefix = strings.TrimSuffix(prefix, "/")
	prefix = strings.TrimSuffix(prefix, string(filepath.Separator))
	suffix = strings.TrimPrefix(suffix, "/")
	suffix = strings.TrimPrefix(suffix, string(filepath.Separator))

	// Check prefix (if any)
	if prefix != "" {
		if !strings.HasPrefix(path, prefix) && !strings.HasPrefix(path, prefix+string(filepath.Separator)) {
			return false
		}
		// Remove matched prefix from path
		path = strings.TrimPrefix(path, prefix)
		path = strings.TrimPrefix(path, string(filepath.Separator))
	}

	// Check suffix (if any)
	if suffix != "" {
		// If suffix contains /, we need to match against path suffix
		if strings.Contains(suffix, "/") || strings.Contains(suffix, string(filepath.Separator)) {
			return strings.HasSuffix(path, suffix) || strings.HasSuffix(path, strings.ReplaceAll(suffix, "/", string(filepath.Separator)))
		}

		// Simple suffix like "*.go"
		matched, _ := filepath.Match(suffix, filepath.Base(path))
		return matched
	}

	// Pattern like "api/**" - matches anything under api
	return true
}

// MatchWithReason filters files and returns why each was matched
func (f *FocusFilter) MatchWithReason(files []string) map[string]string {
	result := make(map[string]string)

	for _, file := range files {
		relPath := file
		if f.basePath != "" {
			if rel, err := filepath.Rel(f.basePath, file); err == nil {
				relPath = rel
			}
		}

		for _, pattern := range f.patterns {
			if f.matchPattern(pattern, relPath) || f.matchPattern(pattern, file) {
				result[file] = pattern
				break
			}
		}
	}

	return result
}

// Patterns returns the configured patterns
func (f *FocusFilter) Patterns() []string {
	return f.patterns
}

// HasPatterns returns true if any patterns are configured
func (f *FocusFilter) HasPatterns() bool {
	return len(f.patterns) > 0
}

// AddPattern adds a pattern to the filter
func (f *FocusFilter) AddPattern(pattern string) {
	f.patterns = append(f.patterns, pattern)
}

// FilterStats provides statistics about a filtering operation
type FilterStats struct {
	TotalFiles   int
	MatchedFiles int
	ByPattern    map[string]int // Files matched per pattern
}

// MatchWithStats filters files and returns statistics
func (f *FocusFilter) MatchWithStats(files []string) ([]string, *FilterStats) {
	stats := &FilterStats{
		TotalFiles: len(files),
		ByPattern:  make(map[string]int),
	}

	if len(f.patterns) == 0 {
		stats.MatchedFiles = len(files)
		return files, stats
	}

	matched := make([]string, 0, len(files))

	for _, file := range files {
		relPath := file
		if f.basePath != "" {
			if rel, err := filepath.Rel(f.basePath, file); err == nil {
				relPath = rel
			}
		}

		for _, pattern := range f.patterns {
			if f.matchPattern(pattern, relPath) || f.matchPattern(pattern, file) {
				matched = append(matched, file)
				stats.ByPattern[pattern]++
				break
			}
		}
	}

	stats.MatchedFiles = len(matched)
	return matched, stats
}
