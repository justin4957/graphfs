# GraphFS REPL Quick Start Guide

## Installation

After building the project, you have two options:

### Option 1: Use Local Binary (for development/testing)
```bash
# Build the binary
go build ./cmd/graphfs

# Run from the project directory
./graphfs repl
```

### Option 2: Install Globally
```bash
# Install to $GOPATH/bin
go install ./cmd/graphfs

# Run from anywhere
graphfs repl
```

## Quick Test

### 1. Start the REPL
```bash
./graphfs repl
```

You should see:
```
Scanning codebase and building graph...
Built graph with 35 modules, 683 triples

GraphFS Interactive REPL
Type .help for commands or enter SPARQL queries
Loaded graph with 35 modules

graphfs>
```

### 2. Try Some Commands

#### Show Statistics
```
graphfs> .stats
```

#### Show Help
```
graphfs> .help
```

#### Run a Simple Query
```
graphfs> SELECT * WHERE { ?s ?p ?o } LIMIT 5
      -> [press Enter on empty line to execute]
```

#### Change Output Format
```
graphfs> .format json
graphfs> SELECT * WHERE { ?s ?p ?o } LIMIT 3
      ->
```

#### Show Example Queries
```
graphfs> .examples
```

### 3. Exit the REPL
```
graphfs> .exit
```

Or press `Ctrl+D`

## Tips

- **Multi-line queries**: Start typing a SPARQL query (SELECT, CONSTRUCT, ASK, DESCRIBE), then press Enter on an empty line to execute
- **History**: Use Up/Down arrow keys to navigate through previous queries
- **Tab completion**: Press Tab to autocomplete commands and keywords
- **Save queries**: Use `.save filename.sparql` to save your last query
- **Load queries**: Use `.load filename.sparql` to load and execute a query file

## Example Session

```bash
$ ./graphfs repl

graphfs> .stats
Graph Statistics:
=================
Total Modules: 35
...

graphfs> SELECT * WHERE { ?s ?p ?o } LIMIT 3
      ->

o                                   | p                                   | s
----------------------------------------------------------------------------------------------------
https://schema.codedoc.org/Method   | http://www.w3.org/1999/02/22-rdf... | <#CryptoHelper.GenerateToken>
...

Query executed in 179.583µs
Returned 3 results

graphfs> .format json
Output format set to: json

graphfs> .exit
Goodbye!
```

## Troubleshooting

### "unknown command 'repl'" Error
This means you're running an old version of graphfs. Use `./graphfs repl` instead, or reinstall:
```bash
go install ./cmd/graphfs
```

### REPL Not Starting
Make sure you're in a directory with a `.graphfs/config.yaml` file, or the REPL will scan the current directory for code.

### No Results from Queries
Try simpler queries first:
```
SELECT * WHERE { ?s ?p ?o } LIMIT 10
```

This will show you what data is available in the graph.
