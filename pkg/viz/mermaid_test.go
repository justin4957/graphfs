package viz

import (
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createMermaidTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test-project", tripleStore)

	// API layer
	api := &graph.Module{
		Path:         "api/handlers.go",
		URI:          "<#handlers.go>",
		Name:         "handlers.go",
		Description:  "API handlers",
		Language:     "go",
		Layer:        "api",
		Tags:         []string{"api", "http"},
		Dependencies: []string{"services/auth.go", "services/users.go"},
		Exports:      []string{"HandleRequest"},
	}
	g.AddModule(api)

	// Service layer
	authService := &graph.Module{
		Path:         "services/auth.go",
		URI:          "<#auth.go>",
		Name:         "auth.go",
		Description:  "Authentication service",
		Language:     "go",
		Layer:        "service",
		Tags:         []string{"service", "auth"},
		Dependencies: []string{"data/users.go"},
		Exports:      []string{"AuthService"},
	}
	g.AddModule(authService)

	userService := &graph.Module{
		Path:         "services/users.go",
		URI:          "<#users.go>",
		Name:         "users.go",
		Description:  "User service",
		Language:     "go",
		Layer:        "service",
		Tags:         []string{"service", "users"},
		Dependencies: []string{"data/users.go"},
		Exports:      []string{"UserService"},
	}
	g.AddModule(userService)

	// Data layer
	userData := &graph.Module{
		Path:        "data/users.go",
		URI:         "<#users-data.go>",
		Name:        "users.go",
		Description: "User data access",
		Language:    "go",
		Layer:       "data",
		Tags:        []string{"data", "database"},
		Exports:     []string{"UserRepository"},
	}
	g.AddModule(userData)

	return g
}

func TestGenerateMermaid_Flowchart(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify basic structure
	if !strings.HasPrefix(mermaid, "flowchart TD") {
		t.Error("Missing flowchart declaration")
	}

	// Verify nodes are present
	nodes := []string{"api_handlers_go", "services_auth_go", "services_users_go", "data_users_go"}
	for _, node := range nodes {
		if !strings.Contains(mermaid, node) {
			t.Errorf("Missing node: %s", node)
		}
	}

	// Verify edges are present
	edges := [][2]string{
		{"api_handlers_go", "services_auth_go"},
		{"api_handlers_go", "services_users_go"},
		{"services_auth_go", "data_users_go"},
	}
	for _, edge := range edges {
		edgePattern := edge[0] + " --> " + edge[1]
		if !strings.Contains(mermaid, edgePattern) {
			t.Errorf("Missing edge: %s --> %s", edge[0], edge[1])
		}
	}
}

func TestGenerateMermaid_Graph(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidGraph,
		Direction: "LR",
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify graph declaration
	if !strings.HasPrefix(mermaid, "graph LR") {
		t.Error("Missing graph LR declaration")
	}

	// Verify nodes
	if !strings.Contains(mermaid, "api_handlers_go") {
		t.Error("Missing API handler node")
	}
}

func TestGenerateMermaid_ClassDiagram(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type: MermaidClass,
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify class diagram declaration
	if !strings.HasPrefix(mermaid, "classDiagram") {
		t.Error("Missing classDiagram declaration")
	}

	// Verify classes are defined
	if !strings.Contains(mermaid, "class handlers") {
		t.Error("Missing handlers class")
	}
	if !strings.Contains(mermaid, "class auth") {
		t.Error("Missing auth class")
	}

	// Verify exports are shown
	if !strings.Contains(mermaid, "HandleRequest()") {
		t.Error("Missing exported function")
	}

	// Verify relationships
	if !strings.Contains(mermaid, "-->") {
		t.Error("Missing class relationships")
	}
}

func TestGenerateMermaid_WithSubgraphs(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:         MermaidFlowchart,
		Direction:    "TD",
		UseSubgraphs: true,
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify subgraphs are created
	if !strings.Contains(mermaid, "subgraph") {
		t.Error("Missing subgraph declaration")
	}

	// Verify layer subgraphs
	if !strings.Contains(mermaid, "Api Layer") {
		t.Error("Missing API layer subgraph")
	}
	if !strings.Contains(mermaid, "Service Layer") {
		t.Error("Missing service layer subgraph")
	}
	if !strings.Contains(mermaid, "Data Layer") {
		t.Error("Missing data layer subgraph")
	}
}

func TestGenerateMermaid_WithLayerColors(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
		ColorBy:   "layer",
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify CSS classes are defined
	if !strings.Contains(mermaid, "classDef") {
		t.Error("Missing classDef declarations")
	}

	// Verify colors are applied
	if !strings.Contains(mermaid, "fill:") {
		t.Error("Missing color styling")
	}

	// Verify class assignments
	if !strings.Contains(mermaid, "class ") {
		t.Error("Missing class assignments")
	}
}

func TestGenerateMermaid_WithLanguageColors(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
		ColorBy:   "language",
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Verify styling is present
	if !strings.Contains(mermaid, "classDef") {
		t.Error("Missing style definitions")
	}

	// Verify Go color is used (all modules are Go)
	if !strings.Contains(mermaid, "#00ADD8") {
		t.Error("Missing Go language color")
	}
}

func TestGenerateMermaid_WithLayerFilter(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
		Filter: &FilterOptions{
			Layers: []string{"service"},
		},
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Should include service layer
	if !strings.Contains(mermaid, "services_auth_go") {
		t.Error("Missing service module")
	}

	// Should NOT include API layer
	if strings.Contains(mermaid, "api_handlers_go") {
		t.Error("API layer should be filtered out")
	}
}

