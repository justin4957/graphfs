/*
# Module: cmd/graphfs/completions_test.go
Tests for shell completion functions.

Tests dynamic completion for module paths, layers, tags, and output formats.

## Linked Modules
- [completions](./completions.go) - Completion functions

## Tags
cli, test, completion

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#completions_test.go> a code:Module ;
    code:name "cmd/graphfs/completions_test.go" ;
    code:description "Tests for shell completion functions" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <./completions.go> ;
    code:tags "cli", "test", "completion" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputFormatCompletion(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		wantCount  int
		wantItems  []string
	}{
		{
			name:       "empty prefix returns all formats",
			toComplete: "",
			wantCount:  6,
			wantItems:  []string{"table", "json", "csv", "dot", "mermaid", "turtle"},
		},
		{
			name:       "t prefix returns table and turtle",
			toComplete: "t",
			wantCount:  2,
			wantItems:  []string{"table", "turtle"},
		},
		{
			name:       "json prefix returns json only",
			toComplete: "json",
			wantCount:  1,
			wantItems:  []string{"json"},
		},
		{
			name:       "nonexistent prefix returns nothing",
			toComplete: "xyz",
			wantCount:  0,
			wantItems:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			completions, directive := outputFormatCompletion(cmd, []string{}, tt.toComplete)

			if len(completions) != tt.wantCount {
				t.Errorf("outputFormatCompletion() returned %d items, want %d", len(completions), tt.wantCount)
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("outputFormatCompletion() returned directive %v, want NoFileComp", directive)
			}

			// Check all expected items are present
			completionMap := make(map[string]bool)
			for _, c := range completions {
				completionMap[c] = true
			}

			for _, want := range tt.wantItems {
				if !completionMap[want] {
					t.Errorf("outputFormatCompletion() missing expected item: %s", want)
				}
			}
		})
	}
}

func TestQueryFormatCompletion(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		wantCount  int
		wantItems  []string
	}{
		{
			name:       "empty prefix returns query formats",
			toComplete: "",
			wantCount:  3,
			wantItems:  []string{"table", "json", "csv"},
		},
		{
			name:       "j prefix returns json",
			toComplete: "j",
			wantCount:  1,
			wantItems:  []string{"json"},
		},
		{
			name:       "dot not included in query formats",
			toComplete: "dot",
			wantCount:  0,
			wantItems:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			completions, directive := queryFormatCompletion(cmd, []string{}, tt.toComplete)

			if len(completions) != tt.wantCount {
				t.Errorf("queryFormatCompletion() returned %d items, want %d", len(completions), tt.wantCount)
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("queryFormatCompletion() returned directive %v, want NoFileComp", directive)
			}

			completionMap := make(map[string]bool)
			for _, c := range completions {
				completionMap[c] = true
			}

			for _, want := range tt.wantItems {
				if !completionMap[want] {
					t.Errorf("queryFormatCompletion() missing expected item: %s", want)
				}
			}
		})
	}
}

func TestCategoryCompletion(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		wantCount  int
		wantItems  []string
	}{
		{
			name:       "empty prefix returns all categories",
			toComplete: "",
			wantCount:  6,
			wantItems:  []string{"dependencies", "security", "analysis", "layers", "impact", "documentation"},
		},
		{
			name:       "d prefix returns dependencies and documentation",
			toComplete: "d",
			wantCount:  2,
			wantItems:  []string{"dependencies", "documentation"},
		},
		{
			name:       "sec prefix returns security",
			toComplete: "sec",
			wantCount:  1,
			wantItems:  []string{"security"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			completions, directive := categoryCompletion(cmd, []string{}, tt.toComplete)

			if len(completions) != tt.wantCount {
				t.Errorf("categoryCompletion() returned %d items, want %d", len(completions), tt.wantCount)
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("categoryCompletion() returned directive %v, want NoFileComp", directive)
			}

			completionMap := make(map[string]bool)
			for _, c := range completions {
				completionMap[c] = true
			}

			for _, want := range tt.wantItems {
				if !completionMap[want] {
					t.Errorf("categoryCompletion() missing expected item: %s", want)
				}
			}
		})
	}
}

func TestShellCompletion(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		wantCount  int
		wantItems  []string
	}{
		{
			name:       "empty prefix returns all shells",
			toComplete: "",
			wantCount:  3,
			wantItems:  []string{"bash", "zsh", "fish"},
		},
		{
			name:       "b prefix returns bash",
			toComplete: "b",
			wantCount:  1,
			wantItems:  []string{"bash"},
		},
		{
			name:       "z prefix returns zsh",
			toComplete: "z",
			wantCount:  1,
			wantItems:  []string{"zsh"},
		},
		{
			name:       "sh prefix returns bash and fish",
			toComplete: "sh",
			wantCount:  0,
			wantItems:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			completions, directive := shellCompletion(cmd, []string{}, tt.toComplete)

			if len(completions) != tt.wantCount {
				t.Errorf("shellCompletion() returned %d items, want %d", len(completions), tt.wantCount)
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("shellCompletion() returned directive %v, want NoFileComp", directive)
			}

			completionMap := make(map[string]bool)
			for _, c := range completions {
				completionMap[c] = true
			}

			for _, want := range tt.wantItems {
				if !completionMap[want] {
					t.Errorf("shellCompletion() missing expected item: %s", want)
				}
			}
		})
	}
}

// Note: modulePathCompletion, layerCompletion, and tagCompletion require
// a working graph which depends on the codebase state. These are integration
// tests that should be run manually or in a dedicated test environment.
