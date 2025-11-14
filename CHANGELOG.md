# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-11-14

### Added - Phase 1 Complete

#### Core Infrastructure
- RDF/Turtle parser for LinkedDoc headers (#3)
- Filesystem scanner with language detection and ignore patterns (#2)
- In-memory triple store with efficient indexing (#1)
- SPARQL query engine with SELECT, WHERE, FILTER, GROUP BY support (#4)
- Knowledge graph builder with validation (#5)
- Complete CLI tool with init, scan, and query commands (#6)

#### CLI Features
- `graphfs init` - Initialize GraphFS in a directory
- `graphfs scan` - Build knowledge graph with progress reporting and validation
- `graphfs query` - Execute SPARQL queries with multiple output formats
- Configuration system with YAML support
- Beautiful table output, JSON, and CSV export
- Progress bars and user-friendly error messages

#### Documentation
- Comprehensive README with quick start guide
- Query examples guide with 9+ working SPARQL examples
- Pre-built query files in examples/minimal-app/queries/
- Interactive demo script (try-it.sh)
- Complete API documentation (GoDoc)
- CLI documentation in cmd/graphfs/README.md

#### Testing
- Unit tests for all core components
- Integration tests with examples/minimal-app
- Test coverage across scanner, parser, graph builder, and query engine
- CI/CD pipeline with GitHub Actions

#### Example Queries
- list-all-modules.sparql - List modules with descriptions
- list-service-modules.sparql - Filter modules by layer
- find-dependencies.sparql - Show module dependencies
- list-exports.sparql - List all exports
- find-by-tag.sparql - Find modules by tag
- find-leaf-modules.sparql - Find modules with no dependencies
- module-stats.sparql - Statistics by layer
- find-functions.sparql - List all functions

### Implementation Details

#### Scanner (#2)
- Recursive directory traversal
- .gitignore and .graphfsignore support
- Language detection for multiple file types
- LinkedDoc detection
- Concurrent scanning for performance
- File size limits and ignore patterns

#### Parser (#3)
- RDF/Turtle syntax support
- LinkedDoc comment block extraction
- @prefix declarations
- URI, literal, and blank node support
- Error handling with line numbers

#### Triple Store (#1)
- Subject-Predicate-Object indexing
- Efficient query matching
- Multiple index structures (SPO, POS, OSP)
- Memory-efficient storage

#### Query Engine (#4)
- SPARQL SELECT queries
- WHERE clause pattern matching
- FILTER with CONTAINS, string operations
- GROUP BY and COUNT aggregations
- LIMIT and OFFSET support
- Variable binding and substitution

#### Graph Builder (#5)
- Module extraction from RDF triples
- Dependency graph construction
- Reverse dependency tracking
- Graph validation with error reporting
- Statistics collection (modules, triples, relationships)
- Circular dependency detection

#### CLI (#6)
- Cobra framework integration
- Viper configuration management
- Beautiful table formatting with go-pretty
- Progress bars with schollz/progressbar
- Multiple output formats (table, JSON, CSV)
- Comprehensive help text

### Known Limitations

- PREFIX declarations not yet supported in queries (use full URIs)
- Literal value matching in WHERE clauses has limitations (use FILTER instead)
- No persistent storage (graph rebuilt on each query)
- No incremental updates (full scan required)

### Phase 1 Success Metrics

✅ All core features implemented
✅ examples/minimal-app fully queryable
✅ 8+ example queries working
✅ Documentation complete
✅ CLI polished and user-friendly
✅ Ready for Phase 2

## [Unreleased]

### Planned for Phase 2
- SPARQL HTTP endpoint
- GraphQL schema generation and server
- REST API for common queries
- Query result caching
- PREFIX support in queries
- Additional SPARQL features (OPTIONAL, UNION, etc.)

### Planned for Phase 3
- Dependency graph analysis
- Impact analysis engine
- Architecture rule validation
- Dead code detection
- Security boundary analysis

### Planned for Phase 4
- Markdown documentation generator
- SVG/GraphViz diagram generation
- Interactive HTML explorer
- API reference generator

[0.1.0]: https://github.com/justin4957/graphfs/releases/tag/v0.1.0