func TestGenerateMermaid_WithTagFilter(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
		Filter: &FilterOptions{
			Tags: []string{"auth"},
		},
	}

	mermaid, err := GenerateMermaid(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaid failed: %v", err)
	}

	// Should include auth service
	if !strings.Contains(mermaid, "services_auth_go") {
		t.Error("Missing auth service")
	}

	// Should NOT include user service (no auth tag)
	if strings.Contains(mermaid, "services_users_go") {
		t.Error("User service should be filtered out")
	}
}

func TestGenerateMermaidMarkdown(t *testing.T) {
	g := createMermaidTestGraph()
	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
		Title:     "Dependency Graph",
	}

	markdown, err := GenerateMermaidMarkdown(g, opts)
	if err != nil {
		t.Fatalf("GenerateMermaidMarkdown failed: %v", err)
	}

	// Verify markdown structure
	if !strings.Contains(markdown, "## Dependency Graph") {
		t.Error("Missing title")
	}

	if !strings.Contains(markdown, "```mermaid") {
		t.Error("Missing mermaid code block start")
	}

	if !strings.Contains(markdown, "```\n") {
		t.Error("Missing mermaid code block end")
	}

	if !strings.Contains(markdown, "flowchart TD") {
		t.Error("Missing flowchart content")
	}
}

func TestGenerateMermaidForImpact(t *testing.T) {
	g := createMermaidTestGraph()

	// Create impact result
	impact := &analysis.ImpactResult{
		TargetModule:     "data/users.go",
		DirectDependents: []string{"services/auth.go", "services/users.go"},
		TransitiveDependents: map[string]int{
			"services/auth.go":  1,
			"services/users.go": 1,
			"api/handlers.go":   2,
		},
		TotalImpactedModules: 3,
	}

	opts := MermaidOptions{
		Direction: "BT", // Bottom to top to show impact flow
	}

	mermaid, err := GenerateMermaidForImpact(g, impact, opts)
	if err != nil {
		t.Fatalf("GenerateMermaidForImpact failed: %v", err)
	}

	// Verify flowchart
	if !strings.Contains(mermaid, "flowchart BT") {
		t.Error("Missing flowchart BT declaration")
	}

	// Verify target module is present
	if !strings.Contains(mermaid, "data_users_go") {
		t.Error("Missing target module")
	}

	// Verify impacted modules
	if !strings.Contains(mermaid, "services_auth_go") {
		t.Error("Missing directly impacted module")
	}
	if !strings.Contains(mermaid, "api_handlers_go") {
		t.Error("Missing transitively impacted module")
	}

	// Verify styling for impact levels
	if !strings.Contains(mermaid, "classDef changed") {
		t.Error("Missing changed module style")
	}
	if !strings.Contains(mermaid, "classDef direct") {
		t.Error("Missing direct impact style")
	}
	if !strings.Contains(mermaid, "classDef transitive") {
		t.Error("Missing transitive impact style")
	}

	// Verify styles are applied
	if !strings.Contains(mermaid, "class data_users_go changed") {
		t.Error("Changed style not applied to target")
	}
}

func TestEscapeMermaidLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "\"simple\""},
		{"with\"quotes\"", "\"with&quot;quotes&quot;\""},
		{"with[brackets]", "\"with(brackets)\""},
		{"complex\"label[test]", "\"complex&quot;label(test)\""},
	}

	for _, tt := range tests {
		result := escapeMermaidLabel(tt.input)
		if result != tt.expected {
			t.Errorf("escapeMermaidLabel(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeNodeID(t *testing.T) {
	gen := &MermaidGenerator{
		nodeIDs: make(map[string]string),
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"api/handlers.go", "api_handlers_go"},
		{"services/auth.go", "services_auth_go"},
		{"pkg/graph/graph.go", "pkg_graph_graph_go"},
		{"internal-store-store.go", "internal_store_store_go"},
	}

	for _, tt := range tests {
		result := gen.sanitizeNodeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeNodeID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetNodeShape(t *testing.T) {
	gen := &MermaidGenerator{}

	tests := []struct {
		layer    string
		expected string
	}{
		{"api", "["},
		{"server", "["},
		{"service", "("},
		{"data", "{"},
		{"storage", "{"},
		{"model", "[/"},
		{"unknown", "["},
	}

	for _, tt := range tests {
		module := &graph.Module{Layer: tt.layer}
		result := gen.getNodeShape(module)
		if result != tt.expected {
			t.Errorf("getNodeShape(layer=%q) = %q, want %q", tt.layer, result, tt.expected)
		}
	}
}

func TestGenerateMermaid_EmptyGraph(t *testing.T) {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("empty", tripleStore)

	opts := MermaidOptions{
		Type:      MermaidFlowchart,
		Direction: "TD",
	}

	_, err := GenerateMermaid(g, opts)
	if err == nil {
		t.Error("Expected error for empty graph, got nil")
	}
}

func TestGenerateMermaid_DirectionOptions(t *testing.T) {
	g := createMermaidTestGraph()

	directions := []string{"TD", "LR", "BT", "RL"}
	for _, dir := range directions {
		opts := MermaidOptions{
			Type:      MermaidFlowchart,
			Direction: dir,
		}

		mermaid, err := GenerateMermaid(g, opts)
		if err != nil {
			t.Fatalf("GenerateMermaid with direction %s failed: %v", dir, err)
		}

		if !strings.Contains(mermaid, "flowchart "+dir) {
			t.Errorf("Missing direction %s in output", dir)
		}
	}
}
