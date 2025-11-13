# SPARQL Query Engine

A lightweight SPARQL query engine for querying RDF triple stores extracted from LinkedDoc metadata.

## Features

- **SELECT Queries**: Query triples with variables and patterns
- **PREFIX Support**: Define and use namespace prefixes
- **FILTER Clauses**: Filter results with REGEX, CONTAINS, equality, and inequality
- **Query Modifiers**:
  - `DISTINCT`: Remove duplicate results
  - `LIMIT`: Limit number of results
  - `OFFSET`: Skip first N results
  - `ORDER BY`: Sort results (ASC/DESC)
- **Pattern Matching**: Subject-predicate-object triple patterns with wildcards
- **Variable Binding**: Bind and propagate variable values across patterns

## Installation

```bash
go get github.com/justin4957/graphfs/pkg/query
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/justin4957/graphfs/internal/store"
    "github.com/justin4957/graphfs/pkg/query"
)

func main() {
    // Create a triple store
    ts := store.NewTripleStore()

    // Add some triples
    ts.Add("<#main.go>", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
    ts.Add("<#main.go>", "https://schema.codedoc.org/name", "main.go")

    // Create executor
    executor := query.NewExecutor(ts)

    // Execute SPARQL query
    queryStr := `
        PREFIX code: <https://schema.codedoc.org/>
        PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
        SELECT ?module ?name WHERE {
            ?module rdf:type code:Module .
            ?module code:name ?name .
        }
    `

    result, err := executor.ExecuteString(queryStr)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d modules\n", result.Count)
    for _, binding := range result.Bindings {
        fmt.Printf("Module: %s (name: %s)\n", binding["module"], binding["name"])
    }
}
```

## Usage Examples

### Basic SELECT Query

```go
queryStr := `
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?name WHERE {
        ?module code:name ?name .
    }
`
```

### Using FILTER with REGEX

```go
queryStr := `
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?name WHERE {
        ?module code:name ?name .
        FILTER(REGEX(?name, "main"))
    }
`
```

### Query with LIMIT and OFFSET

```go
queryStr := `
    SELECT ?s ?p ?o WHERE {
        ?s ?p ?o .
    }
    LIMIT 10 OFFSET 5
`
```

### DISTINCT Results

```go
queryStr := `
    SELECT DISTINCT ?module WHERE {
        ?module ?predicate ?object .
    }
`
```

### ORDER BY

```go
queryStr := `
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?name WHERE {
        ?module code:name ?name .
    }
    ORDER BY ASC(?name)
`
```

### Finding Dependencies

```go
queryStr := `
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?dependency WHERE {
        ?module code:linksTo ?dependency .
    }
`
```

## Query Structure

### SELECT Clause

```sparql
SELECT ?var1 ?var2 ...
SELECT *                  # Select all variables
SELECT DISTINCT ?var      # Remove duplicates
```

### WHERE Clause

Triple patterns with variables:
```sparql
WHERE {
    ?subject ?predicate ?object .
    ?subject <http://example.org/property> "literal value" .
    <#specific> ?predicate ?object .
}
```

### PREFIX Declarations

```sparql
PREFIX code: <https://schema.codedoc.org/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
```

### FILTER Expressions

Supported filter functions:

- **REGEX**: Pattern matching
  ```sparql
  FILTER(REGEX(?name, "pattern"))
  ```

- **CONTAINS**: Substring search
  ```sparql
  FILTER(CONTAINS(?name, "substring"))
  ```

- **Equality**: Exact match
  ```sparql
  FILTER(?var = "value")
  ```

- **Inequality**: Not equal
  ```sparql
  FILTER(?var != "value")
  ```

### Query Modifiers

```sparql
LIMIT 10              # Return at most 10 results
OFFSET 5              # Skip first 5 results
DISTINCT              # Remove duplicate results
ORDER BY ASC(?var)    # Sort ascending
ORDER BY DESC(?var)   # Sort descending
```

## API Reference

### Executor

```go
type Executor struct {
    // Contains filtered or unexported fields
}

// NewExecutor creates a new query executor
func NewExecutor(tripleStore *store.TripleStore) *Executor

// Execute executes a parsed query
func (e *Executor) Execute(query *Query) (*QueryResult, error)

// ExecuteString parses and executes a SPARQL query string
func (e *Executor) ExecuteString(queryStr string) (*QueryResult, error)
```

### QueryResult

```go
type QueryResult struct {
    Variables []string             // Variable names (without ?)
    Bindings  []map[string]string  // Variable bindings for each result
    Count     int                  // Number of results
}
```

### Query Parsing

```go
// ParseQuery parses a SPARQL query string
func ParseQuery(queryStr string) (*Query, error)
```

## Supported SPARQL Features

### ✅ Supported

- SELECT queries
- PREFIX declarations
- Triple patterns with variables
- FILTER (REGEX, CONTAINS, =, !=)
- DISTINCT
- LIMIT / OFFSET
- ORDER BY (ASC/DESC)
- Multiple triple patterns (joins)
- Specific subject/predicate/object matching

### ❌ Not Yet Supported

- CONSTRUCT, ASK, DESCRIBE queries
- OPTIONAL patterns
- UNION
- Named graphs
- Property paths
- Aggregation (COUNT, SUM, etc.)
- GROUP BY / HAVING
- BIND
- Subqueries
- Negation (NOT EXISTS, MINUS)
- Advanced filter functions (STR, LANG, DATATYPE, etc.)

## Integration with GraphFS

The query engine integrates seamlessly with the GraphFS pipeline:

```go
// 1. Scan codebase
scanner := scanner.NewScanner()
result, _ := scanner.Scan("/path/to/code", scanner.ScanOptions{})

// 2. Parse LinkedDoc metadata
parser := parser.NewParser()
ts := store.NewTripleStore()

for _, file := range result.Files {
    if file.HasLinkedDoc {
        triples, _ := parser.Parse(file.Path)
        for _, triple := range triples {
            // Add to store...
        }
    }
}

// 3. Query the knowledge graph
executor := query.NewExecutor(ts)
queryResult, _ := executor.ExecuteString(`
    PREFIX code: <https://schema.codedoc.org/>
    SELECT ?module ?dependency WHERE {
        ?module code:linksTo ?dependency .
    }
`)
```

## Performance

- **Pattern Matching**: O(1) lookups via triple store indexes
- **Filtering**: O(n) where n is the number of bindings
- **Sorting**: O(n log n) for ORDER BY
- **Memory**: Bindings are stored in memory during query execution

## Testing

Run the test suite:

```bash
# Unit tests
go test ./pkg/query

# Integration tests (requires examples/minimal-app)
go test ./pkg/query -run TestIntegration

# Example tests
go test ./pkg/query -run Example
```

## Examples

See [example_test.go](example_test.go) for complete working examples.

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](../../LICENSE) for details.
