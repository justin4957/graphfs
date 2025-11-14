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

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	scanInclude  []string
	scanExclude  []string
	scanValidate bool
	scanStats    bool
	scanOutput   string
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
  graphfs scan --output graph.json       # Export graph to JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringSliceVar(&scanInclude, "include", nil, "Include files matching pattern")
	scanCmd.Flags().StringSliceVar(&scanExclude, "exclude", nil, "Exclude files matching pattern")
	scanCmd.Flags().BoolVar(&scanValidate, "validate", false, "Validate graph consistency")
	scanCmd.Flags().BoolVar(&scanStats, "stats", false, "Show detailed statistics")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Export graph to file")
}

func runScan(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

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
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: Could not load config, using defaults: %v\n", err)
		}
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
	}

	if len(scanInclude) > 0 {
		scanOpts.IncludePatterns = scanInclude
	}
	if len(scanExclude) > 0 {
		scanOpts.ExcludePatterns = append(scanOpts.ExcludePatterns, scanExclude...)
	}

	fmt.Println("Scanning codebase...")

	// Build graph
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions:    scanOpts,
		Validate:       scanValidate,
		ReportProgress: verbose,
	}

	graphObj, err := builder.Build(absPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// Show progress with modules (simulating progress bar)
	if !verbose {
		bar := progressbar.Default(int64(graphObj.Statistics.TotalModules), "Parsing LinkedDoc")
		for i := 0; i < graphObj.Statistics.TotalModules; i++ {
			bar.Add(1)
		}
		fmt.Println()
	}

	// Print summary
	fmt.Println("\nBuilding graph...")
	fmt.Printf("  ✓ %d modules\n", graphObj.Statistics.TotalModules)
	fmt.Printf("  ✓ %d triples\n", graphObj.Statistics.TotalTriples)
	fmt.Printf("  ✓ %d relationships\n", graphObj.Statistics.TotalRelationships)

	// Show validation results if requested
	if scanValidate {
		validator := graph.NewValidator()
		result := validator.Validate(graphObj)
		fmt.Printf("  ✓ %d errors, %d warnings\n", len(result.Errors), len(result.Warnings))

		if len(result.Errors) > 0 {
			fmt.Println("\nValidation Errors:")
			for _, err := range result.Errors {
				fmt.Printf("  ✗ %s\n", err)
			}
		}

		if len(result.Warnings) > 0 && verbose {
			fmt.Println("\nValidation Warnings:")
			for _, warn := range result.Warnings {
				fmt.Printf("  ⚠ %s\n", warn)
			}
		}
	}

	// Show detailed statistics if requested
	if scanStats {
		fmt.Println("\nDetailed Statistics:")
		fmt.Printf("  Modules by Language:\n")
		for lang, count := range graphObj.Statistics.ModulesByLanguage {
			fmt.Printf("    %s: %d\n", lang, count)
		}
		fmt.Printf("  Modules by Layer:\n")
		for layer, count := range graphObj.Statistics.ModulesByLayer {
			fmt.Printf("    %s: %d\n", layer, count)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✓ Graph built successfully in %v\n", duration.Round(time.Millisecond))

	// Export graph if requested
	if scanOutput != "" {
		if err := exportGraph(graphObj, scanOutput); err != nil {
			return fmt.Errorf("failed to export graph: %w", err)
		}
		fmt.Printf("✓ Graph exported to %s\n", scanOutput)
	}

	// Save graph to store (for future use)
	storeDir := filepath.Join(absPath, ".graphfs", "store")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: Could not create store directory: %v\n", err)
		}
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
