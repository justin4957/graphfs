/*
# Module: cmd/graphfs/cmd_test.go
Tests for CLI commands.

Provides unit and integration tests for all CLI commands.

## Linked Modules
- [root](./root.go) - Root command
- [cmd_init](./cmd_init.go) - Init command
- [cmd_scan](./cmd_scan.go) - Scan command
- [cmd_query](./cmd_query.go) - Query command

## Tags
cli, test, integration

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_test.go> a code:Module ;

	code:name "cmd/graphfs/cmd_test.go" ;
	code:description "Tests for CLI commands" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <./cmd_init.go>, <./cmd_scan.go>, <./cmd_query.go> ;
	code:tags "cli", "test", "integration" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Run init command
	err := runInit(initCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify .graphfs directory was created
	graphfsDir := filepath.Join(tmpDir, ".graphfs")
	if _, err := os.Stat(graphfsDir); os.IsNotExist(err) {
		t.Errorf(".graphfs directory was not created")
	}

	// Verify config file was created
	configFile := filepath.Join(graphfsDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("config.yaml was not created")
	}

	// Verify .graphfsignore was created
	ignoreFile := filepath.Join(tmpDir, ".graphfsignore")
	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		t.Errorf(".graphfsignore was not created")
	}

	// Verify store directory was created
	storeDir := filepath.Join(graphfsDir, "store")
	if _, err := os.Stat(storeDir); os.IsNotExist(err) {
		t.Errorf("store directory was not created")
	}
}

func TestInitCommandExistingDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Initialize twice
	err := runInit(initCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	err = runInit(initCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	// Should not error on re-initialization
}

func TestConfigLoading(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := saveDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load config
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify config values
	if config.Version != 1 {
		t.Errorf("expected version 1, got %d", config.Version)
	}

	if len(config.Scan.Include) == 0 {
		t.Errorf("expected include patterns, got none")
	}

	if config.Query.DefaultLimit != 100 {
		t.Errorf("expected default limit 100, got %d", config.Query.DefaultLimit)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Version != 1 {
		t.Errorf("expected version 1, got %d", config.Version)
	}

	if len(config.Scan.Include) == 0 {
		t.Errorf("expected default include patterns")
	}

	if len(config.Scan.Exclude) == 0 {
		t.Errorf("expected default exclude patterns")
	}

	if config.Scan.MaxFileSize != 1048576 {
		t.Errorf("expected max file size 1048576, got %d", config.Scan.MaxFileSize)
	}
}

func TestScanCommandWithMinimalApp(t *testing.T) {
	// This test requires the examples/minimal-app directory
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Get path to minimal-app
	minimalAppPath := filepath.Join("..", "..", "examples", "minimal-app")
	if _, err := os.Stat(minimalAppPath); os.IsNotExist(err) {
		t.Skip("examples/minimal-app not found")
	}

	// Create temporary directory for output
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "graph.json")

	// Set output flag
	scanOutput = outputFile
	scanValidate = true
	scanStats = true

	// Run scan command
	err := runScan(scanCmd, []string{minimalAppPath})
	if err != nil {
		t.Fatalf("scan command failed: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("output file was not created")
	}
}
