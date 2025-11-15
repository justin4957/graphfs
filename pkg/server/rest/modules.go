/*
# Module: pkg/server/rest/modules.go
Module endpoints for REST API.

Implements module listing, retrieval, search, and relationship queries.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure
- [./handler](./handler.go) - REST handler

## Tags
rest, api, modules

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#rest-modules.go> a code:Module ;
    code:name "pkg/server/rest/modules.go" ;
    code:description "Module endpoints for REST API" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go>, <./handler.go> ;
    code:tags "rest", "api", "modules" .
<!-- End LinkedDoc RDF -->
*/

package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
)

// ModuleResponse represents a module in the API response
type ModuleResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Description  string            `json:"description,omitempty"`
	Language     string            `json:"language,omitempty"`
	Layer        string            `json:"layer,omitempty"`
	Tags         []string          `json:"tags"`
	Exports      []string          `json:"exports"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Dependents   []string          `json:"dependents,omitempty"`
	Links        map[string]string `json:"links"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data  interface{}       `json:"data"`
	Meta  map[string]int    `json:"meta"`
	Links map[string]string `json:"links,omitempty"`
}

// toModuleResponse converts a graph.Module to API response format
func (h *Handler) toModuleResponse(mod *graph.Module, includeDeps bool) ModuleResponse {
	resp := ModuleResponse{
		ID:          mod.URI,
		Name:        mod.Name,
		Path:        mod.Path,
		Description: mod.Description,
		Language:    mod.Language,
		Layer:       mod.Layer,
		Tags:        mod.Tags,
		Exports:     mod.Exports,
		Links: map[string]string{
			"self":         fmt.Sprintf("/api/v1/modules/%s", mod.URI),
			"dependencies": fmt.Sprintf("/api/v1/modules/%s/dependencies", mod.URI),
			"dependents":   fmt.Sprintf("/api/v1/modules/%s/dependents", mod.URI),
		},
	}

	if mod.Tags == nil {
		resp.Tags = []string{}
	}
	if mod.Exports == nil {
		resp.Exports = []string{}
	}

	if includeDeps {
		resp.Dependencies = mod.Dependencies
		resp.Dependents = mod.Dependents
		if resp.Dependencies == nil {
			resp.Dependencies = []string{}
		}
		if resp.Dependents == nil {
			resp.Dependents = []string{}
		}
	}

	return resp
}

// handleModules handles GET /api/v1/modules
func (h *Handler) handleModules(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Parse query parameters
	language, layer, tag, limit, offset := h.parseQueryParams(r)

	// Filter modules
	filtered := h.filterModules(language, layer, tag)

	// Apply pagination
	total := len(filtered)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	// Convert to response format
	modules := make([]ModuleResponse, len(paginated))
	for i, mod := range paginated {
		modules[i] = h.toModuleResponse(mod, false)
	}

	// Build response
	response := ListResponse{
		Data: modules,
		Meta: map[string]int{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"count":  len(modules),
		},
	}

	// Add pagination links
	links := make(map[string]string)
	queryBase := buildQueryString(r, language, layer, tag)

	links["self"] = fmt.Sprintf("/api/v1/modules?%slimit=%d&offset=%d", queryBase, limit, offset)

	if end < total {
		links["next"] = fmt.Sprintf("/api/v1/modules?%slimit=%d&offset=%d", queryBase, limit, end)
	}
	if start > 0 {
		prevOffset := start - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links["prev"] = fmt.Sprintf("/api/v1/modules?%slimit=%d&offset=%d", queryBase, limit, prevOffset)
	}

	response.Links = links

	h.writeJSON(w, http.StatusOK, response)
}

// handleModulesWithID handles GET /api/v1/modules/:id and subpaths
func (h *Handler) handleModulesWithID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract module ID and check for subpaths
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/v1/modules/")

	// Handle subpaths
	if strings.Contains(path, "/dependencies") {
		h.handleModuleDependencies(w, r, path)
		return
	}
	if strings.Contains(path, "/dependents") {
		h.handleModuleDependents(w, r, path)
		return
	}

	// Get single module
	moduleID := path

	// Find module by URI, path, or name
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

	h.writeJSON(w, http.StatusOK, h.toModuleResponse(module, true))
}

// handleModuleDependencies handles GET /api/v1/modules/:id/dependencies
func (h *Handler) handleModuleDependencies(w http.ResponseWriter, r *http.Request, path string) {
	moduleID := strings.Split(path, "/dependencies")[0]

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

	// Get dependencies
	deps := make([]ModuleResponse, 0)
	for _, depPath := range module.Dependencies {
		if depMod := h.graph.GetModule(depPath); depMod != nil {
			deps = append(deps, h.toModuleResponse(depMod, false))
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"module":       h.toModuleResponse(module, false),
		"dependencies": deps,
		"count":        len(deps),
	})
}

// handleModuleDependents handles GET /api/v1/modules/:id/dependents
func (h *Handler) handleModuleDependents(w http.ResponseWriter, r *http.Request, path string) {
	moduleID := strings.Split(path, "/dependents")[0]

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

	// Get dependents
	deps := make([]ModuleResponse, 0)
	for _, depPath := range module.Dependents {
		if depMod := h.graph.GetModule(depPath); depMod != nil {
			deps = append(deps, h.toModuleResponse(depMod, false))
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"module":     h.toModuleResponse(module, false),
		"dependents": deps,
		"count":      len(deps),
	})
}

// handleModulesSearch handles GET /api/v1/modules/search
func (h *Handler) handleModulesSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		h.writeJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method != "GET" {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_QUERY", "Query parameter 'q' is required")
		return
	}

	queryLower := strings.ToLower(query)
	var results []*graph.Module

	for _, mod := range h.graph.Modules {
		// Search in description
		if strings.Contains(strings.ToLower(mod.Description), queryLower) {
			results = append(results, mod)
			continue
		}

		// Search in name
		if strings.Contains(strings.ToLower(mod.Name), queryLower) {
			results = append(results, mod)
			continue
		}

		// Search in path
		if strings.Contains(strings.ToLower(mod.Path), queryLower) {
			results = append(results, mod)
			continue
		}

		// Search in tags
		for _, tag := range mod.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, mod)
				break
			}
		}
	}

	// Convert to response format
	modules := make([]ModuleResponse, len(results))
	for i, mod := range results {
		modules[i] = h.toModuleResponse(mod, false)
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"results": modules,
		"count":   len(modules),
	})
}

// buildQueryString builds a query string from filter parameters
func buildQueryString(r *http.Request, language, layer, tag string) string {
	var parts []string
	if language != "" {
		parts = append(parts, fmt.Sprintf("language=%s", language))
	}
	if layer != "" {
		parts = append(parts, fmt.Sprintf("layer=%s", layer))
	}
	if tag != "" {
		parts = append(parts, fmt.Sprintf("tag=%s", tag))
	}
	if len(parts) > 0 {
		return strings.Join(parts, "&") + "&"
	}
	return ""
}
