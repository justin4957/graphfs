package rules

import (
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test", tripleStore)

	// Module with all metadata
	main := &graph.Module{
		Path:         "main.go",
		URI:          "<#main.go>",
		Name:         "main.go",
		Description:  "Main entry point",
		Layer:        "main",
		Tags:         []string{"entrypoint"},
		Dependencies: []string{"services/auth.go"},
		Exports:      []string{"main"},
	}
	g.AddModule(main)

	// Add RDF triples for main.go
	g.Store.Add("<#main.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	g.Store.Add("<#main.go>", "https://schema.codedoc.org/layer", "main")
	g.Store.Add("<#main.go>", "https://schema.codedoc.org/description", "Main entry point")
	g.Store.Add("<#main.go>", "https://schema.codedoc.org/exports", "main")
	g.Store.Add("<#main.go>", "https://schema.codedoc.org/tags", "entrypoint")

	// Module with exports but no description
	auth := &graph.Module{
		Path:         "services/auth.go",
		URI:          "<#auth.go>",
		Name:         "services/auth.go",
		Layer:        "service",
		Tags:         []string{"security"},
		Dependencies: []string{},
		Exports:      []string{"AuthService"},
	}
	g.AddModule(auth)

	// Add RDF triples for auth.go (no description!)
	g.Store.Add("<#auth.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	g.Store.Add("<#auth.go>", "https://schema.codedoc.org/layer", "service")
	g.Store.Add("<#auth.go>", "https://schema.codedoc.org/exports", "AuthService")

	// Module with no layer
	helper := &graph.Module{
		Path:        "utils/helper.go",
		URI:         "<#helper.go>",
		Name:        "utils/helper.go",
		Description: "Helper functions",
		Tags:        []string{"utility"},
		Exports:     []string{"HelperFunc"},
	}
	g.AddModule(helper)

	// Add RDF triples for helper.go (no layer!)
	g.Store.Add("<#helper.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	g.Store.Add("<#helper.go>", "https://schema.codedoc.org/description", "Helper functions")
	g.Store.Add("<#helper.go>", "https://schema.codedoc.org/exports", "HelperFunc")

	// Module with no tags
	user := &graph.Module{
		Path:        "models/user.go",
		URI:         "<#user.go>",
		Name:        "models/user.go",
		Description: "User model",
		Layer:       "model",
	}
	g.AddModule(user)

	// Add RDF triples for user.go (no tags, no exports)
	g.Store.Add("<#user.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	g.Store.Add("<#user.go>", "https://schema.codedoc.org/layer", "model")
	g.Store.Add("<#user.go>", "https://schema.codedoc.org/description", "User model")

	return g
}

func TestNewEngine(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if engine.graph != g {
		t.Error("Graph not set correctly")
	}

	if engine.parser == nil {
		t.Error("Parser not initialized")
	}

	if engine.evaluator == nil {
		t.Error("Evaluator not initialized")
	}
}

func TestParser_Parse(t *testing.T) {
	yaml := `
version: "1.0"
name: "Test Rules"
rules:
  - id: test-rule
    name: "Test Rule"
    description: "A test rule"
    severity: error
    pattern: "SELECT ?module WHERE { ?module a code:Module }"
    expect: 0
    enabled: true
    suggestion: "Fix it"
`

	parser := NewParser()
	ruleSet, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Failed to parse rules: %v", err)
	}

	if ruleSet.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", ruleSet.Version)
	}

	if len(ruleSet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(ruleSet.Rules))
	}

	rule := ruleSet.Rules[0]
	if rule.ID != "test-rule" {
		t.Errorf("Expected ID 'test-rule', got %s", rule.ID)
	}

	if rule.Severity != SeverityError {
		t.Errorf("Expected severity error, got %s", rule.Severity)
	}

	if !rule.Enabled {
		t.Error("Expected rule to be enabled")
	}
}

func TestParser_ValidateMissingFields(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name: "missing version",
			yaml: `
rules:
  - id: test
    name: Test
    severity: error
    pattern: SELECT
`,
			expected: "missing version",
		},
		{
			name: "no rules",
			yaml: `
version: "1.0"
rules: []
`,
			expected: "no rules defined",
		},
		{
			name: "missing rule ID",
			yaml: `
version: "1.0"
rules:
  - name: Test
    severity: error
    pattern: SELECT
`,
			expected: "missing ID",
		},
		{
			name: "missing pattern",
			yaml: `
version: "1.0"
rules:
  - id: test
    name: Test
    severity: error
`,
			expected: "missing pattern",
		},
		{
			name: "invalid severity",
			yaml: `
version: "1.0"
rules:
  - id: test
    name: Test
    severity: critical
    pattern: SELECT
`,
			expected: "invalid severity",
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tt.yaml))
			if err == nil {
				t.Fatal("Expected error but got none")
			}

			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestValidationResult_Success(t *testing.T) {
	result := &ValidationResult{
		ErrorCount:   0,
		WarningCount: 2,
	}

	if !result.Success() {
		t.Error("Expected success when no errors")
	}

	result.ErrorCount = 1
	if result.Success() {
		t.Error("Expected failure when errors present")
	}
}

