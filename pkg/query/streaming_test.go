/*
# Module: pkg/query/streaming_test.go
Tests for streaming query executor.

## Linked Modules
- [streaming](./streaming.go) - Streaming executor implementation

## Tags
test, query, streaming, pagination

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#streaming_test.go> a code:Module ;
    code:name "pkg/query/streaming_test.go" ;
    code:description "Tests for streaming query executor" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./streaming.go> ;
    code:tags "test", "query", "streaming", "pagination" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"sync"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

// createTestStore creates a triple store with test data
func createTestStore() *store.TripleStore {
	tripleStore := store.NewTripleStore()

	// Add 100 test triples
	for i := 0; i < 100; i++ {
		tripleStore.Add(
			"module"+string(rune('A'+i%26)),
			"imports",
			"module"+string(rune('A'+(i+1)%26)),
		)
	}

	return tripleStore
}

func TestStreamingExecutorCreation(t *testing.T) {
	tripleStore := createTestStore()

	// Test with nil config (should use defaults)
	executor := NewStreamingExecutor(tripleStore, nil)
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
	if executor.config.PageSize != 100 {
		t.Errorf("expected default page size 100, got %d", executor.config.PageSize)
	}

	// Test with custom config
	config := &StreamConfig{
		PageSize:   50,
		BufferSize: 25,
	}
	executor = NewStreamingExecutor(tripleStore, config)
	if executor.config.PageSize != 50 {
		t.Errorf("expected page size 50, got %d", executor.config.PageSize)
	}
	if executor.config.BufferSize != 25 {
		t.Errorf("expected buffer size 25, got %d", executor.config.BufferSize)
	}

	// Test with invalid config values (should use defaults)
	config = &StreamConfig{
		PageSize:   -1,
		BufferSize: 0,
	}
	executor = NewStreamingExecutor(tripleStore, config)
	if executor.config.PageSize != 100 {
		t.Errorf("expected default page size 100 for invalid value, got %d", executor.config.PageSize)
	}
}

func TestResultStreamForEach(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)

	count := 0
	err = stream.ForEach(func(binding map[string]string) error {
		count++
		if _, ok := binding["s"]; !ok {
			t.Error("expected binding to have 's' variable")
		}
		if _, ok := binding["o"]; !ok {
			t.Error("expected binding to have 'o' variable")
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count == 0 {
		t.Error("expected at least one result")
	}
}

func TestResultStreamCollect(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)
	results, err := stream.Collect()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestResultStreamCollectPage(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 100}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)
	page, err := stream.CollectPage(5)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(page) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(page))
	}
}

func TestExecutePaginated(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Test first page
	result, err := executor.ExecutePaginated(query, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", result.PageSize)
	}

	if len(result.Bindings) > 10 {
		t.Errorf("expected at most 10 results, got %d", len(result.Bindings))
	}

	if result.TotalCount == 0 {
		t.Error("expected non-zero total count")
	}

	if result.TotalPages == 0 {
		t.Error("expected non-zero total pages")
	}

	// Test second page if available
	if result.HasMore {
		result2, err := executor.ExecutePaginated(query, 2, 10)
		if err != nil {
			t.Fatalf("unexpected error on page 2: %v", err)
		}

		if result2.Page != 2 {
			t.Errorf("expected page 2, got %d", result2.Page)
		}

		// Results should be different from first page
		if len(result2.Bindings) > 0 && len(result.Bindings) > 0 {
			if result2.Bindings[0]["s"] == result.Bindings[0]["s"] &&
				result2.Bindings[0]["o"] == result.Bindings[0]["o"] {
				t.Error("expected different results on different pages")
			}
		}
	}
}

func TestExecuteStringPaginated(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	result, err := executor.ExecuteStringPaginated("SELECT ?s ?o WHERE { ?s <imports> ?o }", 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PageSize != 5 {
		t.Errorf("expected page size 5, got %d", result.PageSize)
	}

	if len(result.Bindings) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(result.Bindings))
	}
}

func TestExecuteStringStream(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	stream, err := executor.ExecuteStringStream("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := stream.Collect()
	if err != nil {
		t.Fatalf("unexpected error collecting: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestWithPagination(t *testing.T) {
	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s ?p ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Test applying pagination
	paginatedQuery := query.WithPagination(10, 20)

	if paginatedQuery.Select.Limit != 10 {
		t.Errorf("expected limit 10, got %d", paginatedQuery.Select.Limit)
	}

	if paginatedQuery.Select.Offset != 20 {
		t.Errorf("expected offset 20, got %d", paginatedQuery.Select.Offset)
	}

	// Original query should be unchanged
	if query.Select.Limit != 0 {
		t.Errorf("original query limit should be 0, got %d", query.Select.Limit)
	}

	if query.Select.Offset != 0 {
		t.Errorf("original query offset should be 0, got %d", query.Select.Offset)
	}
}

func TestResultStreamCount(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o } LIMIT 5")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)

	// Count should start at 0
	if stream.Count() != 0 {
		t.Errorf("expected initial count 0, got %d", stream.Count())
	}

	// Consume all results
	_, err = stream.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count should be updated
	if stream.Count() == 0 {
		t.Error("expected non-zero count after collecting")
	}
}

func TestConcurrentStreamAccess(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10, BufferSize: 5}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalCount := 0

	// Multiple goroutines reading from the stream
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for binding := range stream.Results {
				mu.Lock()
				if binding != nil {
					totalCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have processed all results
	if totalCount == 0 {
		t.Error("expected to process at least some results")
	}
}

func TestPaginatedResultHasMore(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 5}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// First page with small page size should have more
	result, err := executor.ExecutePaginated(query, 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount > 5 && !result.HasMore {
		t.Error("expected HasMore to be true when there are more results")
	}

	// Last page should not have more
	lastPage := result.TotalPages
	result, err = executor.ExecutePaginated(query, lastPage, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HasMore {
		t.Error("expected HasMore to be false on last page")
	}
}

func TestDefaultStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()

	if config.PageSize != 100 {
		t.Errorf("expected default page size 100, got %d", config.PageSize)
	}

	if config.BufferSize != 100 {
		t.Errorf("expected default buffer size 100, got %d", config.BufferSize)
	}

	if config.ReportProgress {
		t.Error("expected ReportProgress to be false by default")
	}
}

func TestStreamWithProgress(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{
		PageSize:       10,
		ReportProgress: true,
	}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o } LIMIT 20")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	progressCalls := 0
	progressCallback := func(current, total int) {
		progressCalls++
		if current < 0 || current > total {
			t.Errorf("invalid progress: current=%d, total=%d", current, total)
		}
	}

	stream := executor.ExecuteStreamWithProgress(query, progressCallback)
	_, err = stream.Collect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if progressCalls == 0 {
		t.Error("expected progress callback to be called")
	}
}

func TestPaginatedQueryInvalidPage(t *testing.T) {
	tripleStore := createTestStore()
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Page 0 should be treated as page 1
	result, err := executor.ExecutePaginated(query, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page 0 to be treated as page 1, got %d", result.Page)
	}

	// Negative page should also be treated as page 1
	result, err = executor.ExecutePaginated(query, -5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected negative page to be treated as page 1, got %d", result.Page)
	}
}

func TestEmptyResultStream(t *testing.T) {
	tripleStore := store.NewTripleStore() // Empty store
	config := &StreamConfig{PageSize: 10}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <nonexistent> ?o }")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	stream := executor.ExecuteStream(query)
	results, err := stream.Collect()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
