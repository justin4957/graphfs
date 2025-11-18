/*
# Module: pkg/query/templates_test.go
Tests for query template system.

Tests template rendering, variable substitution, and template management.

## Linked Modules
- [templates](./templates.go) - Query templates

## Tags
query, templates, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#templates_test.go> a code:Module ;
    code:name "pkg/query/templates_test.go" ;
    code:description "Tests for query template system" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./templates.go> ;
    code:tags "query", "templates", "test" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltInTemplates(t *testing.T) {
	// Verify we have the expected number of templates
	if len(BuiltInTemplates) < 15 {
		t.Errorf("Expected at least 15 built-in templates, got %d", len(BuiltInTemplates))
	}

	// Verify all templates have required fields
	for _, tmpl := range BuiltInTemplates {
		if tmpl.Name == "" {
			t.Errorf("Template missing name")
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s missing description", tmpl.Name)
		}
		if tmpl.Category == "" {
			t.Errorf("Template %s missing category", tmpl.Name)
		}
		if tmpl.Query == "" {
			t.Errorf("Template %s missing query", tmpl.Name)
		}
	}
}

func TestTemplateManagerGetTemplate(t *testing.T) {
	tm := NewTemplateManager("")

	// Test getting existing template
	tmpl, err := tm.GetTemplate("find-dependencies")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}
	if tmpl.Name != "find-dependencies" {
		t.Errorf("Expected template name 'find-dependencies', got '%s'", tmpl.Name)
	}

	// Test getting non-existent template
	_, err = tm.GetTemplate("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent template")
	}
}

func TestTemplateManagerListTemplates(t *testing.T) {
	tm := NewTemplateManager("")

	// Test listing all templates
	templates := tm.ListTemplates("")
	if len(templates) < 15 {
		t.Errorf("Expected at least 15 templates, got %d", len(templates))
	}

	// Test filtering by category
	depTemplates := tm.ListTemplates("dependencies")
	if len(depTemplates) == 0 {
		t.Errorf("Expected dependency templates, got none")
	}
	for _, tmpl := range depTemplates {
		if tmpl.Category != "dependencies" {
			t.Errorf("Expected category 'dependencies', got '%s'", tmpl.Category)
		}
	}
}

func TestTemplateManagerGetCategories(t *testing.T) {
	tm := NewTemplateManager("")

	categories := tm.GetCategories()
	if len(categories) == 0 {
		t.Errorf("Expected categories, got none")
	}

	// Check for expected categories
	expectedCategories := []string{"dependencies", "security", "analysis", "layers", "impact", "documentation"}
	foundCategories := make(map[string]bool)
	for _, cat := range categories {
		foundCategories[cat] = true
	}

	for _, expected := range expectedCategories {
		if !foundCategories[expected] {
			t.Errorf("Expected category '%s' not found", expected)
		}
	}
}

func TestTemplateRender(t *testing.T) {
	tm := NewTemplateManager("")

	tests := []struct {
		name          string
		templateName  string
		variables     map[string]string
		expectError   bool
		expectedQuery string
	}{
		{
			name:         "find-dependencies with module",
			templateName: "find-dependencies",
			variables: map[string]string{
				"module": "api/handlers.go",
			},
			expectError: false,
			expectedQuery: `SELECT ?dep WHERE {
    <#api/handlers.go> <#imports> ?dep .
}`,
		},
		{
			name:         "find-dependencies with default",
			templateName: "find-dependencies",
			variables:    map[string]string{},
			expectError:  false,
			expectedQuery: `SELECT ?dep WHERE {
    <#api/handlers.go> <#imports> ?dep .
}`,
		},
		{
			name:         "circular-deps no variables",
			templateName: "circular-deps",
			variables:    map[string]string{},
			expectError:  false,
			expectedQuery: `SELECT ?a ?b WHERE {
    ?a <#imports> ?b .
    ?b <#imports> ?a .
}`,
		},
		{
			name:         "complex-modules with threshold",
			templateName: "complex-modules",
			variables: map[string]string{
				"threshold": "5",
			},
			expectError: false,
			expectedQuery: `SELECT ?module (COUNT(?dep) as ?depCount) WHERE {
    ?module <#imports> ?dep .
}
GROUP BY ?module
HAVING (COUNT(?dep) > 5)
ORDER BY DESC(?depCount)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := tm.RenderTemplate(tt.templateName, tt.variables)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if strings.TrimSpace(rendered) != strings.TrimSpace(tt.expectedQuery) {
					t.Errorf("Query mismatch.\nExpected:\n%s\n\nGot:\n%s", tt.expectedQuery, rendered)
				}
			}
		})
	}
}

func TestTemplateRenderMissingVariable(t *testing.T) {
	tm := NewTemplateManager("")

	// Test with missing required variable
	_, err := tm.RenderTemplate("find-usages", map[string]string{})
	if err == nil {
		t.Errorf("Expected error for missing required variable")
	}
}

func TestTemplateManagerSaveCustomTemplate(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	// Create custom template
	customTemplate := &QueryTemplate{
		Name:        "test-template",
		Description: "Test template",
		Category:    "test",
		Query: `SELECT ?s WHERE {
    ?s <#test> ?o .
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run test-template",
	}

	// Save template
	err := tm.SaveCustomTemplate(customTemplate)
	if err != nil {
		t.Fatalf("Failed to save custom template: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "test-template.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Template file not created")
	}

	// Verify template can be retrieved
	tmpl, err := tm.GetTemplate("test-template")
	if err != nil {
		t.Fatalf("Failed to get saved template: %v", err)
	}
	if tmpl.Name != "test-template" {
		t.Errorf("Expected template name 'test-template', got '%s'", tmpl.Name)
	}
}

func TestTemplateManagerLoadCustomTemplates(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create a custom template file
	templateJSON := `{
  "name": "custom-query",
  "description": "Custom query template",
  "category": "custom",
  "query": "SELECT ?s WHERE { ?s ?p ?o }",
  "variables": [],
  "example": "graphfs examples run custom-query"
}`
	filePath := filepath.Join(tempDir, "custom-query.json")
	if err := os.WriteFile(filePath, []byte(templateJSON), 0644); err != nil {
		t.Fatalf("Failed to create test template file: %v", err)
	}

	// Create template manager (should load custom templates)
	tm := NewTemplateManager(tempDir)

	// Verify custom template is loaded
	tmpl, err := tm.GetTemplate("custom-query")
	if err != nil {
		t.Fatalf("Failed to get custom template: %v", err)
	}
	if tmpl.Name != "custom-query" {
		t.Errorf("Expected template name 'custom-query', got '%s'", tmpl.Name)
	}
}

func TestTemplateManagerExportTemplate(t *testing.T) {
	tm := NewTemplateManager("")

	// Create temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-query.sparql")

	// Export template
	variables := map[string]string{
		"module": "test/module.go",
	}
	err := tm.ExportTemplate("find-dependencies", outputPath, variables)
	if err != nil {
		t.Fatalf("Failed to export template: %v", err)
	}

	// Verify file exists and contains expected content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	expectedContent := `SELECT ?dep WHERE {
    <#test/module.go> <#imports> ?dep .
}`
	if strings.TrimSpace(string(content)) != strings.TrimSpace(expectedContent) {
		t.Errorf("Exported content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(content))
	}
}

func TestTemplateVariableDefaults(t *testing.T) {
	tm := NewTemplateManager("")

	// Get template with default variables
	tmpl, err := tm.GetTemplate("complex-modules")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	// Find the threshold variable
	var thresholdVar *Variable
	for i := range tmpl.Variables {
		if tmpl.Variables[i].Name == "threshold" {
			thresholdVar = &tmpl.Variables[i]
			break
		}
	}

	if thresholdVar == nil {
		t.Fatalf("Template 'complex-modules' missing 'threshold' variable")
	}

	if thresholdVar.Default != "10" {
		t.Errorf("Expected default value '10', got '%s'", thresholdVar.Default)
	}

	// Render with default
	rendered, err := tm.Render(tmpl, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to render with default: %v", err)
	}

	if !strings.Contains(rendered, "10") {
		t.Errorf("Expected rendered query to contain default value '10'")
	}
}

func TestTemplateCategories(t *testing.T) {
	expectedCategories := map[string]int{
		"dependencies":  0,
		"security":      0,
		"analysis":      0,
		"layers":        0,
		"impact":        0,
		"documentation": 0,
	}

	// Count templates per category
	for _, tmpl := range BuiltInTemplates {
		if _, ok := expectedCategories[tmpl.Category]; ok {
			expectedCategories[tmpl.Category]++
		} else {
			t.Errorf("Unexpected category: %s", tmpl.Category)
		}
	}

	// Verify each category has at least one template
	for cat, count := range expectedCategories {
		if count == 0 {
			t.Errorf("Category '%s' has no templates", cat)
		}
	}
}
