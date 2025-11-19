/*
# Module: pkg/repl/completer.go
Autocomplete functionality for REPL.

Provides intelligent autocomplete for SPARQL keywords, module paths,
predicates, and REPL commands with context-aware suggestions.

## Linked Modules
- [repl](./repl.go) - REPL core
- [../graph](../graph/graph.go) - Graph data structure

## Tags
repl, autocomplete, completion

## Exports
Completer, NewCompleter

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#completer.go> a code:Module ;
    code:name "pkg/repl/completer.go" ;
    code:description "Autocomplete functionality for REPL" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <./repl.go>, <../graph/graph.go> ;
    code:exports <#Completer>, <#NewCompleter> ;
    code:tags "repl", "autocomplete", "completion" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"strings"
	"unicode"

	"github.com/chzyer/readline"
	"github.com/justin4957/graphfs/pkg/graph"
)

// Completer provides autocomplete functionality
type Completer struct {
	graph      *graph.Graph
	commands   []readline.PrefixCompleterInterface
	keywords   []string
	predicates []string
	modules    []string
}

// NewCompleter creates a new completer
func NewCompleter(g *graph.Graph) *Completer {
	c := &Completer{
		graph:    g,
		keywords: getSPARQLKeywords(),
	}

	// Build module and predicate lists
	c.buildModuleList()
	c.buildPredicateList()
	c.buildCommandList()

	return c
}

// buildCommandList creates the command autocomplete tree
func (c *Completer) buildCommandList() {
	c.commands = []readline.PrefixCompleterInterface{
		// REPL commands
		readline.PcItem(".help"),
		readline.PcItem(".format",
			readline.PcItem("table"),
			readline.PcItem("json"),
			readline.PcItem("csv"),
		),
		readline.PcItem(".load"),
		readline.PcItem(".save"),
		readline.PcItem(".history"),
		readline.PcItem(".clear"),
		readline.PcItem(".schema"),
		readline.PcItem(".examples"),
		readline.PcItem(".stats"),
		readline.PcItem(".modules"),
		readline.PcItem(".predicates"),
		readline.PcItem(".exit"),
		readline.PcItem(".quit"),

		// SPARQL keywords
		readline.PcItem("SELECT"),
		readline.PcItem("SELECT DISTINCT"),
		readline.PcItem("WHERE"),
		readline.PcItem("CONSTRUCT"),
		readline.PcItem("ASK"),
		readline.PcItem("DESCRIBE"),
		readline.PcItem("PREFIX"),
		readline.PcItem("FILTER"),
		readline.PcItem("OPTIONAL"),
		readline.PcItem("UNION"),
		readline.PcItem("LIMIT"),
		readline.PcItem("OFFSET"),
		readline.PcItem("ORDER BY"),
		readline.PcItem("ORDER BY ASC"),
		readline.PcItem("ORDER BY DESC"),
		readline.PcItem("DISTINCT"),
		readline.PcItem("GROUP BY"),
		readline.PcItem("HAVING"),
		readline.PcItem("COUNT"),
		readline.PcItem("NOT EXISTS"),
	}

	// Add predicates as completions
	for _, pred := range c.predicates {
		c.commands = append(c.commands, readline.PcItem(pred))
	}
}

// buildModuleList extracts module paths from the graph
func (c *Completer) buildModuleList() {
	seen := make(map[string]bool)
	for _, mod := range c.graph.Modules {
		if mod.Name != "" && !seen[mod.Name] {
			c.modules = append(c.modules, "<#"+mod.Name+">")
			seen[mod.Name] = true
		}
		if mod.Path != "" && !seen[mod.Path] {
			c.modules = append(c.modules, "<#"+mod.Path+">")
			seen[mod.Path] = true
		}
	}
}

// buildPredicateList extracts common predicates
func (c *Completer) buildPredicateList() {
	// Common code ontology predicates
	predicates := []string{
		"<#imports>",
		"<#exports>",
		"<#name>",
		"<#description>",
		"<#language>",
		"<#layer>",
		"<#zone>",
		"<#tags>",
		"<#linksTo>",
		"<#dependsOn>",
		"<#type>",
		"<#version>",
		"code:Module",
		"code:name",
		"code:description",
		"code:language",
		"code:layer",
		"code:zone",
		"code:tags",
		"code:imports",
		"code:exports",
		"code:linksTo",
		"code:dependsOn",
		"a",
		"rdf:type",
	}

	c.predicates = predicates
}

// GetCompleter returns a readline completer
func (c *Completer) GetCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(c.commands...)
}

// GetAutoCompleteFunc returns a custom autocomplete function for context-aware completion
func (c *Completer) GetAutoCompleteFunc() readline.AutoCompleter {
	return &contextCompleter{c}
}

// contextCompleter implements readline.AutoCompleter for context-aware completion
type contextCompleter struct {
	completer *Completer
}

// Do implements the readline.AutoCompleter interface
func (cc *contextCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])

	// Get the word being typed
	words := strings.Fields(lineStr)
	if len(words) == 0 {
		return nil, 0
	}

	// Get the last word (what we're completing)
	lastWord := ""
	if pos > 0 && !unicode.IsSpace(rune(line[pos-1])) {
		lastWord = words[len(words)-1]
	}

	// Determine what to suggest based on context
	var suggestions []string

	// Check what context we're in
	if strings.HasPrefix(lastWord, ".") {
		// Command completion
		suggestions = []string{
			".help", ".format", ".load", ".save", ".history",
			".clear", ".schema", ".examples", ".stats",
			".modules", ".predicates", ".exit", ".quit",
		}
	} else if strings.HasPrefix(lastWord, "<#") || strings.HasPrefix(lastWord, "<") {
		// Module or predicate completion
		suggestions = append(suggestions, cc.completer.predicates...)
		suggestions = append(suggestions, cc.completer.modules...)
	} else if strings.HasPrefix(lastWord, "code:") {
		// Code ontology completion
		for _, pred := range cc.completer.predicates {
			if strings.HasPrefix(pred, "code:") {
				suggestions = append(suggestions, pred)
			}
		}
	} else {
		// Keyword completion
		suggestions = cc.completer.keywords
		// Also add commands and predicates
		suggestions = append(suggestions, ".help", ".format", ".modules", ".predicates")
	}

	// Filter suggestions by prefix
	var matches []string
	lowerLast := strings.ToLower(lastWord)
	for _, suggestion := range suggestions {
		if strings.HasPrefix(strings.ToLower(suggestion), lowerLast) {
			matches = append(matches, suggestion)
		}
	}

	// Convert matches to readline format
	if len(matches) == 0 {
		return nil, 0
	}

	// Calculate the length to replace
	length = len(lastWord)

	// Convert matches to [][]rune
	newLine = make([][]rune, len(matches))
	for i, match := range matches {
		// Only return the part after what's already typed
		completion := match[len(lastWord):]
		newLine[i] = []rune(completion)
	}

	return newLine, length
}

// getSPARQLKeywords returns common SPARQL keywords
func getSPARQLKeywords() []string {
	return []string{
		"SELECT", "WHERE", "CONSTRUCT", "ASK", "DESCRIBE",
		"PREFIX", "FILTER", "OPTIONAL", "UNION", "LIMIT",
		"OFFSET", "ORDER", "BY", "ASC", "DESC", "DISTINCT",
		"GROUP", "HAVING", "COUNT", "SUM", "AVG", "MIN", "MAX",
		"BIND", "VALUES", "NOT", "EXISTS", "REGEX", "LANG",
		"DATATYPE", "STR", "STRLEN", "SUBSTR", "UCASE", "LCASE",
		"STRSTARTS", "STRENDS", "CONTAINS", "CONCAT", "REPLACE",
	}
}

// GetModules returns the list of module paths
func (c *Completer) GetModules() []string {
	return c.modules
}

// GetPredicates returns the list of predicates
func (c *Completer) GetPredicates() []string {
	return c.predicates
}

// GetKeywords returns SPARQL keywords
func (c *Completer) GetKeywords() []string {
	return c.keywords
}

// FilterSuggestions filters suggestions based on prefix
func FilterSuggestions(suggestions []string, prefix string) []string {
	if prefix == "" {
		return suggestions
	}

	prefix = strings.ToLower(prefix)
	filtered := make([]string, 0)

	for _, suggestion := range suggestions {
		if strings.HasPrefix(strings.ToLower(suggestion), prefix) {
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}
