package store

import (
	"sync"
	"testing"
)

func TestTripleStore_Add(t *testing.T) {
	store := NewTripleStore()

	err := store.Add("subject1", "predicate1", "object1")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if store.Count() != 1 {
		t.Errorf("Count() = %d, want 1", store.Count())
	}

	// Adding same triple again should not error
	err = store.Add("subject1", "predicate1", "object1")
	if err != nil {
		t.Errorf("Add() duplicate should not error, got %v", err)
	}

	// Count should still be 1
	if store.Count() != 1 {
		t.Errorf("Count() after duplicate = %d, want 1", store.Count())
	}
}

func TestTripleStore_AddTriple(t *testing.T) {
	store := NewTripleStore()

	triple := NewTriple("subject1", "predicate1", "object1")
	err := store.AddTriple(triple)
	if err != nil {
		t.Fatalf("AddTriple() error = %v", err)
	}

	if store.Count() != 1 {
		t.Errorf("Count() = %d, want 1", store.Count())
	}
}

func TestTripleStore_BulkAdd(t *testing.T) {
	store := NewTripleStore()

	triples := []Triple{
		NewTriple("s1", "p1", "o1"),
		NewTriple("s1", "p2", "o2"),
		NewTriple("s2", "p1", "o3"),
	}

	err := store.BulkAdd(triples)
	if err != nil {
		t.Fatalf("BulkAdd() error = %v", err)
	}

	if store.Count() != 3 {
		t.Errorf("Count() = %d, want 3", store.Count())
	}
}

func TestTripleStore_Find(t *testing.T) {
	store := NewTripleStore()

	// Add test data
	store.Add("s1", "p1", "o1")
	store.Add("s1", "p2", "o2")
	store.Add("s2", "p1", "o3")
	store.Add("s2", "p2", "o1")

	tests := []struct {
		name      string
		subject   string
		predicate string
		object    string
		wantCount int
	}{
		{
			name:      "find all",
			subject:   "",
			predicate: "",
			object:    "",
			wantCount: 4,
		},
		{
			name:      "find by subject",
			subject:   "s1",
			predicate: "",
			object:    "",
			wantCount: 2,
		},
		{
			name:      "find by predicate",
			subject:   "",
			predicate: "p1",
			object:    "",
			wantCount: 2,
		},
		{
			name:      "find by object",
			subject:   "",
			predicate: "",
			object:    "o1",
			wantCount: 2,
		},
		{
			name:      "find by subject and predicate",
			subject:   "s1",
			predicate: "p1",
			object:    "",
			wantCount: 1,
		},
		{
			name:      "find by predicate and object",
			subject:   "",
			predicate: "p1",
			object:    "o1",
			wantCount: 1,
		},
		{
			name:      "find by subject and object",
			subject:   "s1",
			predicate: "",
			object:    "o1",
			wantCount: 1,
		},
		{
			name:      "find exact triple",
			subject:   "s1",
			predicate: "p1",
			object:    "o1",
			wantCount: 1,
		},
		{
			name:      "find non-existent",
			subject:   "s3",
			predicate: "",
			object:    "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := store.Find(tt.subject, tt.predicate, tt.object)
			if len(results) != tt.wantCount {
				t.Errorf("Find() returned %d results, want %d", len(results), tt.wantCount)
				for _, r := range results {
					t.Logf("  %s", r.String())
				}
			}
		})
	}
}

func TestTripleStore_Get(t *testing.T) {
	store := NewTripleStore()

	store.Add("s1", "p1", "o1")
	store.Add("s1", "p2", "o2")
	store.Add("s1", "p1", "o3")

	props := store.Get("s1")

	if len(props) != 2 {
		t.Errorf("Get() returned %d properties, want 2", len(props))
	}

	if len(props["p1"]) != 2 {
		t.Errorf("Get() p1 has %d objects, want 2", len(props["p1"]))
	}

	if len(props["p2"]) != 1 {
		t.Errorf("Get() p2 has %d objects, want 1", len(props["p2"]))
	}
}

func TestTripleStore_Delete(t *testing.T) {
	store := NewTripleStore()

	// Add test data
	store.Add("s1", "p1", "o1")
	store.Add("s1", "p2", "o2")
	store.Add("s2", "p1", "o3")

	// Delete specific triple
	err := store.Delete("s1", "p1", "o1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if store.Count() != 2 {
		t.Errorf("Count() after delete = %d, want 2", store.Count())
	}

	// Verify triple is gone
	results := store.Find("s1", "p1", "o1")
	if len(results) != 0 {
		t.Errorf("Find() after delete returned %d results, want 0", len(results))
	}

	// Delete by pattern
	store.Delete("s1", "", "")
	if store.Count() != 1 {
		t.Errorf("Count() after pattern delete = %d, want 1", store.Count())
	}
}