func TestValidationResult_GetViolationsBySeverity(t *testing.T) {
	rule1 := &Rule{ID: "r1", Severity: SeverityError}
	rule2 := &Rule{ID: "r2", Severity: SeverityWarning}

	result := &ValidationResult{
		Violations: []Violation{
			{Rule: rule1, Message: "Error 1"},
			{Rule: rule2, Message: "Warning 1"},
			{Rule: rule1, Message: "Error 2"},
		},
	}

	errors := result.GetViolationsBySeverity(SeverityError)
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	warnings := result.GetViolationsBySeverity(SeverityWarning)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}
}

func TestValidationResult_GetViolationsByRule(t *testing.T) {
	rule1 := &Rule{ID: "r1"}
	rule2 := &Rule{ID: "r2"}

	result := &ValidationResult{
		Violations: []Violation{
			{Rule: rule1, Message: "V1"},
			{Rule: rule2, Message: "V2"},
			{Rule: rule1, Message: "V3"},
		},
	}

	grouped := result.GetViolationsByRule()
	if len(grouped) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(grouped))
	}

	if len(grouped["r1"]) != 2 {
		t.Errorf("Expected 2 violations for r1, got %d", len(grouped["r1"]))
	}

	if len(grouped["r2"]) != 1 {
		t.Errorf("Expected 1 violation for r2, got %d", len(grouped["r2"]))
	}
}

func TestGetBuiltInRules(t *testing.T) {
	rules := GetBuiltInRules()

	if len(rules) == 0 {
		t.Fatal("Expected built-in rules")
	}

	// Check that all built-in rules are valid
	for _, rule := range rules {
		if rule.ID == "" {
			t.Error("Built-in rule missing ID")
		}
		if rule.Name == "" {
			t.Error("Built-in rule missing name")
		}
		if rule.Pattern == "" {
			t.Error("Built-in rule missing pattern")
		}
		if rule.Severity == "" {
			t.Error("Built-in rule missing severity")
		}
	}
}

func TestReporter_FormatText(t *testing.T) {
	rule := &Rule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Severity: SeverityError,
	}

	result := &ValidationResult{
		TotalRules:   2,
		PassedRules:  []*Rule{{ID: "passed", Name: "Passed Rule"}},
		FailedRules:  []*Rule{rule},
		ErrorCount:   1,
		WarningCount: 0,
		Duration:     150,
		Violations: []Violation{
			{
				Rule:     rule,
				Message:  "Test violation",
				FilePath: "test.go",
			},
		},
	}

	reporter := NewReporter(FormatText)
	output := reporter.Report(result)

	if !strings.Contains(output, "Test Rule") {
		t.Error("Output should contain rule name")
	}

	if !strings.Contains(output, "Test violation") {
		t.Error("Output should contain violation message")
	}

	if !strings.Contains(output, "test.go") {
		t.Error("Output should contain file path")
	}

	if !strings.Contains(output, "FAILED") {
		t.Error("Output should indicate failure")
	}
}

func TestReporter_FormatJSON(t *testing.T) {
	rule := &Rule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Severity: SeverityError,
	}

	result := &ValidationResult{
		TotalRules:   1,
		ErrorCount:   1,
		WarningCount: 0,
		Duration:     100,
		Violations: []Violation{
			{
				Rule:     rule,
				Message:  "Test violation",
				FilePath: "test.go",
			},
		},
	}

	reporter := NewReporter(FormatJSON)
	output := reporter.Report(result)

	if !strings.Contains(output, `"rule_id": "test-rule"`) {
		t.Error("JSON should contain rule ID")
	}

	if !strings.Contains(output, `"error_count": 1`) {
		t.Error("JSON should contain error count")
	}

	if !strings.Contains(output, `"success": false`) {
		t.Error("JSON should indicate failure")
	}
}

func TestReporter_FormatJUnit(t *testing.T) {
	rule := &Rule{
		ID:       "test-rule",
		Name:     "Test Rule",
		Severity: SeverityError,
	}

	result := &ValidationResult{
		TotalRules:  1,
		FailedRules: []*Rule{rule},
		ErrorCount:  1,
		Duration:    100,
		Violations: []Violation{
			{
				Rule:     rule,
				Message:  "Test violation",
				FilePath: "test.go",
			},
		},
	}

	reporter := NewReporter(FormatJUnit)
	output := reporter.Report(result)

	if !strings.Contains(output, `<testsuite`) {
		t.Error("JUnit XML should contain testsuite")
	}

	if !strings.Contains(output, `<testcase`) {
		t.Error("JUnit XML should contain testcase")
	}

	if !strings.Contains(output, `<failure`) {
		t.Error("JUnit XML should contain failure")
	}

	if !strings.Contains(output, "Test Rule") {
		t.Error("JUnit XML should contain rule name")
	}
}

