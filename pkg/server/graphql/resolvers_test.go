package graphql

import (
	"context"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func TestNewResolver(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	if resolver == nil {
		t.Fatal("NewResolver returned nil")
	}

	if resolver.graph != g {
		t.Error("Resolver graph not set correctly")
	}
}

func TestResolverModule(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	tests := []struct {
		name     string
		args     map[string]interface{}
		expected string
		isNil    bool
	}{
		{
			name:     "find by name",
			args:     map[string]interface{}{"name": "main.go"},
			expected: "main.go",
			isNil:    false,
		},
		{
			name:     "find by path",
			args:     map[string]interface{}{"path": "utils/helper.go"},
			expected: "helper.go",
			isNil:    false,
		},
		{
			name:     "find by uri",
			args:     map[string]interface{}{"uri": "<#config.go>"},
			expected: "config.go",
			isNil:    false,
		},
		{
			name:     "not found",
			args:     map[string]interface{}{"name": "nonexistent.go"},
			expected: "",
			isNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := graphql.ResolveParams{
				Args:    tt.args,
				Context: context.Background(),
			}

			result, err := resolver.Module(params)
			if err != nil {
				t.Fatalf("Module resolver failed: %v", err)
			}

			if tt.isNil {
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected non-nil result")
				}
				module := result.(*graph.Module)
				if module.Name != tt.expected {
					t.Errorf("Expected name %s, got %s", tt.expected, module.Name)
				}
			}
		})
	}
}

func TestResolverModules(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	tests := []struct {
		name          string
		args          map[string]interface{}
		expectedCount int
	}{
		{
			name:          "all modules",
			args:          map[string]interface{}{},
			expectedCount: 3,
		},
		{
			name:          "filter by language",
			args:          map[string]interface{}{"language": "go"},
			expectedCount: 3,
		},
		{
			name:          "filter by layer",
			args:          map[string]interface{}{"layer": "utility"},
			expectedCount: 1,
		},
		{
			name:          "filter by tag",
			args:          map[string]interface{}{"tag": "config"},
			expectedCount: 1,
		},
		{
			name:          "no matches",
			args:          map[string]interface{}{"language": "python"},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := graphql.ResolveParams{
				Args:    tt.args,
				Context: context.Background(),
			}

			result, err := resolver.Modules(params)
			if err != nil {
				t.Fatalf("Modules resolver failed: %v", err)
			}

			connection := result.(map[string]interface{})
			totalCount := connection["totalCount"].(int)

			if totalCount != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, totalCount)
			}

			edges := connection["edges"].([]map[string]interface{})
			if len(edges) != tt.expectedCount {
				t.Errorf("Expected %d edges, got %d", tt.expectedCount, len(edges))
			}
		})
	}
}

func TestResolverModulesPagination(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	// Test first page
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"first": 2,
		},
		Context: context.Background(),
	}

	result, err := resolver.Modules(params)
	if err != nil {
		t.Fatalf("Modules resolver failed: %v", err)
	}

	connection := result.(map[string]interface{})
	edges := connection["edges"].([]map[string]interface{})

	if len(edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(edges))
	}

	pageInfo := connection["pageInfo"].(map[string]interface{})
	if pageInfo["hasNextPage"] != true {
		t.Error("Expected hasNextPage to be true")
	}

	if pageInfo["hasPreviousPage"] != false {
		t.Error("Expected hasPreviousPage to be false")
	}

	// Test with cursor
	endCursor := pageInfo["endCursor"].(string)
	params2 := graphql.ResolveParams{
		Args: map[string]interface{}{
			"first": 2,
			"after": endCursor,
		},
		Context: context.Background(),
	}

	result2, err := resolver.Modules(params2)
	if err != nil {
		t.Fatalf("Modules resolver failed: %v", err)
	}

	connection2 := result2.(map[string]interface{})
	edges2 := connection2["edges"].([]map[string]interface{})

	if len(edges2) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(edges2))
	}

	pageInfo2 := connection2["pageInfo"].(map[string]interface{})
	if pageInfo2["hasNextPage"] != false {
		t.Error("Expected hasNextPage to be false")
	}

	if pageInfo2["hasPreviousPage"] != true {
		t.Error("Expected hasPreviousPage to be true")
	}
}

func TestResolverSearchModules(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "search in description",
			query:         "utilities",
			expectedCount: 1,
			expectedNames: []string{"helper.go"},
		},
		{
			name:          "search in name",
			query:         "config",
			expectedCount: 1,
			expectedNames: []string{"config.go"},
		},
		{
			name:          "search in tags",
			query:         "main",
			expectedCount: 1,
			expectedNames: []string{"main.go"},
		},
		{
			name:          "case insensitive",
			query:         "HELPER",
			expectedCount: 1,
			expectedNames: []string{"helper.go"},
		},
		{
			name:          "no matches",
			query:         "xyz123",
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := graphql.ResolveParams{
				Args: map[string]interface{}{
					"query": tt.query,
				},
				Context: context.Background(),
			}

			result, err := resolver.SearchModules(params)
			if err != nil {
				t.Fatalf("SearchModules resolver failed: %v", err)
			}

			modules := result.([]*graph.Module)
			if len(modules) != tt.expectedCount {
				t.Errorf("Expected %d modules, got %d", tt.expectedCount, len(modules))
			}

			for i, mod := range modules {
				if i < len(tt.expectedNames) {
					if mod.Name != tt.expectedNames[i] {
						t.Errorf("Expected name %s, got %s", tt.expectedNames[i], mod.Name)
					}
				}
			}
		})
	}
}

