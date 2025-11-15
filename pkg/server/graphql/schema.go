/*
# Module: pkg/server/graphql/schema.go
GraphQL schema definition for GraphFS.

Defines GraphQL types and builds executable schema.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./resolvers](./resolvers.go) - GraphQL resolvers

## Tags
graphql, schema, server

## Exports
BuildSchema, ModuleType, QueryType

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#schema.go> a code:Module ;
    code:name "pkg/server/graphql/schema.go" ;
    code:description "GraphQL schema definition for GraphFS" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./resolvers.go> ;
    code:exports <#BuildSchema>, <#ModuleType>, <#QueryType> ;
    code:tags "graphql", "schema", "server" .
<!-- End LinkedDoc RDF -->
*/

package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/justin4957/graphfs/pkg/graph"
)

var (
	// ModuleType represents a code module in GraphQL
	ModuleType *graphql.Object

	// ExportType represents an exported symbol
	ExportType *graphql.Object

	// GraphStatsType represents graph statistics
	GraphStatsType *graphql.Object

	// LanguageStatsType represents language statistics
	LanguageStatsType *graphql.Object

	// LayerStatsType represents layer statistics
	LayerStatsType *graphql.Object

	// PageInfoType represents pagination information
	PageInfoType *graphql.Object

	// ModuleEdgeType represents a module edge
	ModuleEdgeType *graphql.Object

	// ModuleConnectionType represents a module connection
	ModuleConnectionType *graphql.Object
)

// BuildSchema builds the GraphQL schema
func BuildSchema(g *graph.Graph) (graphql.Schema, error) {
	// Initialize types
	initTypes()

	// Create resolver
	resolver := NewResolver(g)

	// Create Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Query",
		Description: "Root query type",
		Fields: graphql.Fields{
			"module": &graphql.Field{
				Type:        ModuleType,
				Description: "Get a single module by name, path, or URI",
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Module name",
					},
					"path": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Module path",
					},
					"uri": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Module URI",
					},
				},
				Resolve: resolver.Module,
			},
			"modules": &graphql.Field{
				Type:        ModuleConnectionType,
				Description: "List all modules with optional filtering",
				Args: graphql.FieldConfigArgument{
					"language": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Filter by programming language",
					},
					"layer": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Filter by architectural layer",
					},
					"tag": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Filter by tag",
					},
					"first": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "Maximum number of results",
					},
					"after": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Cursor for pagination",
					},
				},
				Resolve: resolver.Modules,
			},
			"searchModules": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(ModuleType)),
				Description: "Search modules by description",
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Search query",
					},
				},
				Resolve: resolver.SearchModules,
			},
			"stats": &graphql.Field{
				Type:        GraphStatsType,
				Description: "Get graph statistics",
				Resolve:     resolver.Stats,
			},
		},
	})

	// Create schema config
	schemaConfig := graphql.SchemaConfig{
		Query: queryType,
	}

	// Create and return schema
	return graphql.NewSchema(schemaConfig)
}

