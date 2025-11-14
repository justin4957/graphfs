# Example SPARQL Queries

This directory contains pre-built SPARQL queries you can run against the minimal-app knowledge graph.

## Quick Start

```bash
# Initialize and scan the project
graphfs init
graphfs scan

# Run a query from file
graphfs query --file queries/list-all-modules.sparql

# Get JSON output
graphfs query --file queries/list-all-modules.sparql --format json

# Get CSV output
graphfs query --file queries/list-all-modules.sparql --format csv
```

## Available Queries

### list-all-modules.sparql
Lists all modules with their descriptions.

```bash
graphfs query --file queries/list-all-modules.sparql
```

### list-service-modules.sparql
Lists all modules in the service layer.

```bash
graphfs query --file queries/list-service-modules.sparql
```

### find-dependencies.sparql
Shows all module dependencies.

```bash
graphfs query --file queries/find-dependencies.sparql
```

### list-exports.sparql
Lists all exports from each module.

```bash
graphfs query --file queries/list-exports.sparql
```

## Creating Your Own Queries

See [../examples/query-examples.md](../examples/query-examples.md) for more query examples and patterns.

### Basic Query Structure

```sparql
SELECT ?variable1 ?variable2 WHERE {
  ?subject <predicate-uri> ?object .
}
```

### Common Patterns

**Find all modules:**
```sparql
?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
```

**Filter by string matching:**
```sparql
FILTER(CONTAINS(?name, "auth"))
```

**Limit results:**
```sparql
} LIMIT 10
```

## Tips

1. Start with the pre-built queries and modify them
2. Use `LIMIT` to control output size during development
3. Use `--format json` for programmatic processing
4. Save complex queries to files in this directory
