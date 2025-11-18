/*
# Module: pkg/cli/output.go
Rich CLI output formatter with colors, tables, and progress indicators.

Provides formatted output with color coding, table rendering, and progress bars
for a better user experience in the CLI.

## Linked Modules
- [../../cmd/graphfs](../../cmd/graphfs/root.go) - CLI commands

## Tags
cli, output, formatting, colors

## Exports
OutputFormatter, NewOutputFormatter

<!-- LinkedDoc RDF --
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#output.go> a code:Module ;
    code:name "pkg/cli/output.go" ;
    code:description "Rich CLI output formatter" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../cmd/graphfs/root.go> ;
    code:exports <#OutputFormatter>, <#NewOutputFormatter> ;
    code:tags "cli", "output", "formatting", "colors" .
<!-- End LinkedDoc RDF -->
*/

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
)

// OutputFormatter handles formatted CLI output
type OutputFormatter struct {
	writer  io.Writer
	quiet   bool
	verbose bool
	noColor bool
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(quiet, verbose, noColor bool) *OutputFormatter {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}

	// Disable color globally if requested
	if noColor {
		color.NoColor = true
	}

	return &OutputFormatter{
		writer:  os.Stdout,
		quiet:   quiet,
		verbose: verbose,
		noColor: noColor,
	}
}

// Success prints a success message in green
func (o *OutputFormatter) Success(format string, args ...interface{}) {
	if o.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		fmt.Fprintf(o.writer, "✓ %s\n", msg)
	} else {
		color.New(color.FgGreen, color.Bold).Fprintf(o.writer, "✓ %s\n", msg)
	}
}

// Error prints an error message in red
func (o *OutputFormatter) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	} else {
		color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "✗ %s\n", msg)
	}
}

// Warning prints a warning message in yellow
func (o *OutputFormatter) Warning(format string, args ...interface{}) {
	if o.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		fmt.Fprintf(o.writer, "⚠ %s\n", msg)
	} else {
		color.New(color.FgYellow).Fprintf(o.writer, "⚠ %s\n", msg)
	}
}

// Info prints an informational message in cyan
func (o *OutputFormatter) Info(format string, args ...interface{}) {
	if o.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		fmt.Fprintf(o.writer, "%s\n", msg)
	} else {
		color.New(color.FgCyan).Fprintf(o.writer, "%s\n", msg)
	}
}

// Debug prints a debug message (only in verbose mode)
func (o *OutputFormatter) Debug(format string, args ...interface{}) {
	if !o.verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		fmt.Fprintf(o.writer, "[DEBUG] %s\n", msg)
	} else {
		color.New(color.FgHiBlack).Fprintf(o.writer, "[DEBUG] %s\n", msg)
	}
}

// Println prints a plain message (respects quiet mode)
func (o *OutputFormatter) Println(format string, args ...interface{}) {
	if o.quiet {
		return
	}
	fmt.Fprintf(o.writer, format+"\n", args...)
}

// Header prints a section header
func (o *OutputFormatter) Header(text string) {
	if o.quiet {
		return
	}
	if o.noColor {
		fmt.Fprintf(o.writer, "\n%s\n%s\n", text, strings.Repeat("=", len(text)))
	} else {
		color.New(color.FgCyan, color.Bold).Fprintf(o.writer, "\n%s\n", text)
		color.New(color.FgCyan).Fprintf(o.writer, "%s\n", strings.Repeat("─", len(text)))
	}
}

// Table renders a formatted table without borders (simpler display)
func (o *OutputFormatter) Table(headers []string, rows [][]string) {
	if o.quiet && len(rows) == 0 {
		return
	}

	// Create simple table writer
	table := tablewriter.NewWriter(o.writer)

	// Set headers - convert []string to []any
	headerInterface := make([]interface{}, len(headers))
	for i, v := range headers {
		headerInterface[i] = v
	}
	table.Header(headerInterface...)

	// Add rows
	for _, row := range rows {
		rowInterface := make([]interface{}, len(row))
		for i, v := range row {
			rowInterface[i] = v
		}
		table.Append(rowInterface...)
	}

	// Render
	table.Render()
}

// TableWithBorders renders a table with borders
func (o *OutputFormatter) TableWithBorders(headers []string, rows [][]string) {
	if o.quiet && len(rows) == 0 {
		return
	}

	// Create table writer
	table := tablewriter.NewWriter(o.writer)

	// Set headers - convert []string to []any
	headerInterface := make([]interface{}, len(headers))
	for i, v := range headers {
		headerInterface[i] = v
	}
	table.Header(headerInterface...)

	// Add rows
	for _, row := range rows {
		rowInterface := make([]interface{}, len(row))
		for i, v := range row {
			rowInterface[i] = v
		}
		table.Append(rowInterface...)
	}

	// Render
	table.Render()
}

// ProgressBar creates a new progress bar
func (o *OutputFormatter) ProgressBar(total int, description string) *progressbar.ProgressBar {
	if o.quiet {
		// Return a silent progress bar
		return progressbar.NewOptions(total,
			progressbar.OptionSetWriter(io.Discard),
			progressbar.OptionSetVisibility(false),
		)
	}

	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(o.writer),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(!o.noColor),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

// KeyValue prints a key-value pair
func (o *OutputFormatter) KeyValue(key string, value interface{}) {
	if o.quiet {
		return
	}
	if o.noColor {
		fmt.Fprintf(o.writer, "  • %s: %v\n", key, value)
	} else {
		color.New(color.FgWhite).Fprintf(o.writer, "  • ")
		color.New(color.FgCyan).Fprintf(o.writer, "%s: ", key)
		color.New(color.FgWhite).Fprintf(o.writer, "%v\n", value)
	}
}

// BulletList prints a bulleted list
func (o *OutputFormatter) BulletList(items []string) {
	if o.quiet {
		return
	}
	for _, item := range items {
		if o.noColor {
			fmt.Fprintf(o.writer, "  • %s\n", item)
		} else {
			color.New(color.FgWhite).Fprintf(o.writer, "  • %s\n", item)
		}
	}
}

// Separator prints a horizontal line
func (o *OutputFormatter) Separator() {
	if o.quiet {
		return
	}
	if o.noColor {
		fmt.Fprintln(o.writer, strings.Repeat("─", 80))
	} else {
		color.New(color.FgHiBlack).Fprintln(o.writer, strings.Repeat("─", 80))
	}
}

// IsQuiet returns whether quiet mode is enabled
func (o *OutputFormatter) IsQuiet() bool {
	return o.quiet
}

// IsVerbose returns whether verbose mode is enabled
func (o *OutputFormatter) IsVerbose() bool {
	return o.verbose
}
