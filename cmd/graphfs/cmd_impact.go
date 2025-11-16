package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/spf13/cobra"
)

var (
	impactModules []string
	impactFormat  string
	impactCompare bool
)

var impactCmd = &cobra.Command{
	Use:   "impact [module-path]",
	Short: "Analyze the impact of modifying a module",
	Long: `Analyze the impact of modifying or removing a module from the codebase.

Shows:
- Direct and transitive dependencies
- Risk level assessment
- Affected modules by layer
- Recommendations for safe changes

Examples:
  # Analyze impact of a single module
  graphfs impact pkg/graph/graph.go

  # Analyze impact of multiple modules
  graphfs impact --modules pkg/graph/graph.go,pkg/graph/builder.go

  # Compare impacts of multiple modules
  graphfs impact --compare --modules pkg/graph/graph.go,pkg/analysis/impact.go

  # Output as JSON
  graphfs impact pkg/graph/graph.go --format json`,
	RunE: runImpact,
}

func init() {
	rootCmd.AddCommand(impactCmd)

	impactCmd.Flags().StringSliceVarP(&impactModules, "modules", "m", nil, "Comma-separated list of modules to analyze")
	impactCmd.Flags().StringVarP(&impactFormat, "format", "f", "text", "Output format (text, json)")
	impactCmd.Flags().BoolVarP(&impactCompare, "compare", "c", false, "Compare impacts of multiple modules")
}

