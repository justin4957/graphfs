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
	"runtime"
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
	UseDefaults     bool          // Use default ignore patterns
	Concurrent      bool          // Enable concurrent scanning
	Workers         int           // Number of parallel workers (0 = NumCPU)
	StrictMode      bool          // Abort on first error
	MaxErrors       int           // Stop after N errors (0 = unlimited)
	Timeout         time.Duration // Overall operation timeout (0 = no timeout)
	FileTimeout     time.Duration // Per-file parse timeout (0 = no timeout)
}

// DefaultScanOptions returns default scan options
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		MaxFileSize:    1024 * 1024, // 1MB
		FollowSymlinks: false,
		IgnoreFiles:    []string{".gitignore", ".graphfsignore"},
		UseDefaults:    true,
		Concurrent:     true,
		Workers:        0, // 0 = use NumCPU
	}
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	Files        []*FileInfo
	TotalFiles   int
	TotalBytes   int64
	Errors       *ErrorCollector
	FilesScanned int
	FilesFailed  int
	Duration     time.Duration
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
		Errors: NewErrorCollector(),
	}

	// Scan based on concurrency setting
	if opts.Concurrent {
		err = s.scanConcurrent(absPath, ignoreMatcher, opts, result)
	} else {
		err = s.scanSequential(absPath, ignoreMatcher, opts, result)
	}

	// Check if scan was aborted due to strict mode or max errors
	if err != nil {
		return nil, err
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
func (s *Scanner) scanSequential(rootPath string, ignoreMatcher *IgnoreMatcher, opts ScanOptions, result *ScanResult) error {
	var scanErr error

	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors.Add(path, err)
			result.FilesScanned++
			result.FilesFailed++

			// Strict mode: abort immediately
			if opts.StrictMode {
				scanErr = fmt.Errorf("strict mode: %s: %w", path, err)
				return scanErr
			}

			// Max errors check
			if opts.MaxErrors > 0 && result.FilesFailed >= opts.MaxErrors {
				scanErr = fmt.Errorf("max errors (%d) reached", opts.MaxErrors)
				return scanErr
			}

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

		result.FilesScanned++

		// Scan the file
		fileInfo, err := s.ScanFile(path)
		if err != nil {
			result.Errors.Add(path, err)
			result.FilesFailed++

			// Strict mode: abort immediately
			if opts.StrictMode {
				scanErr = fmt.Errorf("strict mode: %s: %w", path, err)
				return scanErr
			}

			// Max errors check
			if opts.MaxErrors > 0 && result.FilesFailed >= opts.MaxErrors {
				scanErr = fmt.Errorf("max errors (%d) reached", opts.MaxErrors)
				return scanErr
			}

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

	return scanErr
}

// scanConcurrent performs concurrent directory scanning
func (s *Scanner) scanConcurrent(rootPath string, ignoreMatcher *IgnoreMatcher, opts ScanOptions, result *ScanResult) error {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		fileChan  = make(chan string, 100)
		errorChan = make(chan error, 1)
		scanErr   error
	)

	// Determine number of workers
	numWorkers := opts.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				mu.Lock()
				result.FilesScanned++
				mu.Unlock()

				fileInfo, err := s.ScanFile(path)
				if err != nil {
					mu.Lock()
					result.Errors.Add(path, err)
					result.FilesFailed++
					failCount := result.FilesFailed
					mu.Unlock()

					// Strict mode: signal error
					if opts.StrictMode {
						select {
						case errorChan <- fmt.Errorf("strict mode: %s: %w", path, err):
						default:
						}
						return
					}

					// Max errors check
					if opts.MaxErrors > 0 && failCount >= opts.MaxErrors {
						select {
						case errorChan <- fmt.Errorf("max errors (%d) reached", opts.MaxErrors):
						default:
						}
						return
					}

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
		// Check if we hit an error from workers
		select {
		case scanErr = <-errorChan:
			return scanErr
		default:
		}

		if err != nil {
			mu.Lock()
			result.Errors.Add(path, err)
			result.FilesScanned++
			result.FilesFailed++
			failCount := result.FilesFailed
			mu.Unlock()

			// Strict mode: abort immediately
			if opts.StrictMode {
				scanErr = fmt.Errorf("strict mode: %s: %w", path, err)
				return scanErr
			}

			// Max errors check
			if opts.MaxErrors > 0 && failCount >= opts.MaxErrors {
				scanErr = fmt.Errorf("max errors (%d) reached", opts.MaxErrors)
				return scanErr
			}

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

	// Check for errors from workers after completion
	select {
	case scanErr = <-errorChan:
	default:
	}

	return scanErr
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
