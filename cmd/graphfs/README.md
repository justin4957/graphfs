<!--
# Module: cmd/graphfs/README.md
Documentation for GraphFS CLI.

## Linked Modules
- [main](./main.go) - CLI entry point
- [root](./root.go) - Root command

## Tags
documentation, cli, readme

LinkedDoc RDF:
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#README.md> a code:Module ;
    code:name "cmd/graphfs/README.md" ;
    code:description "Documentation for GraphFS CLI" ;
    code:language "markdown" ;
    code:layer "cli" ;
    code:linksTo <./main.go>, <./root.go> ;
    code:tags "documentation", "cli", "readme" .
-->

# GraphFS CLI

Command-line interface for GraphFS - Semantic Code Filesystem Toolkit.

## Installation

```bash
go install github.com/justin4957/graphfs/cmd/graphfs@latest
```

## Commands

### graphfs init

Initialize GraphFS in a directory.

```bash
graphfs init [path]
```

**What it does:**
- Creates `.graphfs/` directory
- Creates `.graphfs/config.yaml` with default settings
- Creates `.graphfsignore` file for exclusion patterns
- Initializes empty triple store

**Examples:**
```bash
# Initialize in current directory
graphfs init

# Initialize in specific directory
graphfs init /path/to/project
```

### graphfs scan

Scan codebase and build knowledge graph.

```bash
graphfs scan [path] [options]
```

**Options:**
- `--include <pattern>` - Include files matching pattern
- `--exclude <pattern>` - Exclude files matching pattern
- `--validate` - Validate graph consistency
- `--stats` - Show detailed statistics
- `--output <file>` - Export graph to file

**Examples:**
```bash
# Scan current directory
graphfs scan

# Scan with validation
graphfs scan --validate

# Show detailed statistics
graphfs scan --stats

# Export graph to JSON
graphfs scan --output graph.json

# Scan specific directory with custom patterns
graphfs scan /path/to/project --include "**/*.go" --exclude "**/vendor/**"
```

### graphfs query

Execute SPARQL query against knowledge graph.

```bash
graphfs query <query> [options]
```

**Options:**
- `--file <path>` - Read query from file
- `--format <fmt>` - Output format: table, json, csv (default: table)
- `--limit <n>` - Limit number of results (default: 100)
- `--output <file>` - Write results to file

**Examples:**
```bash
# Inline query with table output
graphfs query 'PREFIX code: <https://schema.codedoc.org/> SELECT ?module WHERE { ?module a code:Module }'

# Query from file
graphfs query --file queries/modules.sparql

# Format as JSON
graphfs query 'SELECT * WHERE { ?s ?p ?o } LIMIT 10' --format json

# Save results to file
graphfs query --file queries/dependencies.sparql --output deps.csv --format csv
```

### graphfs version

Show version information.

```bash
graphfs version
```

## Configuration

GraphFS uses a YAML configuration file located at `.graphfs/config.yaml`.

**Default configuration:**
```yaml
version: 1
scan:
  include: ["**/*.go", "**/*.py", "**/*.js", "**/*.ts"]
  exclude: ["**/node_modules/**", "**/vendor/**", "**/.git/**"]
  max_file_size: 1048576  # 1MB
query:
  default_limit: 100
  timeout: 30s
```

**Override with CLI flags:**
```bash
graphfs scan --include "**/*.go" --exclude "**/test/**"
```

**Override with environment variables:**
```bash
export GRAPHFS_SCAN_INCLUDE="**/*.go"
export GRAPHFS_SCAN_EXCLUDE="**/vendor/**"
```

## .graphfsignore

Exclude files and directories from scanning by adding patterns to `.graphfsignore`:

```
# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/

# IDE files
.vscode/
.idea/
```

## Example Workflow

```bash
# 1. Initialize GraphFS in your project
cd my-project
graphfs init

# 2. Scan codebase and build knowledge graph
graphfs scan --validate --stats

# 3. Query the graph
graphfs query 'PREFIX code: <https://schema.codedoc.org/>
SELECT ?module ?description WHERE {
  ?module a code:Module ;
          code:description ?description .
}'

# 4. Find modules by tag
graphfs query 'PREFIX code: <https://schema.codedoc.org/>
SELECT ?module WHERE {
  ?module a code:Module ;
          code:tags "security" .
}'

# 5. Export graph for external analysis
graphfs scan --output graph.json
```

## Global Flags

- `--config <file>` - Config file (default: `.graphfs/config.yaml`)
- `--verbose, -v` - Verbose output
- `--no-color` - Disable colored output
- `--help, -h` - Help for any command
- `--version` - Show version information

## Exit Codes

- `0` - Success
- `1` - Error
- `2` - Invalid arguments

## Architecture

```
cmd/graphfs/
├── main.go          # Entry point
├── root.go          # Root command
├── config.go        # Configuration handling
├── cmd_init.go      # Init command implementation
├── cmd_scan.go      # Scan command implementation
├── cmd_query.go     # Query command implementation
├── output.go        # Output formatting utilities
├── cmd_test.go      # Test suite
└── README.md        # This file
```

## Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management
- [jedib0t/go-pretty](https://github.com/jedib0t/go-pretty) - Table formatting
- [schollz/progressbar](https://github.com/schollz/progressbar) - Progress bars

## Contributing

See main repository README for contribution guidelines.

## License

See main repository LICENSE file.
