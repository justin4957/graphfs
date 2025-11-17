package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/docs"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	docsOutputDir    string
	docsFormat       string
	docsTemplate     string
	docsLayers       []string
	docsTags         []string
	docsDepth        int
	docsIncludeGraph bool
	docsTitle        string
	docsAuthor       string
	docsVersion      string
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation from the knowledge graph",
	Long: `Generate comprehensive documentation from the knowledge graph in multiple formats.

Supports three output formats:
  - single: Single README.md file with all modules
  - multi:  One markdown file per module with an index
  - directory: Organized by directory structure

The documentation includes:
  - Module descriptions and metadata
  - Dependencies and dependents
  - Exported functions and types
  - Cross-links between modules
  - Project statistics and overview
  - Optional frontmatter for static site generators

Examples:
  # Generate single README.md file
  graphfs docs --format single --output ./docs

  # Generate multi-file documentation
  graphfs docs --format multi --output ./docs

  # Generate directory-structured docs
  graphfs docs --format directory --output ./docs

  # Filter by layer
  graphfs docs --format single --layer service --layer api

  # Filter by tag
  graphfs docs --tag security --tag api

  # Include frontmatter for Jekyll/Hugo
  graphfs docs --format single --author "GraphFS Team" --version "1.0.0"`,
	RunE: runDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)

	docsCmd.Flags().StringVarP(&docsOutputDir, "output", "o", "./docs", "Output directory for documentation")
	docsCmd.Flags().StringVarP(&docsFormat, "format", "f", "single", "Output format: single, multi, directory")
	docsCmd.Flags().StringVar(&docsTemplate, "template", "", "Custom template file")
	docsCmd.Flags().StringSliceVar(&docsLayers, "layer", []string{}, "Filter by layer (can be specified multiple times)")
	docsCmd.Flags().StringSliceVar(&docsTags, "tag", []string{}, "Filter by tag (can be specified multiple times)")
	docsCmd.Flags().IntVar(&docsDepth, "depth", 0, "Maximum dependency depth (0 for unlimited)")
	docsCmd.Flags().BoolVar(&docsIncludeGraph, "include-graph", false, "Include dependency graph visualizations")
	docsCmd.Flags().StringVar(&docsTitle, "title", "", "Documentation title (defaults to project name)")
	docsCmd.Flags().StringVar(&docsAuthor, "author", "", "Author name for frontmatter")
	docsCmd.Flags().StringVar(&docsVersion, "version", "", "Version for frontmatter")
}

func runDocs(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absPath)
	}

	// Load configuration
	configPath := filepath.Join(absPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: Could not load config, using defaults: %v\n", err)
		}
		config = DefaultConfig()
	}

	// Validate format
	var format docs.DocsFormat
	switch docsFormat {
	case "single":
		format = docs.DocsSingleFile
	case "multi":
		format = docs.DocsMultiFile
	case "directory":
		format = docs.DocsDirectory
	default:
		return fmt.Errorf("invalid format: %s (must be single, multi, or directory)", docsFormat)
	}

	// Build the knowledge graph
	fmt.Println("Building knowledge graph...")

	// Setup scan options
	scanOpts := scanner.ScanOptions{
		IncludePatterns: config.Scan.Include,
		ExcludePatterns: config.Scan.Exclude,
		MaxFileSize:     config.Scan.MaxFileSize,
		UseDefaults:     true,
		IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
		Concurrent:      true,
	}

	// Build graph
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions:    scanOpts,
		Validate:       false,
		ReportProgress: verbose,
	}

	g, err := builder.Build(absPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	if g.Statistics.TotalModules == 0 {
		return fmt.Errorf("no modules found to document")
	}

	fmt.Printf("Found %d modules\n", g.Statistics.TotalModules)

	// Prepare documentation options
	projectName := filepath.Base(absPath)

	title := docsTitle
	if title == "" {
		title = fmt.Sprintf("%s Documentation", projectName)
	}

	frontMatter := make(map[string]string)
	if docsAuthor != "" {
		frontMatter["author"] = docsAuthor
	}
	if docsVersion != "" {
		frontMatter["version"] = docsVersion
	}
	if len(frontMatter) > 0 {
		frontMatter["title"] = title
		frontMatter["project"] = projectName
	}

	docsOpts := docs.DocsOptions{
		OutputDir:     docsOutputDir,
		Format:        format,
		Template:      docsTemplate,
		IncludeLayers: docsLayers,
		IncludeTags:   docsTags,
		Depth:         docsDepth,
		IncludeGraph:  docsIncludeGraph,
		Title:         title,
		ProjectName:   projectName,
		FrontMatter:   frontMatter,
	}

	// Generate documentation
	fmt.Printf("Generating %s documentation...\n", docsFormat)

	err = docs.GenerateDocs(g, docsOpts)
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Report results
	var outputDesc string
	switch format {
	case docs.DocsSingleFile:
		outputDesc = filepath.Join(docsOutputDir, "README.md")
	case docs.DocsMultiFile:
		outputDesc = fmt.Sprintf("%s (index.md + %d module files)", docsOutputDir, g.Statistics.TotalModules)
	case docs.DocsDirectory:
		outputDesc = fmt.Sprintf("%s (organized by directory structure)", docsOutputDir)
	}

	fmt.Printf("\n✓ Documentation generated successfully\n")
	fmt.Printf("  Output: %s\n", outputDesc)
	fmt.Printf("  Format: %s\n", docsFormat)
	fmt.Printf("  Modules: %d\n", g.Statistics.TotalModules)

	if len(docsLayers) > 0 {
		fmt.Printf("  Filtered by layers: %v\n", docsLayers)
	}
	if len(docsTags) > 0 {
		fmt.Printf("  Filtered by tags: %v\n", docsTags)
	}

	return nil
}
