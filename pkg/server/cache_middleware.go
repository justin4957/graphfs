/*
# Module: pkg/server/cache_middleware.go
HTTP response caching middleware.

Provides caching middleware for HTTP handlers with LRU eviction.

## Linked Modules
- [../cache](../cache/cache.go) - Cache implementation

## Tags
server, cache, middleware, http

## Exports
CacheMiddleware, responseWriter

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cache_middleware.go> a code:Module ;
    code:name "pkg/server/cache_middleware.go" ;
    code:description "HTTP response caching middleware" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../cache/cache.go> ;
    code:exports <#CacheMiddleware>, <#responseWriter> ;
    code:tags "server", "cache", "middleware", "http" .
<!-- End LinkedDoc RDF -->
*/

package server

import (
	"bytes"
	"net/http"

	"github.com/justin4957/graphfs/pkg/cache"
)

// responseWriter captures the response for caching
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// CacheMiddleware wraps an HTTP handler with caching support
func CacheMiddleware(next http.Handler, c *cache.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip caching for non-GET/POST requests
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching if OPTIONS request
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key based on URL and query parameters
		cacheKey := cache.GenerateKey(r.Method, r.URL.String(), r.URL.Query().Get("query"))

		// Check cache
		if cached, found := c.Get(cacheKey); found {
			// Cache hit
			cachedResp := cached.(map[string]interface{})

			// Set headers
			if headers, ok := cachedResp["headers"].(http.Header); ok {
				for key, values := range headers {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
			}

			w.Header().Set("X-Cache", "HIT")

			// Write status code
			if statusCode, ok := cachedResp["statusCode"].(int); ok {
				w.WriteHeader(statusCode)
			}

			// Write body
			if body, ok := cachedResp["body"].([]byte); ok {
				w.Write(body)
			}
			return
		}

		// Cache miss - capture response
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Only cache successful responses
		if rw.statusCode == http.StatusOK {
			// Store in cache
			cachedResp := map[string]interface{}{
				"statusCode": rw.statusCode,
				"headers":    rw.Header().Clone(),
				"body":       rw.body.Bytes(),
			}

			c.Set(cacheKey, cachedResp, int64(rw.body.Len()))

			// Set cache miss header
			w.Header().Set("X-Cache", "MISS")
		}
	})
}
