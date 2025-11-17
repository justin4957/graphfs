/*
# Module: pkg/analysis/security.go
Security boundary analysis for detecting and enforcing security zones.

Analyzes security boundaries, tracks data flow across zones, identifies violations,
and generates security audit reports.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure
- [./zones](./zones.go) - Security zones

## Tags
analysis, security, boundaries

## Exports
SecurityAnalysis, SecurityBoundary, SecurityViolation, AnalyzeSecurity

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#security.go> a code:Module ;
    code:name "pkg/analysis/security.go" ;
    code:description "Security boundary analysis for detecting and enforcing security zones" ;
    code:language "go" ;
    code:layer "analysis" ;
    code:linksTo <../graph/graph.go>, <./zones.go> ;
    code:exports <#SecurityAnalysis>, <#SecurityBoundary>, <#SecurityViolation>, <#AnalyzeSecurity> ;
    code:tags "analysis", "security", "boundaries" .
<!-- End LinkedDoc RDF -->
*/

package analysis

import (
	"fmt"
	"time"

	"github.com/justin4957/graphfs/pkg/graph"
)

// SecurityBoundary represents a boundary between security zones
type SecurityBoundary struct {
	From      SecurityZone
	To        SecurityZone
	Crossings []*BoundaryCrossing
	Allowed   bool
	Policy    string
}

// BoundaryCrossing represents a module crossing a security boundary
type BoundaryCrossing struct {
	Source      *graph.Module
	Destination *graph.Module
	SourceZone  SecurityZone
	DestZone    SecurityZone
	Risk        RiskLevel
	Path        []string
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type           string
	Crossing       *BoundaryCrossing
	Description    string
	Risk           RiskLevel
	Recommendation string
}

// SecurityAnalysis contains the results of security boundary analysis
type SecurityAnalysis struct {
	Zones           map[SecurityZone][]*ModuleZone
	Boundaries      []*SecurityBoundary
	Violations      []*SecurityViolation
	RiskScore       float64 // 0.0 (safe) to 10.0 (critical)
	Recommendations []string
	Duration        time.Duration
}

// SecurityOptions configures security analysis
type SecurityOptions struct {
	StrictMode       bool                // Enforce stricter boundary rules
	AllowedCrossings map[string][]string // Allowed zone crossings
}

// SecurityAnalyzer performs security boundary analysis
type SecurityAnalyzer struct {
	graph   *graph.Graph
	options SecurityOptions
}

// NewSecurityAnalyzer creates a new security analyzer
func NewSecurityAnalyzer(g *graph.Graph, opts SecurityOptions) *SecurityAnalyzer {
	return &SecurityAnalyzer{
		graph:   g,
		options: opts,
	}
}

// AnalyzeSecurity performs comprehensive security boundary analysis
func AnalyzeSecurity(g *graph.Graph, opts SecurityOptions) (*SecurityAnalysis, error) {
	analyzer := NewSecurityAnalyzer(g, opts)
	return analyzer.Analyze()
}

// Analyze performs the security analysis
func (sa *SecurityAnalyzer) Analyze() (*SecurityAnalysis, error) {
	startTime := time.Now()

	analysis := &SecurityAnalysis{
		Violations:      make([]*SecurityViolation, 0),
		Recommendations: make([]string, 0),
	}

	// Step 1: Classify modules into security zones
	analysis.Zones = ClassifyZones(sa.graph)

	// Step 2: Detect boundary crossings
	analysis.Boundaries = sa.detectBoundaries(analysis.Zones)

	// Step 3: Identify violations
	analysis.Violations = sa.detectViolations(analysis.Boundaries)

	// Step 4: Calculate risk score
	analysis.RiskScore = sa.calculateRiskScore(analysis.Violations)

	// Step 5: Generate recommendations
	analysis.Recommendations = sa.generateRecommendations(analysis.Violations)

	analysis.Duration = time.Since(startTime)
	return analysis, nil
}

