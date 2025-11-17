/*
# Module: pkg/viz/dot.go
GraphViz DOT format generation for dependency visualization.

Generates DOT format output for various graph visualizations including
dependency graphs, impact analysis, security zones, and module relationships.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [../analysis](../analysis/impact.go) - Impact analysis
- [../analysis](../analysis/security.go) - Security analysis

## Tags
visualization, graphviz, dot, export

## Exports
GenerateDOT, VizOptions, VizType, RenderToFile

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#dot.go> a code:Module ;
    code:name "pkg/viz/dot.go" ;
    code:description "GraphViz DOT format generation for dependency visualization" ;
    code:language "go" ;
    code:layer "visualization" ;
    code:linksTo <../graph/graph.go>, <../analysis/impact.go>, <../analysis/security.go> ;
    code:exports <#GenerateDOT>, <#VizOptions>, <#VizType>, <#RenderToFile> ;
    code:tags "visualization", "graphviz", "dot", "export" .
<!-- End LinkedDoc RDF -->
*/

package viz

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
)

// VizType represents the type of visualization
type VizType string

const (
	VizDependency VizType = "dependency" // Dependency graph
	VizImpact     VizType = "impact"     // Impact analysis
	VizSecurity   VizType = "security"   // Security zones
	VizLayer      VizType = "layer"      // Layer relationships
)

// VizOptions configures visualization generation
type VizOptions struct {
	Type       VizType                    // Type of visualization
	Layout     string                     // dot, neato, fdp, circo, twopi
	ColorBy    string                     // language, layer, security, tag
	MaxDepth   int                        // Maximum depth (0 = unlimited)
	Filter     *FilterOptions             // Filtering options
	Theme      string                     // Theme name (default, security, language)
	ShowLabels bool                       // Show detailed labels
	Rankdir    string                     // Graph direction (LR, TB, RL, BT)
	Title      string                     // Graph title
	Security   *analysis.SecurityAnalysis // Security analysis results
	Impact     *analysis.ImpactResult     // Impact analysis results
}

// FilterOptions configures graph filtering
type FilterOptions struct {
	IncludePaths []string // Include only these paths (glob patterns)
	ExcludePaths []string // Exclude these paths (glob patterns)
	Layers       []string // Include only these layers
	Tags         []string // Include only modules with these tags
	MinDepth     int      // Minimum depth from root
	MaxDepth     int      // Maximum depth from root
}

// DOTGenerator generates DOT format output
type DOTGenerator struct {
	graph   *graph.Graph
	options VizOptions
	builder strings.Builder
	visited map[string]bool
	depth   map[string]int
}

// NewDOTGenerator creates a new DOT generator
func NewDOTGenerator(g *graph.Graph, opts VizOptions) *DOTGenerator {
	if opts.Rankdir == "" {
		opts.Rankdir = "LR" // Left to right by default
	}
	if opts.Layout == "" {
		opts.Layout = "dot"
	}
	if opts.Theme == "" {
		opts.Theme = "default"
	}

	return &DOTGenerator{
		graph:   g,
		options: opts,
		visited: make(map[string]bool),
		depth:   make(map[string]int),
	}
}

// GenerateDOT generates DOT format for the graph
func GenerateDOT(g *graph.Graph, opts VizOptions) (string, error) {
	gen := NewDOTGenerator(g, opts)
	return gen.Generate()
}

// Generate generates the DOT output
func (dg *DOTGenerator) Generate() (string, error) {
	dg.builder.Reset()

	// Write header
	dg.writeHeader()

	// Write graph attributes
	dg.writeGraphAttributes()

	// Write nodes and edges based on visualization type
	switch dg.options.Type {
	case VizDependency:
		dg.generateDependencyGraph()
	case VizImpact:
		dg.generateImpactGraph()
	case VizSecurity:
		dg.generateSecurityGraph()
	case VizLayer:
		dg.generateLayerGraph()
	default:
		dg.generateDependencyGraph()
	}

	// Write footer
	dg.writeFooter()

	return dg.builder.String(), nil
}

