/*
# Module: pkg/graph/graph.go
Graph data structures for knowledge graph representation.

Defines the Graph and GraphStats types for managing the codebase knowledge graph.

## Linked Modules
- [builder](./builder.go) - Graph builder
- [module](./module.go) - Module data structure
- [../../internal/store](../../internal/store/store.go) - Triple store

## Tags
graph, knowledge-graph, data-structure

## Exports
Graph, GraphStats

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#graph.go> a code:Module ;
    code:name "pkg/graph/graph.go" ;
    code:description "Graph data structures for knowledge graph representation" ;
    code:language "go" ;
    code:layer "graph" ;
    code:linksTo <./builder.go>, <./module.go>, <../../internal/store/store.go> ;
    code:exports <#Graph>, <#GraphStats> ;
    code:tags "graph", "knowledge-graph", "data-structure" .
<!-- End LinkedDoc RDF -->
*/

package graph

import (
	"sync"
	"time"

	"github.com/justin4957/graphfs/internal/store"
)

// Graph represents a codebase knowledge graph
type Graph struct {
	Store      *store.TripleStore // Triple store containing all RDF triples
	Root       string             // Root directory path
	Modules    map[string]*Module // Modules indexed by path
	Statistics GraphStats         // Graph statistics
	mu         sync.Mutex         // Mutex for thread-safe operations
}

// GraphStats provides statistics about the knowledge graph
type GraphStats struct {
	TotalModules       int            // Total number of modules
	TotalTriples       int            // Total number of RDF triples
	TotalRelationships int            // Total number of relationships (linksTo, calls, etc.)
	ModulesByLanguage  map[string]int // Modules grouped by language
	ModulesByLayer     map[string]int // Modules grouped by layer
	BuildDuration      time.Duration  // Time taken to build graph
}

// NewGraph creates a new empty graph
func NewGraph(root string, tripleStore *store.TripleStore) *Graph {
	return &Graph{
		Store:   tripleStore,
		Root:    root,
		Modules: make(map[string]*Module),
		Statistics: GraphStats{
			ModulesByLanguage: make(map[string]int),
			ModulesByLayer:    make(map[string]int),
		},
	}
}

// GetModule returns a module by its path
func (g *Graph) GetModule(path string) *Module {
	return g.Modules[path]
}

// AddModule adds a module to the graph (thread-safe)
func (g *Graph) AddModule(module *Module) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Modules[module.Path] = module
	g.Statistics.TotalModules++

	// Update statistics
	if module.Language != "" {
		g.Statistics.ModulesByLanguage[module.Language]++
	}
	if module.Layer != "" {
		g.Statistics.ModulesByLayer[module.Layer]++
	}
}

// RemoveModule removes a module from the graph
func (g *Graph) RemoveModule(path string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	module, exists := g.Modules[path]
	if !exists {
		return
	}

	// Update statistics
	g.Statistics.TotalModules--
	if module.Language != "" {
		g.Statistics.ModulesByLanguage[module.Language]--
		if g.Statistics.ModulesByLanguage[module.Language] <= 0 {
			delete(g.Statistics.ModulesByLanguage, module.Language)
		}
	}
	if module.Layer != "" {
		g.Statistics.ModulesByLayer[module.Layer]--
		if g.Statistics.ModulesByLayer[module.Layer] <= 0 {
			delete(g.Statistics.ModulesByLayer, module.Layer)
		}
	}

	delete(g.Modules, path)
}

// GetModulesByLanguage returns all modules for a given language
func (g *Graph) GetModulesByLanguage(language string) []*Module {
	var modules []*Module
	for _, module := range g.Modules {
		if module.Language == language {
			modules = append(modules, module)
		}
	}
	return modules
}

// GetModulesByLayer returns all modules for a given layer
func (g *Graph) GetModulesByLayer(layer string) []*Module {
	var modules []*Module
	for _, module := range g.Modules {
		if module.Layer == layer {
			modules = append(modules, module)
		}
	}
	return modules
}

// GetModulesByTag returns all modules with a given tag
func (g *Graph) GetModulesByTag(tag string) []*Module {
	var modules []*Module
	for _, module := range g.Modules {
		for _, t := range module.Tags {
			if t == tag {
				modules = append(modules, module)
				break
			}
		}
	}
	return modules
}
