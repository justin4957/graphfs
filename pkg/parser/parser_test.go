package parser

import (
	"testing"
)

func TestExtractLinkedDoc(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid LinkedDoc block",
			content: `/*
Some documentation

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test> a code:Module .
<!-- End LinkedDoc RDF -->
*/`,
			want:    "@prefix code: <https://schema.codedoc.org/> .\n<#test> a code:Module .",
			wantErr: false,
		},
		{
			name:    "no LinkedDoc block",
			content: `/* Just a comment */`,
			want:    "",
			wantErr: false,
		},
		{
			name: "missing end marker",
			content: `/*
<!-- LinkedDoc RDF -->
<#test> a code:Module .
*/`,
			wantErr:     true,
			errContains: "not closed",
		},
		{
			name: "empty LinkedDoc block",
			content: `/*
<!-- LinkedDoc RDF -->
<!-- End LinkedDoc RDF -->
*/`,
			want:    "",
			wantErr: false,
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ExtractLinkedDoc(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractLinkedDoc() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ExtractLinkedDoc() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractLinkedDoc() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ExtractLinkedDoc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePrefix(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantPrefix string
		wantURI    string
		wantErr    bool
	}{
		{
			name:       "valid prefix",
			line:       "@prefix code: <https://schema.codedoc.org/> .",
			wantPrefix: "code",
			wantURI:    "https://schema.codedoc.org/",
			wantErr:    false,
		},
		{
			name:       "rdf prefix",
			line:       "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .",
			wantPrefix: "rdf",
			wantURI:    "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			wantErr:    false,
		},
		{
			name:    "invalid syntax",
			line:    "@prefix invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			err := parser.parsePrefix(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePrefix() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parsePrefix() unexpected error: %v", err)
				return
			}

			if uri, ok := parser.prefixes[tt.wantPrefix]; !ok || uri != tt.wantURI {
				t.Errorf("parsePrefix() prefix[%s] = %q, want %q", tt.wantPrefix, uri, tt.wantURI)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantTriples int
		checkTriple func([]Triple) bool
	}{
		{
			name: "simple module declaration",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> a code:Module .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 1,
			checkTriple: func(triples []Triple) bool {
				if len(triples) == 0 {
					return false
				}
				t := triples[0]
				return t.Subject == "<#test.go>" &&
					t.Predicate == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" &&
					t.Object.String() == "https://schema.codedoc.org/Module"
			},
		},
		{
			name: "module with properties",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> a code:Module ;
    code:name "test.go" ;
    code:description "Test module" .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 3,
			checkTriple: func(triples []Triple) bool {
				return len(triples) == 3
			},
		},
		{
			name: "multiple values",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> code:exports <#Foo>, <#Bar>, <#Baz> .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 3,
			checkTriple: func(triples []Triple) bool {
				return len(triples) == 3 &&
					triples[0].Object.String() == "#Foo" &&
					triples[1].Object.String() == "#Bar" &&
					triples[2].Object.String() == "#Baz"
			},
		},
		{
			name:        "no LinkedDoc",
			content:     `/* Just a regular comment */`,
			wantTriples: 0,
			checkTriple: func(triples []Triple) bool {
				return len(triples) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			triples, err := parser.ParseString(tt.content)

			if err != nil {
				t.Errorf("ParseString() unexpected error: %v", err)
				return
			}

			if len(triples) != tt.wantTriples {
				t.Errorf("ParseString() got %d triples, want %d", len(triples), tt.wantTriples)
			}

			if tt.checkTriple != nil && !tt.checkTriple(triples) {
				t.Errorf("ParseString() triple validation failed")
				for i, triple := range triples {
					t.Logf("Triple %d: %s -> %s -> %s", i, triple.Subject, triple.Predicate, triple.Object.String())
				}
			}
		})
	}
}

func TestParseObject(t *testing.T) {
	parser := NewParser()
	parser.prefixes["code"] = "https://schema.codedoc.org/"

	tests := []struct {
		name     string
		objStr   string
		wantType string
		wantVal  string
	}{
		{
			name:     "URI with angle brackets",
			objStr:   "<#test>",
			wantType: "uri",
			wantVal:  "#test",
		},
		{
			name:     "prefixed URI",
			objStr:   "code:Module",
			wantType: "uri",
			wantVal:  "https://schema.codedoc.org/Module",
		},
		{
			name:     "quoted literal",
			objStr:   `"test value"`,
			wantType: "literal",
			wantVal:  "test value",
		},
		{
			name:     "unquoted literal",
			objStr:   "true",
			wantType: "literal",
			wantVal:  "true",
		},
		{
			name:     "blank node",
			objStr:   "[...]",
			wantType: "bnode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := parser.parseObject(tt.objStr)

			if obj.Type() != tt.wantType {
				t.Errorf("parseObject() type = %s, want %s", obj.Type(), tt.wantType)
			}

			if tt.wantVal != "" && obj.String() != tt.wantVal {
				t.Errorf("parseObject() value = %s, want %s", obj.String(), tt.wantVal)
			}
		})
	}
}

func TestExpandPrefix(t *testing.T) {
	parser := NewParser()
	parser.prefixes["code"] = "https://schema.codedoc.org/"
	parser.prefixes["rdf"] = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"

	tests := []struct {
		name        string
		prefixedURI string
		want        string
	}{
		{
			name:        "code prefix",
			prefixedURI: "code:Module",
			want:        "https://schema.codedoc.org/Module",
		},
		{
			name:        "rdf prefix",
			prefixedURI: "rdf:type",
			want:        "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		},
		{
			name:        "unknown prefix",
			prefixedURI: "unknown:foo",
			want:        "unknown:foo",
		},
		{
			name:        "no prefix",
			prefixedURI: "plain",
			want:        "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.expandPrefix(tt.prefixedURI)
			if got != tt.want {
				t.Errorf("expandPrefix() = %s, want %s", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && indexSubstring(s, substr) >= 0))
}

func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
