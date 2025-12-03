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

// TestBlankNodeParsing tests W3C Turtle blank node syntax support.
// Addresses issue #72: Fix RDF parser to support Turtle blank node syntax.
func TestBlankNodeParsing(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantTriples int
		checkFn     func(*testing.T, []Triple)
	}{
		{
			name: "simple blank node inline",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> code:linksTo [ code:name "auth" ; code:path "../auth.go" ] .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 1, // 1 triple with BlankNodeObject containing 2 inner triples
			checkFn: func(t *testing.T, triples []Triple) {
				if len(triples) == 0 {
					t.Fatal("Expected at least one triple")
				}
				// Blank node should be parsed as BlankNodeObject
				bn, ok := triples[0].Object.(BlankNodeObject)
				if !ok {
					t.Errorf("First triple object should be blank node, got %T", triples[0].Object)
					return
				}
				if len(bn.Triples) != 2 {
					t.Errorf("Blank node should have 2 inner triples, got %d", len(bn.Triples))
				}
			},
		},
		{
			name: "multi-line blank node",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> code:linksTo [
    code:name "resolver" ;
    code:path "../resolver.go" ;
    code:relationship "uses"
] .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 1, // 1 triple with BlankNodeObject containing 3 inner triples
			checkFn: func(t *testing.T, triples []Triple) {
				if len(triples) == 0 {
					t.Fatal("Expected at least one triple")
				}
				// Blank node should contain 3 property triples
				bn, ok := triples[0].Object.(BlankNodeObject)
				if !ok {
					t.Errorf("Expected blank node object, got %T", triples[0].Object)
					return
				}
				if len(bn.Triples) != 3 {
					t.Errorf("Blank node should have 3 triples, got %d", len(bn.Triples))
				}
				// Verify predicates
				predicates := make(map[string]bool)
				for _, inner := range bn.Triples {
					predicates[inner.Predicate] = true
				}
				expectedPredicates := []string{
					"https://schema.codedoc.org/name",
					"https://schema.codedoc.org/path",
					"https://schema.codedoc.org/relationship",
				}
				for _, pred := range expectedPredicates {
					if !predicates[pred] {
						t.Errorf("Missing expected predicate: %s", pred)
					}
				}
			},
		},
		{
			name: "nested blank nodes",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> code:dependency [
    code:module [ code:name "auth" ; code:version "1.0" ] ;
    code:type "direct"
] .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 1, // 1 triple with nested BlankNodeObject
			checkFn: func(t *testing.T, triples []Triple) {
				if len(triples) == 0 {
					t.Fatal("Expected at least one triple")
				}
				// Outer blank node
				outerBn, ok := triples[0].Object.(BlankNodeObject)
				if !ok {
					t.Errorf("Expected blank node object, got %T", triples[0].Object)
					return
				}
				if len(outerBn.Triples) != 2 {
					t.Errorf("Outer blank node should have 2 triples, got %d", len(outerBn.Triples))
				}
				// Find the nested blank node (module property)
				var nestedBn *BlankNodeObject
				for _, inner := range outerBn.Triples {
					if bn, ok := inner.Object.(BlankNodeObject); ok {
						nestedBn = &bn
						break
					}
				}
				if nestedBn == nil {
					t.Error("Expected nested blank node for code:module")
					return
				}
				if len(nestedBn.Triples) != 2 {
					t.Errorf("Nested blank node should have 2 triples, got %d", len(nestedBn.Triples))
				}
			},
		},
		{
			name: "blank node with URI object",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test.go> code:linksTo [ code:target <../auth.go> ; code:type "import" ] .
<!-- End LinkedDoc RDF -->
*/`,
			wantTriples: 1,
			checkFn: func(t *testing.T, triples []Triple) {
				if len(triples) == 0 {
					t.Fatal("Expected at least one triple")
				}
				bn, ok := triples[0].Object.(BlankNodeObject)
				if !ok {
					t.Errorf("Expected blank node object, got %T", triples[0].Object)
					return
				}
				// Should have a URI object for target
				hasURIObject := false
				for _, inner := range bn.Triples {
					if inner.Object.Type() == "uri" {
						hasURIObject = true
						break
					}
				}
				if !hasURIObject {
					t.Error("Expected at least one URI object in blank node")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			triples, err := parser.ParseString(tt.content)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			if len(triples) != tt.wantTriples {
				t.Errorf("ParseString() got %d triples, want %d", len(triples), tt.wantTriples)
				for i, tr := range triples {
					t.Logf("Triple %d: %s | %s | %s", i, tr.Subject, tr.Predicate, tr.Object.String())
				}
			}

			if tt.checkFn != nil {
				tt.checkFn(t, triples)
			}
		})
	}
}
