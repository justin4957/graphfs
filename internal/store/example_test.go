package store_test

import (
	"fmt"

	"github.com/justin4957/graphfs/internal/store"
)

// Example demonstrates basic triple store usage
func Example() {
	ts := store.NewTripleStore()

	// Add triples
	ts.Add("main.go", "exports", "main")
	ts.Add("main.go", "linksTo", "utils.go")
	ts.Add("utils.go", "exports", "helper")

	fmt.Printf("Total triples: %d\n", ts.Count())

	// Output:
	// Total triples: 3
}

// Example_patternMatching demonstrates pattern matching with wildcards
func Example_patternMatching() {
	ts := store.NewTripleStore()

	ts.Add("main.go", "linksTo", "utils.go")
	ts.Add("main.go", "exports", "main")
	ts.Add("utils.go", "exports", "helper")

	// Find by subject
	mainTriples := ts.Find("main.go", "", "")
	fmt.Printf("main.go triples: %d\n", len(mainTriples))

	// Find by predicate
	exports := ts.Find("", "exports", "")
	fmt.Printf("exports: %d\n", len(exports))

	// Find all
	all := ts.Find("", "", "")
	fmt.Printf("all: %d\n", len(all))

	// Output:
	// main.go triples: 2
	// exports: 2
	// all: 3
}

// Example_bulkAdd demonstrates bulk insertion
func Example_bulkAdd() {
	ts := store.NewTripleStore()

	triples := []store.Triple{
		store.NewTriple("s1", "p1", "o1"),
		store.NewTriple("s2", "p2", "o2"),
		store.NewTriple("s3", "p3", "o3"),
	}

	ts.BulkAdd(triples)

	fmt.Printf("Count: %d\n", ts.Count())

	// Output:
	// Count: 3
}

// Example_getProperties demonstrates getting all properties for a subject
func Example_getProperties() {
	ts := store.NewTripleStore()

	ts.Add("main.go", "exports", "main")
	ts.Add("main.go", "linksTo", "utils.go")
	ts.Add("main.go", "linksTo", "types.go")

	props := ts.Get("main.go")

	fmt.Printf("Properties: %d\n", len(props))
	fmt.Printf("exports: %d\n", len(props["exports"]))
	fmt.Printf("linksTo: %d\n", len(props["linksTo"]))

	// Output:
	// Properties: 2
	// exports: 1
	// linksTo: 2
}

// Example_delete demonstrates deleting triples
func Example_delete() {
	ts := store.NewTripleStore()

	ts.Add("s1", "p1", "o1")
	ts.Add("s1", "p2", "o2")
	ts.Add("s2", "p1", "o3")

	fmt.Printf("Before: %d\n", ts.Count())

	// Delete specific triple
	ts.Delete("s1", "p1", "o1")

	fmt.Printf("After delete one: %d\n", ts.Count())

	// Delete by pattern
	ts.Delete("s1", "", "")

	fmt.Printf("After delete pattern: %d\n", ts.Count())

	// Output:
	// Before: 3
	// After delete one: 2
	// After delete pattern: 1
}

// Example_statistics demonstrates statistics methods
func Example_statistics() {
	ts := store.NewTripleStore()

	ts.Add("s1", "p1", "o1")
	ts.Add("s1", "p2", "o2")
	ts.Add("s2", "p1", "o1")

	fmt.Printf("Count: %d\n", ts.Count())
	fmt.Printf("Subjects: %d\n", len(ts.Subjects()))
	fmt.Printf("Predicates: %d\n", len(ts.Predicates()))
	fmt.Printf("Objects: %d\n", len(ts.Objects()))

	// Output:
	// Count: 3
	// Subjects: 2
	// Predicates: 2
	// Objects: 2
}

// Example_triple demonstrates triple operations
func Example_triple() {
	t1 := store.NewTriple("subject", "predicate", "object")
	t2 := store.NewTriple("subject", "predicate", "object")
	t3 := store.NewTriple("subject", "predicate", "other")

	fmt.Printf("t1 == t2: %v\n", t1.Equals(t2))
	fmt.Printf("t1 == t3: %v\n", t1.Equals(t3))
	fmt.Printf("String: %s\n", t1.String())

	// Output:
	// t1 == t2: true
	// t1 == t3: false
	// String: subject -> predicate -> object
}
