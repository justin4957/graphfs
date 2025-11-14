/*
# Module: pkg/parser/triple.go
RDF triple data structure.

Represents a Subject-Predicate-Object triple extracted from LinkedDoc headers.
Supports literal values, URIs, and blank nodes.

## Linked Modules
None (core data structure)

## Tags
parser, rdf, data-structure

## Exports
Triple, TripleObject, LiteralObject, URIObject, BlankNodeObject

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#pkg/parser/triple.go> a code:Module ;
    code:name "pkg/parser/triple.go" ;
    code:description "RDF triple data structure" ;
    code:language "go" ;
    code:layer "data" ;
    code:exports <#Triple>, <#TripleObject>, <#LiteralObject>, <#URIObject>, <#BlankNodeObject> ;
    code:tags "parser", "rdf", "data-structure" ;
    code:isLeaf true .

<#Triple> a code:Type ;
    code:name "Triple" ;
    code:kind "struct" ;
    code:description "RDF Subject-Predicate-Object triple" ;
    code:hasField <#Triple.Subject>, <#Triple.Predicate>, <#Triple.Object> .
<!-- End LinkedDoc RDF -->
*/

package parser

import "fmt"

// Triple represents an RDF Subject-Predicate-Object triple
type Triple struct {
	Subject   string
	Predicate string
	Object    TripleObject
}

// TripleObject represents the object part of a triple
// Can be a literal, URI, or blank node
type TripleObject interface {
	String() string
	Type() string
}

// LiteralObject represents a literal value (string, number, etc.)
type LiteralObject struct {
	Value string
}

func (l LiteralObject) String() string {
	return l.Value
}

func (l LiteralObject) Type() string {
	return "literal"
}

// URIObject represents a URI reference
type URIObject struct {
	URI string
}

func (u URIObject) String() string {
	return u.URI
}

func (u URIObject) Type() string {
	return "uri"
}

// BlankNodeObject represents a blank node with properties
type BlankNodeObject struct {
	Triples []Triple
}

func (b BlankNodeObject) String() string {
	return fmt.Sprintf("[%d triples]", len(b.Triples))
}

func (b BlankNodeObject) Type() string {
	return "bnode"
}

// NewLiteral creates a new literal object
func NewLiteral(value string) TripleObject {
	return LiteralObject{Value: value}
}

// NewURI creates a new URI object
func NewURI(uri string) TripleObject {
	return URIObject{URI: uri}
}

// NewBlankNode creates a new blank node object
func NewBlankNode(triples []Triple) TripleObject {
	return BlankNodeObject{Triples: triples}
}
