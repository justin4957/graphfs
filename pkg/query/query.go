/*
# Module: pkg/query/query.go
SPARQL query data structures.

Defines the data structures for representing parsed SPARQL queries.

## Linked Modules
None (core data structure)

## Tags
query, sparql, data-structure

## Exports
Query, SelectQuery, TriplePattern, Filter, OrderBy

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#query.go> a code:Module ;
    code:name "pkg/query/query.go" ;
    code:description "SPARQL query data structures" ;
    code:language "go" ;
    code:layer "query" ;
    code:exports <#Query>, <#SelectQuery>, <#TriplePattern>, <#Filter> ;
    code:tags "query", "sparql", "data-structure" .
<!-- End LinkedDoc RDF -->
*/

package query

// Query represents a SPARQL query
type Query struct {
	Type   QueryType
	Select *SelectQuery
}

// QueryType represents the type of SPARQL query
type QueryType string

const (
	SelectQueryType QueryType = "SELECT"
)

// SelectQuery represents a SELECT query
type SelectQuery struct {
	Variables []string          // Variables to select (e.g., ["?subject", "?predicate"])
	Distinct  bool              // DISTINCT modifier
	Where     []TriplePattern   // WHERE clause triple patterns
	Filters   []Filter          // FILTER clauses
	OrderBy   []OrderBy         // ORDER BY clauses
	Limit     int               // LIMIT (0 = no limit)
	Offset    int               // OFFSET (0 = no offset)
	Prefixes  map[string]string // Prefix declarations
}

// TriplePattern represents a triple pattern in WHERE clause
type TriplePattern struct {
	Subject   string // Can be variable (?var), URI (<uri>), or literal
	Predicate string
	Object    string
}

// Filter represents a FILTER clause
type Filter struct {
	Expression string // Filter expression (simplified for now)
}

// OrderBy represents an ORDER BY clause
type OrderBy struct {
	Variable   string
	Descending bool
}

// IsVariable checks if a string is a SPARQL variable
func IsVariable(s string) bool {
	return len(s) > 0 && s[0] == '?'
}

// IsURI checks if a string is a URI reference
func IsURI(s string) bool {
	return len(s) > 1 && s[0] == '<' && s[len(s)-1] == '>'
}

// IsLiteral checks if a string is a quoted literal
func IsLiteral(s string) bool {
	return len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"'
}

// StripVariable removes the ? prefix from a variable
func StripVariable(s string) string {
	if IsVariable(s) {
		return s[1:]
	}
	return s
}

// StripURI removes the < > brackets from a URI
func StripURI(s string) string {
	if IsURI(s) {
		return s[1 : len(s)-1]
	}
	return s
}

// StripLiteral removes the quotes from a literal string
func StripLiteral(s string) string {
	if IsLiteral(s) {
		return s[1 : len(s)-1]
	}
	return s
}
