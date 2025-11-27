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
	"strings"

	"github.com/justin4957/graphfs/pkg/cli"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	queryFile     string
	queryFormat   string
	queryLimit    int
	queryOutput   string
	queryOffset   int
	queryPageSize int
	queryStream   bool
	queryPage     int
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query <query>",
	Short: "Execute SPARQL query against knowledge graph",
	Long: `Execute SPARQL query against the knowledge graph.

The query command allows you to query the knowledge graph using SPARQL.
Results can be formatted as table, JSON, or CSV.

Supports streaming and pagination for large result sets:
  --stream     Stream results incrementally (memory efficient)
  --page N     Show specific page of results
  --page-size  Number of results per page (default: 100)
  --offset     Skip first N results
  --limit      Limit total results

Examples:
  # Inline query
  graphfs query 'SELECT * WHERE { ?s ?p ?o } LIMIT 10'

  # Query from file
  graphfs query --file queries/modules.sparql

  # Format as JSON
  graphfs query 'SELECT ?module WHERE { ?module a code:Module }' --format json

  # Save to file
  graphfs query --file queries/deps.sparql --output results.json

  # Stream large results
  graphfs query --stream 'SELECT * WHERE { ?s ?p ?o }'

  # Paginate results
  graphfs query --page 1 --page-size 50 'SELECT * WHERE { ?s ?p ?o }'

  # Skip first 100 results
  graphfs query --offset 100 --limit 50 'SELECT * WHERE { ?s ?p ?o }'`,
	Args: cobra.MaximumNArgs(1),
	RunE: runQuery,
}

func init() {
	queryCmd.Flags().StringVarP(&queryFile, "file", "f", "", "Read query from file")
	queryCmd.Flags().StringVar(&queryFormat, "format", "table", "Output format: table, json, csv")
	queryCmd.Flags().IntVarP(&queryLimit, "limit", "l", 0, "Limit number of results (0 = no limit)")
	queryCmd.Flags().StringVarP(&queryOutput, "output", "o", "", "Write results to file")
	queryCmd.Flags().IntVar(&queryOffset, "offset", 0, "Skip first N results")
	queryCmd.Flags().IntVar(&queryPageSize, "page-size", 100, "Number of results per page")
	queryCmd.Flags().BoolVar(&queryStream, "stream", false, "Stream results incrementally")
	queryCmd.Flags().IntVar(&queryPage, "page", 0, "Show specific page of results (1-indexed)")
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Create output formatter
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

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
	out.Debug("Building knowledge graph...")

	builder := graph.NewBuilder()
	graphObj, err := builder.Build(currentDir, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
			IgnoreFiles: []string{".gitignore", ".graphfsignore"},
			Concurrent:  true,
		},
		ReportProgress: verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	out.Debug("Graph loaded: %d modules, %d triples",
		graphObj.Statistics.TotalModules,
		graphObj.Statistics.TotalTriples)

	// Route to appropriate execution mode
	if queryStream {
		return runStreamingQuery(graphObj, queryString, out)
	} else if queryPage > 0 {
		return runPaginatedQuery(graphObj, queryString, out)
	} else {
		return runNormalQuery(graphObj, queryString, out)
	}
}

// runNormalQuery executes a query normally (all results at once)
func runNormalQuery(graphObj *graph.Graph, queryString string, out *cli.OutputFormatter) error {
	// Parse the query to apply CLI flags
	parsedQuery, err := query.ParseQuery(queryString)
	if err != nil {
		return fmt.Errorf("query parse failed: %w", err)
	}

	// Apply CLI limit/offset if specified and not already in query
	if parsedQuery.Select != nil {
		if queryLimit > 0 && parsedQuery.Select.Limit == 0 {
			parsedQuery.Select.Limit = queryLimit
		}
		if queryOffset > 0 && parsedQuery.Select.Offset == 0 {
			parsedQuery.Select.Offset = queryOffset
		}
	}

	// Execute query
	executor := query.NewExecutor(graphObj.Store)
	result, err := executor.Execute(parsedQuery)
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
		out.Success("Results written to %s", queryOutput)
	} else {
		fmt.Println(output)
	}

	return nil
}

