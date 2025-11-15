# GraphFS REST API Testing Guide

## Prerequisites

1. Build GraphFS:
```bash
go build ./cmd/graphfs
```

2. Initialize a GraphFS project (if not already done):
```bash
./graphfs init
```

## Step-by-Step Testing Instructions

### Step 1: Start the Server

```bash
./graphfs serve
```

**Expected output:**
```
Scanning codebase and building graph...
Built graph with 30 modules, 626 triples
Starting GraphFS server on http://localhost:8080
SPARQL endpoint: http://localhost:8080/sparql
GraphQL endpoint: http://localhost:8080/graphql
GraphQL Playground: http://localhost:8080/graphql
REST API: http://localhost:8080/api/v1
```

### Step 2: Verify Server is Running

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{"status":"ok"}
```

### Step 3: Check Available Endpoints

```bash
curl http://localhost:8080/ | python3 -m json.tool
```

**Expected response includes REST API info:**
```json
{
  "name": "GraphFS API",
  "version": "0.2.0",
  "endpoints": {
    "rest": {
      "path": "/api/v1",
      "methods": ["GET"],
      "description": "RESTful API for common queries",
      "endpoints": {
        "modules": "/api/v1/modules",
        "search": "/api/v1/modules/search?q=query",
        "stats": "/api/v1/analysis/stats",
        "tags": "/api/v1/tags",
        "exports": "/api/v1/exports"
      }
    }
  }
}
```

## REST API Endpoint Tests

### 1. List All Modules

```bash
curl "http://localhost:8080/api/v1/modules" | python3 -m json.tool
```

**Returns:** Paginated list of all modules with metadata

**Example response:**
```json
{
  "data": [
    {
      "id": "<#scanner.go>",
      "name": "pkg/scanner/scanner.go",
      "path": "pkg/scanner/scanner.go",
      "description": "Filesystem scanner for GraphFS",
      "language": "go",
      "layer": "scanner",
      "tags": ["scanner", "filesystem"],
      "exports": ["#Scanner", "#NewScanner"],
      "links": {
        "self": "/api/v1/modules/<#scanner.go>",
        "dependencies": "/api/v1/modules/<#scanner.go>/dependencies",
        "dependents": "/api/v1/modules/<#scanner.go>/dependents"
      }
    }
  ],
  "meta": {
    "total": 30,
    "limit": 50,
    "offset": 0,
    "count": 30
  }
}
```

### 2. List Modules with Filters

**Filter by language:**
```bash
curl "http://localhost:8080/api/v1/modules?language=go" | python3 -m json.tool
```

**Filter by layer:**
```bash
curl "http://localhost:8080/api/v1/modules?layer=server" | python3 -m json.tool
```

**Filter by tag:**
```bash
curl "http://localhost:8080/api/v1/modules?tag=scanner" | python3 -m json.tool
```

**Combined filters with pagination:**
```bash
curl "http://localhost:8080/api/v1/modules?language=go&layer=server&limit=5&offset=0" | python3 -m json.tool
```

### 3. Get Specific Module

```bash
curl "http://localhost:8080/api/v1/modules/pkg/scanner/scanner.go" | python3 -m json.tool
```

**Returns:** Full module details including dependencies and dependents

### 4. Get Module Dependencies

```bash
curl "http://localhost:8080/api/v1/modules/pkg/scanner/scanner.go/dependencies" | python3 -m json.tool
```

**Returns:**
```json
{
  "module": { /* module details */ },
  "dependencies": [ /* array of dependency modules */ ],
  "count": 3
}
```

### 5. Get Module Dependents

```bash
curl "http://localhost:8080/api/v1/modules/pkg/scanner/scanner.go/dependents" | python3 -m json.tool
```

### 6. Search Modules

```bash
curl "http://localhost:8080/api/v1/modules/search?q=scanner" | python3 -m json.tool
```

**Searches in:** module name, description, path, and tags

**Returns:**
```json
{
  "query": "scanner",
  "results": [ /* matching modules */ ],
  "count": 3
}
```

### 7. Get Graph Statistics

```bash
curl "http://localhost:8080/api/v1/analysis/stats" | python3 -m json.tool
```

**Returns:**
```json
{
  "totalModules": 30,
  "totalTriples": 626,
  "totalRelationships": 44,
  "totalExports": 73,
  "totalTags": 54,
  "modulesByLanguage": [
    {"language": "go", "count": 30}
  ],
  "modulesByLayer": [
    {"layer": "server", "count": 9},
    {"layer": "scanner", "count": 3}
  ],
  "mostCommonLanguage": "go",
  "mostCommonLayer": "server",
  "averageDependencies": 1.47
}
```

### 8. Impact Analysis

```bash
curl "http://localhost:8080/api/v1/analysis/impact/pkg/scanner/scanner.go?depth=2" | python3 -m json.tool
```

**Returns:**
```json
{
  "module": { /* module being analyzed */ },
  "depth": 2,
  "impactedModules": [ /* modules that depend on this */ ],
  "impactCount": 5,
  "directDependents": 3
}
```

### 9. List All Tags

```bash
curl "http://localhost:8080/api/v1/tags" | python3 -m json.tool
```

**Returns:**
```json
{
  "tags": [
    {
      "tag": "scanner",
      "count": 3,
      "links": {
        "modules": "/api/v1/tags/scanner/modules"
      }
    }
  ],
  "total": 54
}
```

### 10. Get Modules by Tag

```bash
curl "http://localhost:8080/api/v1/tags/scanner/modules" | python3 -m json.tool
```

### 11. List All Exports

```bash
curl "http://localhost:8080/api/v1/exports" | python3 -m json.tool
```

### 12. Search Exports by Name

```bash
curl "http://localhost:8080/api/v1/exports?name=Scanner" | python3 -m json.tool
```

**Returns:**
```json
{
  "exports": [
    {
      "name": "#Scanner",
      "module": { /* module that exports this */ }
    }
  ],
  "count": 2
}
```

## Common Issues

### Issue: 404 Not Found

**Solution:** Make sure you're using the exact paths as shown above. The server distinguishes between:
- `/api/v1/modules` (list all)
- `/api/v1/modules/` (get by ID - requires trailing slash + ID)
- `/api/v1/modules/search` (search endpoint)

### Issue: Empty Results

**Cause:** Your codebase may not have LinkedDoc metadata yet.

**Solution:** Add LinkedDoc comments to your files. See the [GraphFS documentation](https://github.com/justin4957/graphfs) for examples.

### Issue: Server Not Responding

**Check:**
1. Is the server running? `curl http://localhost:8080/health`
2. Is the port already in use? Try a different port: `./graphfs serve --port 8081`

