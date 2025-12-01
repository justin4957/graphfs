/*
# Module: pkg/shadow/index.go
Shadow index for fast lookups and queries.

Provides in-memory indexing of shadow entries for efficient querying
by various attributes like tags, concepts, language, and layer.

## Linked Modules
- [shadow](./shadow.go) - Shadow file system manager
- [entry](./entry.go) - Shadow entry data structure

## Tags
shadow, index, query, lookup

## Exports
Index, NewIndex, IndexEntry

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#index.go> a code:Module ;
    code:name "pkg/shadow/index.go" ;
    code:description "Shadow index for fast lookups and queries" ;
    code:language "go" ;
    code:layer "shadow" ;
    code:linksTo <./shadow.go>, <./entry.go> ;
    code:exports <#Index>, <#NewIndex>, <#IndexEntry> ;
    code:tags "shadow", "index", "query", "lookup" .
<!-- End LinkedDoc RDF -->
*/

package shadow

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// IndexEntry represents a lightweight index record for a shadow entry
type IndexEntry struct {
	Path        string      `json:"path"`
	URI         string      `json:"uri,omitempty"`
	Name        string      `json:"name,omitempty"`
	Language    string      `json:"language,omitempty"`
	Layer       string      `json:"layer,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Concepts    []string    `json:"concepts,omitempty"`
	Source      EntrySource `json:"source"`
	UpdatedAt   time.Time   `json:"updated_at"`
	TripleCount int         `json:"triple_count"`
	HasManual   bool        `json:"has_manual"`
}

// Index provides fast lookups for shadow entries
type Index struct {
	// Version for format compatibility
	Version string `json:"version"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at"`

	// Entries indexed by path
	Entries map[string]*IndexEntry `json:"entries"`

	// Inverted indexes for fast lookups
	ByTag      map[string][]string `json:"by_tag"`
	ByConcept  map[string][]string `json:"by_concept"`
	ByLanguage map[string][]string `json:"by_language"`
	ByLayer    map[string][]string `json:"by_layer"`

	// Statistics
	Stats IndexStats `json:"stats"`

	// Mutex for thread-safe operations
	mu sync.RWMutex `json:"-"`
}

// IndexStats tracks index statistics
type IndexStats struct {
	TotalEntries  int            `json:"total_entries"`
	TotalTriples  int            `json:"total_triples"`
	ManualEntries int            `json:"manual_entries"`
	AutoEntries   int            `json:"auto_entries"`
	MixedEntries  int            `json:"mixed_entries"`
	LanguageCount map[string]int `json:"language_count"`
	LayerCount    map[string]int `json:"layer_count"`
	TagCount      map[string]int `json:"tag_count"`
	ConceptCount  map[string]int `json:"concept_count"`
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	now := time.Now()
	return &Index{
		Version:    ShadowVersion,
		CreatedAt:  now,
		UpdatedAt:  now,
		Entries:    make(map[string]*IndexEntry),
		ByTag:      make(map[string][]string),
		ByConcept:  make(map[string][]string),
		ByLanguage: make(map[string][]string),
		ByLayer:    make(map[string][]string),
		Stats: IndexStats{
			LanguageCount: make(map[string]int),
			LayerCount:    make(map[string]int),
			TagCount:      make(map[string]int),
			ConceptCount:  make(map[string]int),
		},
	}
}

// Add adds or updates an entry in the index
func (idx *Index) Add(path string, entry *Entry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove existing entry from inverted indexes if present
	if existing, ok := idx.Entries[path]; ok {
		idx.removeFromInvertedIndexes(path, existing)
	}

	// Create index entry
	indexEntry := &IndexEntry{
		Path:        path,
		Source:      entry.Source,
		UpdatedAt:   entry.UpdatedAt,
		TripleCount: len(entry.Triples),
		HasManual:   entry.HasManualData(),
		Concepts:    entry.Concepts,
	}

	// Extract module info if present
	if entry.Module != nil {
		indexEntry.URI = entry.Module.URI
		indexEntry.Name = entry.Module.Name
		indexEntry.Language = entry.Module.Language
		indexEntry.Layer = entry.Module.Layer
		indexEntry.Tags = entry.Module.Tags
	}

	// Store entry
	idx.Entries[path] = indexEntry

	// Update inverted indexes
	idx.addToInvertedIndexes(path, indexEntry)

	// Update stats
	idx.updateStats()
	idx.UpdatedAt = time.Now()
}

// Remove removes an entry from the index
func (idx *Index) Remove(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if entry, ok := idx.Entries[path]; ok {
		idx.removeFromInvertedIndexes(path, entry)
		delete(idx.Entries, path)
		idx.updateStats()
		idx.UpdatedAt = time.Now()
	}
}

// Get retrieves an index entry by path
func (idx *Index) Get(path string) (*IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entry, ok := idx.Entries[path]
	return entry, ok
}

