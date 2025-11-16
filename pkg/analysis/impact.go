/*
# Module: pkg/analysis/impact.go
Impact analysis engine for GraphFS.

Provides impact analysis capabilities to assess the effects of modifying or
removing modules from the codebase. Calculates direct and transitive impacts,
risk levels, and provides recommendations.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [./graph_algorithms](./graph_algorithms.go) - Graph algorithms

## Tags
analysis, impact-analysis, refactoring, risk-assessment

## Exports
ImpactAnalysis, ImpactResult, RiskLevel, AnalyzeImpact

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#impact.go> a code:Module ;
    code:name "pkg/analysis/impact.go" ;
    code:description "Impact analysis engine for GraphFS" ;
    code:language "go" ;
    code:layer "analysis" ;
    code:linksTo <../graph/graph.go>, <./graph_algorithms.go> ;
    code:exports <#ImpactAnalysis>, <#ImpactResult>, <#RiskLevel>, <#AnalyzeImpact> ;
    code:tags "analysis", "impact-analysis", "refactoring", "risk-assessment" .
<!-- End LinkedDoc RDF -->
*/

package analysis

import (
	"fmt"
	"sort"

	"github.com/justin4957/graphfs/pkg/graph"
)

// RiskLevel represents the risk level of making changes to a module
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// ImpactResult contains the results of an impact analysis
type ImpactResult struct {
	TargetModule string // Module being analyzed

	// Direct impact
	DirectDependents   []string // Modules that directly depend on target
	DirectDependencies []string // Modules that target directly depends on

	// Transitive impact
	TransitiveDependents map[string]int // All modules affected (with depth)
	TotalImpactedModules int            // Total number of modules affected

	// Impact by layer
	ImpactByLayer map[string]int // Number of impacted modules per layer

	// Risk assessment
	RiskLevel       RiskLevel  // Overall risk level
	RiskFactors     []string   // Factors contributing to risk
	Recommendations []string   // Recommended actions
	CriticalPaths   [][]string // Critical dependency paths
	BreakingChanges bool       // Whether changes would be breaking

	// Additional metrics
	MaxImpactDepth   int     // Maximum depth of impact
	ImpactPercentage float64 // Percentage of total modules impacted
}

// ImpactAnalysis provides impact analysis capabilities
type ImpactAnalysis struct {
	graph *graph.Graph
}

// NewImpactAnalysis creates a new impact analysis engine
func NewImpactAnalysis(g *graph.Graph) *ImpactAnalysis {
	return &ImpactAnalysis{
		graph: g,
	}
}

// AnalyzeImpact performs a comprehensive impact analysis for a module
func (ia *ImpactAnalysis) AnalyzeImpact(modulePath string) (*ImpactResult, error) {
	// Check if module exists
	module, exists := ia.graph.Modules[modulePath]
	if !exists {
		return nil, fmt.Errorf("module not found: %s", modulePath)
	}

	result := &ImpactResult{
		TargetModule:       modulePath,
		DirectDependents:   make([]string, 0),
		DirectDependencies: make([]string, 0),
		ImpactByLayer:      make(map[string]int),
		RiskFactors:        make([]string, 0),
		Recommendations:    make([]string, 0),
		CriticalPaths:      make([][]string, 0),
	}

	// Get direct dependents and dependencies
	result.DirectDependents = ia.getDirectDependents(modulePath)
	result.DirectDependencies = module.Dependencies

	// Get transitive dependents (modules impacted by changes)
	result.TransitiveDependents = TransitiveDependents(ia.graph, modulePath)
	result.TotalImpactedModules = len(result.TransitiveDependents)

	// Calculate impact by layer
	ia.calculateImpactByLayer(result)

	// Find critical paths (shortest paths to highly impacted modules)
	ia.findCriticalPaths(result)

	// Calculate max impact depth
	result.MaxImpactDepth = ia.calculateMaxDepth(result.TransitiveDependents)

	// Calculate impact percentage
	totalModules := len(ia.graph.Modules)
	if totalModules > 0 {
		result.ImpactPercentage = float64(result.TotalImpactedModules) / float64(totalModules) * 100
	}

	// Assess risk level
	ia.assessRisk(result)

	// Generate recommendations
	ia.generateRecommendations(result)

	return result, nil
}

