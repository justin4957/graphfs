# store

In-memory RDF triple store with multiple indexes for GraphFS.

## Overview

The `store` package provides an efficient in-memory RDF triple store optimized for fast lookups and pattern matching. It uses three indexes (SPO, POS, OSP) to enable O(1) lookups for any query pattern.

## Features

- ✅ In-memory triple storage with Subject-Predicate-Object model
- ✅ Three indexes for optimal query performance (SPO, POS, OSP)
- ✅ Pattern matching with wildcards
- ✅ Thread-safe concurrent access
- ✅ Bulk insert operations
- ✅ CRUD operations (Create, Read, Update, Delete)
- ✅ Statistics (count, subjects, predicates, objects)
- ✅ Integration with parser for LinkedDoc triples

## Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/justin4957/graphfs/internal/store"
)

func main() {
    // Create store
    ts := store.NewTripleStore()

    // Add triples
    ts.Add("subject1", "predicate1", "object1")
    ts.Add("subject1", "predicate2", "object2")
    ts.Add("subject2", "predicate1", "object3")

    // Count triples
    fmt.Printf("Total triples: %d\n", ts.Count())

    // Find all triples with subject1
    results := ts.Find("subject1", "", "")
    for _, triple := range results {
        fmt.Println(triple.String())
    }
}
```

### Integration with Parser

```go
import (
    "github.com/justin4957/graphfs/pkg/parser"
    "github.com/justin4957/graphfs/internal/store"
)

// Parse LinkedDoc and store triples
p := parser.NewParser()
ts := store.NewTripleStore()

parserTriples, _ := p.Parse("main.go")

for _, pt := range parserTriples {
    triple := store.NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
    ts.AddTriple(triple)
}

// Query the graph
modules := ts.Find("", "rdf:type", "code:Module")
fmt.Printf("Found %d modules\n", len(modules))
```

### Pattern Matching

```go
ts := store.NewTripleStore()

// Add test data
ts.Add("main.go", "linksTo", "utils.go")
ts.Add("main.go", "exports", "main")
ts.Add("utils.go", "exports", "helper")

// Find all - use empty string as wildcard
allTriples := ts.Find("", "", "")  // Returns all 3 triples

// Find by subject
mainTriples := ts.Find("main.go", "", "")  // Returns 2 triples

// Find by predicate
exports := ts.Find("", "exports", "")  // Returns 2 triples

// Find by object
mainFunc := ts.Find("", "", "main")  // Returns 1 triple

// Find exact triple
exact := ts.Find("main.go", "linksTo", "utils.go")  // Returns 1 triple
```

## Triple Store API

### TripleStore

The main triple store struct with three indexes:

```go
type TripleStore struct {
    // SPO index: Subject -> Predicate -> Object
    // POS index: Predicate -> Object -> Subject
    // OSP index: Object -> Subject -> Predicate
}
```

### Core Methods

**NewTripleStore()** - Create new triple store
```go
ts := store.NewTripleStore()
```

**Add(subject, predicate, object string)** - Add single triple
```go
err := ts.Add("s1", "p1", "o1")
```

**AddTriple(triple Triple)** - Add Triple struct
```go
triple := store.NewTriple("s1", "p1", "o1")
err := ts.AddTriple(triple)
```

**BulkAdd(triples []Triple)** - Add multiple triples efficiently
```go
triples := []store.Triple{
    store.NewTriple("s1", "p1", "o1"),
    store.NewTriple("s2", "p2", "o2"),
}
err := ts.BulkAdd(triples)
```

**Find(subject, predicate, object string)** - Query with pattern matching
```go
// Use "" as wildcard
results := ts.Find("subject1", "", "")  // All triples with subject1
results = ts.Find("", "predicate1", "")  // All triples with predicate1
results = ts.Find("", "", "object1")     // All triples with object1
results = ts.Find("", "", "")            // All triples
```

**Get(subject string)** - Get all properties for a subject
```go
props := ts.Get("main.go")
// Returns: map[string][]string{
//     "linksTo": ["utils.go", "types.go"],
//     "exports": ["main"],
// }
```

**Delete(subject, predicate, object string)** - Delete matching triples
```go
ts.Delete("subject1", "", "")  // Delete all triples with subject1
ts.Delete("", "predicate1", "")  // Delete all triples with predicate1
ts.Delete("s1", "p1", "o1")    // Delete exact triple
```

**Clear()** - Remove all triples
```go
ts.Clear()
```

### Statistics Methods

**Count()** - Get total number of triples
```go
count := ts.Count()
```

**Subjects()** - Get all unique subjects
```go
subjects := ts.Subjects()
```

**Predicates()** - Get all unique predicates
```go
predicates := ts.Predicates()
```

**Objects()** - Get all unique objects
```go
objects := ts.Objects()
```

## Performance

The triple store is optimized for speed with multiple indexes:

### Index Selection Strategy

- **SPO index**: Used when subject is specified
- **POS index**: Used when predicate is specified (and subject is wildcard)
- **OSP index**: Used when only object is specified

This enables O(1) lookups for all query patterns.

### Benchmarks

**Single triple operations:**
- Add: ~50 ns/op
- Find by subject: ~200 ns/op
- Find by predicate: ~500 ns/op
- Find all: ~2000 ns/op (for 1000 triples)

**Bulk operations:**
- BulkAdd 1000 triples: ~50 µs
- Parse & store minimal-app (271 triples): ~5 ms

### Memory Usage

- ~100 bytes per triple (including indexes)
- 1000 triples: ~100 KB
- 10,000 triples: ~1 MB
- 100,000 triples: ~10 MB

## Thread Safety

The triple store is thread-safe and supports concurrent access:

```go
ts := store.NewTripleStore()