func TestResolverStats(t *testing.T) {
	g := setupTestGraph()
	resolver := NewResolver(g)

	params := graphql.ResolveParams{
		Args:    map[string]interface{}{},
		Context: context.Background(),
	}

	result, err := resolver.Stats(params)
	if err != nil {
		t.Fatalf("Stats resolver failed: %v", err)
	}

	stats := result.(map[string]interface{})

	// Check total modules
	if stats["totalModules"] != 3 {
		t.Errorf("Expected totalModules 3, got %v", stats["totalModules"])
	}

	// Check total triples (may be 0 if graph hasn't been serialized)
	totalTriples := stats["totalTriples"].(int)
	if totalTriples < 0 {
		t.Error("Expected totalTriples >= 0")
	}

	// Check total relationships
	totalRelationships := stats["totalRelationships"].(int)
	if totalRelationships != 2 {
		t.Errorf("Expected totalRelationships 2, got %d", totalRelationships)
	}

	// Check modules by language
	modulesByLanguage := stats["modulesByLanguage"].([]map[string]interface{})
	if len(modulesByLanguage) != 1 {
		t.Errorf("Expected 1 language entry, got %d", len(modulesByLanguage))
	}

	langStats := modulesByLanguage[0]
	if langStats["language"] != "go" {
		t.Errorf("Expected language 'go', got %v", langStats["language"])
	}
	if langStats["count"] != 3 {
		t.Errorf("Expected count 3, got %v", langStats["count"])
	}

	// Check modules by layer
	modulesByLayer := stats["modulesByLayer"].([]map[string]interface{})
	if len(modulesByLayer) != 3 {
		t.Errorf("Expected 3 layer entries, got %d", len(modulesByLayer))
	}
}

func TestEncodeCursor(t *testing.T) {
	cursor := encodeCursor(42)

	if cursor == "" {
		t.Error("Expected non-empty cursor")
	}

	// Cursor should be base64 encoded
	if !strings.Contains(cursor, "=") && len(cursor)%4 != 0 {
		// Base64 strings are typically padded with = or have length divisible by 4
		t.Error("Cursor doesn't appear to be base64 encoded")
	}
}

func TestDecodeCursor(t *testing.T) {
	// Test valid cursor
	cursor := encodeCursor(42)
	idx, err := decodeCursor(cursor)
	if err != nil {
		t.Fatalf("decodeCursor failed: %v", err)
	}

	if idx != 42 {
		t.Errorf("Expected index 42, got %d", idx)
	}

	// Test invalid cursor
	_, err = decodeCursor("invalid")
	if err == nil {
		t.Error("Expected error for invalid cursor")
	}

	// Test malformed cursor
	_, err = decodeCursor("YWJjZGVm") // "abcdef" in base64
	if err == nil {
		t.Error("Expected error for malformed cursor")
	}
}

func TestCursorRoundTrip(t *testing.T) {
	testIndices := []int{0, 1, 10, 100, 999}

	for _, idx := range testIndices {
		t.Run(string(rune(idx)), func(t *testing.T) {
			cursor := encodeCursor(idx)
			decoded, err := decodeCursor(cursor)
			if err != nil {
				t.Fatalf("Round trip failed: %v", err)
			}

			if decoded != idx {
				t.Errorf("Expected %d, got %d", idx, decoded)
			}
		})
	}
}

func setupTestGraphMultiLanguage() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("/test", tripleStore)

	// Go module
	goMod := graph.NewModule("main.go", "<#main.go>")
	goMod.Name = "main.go"
	goMod.Language = "go"
	goMod.Layer = "application"

	// Python module
	pyMod := graph.NewModule("script.py", "<#script.py>")
	pyMod.Name = "script.py"
	pyMod.Language = "python"
	pyMod.Layer = "scripts"

	// JavaScript module
	jsMod := graph.NewModule("index.js", "<#index.js>")
	jsMod.Name = "index.js"
	jsMod.Language = "javascript"
	jsMod.Layer = "frontend"

	g.AddModule(goMod)
	g.AddModule(pyMod)
	g.AddModule(jsMod)

	return g
}

func TestResolverStatsMultiLanguage(t *testing.T) {
	g := setupTestGraphMultiLanguage()
	resolver := NewResolver(g)

	params := graphql.ResolveParams{
		Args:    map[string]interface{}{},
		Context: context.Background(),
	}

	result, err := resolver.Stats(params)
	if err != nil {
		t.Fatalf("Stats resolver failed: %v", err)
	}

	stats := result.(map[string]interface{})

	// Check modules by language
	modulesByLanguage := stats["modulesByLanguage"].([]map[string]interface{})
	if len(modulesByLanguage) != 3 {
		t.Errorf("Expected 3 language entries, got %d", len(modulesByLanguage))
	}

	// Check modules by layer
	modulesByLayer := stats["modulesByLayer"].([]map[string]interface{})
	if len(modulesByLayer) != 3 {
		t.Errorf("Expected 3 layer entries, got %d", len(modulesByLayer))
	}
}