// GetByTag returns all paths with the given tag
func (idx *Index) GetByTag(tag string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	paths := idx.ByTag[tag]
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// GetByConcept returns all paths with the given concept
func (idx *Index) GetByConcept(concept string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	paths := idx.ByConcept[concept]
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// GetByLanguage returns all paths with the given language
func (idx *Index) GetByLanguage(language string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	paths := idx.ByLanguage[language]
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// GetByLayer returns all paths with the given layer
func (idx *Index) GetByLayer(layer string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	paths := idx.ByLayer[layer]
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// Search performs a multi-criteria search on the index
func (idx *Index) Search(query SearchQuery) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []string
	firstFilter := true

	// Start with all entries or filter by specific criteria
	if query.Language != "" {
		results = idx.ByLanguage[query.Language]
		firstFilter = false
	}

	if query.Layer != "" {
		layerResults := idx.ByLayer[query.Layer]
		if firstFilter {
			results = layerResults
			firstFilter = false
		} else {
			results = intersect(results, layerResults)
		}
	}

	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			tagResults := idx.ByTag[tag]
			if firstFilter {
				results = tagResults
				firstFilter = false
			} else {
				results = intersect(results, tagResults)
			}
		}
	}

	if len(query.Concepts) > 0 {
		for _, concept := range query.Concepts {
			conceptResults := idx.ByConcept[concept]
			if firstFilter {
				results = conceptResults
				firstFilter = false
			} else {
				results = intersect(results, conceptResults)
			}
		}
	}

	// If no filters applied, return all paths
	if firstFilter {
		results = make([]string, 0, len(idx.Entries))
		for path := range idx.Entries {
			results = append(results, path)
		}
	}

	// Apply text search filter
	if query.TextQuery != "" {
		results = idx.filterByText(results, query.TextQuery)
	}

	// Apply source filter
	if query.Source != "" {
		results = idx.filterBySource(results, query.Source)
	}

	// Apply manual filter
	if query.HasManual {
		results = idx.filterByManual(results, true)
	}

	// Sort results
	sort.Strings(results)

	// Apply limit and offset
	if query.Offset > 0 {
		if query.Offset >= len(results) {
			return []string{}
		}
		results = results[query.Offset:]
	}

	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results
}

// SearchQuery defines search criteria
type SearchQuery struct {
	Language  string
	Layer     string
	Tags      []string
	Concepts  []string
	TextQuery string
	Source    EntrySource
	HasManual bool
	Limit     int
	Offset    int
}

// ListTags returns all unique tags
func (idx *Index) ListTags() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	tags := make([]string, 0, len(idx.ByTag))
	for tag := range idx.ByTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// ListConcepts returns all unique concepts
func (idx *Index) ListConcepts() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	concepts := make([]string, 0, len(idx.ByConcept))
	for concept := range idx.ByConcept {
		concepts = append(concepts, concept)
	}
	sort.Strings(concepts)
	return concepts
}

// ListLanguages returns all unique languages
func (idx *Index) ListLanguages() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	languages := make([]string, 0, len(idx.ByLanguage))
	for lang := range idx.ByLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages
}

// ListLayers returns all unique layers
func (idx *Index) ListLayers() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	layers := make([]string, 0, len(idx.ByLayer))
	for layer := range idx.ByLayer {
		layers = append(layers, layer)
	}
	sort.Strings(layers)
	return layers
}

// Statistics returns the current index statistics
func (idx *Index) Statistics() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.Stats
}

// Load loads the index from a file
func (idx *Index) Load(path string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var loaded Index
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to parse index file: %w", err)
	}

	// Copy data to current index
	idx.Version = loaded.Version
	idx.CreatedAt = loaded.CreatedAt
	idx.UpdatedAt = loaded.UpdatedAt
	idx.Entries = loaded.Entries
	idx.ByTag = loaded.ByTag
	idx.ByConcept = loaded.ByConcept
	idx.ByLanguage = loaded.ByLanguage
	idx.ByLayer = loaded.ByLayer
	idx.Stats = loaded.Stats

	// Initialize maps if nil
	if idx.Entries == nil {
		idx.Entries = make(map[string]*IndexEntry)
	}
	if idx.ByTag == nil {
		idx.ByTag = make(map[string][]string)
	}
	if idx.ByConcept == nil {
		idx.ByConcept = make(map[string][]string)
	}
	if idx.ByLanguage == nil {
		idx.ByLanguage = make(map[string][]string)
	}
	if idx.ByLayer == nil {
		idx.ByLayer = make(map[string][]string)
	}

	return nil
}

// Save saves the index to a file
func (idx *Index) Save(path string) error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize index: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// Clear removes all entries from the index
func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Entries = make(map[string]*IndexEntry)
	idx.ByTag = make(map[string][]string)
	idx.ByConcept = make(map[string][]string)
	idx.ByLanguage = make(map[string][]string)
	idx.ByLayer = make(map[string][]string)
	idx.Stats = IndexStats{
		LanguageCount: make(map[string]int),
		LayerCount:    make(map[string]int),
		TagCount:      make(map[string]int),
		ConceptCount:  make(map[string]int),
	}
	idx.UpdatedAt = time.Now()
}

