package repl

import (
	"testing"

	"github.com/justin4957/graphfs/pkg/query"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"bool", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"short", "hi", 5, "hi   "},
		{"exact", "hello", 5, "hello"},
		{"long", "hello world", 5, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

func TestFormatTable(t *testing.T) {
	// This is a basic test to ensure formatTable doesn't panic
	config := &Config{
		NoColor: true,
		Prompt:  "test> ",
	}

	result := &query.QueryResult{
		Variables: []string{"s", "p", "o"},
		Bindings: []map[string]string{
			{
				"s": "<#test>",
				"p": "rdf:type",
				"o": "code:Module",
			},
		},
	}

	// Create a minimal REPL instance for testing
	r := &REPL{
		config: config,
		format: "table",
	}

	// Should not panic
	err := r.formatTable(result)
	if err != nil {
		t.Errorf("formatTable() returned error: %v", err)
	}
}

func TestFormatJSON(t *testing.T) {
	config := &Config{
		NoColor: true,
		Prompt:  "test> ",
	}

	result := &query.QueryResult{
		Variables: []string{"s", "p"},
		Bindings: []map[string]string{
			{
				"s": "<#test>",
				"p": "rdf:type",
			},
		},
	}

	r := &REPL{
		config: config,
		format: "json",
	}

	// Should not panic
	err := r.formatJSON(result)
	if err != nil {
		t.Errorf("formatJSON() returned error: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"long", "hello world", 8, "hello..."},
		{"multiline", "hello\nworld", 20, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