// detectBoundaries detects security boundary crossings
func (sa *SecurityAnalyzer) detectBoundaries(zones map[SecurityZone][]*ModuleZone) []*SecurityBoundary {
	boundaries := make(map[string]*SecurityBoundary)

	// Create module to zone mapping
	moduleZones := make(map[string]SecurityZone)
	for zone, modules := range zones {
		for _, mz := range modules {
			moduleZones[mz.Module.Path] = zone
		}
	}

	// Find all boundary crossings
	for _, module := range sa.graph.Modules {
		sourceZone := moduleZones[module.Path]

		for _, depPath := range module.Dependencies {
			depModule := sa.graph.GetModule(depPath)
			if depModule == nil {
				continue
			}

			destZone := moduleZones[depPath]

			// Check if crossing a boundary
			if sourceZone != destZone {
				boundaryKey := string(sourceZone) + "->" + string(destZone)

				if boundaries[boundaryKey] == nil {
					boundaries[boundaryKey] = &SecurityBoundary{
						From:      sourceZone,
						To:        destZone,
						Crossings: make([]*BoundaryCrossing, 0),
						Allowed:   sa.isCrossingAllowed(sourceZone, destZone),
					}
				}

				crossing := &BoundaryCrossing{
					Source:      module,
					Destination: depModule,
					SourceZone:  sourceZone,
					DestZone:    destZone,
					Risk:        sa.assessCrossingRisk(sourceZone, destZone),
					Path:        []string{module.Path, depPath},
				}

				boundaries[boundaryKey].Crossings = append(boundaries[boundaryKey].Crossings, crossing)
			}
		}
	}

	// Convert map to slice
	result := make([]*SecurityBoundary, 0, len(boundaries))
	for _, boundary := range boundaries {
		result = append(result, boundary)
	}

	return result
}

// isCrossingAllowed checks if a boundary crossing is allowed
func (sa *SecurityAnalyzer) isCrossingAllowed(from, to SecurityZone) bool {
	// Default allowed crossings (following principle of least privilege)
	defaultAllowed := map[SecurityZone][]SecurityZone{
		ZonePublic:   {ZoneTrusted, ZoneInternal},           // Public servers can access trusted/internal (but NOT data/admin directly)
		ZoneTrusted:  {ZoneInternal, ZoneData},              // Trusted can access internal and data
		ZoneInternal: {ZoneTrusted, ZonePublic, ZoneData},   // Internal can access data, trusted, public
		ZoneAdmin:    {ZoneTrusted, ZoneInternal, ZoneData}, // Admin has broad access
		ZoneData:     {},                                    // Data layer shouldn't call out
	}

	// Check custom allowed crossings first
	if sa.options.AllowedCrossings != nil {
		if allowed, ok := sa.options.AllowedCrossings[string(from)]; ok {
			for _, allowedTo := range allowed {
				if allowedTo == string(to) {
					return true
				}
			}
			return false // Custom rules override defaults
		}
	}

	// Check default rules
	if allowed, ok := defaultAllowed[from]; ok {
		for _, allowedZone := range allowed {
			if allowedZone == to {
				return true
			}
		}
	}

	return false
}

// assessCrossingRisk assesses the risk level of a boundary crossing
func (sa *SecurityAnalyzer) assessCrossingRisk(from, to SecurityZone) RiskLevel {
	// Critical risk scenarios
	if from == ZonePublic && to == ZoneAdmin {
		return RiskLevelCritical // Public to admin is critical
	}
	if from == ZonePublic && to == ZoneData {
		return RiskLevelCritical // Public to data is critical
	}

	// High risk scenarios
	if from == ZonePublic && to == ZoneInternal {
		return RiskLevelHigh // Public exposing internal details
	}
	if from == ZoneTrusted && to == ZoneAdmin {
		return RiskLevelHigh // Trusted to admin needs careful auth
	}

	// Medium risk scenarios
	if from == ZoneInternal && to == ZonePublic {
		return RiskLevelMedium // Internal leaking to public
	}
	if from == ZoneData && to == ZonePublic {
		return RiskLevelMedium // Data exposed to public
	}

	// Low risk for normal flows
	if from == ZonePublic && to == ZoneTrusted {
		return RiskLevelLow
	}
	if from == ZoneTrusted && to == ZoneData {
		return RiskLevelLow
	}

	return RiskLevelLow
}

