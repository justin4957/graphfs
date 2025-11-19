package diff

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func TestDiffer_compareModules(t *testing.T) {
	differ := NewDiffer(".")

	tests := []struct {
		name    string
		old     *graph.Module
		new     *graph.Module
		wantNil bool
		checkFn func(*testing.T, *ModuleChange)
	}{
		{
			name: "No changes",
			old: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go"},
				Exports:      []string{"Func1"},
				Layer:        "service",
			},
			new: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go"},
				Exports:      []string{"Func1"},
				Layer:        "service",
			},
			wantNil: true,
		},
		{
			name: "Dependency added",
			old: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go"},
			},
			new: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go", "dep2.go"},
			},
			wantNil: false,
			checkFn: func(t *testing.T, c *ModuleChange) {
				if len(c.DepsAdded) != 1 || c.DepsAdded[0] != "dep2.go" {
					t.Errorf("Expected dep2.go to be added, got %v", c.DepsAdded)
				}
			},
		},
		{
			name: "Dependency removed",
			old: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go", "dep2.go"},
			},
			new: &graph.Module{
				Path:         "test.go",
				Dependencies: []string{"dep1.go"},
			},
			wantNil: false,
			checkFn: func(t *testing.T, c *ModuleChange) {
				if len(c.DepsRemoved) != 1 || c.DepsRemoved[0] != "dep2.go" {
					t.Errorf("Expected dep2.go to be removed, got %v", c.DepsRemoved)
				}
			},
		},
		{
			name: "Export added",
			old: &graph.Module{
				Path:    "test.go",
				Exports: []string{"Func1"},
			},
			new: &graph.Module{
				Path:    "test.go",
				Exports: []string{"Func1", "Func2"},
			},
			wantNil: false,
			checkFn: func(t *testing.T, c *ModuleChange) {
				if len(c.ExportsAdded) != 1 || c.ExportsAdded[0] != "Func2" {
					t.Errorf("Expected Func2 to be added, got %v", c.ExportsAdded)
				}
			},
		},
		{
			name: "Layer changed",
			old: &graph.Module{
				Path:  "test.go",
				Layer: "service",
			},
			new: &graph.Module{
				Path:  "test.go",
				Layer: "api",
			},
			wantNil: false,
			checkFn: func(t *testing.T, c *ModuleChange) {
				if !c.LayerChanged {
					t.Error("Expected layer to be marked as changed")
				}
				if c.OldLayer != "service" || c.NewLayer != "api" {
					t.Errorf("Expected layer change from service to api, got %s -> %s", c.OldLayer, c.NewLayer)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := differ.compareModules(tt.old, tt.new)

			if tt.wantNil && change != nil {
				t.Errorf("Expected no change, got %+v", change)
			}

			if !tt.wantNil && change == nil {
				t.Error("Expected change, got nil")
			}

			if change != nil && tt.checkFn != nil {
				tt.checkFn(t, change)
			}
		})
	}
}

func TestDiffer_compareGraphs(t *testing.T) {
	differ := NewDiffer(".")

	// Create old graph
	oldStore := store.NewTripleStore()
	oldGraph := graph.NewGraph(".", oldStore)
	oldGraph.AddModule(&graph.Module{Path: "module1.go", Layer: "service"})
	oldGraph.AddModule(&graph.Module{Path: "module2.go", Layer: "api"})
	oldGraph.AddModule(&graph.Module{Path: "module3.go", Layer: "data"})

	// Create new graph
	newStore := store.NewTripleStore()
	newGraph := graph.NewGraph(".", newStore)
	newGraph.AddModule(&graph.Module{Path: "module1.go", Layer: "service"})  // Unchanged
	newGraph.AddModule(&graph.Module{Path: "module2.go", Layer: "security"}) // Modified (layer changed)
	newGraph.AddModule(&graph.Module{Path: "module4.go", Layer: "service"})  // Added
	// module3.go removed

	diff := differ.compareGraphs(oldGraph, newGraph)

	// Check stats
	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 added module, got %d", len(diff.Added))
	}
	if len(diff.Removed) != 1 {
		t.Errorf("Expected 1 removed module, got %d", len(diff.Removed))
	}
	if len(diff.Modified) != 1 {
		t.Errorf("Expected 1 modified module, got %d", len(diff.Modified))
	}
	if len(diff.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged module, got %d", len(diff.Unchanged))
	}

	// Check specific modules
	if len(diff.Added) > 0 && diff.Added[0].Path != "module4.go" {
		t.Errorf("Expected module4.go to be added, got %s", diff.Added[0].Path)
	}
	if len(diff.Removed) > 0 && diff.Removed[0].Path != "module3.go" {
		t.Errorf("Expected module3.go to be removed, got %s", diff.Removed[0].Path)
	}
	if len(diff.Modified) > 0 && diff.Modified[0].Module.Path != "module2.go" {
		t.Errorf("Expected module2.go to be modified, got %s", diff.Modified[0].Module.Path)
	}
}

func TestGraphDiff_Stats(t *testing.T) {
	diff := &GraphDiff{
		Added:     []*graph.Module{{Path: "new1.go"}, {Path: "new2.go"}},
		Removed:   []*graph.Module{{Path: "old1.go"}},
		Modified:  []*ModuleChange{{Module: &graph.Module{Path: "changed.go"}}},
		Unchanged: []*graph.Module{{Path: "same1.go"}, {Path: "same2.go"}, {Path: "same3.go"}},
	}

	stats := diff.Stats()

	if stats.Added != 2 {
		t.Errorf("Expected 2 added, got %d", stats.Added)
	}
	if stats.Removed != 1 {
		t.Errorf("Expected 1 removed, got %d", stats.Removed)
	}
	if stats.Modified != 1 {
		t.Errorf("Expected 1 modified, got %d", stats.Modified)
	}
	if stats.Unchanged != 3 {
		t.Errorf("Expected 3 unchanged, got %d", stats.Unchanged)
	}
	if stats.Total != 7 {
		t.Errorf("Expected total 7, got %d", stats.Total)
	}
}

func TestToSet(t *testing.T) {
	items := []string{"a", "b", "c"}
	set := toSet(items)

	if len(set) != 3 {
		t.Errorf("Expected set size 3, got %d", len(set))
	}

	if !set["a"] || !set["b"] || !set["c"] {
		t.Error("Set missing expected items")
	}

	if set["d"] {
		t.Error("Set contains unexpected item")
	}
}
