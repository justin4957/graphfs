/*
# Module: pkg/server/graphql/resolvers.go
GraphQL resolvers for GraphFS.

Implements resolver functions for GraphQL queries.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./schema](./schema.go) - GraphQL schema

## Tags
graphql, resolvers, server

## Exports
Resolver, NewResolver

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#resolvers.go> a code:Module ;
    code:name "pkg/server/graphql/resolvers.go" ;
    code:description "GraphQL resolvers for GraphFS" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./schema.go> ;
    code:exports <#Resolver>, <#NewResolver> ;
    code:tags "graphql", "resolvers", "server" .
<!-- End LinkedDoc RDF -->
*/

package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/justin4957/graphfs/pkg/graph"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// graphContextKey is the context key for storing the graph
	graphContextKey contextKey = "graph"
)

// Resolver handles GraphQL query resolution
type Resolver struct {
	graph *graph.Graph
}

// NewResolver creates a new resolver
func NewResolver(g *graph.Graph) *Resolver {
	return &Resolver{graph: g}
}

// Module resolves the module query
func (r *Resolver) Module(p graphql.ResolveParams) (interface{}, error) {
	// Add graph to context for nested resolvers
	ctx := context.WithValue(p.Context, graphContextKey, r.graph)
	p.Context = ctx

	// Get arguments
	name, hasName := p.Args["name"].(string)
	path, hasPath := p.Args["path"].(string)
	uri, hasURI := p.Args["uri"].(string)

	// Find module by name, path, or URI
	if hasName {
		for _, mod := range r.graph.Modules {
			if mod.Name == name {
				return mod, nil
			}
		}
	}

	if hasPath {
		if mod := r.graph.GetModule(path); mod != nil {
			return mod, nil
		}
	}

	if hasURI {
		for _, mod := range r.graph.Modules {
			if mod.URI == uri {
				return mod, nil
			}
		}
	}

	return nil, nil
}

// Modules resolves the modules query with filtering and pagination
func (r *Resolver) Modules(p graphql.ResolveParams) (interface{}, error) {
	// Add graph to context for nested resolvers
	ctx := context.WithValue(p.Context, graphContextKey, r.graph)
	p.Context = ctx

	// Get filter arguments
	language, _ := p.Args["language"].(string)
	layer, _ := p.Args["layer"].(string)
	tag, _ := p.Args["tag"].(string)

	// Get pagination arguments
	first, hasFirst := p.Args["first"].(int)
	after, _ := p.Args["after"].(string)

	// Filter modules
	var filtered []*graph.Module
	for _, mod := range r.graph.Modules {
		// Apply filters
		if language != "" && mod.Language != language {
			continue
		}
		if layer != "" && mod.Layer != layer {
			continue
		}
		if tag != "" {
			hasTag := false
			for _, t := range mod.Tags {
				if t == tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		filtered = append(filtered, mod)
	}

	// Handle pagination
	startIdx := 0
	if after != "" {
		// Decode cursor
		idx, err := decodeCursor(after)
		if err == nil {
			startIdx = idx + 1
		}
	}

	endIdx := len(filtered)
	if hasFirst && startIdx+first < endIdx {
		endIdx = startIdx + first
	}

	// Build edges
	var edges []map[string]interface{}
	for i := startIdx; i < endIdx; i++ {
		edges = append(edges, map[string]interface{}{
			"node":   filtered[i],
			"cursor": encodeCursor(i),
		})
	}

	// Build page info
	pageInfo := map[string]interface{}{
		"hasNextPage":     endIdx < len(filtered),
		"hasPreviousPage": startIdx > 0,
		"startCursor":     nil,
		"endCursor":       nil,
	}

	if len(edges) > 0 {
		pageInfo["startCursor"] = encodeCursor(startIdx)
		pageInfo["endCursor"] = encodeCursor(endIdx - 1)
	}

	// Build connection
	return map[string]interface{}{
		"edges":      edges,
		"pageInfo":   pageInfo,
		"totalCount": len(filtered),
	}, nil
}

// SearchModules resolves the searchModules query
func (r *Resolver) SearchModules(p graphql.ResolveParams) (interface{}, error) {
	// Add graph to context for nested resolvers
	ctx := context.WithValue(p.Context, graphContextKey, r.graph)
	p.Context = ctx

	query, ok := p.Args["query"].(string)
	if !ok || query == "" {
		return []*graph.Module{}, nil
	}

	queryLower := strings.ToLower(query)
	var results []*graph.Module

	for _, mod := range r.graph.Modules {
		// Search in description
		if strings.Contains(strings.ToLower(mod.Description), queryLower) {
			results = append(results, mod)
			continue
		}

		// Search in name
		if strings.Contains(strings.ToLower(mod.Name), queryLower) {
			results = append(results, mod)
			continue
		}

		// Search in tags
		for _, tag := range mod.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, mod)
				break
			}
		}
	}

	return results, nil
}

// Stats resolves the stats query
func (r *Resolver) Stats(p graphql.ResolveParams) (interface{}, error) {
	// Count modules by language
	languageCounts := make(map[string]int)
	for _, mod := range r.graph.Modules {
		if mod.Language != "" {
			languageCounts[mod.Language]++
		}
	}

	var modulesByLanguage []map[string]interface{}
	for lang, count := range languageCounts {
		modulesByLanguage = append(modulesByLanguage, map[string]interface{}{
			"language": lang,
			"count":    count,
		})
	}

	// Count modules by layer
	layerCounts := make(map[string]int)
	for _, mod := range r.graph.Modules {
		if mod.Layer != "" {
			layerCounts[mod.Layer]++
		}
	}

	var modulesByLayer []map[string]interface{}
	for layer, count := range layerCounts {
		modulesByLayer = append(modulesByLayer, map[string]interface{}{
			"layer": layer,
			"count": count,
		})
	}

	// Count total relationships
	totalRelationships := 0
	for _, mod := range r.graph.Modules {
		totalRelationships += len(mod.Dependencies)
	}

	return map[string]interface{}{
		"totalModules":       len(r.graph.Modules),
		"totalTriples":       r.graph.Store.Count(),
		"totalRelationships": totalRelationships,
		"modulesByLanguage":  modulesByLanguage,
		"modulesByLayer":     modulesByLayer,
	}, nil
}

// encodeCursor encodes an index as a base64 cursor
func encodeCursor(idx int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor:%d", idx)))
}

// decodeCursor decodes a base64 cursor to an index
func decodeCursor(cursor string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 || parts[0] != "cursor" {
		return 0, fmt.Errorf("invalid cursor format")
	}

	return strconv.Atoi(parts[1])
}