// writeHeader writes the DOT header
func (dg *DOTGenerator) writeHeader() {
	dg.builder.WriteString("digraph GraphFS {\n")
	if dg.options.Title != "" {
		dg.builder.WriteString(fmt.Sprintf("  labelloc=\"t\";\n  label=\"%s\";\n", escapeLabel(dg.options.Title)))
	}
}

// writeGraphAttributes writes graph-level attributes
func (dg *DOTGenerator) writeGraphAttributes() {
	dg.builder.WriteString(fmt.Sprintf("  rankdir=%s;\n", dg.options.Rankdir))
	dg.builder.WriteString("  node [shape=box, style=filled, fontname=\"Arial\"];\n")
	dg.builder.WriteString("  edge [fontname=\"Arial\", fontsize=10];\n")
	dg.builder.WriteString("  graph [fontname=\"Arial\"];\n\n")
}

// writeFooter writes the DOT footer
func (dg *DOTGenerator) writeFooter() {
	dg.builder.WriteString("}\n")
}

// generateDependencyGraph generates a dependency visualization
func (dg *DOTGenerator) generateDependencyGraph() {
	// Get filtered modules
	modules := dg.getFilteredModules()

	// Write nodes
	dg.builder.WriteString("  // Nodes\n")
	for _, module := range modules {
		dg.writeNode(module)
	}
	dg.builder.WriteString("\n")

	// Write edges
	dg.builder.WriteString("  // Dependencies\n")
	for _, module := range modules {
		dg.writeEdges(module)
	}
}

// generateImpactGraph generates an impact analysis visualization
func (dg *DOTGenerator) generateImpactGraph() {
	if dg.options.Impact == nil {
		dg.generateDependencyGraph()
		return
	}

	impact := dg.options.Impact

	// Get changed module
	changedModule := dg.graph.GetModule(impact.TargetModule)
	if changedModule == nil {
		dg.generateDependencyGraph()
		return
	}

	// Write changed module
	dg.builder.WriteString("  // Changed module\n")
	dg.writeNodeWithColor(changedModule, "#FF5722") // Red-orange
	dg.builder.WriteString("\n")

	// Write directly affected modules (direct dependents)
	dg.builder.WriteString("  // Directly affected\n")
	for _, depPath := range impact.DirectDependents {
		module := dg.graph.GetModule(depPath)
		if module != nil {
			dg.writeNodeWithColor(module, "#FF9800") // Orange
		}
	}
	dg.builder.WriteString("\n")

	// Write transitively affected modules
	dg.builder.WriteString("  // Transitively affected\n")
	for depPath, depth := range impact.TransitiveDependents {
		// Skip direct dependents (depth 1) as they're already shown
		if depth > 1 {
			module := dg.graph.GetModule(depPath)
			if module != nil {
				dg.writeNodeWithColor(module, "#FFC107") // Amber
			}
		}
	}
	dg.builder.WriteString("\n")

	// Collect all modules for edge drawing
	affectedModules := make(map[string]bool)
	affectedModules[impact.TargetModule] = true
	for _, path := range impact.DirectDependents {
		affectedModules[path] = true
	}
	for path := range impact.TransitiveDependents {
		affectedModules[path] = true
	}

	// Write edges for all affected modules
	dg.builder.WriteString("  // Dependencies\n")
	for path := range affectedModules {
		module := dg.graph.GetModule(path)
		if module != nil {
			dg.writeEdges(module)
		}
	}
}

