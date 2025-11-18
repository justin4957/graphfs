package query

import (
	"fmt"
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

func TestQueryPlannerOptimization(t *testing.T) {
	// Create a test store with known statistics
	ts := store.NewTripleStore()

	// Add triples with different cardinalities
	// High cardinality predicate (many triples) - different subjects
	for i := 0; i < 100; i++ {
		ts.Add(fmt.Sprintf("<#module%d>", i), "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "https://schema.codedoc.org/Module")
	}

	// Medium cardinality predicate - different subjects
	for i := 0; i < 50; i++ {
		ts.Add(fmt.Sprintf("<#module%d>", i), "https://schema.codedoc.org/language", "go")
	}

	// Low cardinality predicate (few triples) - different subjects
	for i := 0; i < 10; i++ {
		ts.Add(fmt.Sprintf("<#module%d>", i), "https://schema.codedoc.org/exports", fmt.Sprintf("<#Export%d>", i))
	}

	planner := NewQueryPlanner(ts.Stats())

	// Test query with patterns in suboptimal order
	query := &SelectQuery{
		Variables: []string{"?module"},
		Where: []TriplePattern{
			{Subject: "?module", Predicate: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", Object: "https://schema.codedoc.org/Module"}, // High cardinality
			{Subject: "?module", Predicate: "https://schema.codedoc.org/language", Object: "go"},                                            // Medium cardinality
			{Subject: "?module", Predicate: "https://schema.codedoc.org/exports", Object: "?export"},                                        // Low cardinality
		},
	}

	optimized := planner.OptimizeQuery(query)

	// Verify that the most selective pattern (exports) is first
	if optimized.Where[0].Predicate != "https://schema.codedoc.org/exports" {
		t.Errorf("Expected most selective pattern first, got predicate: %s", optimized.Where[0].Predicate)
	}

	// Verify that the least selective pattern (type) is last
	if optimized.Where[2].Predicate != "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" {
		t.Errorf("Expected least selective pattern last, got predicate: %s", optimized.Where[2].Predicate)
	}
}

func TestQueryPlannerSinglePattern(t *testing.T) {
	ts := store.NewTripleStore()
	planner := NewQueryPlanner(ts.Stats())

	// Single pattern query - should not be modified
	query := &SelectQuery{
		Variables: []string{"?module"},
		Where: []TriplePattern{
			{Subject: "?module", Predicate: "https://schema.codedoc.org/language", Object: "go"},
		},
	}

	optimized := planner.OptimizeQuery(query)

	if len(optimized.Where) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(optimized.Where))
	}

	if optimized.Where[0].Predicate != query.Where[0].Predicate {
		t.Error("Single pattern was modified unexpectedly")
	}
}

func TestQueryPlannerPreservesOtherFields(t *testing.T) {
	ts := store.NewTripleStore()
	planner := NewQueryPlanner(ts.Stats())

	// Query with multiple fields set
	query := &SelectQuery{
		Variables: []string{"?module", "?export"},
		Distinct:  true,
		Where: []TriplePattern{
			{Subject: "?module", Predicate: "https://schema.codedoc.org/language", Object: "go"},
			{Subject: "?module", Predicate: "https://schema.codedoc.org/exports", Object: "?export"},
		},
		Filters: []Filter{{Expression: "?module != \"test\""}},
		OrderBy: []OrderBy{{Variable: "?module", Descending: false}},
		Limit:   10,
		Offset:  5,
		Prefixes: map[string]string{
			"code": "https://schema.codedoc.org/",
		},
	}

	optimized := planner.OptimizeQuery(query)

	// Verify all fields are preserved
	if len(optimized.Variables) != 2 {
		t.Errorf("Variables not preserved: got %d, want 2", len(optimized.Variables))
	}
	if !optimized.Distinct {
		t.Error("Distinct flag not preserved")
	}
	if len(optimized.Filters) != 1 {
		t.Errorf("Filters not preserved: got %d, want 1", len(optimized.Filters))
	}
	if len(optimized.OrderBy) != 1 {
		t.Errorf("OrderBy not preserved: got %d, want 1", len(optimized.OrderBy))
	}
	if optimized.Limit != 10 {
		t.Errorf("Limit not preserved: got %d, want 10", optimized.Limit)
	}
	if optimized.Offset != 5 {
		t.Errorf("Offset not preserved: got %d, want 5", optimized.Offset)
	}
	if len(optimized.Prefixes) != 1 {
		t.Errorf("Prefixes not preserved: got %d, want 1", len(optimized.Prefixes))
	}
}

func TestSelectivityEstimation(t *testing.T) {
	ts := store.NewTripleStore()

	// Add test data with known distribution
	ts.Add("<#mod1>", "code:type", "Module")
	ts.Add("<#mod2>", "code:type", "Module")
	ts.Add("<#mod3>", "code:type", "Module")
	ts.Add("<#mod1>", "code:exports", "Func1")

	planner := NewQueryPlanner(ts.Stats())

	// Test selectivity for bound vs unbound predicates
	highCardinalityPattern := TriplePattern{
		Subject:   "?module",
		Predicate: "code:type",
		Object:    "Module",
	}

	lowCardinalityPattern := TriplePattern{
		Subject:   "?module",
		Predicate: "code:exports",
		Object:    "?export",
	}

	highSel := planner.estimateSelectivity(highCardinalityPattern)
	lowSel := planner.estimateSelectivity(lowCardinalityPattern)

	// Lower selectivity value = more selective (fewer results)
	if lowSel >= highSel {
		t.Errorf("Expected exports (low cardinality) to be more selective than type (high cardinality), got low=%f, high=%f", lowSel, highSel)
	}
}

func TestSelectivityWithUnknownPredicates(t *testing.T) {
	ts := store.NewTripleStore()
	ts.Add("<#mod1>", "code:known", "value")

	planner := NewQueryPlanner(ts.Stats())

	// Pattern with unknown predicate should be very selective
	unknownPattern := TriplePattern{
		Subject:   "?module",
		Predicate: "code:unknown",
		Object:    "?value",
	}

	knownPattern := TriplePattern{
		Subject:   "?module",
		Predicate: "code:known",
		Object:    "?value",
	}

	unknownSel := planner.estimateSelectivity(unknownPattern)
	knownSel := planner.estimateSelectivity(knownPattern)

	// Unknown predicates should be considered highly selective
	if unknownSel >= knownSel {
		t.Errorf("Expected unknown predicate to be more selective, got unknown=%f, known=%f", unknownSel, knownSel)
	}
}

func TestExecutorPlanningToggle(t *testing.T) {
	ts := store.NewTripleStore()

	// Add test data
	ts.Add("<#mod1>", "<type>", "<Module>")
	ts.Add("<#mod1>", "<exports>", "<Func1>")

	executor := NewExecutor(ts)

	query := `
		SELECT ?module
		WHERE {
			?module <type> <Module> .
			?module <exports> ?export
		}
	`

	// Test with planning enabled
	resultWithPlanning, err := executor.ExecuteString(query)
	if err != nil {
		t.Fatalf("Query with planning failed: %v", err)
	}

	// Disable planning
	executor.DisablePlanning()

	// Test with planning disabled
	resultWithoutPlanning, err := executor.ExecuteString(query)
	if err != nil {
		t.Fatalf("Query without planning failed: %v", err)
	}

	// Results should be the same, just potentially different performance
	if resultWithPlanning.Count != resultWithoutPlanning.Count {
		t.Errorf("Results differ: with planning=%d, without=%d", resultWithPlanning.Count, resultWithoutPlanning.Count)
	}

	// Re-enable planning
	executor.EnablePlanning()

	// Verify planning is enabled again
	resultAfterReEnable, err := executor.ExecuteString(query)
	if err != nil {
		t.Fatalf("Query after re-enabling planning failed: %v", err)
	}

	if resultAfterReEnable.Count != resultWithPlanning.Count {
		t.Errorf("Results differ after re-enabling: got %d, want %d", resultAfterReEnable.Count, resultWithPlanning.Count)
	}
}
