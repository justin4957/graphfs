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
	case ".modules":
		return r.cmdModules(args)
	case ".predicates":
		return r.cmdPredicates(args)
	case ".paginate":
		return r.cmdPaginate(args)
	case ".pagesize":
		return r.cmdPageSize(args)
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
  .paginate [on|off]  Toggle interactive pagination for large results
  .pagesize [N]       Set page size for pagination (default: 20)
  .load <file>        Load and execute query from file
  .save <file>        Save last query to file
  .history            Show query history
  .clear              Clear screen
  .schema             Show available predicates and types
  .examples           Show example queries
  .stats              Show graph statistics
  .modules            List all modules (with autocomplete support)
  .predicates         List all predicates (with autocomplete support)
  .exit               Exit REPL (or Ctrl+D)

Query Features:
  - Multi-line queries: Start typing a query and press Enter on empty line to execute
  - Tab completion: Press Tab for commands, keywords, modules, and predicates
  - History: Use Up/Down arrows to navigate query history
  - Ctrl+R: Reverse search through history
  - Syntax highlighting: Color-coded SPARQL queries for better readability
  - Interactive pagination: Navigate through large result sets page by page

Pagination Controls (when enabled):
  [n]ext     - Go to next page
  [p]rev     - Go to previous page
  [f]irst    - Go to first page
  [l]ast     - Go to last page
  [g]oto N   - Go to page N
  [q]uit     - Exit pagination

Examples:
  SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 10
  .format json
  .paginate on
  .pagesize 50
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

// cmdModules lists all available modules
func (r *REPL) cmdModules(args []string) error {
	modules := r.completer.GetModules()

	if len(modules) == 0 {
		r.printInfo("No modules found")
		return nil
	}

	r.printInfo(fmt.Sprintf("Available Modules (%d):", len(modules)))
	r.printInfo("====================")

	// Group by prefix or show first 50
	limit := 50
	if len(args) > 0 {
		// Filter by prefix
		prefix := args[0]
		filtered := FilterSuggestions(modules, prefix)
		if len(filtered) == 0 {
			r.printInfo(fmt.Sprintf("No modules matching '%s'", prefix))
			return nil
		}
		modules = filtered
		limit = len(modules)
	}

	count := 0
	for _, mod := range modules {
		if count >= limit {
			r.printInfo(fmt.Sprintf("\n... and %d more. Use '.modules <prefix>' to filter.", len(modules)-limit))
			break
		}
		fmt.Printf("  %s\n", mod)
		count++
	}

	r.printInfo("\nTip: Use these in queries like: SELECT ?x WHERE { <#module.go> ?p ?x }")
	return nil
}

// cmdPredicates lists all available predicates
func (r *REPL) cmdPredicates(args []string) error {
	predicates := r.completer.GetPredicates()

	if len(predicates) == 0 {
		r.printInfo("No predicates found")
		return nil
	}

	r.printInfo(fmt.Sprintf("Available Predicates (%d):", len(predicates)))
	r.printInfo("====================")

	// Group by type
	hashPredicates := make([]string, 0)
	codePredicates := make([]string, 0)
	otherPredicates := make([]string, 0)

	for _, pred := range predicates {
		if strings.HasPrefix(pred, "<#") {
			hashPredicates = append(hashPredicates, pred)
		} else if strings.HasPrefix(pred, "code:") {
			codePredicates = append(codePredicates, pred)
		} else {
			otherPredicates = append(otherPredicates, pred)
		}
	}

	if len(hashPredicates) > 0 {
		fmt.Println("\nShort form (#):")
		for _, pred := range hashPredicates {
			fmt.Printf("  %-30s\n", pred)
		}
	}

	if len(codePredicates) > 0 {
		fmt.Println("\nCode ontology (code:):")
		for _, pred := range codePredicates {
			fmt.Printf("  %-30s\n", pred)
		}
	}

	if len(otherPredicates) > 0 {
		fmt.Println("\nOther:")
		for _, pred := range otherPredicates {
			fmt.Printf("  %-30s\n", pred)
		}
	}

	r.printInfo("\nTip: Use these in queries like: SELECT ?x WHERE { ?x <#imports> ?dep }")
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

// cmdPaginate toggles interactive pagination
func (r *REPL) cmdPaginate(args []string) error {
	if len(args) == 0 {
		status := "off"
		if r.config.Paginate {
			status = "on"
		}
		r.printInfo(fmt.Sprintf("Pagination: %s (page size: %d)", status, r.config.PageSize))
		r.printInfo("Usage: .paginate [on|off]")
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "on", "true", "1", "yes":
		r.config.Paginate = true
		r.printSuccess("Pagination enabled")
	case "off", "false", "0", "no":
		r.config.Paginate = false
		r.printSuccess("Pagination disabled")
	default:
		return fmt.Errorf("invalid value: %s (use 'on' or 'off')", args[0])
	}

	return nil
}

// cmdPageSize sets the page size for pagination
func (r *REPL) cmdPageSize(args []string) error {
	if len(args) == 0 {
		r.printInfo(fmt.Sprintf("Current page size: %d", r.config.PageSize))
		r.printInfo("Usage: .pagesize <N>")
		return nil
	}

	var size int
	if _, err := fmt.Sscanf(args[0], "%d", &size); err != nil {
		return fmt.Errorf("invalid page size: %s", args[0])
	}

	if size < 1 {
		return fmt.Errorf("page size must be at least 1")
	}

	if size > 1000 {
		return fmt.Errorf("page size too large (max: 1000)")
	}

	r.config.PageSize = size
	r.printSuccess(fmt.Sprintf("Page size set to: %d", size))
	return nil
}
