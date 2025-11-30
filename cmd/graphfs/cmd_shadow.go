/*
# Module: cmd/graphfs/cmd_shadow.go
Shadow command implementation.

Manages the shadow file system for storing graph metadata separately
from the actual codebase.

## Linked Modules
- [root](./root.go) - Root command
- [../../pkg/shadow](../../pkg/shadow/shadow.go) - Shadow file system

## Tags
cli, command, shadow, metadata

## Exports
shadowCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_shadow.go> a code:Module ;

	code:name "cmd/graphfs/cmd_shadow.go" ;
	code:description "Shadow command implementation" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <../../pkg/shadow/shadow.go> ;
	code:exports <#shadowCmd> ;
	code:tags "cli", "command", "shadow", "metadata" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/justin4957/graphfs/pkg/cli"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/justin4957/graphfs/pkg/shadow"
	"github.com/spf13/cobra"
)

var (
	// Shadow build flags
	shadowMerge     bool
	shadowForce     bool
	shadowNoTriples bool
	shadowSkipClean bool
	shadowWorkers   int

	// Shadow query flags
	shadowLanguage string
	shadowLayer    string
	shadowTags     []string
	shadowConcepts []string
	shadowOutput   string

	// Shadow annotate flags
	shadowKey    string
	shadowValue  string
	shadowAuthor string
)

// shadowCmd represents the shadow command
var shadowCmd = &cobra.Command{
	Use:   "shadow",
	Short: "Manage shadow file system for graph metadata",
	Long: `Manage the shadow file system for storing graph metadata separately from source code.

The shadow file system creates a parallel directory structure (.graphfs/shadow/) that
mirrors your codebase. Each source file can have a corresponding shadow file containing:

  - Semantic metadata (RDF triples)
  - Module information (name, description, language, layer, tags)
  - Relationships (dependencies, dependents, exports, calls)
  - Manual annotations (custom key-value pairs)
  - Concepts and semantic tags

This allows you to build a rich semantic graph of your codebase without modifying
the actual source files.

Subcommands:
  init      Initialize shadow file system
  build     Generate shadow entries from LinkedDoc metadata
  sync      Full sync (build + clean orphaned entries)
  query     Query shadow entries by various criteria
  show      Show shadow entry for a specific file
  annotate  Add manual annotations to shadow entries
  stats     Show shadow file system statistics
  clean     Remove orphaned shadow entries

Examples:
  graphfs shadow init                           # Initialize shadow file system
  graphfs shadow build                          # Build shadow entries from codebase
  graphfs shadow build --force                  # Force rebuild all entries
  graphfs shadow sync                           # Build and clean in one operation
  graphfs shadow query --language go            # Find all Go files
  graphfs shadow query --tags api,service       # Find files with specific tags
  graphfs shadow show pkg/shadow/shadow.go      # Show shadow entry for a file
  graphfs shadow annotate pkg/api.go --key "reviewed" --value "true"
  graphfs shadow stats                          # Show statistics`,
}

// shadowInitCmd initializes the shadow file system
var shadowInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize shadow file system",
	Long: `Initialize the shadow file system structure.

Creates the .graphfs/shadow/ directory and index file. This is automatically
called by other shadow commands if needed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowInit,
}

// shadowBuildCmd builds shadow entries from source files
var shadowBuildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Generate shadow entries from LinkedDoc metadata",
	Long: `Scan the codebase and generate shadow entries from LinkedDoc metadata.

By default, this command:
  - Scans all files with LinkedDoc headers
  - Creates shadow entries with module info and relationships
  - Merges with existing entries (preserving manual annotations)
  - Skips files that haven't changed since last build

Flags:
  --force         Force rebuild all entries (overwrites existing)
  --no-merge      Don't merge with existing entries
  --no-triples    Don't include raw RDF triples in entries
  --workers N     Number of parallel workers (default: NumCPU)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowBuild,
}

// shadowSyncCmd performs a full sync operation
var shadowSyncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Full sync (build + clean)",
	Long: `Perform a full synchronization of the shadow file system.

This command:
  1. Builds/updates shadow entries from source files
  2. Removes shadow entries for files that no longer exist

Use --skip-clean to only build without cleaning orphaned entries.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowSync,
}

// shadowQueryCmd queries shadow entries
var shadowQueryCmd = &cobra.Command{
	Use:   "query [path]",
	Short: "Query shadow entries by various criteria",
	Long: `Query shadow entries using various filters.

Filters:
  --language      Filter by programming language
  --layer         Filter by architectural layer
  --tags          Filter by tags (comma-separated, matches all)
  --concepts      Filter by concepts (comma-separated, matches all)

Output formats:
  --output json   Output as JSON
  --output table  Output as table (default)
  --output paths  Output only file paths`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowQuery,
}

// shadowShowCmd shows a shadow entry for a specific file
var shadowShowCmd = &cobra.Command{
	Use:   "show <file>",
	Short: "Show shadow entry for a specific file",
	Long: `Display the shadow entry for a specific source file.

Shows all metadata including:
  - Module information
  - Relationships (dependencies, dependents)
  - RDF triples
  - Manual annotations
  - Custom properties`,
	Args: cobra.ExactArgs(1),
	RunE: runShadowShow,
}

// shadowAnnotateCmd adds manual annotations
var shadowAnnotateCmd = &cobra.Command{
	Use:   "annotate <file>",
	Short: "Add manual annotation to shadow entry",
	Long: `Add or update a manual annotation on a shadow entry.

Annotations are key-value pairs that persist across rebuilds.
Use this to add custom metadata like code review status, ownership, etc.

Example:
  graphfs shadow annotate pkg/api.go --key "reviewed" --value "true" --author "john"
  graphfs shadow annotate pkg/api.go --key "owner" --value "team-backend"`,
	Args: cobra.ExactArgs(1),
	RunE: runShadowAnnotate,
}

// shadowStatsCmd shows shadow file system statistics
var shadowStatsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Show shadow file system statistics",
	Long: `Display statistics about the shadow file system.

Shows:
  - Total entries and triples
  - Entries by source (auto/manual/mixed)
  - Entries by language and layer
  - Tag and concept distribution`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowStats,
}

// shadowCleanCmd removes orphaned shadow entries
var shadowCleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Remove orphaned shadow entries",
	Long: `Remove shadow entries for source files that no longer exist.

This helps keep the shadow file system in sync with the actual codebase.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowClean,
}

// shadowRebuildIndexCmd rebuilds the index
var shadowRebuildIndexCmd = &cobra.Command{
	Use:   "rebuild-index [path]",
	Short: "Rebuild the shadow index",
	Long: `Rebuild the shadow index by scanning all shadow files.

Use this if the index becomes corrupted or out of sync.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShadowRebuildIndex,
}

func init() {
	// Add subcommands
	shadowCmd.AddCommand(shadowInitCmd)
	shadowCmd.AddCommand(shadowBuildCmd)
	shadowCmd.AddCommand(shadowSyncCmd)
	shadowCmd.AddCommand(shadowQueryCmd)
	shadowCmd.AddCommand(shadowShowCmd)
	shadowCmd.AddCommand(shadowAnnotateCmd)
	shadowCmd.AddCommand(shadowStatsCmd)
	shadowCmd.AddCommand(shadowCleanCmd)
	shadowCmd.AddCommand(shadowRebuildIndexCmd)

	// Build flags
	shadowBuildCmd.Flags().BoolVar(&shadowMerge, "merge", true, "Merge with existing entries")
	shadowBuildCmd.Flags().BoolVar(&shadowForce, "force", false, "Force rebuild all entries")
	shadowBuildCmd.Flags().BoolVar(&shadowNoTriples, "no-triples", false, "Don't include raw RDF triples")
	shadowBuildCmd.Flags().IntVarP(&shadowWorkers, "workers", "w", 0, "Number of parallel workers")

	// Sync flags
	shadowSyncCmd.Flags().BoolVar(&shadowMerge, "merge", true, "Merge with existing entries")
	shadowSyncCmd.Flags().BoolVar(&shadowForce, "force", false, "Force rebuild all entries")
	shadowSyncCmd.Flags().BoolVar(&shadowNoTriples, "no-triples", false, "Don't include raw RDF triples")
	shadowSyncCmd.Flags().BoolVar(&shadowSkipClean, "skip-clean", false, "Skip cleaning orphaned entries")
	shadowSyncCmd.Flags().IntVarP(&shadowWorkers, "workers", "w", 0, "Number of parallel workers")

	// Query flags
	shadowQueryCmd.Flags().StringVar(&shadowLanguage, "language", "", "Filter by language")
	shadowQueryCmd.Flags().StringVar(&shadowLayer, "layer", "", "Filter by layer")
	shadowQueryCmd.Flags().StringSliceVar(&shadowTags, "tags", nil, "Filter by tags")
	shadowQueryCmd.Flags().StringSliceVar(&shadowConcepts, "concepts", nil, "Filter by concepts")
	shadowQueryCmd.Flags().StringVarP(&shadowOutput, "output", "o", "table", "Output format (table, json, paths)")

	// Show flags
	shadowShowCmd.Flags().StringVarP(&shadowOutput, "output", "o", "table", "Output format (table, json)")

	// Annotate flags
	shadowAnnotateCmd.Flags().StringVar(&shadowKey, "key", "", "Annotation key (required)")
	shadowAnnotateCmd.Flags().StringVar(&shadowValue, "value", "", "Annotation value (required)")
	shadowAnnotateCmd.Flags().StringVar(&shadowAuthor, "author", "", "Annotation author")
	_ = shadowAnnotateCmd.MarkFlagRequired("key")
	_ = shadowAnnotateCmd.MarkFlagRequired("value")

	// Register shadow command with root
	rootCmd.AddCommand(shadowCmd)
}

func runShadowInit(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Initialize
	if err := shadowFS.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize shadow file system: %w", err)
	}

	out.Success("Shadow file system initialized at %s", shadowFS.ShadowPath())
	return nil
}

func runShadowBuild(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	out.Info("Building shadow file system...")

	// Create and initialize shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize shadow file system: %w", err)
	}

	// Create builder
	builder := shadow.NewBuilder(shadowFS)

	// Configure build options
	opts := shadow.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
			IgnoreFiles: []string{".gitignore", ".graphfsignore"},
			Concurrent:  true,
			Workers:     shadowWorkers,
		},
		MergeExisting:  shadowMerge && !shadowForce,
		ForceOverwrite: shadowForce,
		ReportProgress: verbose,
		Workers:        shadowWorkers,
		IncludeTriples: !shadowNoTriples,
		SkipUnchanged:  !shadowForce,
	}

	// Run build
	result, err := builder.Build(opts)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Print results
	out.Println("")
	out.Success("Shadow build completed successfully")
	out.KeyValue("Total Files", result.TotalFiles)
	out.KeyValue("Processed", result.ProcessedFiles)
	out.KeyValue("New Entries", result.NewEntries)
	out.KeyValue("Updated", result.UpdatedEntries)
	out.KeyValue("Merged", result.MergedEntries)
	out.KeyValue("Skipped", result.SkippedFiles)

	if len(result.Errors) > 0 {
		out.Println("")
		out.Warning("Encountered %d errors:", len(result.Errors))
		for i, buildErr := range result.Errors {
			if i >= 5 {
				out.Warning("  ... and %d more errors", len(result.Errors)-5)
				break
			}
			out.Warning("  - %s: %s", buildErr.Path, buildErr.Message)
		}
	}

	duration := time.Since(startTime)
	out.Println("")
	out.Success("Build completed in %v", duration.Round(time.Millisecond))

	return nil
}

func runShadowSync(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	out.Info("Syncing shadow file system...")

	// Create and initialize shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize shadow file system: %w", err)
	}

	// Create builder
	builder := shadow.NewBuilder(shadowFS)

	// Configure build options
	opts := shadow.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
			IgnoreFiles: []string{".gitignore", ".graphfsignore"},
			Concurrent:  true,
			Workers:     shadowWorkers,
		},
		MergeExisting:  shadowMerge && !shadowForce,
		ForceOverwrite: shadowForce,
		ReportProgress: verbose,
		Workers:        shadowWorkers,
		IncludeTriples: !shadowNoTriples,
		SkipUnchanged:  !shadowForce,
	}

	if shadowSkipClean {
		// Build only
		result, err := builder.Build(opts)
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		out.Println("")
		out.Success("Build completed: %d new, %d updated, %d merged",
			result.NewEntries, result.UpdatedEntries, result.MergedEntries)
	} else {
		// Full sync
		result, err := builder.Sync(opts)
		if err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		out.Println("")
		out.Success("Sync completed successfully")
		out.KeyValue("New Entries", result.BuildResult.NewEntries)
		out.KeyValue("Updated", result.BuildResult.UpdatedEntries)
		out.KeyValue("Merged", result.BuildResult.MergedEntries)
		out.KeyValue("Cleaned", result.CleanResult.RemovedEntries)
	}

	duration := time.Since(startTime)
	out.Println("")
	out.Success("Sync completed in %v", duration.Round(time.Millisecond))

	return nil
}

func runShadowQuery(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Load index
	if err := shadowFS.LoadIndex(); err != nil {
		// Index might not exist, try to rebuild
		if err := shadowFS.RebuildIndex(); err != nil {
			return fmt.Errorf("failed to load or rebuild index: %w", err)
		}
	}

	// Build search query
	query := shadow.SearchQuery{
		Language: shadowLanguage,
		Layer:    shadowLayer,
		Tags:     shadowTags,
		Concepts: shadowConcepts,
	}

	// Execute search
	paths := shadowFS.Index().Search(query)

	if len(paths) == 0 {
		out.Info("No matching shadow entries found")
		return nil
	}

	// Output results
	switch shadowOutput {
	case "json":
		// Get full entries
		var entries []*shadow.Entry
		for _, path := range paths {
			entry, err := shadowFS.Get(filepath.Join(absPath, path))
			if err == nil {
				entries = append(entries, entry)
			}
		}
		data, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(data))

	case "paths":
		for _, path := range paths {
			fmt.Println(path)
		}

	default: // table
		out.Header(fmt.Sprintf("Shadow Entries (%d results)", len(paths)))
		out.Println("")

		headers := []string{"Path", "Language", "Layer", "Tags"}
		var rows [][]string

		for _, path := range paths {
			indexEntry, ok := shadowFS.Index().Get(path)
			if !ok {
				continue
			}

			tags := ""
			if len(indexEntry.Tags) > 0 {
				tags = strings.Join(indexEntry.Tags, ", ")
			}

			rows = append(rows, []string{
				path,
				indexEntry.Language,
				indexEntry.Layer,
				tags,
			})
		}

		out.Table(headers, rows)
	}

	return nil
}

func runShadowShow(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	filePath := args[0]

	// Find project root
	absPath, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Get shadow entry
	sourceFile := filePath
	if !filepath.IsAbs(filePath) {
		sourceFile = filepath.Join(absPath, filePath)
	}

	entry, err := shadowFS.Get(sourceFile)
	if err != nil {
		return fmt.Errorf("shadow entry not found for %s: %w", filePath, err)
	}

	// Output
	if shadowOutput == "json" {
		data, _ := json.MarshalIndent(entry, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Table format
	out.Header(fmt.Sprintf("Shadow Entry: %s", entry.SourcePath))
	out.Println("")

	out.KeyValue("Version", entry.Version)
	out.KeyValue("Source", string(entry.Source))
	out.KeyValue("Created", entry.CreatedAt.Format(time.RFC3339))
	out.KeyValue("Updated", entry.UpdatedAt.Format(time.RFC3339))

	if entry.SourceHash != "" {
		out.KeyValue("Source Hash", entry.SourceHash[:16]+"...")
	}

	if entry.Module != nil {
		out.Println("")
		out.Header("Module Information")
		if entry.Module.URI != "" {
			out.KeyValue("URI", entry.Module.URI)
		}
		if entry.Module.Name != "" {
			out.KeyValue("Name", entry.Module.Name)
		}
		if entry.Module.Description != "" {
			out.KeyValue("Description", entry.Module.Description)
		}
		if entry.Module.Language != "" {
			out.KeyValue("Language", entry.Module.Language)
		}
		if entry.Module.Layer != "" {
			out.KeyValue("Layer", entry.Module.Layer)
		}
		if len(entry.Module.Tags) > 0 {
			out.KeyValue("Tags", strings.Join(entry.Module.Tags, ", "))
		}
	}

	if len(entry.Dependencies) > 0 {
		out.Println("")
		out.Header(fmt.Sprintf("Dependencies (%d)", len(entry.Dependencies)))
		for _, dep := range entry.Dependencies {
			out.Println("  - %s -> %s", dep.Type, dep.Target)
		}
	}

	if len(entry.Exports) > 0 {
		out.Println("")
		out.Header(fmt.Sprintf("Exports (%d)", len(entry.Exports)))
		for _, exp := range entry.Exports {
			out.Println("  - %s", exp)
		}
	}

	if len(entry.Triples) > 0 {
		out.Println("")
		out.Header(fmt.Sprintf("RDF Triples (%d)", len(entry.Triples)))
		for i, t := range entry.Triples {
			if i >= 10 {
				out.Println("  ... and %d more triples", len(entry.Triples)-10)
				break
			}
			out.Println("  - <%s> <%s> <%s>", t.Subject, t.Predicate, t.Object)
		}
	}

	if len(entry.Annotations) > 0 {
		out.Println("")
		out.Header(fmt.Sprintf("Annotations (%d)", len(entry.Annotations)))
		for _, a := range entry.Annotations {
			out.Println("  - %s: %v", a.Key, a.Value)
			if a.Author != "" {
				out.Println("    (by %s)", a.Author)
			}
		}
	}

	if len(entry.Concepts) > 0 {
		out.Println("")
		out.Header("Concepts")
		out.Println("  %s", strings.Join(entry.Concepts, ", "))
	}

	return nil
}

func runShadowAnnotate(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	filePath := args[0]

	// Find project root
	absPath, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create shadow file system
	config := shadow.DefaultConfig()
	config.PreserveManual = true
	shadowFS, err := shadow.NewShadowFS(absPath, config)
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize shadow file system: %w", err)
	}

	// Get or create shadow entry
	sourceFile := filePath
	if !filepath.IsAbs(filePath) {
		sourceFile = filepath.Join(absPath, filePath)
	}

	entry, err := shadowFS.Get(sourceFile)
	if err != nil {
		// Create new manual entry
		relPath, _ := filepath.Rel(absPath, sourceFile)
		entry = shadow.NewManualEntry(relPath)
	}

	// Add annotation
	entry.AddAnnotation(shadowKey, shadowValue, shadowAuthor)

	// Update source to mixed if it was auto
	if entry.Source == shadow.SourceAuto {
		entry.Source = shadow.SourceMixed
	}

	// Save entry
	if err := shadowFS.Set(sourceFile, entry); err != nil {
		return fmt.Errorf("failed to save shadow entry: %w", err)
	}

	out.Success("Added annotation '%s' = '%s' to %s", shadowKey, shadowValue, filePath)
	return nil
}

func runShadowStats(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Load index
	if err := shadowFS.LoadIndex(); err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Get statistics
	stats := shadowFS.Index().Statistics()

	out.Header("Shadow File System Statistics")
	out.Println("")

	out.KeyValue("Total Entries", stats.TotalEntries)
	out.KeyValue("Total Triples", stats.TotalTriples)
	out.Println("")

	out.Header("Entries by Source")
	out.KeyValue("Auto-generated", stats.AutoEntries)
	out.KeyValue("Manual", stats.ManualEntries)
	out.KeyValue("Mixed", stats.MixedEntries)

	if len(stats.LanguageCount) > 0 {
		out.Println("")
		out.Header("Entries by Language")
		headers := []string{"Language", "Count"}
		var rows [][]string
		for lang, count := range stats.LanguageCount {
			rows = append(rows, []string{lang, fmt.Sprintf("%d", count)})
		}
		out.Table(headers, rows)
	}

	if len(stats.LayerCount) > 0 {
		out.Println("")
		out.Header("Entries by Layer")
		headers := []string{"Layer", "Count"}
		var rows [][]string
		for layer, count := range stats.LayerCount {
			rows = append(rows, []string{layer, fmt.Sprintf("%d", count)})
		}
		out.Table(headers, rows)
	}

	if len(stats.TagCount) > 0 {
		out.Println("")
		out.Header("Top Tags")
		headers := []string{"Tag", "Count"}
		var rows [][]string
		for tag, count := range stats.TagCount {
			rows = append(rows, []string{tag, fmt.Sprintf("%d", count)})
		}
		out.Table(headers, rows)
	}

	return nil
}

func runShadowClean(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	out.Info("Cleaning orphaned shadow entries...")

	// Create and initialize shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Create builder
	builder := shadow.NewBuilder(shadowFS)

	// Configure build options
	opts := shadow.BuildOptions{
		ReportProgress: verbose,
	}

	// Run clean
	result, err := builder.Clean(opts)
	if err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	out.Println("")
	out.Success("Clean completed")
	out.KeyValue("Total Entries", result.TotalEntries)
	out.KeyValue("Orphaned Entries", len(result.OrphanedEntries))
	out.KeyValue("Removed", result.RemovedEntries)

	if verbose && len(result.OrphanedEntries) > 0 {
		out.Println("")
		out.Header("Removed Entries")
		for _, path := range result.OrphanedEntries {
			out.Println("  - %s", path)
		}
	}

	return nil
}

func runShadowRebuildIndex(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	out.Info("Rebuilding shadow index...")

	// Create shadow file system
	shadowFS, err := shadow.NewShadowFS(absPath, shadow.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create shadow file system: %w", err)
	}

	// Rebuild index
	if err := shadowFS.RebuildIndex(); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	stats := shadowFS.Index().Statistics()

	out.Success("Index rebuilt successfully")
	out.KeyValue("Indexed Entries", stats.TotalEntries)

	return nil
}
