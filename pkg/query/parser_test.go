package query

import (
	"testing"
)

func TestParseQuery_SimpleSelect(t *testing.T) {
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		SELECT ?module WHERE {
			?module a code:Module .
		}
	`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if query.Type != SelectQueryType {
		t.Errorf("Type = %v, want %v", query.Type, SelectQueryType)
	}

	if len(query.Select.Variables) != 1 || query.Select.Variables[0] != "?module" {
		t.Errorf("Variables = %v, want [?module]", query.Select.Variables)
	}

	if len(query.Select.Where) != 1 {
		t.Fatalf("Where patterns = %d, want 1", len(query.Select.Where))
	}

	pattern := query.Select.Where[0]
	if pattern.Subject != "?module" {
		t.Errorf("Subject = %v, want ?module", pattern.Subject)
	}
}

func TestParseQuery_MultipleVariables(t *testing.T) {
	queryStr := `SELECT ?s ?p ?o WHERE { ?s ?p ?o . }`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if len(query.Select.Variables) != 3 {
		t.Errorf("Variables count = %d, want 3", len(query.Select.Variables))
	}
}

func TestParseQuery_SelectAll(t *testing.T) {
	queryStr := `SELECT * WHERE { ?s ?p ?o . }`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if len(query.Select.Variables) != 1 || query.Select.Variables[0] != "*" {
		t.Errorf("Variables = %v, want [*]", query.Select.Variables)
	}
}

func TestParseQuery_WithLimit(t *testing.T) {
	queryStr := `SELECT ?s WHERE { ?s ?p ?o . } LIMIT 10`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if query.Select.Limit != 10 {
		t.Errorf("Limit = %d, want 10", query.Select.Limit)
	}
}

func TestParseQuery_WithOffset(t *testing.T) {
	queryStr := `SELECT ?s WHERE { ?s ?p ?o . } OFFSET 5`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if query.Select.Offset != 5 {
		t.Errorf("Offset = %d, want 5", query.Select.Offset)
	}
}

func TestParseQuery_WithDistinct(t *testing.T) {
	queryStr := `SELECT DISTINCT ?s WHERE { ?s ?p ?o . }`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if !query.Select.Distinct {
		t.Error("Distinct should be true")
	}
}

func TestParseQuery_WithFilter(t *testing.T) {
	queryStr := `
		SELECT ?module WHERE {
			?module a code:Module .
			FILTER(REGEX(?module, "main"))
		}
	`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if len(query.Select.Filters) != 1 {
		t.Errorf("Filters count = %d, want 1", len(query.Select.Filters))
	}
}

func TestParseQuery_PrefixExpansion(t *testing.T) {
	queryStr := `
		PREFIX code: <https://schema.codedoc.org/>
		PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
		SELECT ?module WHERE {
			?module rdf:type code:Module .
		}
	`

	query, err := ParseQuery(queryStr)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	if len(query.Select.Prefixes) != 2 {
		t.Errorf("Prefixes count = %d, want 2", len(query.Select.Prefixes))
	}

	if query.Select.Prefixes["code"] != "https://schema.codedoc.org/" {
		t.Errorf("code prefix = %v", query.Select.Prefixes["code"])
	}
}

func TestIsVariable(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"?var", true},
		{"?module", true},
		{"module", false},
		{"<uri>", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsVariable(tt.input)
		if got != tt.want {
			t.Errorf("IsVariable(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsURI(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"<http://example.org>", true},
		{"<#module>", true},
		{"?var", false},
		{"module", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsURI(tt.input)
		if got != tt.want {
			t.Errorf("IsURI(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestStripVariable(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"?var", "var"},
		{"?module", "module"},
		{"module", "module"},
	}

	for _, tt := range tests {
		got := StripVariable(tt.input)
		if got != tt.want {
			t.Errorf("StripVariable(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStripURI(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<http://example.org>", "http://example.org"},
		{"<#module>", "#module"},
		{"plain", "plain"},
	}

	for _, tt := range tests {
		got := StripURI(tt.input)
		if got != tt.want {
			t.Errorf("StripURI(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
