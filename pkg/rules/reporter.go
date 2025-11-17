/*
# Module: pkg/rules/reporter.go
Violation reporter for formatting and displaying rule violations.

Provides multiple output formats for rule violations including text, JSON, and JUnit XML.

## Linked Modules
- [./rule](./rule.go) - Rule data structures

## Tags
rules, reporter, output

## Exports
Reporter, FormatText, FormatJSON

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#reporter.go> a code:Module ;
    code:name "pkg/rules/reporter.go" ;
    code:description "Violation reporter for formatting and displaying rule violations" ;
    code:language "go" ;
    code:layer "rules" ;
    code:linksTo <./rule.go> ;
    code:exports <#Reporter>, <#FormatText>, <#FormatJSON> ;
    code:tags "rules", "reporter", "output" .
<!-- End LinkedDoc RDF -->
*/

package rules

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// OutputFormat represents the output format for violations
type OutputFormat string

const (
	FormatText  OutputFormat = "text"
	FormatJSON  OutputFormat = "json"
	FormatJUnit OutputFormat = "junit"
)

// Reporter formats and reports rule violations
type Reporter struct {
	format OutputFormat
}

// NewReporter creates a new reporter with the specified format
func NewReporter(format OutputFormat) *Reporter {
	return &Reporter{
		format: format,
	}
}

// Report formats and outputs the validation result
func (r *Reporter) Report(result *ValidationResult) string {
	switch r.format {
	case FormatJSON:
		return r.formatJSON(result)
	case FormatJUnit:
		return r.formatJUnit(result)
	default:
		return r.formatText(result)
	}
}

// formatText formats the result as human-readable text
func (r *Reporter) formatText(result *ValidationResult) string {
	var output strings.Builder

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	// Header
	cyan.Fprintf(&output, "📋 Architecture Validation Results\n\n")

	// Group violations by rule
	violationsByRule := result.GetViolationsByRule()

	// Report passed rules
	if len(result.PassedRules) > 0 {
		green.Fprintf(&output, "✅ Passed Rules (%d):\n", len(result.PassedRules))
		for _, rule := range result.PassedRules {
			output.WriteString(fmt.Sprintf("  • %s\n", rule.Name))
		}
		output.WriteString("\n")
	}

	// Report failed rules with violations
	if len(result.FailedRules) > 0 {
		// Sort by severity
		errorRules := make([]*Rule, 0)
		warningRules := make([]*Rule, 0)
		infoRules := make([]*Rule, 0)

		for _, rule := range result.FailedRules {
			switch rule.Severity {
			case SeverityError:
				errorRules = append(errorRules, rule)
			case SeverityWarning:
				warningRules = append(warningRules, rule)
			case SeverityInfo:
				infoRules = append(infoRules, rule)
			}
		}

		// Report errors
		if len(errorRules) > 0 {
			red.Fprintf(&output, "❌ Failed Rules - Errors (%d):\n", len(errorRules))
			for _, rule := range errorRules {
				violations := violationsByRule[rule.ID]
				output.WriteString(fmt.Sprintf("  • %s (%d violations)\n", rule.Name, len(violations)))
				r.formatViolations(&output, violations, "    ")
			}
			output.WriteString("\n")
		}

		// Report warnings
		if len(warningRules) > 0 {
			yellow.Fprintf(&output, "⚠️  Failed Rules - Warnings (%d):\n", len(warningRules))
			for _, rule := range warningRules {
				violations := violationsByRule[rule.ID]
				output.WriteString(fmt.Sprintf("  • %s (%d violations)\n", rule.Name, len(violations)))
				r.formatViolations(&output, violations, "    ")
			}
			output.WriteString("\n")
		}

		// Report info
		if len(infoRules) > 0 {
			cyan.Fprintf(&output, "ℹ️  Failed Rules - Info (%d):\n", len(infoRules))
			for _, rule := range infoRules {
				violations := violationsByRule[rule.ID]
				output.WriteString(fmt.Sprintf("  • %s (%d violations)\n", rule.Name, len(violations)))
				r.formatViolations(&output, violations, "    ")
			}
			output.WriteString("\n")
		}
	}

	// Summary
	cyan.Fprintf(&output, "📊 Summary:\n")
	output.WriteString(fmt.Sprintf("  • Total Rules: %d\n", result.TotalRules))
	output.WriteString(fmt.Sprintf("  • Passed: %d\n", len(result.PassedRules)))
	output.WriteString(fmt.Sprintf("  • Failed: %d\n", len(result.FailedRules)))

	if result.ErrorCount > 0 {
		red.Fprintf(&output, "  • Errors: %d\n", result.ErrorCount)
	}
	if result.WarningCount > 0 {
		yellow.Fprintf(&output, "  • Warnings: %d\n", result.WarningCount)
	}
	if result.InfoCount > 0 {
		output.WriteString(fmt.Sprintf("  • Info: %d\n", result.InfoCount))
	}

	output.WriteString(fmt.Sprintf("  • Duration: %dms\n", result.Duration))

	// Overall status
	output.WriteString("\n")
	if result.Success() {
		green.Fprintf(&output, "✅ Validation PASSED\n")
	} else {
		red.Fprintf(&output, "❌ Validation FAILED\n")
	}

	return output.String()
}

