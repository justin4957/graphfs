/*
# Module: cmd/graphfs/completions.go
Dynamic shell completion functions.

Provides context-aware completion for module paths, layers, tags, and output formats
by loading the knowledge graph from cache.

## Linked Modules
- [root](./root.go) - Root command
- [../../pkg/graph](../../pkg/graph/graph.go) - Graph data structure

## Tags
cli, completion, autocomplete

## Exports
modulePathCompletion, layerCompletion, tagCompletion, outputFormatCompletion

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#completions.go> a code:Module ;
    code:name "cmd/graphfs/completions.go" ;
    code:description "Dynamic shell completion functions" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <./root.go>, <../../pkg/graph/graph.go> ;
    code:exports <#modulePathCompletion>, <#layerCompletion>, <#tagCompletion>, <#outputFormatCompletion> ;
    code:tags "cli", "completion", "autocomplete" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
)

// loadGraphForCompletion loads the graph for completion purposes
// It tries to load quickly without validation
func loadGraphForCompletion() (*graph.Graph, error) {
	// Get current directory
	rootPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Load configuration
	configPath := filepath.Join(rootPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		// If no config, use defaults
		config = &Config{
			Scan: ScanConfig{
				Include:     []string{},
				Exclude:     []string{},
				MaxFileSize: 1024 * 1024, // 1MB
			},
		}
	}

	// Create scanner with options
	scanOpts := scanner.ScanOptions{
		IncludePatterns: config.Scan.Include,
		ExcludePatterns: config.Scan.Exclude,
		MaxFileSize:     config.Scan.MaxFileSize,
		UseDefaults:     true,
		IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
		Concurrent:      true,
	}

	// Build graph without validation for speed
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanOpts,
		Validate:    false, // Skip validation for completion performance
	}

	return builder.Build(rootPath, buildOpts)
}

// modulePathCompletion provides completion for module paths
func modulePathCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	g, err := loadGraphForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for path := range g.Modules {
		if strings.HasPrefix(path, toComplete) {
			completions = append(completions, path)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// layerCompletion provides completion for layer values
func layerCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	g, err := loadGraphForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract unique layers from modules
	layerSet := make(map[string]bool)
	for _, module := range g.Modules {
		if module.Layer != "" {
			layerSet[module.Layer] = true
		}
	}

	var completions []string
	for layer := range layerSet {
		if strings.HasPrefix(layer, toComplete) {
			completions = append(completions, layer)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// tagCompletion provides completion for tag values
func tagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	g, err := loadGraphForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract unique tags from modules
	tagSet := make(map[string]bool)
	for _, module := range g.Modules {
		for _, tag := range module.Tags {
			tagSet[tag] = true
		}
	}

	var completions []string
	for tag := range tagSet {
		if strings.HasPrefix(tag, toComplete) {
			completions = append(completions, tag)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// outputFormatCompletion provides completion for output format flags
func outputFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "json", "csv", "dot", "mermaid", "turtle"}

	var completions []string
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			completions = append(completions, format)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// queryFormatCompletion provides completion for query output format flags
func queryFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "json", "csv"}

	var completions []string
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			completions = append(completions, format)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// categoryCompletion provides completion for example query categories
func categoryCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	categories := []string{"dependencies", "security", "analysis", "layers", "impact", "documentation"}

	var completions []string
	for _, cat := range categories {
		if strings.HasPrefix(cat, toComplete) {
			completions = append(completions, cat)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// shellCompletion provides completion for shell types
func shellCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	shells := []string{"bash", "zsh", "fish"}

	var completions []string
	for _, shell := range shells {
		if strings.HasPrefix(shell, toComplete) {
			completions = append(completions, shell)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// registerCompletions registers all completion functions for commands
func registerCompletions() error {
	// Register completion for impact command (module path)
	if err := impactCmd.MarkFlagFilename("viz"); err != nil {
		return fmt.Errorf("failed to mark impact viz flag: %w", err)
	}
	impactCmd.ValidArgsFunction = modulePathCompletion

	// Register completion for viz command
	if err := vizCmd.RegisterFlagCompletionFunc("format", outputFormatCompletion); err != nil {
		return fmt.Errorf("failed to register viz format completion: %w", err)
	}
	if err := vizCmd.RegisterFlagCompletionFunc("module", modulePathCompletion); err != nil {
		return fmt.Errorf("failed to register viz module completion: %w", err)
	}
	if err := vizCmd.RegisterFlagCompletionFunc("layer", layerCompletion); err != nil {
		return fmt.Errorf("failed to register viz layer completion: %w", err)
	}
	if err := vizCmd.MarkFlagFilename("output"); err != nil {
		return fmt.Errorf("failed to mark viz output flag: %w", err)
	}

	// Register completion for query command
	if err := queryCmd.RegisterFlagCompletionFunc("format", queryFormatCompletion); err != nil {
		return fmt.Errorf("failed to register query format completion: %w", err)
	}
	if err := queryCmd.MarkFlagFilename("file"); err != nil {
		return fmt.Errorf("failed to mark query file flag: %w", err)
	}
	if err := queryCmd.MarkFlagFilename("output"); err != nil {
		return fmt.Errorf("failed to mark query output flag: %w", err)
	}

	// Register completion for dead-code command
	if err := deadCodeCmd.MarkFlagFilename("script"); err != nil {
		return fmt.Errorf("failed to mark dead-code script flag: %w", err)
	}
	if err := deadCodeCmd.MarkFlagDirname("target"); err != nil {
		return fmt.Errorf("failed to mark dead-code target flag: %w", err)
	}

	// Register completion for security command
	if err := securityCmd.MarkFlagDirname("target"); err != nil {
		return fmt.Errorf("failed to mark security target flag: %w", err)
	}
	if err := securityCmd.MarkFlagFilename("viz"); err != nil {
		return fmt.Errorf("failed to mark security viz flag: %w", err)
	}

	// Register completion for docs command
	if err := docsCmd.RegisterFlagCompletionFunc("layer", layerCompletion); err != nil {
		return fmt.Errorf("failed to register docs layer completion: %w", err)
	}
	if err := docsCmd.RegisterFlagCompletionFunc("tag", tagCompletion); err != nil {
		return fmt.Errorf("failed to register docs tag completion: %w", err)
	}
	if err := docsCmd.MarkFlagDirname("output"); err != nil {
		return fmt.Errorf("failed to mark docs output flag: %w", err)
	}
	if err := docsCmd.MarkFlagFilename("template"); err != nil {
		return fmt.Errorf("failed to mark docs template flag: %w", err)
	}

	// Register completion for examples command
	if err := examplesListCmd.RegisterFlagCompletionFunc("category", categoryCompletion); err != nil {
		return fmt.Errorf("failed to register examples list category completion: %w", err)
	}

	// Register completion for completion command itself
	completionCmd.ValidArgsFunction = shellCompletion

	return nil
}
