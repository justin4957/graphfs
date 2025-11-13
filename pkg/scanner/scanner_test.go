package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_ScanFile(t *testing.T) {
	scanner := NewScanner()

	// Test scanning a Go file from minimal-app
	testFile := filepath.Join("..", "..", "examples", "minimal-app", "main.go")

	fileInfo, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	if fileInfo.Language != "Go" {
		t.Errorf("Language = %q, want %q", fileInfo.Language, "Go")
	}

	if fileInfo.Size == 0 {
		t.Error("File size should not be zero")
	}

	if !fileInfo.HasLinkedDoc {
		t.Error("main.go should have LinkedDoc")
	}

	t.Logf("Scanned file: %s, Language: %s, Size: %d bytes, HasLinkedDoc: %v",
		fileInfo.Path, fileInfo.Language, fileInfo.Size, fileInfo.HasLinkedDoc)
}

func TestScanner_Scan_MinimalApp(t *testing.T) {
	scanner := NewScanner()

	minimalAppPath := filepath.Join("..", "..", "examples", "minimal-app")

	// Check if minimal-app exists
	if _, err := os.Stat(minimalAppPath); os.IsNotExist(err) {
		t.Skip("minimal-app directory not found")
	}

	opts := DefaultScanOptions()
	opts.Concurrent = false // Use sequential for predictable testing

	result, err := scanner.Scan(minimalAppPath, opts)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("Scan() returned no files")
	}

	// Count files by language
	langCounts := make(map[string]int)
	linkedDocCount := 0

	for _, file := range result.Files {
		langCounts[file.Language]++
		if file.HasLinkedDoc {
			linkedDocCount++
		}
		t.Logf("Found: %s (%s, %d bytes, LinkedDoc: %v)",
			filepath.Base(file.Path), file.Language, file.Size, file.HasLinkedDoc)
	}

	// minimal-app has 7 Go files (main.go, user.go, auth.go, user service, logger, crypto, validator)
	if langCounts["Go"] < 7 {
		t.Errorf("Expected at least 7 Go files, got %d", langCounts["Go"])
	}

	// All minimal-app files should have LinkedDoc
	if linkedDocCount < 7 {
		t.Errorf("Expected at least 7 files with LinkedDoc, got %d", linkedDocCount)
	}

	t.Logf("Scan completed: %d files, %d bytes, %v duration",
		result.TotalFiles, result.TotalBytes, result.Duration)
}

func TestScanner_Scan_WithOptions(t *testing.T) {
	scanner := NewScanner()

	minimalAppPath := filepath.Join("..", "..", "examples", "minimal-app")

	if _, err := os.Stat(minimalAppPath); os.IsNotExist(err) {
		t.Skip("minimal-app directory not found")
	}

	tests := []struct {
		name    string
		opts    ScanOptions
		wantMin int // Minimum number of files expected
	}{
		{
			name:    "default options",
			opts:    DefaultScanOptions(),
			wantMin: 7,
		},
		{
			name: "with exclude pattern",
			opts: ScanOptions{
				ExcludePatterns: []string{"**/services/**"},
				UseDefaults:     true,
			},
			wantMin: 4, // Should exclude services/*.go
		},
		{
			name: "concurrent scanning",
			opts: ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
			},
			wantMin: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scanner.Scan(minimalAppPath, tt.opts)
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			if len(result.Files) < tt.wantMin {
				t.Errorf("Expected at least %d files, got %d", tt.wantMin, len(result.Files))
			}

			t.Logf("Scanned %d files in %v", result.TotalFiles, result.Duration)
		})
	}
}

func TestScanner_Scan_IgnorePatterns(t *testing.T) {
	scanner := NewScanner()

	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"main.go",
		"util.go",
		"node_modules/package/index.js",
		".git/config",
		"vendor/lib/lib.go",
		"build/output.js",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("package main"), 0644)
	}

	opts := DefaultScanOptions()
	result, err := scanner.Scan(tmpDir, opts)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only find main.go and util.go (node_modules, .git, vendor, build are ignored)
	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files (main.go, util.go), got %d", len(result.Files))
		for _, file := range result.Files {
			t.Logf("Found: %s", file.Path)
		}
	}

	// Verify the found files are the expected ones
	foundFiles := make(map[string]bool)
	for _, file := range result.Files {
		base := filepath.Base(file.Path)
		foundFiles[base] = true
	}

	if !foundFiles["main.go"] {
		t.Error("main.go should be found")
	}
	if !foundFiles["util.go"] {
		t.Error("util.go should be found")
	}
}

func TestScanner_Scan_NonExistentPath(t *testing.T) {
	scanner := NewScanner()

	_, err := scanner.Scan("/nonexistent/path", DefaultScanOptions())
	if err == nil {
		t.Error("Scan() should return error for nonexistent path")
	}
}

func TestScanner_Scan_MaxFileSize(t *testing.T) {
	scanner := NewScanner()

	tmpDir := t.TempDir()

	// Create small file
	smallFile := filepath.Join(tmpDir, "small.go")
	os.WriteFile(smallFile, []byte("package main"), 0644)

	// Create large file
	largeFile := filepath.Join(tmpDir, "large.go")
	largeContent := make([]byte, 2048)
	os.WriteFile(largeFile, largeContent, 0644)

	opts := DefaultScanOptions()
	opts.MaxFileSize = 1024 // 1KB limit

	result, err := scanner.Scan(tmpDir, opts)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only find small.go
	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file (small.go), got %d", len(result.Files))
	}

	if len(result.Files) > 0 && filepath.Base(result.Files[0].Path) != "small.go" {
		t.Errorf("Expected small.go, got %s", filepath.Base(result.Files[0].Path))
	}
}

func TestDefaultScanOptions(t *testing.T) {
	opts := DefaultScanOptions()

	if opts.MaxFileSize == 0 {
		t.Error("MaxFileSize should not be zero")
	}

	if opts.FollowSymlinks {
		t.Error("FollowSymlinks should be false by default")
	}

	if !opts.UseDefaults {
		t.Error("UseDefaults should be true")
	}

	if !opts.Concurrent {
		t.Error("Concurrent should be true")
	}

	if len(opts.IgnoreFiles) == 0 {
		t.Error("IgnoreFiles should contain default values")
	}
}

func BenchmarkScanner_Scan(b *testing.B) {
	scanner := NewScanner()
	minimalAppPath := filepath.Join("..", "..", "examples", "minimal-app")

	if _, err := os.Stat(minimalAppPath); os.IsNotExist(err) {
		b.Skip("minimal-app directory not found")
	}

	opts := DefaultScanOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan(minimalAppPath, opts)
	}
}

func BenchmarkScanner_ScanFile(b *testing.B) {
	scanner := NewScanner()
	testFile := filepath.Join("..", "..", "examples", "minimal-app", "main.go")

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skip("test file not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.ScanFile(testFile)
	}
}
