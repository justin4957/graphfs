/*
# Module: pkg/doctor/doctor.go
Health check system for GraphFS diagnostics.

Provides health checks for system diagnostics, cache integrity, configuration
validation, and performance benchmarking.

## Linked Modules
- [../cache](../cache/cache.go) - Cache management
- [../parser](../parser/parser.go) - Parser
- [../graph](../graph/graph.go) - Graph builder

## Tags
diagnostics, health-check, troubleshooting

## Exports
CheckStatus, HealthCheck, RunAllChecks

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#doctor.go> a code:Module ;
    code:name "pkg/doctor/doctor.go" ;
    code:description "Health check system for GraphFS diagnostics" ;
    code:language "go" ;
    code:layer "diagnostics" ;
    code:linksTo <../cache/cache.go>, <../parser/parser.go>, <../graph/graph.go> ;
    code:exports <#CheckStatus>, <#HealthCheck>, <#RunAllChecks> ;
    code:tags "diagnostics", "health-check", "troubleshooting" .
<!-- End LinkedDoc RDF -->
*/

package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/justin4957/graphfs/pkg/cache"
	"github.com/justin4957/graphfs/pkg/parser"
)

// CheckStatus represents the status of a health check
type CheckStatus int

const (
	// StatusOK indicates the check passed
	StatusOK CheckStatus = iota
	// StatusWarning indicates a non-critical issue
	StatusWarning
	// StatusError indicates a critical issue
	StatusError
)

// String returns a string representation of the check status
func (s CheckStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarning:
		return "WARNING"
	case StatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// HealthCheck represents a single diagnostic check
type HealthCheck struct {
	Name    string
	Status  CheckStatus
	Message string
	Fix     string
}

// CheckGraphFSVersion checks if GraphFS version is available
func CheckGraphFSVersion(version string) HealthCheck {
	if version == "" {
		version = "unknown"
	}

	return HealthCheck{
		Name:    "GraphFS version",
		Status:  StatusOK,
		Message: fmt.Sprintf("v%s", version),
	}
}

// CheckGoVersion checks if Go version meets minimum requirements
func CheckGoVersion() HealthCheck {
	goVersion := runtime.Version()

	// Parse version (e.g., "go1.23.0")
	if !strings.HasPrefix(goVersion, "go1.") {
		return HealthCheck{
			Name:    "Go version",
			Status:  StatusError,
			Message: fmt.Sprintf("Unexpected Go version format: %s", goVersion),
		}
	}

	// Extract minor version
	parts := strings.Split(goVersion[4:], ".")
	if len(parts) < 2 {
		return HealthCheck{
			Name:    "Go version",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Could not parse Go version: %s", goVersion),
		}
	}

	// Check minimum version (Go 1.21+)
	minor := parts[0]
	if minor < "21" {
		return HealthCheck{
			Name:    "Go version",
			Status:  StatusError,
			Message: fmt.Sprintf("%s (requires Go 1.21+)", goVersion),
			Fix:     "Upgrade Go: https://golang.org/dl/",
		}
	}

	return HealthCheck{
		Name:    "Go version",
		Status:  StatusOK,
		Message: goVersion,
	}
}

// CheckGraphVizInstalled checks if GraphViz is installed (optional)
func CheckGraphVizInstalled() HealthCheck {
	_, err := exec.LookPath("dot")
	if err != nil {
		fix := "Install GraphViz for visualization support"
		if runtime.GOOS == "darwin" {
			fix = "Install: brew install graphviz"
		} else if runtime.GOOS == "linux" {
			fix = "Install: sudo apt-get install graphviz (or equivalent)"
		}

		return HealthCheck{
			Name:    "GraphViz installation",
			Status:  StatusWarning,
			Message: "GraphViz not found (optional for visualizations)",
			Fix:     fix,
		}
	}

	return HealthCheck{
		Name:    "GraphViz installation",
		Status:  StatusOK,
		Message: "GraphViz available",
	}
}

// CheckCacheIntegrity checks if the cache is healthy
func CheckCacheIntegrity(rootPath string) HealthCheck {
	cacheDir := filepath.Join(rootPath, ".graphfs", "cache")

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return HealthCheck{
			Name:    "Cache directory",
			Status:  StatusOK,
			Message: "No cache (will be created on first scan)",
		}
	}

	// Try to open cache
	cacheManager, err := cache.NewManager(cacheDir)
	if err != nil {
		return HealthCheck{
			Name:    "Cache integrity",
			Status:  StatusError,
			Message: fmt.Sprintf("Cache corrupted: %v", err),
			Fix:     "Delete cache directory: rm -rf .graphfs/cache",
		}
	}
	defer cacheManager.Close()

	stats, err := cacheManager.Stats()
	if err != nil {
		return HealthCheck{
			Name:    "Cache integrity",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Could not get cache stats: %v", err),
		}
	}

	return HealthCheck{
		Name:    "Cache integrity",
		Status:  StatusOK,
		Message: fmt.Sprintf("%d cached modules", stats.ModuleCount),
	}
}

