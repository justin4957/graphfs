/*
# Module: pkg/analysis/coverage.go
Usage coverage analysis for tracking module and symbol usage.

Analyzes how modules and symbols are used throughout the codebase to determine
coverage and identify underutilized components.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [./deadcode](./deadcode.go) - Dead code detection

## Tags
analysis, coverage, usage

## Exports
CoverageAnalysis, ModuleCoverage, AnalyzeCoverage

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#coverage.go> a code:Module ;
    code:name "pkg/analysis/coverage.go" ;
    code:description "Usage coverage analysis for tracking module and symbol usage" ;
    code:language "go" ;
    code:layer "analysis" ;
    code:linksTo <../graph/graph.go>, <./deadcode.go> ;
    code:exports <#CoverageAnalysis>, <#ModuleCoverage>, <#AnalyzeCoverage> ;
    code:tags "analysis", "coverage", "usage" .
<!-- End LinkedDoc RDF -->
*/

package analysis

import (
	"sort"

	"github.com/justin4957/graphfs/pkg/graph"
)

// CoverageAnalysis represents usage coverage of the codebase
type CoverageAnalysis struct {
	ModuleCoverage      []*ModuleCoverage
	TotalModules        int
	ReferencedModules   int
	UnreferencedModules int
	CoveragePercent     float64
	HighUsageModules    []*ModuleCoverage // Top 10% most referenced
	LowUsageModules     []*ModuleCoverage // Bottom 10% least referenced
}

// ModuleCoverage represents usage statistics for a single module
type ModuleCoverage struct {
	Module         *graph.Module
	IncomingRefs   int     // Number of modules that depend on this one
	OutgoingRefs   int     // Number of modules this one depends on
	UsageScore     float64 // Normalized usage score (0.0-1.0)
	IsEntryPoint   bool    // Whether this is an entry point
	IsDeadCode     bool    // Whether this appears to be dead code
	DirectUsers    []string
	TransitiveRefs int // Number of transitive references
}

// AnalyzeCoverage performs usage coverage analysis
func AnalyzeCoverage(g *graph.Graph) *CoverageAnalysis {
	analyzer := &coverageAnalyzer{
		graph:             g,
		incomingRefs:      make(map[string][]string),
		outgoingRefs:      make(map[string][]string),
		transitiveRefs:    make(map[string]int),
		moduleCoverageMap: make(map[string]*ModuleCoverage),
	}

	return analyzer.analyze()
}

type coverageAnalyzer struct {
	graph             *graph.Graph
	incomingRefs      map[string][]string
	outgoingRefs      map[string][]string
	transitiveRefs    map[string]int
	moduleCoverageMap map[string]*ModuleCoverage
}

func (a *coverageAnalyzer) analyze() *CoverageAnalysis {
	// Build reference maps
	a.buildReferenceMaps()

	// Calculate transitive references
	a.calculateTransitiveReferences()

	// Create module coverage entries
	coverages := make([]*ModuleCoverage, 0, len(a.graph.Modules))
	for _, module := range a.graph.Modules {
		coverage := a.analyzeModule(module)
		coverages = append(coverages, coverage)
		a.moduleCoverageMap[module.Path] = coverage
	}

	// Calculate statistics
	analysis := &CoverageAnalysis{
		ModuleCoverage: coverages,
		TotalModules:   len(coverages),
	}

	for _, cov := range coverages {
		if cov.IncomingRefs > 0 {
			analysis.ReferencedModules++
		} else {
			analysis.UnreferencedModules++
		}
	}

	if analysis.TotalModules > 0 {
		analysis.CoveragePercent = float64(analysis.ReferencedModules) / float64(analysis.TotalModules) * 100.0
	}

	// Sort by usage score
	sort.Slice(coverages, func(i, j int) bool {
		return coverages[i].UsageScore > coverages[j].UsageScore
	})

	// Identify high and low usage modules
	topN := len(coverages) / 10
	if topN < 1 {
		topN = 1
	}
	if topN > len(coverages) {
		topN = len(coverages)
	}

	analysis.HighUsageModules = coverages[:topN]
	if len(coverages) >= topN {
		analysis.LowUsageModules = coverages[len(coverages)-topN:]
	}

	return analysis
}

