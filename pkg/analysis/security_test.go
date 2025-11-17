package analysis

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraphForSecurity() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test", tripleStore)

	// Public API module
	publicAPI := &graph.Module{
		Path:         "api/handlers/public.go",
		URI:          "<#public.go>",
		Name:         "public.go",
		Description:  "Public API handlers",
		Layer:        "api",
		Tags:         []string{"public", "api", "http"},
		Dependencies: []string{"services/auth.go", "admin/users.go"},
		Exports:      []string{"HandlePublicRequest"},
	}
	g.AddModule(publicAPI)

	// Auth service (trusted)
	authService := &graph.Module{
		Path:         "services/auth.go",
		URI:          "<#auth.go>",
		Name:         "auth.go",
		Description:  "Authentication service",
		Layer:        "services",
		Tags:         []string{"auth", "service"},
		Dependencies: []string{"internal/store/users.go"},
		Exports:      []string{"AuthService"},
	}
	g.AddModule(authService)

	// Admin module
	adminUsers := &graph.Module{
		Path:         "admin/users.go",
		URI:          "<#admin-users.go>",
		Name:         "users.go",
		Description:  "User administration",
		Layer:        "admin",
		Tags:         []string{"admin", "privileged"},
		Dependencies: []string{"internal/store/users.go"},
		Exports:      []string{"AdminUserService"},
	}
	g.AddModule(adminUsers)

	// Data layer module
	userStore := &graph.Module{
		Path:         "internal/store/users.go",
		URI:          "<#users-store.go>",
		Name:         "users.go",
		Description:  "User data storage",
		Layer:        "data",
		Tags:         []string{"database", "storage"},
		Dependencies: []string{},
		Exports:      []string{"UserStore"},
	}
	g.AddModule(userStore)

	return g
}

func TestClassifyZones(t *testing.T) {
	g := createTestGraphForSecurity()
	zones := ClassifyZones(g)

	if len(zones) == 0 {
		t.Fatal("Expected zones to be classified")
	}

	// Check that we have modules in different zones
	hasPublic := len(zones[ZonePublic]) > 0
	hasAdmin := len(zones[ZoneAdmin]) > 0
	hasData := len(zones[ZoneData]) > 0

	if !hasPublic {
		t.Error("Expected public zone modules")
	}
	if !hasAdmin {
		t.Error("Expected admin zone modules")
	}
	if !hasData {
		t.Error("Expected data zone modules")
	}

	t.Logf("Zones classified: %d zones", len(zones))
	for zone, modules := range zones {
		t.Logf("  %s: %d modules", zone, len(modules))
	}
}

