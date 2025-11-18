/*
# Module: pkg/repl/repl.go
Interactive REPL for GraphFS queries.

Provides an interactive Read-Eval-Print Loop for exploring the knowledge graph
with SPARQL queries, syntax highlighting, and tab completion.

## Linked Modules
- [../query](../query/executor.go) - Query executor
- [../graph](../graph/graph.go) - Graph data structure

## Tags
repl, interactive, cli, sparql

## Exports
REPL, Config, New

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#repl.go> a code:Module ;
    code:name "pkg/repl/repl.go" ;
    code:description "Interactive REPL for GraphFS queries" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <../query/executor.go>, <../graph/graph.go> ;
    code:exports <#REPL>, <#Config>, <#New> ;
    code:tags "repl", "interactive", "cli", "sparql" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
)

// Config holds REPL configuration
type Config struct {
	HistoryFile string
	Prompt      string
	NoColor     bool
}

// REPL is the interactive Read-Eval-Print Loop
type REPL struct {
	config      *Config
	executor    *query.Executor
	graph       *graph.Graph
	rl          *readline.Instance
	format      string
	history     []string
	completer   *Completer
	highlighter *Highlighter
}

// New creates a new REPL instance
func New(executor *query.Executor, g *graph.Graph, config *Config) (*REPL, error) {
	if config == nil {
		config = &Config{
			HistoryFile: filepath.Join(os.TempDir(), ".graphfs_history"),
			Prompt:      "graphfs> ",
			NoColor:     false,
		}
	}

	// Configure readline
	rlConfig := &readline.Config{
		Prompt:          config.Prompt,
		HistoryFile:     config.HistoryFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %w", err)
	}

	// Create completer and highlighter
	completer := NewCompleter(g)
	highlighter := NewHighlighter(config.NoColor)

	repl := &REPL{
		config:      config,
		executor:    executor,
		graph:       g,
		rl:          rl,
		format:      "table",
		history:     make([]string, 0),
		completer:   completer,
		highlighter: highlighter,
	}

	// Set up autocomplete
	repl.setupAutocomplete()

	return repl, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	defer r.rl.Close()

	r.printWelcome()

	var multilineQuery strings.Builder
	inMultiline := false

	for {
		var line string
		var err error

		if inMultiline {
			r.rl.SetPrompt("      -> ")
			line, err = r.rl.Readline()
		} else {
			r.rl.SetPrompt(r.config.Prompt)
			line, err = r.rl.Readline()
		}

		if err != nil {
			if err == readline.ErrInterrupt {
				if inMultiline {
					multilineQuery.Reset()
					inMultiline = false
					continue
				}
				if len(line) == 0 {
					break
				}
			} else if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)

		// Handle empty lines
		if line == "" {
			if inMultiline {
				// Execute the multiline query
				queryStr := multilineQuery.String()
				multilineQuery.Reset()
				inMultiline = false
				r.rl.SetPrompt(r.config.Prompt)
				r.executeQuery(queryStr)
			}
			continue
		}

		// Handle REPL commands
		if strings.HasPrefix(line, ".") {
			if inMultiline {
				r.printError("Cannot use commands in multiline mode. Press Enter on empty line to execute query.")
				continue
			}
			if err := r.handleCommand(line); err != nil {
				if err == io.EOF {
					break
				}
				r.printError(err.Error())
			}
			continue
		}

		// Check if starting multiline query
		if !inMultiline && (strings.HasPrefix(strings.ToUpper(line), "SELECT") ||
			strings.HasPrefix(strings.ToUpper(line), "CONSTRUCT") ||
			strings.HasPrefix(strings.ToUpper(line), "ASK") ||
			strings.HasPrefix(strings.ToUpper(line), "DESCRIBE")) {
			inMultiline = true
			multilineQuery.WriteString(line)
			multilineQuery.WriteString("\n")
			continue
		}

		// Continue building multiline query
		if inMultiline {
			multilineQuery.WriteString(line)
			multilineQuery.WriteString("\n")
			continue
		}

		// Single line query
		r.executeQuery(line)
	}

	r.printGoodbye()
	return nil
}

// executeQuery executes a SPARQL query and displays results
func (r *REPL) executeQuery(queryStr string) {
	queryStr = strings.TrimSpace(queryStr)
	if queryStr == "" {
		return
	}

	// Add to history
	r.history = append(r.history, queryStr)

	// Execute query with timing
	start := time.Now()
	result, err := r.executor.ExecuteString(queryStr)
	duration := time.Since(start)

	if err != nil {
		r.printError(fmt.Sprintf("Query error: %v", err))
		return
	}

	// Format and display results
	if err := r.formatResult(result); err != nil {
		r.printError(fmt.Sprintf("Format error: %v", err))
		return
	}

	// Print execution time and result count
	r.printInfo(fmt.Sprintf("Query executed in %v", duration))
	if result != nil && result.Bindings != nil {
		r.printInfo(fmt.Sprintf("Returned %d results", len(result.Bindings)))
	}
}

// setupAutocomplete configures tab completion
func (r *REPL) setupAutocomplete() {
	r.rl.Config.AutoComplete = r.completer.GetCompleter()
}

// printWelcome displays the welcome message
func (r *REPL) printWelcome() {
	if r.config.NoColor {
		fmt.Println("GraphFS Interactive REPL")
		fmt.Println("Type .help for commands or enter SPARQL queries")
		fmt.Printf("Loaded graph with %d modules\n", len(r.graph.Modules))
		fmt.Println()
		fmt.Println("Features:")
		fmt.Println("  - Tab completion for commands, keywords, modules, and predicates")
		fmt.Println("  - Multi-line query editing")
		fmt.Println("  - Query history with Up/Down arrows and Ctrl+R search")
		fmt.Println("  - Syntax highlighting")
		fmt.Println()
	} else {
		cyan := color.New(color.FgCyan, color.Bold)
		cyan.Println("GraphFS Interactive REPL")
		fmt.Println("Type .help for commands or enter SPARQL queries")
		fmt.Printf("Loaded graph with %d modules\n", len(r.graph.Modules))
		fmt.Println()
		green := color.New(color.FgGreen)
		green.Println("Features:")
		fmt.Println("  - Tab completion for commands, keywords, modules, and predicates")
		fmt.Println("  - Multi-line query editing")
		fmt.Println("  - Query history with Up/Down arrows and Ctrl+R search")
		fmt.Println("  - Syntax highlighting")
		fmt.Println()
	}
}

// printGoodbye displays the goodbye message
func (r *REPL) printGoodbye() {
	fmt.Println("\nGoodbye!")
}

// printError displays an error message
func (r *REPL) printError(msg string) {
	if r.config.NoColor {
		fmt.Fprintf(r.rl.Stderr(), "Error: %s\n", msg)
	} else {
		red := color.New(color.FgRed)
		red.Fprintf(r.rl.Stderr(), "Error: %s\n", msg)
	}
}

// printInfo displays an info message
func (r *REPL) printInfo(msg string) {
	if r.config.NoColor {
		fmt.Println(msg)
	} else {
		cyan := color.New(color.FgCyan)
		cyan.Println(msg)
	}
}

// printSuccess displays a success message
func (r *REPL) printSuccess(msg string) {
	if r.config.NoColor {
		fmt.Println(msg)
	} else {
		green := color.New(color.FgGreen)
		green.Println(msg)
	}
}
