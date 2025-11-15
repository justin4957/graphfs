/*
# Module: pkg/server/rest/tags.go
Tag and export endpoints for REST API.

Implements tag listing and export endpoints.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./handler](./handler.go) - REST handler

## Tags
rest, api, tags, exports

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#rest-tags.go> a code:Module ;
    code:name "pkg/server/rest/tags.go" ;
    code:description "Tag and export endpoints for REST API" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./handler.go> ;
    code:tags "rest", "api", "tags", "exports" .
<!-- End LinkedDoc RDF -->
*/

package rest

import (
	"net/http"
	"strings"
)

// handleTags handles GET /api/v1/tags
func (h *Handler) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Collect all unique tags with counts
	tagCounts := make(map[string]int)
	for _, mod := range h.graph.Modules {
		for _, tag := range mod.Tags {
			tagCounts[tag]++
		}
	}

	// Build response
	tags := make([]map[string]interface{}, 0, len(tagCounts))
	for tag, count := range tagCounts {
		tags = append(tags, map[string]interface{}{
			"tag":   tag,
			"count": count,
			"links": map[string]string{
				"modules": "/api/v1/tags/" + tag + "/modules",
			},
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tags":  tags,
		"total": len(tags),
	})
}

// handleTagModules handles GET /api/v1/tags/:tag/modules
func (h *Handler) handleTagModules(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract tag
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tags/")
	path = strings.TrimSuffix(path, "/modules")
	tag := path

	if tag == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_TAG", "Tag parameter is required")
		return
	}

	// Find modules with this tag
	var modules []ModuleResponse
	for _, mod := range h.graph.Modules {
		for _, t := range mod.Tags {
			if t == tag {
				modules = append(modules, h.toModuleResponse(mod, false))
				break
			}
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tag":     tag,
		"modules": modules,
		"count":   len(modules),
	})
}

// handleExports handles GET /api/v1/exports
func (h *Handler) handleExports(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Check for search query
	query := r.URL.Query().Get("name")

	// Collect all exports
	type ExportInfo struct {
		Name   string         `json:"name"`
		Module ModuleResponse `json:"module"`
	}

	var exports []ExportInfo
	for _, mod := range h.graph.Modules {
		for _, exp := range mod.Exports {
			// Filter by query if provided
			if query != "" && !strings.Contains(strings.ToLower(exp), strings.ToLower(query)) {
				continue
			}

			exports = append(exports, ExportInfo{
				Name:   exp,
				Module: h.toModuleResponse(mod, false),
			})
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"exports": exports,
		"count":   len(exports),
	})
}
