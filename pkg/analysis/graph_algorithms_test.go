package analysis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/justin4957/graphfs/pkg/graph"
)

// Helper function to create test graphs
func createTestGraph() *graph.Graph {
	g := &graph.Graph{
		Modules: make(map[string]*graph.Module),
	}

	// Create a simple DAG:
	// A -> B -> D
	// A -> C -> D
	// E (isolated)

	g.Modules["A"] = &graph.Module{
		Path:         "A",
		Dependencies: []string{"B", "C"},
	}

	g.Modules["B"] = &graph.Module{
		Path:         "B",
		Dependencies: []string{"D"},
	}

	g.Modules["C"] = &graph.Module{
		Path:         "C",
		Dependencies: []string{"D"},
	}

	g.Modules["D"] = &graph.Module{
		Path:         "D",
		Dependencies: []string{},
	}

	g.Modules["E"] = &graph.Module{
		Path:         "E",
		Dependencies: []string{},
	}

	return g
}

func createCyclicGraph() *graph.Graph {
	g := &graph.Graph{
		Modules: make(map[string]*graph.Module),
	}

	// Create a graph with a cycle:
	// A -> B -> C -> A
	// D (isolated)

	g.Modules["A"] = &graph.Module{
		Path:         "A",
		Dependencies: []string{"B"},
	}

	g.Modules["B"] = &graph.Module{
		Path:         "B",
		Dependencies: []string{"C"},
	}

	g.Modules["C"] = &graph.Module{
		Path:         "C",
		Dependencies: []string{"A"},
	}

	g.Modules["D"] = &graph.Module{
		Path:         "D",
		Dependencies: []string{},
	}

	return g
}