// generateSecurityGraph generates a security zone visualization
func (dg *DOTGenerator) generateSecurityGraph() {
	if dg.options.Security == nil {
		dg.generateDependencyGraph()
		return
	}

	sec := dg.options.Security

	// Group modules by zone using subgraphs
	zoneOrder := []analysis.SecurityZone{
		analysis.ZonePublic,
		analysis.ZoneTrusted,
		analysis.ZoneInternal,
		analysis.ZoneAdmin,
		analysis.ZoneData,
		analysis.ZoneUnknown,
	}

	zoneColors := map[analysis.SecurityZone]string{
		analysis.ZonePublic:   "#4CAF50", // Green
		analysis.ZoneTrusted:  "#2196F3", // Blue
		analysis.ZoneInternal: "#9E9E9E", // Gray
		analysis.ZoneAdmin:    "#F44336", // Red
		analysis.ZoneData:     "#FF9800", // Orange
		analysis.ZoneUnknown:  "#E0E0E0", // Light gray
	}

	idx := 0
	for _, zone := range zoneOrder {
		modules, ok := sec.Zones[zone]
		if !ok || len(modules) == 0 {
			continue
		}

		dg.builder.WriteString(fmt.Sprintf("  // Zone: %s\n", zone))
		dg.builder.WriteString(fmt.Sprintf("  subgraph cluster_%d {\n", idx))
		// Capitalize first letter of zone name
		zoneName := string(zone)
		if len(zoneName) > 0 {
			zoneName = strings.ToUpper(zoneName[:1]) + zoneName[1:]
		}
		dg.builder.WriteString(fmt.Sprintf("    label=\"%s Zone\";\n", zoneName))
		dg.builder.WriteString("    style=filled;\n")
		dg.builder.WriteString(fmt.Sprintf("    fillcolor=\"%s30\";\n", zoneColors[zone])) // 30 = transparency
		dg.builder.WriteString(fmt.Sprintf("    color=\"%s\";\n\n", zoneColors[zone]))

		for _, mz := range modules {
			if dg.shouldIncludeModule(mz.Module) {
				dg.writeNodeWithColorInCluster(mz.Module, zoneColors[zone])
			}
		}

		dg.builder.WriteString("  }\n\n")
		idx++
	}

	// Write edges (outside clusters)
	dg.builder.WriteString("  // Dependencies\n")
	for _, modules := range sec.Zones {
		for _, mz := range modules {
			if dg.shouldIncludeModule(mz.Module) {
				dg.writeSecurityEdges(mz.Module, sec)
			}
		}
	}
}

// generateLayerGraph generates a layer-based visualization
func (dg *DOTGenerator) generateLayerGraph() {
	// Group modules by layer
	layerMap := make(map[string][]*graph.Module)
	for _, module := range dg.graph.Modules {
		if dg.shouldIncludeModule(module) {
			layer := module.Layer
			if layer == "" {
				layer = "unknown"
			}
			layerMap[layer] = append(layerMap[layer], module)
		}
	}

	// Get sorted layer names
	layers := make([]string, 0, len(layerMap))
	for layer := range layerMap {
		layers = append(layers, layer)
	}
	sort.Strings(layers)

	// Write subgraphs for each layer
	for idx, layer := range layers {
		modules := layerMap[layer]
		dg.builder.WriteString(fmt.Sprintf("  // Layer: %s\n", layer))
		dg.builder.WriteString(fmt.Sprintf("  subgraph cluster_%d {\n", idx))
		dg.builder.WriteString(fmt.Sprintf("    label=\"Layer: %s\";\n", layer))
		dg.builder.WriteString("    style=filled;\n")
		dg.builder.WriteString("    fillcolor=\"#f0f0f0\";\n\n")

		for _, module := range modules {
			dg.writeNode(module)
		}

		dg.builder.WriteString("  }\n\n")
	}

	// Write edges
	dg.builder.WriteString("  // Dependencies\n")
	for _, module := range dg.graph.Modules {
		if dg.shouldIncludeModule(module) {
			dg.writeEdges(module)
		}
	}
}

// writeNode writes a node with default styling
func (dg *DOTGenerator) writeNode(module *graph.Module) {
	color := dg.getNodeColor(module)
	dg.writeNodeWithColor(module, color)
}

// writeNodeWithColor writes a node with a specific color
func (dg *DOTGenerator) writeNodeWithColor(module *graph.Module, color string) {
	nodeID := dg.getNodeID(module)
	label := dg.getNodeLabel(module)

	dg.builder.WriteString(fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\"];\n",
		nodeID, color, escapeLabel(label)))
}

// writeNodeWithColorInCluster writes a node inside a cluster
func (dg *DOTGenerator) writeNodeWithColorInCluster(module *graph.Module, color string) {
	nodeID := dg.getNodeID(module)
	label := dg.getNodeLabel(module)

	dg.builder.WriteString(fmt.Sprintf("    \"%s\" [fillcolor=\"%s\", label=\"%s\"];\n",
		nodeID, color, escapeLabel(label)))
}

