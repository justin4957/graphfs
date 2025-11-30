/*
# Module: pkg/shadow/shadow_test.go
Tests for shadow file system functionality.

## Tags
shadow, test

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#shadow_test.go> a code:Module ;
    code:name "pkg/shadow/shadow_test.go" ;
    code:description "Tests for shadow file system functionality" ;
    code:language "go" ;
    code:layer "test" ;
    code:linksTo <./shadow.go>, <./entry.go>, <./index.go> ;
    code:tags "shadow", "test" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewShadowFS(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create shadow file system
	config := DefaultConfig()
	shadowFS, err := NewShadowFS(tmpDir, config)
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	// Verify paths
	if shadowFS.RootPath() != tmpDir {
		t.Errorf("Expected root path %s, got %s", tmpDir, shadowFS.RootPath())
	}

	expectedShadowPath := filepath.Join(tmpDir, DefaultShadowDir)
	if shadowFS.ShadowPath() != expectedShadowPath {
		t.Errorf("Expected shadow path %s, got %s", expectedShadowPath, shadowFS.ShadowPath())
	}
}

func TestShadowFSInitialize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	// Initialize
	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	// Verify shadow directory exists
	if _, err := os.Stat(shadowFS.ShadowPath()); os.IsNotExist(err) {
		t.Error("Shadow directory was not created")
	}

	// Verify index file exists
	indexPath := filepath.Join(shadowFS.ShadowPath(), "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("Index file was not created")
	}
}

func TestShadowFSGetShadowPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	// Test with relative path
	sourcePath := filepath.Join(tmpDir, "pkg", "module.go")
	shadowPath, err := shadowFS.GetShadowPath(sourcePath)
	if err != nil {
		t.Fatalf("Failed to get shadow path: %v", err)
	}

	expectedPath := filepath.Join(shadowFS.ShadowPath(), "pkg", "module.go"+ShadowExtension)
	if shadowPath != expectedPath {
		t.Errorf("Expected shadow path %s, got %s", expectedPath, shadowPath)
	}
}

func TestShadowFSSetAndGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	// Create entry
	sourcePath := filepath.Join(tmpDir, "test.go")
	entry := NewAutoEntry("test.go")
	entry.SetModule("<#test.go>", "test.go", "Test module", "go", "test", []string{"testing"})
	entry.AddTriple("<#test.go>", "code:description", "Test module", SourceAuto)

	// Set entry
	if err := shadowFS.Set(sourcePath, entry); err != nil {
		t.Fatalf("Failed to set shadow entry: %v", err)
	}

	// Get entry
	retrieved, err := shadowFS.Get(sourcePath)
	if err != nil {
		t.Fatalf("Failed to get shadow entry: %v", err)
	}

	// Verify entry
	if retrieved.SourcePath != entry.SourcePath {
		t.Errorf("Expected source path %s, got %s", entry.SourcePath, retrieved.SourcePath)
	}

	if retrieved.Module == nil {
		t.Fatal("Expected module to be set")
	}

	if retrieved.Module.Name != "test.go" {
		t.Errorf("Expected module name 'test.go', got %s", retrieved.Module.Name)
	}

	if len(retrieved.Triples) != 1 {
		t.Errorf("Expected 1 triple, got %d", len(retrieved.Triples))
	}
}

func TestShadowFSExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	sourcePath := filepath.Join(tmpDir, "test.go")

	// Should not exist initially
	if shadowFS.Exists(sourcePath) {
		t.Error("Entry should not exist initially")
	}

	// Create entry
	entry := NewAutoEntry("test.go")
	if err := shadowFS.Set(sourcePath, entry); err != nil {
		t.Fatalf("Failed to set shadow entry: %v", err)
	}

	// Should exist now
	if !shadowFS.Exists(sourcePath) {
		t.Error("Entry should exist after being set")
	}
}

func TestShadowFSDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	sourcePath := filepath.Join(tmpDir, "test.go")

	// Create entry
	entry := NewAutoEntry("test.go")
	if err := shadowFS.Set(sourcePath, entry); err != nil {
		t.Fatalf("Failed to set shadow entry: %v", err)
	}

	// Verify it exists
	if !shadowFS.Exists(sourcePath) {
		t.Fatal("Entry should exist after being set")
	}

	// Delete entry
	if err := shadowFS.Delete(sourcePath); err != nil {
		t.Fatalf("Failed to delete shadow entry: %v", err)
	}

	// Verify it's deleted
	if shadowFS.Exists(sourcePath) {
		t.Error("Entry should not exist after deletion")
	}
}

func TestShadowFSMerge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultConfig()
	config.PreserveManual = true
	shadowFS, err := NewShadowFS(tmpDir, config)
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	sourcePath := filepath.Join(tmpDir, "test.go")

	// Create initial entry with manual annotation
	entry1 := NewManualEntry("test.go")
	entry1.AddAnnotation("reviewed", true, "tester")
	entry1.AddTriple("<#test.go>", "code:manual", "value1", SourceManual)

	if err := shadowFS.Set(sourcePath, entry1); err != nil {
		t.Fatalf("Failed to set initial shadow entry: %v", err)
	}

	// Create new entry with auto data
	entry2 := NewAutoEntry("test.go")
	entry2.SetModule("<#test.go>", "test.go", "Updated description", "go", "test", []string{"updated"})
	entry2.AddTriple("<#test.go>", "code:auto", "value2", SourceAuto)

	// Merge entries
	if err := shadowFS.Merge(sourcePath, entry2); err != nil {
		t.Fatalf("Failed to merge shadow entry: %v", err)
	}

	// Get merged entry
	merged, err := shadowFS.Get(sourcePath)
	if err != nil {
		t.Fatalf("Failed to get merged shadow entry: %v", err)
	}

	// Verify manual annotation preserved
	if len(merged.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(merged.Annotations))
	}

	// Verify module info updated
	if merged.Module == nil || merged.Module.Description != "Updated description" {
		t.Error("Module info should be updated from new entry")
	}

	// Verify both triples exist
	if len(merged.Triples) < 2 {
		t.Errorf("Expected at least 2 triples, got %d", len(merged.Triples))
	}

	// Verify source is mixed
	if merged.Source != SourceMixed {
		t.Errorf("Expected source to be mixed, got %s", merged.Source)
	}
}

func TestShadowFSList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	// Create multiple entries
	paths := []string{
		filepath.Join(tmpDir, "pkg", "a.go"),
		filepath.Join(tmpDir, "pkg", "b.go"),
		filepath.Join(tmpDir, "cmd", "main.go"),
	}

	for i, path := range paths {
		entry := NewAutoEntry(path)
		entry.SetModule("<#test>", path, "Test module", "go", "test", nil)
		if err := shadowFS.Set(path, entry); err != nil {
			t.Fatalf("Failed to set entry %d: %v", i, err)
		}
	}

	// List all entries
	entries, err := shadowFS.List()
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

// Entry tests

func TestNewEntry(t *testing.T) {
	entry := NewEntry("test.go", SourceAuto)

	if entry.SourcePath != "test.go" {
		t.Errorf("Expected source path 'test.go', got %s", entry.SourcePath)
	}

	if entry.Source != SourceAuto {
		t.Errorf("Expected source 'auto', got %s", entry.Source)
	}

	if entry.Version != ShadowVersion {
		t.Errorf("Expected version %s, got %s", ShadowVersion, entry.Version)
	}

	if entry.CreatedAt.IsZero() {
		t.Error("Created time should not be zero")
	}
}

func TestEntryAddTriple(t *testing.T) {
	entry := NewAutoEntry("test.go")

	entry.AddTriple("<#test>", "code:name", "Test", SourceAuto)
	entry.AddTriple("<#test>", "code:description", "A test", SourceAuto)

	// Try to add duplicate
	entry.AddTriple("<#test>", "code:name", "Test", SourceAuto)

	if len(entry.Triples) != 2 {
		t.Errorf("Expected 2 triples (no duplicates), got %d", len(entry.Triples))
	}
}

func TestEntryAddDependency(t *testing.T) {
	entry := NewAutoEntry("test.go")

	entry.AddDependency("linksTo", "other.go", SourceAuto)
	entry.AddDependency("linksTo", "another.go", SourceAuto)

	// Try to add duplicate
	entry.AddDependency("linksTo", "other.go", SourceAuto)

	if len(entry.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies (no duplicates), got %d", len(entry.Dependencies))
	}
}

func TestEntryAnnotation(t *testing.T) {
	entry := NewManualEntry("test.go")

	// Add annotation
	entry.AddAnnotation("reviewed", true, "tester")

	// Get annotation
	val, ok := entry.GetAnnotation("reviewed")
	if !ok {
		t.Fatal("Expected annotation to exist")
	}

	if val != true {
		t.Errorf("Expected annotation value true, got %v", val)
	}

	// Update annotation
	entry.AddAnnotation("reviewed", false, "tester2")

	val, _ = entry.GetAnnotation("reviewed")
	if val != false {
		t.Errorf("Expected updated annotation value false, got %v", val)
	}

	// Verify still only one annotation
	if len(entry.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(entry.Annotations))
	}
}

func TestEntryValidate(t *testing.T) {
	// Valid entry
	entry := NewAutoEntry("test.go")
	if err := entry.Validate(); err != nil {
		t.Errorf("Valid entry should not fail validation: %v", err)
	}

	// Invalid: missing source path
	entry2 := &Entry{
		Version: ShadowVersion,
		Source:  SourceAuto,
	}
	if err := entry2.Validate(); err == nil {
		t.Error("Entry without source path should fail validation")
	}

	// Invalid: missing version
	entry3 := &Entry{
		SourcePath: "test.go",
		Source:     SourceAuto,
	}
	if err := entry3.Validate(); err == nil {
		t.Error("Entry without version should fail validation")
	}
}

func TestEntryHasManualData(t *testing.T) {
	// Auto entry without manual data
	entry1 := NewAutoEntry("test.go")
	if entry1.HasManualData() {
		t.Error("Auto entry without manual data should return false")
	}

	// Auto entry with manual triple
	entry2 := NewAutoEntry("test.go")
	entry2.AddTriple("<#test>", "code:manual", "value", SourceManual)
	if !entry2.HasManualData() {
		t.Error("Entry with manual triple should return true")
	}

	// Auto entry with annotation
	entry3 := NewAutoEntry("test.go")
	entry3.AddAnnotation("key", "value", "author")
	if !entry3.HasManualData() {
		t.Error("Entry with annotation should return true")
	}

	// Manual entry
	entry4 := NewManualEntry("test.go")
	if !entry4.HasManualData() {
		t.Error("Manual entry should return true")
	}
}

// Index tests

func TestNewIndex(t *testing.T) {
	idx := NewIndex()

	if idx.Version != ShadowVersion {
		t.Errorf("Expected version %s, got %s", ShadowVersion, idx.Version)
	}

	if len(idx.Entries) != 0 {
		t.Error("New index should have no entries")
	}
}

func TestIndexAddAndGet(t *testing.T) {
	idx := NewIndex()

	entry := NewAutoEntry("test.go")
	entry.SetModule("<#test>", "test.go", "Test module", "go", "api", []string{"testing", "api"})
	entry.AddConcept("authentication")

	idx.Add("test.go", entry)

	// Verify entry exists
	indexEntry, ok := idx.Get("test.go")
	if !ok {
		t.Fatal("Expected entry to exist")
	}

	if indexEntry.Language != "go" {
		t.Errorf("Expected language 'go', got %s", indexEntry.Language)
	}

	if indexEntry.Layer != "api" {
		t.Errorf("Expected layer 'api', got %s", indexEntry.Layer)
	}

	// Verify inverted indexes
	if paths := idx.GetByLanguage("go"); len(paths) != 1 {
		t.Errorf("Expected 1 entry for language 'go', got %d", len(paths))
	}

	if paths := idx.GetByLayer("api"); len(paths) != 1 {
		t.Errorf("Expected 1 entry for layer 'api', got %d", len(paths))
	}

	if paths := idx.GetByTag("testing"); len(paths) != 1 {
		t.Errorf("Expected 1 entry for tag 'testing', got %d", len(paths))
	}

	if paths := idx.GetByConcept("authentication"); len(paths) != 1 {
		t.Errorf("Expected 1 entry for concept 'authentication', got %d", len(paths))
	}
}

func TestIndexRemove(t *testing.T) {
	idx := NewIndex()

	entry := NewAutoEntry("test.go")
	entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag1"})

	idx.Add("test.go", entry)

	// Verify entry exists
	if _, ok := idx.Get("test.go"); !ok {
		t.Fatal("Entry should exist")
	}

	// Remove entry
	idx.Remove("test.go")

	// Verify entry removed
	if _, ok := idx.Get("test.go"); ok {
		t.Error("Entry should not exist after removal")
	}

	// Verify inverted indexes cleaned up
	if paths := idx.GetByLanguage("go"); len(paths) != 0 {
		t.Errorf("Expected 0 entries for language 'go' after removal, got %d", len(paths))
	}
}

func TestIndexSearch(t *testing.T) {
	idx := NewIndex()

	// Add test entries
	entries := []struct {
		path     string
		language string
		layer    string
		tags     []string
	}{
		{"pkg/api/handler.go", "go", "api", []string{"http", "api"}},
		{"pkg/api/router.go", "go", "api", []string{"http", "routing"}},
		{"pkg/db/store.go", "go", "data", []string{"database", "storage"}},
		{"web/app.ts", "typescript", "frontend", []string{"ui", "app"}},
	}

	for _, e := range entries {
		entry := NewAutoEntry(e.path)
		entry.SetModule("<#"+e.path+">", e.path, "Test", e.language, e.layer, e.tags)
		idx.Add(e.path, entry)
	}

	// Search by language
	results := idx.Search(SearchQuery{Language: "go"})
	if len(results) != 3 {
		t.Errorf("Expected 3 Go entries, got %d", len(results))
	}

	// Search by layer
	results = idx.Search(SearchQuery{Layer: "api"})
	if len(results) != 2 {
		t.Errorf("Expected 2 API entries, got %d", len(results))
	}

	// Search by tag
	results = idx.Search(SearchQuery{Tags: []string{"http"}})
	if len(results) != 2 {
		t.Errorf("Expected 2 entries with 'http' tag, got %d", len(results))
	}

	// Combined search
	results = idx.Search(SearchQuery{
		Language: "go",
		Layer:    "api",
		Tags:     []string{"http"},
	})
	if len(results) != 2 {
		t.Errorf("Expected 2 entries matching all criteria, got %d", len(results))
	}

	// Search with limit
	results = idx.Search(SearchQuery{Language: "go", Limit: 2})
	if len(results) != 2 {
		t.Errorf("Expected 2 entries with limit, got %d", len(results))
	}
}

func TestIndexSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and populate index
	idx := NewIndex()
	entry := NewAutoEntry("test.go")
	entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag1", "tag2"})
	idx.Add("test.go", entry)

	// Save index
	indexPath := filepath.Join(tmpDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("Failed to save index: %v", err)
	}

	// Load index into new instance
	idx2 := NewIndex()
	if err := idx2.Load(indexPath); err != nil {
		t.Fatalf("Failed to load index: %v", err)
	}

	// Verify loaded data
	if idx2.Count() != 1 {
		t.Errorf("Expected 1 entry, got %d", idx2.Count())
	}

	indexEntry, ok := idx2.Get("test.go")
	if !ok {
		t.Fatal("Expected entry to exist in loaded index")
	}

	if indexEntry.Language != "go" {
		t.Errorf("Expected language 'go', got %s", indexEntry.Language)
	}
}

func TestIndexStatistics(t *testing.T) {
	idx := NewIndex()

	// Add various entries
	autoEntry := NewAutoEntry("auto.go")
	autoEntry.SetModule("<#auto>", "auto.go", "Auto", "go", "api", []string{"tag1"})
	autoEntry.AddTriple("<#auto>", "code:test", "value", SourceAuto)
	idx.Add("auto.go", autoEntry)

	manualEntry := NewManualEntry("manual.go")
	manualEntry.SetModule("<#manual>", "manual.go", "Manual", "go", "data", []string{"tag2"})
	idx.Add("manual.go", manualEntry)

	mixedEntry := NewEntry("mixed.go", SourceMixed)
	mixedEntry.SetModule("<#mixed>", "mixed.go", "Mixed", "typescript", "frontend", nil)
	idx.Add("mixed.go", mixedEntry)

	// Get statistics
	stats := idx.Statistics()

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}

	if stats.AutoEntries != 1 {
		t.Errorf("Expected 1 auto entry, got %d", stats.AutoEntries)
	}

	if stats.ManualEntries != 1 {
		t.Errorf("Expected 1 manual entry, got %d", stats.ManualEntries)
	}

	if stats.MixedEntries != 1 {
		t.Errorf("Expected 1 mixed entry, got %d", stats.MixedEntries)
	}

	if stats.LanguageCount["go"] != 2 {
		t.Errorf("Expected 2 Go entries, got %d", stats.LanguageCount["go"])
	}

	if stats.LanguageCount["typescript"] != 1 {
		t.Errorf("Expected 1 TypeScript entry, got %d", stats.LanguageCount["typescript"])
	}
}

func TestIndexListMethods(t *testing.T) {
	idx := NewIndex()

	// Add test entries
	entry1 := NewAutoEntry("a.go")
	entry1.SetModule("<#a>", "a.go", "A", "go", "api", []string{"tag1", "tag2"})
	entry1.AddConcept("concept1")
	idx.Add("a.go", entry1)

	entry2 := NewAutoEntry("b.py")
	entry2.SetModule("<#b>", "b.py", "B", "python", "data", []string{"tag2", "tag3"})
	entry2.AddConcept("concept2")
	idx.Add("b.py", entry2)

	// Test ListTags
	tags := idx.ListTags()
	if len(tags) != 3 {
		t.Errorf("Expected 3 unique tags, got %d", len(tags))
	}

	// Test ListLanguages
	languages := idx.ListLanguages()
	if len(languages) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(languages))
	}

	// Test ListLayers
	layers := idx.ListLayers()
	if len(layers) != 2 {
		t.Errorf("Expected 2 layers, got %d", len(layers))
	}

	// Test ListConcepts
	concepts := idx.ListConcepts()
	if len(concepts) != 2 {
		t.Errorf("Expected 2 concepts, got %d", len(concepts))
	}
}

func TestIndexClear(t *testing.T) {
	idx := NewIndex()

	// Add entry
	entry := NewAutoEntry("test.go")
	entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag1"})
	idx.Add("test.go", entry)

	if idx.Count() != 1 {
		t.Fatal("Expected 1 entry before clear")
	}

	// Clear index
	idx.Clear()

	if idx.Count() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", idx.Count())
	}

	// Verify inverted indexes cleared
	if len(idx.ByLanguage) != 0 {
		t.Error("Language index should be empty after clear")
	}
}

// Benchmark tests

func BenchmarkIndexAdd(b *testing.B) {
	idx := NewIndex()

	for i := 0; i < b.N; i++ {
		entry := NewAutoEntry("test.go")
		entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag1", "tag2"})
		idx.Add("test.go", entry)
	}
}

func BenchmarkIndexSearch(b *testing.B) {
	idx := NewIndex()

	// Populate index
	for i := 0; i < 1000; i++ {
		entry := NewAutoEntry("test.go")
		entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag1", "tag2"})
		idx.Add("test.go", entry)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		idx.Search(SearchQuery{Language: "go", Layer: "api"})
	}
}

// Helper function tests

func TestIsShadowFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.shadow.json", true},
		{"path/to/file.shadow.json", true},
		{"test.json", false},
		{"test.go", false},
		{".shadow.json", false}, // Too short
	}

	for _, tt := range tests {
		result := isShadowFile(tt.path)
		if result != tt.expected {
			t.Errorf("isShadowFile(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

// Concurrency tests

func TestShadowFSConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shadow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	shadowFS, err := NewShadowFS(tmpDir, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create shadow file system: %v", err)
	}

	if err := shadowFS.Initialize(); err != nil {
		t.Fatalf("Failed to initialize shadow file system: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			path := filepath.Join(tmpDir, "test.go")
			entry := NewAutoEntry("test.go")
			entry.SetModule("<#test>", "test.go", "Test", "go", "api", nil)

			// Set
			_ = shadowFS.Set(path, entry)

			// Get
			_, _ = shadowFS.Get(path)

			// Exists
			_ = shadowFS.Exists(path)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

func TestIndexConcurrentAccess(t *testing.T) {
	idx := NewIndex()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			entry := NewAutoEntry("test.go")
			entry.SetModule("<#test>", "test.go", "Test", "go", "api", []string{"tag"})

			// Add
			idx.Add("test.go", entry)

			// Get
			_, _ = idx.Get("test.go")

			// Search
			_ = idx.Search(SearchQuery{Language: "go"})

			// Stats
			_ = idx.Statistics()

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}
