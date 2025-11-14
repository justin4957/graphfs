/*
# Module: cmd/graphfs/cmd_query.go
Query command implementation.

Executes SPARQL queries against the knowledge graph.

## Linked Modules
- [root](./root.go) - Root command
- [output](./output.go) - Output formatting
- [../../pkg/query](../../pkg/query/engine.go) - Query engine

## Tags
cli, command, query, sparql

## Exports
queryCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_query.go> a code:Module ;

	code:name "cmd/graphfs/cmd_query.go" ;
	code:description "Query command implementation" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <./output.go>, <../../pkg/query/engine.go> ;
	code:exports <#queryCmd> ;
	code:tags "cli", "command", "query", "sparql" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/spf13/cobra"
)

var (
	queryFile   string
	queryFormat string
	queryLimit  int
	queryOutput string
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query <query>",
	Short: "Execute SPARQL query against knowledge graph",
	Long: `Execute SPARQL query against the knowledge graph.

The query command allows you to query the knowledge graph using SPARQL.
Results can be formatted as table, JSON, or CSV.

Examples:
  # Inline query
  graphfs query 'SELECT * WHERE { ?s ?p ?o } LIMIT 10'

  # Query from file
  graphfs query --file queries/modules.sparql

  # Format as JSON
  graphfs query 'SELECT ?module WHERE { ?module a code:Module }' --format json

  # Save to file
  graphfs query --file queries/deps.sparql --output results.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runQuery,
}

func init() {
	queryCmd.Flags().StringVarP(&queryFile, "file", "f", "", "Read query from file")
	queryCmd.Flags().StringVar(&queryFormat, "format", "table", "Output format: table, json, csv")
	queryCmd.Flags().IntVarP(&queryLimit, "limit", "l", 100, "Limit number of results")
	queryCmd.Flags().StringVarP(&queryOutput, "output", "o", "", "Write results to file")
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Get query string
	var queryString string
	if queryFile != "" {
		// Read query from file
		data, err := os.ReadFile(queryFile)
		if err != nil {
			return fmt.Errorf("failed to read query file: %w", err)
		}
		queryString = string(data)
	} else if len(args) > 0 {
		// Use inline query
		queryString = args[0]
	} else {
		return fmt.Errorf("query string or --file required")
	}

	// Determine current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if GraphFS is initialized
	graphfsDir := filepath.Join(currentDir, ".graphfs")
	if _, err := os.Stat(graphfsDir); os.IsNotExist(err) {
		return fmt.Errorf("GraphFS not initialized. Run 'graphfs init' first")
	}

	// Build graph (in a real implementation, we would load from store)
	if verbose {
		fmt.Println("Building knowledge graph...")
	}

	builder := graph.NewBuilder()
	graphObj, err := builder.Build(currentDir, graph.BuildOptions{
		ReportProgress: verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	if verbose {
		fmt.Printf("Graph loaded: %d modules, %d triples\n",
			graphObj.Statistics.TotalModules,
			graphObj.Statistics.TotalTriples)
	}

	// Execute query
	engine := query.NewEngine(graphObj.Store)
	result, err := engine.Execute(queryString)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Format and output results
	var output string
	switch queryFormat {
	case "json":
		output, err = formatJSON(result)
	case "csv":
		output, err = formatCSV(result)
	case "table":
		output, err = formatTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", queryFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	// Write to file or stdout
	if queryOutput != "" {
		if err := os.WriteFile(queryOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("Results written to %s\n", queryOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}

// formatJSON formats query results as JSON
func formatJSON(result *query.QueryResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatCSV formats query results as CSV
func formatCSV(result *query.QueryResult) (string, error) {
	if len(result.Bindings) == 0 {
		return "", nil
	}

	// Create in-memory CSV writer
	var output [][]string

	// Add header
	output = append(output, result.Variables)

	// Add rows
	for _, binding := range result.Bindings {
		row := make([]string, len(result.Variables))
		for i, variable := range result.Variables {
			if value, ok := binding[variable]; ok {
				row[i] = value
			}
		}
		output = append(output, row)
	}

	// Convert to CSV string
	var csvData string
	writer := csv.NewWriter(&stringWriter{&csvData})
	writer.WriteAll(output)
	writer.Flush()

	return csvData, writer.Error()
}

// stringWriter implements io.Writer for strings
type stringWriter struct {
	s *string
}

func (sw *stringWriter) Write(p []byte) (n int, err error) {
	*sw.s += string(p)
	return len(p), nil
}
