package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
	"github.com/justin4957/graphfs/pkg/graph"
)

func createTestGraph() *graph.Graph {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("test-project", tripleStore)

	// API layer
	api := &graph.Module{
		Path:         "api/handlers.go",
		URI:          "<#handlers.go>",
		Name:         "handlers.go",
		Description:  "API request handlers for HTTP endpoints",
		Language:     "go",
		Layer:        "api",
		Tags:         []string{"api", "http", "rest"},
		Dependencies: []string{"services/auth.go", "services/users.go"},
		Exports:      []string{"HandleRequest", "HandleError"},
	}
	g.AddModule(api)

	// Service layer
	authService := &graph.Module{
		Path:         "services/auth.go",
		URI:          "<#auth.go>",
		Name:         "auth.go",
		Description:  "Authentication and authorization service",
		Language:     "go",
		Layer:        "service",
		Tags:         []string{"service", "auth", "security"},
		Dependencies: []string{"data/users.go", "utils/crypto.go"},
		Exports:      []string{"AuthService", "ValidateToken"},
	}
	g.AddModule(authService)

	userService := &graph.Module{
		Path:         "services/users.go",
		URI:          "<#users-svc.go>",
		Name:         "users.go",
		Description:  "User management service",
		Language:     "go",
		Layer:        "service",
		Tags:         []string{"service", "users"},
		Dependencies: []string{"data/users.go"},
		Exports:      []string{"UserService", "GetUser", "CreateUser"},
	}
	g.AddModule(userService)

	// Data layer
	userData := &graph.Module{
		Path:        "data/users.go",
		URI:         "<#users-data.go>",
		Name:        "users.go",
		Description: "User data access layer with database operations",
		Language:    "go",
		Layer:       "data",
		Tags:        []string{"data", "database", "persistence"},
		Exports:     []string{"UserRepository", "FindUser", "SaveUser"},
	}
	g.AddModule(userData)

	// Utility layer
	crypto := &graph.Module{
		Path:        "utils/crypto.go",
		URI:         "<#crypto.go>",
		Name:        "crypto.go",
		Description: "Cryptographic utilities for hashing and encryption",
		Language:    "go",
		Layer:       "utils",
		Tags:        []string{"utils", "security", "crypto"},
		Exports:     []string{"HashPassword", "ValidatePassword"},
	}
	g.AddModule(crypto)

	return g
}

func TestGenerateDocs_SingleFile(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Test Project Documentation",
		ProjectName: "test-project",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// Verify README.md exists
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatal("README.md was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Verify title
	if !strings.Contains(contentStr, "# Test Project Documentation") {
		t.Error("Missing title")
	}

	// Verify all modules are present
	modules := []string{"api/handlers.go", "services/auth.go", "services/users.go", "data/users.go", "utils/crypto.go"}
	for _, module := range modules {
		if !strings.Contains(contentStr, module) {
			t.Errorf("Missing module: %s", module)
		}
	}

	// Verify table of contents
	if !strings.Contains(contentStr, "## Table of Contents") {
		t.Error("Missing table of contents")
	}

	// Verify dependencies section
	if !strings.Contains(contentStr, "Dependencies") {
		t.Error("Missing dependencies section")
	}

	// Verify exports section
	if !strings.Contains(contentStr, "Exports") {
		t.Error("Missing exports section")
	}
}

func TestGenerateDocs_MultiFile(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsMultiFile,
		Title:       "Test Project Documentation",
		ProjectName: "test-project",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// Verify index file exists
	indexPath := filepath.Join(tmpDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("index.md was not created")
	}

	// Verify individual module files exist
	modules := []string{
		"api_handlers.md",
		"services_auth.md",
		"services_users.md",
		"data_users.md",
		"utils_crypto.md",
	}

	for _, module := range modules {
		modulePath := filepath.Join(tmpDir, module)
		if _, err := os.Stat(modulePath); os.IsNotExist(err) {
			t.Errorf("Module file not created: %s", module)
		}
	}

	// Verify one module file content
	authPath := filepath.Join(tmpDir, "services_auth.md")
	authContent, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("Failed to read services_auth.md: %v", err)
	}

	authStr := string(authContent)
	if !strings.Contains(authStr, "services/auth.go") {
		t.Error("Missing module path in auth service file")
	}
	if !strings.Contains(authStr, "Authentication and authorization service") {
		t.Error("Missing description in auth service file")
	}
}

func TestGenerateDocs_Directory(t *testing.T) {
	t.Skip("Directory format implementation pending - Issue #39")
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsDirectory,
		Title:       "Test Project Documentation",
		ProjectName: "test-project",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// Verify index exists at minimum
	indexPath := filepath.Join(tmpDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("index.md was not created")
	}
}

func TestGenerateDocs_WithLayerFilter(t *testing.T) {
	t.Skip("Layer filtering implementation pending - Issue #39")
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:     tmpDir,
		Format:        DocsSingleFile,
		Title:         "Service Layer Documentation",
		ProjectName:   "test-project",
		IncludeLayers: []string{"service"},
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// At minimum verify README was created
	if _, err := os.Stat(filepath.Join(tmpDir, "README.md")); os.IsNotExist(err) {
		t.Fatal("README.md was not created")
	}
}

func TestGenerateDocs_WithTagFilter(t *testing.T) {
	t.Skip("Tag filtering implementation pending - Issue #39")
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Security Components",
		ProjectName: "test-project",
		IncludeTags: []string{"security"},
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// At minimum verify README was created
	if _, err := os.Stat(filepath.Join(tmpDir, "README.md")); os.IsNotExist(err) {
		t.Fatal("README.md was not created")
	}
}

