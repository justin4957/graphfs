/*
# Module: pkg/graph/module.go
Module data structure for representing code modules.

Defines the Module type with metadata and relationship information.

## Linked Modules
- [graph](./graph.go) - Graph data structure

## Tags
graph, module, data-structure

## Exports
Module

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#module.go> a code:Module ;
    code:name "pkg/graph/module.go" ;
    code:description "Module data structure for representing code modules" ;
    code:language "go" ;
    code:layer "graph" ;
    code:linksTo <./graph.go> ;
    code:exports <#Module> ;
    code:tags "graph", "module", "data-structure" .
<!-- End LinkedDoc RDF -->
*/

package graph

// Module represents a code module in the knowledge graph
type Module struct {
	// Identity
	Path        string // File path relative to root
	URI         string // Unique URI identifier (e.g., <#main.go>)
	Name        string // Display name
	Description string // Module description

	// Metadata
	Language string   // Programming language
	Layer    string   // Architectural layer (e.g., "services", "utils")
	Tags     []string // Tags for categorization

	// Relationships
	Dependencies []string // Modules this module depends on (linksTo)
	Dependents   []string // Modules that depend on this module (reverse linksTo)
	Exports      []string // Exported symbols/functions
	Calls        []string // Functions this module calls

	// Additional properties
	Properties map[string][]string // Additional RDF properties
}

// NewModule creates a new module
func NewModule(path, uri string) *Module {
	return &Module{
		Path:         path,
		URI:          uri,
		Dependencies: []string{},
		Dependents:   []string{},
		Exports:      []string{},
		Calls:        []string{},
		Tags:         []string{},
		Properties:   make(map[string][]string),
	}
}

// AddDependency adds a dependency to this module
func (m *Module) AddDependency(dep string) {
	// Avoid duplicates
	for _, existing := range m.Dependencies {
		if existing == dep {
			return
		}
	}
	m.Dependencies = append(m.Dependencies, dep)
}

// AddDependent adds a dependent (reverse dependency) to this module
func (m *Module) AddDependent(dep string) {
	// Avoid duplicates
	for _, existing := range m.Dependents {
		if existing == dep {
			return
		}
	}
	m.Dependents = append(m.Dependents, dep)
}

// AddExport adds an exported symbol
func (m *Module) AddExport(export string) {
	// Avoid duplicates
	for _, existing := range m.Exports {
		if existing == export {
			return
		}
	}
	m.Exports = append(m.Exports, export)
}

// AddCall adds a function call
func (m *Module) AddCall(call string) {
	// Avoid duplicates
	for _, existing := range m.Calls {
		if existing == call {
			return
		}
	}
	m.Calls = append(m.Calls, call)
}

// AddTag adds a tag
func (m *Module) AddTag(tag string) {
	// Avoid duplicates
	for _, existing := range m.Tags {
		if existing == tag {
			return
		}
	}
	m.Tags = append(m.Tags, tag)
}

// AddProperty adds a property value
func (m *Module) AddProperty(predicate, value string) {
	m.Properties[predicate] = append(m.Properties[predicate], value)
}

// HasCircularDependency checks if adding a dependency would create a cycle
func (m *Module) HasCircularDependency(target string, graph *Graph) bool {
	return m.hasCircularDependencyRecursive(target, graph, make(map[string]bool))
}

func (m *Module) hasCircularDependencyRecursive(target string, graph *Graph, visited map[string]bool) bool {
	if m.URI == target || m.Path == target {
		return true
	}

	if visited[m.URI] {
		return false
	}
	visited[m.URI] = true

	// Check all dependencies recursively
	for _, dep := range m.Dependencies {
		depModule := graph.GetModule(dep)
		if depModule == nil {
			// Try to find by URI
			for _, mod := range graph.Modules {
				if mod.URI == dep {
					depModule = mod
					break
				}
			}
		}

		if depModule != nil && depModule.hasCircularDependencyRecursive(target, graph, visited) {
			return true
		}
	}

	return false
}
