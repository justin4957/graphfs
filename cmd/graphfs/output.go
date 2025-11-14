/*
# Module: cmd/graphfs/output.go
Output formatting utilities for CLI.

Provides table formatting for query results using go-pretty.

## Linked Modules
- [cmd_query](./cmd_query.go) - Query command
- [../../pkg/query](../../pkg/query/engine.go) - Query engine

## Tags
cli, output, formatting, table

## Exports
formatTable

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#output.go> a code:Module ;

	code:name "cmd/graphfs/output.go" ;
	code:description "Output formatting utilities for CLI" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./cmd_query.go>, <../../pkg/query/engine.go> ;
	code:exports <#formatTable> ;
	code:tags "cli", "output", "formatting", "table" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/justin4957/graphfs/pkg/query"
)

// formatTable formats query results as a pretty table
func formatTable(result *query.QueryResult) (string, error) {
	if len(result.Bindings) == 0 {
		return "No results found.", nil
	}

	// Create table writer
	t := table.NewWriter()

	// Add header
	header := make(table.Row, len(result.Variables))
	for i, variable := range result.Variables {
		header[i] = variable
	}
	t.AppendHeader(header)

	// Add rows
	for _, binding := range result.Bindings {
		row := make(table.Row, len(result.Variables))
		for i, variable := range result.Variables {
			if value, ok := binding[variable]; ok {
				row[i] = value
			} else {
				row[i] = ""
			}
		}
		t.AppendRow(row)
	}

	// Configure style
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false

	return t.Render(), nil
}
