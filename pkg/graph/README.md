# Graph Builder

Knowledge graph builder for GraphFS that orchestrates scanner, parser, and triple store to build and query codebase knowledge graphs.

## Features

- **Graph Building**: Scan and parse codebases into RDF knowledge graphs
- **Module Management**: Extract and manage code module metadata
- **Dependency Tracking**: Build dependency graphs with reverse lookups
- **Validation**: Comprehensive graph validation with error and warning reporting
- **Query Helpers**: Convenient methods for querying modules by language, layer, tags
- **Statistics**: Detailed graph statistics and metrics

## Installation

```bash
go get github.com/justin4957/graphfs/pkg/graph
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/justin4957/graphfs/pkg/graph"
    "github.com/justin4957/graphfs/pkg/scanner"
)

func main() {
    // Create builder
    builder := graph.NewBuilder()

    // Build knowledge graph from codebase
    g, err := builder.Build("/path/to/codebase", graph.BuildOptions{
        ScanOptions: scanner.ScanOptions{
            UseDefaults: true,
        },
        Validate:       true,
        ReportProgress: true,
    })

    if err != nil {
        panic(err)
    }

    fmt.Printf("Built graph: %d modules, %d triples\n",
        g.Statistics.TotalModules,
        g.Statistics.TotalTriples)
}
```

## Core Types

### Graph

The main knowledge graph structure:

```go
type Graph struct {
    Store      *store.TripleStore    // RDF triple store
    Root       string                // Root directory path
    Modules    map[string]*Module    // Modules indexed by path
    Statistics GraphStats            // Graph statistics
}
```

### Module

Represents a code module:

```go
type Module struct {
    // Identity
    Path        string
    URI         string
    Name        string
    Description string

    // Metadata
    Language string
    Layer    string
    Tags     []string

    // Relationships
    Dependencies []string  // Modules this depends on
    Dependents   []string  // Modules that depend on this
    Exports      []string  // Exported symbols
    Calls        []string  // Function calls

    // Additional properties
    Properties map[string][]string
}
```

### Builder

Orchestrates graph construction:

```go
type Builder struct {
    // Internal scanners, parsers, validators
}

func NewBuilder() *Builder

func (b *Builder) Build(rootPath string, opts BuildOptions) (*Graph, error)
func (b *Builder) Rebuild(rootPath string, opts BuildOptions) (*Graph, error)
```

## Usage Examples

### Build a Graph

```go
builder := graph.NewBuilder()

g, err := builder.Build("./myproject", graph.BuildOptions{
    ScanOptions: scanner.ScanOptions{
        UseDefaults: true,
        Concurrent:  true,
    },
    Validate:       true,
    ReportProgress: true,
})
```

### Query Modules

```go
// Get modules by language
goModules := g.GetModulesByLanguage("go")
fmt.Printf("Found %d Go modules\n", len(goModules))

// Get modules by layer
serviceModules := g.GetModulesByLayer("services")

// Get modules by tag
entrypoints := g.GetModulesByTag("entrypoint")
```

### Module Metadata

```go
module := g.GetModule("main.go")

fmt.Printf("Name: %s\n", module.Name)
fmt.Printf("Language: %s\n", module.Language)
fmt.Printf("Dependencies: %v\n", module.Dependencies)
fmt.Printf("Exports: %v\n", module.Exports)
```

### Dependency Analysis

```go
// Get direct dependencies
directDeps := g.GetDirectDependencies("main.go")

// Get all transitive dependencies
allDeps := g.GetTransitiveDependencies("main.go")

// Get modules that depend on this one (reverse dependencies)
dependents := g.GetDependents("utils.go")
```

### Validation

```go
validator := graph.NewValidator()
result := validator.Validate(g)

// Check for errors
if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        fmt.Printf("Error in %s: %s\n", err.Module, err.Message)
    }
}

// Check for warnings
if len(result.Warnings) > 0 {
    for _, warn := range result.Warnings {
        fmt.Printf("Warning in %s: %s\n", warn.Module, warn.Message)
    }
}
```

## Build Options

```go
type BuildOptions struct {
    ScanOptions     scanner.ScanOptions  // Scanner configuration
    Validate        bool                 // Validate graph after building
    ReportProgress  bool                 // Print progress messages
}
```

## Graph Statistics

The `GraphStats` type provides metrics about the graph:

