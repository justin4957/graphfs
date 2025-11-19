/*
# Module: pkg/graph/builder.go
Graph builder implementation.

Orchestrates scanner, parser, and triple store to build the knowledge graph.

## Linked Modules
- [graph](./graph.go) - Graph data structure
- [module](./module.go) - Module data structure
- [validator](./validator.go) - Graph validation
- [../../pkg/scanner](../../pkg/scanner/scanner.go) - File scanner
- [../../pkg/parser](../../pkg/parser/parser.go) - LinkedDoc parser
- [../../internal/store](../../internal/store/store.go) - Triple store

## Tags
graph, builder, orchestration

## Exports
Builder, BuildOptions, NewBuilder

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#builder.go> a code:Module ;
    code:name "pkg/graph/builder.go" ;
    code:description "Graph builder implementation" ;
    code:language "go" ;
    code:layer "graph" ;
    code:linksTo <./graph.go>, <./module.go>, <./validator.go>,
                 <../../pkg/scanner/scanner.go>, <../../pkg/parser/parser.go>,
                 <../../internal/store/store.go> ;
    code:exports <#Builder>, <#BuildOptions>, <#NewBuilder> ;
    code:tags "graph", "builder", "orchestration" .
<!-- End LinkedDoc RDF -->
*/

package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/cache"
	"github.com/justin4957/graphfs/pkg/parser"
	"github.com/justin4957/graphfs/pkg/scanner"
)

// Builder builds knowledge graphs from codebases
type Builder struct {
	scanner      *scanner.Scanner
	parser       *parser.Parser
	validator    *Validator
	cacheManager *cache.Manager
}

// BuildOptions configures graph building
type BuildOptions struct {
	ScanOptions    scanner.ScanOptions // Scanner configuration
	Validate       bool                // Validate graph after building
	ReportProgress bool                // Report progress during build
	UseCache       bool                // Enable persistent caching
}

// NewBuilder creates a new graph builder
func NewBuilder() *Builder {
	return &Builder{
		scanner:   scanner.NewScanner(),
		parser:    parser.NewParser(),
		validator: NewValidator(),
	}
}

