# GraphFS Phase 1 Implementation Plan

## Overview

This document provides the complete Phase 1 implementation plan for GraphFS, starting with a minimal working example and progressing through all core infrastructure components.

## ✅ Completed: Minimal Example Application

Location: `examples/minimal-app/`

A fully functional minimal application with stubbed code but **working LinkedDoc headers** that demonstrate:
- Complete LinkedDoc+RDF format
- Realistic module relationships
- Security boundaries and architecture rules
- All semantic metadata types

### Structure
```
examples/minimal-app/
├── main.go                   # Entry point
├── models/
│   └── user.go              # Data model
├── services/
│   ├── auth.go              # Authentication (security-critical)
│   └── user.go              # User management (CRUD)
└── utils/
    ├── logger.go            # Logging utility
    ├── crypto.go            # Cryptographic utilities (security-critical)
    └── validator.go         # Data validation
```

### Build & Test
```bash
cd examples/minimal-app
go build -o minimal-app main.go
./minimal-app
```

All development and testing references this working filesystem from the start!

## Phase 1 GitHub Issues

All Phase 1 work has been broken down into 7 GitHub issues:

### [#1 - LinkedDoc+RDF Parser](https://github.com/justin4957/graphfs/issues/1)
**Priority:** HIGH | **Week 1 (Days 1-3)**

Parse LinkedDoc comment blocks and extract RDF/Turtle triples.

**Key Deliverables:**
- Extract `<!-- LinkedDoc RDF -->` sections from comments
- Parse Subject-Predicate-Object triples
- Support @prefix declarations
- Handle blank nodes and multi-line triples
- Test against all files in `examples/minimal-app/`

**API:**
```go
parser := parser.NewParser()
triples, err := parser.Parse("examples/minimal-app/main.go")
```

---

### [#2 - Filesystem Scanner](https://github.com/justin4957/graphfs/issues/2)
**Priority:** HIGH | **Week 1 (Days 2-4)**

Recursive filesystem scanner with language detection.

**Key Deliverables:**
- Recursive directory traversal
- Language detection (Go, Python, JS, etc.)
- .gitignore/.graphfsignore support
- Concurrent scanning for performance
- Scan 10,000+ files in < 5 seconds

**API:**
```go
scanner := scanner.NewScanner()
result, err := scanner.Scan("examples/minimal-app", opts)
```

---

### [#3 - In-Memory Triple Store](https://github.com/justin4957/graphfs/issues/3)
**Priority:** HIGH | **Week 1 (Days 3-5)**

RDF triple store with multiple indexing strategies.

**Key Deliverables:**
- Subject-Predicate-Object storage
- Multiple indexes (SPO, POS, OSP)
- Pattern matching with wildcards
- Store 100,000+ triples efficiently
- Insert 10,000 triples/second

**API:**
```go
store := store.NewTripleStore()
store.Add("main.go", "code:linksTo", "services/user.go")
results := store.Find("main.go", "code:linksTo", "")
```

---

### [#4 - Basic SPARQL Query Engine](https://github.com/justin4957/graphfs/issues/4)
**Priority:** HIGH | **Week 1-2 (Days 5-7)**

SPARQL 1.1 query engine with SELECT, FILTER, OPTIONAL.

**Key Deliverables:**
- SELECT queries with WHERE clause
- FILTER clause (=, !=, <, >, &&, ||, REGEX)
- OPTIONAL and UNION clauses
- LIMIT, OFFSET, ORDER BY, DISTINCT
- Execute queries in < 10ms

**Example Query:**
```sparql
PREFIX code: <https://schema.codedoc.org/>
PREFIX sec: <https://schema.codedoc.org/security/>

SELECT ?module ?description WHERE {
  ?module a code:Module ;
          sec:securityCritical true ;
          code:description ?description .
}
```

---

### [#5 - Graph Builder](https://github.com/justin4957/graphfs/issues/5)
**Priority:** MEDIUM | **Week 2 (Days 6-8)**

Orchestrates scanner, parser, and store to build knowledge graph.

**Key Deliverables:**
- Coordinate all components
- Build complete knowledge graph from directory
- Validate graph consistency
- Detect circular dependencies
- Generate statistics

**API:**
```go
builder := graph.NewBuilder(scanner, parser, store)
graph, err := builder.Build("examples/minimal-app", opts)
fmt.Printf("Built graph: %d modules, %d triples\n",
    graph.Statistics.TotalModules,
    graph.Statistics.TotalTriples)
```

---

### [#6 - CLI Commands (init, scan, query)](https://github.com/justin4957/graphfs/issues/6)
**Priority:** MEDIUM | **Week 2 (Days 9-14)**

Command-line interface with three core commands.

**Commands:**
```bash
# Initialize GraphFS
graphfs init [path]

# Scan codebase and build graph
graphfs scan [path] --validate --stats

# Execute SPARQL query
graphfs query 'SELECT * WHERE { ?s ?p ?o }' --format table
graphfs query --file queries/security.sparql --format json
```

**Key Deliverables:**
- `graphfs init` - Create `.graphfs/` directory and config
- `graphfs scan` - Build knowledge graph with progress bar
- `graphfs query` - Execute SPARQL with multiple output formats
- Configuration file support
- Beautiful table/JSON/CSV output

