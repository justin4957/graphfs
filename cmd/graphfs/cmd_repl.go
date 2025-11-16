/*
# Module: cmd/graphfs/cmd_repl.go
CLI command for interactive REPL.

Implements the 'graphfs repl' command for interactive query sessions.

## Linked Modules
- [../../pkg/repl](../../pkg/repl/repl.go) - REPL implementation
- [main](./main.go) - CLI entry point

## Tags
cli, repl, commands

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_repl.go> a code:Module ;
    code:name "cmd/graphfs/cmd_repl.go" ;
    code:description "CLI command for interactive REPL" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/repl/repl.go>, <./main.go> ;
    code:tags "cli", "repl", "commands" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/justin4957/graphfs/pkg/repl"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:     "repl",
	Aliases: []string{"interactive"},
	Short:   "Start interactive REPL for queries",
	Long: `Start an interactive Read-Eval-Print Loop for exploring the knowledge graph.

The REPL provides an interactive shell for executing SPARQL queries with:
- Multi-line query editing
- Command history (up/down arrows)
- Tab completion for keywords and commands
- Multiple output formats (table, JSON, CSV)
- Syntax highlighting and colored output

REPL Commands:
  .help               Show help and available commands
  .format [fmt]       Change output format (table, json, csv)
  .load <file>        Load and execute query from file
  .save <file>        Save last query to file
  .history            Show query history
  .clear              Clear screen
  .schema             Show available predicates and types
  .examples           Show example queries
  .stats              Show graph statistics
  .exit               Exit REPL (or Ctrl+D)

Examples:
  # Start REPL
  graphfs repl

  # Example query in REPL
  graphfs> SELECT ?name ?lang WHERE {
        ->   ?m a code:Module ;
        ->      code:name ?name ;
        ->      code:language ?lang
        -> } LIMIT 10
        -> [press Enter on empty line to execute]

  # Change output format
  graphfs> .format json

  # Show graph statistics
  graphfs> .stats

  # Load query from file
  graphfs> .load my-query.sparql
`,
	RunE: runREPL,
}

func init() {
	rootCmd.AddCommand(replCmd)
}

func runREPL(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Built graph with %d modules, %d triples\n\n", len(g.Modules), g.Store.Count())

	// Create query executor
	executor := query.NewExecutor(g.Store)

	// Create REPL config
	replConfig := &repl.Config{
		HistoryFile: filepath.Join(os.TempDir(), ".graphfs_history"),
		Prompt:      "graphfs> ",
		NoColor:     noColor,
	}

	// Create and start REPL
	r, err := repl.New(executor, g, replConfig)
	if err != nil {
		return fmt.Errorf("failed to create REPL: %w", err)
	}

	return r.Run()
}
