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

	"github.com/justin4957/graphfs/pkg/query"
)

// Config holds server configuration
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	EnableCORS   bool
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnableCORS:   true,
	}
}

// Server is the HTTP server for GraphFS
type Server struct {
	config   *Config
	executor *query.Executor
	server   *http.Server
}

// NewServer creates a new HTTP server
func NewServer(config *Config, executor *query.Executor) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	return &Server{
		config:   config,
		executor: executor,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// SPARQL endpoint
	sparqlHandler := NewSPARQLHandler(s.executor, s.config.EnableCORS)
	mux.Handle("/sparql", sparqlHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

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
	response := `{
  "name": "GraphFS API",
  "version": "0.2.0",
  "endpoints": {
    "sparql": {
      "path": "/sparql",
      "methods": ["GET", "POST"],
      "description": "SPARQL query endpoint",
      "formats": ["json", "csv", "tsv", "xml"]
    },
    "health": {
      "path": "/health",
      "methods": ["GET"],
      "description": "Health check endpoint"
    }
  }
}`
	w.Write([]byte(response))
}
