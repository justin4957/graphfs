/*
# Module: pkg/query/streaming.go
Streaming query executor for large result sets.

Provides streaming and pagination support to handle massive query results
without loading everything into memory.

## Linked Modules
- [executor](./executor.go) - Base query executor
- [query](./query.go) - Query data structures

## Tags
query, streaming, pagination, memory-efficient

## Exports
StreamingExecutor, ResultStream, StreamConfig

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#streaming.go> a code:Module ;
    code:name "pkg/query/streaming.go" ;
    code:description "Streaming query executor for large result sets" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./executor.go>, <./query.go> ;
    code:exports <#StreamingExecutor>, <#ResultStream>, <#StreamConfig> ;
    code:tags "query", "streaming", "pagination", "memory-efficient" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"fmt"
	"sync"

	"github.com/justin4957/graphfs/internal/store"
)

// StreamConfig holds configuration for streaming query execution
type StreamConfig struct {
	PageSize       int  // Number of results per page (default: 100)
	BufferSize     int  // Channel buffer size (default: PageSize)
	ReportProgress bool // Whether to report progress during streaming
}

// DefaultStreamConfig returns a default streaming configuration
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		PageSize:       100,
		BufferSize:     100,
		ReportProgress: false,
	}
}

// StreamingExecutor executes queries with streaming support
type StreamingExecutor struct {
	executor *Executor
	config   *StreamConfig
}

// NewStreamingExecutor creates a new streaming executor
func NewStreamingExecutor(tripleStore *store.TripleStore, config *StreamConfig) *StreamingExecutor {
	if config == nil {
		config = DefaultStreamConfig()
	}
	if config.PageSize <= 0 {
		config.PageSize = 100
	}
	if config.BufferSize <= 0 {
		config.BufferSize = config.PageSize
	}

	return &StreamingExecutor{
		executor: NewExecutor(tripleStore),
		config:   config,
	}
}

// ResultStream represents a streaming result set
type ResultStream struct {
	Results    chan map[string]string // Channel for streaming individual bindings
	Errors     chan error             // Channel for errors
	Done       chan struct{}          // Signal when streaming is complete
	Variables  []string               // Variable names for the result set
	TotalCount int                    // Total count of results (populated after streaming completes)

	mu     sync.Mutex
	count  int  // Current count of streamed results
	closed bool // Whether the stream has been closed
}

// NewResultStream creates a new result stream
func NewResultStream(bufferSize int, variables []string) *ResultStream {
	return &ResultStream{
		Results:   make(chan map[string]string, bufferSize),
		Errors:    make(chan error, 1),
		Done:      make(chan struct{}),
		Variables: variables,
	}
}

// Close closes the result stream
func (s *ResultStream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true
		close(s.Results)
		close(s.Errors)
		close(s.Done)
	}
}

// Count returns the current count of streamed results
func (s *ResultStream) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// incrementCount safely increments the result count
func (s *ResultStream) incrementCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
}

// ProgressCallback is called to report streaming progress
type ProgressCallback func(current, total int)

// ExecuteStream executes a query and streams results
func (e *StreamingExecutor) ExecuteStream(query *Query) *ResultStream {
	return e.ExecuteStreamWithProgress(query, nil)
}

// ExecuteStreamWithProgress executes a query with progress reporting
func (e *StreamingExecutor) ExecuteStreamWithProgress(query *Query, progressCallback ProgressCallback) *ResultStream {
	// First, execute the full query to get results
	// In a real implementation with a proper database, this would use cursors
	result, err := e.executor.Execute(query)

	// Determine variables for the stream
	var variables []string
	if result != nil {
		variables = result.Variables
	}

	stream := NewResultStream(e.config.BufferSize, variables)

	go func() {
		defer stream.Close()

		if err != nil {
			stream.Errors <- err
			return
		}

		if result == nil || len(result.Bindings) == 0 {
			stream.TotalCount = 0
			return
		}

		totalResults := len(result.Bindings)
		stream.TotalCount = totalResults

		// Stream results in batches
		for i, binding := range result.Bindings {
			select {
			case stream.Results <- binding:
				stream.incrementCount()

				// Report progress if callback is provided
				if progressCallback != nil && e.config.ReportProgress {
					progressCallback(i+1, totalResults)
				}
			case <-stream.Done:
				// Stream was closed early
				return
			}
		}
	}()

	return stream
}

// ExecuteStringStream parses and executes a query string with streaming
func (e *StreamingExecutor) ExecuteStringStream(queryStr string) (*ResultStream, error) {
	query, err := ParseQuery(queryStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return e.ExecuteStream(query), nil
}

// ForEach iterates over all results in the stream
func (s *ResultStream) ForEach(fn func(binding map[string]string) error) error {
	// Drain all results from the Results channel
	// The Results channel is closed when streaming completes
	for binding := range s.Results {
		if err := fn(binding); err != nil {
			return err
		}
	}

	// After Results channel is closed, check for any errors
	select {
	case err := <-s.Errors:
		if err != nil {
			return err
		}
	default:
		// No error
	}

	return nil
}

// Collect gathers all results from the stream into a slice
// Use with caution for large result sets as this defeats the purpose of streaming
func (s *ResultStream) Collect() ([]map[string]string, error) {
	var results []map[string]string

	err := s.ForEach(func(binding map[string]string) error {
		results = append(results, binding)
		return nil
	})

	return results, err
}

// CollectPage collects up to `limit` results from the stream
func (s *ResultStream) CollectPage(limit int) ([]map[string]string, error) {
	var results []map[string]string
	count := 0

	for {
		if count >= limit {
			return results, nil
		}

		select {
		case binding, ok := <-s.Results:
			if !ok {
				// Channel closed, check for errors
				select {
				case err := <-s.Errors:
					if err != nil {
						return results, err
					}
				default:
				}
				return results, nil
			}
			results = append(results, binding)
			count++

		case err := <-s.Errors:
			return results, err
		}
	}
}

// PaginatedResult represents a page of results
type PaginatedResult struct {
	Bindings   []map[string]string // Results for this page
	Variables  []string            // Variable names
	Page       int                 // Current page number (1-indexed)
	PageSize   int                 // Size of each page
	TotalCount int                 // Total number of results
	TotalPages int                 // Total number of pages
	HasMore    bool                // Whether there are more pages
}

// ExecutePaginated executes a query with pagination
func (e *StreamingExecutor) ExecutePaginated(query *Query, page, pageSize int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = e.config.PageSize
	}

	// Clone the query and apply pagination
	paginatedQuery := query.WithPagination(pageSize, (page-1)*pageSize)

	// Execute the paginated query
	result, err := e.executor.Execute(paginatedQuery)
	if err != nil {
		return nil, err
	}

	// Get total count by executing without limit/offset
	countQuery := query.WithPagination(0, 0)
	countResult, err := e.executor.Execute(countQuery)
	if err != nil {
		return nil, err
	}

	totalCount := len(countResult.Bindings)
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginatedResult{
		Bindings:   result.Bindings,
		Variables:  result.Variables,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}, nil
}

// ExecuteStringPaginated parses and executes a query string with pagination
func (e *StreamingExecutor) ExecuteStringPaginated(queryStr string, page, pageSize int) (*PaginatedResult, error) {
	query, err := ParseQuery(queryStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return e.ExecutePaginated(query, page, pageSize)
}

// WithPagination creates a copy of the query with pagination applied
func (q *Query) WithPagination(limit, offset int) *Query {
	newQuery := *q
	if newQuery.Select != nil {
		newSelect := *newQuery.Select
		if limit > 0 {
			newSelect.Limit = limit
		}
		if offset > 0 {
			newSelect.Offset = offset
		}
		newQuery.Select = &newSelect
	}
	return &newQuery
}