---

### [#7 - Documentation, Testing, Example Queries](https://github.com/justin4957/graphfs/issues/7)
**Priority:** MEDIUM | **Week 2 (Days 12-14)**

Complete Phase 1 with documentation and example queries.

**Key Deliverables:**
- 10+ example SPARQL queries in `.graphfs/queries/`
- User guide and developer guide
- 80%+ test coverage
- CI/CD pipeline (GitHub Actions)
- Performance benchmarks
- Release v0.1.0

**Example Queries:**
- List all modules
- Find dependencies
- Find security-critical modules
- Detect circular dependencies
- Impact analysis (transitive dependencies)
- Architecture rule violations

---

## Development Order

### Week 1: Core Components
1. **Days 1-3:** Parser (#1)
   - Start with simple triple parsing
   - Test incrementally with minimal-app files
   - Add complex features (blank nodes, prefixes)

2. **Days 2-4:** Scanner (#2) - Can parallelize with Parser
   - Basic directory traversal
   - Language detection
   - Ignore patterns

3. **Days 3-5:** Triple Store (#3)
   - Basic SPO storage
   - Add indexes incrementally
   - Performance testing

4. **Days 5-7:** SPARQL Engine (#4)
   - Start with basic SELECT
   - Add FILTER, OPTIONAL incrementally
   - Test against minimal-app queries

### Week 2: Integration & Polish
5. **Days 6-8:** Graph Builder (#5)
   - Integrate all components
   - Build graph from minimal-app
   - Validation and statistics

6. **Days 9-14:** CLI & Documentation (#6, #7)
   - Implement CLI commands
   - Write example queries
   - Documentation and testing
   - CI/CD setup

## Testing Strategy

### Unit Tests
- Each component tested independently
- Mock dependencies
- Edge cases and error conditions

### Integration Tests
All tests use `examples/minimal-app/` as test fixture:
```go
func TestBuildMinimalApp(t *testing.T) {
    builder := graph.NewBuilder(scanner, parser, store)
    graph, err := builder.Build("../../examples/minimal-app", opts)

    assert.NoError(t, err)
    assert.Equal(t, 8, graph.Statistics.TotalModules)
    assert.True(t, graph.Statistics.TotalTriples > 40)
}
```

### Query Tests
Test all example queries against minimal-app:
```go
func TestSecurityCriticalQuery(t *testing.T) {
    result, err := engine.Execute(`
        PREFIX sec: <https://schema.codedoc.org/security/>
        SELECT ?module WHERE {
            ?module sec:securityCritical true .
        }
    `)

    assert.Contains(t, result.Bindings, "services/auth.go")
    assert.Contains(t, result.Bindings, "utils/crypto.go")
}
```

## Success Criteria

Phase 1 is complete when:

- ✅ All 7 issues (#1-#7) closed
- ✅ `examples/minimal-app/` fully scannable and queryable
- ✅ All example queries return correct results
- ✅ 80%+ test coverage
- ✅ Documentation complete
- ✅ CI/CD passing
- ✅ Performance targets met:
  - Scan 1000+ files in < 10 seconds
  - Query execution < 10ms
  - Store 100,000+ triples
- ✅ Release v0.1.0 tagged

## Key Features Delivered

After Phase 1, users can:

1. **Add LinkedDoc headers** to their code (see examples/minimal-app for format)
2. **Initialize GraphFS** in their project: `graphfs init`
3. **Scan codebase**: `graphfs scan .`
4. **Query the graph** using SPARQL:
   ```bash
   graphfs query 'PREFIX code: <https://schema.codedoc.org/>
   SELECT ?module WHERE { ?module a code:Module }'
   ```
5. **Analyze dependencies**: Find what modules depend on what
6. **Find security-critical code**: Query by security metadata
7. **Detect circular dependencies**: Graph validation
8. **Impact analysis**: Understand downstream effects of changes

## Example Workflow

```bash
# Clone repo
git clone https://github.com/justin4957/graphfs.git
cd graphfs

# Build GraphFS
go build -o graphfs cmd/graphfs/main.go

# Try the minimal example
cd examples/minimal-app
../../graphfs init
../../graphfs scan . --validate --stats

# Query the graph
../../graphfs query 'PREFIX code: <https://schema.codedoc.org/>
SELECT ?module ?description WHERE {
  ?module a code:Module ;
          code:description ?description .
}'

# Use example queries
../../graphfs query --file ../../.graphfs/queries/find-security-critical.sparql

# Add to your own project
cd ~/my-project
# [Add LinkedDoc headers to your code]
graphfs init
graphfs scan .
graphfs query --file queries/find-dependencies.sparql
```

## Next Steps (Phase 2)

After Phase 1 completion:
- GraphQL API endpoint
- SPARQL HTTP endpoint
- REST API for common queries
- Interactive CLI mode
- Query result caching

See [ROADMAP.md](ROADMAP.md) for full Phase 2 details.

---

**Last Updated:** 2025-11-13
**Status:** Phase 1 issues created, minimal example complete
**Next Milestone:** Complete issue #1 (Parser)
