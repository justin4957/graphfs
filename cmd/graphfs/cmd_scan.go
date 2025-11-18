/*
# Module: cmd/graphfs/cmd_scan.go
Scan command implementation.

Scans a codebase and builds the knowledge graph using the graph builder.

## Linked Modules
- [root](./root.go) - Root command
- [config](./config.go) - Configuration handling
- [../../pkg/graph](../../pkg/graph/builder.go) - Graph builder
- [../../pkg/scanner](../../pkg/scanner/scanner.go) - Scanner

## Tags
cli, command, scan

## Exports
scanCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_scan.go> a code:Module ;

	code:name "cmd/graphfs/cmd_scan.go" ;
	code:description "Scan command implementation" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <./config.go>, <../../pkg/graph/builder.go>,
	             <../../pkg/scanner/scanner.go> ;
	code:exports <#scanCmd> ;
	code:tags "cli", "command", "scan" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/justin4957/graphfs/pkg/cli"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	scanInclude  []string
	scanExclude  []string
	scanValidate bool
	scanStats    bool
	scanOutput   string
	scanNoCache  bool
	scanWorkers  int
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan codebase and build knowledge graph",
	Long: `Scan codebase and build knowledge graph from LinkedDoc metadata.

The scan command discovers all files with LinkedDoc metadata, parses them,
and builds a queryable knowledge graph stored in .graphfs/store.db.

Examples:
  graphfs scan                           # Scan current directory
  graphfs scan /path/to/project          # Scan specific directory
  graphfs scan --validate                # Scan with validation
  graphfs scan --stats                   # Show detailed statistics
  graphfs scan --output graph.json       # Export graph to JSON
  graphfs scan --workers 4               # Use 4 parallel workers
  graphfs scan --workers 1               # Sequential processing`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringSliceVar(&scanInclude, "include", nil, "Include files matching pattern")
	scanCmd.Flags().StringSliceVar(&scanExclude, "exclude", nil, "Exclude files matching pattern")
	scanCmd.Flags().BoolVar(&scanValidate, "validate", false, "Validate graph consistency")
	scanCmd.Flags().BoolVar(&scanStats, "stats", false, "Show detailed statistics")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Export graph to file")
	scanCmd.Flags().BoolVar(&scanNoCache, "no-cache", false, "Disable persistent caching")
	scanCmd.Flags().IntVarP(&scanWorkers, "workers", "w", 0, "Number of parallel workers (0 = NumCPU)")
}

func runScan(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Create output formatter
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absPath)
	}

	// Load configuration
	configPath := filepath.Join(absPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		out.Debug("Could not load config, using defaults: %v", err)
		config = DefaultConfig()
	}

	// Override with command-line flags
	scanOpts := scanner.ScanOptions{
		IncludePatterns: config.Scan.Include,
		ExcludePatterns: config.Scan.Exclude,
		MaxFileSize:     config.Scan.MaxFileSize,
		UseDefaults:     true,
		IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
		Concurrent:      true,
		Workers:         scanWorkers, // 0 = use NumCPU
	}

	if len(scanInclude) > 0 {
		scanOpts.IncludePatterns = scanInclude
	}
	if len(scanExclude) > 0 {
		scanOpts.ExcludePatterns = append(scanOpts.ExcludePatterns, scanExclude...)
	}

	out.Info("📊 Scanning codebase...")

	// Build graph
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions:    scanOpts,
		Validate:       scanValidate,
		ReportProgress: verbose,
		UseCache:       !scanNoCache, // Enable cache by default unless --no-cache is set
	}

	graphObj, err := builder.Build(absPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// Print summary
	out.Println("")
	out.Success("Knowledge graph built successfully")
	out.KeyValue("Modules", graphObj.Statistics.TotalModules)
	out.KeyValue("Triples", graphObj.Statistics.TotalTriples)
	out.KeyValue("Relationships", graphObj.Statistics.TotalRelationships)

	// Show validation results if requested
	if scanValidate {
		validator := graph.NewValidator()
		result := validator.Validate(graphObj)

		if len(result.Errors) > 0 {
			out.Println("")
			out.Error("Validation found %d errors", len(result.Errors))
			for _, err := range result.Errors {
				out.Error("  %s", err)
			}
		}

		if len(result.Warnings) > 0 {
			out.Println("")
			out.Warning("Validation found %d warnings", len(result.Warnings))
			if verbose {
				for _, warn := range result.Warnings {
					out.Warning("  %s", warn)
				}
			}
		}

		if len(result.Errors) == 0 && len(result.Warnings) == 0 {
			out.Success("Validation passed with no errors or warnings")
		}
	}

	// Show detailed statistics if requested
	if scanStats {
		out.Header("Detailed Statistics")

		if len(graphObj.Statistics.ModulesByLanguage) > 0 {
			out.Println("\nModules by Language:")
			headers := []string{"Language", "Count"}
			rows := [][]string{}
			for lang, count := range graphObj.Statistics.ModulesByLanguage {
				rows = append(rows, []string{lang, fmt.Sprintf("%d", count)})
			}
			out.Table(headers, rows)
		}

		if len(graphObj.Statistics.ModulesByLayer) > 0 {
			out.Println("\nModules by Layer:")
			headers := []string{"Layer", "Count"}
			rows := [][]string{}
			for layer, count := range graphObj.Statistics.ModulesByLayer {
				rows = append(rows, []string{layer, fmt.Sprintf("%d", count)})
			}
			out.Table(headers, rows)
		}
	}

	duration := time.Since(startTime)
	out.Println("")
	out.Success("Graph built in %v", duration.Round(time.Millisecond))

	// Export graph if requested
	if scanOutput != "" {
		if err := exportGraph(graphObj, scanOutput); err != nil {
			return fmt.Errorf("failed to export graph: %w", err)
		}
		out.Success("Graph exported to %s", scanOutput)
	}

	// Save graph to store (for future use)
	storeDir := filepath.Join(absPath, ".graphfs", "store")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		out.Debug("Could not create store directory: %v", err)
	}

	return nil
}

// exportGraph exports the graph to a file in JSON format
func exportGraph(g *graph.Graph, filename string) error {
	// Create export structure
	export := map[string]interface{}{
		"root":       g.Root,
		"modules":    g.Modules,
		"statistics": g.Statistics,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal graph: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
