/*
# Module: pkg/schema/graphql/generator.go
GraphQL schema generator for GraphFS knowledge graphs.

Generates GraphQL Schema Definition Language (SDL) from RDF knowledge graphs.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [../../../internal/store](../../../internal/store/store.go) - Triple store

## Tags
graphql, schema, generator, sdl

## Exports
Generator, GenerateSchema, GenerateOptions

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#generator.go> a code:Module ;
    code:name "pkg/schema/graphql/generator.go" ;
    code:description "GraphQL schema generator for GraphFS knowledge graphs" ;
    code:language "go" ;
    code:layer "schema" ;
    code:linksTo <../../graph/graph.go>, <../../../internal/store/store.go> ;
    code:exports <#Generator>, <#GenerateSchema>, <#GenerateOptions> ;
    code:tags "graphql", "schema", "generator", "sdl" .
<!-- End LinkedDoc RDF -->
*/

package graphql

import (
	"fmt"
	"sort"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
)

// Generator generates GraphQL schemas from knowledge graphs
type Generator struct {
	graph *graph.Graph
}

// GenerateOptions configures schema generation
type GenerateOptions struct {
	IncludeTypes []string          // Types to include (empty = all)
	ExcludeTypes []string          // Types to exclude
	CustomTypes  map[string]string // Custom type mappings
}

// DefaultGenerateOptions returns default options
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		IncludeTypes: []string{},
		ExcludeTypes: []string{},
		CustomTypes:  make(map[string]string),
	}
}

// NewGenerator creates a new GraphQL schema generator
func NewGenerator(g *graph.Graph) *Generator {
	return &Generator{
		graph: g,
	}
}

// GenerateSchema generates GraphQL SDL from the knowledge graph
func (gen *Generator) GenerateSchema(opts GenerateOptions) (string, error) {
	var sb strings.Builder

	// Write header
	sb.WriteString("# GraphQL Schema for GraphFS Knowledge Graph\n")
	sb.WriteString("# Auto-generated from RDF knowledge graph\n")
	sb.WriteString("# Do not edit manually\n\n")

	// Write schema definition
	sb.WriteString("schema {\n")
	sb.WriteString("  query: Query\n")
	sb.WriteString("}\n\n")

	// Generate Module type
	sb.WriteString(gen.generateModuleType())
	sb.WriteString("\n")

	// Generate Export type
	sb.WriteString(gen.generateExportType())
	sb.WriteString("\n")

	// Generate Query type
	sb.WriteString(gen.generateQueryType())
	sb.WriteString("\n")

	// Generate connection types for pagination
	sb.WriteString(gen.generateConnectionTypes())

	return sb.String(), nil
}

