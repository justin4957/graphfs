/*
# Module: pkg/diff/differ.go
Graph diff implementation for analyzing changes between commits.

Compares knowledge graphs between Git commits to detect added, removed, and
modified modules, along with their dependency and export changes.

## Linked Modules
- [../graph](../graph/graph.go) - Graph building
- [../scanner](../scanner/scanner.go) - File scanning

## Tags
diff, git, analysis

## Exports
Differ, GraphDiff, ModuleChange, NewDiffer

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#differ.go> a code:Module ;
    code:name "pkg/diff/differ.go" ;
    code:description "Graph diff implementation for analyzing changes between commits" ;
    code:language "go" ;
    code:layer "diff" ;
    code:linksTo <../graph/graph.go>, <../scanner/scanner.go> ;
    code:exports <#Differ>, <#GraphDiff>, <#ModuleChange>, <#NewDiffer> ;
    code:tags "diff", "git", "analysis" .
<!-- End LinkedDoc RDF -->
*/

package diff

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
)

// GraphDiff represents the differences between two knowledge graphs
type GraphDiff struct {
	Added     []*graph.Module // Modules added in new graph
	Removed   []*graph.Module // Modules removed from old graph
	Modified  []*ModuleChange // Modules that changed
	Unchanged []*graph.Module // Modules that didn't change
	OldGraph  *graph.Graph    // Reference graph
	NewGraph  *graph.Graph    // Current graph
}

// ModuleChange represents changes to a single module
type ModuleChange struct {
	Module         *graph.Module // The module in new graph
	DepsAdded      []string      // Dependencies added
	DepsRemoved    []string      // Dependencies removed
	ExportsAdded   []string      // Exports added
	ExportsRemoved []string      // Exports removed
	LayerChanged   bool          // Whether layer changed
	OldLayer       string        // Previous layer
	NewLayer       string        // Current layer
}

// Differ compares knowledge graphs between commits
type Differ struct {
	gitRepo string
}

// NewDiffer creates a new graph differ for a Git repository
func NewDiffer(repo string) *Differ {
	return &Differ{gitRepo: repo}
}

// Diff compares the current state against a Git reference
func (d *Differ) Diff(ref string) (*GraphDiff, error) {
	// Build graph for current state
	currentGraph, err := d.buildCurrentGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to build current graph: %w", err)
	}

	// Build graph for reference commit
	refGraph, err := d.buildGraphAtRef(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph at %s: %w", ref, err)
	}

	// Compare graphs
	diff := d.compareGraphs(refGraph, currentGraph)
	diff.OldGraph = refGraph
	diff.NewGraph = currentGraph

	return diff, nil
}

// buildCurrentGraph builds the graph for the current working directory
func (d *Differ) buildCurrentGraph() (*graph.Graph, error) {
	builder := graph.NewBuilder()
	return builder.Build(d.gitRepo, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			MaxFileSize:    1024 * 1024,
			FollowSymlinks: false,
			IgnoreFiles:    []string{".gitignore", ".graphfsignore"},
			UseDefaults:    true,
			Concurrent:     true,
		},
		Validate:       false,
		ReportProgress: false,
		UseCache:       true,
	})
}

// buildGraphAtRef builds the graph for a specific Git reference
func (d *Differ) buildGraphAtRef(ref string) (*graph.Graph, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "graphfs-diff-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use git worktree to checkout ref to temp directory
	cmd := exec.Command("git", "worktree", "add", "--detach", tmpDir, ref)
	cmd.Dir = d.gitRepo
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git worktree failed: %w\nOutput: %s", err, string(output))
	}

	// Ensure worktree is removed even if build fails
	defer func() {
		cmd := exec.Command("git", "worktree", "remove", "--force", tmpDir)
		cmd.Dir = d.gitRepo
		cmd.Run()
	}()

	// Build graph from temp directory
	builder := graph.NewBuilder()
	return builder.Build(tmpDir, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			MaxFileSize:    1024 * 1024,
			FollowSymlinks: false,
			IgnoreFiles:    []string{".gitignore", ".graphfsignore"},
			UseDefaults:    true,
			Concurrent:     true,
		},
		Validate:       false,
		ReportProgress: false,
		UseCache:       false, // Don't use cache for ref builds
	})
}

// compareGraphs compares two graphs and returns the differences
func (d *Differ) compareGraphs(old, new *graph.Graph) *GraphDiff {
	diff := &GraphDiff{
		Added:     []*graph.Module{},
		Removed:   []*graph.Module{},
		Modified:  []*ModuleChange{},
		Unchanged: []*graph.Module{},
	}

	// Create maps for quick lookup
	oldModules := make(map[string]*graph.Module)
	for _, m := range old.Modules {
		oldModules[m.Path] = m
	}

	newModules := make(map[string]*graph.Module)
	for _, m := range new.Modules {
		newModules[m.Path] = m
	}

	// Find added and modified modules
	for path, newMod := range newModules {
		oldMod, existed := oldModules[path]
		if !existed {
			diff.Added = append(diff.Added, newMod)
		} else {
			change := d.compareModules(oldMod, newMod)
			if change != nil {
				diff.Modified = append(diff.Modified, change)
			} else {
				diff.Unchanged = append(diff.Unchanged, newMod)
			}
		}
	}

	// Find removed modules
	for path, oldMod := range oldModules {
		if _, exists := newModules[path]; !exists {
			diff.Removed = append(diff.Removed, oldMod)
		}
	}

	return diff
}

// compareModules compares two module versions and returns changes
func (d *Differ) compareModules(old, new *graph.Module) *ModuleChange {
	change := &ModuleChange{Module: new}
	hasChanges := false

	// Compare dependencies
	oldDeps := toSet(old.Dependencies)
	newDeps := toSet(new.Dependencies)

	for dep := range newDeps {
		if !oldDeps[dep] {
			change.DepsAdded = append(change.DepsAdded, dep)
			hasChanges = true
		}
	}

	for dep := range oldDeps {
		if !newDeps[dep] {
			change.DepsRemoved = append(change.DepsRemoved, dep)
			hasChanges = true
		}
	}

	// Compare exports
	oldExports := toSet(old.Exports)
	newExports := toSet(new.Exports)

	for exp := range newExports {
		if !oldExports[exp] {
			change.ExportsAdded = append(change.ExportsAdded, exp)
			hasChanges = true
		}
	}

	for exp := range oldExports {
		if !newExports[exp] {
			change.ExportsRemoved = append(change.ExportsRemoved, exp)
			hasChanges = true
		}
	}

	// Compare layer
	if old.Layer != new.Layer {
		change.LayerChanged = true
		change.OldLayer = old.Layer
		change.NewLayer = new.Layer
		hasChanges = true
	}

	if hasChanges {
		return change
	}
	return nil
}

// toSet converts a slice to a set (map[string]bool)
func toSet(items []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range items {
		set[item] = true
	}
	return set
}

// Stats returns summary statistics for the diff
func (d *GraphDiff) Stats() DiffStats {
	return DiffStats{
		Added:     len(d.Added),
		Removed:   len(d.Removed),
		Modified:  len(d.Modified),
		Unchanged: len(d.Unchanged),
		Total:     len(d.Added) + len(d.Removed) + len(d.Modified) + len(d.Unchanged),
	}
}

// DiffStats contains summary statistics
type DiffStats struct {
	Added     int
	Removed   int
	Modified  int
	Unchanged int
	Total     int
}
