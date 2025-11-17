package viz

import (
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/analysis"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test", tripleStore)

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
		URI:          "<#users-svc.go>",
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

func TestGenerateDOT_Dependency(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify DOT structure
	if !strings.Contains(dot, "digraph GraphFS") {
		t.Error("Missing digraph declaration")
	}
	if !strings.Contains(dot, "rankdir=LR") {
		t.Error("Missing rankdir attribute")
	}

	// Verify nodes are present
	nodes := []string{"api/handlers.go", "services/auth.go", "services/users.go", "data/users.go"}
	for _, node := range nodes {
		if !strings.Contains(dot, node) {
			t.Errorf("Missing node: %s", node)
		}
	}

	// Verify edges are present
	edges := [][2]string{
		{"api/handlers.go", "services/auth.go"},
		{"api/handlers.go", "services/users.go"},
		{"services/auth.go", "data/users.go"},
		{"services/users.go", "data/users.go"},
	}
	for _, edge := range edges {
		edgePattern := edge[0] + "\" -> \"" + edge[1]
		if !strings.Contains(dot, edgePattern) {
			t.Errorf("Missing edge: %s -> %s", edge[0], edge[1])
		}
	}
}

func TestGenerateDOT_Layer(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizLayer,
		Rankdir: "TB",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify subgraphs for layers
	if !strings.Contains(dot, "subgraph cluster_") {
		t.Error("Missing layer subgraphs")
	}
	if !strings.Contains(dot, "Layer: api") {
		t.Error("Missing API layer")
	}
	if !strings.Contains(dot, "Layer: service") {
		t.Error("Missing service layer")
	}
	if !strings.Contains(dot, "Layer: data") {
		t.Error("Missing data layer")
	}
}