func TestTopologicalSort(t *testing.T) {
	t.Run("valid DAG", func(t *testing.T) {
		g := createTestGraph()
		result, err := TopologicalSort(g)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 5 {
			t.Fatalf("Expected 5 modules, got %d", len(result))
		}

		// Verify ordering: dependencies must come before dependents
		positions := make(map[string]int)
		for i, module := range result {
			positions[module.Path] = i
		}

		// D must come before B and C
		if positions["D"] >= positions["B"] || positions["D"] >= positions["C"] {
			t.Error("D should come before B and C")
		}

		// B and C must come before A
		if positions["B"] >= positions["A"] || positions["C"] >= positions["A"] {
			t.Error("B and C should come before A")
		}
	})

	t.Run("cyclic graph", func(t *testing.T) {
		g := createCyclicGraph()
		_, err := TopologicalSort(g)

		if err == nil {
			t.Fatal("Expected error for cyclic graph, got nil")
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		g := &graph.Graph{
			Modules: make(map[string]*graph.Module),
		}

		result, err := TopologicalSort(g)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 0 {
			t.Fatalf("Expected 0 modules, got %d", len(result))
		}
	})
}

func TestShortestPath(t *testing.T) {
	g := createTestGraph()

	tests := []struct {
		name     string
		from     string
		to       string
		expected []string
	}{
		{
			name:     "direct dependency",
			from:     "A",
			to:       "B",
			expected: []string{"A", "B"},
		},
		{
			name:     "transitive dependency",
			from:     "A",
			to:       "D",
			expected: []string{"A", "B", "D"}, // or A -> C -> D
		},
		{
			name:     "same module",
			from:     "A",
			to:       "A",
			expected: []string{"A"},
		},
		{
			name:     "no path",
			from:     "D",
			to:       "A",
			expected: nil,
		},
		{
			name:     "isolated module",
			from:     "E",
			to:       "A",
			expected: nil,
		},
		{
			name:     "non-existent source",
			from:     "X",
			to:       "A",
			expected: nil,
		},
		{
			name:     "non-existent target",
			from:     "A",
			to:       "X",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortestPath(g, tt.from, tt.to)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				// For A -> D, both A -> B -> D and A -> C -> D are valid
				if tt.from == "A" && tt.to == "D" {
					if len(result) != 3 || result[0] != "A" || result[2] != "D" {
						t.Errorf("Expected path of length 3 from A to D, got %v", result)
					}
				} else {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestStronglyConnectedComponents(t *testing.T) {
	t.Run("DAG with no cycles", func(t *testing.T) {
		g := createTestGraph()
		sccs := StronglyConnectedComponents(g)

		// In a DAG, each node is its own SCC
		if len(sccs) != 5 {
			t.Fatalf("Expected 5 SCCs, got %d", len(sccs))
		}

		for _, scc := range sccs {
			if len(scc) != 1 {
				t.Errorf("Expected singleton SCC, got %v", scc)
			}
		}
	})

	t.Run("graph with cycle", func(t *testing.T) {
		g := createCyclicGraph()
		sccs := StronglyConnectedComponents(g)

		// Should have one SCC with {A, B, C} and one with {D}
		if len(sccs) != 2 {
			t.Fatalf("Expected 2 SCCs, got %d", len(sccs))
		}

		// Find the cycle SCC
		var cycleSCC []string
		for _, scc := range sccs {
			if len(scc) == 3 {
				cycleSCC = scc
				break
			}
		}

		if cycleSCC == nil {
			t.Fatal("Expected to find SCC with 3 nodes")
		}

		expected := []string{"A", "B", "C"}
		sort.Strings(cycleSCC)
		if !reflect.DeepEqual(cycleSCC, expected) {
			t.Errorf("Expected cycle SCC %v, got %v", expected, cycleSCC)
		}
	})
}

func TestTransitiveDependencies(t *testing.T) {
	g := createTestGraph()

	tests := []struct {
		name     string
		module   string
		expected map[string]int
	}{
		{
			name:   "module A",
			module: "A",
			expected: map[string]int{
				"B": 1,
				"C": 1,
				"D": 2,
			},
		},
		{
			name:   "module B",
			module: "B",
			expected: map[string]int{
				"D": 1,
			},
		},
		{
			name:     "module D (leaf)",
			module:   "D",
			expected: map[string]int{},
		},
		{
			name:     "module E (isolated)",
			module:   "E",
			expected: map[string]int{},
		},
		{
			name:     "non-existent module",
			module:   "X",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransitiveDependencies(g, tt.module)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTransitiveDependents(t *testing.T) {
	g := createTestGraph()

	tests := []struct {
		name     string
		module   string
		expected map[string]int
	}{
		{
			name:   "module D (depended on by all)",
			module: "D",
			expected: map[string]int{
				"B": 1,
				"C": 1,
				"A": 2,
			},
		},
		{
			name:   "module B",
			module: "B",
			expected: map[string]int{
				"A": 1,
			},
		},
		{
			name:     "module A (root)",
			module:   "A",
			expected: map[string]int{},
		},
		{
			name:     "module E (isolated)",
			module:   "E",
			expected: map[string]int{},
		},
		{
			name:     "non-existent module",
			module:   "X",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransitiveDependents(g, tt.module)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCyclicDependencies(t *testing.T) {
	t.Run("DAG with no cycles", func(t *testing.T) {
		g := createTestGraph()
		cycles := CyclicDependencies(g)

		if len(cycles) != 0 {
			t.Errorf("Expected 0 cycles, got %d: %v", len(cycles), cycles)
		}
	})

	t.Run("graph with cycle", func(t *testing.T) {
		g := createCyclicGraph()
		cycles := CyclicDependencies(g)

		if len(cycles) != 1 {
			t.Fatalf("Expected 1 cycle, got %d", len(cycles))
		}

		cycle := cycles[0]
		sort.Strings(cycle)
		expected := []string{"A", "B", "C"}

		if !reflect.DeepEqual(cycle, expected) {
			t.Errorf("Expected cycle %v, got %v", expected, cycle)
		}
	})

	t.Run("self-loop", func(t *testing.T) {
		g := &graph.Graph{
			Modules: make(map[string]*graph.Module),
		}

		g.Modules["A"] = &graph.Module{
			Path:         "A",
			Dependencies: []string{"A"}, // Self-loop
		}

		cycles := CyclicDependencies(g)

		if len(cycles) != 1 {
			t.Fatalf("Expected 1 cycle (self-loop), got %d", len(cycles))
		}

		if len(cycles[0]) != 1 || cycles[0][0] != "A" {
			t.Errorf("Expected self-loop cycle [A], got %v", cycles[0])
		}
	})
}

func TestDependencyDepth(t *testing.T) {
	g := createTestGraph()

	tests := []struct {
		name     string
		module   string
		expected int
	}{
		{
			name:     "module A (deepest)",
			module:   "A",
			expected: 2,
		},
		{
			name:     "module B",
			module:   "B",
			expected: 1,
		},
		{
			name:     "module C",
			module:   "C",
			expected: 1,
		},
		{
			name:     "module D (leaf)",
			module:   "D",
			expected: 0,
		},
		{
			name:     "module E (isolated)",
			module:   "E",
			expected: 0,
		},
		{
			name:     "non-existent module",
			module:   "X",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DependencyDepth(g, tt.module)

			if result != tt.expected {
				t.Errorf("Expected depth %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestComplexGraph(t *testing.T) {
	// Create a more complex graph for integration testing
	g := &graph.Graph{
		Modules: make(map[string]*graph.Module),
	}

	// Diamond dependency pattern:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	//     |
	//     E

	g.Modules["A"] = &graph.Module{
		Path:         "A",
		Dependencies: []string{"B", "C"},
	}

	g.Modules["B"] = &graph.Module{
		Path:         "B",
		Dependencies: []string{"D"},
	}

	g.Modules["C"] = &graph.Module{
		Path:         "C",
		Dependencies: []string{"D"},
	}

	g.Modules["D"] = &graph.Module{
		Path:         "D",
		Dependencies: []string{"E"},
	}

	g.Modules["E"] = &graph.Module{
		Path:         "E",
		Dependencies: []string{},
	}

	t.Run("topological sort", func(t *testing.T) {
		result, err := TopologicalSort(g)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 5 {
			t.Fatalf("Expected 5 modules, got %d", len(result))
		}

		// Verify valid ordering
		positions := make(map[string]int)
		for i, module := range result {
			positions[module.Path] = i
		}

		if positions["E"] >= positions["D"] {
			t.Error("E should come before D")
		}
		if positions["D"] >= positions["B"] || positions["D"] >= positions["C"] {
			t.Error("D should come before B and C")
		}
		if positions["B"] >= positions["A"] || positions["C"] >= positions["A"] {
			t.Error("B and C should come before A")
		}
	})

	t.Run("transitive dependencies of A", func(t *testing.T) {
		deps := TransitiveDependencies(g, "A")

		expected := map[string]int{
			"B": 1,
			"C": 1,
			"D": 2,
			"E": 3,
		}

		if !reflect.DeepEqual(deps, expected) {
			t.Errorf("Expected %v, got %v", expected, deps)
		}
	})

	t.Run("dependency depth", func(t *testing.T) {
		depth := DependencyDepth(g, "A")
		if depth != 3 {
			t.Errorf("Expected depth 3, got %d", depth)
		}
	})
}
