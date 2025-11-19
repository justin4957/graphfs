package scanner

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorCollector_Add(t *testing.T) {
	ec := NewErrorCollector()

	if ec.HasErrors() {
		t.Error("New collector should not have errors")
	}

	ec.Add("file1.go", errors.New("test error 1"))
	ec.Add("file2.go", errors.New("test error 2"))

	if !ec.HasErrors() {
		t.Error("Collector should have errors")
	}

	if ec.Count() != 2 {
		t.Errorf("Expected 2 errors, got %d", ec.Count())
	}
}

func TestErrorCollector_AddWithLine(t *testing.T) {
	ec := NewErrorCollector()

	ec.AddWithLine("file.go", 42, errors.New("syntax error"))

	errs := ec.Errors()
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}

	if errs[0].Line != 42 {
		t.Errorf("Expected line 42, got %d", errs[0].Line)
	}

	expected := "file.go:42:"
	if !strings.Contains(errs[0].Error(), expected) {
		t.Errorf("Expected error to contain %q, got %q", expected, errs[0].Error())
	}
}

func TestErrorCollector_Report(t *testing.T) {
	ec := NewErrorCollector()

	// Empty collector
	report := ec.Report()
	if report != "" {
		t.Error("Empty collector should return empty report")
	}

	// Add errors
	for i := 0; i < 15; i++ {
		ec.Add("file.go", errors.New("error"))
	}

	report = ec.Report()
	if !strings.Contains(report, "⚠️") {
		t.Error("Report should contain warning emoji")
	}

	if !strings.Contains(report, "15 error(s)") {
		t.Error("Report should mention 15 errors")
	}

	// Should show "and 5 more" since we show max 10
	if !strings.Contains(report, "and 5 more") {
		t.Error("Report should indicate more errors")
	}
}

func TestErrorCollector_Clear(t *testing.T) {
	ec := NewErrorCollector()

	ec.Add("file.go", errors.New("error"))
	if !ec.HasErrors() {
		t.Error("Should have errors before clear")
	}

	ec.Clear()
	if ec.HasErrors() {
		t.Error("Should not have errors after clear")
	}

	if ec.Count() != 0 {
		t.Errorf("Expected 0 errors after clear, got %d", ec.Count())
	}
}

func TestErrorCollector_Concurrent(t *testing.T) {
	ec := NewErrorCollector()

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			ec.Add("file.go", errors.New("concurrent error"))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if ec.Count() != 10 {
		t.Errorf("Expected 10 errors from concurrent adds, got %d", ec.Count())
	}
}

func TestScanError_Error(t *testing.T) {
	// Without line number
	err := ScanError{
		File:    "test.go",
		Message: "parse error",
		Err:     errors.New("underlying error"),
	}

	expected := "test.go: parse error"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}

	// With line number
	err.Line = 42
	expected = "test.go:42: parse error"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestScanError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := ScanError{
		File:    "test.go",
		Message: "parse error",
		Err:     underlying,
	}

	if err.Unwrap() != underlying {
		t.Error("Unwrap should return underlying error")
	}
}
