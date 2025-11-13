package store

import (
	"path/filepath"
	"testing"

	"github.com/justin4957/graphfs/pkg/parser"
)

// TestIntegration_ParseAndStore tests parsing LinkedDoc and storing in triple store
func TestIntegration_ParseAndStore(t *testing.T) {
	// Create parser and store
	p := parser.NewParser()
	store := NewTripleStore()

	// Parse main.go from minimal-app
	testFile := filepath.Join("..", "..", "examples", "minimal-app", "main.go")

	parserTriples, err := p.Parse(testFile)
	if err != nil {
		t.Fatalf("Parser.Parse() error = %v", err)
	}

	if len(parserTriples) == 0 {
		t.Fatal("Parser returned no triples")
	}

	t.Logf("Parsed %d triples from main.go", len(parserTriples))

	// Convert parser triples to store triples and add to store
	for _, pt := range parserTriples {
		triple := NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
		err := store.AddTriple(triple)
		if err != nil {
			t.Fatalf("Store.AddTriple() error = %v", err)
		}
	}

	// Verify triples were stored
	if store.Count() != len(parserTriples) {
		t.Errorf("Store count = %d, want %d", store.Count(), len(parserTriples))
	}

	// Query for module type
	moduleTriples := store.Find("", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	if len(moduleTriples) == 0 {
		t.Error("Expected to find Module type declarations")
	}

	t.Logf("Found %d modules", len(moduleTriples))

	// Query for module name
	nameTriples := store.Find("<#main.go>", "https://schema.codedoc.org/name", "")
	if len(nameTriples) == 0 {
		t.Error("Expected to find module name")
	} else {
		t.Logf("Module name: %s", nameTriples[0].Object)
	}

	// Query for dependencies (linksTo)
	deps := store.Find("<#main.go>", "https://schema.codedoc.org/linksTo", "")
	t.Logf("Found %d dependencies for main.go", len(deps))

	for _, dep := range deps {
		t.Logf("  Dependency: %s", dep.Object)
	}

	// Get all properties of main.go
	props := store.Get("<#main.go>")
	t.Logf("main.go has %d different properties", len(props))

	for pred, objects := range props {
		t.Logf("  %s: %d values", pred, len(objects))
	}
}

// TestIntegration_ParseAllMinimalApp tests parsing all files from minimal-app
func TestIntegration_ParseAllMinimalApp(t *testing.T) {
	p := parser.NewParser()
	store := NewTripleStore()

	files := []string{
		"main.go",
		"models/user.go",
		"services/auth.go",
		"services/user.go",
		"utils/logger.go",
		"utils/crypto.go",
		"utils/validator.go",
	}

	totalParsed := 0

	for _, file := range files {
		path := filepath.Join("..", "..", "examples", "minimal-app", file)

		parserTriples, err := p.Parse(path)
		if err != nil {
			t.Errorf("Parse(%s) error = %v", file, err)
			continue
		}

		totalParsed += len(parserTriples)

		// Add to store
		for _, pt := range parserTriples {
			triple := NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
			store.AddTriple(triple)
		}

		t.Logf("%s: %d triples", file, len(parserTriples))
	}

	t.Logf("\nTotal triples parsed: %d", totalParsed)
	t.Logf("Triples in store: %d", store.Count())

	// Query statistics
	modules := store.Find("", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	functions := store.Find("", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Function")
	types := store.Find("", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Type")

	t.Logf("\nStatistics:")
	t.Logf("  Modules: %d", len(modules))
	t.Logf("  Functions: %d", len(functions))
	t.Logf("  Types: %d", len(types))
	t.Logf("  Unique subjects: %d", len(store.Subjects()))
	t.Logf("  Unique predicates: %d", len(store.Predicates()))
	t.Logf("  Unique objects: %d", len(store.Objects()))

	// Verify we got a reasonable number of triples
	if totalParsed < 200 {
		t.Errorf("Expected at least 200 triples from minimal-app, got %d", totalParsed)
	}

	// Query for security-critical modules
	securityTriples := store.Find("", "https://schema.codedoc.org/security/securityCritical", "true")
	t.Logf("\nSecurity-critical modules: %d", len(securityTriples))

	for _, triple := range securityTriples {
		t.Logf("  %s is security-critical", triple.Subject)
	}
}

// TestIntegration_QueryPatterns tests various query patterns
func TestIntegration_QueryPatterns(t *testing.T) {
	p := parser.NewParser()
	store := NewTripleStore()

	// Parse auth.go (has security metadata)
	authFile := filepath.Join("..", "..", "examples", "minimal-app", "services", "auth.go")

	parserTriples, err := p.Parse(authFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	for _, pt := range parserTriples {
		triple := NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
		store.AddTriple(triple)
	}

	tests := []struct {
		name      string
		subject   string
		predicate string
		object    string
		minCount  int
	}{
		{
			name:      "find all triples",
			subject:   "",
			predicate: "",
			object:    "",
			minCount:  30,
		},
		{
			name:      "find module exports",
			subject:   "",
			predicate: "https://schema.codedoc.org/exports",
			object:    "",
			minCount:  1,
		},
		{
			name:      "find security boundaries",
			subject:   "",
			predicate: "https://schema.codedoc.org/security/securityBoundary",
			object:    "",
			minCount:  0, // May or may not be present
		},
		{
			name:      "find module tags",
			subject:   "",
			predicate: "https://schema.codedoc.org/tags",
			object:    "",
			minCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := store.Find(tt.subject, tt.predicate, tt.object)

			if len(results) < tt.minCount {
				t.Errorf("Find() returned %d results, want at least %d", len(results), tt.minCount)
			}

			t.Logf("Query matched %d triples", len(results))
		})
	}
}

// BenchmarkIntegration_ParseAndStore benchmarks parsing and storing
func BenchmarkIntegration_ParseAndStore(b *testing.B) {
	testFile := filepath.Join("..", "..", "examples", "minimal-app", "main.go")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewParser()
		store := NewTripleStore()

		parserTriples, _ := p.Parse(testFile)

		for _, pt := range parserTriples {
			triple := NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
			store.AddTriple(triple)
		}
	}
}
