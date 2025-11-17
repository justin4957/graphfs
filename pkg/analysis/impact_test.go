package analysis

import (
	"testing"

	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraphForImpact() *graph.Graph {
	g := &graph.Graph{
		Modules: make(map[string]*graph.Module),
	}

	// Create a realistic dependency structure:
	//
	//        API (handlers)
	//       /   \
	//   ServiceA  ServiceB (services)
	//      |        |
	//   UtilsA   UtilsB (utils)
	//       \     /
	//        Core
	//
	// Plus some isolated modules

	g.Modules["handlers/api.go"] = &graph.Module{
		Path:         "handlers/api.go",
		Layer:        "handlers",
		Dependencies: []string{"services/serviceA.go", "services/serviceB.go"},
	}

	g.Modules["services/serviceA.go"] = &graph.Module{
		Path:         "services/serviceA.go",
		Layer:        "services",
		Dependencies: []string{"utils/utilsA.go"},
	}

	g.Modules["services/serviceB.go"] = &graph.Module{
		Path:         "services/serviceB.go",
		Layer:        "services",
		Dependencies: []string{"utils/utilsB.go"},
	}

	g.Modules["utils/utilsA.go"] = &graph.Module{
		Path:         "utils/utilsA.go",
		Layer:        "utils",
		Dependencies: []string{"core/core.go"},
	}

	g.Modules["utils/utilsB.go"] = &graph.Module{
		Path:         "utils/utilsB.go",
		Layer:        "utils",
		Dependencies: []string{"core/core.go"},
	}

	g.Modules["core/core.go"] = &graph.Module{
		Path:         "core/core.go",
		Layer:        "core",
		Dependencies: []string{},
	}

	g.Modules["isolated/module.go"] = &graph.Module{
		Path:         "isolated/module.go",
		Layer:        "utils",
		Dependencies: []string{},
	}

	return g
}

func TestNewImpactAnalysis(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	if ia == nil {
		t.Fatal("Expected non-nil ImpactAnalysis")
	}

	if ia.graph != g {
		t.Error("Graph not set correctly")
	}
}

func TestAnalyzeImpact_CoreModule(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("core/core.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Core is used by everyone
	if result.TargetModule != "core/core.go" {
		t.Errorf("Expected target 'core/core.go', got '%s'", result.TargetModule)
	}

	// Should have 2 direct dependents (utilsA and utilsB)
	if len(result.DirectDependents) != 2 {
		t.Errorf("Expected 2 direct dependents, got %d", len(result.DirectDependents))
	}

	// Should have 0 direct dependencies
	if len(result.DirectDependencies) != 0 {
		t.Errorf("Expected 0 direct dependencies, got %d", len(result.DirectDependencies))
	}

	// Should have transitive impact (utilsA, utilsB, serviceA, serviceB, api)
	if result.TotalImpactedModules != 5 {
		t.Errorf("Expected 5 impacted modules, got %d", result.TotalImpactedModules)
	}

	// Risk should be HIGH or CRITICAL
	if result.RiskLevel != RiskLevelHigh && result.RiskLevel != RiskLevelCritical {
		t.Errorf("Expected HIGH or CRITICAL risk for core module, got %s", result.RiskLevel)
	}

	// Should have multiple layers impacted
	if len(result.ImpactByLayer) < 3 {
		t.Errorf("Expected at least 3 layers impacted, got %d", len(result.ImpactByLayer))
	}
}

func TestAnalyzeImpact_LeafModule(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("handlers/api.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// API is a leaf (no one depends on it)
	if len(result.DirectDependents) != 0 {
		t.Errorf("Expected 0 direct dependents, got %d", len(result.DirectDependents))
	}

	// Should have 2 direct dependencies
	if len(result.DirectDependencies) != 2 {
		t.Errorf("Expected 2 direct dependencies, got %d", len(result.DirectDependencies))
	}

	// Should have no transitive impact
	if result.TotalImpactedModules != 0 {
		t.Errorf("Expected 0 impacted modules, got %d", result.TotalImpactedModules)
	}

	// Risk should be LOW
	if result.RiskLevel != RiskLevelLow {
		t.Errorf("Expected LOW risk for leaf module, got %s", result.RiskLevel)
	}
}

func TestAnalyzeImpact_IsolatedModule(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("isolated/module.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Isolated module has no impact
	if len(result.DirectDependents) != 0 {
		t.Errorf("Expected 0 direct dependents, got %d", len(result.DirectDependents))
	}

	if result.TotalImpactedModules != 0 {
		t.Errorf("Expected 0 impacted modules, got %d", result.TotalImpactedModules)
	}

	if result.RiskLevel != RiskLevelLow {
		t.Errorf("Expected LOW risk for isolated module, got %s", result.RiskLevel)
	}
}

func TestAnalyzeImpact_NonExistentModule(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	_, err := ia.AnalyzeImpact("nonexistent/module.go")
	if err == nil {
		t.Fatal("Expected error for non-existent module")
	}
}

func TestAnalyzeImpact_MiddleModule(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("services/serviceA.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// ServiceA should have 1 direct dependent (api)
	if len(result.DirectDependents) != 1 {
		t.Errorf("Expected 1 direct dependent, got %d", len(result.DirectDependents))
	}

	// Should have 1 direct dependency (utilsA)
	if len(result.DirectDependencies) != 1 {
		t.Errorf("Expected 1 direct dependency, got %d", len(result.DirectDependencies))
	}

	// Should have transitive impact (api)
	if result.TotalImpactedModules != 1 {
		t.Errorf("Expected 1 impacted module, got %d", result.TotalImpactedModules)
	}

	// Risk should be LOW or MEDIUM
	if result.RiskLevel != RiskLevelLow && result.RiskLevel != RiskLevelMedium {
		t.Errorf("Expected LOW or MEDIUM risk, got %s", result.RiskLevel)
	}
}

func TestAnalyzeMultipleModules(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	modules := []string{"services/serviceA.go", "services/serviceB.go"}
	result, err := ia.AnalyzeMultipleModules(modules)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Both services are used by api
	if len(result.DirectDependents) != 1 {
		t.Errorf("Expected 1 combined direct dependent, got %d", len(result.DirectDependents))
	}

	// Both services depend on different utils
	if len(result.DirectDependencies) != 2 {
		t.Errorf("Expected 2 combined direct dependencies, got %d", len(result.DirectDependencies))
	}

	// Should have transitive impact (api)
	if result.TotalImpactedModules != 1 {
		t.Errorf("Expected 1 impacted module, got %d", result.TotalImpactedModules)
	}

	// Target should indicate multiple modules
	if result.TargetModule != "2 modules" {
		t.Errorf("Expected target '2 modules', got '%s'", result.TargetModule)
	}
}

func TestAnalyzeMultipleModules_Empty(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	_, err := ia.AnalyzeMultipleModules([]string{})
	if err == nil {
		t.Fatal("Expected error for empty module list")
	}
}

func TestAnalyzeMultipleModules_NonExistent(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	modules := []string{"services/serviceA.go", "nonexistent/module.go"}
	_, err := ia.AnalyzeMultipleModules(modules)
	if err == nil {
		t.Fatal("Expected error for non-existent module")
	}
}

func TestRiskLevels(t *testing.T) {
	tests := []struct {
		name         string
		module       string
		expectedRisk RiskLevel
	}{
		{
			name:         "core module - high impact",
			module:       "core/core.go",
			expectedRisk: RiskLevelHigh, // or CRITICAL
		},
		{
			name:         "leaf module - low impact",
			module:       "handlers/api.go",
			expectedRisk: RiskLevelLow,
		},
		{
			name:         "isolated module - low impact",
			module:       "isolated/module.go",
			expectedRisk: RiskLevelLow,
		},
	}

	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ia.AnalyzeImpact(tt.module)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// For core module, accept either HIGH or CRITICAL
			if tt.module == "core/core.go" {
				if result.RiskLevel != RiskLevelHigh && result.RiskLevel != RiskLevelCritical {
					t.Errorf("Expected HIGH or CRITICAL risk, got %s", result.RiskLevel)
				}
			} else if result.RiskLevel != tt.expectedRisk {
				t.Errorf("Expected risk %s, got %s", tt.expectedRisk, result.RiskLevel)
			}
		})
	}
}

func TestImpactMetrics(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("core/core.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check max depth
	if result.MaxImpactDepth < 1 {
		t.Errorf("Expected max depth >= 1, got %d", result.MaxImpactDepth)
	}

	// Check impact percentage
	if result.ImpactPercentage <= 0 || result.ImpactPercentage > 100 {
		t.Errorf("Invalid impact percentage: %.2f", result.ImpactPercentage)
	}

	// Check that recommendations exist
	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}

	// Check that risk factors exist for high-impact module
	if len(result.RiskFactors) == 0 {
		t.Error("Expected risk factors to be identified")
	}
}

func TestImpactByLayer(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	result, err := ia.AnalyzeImpact("core/core.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Core should impact multiple layers
	if len(result.ImpactByLayer) < 2 {
		t.Errorf("Expected at least 2 layers impacted, got %d", len(result.ImpactByLayer))
	}

	// Check that layer counts are positive
	for layer, count := range result.ImpactByLayer {
		if count <= 0 {
			t.Errorf("Invalid count for layer %s: %d", layer, count)
		}
	}
}

func TestCompareImpacts(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	modules := []string{"core/core.go", "services/serviceA.go", "handlers/api.go"}
	results, err := ia.CompareImpacts(modules)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify each module has a result
	for _, module := range modules {
		if _, exists := results[module]; !exists {
			t.Errorf("Missing result for module %s", module)
		}
	}

	// Core should have higher impact than api
	coreImpact := results["core/core.go"].TotalImpactedModules
	apiImpact := results["handlers/api.go"].TotalImpactedModules

	if coreImpact <= apiImpact {
		t.Errorf("Expected core impact (%d) > api impact (%d)", coreImpact, apiImpact)
	}
}

func TestCompareImpacts_NonExistent(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	modules := []string{"core/core.go", "nonexistent/module.go"}
	_, err := ia.CompareImpacts(modules)
	if err == nil {
		t.Fatal("Expected error for non-existent module")
	}
}

func TestBreakingChanges(t *testing.T) {
	g := createTestGraphForImpact()
	ia := NewImpactAnalysis(g)

	tests := []struct {
		name           string
		module         string
		expectBreaking bool
	}{
		{
			name:           "core module - breaking",
			module:         "core/core.go",
			expectBreaking: true,
		},
		{
			name:           "leaf module - not breaking",
			module:         "handlers/api.go",
			expectBreaking: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ia.AnalyzeImpact(tt.module)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.BreakingChanges != tt.expectBreaking {
				t.Errorf("Expected breaking=%v, got %v", tt.expectBreaking, result.BreakingChanges)
			}
		})
	}
}