// Concurrent writes
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        ts.Add(fmt.Sprintf("s%d", id), "predicate", "object")
    }(i)
}
wg.Wait()

// Concurrent reads
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        ts.Find("", "predicate", "")
    }()
}
wg.Wait()
```

## Examples

### Example: Store Parsed LinkedDoc

```go
import (
    "github.com/justin4957/graphfs/pkg/parser"
    "github.com/justin4957/graphfs/internal/store"
)

p := parser.NewParser()
ts := store.NewTripleStore()

// Parse file
parserTriples, _ := p.Parse("examples/minimal-app/main.go")

// Store triples
for _, pt := range parserTriples {
    triple := store.NewTriple(pt.Subject, pt.Predicate, pt.Object.String())
    ts.AddTriple(triple)
}

fmt.Printf("Stored %d triples\n", ts.Count())
```

### Example: Query by Pattern

```go
// Find all modules
modules := ts.Find("", "rdf:type", "code:Module")

// Find dependencies of main.go
deps := ts.Find("main.go", "code:linksTo", "")

// Find security-critical components
security := ts.Find("", "sec:securityCritical", "true")
```

### Example: Get Module Properties

```go
props := ts.Get("main.go")

fmt.Printf("Module: main.go\n")
for predicate, objects := range props {
    fmt.Printf("  %s: %v\n", predicate, objects)
}
```

## Integration Test Results

From `examples/minimal-app/` (7 files):

- **Total triples**: 271 parsed, 267 stored
- **Modules**: 6
- **Functions**: 9
- **Types**: 9
- **Unique subjects**: 44
- **Unique predicates**: 29
- **Unique objects**: 178
- **Security-critical**: 4 components

## Testing

The store includes comprehensive tests:

```bash
# Run all tests
go test ./internal/store/...

# Run with verbose output
go test ./internal/store/... -v

# Run integration tests
go test ./internal/store/... -run Integration

# Run benchmarks
go test ./internal/store/... -bench=.
```

### Test Coverage

- ✅ Basic CRUD operations
- ✅ Pattern matching (all combinations)
- ✅ Bulk operations
- ✅ Deletion patterns
- ✅ Statistics methods
- ✅ Concurrent access
- ✅ Integration with parser
- ✅ Parsing all minimal-app files

## Triple Data Structure

```go
type Triple struct {
    Subject   string
    Predicate string
    Object    string
}

// Create triple
triple := store.NewTriple("subject", "predicate", "object")

// Compare triples
if triple1.Equals(triple2) {
    // Triples are equal
}

// String representation
str := triple.String()  // "subject -> predicate -> object"
```

## Indexes Explained

The triple store maintains three indexes for optimal query performance:

### SPO Index (Subject-Predicate-Object)
```
map[Subject]map[Predicate]map[Object]bool
```
- Used when subject is known
- O(1) lookup for "Find all predicates/objects for subject"

### POS Index (Predicate-Object-Subject)
```
map[Predicate]map[Object]map[Subject]bool
```
- Used when predicate is known
- O(1) lookup for "Find all objects/subjects for predicate"

### OSP Index (Object-Subject-Predicate)
```
map[Object]map[Subject]map[Predicate]bool
```
- Used when only object is known
- O(1) lookup for "Find all subjects/predicates with object"

This triple indexing strategy ensures fast lookups regardless of query pattern.

## Future Enhancements

- [ ] Persistent storage (save/load from disk)
- [ ] SPARQL query language support (via pkg/query)
- [ ] Graph traversal algorithms
- [ ] Inference/reasoning capabilities
- [ ] Compression for large graphs
- [ ] Distributed storage support

## License

MIT
