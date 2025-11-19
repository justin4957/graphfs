/*
# Module: cmd/graphfs/cmd_doctor.go
Doctor command for system diagnostics.

Implements the 'graphfs doctor' command for running health checks and diagnostics.

## Linked Modules
- [../../pkg/doctor](../../pkg/doctor/doctor.go) - Health check system
- [root](./root.go) - Root command

## Tags
cli, diagnostics, doctor

## Exports
doctorCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_doctor.go> a code:Module ;
    code:name "cmd/graphfs/cmd_doctor.go" ;
    code:description "Doctor command for system diagnostics" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/doctor/doctor.go>, <./root.go> ;
    code:exports <#doctorCmd> ;
    code:tags "cli", "diagnostics", "doctor" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics and health checks",
	Long: `Run comprehensive diagnostics to check GraphFS installation,
configuration, and performance. Reports issues and provides
recommendations for fixes.

The doctor command checks:
- GraphFS and Go versions
- Optional dependencies (GraphViz)
- Cache integrity
- Configuration file
- Disk space
- File permissions
- Parser performance

Exit Codes:
  0 - All checks passed (or only warnings)
  1 - One or more critical errors found

Examples:
  graphfs doctor           # Run all diagnostics
  graphfs doctor --verbose # Show detailed output`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Color setup
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	// Disable colors if --no-color flag is set
	if noColor {
		color.NoColor = true
	}

	cyan.Println("🔍 Running GraphFS Diagnostics")
	fmt.Println()

	// Determine root path
	rootPath, err := os.Getwd()
	if err != nil {
		rootPath = "."
	}
	rootPath, _ = filepath.Abs(rootPath)

	// Run all health checks
	checks := doctor.RunAllChecks(rootPath, Version)

	issues := 0
	warnings := 0

	// Display results
	for _, check := range checks {
		switch check.Status {
		case doctor.StatusOK:
			green.Printf("✓ %s", check.Name)
			if verbose && check.Message != "" {
				fmt.Printf(" - %s", check.Message)
			}
			fmt.Println()

		case doctor.StatusWarning:
			yellow.Printf("⚠ %s\n", check.Name)
			if check.Message != "" {
				yellow.Printf("  %s\n", check.Message)
			}
			if check.Fix != "" {
				fmt.Printf("  Fix: %s\n", check.Fix)
			}
			warnings++

		case doctor.StatusError:
			red.Printf("✗ %s\n", check.Name)
			if check.Message != "" {
				red.Printf("  %s\n", check.Message)
			}
			if check.Fix != "" {
				fmt.Printf("  Fix: %s\n", check.Fix)
			}
			issues++
		}
	}

	fmt.Println()

	// Summary
	if issues == 0 && warnings == 0 {
		green.Println("✓ All checks passed! GraphFS is healthy.")
		return nil
	}

	if issues > 0 {
		red.Printf("❌ %d critical issue(s) found\n", issues)
	}
	if warnings > 0 {
		yellow.Printf("⚠️  %d warning(s) found\n", warnings)
	}

	// Exit with error code if critical issues found
	if issues > 0 {
		os.Exit(1)
	}

	return nil
}