// detectViolations identifies security policy violations
func (sa *SecurityAnalyzer) detectViolations(boundaries []*SecurityBoundary) []*SecurityViolation {
	violations := make([]*SecurityViolation, 0)

	for _, boundary := range boundaries {
		if !boundary.Allowed {
			for _, crossing := range boundary.Crossings {
				violation := &SecurityViolation{
					Type:           "unauthorized_boundary_crossing",
					Crossing:       crossing,
					Risk:           crossing.Risk,
					Description:    fmt.Sprintf("Unauthorized crossing from %s to %s zone", crossing.SourceZone, crossing.DestZone),
					Recommendation: sa.getRecommendationForViolation(crossing),
				}
				violations = append(violations, violation)
			}
		} else if sa.options.StrictMode {
			// In strict mode, flag high-risk crossings even if allowed
			for _, crossing := range boundary.Crossings {
				if crossing.Risk == RiskLevelCritical || crossing.Risk == RiskLevelHigh {
					violation := &SecurityViolation{
						Type:           "high_risk_crossing",
						Crossing:       crossing,
						Risk:           crossing.Risk,
						Description:    fmt.Sprintf("High-risk crossing from %s to %s zone requires validation", crossing.SourceZone, crossing.DestZone),
						Recommendation: "Add input validation, authentication, and rate limiting",
					}
					violations = append(violations, violation)
				}
			}
		}
	}

	return violations
}

// getRecommendationForViolation generates a recommendation for a violation
func (sa *SecurityAnalyzer) getRecommendationForViolation(crossing *BoundaryCrossing) string {
	switch crossing.Risk {
	case RiskLevelCritical:
		return fmt.Sprintf("CRITICAL: Remove direct access from %s to %s. Use proper authentication and authorization layers.", crossing.Source.Path, crossing.Destination.Path)
	case RiskLevelHigh:
		return fmt.Sprintf("Add authentication middleware and input validation for %s before accessing %s", crossing.Source.Path, crossing.Destination.Path)
	case RiskLevelMedium:
		return "Review data flow and add sanitization if needed"
	default:
		return "Monitor this crossing and ensure proper error handling"
	}
}

// calculateRiskScore calculates overall security risk score (0-10)
func (sa *SecurityAnalyzer) calculateRiskScore(violations []*SecurityViolation) float64 {
	if len(violations) == 0 {
		return 0.0
	}

	totalRisk := 0.0
	riskWeights := map[RiskLevel]float64{
		RiskLevelCritical: 10.0,
		RiskLevelHigh:     7.0,
		RiskLevelMedium:   4.0,
		RiskLevelLow:      2.0,
	}

	for _, violation := range violations {
		totalRisk += riskWeights[violation.Risk]
	}

	// Normalize to 0-10 scale
	avgRisk := totalRisk / float64(len(violations))
	if avgRisk > 10.0 {
		avgRisk = 10.0
	}

	return avgRisk
}

// generateRecommendations generates security recommendations
func (sa *SecurityAnalyzer) generateRecommendations(violations []*SecurityViolation) []string {
	recommendations := make([]string, 0)

	if len(violations) == 0 {
		recommendations = append(recommendations, "No security violations detected. Continue monitoring.")
		return recommendations
	}

	// Count violations by risk
	criticalCount := 0
	highCount := 0

	for _, v := range violations {
		if v.Risk == RiskLevelCritical {
			criticalCount++
		} else if v.Risk == RiskLevelHigh {
			highCount++
		}
	}

	if criticalCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Address %d CRITICAL security violations immediately", criticalCount))
	}
	if highCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review and mitigate %d HIGH-risk security issues", highCount))
	}

	recommendations = append(recommendations, "Implement proper authentication and authorization layers")
	recommendations = append(recommendations, "Add input validation and sanitization")
	recommendations = append(recommendations, "Enable security audit logging")

	return recommendations
}

// HasViolations returns true if there are any security violations
func (sa *SecurityAnalysis) HasViolations() bool {
	return len(sa.Violations) > 0
}

// GetCriticalViolations returns only critical violations
func (sa *SecurityAnalysis) GetCriticalViolations() []*SecurityViolation {
	critical := make([]*SecurityViolation, 0)
	for _, v := range sa.Violations {
		if v.Risk == RiskLevelCritical {
			critical = append(critical, v)
		}
	}
	return critical
}

// GetHighRiskViolations returns high and critical violations
func (sa *SecurityAnalysis) GetHighRiskViolations() []*SecurityViolation {
	highRisk := make([]*SecurityViolation, 0)
	for _, v := range sa.Violations {
		if v.Risk == RiskLevelCritical || v.Risk == RiskLevelHigh {
			highRisk = append(highRisk, v)
		}
	}
	return highRisk
}
