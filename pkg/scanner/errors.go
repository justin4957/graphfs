/*
# Module: pkg/scanner/errors.go
Error collection and reporting for scanner.

Collects errors encountered during scanning and provides summary reporting
for graceful error recovery with partial results.

## Linked Modules
- [scanner](./scanner.go) - Scanner implementation

## Tags
scanner, errors, recovery

## Exports
ErrorCollector, ScanError

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#errors.go> a code:Module ;
    code:name "pkg/scanner/errors.go" ;
    code:description "Error collection and reporting for scanner" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./scanner.go> ;
    code:exports <#ErrorCollector>, <#ScanError> ;
    code:tags "scanner", "errors", "recovery" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"fmt"
	"strings"
	"sync"
)

// ScanError represents an error encountered during scanning
type ScanError struct {
	File    string
	Line    int
	Message string
	Err     error
}

// Error implements the error interface
func (e *ScanError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.File, e.Message)
}

// Unwrap returns the underlying error
func (e *ScanError) Unwrap() error {
	return e.Err
}

// ErrorCollector collects errors encountered during scanning
type ErrorCollector struct {
	errors []ScanError
	mu     sync.Mutex
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]ScanError, 0),
	}
}

// Add adds an error to the collection
func (ec *ErrorCollector) Add(file string, err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = append(ec.errors, ScanError{
		File:    file,
		Message: err.Error(),
		Err:     err,
	})
}

// AddWithLine adds an error with line number to the collection
func (ec *ErrorCollector) AddWithLine(file string, line int, err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = append(ec.errors, ScanError{
		File:    file,
		Line:    line,
		Message: err.Error(),
		Err:     err,
	})
}

// HasErrors returns true if any errors have been collected
func (ec *ErrorCollector) HasErrors() bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors) > 0
}

// Count returns the number of errors collected
func (ec *ErrorCollector) Count() int {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors)
}

// Errors returns all collected errors
func (ec *ErrorCollector) Errors() []ScanError {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]ScanError, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// Report generates a formatted error report
func (ec *ErrorCollector) Report() string {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.errors) == 0 {
		return ""
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\n⚠️  %d error(s) encountered:\n\n", len(ec.errors)))

	// Show first 10 errors
	maxShow := 10
	for i, err := range ec.errors {
		if i >= maxShow {
			remaining := len(ec.errors) - maxShow
			buf.WriteString(fmt.Sprintf("\n... and %d more error(s)\n", remaining))
			break
		}
		buf.WriteString(fmt.Sprintf("  • %s\n", err.Error()))
	}

	return buf.String()
}

// Clear clears all collected errors
func (ec *ErrorCollector) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = make([]ScanError, 0)
}
