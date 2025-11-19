package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGraphFSVersion(t *testing.T) {
	tests := []struct {
		version string
		wantMsg string
	}{
		{"1.0.0", "v1.0.0"},
		{"", "vunknown"},
		{"2.3.4-beta", "v2.3.4-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			check := CheckGraphFSVersion(tt.version)

			if check.Status != StatusOK {
				t.Errorf("Expected StatusOK, got %v", check.Status)
			}

			if !strings.Contains(check.Message, tt.wantMsg) {
				t.Errorf("Expected message to contain %q, got %q", tt.wantMsg, check.Message)
			}
		})
	}
}

func TestCheckGoVersion(t *testing.T) {
	check := CheckGoVersion()

	// Should always succeed in tests (we're running with supported Go)
	if check.Status == StatusError {
		t.Errorf("Go version check failed: %s", check.Message)
	}

	if check.Name != "Go version" {
		t.Errorf("Expected name 'Go version', got %q", check.Name)
	}
}

func TestCheckGraphVizInstalled(t *testing.T) {
	check := CheckGraphVizInstalled()

	// Check should return OK or Warning (not Error)
	if check.Status == StatusError {
		t.Errorf("GraphViz check should not return Error status")
	}

	if check.Name != "GraphViz installation" {
		t.Errorf("Expected name 'GraphViz installation', got %q", check.Name)
	}
}

func TestCheckCacheIntegrity(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "graphfs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// No cache directory - should be OK
	check := CheckCacheIntegrity(tmpDir)
	if check.Status != StatusOK {
		t.Errorf("Expected StatusOK for missing cache, got %v", check.Status)
	}

	// Create cache directory
	cacheDir := filepath.Join(tmpDir, ".graphfs", "cache")
	os.MkdirAll(cacheDir, 0755)

	// Empty cache directory - might fail or warn
	check = CheckCacheIntegrity(tmpDir)
	// Just verify it doesn't panic
	_ = check.Status
}

func TestCheckConfigFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "graphfs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// No config file
	check := CheckConfigFile(tmpDir)
	if check.Status != StatusWarning {
		t.Errorf("Expected StatusWarning for missing config, got %v", check.Status)
	}

	// Create config file
	configDir := filepath.Join(tmpDir, ".graphfs")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "config.yaml")
	os.WriteFile(configFile, []byte("test: config"), 0644)

	check = CheckConfigFile(tmpDir)
	if check.Status != StatusOK {
		t.Errorf("Expected StatusOK for existing config, got %v", check.Status)
	}
}

func TestCheckDiskSpace(t *testing.T) {
	check := CheckDiskSpace(".")

	// Should not error (unless something is seriously wrong)
	if check.Status == StatusError && !strings.Contains(check.Message, "Low disk space") {
		t.Errorf("Unexpected error in disk space check: %s", check.Message)
	}

	if check.Name != "Disk space" {
		t.Errorf("Expected name 'Disk space', got %q", check.Name)
	}
}

func TestCheckPermissions(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "graphfs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	check := CheckPermissions(tmpDir)
	if check.Status != StatusOK {
		t.Errorf("Expected StatusOK for temp dir permissions, got %v: %s", check.Status, check.Message)
	}
}

func TestCheckPerformance(t *testing.T) {
	check := CheckPerformance()

	// Parser should work
	if check.Status == StatusError {
		t.Errorf("Parser performance check failed: %s", check.Message)
	}

	if check.Name != "Parser performance" {
		t.Errorf("Expected name 'Parser performance', got %q", check.Name)
	}
}

func TestRunAllChecks(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "graphfs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	checks := RunAllChecks(tmpDir, "test-version")

	// Should have 8 checks
	expectedChecks := 8
	if len(checks) != expectedChecks {
		t.Errorf("Expected %d checks, got %d", expectedChecks, len(checks))
	}

	// Verify all checks have names
	for i, check := range checks {
		if check.Name == "" {
			t.Errorf("Check %d has empty name", i)
		}
	}
}

func TestCheckStatus_String(t *testing.T) {
	tests := []struct {
		status CheckStatus
		want   string
	}{
		{StatusOK, "OK"},
		{StatusWarning, "WARNING"},
		{StatusError, "ERROR"},
		{CheckStatus(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("Status.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
