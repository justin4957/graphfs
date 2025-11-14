/*
# Module: pkg/query/parser.go
SPARQL query parser.

Parses SPARQL SELECT queries with WHERE, FILTER, ORDER BY, LIMIT, and OFFSET.

## Linked Modules
- [query](./query.go) - Query data structures

## Tags
query, sparql, parser

## Exports
ParseQuery

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#pkg/query/parser.go> a code:Module ;
    code:name "pkg/query/parser.go" ;
    code:description "SPARQL query parser" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./query.go> ;
    code:exports <#ParseQuery> ;
    code:tags "query", "sparql", "parser" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseQuery parses a SPARQL query string
func ParseQuery(queryStr string) (*Query, error) {
	queryStr = strings.TrimSpace(queryStr)

	// Detect query type (check for SELECT anywhere in the query, not just at start)
	if strings.Contains(strings.ToUpper(queryStr), "SELECT") {
		selectQuery, err := parseSelectQuery(queryStr)
		if err != nil {
			return nil, err
		}
		return &Query{
			Type:   SelectQueryType,
			Select: selectQuery,
		}, nil
	}

	return nil, fmt.Errorf("unsupported query type")
}

// parseSelectQuery parses a SELECT query
func parseSelectQuery(queryStr string) (*SelectQuery, error) {
	query := &SelectQuery{
		Prefixes: make(map[string]string),
		Limit:    0,
		Offset:   0,
	}

	// Extract PREFIX declarations
	prefixRegex := regexp.MustCompile(`(?i)PREFIX\s+(\w+):\s+<([^>]+)>`)
	prefixMatches := prefixRegex.FindAllStringSubmatch(queryStr, -1)
	for _, match := range prefixMatches {
		if len(match) == 3 {
			query.Prefixes[match[1]] = match[2]
		}
	}

	// Remove PREFIX declarations for easier parsing
	queryStr = prefixRegex.ReplaceAllString(queryStr, "")

	// Check for DISTINCT
	if regexp.MustCompile(`(?i)\bSELECT\s+DISTINCT\b`).MatchString(queryStr) {
		query.Distinct = true
		queryStr = regexp.MustCompile(`(?i)\bDISTINCT\b`).ReplaceAllString(queryStr, "")
	}

	// Extract variables from SELECT clause
	selectRegex := regexp.MustCompile(`(?i)SELECT\s+([\s\S]*?)\s+WHERE`)
	selectMatch := selectRegex.FindStringSubmatch(queryStr)
	if selectMatch == nil {
		return nil, fmt.Errorf("invalid SELECT query: missing WHERE clause")
	}

	varsStr := strings.TrimSpace(selectMatch[1])
	if varsStr == "*" {
		query.Variables = []string{"*"}
	} else {
		// Extract variables
		varRegex := regexp.MustCompile(`\?(\w+)`)
		varMatches := varRegex.FindAllString(varsStr, -1)
		query.Variables = varMatches
	}

	// Extract WHERE clause
	whereRegex := regexp.MustCompile(`(?i)WHERE\s*\{([\s\S]*?)\}`)
	whereMatch := whereRegex.FindStringSubmatch(queryStr)
	if whereMatch == nil {
		return nil, fmt.Errorf("invalid WHERE clause")
	}

	whereClause := whereMatch[1]

	// Parse triple patterns
	patterns, err := parseTriplePatterns(whereClause, query.Prefixes)
	if err != nil {
		return nil, err
	}
	query.Where = patterns

	// Extract FILTER clauses - must handle nested parentheses
	query.Filters = extractFilters(whereClause)

	// Extract ORDER BY
	orderByRegex := regexp.MustCompile(`(?i)ORDER\s+BY\s+(ASC|DESC)?\s*\(\s*\?(\w+)\s*\)`)
	orderByMatch := orderByRegex.FindStringSubmatch(queryStr)
	if orderByMatch != nil {
		orderBy := OrderBy{
			Variable:   "?" + orderByMatch[2],
			Descending: strings.ToUpper(orderByMatch[1]) == "DESC",
		}
		query.OrderBy = append(query.OrderBy, orderBy)
	}

	// Extract LIMIT
	limitRegex := regexp.MustCompile(`(?i)LIMIT\s+(\d+)`)
	limitMatch := limitRegex.FindStringSubmatch(queryStr)
	if limitMatch != nil {
		limit, _ := strconv.Atoi(limitMatch[1])
		query.Limit = limit
	}

	// Extract OFFSET
	offsetRegex := regexp.MustCompile(`(?i)OFFSET\s+(\d+)`)
	offsetMatch := offsetRegex.FindStringSubmatch(queryStr)
	if offsetMatch != nil {
		offset, _ := strconv.Atoi(offsetMatch[1])
		query.Offset = offset
	}

	return query, nil
}

// extractFilters extracts FILTER clauses with balanced parentheses
func extractFilters(whereClause string) []Filter {
	var filters []Filter

	// Find all FILTER keywords
	filterKeyword := regexp.MustCompile(`(?i)\bFILTER\s*\(`)
	matches := filterKeyword.FindAllStringIndex(whereClause, -1)

	for _, match := range matches {
		startIdx := match[1] // Position after "FILTER("

		// Find matching closing parenthesis
		depth := 1
		endIdx := startIdx
		for endIdx < len(whereClause) && depth > 0 {
			if whereClause[endIdx] == '(' {
				depth++
			} else if whereClause[endIdx] == ')' {
				depth--
			}
			endIdx++
		}

		if depth == 0 {
			// Extract expression between balanced parentheses
			expression := strings.TrimSpace(whereClause[startIdx : endIdx-1])
			filters = append(filters, Filter{Expression: expression})
		}
	}

	return filters
}

// splitTriples splits a WHERE clause by periods, but not periods inside URIs
func splitTriples(whereClause string) []string {
	var triples []string
	var current strings.Builder
	inURI := false

	for _, ch := range whereClause {
		if ch == '<' {
			inURI = true
			current.WriteRune(ch)
		} else if ch == '>' {
			inURI = false
			current.WriteRune(ch)
		} else if ch == '.' && !inURI {
			// Triple terminator
			if current.Len() > 0 {
				triples = append(triples, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}

	// Add final triple if any
	if current.Len() > 0 {
		triples = append(triples, current.String())
	}

	return triples
}

// parseTriplePatterns parses triple patterns from WHERE clause
func parseTriplePatterns(whereClause string, prefixes map[string]string) ([]TriplePattern, error) {
	var patterns []TriplePattern

	// Split by period (end of triple) but not periods inside URIs
	lines := splitTriples(whereClause)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle semicolon (same subject continuation)
		parts := strings.Split(line, ";")

		var currentSubject string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Split by whitespace
			tokens := strings.Fields(part)
			if len(tokens) < 3 {
				// Check if this is a continuation (has only 2 tokens)
				if len(tokens) == 2 && currentSubject != "" {
					// Subject is from previous pattern
					tokens = append([]string{currentSubject}, tokens...)
				} else {
					continue
				}
			}

			subject := expandPrefix(tokens[0], prefixes)
			predicate := expandPrefix(tokens[1], prefixes)

			// Handle "a" as rdf:type
			if predicate == "a" {
				predicate = "rdf:type"
				predicate = expandPrefix(predicate, prefixes)
			}

			// Object is everything after predicate
			object := strings.Join(tokens[2:], " ")
			object = expandPrefix(object, prefixes)

			patterns = append(patterns, TriplePattern{
				Subject:   subject,
				Predicate: predicate,
				Object:    object,
			})

			currentSubject = subject
		}
	}

	return patterns, nil
}

// expandPrefix expands a prefixed URI
func expandPrefix(term string, prefixes map[string]string) string {
	term = strings.TrimSpace(term)

	// Check if it's a prefixed URI
	if strings.Contains(term, ":") && !strings.HasPrefix(term, "<") && !strings.HasPrefix(term, "?") && !strings.HasPrefix(term, "\"") {
		parts := strings.SplitN(term, ":", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			localName := parts[1]

			if baseURI, ok := prefixes[prefix]; ok {
				// Return expanded URI with angle brackets for SPARQL syntax
				return "<" + baseURI + localName + ">"
			}
		}
	}

	// Return as-is if it's already a variable, URI, or literal
	return term
}
