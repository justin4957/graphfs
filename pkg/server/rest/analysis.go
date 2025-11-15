/*
# Module: pkg/server/rest/analysis.go
Analysis endpoints for REST API.

Implements statistics and impact analysis endpoints.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./handler](./handler.go) - REST handler

## Tags
rest, api, analysis

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#rest-analysis.go> a code:Module ;
    code:name "pkg/server/rest/analysis.go" ;
    code:description "Analysis endpoints for REST API" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./handler.go> ;
    code:tags "rest", "api", "analysis" .
<!-- End LinkedDoc RDF -->
*/

package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
)

// handleAnalysisStats handles GET /api/v1/analysis/stats
func (h *Handler) handleAnalysisStats(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Count modules by language
	languageCounts := make(map[string]int)
	for _, mod := range h.graph.Modules {
		if mod.Language != "" {
			languageCounts[mod.Language]++
		}
	}

	var modulesByLanguage []map[string]interface{}
	for lang, count := range languageCounts {
		modulesByLanguage = append(modulesByLanguage, map[string]interface{}{
			"language": lang,
			"count":    count,
		})
	}

	// Count modules by layer
	layerCounts := make(map[string]int)
	for _, mod := range h.graph.Modules {
		if mod.Layer != "" {
			layerCounts[mod.Layer]++
		}
	}

	var modulesByLayer []map[string]interface{}
	for layer, count := range layerCounts {
		modulesByLayer = append(modulesByLayer, map[string]interface{}{
			"layer": layer,
			"count": count,
		})
	}

	// Count total relationships
	totalRelationships := 0
	for _, mod := range h.graph.Modules {
		totalRelationships += len(mod.Dependencies)
	}

	// Count all tags
	tagCounts := make(map[string]int)
	for _, mod := range h.graph.Modules {
		for _, tag := range mod.Tags {
			tagCounts[tag]++
		}
	}

	// Count all exports
	totalExports := 0
	for _, mod := range h.graph.Modules {
		totalExports += len(mod.Exports)
	}

	stats := map[string]interface{}{
		"totalModules":        len(h.graph.Modules),
		"totalTriples":        h.graph.Store.Count(),
		"totalRelationships":  totalRelationships,
		"totalExports":        totalExports,
		"totalTags":           len(tagCounts),
		"modulesByLanguage":   modulesByLanguage,
		"modulesByLayer":      modulesByLayer,
		"mostCommonLanguage":  getMostCommon(languageCounts),
		"mostCommonLayer":     getMostCommon(layerCounts),
		"averageDependencies": float64(totalRelationships) / float64(len(h.graph.Modules)),
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// handleAnalysisImpact handles GET /api/v1/analysis/impact/:module
func (h *Handler) handleAnalysisImpact(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract module ID
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/analysis/impact/")
	moduleID := path

	// Parse depth parameter
	depth := 1 // default
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if parsed, err := strconv.Atoi(depthStr); err == nil && parsed > 0 {
			depth = parsed
			if depth > 10 {
				depth = 10 // max depth
			}
		}
	}

	// Find module
	var module *graph.Module
	for _, mod := range h.graph.Modules {
		if mod.URI == moduleID || mod.Path == moduleID || mod.Name == moduleID {
			module = mod
			break
		}
	}

	if module == nil {
		h.writeError(w, http.StatusNotFound, "MODULE_NOT_FOUND",
			fmt.Sprintf("Module '%s' not found", moduleID))
		return
	}

	// Calculate impact - get all dependents recursively
	impactedModules := h.getImpactedModules(module, depth)

	// Convert to response format
	impacted := make([]ModuleResponse, len(impactedModules))
	for i, mod := range impactedModules {
		impacted[i] = h.toModuleResponse(mod, false)
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"module":           h.toModuleResponse(module, false),
		"depth":            depth,
		"impactedModules":  impacted,
		"impactCount":      len(impacted),
		"directDependents": len(module.Dependents),
	})
}

// getImpactedModules gets all modules impacted by changes to the given module
func (h *Handler) getImpactedModules(module *graph.Module, maxDepth int) []*graph.Module {
	visited := make(map[string]bool)
	var result []*graph.Module

	var traverse func(*graph.Module, int)
	traverse = func(mod *graph.Module, depth int) {
		if depth > maxDepth {
			return
		}

		for _, depPath := range mod.Dependents {
			if visited[depPath] {
				continue
			}

			depMod := h.graph.GetModule(depPath)
			if depMod == nil {
				continue
			}

			visited[depPath] = true
			result = append(result, depMod)

			// Recurse
			traverse(depMod, depth+1)
		}
	}

	traverse(module, 0)
	return result
}

// getMostCommon returns the most common key in a map
func getMostCommon(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}

	maxCount := 0
	var maxKey string
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			maxKey = key
		}
	}
	return maxKey
}
