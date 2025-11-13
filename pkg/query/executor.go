/*
# Module: pkg/query/executor.go
SPARQL query executor.

Executes parsed SPARQL queries against the triple store.

## Linked Modules
- [query](./query.go) - Query data structures
- [../../internal/store](../../internal/store/store.go) - Triple store

## Tags
query, sparql, executor

## Exports
Executor, NewExecutor, QueryResult, Binding

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#executor.go> a code:Module ;
    code:name "pkg/query/executor.go" ;
    code:description "SPARQL query executor" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./query.go>, <../../internal/store/store.go> ;
    code:exports <#Executor>, <#NewExecutor>, <#QueryResult> ;
    code:tags "query", "sparql", "executor" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/justin4957/graphfs/internal/store"
)

// Executor executes SPARQL queries against a triple store
type Executor struct {
	store *store.TripleStore
}

// NewExecutor creates a new query executor
func NewExecutor(tripleStore *store.TripleStore) *Executor {
	return &Executor{
		store: tripleStore,
	}
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Variables []string             // Variable names (without ?)
	Bindings  []map[string]string  // Variable bindings for each result
	Count     int                  // Number of results
}

// Execute executes a parsed query
func (e *Executor) Execute(query *Query) (*QueryResult, error) {
	if query.Type == SelectQueryType {
		return e.executeSelect(query.Select)
	}

	return nil, fmt.Errorf("unsupported query type: %s", query.Type)
}

// ExecuteString parses and executes a SPARQL query string
func (e *Executor) ExecuteString(queryStr string) (*QueryResult, error) {
	query, err := ParseQuery(queryStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return e.Execute(query)
}

// executeSelect executes a SELECT query
func (e *Executor) executeSelect(query *SelectQuery) (*QueryResult, error) {
	// Start with all possible bindings
	bindings := []map[string]string{{}}

	// Process each triple pattern
	for _, pattern := range query.Where {
		bindings = e.matchPattern(pattern, bindings, query.Prefixes)
	}

	// Apply filters
	for _, filter := range query.Filters {
		bindings = e.applyFilter(filter, bindings)
	}

	// Apply ORDER BY
	if len(query.OrderBy) > 0 {
		bindings = e.applyOrderBy(query.OrderBy[0], bindings)
	}

	// Apply OFFSET
	if query.Offset > 0 && query.Offset < len(bindings) {
		bindings = bindings[query.Offset:]
	} else if query.Offset >= len(bindings) {
		bindings = []map[string]string{}
	}

	// Apply LIMIT
	if query.Limit > 0 && query.Limit < len(bindings) {
		bindings = bindings[:query.Limit]
	}

	// Apply DISTINCT
	if query.Distinct {
		bindings = e.applyDistinct(bindings, query.Variables)
	}

	// Project variables
	result := &QueryResult{
		Bindings: bindings,
		Count:    len(bindings),
	}

	// Determine variables to return
	if len(query.Variables) == 1 && query.Variables[0] == "*" {
		// Return all variables
		varSet := make(map[string]bool)
		for _, binding := range bindings {
			for v := range binding {
				varSet[v] = true
			}
		}
		for v := range varSet {
			result.Variables = append(result.Variables, v)
		}
		sort.Strings(result.Variables)
	} else {
		// Return specified variables
		for _, v := range query.Variables {
			result.Variables = append(result.Variables, StripVariable(v))
		}
	}

	return result, nil
}

// matchPattern matches a triple pattern against the store
func (e *Executor) matchPattern(pattern TriplePattern, currentBindings []map[string]string, prefixes map[string]string) []map[string]string {
	var newBindings []map[string]string

	for _, binding := range currentBindings {
		// Resolve pattern with current bindings
		subject := e.resolveValue(pattern.Subject, binding, prefixes)
		predicate := e.resolveValue(pattern.Predicate, binding, prefixes)
		object := e.resolveValue(pattern.Object, binding, prefixes)

		// Query triple store
		triples := e.store.Find(subject, predicate, object)

		// Create new bindings for each matching triple
		for _, triple := range triples {
			newBinding := make(map[string]string)
			// Copy existing bindings
			for k, v := range binding {
				newBinding[k] = v
			}

			// Add new variable bindings
			if IsVariable(pattern.Subject) {
				newBinding[StripVariable(pattern.Subject)] = triple.Subject
			}
			if IsVariable(pattern.Predicate) {
				newBinding[StripVariable(pattern.Predicate)] = triple.Predicate
			}
			if IsVariable(pattern.Object) {
				newBinding[StripVariable(pattern.Object)] = triple.Object
			}

			newBindings = append(newBindings, newBinding)
		}
	}

	return newBindings
}

// resolveValue resolves a pattern value with variable bindings
func (e *Executor) resolveValue(value string, binding map[string]string, prefixes map[string]string) string {
	if IsVariable(value) {
		// Look up variable in bindings
		if boundValue, ok := binding[StripVariable(value)]; ok {
			return boundValue
		}
		// Unbound variable - use empty string as wildcard
		return ""
	}

	// Strip URI brackets if present - the store uses canonical form without brackets
	// Exception: subjects that start with < keep their brackets as they're part of the identifier
	if IsURI(value) {
		stripped := StripURI(value)
		// If it starts with # or ./ it's a local reference that should keep brackets in store
		if strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, "./") || strings.HasPrefix(stripped, "../") {
			return value // Keep brackets for local references
		}
		// For http/https URIs, use canonical form without brackets
		return stripped
	}

	return value
}

// applyFilter applies a FILTER clause to bindings
func (e *Executor) applyFilter(filter Filter, bindings []map[string]string) []map[string]string {
	var filtered []map[string]string

	for _, binding := range bindings {
		if e.evaluateFilter(filter.Expression, binding) {
			filtered = append(filtered, binding)
		}
	}

	return filtered
}

// evaluateFilter evaluates a filter expression (simplified)
func (e *Executor) evaluateFilter(expression string, binding map[string]string) bool {
	// Replace variables with their values
	expr := expression
	for varName, value := range binding {
		expr = strings.ReplaceAll(expr, "?"+varName, `"`+value+`"`)
	}

	// REGEX filter: REGEX(?var, "pattern")
	regexPattern := regexp.MustCompile(`REGEX\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)
	if match := regexPattern.FindStringSubmatch(expr); match != nil {
		value := match[1]
		pattern := match[2]
		matched, _ := regexp.MatchString(pattern, value)
		return matched
	}

	// CONTAINS: CONTAINS(?var, "substring")
	containsPattern := regexp.MustCompile(`CONTAINS\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)
	if match := containsPattern.FindStringSubmatch(expr); match != nil {
		value := match[1]
		substring := match[2]
		return strings.Contains(value, substring)
	}

	// Equality: ?var = "value"
	eqPattern := regexp.MustCompile(`"([^"]+)"\s*=\s*"([^"]+)"`)
	if match := eqPattern.FindStringSubmatch(expr); match != nil {
		return match[1] == match[2]
	}

	// Inequality: ?var != "value"
	neqPattern := regexp.MustCompile(`"([^"]+)"\s*!=\s*"([^"]+)"`)
	if match := neqPattern.FindStringSubmatch(expr); match != nil {
		return match[1] != match[2]
	}

	// Default: assume true if we can't parse
	return true
}

// applyOrderBy sorts bindings by a variable
func (e *Executor) applyOrderBy(orderBy OrderBy, bindings []map[string]string) []map[string]string {
	varName := StripVariable(orderBy.Variable)

	sort.Slice(bindings, func(i, j int) bool {
		valI := bindings[i][varName]
		valJ := bindings[j][varName]

		if orderBy.Descending {
			return valI > valJ
		}
		return valI < valJ
	})

	return bindings
}

// applyDistinct removes duplicate bindings
func (e *Executor) applyDistinct(bindings []map[string]string, variables []string) []map[string]string {
	seen := make(map[string]bool)
	var unique []map[string]string

	for _, binding := range bindings {
		// Create key from relevant variable values
		var key string
		if len(variables) == 1 && variables[0] == "*" {
			// All variables
			var keys []string
			for k, v := range binding {
				keys = append(keys, k+"="+v)
			}
			sort.Strings(keys)
			key = strings.Join(keys, "|")
		} else {
			// Specified variables
			var values []string
			for _, v := range variables {
				varName := StripVariable(v)
				values = append(values, binding[varName])
			}
			key = strings.Join(values, "|")
		}

		if !seen[key] {
			seen[key] = true
			unique = append(unique, binding)
		}
	}

	return unique
}
