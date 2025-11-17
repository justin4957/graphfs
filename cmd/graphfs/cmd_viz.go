package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/justin4957/graphfs/pkg/viz"
	"github.com/spf13/cobra"
)

var (
	vizType       string
	vizOutput     string
	vizLayout     string
	vizColorBy    string
	vizFormat     string
	vizTitle      string
	vizShowLabels bool
	vizRankdir    string
	vizLayers     []string
	vizTags       []string
	vizTarget     string
	vizModule     string
)

var vizCmd = &cobra.Command{
	Use:   "viz",
	Short: "Generate graph visualizations",
	Long: `Generate visual representations of dependency graphs, impact analysis,
security zones, and module relationships using GraphViz DOT format.

Visualization Types:
  • dependency - Full dependency graph
  • impact     - Impact analysis (requires --module)
  • security   - Security zone boundaries
  • layer      - Layer-based grouping

Output Formats:
  • dot     - DOT source file (default)
  • svg     - SVG vector graphics (requires graphviz)
  • png     - PNG raster graphics (requires graphviz)
  • pdf     - PDF document (requires graphviz)
  • mermaid - Mermaid diagram syntax (.mmd)
  • md      - Mermaid embedded in Markdown

Color Schemes:
  • language - Color by programming language
  • layer    - Color by architectural layer
  • default  - Default color scheme

Layout Engines:
  • dot    - Hierarchical layout (default)
  • neato  - Spring model layout
  • fdp    - Force-directed layout
  • circo  - Circular layout
  • twopi  - Radial layout

Examples:
  # Generate dependency graph as SVG
  graphfs viz --type dependency --output deps.svg

  # Layer-based visualization with custom layout
  graphfs viz --type layer --layout neato --output layers.png

  # Filtered dependency graph
  graphfs viz --type dependency --layer service --output services.svg

  # Security zones visualization
  graphfs viz --type security --output security.pdf

  # With title and labels
  graphfs viz --type dependency --title "My Project" --labels --output graph.svg

  # Generate Mermaid diagram
  graphfs viz --format mermaid --type dependency --output deps.mmd

  # Mermaid embedded in Markdown
  graphfs viz --format md --type dependency --title "Architecture" --output README.md`,
	RunE: runViz,
}

func init() {
	rootCmd.AddCommand(vizCmd)

	vizCmd.Flags().StringVarP(&vizType, "type", "t", "dependency",
		"Visualization type (dependency, impact, security, layer)")
	vizCmd.Flags().StringVarP(&vizOutput, "output", "o", "graph.dot",
		"Output file path")
	vizCmd.Flags().StringVarP(&vizLayout, "layout", "l", "dot",
		"GraphViz layout engine (dot, neato, fdp, circo, twopi)")
	vizCmd.Flags().StringVarP(&vizColorBy, "color-by", "c", "default",
		"Color scheme (language, layer, default)")
	vizCmd.Flags().StringVarP(&vizFormat, "format", "f", "",
		"Output format (dot, svg, png, pdf, mermaid, md) - auto-detected from extension")
	vizCmd.Flags().StringVar(&vizTitle, "title", "",
		"Graph title")
	vizCmd.Flags().BoolVar(&vizShowLabels, "labels", false,
		"Show detailed labels with descriptions")
	vizCmd.Flags().StringVar(&vizRankdir, "rankdir", "LR",
		"Graph direction (LR, TB, RL, BT)")
	vizCmd.Flags().StringSliceVar(&vizLayers, "layer", []string{},
		"Filter by layer(s)")
	vizCmd.Flags().StringSliceVar(&vizTags, "tag", []string{},
		"Filter by tag(s)")
	vizCmd.Flags().StringVarP(&vizTarget, "target", "d", ".",
		"Target directory to analyze")
	vizCmd.Flags().StringVarP(&vizModule, "module", "m", "",
		"Module for impact visualization")
}

