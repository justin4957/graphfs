package scanner_test

import (
	"fmt"
	"log"

	"github.com/justin4957/graphfs/pkg/scanner"
)

// Example demonstrates basic scanner usage
func Example() {
	s := scanner.NewScanner()

	result, err := s.Scan("../../examples/minimal-app", scanner.DefaultScanOptions())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d files\n", result.TotalFiles)
	fmt.Printf("Total size: %d bytes\n", result.TotalBytes)

	// Output will vary based on file sizes, so we just check it ran
	// Output:
	// Found 7 files
}

// Example_languageDetection demonstrates language detection
func Example_languageDetection() {
	languages := []string{
		scanner.DetectLanguage("main.go"),
		scanner.DetectLanguage("script.py"),
		scanner.DetectLanguage("app.js"),
		scanner.DetectLanguage("Component.tsx"),
	}

	for _, lang := range languages {
		fmt.Println(lang)
	}

	// Output:
	// Go
	// Python
	// JavaScript
	// TypeScript
}

// Example_customOptions demonstrates scanning with custom options
func Example_customOptions() {
	s := scanner.NewScanner()

	opts := scanner.ScanOptions{
		ExcludePatterns: []string{"**/test/**"},
		MaxFileSize:     100 * 1024, // 100KB
		UseDefaults:     true,
		Concurrent:      false,
	}

	result, err := s.Scan("../../examples/minimal-app", opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Files: %d\n", result.TotalFiles)

	// Output:
	// Files: 7
}

// Example_ignorePatterns demonstrates ignore pattern matching
func Example_ignorePatterns() {
	matcher := scanner.NewIgnoreMatcher([]string{
		".git",
		"node_modules",
		"*.pyc",
	})

	paths := []string{
		"src/main.go",
		".git/config",
		"node_modules/package/index.js",
		"script.pyc",
	}

	for _, path := range paths {
		if matcher.ShouldIgnore(path) {
			fmt.Printf("%s: ignored\n", path)
		} else {
			fmt.Printf("%s: included\n", path)
		}
	}

	// Output:
	// src/main.go: included
	// .git/config: ignored
	// node_modules/package/index.js: ignored
	// script.pyc: ignored
}

// Example_scanFile demonstrates scanning a single file
func Example_scanFile() {
	s := scanner.NewScanner()

	fileInfo, err := s.ScanFile("../../examples/minimal-app/main.go")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Language: %s\n", fileInfo.Language)
	fmt.Printf("Has LinkedDoc: %v\n", fileInfo.HasLinkedDoc)

	// Output:
	// Language: Go
	// Has LinkedDoc: true
}

// Example_filterByLanguage demonstrates filtering files by language
func Example_filterByLanguage() {
	s := scanner.NewScanner()

	result, err := s.Scan("../../examples/minimal-app", scanner.DefaultScanOptions())
	if err != nil {
		log.Fatal(err)
	}

	goFiles := 0
	for _, file := range result.Files {
		if file.Language == "Go" {
			goFiles++
		}
	}

	fmt.Printf("Go files: %d\n", goFiles)

	// Output:
	// Go files: 7
}

// Example_linkedDocDetection demonstrates LinkedDoc detection
func Example_linkedDocDetection() {
	s := scanner.NewScanner()

	result, err := s.Scan("../../examples/minimal-app", scanner.DefaultScanOptions())
	if err != nil {
		log.Fatal(err)
	}

	withLinkedDoc := 0
	for _, file := range result.Files {
		if file.HasLinkedDoc {
			withLinkedDoc++
		}
	}

	fmt.Printf("Files with LinkedDoc: %d/%d\n", withLinkedDoc, result.TotalFiles)

	// Output:
	// Files with LinkedDoc: 7/7
}
