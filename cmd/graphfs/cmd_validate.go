package main

import (
	"fmt"
	"os"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/rules"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	validateRulesFile string
	validateFormat    string
	validateSeverity  string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate architectural rules against the knowledge graph",
	Long: `Validate architectural rules against the knowledge graph.

Executes SPARQL-based rules to enforce architectural constraints and design principles.

Examples:
  # Validate with rules file
  graphfs validate --rules .graphfs-rules.yml

  # Output as JSON
  graphfs validate --rules .graphfs-rules.yml --format json

  # Output as JUnit XML for CI/CD
  graphfs validate --rules .graphfs-rules.yml --format junit > results.xml

  # Only check error-level rules
  graphfs validate --rules .graphfs-rules.yml --severity error`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&validateRulesFile, "rules", "r", "", "Path to rules file (YAML)")
	validateCmd.Flags().StringVarP(&validateFormat, "format", "f", "text", "Output format (text, json, junit)")
	validateCmd.Flags().StringVarP(&validateSeverity, "severity", "s", "info", "Minimum severity level (info, warning, error)")
	validateCmd.MarkFlagRequired("rules")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."

	// Build knowledge graph
	fmt.Fprintln(os.Stderr, "Building knowledge graph...")
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		Validate:       false,
		ReportProgress: false,
	}

	g, err := builder.Build(targetPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Loaded %d modules\n\n", len(g.Modules))

	// Parse severity level
	var minSeverity rules.Severity
	switch validateSeverity {
	case "error":
		minSeverity = rules.SeverityError
	case "warning":
		minSeverity = rules.SeverityWarning
	case "info":
		minSeverity = rules.SeverityInfo
	default:
		return fmt.Errorf("invalid severity level: %s (must be info, warning, or error)", validateSeverity)
	}

	// Create engine and validate
	engine := rules.NewEngine(g)

	// Parse rules file
	ruleSet, err := rules.ParseRules(validateRulesFile)
	if err != nil {
		return fmt.Errorf("failed to parse rules: %w", err)
	}

	// Validate with filters
	result, err := engine.ValidateWithFilter(ruleSet.Rules, nil, minSeverity)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Report results
	var format rules.OutputFormat
	switch validateFormat {
	case "json":
		format = rules.FormatJSON
	case "junit":
		format = rules.FormatJUnit
	default:
		format = rules.FormatText
	}

	reporter := rules.NewReporter(format)
	output := reporter.Report(result)
	fmt.Print(output)

	// Exit with error code if validation failed
	if !result.Success() {
		os.Exit(1)
	}

	return nil
}
