/*
# Module: pkg/repl/completer_test.go
Tests for autocomplete functionality.

Tests the completer with various graph configurations.

## Linked Modules
- [completer](./completer.go) - Completer

## Tags
repl, test, autocomplete

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#completer_test.go> a code:Module ;
    code:name "pkg/repl/completer_test.go" ;
    code:description "Tests for autocomplete functionality" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <./completer.go> ;
    code:tags "repl", "test", "autocomplete" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func TestNewCompleter(t *testing.T) {
	g := createTestGraph(t)
	completer := NewCompleter(g)

	if completer == nil {
		t.Fatal("Expected non-nil completer")
	}

	if completer.graph != g {
		t.Error("Completer graph mismatch")
	}
}

func TestCompleterGetModules(t *testing.T) {
	g := createTestGraph(t)
	completer := NewCompleter(g)

	modules := completer.GetModules()
	if len(modules) == 0 {
		t.Error("Expected modules, got none")
	}

	// Check that modules are formatted correctly
	for _, mod := range modules {
		if len(mod) < 3 || mod[0] != '<' || mod[1] != '#' {
			t.Errorf("Module not properly formatted: %s", mod)
		}
	}
}

func TestCompleterGetPredicates(t *testing.T) {
	g := createTestGraph(t)
	completer := NewCompleter(g)

	predicates := completer.GetPredicates()
	if len(predicates) == 0 {
		t.Error("Expected predicates, got none")
	}

	// Check for common predicates
	foundImports := false
	foundExports := false
	for _, pred := range predicates {
		if pred == "<#imports>" {
			foundImports = true
		}
		if pred == "<#exports>" {
			foundExports = true
		}
	}

	if !foundImports {
		t.Error("Expected <#imports> predicate")
	}
	if !foundExports {
		t.Error("Expected <#exports> predicate")
	}
}

func TestCompleterGetKeywords(t *testing.T) {
	g := createTestGraph(t)
	completer := NewCompleter(g)

	keywords := completer.GetKeywords()
	if len(keywords) == 0 {
		t.Error("Expected keywords, got none")
	}

	// Check for common keywords
	foundSelect := false
	foundWhere := false
	for _, kw := range keywords {
		if kw == "SELECT" {
			foundSelect = true
		}
		if kw == "WHERE" {
			foundWhere = true
		}
	}

	if !foundSelect {
		t.Error("Expected SELECT keyword")
	}
	if !foundWhere {
		t.Error("Expected WHERE keyword")
	}
}

func TestFilterSuggestions(t *testing.T) {
	suggestions := []string{
		"SELECT",
		"WHERE",
		"SELECT DISTINCT",
		"FILTER",
		"FORMAT",
	}

	tests := []struct {
		prefix   string
		expected int
	}{
		{"", 5},            // No prefix returns all
		{"SEL", 2},         // "SELECT" and "SELECT DISTINCT"
		{"WHERE", 1},       // Exact match
		{"FI", 1},          // "FILTER"
		{"FOR", 1},         // "FORMAT"
		{"NONEXISTENT", 0}, // No matches
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			filtered := FilterSuggestions(suggestions, tt.prefix)
			if len(filtered) != tt.expected {
				t.Errorf("Expected %d suggestions for prefix '%s', got %d",
					tt.expected, tt.prefix, len(filtered))
			}
		})
	}
}

func TestGetSPARQLKeywords(t *testing.T) {
	keywords := getSPARQLKeywords()
	if len(keywords) == 0 {
		t.Error("Expected SPARQL keywords, got none")
	}

	// Verify essential keywords are present
	essentialKeywords := []string{
		"SELECT", "WHERE", "CONSTRUCT", "ASK", "DESCRIBE",
		"FILTER", "OPTIONAL", "UNION", "LIMIT", "OFFSET",
	}

	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[kw] = true
	}

	for _, essential := range essentialKeywords {
		if !keywordMap[essential] {
			t.Errorf("Expected essential keyword '%s' not found", essential)
		}
	}
}

// createTestGraph creates a simple test graph
func createTestGraph(t *testing.T) *graph.Graph {
	tripleStore := store.NewTripleStore()

	modules := map[string]*graph.Module{
		"test/module1.go": {
			Name:     "test/module1.go",
			Path:     "test/module1.go",
			Language: "go",
			Layer:    "api",
		},
		"test/module2.go": {
			Name:     "test/module2.go",
			Path:     "test/module2.go",
			Language: "go",
			Layer:    "service",
		},
	}

	return &graph.Graph{
		Modules: modules,
		Store:   tripleStore,
	}
}