// CheckConfigFile checks if the configuration file exists
func CheckConfigFile(rootPath string) HealthCheck {
	configPath := filepath.Join(rootPath, ".graphfs", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return HealthCheck{
			Name:    "Configuration file",
			Status:  StatusWarning,
			Message: "No config file (using defaults)",
			Fix:     "Create config: graphfs init",
		}
	}

	return HealthCheck{
		Name:    "Configuration file",
		Status:  StatusOK,
		Message: "Configuration found",
	}
}

// CheckDiskSpace checks available disk space
func CheckDiskSpace(rootPath string) HealthCheck {
	var stat syscall.Statfs_t
	err := syscall.Statfs(rootPath, &stat)
	if err != nil {
		return HealthCheck{
			Name:    "Disk space",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Could not check disk space: %v", err),
		}
	}

	// Calculate available space in GB
	availableGB := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	if availableGB < 1.0 {
		return HealthCheck{
			Name:    "Disk space",
			Status:  StatusError,
			Message: fmt.Sprintf("Low disk space: %.1f GB available", availableGB),
			Fix:     "Free up disk space",
		}
	}

	if availableGB < 5.0 {
		return HealthCheck{
			Name:    "Disk space",
			Status:  StatusWarning,
			Message: fmt.Sprintf("%.1f GB available (consider freeing space)", availableGB),
		}
	}

	return HealthCheck{
		Name:    "Disk space",
		Status:  StatusOK,
		Message: fmt.Sprintf("%.1f GB available", availableGB),
	}
}

// CheckPermissions checks if the current directory is readable/writable
func CheckPermissions(rootPath string) HealthCheck {
	// Check read permission
	_, err := os.ReadDir(rootPath)
	if err != nil {
		return HealthCheck{
			Name:    "File permissions",
			Status:  StatusError,
			Message: fmt.Sprintf("Cannot read directory: %v", err),
			Fix:     "Check directory permissions",
		}
	}

	// Check write permission by creating a temp file
	testFile := filepath.Join(rootPath, ".graphfs_permission_test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return HealthCheck{
			Name:    "File permissions",
			Status:  StatusError,
			Message: fmt.Sprintf("Cannot write to directory: %v", err),
			Fix:     "Check directory permissions",
		}
	}
	os.Remove(testFile)

	return HealthCheck{
		Name:    "File permissions",
		Status:  StatusOK,
		Message: "Read/write access OK",
	}
}

// CheckPerformance runs a quick performance benchmark
func CheckPerformance() HealthCheck {
	// Quick performance benchmark
	start := time.Now()

	// Parse a small test file
	p := parser.NewParser()
	testCode := `package test
func Example() {
	// Test function
}`
	_, err := p.ParseString(testCode)

	duration := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:    "Parser performance",
			Status:  StatusError,
			Message: fmt.Sprintf("Parser error: %v", err),
		}
	}

	if duration > 100*time.Millisecond {
		return HealthCheck{
			Name:    "Parser performance",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Slow parsing (%v)", duration),
		}
	}

	return HealthCheck{
		Name:    "Parser performance",
		Status:  StatusOK,
		Message: fmt.Sprintf("Parse time: %v", duration),
	}
}

// RunAllChecks runs all health checks
func RunAllChecks(rootPath, version string) []HealthCheck {
	return []HealthCheck{
		CheckGraphFSVersion(version),
		CheckGoVersion(),
		CheckGraphVizInstalled(),
		CheckCacheIntegrity(rootPath),
		CheckConfigFile(rootPath),
		CheckDiskSpace(rootPath),
		CheckPermissions(rootPath),
		CheckPerformance(),
	}
}