// writeEdges writes edges for a module's dependencies
func (dg *DOTGenerator) writeEdges(module *graph.Module) {
	fromID := dg.getNodeID(module)

	for _, depPath := range module.Dependencies {
		depModule := dg.graph.GetModule(depPath)
		if depModule == nil || !dg.shouldIncludeModule(depModule) {
			continue
		}

		toID := dg.getNodeID(depModule)
		dg.builder.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", fromID, toID))
	}
}

// writeSecurityEdges writes edges with security violation highlighting
func (dg *DOTGenerator) writeSecurityEdges(module *graph.Module, sec *analysis.SecurityAnalysis) {
	fromID := dg.getNodeID(module)

	// Check if any dependencies are violations
	violationPaths := make(map[string]bool)
	for _, violation := range sec.Violations {
		if violation.Crossing != nil && violation.Crossing.Source.Path == module.Path {
			violationPaths[violation.Crossing.Destination.Path] = true
		}
	}

	for _, depPath := range module.Dependencies {
		depModule := dg.graph.GetModule(depPath)
		if depModule == nil || !dg.shouldIncludeModule(depModule) {
			continue
		}

		toID := dg.getNodeID(depModule)

		// Highlight violations
		if violationPaths[depPath] {
			dg.builder.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [color=red, penwidth=2.0, style=bold];\n", fromID, toID))
		} else {
			dg.builder.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", fromID, toID))
		}
	}
}

// getNodeID returns a unique ID for a node
func (dg *DOTGenerator) getNodeID(module *graph.Module) string {
	return module.Path
}

// getNodeLabel returns the label for a node
func (dg *DOTGenerator) getNodeLabel(module *graph.Module) string {
	if dg.options.ShowLabels {
		return fmt.Sprintf("%s\n%s", filepath.Base(module.Path), module.Description)
	}
	return filepath.Base(module.Path)
}

// getNodeColor returns the color for a node based on ColorBy option
func (dg *DOTGenerator) getNodeColor(module *graph.Module) string {
	switch dg.options.ColorBy {
	case "language":
		return dg.getLanguageColor(module)
	case "layer":
		return dg.getLayerColor(module)
	default:
		return "#90CAF9" // Light blue default
	}
}

// getLanguageColor returns color based on language
func (dg *DOTGenerator) getLanguageColor(module *graph.Module) string {
	switch strings.ToLower(module.Language) {
	case "go":
		return "#00ADD8" // Go cyan
	case "python":
		return "#3776AB" // Python blue
	case "javascript", "js":
		return "#F7DF1E" // JavaScript yellow
	case "typescript", "ts":
		return "#3178C6" // TypeScript blue
	case "rust":
		return "#CE422B" // Rust orange
	case "java":
		return "#ED8B00" // Java orange
	default:
		return "#90CAF9" // Light blue
	}
}

// getLayerColor returns color based on layer
func (dg *DOTGenerator) getLayerColor(module *graph.Module) string {
	colors := map[string]string{
		"api":      "#4CAF50", // Green
		"service":  "#2196F3", // Blue
		"data":     "#FF9800", // Orange
		"internal": "#9E9E9E", // Gray
		"util":     "#E0E0E0", // Light gray
	}

	if color, ok := colors[strings.ToLower(module.Layer)]; ok {
		return color
	}
	return "#90CAF9" // Light blue default
}

// getFilteredModules returns modules after applying filters
func (dg *DOTGenerator) getFilteredModules() []*graph.Module {
	filtered := make([]*graph.Module, 0)

	for _, module := range dg.graph.Modules {
		if dg.shouldIncludeModule(module) {
			filtered = append(filtered, module)
		}
	}

	return filtered
}

// shouldIncludeModule checks if a module should be included
func (dg *DOTGenerator) shouldIncludeModule(module *graph.Module) bool {
	if dg.options.Filter == nil {
		return true
	}

	filter := dg.options.Filter

	// Check layer filter
	if len(filter.Layers) > 0 {
		found := false
		for _, layer := range filter.Layers {
			if strings.EqualFold(module.Layer, layer) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tag filter
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
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

// escapeLabel escapes special characters in DOT labels
func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
