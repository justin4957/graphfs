/*
Benchmarks for parallel vs sequential graph building.
*/

package graph

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/justin4957/graphfs/pkg/scanner"
)

// BenchmarkBuild_Sequential benchmarks sequential graph building
func BenchmarkBuild_Sequential(b *testing.B) {
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     1, // Sequential
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false, // Disable cache for fair comparison
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_Parallel2 benchmarks parallel graph building with 2 workers
func BenchmarkBuild_Parallel2(b *testing.B) {
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     2,
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_Parallel4 benchmarks parallel graph building with 4 workers
func BenchmarkBuild_Parallel4(b *testing.B) {
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     4,
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_Parallel8 benchmarks parallel graph building with 8 workers
func BenchmarkBuild_Parallel8(b *testing.B) {
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     8,
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_ParallelDefault benchmarks parallel graph building with NumCPU workers
func BenchmarkBuild_ParallelDefault(b *testing.B) {
	minimalAppPath := "../../examples/minimal-app"
	absPath, err := filepath.Abs(minimalAppPath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	numCPU := runtime.NumCPU()
	b.Logf("Using %d workers (NumCPU)", numCPU)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     0, // 0 = NumCPU
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_LargeCodebase benchmarks on the graphfs codebase itself
func BenchmarkBuild_LargeCodebase_Sequential(b *testing.B) {
	// Use current directory (graphfs codebase)
	absPath, err := filepath.Abs("../..")
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     1,
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

// BenchmarkBuild_LargeCodebase_Parallel benchmarks parallel on graphfs codebase
func BenchmarkBuild_LargeCodebase_Parallel(b *testing.B) {
	absPath, err := filepath.Abs("../..")
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.Build(absPath, BuildOptions{
			ScanOptions: scanner.ScanOptions{
				UseDefaults: true,
				Concurrent:  true,
				Workers:     0, // NumCPU
			},
			Validate:       false,
			ReportProgress: false,
			UseCache:       false,
		})
		if err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}
