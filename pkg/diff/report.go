/*
# Module: pkg/diff/report.go
Report formatting for graph diffs.

Generates human-readable and structured reports of graph differences in
multiple formats (text, JSON, Markdown).

## Linked Modules
- [differ](./differ.go) - Graph diff

## Tags
diff, reporting, formatting

## Exports
ReportFormat, FormatDiff, FormatJSON, FormatMarkdown, FormatText

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#report.go> a code:Module ;
    code:name "pkg/diff/report.go" ;
    code:description "Report formatting for graph diffs" ;
    code:language "go" ;
    code:layer "diff" ;
    code:linksTo <./differ.go> ;
    code:exports <#ReportFormat>, <#FormatDiff> ;
    code:tags "diff", "reporting", "formatting" .
<!-- End LinkedDoc RDF -->
*/

package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// ReportFormat specifies the output format for diff reports
type ReportFormat string

const (
	FormatText     ReportFormat = "text"
	FormatJSON     ReportFormat = "json"
	FormatMarkdown ReportFormat = "md"
)

// FormatDiff formats a graph diff according to the specified format
func FormatDiff(diff *GraphDiff, format ReportFormat, colorize bool) (string, error) {
	switch format {
	case FormatText:
		return formatText(diff, colorize), nil
	case FormatJSON:
		return formatJSON(diff)
	case FormatMarkdown:
		return formatMarkdown(diff), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

// formatText generates a text report with optional colors
func formatText(diff *GraphDiff, colorize bool) string {
	var b strings.Builder

	// Setup colors
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan, color.Bold)

	if !colorize {
		color.NoColor = true
		defer func() { color.NoColor = false }()
	}

	stats := diff.Stats()

	// Title
	cyan.Fprintln(&b, "📊 Graph Changes")
	fmt.Fprintln(&b)

	// Summary
	fmt.Fprintln(&b, "Summary:")
	if stats.Added > 0 {
		green.Fprintf(&b, "  • %d modules added\n", stats.Added)
	}
	if stats.Removed > 0 {
		red.Fprintf(&b, "  • %d modules removed\n", stats.Removed)
	}
	if stats.Modified > 0 {
		yellow.Fprintf(&b, "  • %d modules modified\n", stats.Modified)
	}
	fmt.Fprintf(&b, "  • %d modules unchanged\n", stats.Unchanged)
	fmt.Fprintln(&b)

	// Added modules
	if len(diff.Added) > 0 {
		green.Fprintln(&b, "Added Modules:")
		sort.Slice(diff.Added, func(i, j int) bool {
			return diff.Added[i].Path < diff.Added[j].Path
		})
		for _, mod := range diff.Added {
			green.Fprintf(&b, "  + %s\n", mod.Path)
		}
		fmt.Fprintln(&b)
	}

	// Removed modules
	if len(diff.Removed) > 0 {
		red.Fprintln(&b, "Removed Modules:")
		sort.Slice(diff.Removed, func(i, j int) bool {
			return diff.Removed[i].Path < diff.Removed[j].Path
		})
		for _, mod := range diff.Removed {
			red.Fprintf(&b, "  - %s\n", mod.Path)
		}
		fmt.Fprintln(&b)
	}

	// Modified modules
	if len(diff.Modified) > 0 {
		yellow.Fprintln(&b, "Modified Modules:")
		sort.Slice(diff.Modified, func(i, j int) bool {
			return diff.Modified[i].Module.Path < diff.Modified[j].Module.Path
		})
		for _, change := range diff.Modified {
			yellow.Fprintf(&b, "  ~ %s\n", change.Module.Path)

			// Layer changes
			if change.LayerChanged {
				fmt.Fprintf(&b, "    • Layer: %s → %s\n", change.OldLayer, change.NewLayer)
			}

			// Dependency changes
			if len(change.DepsAdded) > 0 || len(change.DepsRemoved) > 0 {
				added := len(change.DepsAdded)
				removed := len(change.DepsRemoved)
				fmt.Fprintf(&b, "    • Dependencies: +%d -%d\n", added, removed)

				for _, dep := range change.DepsAdded {
					green.Fprintf(&b, "      + %s\n", dep)
				}
				for _, dep := range change.DepsRemoved {
					red.Fprintf(&b, "      - %s\n", dep)
				}
			}

			// Export changes
			if len(change.ExportsAdded) > 0 || len(change.ExportsRemoved) > 0 {
				added := len(change.ExportsAdded)
				removed := len(change.ExportsRemoved)
				fmt.Fprintf(&b, "    • Exports: +%d -%d\n", added, removed)

				for _, exp := range change.ExportsAdded {
					green.Fprintf(&b, "      + %s\n", exp)
				}
				for _, exp := range change.ExportsRemoved {
					red.Fprintf(&b, "      - %s\n", exp)
				}
			}

			fmt.Fprintln(&b)
		}
	}

	return b.String()
}

// formatJSON generates a JSON report
func formatJSON(diff *GraphDiff) (string, error) {
	// Create a simplified structure for JSON output
	type jsonModule struct {
		Path  string `json:"path"`
		Layer string `json:"layer,omitempty"`
	}

	type jsonChange struct {
		Path           string   `json:"path"`
		DepsAdded      []string `json:"deps_added,omitempty"`
		DepsRemoved    []string `json:"deps_removed,omitempty"`
		ExportsAdded   []string `json:"exports_added,omitempty"`
		ExportsRemoved []string `json:"exports_removed,omitempty"`
		LayerChanged   bool     `json:"layer_changed,omitempty"`
		OldLayer       string   `json:"old_layer,omitempty"`
		NewLayer       string   `json:"new_layer,omitempty"`
	}

	output := struct {
		Stats    DiffStats    `json:"stats"`
		Added    []jsonModule `json:"added"`
		Removed  []jsonModule `json:"removed"`
		Modified []jsonChange `json:"modified"`
	}{
		Stats:    diff.Stats(),
		Added:    make([]jsonModule, len(diff.Added)),
		Removed:  make([]jsonModule, len(diff.Removed)),
		Modified: make([]jsonChange, len(diff.Modified)),
	}

	for i, mod := range diff.Added {
		output.Added[i] = jsonModule{Path: mod.Path, Layer: mod.Layer}
	}

	for i, mod := range diff.Removed {
		output.Removed[i] = jsonModule{Path: mod.Path, Layer: mod.Layer}
	}

	for i, change := range diff.Modified {
		output.Modified[i] = jsonChange{
			Path:           change.Module.Path,
			DepsAdded:      change.DepsAdded,
			DepsRemoved:    change.DepsRemoved,
			ExportsAdded:   change.ExportsAdded,
			ExportsRemoved: change.ExportsRemoved,
			LayerChanged:   change.LayerChanged,
			OldLayer:       change.OldLayer,
			NewLayer:       change.NewLayer,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatMarkdown generates a Markdown report
func formatMarkdown(diff *GraphDiff) string {
	var b strings.Builder

	stats := diff.Stats()

	// Title
	fmt.Fprintln(&b, "# Graph Changes")
	fmt.Fprintln(&b)

	// Summary
	fmt.Fprintln(&b, "## Summary")
	fmt.Fprintln(&b)
	if stats.Added > 0 {
		fmt.Fprintf(&b, "- **%d modules added**\n", stats.Added)
	}
	if stats.Removed > 0 {
		fmt.Fprintf(&b, "- **%d modules removed**\n", stats.Removed)
	}
	if stats.Modified > 0 {
		fmt.Fprintf(&b, "- **%d modules modified**\n", stats.Modified)
	}
	fmt.Fprintf(&b, "- %d modules unchanged\n", stats.Unchanged)
	fmt.Fprintln(&b)

	// Added modules
	if len(diff.Added) > 0 {
		fmt.Fprintln(&b, "## Added Modules")
		fmt.Fprintln(&b)
		sort.Slice(diff.Added, func(i, j int) bool {
			return diff.Added[i].Path < diff.Added[j].Path
		})
		for _, mod := range diff.Added {
			fmt.Fprintf(&b, "- `%s`\n", mod.Path)
		}
		fmt.Fprintln(&b)
	}

	// Removed modules
	if len(diff.Removed) > 0 {
		fmt.Fprintln(&b, "## Removed Modules")
		fmt.Fprintln(&b)
		sort.Slice(diff.Removed, func(i, j int) bool {
			return diff.Removed[i].Path < diff.Removed[j].Path
		})
		for _, mod := range diff.Removed {
			fmt.Fprintf(&b, "- `%s`\n", mod.Path)
		}
		fmt.Fprintln(&b)
	}

	// Modified modules
	if len(diff.Modified) > 0 {
		fmt.Fprintln(&b, "## Modified Modules")
		fmt.Fprintln(&b)
		sort.Slice(diff.Modified, func(i, j int) bool {
			return diff.Modified[i].Module.Path < diff.Modified[j].Module.Path
		})
		for _, change := range diff.Modified {
			fmt.Fprintf(&b, "### `%s`\n\n", change.Module.Path)

			// Layer changes
			if change.LayerChanged {
				fmt.Fprintf(&b, "- **Layer changed**: `%s` → `%s`\n", change.OldLayer, change.NewLayer)
			}

			// Dependency changes
			if len(change.DepsAdded) > 0 {
				fmt.Fprintln(&b, "- **Dependencies added**:")
				for _, dep := range change.DepsAdded {
					fmt.Fprintf(&b, "  - `%s`\n", dep)
				}
			}
			if len(change.DepsRemoved) > 0 {
				fmt.Fprintln(&b, "- **Dependencies removed**:")
				for _, dep := range change.DepsRemoved {
					fmt.Fprintf(&b, "  - `%s`\n", dep)
				}
			}

			// Export changes
			if len(change.ExportsAdded) > 0 {
				fmt.Fprintln(&b, "- **Exports added**:")
				for _, exp := range change.ExportsAdded {
					fmt.Fprintf(&b, "  - `%s`\n", exp)
				}
			}
			if len(change.ExportsRemoved) > 0 {
				fmt.Fprintln(&b, "- **Exports removed**:")
				for _, exp := range change.ExportsRemoved {
					fmt.Fprintf(&b, "  - `%s`\n", exp)
				}
			}

			fmt.Fprintln(&b)
		}
	}

	return b.String()
}
