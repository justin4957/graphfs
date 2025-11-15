package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestHandleModules(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Meta["total"] != 3 {
		t.Errorf("Expected 3 modules, got %d", response.Meta["total"])
	}
}

func TestHandleModulesWithFilter(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules?layer=utility", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Meta["total"] != 1 {
		t.Errorf("Expected 1 module, got %d", response.Meta["total"])
	}
}

func TestHandleModulesWithPagination(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules?limit=2&offset=0", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Meta["count"] != 2 {
		t.Errorf("Expected 2 modules in page, got %d", response.Meta["count"])
	}

	if response.Meta["total"] != 3 {
		t.Errorf("Expected total 3 modules, got %d", response.Meta["total"])
	}

	if response.Links["next"] == "" {
		t.Error("Expected next link")
	}
}

func TestHandleModuleByID(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules/main.go", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ModuleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != "main.go" {
		t.Errorf("Expected name main.go, got %s", response.Name)
	}

	if len(response.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(response.Dependencies))
	}
}

func TestHandleModuleNotFound(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules/nonexistent.go", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleModulesSearch(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules/search?q=utilities", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["count"].(float64) != 1 {
		t.Errorf("Expected 1 result, got %v", response["count"])
	}
}

func TestHandleAnalysisStats(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/analysis/stats", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["totalModules"].(float64) != 3 {
		t.Errorf("Expected 3 modules, got %v", response["totalModules"])
	}

	if response["totalRelationships"].(float64) != 2 {
		t.Errorf("Expected 2 relationships, got %v", response["totalRelationships"])
	}
}

func TestHandleTags(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/tags", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["total"].(float64) < 1 {
		t.Error("Expected at least 1 tag")
	}
}

func TestHandleExports(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/exports", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["count"].(float64) < 1 {
		t.Error("Expected at least 1 export")
	}
}

func TestCORSHeaders(t *testing.T) {
	g := setupTestGraph()
	handler := NewHandler(g, true)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v1/modules", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header")
	}
}
