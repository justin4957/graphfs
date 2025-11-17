/*
# Module: pkg/rules/parser.go
YAML rule parser for loading and validating rule definitions.

Parses rule definitions from YAML files and validates their structure.

## Linked Modules
- [./rule](./rule.go) - Rule data structures

## Tags
rules, parser, yaml

## Exports
Parser, ParseRules, ParseRuleSet

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#parser.go> a code:Module ;
    code:name "pkg/rules/parser.go" ;
    code:description "YAML rule parser for loading and validating rule definitions" ;
    code:language "go" ;
    code:layer "rules" ;
    code:linksTo <./rule.go> ;
    code:exports <#Parser>, <#ParseRules>, <#ParseRuleSet> ;
    code:tags "rules", "parser", "yaml" .
<!-- End LinkedDoc RDF -->
*/

package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser parses rule definitions from YAML
type Parser struct{}

// NewParser creates a new rule parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses rules from a YAML file
func (p *Parser) ParseFile(filePath string) (*RuleSet, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses rules from YAML data
func (p *Parser) Parse(data []byte) (*RuleSet, error) {
	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the rule set
	if err := p.validate(&ruleSet); err != nil {
		return nil, err
	}

	// Set default values
	p.setDefaults(&ruleSet)

	return &ruleSet, nil
}

// validate validates the rule set structure
func (p *Parser) validate(ruleSet *RuleSet) error {
	if ruleSet.Version == "" {
		return fmt.Errorf("missing version field")
	}

	if len(ruleSet.Rules) == 0 {
		return fmt.Errorf("no rules defined")
	}

	// Validate each rule
	ids := make(map[string]bool)
	for i, rule := range ruleSet.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule %d: missing ID", i)
		}

		if ids[rule.ID] {
			return fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		ids[rule.ID] = true

		if rule.Name == "" {
			return fmt.Errorf("rule %s: missing name", rule.ID)
		}

		if rule.Pattern == "" {
			return fmt.Errorf("rule %s: missing pattern", rule.ID)
		}

		if rule.Severity == "" {
			return fmt.Errorf("rule %s: missing severity", rule.ID)
		}

		if rule.Severity != SeverityError && rule.Severity != SeverityWarning && rule.Severity != SeverityInfo {
			return fmt.Errorf("rule %s: invalid severity '%s' (must be error, warning, or info)", rule.ID, rule.Severity)
		}
	}

	return nil
}

// setDefaults sets default values for rules
func (p *Parser) setDefaults(ruleSet *RuleSet) {
	for _, rule := range ruleSet.Rules {
		// Enable by default if not specified
		if rule.Enabled == false && rule.Severity != "" {
			rule.Enabled = true
		}

		// Default expect to 0 (no violations)
		if rule.Expect == 0 && rule.Pattern != "" {
			rule.Expect = 0
		}
	}
}

// ParseRules is a convenience function to parse rules from a file
func ParseRules(filePath string) (*RuleSet, error) {
	parser := NewParser()
	return parser.ParseFile(filePath)
}

// ParseRuleSet is a convenience function to parse rules from YAML data
func ParseRuleSet(data []byte) (*RuleSet, error) {
	parser := NewParser()
	return parser.Parse(data)
}
