/*
# Module: cmd/graphfs/cmd_watch.go
Watch command for live file system monitoring.

Implements the 'graphfs watch' command for monitoring file changes and
automatically re-running queries or regenerating visualizations.

## Linked Modules
- [../../pkg/watch](../../pkg/watch/watcher.go) - File system watcher
- [../../pkg/graph](../../pkg/graph/graph.go) - Graph building
- [../../pkg/query](../../pkg/query/engine.go) - Query engine
- [root](./root.go) - Root command

## Tags
cli, watch, monitoring

## Exports
watchCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_watch.go> a code:Module ;
    code:name "cmd/graphfs/cmd_watch.go" ;
    code:description "Watch command for live file system monitoring" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/watch/watcher.go>, <../../pkg/graph/graph.go>,
                 <../../pkg/query/engine.go>, <./root.go> ;
    code:exports <#watchCmd> ;
    code:tags "cli", "watch", "monitoring" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/justin4957/graphfs/pkg/viz"
	"github.com/justin4957/graphfs/pkg/watch"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [query]",
	Short: "Watch for file changes and update results",
	Long: `Watch the codebase for file changes and automatically
re-run queries or regenerate visualizations.

The watch command monitors your codebase for changes and automatically
updates results in real-time, enabling rapid development workflows.

Examples:
  # Watch and re-run query
  graphfs watch "SELECT * WHERE { ?s <#imports> ?o }"

  # Watch specific directory
  graphfs watch --path services/ "SELECT * WHERE { ... }"

  # Watch and regenerate visualization
  graphfs watch --viz --output graph.svg

  # Watch with custom debounce time
  graphfs watch --debounce 500ms "SELECT * WHERE { ... }"

Exit Codes:
  0 - Watch completed successfully (Ctrl+C)
  1 - Error during setup or execution`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

var (
	watchPath     string
	watchViz      bool
	watchOutput   string
	watchDebounce time.Duration
	watchVerbose  bool
)

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringVarP(&watchPath, "path", "p", ".", "Path to watch")
	watchCmd.Flags().BoolVar(&watchViz, "viz", false, "Regenerate visualization on changes")
	watchCmd.Flags().StringVarP(&watchOutput, "output", "o", "", "Output file for visualization")
	watchCmd.Flags().DurationVar(&watchDebounce, "debounce", 300*time.Millisecond, "Debounce duration for batching changes")
	watchCmd.Flags().BoolVarP(&watchVerbose, "verbose", "v", false, "Enable verbose output")
}

func runWatch(cmd *cobra.Command, args []string) error {
	// Color setup
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	gray := color.New(color.FgHiBlack)

	// Disable colors if --no-color flag is set
	if noColor {
		color.NoColor = true
	}

	// Validate flags
	if watchViz && watchOutput == "" {
		return fmt.Errorf("--output is required when using --viz")
	}

	var queryString string
	if len(args) > 0 {
		queryString = args[0]
	}

	if queryString == "" && !watchViz {
		return fmt.Errorf("either a query or --viz flag is required")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(watchPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	cyan.Println("🔍 Building initial graph...")

	// Build initial graph
	builder := graph.NewBuilder()
	opts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			MaxFileSize:    1024 * 1024, // 1MB
			FollowSymlinks: false,
			IgnoreFiles:    []string{".gitignore", ".graphfsignore"},
			UseDefaults:    true,
			Concurrent:     true,
			Workers:        0,
		},
		Validate:       false,
		ReportProgress: watchVerbose,
		UseCache:       true,
	}

	g, err := builder.Build(absPath, opts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	green.Printf("✓ Graph built: %d modules\n", g.Statistics.TotalModules)
	fmt.Println()

	// Execute initial query if provided
	var executor *query.Executor
	if queryString != "" {
		executor = query.NewExecutor(g.Store)
		if err := executeQuery(executor, queryString, green, yellow, red); err != nil {
			return err
		}
		fmt.Println()
	}

	// Generate initial visualization if requested
	if watchViz {
		if err := generateViz(g, watchOutput, green, yellow); err != nil {
			return err
		}
		fmt.Println()
	}

	// Setup watcher
	watchOpts := watch.WatchOptions{
		Path:     absPath,
		Debounce: watchDebounce,
		Verbose:  watchVerbose,
	}

	watcher, err := watch.NewWatcher(g, watchOpts, func(graph *graph.Graph, changedFiles []string) {
		// Show what changed
		gray.Printf("\n[Change detected: %d file(s)]\n", len(changedFiles))
		if watchVerbose {
			for _, file := range changedFiles {
				relPath, _ := filepath.Rel(absPath, file)
				fmt.Printf("  • %s\n", relPath)
			}
		}

		// Re-run query if provided
		if queryString != "" {
			if err := executeQuery(executor, queryString, green, yellow, red); err != nil {
				red.Printf("Query error: %v\n", err)
			}
		}

		// Regenerate visualization if requested
		if watchViz {
			if err := generateViz(graph, watchOutput, green, yellow); err != nil {
				red.Printf("Visualization error: %v\n", err)
			}
		}

		fmt.Println()
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	watcher.Start()
	defer watcher.Stop()

	cyan.Printf("👀 Watching for changes in %s\n", watchPath)
	gray.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	green.Println("✓ Watch stopped")

	return nil
}

// executeQuery runs a query and displays results
func executeQuery(executor *query.Executor, queryString string, green, yellow, red *color.Color) error {
	startTime := time.Now()

	result, err := executor.ExecuteString(queryString)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	duration := time.Since(startTime)

	// Display results
	if result.Count == 0 {
		yellow.Println("No results")
		return nil
	}

	green.Printf("Results (%d) in %v:\n", result.Count, duration)
	for _, bindings := range result.Bindings {
		for key, value := range bindings {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}

	return nil
}

// generateViz generates a visualization and saves it
func generateViz(g *graph.Graph, output string, green, yellow *color.Color) error {
	startTime := time.Now()

	// Determine viz type from extension
	ext := filepath.Ext(output)
	var format viz.OutputFormat
	switch ext {
	case ".svg":
		format = viz.FormatSVG
	case ".png":
		format = viz.FormatPNG
	case ".dot":
		format = viz.FormatDOT
	default:
		return fmt.Errorf("unsupported output format: %s (use .svg, .png, or .dot)", ext)
	}

	// Generate visualization
	if err := viz.RenderToFile(g, viz.RenderOptions{
		VizOptions: viz.VizOptions{
			Type:       viz.VizDependency,
			ShowLabels: true,
		},
		Output: output,
		Format: format,
	}); err != nil {
		return fmt.Errorf("failed to generate visualization: %w", err)
	}

	duration := time.Since(startTime)
	green.Printf("✓ Visualization updated (%v) → %s\n", duration, output)

	return nil
}
