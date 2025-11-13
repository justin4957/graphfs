/*
# Module: pkg/scanner/scanner.go
Filesystem scanner for GraphFS.

Recursively scans directories to find source code files with language detection,
ignore pattern filtering, and LinkedDoc detection.

## Linked Modules
- [language](./language.go) - Language detection
- [ignore](./ignore.go) - Ignore pattern matching
- [../parser](../parser/parser.go) - LinkedDoc detection

## Tags
scanner, filesystem, recursive

## Exports
Scanner, NewScanner, ScanOptions, ScanResult, FileInfo

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#scanner.go> a code:Module ;
    code:name "pkg/scanner/scanner.go" ;
    code:description "Filesystem scanner for GraphFS" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./language.go>, <./ignore.go>, <../parser/parser.go> ;
    code:exports <#Scanner>, <#NewScanner>, <#ScanOptions>, <#ScanResult>, <#FileInfo> ;
    code:tags "scanner", "filesystem", "recursive" .

<#Scanner> a code:Type ;
    code:name "Scanner" ;
    code:kind "struct" ;
    code:description "Recursive filesystem scanner" ;
    code:hasMethod <#Scanner.Scan>, <#Scanner.ScanFile> .

<#Scanner.Scan> a code:Method ;
    code:name "Scan" ;
    code:description "Scans a directory recursively" .

<#Scanner.ScanFile> a code:Method ;
    code:name "ScanFile" ;
    code:description "Scans a single file" .

<#NewScanner> a code:Function ;
    code:name "NewScanner" ;
    code:description "Creates new scanner instance" ;
    code:returns <#Scanner> .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/justin4957/graphfs/pkg/parser"
)

// Scanner recursively scans directories for source files
type Scanner struct {
	parser *parser.Parser
}

// ScanOptions configures the scanner behavior
type ScanOptions struct {
	IncludePatterns []string
	ExcludePatterns []string
	MaxFileSize     int64
	FollowSymlinks  bool
	IgnoreFiles     []string
	UseDefaults     bool // Use default ignore patterns
	Concurrent      bool // Enable concurrent scanning
}

// DefaultScanOptions returns default scan options
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		MaxFileSize: 1024 * 1024, // 1MB
		FollowSymlinks: false,
		IgnoreFiles: []string{".gitignore", ".graphfsignore"},
		UseDefaults: true,
		Concurrent:  true,
	}
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	Files      []*FileInfo
	TotalFiles int
	TotalBytes int64
	Errors     []error
	Duration   time.Duration
}

// FileInfo contains information about a scanned file
type FileInfo struct {
	Path         string
	Language     string
	Size         int64
	ModTime      time.Time
	HasLinkedDoc bool
}

// NewScanner creates a new filesystem scanner
func NewScanner() *Scanner {
	return &Scanner{
		parser: parser.NewParser(),
	}
}

// Scan recursively scans a directory
func (s *Scanner) Scan(rootPath string, opts ScanOptions) (*ScanResult, error) {
	startTime := time.Now()

	// Resolve absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", absPath)
	}

	// Build ignore matcher
	ignoreMatcher := s.buildIgnoreMatcher(absPath, opts)

	result := &ScanResult{
		Files:  make([]*FileInfo, 0),
		Errors: make([]error, 0),
	}

	// Scan based on concurrency setting
	if opts.Concurrent {
		s.scanConcurrent(absPath, ignoreMatcher, opts, result)
	} else {
		s.scanSequential(absPath, ignoreMatcher, opts, result)
	}

	result.Duration = time.Since(startTime)
	result.TotalFiles = len(result.Files)

	// Calculate total bytes
	for _, file := range result.Files {
		result.TotalBytes += file.Size
	}

	return result, nil
}

// scanSequential performs sequential directory scanning
func (s *Scanner) scanSequential(rootPath string, ignoreMatcher *IgnoreMatcher, opts ScanOptions, result *ScanResult) {
	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil
		}

		// Get relative path for ignore matching
		relPath, _ := filepath.Rel(rootPath, path)

		// Check if should ignore
		if ignoreMatcher.ShouldIgnore(relPath) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Handle symlinks
		if !opts.FollowSymlinks {
			info, err := d.Info()
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
		}

		// Scan the file
		fileInfo, err := s.ScanFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error scanning %s: %w", path, err))
			return nil
		}

		// Check file size limit
		if opts.MaxFileSize > 0 && fileInfo.Size > opts.MaxFileSize {
			return nil
		}

		// Only include source files (not unknown language)
		if fileInfo.Language != "unknown" {
			result.Files = append(result.Files, fileInfo)
		}

		return nil
	})
}

// scanConcurrent performs concurrent directory scanning
func (s *Scanner) scanConcurrent(rootPath string, ignoreMatcher *IgnoreMatcher, opts ScanOptions, result *ScanResult) {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		fileChan = make(chan string, 100)
	)

	// Start worker pool
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				fileInfo, err := s.ScanFile(path)
				if err != nil {
					mu.Lock()
					result.Errors = append(result.Errors, fmt.Errorf("error scanning %s: %w", path, err))
					mu.Unlock()
					continue
				}

				// Check file size limit
				if opts.MaxFileSize > 0 && fileInfo.Size > opts.MaxFileSize {
					continue
				}

				// Only include source files
				if fileInfo.Language != "unknown" {
					mu.Lock()
					result.Files = append(result.Files, fileInfo)
					mu.Unlock()
				}
			}
		}()
	}

	// Walk directory and send files to workers
	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			mu.Lock()
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			mu.Unlock()
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)

		if ignoreMatcher.ShouldIgnore(relPath) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !opts.FollowSymlinks {
			info, err := d.Info()
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
		}

		fileChan <- path
		return nil
	})

	close(fileChan)
	wg.Wait()
}

// ScanFile scans a single file
func (s *Scanner) ScanFile(filePath string) (*FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileInfo := &FileInfo{
		Path:     filePath,
		Language: DetectLanguage(filePath),
		Size:     info.Size(),
		ModTime:  info.ModTime(),
	}

	// Check if file has LinkedDoc (only for source files)
	if fileInfo.Language != "unknown" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			linkedDoc, _ := s.parser.ExtractLinkedDoc(string(content))
			fileInfo.HasLinkedDoc = linkedDoc != ""
		}
	}

	return fileInfo, nil
}

// buildIgnoreMatcher builds the ignore matcher from options
func (s *Scanner) buildIgnoreMatcher(rootPath string, opts ScanOptions) *IgnoreMatcher {
	var patterns []string

	// Add default patterns if requested
	if opts.UseDefaults {
		patterns = append(patterns, DefaultIgnorePatterns()...)
	}

	// Add exclude patterns from options
	patterns = append(patterns, opts.ExcludePatterns...)

	// Load ignore files
	for _, ignoreFile := range opts.IgnoreFiles {
		ignoreFilePath := filepath.Join(rootPath, ignoreFile)
		if content, err := os.ReadFile(ignoreFilePath); err == nil {
			filePatterns := ParseIgnoreFile(string(content))
			patterns = append(patterns, filePatterns...)
		}
	}

	return NewIgnoreMatcher(patterns)
}
