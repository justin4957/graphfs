# scanner

Filesystem scanner for GraphFS with language detection and ignore pattern support.

## Overview

The `scanner` package recursively scans directories to find source code files, detect programming languages, and identify files with LinkedDoc metadata. It supports ignore patterns (.gitignore-style), concurrent scanning, and extensive filtering options.

## Features

- ✅ Recursive directory scanning
- ✅ Language detection for 14+ programming languages
- ✅ LinkedDoc detection (HasLinkedDoc flag)
- ✅ .gitignore-style ignore patterns
- ✅ Custom exclude/include patterns
- ✅ Concurrent scanning with worker pools
- ✅ File size limits
- ✅ Symlink handling
- ✅ Default ignore patterns (node_modules, vendor, .git, etc.)
- ✅ .gitignore and .graphfsignore file support

## Usage

### Basic Scanning

```go
package main

import (
    "fmt"
    "log"
    "github.com/justin4957/graphfs/pkg/scanner"
)

func main() {
    // Create scanner
    s := scanner.NewScanner()

    // Scan directory with default options
    result, err := s.Scan("examples/minimal-app", scanner.DefaultScanOptions())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d files (%d bytes) in %v\n",
        result.TotalFiles,
        result.TotalBytes,
        result.Duration)

    for _, file := range result.Files {
        fmt.Printf("  %s (%s, LinkedDoc: %v)\n",
            file.Path,
            file.Language,
            file.HasLinkedDoc)
    }
}
```

### Custom Scan Options

```go
opts := scanner.ScanOptions{
    ExcludePatterns: []string{"**/test/**", "*.test.go"},
    MaxFileSize:     500 * 1024, // 500KB
    FollowSymlinks:  false,
    UseDefaults:     true, // Use default ignore patterns
    Concurrent:      true, // Enable concurrent scanning
}

result, err := s.Scan("/path/to/project", opts)
```

### Scan Single File

```go
s := scanner.NewScanner()

fileInfo, err := s.ScanFile("examples/minimal-app/main.go")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Language: %s\n", fileInfo.Language)
fmt.Printf("Size: %d bytes\n", fileInfo.Size)
fmt.Printf("Has LinkedDoc: %v\n", fileInfo.HasLinkedDoc)
```

## Language Detection

The scanner automatically detects programming languages based on file extensions:

```go
language := scanner.DetectLanguage("main.go")        // Returns: "Go"
language = scanner.DetectLanguage("script.py")       // Returns: "Python"
language = scanner.DetectLanguage("Component.tsx")   // Returns: "TypeScript"
```

### Supported Languages

- **Go** (.go)
- **Python** (.py, .pyw)
- **JavaScript** (.js, .mjs, .cjs)
- **TypeScript** (.ts, .tsx)
- **Java** (.java)
- **Rust** (.rs)
- **C** (.c, .h)
- **C++** (.cpp, .cc, .cxx, .hpp, .hxx, .h++)
- **C#** (.cs)
- **Ruby** (.rb)
- **PHP** (.php)
- **Swift** (.swift)
- **Kotlin** (.kt, .kts)
- **Scala** (.scala)

### Register Custom Languages

```go
scanner.RegisterLanguage("elixir", "Elixir", []string{".ex", ".exs"})

lang := scanner.DetectLanguage("app.ex") // Returns: "Elixir"
```

## Ignore Patterns

### Default Ignore Patterns

The scanner includes sensible defaults to skip common directories:

- Version control: `.git`, `.svn`, `.hg`
- Dependencies: `node_modules`, `vendor`, `target`
- Build output: `dist`, `build`, `out`, `bin`
- IDE: `.idea`, `.vscode`, `.vs`
- OS files: `.DS_Store`, `Thumbs.db`
- Build artifacts: `*.exe`, `*.dll`, `*.so`, `*.pyc`

### Custom Ignore Patterns

```go
opts := scanner.ScanOptions{
    ExcludePatterns: []string{
        "**/test/**",      // Exclude test directories
        "*.test.go",       // Exclude test files
        "vendor",          // Exclude vendor directory
        "*.generated.go",  // Exclude generated files
    },
    UseDefaults: true, // Also use default patterns
}
```

### .gitignore Support

The scanner automatically reads `.gitignore` and `.graphfsignore` files:

```go
opts := scanner.DefaultScanOptions()
opts.IgnoreFiles = []string{".gitignore", ".graphfsignore", ".customignore"}

result, err := s.Scan("/path/to/project", opts)
```

**.graphfsignore example:**
```
# GraphFS specific ignores
*.generated.go
**/mocks/**
**/fixtures/**
```

## Scan Options

### ScanOptions

