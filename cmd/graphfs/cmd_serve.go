/*
# Module: cmd/graphfs/cmd_serve.go
CLI command to start GraphFS HTTP server.

Starts HTTP server for SPARQL and future GraphQL/REST endpoints.

## Linked Modules
- [../../pkg/server](../../pkg/server/server.go) - HTTP server
- [../../pkg/scanner](../../pkg/scanner/scanner.go) - Filesystem scanner
- [../../pkg/graph](../../pkg/graph/graph.go) - Graph builder

## Tags
cli, server, command

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_serve.go> a code:Module ;
    code:name "cmd/graphfs/cmd_serve.go" ;
    code:description "CLI command to start GraphFS HTTP server" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/server/server.go>, <../../pkg/scanner/scanner.go>, <../../pkg/graph/graph.go> ;
    code:tags "cli", "server", "command" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/justin4957/graphfs/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start GraphFS HTTP server",
	Long: `Start HTTP server for SPARQL queries and GraphQL/REST APIs.

The server scans the codebase, builds the knowledge graph, and exposes
query endpoints via HTTP. The graph is loaded at startup and kept in memory.

Examples:
  # Start server on default port 8080
  graphfs serve

  # Start server on custom port
  graphfs serve --port 9000

  # Start server on all interfaces
  graphfs serve --host 0.0.0.0 --port 8080

  # Query the server
  curl http://localhost:8080/sparql?query=SELECT+*+WHERE+{+?s+?p+?o+}+LIMIT+10
`,
	RunE: runServe,
}

var (
	serveHost string
	servePort int
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host to bind server to")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get root path
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration
	configPath := filepath.Join(rootPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Scanning codebase and building graph...")

	// Create scanner with options
	scanOpts := scanner.ScanOptions{
		IncludePatterns: config.Scan.Include,
		ExcludePatterns: config.Scan.Exclude,
		MaxFileSize:     config.Scan.MaxFileSize,
		UseDefaults:     true,
		IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
		Concurrent:      true,
	}

	// Build graph using builder
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanOpts,
		Validate:    true,
	}

	g, err := builder.Build(rootPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Printf("Built graph with %d modules, %d triples\n", len(g.Modules), g.Store.Count())

	// Create query executor
	executor := query.NewExecutor(g.Store)

	// Create server configuration
	serverConfig := &server.Config{
		Host:             serveHost,
		Port:             servePort,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		EnableCORS:       viper.GetBool("server.cors"),
		EnableGraphQL:    true,
		EnablePlayground: true,
		EnableREST:       true,
	}

	// Create and start server with GraphQL support
	srv := server.NewServerWithGraph(serverConfig, executor, g)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Stop(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
