package parser

import (
	"path/filepath"
	"testing"
)

// TestParseMinimalAppMainGo tests parsing the main.go from minimal-app
func TestParseMinimalAppMainGo(t *testing.T) {
	parser := NewParser()
	path := filepath.Join("..", "..", "examples", "minimal-app", "main.go")

	triples, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(triples) == 0 {
		t.Fatal("Parse() returned no triples")
	}

	// Check for key triples
	foundModule := false
	foundName := false
	foundLinksTo := false
	foundExports := false

	for _, triple := range triples {
		t.Logf("Triple: %s -> %s -> %s (%s)",
			triple.Subject, triple.Predicate, triple.Object.String(), triple.Object.Type())

		// Check for module declaration
		if triple.Predicate == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" &&
			triple.Object.String() == "https://schema.codedoc.org/Module" {
			foundModule = true
		}

		// Check for name property
		if contains(triple.Predicate, "name") && triple.Object.String() == "main.go" {
			foundName = true
		}

		// Check for linksTo relationships
		if contains(triple.Predicate, "linksTo") {
			foundLinksTo = true
		}

		// Check for exports
		if contains(triple.Predicate, "exports") {
			foundExports = true
		}
	}

	if !foundModule {
		t.Error("Parse() did not find code:Module declaration")
	}

	if !foundName {
		t.Error("Parse() did not find code:name property")
	}

	if !foundLinksTo {
		t.Error("Parse() did not find code:linksTo relationships")
	}

	if !foundExports {
		t.Error("Parse() did not find code:exports")
	}

	t.Logf("Successfully parsed %d triples from main.go", len(triples))
}

// TestParseMinimalAppAuthService tests parsing services/auth.go
func TestParseMinimalAppAuthService(t *testing.T) {
	parser := NewParser()
	path := filepath.Join("..", "..", "examples", "minimal-app", "services", "auth.go")

	triples, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(triples) == 0 {
		t.Fatal("Parse() returned no triples")
	}

	// Check for security metadata
	foundSecurityBoundary := false
	foundSecurityCritical := false

	for _, triple := range triples {
		if contains(triple.Predicate, "security") {
			if contains(triple.Predicate, "securityBoundary") {
				foundSecurityBoundary = true
			}
			if contains(triple.Predicate, "securityCritical") {
				foundSecurityCritical = true
			}
		}
	}

	if !foundSecurityBoundary && !foundSecurityCritical {
		t.Log("Warning: No security metadata found in auth.go (expected security boundary)")
	}

	t.Logf("Successfully parsed %d triples from auth.go", len(triples))
}

// TestParseMinimalAppUserModel tests parsing models/user.go
func TestParseMinimalAppUserModel(t *testing.T) {
	parser := NewParser()
	path := filepath.Join("..", "..", "examples", "minimal-app", "models", "user.go")

	triples, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(triples) == 0 {
		t.Fatal("Parse() returned no triples")
	}

	// Check for type declarations
	foundUserType := false
	foundUserRoleType := false
	foundValidateFunction := false

	for _, triple := range triples {
		if contains(triple.Object.String(), "User") && triple.Object.Type() == "uri" {
			foundUserType = true
		}
		if contains(triple.Object.String(), "UserRole") {
			foundUserRoleType = true
		}
		if contains(triple.Object.String(), "ValidateUser") {
			foundValidateFunction = true
		}
	}

	if !foundUserType {
		t.Log("Warning: User type declaration not found")
	}

	if !foundUserRoleType {
		t.Log("Warning: UserRole type declaration not found")
	}

	if !foundValidateFunction {
		t.Log("Warning: ValidateUser function declaration not found")
	}

	t.Logf("Successfully parsed %d triples from user.go", len(triples))
}

// TestParseAllMinimalAppFiles tests parsing all files in minimal-app
func TestParseAllMinimalAppFiles(t *testing.T) {
	parser := NewParser()

	files := []string{
		"main.go",
		"models/user.go",
		"services/auth.go",
		"services/user.go",
		"utils/logger.go",
		"utils/crypto.go",
		"utils/validator.go",
	}

	totalTriples := 0

	for _, file := range files {
		path := filepath.Join("..", "..", "examples", "minimal-app", file)

		triples, err := parser.Parse(path)
		if err != nil {
			t.Errorf("Parse(%s) error = %v", file, err)
			continue
		}

		if len(triples) == 0 {
			t.Errorf("Parse(%s) returned no triples", file)
			continue
		}

		totalTriples += len(triples)
		t.Logf("%s: %d triples", file, len(triples))
	}

	t.Logf("\nTotal triples parsed from minimal-app: %d", totalTriples)

	if totalTriples < 40 {
		t.Errorf("Expected at least 40 total triples from minimal-app, got %d", totalTriples)
	}
}

// TestParseLinkedDocSchemaTypes tests that all schema types are recognized
func TestParseLinkedDocSchemaTypes(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		check   func([]Triple) bool
	}{
		{
			name: "code:Module type",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#test> a code:Module .
<!-- End LinkedDoc RDF -->
*/`,
			check: func(triples []Triple) bool {
				return len(triples) > 0 &&
					contains(triples[0].Object.String(), "Module")
			},
		},
		{
			name: "code:Function type",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<#foo> a code:Function .
<!-- End LinkedDoc RDF -->
*/`,
			check: func(triples []Triple) bool {
				return len(triples) > 0 &&
					contains(triples[0].Object.String(), "Function")
			},
		},
		{
			name: "security metadata",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix sec: <https://schema.codedoc.org/security/> .
<#test> sec:securityCritical true .
<!-- End LinkedDoc RDF -->
*/`,
			check: func(triples []Triple) bool {
				return len(triples) > 0 &&
					contains(triples[0].Predicate, "securityCritical")
			},
		},
		{
			name: "architecture rules",
			content: `/*
<!-- LinkedDoc RDF -->
@prefix arch: <https://schema.codedoc.org/architecture/> .
<#rule1> a arch:Rule .
<!-- End LinkedDoc RDF -->
*/`,
			check: func(triples []Triple) bool {
				return len(triples) > 0 &&
					contains(triples[0].Object.String(), "Rule")
			},
		},
	}

	parser := NewParser()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			triples, err := parser.ParseString(tc.content)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			if !tc.check(triples) {
				t.Errorf("Schema type check failed")
				for i, triple := range triples {
					t.Logf("Triple %d: %s -> %s -> %s",
						i, triple.Subject, triple.Predicate, triple.Object.String())
				}
			}
		})
	}
}