// Count returns the total number of entries
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.Entries)
}

// addToInvertedIndexes adds entry to all inverted indexes
func (idx *Index) addToInvertedIndexes(path string, entry *IndexEntry) {
	// Add to language index
	if entry.Language != "" {
		idx.ByLanguage[entry.Language] = append(idx.ByLanguage[entry.Language], path)
	}

	// Add to layer index
	if entry.Layer != "" {
		idx.ByLayer[entry.Layer] = append(idx.ByLayer[entry.Layer], path)
	}

	// Add to tag indexes
	for _, tag := range entry.Tags {
		idx.ByTag[tag] = append(idx.ByTag[tag], path)
	}

	// Add to concept indexes
	for _, concept := range entry.Concepts {
		idx.ByConcept[concept] = append(idx.ByConcept[concept], path)
	}
}

// removeFromInvertedIndexes removes entry from all inverted indexes
func (idx *Index) removeFromInvertedIndexes(path string, entry *IndexEntry) {
	// Remove from language index
	if entry.Language != "" {
		idx.ByLanguage[entry.Language] = removeFromSlice(idx.ByLanguage[entry.Language], path)
		if len(idx.ByLanguage[entry.Language]) == 0 {
			delete(idx.ByLanguage, entry.Language)
		}
	}

	// Remove from layer index
	if entry.Layer != "" {
		idx.ByLayer[entry.Layer] = removeFromSlice(idx.ByLayer[entry.Layer], path)
		if len(idx.ByLayer[entry.Layer]) == 0 {
			delete(idx.ByLayer, entry.Layer)
		}
	}

	// Remove from tag indexes
	for _, tag := range entry.Tags {
		idx.ByTag[tag] = removeFromSlice(idx.ByTag[tag], path)
		if len(idx.ByTag[tag]) == 0 {
			delete(idx.ByTag, tag)
		}
	}

	// Remove from concept indexes
	for _, concept := range entry.Concepts {
		idx.ByConcept[concept] = removeFromSlice(idx.ByConcept[concept], path)
		if len(idx.ByConcept[concept]) == 0 {
			delete(idx.ByConcept, concept)
		}
	}
}

// updateStats recalculates index statistics
func (idx *Index) updateStats() {
	idx.Stats = IndexStats{
		TotalEntries:  len(idx.Entries),
		LanguageCount: make(map[string]int),
		LayerCount:    make(map[string]int),
		TagCount:      make(map[string]int),
		ConceptCount:  make(map[string]int),
	}

	for _, entry := range idx.Entries {
		idx.Stats.TotalTriples += entry.TripleCount

		switch entry.Source {
		case SourceManual:
			idx.Stats.ManualEntries++
		case SourceAuto:
			idx.Stats.AutoEntries++
		case SourceMixed:
			idx.Stats.MixedEntries++
		}

		if entry.Language != "" {
			idx.Stats.LanguageCount[entry.Language]++
		}
		if entry.Layer != "" {
			idx.Stats.LayerCount[entry.Layer]++
		}
		for _, tag := range entry.Tags {
			idx.Stats.TagCount[tag]++
		}
		for _, concept := range entry.Concepts {
			idx.Stats.ConceptCount[concept]++
		}
	}
}

// filterByText filters results by text query
func (idx *Index) filterByText(paths []string, query string) []string {
	query = strings.ToLower(query)
	var results []string

	for _, path := range paths {
		entry := idx.Entries[path]
		if entry == nil {
			continue
		}

		// Search in path
		if strings.Contains(strings.ToLower(path), query) {
			results = append(results, path)
			continue
		}

		// Search in name
		if strings.Contains(strings.ToLower(entry.Name), query) {
			results = append(results, path)
			continue
		}

		// Search in URI
		if strings.Contains(strings.ToLower(entry.URI), query) {
			results = append(results, path)
			continue
		}
	}

	return results
}

// filterBySource filters results by entry source
func (idx *Index) filterBySource(paths []string, source EntrySource) []string {
	var results []string

	for _, path := range paths {
		entry := idx.Entries[path]
		if entry != nil && entry.Source == source {
			results = append(results, path)
		}
	}

	return results
}

// filterByManual filters results by manual data presence
func (idx *Index) filterByManual(paths []string, hasManual bool) []string {
	var results []string

	for _, path := range paths {
		entry := idx.Entries[path]
		if entry != nil && entry.HasManual == hasManual {
			results = append(results, path)
		}
	}

	return results
}

// intersect returns the intersection of two string slices
func intersect(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}

	var result []string
	for _, s := range b {
		if set[s] {
			result = append(result, s)
		}
	}

	return result
}

// removeFromSlice removes a string from a slice
func removeFromSlice(slice []string, str string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != str {
			result = append(result, s)
		}
	}
	return result
}