func (a *coverageAnalyzer) buildReferenceMaps() {
	for _, module := range a.graph.Modules {
		a.outgoingRefs[module.Path] = module.Dependencies

		for _, dep := range module.Dependencies {
			a.incomingRefs[dep] = append(a.incomingRefs[dep], module.Path)
		}
	}
}

func (a *coverageAnalyzer) calculateTransitiveReferences() {
	for _, module := range a.graph.Modules {
		visited := make(map[string]bool)
		a.countTransitive(module.Path, visited)
		a.transitiveRefs[module.Path] = len(visited) - 1 // Exclude self
	}
}

func (a *coverageAnalyzer) countTransitive(modulePath string, visited map[string]bool) {
	if visited[modulePath] {
		return
	}
	visited[modulePath] = true

	for _, user := range a.incomingRefs[modulePath] {
		a.countTransitive(user, visited)
	}
}

func (a *coverageAnalyzer) analyzeModule(module *graph.Module) *ModuleCoverage {
	incoming := len(a.incomingRefs[module.Path])
	outgoing := len(a.outgoingRefs[module.Path])
	transitive := a.transitiveRefs[module.Path]

	// Calculate usage score (0.0-1.0)
	// Factors: incoming refs, transitive refs, whether it's exported
	score := 0.0

	if incoming > 0 {
		// Base score from incoming references (normalized by total modules)
		score += float64(incoming) / float64(len(a.graph.Modules)) * 0.5

		// Bonus for transitive usage
		score += float64(transitive) / float64(len(a.graph.Modules)) * 0.3

		// Bonus for having exports
		if len(module.Exports) > 0 {
			score += 0.2
		}
	}

	if score > 1.0 {
		score = 1.0
	}

	return &ModuleCoverage{
		Module:         module,
		IncomingRefs:   incoming,
		OutgoingRefs:   outgoing,
		TransitiveRefs: transitive,
		UsageScore:     score,
		IsEntryPoint:   a.isEntryPoint(module),
		IsDeadCode:     incoming == 0 && !a.isEntryPoint(module),
		DirectUsers:    a.incomingRefs[module.Path],
	}
}

func (a *coverageAnalyzer) isEntryPoint(module *graph.Module) bool {
	// Check for main or init functions
	for _, export := range module.Exports {
		if export == "main" || export == "init" {
			return true
		}
	}
	return false
}

// GetUnusedModules returns modules with no incoming references
func (a *CoverageAnalysis) GetUnusedModules() []*ModuleCoverage {
	unused := make([]*ModuleCoverage, 0)
	for _, cov := range a.ModuleCoverage {
		if cov.IncomingRefs == 0 && !cov.IsEntryPoint {
			unused = append(unused, cov)
		}
	}
	return unused
}

// GetHighUsageModules returns modules with high usage (top 20%)
func (a *CoverageAnalysis) GetHighUsageModules() []*ModuleCoverage {
	threshold := 0.7
	high := make([]*ModuleCoverage, 0)
	for _, cov := range a.ModuleCoverage {
		if cov.UsageScore >= threshold {
			high = append(high, cov)
		}
	}
	return high
}

// GetLowUsageModules returns modules with low usage (bottom 20%)
func (a *CoverageAnalysis) GetLowUsageModules() []*ModuleCoverage {
	threshold := 0.3
	low := make([]*ModuleCoverage, 0)
	for _, cov := range a.ModuleCoverage {
		if cov.UsageScore < threshold && !cov.IsEntryPoint {
			low = append(low, cov)
		}
	}
	return low
}