// runStreamingQuery executes a query with streaming output
func runStreamingQuery(graphObj *graph.Graph, queryString string, out *cli.OutputFormatter) error {
	config := &query.StreamConfig{
		PageSize:       queryPageSize,
		BufferSize:     queryPageSize,
		ReportProgress: verbose,
	}

	streamExecutor := query.NewStreamingExecutor(graphObj.Store, config)

	// Progress callback for verbose mode
	var progressCallback query.ProgressCallback
	if verbose {
		progressCallback = func(current, total int) {
			if current%1000 == 0 || current == total {
				fmt.Fprintf(os.Stderr, "\rStreaming: %d/%d results...", current, total)
			}
		}
	}

	parsedQuery, err := query.ParseQuery(queryString)
	if err != nil {
		return fmt.Errorf("query parse failed: %w", err)
	}

	stream := streamExecutor.ExecuteStreamWithProgress(parsedQuery, progressCallback)

	fmt.Println("Streaming results...")
	if len(stream.Variables) > 0 {
		fmt.Println("Variables:", stream.Variables)
	}
	fmt.Println()

	count := 0
	err = stream.ForEach(func(binding map[string]string) error {
		count++
		switch queryFormat {
		case "json":
			data, jsonErr := json.Marshal(binding)
			if jsonErr != nil {
				return jsonErr
			}
			fmt.Println(string(data))
		case "csv":
			var values []string
			for _, v := range stream.Variables {
				values = append(values, binding[v])
			}
			fmt.Println(strings.Join(values, ","))
		default: // table format
			fmt.Printf("%d. %v\n", count, binding)
		}
		return nil
	})

	if verbose {
		fmt.Fprintln(os.Stderr) // Clear progress line
	}

	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Printf("\nTotal: %d results streamed\n", count)
	return nil
}

// runPaginatedQuery executes a query with pagination
func runPaginatedQuery(graphObj *graph.Graph, queryString string, out *cli.OutputFormatter) error {
	config := &query.StreamConfig{
		PageSize: queryPageSize,
	}

	streamExecutor := query.NewStreamingExecutor(graphObj.Store, config)

	parsedQuery, err := query.ParseQuery(queryString)
	if err != nil {
		return fmt.Errorf("query parse failed: %w", err)
	}

	paginatedResult, err := streamExecutor.ExecutePaginated(parsedQuery, queryPage, queryPageSize)
	if err != nil {
		return fmt.Errorf("paginated query failed: %w", err)
	}

	// Convert to QueryResult for formatting
	result := &query.QueryResult{
		Variables: paginatedResult.Variables,
		Bindings:  paginatedResult.Bindings,
		Count:     len(paginatedResult.Bindings),
	}

	// Format and output results
	var output string
	switch queryFormat {
	case "json":
		// Include pagination metadata in JSON output
		paginationData := map[string]interface{}{
			"page":       paginatedResult.Page,
			"pageSize":   paginatedResult.PageSize,
			"totalCount": paginatedResult.TotalCount,
			"totalPages": paginatedResult.TotalPages,
			"hasMore":    paginatedResult.HasMore,
			"variables":  paginatedResult.Variables,
			"bindings":   paginatedResult.Bindings,
		}
		data, jsonErr := json.MarshalIndent(paginationData, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to marshal JSON: %w", jsonErr)
		}
		output = string(data)
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
		out.Success("Results written to %s", queryOutput)
	} else {
		fmt.Println(output)
		// Print pagination info for table/csv format
		if queryFormat != "json" {
			fmt.Printf("\nPage %d of %d (showing %d of %d total results)\n",
				paginatedResult.Page,
				paginatedResult.TotalPages,
				len(paginatedResult.Bindings),
				paginatedResult.TotalCount)
			if paginatedResult.HasMore {
				fmt.Printf("Use --page %d to see more results\n", paginatedResult.Page+1)
			}
		}
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
