package viz

import (
	"fmt"
	"strings"

	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
)

// MermaidType represents the type of Mermaid diagram
type MermaidType string

const (
	MermaidFlowchart MermaidType = "flowchart"
	MermaidGraph     MermaidType = "graph"
	MermaidClass     MermaidType = "class"
)

// MermaidOptions configures Mermaid diagram generation
type MermaidOptions struct {
	Type         MermaidType
	Direction    string // TD, LR, BT, RL
	ColorBy      string // layer, language, security
	Theme        string // default, dark, forest, neutral
	MaxDepth     int
	Filter       *FilterOptions
	Links        bool // Add clickable links
	Title        string
	UseSubgraphs bool // Group nodes by layer/package
}

// MermaidGenerator generates Mermaid diagram syntax
type MermaidGenerator struct {
	graph   *graph.Graph
	options MermaidOptions
	builder strings.Builder
	nodeIDs map[string]string // Map module paths to sanitized IDs
	colors  map[string]string // Map for node colors
}

// GenerateMermaid generates a Mermaid diagram from the graph
func GenerateMermaid(g *graph.Graph, opts MermaidOptions) (string, error) {
	gen := &MermaidGenerator{
		graph:   g,
		options: opts,
		nodeIDs: make(map[string]string),
		colors:  make(map[string]string),
	}

	return gen.generate()
}

// GenerateMermaidMarkdown generates Mermaid embedded in Markdown code block
func GenerateMermaidMarkdown(g *graph.Graph, opts MermaidOptions) (string, error) {
	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if opts.Title != "" {
		sb.WriteString("## ")
		sb.WriteString(opts.Title)
		sb.WriteString("\n\n")
	}
	sb.WriteString("```mermaid\n")
	sb.WriteString(mermaid)
	sb.WriteString("\n```\n")

	return sb.String(), nil
}

// generate creates the Mermaid diagram
func (mg *MermaidGenerator) generate() (string, error) {
	// Set defaults
	if mg.options.Direction == "" {
		mg.options.Direction = "TD"
	}
	if mg.options.Type == "" {
		mg.options.Type = MermaidFlowchart
	}

	// Generate diagram based on type
	switch mg.options.Type {
	case MermaidFlowchart:
		return mg.generateFlowchart()
	case MermaidGraph:
		return mg.generateGraph()
	case MermaidClass:
		return mg.generateClassDiagram()
	default:
		return "", fmt.Errorf("unsupported Mermaid type: %s", mg.options.Type)
	}
}

// generateFlowchart generates a flowchart diagram
func (mg *MermaidGenerator) generateFlowchart() (string, error) {
	mg.builder.WriteString(fmt.Sprintf("flowchart %s\n", mg.options.Direction))

	// Collect modules to include
	modules := mg.getFilteredModules()
	if len(modules) == 0 {
		return "", fmt.Errorf("no modules to display")
	}

	// Generate node IDs
	for _, module := range modules {
		mg.nodeIDs[module.Path] = mg.sanitizeNodeID(module.Path)
	}

	// Generate nodes and edges based on organization
	if mg.options.UseSubgraphs {
		mg.generateWithSubgraphs(modules)
	} else {
		mg.generateNodesAndEdges(modules)
	}

	// Add styling
	mg.addStyling(modules)

	return mg.builder.String(), nil
}

// generateGraph generates a graph diagram (similar to flowchart but with different syntax)
func (mg *MermaidGenerator) generateGraph() (string, error) {
	mg.builder.WriteString(fmt.Sprintf("graph %s\n", mg.options.Direction))

	modules := mg.getFilteredModules()
	if len(modules) == 0 {
		return "", fmt.Errorf("no modules to display")
	}

	for _, module := range modules {
		mg.nodeIDs[module.Path] = mg.sanitizeNodeID(module.Path)
	}

	if mg.options.UseSubgraphs {
		mg.generateWithSubgraphs(modules)
	} else {
		mg.generateNodesAndEdges(modules)
	}

	mg.addStyling(modules)

	return mg.builder.String(), nil
}

