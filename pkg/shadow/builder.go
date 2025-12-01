/*
# Module: pkg/shadow/builder.go
Shadow builder for generating shadow entries from source files.

Integrates with the graph builder to create shadow entries from parsed
LinkedDoc metadata, enabling automatic shadow file generation.

## Linked Modules
- [shadow](./shadow.go) - Shadow file system manager
- [entry](./entry.go) - Shadow entry data structure
- [../graph/builder](../graph/builder.go) - Graph builder

## Tags
shadow, builder, generation, integration

## Exports
Builder, NewBuilder, BuildOptions

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#builder.go> a code:Module ;
    code:name "pkg/shadow/builder.go" ;
    code:description "Shadow builder for generating shadow entries from source files" ;
    code:language "go" ;
    code:layer "shadow" ;
    code:linksTo <./shadow.go>, <./entry.go>, <../graph/builder.go> ;
    code:exports <#Builder>, <#NewBuilder>, <#BuildOptions> ;
    code:tags "shadow", "builder", "generation", "integration" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/justin4957/graphfs/pkg/parser"
	"github.com/justin4957/graphfs/pkg/scanner"
)

// BuildOptions configures shadow building
type BuildOptions struct {
	// ScanOptions configures file scanning
	ScanOptions scanner.ScanOptions

	// MergeExisting preserves existing manual annotations
	MergeExisting bool

	// ForceOverwrite overwrites all existing entries
	ForceOverwrite bool

	// ReportProgress reports build progress
	ReportProgress bool

	// Workers is the number of parallel workers (0 = NumCPU)
	Workers int

	// IncludeTriples includes raw RDF triples in shadow entries
	IncludeTriples bool

	// SkipUnchanged skips files that haven't changed since last build
	SkipUnchanged bool
}

// DefaultBuildOptions returns default build options
func DefaultBuildOptions() BuildOptions {
	return BuildOptions{
		ScanOptions:    scanner.ScanOptions{UseDefaults: true},
		MergeExisting:  true,
		ForceOverwrite: false,
		ReportProgress: false,
		Workers:        0,
		IncludeTriples: true,
		SkipUnchanged:  true,
	}
}

// Builder builds shadow entries from source files
type Builder struct {
	shadowFS *ShadowFS
	scanner  *scanner.Scanner
	parser   *parser.Parser
}

// BuildResult contains the results of a shadow build operation
type BuildResult struct {
	TotalFiles     int
	ProcessedFiles int
	SkippedFiles   int
	NewEntries     int
	UpdatedEntries int
	MergedEntries  int
	Errors         []BuildError
	Duration       time.Duration
}

// BuildError represents an error during shadow building
type BuildError struct {
	Path    string
	Message string
	Err     error
}

func (e BuildError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// NewBuilder creates a new shadow builder
func NewBuilder(shadowFS *ShadowFS) *Builder {
	return &Builder{
		shadowFS: shadowFS,
		scanner:  scanner.NewScanner(),
		parser:   parser.NewParser(),
	}
}

// Build generates shadow entries for all source files
func (b *Builder) Build(opts BuildOptions) (*BuildResult, error) {
	startTime := time.Now()
	result := &BuildResult{}

	// Scan for files
	if opts.ReportProgress {
		fmt.Println("Scanning codebase for LinkedDoc files...")
	}

	scanResult, err := b.scanner.Scan(b.shadowFS.RootPath(), opts.ScanOptions)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Filter files with LinkedDoc
	var linkedDocFiles []scanner.FileInfo
	for _, file := range scanResult.Files {
		if file.HasLinkedDoc {
			linkedDocFiles = append(linkedDocFiles, *file)
		}
	}

	result.TotalFiles = len(linkedDocFiles)

	if opts.ReportProgress {
		fmt.Printf("Found %d files with LinkedDoc metadata\n", result.TotalFiles)
	}

	// Determine number of workers
	numWorkers := opts.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	// Process files in parallel
	var wg sync.WaitGroup
	fileChan := make(chan scanner.FileInfo, len(linkedDocFiles))
	resultChan := make(chan fileResult, len(linkedDocFiles))

	var processed, skipped, newEntries, updated, merged atomic.Int64
	var errorsMu sync.Mutex
	var errors []BuildError

	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerParser := parser.NewParser()

			for file := range fileChan {
				fr := b.processFile(file, workerParser, opts)
				resultChan <- fr
			}
		}()
	}

	// Send files to workers
	go func() {
		for _, file := range linkedDocFiles {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for fr := range resultChan {
		if fr.err != nil {
			errorsMu.Lock()
			errors = append(errors, BuildError{
				Path:    fr.path,
				Message: fr.err.Error(),
				Err:     fr.err,
			})
			errorsMu.Unlock()
			continue
		}

		processed.Add(1)

		switch fr.status {
		case statusSkipped:
			skipped.Add(1)
		case statusNew:
			newEntries.Add(1)
		case statusUpdated:
			updated.Add(1)
		case statusMerged:
			merged.Add(1)
		}
	}

	result.ProcessedFiles = int(processed.Load())
	result.SkippedFiles = int(skipped.Load())
	result.NewEntries = int(newEntries.Load())
	result.UpdatedEntries = int(updated.Load())
	result.MergedEntries = int(merged.Load())
	result.Errors = errors
	result.Duration = time.Since(startTime)

	// Save index
	if err := b.shadowFS.SaveIndex(); err != nil {
		return result, fmt.Errorf("failed to save index: %w", err)
	}

	if opts.ReportProgress {
		fmt.Printf("Shadow build complete: %d new, %d updated, %d merged, %d skipped in %v\n",
			result.NewEntries, result.UpdatedEntries, result.MergedEntries, result.SkippedFiles, result.Duration)
		if len(result.Errors) > 0 {
			fmt.Printf("Encountered %d errors during build\n", len(result.Errors))
		}
	}

	return result, nil
}

// BuildFile generates a shadow entry for a single file
func (b *Builder) BuildFile(sourcePath string, opts BuildOptions) error {
	// Resolve absolute path
	absPath := sourcePath
	if !filepath.IsAbs(sourcePath) {
		absPath = filepath.Join(b.shadowFS.RootPath(), sourcePath)
	}

	// Check file exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Create file info
	fileInfo := scanner.FileInfo{
		Path:         absPath,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		HasLinkedDoc: true,
	}

	fr := b.processFile(fileInfo, b.parser, opts)
	if fr.err != nil {
		return fr.err
	}

	return nil
}

type resultStatus int

const (
	statusSkipped resultStatus = iota
	statusNew
	statusUpdated
	statusMerged
)

type fileResult struct {
	path   string
	status resultStatus
	err    error
}

// processFile processes a single file and creates/updates shadow entry
func (b *Builder) processFile(file scanner.FileInfo, p *parser.Parser, opts BuildOptions) fileResult {
	result := fileResult{path: file.Path}

	// Calculate file hash
	hash, err := calculateFileHash(file.Path)
	if err != nil {
		result.err = fmt.Errorf("failed to calculate hash: %w", err)
		return result
	}

	// Get relative path
	relPath, err := filepath.Rel(b.shadowFS.RootPath(), file.Path)
	if err != nil {
		relPath = file.Path
	}

	// Check if we should skip unchanged files
	if opts.SkipUnchanged && !opts.ForceOverwrite {
		existing, _ := b.shadowFS.Get(file.Path)
		if existing != nil && existing.SourceHash == hash {
			result.status = statusSkipped
			return result
		}
	}

	// Parse LinkedDoc metadata
	triples, err := p.Parse(file.Path)
	if err != nil {
		result.err = fmt.Errorf("failed to parse: %w", err)
		return result
	}

	// Create shadow entry
	entry := NewAutoEntry(relPath)
	entry.SourceHash = hash

	// Extract module information from triples
	b.extractModuleInfo(entry, triples, relPath)

	// Add raw triples if enabled
	if opts.IncludeTriples {
		for _, t := range triples {
			var objStr string
			switch obj := t.Object.(type) {
			case parser.LiteralObject:
				objStr = obj.Value
			case parser.URIObject:
				objStr = obj.URI
			default:
				continue
			}

			entry.AddTriple(t.Subject, t.Predicate, objStr, SourceAuto)
		}
	}

	// Handle existing entries
	existing, _ := b.shadowFS.Get(file.Path)
	if existing != nil {
		if opts.ForceOverwrite {
			if err := b.shadowFS.Set(file.Path, entry); err != nil {
				result.err = err
				return result
			}
			result.status = statusUpdated
		} else if opts.MergeExisting {
			if err := b.shadowFS.Merge(file.Path, entry); err != nil {
				result.err = err
				return result
			}
			result.status = statusMerged
		} else {
			result.status = statusSkipped
		}
	} else {
		if err := b.shadowFS.Set(file.Path, entry); err != nil {
			result.err = err
			return result
		}
		result.status = statusNew
	}

	return result
}

// extractModuleInfo extracts module information from parsed triples
func (b *Builder) extractModuleInfo(entry *Entry, triples []parser.Triple, modulePath string) {
	var moduleURI string
	var name, description, language, layer string
	var tags []string
	var dependencies []string
	var exports []string
	var calls []string

	// First pass: find module URI
	for _, t := range triples {
		var objStr string
		switch obj := t.Object.(type) {
		case parser.LiteralObject:
			objStr = obj.Value
		case parser.URIObject:
			objStr = obj.URI
		default:
			continue
		}

		if strings.Contains(t.Predicate, "rdf-syntax-ns#type") && strings.Contains(objStr, "Module") {
			moduleURI = t.Subject
			break
		}
	}

	// Second pass: extract properties for the module
	for _, t := range triples {
		if t.Subject != moduleURI && moduleURI != "" {
			continue
		}

		var objStr string
		switch obj := t.Object.(type) {
		case parser.LiteralObject:
			objStr = obj.Value
		case parser.URIObject:
			objStr = obj.URI
		default:
			continue
		}

		switch {
		case strings.HasSuffix(t.Predicate, "name"):
			name = objStr
		case strings.HasSuffix(t.Predicate, "description"):
			description = objStr
		case strings.HasSuffix(t.Predicate, "language"):
			language = objStr
		case strings.HasSuffix(t.Predicate, "layer"):
			layer = objStr
		case strings.HasSuffix(t.Predicate, "tags"):
			tags = append(tags, objStr)
		case strings.HasSuffix(t.Predicate, "linksTo"):
			dep := resolveDependencyPath(objStr, modulePath)
			dependencies = append(dependencies, dep)
		case strings.HasSuffix(t.Predicate, "exports"):
			exports = append(exports, objStr)
		case strings.HasSuffix(t.Predicate, "calls"):
			calls = append(calls, objStr)
		}
	}

	// Set module info
	if moduleURI != "" || name != "" {
		entry.SetModule(moduleURI, name, description, language, layer, tags)
	}

	// Add relationships
	for _, dep := range dependencies {
		entry.AddDependency("linksTo", dep, SourceAuto)
	}

	// Add exports and calls
	for _, exp := range exports {
		entry.AddExport(exp)
	}
	for _, call := range calls {
		entry.AddCall(call)
	}
}

// Clean removes shadow entries for files that no longer exist
func (b *Builder) Clean(opts BuildOptions) (*CleanResult, error) {
	startTime := time.Now()
	result := &CleanResult{}

	if opts.ReportProgress {
		fmt.Println("Cleaning orphaned shadow entries...")
	}

	// Get all shadow entries
	entries, err := b.shadowFS.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list shadow entries: %w", err)
	}

	result.TotalEntries = len(entries)

	// Check each entry
	for _, entry := range entries {
		sourcePath := filepath.Join(b.shadowFS.RootPath(), entry.SourcePath)

		// Check if source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// File no longer exists
			result.OrphanedEntries = append(result.OrphanedEntries, entry.SourcePath)

			// Delete shadow entry
			if err := b.shadowFS.Delete(sourcePath); err != nil {
				result.Errors = append(result.Errors, BuildError{
					Path:    entry.SourcePath,
					Message: "failed to delete orphaned entry",
					Err:     err,
				})
			} else {
				result.RemovedEntries++
			}
		}
	}

	result.Duration = time.Since(startTime)

	// Save index
	if err := b.shadowFS.SaveIndex(); err != nil {
		return result, fmt.Errorf("failed to save index: %w", err)
	}

	if opts.ReportProgress {
		fmt.Printf("Clean complete: %d orphaned entries removed in %v\n",
			result.RemovedEntries, result.Duration)
	}

	return result, nil
}

// CleanResult contains the results of a clean operation
type CleanResult struct {
	TotalEntries    int
	OrphanedEntries []string
	RemovedEntries  int
	Errors          []BuildError
	Duration        time.Duration
}

// Sync performs a full build and clean
func (b *Builder) Sync(opts BuildOptions) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{}

	// Build new/updated entries
	buildResult, err := b.Build(opts)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}
	result.BuildResult = buildResult

	// Clean orphaned entries
	cleanResult, err := b.Clean(opts)
	if err != nil {
		return nil, fmt.Errorf("clean failed: %w", err)
	}
	result.CleanResult = cleanResult

	result.Duration = time.Since(startTime)

	if opts.ReportProgress {
		fmt.Printf("Sync complete in %v\n", result.Duration)
	}

	return result, nil
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	BuildResult *BuildResult
	CleanResult *CleanResult
	Duration    time.Duration
}

// calculateFileHash computes SHA256 hash of file contents
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// resolveDependencyPath resolves a dependency path relative to the module's location
func resolveDependencyPath(depPath, modulePath string) string {
	// Remove angle brackets if present (RDF URI notation)
	depPath = strings.TrimPrefix(depPath, "<")
	depPath = strings.TrimSuffix(depPath, ">")

	// If it's already an absolute path or doesn't contain relative markers, return as-is
	if !strings.Contains(depPath, "..") && !strings.HasPrefix(depPath, "./") {
		return depPath
	}

	// Get the directory of the module
	moduleDir := filepath.Dir(modulePath)

	// Resolve the relative path
	resolvedPath := filepath.Join(moduleDir, depPath)

	// Clean the path to normalize it
	return filepath.Clean(resolvedPath)
}
