/*
# Module: internal/store/store.go
In-memory RDF triple store with multiple indexes.

Implements an efficient in-memory triple store with SPO, POS, and OSP indexes
for fast lookups and pattern matching.

## Linked Modules
- [triple](./triple.go) - Triple data structure

## Tags
store, rdf, triplestore, in-memory

## Exports
TripleStore, NewTripleStore

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#store.go> a code:Module ;
    code:name "internal/store/store.go" ;
    code:description "In-memory RDF triple store with multiple indexes" ;
    code:language "go" ;
    code:layer "storage" ;
    code:linksTo <./triple.go> ;
    code:exports <#TripleStore>, <#NewTripleStore> ;
    code:tags "store", "rdf", "triplestore", "in-memory" .

<#TripleStore> a code:Type ;
    code:name "TripleStore" ;
    code:kind "struct" ;
    code:description "In-memory triple store with multiple indexes" ;
    code:hasMethod <#TripleStore.Add>, <#TripleStore.BulkAdd>, <#TripleStore.Find>, <#TripleStore.Delete>, <#TripleStore.Clear> .

<#NewTripleStore> a code:Function ;
    code:name "NewTripleStore" ;
    code:description "Creates new triple store instance" ;
    code:returns <#TripleStore> .
<!-- End LinkedDoc RDF -->
*/

package store

import (
	"fmt"
	"sync"
)

// IndexStats contains statistics for query optimization
type IndexStats struct {
	PredicateCounts map[string]int // Number of triples per predicate
	SubjectCounts   map[string]int // Number of triples per subject
	ObjectCounts    map[string]int // Number of triples per object
	TotalTriples    int            // Total number of triples
}

// TripleStore is an in-memory RDF triple store with multiple indexes
type TripleStore struct {
	mu sync.RWMutex

	// SPO index: Subject -> Predicate -> Object -> exists
	spo map[string]map[string]map[string]bool

	// POS index: Predicate -> Object -> Subject -> exists
	pos map[string]map[string]map[string]bool

	// OSP index: Object -> Subject -> Predicate -> exists
	osp map[string]map[string]map[string]bool

	// Count of triples
	count int

	// Statistics for query optimization
	stats IndexStats
}

// NewTripleStore creates a new in-memory triple store
func NewTripleStore() *TripleStore {
	return &TripleStore{
		spo:   make(map[string]map[string]map[string]bool),
		pos:   make(map[string]map[string]map[string]bool),
		osp:   make(map[string]map[string]map[string]bool),
		count: 0,
		stats: IndexStats{
			PredicateCounts: make(map[string]int),
			SubjectCounts:   make(map[string]int),
			ObjectCounts:    make(map[string]int),
			TotalTriples:    0,
		},
	}
}

// Add inserts a triple into the store
func (ts *TripleStore) Add(subject, predicate, object string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Check if triple already exists
	if ts.existsUnsafe(subject, predicate, object) {
		return nil // Already exists, no error
	}

	// Add to SPO index
	if ts.spo[subject] == nil {
		ts.spo[subject] = make(map[string]map[string]bool)
	}
	if ts.spo[subject][predicate] == nil {
		ts.spo[subject][predicate] = make(map[string]bool)
	}
	ts.spo[subject][predicate][object] = true

	// Add to POS index
	if ts.pos[predicate] == nil {
		ts.pos[predicate] = make(map[string]map[string]bool)
	}
	if ts.pos[predicate][object] == nil {
		ts.pos[predicate][object] = make(map[string]bool)
	}
	ts.pos[predicate][object][subject] = true

	// Add to OSP index
	if ts.osp[object] == nil {
		ts.osp[object] = make(map[string]map[string]bool)
	}
	if ts.osp[object][subject] == nil {
		ts.osp[object][subject] = make(map[string]bool)
	}
	ts.osp[object][subject][predicate] = true

	// Update statistics
	ts.stats.PredicateCounts[predicate]++
	ts.stats.SubjectCounts[subject]++
	ts.stats.ObjectCounts[object]++
	ts.stats.TotalTriples++

	ts.count++
	return nil
}

// AddTriple inserts a Triple struct into the store
func (ts *TripleStore) AddTriple(triple Triple) error {
	return ts.Add(triple.Subject, triple.Predicate, triple.Object)
}

// BulkAdd inserts multiple triples efficiently
func (ts *TripleStore) BulkAdd(triples []Triple) error {
	for _, triple := range triples {
		if err := ts.AddTriple(triple); err != nil {
			return err
		}
	}
	return nil
}