// generateClassDiagram generates a UML class diagram
func (mg *MermaidGenerator) generateClassDiagram() (string, error) {
	mg.builder.WriteString("classDiagram\n")

	modules := mg.getFilteredModules()
	if len(modules) == 0 {
		return "", fmt.Errorf("no modules to display")
	}

	// Generate classes
	for _, module := range modules {
		className := mg.sanitizeClassName(module.Name)
		mg.builder.WriteString(fmt.Sprintf("    class %s {\n", className))

		// Add metadata as attributes
		if module.Layer != "" {
			mg.builder.WriteString(fmt.Sprintf("        +layer: %s\n", module.Layer))
		}
		if module.Language != "" {
			mg.builder.WriteString(fmt.Sprintf("        +language: %s\n", module.Language))
		}

		// Add exports as methods
		for _, export := range module.Exports {
			mg.builder.WriteString(fmt.Sprintf("        +%s()\n", export))
		}

		mg.builder.WriteString("    }\n")
	}

	// Generate relationships
	mg.builder.WriteString("\n")
	for _, module := range modules {
		className := mg.sanitizeClassName(module.Name)
		for _, dep := range module.Dependencies {
			depModule := mg.graph.GetModule(dep)
			if depModule != nil && mg.shouldIncludeModule(depModule) {
				depClassName := mg.sanitizeClassName(depModule.Name)
				mg.builder.WriteString(fmt.Sprintf("    %s --> %s : depends on\n",
					className, depClassName))
			}
		}
	}

	return mg.builder.String(), nil
}

// generateNodesAndEdges generates nodes and edges without subgraphs
func (mg *MermaidGenerator) generateNodesAndEdges(modules []*graph.Module) {
	// Generate nodes
	for _, module := range modules {
		nodeID := mg.nodeIDs[module.Path]
		label := mg.getNodeLabel(module)

		// Use different shapes based on layer or type
		shape := mg.getNodeShape(module)

		mg.builder.WriteString(fmt.Sprintf("    %s%s%s\n",
			nodeID, shape, escapeMermaidLabel(label)))
	}

	// Generate edges
	mg.builder.WriteString("\n")
	for _, module := range modules {
		fromID := mg.nodeIDs[module.Path]
		for _, dep := range module.Dependencies {
			if toID, exists := mg.nodeIDs[dep]; exists {
				mg.builder.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
			}
		}
	}
}

// generateWithSubgraphs generates nodes grouped by layer in subgraphs
func (mg *MermaidGenerator) generateWithSubgraphs(modules []*graph.Module) {
	// Group modules by layer
	layerMap := make(map[string][]*graph.Module)
	for _, module := range modules {
		layer := module.Layer
		if layer == "" {
			layer = "unknown"
		}
		layerMap[layer] = append(layerMap[layer], module)
	}

	// Generate subgraphs
	for layer, layerModules := range layerMap {
		layerLabel := strings.Title(layer) + " Layer"
		mg.builder.WriteString(fmt.Sprintf("    subgraph %s[\"%s\"]\n",
			mg.sanitizeNodeID(layer), layerLabel))

		for _, module := range layerModules {
			nodeID := mg.nodeIDs[module.Path]
			label := mg.getNodeLabel(module)
			shape := mg.getNodeShape(module)

			mg.builder.WriteString(fmt.Sprintf("        %s%s%s\n",
				nodeID, shape, escapeMermaidLabel(label)))
		}

		mg.builder.WriteString("    end\n\n")
	}

	// Generate edges between nodes
	for _, module := range modules {
		fromID := mg.nodeIDs[module.Path]
		for _, dep := range module.Dependencies {
			if toID, exists := mg.nodeIDs[dep]; exists {
				mg.builder.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
			}
		}
	}
}

// addStyling adds CSS styling to nodes
func (mg *MermaidGenerator) addStyling(modules []*graph.Module) {
	if mg.options.ColorBy == "" {
		return
	}

	mg.builder.WriteString("\n")

	// Define color classes based on ColorBy option
	switch mg.options.ColorBy {
	case "layer":
		mg.addLayerStyling(modules)
	case "language":
		mg.addLanguageStyling(modules)
	case "security":
		// Security styling requires security analysis
		// This would be added if security analysis is provided
	}
}