// generateModuleType generates the Module GraphQL type
func (gen *Generator) generateModuleType() string {
	var sb strings.Builder

	sb.WriteString("\"\"\"Represents a code module in the knowledge graph\"\"\"\n")
	sb.WriteString("type Module {\n")
	sb.WriteString("  \"\"\"Unique identifier\"\"\"\n")
	sb.WriteString("  id: ID!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"URI identifier (e.g., <#main.go>)\"\"\"\n")
	sb.WriteString("  uri: String!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Module name\"\"\"\n")
	sb.WriteString("  name: String!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Module description\"\"\"\n")
	sb.WriteString("  description: String\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"File path relative to root\"\"\"\n")
	sb.WriteString("  path: String!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Programming language\"\"\"\n")
	sb.WriteString("  language: String\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Architectural layer\"\"\"\n")
	sb.WriteString("  layer: String\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Tags for categorization\"\"\"\n")
	sb.WriteString("  tags: [String!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Modules this module depends on\"\"\"\n")
	sb.WriteString("  dependencies: [Module!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Modules that depend on this module\"\"\"\n")
	sb.WriteString("  dependents: [Module!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Exported symbols/functions\"\"\"\n")
	sb.WriteString("  exports: [Export!]!\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateExportType generates the Export GraphQL type
func (gen *Generator) generateExportType() string {
	var sb strings.Builder

	sb.WriteString("\"\"\"Represents an exported symbol from a module\"\"\"\n")
	sb.WriteString("type Export {\n")
	sb.WriteString("  \"\"\"Export name\"\"\"\n")
	sb.WriteString("  name: String!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Module that exports this symbol\"\"\"\n")
	sb.WriteString("  module: Module!\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateQueryType generates the Query root type
func (gen *Generator) generateQueryType() string {
	var sb strings.Builder

	sb.WriteString("\"\"\"Root query type\"\"\"\n")
	sb.WriteString("type Query {\n")
	sb.WriteString("  \"\"\"Get a single module by name or path\"\"\"\n")
	sb.WriteString("  module(name: String, path: String, uri: String): Module\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"List all modules with optional filtering\"\"\"\n")
	sb.WriteString("  modules(\n")
	sb.WriteString("    \"\"\"Filter by programming language\"\"\"\n")
	sb.WriteString("    language: String\n")
	sb.WriteString("    \n")
	sb.WriteString("    \"\"\"Filter by architectural layer\"\"\"\n")
	sb.WriteString("    layer: String\n")
	sb.WriteString("    \n")
	sb.WriteString("    \"\"\"Filter by tag\"\"\"\n")
	sb.WriteString("    tag: String\n")
	sb.WriteString("    \n")
	sb.WriteString("    \"\"\"Maximum number of results\"\"\"\n")
	sb.WriteString("    first: Int\n")
	sb.WriteString("    \n")
	sb.WriteString("    \"\"\"Cursor for pagination\"\"\"\n")
	sb.WriteString("    after: String\n")
	sb.WriteString("  ): ModuleConnection!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Search modules by description\"\"\"\n")
	sb.WriteString("  searchModules(query: String!): [Module!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Get graph statistics\"\"\"\n")
	sb.WriteString("  stats: GraphStats!\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateConnectionTypes generates pagination connection types
func (gen *Generator) generateConnectionTypes() string {
	var sb strings.Builder

	// ModuleConnection
	sb.WriteString("\"\"\"Connection type for module pagination\"\"\"\n")
	sb.WriteString("type ModuleConnection {\n")
	sb.WriteString("  \"\"\"List of module edges\"\"\"\n")
	sb.WriteString("  edges: [ModuleEdge!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Pagination information\"\"\"\n")
	sb.WriteString("  pageInfo: PageInfo!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Total count of modules\"\"\"\n")
	sb.WriteString("  totalCount: Int!\n")
	sb.WriteString("}\n\n")

	// ModuleEdge
	sb.WriteString("\"\"\"Edge type for module connections\"\"\"\n")
	sb.WriteString("type ModuleEdge {\n")
	sb.WriteString("  \"\"\"The module\"\"\"\n")
	sb.WriteString("  node: Module!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Cursor for this edge\"\"\"\n")
	sb.WriteString("  cursor: String!\n")
	sb.WriteString("}\n\n")

	// PageInfo
	sb.WriteString("\"\"\"Information about pagination\"\"\"\n")
	sb.WriteString("type PageInfo {\n")
	sb.WriteString("  \"\"\"Whether there are more results\"\"\"\n")
	sb.WriteString("  hasNextPage: Boolean!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Whether there are previous results\"\"\"\n")
	sb.WriteString("  hasPreviousPage: Boolean!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Cursor of the first edge\"\"\"\n")
	sb.WriteString("  startCursor: String\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Cursor of the last edge\"\"\"\n")
	sb.WriteString("  endCursor: String\n")
	sb.WriteString("}\n\n")

	// GraphStats
	sb.WriteString("\"\"\"Statistics about the knowledge graph\"\"\"\n")
	sb.WriteString("type GraphStats {\n")
	sb.WriteString("  \"\"\"Total number of modules\"\"\"\n")
	sb.WriteString("  totalModules: Int!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Total number of RDF triples\"\"\"\n")
	sb.WriteString("  totalTriples: Int!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Total number of relationships\"\"\"\n")
	sb.WriteString("  totalRelationships: Int!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Modules grouped by language\"\"\"\n")
	sb.WriteString("  modulesByLanguage: [LanguageStats!]!\n")
	sb.WriteString("  \n")
	sb.WriteString("  \"\"\"Modules grouped by layer\"\"\"\n")
	sb.WriteString("  modulesByLayer: [LayerStats!]!\n")
	sb.WriteString("}\n\n")

	// LanguageStats
	sb.WriteString("\"\"\"Module statistics by language\"\"\"\n")
	sb.WriteString("type LanguageStats {\n")
	sb.WriteString("  language: String!\n")
	sb.WriteString("  count: Int!\n")
	sb.WriteString("}\n\n")

	// LayerStats
	sb.WriteString("\"\"\"Module statistics by layer\"\"\"\n")
	sb.WriteString("type LayerStats {\n")
	sb.WriteString("  layer: String!\n")
	sb.WriteString("  count: Int!\n")
	sb.WriteString("}\n")

	return sb.String()
}

// GenerateSchemaFromGraph is a convenience function to generate schema
func GenerateSchemaFromGraph(g *graph.Graph, opts GenerateOptions) (string, error) {
	gen := NewGenerator(g)
	return gen.GenerateSchema(opts)
}

// GetAvailableTypes returns a list of all GraphQL types that will be generated
func (gen *Generator) GetAvailableTypes() []string {
	types := []string{
		"Module",
		"Export",
		"Query",
		"ModuleConnection",
		"ModuleEdge",
		"PageInfo",
		"GraphStats",
		"LanguageStats",
		"LayerStats",
	}
	sort.Strings(types)
	return types
}

// ValidateSchema performs basic validation on the generated schema
func (gen *Generator) ValidateSchema(schema string) error {
	// Check for required types
	requiredTypes := []string{"Query", "Module"}
	for _, reqType := range requiredTypes {
		if !strings.Contains(schema, "type "+reqType) {
			return fmt.Errorf("missing required type: %s", reqType)
		}
	}

	// Check schema definition
	if !strings.Contains(schema, "schema {") {
		return fmt.Errorf("missing schema definition")
	}

	return nil
}
