package query

import (
	"fmt"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

// setupBenchmarkStore creates a large triple store for benchmarking
func setupBenchmarkStore() *store.TripleStore {
	ts := store.NewTripleStore()

	// Add realistic code graph data
	// Simulate a large codebase with 10,000 modules

	// Common predicates (very high cardinality)
	commonPredicates := []string{
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		"https://schema.codedoc.org/name",
		"https://schema.codedoc.org/description",
	}

	// Less common predicates (medium cardinality)
	mediumPredicates := []string{
		"https://schema.codedoc.org/language",
		"https://schema.codedoc.org/layer",
		"https://schema.codedoc.org/linksTo",
	}

	// Rare predicates (low cardinality)
	rarePredicates := []string{
		"https://schema.codedoc.org/exports",
		"https://schema.codedoc.org/tags",
	}

	// Add 10,000 modules
	for i := 0; i < 10000; i++ {
		moduleURI := fmt.Sprintf("<#module%d.go>", i)

		// Common predicates (high cardinality)
		ts.Add(moduleURI, commonPredicates[0], "https://schema.codedoc.org/Module")
		ts.Add(moduleURI, commonPredicates[1], fmt.Sprintf("\"module%d\"", i))
		ts.Add(moduleURI, commonPredicates[2], fmt.Sprintf("\"Description for module %d\"", i))

		// Language (medium cardinality - only a few languages)
		language := []string{"go", "python", "javascript", "rust"}[i%4]
		ts.Add(moduleURI, mediumPredicates[0], language)

		// Layer (medium cardinality - architectural layers)
		layer := []string{"api", "service", "storage", "cli"}[i%4]
		ts.Add(moduleURI, mediumPredicates[1], layer)

		// Links (sparse - only some modules link to others)
		if i%3 == 0 && i > 0 {
			targetModule := fmt.Sprintf("<#module%d.go>", i-1)
			ts.Add(moduleURI, mediumPredicates[2], targetModule)
		}

		// Exports (rare - only some modules export)
		if i%5 == 0 {
			ts.Add(moduleURI, rarePredicates[0], fmt.Sprintf("<#Export%d>", i))
		}

		// Tags (rare - only some modules have tags)
		if i%7 == 0 {
			ts.Add(moduleURI, rarePredicates[1], "\"api\"")
		}
	}

	return ts
}

// BenchmarkQueryWithoutOptimization benchmarks queries without optimization
func BenchmarkQueryWithoutOptimization(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)
	executor.DisablePlanning() // Disable optimization

	// Complex query: find all Go modules in the API layer that have exports
	// This query has patterns with very different selectivities:
	// - language="go" is medium selectivity (25% of modules)
	// - layer="api" is medium selectivity (25% of modules)
	// - has exports is rare (20% of modules)
	// Without optimization, this processes in declaration order
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module
		WHERE {
			?module code:language "go" .
			?module code:layer "api" .
			?module code:exports ?export
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkQueryWithOptimization benchmarks queries with optimization
func BenchmarkQueryWithOptimization(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)
	// Planning is enabled by default

	// Same complex query as above
	// With optimization, the planner will reorder to start with the most selective pattern
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module
		WHERE {
			?module code:language "go" .
			?module code:layer "api" .
			?module code:exports ?export
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexJoinWithoutOptimization tests a complex join query without optimization
func BenchmarkComplexJoinWithoutOptimization(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)
	executor.DisablePlanning()

	// Complex join: find modules that link to other modules in the same layer
	// Very inefficient without optimization due to cartesian product
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?target
		WHERE {
			?module code:layer ?layer .
			?target code:layer ?layer .
			?module code:linksTo ?target
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComplexJoinWithOptimization tests a complex join query with optimization
func BenchmarkComplexJoinWithOptimization(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)
	// Planning enabled by default

	// Same complex join query
	// Optimizer should start with the most selective pattern (linksTo)
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?target
		WHERE {
			?module code:layer ?layer .
			?target code:layer ?layer .
			?module code:linksTo ?target
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimpleQuery benchmarks a simple query (optimization has minimal impact)
func BenchmarkSimpleQuery(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)

	// Simple query with one pattern - optimization won't change anything
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module
		WHERE {
			?module code:language "go"
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHighlySelectiveFirstPattern tests when first pattern is already optimal
func BenchmarkHighlySelectiveFirstPattern(b *testing.B) {
	ts := setupBenchmarkStore()
	executor := NewExecutor(ts)

	// Query where first pattern is already most selective
	// Optimization should have minimal overhead
	query := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module
		WHERE {
			?module code:exports ?export .
			?module code:language "go" .
			?module code:layer "api"
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteString(query)
		if err != nil {
			b.Fatal(err)
		}
	}
}