func TestSecurityAnalysis(t *testing.T) {
	g := createTestGraphForSecurity()
	opts := SecurityOptions{
		StrictMode: false,
	}

	analysis, err := AnalyzeSecurity(g, opts)
	if err != nil {
		t.Fatalf("Security analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("Analysis is nil")
	}

	// Should detect the public->admin crossing as violation
	if !analysis.HasViolations() {
		t.Error("Expected to find security violations (public->admin crossing)")
	}

	t.Logf("Security Analysis Results:")
	t.Logf("  Risk Score: %.1f/10.0", analysis.RiskScore)
	t.Logf("  Violations: %d", len(analysis.Violations))
	t.Logf("  Boundaries: %d", len(analysis.Boundaries))
}

func TestDetectViolations(t *testing.T) {
	g := createTestGraphForSecurity()
	opts := SecurityOptions{
		StrictMode: false,
	}

	analysis, err := AnalyzeSecurity(g, opts)
	if err != nil {
		t.Fatalf("Security analysis failed: %v", err)
	}

	// Should have at least one violation (public->admin)
	criticalViolations := analysis.GetCriticalViolations()
	if len(criticalViolations) == 0 {
		t.Error("Expected to find critical violations for public->admin crossing")
	}

	for _, v := range criticalViolations {
		t.Logf("Critical Violation: %s", v.Description)
		t.Logf("  Risk: %s", v.Risk)
		t.Logf("  Recommendation: %s", v.Recommendation)
	}
}

func TestZoneClassification(t *testing.T) {
	g := createTestGraphForSecurity()
	classifier := NewZoneClassifier(g)

	// Test public API classification
	publicModule := g.GetModule("api/handlers/public.go")
	if publicModule == nil {
		t.Fatal("Public module not found")
	}

	mz := classifier.ClassifyModule(publicModule)
	if mz.Zone != ZonePublic {
		t.Errorf("Expected public zone, got %s", mz.Zone)
	}
	if mz.Confidence < 0.7 {
		t.Errorf("Expected high confidence (>0.7), got %.2f", mz.Confidence)
	}

	// Test admin classification
	adminModule := g.GetModule("admin/users.go")
	if adminModule != nil {
		mz = classifier.ClassifyModule(adminModule)
		if mz.Zone != ZoneAdmin {
			t.Errorf("Expected admin zone, got %s", mz.Zone)
		}
	}

	// Test data layer classification
	dataModule := g.GetModule("internal/store/users.go")
	if dataModule != nil {
		mz = classifier.ClassifyModule(dataModule)
		if mz.Zone != ZoneData {
			t.Errorf("Expected data zone, got %s", mz.Zone)
		}
	}
}

func TestRiskAssessment(t *testing.T) {
	g := createTestGraphForSecurity()
	analyzer := NewSecurityAnalyzer(g, SecurityOptions{})

	tests := []struct {
		from     SecurityZone
		to       SecurityZone
		expected RiskLevel
	}{
		{ZonePublic, ZoneAdmin, RiskLevelCritical},
		{ZonePublic, ZoneData, RiskLevelCritical},
		{ZonePublic, ZoneInternal, RiskLevelHigh},
		{ZonePublic, ZoneTrusted, RiskLevelLow},
		{ZoneTrusted, ZoneData, RiskLevelLow},
	}

	for _, tt := range tests {
		risk := analyzer.assessCrossingRisk(tt.from, tt.to)
		if risk != tt.expected {
			t.Errorf("Risk assessment for %s->%s: expected %s, got %s",
				tt.from, tt.to, tt.expected, risk)
		}
	}
}

func TestAllowedCrossings(t *testing.T) {
	g := createTestGraphForSecurity()
	analyzer := NewSecurityAnalyzer(g, SecurityOptions{})

	tests := []struct {
		from     SecurityZone
		to       SecurityZone
		expected bool
	}{
		{ZonePublic, ZoneTrusted, true},
		{ZonePublic, ZoneInternal, true}, // Updated: now allowed
		{ZonePublic, ZoneAdmin, false},
		{ZonePublic, ZoneData, false},
		{ZoneTrusted, ZoneData, true},
		{ZoneTrusted, ZoneInternal, true},
		{ZoneInternal, ZoneData, true},   // Updated: now allowed
		{ZoneInternal, ZonePublic, true}, // Updated: now allowed
		{ZoneData, ZonePublic, false},
	}

	for _, tt := range tests {
		allowed := analyzer.isCrossingAllowed(tt.from, tt.to)
		if allowed != tt.expected {
			t.Errorf("Crossing %s->%s: expected allowed=%v, got %v",
				tt.from, tt.to, tt.expected, allowed)
		}
	}
}

func TestStrictMode(t *testing.T) {
	g := createTestGraphForSecurity()

	// Test with strict mode enabled
	strictOpts := SecurityOptions{
		StrictMode: true,
	}

	strictAnalysis, err := AnalyzeSecurity(g, strictOpts)
	if err != nil {
		t.Fatalf("Strict mode analysis failed: %v", err)
	}

	// Test without strict mode
	normalOpts := SecurityOptions{
		StrictMode: false,
	}

	normalAnalysis, err := AnalyzeSecurity(g, normalOpts)
	if err != nil {
		t.Fatalf("Normal mode analysis failed: %v", err)
	}

	// Strict mode should find at least as many violations as normal mode
	// (may be equal if no allowed high-risk crossings exist)
	if len(strictAnalysis.Violations) < len(normalAnalysis.Violations) {
		t.Error("Expected strict mode to find at least as many violations as normal mode")
	}

	t.Logf("Normal mode violations: %d", len(normalAnalysis.Violations))
	t.Logf("Strict mode violations: %d", len(strictAnalysis.Violations))
}
