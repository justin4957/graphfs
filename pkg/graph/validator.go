/*
# Module: pkg/graph/validator.go
Graph validation implementation.

Validates graph consistency, detects circular dependencies, and checks for common issues.

## Linked Modules
- [graph](./graph.go) - Graph data structure
- [module](./module.go) - Module data structure

## Tags
graph, validation, consistency

## Exports
Validator, ValidationResult, ValidationError, ValidationWarning, NewValidator

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#validator.go> a code:Module ;
    code:name "pkg/graph/validator.go" ;
    code:description "Graph validation implementation" ;
    code:language "go" ;
    code:layer "graph" ;
    code:linksTo <./graph.go>, <./module.go> ;
    code:exports <#Validator>, <#ValidationResult>, <#ValidationError>, <#ValidationWarning>, <#NewValidator> ;
    code:tags "graph", "validation", "consistency" .
<!-- End LinkedDoc RDF -->
*/

package graph

import (
	"fmt"
	"strings"
)

// Validator validates knowledge graphs
type Validator struct{}

// ValidationResult contains validation errors and warnings
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error
type ValidationError struct {
	Module  string
	Message string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Module  string
	Message string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs comprehensive graph validation
func (v *Validator) Validate(graph *Graph) ValidationResult {
	result := ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Check for required fields
	v.validateRequiredFields(graph, &result)

	// Validate dependencies
	v.validateDependencies(graph, &result)

	// Detect circular dependencies
	v.detectCircularDependencies(graph, &result)

	// Check for duplicate URIs
	v.checkDuplicateURIs(graph, &result)

	// Validate URI format
	v.validateURIFormat(graph, &result)

	// Best practices warnings
	v.checkBestPractices(graph, &result)

	return result
}

// validateRequiredFields checks that modules have required fields
func (v *Validator) validateRequiredFields(graph *Graph, result *ValidationResult) {
	for path, module := range graph.Modules {
		if module.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Module:  path,
				Message: "missing required field: name",
			})
		}

		if module.Description == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: "missing description",
			})
		}

		if module.Language == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: "missing language",
			})
		}
	}
}

// validateDependencies checks that all dependencies exist
func (v *Validator) validateDependencies(graph *Graph, result *ValidationResult) {
	for path, module := range graph.Modules {
		for _, dep := range module.Dependencies {
			if !v.dependencyExists(graph, dep) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Module:  path,
					Message: fmt.Sprintf("dependency not found: %s", dep),
				})
			}
		}
	}
}

// dependencyExists checks if a dependency target exists in the graph
func (v *Validator) dependencyExists(graph *Graph, dep string) bool {
	// Check direct path
	if graph.GetModule(dep) != nil {
		return true
	}

	// Check URI
	for _, module := range graph.Modules {
		if module.URI == dep {
			return true
		}
	}

	// Check name or relative path
	for _, module := range graph.Modules {
		if module.Name == dep || strings.HasSuffix(module.Path, dep) {
			return true
		}
	}

	return false
}

// detectCircularDependencies detects circular dependency chains
func (v *Validator) detectCircularDependencies(graph *Graph, result *ValidationResult) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for path, module := range graph.Modules {
		if !visited[module.URI] {
			if v.hasCycleDFS(module, graph, visited, recStack, []string{}) {
				result.Errors = append(result.Errors, ValidationError{
					Module:  path,
					Message: "circular dependency detected",
				})
			}
		}
	}
}

// hasCycleDFS performs DFS to detect cycles
func (v *Validator) hasCycleDFS(module *Module, graph *Graph, visited, recStack map[string]bool, path []string) bool {
	visited[module.URI] = true
	recStack[module.URI] = true

	for _, dep := range module.Dependencies {
		depModule := graph.GetModule(dep)
		if depModule == nil {
			// Try to find by URI
			for _, mod := range graph.Modules {
				if mod.URI == dep {
					depModule = mod
					break
				}
			}
		}

		if depModule != nil {
			if !visited[depModule.URI] {
				if v.hasCycleDFS(depModule, graph, visited, recStack, append(path, module.URI)) {
					return true
				}
			} else if recStack[depModule.URI] {
				// Found a cycle
				return true
			}
		}
	}

	recStack[module.URI] = false
	return false
}

// checkDuplicateURIs checks for duplicate module URIs
func (v *Validator) checkDuplicateURIs(graph *Graph, result *ValidationResult) {
	uriMap := make(map[string][]string)

	for path, module := range graph.Modules {
		uriMap[module.URI] = append(uriMap[module.URI], path)
	}

	for uri, paths := range uriMap {
		if len(paths) > 1 {
			for _, path := range paths {
				result.Errors = append(result.Errors, ValidationError{
					Module:  path,
					Message: fmt.Sprintf("duplicate URI %s found in: %v", uri, paths),
				})
			}
		}
	}
}

// validateURIFormat checks URI formatting
func (v *Validator) validateURIFormat(graph *Graph, result *ValidationResult) {
	for path, module := range graph.Modules {
		if module.URI == "" {
			result.Errors = append(result.Errors, ValidationError{
				Module:  path,
				Message: "missing URI",
			})
			continue
		}

		// Check basic URI format
		if !strings.HasPrefix(module.URI, "<") || !strings.HasSuffix(module.URI, ">") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: fmt.Sprintf("URI should be wrapped in angle brackets: %s", module.URI),
			})
		}
	}
}

// checkBestPractices checks for best practice violations
func (v *Validator) checkBestPractices(graph *Graph, result *ValidationResult) {
	for path, module := range graph.Modules {
		// Check for missing tags
		if len(module.Tags) == 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: "no tags specified",
			})
		}

		// Check for missing exports
		if len(module.Exports) == 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: "no exports specified",
			})
		}

		// Check for too many dependencies (code smell)
		if len(module.Dependencies) > 10 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Module:  path,
				Message: fmt.Sprintf("high number of dependencies (%d) - consider refactoring", len(module.Dependencies)),
			})
		}
	}
}
