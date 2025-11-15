package graphql

import (
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func setupTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("/test", tripleStore)

	// Add test modules
	module1 := graph.NewModule("main.go", "<#main.go>")
	module1.Name = "main.go"
	module1.Description = "Main application"
	module1.Language = "go"
	module1.Layer = "application"
	module1.Tags = []string{"entry", "main"}
	module1.Exports = []string{"main"}

	module2 := graph.NewModule("utils/helper.go", "<#helper.go>")
	module2.Name = "helper.go"
	module2.Description = "Helper utilities"
	module2.Language = "go"
	module2.Layer = "utility"
	module2.Tags = []string{"utility"}
	module2.Exports = []string{"DoSomething", "Helper"}

	module1.Dependencies = []string{"utils/helper.go"}
	module2.Dependents = []string{"main.go"}

	g.AddModule(module1)
	g.AddModule(module2)

	return g
}

func TestNewGenerator(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if gen.graph != g {
		t.Error("Generator graph not set correctly")
	}
}

func TestGenerateSchema(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	schema, err := gen.GenerateSchema(DefaultGenerateOptions())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	if schema == "" {
		t.Fatal("Generated schema is empty")
	}

	// Check for required types
	requiredTypes := []string{
		"type Module",
		"type Export",
		"type Query",
		"type ModuleConnection",
		"type PageInfo",
		"type GraphStats",
	}

	for _, reqType := range requiredTypes {
		if !strings.Contains(schema, reqType) {
			t.Errorf("Schema missing required type: %s", reqType)
		}
	}
}

func TestGenerateModuleType(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	moduleType := gen.generateModuleType()

	// Check for required fields
	requiredFields := []string{
		"id: ID!",
		"uri: String!",
		"name: String!",
		"description: String",
		"path: String!",
		"language: String",
		"layer: String",
		"tags: [String!]!",
		"dependencies: [Module!]!",
		"dependents: [Module!]!",
		"exports: [Export!]!",
	}

	for _, field := range requiredFields {
		if !strings.Contains(moduleType, field) {
			t.Errorf("Module type missing field: %s", field)
		}
	}
}

func TestGenerateQueryType(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	queryType := gen.generateQueryType()

	// Check for required query fields
	requiredQueries := []string{
		"module(",
		"modules(",
		"searchModules(",
		"stats:",
	}

	for _, query := range requiredQueries {
		if !strings.Contains(queryType, query) {
			t.Errorf("Query type missing query: %s", query)
		}
	}

	// Check for filter parameters
	if !strings.Contains(queryType, "language: String") {
		t.Error("modules query missing language parameter")
	}

	if !strings.Contains(queryType, "layer: String") {
		t.Error("modules query missing layer parameter")
	}

	if !strings.Contains(queryType, "tag: String") {
		t.Error("modules query missing tag parameter")
	}
}

func TestGenerateConnectionTypes(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	connectionTypes := gen.generateConnectionTypes()

	// Check for pagination types
	requiredTypes := []string{
		"type ModuleConnection",
		"type ModuleEdge",
		"type PageInfo",
		"hasNextPage: Boolean!",
		"hasPreviousPage: Boolean!",
		"startCursor: String",
		"endCursor: String",
	}

	for _, reqType := range requiredTypes {
		if !strings.Contains(connectionTypes, reqType) {
			t.Errorf("Connection types missing: %s", reqType)
		}
	}
}

func TestGenerateSchemaHeader(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	schema, err := gen.GenerateSchema(DefaultGenerateOptions())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// Check for header comments
	if !strings.Contains(schema, "# GraphQL Schema for GraphFS") {
		t.Error("Schema missing header comment")
	}

	// Check for schema definition
	if !strings.Contains(schema, "schema {") {
		t.Error("Schema missing schema definition")
	}

	if !strings.Contains(schema, "query: Query") {
		t.Error("Schema missing Query root type definition")
	}
}

func TestGetAvailableTypes(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	types := gen.GetAvailableTypes()

	if len(types) == 0 {
		t.Fatal("GetAvailableTypes returned empty list")
	}

	// Check for essential types
	essentialTypes := []string{"Module", "Query", "Export"}
	for _, essentialType := range essentialTypes {
		found := false
		for _, t := range types {
			if t == essentialType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAvailableTypes missing essential type: %s", essentialType)
		}
	}

	// Check that types are sorted
	for i := 1; i < len(types); i++ {
		if types[i-1] > types[i] {
			t.Error("GetAvailableTypes not sorted")
			break
		}
	}
}

func TestValidateSchema(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	// Test valid schema
	validSchema, err := gen.GenerateSchema(DefaultGenerateOptions())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	if err := gen.ValidateSchema(validSchema); err != nil {
		t.Errorf("ValidateSchema failed for valid schema: %v", err)
	}

	// Test invalid schemas
	testCases := []struct {
		name   string
		schema string
		errMsg string
	}{
		{
			name:   "missing Query type",
			schema: "type Module { id: ID! }",
			errMsg: "missing required type: Query",
		},
		{
			name:   "missing Module type",
			schema: "type Query { test: String }",
			errMsg: "missing required type: Module",
		},
		{
			name:   "missing schema definition",
			schema: "type Query { test: String } type Module { id: ID! }",
			errMsg: "missing schema definition",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := gen.ValidateSchema(tc.schema)
			if err == nil {
				t.Errorf("Expected validation error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.errMsg, err.Error())
			}
		})
	}
}

func TestGenerateSchemaFromGraph(t *testing.T) {
	g := setupTestGraph()

	schema, err := GenerateSchemaFromGraph(g, DefaultGenerateOptions())
	if err != nil {
		t.Fatalf("GenerateSchemaFromGraph failed: %v", err)
	}

	if schema == "" {
		t.Fatal("Generated schema is empty")
	}

	// Verify it's valid
	gen := NewGenerator(g)
	if err := gen.ValidateSchema(schema); err != nil {
		t.Errorf("Generated schema is invalid: %v", err)
	}
}

func TestDefaultGenerateOptions(t *testing.T) {
	opts := DefaultGenerateOptions()

	if opts.IncludeTypes == nil {
		t.Error("IncludeTypes should be initialized")
	}

	if opts.ExcludeTypes == nil {
		t.Error("ExcludeTypes should be initialized")
	}

	if opts.CustomTypes == nil {
		t.Error("CustomTypes should be initialized")
	}
}

func TestSchemaDescriptions(t *testing.T) {
	g := setupTestGraph()
	gen := NewGenerator(g)

	schema, err := gen.GenerateSchema(DefaultGenerateOptions())
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// Check for GraphQL description strings (triple quotes)
	if !strings.Contains(schema, "\"\"\"") {
		t.Error("Schema missing GraphQL description strings")
	}

	// Check for specific descriptions
	descriptions := []string{
		"Represents a code module",
		"Root query type",
		"Information about pagination",
	}

	for _, desc := range descriptions {
		if !strings.Contains(schema, desc) {
			t.Errorf("Schema missing description: %s", desc)
		}
	}
}