// AnalyzeMultipleModules analyzes the combined impact of changes to multiple modules
func (ia *ImpactAnalysis) AnalyzeMultipleModules(modulePaths []string) (*ImpactResult, error) {
	if len(modulePaths) == 0 {
		return nil, fmt.Errorf("no modules specified")
	}

	// Verify all modules exist
	for _, path := range modulePaths {
		if _, exists := ia.graph.Modules[path]; !exists {
			return nil, fmt.Errorf("module not found: %s", path)
		}
	}

	result := &ImpactResult{
		TargetModule:         fmt.Sprintf("%d modules", len(modulePaths)),
		DirectDependents:     make([]string, 0),
		DirectDependencies:   make([]string, 0),
		TransitiveDependents: make(map[string]int),
		ImpactByLayer:        make(map[string]int),
		RiskFactors:          make([]string, 0),
		Recommendations:      make([]string, 0),
		CriticalPaths:        make([][]string, 0),
	}

	// Combine impacts from all modules
	directDependentsSet := make(map[string]bool)
	directDependenciesSet := make(map[string]bool)

	for _, modulePath := range modulePaths {
		module := ia.graph.Modules[modulePath]

		// Collect direct dependents
		for _, dep := range ia.getDirectDependents(modulePath) {
			directDependentsSet[dep] = true
		}

		// Collect direct dependencies
		for _, dep := range module.Dependencies {
			directDependenciesSet[dep] = true
		}

		// Merge transitive dependents (keep minimum depth)
		transitive := TransitiveDependents(ia.graph, modulePath)
		for path, depth := range transitive {
			if existingDepth, exists := result.TransitiveDependents[path]; !exists || depth < existingDepth {
				result.TransitiveDependents[path] = depth
			}
		}
	}

	// Convert sets to slices
	for dep := range directDependentsSet {
		result.DirectDependents = append(result.DirectDependents, dep)
	}
	for dep := range directDependenciesSet {
		result.DirectDependencies = append(result.DirectDependencies, dep)
	}

	sort.Strings(result.DirectDependents)
	sort.Strings(result.DirectDependencies)

	result.TotalImpactedModules = len(result.TransitiveDependents)

	// Calculate metrics
	ia.calculateImpactByLayer(result)
	result.MaxImpactDepth = ia.calculateMaxDepth(result.TransitiveDependents)

	totalModules := len(ia.graph.Modules)
	if totalModules > 0 {
		result.ImpactPercentage = float64(result.TotalImpactedModules) / float64(totalModules) * 100
	}

	// Assess risk and generate recommendations
	ia.assessRisk(result)
	ia.generateRecommendations(result)

	return result, nil
}

// getDirectDependents returns modules that directly depend on the target
func (ia *ImpactAnalysis) getDirectDependents(modulePath string) []string {
	dependents := make([]string, 0)

	for path, module := range ia.graph.Modules {
		for _, dep := range module.Dependencies {
			if dep == modulePath {
				dependents = append(dependents, path)
				break
			}
		}
	}

	sort.Strings(dependents)
	return dependents
}

// calculateImpactByLayer calculates the number of impacted modules per layer
func (ia *ImpactAnalysis) calculateImpactByLayer(result *ImpactResult) {
	for modulePath := range result.TransitiveDependents {
		if module, exists := ia.graph.Modules[modulePath]; exists {
			layer := module.Layer
			if layer == "" {
				layer = "unknown"
			}
			result.ImpactByLayer[layer]++
		}
	}
}

// findCriticalPaths finds the shortest paths to highly impacted modules
func (ia *ImpactAnalysis) findCriticalPaths(result *ImpactResult) {
	// Find modules with high impact (many dependents)
	criticalModules := make([]string, 0)

	for modulePath := range result.TransitiveDependents {
		dependents := ia.getDirectDependents(modulePath)
		if len(dependents) >= 3 { // Threshold for "critical"
			criticalModules = append(criticalModules, modulePath)
		}
	}

	// Limit to top 5 critical modules
	if len(criticalModules) > 5 {
		sort.Strings(criticalModules)
		criticalModules = criticalModules[:5]
	}

	// Find shortest paths from target to critical modules
	for _, criticalModule := range criticalModules {
		path := ShortestPath(ia.graph, criticalModule, result.TargetModule)
		if len(path) > 1 {
			result.CriticalPaths = append(result.CriticalPaths, path)
		}
	}
}

