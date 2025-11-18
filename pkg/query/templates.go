/*
# Module: pkg/query/templates.go
Query template system for common SPARQL patterns.

Provides pre-built query templates with variable substitution
to help users run common queries without knowing SPARQL syntax.

## Linked Modules
- [query](./query.go) - Query data structures
- [executor](./executor.go) - Query executor

## Tags
query, templates, sparql, examples

## Exports
QueryTemplate, Variable, BuiltInTemplates, TemplateManager

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#templates.go> a code:Module ;
    code:name "pkg/query/templates.go" ;
    code:description "Query template system for common SPARQL patterns" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./query.go>, <./executor.go> ;
    code:exports <#QueryTemplate>, <#Variable>, <#BuiltInTemplates>, <#TemplateManager> ;
    code:tags "query", "templates", "sparql", "examples" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// QueryTemplate represents a parameterized SPARQL query template
type QueryTemplate struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Query       string     `json:"query"`
	Variables   []Variable `json:"variables"`
	Example     string     `json:"example"`
}

// Variable represents a template variable that can be substituted
type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// BuiltInTemplates contains all predefined query templates
var BuiltInTemplates = []QueryTemplate{
	{
		Name:        "find-dependencies",
		Description: "Find all dependencies of a module",
		Category:    "dependencies",
		Query: `SELECT ?dep WHERE {
    <#{{.module}}> <#imports> ?dep .
}`,
		Variables: []Variable{
			{Name: "module", Description: "Module path", Default: "api/handlers.go"},
		},
		Example: "graphfs examples run find-dependencies --module=api/handlers.go",
	},
	{
		Name:        "find-usages",
		Description: "Find all modules that import a given module",
		Category:    "dependencies",
		Query: `SELECT ?user WHERE {
    ?user <#imports> <#{{.module}}> .
}`,
		Variables: []Variable{
			{Name: "module", Description: "Module path"},
		},
		Example: "graphfs examples run find-usages --module=pkg/graph/graph.go",
	},
	{
		Name:        "circular-deps",
		Description: "Find circular dependencies",
		Category:    "dependencies",
		Query: `SELECT ?a ?b WHERE {
    ?a <#imports> ?b .
    ?b <#imports> ?a .
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run circular-deps",
	},
	{
		Name:        "dependency-tree",
		Description: "Find all transitive dependencies of a module",
		Category:    "dependencies",
		Query: `SELECT ?dep WHERE {
    <#{{.module}}> <#imports>+ ?dep .
}`,
		Variables: []Variable{
			{Name: "module", Description: "Module path"},
		},
		Example: "graphfs examples run dependency-tree --module=cmd/graphfs/main.go",
	},
	{
		Name:        "security-violations",
		Description: "Find security zone boundary violations",
		Category:    "security",
		Query: `SELECT ?from ?to ?fromZone ?toZone WHERE {
    ?from <#imports> ?to .
    ?from <#zone> ?fromZone .
    ?to <#zone> ?toZone .
    FILTER(?fromZone != ?toZone && ?toZone = "trusted")
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run security-violations",
	},
	{
		Name:        "zone-crossings",
		Description: "Find all imports that cross security zones",
		Category:    "security",
		Query: `SELECT ?from ?to ?fromZone ?toZone WHERE {
    ?from <#imports> ?to .
    ?from <#zone> ?fromZone .
    ?to <#zone> ?toZone .
    FILTER(?fromZone != ?toZone)
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run zone-crossings",
	},
	{
		Name:        "trust-boundaries",
		Description: "Find modules in trusted zones and their importers",
		Category:    "security",
		Query: `SELECT ?trusted ?importer WHERE {
    ?trusted <#zone> "trusted" .
    ?importer <#imports> ?trusted .
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run trust-boundaries",
	},
	{
		Name:        "unused-exports",
		Description: "Find exported functions that are never used",
		Category:    "analysis",
		Query: `SELECT ?module ?export WHERE {
    ?module <#exports> ?export .
    FILTER NOT EXISTS {
        ?other <#imports> ?module .
    }
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run unused-exports",
	},
	{
		Name:        "dead-code",
		Description: "Find unreachable code modules",
		Category:    "analysis",
		Query: `SELECT ?module WHERE {
    ?module a <#Module> .
    FILTER NOT EXISTS {
        ?other <#imports> ?module .
    }
    FILTER(?module != <#main>)
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run dead-code",
	},
	{
		Name:        "complex-modules",
		Description: "Find modules with many dependencies",
		Category:    "analysis",
		Query: `SELECT ?module (COUNT(?dep) as ?depCount) WHERE {
    ?module <#imports> ?dep .
}
GROUP BY ?module
HAVING (COUNT(?dep) > {{.threshold}})
ORDER BY DESC(?depCount)`,
		Variables: []Variable{
			{Name: "threshold", Description: "Minimum number of dependencies", Default: "10"},
		},
		Example: "graphfs examples run complex-modules --threshold=10",
	},
	{
		Name:        "hot-paths",
		Description: "Find most frequently imported modules",
		Category:    "analysis",
		Query: `SELECT ?module (COUNT(?importer) as ?importCount) WHERE {
    ?importer <#imports> ?module .
}
GROUP BY ?module
ORDER BY DESC(?importCount)
LIMIT {{.limit}}`,
		Variables: []Variable{
			{Name: "limit", Description: "Number of top modules to show", Default: "10"},
		},
		Example: "graphfs examples run hot-paths --limit=20",
	},
	{
		Name:        "layer-violations",
		Description: "Find violations of layered architecture",
		Category:    "layers",
		Query: `SELECT ?from ?to ?fromLayer ?toLayer WHERE {
    ?from <#imports> ?to .
    ?from <#layer> ?fromLayer .
    ?to <#layer> ?toLayer .
    FILTER(?fromLayer = "{{.lowerLayer}}" && ?toLayer = "{{.upperLayer}}")
}`,
		Variables: []Variable{
			{Name: "lowerLayer", Description: "Lower layer name", Default: "data"},
			{Name: "upperLayer", Description: "Upper layer name", Default: "cli"},
		},
		Example: "graphfs examples run layer-violations --lowerLayer=data --upperLayer=cli",
	},
	{
		Name:        "layer-dependencies",
		Description: "Show dependencies between layers",
		Category:    "layers",
		Query: `SELECT ?fromLayer ?toLayer (COUNT(*) as ?count) WHERE {
    ?from <#imports> ?to .
    ?from <#layer> ?fromLayer .
    ?to <#layer> ?toLayer .
    FILTER(?fromLayer != ?toLayer)
}
GROUP BY ?fromLayer ?toLayer
ORDER BY DESC(?count)`,
		Variables: []Variable{},
		Example:   "graphfs examples run layer-dependencies",
	},
	{
		Name:        "layer-stats",
		Description: "Statistics for each architectural layer",
		Category:    "layers",
		Query: `SELECT ?layer (COUNT(?module) as ?moduleCount) WHERE {
    ?module <#layer> ?layer .
}
GROUP BY ?layer
ORDER BY DESC(?moduleCount)`,
		Variables: []Variable{},
		Example:   "graphfs examples run layer-stats",
	},
	{
		Name:        "change-impact",
		Description: "Find all modules affected by changing a module",
		Category:    "impact",
		Query: `SELECT ?affected WHERE {
    ?affected <#imports>+ <#{{.module}}> .
}`,
		Variables: []Variable{
			{Name: "module", Description: "Module path to analyze"},
		},
		Example: "graphfs examples run change-impact --module=pkg/graph/builder.go",
	},
	{
		Name:        "undocumented-modules",
		Description: "Find modules without LinkedDoc comments",
		Category:    "documentation",
		Query: `SELECT ?module WHERE {
    ?module a <#Module> .
    FILTER NOT EXISTS {
        ?module <#description> ?desc .
    }
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run undocumented-modules",
	},
	{
		Name:        "missing-exports",
		Description: "Find modules that don't export anything",
		Category:    "documentation",
		Query: `SELECT ?module WHERE {
    ?module a <#Module> .
    FILTER NOT EXISTS {
        ?module <#exports> ?export .
    }
}`,
		Variables: []Variable{},
		Example:   "graphfs examples run missing-exports",
	},
	{
		Name:        "api-surface",
		Description: "Find all public API exports by package",
		Category:    "documentation",
		Query: `SELECT ?module ?export WHERE {
    ?module <#layer> "{{.layer}}" .
    ?module <#exports> ?export .
}
ORDER BY ?module`,
		Variables: []Variable{
			{Name: "layer", Description: "Layer name", Default: "api"},
		},
		Example: "graphfs examples run api-surface --layer=api",
	},
}

// TemplateManager manages query templates
type TemplateManager struct {
	customTemplatesDir string
	templates          map[string]*QueryTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(customTemplatesDir string) *TemplateManager {
	tm := &TemplateManager{
		customTemplatesDir: customTemplatesDir,
		templates:          make(map[string]*QueryTemplate),
	}

	// Load built-in templates
	for i := range BuiltInTemplates {
		tm.templates[BuiltInTemplates[i].Name] = &BuiltInTemplates[i]
	}

	// Load custom templates if directory exists
	if customTemplatesDir != "" {
		tm.loadCustomTemplates()
	}

	return tm
}

// GetTemplate retrieves a template by name
func (tm *TemplateManager) GetTemplate(templateName string) (*QueryTemplate, error) {
	tmpl, ok := tm.templates[templateName]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}
	return tmpl, nil
}

// ListTemplates returns all templates, optionally filtered by category
func (tm *TemplateManager) ListTemplates(category string) []*QueryTemplate {
	var templates []*QueryTemplate
	for _, tmpl := range tm.templates {
		if category == "" || tmpl.Category == category {
			templates = append(templates, tmpl)
		}
	}
	return templates
}

// GetCategories returns all unique template categories
func (tm *TemplateManager) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, tmpl := range tm.templates {
		categoryMap[tmpl.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}

// RenderTemplate renders a template with the given variables
func (tm *TemplateManager) RenderTemplate(templateName string, variables map[string]string) (string, error) {
	tmpl, err := tm.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	return tm.Render(tmpl, variables)
}

// Render renders a template with the given variables
func (tm *TemplateManager) Render(tmpl *QueryTemplate, variables map[string]string) (string, error) {
	// Apply defaults for missing variables
	vars := make(map[string]string)
	for _, v := range tmpl.Variables {
		if val, ok := variables[v.Name]; ok {
			vars[v.Name] = val
		} else if v.Default != "" {
			vars[v.Name] = v.Default
		} else {
			return "", fmt.Errorf("required variable missing: %s", v.Name)
		}
	}

	// Parse and execute template
	goTemplate, err := template.New(tmpl.Name).Parse(tmpl.Query)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := goTemplate.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// SaveCustomTemplate saves a custom template to disk
func (tm *TemplateManager) SaveCustomTemplate(tmpl *QueryTemplate) error {
	if tm.customTemplatesDir == "" {
		return fmt.Errorf("custom templates directory not configured")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(tm.customTemplatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Serialize template to JSON
	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Write to file
	filePath := filepath.Join(tm.customTemplatesDir, tmpl.Name+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	// Add to in-memory cache
	tm.templates[tmpl.Name] = tmpl

	return nil
}

// ExportTemplate exports a template to a .sparql file
func (tm *TemplateManager) ExportTemplate(templateName string, outputPath string, variables map[string]string) error {
	// Render template
	rendered, err := tm.RenderTemplate(templateName, variables)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// loadCustomTemplates loads custom templates from the templates directory
func (tm *TemplateManager) loadCustomTemplates() {
	entries, err := os.ReadDir(tm.customTemplatesDir)
	if err != nil {
		return // Directory doesn't exist or can't be read
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(tm.customTemplatesDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var tmpl QueryTemplate
		if err := json.Unmarshal(data, &tmpl); err != nil {
			continue
		}

		tm.templates[tmpl.Name] = &tmpl
	}
}
