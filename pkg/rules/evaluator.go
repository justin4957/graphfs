/*
# Module: pkg/rules/evaluator.go
SPARQL-based rule evaluator for executing rules against the knowledge graph.

Evaluates architectural rules using SPARQL queries and detects violations.

## Linked Modules
- [./rule](./rule.go) - Rule data structures
- [../graph](../graph/graph.go) - Graph data structure
- [../query](../query/sparql.go) - SPARQL query engine

## Tags
rules, evaluator, sparql

## Exports
Evaluator, EvaluateRule

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#evaluator.go> a code:Module ;
    code:name "pkg/rules/evaluator.go" ;
    code:description "SPARQL-based rule evaluator for executing rules" ;
    code:language "go" ;
    code:layer "rules" ;
    code:linksTo <./rule.go>, <../graph/graph.go>, <../query/sparql.go> ;
    code:exports <#Evaluator>, <#EvaluateRule> ;
    code:tags "rules", "evaluator", "sparql" .
<!-- End LinkedDoc RDF -->
*/

package rules

import (
	"fmt"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
)

// Evaluator evaluates rules against a knowledge graph
type Evaluator struct {
	graph    *graph.Graph
	executor *query.Executor
}

// NewEvaluator creates a new rule evaluator
func NewEvaluator(g *graph.Graph) *Evaluator {
	return &Evaluator{
		graph:    g,
		executor: query.NewExecutor(g.Store),
	}
}

// EvaluateRule evaluates a single rule and returns violations
func (e *Evaluator) EvaluateRule(rule *Rule) ([]Violation, error) {
	// Execute SPARQL query
	results, err := e.executor.ExecuteString(rule.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rule %s: %w", rule.ID, err)
	}

	// Check if rule passes or fails
	actualCount := results.Count
	expectedCount := rule.Expect

	// If actual count matches expected, rule passes
	if actualCount == expectedCount {
		return []Violation{}, nil
	}

	// Rule failed - create violations
	violations := make([]Violation, 0)

	// If we have result bindings, create a violation for each
	if len(results.Bindings) > 0 {
		for _, binding := range results.Bindings {
			violation := e.createViolation(rule, binding)
			violations = append(violations, violation)
		}
	} else {
		// No results but expected some - create a generic violation
		violation := Violation{
			Rule:       rule,
			Message:    fmt.Sprintf("Expected %d results, got %d", expectedCount, actualCount),
			Suggestion: rule.Suggestion,
			Details:    make(map[string]any),
		}
		violations = append(violations, violation)
	}

	return violations, nil
}

// createViolation creates a violation from a SPARQL result row
func (e *Evaluator) createViolation(rule *Rule, row map[string]string) Violation {
	violation := Violation{
		Rule:       rule,
		Details:    make(map[string]any),
		Suggestion: rule.Suggestion,
	}

	// Extract module information from results
	for key, value := range row {
		violation.Details[key] = value

		// Try to find the module
		if strings.Contains(key, "module") || key == "?module" || key == "module" {
			if module := e.findModule(value); module != nil {
				violation.Module = module
				violation.FilePath = module.Path
			}
		}

		// Check for other common variables
		if strings.Contains(key, "source") || key == "?source" || key == "source" {
			if module := e.findModule(value); module != nil {
				violation.Module = module
				violation.FilePath = module.Path
			}
		}
	}

	// Build violation message
	violation.Message = e.buildViolationMessage(rule, row)

	return violation
}

// findModule finds a module by URI or path
func (e *Evaluator) findModule(identifier string) *graph.Module {
	// Try direct path lookup
	if module := e.graph.GetModule(identifier); module != nil {
		return module
	}

	// Try URI lookup
	for _, module := range e.graph.Modules {
		if module.URI == identifier {
			return module
		}
		if module.Name == identifier {
			return module
		}
	}

	return nil
}

// buildViolationMessage builds a descriptive violation message
func (e *Evaluator) buildViolationMessage(rule *Rule, row map[string]string) string {
	// Start with rule name
	msg := rule.Name

	// Add details from result row
	if len(row) > 0 {
		details := make([]string, 0)
		for key, value := range row {
			// Clean up variable names
			cleanKey := strings.TrimPrefix(key, "?")
			details = append(details, fmt.Sprintf("%s=%s", cleanKey, value))
		}
		if len(details) > 0 {
			msg += ": " + strings.Join(details, ", ")
		}
	}

	return msg
}

// EvaluateAll evaluates all rules in a rule set
func (e *Evaluator) EvaluateAll(rules []*Rule) ([]Violation, error) {
	allViolations := make([]Violation, 0)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		violations, err := e.EvaluateRule(rule)
		if err != nil {
			return nil, err
		}

		allViolations = append(allViolations, violations...)
	}

	return allViolations, nil
}
