/*
# Module: pkg/scanner/language.go
Language detection for source files.

Detects programming language based on file extension.
Provides extensible language registry for adding new languages.

## Linked Modules
None (utility module with no dependencies)

## Tags
scanner, language-detection, utility

## Exports
Language, DetectLanguage, RegisterLanguage, SupportedLanguages

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#language.go> a code:Module ;
    code:name "pkg/scanner/language.go" ;
    code:description "Language detection for source files" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:exports <#Language>, <#DetectLanguage>, <#RegisterLanguage>, <#SupportedLanguages> ;
    code:tags "scanner", "language-detection", "utility" ;
    code:isLeaf true .

<#DetectLanguage> a code:Function ;
    code:name "DetectLanguage" ;
    code:description "Detects programming language from file path" .

<#RegisterLanguage> a code:Function ;
    code:name "RegisterLanguage" ;
    code:description "Registers a language with extensions" .

<#SupportedLanguages> a code:Function ;
    code:name "SupportedLanguages" ;
    code:description "Returns list of supported languages" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"path/filepath"
	"strings"
)

// Language represents a programming language
type Language struct {
	Name       string
	Extensions []string
}

// Built-in language registry
var languageRegistry = map[string]*Language{
	"go": {
		Name:       "Go",
		Extensions: []string{".go"},
	},
	"python": {
		Name:       "Python",
		Extensions: []string{".py", ".pyw"},
	},
	"javascript": {
		Name:       "JavaScript",
		Extensions: []string{".js", ".mjs", ".cjs"},
	},
	"typescript": {
		Name:       "TypeScript",
		Extensions: []string{".ts", ".tsx"},
	},
	"java": {
		Name:       "Java",
		Extensions: []string{".java"},
	},
	"rust": {
		Name:       "Rust",
		Extensions: []string{".rs"},
	},
	"c": {
		Name:       "C",
		Extensions: []string{".c", ".h"},
	},
	"cpp": {
		Name:       "C++",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h++"},
	},
	"csharp": {
		Name:       "C#",
		Extensions: []string{".cs"},
	},
	"ruby": {
		Name:       "Ruby",
		Extensions: []string{".rb"},
	},
	"php": {
		Name:       "PHP",
		Extensions: []string{".php"},
	},
	"swift": {
		Name:       "Swift",
		Extensions: []string{".swift"},
	},
	"kotlin": {
		Name:       "Kotlin",
		Extensions: []string{".kt", ".kts"},
	},
	"scala": {
		Name:       "Scala",
		Extensions: []string{".scala"},
	},
}

// Extension to language mapping (built from registry)
var extensionMap map[string]string

func init() {
	buildExtensionMap()
}

// buildExtensionMap builds the extension to language mapping
func buildExtensionMap() {
	extensionMap = make(map[string]string)
	for langKey, lang := range languageRegistry {
		for _, ext := range lang.Extensions {
			extensionMap[strings.ToLower(ext)] = langKey
		}
	}
}

// DetectLanguage detects the programming language from a file path
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return "unknown"
	}

	if langKey, ok := extensionMap[ext]; ok {
		return languageRegistry[langKey].Name
	}

	return "unknown"
}

// RegisterLanguage registers a new language or updates an existing one
func RegisterLanguage(key string, name string, extensions []string) {
	languageRegistry[key] = &Language{
		Name:       name,
		Extensions: extensions,
	}
	buildExtensionMap()
}

// SupportedLanguages returns a list of all supported languages
func SupportedLanguages() []string {
	languages := make([]string, 0, len(languageRegistry))
	for _, lang := range languageRegistry {
		languages = append(languages, lang.Name)
	}
	return languages
}

// GetLanguage returns the Language struct for a given key
func GetLanguage(key string) (*Language, bool) {
	lang, ok := languageRegistry[key]
	return lang, ok
}
