/*
# Module: pkg/docs/markdown.go
Markdown documentation generator from LinkedDoc metadata and code analysis.

Generates comprehensive module documentation including dependencies,
exports, usage examples, and relationships.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [../analysis](../analysis/impact.go) - Impact analysis

## Tags
documentation, markdown, generator

## Exports
GenerateDocs, GenerateModuleDocs, DocsOptions, DocsFormat

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#markdown.go> a code:Module ;
    code:name "pkg/docs/markdown.go" ;
    code:description "Markdown documentation generator" ;
    code:language "go" ;
    code:layer "documentation" ;
    code:linksTo <../graph/graph.go>, <../analysis/impact.go> ;
    code:exports <#GenerateDocs>, <#GenerateModuleDocs>, <#DocsOptions> ;
    code:tags "documentation", "markdown", "generator" .
<!-- End LinkedDoc RDF -->
*/

package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/justin4957/graphfs/pkg/graph"
)

// DocsFormat represents the documentation output format
type DocsFormat string

const (
	DocsSingleFile DocsFormat = "single"    // Single README.md file
	DocsMultiFile  DocsFormat = "multi"     // One file per module
	DocsDirectory  DocsFormat = "directory" // Organized by directory structure
)

// DocsOptions configures documentation generation
type DocsOptions struct {
	OutputDir     string            // Output directory
	Format        DocsFormat        // Output format
	Template      string            // Custom template path
	IncludeLayers []string          // Include only these layers
	IncludeTags   []string          // Include only modules with these tags
	Depth         int               // Max dependency depth (0 = unlimited)
	IncludeGraph  bool              // Include dependency graphs
	FrontMatter   map[string]string // Frontmatter for static site generators
	Title         string            // Documentation title
	ProjectName   string            // Project name
}

// ModuleDoc represents documentation for a single module
type ModuleDoc struct {
	Module       *graph.Module
	Dependencies []*graph.Module
	Dependents   []*graph.Module
	RelatedPath  string // Relative path for links
}

// DocsGenerator generates markdown documentation
type DocsGenerator struct {
	graph   *graph.Graph
	options DocsOptions
	modules []*ModuleDoc
}

// NewDocsGenerator creates a new documentation generator
func NewDocsGenerator(g *graph.Graph, opts DocsOptions) *DocsGenerator {
	if opts.Format == "" {
		opts.Format = DocsMultiFile
	}
	if opts.ProjectName == "" {
		opts.ProjectName = filepath.Base(g.Root)
	}
	if opts.Title == "" {
		opts.Title = fmt.Sprintf("%s Documentation", opts.ProjectName)
	}

	return &DocsGenerator{
		graph:   g,
		options: opts,
		modules: make([]*ModuleDoc, 0),
	}
}

// GenerateDocs generates documentation for the entire graph
func GenerateDocs(g *graph.Graph, opts DocsOptions) error {
	gen := NewDocsGenerator(g, opts)
	return gen.Generate()
}

