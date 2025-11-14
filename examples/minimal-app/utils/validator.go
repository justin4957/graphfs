/*
# Module: utils/validator.go
Data validation utilities.

Provides validation functions for common data types like email, phone, etc.
Used by models and services for input validation.

## Linked Modules
None (utility module with no dependencies)

## Tags
utility, validation, data-quality

## Exports
ValidateEmail, ValidationError, NewValidationError, PermissionError, NewPermissionError

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<#utils/validator.go> a code:Module ;
    code:name "utils/validator.go" ;
    code:description "Data validation utilities" ;
    code:language "go" ;
    code:layer "utility" ;
    code:exports <#ValidateEmail>, <#ValidationError>, <#NewValidationError>, <#PermissionError>, <#NewPermissionError> ;
    code:tags "utility", "validation", "data-quality" ;
    code:isLeaf true .

<#ValidateEmail> a code:Function ;
    code:name "ValidateEmail" ;
    code:description "Validates email format" ;
    code:returns "bool" .

<#ValidationError> a code:Type ;
    code:name "ValidationError" ;
    code:kind "struct" ;
    code:description "Error for validation failures" ;
    code:implements "error" .

<#NewValidationError> a code:Function ;
    code:name "NewValidationError" ;
    code:description "Creates new validation error" ;
    code:returns <#ValidationError> .

<#PermissionError> a code:Type ;
    code:name "PermissionError" ;
    code:kind "struct" ;
    code:description "Error for permission failures" ;
    code:implements "error" .

<#NewPermissionError> a code:Function ;
    code:name "NewPermissionError" ;
    code:description "Creates new permission error" ;
    code:returns <#PermissionError> .
<!-- End LinkedDoc RDF -->
*/

package utils

import (
	"fmt"
	"regexp"
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// PermissionError represents a permission error
type PermissionError struct {
	Message string
}

// Error implements the error interface
func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission error: %s", e.Message)
}

// NewPermissionError creates a new permission error
func NewPermissionError(message string) error {
	return &PermissionError{Message: message}
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	// Stubbed email validation - basic regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