// Find queries triples matching the pattern (use "" for wildcard)
func (ts *TripleStore) Find(subject, predicate, object string) []Triple {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	var results []Triple

	// All wildcards - return all triples
	if subject == "" && predicate == "" && object == "" {
		for s, pMap := range ts.spo {
			for p, oMap := range pMap {
				for o := range oMap {
					results = append(results, Triple{Subject: s, Predicate: p, Object: o})
				}
			}
		}
		return results
	}

	// Use most specific index
	if subject != "" {
		// Use SPO index
		if pMap, ok := ts.spo[subject]; ok {
			if predicate != "" {
				// S and P specified
				if oMap, ok := pMap[predicate]; ok {
					if object != "" {
						// All specified
						if oMap[object] {
							results = append(results, Triple{Subject: subject, Predicate: predicate, Object: object})
						}
					} else {
						// S and P specified, O wildcard
						for o := range oMap {
							results = append(results, Triple{Subject: subject, Predicate: predicate, Object: o})
						}
					}
				}
			} else {
				// S specified, P and O wildcards
				for p, oMap := range pMap {
					if object != "" {
						// S and O specified, P wildcard
						if oMap[object] {
							results = append(results, Triple{Subject: subject, Predicate: p, Object: object})
						}
					} else {
						// S specified, P and O wildcards
						for o := range oMap {
							results = append(results, Triple{Subject: subject, Predicate: p, Object: o})
						}
					}
				}
			}
		}
	} else if predicate != "" {
		// Use POS index
		if oMap, ok := ts.pos[predicate]; ok {
			if object != "" {
				// P and O specified, S wildcard
				if sMap, ok := oMap[object]; ok {
					for s := range sMap {
						results = append(results, Triple{Subject: s, Predicate: predicate, Object: object})
					}
				}
			} else {
				// P specified, S and O wildcards
				for o, sMap := range oMap {
					for s := range sMap {
						results = append(results, Triple{Subject: s, Predicate: predicate, Object: o})
					}
				}
			}
		}
	} else if object != "" {
		// Use OSP index (O specified, S and P wildcards)
		if sMap, ok := ts.osp[object]; ok {
			for s, pMap := range sMap {
				for p := range pMap {
					results = append(results, Triple{Subject: s, Predicate: p, Object: object})
				}
			}
		}
	}

	return results
}

// Get retrieves all properties for a subject as a map
func (ts *TripleStore) Get(subject string) map[string][]string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make(map[string][]string)

	if pMap, ok := ts.spo[subject]; ok {
		for p, oMap := range pMap {
			var objects []string
			for o := range oMap {
				objects = append(objects, o)
			}
			result[p] = objects
		}
	}

	return result
}

// Delete removes matching triples
func (ts *TripleStore) Delete(subject, predicate, object string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Find matching triples first
	matches := ts.findUnsafe(subject, predicate, object)

	// Delete each match
	for _, triple := range matches {
		ts.deleteTripleUnsafe(triple.Subject, triple.Predicate, triple.Object)
	}

	return nil
}

// Clear removes all triples
func (ts *TripleStore) Clear() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.spo = make(map[string]map[string]map[string]bool)
	ts.pos = make(map[string]map[string]map[string]bool)
	ts.osp = make(map[string]map[string]map[string]bool)
	ts.count = 0

	// Reset statistics
	ts.stats = IndexStats{
		PredicateCounts: make(map[string]int),
		SubjectCounts:   make(map[string]int),
		ObjectCounts:    make(map[string]int),
		TotalTriples:    0,
	}

	return nil
}

// Count returns the total number of triples
func (ts *TripleStore) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.count
}

// Subjects returns all unique subjects
func (ts *TripleStore) Subjects() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	subjects := make([]string, 0, len(ts.spo))
	for s := range ts.spo {
		subjects = append(subjects, s)
	}
	return subjects
}

// Predicates returns all unique predicates
func (ts *TripleStore) Predicates() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	predicates := make([]string, 0, len(ts.pos))
	for p := range ts.pos {
		predicates = append(predicates, p)
	}
	return predicates
}

// Objects returns all unique objects
func (ts *TripleStore) Objects() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	objects := make([]string, 0, len(ts.osp))
	for o := range ts.osp {
		objects = append(objects, o)
	}
	return objects
}

// existsUnsafe checks if a triple exists (no locking)
func (ts *TripleStore) existsUnsafe(subject, predicate, object string) bool {
	if pMap, ok := ts.spo[subject]; ok {
		if oMap, ok := pMap[predicate]; ok {
			return oMap[object]
		}
	}
	return false
}

