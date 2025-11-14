/*
# Module: pkg/parser/parser.go
LinkedDoc+RDF parser implementation.

Extracts RDF/Turtle triples from LinkedDoc comment blocks in source code.
Supports @prefix declarations, URIs, literals, and blank nodes.

## Linked Modules
- [triple](./triple.go) - Triple data structure

## Tags
parser, rdf, turtle, linkeddoc

## Exports
Parser, NewParser, ParseError

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#pkg/parser/parser.go> a code:Module ;
    code:name "pkg/parser/parser.go" ;
    code:description "LinkedDoc+RDF parser implementation" ;
    code:language "go" ;
    code:layer "parser" ;
    code:linksTo <./triple.go> ;
    code:exports <#Parser>, <#NewParser>, <#ParseError> ;
    code:tags "parser", "rdf", "turtle", "linkeddoc" .

<#Parser> a code:Type ;
    code:name "Parser" ;
    code:kind "struct" ;
    code:description "LinkedDoc parser" ;
    code:hasMethod <#Parser.Parse>, <#Parser.ParseString>, <#Parser.ExtractLinkedDoc> .

<#Parser.Parse> a code:Method ;
    code:name "Parse" ;
    code:description "Parses LinkedDoc from a file" .

<#Parser.ParseString> a code:Method ;
    code:name "ParseString" ;
    code:description "Parses LinkedDoc from a string" .

<#Parser.ExtractLinkedDoc> a code:Method ;
    code:name "ExtractLinkedDoc" ;
    code:description "Extracts LinkedDoc RDF block from content" .
<!-- End LinkedDoc RDF -->
*/

package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Parser extracts RDF triples from LinkedDoc headers
type Parser struct {
	prefixes map[string]string
}

// ParseError represents a parsing error
type ParseError struct {
	Message string
	Line    int
}

func (e ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error at line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// NewParser creates a new LinkedDoc parser
func NewParser() *Parser {
	return &Parser{
		prefixes: make(map[string]string),
	}
}

// Parse extracts RDF triples from a source file
func (p *Parser) Parse(filePath string) ([]Triple, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseString(string(content))
}

// ParseString extracts RDF triples from a string
func (p *Parser) ParseString(content string) ([]Triple, error) {
	// Reset prefixes for each parse with standard RDF prefix
	p.prefixes = map[string]string{
		"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	}

	linkedDoc, err := p.ExtractLinkedDoc(content)
	if err != nil {
		return nil, err
	}

	if linkedDoc == "" {
		// No LinkedDoc block found - not an error, just return empty
		return []Triple{}, nil
	}

	return p.parseRDF(linkedDoc)
}

// ExtractLinkedDoc extracts the LinkedDoc RDF block from content
func (p *Parser) ExtractLinkedDoc(content string) (string, error) {
	startMarker := "<!-- LinkedDoc RDF -->"
	endMarker := "<!-- End LinkedDoc RDF -->"

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return "", nil // No LinkedDoc block
	}

	endIdx := strings.Index(content, endMarker)
	if endIdx == -1 {
		return "", ParseError{Message: "LinkedDoc block not closed (missing <!-- End LinkedDoc RDF -->)"}
	}

	if endIdx <= startIdx {
		return "", ParseError{Message: "Invalid LinkedDoc block (end marker before start)"}
	}

	// Extract content between markers
	rdfContent := content[startIdx+len(startMarker) : endIdx]
	return strings.TrimSpace(rdfContent), nil
}

// parseRDF parses RDF/Turtle triples from LinkedDoc content
func (p *Parser) parseRDF(content string) ([]Triple, error) {
	var triples []Triple
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	var currentSubject string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse @prefix declarations
		if strings.HasPrefix(line, "@prefix") {
			if err := p.parsePrefix(line); err != nil {
				return nil, ParseError{Message: err.Error(), Line: lineNum}
			}
			continue
		}

		// Parse triples
		parsedTriples, newSubject, err := p.parseTripleLine(line, currentSubject, lineNum)
		if err != nil {
			return nil, err
		}

		if newSubject != "" {
			currentSubject = newSubject
		}

		triples = append(triples, parsedTriples...)

		// Check if line ends with continuation
		if !strings.HasSuffix(line, ";") && !strings.HasSuffix(line, ",") && strings.HasSuffix(line, ".") {
			currentSubject = ""
		}
	}

	return triples, nil
}

