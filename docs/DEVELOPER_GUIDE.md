# GraphFS Developer Guide

This guide is for developers who want to contribute to GraphFS or understand its internal architecture.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Component Descriptions](#component-descriptions)
3. [Development Setup](#development-setup)
4. [Adding New Features](#adding-new-features)
5. [Testing Guidelines](#testing-guidelines)
6. [Performance Optimization](#performance-optimization)
7. [Code Style](#code-style)

## Architecture Overview

GraphFS follows a layered architecture:

```
┌─────────────────────────────────────────┐
│           CLI Layer (cmd/)              │
│   (User Interface, Commands, Output)    │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────┴───────────────────────┐
│        Application Layer (pkg/)         │
│   (Graph Builder, Query Engine)         │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────┴───────────────────────┐
│      Core Services Layer (pkg/)         │
│   (Scanner, Parser, Query Executor)     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────┴───────────────────────┐
│     Infrastructure Layer (internal/)    │
│      (Triple Store, Indexing)           │
└─────────────────────────────────────────┘
```

### Data Flow

```
Source Code Files
      ↓
   Scanner (pkg/scanner)
      ↓
   Parser (pkg/parser) → RDF Triples
      ↓
   Graph Builder (pkg/graph) → Knowledge Graph
      ↓
   Triple Store (internal/store)
      ↓
   Query Engine (pkg/query) → Results
      ↓
   CLI Output (cmd/graphfs)
```

## Component Descriptions

### 1. Scanner (`pkg/scanner/`)

**Purpose:** Recursively scan directories to find source files with LinkedDoc metadata.

**Key Files:**
- `scanner.go` - Main scanner implementation
- `language.go` - Language detection
- `ignore.go` - .gitignore-style pattern matching

**Responsibilities:**
- Recursive directory traversal
- File type detection
- Ignore pattern matching (.gitignore, .graphfsignore)
- LinkedDoc detection
- Concurrent scanning for performance

**Extension Points:**
- Add new language detections in `language.go`
- Add custom ignore patterns in `ignore.go`

### 2. Parser (`pkg/parser/`)

**Purpose:** Parse RDF/Turtle triples from LinkedDoc comment blocks.

**Key Files:**
- `parser.go` - Main parser implementation
- `triple.go` - Triple data structure
- `lexer.go` - Tokenization

**Responsibilities:**
- Extract LinkedDoc blocks from comments
- Parse RDF/Turtle syntax
- Handle @prefix declarations
- Support URIs, literals, and blank nodes
- Error reporting with line numbers

**Extension Points:**
- Add new RDF syntax support
- Enhance error messages
- Support additional literal types

### 3. Triple Store (`internal/store/`)

**Purpose:** Efficient storage and retrieval of RDF triples.

**Key Files:**
- `store.go` - Triple store implementation
- `triple.go` - Triple structure
- `index.go` - Indexing structures

**Responsibilities:**
- SPO (Subject-Predicate-Object) storage
- Multiple indexes (SPO, POS, OSP) for fast lookups
- Pattern matching for queries
- Memory-efficient storage

**Extension Points:**
- Add persistent storage (RocksDB, PostgreSQL)
- Optimize indexing strategies
- Add caching layer

### 4. Graph Builder (`pkg/graph/`)

**Purpose:** Build semantic knowledge graph from RDF triples.

**Key Files:**
- `builder.go` - Graph construction orchestration
- `graph.go` - Graph data structure
- `module.go` - Module representation
- `validator.go` - Graph validation

**Responsibilities:**
- Coordinate scanner, parser, and store
- Extract modules from triples
- Build dependency relationships
- Validate graph consistency
- Collect statistics

**Extension Points:**
- Add incremental updates
- Enhance validation rules
- Add graph algorithms

### 5. Query Engine (`pkg/query/`)

**Purpose:** Execute SPARQL queries against the knowledge graph.

**Key Files:**
- `executor.go` - Query execution
- `parser.go` - SPARQL query parsing
- `query.go` - Query data structures

**Responsibilities:**
- Parse SPARQL queries
- Execute SELECT queries
- Pattern matching in WHERE clauses
- FILTER evaluation
- GROUP BY and aggregations
- Result binding and formatting

**Extension Points:**
- Add PREFIX support
- Implement OPTIONAL, UNION
- Add UPDATE queries
- Optimize query execution

### 6. CLI (`cmd/graphfs/`)

**Purpose:** User-facing command-line interface.

**Key Files:**
- `main.go` - Entry point
- `root.go` - Root command setup
- `cmd_init.go` - Init command
- `cmd_scan.go` - Scan command
- `cmd_query.go` - Query command
- `config.go` - Configuration management
- `output.go` - Output formatting

**Responsibilities:**
- Command-line argument parsing
- Configuration loading
- Progress reporting
- Result formatting (table, JSON, CSV)
- Error handling and user feedback

**Extension Points:**
- Add new commands
- Enhance output formats
- Add interactive mode

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional)

### Clone and Build

```bash
git clone https://github.com/justin4957/graphfs.git
cd graphfs
go mod download
go build ./cmd/graphfs
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./pkg/scanner
go test ./pkg/parser
go test ./pkg/query

# Run with race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Run Benchmarks

```bash
go test -bench=. ./pkg/scanner
go test -bench=. ./pkg/query
```

### Install Locally

```bash
go install ./cmd/graphfs
```

## Adding New Features

### Adding a New CLI Command

1. Create command file: `cmd/graphfs/cmd_mycommand.go`
2. Define command structure:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  `Detailed description`,
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringVar(&myFlag, "flag", "", "Flag description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

3. Add tests: `cmd/graphfs/cmd_test.go`
4. Update documentation

### Adding a New Query Feature

1. Update query parser in `pkg/query/parser.go`
2. Enhance executor in `pkg/query/executor.go`
3. Add data structures in `pkg/query/query.go`
4. Write tests in `pkg/query/executor_test.go`
5. Add example queries

### Adding Language Support

1. Update language detection in `pkg/scanner/language.go`:

```go
func detectLanguage(filename string) string {
    ext := filepath.Ext(filename)
    switch ext {
    case ".newlang":
        return "newlang"
    // ...
    }
}
```

2. Update comment extraction in `pkg/parser/parser.go`
3. Test with sample files

## Testing Guidelines

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Setup
    setup := createTestSetup()
    defer cleanup(setup)

    // Execute
    result := featureUnderTest(setup)

    // Verify
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Test Categories

1. **Unit Tests** - Test individual functions/methods
2. **Integration Tests** - Test component interactions
3. **Example Tests** - Executable examples (Example_xxx functions)
4. **Benchmark Tests** - Performance tests (Benchmark_xxx functions)

### Test Coverage Goals

- Overall: 80%+
- Core packages (scanner, parser, graph, query): 90%+
- CLI commands: 70%+
- Internal packages: 80%+

### Running Specific Tests

```bash
# Run single test
go test -run TestScannerBasic ./pkg/scanner

# Run tests matching pattern
go test -run "Scanner.*" ./pkg/scanner

# Run with short flag (skip slow tests)
go test -short ./...
```

## Performance Optimization

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./pkg/query
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./pkg/scanner
go tool pprof mem.prof

# Generate profile visualization
go tool pprof -http=:8080 cpu.prof
```

### Performance Tips

1. **Scanner Optimization**
   - Use concurrent scanning for large codebases
   - Efficient ignore pattern matching
   - Stream file reading instead of loading entire files

2. **Query Optimization**
   - Use appropriate indexes in triple store
   - Implement query planning for complex queries
   - Cache frequently used query results

3. **Memory Optimization**
   - String interning for repeated URIs
   - Efficient data structures (maps vs slices)
   - Avoid unnecessary allocations

### Benchmarking

```go
func BenchmarkScanLargeDirectory(b *testing.B) {
    // Setup
    dir := createLargeTempDirectory()
    defer os.RemoveAll(dir)

    scanner := NewScanner()
    opts := DefaultScanOptions()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanner.Scan(dir, opts)
    }
}
```

## Code Style

### General Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write clear, self-documenting code
- Add comments for exported functions/types
- Keep functions small and focused

### Naming Conventions

- **Packages**: Short, lowercase, no underscores (e.g., `scanner`, `query`)
- **Files**: Lowercase with underscores (e.g., `triple_store.go`)
- **Types**: PascalCase (e.g., `TripleStore`, `QueryResult`)
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase (e.g., `moduleName`, `queryResult`)
- **Constants**: PascalCase or ALL_CAPS for exported

### Documentation

All exported types, functions, and methods must have documentation:

```go
// Scanner recursively scans directories for source files with LinkedDoc metadata.
// It respects .gitignore and .graphfsignore patterns and supports concurrent scanning.
type Scanner struct {
    parser *parser.Parser
}

// Scan performs a recursive directory scan starting from rootPath.
// It returns a ScanResult containing all discovered files with LinkedDoc metadata.
// The opts parameter configures include/exclude patterns, file size limits, etc.
func (s *Scanner) Scan(rootPath string, opts ScanOptions) (*ScanResult, error) {
    // Implementation
}
```

### Error Handling

```go
// Good - wrap errors with context
if err != nil {
    return fmt.Errorf("failed to parse file %s: %w", filename, err)
}

// Bad - lose error context
if err != nil {
    return err
}
```

### Testing Best Practices

```go
// Use table-driven tests
func TestModuleParsing(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *Module
        wantErr  bool
    }{
        {
            name:  "valid module",
            input: "...",
            expected: &Module{...},
            wantErr: false,
        },
        // More test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := parseModule(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Contributing Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and add tests
4. Run tests: `go test ./...`
5. Run linter: `golangci-lint run`
6. Commit with clear messages
7. Push to your fork
8. Create a pull request

### Commit Messages

```
type(scope): Short description

Longer description if needed.

- Bullet points for details
- Another detail

Closes #123
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`

## Release Process

1. Update CHANGELOG.md
2. Update version in code
3. Create git tag: `git tag v0.x.0`
4. Push tag: `git push origin v0.x.0`
5. GitHub Actions builds and creates release
6. Update documentation

## Getting Help

- Read the [User Guide](USER_GUIDE.md)
- Check [existing issues](https://github.com/justin4957/graphfs/issues)
- Ask questions in discussions
- Join our community

---

**Happy Contributing!** 🎉
