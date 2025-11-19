/*
# Module: cmd/graphfs/cmd_diff.go
Diff command for analyzing changes between commits.

Implements the 'graphfs diff' command for comparing knowledge graphs
between Git commits to understand impact of changes.

## Linked Modules
- [../../pkg/diff](../../pkg/diff/differ.go) - Graph diffing
- [root](./root.go) - Root command

## Tags
cli, diff, git

## Exports
diffCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_diff.go> a code:Module ;
    code:name "cmd/graphfs/cmd_diff.go" ;
    code:description "Diff command for analyzing changes between commits" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/diff/differ.go>, <./root.go> ;
    code:exports <#diffCmd> ;
    code:tags "cli", "diff", "git" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [ref]",
	Short: "Analyze changes between commits",
	Long: `Compare the knowledge graph between the current state
and a Git reference (commit, branch, tag).

The diff command helps you understand the impact of code changes
by comparing module dependencies, exports, and architectural layers
between different versions of your codebase.

Examples:
  # Diff against main branch
  graphfs diff main

  # Diff against specific commit
  graphfs diff abc123

  # Diff against previous commit
  graphfs diff HEAD~1

  # Export diff as JSON
  graphfs diff --format json main

  # Export diff as Markdown
  graphfs diff --format md --output CHANGES.md main

Exit Codes:
  0 - Diff completed successfully
  1 - Error during diff analysis`,
	Args: cobra.ExactArgs(1),
	RunE: runDiff,
}

var (
	diffOutput string
	diffFormat string
)

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVarP(&diffOutput, "output", "o", "", "Output file for report")
	diffCmd.Flags().StringVar(&diffFormat, "format", "text", "Output format (text, json, md)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	ref := args[0]

	// Get current working directory (must be a git repo)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate format
	var format diff.ReportFormat
	switch diffFormat {
	case "text":
		format = diff.FormatText
	case "json":
		format = diff.FormatJSON
	case "md", "markdown":
		format = diff.FormatMarkdown
	default:
		return fmt.Errorf("unknown format: %s (supported: text, json, md)", diffFormat)
	}

	// Create differ
	differ := diff.NewDiffer(absPath)

	// Show progress
	if !noColor {
		fmt.Printf("📊 Analyzing changes since %s...\n\n", ref)
	} else {
		fmt.Printf("Analyzing changes since %s...\n\n", ref)
	}

	// Perform diff
	result, err := differ.Diff(ref)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	// Format output
	report, err := diff.FormatDiff(result, format, !noColor)
	if err != nil {
		return fmt.Errorf("failed to format diff: %w", err)
	}

	// Write output
	if diffOutput != "" {
		if err := os.WriteFile(diffOutput, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("✓ Diff report written to %s\n", diffOutput)
	} else {
		fmt.Print(report)
	}

	return nil
}
