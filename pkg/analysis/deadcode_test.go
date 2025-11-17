package analysis

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraphForDeadCode() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test", tripleStore)

	// Entry point module (main)
	main := &graph.Module{
		Path:         "cmd/main.go",
		URI:          "<#main.go>",
		Name:         "main.go",
		Description:  "Main entry point",
		Layer:        "cmd",
		Tags:         []string{"entrypoint"},
		Dependencies: []string{"services/auth.go", "services/api.go"},
		Exports:      []string{"main"},
	}
	g.AddModule(main)

	// Active service module (referenced by main)
	auth := &graph.Module{
		Path:         "services/auth.go",
		URI:          "<#auth.go>",
		Name:         "auth.go",
		Description:  "Authentication service",
		Layer:        "services",
		Tags:         []string{"auth", "security"},
		Dependencies: []string{"utils/crypto.go"},
		Exports:      []string{"AuthService", "Login"},
	}
	g.AddModule(auth)

	// Another active module
	api := &graph.Module{
		Path:         "services/api.go",
		URI:          "<#api.go>",
		Name:         "api.go",
		Description:  "API handlers",
		Layer:        "services",
		Tags:         []string{"api", "http"},
		Dependencies: []string{"services/auth.go"},
		Exports:      []string{"APIServer"},
	}
	g.AddModule(api)

	// Utility module (referenced by auth)
	crypto := &graph.Module{
		Path:         "utils/crypto.go",
		URI:          "<#crypto.go>",
		Name:         "crypto.go",
		Description:  "Cryptographic utilities",
		Layer:        "utils",
		Tags:         []string{"crypto", "security"},
		Dependencies: []string{},
		Exports:      []string{"Hash", "Verify"},
	}
	g.AddModule(crypto)

	// Dead code: unreferenced module
	legacy := &graph.Module{
		Path:         "utils/legacy_helper.go",
		URI:          "<#legacy_helper.go>",
		Name:         "legacy_helper.go",
		Description:  "Legacy helper functions",
		Layer:        "utils",
		Tags:         []string{"legacy"},
		Dependencies: []string{},
		Exports:      []string{"OldHelper"},
	}
	g.AddModule(legacy)

	// Dead code: internal module with no refs
	internal := &graph.Module{
		Path:         "internal/experimental.go",
		URI:          "<#experimental.go>",
		Name:         "experimental.go",
		Description:  "Experimental features",
		Layer:        "internal",
		Tags:         []string{"experimental", "wip"},
		Dependencies: []string{},
		Exports:      []string{},
	}
	g.AddModule(internal)

	// Dead code: module that might use reflection
	plugin := &graph.Module{
		Path:         "plugins/old_plugin.go",
		URI:          "<#old_plugin.go>",
		Name:         "old_plugin.go",
		Description:  "Old plugin system",
		Layer:        "plugins",
		Tags:         []string{"plugin", "reflect"},
		Dependencies: []string{},
		Exports:      []string{"PluginInterface"},
	}
	g.AddModule(plugin)

	return g
}

func TestDetectDeadCode(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence:   0.5,
		ExcludePatterns: []string{},
		IncludeTests:    false,
		AggressiveMode:  false,
	}

	analysis, err := DetectDeadCode(g, opts)
	if err != nil {
		t.Fatalf("DetectDeadCode failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("Analysis is nil")
	}

	// Should find some dead code
	if len(analysis.UnreferencedModules) == 0 {
		t.Error("Expected to find unreferenced modules")
	}

	t.Logf("Found %d unreferenced modules", len(analysis.UnreferencedModules))
	for _, dm := range analysis.UnreferencedModules {
		t.Logf("  - %s (confidence: %.2f, safe: %v)", dm.Module.Path, dm.Confidence, dm.SafeToRemove)
	}
}

func TestDetector_FindUnreferencedModules(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence:  0.5,
		AggressiveMode: false,
	}

	detector := NewDetector(g, opts)
	deadModules := detector.findUnreferencedModules()

	if len(deadModules) == 0 {
		t.Error("Expected to find dead modules")
	}

	// Should not include main.go (entry point)
	for _, dm := range deadModules {
		if dm.Module.Path == "cmd/main.go" {
			t.Error("Main module should not be flagged as dead code")
		}
	}

	// Should not include referenced modules
	for _, dm := range deadModules {
		if dm.Module.Path == "services/auth.go" || dm.Module.Path == "utils/crypto.go" {
			t.Errorf("Referenced module %s should not be flagged as dead code", dm.Module.Path)
		}
	}
}