// initTypes initializes all GraphQL types
func initTypes() {
	// LanguageStatsType
	LanguageStatsType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "LanguageStats",
		Description: "Module statistics by language",
		Fields: graphql.Fields{
			"language": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Programming language",
			},
			"count": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Number of modules",
			},
		},
	})

	// LayerStatsType
	LayerStatsType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "LayerStats",
		Description: "Module statistics by layer",
		Fields: graphql.Fields{
			"layer": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Architectural layer",
			},
			"count": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Number of modules",
			},
		},
	})

	// GraphStatsType
	GraphStatsType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "GraphStats",
		Description: "Statistics about the knowledge graph",
		Fields: graphql.Fields{
			"totalModules": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Total number of modules",
			},
			"totalTriples": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Total number of RDF triples",
			},
			"totalRelationships": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Total number of relationships",
			},
			"modulesByLanguage": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(LanguageStatsType)),
				Description: "Modules grouped by language",
			},
			"modulesByLayer": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(LayerStatsType)),
				Description: "Modules grouped by layer",
			},
		},
	})

	// PageInfoType
	PageInfoType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "PageInfo",
		Description: "Information about pagination",
		Fields: graphql.Fields{
			"hasNextPage": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Whether there are more results",
			},
			"hasPreviousPage": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Whether there are previous results",
			},
			"startCursor": &graphql.Field{
				Type:        graphql.String,
				Description: "Cursor of the first edge",
			},
			"endCursor": &graphql.Field{
				Type:        graphql.String,
				Description: "Cursor of the last edge",
			},
		},
	})

	// ExportType
	ExportType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Export",
		Description: "Represents an exported symbol from a module",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Export name",
			},
		},
	})

	// ModuleType (forward reference, actual fields defined below)
	ModuleType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Module",
		Description: "Represents a code module in the knowledge graph",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.ID),
				Description: "Unique identifier",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						return module.URI, nil
					}
					return nil, nil
				},
			},
			"uri": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "URI identifier",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						return module.URI, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Module name",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						return module.Name, nil
					}
					return nil, nil
				},
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "Module description",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						if module.Description == "" {
							return nil, nil
						}
						return module.Description, nil
					}
					return nil, nil
				},
			},
			"path": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "File path relative to root",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						return module.Path, nil
					}
					return nil, nil
				},
			},
			"language": &graphql.Field{
				Type:        graphql.String,
				Description: "Programming language",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						if module.Language == "" {
							return nil, nil
						}
						return module.Language, nil
					}
					return nil, nil
				},
			},
			"layer": &graphql.Field{
				Type:        graphql.String,
				Description: "Architectural layer",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						if module.Layer == "" {
							return nil, nil
						}
						return module.Layer, nil
					}
					return nil, nil
				},
			},
			"tags": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
				Description: "Tags for categorization",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						return module.Tags, nil
					}
					return []string{}, nil
				},
			},
			"exports": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(ExportType)),
				Description: "Exported symbols/functions",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if module, ok := p.Source.(*graph.Module); ok {
						exports := make([]map[string]interface{}, len(module.Exports))
						for i, exp := range module.Exports {
							exports[i] = map[string]interface{}{
								"name": exp,
							}
						}
						return exports, nil
					}
					return []interface{}{}, nil
				},
			},
		},
	})

	// ModuleEdgeType
	ModuleEdgeType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "ModuleEdge",
		Description: "Edge type for module connections",
		Fields: graphql.Fields{
			"node": &graphql.Field{
				Type:        graphql.NewNonNull(ModuleType),
				Description: "The module",
			},
			"cursor": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Cursor for this edge",
			},
		},
	})

	// ModuleConnectionType
	ModuleConnectionType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "ModuleConnection",
		Description: "Connection type for module pagination",
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(ModuleEdgeType)),
				Description: "List of module edges",
			},
			"pageInfo": &graphql.Field{
				Type:        graphql.NewNonNull(PageInfoType),
				Description: "Pagination information",
			},
			"totalCount": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Total count of modules",
			},
		},
	})

	// Add dependency fields to ModuleType (circular reference)
	ModuleType.AddFieldConfig("dependencies", &graphql.Field{
		Type:        graphql.NewList(graphql.NewNonNull(ModuleType)),
		Description: "Modules this module depends on",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if module, ok := p.Source.(*graph.Module); ok {
				// Get graph from context
				if g, ok := p.Context.Value("graph").(*graph.Graph); ok {
					var deps []*graph.Module
					for _, depPath := range module.Dependencies {
						if depMod := g.GetModule(depPath); depMod != nil {
							deps = append(deps, depMod)
						}
					}
					return deps, nil
				}
			}
			return []*graph.Module{}, nil
		},
	})

	ModuleType.AddFieldConfig("dependents", &graphql.Field{
		Type:        graphql.NewList(graphql.NewNonNull(ModuleType)),
		Description: "Modules that depend on this module",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if module, ok := p.Source.(*graph.Module); ok {
				// Get graph from context
				if g, ok := p.Context.Value("graph").(*graph.Graph); ok {
					var deps []*graph.Module
					for _, depPath := range module.Dependents {
						if depMod := g.GetModule(depPath); depMod != nil {
							deps = append(deps, depMod)
						}
					}
					return deps, nil
				}
			}
			return []*graph.Module{}, nil
		},
	})
}
