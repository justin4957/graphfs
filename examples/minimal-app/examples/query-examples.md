# GraphFS Query Examples

This document provides working examples of SPARQL queries for the minimal-app project.

**Note**: PREFIX declarations are not yet supported in the current implementation. Use full URIs for now.

## Basic Queries

### 1. List All Modules

Get all modules with their descriptions:

```sparql
SELECT ?module ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/description> ?description .
}
```

### 2. List Modules by Layer

Find all service layer modules:

```sparql
SELECT ?module ?name ?layer WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/layer> ?layer .
  FILTER(CONTAINS(?layer, "service"))
}
```

### 3. Find Modules with Specific Tags

Find all modules tagged with "security":

```sparql
SELECT ?module ?name WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/tags> ?tag .
  FILTER(CONTAINS(?tag, "security"))
}
```

### 4. Find Module Dependencies

Find all modules that link to the auth service:

```sparql
SELECT ?module ?dependency WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/linksTo> ?dependency .
  FILTER(CONTAINS(STR(?dependency), "auth.go"))
}
```

### 5. List All Utility Modules

```sparql
SELECT ?module ?name ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/description> ?description .
  ?module <https://schema.codedoc.org/layer> "utility" .
}
```

### 6. Count Modules by Language

```sparql
SELECT ?language (COUNT(?module) as ?count) WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/language> ?language .
}
GROUP BY ?language
```

## Advanced Queries

### 7. Find All Module Exports

```sparql
SELECT ?module ?export WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/exports> ?export .
}
```

### 8. Find Functions and Their Descriptions

```sparql
SELECT ?function ?name ?description WHERE {
  ?function <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Function> .
  ?function <https://schema.codedoc.org/name> ?name .
  ?function <https://schema.codedoc.org/description> ?description .
}
```

### 9. Find All Types (Structs/Classes)

```sparql
SELECT ?type ?name ?kind WHERE {
  ?type <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Type> .
  ?type <https://schema.codedoc.org/name> ?name .
  ?type <https://schema.codedoc.org/kind> ?kind .
}
```

## Using Query Files

Save queries to files and execute them:

```bash
# Save query to file
cat > queries/list-modules.sparql << 'SPARQL'
SELECT ?module ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/description> ?description .
}
SPARQL

# Execute from file
graphfs query --file queries/list-modules.sparql

# With JSON output
graphfs query --file queries/list-modules.sparql --format json

# With CSV output
graphfs query --file queries/list-modules.sparql --format csv
```

## Output Formats

### Table (default)

```bash
graphfs query 'SELECT ?module ?name WHERE { 
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> . 
  ?module <https://schema.codedoc.org/name> ?name . 
} LIMIT 3'
```

### JSON

```bash
graphfs query 'SELECT ?module ?name WHERE { 
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> . 
  ?module <https://schema.codedoc.org/name> ?name . 
} LIMIT 3' --format json
```

### CSV

```bash
graphfs query 'SELECT ?module ?name WHERE { 
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> . 
  ?module <https://schema.codedoc.org/name> ?name . 
} LIMIT 3' --format csv
```

## Common URI Prefixes

Since PREFIX declarations are not yet supported, here are the full URIs you'll need:

- **RDF type**: `<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>`
- **Code schema base**: `<https://schema.codedoc.org/>`
- **Code Module**: `<https://schema.codedoc.org/Module>`
- **Code properties**:
  - `<https://schema.codedoc.org/name>`
  - `<https://schema.codedoc.org/description>`
  - `<https://schema.codedoc.org/language>`
  - `<https://schema.codedoc.org/layer>`
  - `<https://schema.codedoc.org/tags>`
  - `<https://schema.codedoc.org/linksTo>`
  - `<https://schema.codedoc.org/exports>`

## Tips

1. **Use LIMIT** to control result size: `LIMIT 10`
2. **Use FILTER** for string matching: `FILTER(CONTAINS(?name, "auth"))`
3. **Use DISTINCT** to remove duplicates: `SELECT DISTINCT ?module WHERE ...`
4. **Save complex queries** to files in the `queries/` directory
5. **Export results** to JSON for programmatic processing

## Coming Soon (Phase 2)

- PREFIX support for shorter queries
- More SPARQL features (OPTIONAL, UNION, etc.)
- GraphQL queries
- REST API endpoints
