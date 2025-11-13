/*
# Module: internal/store/triple.go
Triple data structure for RDF storage.

Represents a Subject-Predicate-Object triple for the triple store.
Compatible with parser.Triple but optimized for storage.

## Linked Modules
None (core data structure)

## Tags
store, rdf, data-structure

## Exports
Triple

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#triple.go> a code:Module ;
    code:name "internal/store/triple.go" ;
    code:description "Triple data structure for RDF storage" ;
    code:language "go" ;
    code:layer "storage" ;
    code:exports <#Triple> ;
    code:tags "store", "rdf", "data-structure" ;
    code:isLeaf true .

<#Triple> a code:Type ;
    code:name "Triple" ;
    code:kind "struct" ;
    code:description "Subject-Predicate-Object triple for storage" ;
    code:hasField <#Triple.Subject>, <#Triple.Predicate>, <#Triple.Object> .
<!-- End LinkedDoc RDF -->
*/

package store

// Triple represents an RDF Subject-Predicate-Object triple
type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

// NewTriple creates a new triple
func NewTriple(subject, predicate, object string) Triple {
	return Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}
}

// Equals checks if two triples are equal
func (t Triple) Equals(other Triple) bool {
	return t.Subject == other.Subject &&
		t.Predicate == other.Predicate &&
		t.Object == other.Object
}

// String returns a string representation of the triple
func (t Triple) String() string {
	return t.Subject + " -> " + t.Predicate + " -> " + t.Object
}
