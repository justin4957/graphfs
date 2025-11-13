package query_test

import (
	"fmt"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/query"
)

func Example_basicQuery() {
	// Create a triple store
	ts := store.NewTripleStore()

	// Add some triples
	ts.Add("<#main.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")
	ts.Add("<#utils.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	ts.Add("<#utils.go>", "https://schema.codedoc.org/name", "utils.go")

	// Create executor
	executor := query.NewExecutor(ts)

	// Execute a simple SPARQL query
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
		SELECT ?module ?name WHERE {
			?module rdf:type code:Module .
			?module code:name ?name .
		}
	`

	result, _ := executor.ExecuteString(queryStr)

	fmt.Printf("Found %d modules\n", result.Count)
	// Output: Found 2 modules
}

func Example_filterQuery() {
	// Create a triple store
	ts := store.NewTripleStore()

	// Add some triples
	ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")
	ts.Add("<#utils.go>", "https://schema.codedoc.org/name", "utils.go")
	ts.Add("<#test.go>", "https://schema.codedoc.org/name", "test.go")

	// Create executor
	executor := query.NewExecutor(ts)

	// Execute a query with FILTER
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module ?name WHERE {
			?module code:name ?name .
			FILTER(REGEX(?name, "^main"))
		}
	`

	result, _ := executor.ExecuteString(queryStr)

	for _, binding := range result.Bindings {
		fmt.Printf("Module: %s\n", binding["name"])
	}
	// Output: Module: main.go
}

func Example_limitOffset() {
	// Create a triple store
	ts := store.NewTripleStore()

	// Add some triples
	ts.Add("<#file1.go>", "https://schema.codedoc.org/name", "file1.go")
	ts.Add("<#file2.go>", "https://schema.codedoc.org/name", "file2.go")
	ts.Add("<#file3.go>", "https://schema.codedoc.org/name", "file3.go")

	// Create executor
	executor := query.NewExecutor(ts)

	// Execute a query with LIMIT
	queryStr := `
		SELECT ?file WHERE {
			?file <https://schema.codedoc.org/name> ?name .
		}
		LIMIT 2
	`

	result, _ := executor.ExecuteString(queryStr)

	fmt.Printf("Found %d results (limited to 2)\n", result.Count)
	// Output: Found 2 results (limited to 2)
}

func Example_distinct() {
	// Create a triple store with duplicate subjects
	ts := store.NewTripleStore()

	// Add triples - same module with multiple properties
	ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")
	ts.Add("<#main.go>", "https://schema.codedoc.org/language", "go")
	ts.Add("<#main.go>", "https://schema.codedoc.org/layer", "app")

	// Create executor
	executor := query.NewExecutor(ts)

	// Query without DISTINCT would return 3 results
	queryStr := `
		SELECT DISTINCT ?module WHERE {
			?module ?p ?o .
		}
	`

	result, _ := executor.ExecuteString(queryStr)

	fmt.Printf("Found %d unique modules\n", result.Count)
	// Output: Found 1 unique modules
}

func Example_dependencies() {
	// Create a triple store with dependency information
	ts := store.NewTripleStore()

	// Add module and dependency triples
	ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")
	ts.Add("<#main.go>", "https://schema.codedoc.org/linksTo", "./utils.go")
	ts.Add("<#main.go>", "https://schema.codedoc.org/linksTo", "./models.go")

	// Create executor
	executor := query.NewExecutor(ts)

	// Find all dependencies
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?dependency WHERE {
			<#main.go> code:linksTo ?dependency .
		}
	`

	result, _ := executor.ExecuteString(queryStr)

	fmt.Printf("main.go has %d dependencies\n", result.Count)
	// Output: main.go has 2 dependencies
}
