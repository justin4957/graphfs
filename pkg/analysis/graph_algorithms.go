/*
# Module: pkg/analysis/graph_algorithms.go
Dependency graph algorithms for GraphFS analysis.

Provides fundamental graph algorithms for analyzing module dependencies including
topological sorting, shortest paths, strongly connected components, and transitive
dependency calculations.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [./types](./types.go) - Analysis result types

## Tags
analysis, graph-algorithms, dependencies, topology

## Exports
TopologicalSort, ShortestPath, StronglyConnectedComponents, TransitiveDependencies

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#graph_algorithms.go> a code:Module ;
    code:name "pkg/analysis/graph_algorithms.go" ;
    code:description "Dependency graph algorithms for GraphFS analysis" ;
    code:language "go" ;
    code:layer "analysis" ;
    code:linksTo <../graph/graph.go>, <./types.go> ;
    code:exports <#TopologicalSort>, <#ShortestPath>, <#StronglyConnectedComponents>, <#TransitiveDependencies> ;
    code:tags "analysis", "graph-algorithms", "dependencies", "topology" .
<!-- End LinkedDoc RDF -->
*/

package analysis

import (
	"fmt"
	"sort"

	"github.com/justin4957/graphfs/pkg/graph"
)

// TopologicalSort performs a topological sort on the module dependency graph.
// Returns modules in an order where dependencies come before dependents.
// Returns an error if the graph contains cycles.
func TopologicalSort(g *graph.Graph) ([]*graph.Module, error) {
	// Build adjacency list and in-degree map
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize all modules with in-degree 0
	for path := range g.Modules {
		inDegree[path] = 0
		adjList[path] = []string{}
	}

	// Build graph structure
	for path, module := range g.Modules {
		for _, dep := range module.Dependencies {
			adjList[dep] = append(adjList[dep], path)
			inDegree[path]++
		}
	}

	// Kahn's algorithm for topological sort
	queue := []string{}
	for path, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, path)
		}
	}

	// Sort queue for deterministic ordering
	sort.Strings(queue)

	result := []*graph.Module{}
	visited := 0

	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]

		if module, exists := g.Modules[current]; exists {
			result = append(result, module)
			visited++
		}

		// Process neighbors
		neighbors := adjList[current]
		sort.Strings(neighbors) // Deterministic ordering

		for _, neighbor := range neighbors {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if visited != len(g.Modules) {
		return nil, fmt.Errorf("graph contains cycles, cannot perform topological sort")
	}

	return result, nil
}

// ShortestPath finds the shortest dependency path between two modules using BFS.
// Returns the path as a slice of module paths, or nil if no path exists.
func ShortestPath(g *graph.Graph, fromPath, toPath string) []string {
	// Check if modules exist
	if _, exists := g.Modules[fromPath]; !exists {
		return nil
	}
	if _, exists := g.Modules[toPath]; !exists {
		return nil
	}

	// Special case: same module
	if fromPath == toPath {
		return []string{fromPath}
	}

	// BFS to find shortest path
	queue := [][]string{{fromPath}}
	visited := make(map[string]bool)
	visited[fromPath] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]
		module := g.Modules[current]

		// Check all dependencies
		for _, dep := range module.Dependencies {
			if visited[dep] {
				continue
			}

			newPath := append([]string{}, path...)
			newPath = append(newPath, dep)

			if dep == toPath {
				return newPath
			}

			visited[dep] = true
			queue = append(queue, newPath)
		}
	}

	return nil // No path found
}

