package graph_test

import (
	"fmt"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
)

func Example_buildGraph() {
	// Build a knowledge graph from a codebase
	builder := graph.NewBuilder()

	minimalAppPath := "../../examples/minimal-app"
	absPath, _ := filepath.Abs(minimalAppPath)

	g, _ := builder.Build(absPath, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		Validate:       false,
		ReportProgress: false,
	})

	fmt.Printf("Built graph with %d modules\n", g.Statistics.TotalModules)
	// Output: Built graph with 7 modules
}

func Example_queryModules() {
	builder := graph.NewBuilder()

	minimalAppPath := "../../examples/minimal-app"
	absPath, _ := filepath.Abs(minimalAppPath)

	g, _ := builder.Build(absPath, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		ReportProgress: false,
	})

	// Get modules by language
	goModules := g.GetModulesByLanguage("go")
	fmt.Printf("Found %d Go modules\n", len(goModules))

	// Output: Found 7 Go modules
}

func Example_moduleMetadata() {
	builder := graph.NewBuilder()

	minimalAppPath := "../../examples/minimal-app"
	absPath, _ := filepath.Abs(minimalAppPath)

	g, _ := builder.Build(absPath, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		ReportProgress: false,
	})

	// Find main.go module
	var mainModule *graph.Module
	for _, module := range g.Modules {
		if module.Name == "main.go" {
			mainModule = module
			break
		}
	}

	if mainModule != nil {
		fmt.Printf("Module: %s\n", mainModule.Name)
		fmt.Printf("Language: %s\n", mainModule.Language)
		fmt.Printf("Dependencies: %d\n", len(mainModule.Dependencies))
	}

	// Output:
	// Module: main.go
	// Language: go
	// Dependencies: 4
}

func Example_validation() {
	builder := graph.NewBuilder()

	minimalAppPath := "../../examples/minimal-app"
	absPath, _ := filepath.Abs(minimalAppPath)

	g, _ := builder.Build(absPath, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
		},
		ReportProgress: false,
	})

	// Validate the graph
	validator := graph.NewValidator()
	result := validator.Validate(g)

	fmt.Printf("Validation: %d errors, %d warnings\n",
		len(result.Errors), len(result.Warnings))

	// Output: Validation: 2 errors, 11 warnings
}
