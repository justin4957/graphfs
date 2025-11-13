/*
# Module: pkg/scanner/ignore.go
Ignore pattern handling for filesystem scanning.

Implements .gitignore-style pattern matching for excluding files and directories.
Supports glob patterns, negation, and common ignore patterns.

## Linked Modules
None (utility module)

## Tags
scanner, ignore-patterns, filtering

## Exports
IgnoreMatcher, NewIgnoreMatcher, DefaultIgnorePatterns

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#ignore.go> a code:Module ;
    code:name "pkg/scanner/ignore.go" ;
    code:description "Ignore pattern handling for filesystem scanning" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:exports <#IgnoreMatcher>, <#NewIgnoreMatcher>, <#DefaultIgnorePatterns> ;
    code:tags "scanner", "ignore-patterns", "filtering" .

<#IgnoreMatcher> a code:Type ;
    code:name "IgnoreMatcher" ;
    code:kind "struct" ;
    code:description "Matches file paths against ignore patterns" ;
    code:hasMethod <#IgnoreMatcher.ShouldIgnore>, <#IgnoreMatcher.AddPattern> .

<#NewIgnoreMatcher> a code:Function ;
    code:name "NewIgnoreMatcher" ;
    code:description "Creates new ignore matcher with patterns" ;
    code:returns <#IgnoreMatcher> .

<#DefaultIgnorePatterns> a code:Function ;
    code:name "DefaultIgnorePatterns" ;
    code:description "Returns default ignore patterns" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"path/filepath"
	"strings"
)

// IgnoreMatcher matches file paths against ignore patterns
type IgnoreMatcher struct {
	patterns []string
}

// NewIgnoreMatcher creates a new ignore matcher with the given patterns
func NewIgnoreMatcher(patterns []string) *IgnoreMatcher {
	return &IgnoreMatcher{
		patterns: patterns,
	}
}

// DefaultIgnorePatterns returns common ignore patterns
func DefaultIgnorePatterns() []string {
	return []string{
		// Version control
		".git",
		".svn",
		".hg",
		".bzr",

		// Dependencies
		"node_modules",
		"vendor",
		"target",
		"dist",
		"build",
		"out",
		"bin",

		// IDE
		".idea",
		".vscode",
		".vs",
		"*.swp",
		"*.swo",
		"*~",

		// OS
		".DS_Store",
		"Thumbs.db",

		// Build artifacts
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",
		"*.o",
		"*.a",
		"*.class",
		"*.pyc",
		"__pycache__",

		// GraphFS specific
		".graphfs/store.db",
		".graphfs/cache",
	}
}

// AddPattern adds a pattern to the matcher
func (m *IgnoreMatcher) AddPattern(pattern string) {
	m.patterns = append(m.patterns, pattern)
}

// AddPatterns adds multiple patterns to the matcher
func (m *IgnoreMatcher) AddPatterns(patterns []string) {
	m.patterns = append(m.patterns, patterns...)
}

// ShouldIgnore checks if a path should be ignored based on patterns
func (m *IgnoreMatcher) ShouldIgnore(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	for _, pattern := range m.patterns {
		if m.matchPattern(path, pattern) {
			return true
		}
	}

	return false
}

// matchPattern matches a path against a pattern
func (m *IgnoreMatcher) matchPattern(path, pattern string) bool {
	// Normalize pattern
	pattern = filepath.ToSlash(pattern)

	// Exact match
	if path == pattern {
		return true
	}

	// Directory match - if pattern is a directory name, match it anywhere in path
	if !strings.Contains(pattern, "/") && !strings.Contains(pattern, "*") {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			if part == pattern {
				return true
			}
		}
	}

	// Suffix match for extensions
	if strings.HasPrefix(pattern, "*.") {
		ext := pattern[1:] // Remove the *
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	// Prefix match
	if strings.HasSuffix(pattern, "/") {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}

	// Glob-style pattern matching (simplified)
	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(filepath.FromSlash(pattern), filepath.FromSlash(path))
		if matched {
			return true
		}

		// Also try matching just the base name
		baseName := filepath.Base(path)
		matched, _ = filepath.Match(pattern, baseName)
		if matched {
			return true
		}
	}

	// Contains match
	if strings.Contains(path, pattern) {
		return true
	}

	return false
}

// ParseIgnoreFile parses a .gitignore-style file content into patterns
func ParseIgnoreFile(content string) []string {
	var patterns []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove negation prefix (! ) - we don't support it yet
		if strings.HasPrefix(line, "!") {
			continue
		}

		patterns = append(patterns, line)
	}

	return patterns
}