// StronglyConnectedComponents finds all strongly connected components (SCCs) in the dependency graph.
// An SCC is a maximal set of modules where each module is reachable from every other module.
// Returns a slice of SCCs, where each SCC is a slice of module paths.
func StronglyConnectedComponents(g *graph.Graph) [][]string {
	// Tarjan's algorithm for finding SCCs
	index := 0
	stack := []string{}
	indices := make(map[string]int)
	lowLinks := make(map[string]int)
	onStack := make(map[string]bool)
	sccs := [][]string{}

	var strongConnect func(string)
	strongConnect = func(v string) {
		indices[v] = index
		lowLinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		// Consider successors (dependencies)
		if module, exists := g.Modules[v]; exists {
			for _, w := range module.Dependencies {
				if _, visited := indices[w]; !visited {
					strongConnect(w)
					if lowLinks[w] < lowLinks[v] {
						lowLinks[v] = lowLinks[w]
					}
				} else if onStack[w] {
					if indices[w] < lowLinks[v] {
						lowLinks[v] = indices[w]
					}
				}
			}
		}

		// If v is a root node, pop the stack and create an SCC
		if lowLinks[v] == indices[v] {
			scc := []string{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			// Sort for deterministic output
			sort.Strings(scc)
			sccs = append(sccs, scc)
		}
	}

	// Process all modules
	modules := []string{}
	for path := range g.Modules {
		modules = append(modules, path)
	}
	sort.Strings(modules) // Deterministic ordering

	for _, path := range modules {
		if _, visited := indices[path]; !visited {
			strongConnect(path)
		}
	}

	return sccs
}

// TransitiveDependencies calculates all transitive dependencies for a given module.
// Returns a map where keys are module paths and values are their depth from the root.
func TransitiveDependencies(g *graph.Graph, modulePath string) map[string]int {
	// Check if module exists
	if _, exists := g.Modules[modulePath]; !exists {
		return nil
	}

	// BFS to find all transitive dependencies
	deps := make(map[string]int)
	queue := []struct {
		path  string
		depth int
	}{{modulePath, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Skip if already visited at a shallower depth
		if existingDepth, visited := deps[current.path]; visited && existingDepth <= current.depth {
			continue
		}

		deps[current.path] = current.depth

		// Add dependencies to queue
		if module, exists := g.Modules[current.path]; exists {
			for _, dep := range module.Dependencies {
				queue = append(queue, struct {
					path  string
					depth int
				}{dep, current.depth + 1})
			}
		}
	}

	// Remove the root module itself
	delete(deps, modulePath)

	return deps
}

// TransitiveDependents calculates all modules that transitively depend on a given module.
// Returns a map where keys are dependent module paths and values are their depth from the target.
func TransitiveDependents(g *graph.Graph, modulePath string) map[string]int {
	// Check if module exists
	if _, exists := g.Modules[modulePath]; !exists {
		return nil
	}

	// Build reverse dependency graph
	reverseDeps := make(map[string][]string)
	for path, module := range g.Modules {
		for _, dep := range module.Dependencies {
			reverseDeps[dep] = append(reverseDeps[dep], path)
		}
	}

	// BFS to find all transitive dependents
	dependents := make(map[string]int)
	queue := []struct {
		path  string
		depth int
	}{{modulePath, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Skip if already visited at a shallower depth
		if existingDepth, visited := dependents[current.path]; visited && existingDepth <= current.depth {
			continue
		}

		dependents[current.path] = current.depth

		// Add dependents to queue
		for _, dependent := range reverseDeps[current.path] {
			queue = append(queue, struct {
				path  string
				depth int
			}{dependent, current.depth + 1})
		}
	}

	// Remove the target module itself
	delete(dependents, modulePath)

	return dependents
}

// CyclicDependencies detects all cyclic dependencies in the graph.
// Returns a slice of cycles, where each cycle is a slice of module paths.
func CyclicDependencies(g *graph.Graph) [][]string {
	sccs := StronglyConnectedComponents(g)
	cycles := [][]string{}

	// Only SCCs with more than one module are true cycles
	// (or single-node cycles with self-loops)
	for _, scc := range sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		} else if len(scc) == 1 {
			// Check for self-loop
			module := g.Modules[scc[0]]
			for _, dep := range module.Dependencies {
				if dep == scc[0] {
					cycles = append(cycles, scc)
					break
				}
			}
		}
	}

	return cycles
}

// DependencyDepth calculates the maximum dependency depth for a module.
// The depth is the longest path from the module to a leaf (module with no dependencies).
func DependencyDepth(g *graph.Graph, modulePath string) int {
	// Check if module exists
	if _, exists := g.Modules[modulePath]; !exists {
		return -1
	}

	// Memoization for efficiency
	memo := make(map[string]int)

	var calculateDepth func(string) int
	calculateDepth = func(path string) int {
		if depth, exists := memo[path]; exists {
			return depth
		}

		module, exists := g.Modules[path]
		if !exists || len(module.Dependencies) == 0 {
			memo[path] = 0
			return 0
		}

		maxDepth := 0
		for _, dep := range module.Dependencies {
			depth := calculateDepth(dep)
			if depth+1 > maxDepth {
				maxDepth = depth + 1
			}
		}

		memo[path] = maxDepth
		return maxDepth
	}

	return calculateDepth(modulePath)
}