// formatViolations formats individual violations
func (r *Reporter) formatViolations(output *strings.Builder, violations []Violation, indent string) {
	for _, v := range violations {
		if v.FilePath != "" {
			if v.LineNumber > 0 {
				output.WriteString(fmt.Sprintf("%s- %s:%d - %s\n", indent, v.FilePath, v.LineNumber, v.Message))
			} else {
				output.WriteString(fmt.Sprintf("%s- %s - %s\n", indent, v.FilePath, v.Message))
			}
		} else {
			output.WriteString(fmt.Sprintf("%s- %s\n", indent, v.Message))
		}

		if v.Suggestion != "" {
			output.WriteString(fmt.Sprintf("%s  💡 %s\n", indent, v.Suggestion))
		}
	}
}

// formatJSON formats the result as JSON
func (r *Reporter) formatJSON(result *ValidationResult) string {
	type jsonViolation struct {
		RuleID     string         `json:"rule_id"`
		RuleName   string         `json:"rule_name"`
		Severity   Severity       `json:"severity"`
		Message    string         `json:"message"`
		FilePath   string         `json:"file_path,omitempty"`
		LineNumber int            `json:"line_number,omitempty"`
		Suggestion string         `json:"suggestion,omitempty"`
		Details    map[string]any `json:"details,omitempty"`
	}

	type jsonResult struct {
		TotalRules   int             `json:"total_rules"`
		PassedRules  int             `json:"passed_rules"`
		FailedRules  int             `json:"failed_rules"`
		ErrorCount   int             `json:"error_count"`
		WarningCount int             `json:"warning_count"`
		InfoCount    int             `json:"info_count"`
		Success      bool            `json:"success"`
		Duration     int64           `json:"duration_ms"`
		Violations   []jsonViolation `json:"violations"`
	}

	violations := make([]jsonViolation, 0, len(result.Violations))
	for _, v := range result.Violations {
		violations = append(violations, jsonViolation{
			RuleID:     v.Rule.ID,
			RuleName:   v.Rule.Name,
			Severity:   v.Rule.Severity,
			Message:    v.Message,
			FilePath:   v.FilePath,
			LineNumber: v.LineNumber,
			Suggestion: v.Suggestion,
			Details:    v.Details,
		})
	}

	jsonRes := jsonResult{
		TotalRules:   result.TotalRules,
		PassedRules:  len(result.PassedRules),
		FailedRules:  len(result.FailedRules),
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		InfoCount:    result.InfoCount,
		Success:      result.Success(),
		Duration:     result.Duration,
		Violations:   violations,
	}

	data, _ := json.MarshalIndent(jsonRes, "", "  ")
	return string(data)
}

// formatJUnit formats the result as JUnit XML
func (r *Reporter) formatJUnit(result *ValidationResult) string {
	type junitFailure struct {
		Message string `xml:"message,attr"`
		Type    string `xml:"type,attr"`
		Content string `xml:",chardata"`
	}

	type testCase struct {
		XMLName   xml.Name      `xml:"testcase"`
		Name      string        `xml:"name,attr"`
		ClassName string        `xml:"classname,attr"`
		Time      float64       `xml:"time,attr"`
		Failure   *junitFailure `xml:"failure,omitempty"`
	}

	type testSuite struct {
		XMLName  xml.Name   `xml:"testsuite"`
		Name     string     `xml:"name,attr"`
		Tests    int        `xml:"tests,attr"`
		Failures int        `xml:"failures,attr"`
		Errors   int        `xml:"errors,attr"`
		Time     float64    `xml:"time,attr"`
		Cases    []testCase `xml:"testcase"`
	}

	cases := make([]testCase, 0)

	// Add passed rules as test cases
	for _, rule := range result.PassedRules {
		cases = append(cases, testCase{
			Name:      rule.Name,
			ClassName: "architecture.rules",
			Time:      0,
		})
	}

	// Add failed rules as test cases with failures
	violationsByRule := result.GetViolationsByRule()
	for _, rule := range result.FailedRules {
		violations := violationsByRule[rule.ID]
		failureContent := ""
		for _, v := range violations {
			failureContent += fmt.Sprintf("%s\n", v.Message)
			if v.FilePath != "" {
				failureContent += fmt.Sprintf("  File: %s\n", v.FilePath)
			}
			if v.Suggestion != "" {
				failureContent += fmt.Sprintf("  Suggestion: %s\n", v.Suggestion)
			}
		}

		cases = append(cases, testCase{
			Name:      rule.Name,
			ClassName: "architecture.rules",
			Time:      0,
			Failure: &junitFailure{
				Message: fmt.Sprintf("%d violations", len(violations)),
				Type:    string(rule.Severity),
				Content: failureContent,
			},
		})
	}

	suite := testSuite{
		Name:     "Architecture Rules",
		Tests:    result.TotalRules,
		Failures: len(result.FailedRules),
		Errors:   0,
		Time:     float64(result.Duration) / 1000.0,
		Cases:    cases,
	}

	data, _ := xml.MarshalIndent(suite, "", "  ")
	return xml.Header + string(data)
}

// ReportViolationsByRule reports violations grouped by rule
func (r *Reporter) ReportViolationsByRule(violations []Violation) string {
	grouped := make(map[string][]Violation)
	rules := make(map[string]*Rule)

	for _, v := range violations {
		grouped[v.Rule.ID] = append(grouped[v.Rule.ID], v)
		rules[v.Rule.ID] = v.Rule
	}

	// Sort rule IDs
	ruleIDs := make([]string, 0, len(grouped))
	for id := range grouped {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)

	var output strings.Builder
	for _, id := range ruleIDs {
		rule := rules[id]
		ruleViolations := grouped[id]

		output.WriteString(fmt.Sprintf("\n%s (%d violations):\n", rule.Name, len(ruleViolations)))
		for _, v := range ruleViolations {
			output.WriteString(fmt.Sprintf("  - %s\n", v.Message))
		}
	}

	return output.String()
}
