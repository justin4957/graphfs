/*
# Module: pkg/repl/formatter.go
Output formatters for REPL results.

Provides formatting for query results in table, JSON, and CSV formats.

## Linked Modules
- [../query](../query/executor.go) - Query result types
- [repl](./repl.go) - REPL core

## Tags
repl, formatter, output

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#formatter.go> a code:Module ;
    code:name "pkg/repl/formatter.go" ;
    code:description "Output formatters for REPL results" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <../query/executor.go>, <./repl.go> ;
    code:tags "repl", "formatter", "output" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/query"
)

// formatResult formats and displays query results
func (r *REPL) formatResult(result *query.QueryResult) error {
	if result == nil {
		r.printInfo("Query executed successfully (no results)")
		return nil
	}

	switch r.format {
	case "table":
		return r.formatTable(result)
	case "json":
		return r.formatJSON(result)
	case "csv":
		return r.formatCSV(result)
	default:
		return fmt.Errorf("unknown format: %s", r.format)
	}
}

// formatTable formats results as a table
func (r *REPL) formatTable(result *query.QueryResult) error {
	if len(result.Bindings) == 0 {
		r.printInfo("No results")
		return nil
	}

	// Get column headers
	vars := result.Variables
	if len(vars) == 0 {
		r.printInfo("No results")
		return nil
	}

	// Calculate column widths
	colWidths := make(map[string]int)
	for _, v := range vars {
		colWidths[v] = len(v)
	}

	for _, binding := range result.Bindings {
		for _, v := range vars {
			if val, ok := binding[v]; ok {
				valueStr := formatValue(val)
				if len(valueStr) > colWidths[v] {
					colWidths[v] = len(valueStr)
				}
			}
		}
	}

	// Cap column widths at 50 characters
	for v := range colWidths {
		if colWidths[v] > 50 {
			colWidths[v] = 50
		}
	}

	// Print header
	var headerParts []string
	for _, v := range vars {
		headerParts = append(headerParts, padRight(v, colWidths[v]))
	}

	if r.config.NoColor {
		fmt.Println(strings.Join(headerParts, " | "))
		fmt.Println(strings.Repeat("-", sumWidths(colWidths, len(vars))))
	} else {
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Println(strings.Join(headerParts, " | "))
		fmt.Println(strings.Repeat("-", sumWidths(colWidths, len(vars))))
	}

	// Print rows
	for _, binding := range result.Bindings {
		var rowParts []string
		for _, v := range vars {
			val := ""
			if bv, ok := binding[v]; ok {
				val = formatValue(bv)
				if len(val) > 50 {
					val = val[:47] + "..."
				}
			}
			rowParts = append(rowParts, padRight(val, colWidths[v]))
		}
		fmt.Println(strings.Join(rowParts, " | "))
	}

	return nil
}

// formatJSON formats results as JSON
func (r *REPL) formatJSON(result *query.QueryResult) error {
	data, err := json.MarshalIndent(result.Bindings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// formatCSV formats results as CSV
func (r *REPL) formatCSV(result *query.QueryResult) error {
	if len(result.Bindings) == 0 {
		r.printInfo("No results")
		return nil
	}

	w := csv.NewWriter(r.rl.Stdout())

	// Write header
	if err := w.Write(result.Variables); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, binding := range result.Bindings {
		row := make([]string, len(result.Variables))
		for i, v := range result.Variables {
			if val, ok := binding[v]; ok {
				row[i] = formatValue(val)
			}
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	w.Flush()
	return w.Error()
}

// formatValue formats a binding value as a string
func formatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// padRight pads a string to the right
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// sumWidths calculates the total width for the table
func sumWidths(widths map[string]int, numCols int) int {
	total := 0
	for _, w := range widths {
		total += w
	}
	// Add separators (3 chars per column minus the last one)
	total += (numCols - 1) * 3
	return total
}