## Integration Examples

### Python

```python
import requests

# List modules
response = requests.get('http://localhost:8080/api/v1/modules', params={
    'language': 'go',
    'limit': 10
})
modules = response.json()['data']

print(f"Found {len(modules)} modules")
for mod in modules:
    print(f"- {mod['name']} ({mod['layer']})")
```

### JavaScript/Node.js

```javascript
const fetch = require('node-fetch');

async function getModules() {
  const response = await fetch('http://localhost:8080/api/v1/modules?language=go');
  const data = await response.json();

  console.log(`Total modules: ${data.meta.total}`);
  data.data.forEach(mod => {
    console.log(`- ${mod.name}`);
  });
}

getModules();
```

### Bash Script

```bash
#!/bin/bash

# Get all server modules
curl -s "http://localhost:8080/api/v1/modules?layer=server" \
  | python3 -m json.tool \
  | grep '"name"' \
  | cut -d'"' -f4

# Get statistics
curl -s "http://localhost:8080/api/v1/analysis/stats" \
  | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"Modules: {data['totalModules']}, Languages: {len(data['modulesByLanguage'])}\")"
```

## Testing Checklist

- [ ] Server starts successfully
- [ ] Health check returns OK
- [ ] List modules endpoint works
- [ ] Filtering by language works
- [ ] Filtering by layer works
- [ ] Filtering by tag works
- [ ] Pagination works (limit & offset)
- [ ] Get module by ID/path works
- [ ] Dependencies endpoint works
- [ ] Dependents endpoint works
- [ ] Search endpoint works
- [ ] Stats endpoint returns data
- [ ] Impact analysis works
- [ ] Tags endpoint works
- [ ] Exports endpoint works
- [ ] CORS headers present (check with browser devtools)
