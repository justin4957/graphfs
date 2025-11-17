/*
# Module: pkg/rules/rule.go
Rule data structures and types for architecture validation.

Defines the core types for architectural rules, violations, and validation results.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure

## Tags
rules, validation, architecture

## Exports
Rule, Severity, Violation, ValidationResult

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#rule.go> a code:Module ;
    code:name "pkg/rules/rule.go" ;
    code:description "Rule data structures and types for architecture validation" ;
    code:language "go" ;
    code:layer "rules" ;
    code:linksTo <../graph/graph.go> ;
    code:exports <#Rule>, <#Severity>, <#Violation>, <#ValidationResult> ;
    code:tags "rules", "validation", "architecture" .
<!-- End LinkedDoc RDF -->
*/

package rules

import (
	"github.com/justin4957/graphfs/pkg/graph"
)

// Severity represents the severity level of a rule violation
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Rule represents an architectural validation rule
type Rule struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Severity    Severity `yaml:"severity"`
	Pattern     string   `yaml:"pattern"`    // SPARQL query
	Expect      int      `yaml:"expect"`     // Expected result count
	Enabled     bool     `yaml:"enabled"`    // Whether rule is enabled
	Tags        []string `yaml:"tags"`       // Rule tags for filtering
	Suggestion  string   `yaml:"suggestion"` // Default suggestion for violations
}

// Violation represents a rule violation
type Violation struct {
	Rule       *Rule          // The rule that was violated
	Module     *graph.Module  // The module involved in the violation
	Message    string         // Violation message
	FilePath   string         // File path where violation occurred
	LineNumber int            // Line number (0 if unknown)
	Suggestion string         // Suggested fix
	Details    map[string]any // Additional details from SPARQL results
}

// ValidationResult contains the results of rule validation
type ValidationResult struct {
	Violations   []Violation // All violations found
	PassedRules  []*Rule     // Rules that passed
	FailedRules  []*Rule     // Rules that failed
	SkippedRules []*Rule     // Rules that were skipped
	TotalRules   int         // Total number of rules evaluated
	ErrorCount   int         // Number of error-level violations
	WarningCount int         // Number of warning-level violations
	InfoCount    int         // Number of info-level violations
	Duration     int64       // Execution duration in milliseconds
}

// RuleSet represents a collection of rules
type RuleSet struct {
	Version string  `yaml:"version"`
	Name    string  `yaml:"name"`
	Rules   []*Rule `yaml:"rules"`
}

// HasErrors returns true if there are any error-level violations
func (r *ValidationResult) HasErrors() bool {
	return r.ErrorCount > 0
}

// HasWarnings returns true if there are any warning-level violations
func (r *ValidationResult) HasWarnings() bool {
	return r.WarningCount > 0
}

// Success returns true if there are no error-level violations
func (r *ValidationResult) Success() bool {
	return r.ErrorCount == 0
}

// GetViolationsBySeverity returns violations filtered by severity
func (r *ValidationResult) GetViolationsBySeverity(severity Severity) []Violation {
	violations := make([]Violation, 0)
	for _, v := range r.Violations {
		if v.Rule.Severity == severity {
			violations = append(violations, v)
		}
	}
	return violations
}

// GetViolationsByRule returns violations grouped by rule ID
func (r *ValidationResult) GetViolationsByRule() map[string][]Violation {
	grouped := make(map[string][]Violation)
	for _, v := range r.Violations {
		grouped[v.Rule.ID] = append(grouped[v.Rule.ID], v)
	}
	return grouped
}
