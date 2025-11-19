/*
# Module: pkg/repl/highlighter.go
Syntax highlighting for SPARQL queries.

Provides color highlighting for SPARQL keywords, URIs, strings, and comments.

## Linked Modules
- [repl](./repl.go) - REPL core

## Tags
repl, syntax, highlighting, color

## Exports
Highlighter, HighlightQuery

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#highlighter.go> a code:Module ;
    code:name "pkg/repl/highlighter.go" ;
    code:description "Syntax highlighting for SPARQL queries" ;
    code:language "go" ;
    code:layer "repl" ;
    code:linksTo <./repl.go> ;
    code:exports <#Highlighter>, <#HighlightQuery> ;
    code:tags "repl", "syntax", "highlighting", "color" .
<!-- End LinkedDoc RDF -->
*/

package repl

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Highlighter provides syntax highlighting for SPARQL
type Highlighter struct {
	noColor       bool
	keywordColor  *color.Color
	uriColor      *color.Color
	stringColor   *color.Color
	commentColor  *color.Color
	variableColor *color.Color
}

// NewHighlighter creates a new syntax highlighter
func NewHighlighter(noColor bool) *Highlighter {
	return &Highlighter{
		noColor:       noColor,
		keywordColor:  color.New(color.FgCyan, color.Bold),
		uriColor:      color.New(color.FgGreen),
		stringColor:   color.New(color.FgYellow),
		commentColor:  color.New(color.FgHiBlack),
		variableColor: color.New(color.FgMagenta),
	}
}

// HighlightQuery applies syntax highlighting to a SPARQL query
func (h *Highlighter) HighlightQuery(query string) string {
	if h.noColor {
		return query
	}

	// Patterns to match
	keywordPattern := regexp.MustCompile(`\b(SELECT|WHERE|CONSTRUCT|ASK|DESCRIBE|PREFIX|FILTER|OPTIONAL|UNION|LIMIT|OFFSET|ORDER BY|DISTINCT|GROUP BY|HAVING|COUNT|SUM|AVG|MIN|MAX|BIND|VALUES|NOT EXISTS|ASC|DESC)\b`)
	uriPattern := regexp.MustCompile(`<[^>]+>`)
	stringPattern := regexp.MustCompile(`"[^"]*"`)
	variablePattern := regexp.MustCompile(`\?[a-zA-Z_][a-zA-Z0-9_]*`)
	commentPattern := regexp.MustCompile(`#.*$`)

	// Apply highlighting in order (comments last to avoid conflicts)
	result := query

	// Highlight URIs
	result = uriPattern.ReplaceAllStringFunc(result, func(match string) string {
		return h.uriColor.Sprint(match)
	})

	// Highlight strings
	result = stringPattern.ReplaceAllStringFunc(result, func(match string) string {
		return h.stringColor.Sprint(match)
	})

	// Highlight variables
	result = variablePattern.ReplaceAllStringFunc(result, func(match string) string {
		return h.variableColor.Sprint(match)
	})

	// Highlight keywords (case-insensitive)
	result = keywordPattern.ReplaceAllStringFunc(result, func(match string) string {
		return h.keywordColor.Sprint(strings.ToUpper(match))
	})

	// Highlight comments
	result = commentPattern.ReplaceAllStringFunc(result, func(match string) string {
		return h.commentColor.Sprint(match)
	})

	return result
}

// HighlightQuery is a convenience function for highlighting
func HighlightQuery(query string, noColor bool) string {
	h := NewHighlighter(noColor)
	return h.HighlightQuery(query)
}
