/*
# Module: pkg/server/rest/handler.go
REST API handler for GraphFS.

Provides RESTful HTTP endpoints for common code analysis queries.

## Linked Modules
- [../../graph](../../graph/graph.go) - Graph data structure

## Tags
rest, api, http, server

## Exports
Handler, NewHandler

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#rest-handler.go> a code:Module ;
    code:name "pkg/server/rest/handler.go" ;
    code:description "REST API handler for GraphFS" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../../graph/graph.go> ;
    code:exports <#Handler>, <#NewHandler> ;
    code:tags "rest", "api", "http", "server" .
<!-- End LinkedDoc RDF -->
*/

package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/justin4957/graphfs/pkg/cache"
	"github.com/justin4957/graphfs/pkg/graph"
)

// Handler handles REST API requests
type Handler struct {
	graph      *graph.Graph
	enableCORS bool
}

// NewHandler creates a new REST API handler
func NewHandler(g *graph.Graph, enableCORS bool) *Handler {
	return &Handler{
		graph:      g,
		enableCORS: enableCORS,
	}
}

// RegisterRoutes registers all REST API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Module endpoints - register specific paths before wildcards
	mux.HandleFunc("/api/v1/modules/search", h.handleModulesSearch)
	mux.HandleFunc("/api/v1/modules/", h.handleModulesWithID)
	mux.HandleFunc("/api/v1/modules", h.handleModules)

	// Analysis endpoints
	mux.HandleFunc("/api/v1/analysis/stats", h.handleAnalysisStats)
	mux.HandleFunc("/api/v1/analysis/impact/", h.handleAnalysisImpact)

	// Tag endpoints
	mux.HandleFunc("/api/v1/tags/", h.handleTagModules)
	mux.HandleFunc("/api/v1/tags", h.handleTags)

	// Export endpoints
	mux.HandleFunc("/api/v1/exports", h.handleExports)
}

// RegisterRoutesWithCache registers all REST API routes with caching
func (h *Handler) RegisterRoutesWithCache(mux *http.ServeMux, c *cache.Cache) {
	// Import the server package type
	cacheMiddleware := func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip caching for non-GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := cache.GenerateKey(r.URL.Path, r.URL.RawQuery)

			// Check cache
			if cached, found := c.Get(cacheKey); found {
				// Write cached response
				if cachedResp, ok := cached.(map[string]interface{}); ok {
					if headers, ok := cachedResp["headers"].(http.Header); ok {
						for key, values := range headers {
							for _, value := range values {
								w.Header().Add(key, value)
							}
						}
					}
					w.Header().Set("X-Cache", "HIT")
					if statusCode, ok := cachedResp["statusCode"].(int); ok {
						w.WriteHeader(statusCode)
					}
					if body, ok := cachedResp["body"].([]byte); ok {
						w.Write(body)
					}
					return
				}
			}

			// Cache miss - capture response
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           make([]byte, 0),
			}

			// Set cache miss header before executing handler
			recorder.Header().Set("X-Cache", "MISS")

			next.ServeHTTP(recorder, r)

			// Cache successful responses
			if recorder.statusCode == http.StatusOK {
				cachedResp := map[string]interface{}{
					"statusCode": recorder.statusCode,
					"headers":    recorder.Header().Clone(),
					"body":       recorder.body,
				}
				c.Set(cacheKey, cachedResp, int64(len(recorder.body)))
			}
		})
	}

	// Module endpoints with caching
	mux.Handle("/api/v1/modules/search", cacheMiddleware(h.handleModulesSearch))
	mux.Handle("/api/v1/modules/", cacheMiddleware(h.handleModulesWithID))
	mux.Handle("/api/v1/modules", cacheMiddleware(h.handleModules))

	// Analysis endpoints with caching
	mux.Handle("/api/v1/analysis/stats", cacheMiddleware(h.handleAnalysisStats))
	mux.Handle("/api/v1/analysis/impact/", cacheMiddleware(h.handleAnalysisImpact))

	// Tag endpoints with caching
	mux.Handle("/api/v1/tags/", cacheMiddleware(h.handleTagModules))
	mux.Handle("/api/v1/tags", cacheMiddleware(h.handleTags))

	// Export endpoints with caching
	mux.Handle("/api/v1/exports", cacheMiddleware(h.handleExports))
}

// responseRecorder captures the HTTP response
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if h.enableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"status":  statusCode,
		},
	})
}

// parseQueryParams parses common query parameters
func (h *Handler) parseQueryParams(r *http.Request) (language, layer, tag string, limit, offset int) {
	language = r.URL.Query().Get("language")
	layer = r.URL.Query().Get("layer")
	tag = r.URL.Query().Get("tag")

	limit = 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset = 0 // default
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return
}

// filterModules filters modules based on criteria
func (h *Handler) filterModules(language, layer, tag string) []*graph.Module {
	var filtered []*graph.Module

	for _, mod := range h.graph.Modules {
		// Apply filters
		if language != "" && mod.Language != language {
			continue
		}
		if layer != "" && mod.Layer != layer {
			continue
		}
		if tag != "" {
			hasTag := false
			for _, t := range mod.Tags {
				if t == tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		filtered = append(filtered, mod)
	}

	return filtered
}