// Build constructs the knowledge graph from a codebase
func (b *Builder) Build(rootPath string, opts BuildOptions) (*Graph, error) {
	startTime := time.Now()

	// Resolve absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Initialize cache if enabled
	if opts.UseCache && b.cacheManager == nil {
		cacheManager, err := cache.NewManager(absRoot)
		if err != nil {
			if opts.ReportProgress {
				fmt.Printf("Warning: failed to initialize cache: %v\n", err)
			}
			opts.UseCache = false // Disable cache if initialization fails
		} else {
			b.cacheManager = cacheManager
			defer b.cacheManager.Close()
		}
	}

	// Create new triple store and graph
	tripleStore := store.NewTripleStore()
	graph := NewGraph(absRoot, tripleStore)

	// Scan for files
	if opts.ReportProgress {
		fmt.Println("Scanning codebase...")
	}

	scanResult, err := b.scanner.Scan(absRoot, opts.ScanOptions)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Report scan errors if any (for partial results)
	if scanResult.Errors.HasErrors() && opts.ReportProgress {
		fmt.Fprint(os.Stderr, scanResult.Errors.Report())
		fmt.Fprintf(os.Stderr, "\n✓ Partial scan results: %d files scanned, %d files failed\n",
			scanResult.FilesScanned, scanResult.FilesFailed)
	}

	if opts.ReportProgress {
		fmt.Printf("Found %d files with LinkedDoc metadata\n", len(scanResult.Files))
	}

	// Parse each file and build graph
	if opts.ReportProgress {
		fmt.Println("Parsing LinkedDoc metadata...")
	}

	// Determine number of workers (use scan workers setting)
	numWorkers := opts.ScanOptions.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	// Use atomic counters for thread-safe cache statistics
	var cacheHits atomic.Int64
	var cacheMisses atomic.Int64

	// Filter files with LinkedDoc
	var linkedDocFiles []scanner.FileInfo
	for _, file := range scanResult.Files {
		if file.HasLinkedDoc {
			linkedDocFiles = append(linkedDocFiles, *file)
		}
	}

	// Process files in parallel
	var wg sync.WaitGroup
	fileChan := make(chan scanner.FileInfo, len(linkedDocFiles))

	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each worker gets its own parser to avoid race conditions
			workerParser := parser.NewParser()

			for file := range fileChan {
				// Try to get module from cache
				if opts.UseCache && b.cacheManager != nil {
					if cachedData, found := b.cacheManager.Get(file.Path); found {
						// Unmarshal the cached module
						var cachedModule Module
						if err := json.Unmarshal(cachedData.ModuleJSON, &cachedModule); err == nil {
							// Add module to graph (thread-safe)
							graph.AddModule(&cachedModule)

							// Restore triples to graph store (thread-safe)
							for _, triple := range cachedData.Triples {
								if err := graph.Store.Add(triple.Subject, triple.Predicate, triple.Object); err != nil {
									// Log error but continue - this shouldn't break the build
									if opts.ReportProgress {
										fmt.Printf("Warning: failed to restore triple for %s: %v\n", file.Path, err)
									}
								}
							}

							cacheHits.Add(1)
							continue
						}
						// If unmarshal fails, fall through to re-parse
					}
					cacheMisses.Add(1)
				}

				if err := b.processFile(file, graph, absRoot, opts.UseCache, workerParser); err != nil {
					if opts.ReportProgress {
						fmt.Printf("Warning: failed to process %s: %v\n", file.Path, err)
					}
				}
			}
		}()
	}

	// Send files to workers
	for _, file := range linkedDocFiles {
		fileChan <- file
	}
	close(fileChan)

	// Wait for all workers to complete
	wg.Wait()

	// Report cache statistics
	if opts.ReportProgress && opts.UseCache && b.cacheManager != nil {
		hits := cacheHits.Load()
		misses := cacheMisses.Load()
		total := hits + misses
		hitRate := 0.0
		if total > 0 {
			hitRate = float64(hits) / float64(total) * 100
		}
		fmt.Printf("Cache: %d hits, %d misses (%.1f%% hit rate)\n", hits, misses, hitRate)
	}

	// Update statistics
	graph.Statistics.TotalTriples = tripleStore.Count()
	graph.Statistics.BuildDuration = time.Since(startTime)

	// Count relationships
	graph.Statistics.TotalRelationships = b.countRelationships(graph)

	// Build dependency graph (reverse dependencies)
	b.buildDependencyGraph(graph)

	// Validate if requested
	if opts.Validate {
		if opts.ReportProgress {
			fmt.Println("Validating graph...")
		}

		validationResult := b.validator.Validate(graph)
		if len(validationResult.Errors) > 0 {
			// Build detailed error message
			errorMsg := fmt.Sprintf("validation failed with %d errors:\n", len(validationResult.Errors))
			for _, err := range validationResult.Errors {
				errorMsg += fmt.Sprintf("  - %s: %s\n", err.Module, err.Message)
			}
			return graph, fmt.Errorf("%s", errorMsg)
		}

		if opts.ReportProgress && len(validationResult.Warnings) > 0 {
			fmt.Printf("Validation completed with %d warnings\n", len(validationResult.Warnings))
		}
	}

	if opts.ReportProgress {
		fmt.Printf("Graph built: %d modules, %d triples, %d relationships in %v\n",
			graph.Statistics.TotalModules,
			graph.Statistics.TotalTriples,
			graph.Statistics.TotalRelationships,
			graph.Statistics.BuildDuration)
	}

	return graph, nil
}

// Rebuild clears and rebuilds the graph
func (b *Builder) Rebuild(rootPath string, opts BuildOptions) (*Graph, error) {
	// Just call Build - it creates a new graph each time
	return b.Build(rootPath, opts)
}

// processFile parses a file and adds it to the graph
func (b *Builder) processFile(file scanner.FileInfo, graph *Graph, rootPath string, useCache bool, p *parser.Parser) error {
	// Parse LinkedDoc metadata
	triples, err := p.Parse(file.Path)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	// Get relative path
	relPath, err := filepath.Rel(rootPath, file.Path)
	if err != nil {
		relPath = file.Path
	}

	// Create module if it doesn't exist
	var moduleURI string
	var module *Module

	// Collect triples for caching
	var cacheTriples []cache.Triple

	// Process triples
	for _, triple := range triples {
		// Add to triple store
		var objectStr string
		switch obj := triple.Object.(type) {
		case parser.LiteralObject:
			objectStr = obj.Value
		case parser.URIObject:
			objectStr = obj.URI
		case parser.BlankNodeObject:
			// Skip blank nodes for now
			continue
		}

		if err := graph.Store.Add(triple.Subject, triple.Predicate, objectStr); err != nil {
			return fmt.Errorf("failed to add triple: %w", err)
		}

		// Store triple for caching
		cacheTriples = append(cacheTriples, cache.Triple{
			Subject:   triple.Subject,
			Predicate: triple.Predicate,
			Object:    objectStr,
		})

		// Extract module information
		if strings.Contains(triple.Predicate, "rdf-syntax-ns#type") &&
			strings.Contains(objectStr, "Module") {
			moduleURI = triple.Subject
			if module == nil {
				module = NewModule(relPath, moduleURI)
			}
		}

		if module != nil && triple.Subject == moduleURI {
			b.extractModuleProperty(module, triple.Predicate, objectStr, relPath)
		}
	}

	// Add module to graph if we found one
	if module != nil {
		graph.AddModule(module)

		// Cache the module and its triples if caching is enabled
		if useCache && b.cacheManager != nil {
			// Ignore cache write errors - caching is not critical
			_ = b.cacheManager.Set(file.Path, module, cacheTriples)
		}
	}

	return nil
}

