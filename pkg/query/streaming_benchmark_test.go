/*
# Module: pkg/query/streaming_benchmark_test.go
Benchmarks for streaming query executor.

Demonstrates memory efficiency of streaming vs. non-streaming execution.

## Linked Modules
- [streaming](./streaming.go) - Streaming executor implementation

## Tags
benchmark, query, streaming, memory

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#streaming_benchmark_test.go> a code:Module ;
    code:name "pkg/query/streaming_benchmark_test.go" ;
    code:description "Benchmarks for streaming query executor" ;
    code:language "go" ;
    code:layer "query" ;
    code:linksTo <./streaming.go> ;
    code:tags "benchmark", "query", "streaming", "memory" .
<!-- End LinkedDoc RDF -->
*/

package query

import (
	"runtime"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

// createLargeTestStore creates a triple store with many triples for benchmarking
func createLargeTestStore(numTriples int) *store.TripleStore {
	tripleStore := store.NewTripleStore()

	for i := 0; i < numTriples; i++ {
		tripleStore.Add(
			"module"+string(rune('A'+i%26))+string(rune('0'+i%10)),
			"imports",
			"module"+string(rune('A'+(i+1)%26))+string(rune('0'+(i+1)%10)),
		)
	}

	return tripleStore
}

func BenchmarkNormalExecution(b *testing.B) {
	tripleStore := createLargeTestStore(10000)
	executor := NewExecutor(tripleStore)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(query)
		if err != nil {
			b.Fatalf("query failed: %v", err)
		}
	}
}

func BenchmarkStreamingExecution(b *testing.B) {
	tripleStore := createLargeTestStore(10000)
	config := &StreamConfig{PageSize: 100}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := executor.ExecuteStream(query)
		count := 0
		err := stream.ForEach(func(binding map[string]string) error {
			count++
			return nil
		})
		if err != nil {
			b.Fatalf("streaming failed: %v", err)
		}
	}
}

func BenchmarkPaginatedExecution(b *testing.B) {
	tripleStore := createLargeTestStore(10000)
	config := &StreamConfig{PageSize: 100}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fetch all pages
		page := 1
		for {
			result, err := executor.ExecutePaginated(query, page, 100)
			if err != nil {
				b.Fatalf("paginated query failed: %v", err)
			}
			if !result.HasMore {
				break
			}
			page++
		}
	}
}

// BenchmarkMemoryUsageNormal measures memory usage for normal execution
func BenchmarkMemoryUsageNormal(b *testing.B) {
	tripleStore := createLargeTestStore(50000)
	executor := NewExecutor(tripleStore)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := executor.Execute(query)
		if err != nil {
			b.Fatalf("query failed: %v", err)
		}
		// Keep result to prevent optimization
		_ = len(result.Bindings)
	}

	runtime.ReadMemStats(&memAfter)
	b.ReportMetric(float64(memAfter.Alloc-memBefore.Alloc)/float64(b.N), "bytes/op")
}

// BenchmarkMemoryUsageStreaming measures memory usage for streaming execution
func BenchmarkMemoryUsageStreaming(b *testing.B) {
	tripleStore := createLargeTestStore(50000)
	config := &StreamConfig{PageSize: 100, BufferSize: 100}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := executor.ExecuteStream(query)
		count := 0
		err := stream.ForEach(func(binding map[string]string) error {
			count++
			// Process binding immediately without storing
			return nil
		})
		if err != nil {
			b.Fatalf("streaming failed: %v", err)
		}
	}

	runtime.ReadMemStats(&memAfter)
	b.ReportMetric(float64(memAfter.Alloc-memBefore.Alloc)/float64(b.N), "bytes/op")
}

// BenchmarkDifferentPageSizes benchmarks various page sizes
func BenchmarkDifferentPageSizes(b *testing.B) {
	tripleStore := createLargeTestStore(10000)

	pageSizes := []int{10, 50, 100, 500, 1000}

	for _, pageSize := range pageSizes {
		b.Run("PageSize"+string(rune('0'+pageSize/100)), func(b *testing.B) {
			config := &StreamConfig{PageSize: pageSize}
			executor := NewStreamingExecutor(tripleStore, config)

			query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
			if err != nil {
				b.Fatalf("failed to parse query: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stream := executor.ExecuteStream(query)
				_, _ = stream.Collect()
			}
		})
	}
}

// BenchmarkCollectPage benchmarks collecting pages of different sizes
func BenchmarkCollectPage(b *testing.B) {
	tripleStore := createLargeTestStore(10000)
	config := &StreamConfig{PageSize: 1000}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	b.Run("CollectPage10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := executor.ExecuteStream(query)
			_, _ = stream.CollectPage(10)
		}
	})

	b.Run("CollectPage100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := executor.ExecuteStream(query)
			_, _ = stream.CollectPage(100)
		}
	})

	b.Run("CollectPage1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream := executor.ExecuteStream(query)
			_, _ = stream.CollectPage(1000)
		}
	})
}

// BenchmarkStreamingThroughput measures results processed per second
func BenchmarkStreamingThroughput(b *testing.B) {
	tripleStore := createLargeTestStore(100000)
	config := &StreamConfig{PageSize: 1000, BufferSize: 1000}
	executor := NewStreamingExecutor(tripleStore, config)

	query, err := ParseQuery("SELECT ?s ?o WHERE { ?s <imports> ?o }")
	if err != nil {
		b.Fatalf("failed to parse query: %v", err)
	}

	b.ResetTimer()
	totalResults := 0
	for i := 0; i < b.N; i++ {
		stream := executor.ExecuteStream(query)
		count := 0
		_ = stream.ForEach(func(binding map[string]string) error {
			count++
			return nil
		})
		totalResults += count
	}

	b.ReportMetric(float64(totalResults)/b.Elapsed().Seconds(), "results/sec")
}