```go
type ScanOptions struct {
    IncludePatterns []string  // Patterns to include (not yet implemented)
    ExcludePatterns []string  // Patterns to exclude
    MaxFileSize     int64     // Maximum file size in bytes (default: 1MB)
    FollowSymlinks  bool      // Follow symbolic links (default: false)
    IgnoreFiles     []string  // Ignore files to read (default: [".gitignore", ".graphfsignore"])
    UseDefaults     bool      // Use default ignore patterns (default: true)
    Concurrent      bool      // Enable concurrent scanning (default: true)
}
```

### Default Options

```go
opts := scanner.DefaultScanOptions()
// Returns:
// - MaxFileSize: 1MB
// - FollowSymlinks: false
// - IgnoreFiles: [".gitignore", ".graphfsignore"]
// - UseDefaults: true
// - Concurrent: true
```

## Scan Results

### ScanResult

```go
type ScanResult struct {
    Files      []*FileInfo  // List of scanned files
    TotalFiles int          // Total number of files found
    TotalBytes int64        // Total size in bytes
    Errors     []error      // Any errors encountered
    Duration   time.Duration // Scan duration
}
```

### FileInfo

```go
type FileInfo struct {
    Path         string    // Absolute file path
    Language     string    // Detected language
    Size         int64     // File size in bytes
    ModTime      time.Time // Last modification time
    HasLinkedDoc bool      // Whether file has LinkedDoc metadata
}
```

## Performance

The scanner is optimized for large codebases:

- **Concurrent scanning**: Uses worker pools for parallel file processing
- **Efficient filtering**: Early exit for ignored directories
- **Memory efficient**: Streams file information, doesn't load entire files
- **Fast language detection**: O(1) extension lookup

### Benchmarks

Scanning `examples/minimal-app` (7 files):
- Sequential: ~10ms
- Concurrent: ~5ms

Scanning 1000+ files:
- < 1 second with concurrent mode
- < 5 seconds sequential mode

## Examples

### Example: Find All Go Files

```go
s := scanner.NewScanner()
result, _ := s.Scan("/path/to/project", scanner.DefaultScanOptions())

for _, file := range result.Files {
    if file.Language == "Go" {
        fmt.Println(file.Path)
    }
}
```

### Example: Find Files with LinkedDoc

```go
s := scanner.NewScanner()
result, _ := s.Scan("/path/to/project", scanner.DefaultScanOptions())

linkedDocFiles := 0
for _, file := range result.Files {
    if file.HasLinkedDoc {
        linkedDocFiles++
        fmt.Printf("%s has LinkedDoc metadata\n", file.Path)
    }
}

fmt.Printf("\nTotal files with LinkedDoc: %d/%d\n", linkedDocFiles, result.TotalFiles)
```

### Example: Scan with Size Limit

```go
opts := scanner.DefaultScanOptions()
opts.MaxFileSize = 100 * 1024 // 100KB limit

result, _ := s.Scan("/path/to/project", opts)
// Only files <= 100KB will be included
```

### Example: Custom Ignore Matcher

```go
matcher := scanner.NewIgnoreMatcher(scanner.DefaultIgnorePatterns())
matcher.AddPattern("**/test/**")
matcher.AddPattern("*.generated.go")

if matcher.ShouldIgnore("src/test/helper.go") {
    fmt.Println("File should be ignored")
}
```

## Testing

The scanner includes comprehensive tests:

```bash
# Run all tests
go test ./pkg/scanner/...

# Run with verbose output
go test ./pkg/scanner/... -v

# Run specific test
go test ./pkg/scanner/... -run TestScanner_Scan_MinimalApp

# Run benchmarks
go test ./pkg/scanner/... -bench=.
```

### Test Coverage

- ✅ Language detection for all supported languages
- ✅ Ignore pattern matching (exact, glob, directory)
- ✅ .gitignore file parsing
- ✅ Scanning minimal-app integration tests
- ✅ Concurrent vs sequential scanning
- ✅ File size limits
- ✅ Symlink handling
- ✅ Error handling

## Integration with GraphFS

The scanner is designed to work seamlessly with other GraphFS components:

```go
// 1. Scan codebase
scanner := scanner.NewScanner()
result, _ := scanner.Scan("/path/to/project", scanner.DefaultScanOptions())

// 2. Parse LinkedDoc from files with metadata
parser := parser.NewParser()
for _, file := range result.Files {
    if file.HasLinkedDoc {
        triples, _ := parser.Parse(file.Path)
        // Store triples in triple store...
    }
}

// 3. Build knowledge graph
// (See pkg/graph for graph building)
```

## Future Enhancements

- [ ] Include patterns support
- [ ] Watch mode for real-time scanning
- [ ] Progress callbacks for large scans
- [ ] Binary file detection
- [ ] More advanced glob patterns (**, ?)
- [ ] Negation patterns (!)
- [ ] Custom language detection via config

## License

MIT