// parsePrefix parses an @prefix declaration
func (p *Parser) parsePrefix(line string) error {
	// Format: @prefix name: <uri> .
	re := regexp.MustCompile(`@prefix\s+(\w+):\s+<([^>]+)>\s*\.`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 3 {
		return fmt.Errorf("invalid @prefix syntax: %s", line)
	}

	prefix := matches[1]
	uri := matches[2]
	p.prefixes[prefix] = uri

	return nil
}

// parseTripleLine parses a line containing triples
func (p *Parser) parseTripleLine(line string, currentSubject string, lineNum int) ([]Triple, string, error) {
	var triples []Triple

	// Remove trailing punctuation for parsing
	line = strings.TrimRight(line, ";.,")

	// Check if this is a new subject declaration
	if strings.Contains(line, " a ") || strings.Contains(line, " code:") || strings.Contains(line, " rdf:") || strings.Contains(line, " rdfs:") || strings.Contains(line, " sec:") || strings.Contains(line, " arch:") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			subject := p.expandPrefix(strings.TrimSpace(parts[0]))
			rest := strings.TrimSpace(parts[1])

			// Parse predicate-object pairs
			poTriples, err := p.parsePredicateObjects(subject, rest, lineNum)
			if err != nil {
				return nil, "", err
			}

			triples = append(triples, poTriples...)
			return triples, subject, nil
		}
	}

	// Continuation of previous subject
	if currentSubject != "" {
		poTriples, err := p.parsePredicateObjects(currentSubject, line, lineNum)
		if err != nil {
			return nil, "", err
		}
		triples = append(triples, poTriples...)
		return triples, currentSubject, nil
	}

	return triples, "", nil
}

// parsePredicateObjects parses predicate-object pairs
func (p *Parser) parsePredicateObjects(subject, line string, lineNum int) ([]Triple, error) {
	var triples []Triple

	// Split by semicolon for multiple predicate-object pairs
	pairs := strings.Split(line, ";")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split predicate and objects
		parts := strings.SplitN(pair, " ", 2)
		if len(parts) < 2 {
			continue
		}

		predicateStr := strings.TrimSpace(parts[0])

		// Handle "a" as shorthand for rdf:type
		if predicateStr == "a" {
			predicateStr = "rdf:type"
		}

		predicate := p.expandPrefix(predicateStr)
		objectsStr := strings.TrimSpace(parts[1])

		// Split objects by comma
		objects := strings.Split(objectsStr, ",")

		for _, objStr := range objects {
			objStr = strings.TrimSpace(objStr)
			if objStr == "" {
				continue
			}

			obj := p.parseObject(objStr)
			triples = append(triples, Triple{
				Subject:   subject,
				Predicate: predicate,
				Object:    obj,
			})
		}
	}

	return triples, nil
}

// parseObject parses an object value
func (p *Parser) parseObject(objStr string) TripleObject {
	objStr = strings.TrimSpace(objStr)

	// URI reference: <uri> or prefix:name
	if strings.HasPrefix(objStr, "<") && strings.HasSuffix(objStr, ">") {
		uri := objStr[1 : len(objStr)-1]
		return NewURI(uri)
	}

	// Prefixed URI
	if strings.Contains(objStr, ":") && !strings.HasPrefix(objStr, "\"") {
		return NewURI(p.expandPrefix(objStr))
	}

	// Literal string: "value"
	if strings.HasPrefix(objStr, "\"") && strings.HasSuffix(objStr, "\"") {
		literal := objStr[1 : len(objStr)-1]
		return NewLiteral(literal)
	}

	// Blank node: [...]
	if strings.HasPrefix(objStr, "[") && strings.HasSuffix(objStr, "]") {
		// Simplified blank node handling
		return NewBlankNode([]Triple{})
	}

	// Default: treat as literal
	return NewLiteral(objStr)
}

// expandPrefix expands a prefixed URI to full URI
func (p *Parser) expandPrefix(prefixedURI string) string {
	parts := strings.SplitN(prefixedURI, ":", 2)
	if len(parts) == 2 {
		prefix := parts[0]
		localName := parts[1]

		if baseURI, ok := p.prefixes[prefix]; ok {
			return baseURI + localName
		}
	}

	return prefixedURI
}
