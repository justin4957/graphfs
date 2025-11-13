package scanner

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"Go file", "main.go", "Go"},
		{"Python file", "script.py", "Python"},
		{"JavaScript file", "app.js", "JavaScript"},
		{"TypeScript file", "app.ts", "TypeScript"},
		{"TypeScript React", "Component.tsx", "TypeScript"},
		{"Java file", "Main.java", "Java"},
		{"Rust file", "main.rs", "Rust"},
		{"C file", "program.c", "C"},
		{"C++ file", "program.cpp", "C++"},
		{"C++ header", "header.hpp", "C++"},
		{"C# file", "Program.cs", "C#"},
		{"Ruby file", "script.rb", "Ruby"},
		{"PHP file", "index.php", "PHP"},
		{"Swift file", "App.swift", "Swift"},
		{"Kotlin file", "Main.kt", "Kotlin"},
		{"Scala file", "Main.scala", "Scala"},
		{"Unknown extension", "file.xyz", "unknown"},
		{"No extension", "README", "unknown"},
		{"Path with directory", "src/main.go", "Go"},
		{"Complex path", "/usr/local/src/app/main.go", "Go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(tt.filePath)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestRegisterLanguage(t *testing.T) {
	// Register custom language
	RegisterLanguage("custom", "CustomLang", []string{".custom", ".cst"})

	tests := []struct {
		filePath string
		want     string
	}{
		{"test.custom", "CustomLang"},
		{"test.cst", "CustomLang"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.filePath)
		if got != tt.want {
			t.Errorf("DetectLanguage(%q) after registration = %q, want %q", tt.filePath, got, tt.want)
		}
	}
}

func TestSupportedLanguages(t *testing.T) {
	languages := SupportedLanguages()

	if len(languages) == 0 {
		t.Error("SupportedLanguages() returned empty list")
	}

	// Check for some expected languages
	expectedLangs := []string{"Go", "Python", "JavaScript", "TypeScript", "Java", "Rust"}
	found := make(map[string]bool)

	for _, lang := range languages {
		found[lang] = true
	}

	for _, expected := range expectedLangs {
		if !found[expected] {
			t.Errorf("Expected language %q not found in supported languages", expected)
		}
	}
}

func TestGetLanguage(t *testing.T) {
	lang, ok := GetLanguage("go")
	if !ok {
		t.Fatal("GetLanguage(\"go\") not found")
	}

	if lang.Name != "Go" {
		t.Errorf("Language name = %q, want %q", lang.Name, "Go")
	}

	if len(lang.Extensions) == 0 {
		t.Error("Language extensions is empty")
	}

	// Test unknown language
	_, ok = GetLanguage("nonexistent")
	if ok {
		t.Error("GetLanguage(\"nonexistent\") should return false")
	}
}

func BenchmarkDetectLanguage(b *testing.B) {
	testPaths := []string{
		"main.go",
		"script.py",
		"app.js",
		"Component.tsx",
		"Main.java",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		DetectLanguage(path)
	}
}