// extractModuleProperty extracts module properties from RDF predicates
func (b *Builder) extractModuleProperty(module *Module, predicate, value, modulePath string) {
	switch {
	case strings.HasSuffix(predicate, "name"):
		module.Name = value
	case strings.HasSuffix(predicate, "description"):
		module.Description = value
	case strings.HasSuffix(predicate, "language"):
		module.Language = value
	case strings.HasSuffix(predicate, "layer"):
		module.Layer = value
	case strings.HasSuffix(predicate, "linksTo"):
		// Resolve relative path to absolute path relative to project root
		resolvedPath := b.resolveDependencyPath(value, modulePath)
		module.AddDependency(resolvedPath)
	case strings.HasSuffix(predicate, "exports"):
		module.AddExport(value)
	case strings.HasSuffix(predicate, "calls"):
		module.AddCall(value)
	case strings.HasSuffix(predicate, "tags"):
		module.AddTag(value)
	default:
		// Store other properties
		module.AddProperty(predicate, value)
	}
}

// resolveDependencyPath resolves a dependency path relative to the module's location
func (b *Builder) resolveDependencyPath(depPath, modulePath string) string {
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

	// Clean the path to normalize it (removes . and .. components)
	cleanPath := filepath.Clean(resolvedPath)

	return cleanPath
}

// buildDependencyGraph builds reverse dependency relationships
func (b *Builder) buildDependencyGraph(graph *Graph) {
	for _, module := range graph.Modules {
		for _, dep := range module.Dependencies {
			// Find the dependent module
			depModule := b.findModuleByDependency(graph, dep)
			if depModule != nil {
				depModule.AddDependent(module.URI)
			}
		}
	}
}

// findModuleByDependency finds a module by its dependency reference
func (b *Builder) findModuleByDependency(graph *Graph, dep string) *Module {
	// Try direct path match
	if module := graph.GetModule(dep); module != nil {
		return module
	}

	// Try URI match
	for _, module := range graph.Modules {
		if module.URI == dep {
			return module
		}
	}

	// Try name match
	for _, module := range graph.Modules {
		if module.Name == dep || strings.HasSuffix(module.Path, dep) {
			return module
		}
	}

	return nil
}

// countRelationships counts total relationships in the graph
func (b *Builder) countRelationships(graph *Graph) int {
	count := 0
	for _, module := range graph.Modules {
		count += len(module.Dependencies)
		count += len(module.Calls)
	}
	return count
}

// GetTransitiveDependencies returns all transitive dependencies of a module
func (g *Graph) GetTransitiveDependencies(path string) []string {
	module := g.GetModule(path)
	if module == nil {
		return nil
	}

	visited := make(map[string]bool)
	var result []string

	g.getTransitiveDependenciesRecursive(module, visited, &result)

	return result
}

func (g *Graph) getTransitiveDependenciesRecursive(module *Module, visited map[string]bool, result *[]string) {
	if visited[module.URI] {
		return
	}
	visited[module.URI] = true

	for _, dep := range module.Dependencies {
		depModule := g.GetModule(dep)
		if depModule == nil {
			// Try to find by URI
			for _, mod := range g.Modules {
				if mod.URI == dep {
					depModule = mod
					break
				}
			}
		}

		if depModule != nil {
			*result = append(*result, depModule.Path)
			g.getTransitiveDependenciesRecursive(depModule, visited, result)
		}
	}
}

// GetDirectDependencies returns direct dependencies of a module
func (g *Graph) GetDirectDependencies(path string) []string {
	module := g.GetModule(path)
	if module == nil {
		return nil
	}
	return module.Dependencies
}

// GetDependents returns modules that depend on the given module
func (g *Graph) GetDependents(path string) []string {
	module := g.GetModule(path)
	if module == nil {
		return nil
	}
	return module.Dependents
}
