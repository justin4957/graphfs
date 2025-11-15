package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/query"
)

func setupTestExecutor() *query.Executor {
	// Create test triple store
	tripleStore := store.NewTripleStore()

	// Add test data
	tripleStore.Add(
		"<#test.go>",
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		"https://schema.codedoc.org/Module",
	)
	tripleStore.Add(
		"<#test.go>",
		"https://schema.codedoc.org/name",
		"test.go",
	)
	tripleStore.Add(
		"<#test.go>",
		"https://schema.codedoc.org/description",
		"Test module",
	)

	return query.NewExecutor(tripleStore)
}

func TestSPARQLHandler_GET(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "SELECT ?name WHERE { ?s <https://schema.codedoc.org/name> ?name }"
	req := httptest.NewRequest(http.MethodGet, "/sparql?query="+url.QueryEscape(queryStr), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/sparql-results+json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}
}

func TestSPARQLHandler_POST(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "SELECT ?name WHERE { ?s <https://schema.codedoc.org/name> ?name }"
	req := httptest.NewRequest(http.MethodPost, "/sparql", strings.NewReader(queryStr))
	req.Header.Set("Content-Type", "application/sparql-query")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestSPARQLHandler_CSVFormat(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "SELECT ?name WHERE { ?s <https://schema.codedoc.org/name> ?name }"
	req := httptest.NewRequest(http.MethodGet, "/sparql?query="+url.QueryEscape(queryStr)+"&format=csv", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("Expected CSV content type, got %s", contentType)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "name") {
		t.Error("Expected CSV header 'name' in output")
	}
}

func TestSPARQLHandler_XMLFormat(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "SELECT ?name WHERE { ?s <https://schema.codedoc.org/name> ?name }"
	req := httptest.NewRequest(http.MethodGet, "/sparql?query="+url.QueryEscape(queryStr)+"&format=xml", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/sparql-results+xml") {
		t.Errorf("Expected XML content type, got %s", contentType)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "<?xml") {
		t.Error("Expected XML declaration in output")
	}
}

func TestSPARQLHandler_TSVFormat(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "SELECT ?name WHERE { ?s <https://schema.codedoc.org/name> ?name }"
	req := httptest.NewRequest(http.MethodGet, "/sparql?query="+url.QueryEscape(queryStr)+"&format=tsv", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/tab-separated-values") {
		t.Errorf("Expected TSV content type, got %s", contentType)
	}
}

func TestSPARQLHandler_InvalidQuery(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	queryStr := "INVALID QUERY"
	req := httptest.NewRequest(http.MethodGet, "/sparql?query="+url.QueryEscape(queryStr), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestSPARQLHandler_MissingQuery(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	req := httptest.NewRequest(http.MethodGet, "/sparql", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestSPARQLHandler_CORS(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	req := httptest.NewRequest(http.MethodOptions, "/sparql", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", rec.Code)
	}

	corsHeader := rec.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got %s", corsHeader)
	}
}

func TestSPARQLHandler_MethodNotAllowed(t *testing.T) {
	executor := setupTestExecutor()
	handler := NewSPARQLHandler(executor, true)

	req := httptest.NewRequest(http.MethodPut, "/sparql", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}