func TestTripleStore_Clear(t *testing.T) {
	store := NewTripleStore()

	store.Add("s1", "p1", "o1")
	store.Add("s2", "p2", "o2")

	err := store.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", store.Count())
	}

	results := store.Find("", "", "")
	if len(results) != 0 {
		t.Errorf("Find() after Clear() returned %d results, want 0", len(results))
	}
}

func TestTripleStore_Subjects(t *testing.T) {
	store := NewTripleStore()

	store.Add("s1", "p1", "o1")
	store.Add("s2", "p1", "o1")
	store.Add("s1", "p2", "o2")

	subjects := store.Subjects()
	if len(subjects) != 2 {
		t.Errorf("Subjects() returned %d subjects, want 2", len(subjects))
	}

	found := make(map[string]bool)
	for _, s := range subjects {
		found[s] = true
	}

	if !found["s1"] || !found["s2"] {
		t.Errorf("Subjects() missing expected subjects, got %v", subjects)
	}
}

func TestTripleStore_Predicates(t *testing.T) {
	store := NewTripleStore()

	store.Add("s1", "p1", "o1")
	store.Add("s1", "p2", "o1")
	store.Add("s2", "p1", "o2")

	predicates := store.Predicates()
	if len(predicates) != 2 {
		t.Errorf("Predicates() returned %d predicates, want 2", len(predicates))
	}

	found := make(map[string]bool)
	for _, p := range predicates {
		found[p] = true
	}

	if !found["p1"] || !found["p2"] {
		t.Errorf("Predicates() missing expected predicates, got %v", predicates)
	}
}

func TestTripleStore_Objects(t *testing.T) {
	store := NewTripleStore()

	store.Add("s1", "p1", "o1")
	store.Add("s1", "p1", "o2")
	store.Add("s2", "p2", "o1")

	objects := store.Objects()
	if len(objects) != 2 {
		t.Errorf("Objects() returned %d objects, want 2", len(objects))
	}

	found := make(map[string]bool)
	for _, o := range objects {
		found[o] = true
	}

	if !found["o1"] || !found["o2"] {
		t.Errorf("Objects() missing expected objects, got %v", objects)
	}
}

func TestTripleStore_ConcurrentAccess(t *testing.T) {
	store := NewTripleStore()

	var wg sync.WaitGroup
	numGoroutines := 10
	triplesPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < triplesPerGoroutine; j++ {
				store.Add(
					"subject"+string(rune(id)),
					"predicate"+string(rune(j)),
					"object"+string(rune(j)),
				)
			}
		}(i)
	}

	wg.Wait()

	// Should have some triples (exact count may vary due to duplicates)
	count := store.Count()
	if count == 0 {
		t.Error("Count() should be > 0 after concurrent adds")
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Find("", "", "")
			store.Subjects()
			store.Predicates()
			store.Objects()
		}()
	}

	wg.Wait()
}

func TestTriple_Equals(t *testing.T) {
	t1 := NewTriple("s1", "p1", "o1")
	t2 := NewTriple("s1", "p1", "o1")
	t3 := NewTriple("s1", "p1", "o2")

	if !t1.Equals(t2) {
		t.Error("Equal triples should be equal")
	}

	if t1.Equals(t3) {
		t.Error("Different triples should not be equal")
	}
}

func TestTriple_String(t *testing.T) {
	triple := NewTriple("subject", "predicate", "object")
	str := triple.String()

	expected := "subject -> predicate -> object"
	if str != expected {
		t.Errorf("String() = %q, want %q", str, expected)
	}
}

func BenchmarkTripleStore_Add(b *testing.B) {
	store := NewTripleStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Add("subject"+string(rune(i%1000)), "predicate", "object")
	}
}

func BenchmarkTripleStore_Find_BySubject(b *testing.B) {
	store := NewTripleStore()

	// Add test data
	for i := 0; i < 1000; i++ {
		store.Add("subject"+string(rune(i)), "predicate", "object"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Find("subject500", "", "")
	}
}

func BenchmarkTripleStore_Find_ByPredicate(b *testing.B) {
	store := NewTripleStore()

	// Add test data
	for i := 0; i < 1000; i++ {
		store.Add("subject"+string(rune(i)), "predicate", "object"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Find("", "predicate", "")
	}
}

func BenchmarkTripleStore_Find_All(b *testing.B) {
	store := NewTripleStore()

	// Add test data
	for i := 0; i < 1000; i++ {
		store.Add("subject"+string(rune(i)), "predicate", "object"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Find("", "", "")
	}
}
