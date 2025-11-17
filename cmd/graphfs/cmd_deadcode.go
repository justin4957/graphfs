package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	deadCodeConfidence float64
	deadCodeExclude    []string
	deadCodeScript     string
	deadCodeAggressive bool
	deadCodeTarget     string
)

var deadCodeCmd = &cobra.Command{
	Use:   "dead-code",
	Short: "Detect dead code in the codebase",
	Long: `Detect dead code including unreferenced modules, unused dependencies, and unexported symbols.

Analyzes the dependency graph to identify modules with no incoming references,
unexported symbols that are never used, and dependencies that are declared but not used.

Examples:
  # Basic dead code detection
  graphfs dead-code

  # High confidence only
  graphfs dead-code --confidence 0.8

  # Exclude patterns
  graphfs dead-code --exclude "experimental/**" --exclude "**/*_test.go"

  # Generate removal script
  graphfs dead-code --script cleanup.sh

  # Aggressive mode (more likely to flag code as dead)
  graphfs dead-code --aggressive`,
	RunE: runDeadCode,
}

func init() {
	rootCmd.AddCommand(deadCodeCmd)

	deadCodeCmd.Flags().Float64VarP(&deadCodeConfidence, "confidence", "c", 0.5,
		"Minimum confidence threshold (0.0-1.0)")
	deadCodeCmd.Flags().StringSliceVarP(&deadCodeExclude, "exclude", "e", []string{},
		"Glob patterns to exclude")
	deadCodeCmd.Flags().StringVarP(&deadCodeScript, "script", "s", "",
		"Generate cleanup script to file")
	deadCodeCmd.Flags().BoolVarP(&deadCodeAggressive, "aggressive", "a", false,
		"Aggressive mode (more likely to flag code as dead)")
	deadCodeCmd.Flags().StringVarP(&deadCodeTarget, "target", "t", ".",
		"Target directory to analyze")
}

func runDeadCode(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	gray := color.New(color.FgHiBlack)

	cyan.Println("🔍 Dead Code Analysis")
	cyan.Println()

	// Build knowledge graph
	gray.Printf("Building knowledge graph from %s...\n", deadCodeTarget)
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		Validate:       false,
		ReportProgress: false,
	}
	g, err := builder.Build(deadCodeTarget, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}
	gray.Printf("Graph built: %d modules\n\n", len(g.Modules))

	// Configure detection options
	opts := analysis.DeadCodeOptions{
		MinConfidence:   deadCodeConfidence,
		ExcludePatterns: deadCodeExclude,
		AggressiveMode:  deadCodeAggressive,
	}

	// Perform dead code detection
	gray.Println("Analyzing for dead code...")
	result, err := analysis.DetectDeadCode(g, opts)
	if err != nil {
		return fmt.Errorf("dead code detection failed: %w", err)
	}

	// Display results
	if !result.HasDeadCode() {
		green.Println("✅ No dead code found!")
		green.Printf("All %d modules are actively referenced.\n", result.TotalModules)
		return nil
	}

	// Get categorized modules
	safeRemovals := result.GetSafeRemovals()
	needsReview := result.GetNeedsReview()

	// Safe removals
	if len(safeRemovals) > 0 {
		green.Printf("\n✓ Safe to Remove (%d modules):\n", len(safeRemovals))
		for _, dm := range safeRemovals {
			fmt.Printf("  • %s\n", dm.Module.Path)
			gray.Printf("    Reason: %s\n", dm.Reason)
			gray.Printf("    Confidence: %.0f%%\n", dm.Confidence*100)
			if len(dm.Suggestions) > 0 {
				yellow.Printf("    → %s\n", dm.Suggestions[0])
			}
			fmt.Println()
		}
	}

	// Needs review
	if len(needsReview) > 0 {
		yellow.Printf("\n⚠  Needs Review (%d modules):\n", len(needsReview))
		for _, dm := range needsReview {
			fmt.Printf("  • %s\n", dm.Module.Path)
			gray.Printf("    Reason: %s\n", dm.Reason)
			gray.Printf("    Confidence: %.0f%%\n", dm.Confidence*100)
			if len(dm.Suggestions) > 0 {
				cyan.Printf("    → %s\n", dm.Suggestions[0])
			}
			fmt.Println()
		}
	}

	// Coverage analysis
	cyan.Println("\n📊 Usage Coverage:")
	coverage := analysis.AnalyzeCoverage(g)
	fmt.Printf("  • Total modules: %d\n", coverage.TotalModules)
	fmt.Printf("  • Referenced: %d (%.1f%%)\n", coverage.ReferencedModules, coverage.CoveragePercent)
	fmt.Printf("  • Unreferenced: %d (%.1f%%)\n", coverage.UnreferencedModules,
		100-coverage.CoveragePercent)

	if len(coverage.HighUsageModules) > 0 {
		fmt.Printf("\n  Top 3 most used modules:\n")
		for i, cov := range coverage.HighUsageModules {
			if i >= 3 {
				break
			}
			fmt.Printf("    %d. %s (%d references)\n", i+1, cov.Module.Path, cov.IncomingRefs)
		}
	}

	// Generate cleanup plan
	plan := analysis.GenerateCleanupPlan(result)

	// Summary
	cyan.Println("\n📋 Summary:")
	if len(safeRemovals) > 0 {
		fmt.Printf("  • %d modules safe to remove\n", len(safeRemovals))
		fmt.Printf("  • Estimated %d lines can be deleted\n", plan.TotalLines)
	}
	if len(needsReview) > 0 {
		fmt.Printf("  • %d modules need manual review\n", len(needsReview))
	}
	fmt.Printf("  • Overall confidence: %.0f%%\n", result.Confidence*100)
	fmt.Printf("  • Analysis time: %v\n", result.Duration)

	// Generate cleanup script if requested
	if deadCodeScript != "" {
		gray.Printf("\nGenerating cleanup script: %s\n", deadCodeScript)
		script := analysis.GenerateScript(plan)
		if err := os.WriteFile(deadCodeScript, []byte(script), 0755); err != nil {
			return fmt.Errorf("failed to write script: %w", err)
		}
		green.Printf("✓ Script generated successfully\n")
		yellow.Printf("⚠  Review the script carefully before executing!\n")
	}

	// Final recommendations
	if len(safeRemovals) > 0 {
		cyan.Println("\n💡 Next Steps:")
		fmt.Println("  1. Review the safe removal suggestions above")
		if deadCodeScript == "" {
			fmt.Println("  2. Generate a cleanup script: graphfs dead-code --script cleanup.sh")
		} else {
			fmt.Printf("  2. Review the generated script: %s\n", deadCodeScript)
			fmt.Printf("  3. Execute the script: bash %s\n", deadCodeScript)
		}
		if len(needsReview) > 0 {
			fmt.Println("  4. Manually review modules with lower confidence")
		}
	} else if len(needsReview) > 0 {
		cyan.Println("\n💡 Recommendation:")
		fmt.Println("  All detected dead code needs manual review before removal.")
		fmt.Println("  Consider using --aggressive mode for more suggestions.")
	}

	// Exit with error if dead code found (for CI integration)
	if result.HasDeadCode() && len(safeRemovals) > 0 {
		fmt.Println()
		red.Println("❌ Dead code detected")
		os.Exit(1)
	}

	return nil
}