// addLayerStyling adds styling based on module layers
func (mg *MermaidGenerator) addLayerStyling(modules []*graph.Module) {
	layerColors := map[string]string{
		"api":      "#4CAF50",
		"service":  "#2196F3",
		"data":     "#FF9800",
		"utils":    "#9C27B0",
		"model":    "#F44336",
		"graph":    "#00BCD4",
		"parser":   "#8BC34A",
		"analysis": "#FFC107",
		"query":    "#E91E63",
		"server":   "#3F51B5",
	}

	// Collect unique layers
	layers := make(map[string]bool)
	for _, module := range modules {
		if module.Layer != "" {
			layers[module.Layer] = true
		}
	}

	// Define styles
	styleNum := 0
	layerStyles := make(map[string]int)

	for layer := range layers {
		color := layerColors[layer]
		if color == "" {
			color = "#90CAF9" // Default blue
		}

		mg.builder.WriteString(fmt.Sprintf("    classDef style%d fill:%s,stroke:#333,stroke-width:2px\n",
			styleNum, color))
		layerStyles[layer] = styleNum
		styleNum++
	}

	// Apply styles to nodes
	mg.builder.WriteString("\n")
	for _, module := range modules {
		if styleNum, exists := layerStyles[module.Layer]; exists {
			nodeID := mg.nodeIDs[module.Path]
			mg.builder.WriteString(fmt.Sprintf("    class %s style%d\n", nodeID, styleNum))
		}
	}
}

// addLanguageStyling adds styling based on programming language
func (mg *MermaidGenerator) addLanguageStyling(modules []*graph.Module) {
	langColors := map[string]string{
		"go":         "#00ADD8",
		"python":     "#3776AB",
		"javascript": "#F7DF1E",
		"typescript": "#3178C6",
		"java":       "#007396",
		"rust":       "#000000",
	}

	languages := make(map[string]bool)
	for _, module := range modules {
		if module.Language != "" {
			languages[module.Language] = true
		}
	}

	styleNum := 0
	langStyles := make(map[string]int)

	for lang := range languages {
		color := langColors[lang]
		if color == "" {
			color = "#90CAF9"
		}

		mg.builder.WriteString(fmt.Sprintf("    classDef style%d fill:%s,stroke:#333,stroke-width:2px\n",
			styleNum, color))
		langStyles[lang] = styleNum
		styleNum++
	}

	mg.builder.WriteString("\n")
	for _, module := range modules {
		if styleNum, exists := langStyles[module.Language]; exists {
			nodeID := mg.nodeIDs[module.Path]
			mg.builder.WriteString(fmt.Sprintf("    class %s style%d\n", nodeID, styleNum))
		}
	}
}

// getFilteredModules returns modules that pass the filter
func (mg *MermaidGenerator) getFilteredModules() []*graph.Module {
	modules := make([]*graph.Module, 0)
	for _, module := range mg.graph.Modules {
		if mg.shouldIncludeModule(module) {
			modules = append(modules, module)
		}
	}
	return modules
}