func TestEngine_Validate_PassingRules(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	// Rule that should pass - looking for non-existent pattern
	rules := []*Rule{
		{
			ID:       "test-pass",
			Name:     "Test Passing Rule",
			Severity: SeverityError,
			Pattern: `
				PREFIX code: <https://schema.codedoc.org/>
				SELECT ?module WHERE {
					?module code:layer "nonexistent" .
				}
			`,
			Expect:  0,
			Enabled: true,
		},
	}

	result, err := engine.Validate(rules)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Success() {
		t.Error("Expected validation to pass")
	}

	if len(result.PassedRules) != 1 {
		t.Errorf("Expected 1 passed rule, got %d", len(result.PassedRules))
	}

	if len(result.FailedRules) != 0 {
		t.Errorf("Expected 0 failed rules, got %d", len(result.FailedRules))
	}
}

func TestEngine_Validate_FailingRules(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	// Rule that should fail - we expect to find 4 modules but we expect 0
	rules := []*Rule{
		{
			ID:       "test-fail",
			Name:     "Test Failing Rule",
			Severity: SeverityError,
			Pattern: `
				PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
				PREFIX code: <https://schema.codedoc.org/>
				SELECT ?module WHERE {
					?module rdf:type code:Module .
				}
			`,
			Expect:  0, // We expect 0 but will get 4 - this should fail
			Enabled: true,
		},
	}

	result, err := engine.Validate(rules)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Success() {
		t.Errorf("Expected validation to fail, but it passed. Violations: %d, Passed: %d", len(result.Violations), len(result.PassedRules))
	}

	if len(result.FailedRules) != 1 {
		t.Errorf("Expected 1 failed rule, got %d. Violations: %d", len(result.FailedRules), len(result.Violations))
	}

	if result.ErrorCount == 0 {
		t.Error("Expected error count > 0")
	}
}

func TestEngine_ValidateWithFilter_Severity(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	rules := []*Rule{
		{
			ID:       "error-rule",
			Name:     "Error Rule",
			Severity: SeverityError,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   999, // Will fail
			Enabled:  true,
		},
		{
			ID:       "warning-rule",
			Name:     "Warning Rule",
			Severity: SeverityWarning,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   999, // Will fail
			Enabled:  true,
		},
		{
			ID:       "info-rule",
			Name:     "Info Rule",
			Severity: SeverityInfo,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   999, // Will fail
			Enabled:  true,
		},
	}

	// Filter for only errors
	result, err := engine.ValidateWithFilter(rules, nil, SeverityError)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.TotalRules != 1 {
		t.Errorf("Expected 1 rule with error severity, got %d", result.TotalRules)
	}
}

func TestEngine_ValidateWithFilter_Tags(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	rules := []*Rule{
		{
			ID:       "security-rule",
			Name:     "Security Rule",
			Severity: SeverityError,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   4, // We have 4 modules
			Enabled:  true,
			Tags:     []string{"security"},
		},
		{
			ID:       "performance-rule",
			Name:     "Performance Rule",
			Severity: SeverityError,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   4, // We have 4 modules
			Enabled:  true,
			Tags:     []string{"performance"},
		},
	}

	// Filter for only security rules
	result, err := engine.ValidateWithFilter(rules, []string{"security"}, SeverityInfo)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.TotalRules != 1 {
		t.Errorf("Expected 1 security rule, got %d", result.TotalRules)
	}
}

func TestEngine_SkippedRules(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	rules := []*Rule{
		{
			ID:       "enabled-rule",
			Name:     "Enabled Rule",
			Severity: SeverityError,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   4, // We have 4 modules
			Enabled:  true,
		},
		{
			ID:       "disabled-rule",
			Name:     "Disabled Rule",
			Severity: SeverityError,
			Pattern:  "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>\nPREFIX code: <https://schema.codedoc.org/>\nSELECT ?m WHERE { ?m rdf:type code:Module }",
			Expect:   4, // We have 4 modules
			Enabled:  false,
		},
	}

	result, err := engine.Validate(rules)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if len(result.SkippedRules) != 1 {
		t.Errorf("Expected 1 skipped rule, got %d", len(result.SkippedRules))
	}

	if result.TotalRules != 1 {
		t.Errorf("Expected 1 total rule (excluding skipped), got %d", result.TotalRules)
	}
}