func TestDetector_IsEntryPoint(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{}
	detector := NewDetector(g, opts)

	tests := []struct {
		name     string
		path     string
		exports  []string
		expected bool
	}{
		{"Main package", "cmd/main.go", []string{"main"}, true},
		{"Service module", "services/auth.go", []string{"AuthService"}, false},
		{"Module with init", "pkg/config.go", []string{"init"}, true},
		{"Regular module", "utils/helper.go", []string{"Helper"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := &graph.Module{
				Path:    tt.path,
				Exports: tt.exports,
			}
			result := detector.isEntryPoint(module)
			if result != tt.expected {
				t.Errorf("isEntryPoint(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetector_IsTestFile(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{}
	detector := NewDetector(g, opts)

	tests := []struct {
		path     string
		expected bool
	}{
		{"main_test.go", true},
		{"auth_test.go", true},
		{"main.go", false},
		{"auth.go", false},
		{"test_helper.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detector.isTestFile(tt.path)
			if result != tt.expected {
				t.Errorf("isTestFile(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetector_IsInternalPackage(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{}
	detector := NewDetector(g, opts)

	tests := []struct {
		path     string
		expected bool
	}{
		{"internal/helper.go", true},
		{"pkg/internal/util.go", true},
		{"services/internal/impl.go", true},
		{"services/api.go", false},
		{"pkg/graph/graph.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detector.isInternalPackage(tt.path)
			if result != tt.expected {
				t.Errorf("isInternalPackage(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetector_MightBeReflectionUsed(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{}
	detector := NewDetector(g, opts)

	tests := []struct {
		name        string
		module      *graph.Module
		expected    bool
		description string
	}{
		{
			name: "Plugin with reflect tag",
			module: &graph.Module{
				Path:    "plugins/handler.go",
				Tags:    []string{"plugin", "reflect"},
				Exports: []string{"Handler"},
			},
			expected:    true,
			description: "Module with reflect tag should be flagged",
		},
		{
			name: "API handler",
			module: &graph.Module{
				Path:    "api/handler.go",
				Tags:    []string{"api", "handler"},
				Exports: []string{"HandleRequest"},
			},
			expected:    true,
			description: "Module with handler tag and exports should be flagged",
		},
		{
			name: "Internal utility",
			module: &graph.Module{
				Path:    "internal/util.go",
				Tags:    []string{"util"},
				Exports: []string{},
			},
			expected:    false,
			description: "Internal module without reflection tags should not be flagged",
		},
		{
			name: "Regular module",
			module: &graph.Module{
				Path:    "services/auth.go",
				Tags:    []string{"auth"},
				Exports: []string{"AuthService"},
			},
			expected:    true,
			description: "Exported module might use reflection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.mightBeReflectionUsed(tt.module)
			if result != tt.expected {
				t.Errorf("mightBeReflectionUsed(%s) = %v, want %v: %s",
					tt.module.Path, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDetector_IsExperimentalOrWIP(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{}
	detector := NewDetector(g, opts)

	tests := []struct {
		name     string
		module   *graph.Module
		expected bool
	}{
		{
			name: "Experimental tag",
			module: &graph.Module{
				Path: "features/new_feature.go",
				Tags: []string{"experimental"},
			},
			expected: true,
		},
		{
			name: "WIP in path",
			module: &graph.Module{
				Path: "wip/prototype.go",
				Tags: []string{},
			},
			expected: true,
		},
		{
			name: "Draft in description",
			module: &graph.Module{
				Path:        "features/auth.go",
				Description: "Draft implementation of auth",
			},
			expected: true,
		},
		{
			name: "Regular module",
			module: &graph.Module{
				Path:        "services/api.go",
				Description: "API service",
				Tags:        []string{"api"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isExperimentalOrWIP(tt.module)
			if result != tt.expected {
				t.Errorf("isExperimentalOrWIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetector_IsExcluded(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		ExcludePatterns: []string{
			"*_test.go",
			"experimental/*",
			"vendor/**",
		},
	}
	detector := NewDetector(g, opts)

	tests := []struct {
		path     string
		expected bool
	}{
		{"main_test.go", true},
		{"auth_test.go", true},
		{"experimental/feature.go", true},
		{"vendor/pkg/lib.go", false}, // Simple glob doesn't support **
		{"main.go", false},
		{"services/api.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detector.isExcluded(tt.path)
			if result != tt.expected {
				t.Errorf("isExcluded(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDeadCodeAnalysis_GetSafeRemovals(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence: 0.5,
	}

	analysis, err := DetectDeadCode(g, opts)
	if err != nil {
		t.Fatalf("DetectDeadCode failed: %v", err)
	}

	safeRemovals := analysis.GetSafeRemovals()

	for _, dm := range safeRemovals {
		if !dm.SafeToRemove {
			t.Errorf("Module %s in safe removals but SafeToRemove is false", dm.Module.Path)
		}
		if dm.Confidence < 0.8 {
			t.Errorf("Module %s in safe removals but confidence is only %.2f", dm.Module.Path, dm.Confidence)
		}
	}
}

func TestDeadCodeAnalysis_GetNeedsReview(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence: 0.3, // Lower threshold to catch more
	}

	analysis, err := DetectDeadCode(g, opts)
	if err != nil {
		t.Fatalf("DetectDeadCode failed: %v", err)
	}

	needsReview := analysis.GetNeedsReview()

	for _, dm := range needsReview {
		if dm.SafeToRemove {
			t.Errorf("Module %s in needs review but SafeToRemove is true", dm.Module.Path)
		}
	}
}

func TestDeadCodeAnalysis_HasDeadCode(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence: 0.5,
	}

	analysis, err := DetectDeadCode(g, opts)
	if err != nil {
		t.Fatalf("DetectDeadCode failed: %v", err)
	}

	if !analysis.HasDeadCode() {
		t.Error("Expected to find dead code")
	}
}

func TestDeadCodeOptions_Defaults(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{} // No options set

	detector := NewDetector(g, opts)

	if detector.options.MinConfidence != 0.5 {
		t.Errorf("Default MinConfidence = %.2f, want 0.5", detector.options.MinConfidence)
	}

	if detector.options.MaxFileAgeDays != 180 {
		t.Errorf("Default MaxFileAgeDays = %d, want 180", detector.options.MaxFileAgeDays)
	}
}

func TestAnalyzeCoverage(t *testing.T) {
	g := createTestGraphForDeadCode()
	analysis := AnalyzeCoverage(g)

	if analysis == nil {
		t.Fatal("Coverage analysis is nil")
	}

	if analysis.TotalModules != len(g.Modules) {
		t.Errorf("TotalModules = %d, want %d", analysis.TotalModules, len(g.Modules))
	}

	if analysis.ReferencedModules == 0 {
		t.Error("Expected some referenced modules")
	}

	if analysis.CoveragePercent < 0 || analysis.CoveragePercent > 100 {
		t.Errorf("CoveragePercent = %.2f, want 0-100", analysis.CoveragePercent)
	}

	t.Logf("Coverage: %.1f%% (%d/%d modules referenced)",
		analysis.CoveragePercent, analysis.ReferencedModules, analysis.TotalModules)
}

func TestCoverageAnalysis_GetUnusedModules(t *testing.T) {
	g := createTestGraphForDeadCode()
	analysis := AnalyzeCoverage(g)

	unused := analysis.GetUnusedModules()

	// Should find some unused modules (excluding entry points)
	if len(unused) == 0 {
		t.Error("Expected to find unused modules")
	}

	// Verify none are entry points
	for _, cov := range unused {
		if cov.IsEntryPoint {
			t.Errorf("Module %s is marked as entry point but in unused list", cov.Module.Path)
		}
		if cov.IncomingRefs > 0 {
			t.Errorf("Module %s has %d incoming refs but in unused list", cov.Module.Path, cov.IncomingRefs)
		}
	}
}

func TestGenerateCleanupPlan(t *testing.T) {
	g := createTestGraphForDeadCode()
	opts := DeadCodeOptions{
		MinConfidence: 0.5,
	}

	analysis, err := DetectDeadCode(g, opts)
	if err != nil {
		t.Fatalf("DetectDeadCode failed: %v", err)
	}

	plan := GenerateCleanupPlan(analysis)

	if plan == nil {
		t.Fatal("Cleanup plan is nil")
	}

	if !plan.HasActions() {
		t.Error("Expected cleanup plan to have actions")
	}

	if plan.Script == "" {
		t.Error("Expected cleanup script to be generated")
	}

	t.Logf("Cleanup plan: %s", plan.GetEstimatedImpact())
	t.Logf("Safe actions: %d, Review actions: %d", len(plan.SafeActions), len(plan.ReviewActions))
}