// shouldIncludeModule checks if a module should be included based on filters
func (mg *MermaidGenerator) shouldIncludeModule(module *graph.Module) bool {
	if mg.options.Filter == nil {
		return true
	}

	// Filter by layers
	if len(mg.options.Filter.Layers) > 0 {
		found := false
		for _, layer := range mg.options.Filter.Layers {
			if module.Layer == layer {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by tags
	if len(mg.options.Filter.Tags) > 0 {
		found := false
		for _, filterTag := range mg.options.Filter.Tags {
			for _, moduleTag := range module.Tags {
				if moduleTag == filterTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// getNodeLabel returns the display label for a module
func (mg *MermaidGenerator) getNodeLabel(module *graph.Module) string {
	// Use short name for clarity
	return module.Name
}

// getNodeShape returns the Mermaid shape syntax for a module
func (mg *MermaidGenerator) getNodeShape(module *graph.Module) string {
	// Different shapes for different layers
	switch module.Layer {
	case "api", "server":
		return "[" // Rectangle
	case "service":
		return "(" // Rounded rectangle (stadium)
	case "data", "storage":
		return "{" // Rhombus
	case "model":
		return "[/" // Parallelogram
	default:
		return "[" // Default rectangle
	}
}

// sanitizeNodeID creates a valid Mermaid node ID from a path
func (mg *MermaidGenerator) sanitizeNodeID(path string) string {
	// Replace special characters with underscores
	id := strings.ReplaceAll(path, "/", "_")
	id = strings.ReplaceAll(id, ".", "_")
	id = strings.ReplaceAll(id, "-", "_")
	return id
}

// sanitizeClassName creates a valid class name
func (mg *MermaidGenerator) sanitizeClassName(name string) string {
	// Remove file extension and sanitize
	name = strings.TrimSuffix(name, ".go")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// escapeMermaidLabel escapes special characters in labels
func escapeMermaidLabel(label string) string {
	// Escape quotes and special characters
	label = strings.ReplaceAll(label, "\"", "&quot;")
	label = strings.ReplaceAll(label, "[", "(")
	label = strings.ReplaceAll(label, "]", ")")
	return "\"" + label + "\""
}

// GenerateMermaidForImpact generates Mermaid diagram for impact analysis
func GenerateMermaidForImpact(g *graph.Graph, impact *analysis.ImpactResult, opts MermaidOptions) (string, error) {
	opts.Type = MermaidFlowchart
	if opts.Direction == "" {
		opts.Direction = "TD"
	}

	gen := &MermaidGenerator{
		graph:   g,
		options: opts,
		nodeIDs: make(map[string]string),
		colors:  make(map[string]string),
	}

	gen.builder.WriteString(fmt.Sprintf("flowchart %s\n", opts.Direction))

	// Get all affected modules
	affectedPaths := make(map[string]bool)
	affectedPaths[impact.TargetModule] = true
	for _, dep := range impact.DirectDependents {
		affectedPaths[dep] = true
	}
	for path := range impact.TransitiveDependents {
		affectedPaths[path] = true
	}

	// Generate node IDs
	for path := range affectedPaths {
		module := g.GetModule(path)
		if module != nil {
			gen.nodeIDs[path] = gen.sanitizeNodeID(path)
		}
	}

	// Generate nodes with impact coloring
	for path := range affectedPaths {
		module := g.GetModule(path)
		if module == nil {
			continue
		}

		nodeID := gen.nodeIDs[path]
		label := module.Name
		shape := gen.getNodeShape(module)

		gen.builder.WriteString(fmt.Sprintf("    %s%s%s\n",
			nodeID, shape, escapeMermaidLabel(label)))
	}

	// Generate edges
	gen.builder.WriteString("\n")
	for path := range affectedPaths {
		module := g.GetModule(path)
		if module == nil {
			continue
		}

		fromID := gen.nodeIDs[path]
		for _, dep := range module.Dependencies {
			if toID, exists := gen.nodeIDs[dep]; exists {
				gen.builder.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
			}
		}
	}

	// Add impact-based styling
	gen.builder.WriteString("\n")
	gen.builder.WriteString("    classDef changed fill:#FF5722,stroke:#333,stroke-width:3px\n")
	gen.builder.WriteString("    classDef direct fill:#FF9800,stroke:#333,stroke-width:2px\n")
	gen.builder.WriteString("    classDef transitive fill:#FFC107,stroke:#333,stroke-width:1px\n")

	gen.builder.WriteString("\n")
	// Apply styles
	changedID := gen.nodeIDs[impact.TargetModule]
	gen.builder.WriteString(fmt.Sprintf("    class %s changed\n", changedID))

	for _, dep := range impact.DirectDependents {
		if nodeID, exists := gen.nodeIDs[dep]; exists {
			gen.builder.WriteString(fmt.Sprintf("    class %s direct\n", nodeID))
		}
	}

	for path := range impact.TransitiveDependents {
		if path != impact.TargetModule && !contains(impact.DirectDependents, path) {
			if nodeID, exists := gen.nodeIDs[path]; exists {
				gen.builder.WriteString(fmt.Sprintf("    class %s transitive\n", nodeID))
			}
		}
	}

	return gen.builder.String(), nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
