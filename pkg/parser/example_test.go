package parser_test

import (
	"fmt"
	"log"

	"github.com/justin4957/graphfs/pkg/parser"
)

// Example demonstrates basic parser usage
func Example() {
	content := `/*
# Module: example.go
Example module demonstrating LinkedDoc format.

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .

<#example.go> a code:Module ;
    code:name "example.go" ;
    code:description "Example module" ;
    code:exports <#Foo>, <#Bar> ;
    code:tags "example", "demo" .
<!-- End LinkedDoc RDF -->
*/`

	p := parser.NewParser()
	triples, err := p.ParseString(content)
	if err != nil {
		log.Fatal(err)
	}

	for _, triple := range triples {
		fmt.Printf("%s\n", triple.Predicate)
	}

	// Output:
	// http://www.w3.org/1999/02/22-rdf-syntax-ns#type
	// https://schema.codedoc.org/name
	// https://schema.codedoc.org/description
	// https://schema.codedoc.org/exports
	// https://schema.codedoc.org/exports
	// https://schema.codedoc.org/tags
	// https://schema.codedoc.org/tags
}

// Example_parseFile demonstrates parsing a file
func Example_parseFile() {
	p := parser.NewParser()
	triples, err := p.Parse("../../examples/minimal-app/main.go")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed %d triples from main.go\n", len(triples))

	// Find module name
	for _, triple := range triples {
		if triple.Predicate == "https://schema.codedoc.org/name" &&
			triple.Subject == "<#main.go>" {
			fmt.Printf("Module name: %s\n", triple.Object.String())
		}
	}

	// Output:
	// Parsed 19 triples from main.go
	// Module name: main.go
}

// Example_extractLinkedDoc demonstrates extracting the LinkedDoc block
func Example_extractLinkedDoc() {
	content := `/*
Some documentation

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test> a code:Module .
<!-- End LinkedDoc RDF -->
*/`

	p := parser.NewParser()
	rdfContent, err := p.ExtractLinkedDoc(content)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d characters of RDF\n", len(rdfContent))

	// Output:
	// Found 69 characters of RDF
}

// Example_tripleTypes demonstrates different triple object types
func Example_tripleTypes() {
	content := `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#module> a code:Module ;
    code:name "test" ;
    code:linksTo <./other.go> .
<!-- End LinkedDoc RDF -->
*/`

	p := parser.NewParser()
	triples, _ := p.ParseString(content)

	for _, triple := range triples {
		fmt.Printf("%s: %s (%s)\n",
			triple.Predicate[len(triple.Predicate)-7:], // Last 7 chars
			triple.Object.String(),
			triple.Object.Type())
	}

	// Output:
	// ns#type: https://schema.codedoc.org/Module (uri)
	// rg/name: test (literal)
	// linksTo: ./other.go (uri)
}