// findUnsafe finds triples without locking (used internally)
func (ts *TripleStore) findUnsafe(subject, predicate, object string) []Triple {
	var results []Triple

	if subject != "" {
		if pMap, ok := ts.spo[subject]; ok {
			if predicate != "" {
				if oMap, ok := pMap[predicate]; ok {
					if object != "" {
						if oMap[object] {
							results = append(results, Triple{Subject: subject, Predicate: predicate, Object: object})
						}
					} else {
						for o := range oMap {
							results = append(results, Triple{Subject: subject, Predicate: predicate, Object: o})
						}
					}
				}
			} else {
				for p, oMap := range pMap {
					if object != "" {
						if oMap[object] {
							results = append(results, Triple{Subject: subject, Predicate: p, Object: object})
						}
					} else {
						for o := range oMap {
							results = append(results, Triple{Subject: subject, Predicate: p, Object: o})
						}
					}
				}
			}
		}
	} else {
		// Wildcard subject - iterate all
		for s, pMap := range ts.spo {
			if predicate != "" {
				if oMap, ok := pMap[predicate]; ok {
					if object != "" {
						if oMap[object] {
							results = append(results, Triple{Subject: s, Predicate: predicate, Object: object})
						}
					} else {
						for o := range oMap {
							results = append(results, Triple{Subject: s, Predicate: predicate, Object: o})
						}
					}
				}
			} else if object != "" {
				for p, oMap := range pMap {
					if oMap[object] {
						results = append(results, Triple{Subject: s, Predicate: p, Object: object})
					}
				}
			} else {
				for p, oMap := range pMap {
					for o := range oMap {
						results = append(results, Triple{Subject: s, Predicate: p, Object: o})
					}
				}
			}
		}
	}

	return results
}

// deleteTripleUnsafe deletes a specific triple (no locking)
func (ts *TripleStore) deleteTripleUnsafe(subject, predicate, object string) {
	// Remove from SPO index
	if pMap, ok := ts.spo[subject]; ok {
		if oMap, ok := pMap[predicate]; ok {
			delete(oMap, object)
			if len(oMap) == 0 {
				delete(pMap, predicate)
			}
		}
		if len(pMap) == 0 {
			delete(ts.spo, subject)
		}
	}

	// Remove from POS index
	if oMap, ok := ts.pos[predicate]; ok {
		if sMap, ok := oMap[object]; ok {
			delete(sMap, subject)
			if len(sMap) == 0 {
				delete(oMap, object)
			}
		}
		if len(oMap) == 0 {
			delete(ts.pos, predicate)
		}
	}

	// Remove from OSP index
	if sMap, ok := ts.osp[object]; ok {
		if pMap, ok := sMap[subject]; ok {
			delete(pMap, predicate)
			if len(pMap) == 0 {
				delete(sMap, subject)
			}
		}
		if len(sMap) == 0 {
			delete(ts.osp, object)
		}
	}

	// Update statistics
	ts.stats.PredicateCounts[predicate]--
	if ts.stats.PredicateCounts[predicate] <= 0 {
		delete(ts.stats.PredicateCounts, predicate)
	}
	ts.stats.SubjectCounts[subject]--
	if ts.stats.SubjectCounts[subject] <= 0 {
		delete(ts.stats.SubjectCounts, subject)
	}
	ts.stats.ObjectCounts[object]--
	if ts.stats.ObjectCounts[object] <= 0 {
		delete(ts.stats.ObjectCounts, object)
	}
	ts.stats.TotalTriples--

	ts.count--
}

// String returns a string representation of the store statistics
func (ts *TripleStore) String() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return fmt.Sprintf("TripleStore{triples: %d, subjects: %d, predicates: %d, objects: %d}",
		ts.count, len(ts.spo), len(ts.pos), len(ts.osp))
}

// Stats returns a copy of the index statistics for query optimization
func (ts *TripleStore) Stats() IndexStats {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Return a copy to prevent external modification
	statsCopy := IndexStats{
		PredicateCounts: make(map[string]int, len(ts.stats.PredicateCounts)),
		SubjectCounts:   make(map[string]int, len(ts.stats.SubjectCounts)),
		ObjectCounts:    make(map[string]int, len(ts.stats.ObjectCounts)),
		TotalTriples:    ts.stats.TotalTriples,
	}

	for k, v := range ts.stats.PredicateCounts {
		statsCopy.PredicateCounts[k] = v
	}
	for k, v := range ts.stats.SubjectCounts {
		statsCopy.SubjectCounts[k] = v
	}
	for k, v := range ts.stats.ObjectCounts {
		statsCopy.ObjectCounts[k] = v
	}

	return statsCopy
}
