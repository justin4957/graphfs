/*
# Module: pkg/server/graphql/server.go
GraphQL HTTP server for GraphFS.

Provides HTTP handler for GraphQL queries with GraphiQL playground.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./schema](./schema.go) - GraphQL schema
- [./resolvers](./resolvers.go) - GraphQL resolvers

## Tags
graphql, server, http

## Exports
NewHandler

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#graphql-server.go> a code:Module ;
    code:name "pkg/server/graphql/server.go" ;
    code:description "GraphQL HTTP server for GraphFS" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./schema.go>, <./resolvers.go> ;
    code:exports <#NewHandler> ;
    code:tags "graphql", "server", "http" .
<!-- End LinkedDoc RDF -->
*/

package graphql

import (
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/justin4957/graphfs/pkg/graph"
)

// HandlerConfig configures the GraphQL handler
type HandlerConfig struct {
	EnablePlayground bool
	EnableCORS       bool
}

// NewHandler creates a new GraphQL HTTP handler
func NewHandler(g *graph.Graph, config HandlerConfig) (http.Handler, error) {
	// Build schema
	schema, err := BuildSchema(g)
	if err != nil {
		return nil, err
	}

	// Create handler
	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   config.EnablePlayground,
		Playground: config.EnablePlayground,
	})

	// Wrap with CORS if enabled
	if config.EnableCORS {
		return corsHandler(h), nil
	}

	return h, nil
}

// corsHandler wraps an HTTP handler with CORS headers
func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
