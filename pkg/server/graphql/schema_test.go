package graphql

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func setupTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("/test", tripleStore)

	// Add test modules
	module1 := graph.NewModule("main.go", "<#main.go>")
	module1.Name = "main.go"
	module1.Description = "Main application entry point"
	module1.Language = "go"
	module1.Layer = "application"
	module1.Tags = []string{"entry", "main"}
	module1.Exports = []string{"main"}

	module2 := graph.NewModule("utils/helper.go", "<#helper.go>")
	module2.Name = "helper.go"
	module2.Description = "Helper utilities for processing"
	module2.Language = "go"
	module2.Layer = "utility"
	module2.Tags = []string{"utility", "helpers"}
	module2.Exports = []string{"DoSomething", "Helper"}

	module3 := graph.NewModule("config/config.go", "<#config.go>")
	module3.Name = "config.go"
	module3.Description = "Configuration management"
	module3.Language = "go"
	module3.Layer = "core"
	module3.Tags = []string{"config"}
	module3.Exports = []string{"LoadConfig", "Config"}

	module1.Dependencies = []string{"utils/helper.go", "config/config.go"}
	module2.Dependents = []string{"main.go"}
	module3.Dependents = []string{"main.go"}

	g.AddModule(module1)
	g.AddModule(module2)
	g.AddModule(module3)

	return g
}

func TestBuildSchema(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	if schema.QueryType() == nil {
		t.Error("Schema missing Query type")
	}
}

func TestModuleQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test query by name
	query := `{
		module(name: "main.go") {
			name
			description
			language
			layer
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	module := data["module"].(map[string]interface{})

	if module["name"] != "main.go" {
		t.Errorf("Expected name 'main.go', got %v", module["name"])
	}

	if module["language"] != "go" {
		t.Errorf("Expected language 'go', got %v", module["language"])
	}
}

func TestModulesQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test listing all modules
	query := `{
		modules {
			edges {
				node {
					name
					language
				}
			}
			totalCount
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	modules := data["modules"].(map[string]interface{})

	if modules["totalCount"] != 3 {
		t.Errorf("Expected totalCount 3, got %v", modules["totalCount"])
	}

	edges := modules["edges"].([]interface{})
	if len(edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(edges))
	}
}

func TestModulesQueryWithFilter(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test filtering by layer
	query := `{
		modules(layer: "utility") {
			edges {
				node {
					name
					layer
				}
			}
			totalCount
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	modules := data["modules"].(map[string]interface{})

	if modules["totalCount"] != 1 {
		t.Errorf("Expected totalCount 1, got %v", modules["totalCount"])
	}

	edges := modules["edges"].([]interface{})
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(edges))
	}

	edge := edges[0].(map[string]interface{})
	node := edge["node"].(map[string]interface{})

	if node["name"] != "helper.go" {
		t.Errorf("Expected name 'helper.go', got %v", node["name"])
	}
}

func TestModulesQueryWithPagination(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test pagination with first
	query := `{
		modules(first: 2) {
			edges {
				node {
					name
				}
				cursor
			}
			pageInfo {
				hasNextPage
				hasPreviousPage
			}
			totalCount
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	modules := data["modules"].(map[string]interface{})

	edges := modules["edges"].([]interface{})
	if len(edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(edges))
	}

	pageInfo := modules["pageInfo"].(map[string]interface{})
	if pageInfo["hasNextPage"] != true {
		t.Error("Expected hasNextPage to be true")
	}

	if pageInfo["hasPreviousPage"] != false {
		t.Error("Expected hasPreviousPage to be false")
	}
}

func TestSearchModulesQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test searching by description
	query := `{
		searchModules(query: "utilities") {
			name
			description
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	modules := data["searchModules"].([]interface{})

	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}

	module := modules[0].(map[string]interface{})
	if module["name"] != "helper.go" {
		t.Errorf("Expected name 'helper.go', got %v", module["name"])
	}
}

func TestStatsQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test stats query
	query := `{
		stats {
			totalModules
			totalTriples
			totalRelationships
			modulesByLanguage {
				language
				count
			}
			modulesByLayer {
				layer
				count
			}
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	stats := data["stats"].(map[string]interface{})

	if stats["totalModules"] != 3 {
		t.Errorf("Expected totalModules 3, got %v", stats["totalModules"])
	}

	modulesByLanguage := stats["modulesByLanguage"].([]interface{})
	if len(modulesByLanguage) == 0 {
		t.Error("Expected modulesByLanguage to have entries")
	}

	modulesByLayer := stats["modulesByLayer"].([]interface{})
	if len(modulesByLayer) == 0 {
		t.Error("Expected modulesByLayer to have entries")
	}
}

func TestDependenciesQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test querying dependencies
	query := `{
		module(name: "main.go") {
			name
			dependencies {
				name
				description
			}
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	module := data["module"].(map[string]interface{})

	dependencies := module["dependencies"].([]interface{})
	if len(dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(dependencies))
	}
}

func TestExportsQuery(t *testing.T) {
	g := setupTestGraph()

	schema, err := BuildSchema(g)
	if err != nil {
		t.Fatalf("BuildSchema failed: %v", err)
	}

	// Test querying exports
	query := `{
		module(name: "helper.go") {
			name
			exports {
				name
			}
		}
	}`

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), "graph", g),
	})

	if len(result.Errors) > 0 {
		t.Fatalf("Query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	module := data["module"].(map[string]interface{})

	exports := module["exports"].([]interface{})
	if len(exports) != 2 {
		t.Errorf("Expected 2 exports, got %d", len(exports))
	}

	export1 := exports[0].(map[string]interface{})
	if export1["name"] != "DoSomething" && export1["name"] != "Helper" {
		t.Errorf("Unexpected export name: %v", export1["name"])
	}
}
