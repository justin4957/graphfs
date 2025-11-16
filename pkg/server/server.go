/*
# Module: pkg/server/server.go
HTTP server for GraphFS query endpoints.

Provides HTTP server for SPARQL, GraphQL, and REST API endpoints.

## Linked Modules
- [sparql_handler](./sparql_handler.go) - SPARQL HTTP handler
- [../query](../query/executor.go) - Query executor
- [../graph](../graph/graph.go) - Graph builder

## Tags
server, http, api

## Exports
Server, Config, NewServer

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#server.go> a code:Module ;
    code:name "pkg/server/server.go" ;
    code:description "HTTP server for GraphFS query endpoints" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <./sparql_handler.go>, <../query/executor.go> ;
    code:exports <#Server>, <#Config>, <#NewServer> ;
    code:tags "server", "http", "api" .
<!-- End LinkedDoc RDF -->
*/

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/justin4957/graphfs/pkg/cache"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	graphqlserver "github.com/justin4957/graphfs/pkg/server/graphql"
	restserver "github.com/justin4957/graphfs/pkg/server/rest"
)

// Config holds server configuration
type Config struct {
	Host             string
	Port             int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	EnableCORS       bool
	EnableGraphQL    bool
	EnablePlayground bool
	EnableREST       bool
	EnableCache      bool
	CacheMaxEntries  int
	CacheTTL         time.Duration
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:             "localhost",
		Port:             8080,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		EnableCORS:       true,
		EnableGraphQL:    true,
		EnablePlayground: true,
		EnableREST:       true,
		EnableCache:      true,
		CacheMaxEntries:  1000,
		CacheTTL:         5 * time.Minute,
	}
}

// Server is the HTTP server for GraphFS
type Server struct {
	config   *Config
	executor *query.Executor
	graph    *graph.Graph
	server   *http.Server
	cache    *cache.Cache
}

// NewServer creates a new HTTP server
func NewServer(config *Config, executor *query.Executor) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	s := &Server{
		config:   config,
		executor: executor,
	}

	// Initialize cache if enabled
	if config.EnableCache {
		s.cache = cache.NewCache(config.CacheMaxEntries, config.CacheTTL)
	}

	return s
}

// NewServerWithGraph creates a new HTTP server with graph for GraphQL support
func NewServerWithGraph(config *Config, executor *query.Executor, g *graph.Graph) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	s := &Server{
		config:   config,
		executor: executor,
		graph:    g,
	}

	// Initialize cache if enabled
	if config.EnableCache {
		s.cache = cache.NewCache(config.CacheMaxEntries, config.CacheTTL)
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// SPARQL endpoint
	sparqlHandler := NewSPARQLHandler(s.executor, s.config.EnableCORS)
	if s.config.EnableCache && s.cache != nil {
		mux.Handle("/sparql", CacheMiddleware(sparqlHandler, s.cache))
	} else {
		mux.Handle("/sparql", sparqlHandler)
	}

	// GraphQL endpoint (if enabled and graph is available)
	if s.config.EnableGraphQL && s.graph != nil {
		graphqlHandler, err := graphqlserver.NewHandler(s.graph, graphqlserver.HandlerConfig{
			EnablePlayground: s.config.EnablePlayground,
			EnableCORS:       s.config.EnableCORS,
		})
		if err != nil {
			return fmt.Errorf("failed to create GraphQL handler: %w", err)
		}

		if s.config.EnableCache && s.cache != nil {
			mux.Handle("/graphql", CacheMiddleware(graphqlHandler, s.cache))
		} else {
			mux.Handle("/graphql", graphqlHandler)
		}
	}

	// REST API endpoints (if enabled and graph is available)
	if s.config.EnableREST && s.graph != nil {
		restHandler := restserver.NewHandler(s.graph, s.config.EnableCORS)
		if s.config.EnableCache && s.cache != nil {
			restHandler.RegisterRoutesWithCache(mux, s.cache)
		} else {
			restHandler.RegisterRoutes(mux)
		}
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Cache stats endpoint (if cache is enabled)
	if s.config.EnableCache && s.cache != nil {
		mux.HandleFunc("/cache/stats", s.handleCacheStats)
	}

	// Root endpoint with API info
	mux.HandleFunc("/", s.handleRoot)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	log.Printf("Starting GraphFS server on http://%s", addr)
	log.Printf("SPARQL endpoint: http://%s/sparql", addr)
	if s.config.EnableGraphQL && s.graph != nil {
		log.Printf("GraphQL endpoint: http://%s/graphql", addr)
		if s.config.EnablePlayground {
			log.Printf("GraphQL Playground: http://%s/graphql", addr)
		}
	}
	if s.config.EnableREST && s.graph != nil {
		log.Printf("REST API: http://%s/api/v1", addr)
	}
	if s.config.EnableCache && s.cache != nil {
		log.Printf("Cache enabled: %d max entries, %v TTL", s.config.CacheMaxEntries, s.config.CacheTTL)
		log.Printf("Cache stats: http://%s/cache/stats", addr)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// handleRoot provides API information
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Build endpoints info
	endpoints := `{
  "name": "GraphFS API",
  "version": "0.2.0",
  "endpoints": {
    "sparql": {
      "path": "/sparql",
      "methods": ["GET", "POST"],
      "description": "SPARQL query endpoint",
      "formats": ["json", "csv", "tsv", "xml"]
    }`

	if s.config.EnableGraphQL && s.graph != nil {
		endpoints += `,
    "graphql": {
      "path": "/graphql",
      "methods": ["GET", "POST"],
      "description": "GraphQL query endpoint",
      "playground": ` + fmt.Sprintf("%v", s.config.EnablePlayground) + `
    }`
	}

	if s.config.EnableREST && s.graph != nil {
		endpoints += `,
    "rest": {
      "path": "/api/v1",
      "methods": ["GET"],
      "description": "RESTful API for common queries",
      "endpoints": {
        "modules": "/api/v1/modules",
        "search": "/api/v1/modules/search?q=query",
        "stats": "/api/v1/analysis/stats",
        "tags": "/api/v1/tags",
        "exports": "/api/v1/exports"
      }
    }`
	}

	endpoints += `,
    "health": {
      "path": "/health",
      "methods": ["GET"],
      "description": "Health check endpoint"
    }
  }
}`

	w.Write([]byte(endpoints))
}

// handleCacheStats provides cache statistics
func (s *Server) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.cache == nil {
		http.Error(w, "Cache not enabled", http.StatusNotFound)
		return
	}

	stats := s.cache.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{
  "hits": %d,
  "misses": %d,
  "evictions": %d,
  "size": %d,
  "maxSize": %d,
  "totalBytes": %d,
  "hitRate": %.4f
}`, stats.Hits, stats.Misses, stats.Evictions, stats.Size, stats.MaxSize, stats.TotalBytes, stats.HitRate)

	w.Write([]byte(response))
}
