package viz

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
)

// OutputFormat represents the output format
type OutputFormat string

const (
	FormatDOT OutputFormat = "dot" // DOT source
	FormatSVG OutputFormat = "svg" // SVG vector graphics
	FormatPNG OutputFormat = "png" // PNG raster graphics
	FormatPDF OutputFormat = "pdf" // PDF document
)

// RenderOptions configures rendering
type RenderOptions struct {
	VizOptions
	Output string       // Output file path
	Format OutputFormat // Output format
}

// RenderToFile renders a graph to a file
func RenderToFile(g *graph.Graph, opts RenderOptions) error {
	// Generate DOT format
	dotContent, err := GenerateDOT(g, opts.VizOptions)
	if err != nil {
		return fmt.Errorf("failed to generate DOT: %w", err)
	}

	// Determine output format from file extension if not specified
	if opts.Format == "" {
		ext := strings.ToLower(filepath.Ext(opts.Output))
		switch ext {
		case ".dot":
			opts.Format = FormatDOT
		case ".svg":
			opts.Format = FormatSVG
		case ".png":
			opts.Format = FormatPNG
		case ".pdf":
			opts.Format = FormatPDF
		default:
			opts.Format = FormatDOT // Default to DOT
		}
	}

	// For DOT format, just write the file
	if opts.Format == FormatDOT {
		return os.WriteFile(opts.Output, []byte(dotContent), 0644)
	}

	// For other formats, check if GraphViz is available
	if !isGraphVizAvailable() {
		// Fall back to DOT format with warning
		dotPath := strings.TrimSuffix(opts.Output, filepath.Ext(opts.Output)) + ".dot"
		err := os.WriteFile(dotPath, []byte(dotContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write DOT file: %w", err)
		}
		return fmt.Errorf("GraphViz not available, saved as DOT format to %s (install graphviz to render %s)", dotPath, opts.Format)
	}

	// Render using GraphViz
	return renderWithGraphViz(dotContent, opts)
}

// isGraphVizAvailable checks if GraphViz is installed
func isGraphVizAvailable() bool {
	_, err := exec.LookPath("dot")
	return err == nil
}

// renderWithGraphViz renders DOT content using GraphViz
func renderWithGraphViz(dotContent string, opts RenderOptions) error {
	// Determine the GraphViz command based on layout
	cmd := opts.Layout
	if cmd == "" {
		cmd = "dot"
	}

	// Create command
	command := exec.Command(cmd, fmt.Sprintf("-T%s", opts.Format), "-o", opts.Output)
	command.Stdin = strings.NewReader(dotContent)

	// Capture stderr for error messages
	var stderr strings.Builder
	command.Stderr = &stderr

	// Run command
	if err := command.Run(); err != nil {
		return fmt.Errorf("GraphViz rendering failed: %s: %w", stderr.String(), err)
	}

	return nil
}

// GetAvailableLayouts returns available GraphViz layout engines
func GetAvailableLayouts() []string {
	layouts := []string{"dot", "neato", "fdp", "circo", "twopi", "sfdp"}
	available := make([]string, 0)

	for _, layout := range layouts {
		if _, err := exec.LookPath(layout); err == nil {
			available = append(available, layout)
		}
	}

	return available
}

// ValidateLayout checks if a layout engine is available
func ValidateLayout(layout string) error {
	if layout == "" {
		layout = "dot"
	}

	if _, err := exec.LookPath(layout); err != nil {
		return fmt.Errorf("layout engine '%s' not found (install graphviz)", layout)
	}

	return nil
}