// calculateMaxDepth calculates the maximum depth in the transitive dependents map
func (ia *ImpactAnalysis) calculateMaxDepth(transitive map[string]int) int {
	maxDepth := 0
	for _, depth := range transitive {
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

// assessRisk determines the risk level based on impact metrics
func (ia *ImpactAnalysis) assessRisk(result *ImpactResult) {
	riskScore := 0

	// Factor 1: Number of direct dependents
	directCount := len(result.DirectDependents)
	if directCount > 10 {
		riskScore += 3
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("High number of direct dependents (%d)", directCount))
	} else if directCount > 5 {
		riskScore += 2
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("Moderate number of direct dependents (%d)", directCount))
	} else if directCount > 0 {
		riskScore += 1
	}

	// Factor 2: Total impacted modules
	if result.TotalImpactedModules > 20 {
		riskScore += 3
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("Large transitive impact (%d modules)", result.TotalImpactedModules))
	} else if result.TotalImpactedModules > 10 {
		riskScore += 2
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("Moderate transitive impact (%d modules)", result.TotalImpactedModules))
	} else if result.TotalImpactedModules > 0 {
		riskScore += 1
	}

	// Factor 3: Impact percentage
	if result.ImpactPercentage > 30 {
		riskScore += 2
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("High impact percentage (%.1f%%)", result.ImpactPercentage))
	} else if result.ImpactPercentage > 15 {
		riskScore += 1
	}

	// Factor 4: Multiple layers impacted
	if len(result.ImpactByLayer) > 3 {
		riskScore += 2
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("Impacts multiple layers (%d)", len(result.ImpactByLayer)))
	} else if len(result.ImpactByLayer) > 1 {
		riskScore += 1
	}

	// Factor 5: Maximum depth
	if result.MaxImpactDepth > 4 {
		riskScore += 1
		result.RiskFactors = append(result.RiskFactors, fmt.Sprintf("Deep dependency chain (depth %d)", result.MaxImpactDepth))
	}

	// Determine risk level
	if riskScore >= 8 {
		result.RiskLevel = RiskLevelCritical
		result.BreakingChanges = true
	} else if riskScore >= 5 {
		result.RiskLevel = RiskLevelHigh
		result.BreakingChanges = true
	} else if riskScore >= 3 {
		result.RiskLevel = RiskLevelMedium
		result.BreakingChanges = false
	} else {
		result.RiskLevel = RiskLevelLow
		result.BreakingChanges = false
	}

	// Add risk level summary
	if len(result.RiskFactors) == 0 {
		result.RiskFactors = append(result.RiskFactors, "Minimal impact on codebase")
	}
}

// generateRecommendations provides recommendations based on impact analysis
func (ia *ImpactAnalysis) generateRecommendations(result *ImpactResult) {
	switch result.RiskLevel {
	case RiskLevelCritical:
		result.Recommendations = append(result.Recommendations,
			"⚠️  CRITICAL: Coordinate with all affected teams before making changes",
			"Create detailed migration plan and timeline",
			"Implement changes incrementally with feature flags",
			"Set up comprehensive integration testing",
			"Plan for potential rollback strategy",
		)

	case RiskLevelHigh:
		result.Recommendations = append(result.Recommendations,
			"Notify all affected teams of planned changes",
			"Review and update API contracts if needed",
			"Add deprecation warnings before breaking changes",
			"Increase test coverage for impacted modules",
			"Consider staged rollout approach",
		)

	case RiskLevelMedium:
		result.Recommendations = append(result.Recommendations,
			"Review impacted modules before changes",
			"Update related documentation and tests",
			"Communicate changes to relevant teams",
			"Monitor for issues after deployment",
		)

	case RiskLevelLow:
		result.Recommendations = append(result.Recommendations,
			"Standard code review process is sufficient",
			"Update tests for modified functionality",
			"Document any API changes",
		)
	}

	// Layer-specific recommendations
	if len(result.ImpactByLayer) > 2 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Impacts %d architectural layers - review layer boundaries", len(result.ImpactByLayer)),
		)
	}

	// Critical path recommendations
	if len(result.CriticalPaths) > 0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Review %d critical dependency paths identified", len(result.CriticalPaths)),
		)
	}
}

// CompareImpacts compares the impact of modifying multiple modules
func (ia *ImpactAnalysis) CompareImpacts(modulePaths []string) (map[string]*ImpactResult, error) {
	results := make(map[string]*ImpactResult)

	for _, path := range modulePaths {
		result, err := ia.AnalyzeImpact(path)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", path, err)
		}
		results[path] = result
	}

	return results, nil
}
