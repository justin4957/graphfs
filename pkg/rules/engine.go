/*
# Module: pkg/rules/engine.go
Rule engine for validating architectural constraints.

Main engine that orchestrates rule parsing, evaluation, and reporting.

## Linked Modules
- [./rule](./rule.go) - Rule data structures
- [./parser](./parser.go) - Rule parser
- [./evaluator](./evaluator.go) - Rule evaluator
- [./reporter](./reporter.go) - Violation reporter
- [../graph](../graph/graph.go) - Graph data structure

## Tags
rules, engine, validation

## Exports
Engine, ValidateRules

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#engine.go> a code:Module ;
    code:name "pkg/rules/engine.go" ;
    code:description "Rule engine for validating architectural constraints" ;
    code:language "go" ;
    code:layer "rules" ;
    code:linksTo <./rule.go>, <./parser.go>, <./evaluator.go>, <./reporter.go>, <../graph/graph.go> ;
    code:exports <#Engine>, <#ValidateRules> ;
    code:tags "rules", "engine", "validation" .
<!-- End LinkedDoc RDF -->
*/

package rules

import (
	"fmt"
	"time"

	"github.com/justin4957/graphfs/pkg/graph"
)

// Engine is the main rule validation engine
type Engine struct {
	graph     *graph.Graph
	parser    *Parser
	evaluator *Evaluator
}

// NewEngine creates a new rule engine
func NewEngine(g *graph.Graph) *Engine {
	return &Engine{
		graph:     g,
		parser:    NewParser(),
		evaluator: NewEvaluator(g),
	}
}

// ValidateFile validates rules from a file
func (e *Engine) ValidateFile(rulesFile string) (*ValidationResult, error) {
	// Parse rules
	ruleSet, err := e.parser.ParseFile(rulesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	return e.Validate(ruleSet.Rules)
}

// Validate validates a set of rules against the graph
func (e *Engine) Validate(rules []*Rule) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Violations:   make([]Violation, 0),
		PassedRules:  make([]*Rule, 0),
		FailedRules:  make([]*Rule, 0),
		SkippedRules: make([]*Rule, 0),
		TotalRules:   0,
		ErrorCount:   0,
		WarningCount: 0,
		InfoCount:    0,
	}

	// Evaluate each rule
	for _, rule := range rules {
		if !rule.Enabled {
			result.SkippedRules = append(result.SkippedRules, rule)
			continue
		}

		result.TotalRules++

		violations, err := e.evaluator.EvaluateRule(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate rule %s: %w", rule.ID, err)
		}

		if len(violations) > 0 {
			result.FailedRules = append(result.FailedRules, rule)
			result.Violations = append(result.Violations, violations...)

			// Count by severity
			for range violations {
				switch rule.Severity {
				case SeverityError:
					result.ErrorCount++
				case SeverityWarning:
					result.WarningCount++
				case SeverityInfo:
					result.InfoCount++
				}
			}
		} else {
			result.PassedRules = append(result.PassedRules, rule)
		}
	}

	// Calculate duration
	duration := time.Since(startTime)
	result.Duration = duration.Milliseconds()

	return result, nil
}

// ValidateWithFilter validates rules filtered by tags or severity
func (e *Engine) ValidateWithFilter(rules []*Rule, tags []string, minSeverity Severity) (*ValidationResult, error) {
	// Filter rules
	filteredRules := make([]*Rule, 0)

	for _, rule := range rules {
		// Check severity
		if !e.meetsMinimumSeverity(rule.Severity, minSeverity) {
			continue
		}

		// Check tags (if specified)
		if len(tags) > 0 && !e.hasMatchingTag(rule, tags) {
			continue
		}

		filteredRules = append(filteredRules, rule)
	}

	return e.Validate(filteredRules)
}

// meetsMinimumSeverity checks if a severity meets the minimum threshold
func (e *Engine) meetsMinimumSeverity(severity, minSeverity Severity) bool {
	severityOrder := map[Severity]int{
		SeverityInfo:    1,
		SeverityWarning: 2,
		SeverityError:   3,
	}

	return severityOrder[severity] >= severityOrder[minSeverity]
}

// hasMatchingTag checks if a rule has any of the specified tags
func (e *Engine) hasMatchingTag(rule *Rule, tags []string) bool {
	for _, ruleTag := range rule.Tags {
		for _, filterTag := range tags {
			if ruleTag == filterTag {
				return true
			}
		}
	}
	return false
}

// ValidateRules is a convenience function to validate rules from a file
func ValidateRules(g *graph.Graph, rulesFile string) (*ValidationResult, error) {
	engine := NewEngine(g)
	return engine.ValidateFile(rulesFile)
}

// GetBuiltInRules returns a set of built-in rules
func GetBuiltInRules() []*Rule {
	return []*Rule{
		{
			ID:          "no-circular-dependencies",
			Name:        "No circular dependencies",
			Description: "Detects circular dependencies between modules",
			Severity:    SeverityError,
			Pattern: `
				SELECT ?module1 ?module2 WHERE {
					?module1 code:linksTo ?module2 .
					?module2 code:linksTo ?module1 .
					FILTER (?module1 != ?module2)
				}
			`,
			Expect:     0,
			Enabled:    true,
			Suggestion: "Refactor to remove circular dependency by introducing an interface or shared module",
		},
		{
			ID:          "exports-documented",
			Name:        "All exports must be documented",
			Description: "Ensures all modules with exports have documentation",
			Severity:    SeverityWarning,
			Pattern: `
				SELECT ?module WHERE {
					?module code:exports ?export .
					FILTER NOT EXISTS { ?module code:description ?desc }
				}
			`,
			Expect:     0,
			Enabled:    true,
			Suggestion: "Add a description field to the module's LinkedDoc metadata",
		},
		{
			ID:          "modules-have-layer",
			Name:        "All modules must have a layer",
			Description: "Ensures all modules are assigned to an architectural layer",
			Severity:    SeverityWarning,
			Pattern: `
				SELECT ?module WHERE {
					?module a code:Module .
					FILTER NOT EXISTS { ?module code:layer ?layer }
				}
			`,
			Expect:     0,
			Enabled:    true,
			Suggestion: "Add a layer field to the module's LinkedDoc metadata",
		},
		{
			ID:          "modules-have-tags",
			Name:        "All modules should have tags",
			Description: "Ensures modules are properly tagged for categorization",
			Severity:    SeverityInfo,
			Pattern: `
				SELECT ?module WHERE {
					?module a code:Module .
					FILTER NOT EXISTS { ?module code:tags ?tag }
				}
			`,
			Expect:     0,
			Enabled:    true,
			Suggestion: "Add tags to the module's LinkedDoc metadata for better categorization",
		},
	}
}