func TestGenerateDOT_Impact(t *testing.T) {
	g := createTestGraph()

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

	opts := VizOptions{
		Type:    VizImpact,
		Impact:  impact,
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify changed module is highlighted
	if !strings.Contains(dot, "#FF5722") { // Red-orange for changed
		t.Error("Changed module not highlighted")
	}

	// Verify directly affected are highlighted
	if !strings.Contains(dot, "#FF9800") { // Orange for directly affected
		t.Error("Directly affected modules not highlighted")
	}

	// Verify transitively affected are highlighted
	if !strings.Contains(dot, "#FFC107") { // Amber for transitively affected
		t.Error("Transitively affected modules not highlighted")
	}
}

func TestGenerateDOT_Security(t *testing.T) {
	g := createTestGraph()

	// Create security analysis
	secAnalysis := &analysis.SecurityAnalysis{
		Zones: map[analysis.SecurityZone][]*analysis.ModuleZone{
			analysis.ZonePublic: {
				{Module: g.GetModule("api/handlers.go"), Zone: analysis.ZonePublic},
			},
			analysis.ZoneTrusted: {
				{Module: g.GetModule("services/auth.go"), Zone: analysis.ZoneTrusted},
				{Module: g.GetModule("services/users.go"), Zone: analysis.ZoneTrusted},
			},
			analysis.ZoneData: {
				{Module: g.GetModule("data/users.go"), Zone: analysis.ZoneData},
			},
		},
		Violations: []*analysis.SecurityViolation{},
	}

	opts := VizOptions{
		Type:     VizSecurity,
		Security: secAnalysis,
		Rankdir:  "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify security zones are created as subgraphs
	if !strings.Contains(dot, "Public Zone") {
		t.Error("Missing public zone")
	}
	if !strings.Contains(dot, "Trusted Zone") {
		t.Error("Missing trusted zone")
	}
	if !strings.Contains(dot, "Data Zone") {
		t.Error("Missing data zone")
	}

	// Verify zone colors
	zoneColors := []string{"#4CAF50", "#2196F3", "#FF9800"} // Green, Blue, Orange
	for _, color := range zoneColors {
		if !strings.Contains(dot, color) {
			t.Errorf("Missing zone color: %s", color)
		}
	}
}

func TestGenerateDOT_WithTitle(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		Title:   "Test Dependency Graph",
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	if !strings.Contains(dot, "Test Dependency Graph") {
		t.Error("Missing graph title")
	}
}

func TestGenerateDOT_ColorByLanguage(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		ColorBy: "language",
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify Go color is used
	if !strings.Contains(dot, "#00ADD8") { // Go cyan
		t.Error("Missing Go language color")
	}
}

func TestGenerateDOT_ColorByLayer(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		ColorBy: "layer",
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify layer colors
	layerColors := []string{"#4CAF50", "#2196F3", "#FF9800"} // API, Service, Data
	colorFound := false
	for _, color := range layerColors {
		if strings.Contains(dot, color) {
			colorFound = true
			break
		}
	}
	if !colorFound {
		t.Error("Missing layer colors")
	}
}

func TestGenerateDOT_WithFilter(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		Rankdir: "LR",
		Filter: &FilterOptions{
			Layers: []string{"service"},
		},
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify only service layer modules are included
	if !strings.Contains(dot, "services/auth.go") {
		t.Error("Missing service module")
	}
	if !strings.Contains(dot, "services/users.go") {
		t.Error("Missing service module")
	}

	// API and data modules should be excluded
	// Note: They might still appear as dependency targets
	// So we check that they don't have their own node definitions with fillcolor
	if strings.Count(dot, "api/handlers.go") > 1 {
		// Should only appear once as edge target, not as node definition
		t.Error("API module should be filtered out")
	}
}

func TestGenerateDOT_WithTagFilter(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:    VizDependency,
		Rankdir: "LR",
		Filter: &FilterOptions{
			Tags: []string{"auth"},
		},
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify auth service is included
	if !strings.Contains(dot, "services/auth.go") {
		t.Error("Missing auth service")
	}
}

func TestEscapeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`simple`, `simple`},
		{`with\backslash`, `with\\backslash`},
		{`with"quotes"`, `with\"quotes\"`},
		{"with\nnewline", `with\nnewline`},
		{`all\special"chars\n`, `all\\special\"chars\\n`},
	}

	for _, tt := range tests {
		result := escapeLabel(tt.input)
		if result != tt.expected {
			t.Errorf("escapeLabel(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateDOT_ShowLabels(t *testing.T) {
	g := createTestGraph()
	opts := VizOptions{
		Type:       VizDependency,
		ShowLabels: true,
		Rankdir:    "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Verify descriptions are included
	if !strings.Contains(dot, "API handlers") {
		t.Error("Missing module description in label")
	}
	if !strings.Contains(dot, "Authentication service") {
		t.Error("Missing module description in label")
	}
}

func TestGenerateDOT_EmptyGraph(t *testing.T) {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("empty", tripleStore)

	opts := VizOptions{
		Type:    VizDependency,
		Rankdir: "LR",
	}

	dot, err := GenerateDOT(g, opts)
	if err != nil {
		t.Fatalf("GenerateDOT failed: %v", err)
	}

	// Should still have valid DOT structure
	if !strings.Contains(dot, "digraph GraphFS") {
		t.Error("Missing digraph declaration")
	}
}

func TestIsGraphVizAvailable(t *testing.T) {
	// Just test that the function runs without error
	available := isGraphVizAvailable()
	t.Logf("GraphViz available: %v", available)
}

func TestGetAvailableLayouts(t *testing.T) {
	layouts := GetAvailableLayouts()
	t.Logf("Available layouts: %v", layouts)

	// Should return a slice (even if empty)
	if layouts == nil {
		t.Error("GetAvailableLayouts returned nil")
	}
}

func TestValidateLayout(t *testing.T) {
	// Test default layout
	err := ValidateLayout("")
	t.Logf("Validate default layout: %v", err)

	// Test specific layout (may or may not be available)
	err = ValidateLayout("dot")
	t.Logf("Validate 'dot' layout: %v", err)

	// Test invalid layout
	err = ValidateLayout("nonexistent_layout_12345")
	if err == nil {
		t.Error("Should return error for invalid layout")
	}
}
