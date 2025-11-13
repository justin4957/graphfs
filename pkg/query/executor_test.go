package query

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

func setupTestStore() *store.TripleStore {
	ts := store.NewTripleStore()

	// Add test data
	ts.Add("<#main.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")
	ts.Add("<#main.go>", "https://schema.codedoc.org/linksTo", "<./utils.go>")
	ts.Add("<#main.go>", "https://schema.codedoc.org/exports", "<#main>")

	ts.Add("<#utils.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	ts.Add("<#utils.go>", "https://schema.codedoc.org/name", "utils.go")
	ts.Add("<#utils.go>", "https://schema.codedoc.org/exports", "<#helper>")

	return ts
}

func TestExecutor_SimpleSelect(t *testing.T) {
	ts := setupTestStore()
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

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	if len(result.Variables) != 1 || result.Variables[0] != "module" {
		t.Errorf("Variables = %v, want [module]", result.Variables)
	}
}

func TestExecutor_SelectWithMultiplePatterns(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:name ?name .
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2", result.Count)
	}

	// Check that we got both modules
	foundMain := false
	foundUtils := false
	for _, binding := range result.Bindings {
		if binding["name"] == "main.go" {
			foundMain = true
		}
		if binding["name"] == "utils.go" {
			foundUtils = true
		}
	}

	if !foundMain || !foundUtils {
		t.Error("Expected to find both main.go and utils.go")
	}
}

func TestExecutor_SelectAll(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `SELECT * WHERE { ?s ?p ?o . } LIMIT 3`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Count = %d, want 3", result.Count)
	}

	// Should have s, p, o variables
	if len(result.Variables) != 3 {
		t.Errorf("Variables count = %d, want 3", len(result.Variables))
	}
}

func TestExecutor_WithLimit(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `SELECT ?s WHERE { ?s ?p ?o . } LIMIT 2`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Count = %d, want 2 (LIMIT applied)", result.Count)
	}
}

func TestExecutor_WithOffset(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `SELECT ?s WHERE { ?s ?p ?o . } OFFSET 3`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	// Should have fewer results after offset
	if result.Count >= 7 {
		t.Errorf("Count = %d, should be less after OFFSET 3", result.Count)
	}
}

func TestExecutor_WithDistinct(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `SELECT DISTINCT ?s WHERE { ?s ?p ?o . }`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	// Should have only 2 unique subjects
	if result.Count != 2 {
		t.Errorf("Count = %d, want 2 (DISTINCT applied)", result.Count)
	}
}

func TestExecutor_WithFilter_REGEX(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:name ?name .
			FILTER(REGEX(?name, "main"))
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Count = %d, want 1 (filtered by REGEX)", result.Count)
	}

	if result.Bindings[0]["name"] != "main.go" {
		t.Errorf("name = %v, want main.go", result.Bindings[0]["name"])
	}
}

func TestExecutor_SpecificSubject(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?name WHERE {
			<#main.go> code:name ?name .
		}
	`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Count = %d, want 1", result.Count)
	}

	if result.Count > 0 && result.Bindings[0]["name"] != "main.go" {
		t.Errorf("name = %v, want main.go", result.Bindings[0]["name"])
	}
}

func TestExecutor_NoResults(t *testing.T) {
	ts := setupTestStore()
	executor := NewExecutor(ts)

	queryStr := `SELECT ?module WHERE { ?module <http://nonexistent> ?o . }`

	result, err := executor.ExecuteString(queryStr)
	if err != nil {
		t.Fatalf("ExecuteString() error = %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}
