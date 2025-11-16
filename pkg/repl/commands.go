/*
# Module: pkg/repl/commands.go
REPL command handlers.

Implements REPL commands like .help, .format, .load, etc.

## Linked Modules
- [repl](./repl.go) - REPL core

## Tags
repl, commands, cli

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#commands.go> a code:Module ;
    code:name "pkg/repl/commands.go" ;
    code:description "REPL command handlers" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <./repl.go> ;
    code:tags "repl", "commands", "cli" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// handleCommand processes REPL commands
func (r *REPL) handleCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case ".help":
		return r.cmdHelp(args)
	case ".format":
		return r.cmdFormat(args)
	case ".load":
		return r.cmdLoad(args)
	case ".save":
		return r.cmdSave(args)
	case ".history":
		return r.cmdHistory(args)
	case ".clear":
		return r.cmdClear(args)
	case ".schema":
		return r.cmdSchema(args)
	case ".examples":
		return r.cmdExamples(args)
	case ".stats":
		return r.cmdStats(args)
	case ".exit", ".quit":
		return io.EOF
	default:
		return fmt.Errorf("unknown command: %s (type .help for available commands)", cmd)
	}
}

// cmdHelp displays help information
func (r *REPL) cmdHelp(args []string) error {
	help := `
GraphFS REPL Commands:
=====================

Query Commands:
  SELECT ...           Execute a SPARQL SELECT query
  CONSTRUCT ...        Execute a SPARQL CONSTRUCT query
  ASK ...             Execute a SPARQL ASK query
  DESCRIBE ...        Execute a SPARQL DESCRIBE query

REPL Commands:
  .help               Show this help message
  .format [fmt]       Change output format (table, json, csv)
  .load <file>        Load and execute query from file
  .save <file>        Save last query to file
  .history            Show query history
  .clear              Clear screen
  .schema             Show available predicates and types
  .examples           Show example queries
  .stats              Show graph statistics
  .exit               Exit REPL (or Ctrl+D)

Query Features:
  - Multi-line queries: Start typing a query and press Enter on empty line to execute
  - Tab completion: Press Tab for command and keyword completion
  - History: Use Up/Down arrows to navigate query history
  - Colors: Syntax highlighting for better readability

Examples:
  SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 10
  .format json
  .load my-query.sparql
  .stats
`
	fmt.Println(help)
	return nil
}

// cmdFormat changes the output format
func (r *REPL) cmdFormat(args []string) error {
	if len(args) == 0 {
		r.printInfo(fmt.Sprintf("Current format: %s", r.format))
		r.printInfo("Available formats: table, json, csv")
		return nil
	}

	format := strings.ToLower(args[0])
	switch format {
	case "table", "json", "csv":
		r.format = format
		r.printSuccess(fmt.Sprintf("Output format set to: %s", format))
	default:
		return fmt.Errorf("unknown format: %s (available: table, json, csv)", format)
	}

	return nil
}

// cmdLoad loads and executes a query from a file
func (r *REPL) cmdLoad(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: .load <file>")
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	queryStr := string(data)
	r.printInfo(fmt.Sprintf("Loaded query from %s", filename))
	r.executeQuery(queryStr)

	return nil
}

// cmdSave saves the last query to a file
func (r *REPL) cmdSave(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: .save <file>")
	}

	if len(r.history) == 0 {
		return fmt.Errorf("no query in history to save")
	}

	filename := args[0]
	lastQuery := r.history[len(r.history)-1]

	if err := os.WriteFile(filename, []byte(lastQuery), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	r.printSuccess(fmt.Sprintf("Saved last query to %s", filename))
	return nil
}

// cmdHistory shows query history
func (r *REPL) cmdHistory(args []string) error {
	if len(r.history) == 0 {
		r.printInfo("No query history")
		return nil
	}

	r.printInfo("Query History:")
	r.printInfo("==============")
	for i, query := range r.history {
		fmt.Printf("%d: %s\n", i+1, truncate(query, 80))
	}

	return nil
}

// cmdClear clears the screen
func (r *REPL) cmdClear(args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

// cmdSchema shows available predicates and types
func (r *REPL) cmdSchema(args []string) error {
	r.printInfo("Schema Information:")
	r.printInfo("==================")

	// Collect unique predicates
	predicates := make(map[string]int)
	for _, mod := range r.graph.Modules {
		for range mod.Dependencies {
			predicates["code:dependsOn"]++
		}
		if mod.Language != "" {
			predicates["code:language"]++
		}
		if mod.Layer != "" {
			predicates["code:layer"]++
		}
		if len(mod.Tags) > 0 {
			predicates["code:tags"]++
		}
		if len(mod.Exports) > 0 {
			predicates["code:exports"]++
		}
	}

	fmt.Println("\nCommon Predicates:")
	for pred, count := range predicates {
		fmt.Printf("  %-30s (%d occurrences)\n", pred, count)
	}

	fmt.Println("\nCommon Patterns:")
	fmt.Println("  ?module a code:Module")
	fmt.Println("  ?module code:name ?name")
	fmt.Println("  ?module code:language ?lang")
	fmt.Println("  ?module code:layer ?layer")
	fmt.Println("  ?module code:dependsOn ?dependency")

	return nil
}

// cmdExamples shows example queries
func (r *REPL) cmdExamples(args []string) error {
	examples := `
Example Queries:
===============

1. List all modules:
   SELECT ?module ?name WHERE {
     ?module a code:Module .
     ?module code:name ?name
   }

2. Find modules by language:
   SELECT ?name ?desc WHERE {
     ?m a code:Module ;
        code:name ?name ;
        code:language "go" ;
        code:description ?desc
   } LIMIT 10

3. Find module dependencies:
   SELECT ?module ?dependency WHERE {
     ?m a code:Module ;
        code:name ?module ;
        code:dependsOn ?dep .
     ?dep code:name ?dependency
   }

4. Count modules by language:
   SELECT ?lang (COUNT(?m) as ?count) WHERE {
     ?m a code:Module ;
        code:language ?lang
   } GROUP BY ?lang

5. Find modules by tag:
   SELECT ?name WHERE {
     ?m a code:Module ;
        code:name ?name ;
        code:tags "cache"
   }

6. Find modules in a specific layer:
   SELECT ?name ?desc WHERE {
     ?m a code:Module ;
        code:name ?name ;
        code:layer "server" ;
        code:description ?desc
   }

Try copying and pasting these examples or modifying them!
`
	fmt.Println(examples)
	return nil
}

// cmdStats shows graph statistics
func (r *REPL) cmdStats(args []string) error {
	r.printInfo("Graph Statistics:")
	r.printInfo("=================")

	fmt.Printf("Total Modules: %d\n", len(r.graph.Modules))
	fmt.Printf("Total Triples: %d\n", r.graph.Store.Count())

	// Count by language
	langCounts := make(map[string]int)
	layerCounts := make(map[string]int)
	totalDeps := 0
	totalExports := 0

	for _, mod := range r.graph.Modules {
		if mod.Language != "" {
			langCounts[mod.Language]++
		}
		if mod.Layer != "" {
			layerCounts[mod.Layer]++
		}
		totalDeps += len(mod.Dependencies)
		totalExports += len(mod.Exports)
	}

	fmt.Println("\nModules by Language:")
	for lang, count := range langCounts {
		fmt.Printf("  %-15s: %d\n", lang, count)
	}

	fmt.Println("\nModules by Layer:")
	for layer, count := range layerCounts {
		fmt.Printf("  %-15s: %d\n", layer, count)
	}

	fmt.Printf("\nTotal Dependencies: %d\n", totalDeps)
	fmt.Printf("Total Exports: %d\n", totalExports)

	return nil
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