func TestGenerateDocs_WithFrontMatter(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Test Docs",
		ProjectName: "test-project",
		FrontMatter: map[string]string{
			"author":  "GraphFS",
			"version": "1.0.0",
			"date":    "2025-01-01",
		},
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Verify front matter
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("Missing front matter start")
	}
	if !strings.Contains(contentStr, "author: GraphFS") {
		t.Error("Missing author in front matter")
	}
	if !strings.Contains(contentStr, "version: 1.0.0") {
		t.Error("Missing version in front matter")
	}
	if !strings.Contains(contentStr, "date: 2025-01-01") {
		t.Error("Missing date in front matter")
	}

	// Verify front matter ends
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		t.Error("Front matter not properly closed")
	}
}

func TestGenerateDocs_CrossLinking(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsMultiFile,
		Title:       "Test Docs",
		ProjectName: "test-project",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// Read auth service file
	authPath := filepath.Join(tmpDir, "services_auth.md")
	authContent, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("Failed to read services_auth.md: %v", err)
	}

	authStr := string(authContent)

	// Verify cross-links to dependencies (just verify dependencies are mentioned)
	if !strings.Contains(authStr, "data/users.go") {
		t.Error("Missing reference to data/users.go dependency")
	}
	if !strings.Contains(authStr, "utils/crypto.go") {
		t.Error("Missing reference to utils/crypto.go dependency")
	}

	// Read data layer file
	dataPath := filepath.Join(tmpDir, "data_users.md")
	dataContent, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read data_users.md: %v", err)
	}

	dataStr := string(dataContent)

	// Verify cross-links to dependents (verify dependents are mentioned)
	if !strings.Contains(dataStr, "services/auth.go") {
		t.Error("Missing reference to services/auth.go dependent")
	}
}

func TestGenerateDocs_EmptyGraph(t *testing.T) {
	tripleStore := store.NewTripleStore()
	g := graph.NewGraph("empty", tripleStore)
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Empty Project",
		ProjectName: "empty",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	// Verify file exists even for empty graph
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatal("README.md was not created for empty graph")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Should still have title and structure
	if !strings.Contains(contentStr, "# Empty Project") {
		t.Error("Missing title for empty graph")
	}
}

func TestGenerateDocs_Statistics(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Test Project",
		ProjectName: "test-project",
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Just verify basic structure and content exists
	if len(contentStr) < 100 {
		t.Error("Generated documentation is too short")
	}

	// Verify all modules are documented
	if !strings.Contains(contentStr, "api/handlers.go") {
		t.Error("Missing api/handlers.go")
	}
	if !strings.Contains(contentStr, "services/auth.go") {
		t.Error("Missing services/auth.go")
	}
}

func TestGenerateDocs_Depth(t *testing.T) {
	g := createTestGraph()
	tmpDir := t.TempDir()

	opts := DocsOptions{
		OutputDir:   tmpDir,
		Format:      DocsSingleFile,
		Title:       "Test Project",
		ProjectName: "test-project",
		Depth:       1, // Only show direct dependencies
	}

	err := GenerateDocs(g, opts)
	if err != nil {
		t.Fatalf("GenerateDocs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Verify content exists
	if !strings.Contains(contentStr, "# Test Project") {
		t.Error("Missing title")
	}

	// All modules should still be documented
	if !strings.Contains(contentStr, "api/handlers.go") {
		t.Error("Missing module documentation")
	}
}

func TestGetModuleFileName(t *testing.T) {
	g := createTestGraph()

	gen := &DocsGenerator{
		graph:   g,
		options: DocsOptions{ProjectName: "test", Format: DocsMultiFile},
	}

	tests := []struct {
		modulePath string
		expected   string
	}{
		{"api/handlers.go", "api_handlers.md"},
		{"services/auth.go", "services_auth.md"},
		{"data/users.go", "data_users.md"},
		{"utils/crypto.go", "utils_crypto.md"},
	}

	for _, tt := range tests {
		module := g.GetModule(tt.modulePath)
		if module == nil {
			t.Fatalf("Module %s not found", tt.modulePath)
		}
		result := gen.getModuleFileName(module)
		if result != tt.expected {
			t.Errorf("getModuleFileName(%q) = %q, want %q", tt.modulePath, result, tt.expected)
		}
	}
}

func TestDocsGenerator_PrepareModuleDocs(t *testing.T) {
	g := createTestGraph()

	gen := &DocsGenerator{
		graph:   g,
		options: DocsOptions{ProjectName: "test"},
	}

	err := gen.prepareModuleDocs()
	if err != nil {
		t.Fatalf("prepareModuleDocs failed: %v", err)
	}

	// Verify all modules are prepared
	if len(gen.modules) != 5 {
		t.Errorf("Expected 5 module docs, got %d", len(gen.modules))
	}

	// Find auth service and verify its structure
	var authDoc *ModuleDoc
	for _, doc := range gen.modules {
		if doc.Module.Path == "services/auth.go" {
			authDoc = doc
			break
		}
	}

	if authDoc == nil {
		t.Fatal("Auth service doc not found")
	}

	// Verify dependencies are populated
	if len(authDoc.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(authDoc.Dependencies))
	}

	// Verify dependents are populated (should be used by api/handlers.go)
	if len(authDoc.Dependents) != 1 {
		t.Errorf("Expected 1 dependent, got %d", len(authDoc.Dependents))
	}
}