func runViz(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	gray := color.New(color.FgHiBlack)

	cyan.Println("📊 Graph Visualization")
	cyan.Println()

	// Build knowledge graph
	gray.Printf("Building knowledge graph from %s...\n", vizTarget)
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		Validate:       false,
		ReportProgress: false,
	}

	g, err := builder.Build(vizTarget, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}
	gray.Printf("Graph built: %d modules\n\n", len(g.Modules))

	// Parse visualization type
	var vizTypeEnum viz.VizType
	switch vizType {
	case "dependency", "deps", "dep":
		vizTypeEnum = viz.VizDependency
	case "impact":
		vizTypeEnum = viz.VizImpact
	case "security", "sec":
		vizTypeEnum = viz.VizSecurity
	case "layer", "layers":
		vizTypeEnum = viz.VizLayer
	default:
		return fmt.Errorf("invalid visualization type: %s (use: dependency, impact, security, layer)", vizType)
	}

	// Create visualization options
	vizOpts := viz.VizOptions{
		Type:       vizTypeEnum,
		Layout:     vizLayout,
		ColorBy:    vizColorBy,
		Theme:      "default",
		ShowLabels: vizShowLabels,
		Rankdir:    vizRankdir,
		Title:      vizTitle,
	}

	// Add filter if specified
	if len(vizLayers) > 0 || len(vizTags) > 0 {
		vizOpts.Filter = &viz.FilterOptions{
			Layers: vizLayers,
			Tags:   vizTags,
		}
	}

	// For impact visualization, check if module is specified
	if vizTypeEnum == viz.VizImpact {
		if vizModule == "" {
			return fmt.Errorf("impact visualization requires --module flag")
		}

		gray.Printf("Analyzing impact of %s...\n", vizModule)
		ia := analysis.NewImpactAnalysis(g)
		impactResult, err := ia.AnalyzeImpact(vizModule)
		if err != nil {
			return fmt.Errorf("impact analysis failed: %w", err)
		}

		vizOpts.Impact = impactResult
		gray.Printf("Impact: %d modules affected\n\n",
			impactResult.TotalImpactedModules)
	}

	// For security visualization, run security analysis
	if vizTypeEnum == viz.VizSecurity {
		gray.Println("Analyzing security boundaries...")
		secOpts := analysis.SecurityOptions{
			StrictMode: false,
		}

		secAnalysis, err := analysis.AnalyzeSecurity(g, secOpts)
		if err != nil {
			return fmt.Errorf("security analysis failed: %w", err)
		}

		vizOpts.Security = secAnalysis
		gray.Printf("Security zones: %d zones, %d violations\n\n",
			len(secAnalysis.Zones), len(secAnalysis.Violations))
	}

	// Generate visualization
	gray.Printf("Generating %s visualization...\n", vizType)

	// Check if Mermaid format is requested
	isMermaid := vizFormat == "mermaid" || vizFormat == "md" ||
		strings.HasSuffix(vizOutput, ".mmd") || strings.HasSuffix(vizOutput, ".md")

	if isMermaid {
		// Generate Mermaid diagram
		mermaidOpts := viz.MermaidOptions{
			Type:      viz.MermaidFlowchart,
			Direction: vizRankdir,
			ColorBy:   vizColorBy,
			Title:     vizTitle,
		}

		// Add filter if specified
		if len(vizLayers) > 0 || len(vizTags) > 0 {
			mermaidOpts.Filter = &viz.FilterOptions{
				Layers: vizLayers,
				Tags:   vizTags,
			}
		}

		var mermaid string
		var err error

		// Generate based on visualization type
		if vizTypeEnum == viz.VizImpact && vizOpts.Impact != nil {
			mermaid, err = viz.GenerateMermaidForImpact(g, vizOpts.Impact, mermaidOpts)
		} else {
			// Embed in markdown if .md extension
			if vizFormat == "md" || strings.HasSuffix(vizOutput, ".md") {
				mermaid, err = viz.GenerateMermaidMarkdown(g, mermaidOpts)
			} else {
				mermaid, err = viz.GenerateMermaid(g, mermaidOpts)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to generate Mermaid diagram: %w", err)
		}

		// Write to file
		if err := os.WriteFile(vizOutput, []byte(mermaid), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else {
		// Use GraphViz for other formats
		// Validate layout if using GraphViz
		if vizFormat != "" && vizFormat != "dot" {
			if err := viz.ValidateLayout(vizLayout); err != nil {
				gray.Printf("Warning: %v\n", err)
				gray.Println("Falling back to DOT format")
				vizFormat = "dot"
			}
		}

		// Render to file
		renderOpts := viz.RenderOptions{
			VizOptions: vizOpts,
			Output:     vizOutput,
			Format:     viz.OutputFormat(vizFormat),
		}

		if err := viz.RenderToFile(g, renderOpts); err != nil {
			return fmt.Errorf("failed to render visualization: %w", err)
		}
	}

	// Success message
	green.Printf("✓ Visualization saved to %s\n", vizOutput)

	// Show tips based on output format
	if isMermaid {
		cyan.Println("\n💡 Tips:")
		fmt.Println("  • View in GitHub: Commit .md/.mmd file to repository")
		fmt.Println("  • Preview in VS Code: Install Mermaid extension")
		fmt.Println("  • View online: https://mermaid.live/")
		fmt.Println("  • Embed in docs: Copy into any Markdown file")
	} else {
		ext := vizOutput[len(vizOutput)-4:]
		if ext == ".dot" {
			cyan.Println("\n💡 Tips:")
			fmt.Println("  • Convert to SVG: dot -Tsvg graph.dot -o graph.svg")
			fmt.Println("  • Convert to PNG: dot -Tpng graph.dot -o graph.png")
			fmt.Println("  • View online: https://dreampuf.github.io/GraphvizOnline/")
			fmt.Println("  • Install GraphViz: brew install graphviz (or apt/yum/choco)")
		}
	}

	return nil
}
