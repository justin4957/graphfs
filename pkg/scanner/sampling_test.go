/*
# Module: pkg/scanner/sampling_test.go
Tests for sampling strategies.

## Linked Modules
- [sampling](./sampling.go) - Sampling implementation

## Tags
scanner, sampling, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#sampling_test.go> a code:Module ;
    code:name "pkg/scanner/sampling_test.go" ;
    code:description "Tests for sampling strategies" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./sampling.go> ;
    code:tags "scanner", "sampling", "test" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseSamplingStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SamplingStrategy
	}{
		{"random", "random", SampleRandom},
		{"stratified", "stratified", SampleStratified},
		{"recent", "recent", SampleRecent},
		{"unknown defaults to random", "unknown", SampleRandom},
		{"empty defaults to random", "", SampleRandom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSamplingStrategy(tt.input)
			if result != tt.expected {
				t.Errorf("ParseSamplingStrategy(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSamplingStrategyString(t *testing.T) {
	tests := []struct {
		strategy SamplingStrategy
		expected string
	}{
		{SampleRandom, "random"},
		{SampleStratified, "stratified"},
		{SampleRecent, "recent"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.strategy.String()
			if result != tt.expected {
				t.Errorf("SamplingStrategy.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSamplerRandomSample(t *testing.T) {
	files := []string{
		"/path/to/file1.go",
		"/path/to/file2.go",
		"/path/to/file3.go",
		"/path/to/file4.go",
		"/path/to/file5.go",
	}

	sampler := NewSampler(SampleRandom, 3, 42) // Fixed seed for reproducibility
	result := sampler.Sample(files)

	if len(result) != 3 {
		t.Errorf("Sample() returned %d files, want 3", len(result))
	}

	// Verify all returned files are from the original set
	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}
	for _, f := range result {
		if !fileSet[f] {
			t.Errorf("Sample() returned unexpected file: %s", f)
		}
	}
}

func TestSamplerRandomSampleReproducible(t *testing.T) {
	files := []string{
		"/path/to/file1.go",
		"/path/to/file2.go",
		"/path/to/file3.go",
		"/path/to/file4.go",
		"/path/to/file5.go",
	}

	// Same seed should produce same results
	sampler1 := NewSampler(SampleRandom, 3, 42)
	result1 := sampler1.Sample(files)

	sampler2 := NewSampler(SampleRandom, 3, 42)
	result2 := sampler2.Sample(files)

	if len(result1) != len(result2) {
		t.Errorf("Same seed produced different sample sizes: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("Same seed produced different results at index %d: %s vs %s", i, result1[i], result2[i])
		}
	}
}

func TestSamplerSampleSizeExceedsFiles(t *testing.T) {
	files := []string{
		"/path/to/file1.go",
		"/path/to/file2.go",
	}

	sampler := NewSampler(SampleRandom, 10, 42)
	result := sampler.Sample(files)

	if len(result) != len(files) {
		t.Errorf("Sample() returned %d files, want %d (all files)", len(result), len(files))
	}
}

func TestSamplerStratifiedSample(t *testing.T) {
	files := []string{
		"/path/dir1/file1.go",
		"/path/dir1/file2.go",
		"/path/dir1/file3.go",
		"/path/dir2/file1.go",
		"/path/dir2/file2.go",
		"/path/dir3/file1.go",
	}

	sampler := NewSampler(SampleStratified, 3, 42)
	result := sampler.Sample(files)

	if len(result) != 3 {
		t.Errorf("Sample() returned %d files, want 3", len(result))
	}

	// Verify stratification - should have files from multiple directories
	dirs := make(map[string]int)
	for _, f := range result {
		dir := filepath.Dir(f)
		dirs[dir]++
	}

	// With stratified sampling, we should have representation from multiple dirs
	if len(dirs) < 2 {
		t.Errorf("Stratified sampling should include files from multiple directories, got %d", len(dirs))
	}
}

func TestSamplerRecentSample(t *testing.T) {
	// Create temporary files with different modification times
	tempDir := t.TempDir()

	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		files[i] = filepath.Join(tempDir, "file"+string(rune('a'+i))+".go")
		if err := os.WriteFile(files[i], []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		// Set different modification times
		mtime := time.Now().Add(time.Duration(-i) * time.Hour)
		if err := os.Chtimes(files[i], mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	sampler := NewSampler(SampleRecent, 3, 0)
	result := sampler.Sample(files)

	if len(result) != 3 {
		t.Errorf("Sample() returned %d files, want 3", len(result))
	}

	// First file should be the most recent (filea.go)
	if result[0] != files[0] {
		t.Errorf("Most recent file should be first, got %s, want %s", result[0], files[0])
	}
}

func TestSampleWithStats(t *testing.T) {
	files := []string{
		"/path/dir1/file1.go",
		"/path/dir1/file2.go",
		"/path/dir2/file1.go",
		"/path/dir2/file2.go",
	}

	sampler := NewSampler(SampleRandom, 2, 42)
	result, stats := sampler.SampleWithStats(files)

	if stats.TotalFiles != 4 {
		t.Errorf("TotalFiles = %d, want 4", stats.TotalFiles)
	}

	if stats.SampledFiles != len(result) {
		t.Errorf("SampledFiles = %d, want %d", stats.SampledFiles, len(result))
	}

	if stats.Strategy != SampleRandom {
		t.Errorf("Strategy = %v, want %v", stats.Strategy, SampleRandom)
	}
}

func TestSamplerEmptyFiles(t *testing.T) {
	sampler := NewSampler(SampleRandom, 10, 42)
	result := sampler.Sample([]string{})

	if len(result) != 0 {
		t.Errorf("Sample() on empty slice returned %d files, want 0", len(result))
	}
}
