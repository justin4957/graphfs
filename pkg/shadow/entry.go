/*
# Module: pkg/shadow/entry.go
Shadow entry data structure for storing file metadata.

Represents the semantic metadata for a single source file, including
RDF triples, annotations, and provenance information.

## Linked Modules
- [shadow](./shadow.go) - Shadow file system manager

## Tags
shadow, entry, metadata, rdf

## Exports
Entry, NewEntry, LoadEntry, EntrySource

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#entry.go> a code:Module ;
    code:name "pkg/shadow/entry.go" ;
    code:description "Shadow entry data structure for storing file metadata" ;
    code:language "go" ;
    code:layer "shadow" ;
    code:linksTo <./shadow.go> ;
    code:exports <#Entry>, <#NewEntry>, <#LoadEntry>, <#EntrySource> ;
    code:tags "shadow", "entry", "metadata", "rdf" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// EntrySource indicates the origin of shadow metadata
type EntrySource string

const (
	// SourceAuto indicates metadata was auto-generated from parsing
	SourceAuto EntrySource = "auto"

	// SourceManual indicates metadata was manually annotated
	SourceManual EntrySource = "manual"

	// SourceMixed indicates metadata has both auto and manual components
	SourceMixed EntrySource = "mixed"
)

// Triple represents an RDF triple in the shadow file
type Triple struct {
	Subject   string      `json:"subject"`
	Predicate string      `json:"predicate"`
	Object    string      `json:"object"`
	Source    EntrySource `json:"source,omitempty"`
}

// Module represents structured module information
type Module struct {
	URI         string   `json:"uri"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Language    string   `json:"language,omitempty"`
	Layer       string   `json:"layer,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Relationship represents a directed relationship between modules
type Relationship struct {
	Type   string      `json:"type"`
	Target string      `json:"target"`
	Source EntrySource `json:"source,omitempty"`
}

// Annotation represents a manual annotation
type Annotation struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Author    string      `json:"author,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
	UpdatedAt time.Time   `json:"updated_at,omitempty"`
}

// Entry represents a shadow file entry for a source file
type Entry struct {
	// Metadata
	Version    string      `json:"version"`
	SourcePath string      `json:"source_path"`
	SourceHash string      `json:"source_hash,omitempty"`
	Source     EntrySource `json:"source"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`

	// Module information (structured)
	Module *Module `json:"module,omitempty"`

	// Relationships (structured)
	Dependencies []Relationship `json:"dependencies,omitempty"`
	Dependents   []Relationship `json:"dependents,omitempty"`
	Exports      []string       `json:"exports,omitempty"`
	Calls        []string       `json:"calls,omitempty"`

	// Raw RDF triples (for full semantic data)
	Triples []Triple `json:"triples,omitempty"`

	// Manual annotations
	Annotations []Annotation `json:"annotations,omitempty"`

	// Concepts and semantic tags
	Concepts []string `json:"concepts,omitempty"`

	// Custom properties for extensibility
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// NewEntry creates a new shadow entry
func NewEntry(sourcePath string, source EntrySource) *Entry {
	now := time.Now()
	return &Entry{
		Version:    ShadowVersion,
		SourcePath: sourcePath,
		Source:     source,
		CreatedAt:  now,
		UpdatedAt:  now,
		Properties: make(map[string]interface{}),
	}
}

// NewAutoEntry creates a new auto-generated shadow entry
func NewAutoEntry(sourcePath string) *Entry {
	return NewEntry(sourcePath, SourceAuto)
}

// NewManualEntry creates a new manual shadow entry
func NewManualEntry(sourcePath string) *Entry {
	return NewEntry(sourcePath, SourceManual)
}

// LoadEntry loads a shadow entry from a file
func LoadEntry(path string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read shadow file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse shadow file: %w", err)
	}

	return &entry, nil
}

// Save writes the entry to a file
func (e *Entry) Save(path string, prettyPrint bool) error {
	e.UpdatedAt = time.Now()

	var data []byte
	var err error

	if prettyPrint {
		data, err = json.MarshalIndent(e, "", "  ")
	} else {
		data, err = json.Marshal(e)
	}

	if err != nil {
		return fmt.Errorf("failed to serialize shadow entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write shadow file: %w", err)
	}

	return nil
}

// Validate checks if the entry is valid
func (e *Entry) Validate() error {
	if e.SourcePath == "" {
		return fmt.Errorf("source path is required")
	}

	if e.Version == "" {
		return fmt.Errorf("version is required")
	}

	if e.Source == "" {
		return fmt.Errorf("source is required")
	}

	return nil
}

// SetModule sets the module information
func (e *Entry) SetModule(uri, name, description, language, layer string, tags []string) {
	e.Module = &Module{
		URI:         uri,
		Name:        name,
		Description: description,
		Language:    language,
		Layer:       layer,
		Tags:        tags,
	}
	e.UpdatedAt = time.Now()
}

// AddTriple adds an RDF triple to the entry
func (e *Entry) AddTriple(subject, predicate, object string, source EntrySource) {
	// Check for duplicates
	for _, t := range e.Triples {
		if t.Subject == subject && t.Predicate == predicate && t.Object == object {
			return
		}
	}

	e.Triples = append(e.Triples, Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
		Source:    source,
	})
	e.UpdatedAt = time.Now()
}

// AddDependency adds a dependency relationship
func (e *Entry) AddDependency(relType, target string, source EntrySource) {
	// Check for duplicates
	for _, d := range e.Dependencies {
		if d.Type == relType && d.Target == target {
			return
		}
	}

	e.Dependencies = append(e.Dependencies, Relationship{
		Type:   relType,
		Target: target,
		Source: source,
	})
	e.UpdatedAt = time.Now()
}

// AddDependent adds a dependent relationship
func (e *Entry) AddDependent(relType, target string, source EntrySource) {
	// Check for duplicates
	for _, d := range e.Dependents {
		if d.Type == relType && d.Target == target {
			return
		}
	}

	e.Dependents = append(e.Dependents, Relationship{
		Type:   relType,
		Target: target,
		Source: source,
	})
	e.UpdatedAt = time.Now()
}

// AddExport adds an exported symbol
func (e *Entry) AddExport(export string) {
	// Check for duplicates
	for _, exp := range e.Exports {
		if exp == export {
			return
		}
	}
	e.Exports = append(e.Exports, export)
	e.UpdatedAt = time.Now()
}

// AddCall adds a function call
func (e *Entry) AddCall(call string) {
	// Check for duplicates
	for _, c := range e.Calls {
		if c == call {
			return
		}
	}
	e.Calls = append(e.Calls, call)
	e.UpdatedAt = time.Now()
}

// AddConcept adds a semantic concept
func (e *Entry) AddConcept(concept string) {
	// Check for duplicates
	for _, c := range e.Concepts {
		if c == concept {
			return
		}
	}
	e.Concepts = append(e.Concepts, concept)
	e.UpdatedAt = time.Now()
}

// AddAnnotation adds a manual annotation
func (e *Entry) AddAnnotation(key string, value interface{}, author string) {
	now := time.Now()

	// Update existing annotation if key matches
	for i, a := range e.Annotations {
		if a.Key == key {
			e.Annotations[i].Value = value
			e.Annotations[i].UpdatedAt = now
			if author != "" {
				e.Annotations[i].Author = author
			}
			e.UpdatedAt = now
			return
		}
	}

	// Add new annotation
	e.Annotations = append(e.Annotations, Annotation{
		Key:       key,
		Value:     value,
		Author:    author,
		CreatedAt: now,
		UpdatedAt: now,
	})
	e.UpdatedAt = now
}

// GetAnnotation retrieves an annotation by key
func (e *Entry) GetAnnotation(key string) (interface{}, bool) {
	for _, a := range e.Annotations {
		if a.Key == key {
			return a.Value, true
		}
	}
	return nil, false
}

// SetProperty sets a custom property
func (e *Entry) SetProperty(key string, value interface{}) {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	e.Properties[key] = value
	e.UpdatedAt = time.Now()
}

// GetProperty retrieves a custom property
func (e *Entry) GetProperty(key string) (interface{}, bool) {
	if e.Properties == nil {
		return nil, false
	}
	val, ok := e.Properties[key]
	return val, ok
}

// Merge combines this entry with another, preserving manual data
func (e *Entry) Merge(other *Entry, preserveManual bool) *Entry {
	result := &Entry{
		Version:    ShadowVersion,
		SourcePath: e.SourcePath,
		SourceHash: other.SourceHash, // Use newer hash
		Source:     SourceMixed,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  time.Now(),
		Properties: make(map[string]interface{}),
	}

	// Merge module info (prefer newer auto-generated)
	if other.Module != nil {
		result.Module = other.Module
	} else if e.Module != nil {
		result.Module = e.Module
	}

	// Merge triples
	tripleSet := make(map[string]Triple)

	// Add existing triples
	for _, t := range e.Triples {
		key := fmt.Sprintf("%s|%s|%s", t.Subject, t.Predicate, t.Object)
		if preserveManual && t.Source == SourceManual {
			tripleSet[key] = t
		} else if t.Source != SourceManual {
			tripleSet[key] = t
		}
	}

	// Add new triples (may override auto-generated)
	for _, t := range other.Triples {
		key := fmt.Sprintf("%s|%s|%s", t.Subject, t.Predicate, t.Object)
		existing, exists := tripleSet[key]
		if !exists || (!preserveManual || existing.Source != SourceManual) {
			tripleSet[key] = t
		}
	}

	for _, t := range tripleSet {
		result.Triples = append(result.Triples, t)
	}

	// Merge dependencies
	depSet := make(map[string]Relationship)
	for _, d := range e.Dependencies {
		key := fmt.Sprintf("%s|%s", d.Type, d.Target)
		depSet[key] = d
	}
	for _, d := range other.Dependencies {
		key := fmt.Sprintf("%s|%s", d.Type, d.Target)
		if preserveManual {
			if existing, exists := depSet[key]; !exists || existing.Source != SourceManual {
				depSet[key] = d
			}
		} else {
			depSet[key] = d
		}
	}
	for _, d := range depSet {
		result.Dependencies = append(result.Dependencies, d)
	}

	// Merge dependents similarly
	depSetDependents := make(map[string]Relationship)
	for _, d := range e.Dependents {
		key := fmt.Sprintf("%s|%s", d.Type, d.Target)
		depSetDependents[key] = d
	}
	for _, d := range other.Dependents {
		key := fmt.Sprintf("%s|%s", d.Type, d.Target)
		if preserveManual {
			if existing, exists := depSetDependents[key]; !exists || existing.Source != SourceManual {
				depSetDependents[key] = d
			}
		} else {
			depSetDependents[key] = d
		}
	}
	for _, d := range depSetDependents {
		result.Dependents = append(result.Dependents, d)
	}

	// Merge exports and calls (union)
	exportSet := make(map[string]bool)
	for _, exp := range e.Exports {
		exportSet[exp] = true
	}
	for _, exp := range other.Exports {
		exportSet[exp] = true
	}
	for exp := range exportSet {
		result.Exports = append(result.Exports, exp)
	}

	callSet := make(map[string]bool)
	for _, c := range e.Calls {
		callSet[c] = true
	}
	for _, c := range other.Calls {
		callSet[c] = true
	}
	for c := range callSet {
		result.Calls = append(result.Calls, c)
	}

	// Merge concepts (union)
	conceptSet := make(map[string]bool)
	for _, c := range e.Concepts {
		conceptSet[c] = true
	}
	for _, c := range other.Concepts {
		conceptSet[c] = true
	}
	for c := range conceptSet {
		result.Concepts = append(result.Concepts, c)
	}

	// Preserve manual annotations always
	result.Annotations = e.Annotations

	// Merge properties (newer overrides, but preserve keys only in original)
	for k, v := range e.Properties {
		result.Properties[k] = v
	}
	for k, v := range other.Properties {
		result.Properties[k] = v
	}

	return result
}

// HasManualData checks if the entry contains any manual data
func (e *Entry) HasManualData() bool {
	if e.Source == SourceManual || e.Source == SourceMixed {
		return true
	}

	for _, t := range e.Triples {
		if t.Source == SourceManual {
			return true
		}
	}

	for _, d := range e.Dependencies {
		if d.Source == SourceManual {
			return true
		}
	}

	return len(e.Annotations) > 0
}

// GetTriplesByPredicate returns all triples with the given predicate
func (e *Entry) GetTriplesByPredicate(predicate string) []Triple {
	var result []Triple
	for _, t := range e.Triples {
		if t.Predicate == predicate {
			result = append(result, t)
		}
	}
	return result
}

// GetTriplesBySubject returns all triples with the given subject
func (e *Entry) GetTriplesBySubject(subject string) []Triple {
	var result []Triple
	for _, t := range e.Triples {
		if t.Subject == subject {
			result = append(result, t)
		}
	}
	return result
}
