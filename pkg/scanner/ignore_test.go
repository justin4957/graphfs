package scanner

import (
	"testing"
)

func TestIgnoreMatcher_ShouldIgnore(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		want     bool
	}{
		{
			name:     "exact match",
			patterns: []string{".git"},
			path:     ".git",
			want:     true,
		},
		{
			name:     "directory anywhere in path",
			patterns: []string{"node_modules"},
			path:     "src/node_modules/package",
			want:     true,
		},
		{
			name:     "extension match",
			patterns: []string{"*.pyc"},
			path:     "script.pyc",
			want:     true,
		},
		{
			name:     "extension match in subdirectory",
			patterns: []string{"*.pyc"},
			path:     "src/utils/cache.pyc",
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{".git", "*.pyc"},
			path:     "main.go",
			want:     false,
		},
		{
			name:     "vendor directory",
			patterns: []string{"vendor"},
			path:     "vendor/package/file.go",
			want:     true,
		},
		{
			name:     ".idea directory",
			patterns: []string{".idea"},
			path:     ".idea/workspace.xml",
			want:     true,
		},
		{
			name:     "build directory",
			patterns: []string{"build"},
			path:     "project/build/output.js",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewIgnoreMatcher(tt.patterns)
			got := matcher.ShouldIgnore(tt.path)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := DefaultIgnorePatterns()

	if len(patterns) == 0 {
		t.Fatal("DefaultIgnorePatterns() returned empty list")
	}

	// Check for some expected patterns
	expectedPatterns := []string{".git", "node_modules", "vendor", ".idea", ".DS_Store"}
	found := make(map[string]bool)

	for _, pattern := range patterns {
		found[pattern] = true
	}

	for _, expected := range expectedPatterns {
		if !found[expected] {
			t.Errorf("Expected pattern %q not found in default patterns", expected)
		}
	}
}

func TestIgnoreMatcher_AddPattern(t *testing.T) {
	matcher := NewIgnoreMatcher([]string{".git"})

	// Initially doesn't match
	if matcher.ShouldIgnore("node_modules/package") {
		t.Error("Should not ignore node_modules before adding pattern")
	}

	// Add pattern
	matcher.AddPattern("node_modules")

	// Now should match
	if !matcher.ShouldIgnore("node_modules/package") {
		t.Error("Should ignore node_modules after adding pattern")
	}
}

func TestIgnoreMatcher_AddPatterns(t *testing.T) {
	matcher := NewIgnoreMatcher([]string{".git"})

	matcher.AddPatterns([]string{"node_modules", "vendor", "*.pyc"})

	tests := []struct {
		path string
		want bool
	}{
		{".git", true},
		{"node_modules/package", true},
		{"vendor/lib", true},
		{"cache.pyc", true},
		{"main.go", false},
	}

	for _, tt := range tests {
		got := matcher.ShouldIgnore(tt.path)
		if got != tt.want {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestParseIgnoreFile(t *testing.T) {
	content := `# Comment line
.git
node_modules

# Another comment
*.pyc
vendor/

# Empty lines should be skipped

build
!important.txt
`

	patterns := ParseIgnoreFile(content)

	expectedPatterns := []string{".git", "node_modules", "*.pyc", "vendor/", "build"}

	if len(patterns) != len(expectedPatterns) {
		t.Fatalf("ParseIgnoreFile() returned %d patterns, want %d", len(patterns), len(expectedPatterns))
	}

	for i, pattern := range patterns {
		if pattern != expectedPatterns[i] {
			t.Errorf("Pattern[%d] = %q, want %q", i, pattern, expectedPatterns[i])
		}
	}
}

func TestIgnoreMatcher_ComplexPaths(t *testing.T) {
	matcher := NewIgnoreMatcher(DefaultIgnorePatterns())

	tests := []struct {
		path string
		want bool
	}{
		{"src/main.go", false},
		{"src/.git/config", true},
		{"project/node_modules/package/index.js", true},
		{"vendor/github.com/user/repo/file.go", true},
		{".idea/workspace.xml", true},
		{".vscode/settings.json", true},
		{"build/output.js", true},
		{"dist/bundle.js", true},
		{".DS_Store", true},
		{"Thumbs.db", true},
		{"test.pyc", true},
		{"__pycache__/module.pyc", true},
		{"examples/minimal-app/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := matcher.ShouldIgnore(tt.path)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func BenchmarkIgnoreMatcher_ShouldIgnore(b *testing.B) {
	matcher := NewIgnoreMatcher(DefaultIgnorePatterns())
	testPaths := []string{
		"src/main.go",
		"node_modules/package/index.js",
		"vendor/lib/file.go",
		".git/config",
		"build/output.js",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		matcher.ShouldIgnore(path)
	}
}
