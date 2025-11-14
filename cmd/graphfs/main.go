/*
# Module: cmd/graphfs/main.go
Main CLI entry point for GraphFS.

Implements the root command and CLI framework using Cobra.

## Linked Modules
- [root](./root.go) - Root command
- [../../pkg/parser](../../pkg/parser/parser.go) - RDF/Turtle parsing
- [../../pkg/scanner](../../pkg/scanner/scanner.go) - Filesystem scanning
- [../../pkg/graph](../../pkg/graph/graph.go) - Knowledge graph construction

## Tags
cli, main, entrypoint

## Exports
main

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd/graphfs/main.go> a code:Module ;

	code:name "cmd/graphfs/main.go" ;
	code:description "Main CLI entry point for GraphFS" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <../../pkg/parser/parser.go>,
	             <../../pkg/scanner/scanner.go>, <../../pkg/graph/graph.go> ;
	code:exports <#main> ;
	code:tags "cli", "main", "entrypoint" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
