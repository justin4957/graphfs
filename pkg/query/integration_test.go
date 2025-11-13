/*
Integration tests for SPARQL query engine using the minimal-app example.
*/

package query

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/parser"
	"github.com/justin4957/graphfs/pkg/scanner"
)

func setupMinimalAppStore(t *testing.T) *store.TripleStore {
	t.Helper()

	// Setup scanner
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Scan minimal-app directory
	s := scanner.NewScanner()
	scanResult, err := s.Scan(absPath, scanner.ScanOptions{
		UseDefaults: true,
		Concurrent:  true,
	})
	if err != nil {
		t.Fatalf("Failed to scan minimal-app: %v", err)
	}

	// Create triple store
	ts := store.NewTripleStore()

	// Parse each file and add triples
	p := parser.NewParser()
	tripleCount := 0
	for _, file := range scanResult.Files {
		if file.HasLinkedDoc {
			triples, err := p.Parse(file.Path)
			if err != nil {
				t.Logf("Warning: failed to parse %s: %v", file.Path, err)
				continue
			}

			for _, triple := range triples {
				var objectStr string
				switch obj := triple.Object.(type) {
				case parser.LiteralObject:
					objectStr = obj.Value
				case parser.URIObject:
					objectStr = obj.URI
				case parser.BlankNodeObject:
					// For blank nodes, just use a string representation
					objectStr = fmt.Sprintf("_:b%d", len(obj.Triples))
				}

				if err := ts.Add(triple.Subject, triple.Predicate, objectStr); err != nil {
					t.Logf("Warning: failed to add triple: %v", err)
				}
				tripleCount++
			}
		}
	}

	t.Logf("Loaded %d triples from minimal-app", tripleCount)
	return ts
}

func TestIntegration_FindAllModules(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
		SELECT ?module WHERE {
			?module rdf:type code:Module .
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count < 6 {
		t.Errorf("Expected at least 6 modules, got %d", result.Count)
	}

	t.Logf("Found %d modules:", result.Count)
	for i, binding := range result.Bindings {
		t.Logf("  [%d] %v", i, binding["module"])
	}
}

func TestIntegration_FindModuleNames(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:name ?name .
		}
		ORDER BY ASC(?name)
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count < 6 {
		t.Errorf("Expected at least 6 modules with names, got %d", result.Count)
	}

	t.Logf("Found %d modules with names:", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s -> %s", binding["module"], binding["name"])
	}
}

func TestIntegration_FindDependencies(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?dependency WHERE {
			?module code:linksTo ?dependency .
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	t.Logf("Found %d dependencies:", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s -> %s", binding["module"], binding["dependency"])
	}
}

func TestIntegration_FilterByLanguage(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:language ?lang .
			?module code:name ?name .
			FILTER(REGEX(?lang, "go"))
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count < 6 {
		t.Errorf("Expected at least 6 Go modules, got %d", result.Count)
	}

	t.Logf("Found %d Go modules:", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s (%s)", binding["name"], binding["module"])
	}
}

func TestIntegration_FindExports(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT DISTINCT ?module ?export WHERE {
			?module code:exports ?export .
		}
		LIMIT 10
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	t.Logf("Found %d exports (limited to 10):", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s exports %s", binding["module"], binding["export"])
	}
}

func TestIntegration_FindUtilsModules(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:name ?name .
			FILTER(REGEX(?name, "utils/"))
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count < 3 {
		t.Errorf("Expected at least 3 utils modules, got %d", result.Count)
	}

	t.Logf("Found %d utils modules:", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s", binding["name"])
	}
}

func TestIntegration_ComplexQuery(t *testing.T) {
	ts := setupMinimalAppStore(t)
	executor := NewExecutor(ts)

	// Find all modules that link to other modules, with names
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
		SELECT DISTINCT ?module ?name WHERE {
			?module rdf:type code:Module .
			?module code:name ?name .
			?module code:linksTo ?dep .
		}
		ORDER BY ASC(?name)
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	t.Logf("Found %d modules with dependencies:", result.Count)
	for _, binding := range result.Bindings {
		t.Logf("  %s (%s)", binding["name"], binding["module"])
	}
}