func runImpact(cmd *cobra.Command, args []string) error {
	// Determine which modules to analyze
	var modulesToAnalyze []string

	if len(args) > 0 {
		modulesToAnalyze = append(modulesToAnalyze, args[0])
	}

	if len(impactModules) > 0 {
		modulesToAnalyze = append(modulesToAnalyze, impactModules...)
	}

	if len(modulesToAnalyze) == 0 {
		return fmt.Errorf("no module specified. Use: graphfs impact <module-path>")
	}

	// Determine target path
	targetPath := "."
	buildOpts := graph.BuildOptions{
		Validate:       false, // Don't validate - test files may have duplicate URIs
		ReportProgress: false,
	}

	fmt.Fprintln(os.Stderr, "Building knowledge graph...")
	builder := graph.NewBuilder()
	g, err := builder.Build(targetPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Loaded %d modules\n\n", len(g.Modules))

	// Create impact analyzer
	ia := analysis.NewImpactAnalysis(g)

	// Perform analysis
	if impactCompare && len(modulesToAnalyze) > 1 {
		return runCompareImpacts(ia, modulesToAnalyze)
	}

	if len(modulesToAnalyze) == 1 {
		return runSingleImpact(ia, modulesToAnalyze[0])
	}

	return runMultipleImpact(ia, modulesToAnalyze)
}

func runSingleImpact(ia *analysis.ImpactAnalysis, modulePath string) error {
	result, err := ia.AnalyzeImpact(modulePath)
	if err != nil {
		return fmt.Errorf("impact analysis failed: %w", err)
	}

	if impactFormat == "json" {
		return printImpactJSON(result)
	}

	return printImpactText(result)
}

func runMultipleImpact(ia *analysis.ImpactAnalysis, modules []string) error {
	result, err := ia.AnalyzeMultipleModules(modules)
	if err != nil {
		return fmt.Errorf("impact analysis failed: %w", err)
	}

	if impactFormat == "json" {
		return printImpactJSON(result)
	}

	fmt.Printf("🔍 Combined Impact Analysis: %d modules\n\n", len(modules))
	for _, module := range modules {
		fmt.Printf("  • %s\n", module)
	}
	fmt.Println()

	return printImpactText(result)
}

func runCompareImpacts(ia *analysis.ImpactAnalysis, modules []string) error {
	results, err := ia.CompareImpacts(modules)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	fmt.Printf("📊 Impact Comparison: %d modules\n\n", len(modules))

	// Sort by impact (highest first)
	type moduleImpact struct {
		path   string
		result *analysis.ImpactResult
	}

	sorted := make([]moduleImpact, 0, len(results))
	for path, result := range results {
		sorted = append(sorted, moduleImpact{path, result})
	}

	// Simple sort by total impacted modules
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].result.TotalImpactedModules > sorted[i].result.TotalImpactedModules {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Print comparison table
	for i, mi := range sorted {
		printComparisonRow(i+1, mi.path, mi.result)
	}

	return nil
}

func printComparisonRow(rank int, modulePath string, result *analysis.ImpactResult) {
	riskColor := getRiskColor(result.RiskLevel)

	fmt.Printf("%d. %s\n", rank, modulePath)
	fmt.Printf("   Risk: %s | ", riskColor.Sprint(result.RiskLevel))
	fmt.Printf("Impacted: %d modules (%.1f%%) | ",
		result.TotalImpactedModules,
		result.ImpactPercentage)
	fmt.Printf("Direct Dependents: %d\n", len(result.DirectDependents))
	fmt.Println()
}

func printImpactText(result *analysis.ImpactResult) error {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	riskColor := getRiskColor(result.RiskLevel)

	// Header
	cyan.Printf("🔍 Impact Analysis: %s\n\n", result.TargetModule)

	// Risk Level
	fmt.Printf("📊 Risk Level: %s", riskColor.Sprint(result.RiskLevel))
	if result.BreakingChanges {
		color.New(color.FgRed).Printf(" ⚠️  BREAKING CHANGES")
	}
	fmt.Println()

	// Impact Summary
	yellow.Println("Impact Summary:")
	fmt.Printf("  • Direct Dependents: %d\n", len(result.DirectDependents))
	fmt.Printf("  • Direct Dependencies: %d\n", len(result.DirectDependencies))
	fmt.Printf("  • Total Impacted Modules: %d (%.1f%% of codebase)\n",
		result.TotalImpactedModules,
		result.ImpactPercentage)
	fmt.Printf("  • Maximum Impact Depth: %d\n", result.MaxImpactDepth)
	fmt.Printf("  • Layers Impacted: %d\n", len(result.ImpactByLayer))
	fmt.Println()

	// Direct Dependents
	if len(result.DirectDependents) > 0 {
		yellow.Println("Direct Dependents:")
		for _, dep := range result.DirectDependents {
			fmt.Printf("  • %s\n", dep)
		}
		fmt.Println()
	}

	// Impact by Layer
	if len(result.ImpactByLayer) > 0 {
		yellow.Println("Impact by Layer:")
		for layer, count := range result.ImpactByLayer {
			fmt.Printf("  • %s: %d modules\n", layer, count)
		}
		fmt.Println()
	}

	// Risk Factors
	if len(result.RiskFactors) > 0 {
		yellow.Println("Risk Factors:")
		for _, factor := range result.RiskFactors {
			fmt.Printf("  • %s\n", factor)
		}
		fmt.Println()
	}

	// Recommendations
	if len(result.Recommendations) > 0 {
		yellow.Println("💡 Recommendations:")
		for _, rec := range result.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
		fmt.Println()
	}

	// Critical Paths
	if len(result.CriticalPaths) > 0 {
		yellow.Println("🔗 Critical Dependency Paths:")
		for i, path := range result.CriticalPaths {
			fmt.Printf("  %d. %s\n", i+1, strings.Join(path, " → "))
		}
		fmt.Println()
	}

	return nil
}

func printImpactJSON(result *analysis.ImpactResult) error {
	// Simple JSON output (could use json.Marshal for production)
	fmt.Println("{")
	fmt.Printf("  \"target_module\": \"%s\",\n", result.TargetModule)
	fmt.Printf("  \"risk_level\": \"%s\",\n", result.RiskLevel)
	fmt.Printf("  \"breaking_changes\": %v,\n", result.BreakingChanges)
	fmt.Printf("  \"total_impacted_modules\": %d,\n", result.TotalImpactedModules)
	fmt.Printf("  \"impact_percentage\": %.2f,\n", result.ImpactPercentage)
	fmt.Printf("  \"direct_dependents\": %d,\n", len(result.DirectDependents))
	fmt.Printf("  \"direct_dependencies\": %d,\n", len(result.DirectDependencies))
	fmt.Printf("  \"max_impact_depth\": %d,\n", result.MaxImpactDepth)
	fmt.Printf("  \"layers_impacted\": %d\n", len(result.ImpactByLayer))
	fmt.Println("}")
	return nil
}

func getRiskColor(level analysis.RiskLevel) *color.Color {
	switch level {
	case analysis.RiskLevelCritical:
		return color.New(color.FgRed, color.Bold)
	case analysis.RiskLevelHigh:
		return color.New(color.FgRed)
	case analysis.RiskLevelMedium:
		return color.New(color.FgYellow)
	case analysis.RiskLevelLow:
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgWhite)
	}
}