// Generate generates the documentation
func (dg *DocsGenerator) Generate() error {
	// Prepare module docs
	if err := dg.prepareModuleDocs(); err != nil {
		return err
	}

	// Create output directory
	if err := os.MkdirAll(dg.options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate based on format
	switch dg.options.Format {
	case DocsSingleFile:
		return dg.generateSingleFile()
	case DocsMultiFile:
		return dg.generateMultiFile()
	case DocsDirectory:
		return dg.generateDirectory()
	default:
		return fmt.Errorf("unknown format: %s", dg.options.Format)
	}
}

// prepareModuleDocs prepares module documentation structures
func (dg *DocsGenerator) prepareModuleDocs() error {
	for _, module := range dg.graph.Modules {
		// Apply filters
		if !dg.shouldIncludeModule(module) {
			continue
		}

		moduleDoc := &ModuleDoc{
			Module:       module,
			Dependencies: make([]*graph.Module, 0),
			Dependents:   make([]*graph.Module, 0),
		}

		// Get dependencies
		for _, depPath := range module.Dependencies {
			dep := dg.graph.GetModule(depPath)
			if dep != nil {
				moduleDoc.Dependencies = append(moduleDoc.Dependencies, dep)
			}
		}

		// Get dependents
		for _, other := range dg.graph.Modules {
			for _, depPath := range other.Dependencies {
				if depPath == module.Path {
					moduleDoc.Dependents = append(moduleDoc.Dependents, other)
					break
				}
			}
		}

		dg.modules = append(dg.modules, moduleDoc)
	}

	// Sort modules by path
	sort.Slice(dg.modules, func(i, j int) bool {
		return dg.modules[i].Module.Path < dg.modules[j].Module.Path
	})

	return nil
}

// shouldIncludeModule checks if a module should be included
func (dg *DocsGenerator) shouldIncludeModule(module *graph.Module) bool {
	// Filter by layer
	if len(dg.options.IncludeLayers) > 0 {
		found := false
		for _, layer := range dg.options.IncludeLayers {
			if strings.EqualFold(module.Layer, layer) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by tag
	if len(dg.options.IncludeTags) > 0 {
		hasTag := false
		for _, filterTag := range dg.options.IncludeTags {
			for _, moduleTag := range module.Tags {
				if strings.EqualFold(moduleTag, filterTag) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// generateSingleFile generates a single README.md file
func (dg *DocsGenerator) generateSingleFile() error {
	var content strings.Builder

	// Write header
	dg.writeFrontMatter(&content)
	dg.writeHeader(&content, dg.options.Title, 1)
	content.WriteString("\n")

	// Write overview
	dg.writeOverview(&content)
	content.WriteString("\n")

	// Write table of contents
	dg.writeTableOfContents(&content)
	content.WriteString("\n")

	// Write module documentation
	for _, moduleDoc := range dg.modules {
		dg.writeModuleSection(&content, moduleDoc, 2)
		content.WriteString("\n")
	}

	// Write footer
	dg.writeFooter(&content)

	// Write to file
	outputPath := filepath.Join(dg.options.OutputDir, "README.md")
	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// generateMultiFile generates one file per module
func (dg *DocsGenerator) generateMultiFile() error {
	// Generate index file
	if err := dg.generateIndexFile(); err != nil {
		return err
	}

	// Generate individual module files
	for _, moduleDoc := range dg.modules {
		if err := dg.generateModuleFile(moduleDoc); err != nil {
			return err
		}
	}

	return nil
}

// generateDirectory generates files organized by directory structure
func (dg *DocsGenerator) generateDirectory() error {
	// Generate index file
	if err := dg.generateIndexFile(); err != nil {
		return err
	}

	// Group modules by directory
	dirModules := make(map[string][]*ModuleDoc)
	for _, moduleDoc := range dg.modules {
		dir := filepath.Dir(moduleDoc.Module.Path)
		if dir == "." {
			dir = "root"
		}
		dirModules[dir] = append(dirModules[dir], moduleDoc)
	}

	// Generate files for each directory
	for dir, modules := range dirModules {
		if err := dg.generateDirectoryFile(dir, modules); err != nil {
			return err
		}
	}

	return nil
}

// generateIndexFile generates the index file
func (dg *DocsGenerator) generateIndexFile() error {
	var content strings.Builder

	dg.writeFrontMatter(&content)
	dg.writeHeader(&content, dg.options.Title, 1)
	content.WriteString("\n")

	dg.writeOverview(&content)
	content.WriteString("\n")

	// Write module index
	dg.writeHeader(&content, "Modules", 2)
	content.WriteString("\n")

	// Group by layer
	layerModules := make(map[string][]*ModuleDoc)
	for _, moduleDoc := range dg.modules {
		layer := moduleDoc.Module.Layer
		if layer == "" {
			layer = "unknown"
		}
		layerModules[layer] = append(layerModules[layer], moduleDoc)
	}

	// Write layers
	layers := make([]string, 0, len(layerModules))
	for layer := range layerModules {
		layers = append(layers, layer)
	}
	sort.Strings(layers)

	for _, layer := range layers {
		modules := layerModules[layer]
		dg.writeHeader(&content, fmt.Sprintf("Layer: %s", layer), 3)
		content.WriteString("\n")

		for _, moduleDoc := range modules {
			linkPath := dg.getModuleLinkPath(moduleDoc.Module)
			content.WriteString(fmt.Sprintf("- [%s](%s)", moduleDoc.Module.Name, linkPath))
			if moduleDoc.Module.Description != "" {
				content.WriteString(fmt.Sprintf(" - %s", moduleDoc.Module.Description))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	dg.writeFooter(&content)

	outputPath := filepath.Join(dg.options.OutputDir, "index.md")
	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// generateModuleFile generates documentation for a single module
func (dg *DocsGenerator) generateModuleFile(moduleDoc *ModuleDoc) error {
	var content strings.Builder

	// Frontmatter
	if len(dg.options.FrontMatter) > 0 {
		content.WriteString("---\n")
		content.WriteString(fmt.Sprintf("title: %s\n", moduleDoc.Module.Name))
		for key, value := range dg.options.FrontMatter {
			content.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
		content.WriteString("---\n\n")
	}

	dg.writeModuleSection(&content, moduleDoc, 1)
	dg.writeFooter(&content)

	// Write to file
	fileName := dg.getModuleFileName(moduleDoc.Module)
	outputPath := filepath.Join(dg.options.OutputDir, fileName)

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// generateDirectoryFile generates documentation for a directory
func (dg *DocsGenerator) generateDirectoryFile(dir string, modules []*ModuleDoc) error {
	var content strings.Builder

	dg.writeFrontMatter(&content)
	dg.writeHeader(&content, fmt.Sprintf("Directory: %s", dir), 1)
	content.WriteString("\n")

	for _, moduleDoc := range modules {
		dg.writeModuleSection(&content, moduleDoc, 2)
		content.WriteString("\n")
	}

	dg.writeFooter(&content)

	// Write to file
	fileName := strings.ReplaceAll(dir, "/", "_") + ".md"
	outputPath := filepath.Join(dg.options.OutputDir, fileName)
	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

// writeModuleSection writes a module's documentation section
func (dg *DocsGenerator) writeModuleSection(w *strings.Builder, moduleDoc *ModuleDoc, level int) {
	module := moduleDoc.Module

	// Module title
	dg.writeHeader(w, fmt.Sprintf("Module: %s", module.Path), level)
	w.WriteString("\n")

	// Metadata
	if module.Layer != "" {
		w.WriteString(fmt.Sprintf("**Layer:** %s  \n", module.Layer))
	}
	if len(module.Tags) > 0 {
		w.WriteString(fmt.Sprintf("**Tags:** %s  \n", strings.Join(module.Tags, ", ")))
	}
	if module.Language != "" {
		w.WriteString(fmt.Sprintf("**Language:** %s  \n", module.Language))
	}
	w.WriteString("\n")

	// Description
	if module.Description != "" {
		dg.writeHeader(w, "Description", level+1)
		w.WriteString(module.Description)
		w.WriteString("\n\n")
	}

	// Dependencies
	if len(moduleDoc.Dependencies) > 0 {
		dg.writeHeader(w, "Dependencies", level+1)
		w.WriteString("\n")
		for _, dep := range moduleDoc.Dependencies {
			linkPath := dg.getModuleLinkPath(dep)
			w.WriteString(fmt.Sprintf("- [%s](%s)", dep.Path, linkPath))
			if dep.Description != "" {
				w.WriteString(fmt.Sprintf(" - %s", dep.Description))
			}
			w.WriteString("\n")
		}
		w.WriteString("\n")
	}

	// Dependents
	if len(moduleDoc.Dependents) > 0 {
		dg.writeHeader(w, "Dependents", level+1)
		w.WriteString("\n")
		for _, dependent := range moduleDoc.Dependents {
			linkPath := dg.getModuleLinkPath(dependent)
			w.WriteString(fmt.Sprintf("- [%s](%s)", dependent.Path, linkPath))
			if dependent.Description != "" {
				w.WriteString(fmt.Sprintf(" - %s", dependent.Description))
			}
			w.WriteString("\n")
		}
		w.WriteString("\n")
	}

	// Exports
	if len(module.Exports) > 0 {
		dg.writeHeader(w, "Exports", level+1)
		w.WriteString("\n")
		for _, export := range module.Exports {
			w.WriteString(fmt.Sprintf("- `%s`\n", export))
		}
		w.WriteString("\n")
	}
}

// writeFrontMatter writes frontmatter for static site generators
func (dg *DocsGenerator) writeFrontMatter(w *strings.Builder) {
	if len(dg.options.FrontMatter) == 0 {
		return
	}

	w.WriteString("---\n")
	for key, value := range dg.options.FrontMatter {
		w.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	w.WriteString("---\n\n")
}

// writeHeader writes a markdown header
func (dg *DocsGenerator) writeHeader(w *strings.Builder, text string, level int) {
	w.WriteString(strings.Repeat("#", level))
	w.WriteString(" ")
	w.WriteString(text)
	w.WriteString("\n")
}

// writeOverview writes the documentation overview
func (dg *DocsGenerator) writeOverview(w *strings.Builder) {
	dg.writeHeader(w, "Overview", 2)
	w.WriteString("\n")
	w.WriteString(fmt.Sprintf("This documentation covers **%d modules** in the %s project.\n\n",
		len(dg.modules), dg.options.ProjectName))

	// Statistics
	dg.writeHeader(w, "Statistics", 3)
	w.WriteString("\n")
	w.WriteString(fmt.Sprintf("- **Total Modules:** %d\n", len(dg.modules)))

	// Count by layer
	layerCounts := make(map[string]int)
	for _, moduleDoc := range dg.modules {
		layer := moduleDoc.Module.Layer
		if layer == "" {
			layer = "unknown"
		}
		layerCounts[layer]++
	}
	w.WriteString(fmt.Sprintf("- **Layers:** %d\n", len(layerCounts)))
	for layer, count := range layerCounts {
		w.WriteString(fmt.Sprintf("  - %s: %d modules\n", layer, count))
	}
	w.WriteString("\n")
}

// writeTableOfContents writes a table of contents
func (dg *DocsGenerator) writeTableOfContents(w *strings.Builder) {
	dg.writeHeader(w, "Table of Contents", 2)
	w.WriteString("\n")

	for _, moduleDoc := range dg.modules {
		anchor := dg.getModuleAnchor(moduleDoc.Module)
		w.WriteString(fmt.Sprintf("- [%s](#%s)\n", moduleDoc.Module.Path, anchor))
	}
}

// writeFooter writes the documentation footer
func (dg *DocsGenerator) writeFooter(w *strings.Builder) {
	w.WriteString("\n---\n\n")
	w.WriteString(fmt.Sprintf("*Generated by GraphFS on %s*\n",
		time.Now().Format("2006-01-02 15:04:05")))
}

// getModuleFileName returns the filename for a module
func (dg *DocsGenerator) getModuleFileName(module *graph.Module) string {
	// Convert path to filename: pkg/graph/graph.go -> pkg_graph_graph.md
	fileName := strings.ReplaceAll(module.Path, "/", "_")
	fileName = strings.ReplaceAll(fileName, ".go", "")
	return fileName + ".md"
}

// getModuleLinkPath returns the link path for a module
func (dg *DocsGenerator) getModuleLinkPath(module *graph.Module) string {
	switch dg.options.Format {
	case DocsSingleFile:
		return "#" + dg.getModuleAnchor(module)
	case DocsMultiFile:
		return dg.getModuleFileName(module)
	case DocsDirectory:
		dir := filepath.Dir(module.Path)
		if dir == "." {
			dir = "root"
		}
		fileName := strings.ReplaceAll(dir, "/", "_") + ".md"
		anchor := dg.getModuleAnchor(module)
		return fileName + "#" + anchor
	default:
		return "#" + dg.getModuleAnchor(module)
	}
}

// getModuleAnchor returns the anchor for a module
func (dg *DocsGenerator) getModuleAnchor(module *graph.Module) string {
	// Convert to markdown anchor format
	anchor := strings.ToLower(module.Path)
	anchor = strings.ReplaceAll(anchor, "/", "-")
	anchor = strings.ReplaceAll(anchor, ".", "-")
	anchor = strings.ReplaceAll(anchor, "_", "-")
	return "module-" + anchor
}

// GenerateModuleDocs generates documentation for a single module
func GenerateModuleDocs(module *graph.Module, opts DocsOptions) (string, error) {
	var content strings.Builder

	moduleDoc := &ModuleDoc{
		Module:       module,
		Dependencies: make([]*graph.Module, 0),
		Dependents:   make([]*graph.Module, 0),
	}

	gen := &DocsGenerator{
		options: opts,
	}

	gen.writeModuleSection(&content, moduleDoc, 1)
	return content.String(), nil
}