```go
type GraphStats struct {
    TotalModules       int                 // Total number of modules
    TotalTriples       int                 // Total RDF triples
    TotalRelationships int                 // Total relationships
    ModulesByLanguage  map[string]int      // Modules per language
    ModulesByLayer     map[string]int      // Modules per layer
    BuildDuration      time.Duration       // Build time
}
```

Access statistics:

```go
fmt.Printf("Modules: %d\n", g.Statistics.TotalModules)
fmt.Printf("Triples: %d\n", g.Statistics.TotalTriples)
fmt.Printf("Go modules: %d\n", g.Statistics.ModulesByLanguage["go"])
fmt.Printf("Build time: %v\n", g.Statistics.BuildDuration)
```

## Validation

The validator checks for:

### Errors
- Missing required fields (name)
- Missing URIs
- Duplicate module URIs
- Circular dependencies

### Warnings
- Missing descriptions
- Missing language
- Non-existent dependencies
- Missing tags or exports
- Too many dependencies (>10)
- Incorrect URI format

Example:

```go
validator := graph.NewValidator()
result := validator.Validate(g)

fmt.Printf("%d errors, %d warnings\n",
    len(result.Errors),
    len(result.Warnings))
```

## Query Helpers

### Module Queries

```go
// Get specific module
module := g.GetModule("main.go")

// Get modules by criteria
goModules := g.GetModulesByLanguage("go")
serviceModules := g.GetModulesByLayer("services")
coreModules := g.GetModulesByTag("core")
```

### Dependency Queries

```go
// Direct dependencies of a module
deps := g.GetDirectDependencies("main.go")

// All transitive dependencies
allDeps := g.GetTransitiveDependencies("main.go")

// Modules that depend on this module
dependents := g.GetDependents("utils.go")
```

### Circular Dependency Detection

```go
module := g.GetModule("a.go")

// Check if adding a dependency would create a cycle
if module.HasCircularDependency("b.go", g) {
    fmt.Println("Circular dependency detected!")
}
```

## Integration with GraphFS Pipeline

The graph builder integrates with other GraphFS components:

```go
// 1. Scan codebase
// (handled internally by builder)

// 2. Parse LinkedDoc metadata
// (handled internally by builder)

// 3. Build knowledge graph
builder := graph.NewBuilder()
g, _ := builder.Build("/path/to/code", graph.BuildOptions{})

// 4. Query the graph
executor := query.NewExecutor(g.Store)
queryResult, _ := executor.ExecuteString(`
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?name WHERE {
        ?module code:name ?name .
    }
`)
```

## Performance

- **Build Time**: ~1ms per module for typical codebases
- **Memory**: O(n) where n is number of triples
- **Concurrent Scanning**: Parallel file processing for faster builds
- **Indexed Lookups**: O(1) module lookups by path

Typical performance for examples/minimal-app (7 modules, 267 triples):
- Build time: < 1ms
- Memory: < 1MB

## Testing

```bash
# Unit tests
go test ./pkg/graph

# Integration tests (requires examples/minimal-app)
go test ./pkg/graph -run TestIntegration

# Example tests
go test ./pkg/graph -run Example
```

## API Reference

### Builder Methods

```go
func NewBuilder() *Builder
func (b *Builder) Build(rootPath string, opts BuildOptions) (*Graph, error)
func (b *Builder) Rebuild(rootPath string, opts BuildOptions) (*Graph, error)
```

### Graph Methods

```go
func NewGraph(root string, tripleStore *store.TripleStore) *Graph
func (g *Graph) GetModule(path string) *Module
func (g *Graph) AddModule(module *Module)
func (g *Graph) GetModulesByLanguage(language string) []*Module
func (g *Graph) GetModulesByLayer(layer string) []*Module
func (g *Graph) GetModulesByTag(tag string) []*Module
func (g *Graph) GetDirectDependencies(path string) []string
func (g *Graph) GetTransitiveDependencies(path string) []string
func (g *Graph) GetDependents(path string) []string
```

### Module Methods

```go
func NewModule(path, uri string) *Module
func (m *Module) AddDependency(dep string)
func (m *Module) AddDependent(dep string)
func (m *Module) AddExport(export string)
func (m *Module) AddCall(call string)
func (m *Module) AddTag(tag string)
func (m *Module) AddProperty(predicate, value string)
func (m *Module) HasCircularDependency(target string, graph *Graph) bool
```

### Validator Methods

```go
func NewValidator() *Validator
func (v *Validator) Validate(graph *Graph) ValidationResult
```

## Examples

See [example_test.go](example_test.go) for complete working examples.

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](../../LICENSE) for details.
