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
	securityStrict bool
	securityTarget string
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Analyze security boundaries and detect violations",
	Long: `Analyze security boundaries across the codebase.

Classifies modules into security zones, detects boundary crossings,
identifies unauthorized access patterns, and generates security audit reports.

Security Zones:
  • public   - External APIs and public endpoints
  • trusted  - Internal services with authentication
  • internal - Private modules and implementation
  • admin    - Administrative and privileged functions
  • data     - Database and storage layer

Examples:
  # Basic security analysis
  graphfs security

  # Strict mode (flags all high-risk crossings)
  graphfs security --strict

  # Analyze specific directory
  graphfs security --target ./services`,
	RunE: runSecurity,
}

func init() {
	rootCmd.AddCommand(securityCmd)

	securityCmd.Flags().BoolVarP(&securityStrict, "strict", "s", false,
		"Strict mode - flag all high-risk crossings")
	securityCmd.Flags().StringVarP(&securityTarget, "target", "t", ".",
		"Target directory to analyze")
}

func runSecurity(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	gray := color.New(color.FgHiBlack)

	cyan.Println("🔒 Security Boundary Analysis")
	cyan.Println()

	// Build knowledge graph
	gray.Printf("Building knowledge graph from %s...\n", securityTarget)
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		Validate:       false,
		ReportProgress: false,
	}

	g, err := builder.Build(securityTarget, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}
	gray.Printf("Graph built: %d modules\n\n", len(g.Modules))

	// Perform security analysis
	gray.Println("Analyzing security boundaries...")
	opts := analysis.SecurityOptions{
		StrictMode: securityStrict,
	}

	result, err := analysis.AnalyzeSecurity(g, opts)
	if err != nil {
		return fmt.Errorf("security analysis failed: %w", err)
	}

	// Display security zones
	cyan.Println("Security Zones:")
	zoneOrder := []analysis.SecurityZone{
		analysis.ZonePublic,
		analysis.ZoneTrusted,
		analysis.ZoneInternal,
		analysis.ZoneAdmin,
		analysis.ZoneData,
		analysis.ZoneUnknown,
	}

	zoneIcons := map[analysis.SecurityZone]string{
		analysis.ZonePublic:   "🌐",
		analysis.ZoneTrusted:  "🔐",
		analysis.ZoneInternal: "🔒",
		analysis.ZoneAdmin:    "👑",
		analysis.ZoneData:     "💾",
		analysis.ZoneUnknown:  "❓",
	}

	for _, zone := range zoneOrder {
		if modules, ok := result.Zones[zone]; ok && len(modules) > 0 {
			icon := zoneIcons[zone]
			info := analysis.GetZoneInfo(zone)
			fmt.Printf("  %s %s (%d modules) - %s\n",
				icon, zone, len(modules), info.Description)
		}
	}
	fmt.Println()

	// Display boundary crossings
	cyan.Println("Boundary Crossings:")
	if len(result.Boundaries) == 0 {
		green.Println("  ✓ No boundary crossings detected")
	} else {
		for _, boundary := range result.Boundaries {
			crossingCount := len(boundary.Crossings)

			if boundary.Allowed && !securityStrict {
				green.Printf("  ✓ %s → %s (%d crossings, allowed)\n",
					boundary.From, boundary.To, crossingCount)
			} else if !boundary.Allowed {
				red.Printf("  ❌ %s → %s (%d crossings, UNAUTHORIZED)\n",
					boundary.From, boundary.To, crossingCount)

				// Show first few crossings
				maxShow := 3
				for i, crossing := range boundary.Crossings {
					if i >= maxShow {
						gray.Printf("     ... and %d more\n", crossingCount-maxShow)
						break
					}
					fmt.Printf("     • %s → %s\n",
						crossing.Source.Path, crossing.Destination.Path)
					gray.Printf("       Risk: %s\n", crossing.Risk)
				}
			} else {
				yellow.Printf("  ⚠️  %s → %s (%d crossings, needs review)\n",
					boundary.From, boundary.To, crossingCount)
			}
		}
	}
	fmt.Println()

	// Display violations
	if len(result.Violations) > 0 {
		cyan.Printf("Security Violations (%d):\n", len(result.Violations))

		criticalViolations := result.GetCriticalViolations()
		highRiskViolations := result.GetHighRiskViolations()

		if len(criticalViolations) > 0 {
			red.Printf("\n  CRITICAL (%d):\n", len(criticalViolations))
			for _, v := range criticalViolations {
				fmt.Printf("    • %s\n", v.Description)
				if v.Crossing != nil {
					fmt.Printf("      %s → %s\n",
						v.Crossing.Source.Path, v.Crossing.Destination.Path)
				}
				yellow.Printf("      💡 %s\n", v.Recommendation)
			}
		}

		if len(highRiskViolations) > len(criticalViolations) {
			highOnly := len(highRiskViolations) - len(criticalViolations)
			yellow.Printf("\n  HIGH RISK (%d):\n", highOnly)
			count := 0
			for _, v := range highRiskViolations {
				if v.Risk != analysis.RiskLevelCritical {
					fmt.Printf("    • %s\n", v.Description)
					if v.Crossing != nil {
						fmt.Printf("      %s → %s\n",
							v.Crossing.Source.Path, v.Crossing.Destination.Path)
					}
					count++
					if count >= 5 {
						gray.Printf("    ... and %d more\n", highOnly-count)
						break
					}
				}
			}
		}
		fmt.Println()
	}

	// Display risk score
	cyan.Println("Risk Assessment:")
	riskColor := green
	riskLabel := "LOW"

	if result.RiskScore >= 7.0 {
		riskColor = red
		riskLabel = "CRITICAL"
	} else if result.RiskScore >= 5.0 {
		riskColor = red
		riskLabel = "HIGH"
	} else if result.RiskScore >= 3.0 {
		riskColor = yellow
		riskLabel = "MEDIUM"
	}

	fmt.Printf("  Risk Score: ")
	riskColor.Printf("%.1f/10.0 (%s)\n", result.RiskScore, riskLabel)
	fmt.Printf("  Analysis Time: %v\n", result.Duration)
	fmt.Println()

	// Display recommendations
	if len(result.Recommendations) > 0 {
		cyan.Println("💡 Recommendations:")
		for _, rec := range result.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
		fmt.Println()
	}

	// Summary
	if !result.HasViolations() {
		green.Println("✅ No security violations detected")
		green.Println("Security boundaries are properly enforced.")
	} else {
		red.Println("❌ Security violations detected")
		fmt.Printf("Address %d violations to improve security posture.\n", len(result.Violations))

		// Exit with error code for CI integration
		os.Exit(1)
	}

	return nil
}
