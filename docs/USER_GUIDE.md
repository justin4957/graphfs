# GraphFS User Guide

Welcome to GraphFS! This guide will help you get started with using GraphFS to build and query semantic knowledge graphs from your codebase.

## Table of Contents

1. [Installation](#installation)
2. [Getting Started](#getting-started)
3. [Adding LinkedDoc to Your Code](#adding-linkeddoc-to-your-code)
4. [Writing SPARQL Queries](#writing-sparql-queries)
5. [HTTP Server and API](#http-server-and-api)
6. [Common Use Cases](#common-use-cases)
7. [Troubleshooting](#troubleshooting)
8. [FAQ](#faq)

## Installation

### From Source

```bash
git clone https://github.com/justin4957/graphfs.git
cd graphfs
go install ./cmd/graphfs
```

### Verify Installation

```bash
graphfs --help
```

You should see the GraphFS help text with available commands.

## Getting Started

### Quick Start with Example Project

The fastest way to see GraphFS in action is to use the included example project:

```bash
# Navigate to the example
cd examples/minimal-app

# Initialize GraphFS
graphfs init

# Scan the codebase
graphfs scan --validate --stats

# Run a query
graphfs query --file queries/list-all-modules.sparql
```

### Initialize Your Own Project

```bash
# Navigate to your project
cd /path/to/your/project

# Initialize GraphFS
graphfs init

# This creates:
# - .graphfs/config.yaml - Configuration file
# - .graphfsignore - Patterns to ignore during scanning
# - .graphfs/store/ - Directory for future persistent storage
```

### Scan Your Codebase

```bash
# Basic scan
graphfs scan

# Scan with validation
graphfs scan --validate

# Scan with detailed statistics
graphfs scan --stats

# Export graph to JSON
graphfs scan --output graph.json
```

## Adding LinkedDoc to Your Code

LinkedDoc is a format for embedding RDF metadata in code comments. Here's how to add it to your code:

### Basic Module Documentation

```go
/*
# Module: services/auth.go
Authentication and authorization service.

## Linked Modules
- [crypto](../utils/crypto.go) - Encryption utilities
- [session](./session.go) - Session management

## Tags
security, authentication, api

## Exports
AuthService, ValidateToken, CreateSession

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#auth.go> a code:Module ;
    code:name "services/auth.go" ;
    code:description "Authentication and authorization service" ;
    code:language "go" ;
    code:layer "service" ;
    code:linksTo <../utils/crypto.go>, <./session.go> ;
    code:exports <#AuthService>, <#ValidateToken>, <#CreateSession> ;
    code:tags "security", "authentication", "api" .
<!-- End LinkedDoc RDF -->
*/
package services
```

### Module Properties

| Property | Description | Required |
|----------|-------------|----------|
| `code:name` | Module path/name | Yes |
| `code:description` | Brief description | Yes |
| `code:language` | Programming language | Yes |
| `code:layer` | Architecture layer | No |
| `code:linksTo` | Dependencies | No |
| `code:exports` | Public exports | No |
| `code:tags` | Tags for categorization | No |

### Supported Languages

GraphFS can detect and parse LinkedDoc from:
- Go (.go)
- Python (.py)
- JavaScript/TypeScript (.js, .ts, .tsx)
- Java (.java)
- Rust (.rs)
- And more...

### Best Practices

1. **Use Unique URIs**: Each module should have a unique URI (e.g., `<#services/auth.go>`)
2. **Document Dependencies**: List all module dependencies in `code:linksTo`
3. **Tag Appropriately**: Use tags for security-critical, deprecated, or experimental modules
4. **Keep Descriptions Clear**: Write concise, meaningful descriptions
5. **Maintain Consistency**: Use consistent naming and layering conventions

## Writing SPARQL Queries

### Basic Query Structure

```sparql
SELECT ?variable1 ?variable2 WHERE {
  ?subject <predicate-uri> ?object .
}
```

### Important: No PREFIX Support (Yet)

GraphFS doesn't currently support PREFIX declarations. Use full URIs:

**Don't do this:**
```sparql
PREFIX code: <https://schema.codedoc.org/>
SELECT ?module WHERE { ?module a code:Module . }
```

**Do this instead:**
```sparql
SELECT ?module WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
}
```

### Common Query Patterns

#### List All Modules

```sparql
SELECT ?module ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/description> ?description .
}
```

#### Find Modules by Layer

```sparql
SELECT ?module ?name ?layer WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/layer> ?layer .
  FILTER(CONTAINS(?layer, "service"))
}
```

#### Find Dependencies

```sparql
SELECT ?module ?dependency WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/linksTo> ?dependency .
}
```

#### Filter by Tag

```sparql
SELECT ?module ?name ?tag WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/tags> ?tag .
  FILTER(CONTAINS(?tag, "security"))
}
```

### Query Output Formats

```bash
# Table format (default)
graphfs query 'SELECT...'

# JSON format
graphfs query 'SELECT...' --format json

# CSV format
graphfs query 'SELECT...' --format csv

# Save to file
graphfs query 'SELECT...' --output results.json --format json
```

### Using Query Files

Save complex queries to files:

```bash
# Create query file
cat > queries/my-query.sparql << 'EOF'
SELECT ?module ?description WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/description> ?description .
}
EOF

# Execute from file
graphfs query --file queries/my-query.sparql
```

## HTTP Server and API

GraphFS can run as an HTTP server, exposing SPARQL query endpoints for remote access.

### Starting the Server

```bash
# Start server on default port (8080)
graphfs serve

# Start on custom port
graphfs serve --port 9000

# Start on all interfaces
graphfs serve --host 0.0.0.0 --port 8080
```

The server will:
1. Scan your codebase and build the knowledge graph
2. Keep the graph in memory
3. Expose HTTP endpoints for querying

### Available Endpoints

#### GET/POST /sparql
Execute SPARQL queries via HTTP.

**GET Request:**
```bash
curl "http://localhost:8080/sparql?query=SELECT+*+WHERE+{+?s+?p+?o+}+LIMIT+10"
```

**POST Request (recommended for complex queries):**
```bash
curl -X POST http://localhost:8080/sparql \
  -H "Content-Type: application/sparql-query" \
  -d 'SELECT ?module ?description WHERE {
    ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
    ?module <https://schema.codedoc.org/description> ?description .
  }'
```

**Query from file:**
```bash
curl -X POST http://localhost:8080/sparql \
  -H "Content-Type: application/sparql-query" \
  --data-binary @queries/list-all-modules.sparql
```

### Output Formats

The server supports multiple output formats via the `Accept` header or `?format` parameter:

#### JSON (default)
```bash
curl http://localhost:8080/sparql?query=...&format=json
# or
curl -H "Accept: application/sparql-results+json" http://localhost:8080/sparql?query=...
```

**Output:**
```json
{
  "head": {
    "vars": ["module", "description"]
  },
  "results": {
    "bindings": [
      {
        "module": {
          "type": "literal",
          "value": "<#main.go>"
        },
        "description": {
          "type": "literal",
          "value": "Main application entry point"
        }
      }
    ]
  }
}
```

#### CSV
```bash
curl http://localhost:8080/sparql?query=...&format=csv
```

**Output:**
```csv
module,description
<#main.go>,Main application entry point
<#auth.go>,Authentication service
```

#### TSV (Tab-Separated Values)
```bash
curl http://localhost:8080/sparql?query=...&format=tsv
```

**Output:**
```
module	description
<#main.go>	Main application entry point
<#auth.go>	Authentication service
```

#### XML
```bash
curl http://localhost:8080/sparql?query=...&format=xml
```

**Output:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<sparql xmlns="http://www.w3.org/2005/sparql-results#">
  <head>
    <variable name="module"/>
    <variable name="description"/>
  </head>
  <results>
    <result>
      <binding name="module">
        <literal><#main.go></literal>
      </binding>
      <binding name="description">
        <literal>Main application entry point</literal>
      </binding>
    </result>
  </results>
</sparql>
```

### CORS Support

The server includes CORS (Cross-Origin Resource Sharing) support enabled by default, allowing web applications to query the API from different domains.

### Health Check

```bash
curl http://localhost:8080/health
```

**Output:**
```json
{"status":"ok"}
```

### API Information

```bash
curl http://localhost:8080/
```

**Output:**
```json
{
  "name": "GraphFS API",
  "version": "0.2.0",
  "endpoints": {
    "sparql": {
      "path": "/sparql",
      "methods": ["GET", "POST"],
      "description": "SPARQL query endpoint",
      "formats": ["json", "csv", "tsv", "xml"]
    },
    "health": {
      "path": "/health",
      "methods": ["GET"],
      "description": "Health check endpoint"
    }
  }
}
```

### Integration Examples

#### Python
```python
import requests

# Query the GraphFS server
response = requests.post(
    'http://localhost:8080/sparql',
    headers={'Content-Type': 'application/sparql-query'},
    data='''
        SELECT ?module ?description WHERE {
            ?module <https://schema.codedoc.org/description> ?description .
        }
    '''
)

results = response.json()
for binding in results['results']['bindings']:
    print(f"{binding['module']['value']}: {binding['description']['value']}")
```

#### JavaScript/Node.js
```javascript
const fetch = require('node-fetch');

async function queryGraphFS() {
    const query = `
        SELECT ?module ?description WHERE {
            ?module <https://schema.codedoc.org/description> ?description .
        }
    `;

    const response = await fetch('http://localhost:8080/sparql', {
        method: 'POST',
        headers: {'Content-Type': 'application/sparql-query'},
        body: query
    });

    const data = await response.json();
    data.results.bindings.forEach(binding => {
        console.log(`${binding.module.value}: ${binding.description.value}`);
    });
}

queryGraphFS();
```

#### curl with jq for JSON processing
```bash
curl -s -X POST http://localhost:8080/sparql \
  -H "Content-Type: application/sparql-query" \
  -d 'SELECT ?module ?name WHERE { ?module <https://schema.codedoc.org/name> ?name }' \
  | jq '.results.bindings[] | "\(.module.value): \(.name.value)"'
```

## Common Use Cases

### 1. Understanding a New Codebase

```bash
# List all modules
graphfs query --file queries/list-all-modules.sparql

# See module organization by layer
graphfs query --file queries/module-stats.sparql

# Find entry points
graphfs query 'SELECT ?module ?name WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/tags> ?tag .
  FILTER(CONTAINS(?tag, "entrypoint"))
}'
```

### 2. Finding Security-Critical Code

```bash
graphfs query 'SELECT ?module ?name WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/tags> ?tag .
  FILTER(CONTAINS(?tag, "security"))
}'
```

### 3. Impact Analysis

```bash
# Find modules that depend on a specific module
graphfs query 'SELECT ?module ?dependency WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/linksTo> ?dependency .
  FILTER(CONTAINS(STR(?dependency), "auth.go"))
}'
```

### 4. Architecture Validation

```bash
# Find modules in wrong layers (e.g., UI accessing database)
graphfs query 'SELECT ?module ?name ?dependency WHERE {
  ?module <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.codedoc.org/Module> .
  ?module <https://schema.codedoc.org/name> ?name .
  ?module <https://schema.codedoc.org/layer> "ui" .
  ?module <https://schema.codedoc.org/linksTo> ?dependency .
  FILTER(CONTAINS(STR(?dependency), "database"))
}'
```

### 5. Documentation Generation

```bash
# Export entire graph for documentation tools
graphfs scan --output docs/graph.json

# Generate module list with descriptions
graphfs query --file queries/list-all-modules.sparql --format csv > docs/modules.csv
```

## Troubleshooting

### GraphFS Not Initialized

**Error:** `GraphFS not initialized. Run 'graphfs init' first`

**Solution:** Run `graphfs init` in your project directory.

### No Results Found

**Possible causes:**
1. No LinkedDoc metadata in scanned files
2. Query syntax error
3. Using PREFIX declarations (not supported)

**Solutions:**
- Verify files have LinkedDoc comments
- Check query syntax (use full URIs)
- Run `graphfs scan --validate` to check for errors

### Validation Errors

**Error:** `validation failed with N errors`

**Common issues:**
- Duplicate module URIs
- Missing required fields (name, description, language)
- Circular dependencies

**Solution:** Fix the LinkedDoc metadata in your files based on the error messages.

### Slow Scanning

**Possible causes:**
- Large codebase
- Too many files being scanned

**Solutions:**
- Add patterns to `.graphfsignore`
- Use `--exclude` flag to skip directories
- Check `.graphfsignore` is properly configured

## FAQ

### Q: Do I need to add LinkedDoc to every file?

**A:** No, only add LinkedDoc to files you want to include in the knowledge graph. The scanner only processes files with LinkedDoc metadata.

### Q: Can I use GraphFS with existing code?

**A:** Yes! GraphFS doesn't require any changes to your actual code, only adding LinkedDoc comments. You can add them incrementally.

### Q: What if my codebase has 1000+ files?

**A:** GraphFS is designed to handle large codebases. Use `.graphfsignore` to exclude generated files, vendor directories, etc. The concurrent scanner is optimized for performance.

### Q: Can I query across multiple projects?

**A:** Currently, GraphFS scans one project at a time. For multi-project queries, you'll need to combine the graphs manually or wait for Phase 2 features.

### Q: Is the graph persistent?

**A:** Currently, no. The graph is rebuilt on each query command. Persistent storage is planned for future phases.

### Q: Can I use this in CI/CD?

**A:** Yes! GraphFS can be used in CI/CD to validate architecture rules, check for security-critical changes, etc. See the GitHub Actions workflow for examples.

### Q: What query features are supported?

**A:** Currently supported:
- SELECT queries
- WHERE clause pattern matching
- FILTER with CONTAINS and string operations
- GROUP BY and COUNT
- LIMIT and OFFSET

**Not yet supported:**
- PREFIX declarations
- OPTIONAL clauses
- UNION
- Complex property paths (only `+` for transitive)

### Q: How do I contribute?

**A:** See CONTRIBUTING.md for guidelines. We welcome contributions!

## Next Steps

- Explore the [Query Examples](../examples/minimal-app/examples/query-examples.md)
- Read the [Developer Guide](DEVELOPER_GUIDE.md) if you want to contribute
- Check the [CHANGELOG](../CHANGELOG.md) for latest features
- Join our community and report issues on GitHub

---

**Need Help?** Open an issue on GitHub: https://github.com/justin4957/graphfs/issues
