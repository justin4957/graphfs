/*
# Module: pkg/server/sparql_handler.go
HTTP handler for SPARQL queries.

Handles HTTP requests for SPARQL queries with multiple output formats.

## Linked Modules
- [../query](../query/executor.go) - Query executor

## Tags
server, sparql, http, handler

## Exports
SPARQLHandler, NewSPARQLHandler

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#sparql_handler.go> a code:Module ;
    code:name "pkg/server/sparql_handler.go" ;
    code:description "HTTP handler for SPARQL queries" ;
    code:language "go" ;
    code:layer "server" ;
    code:linksTo <../query/executor.go> ;
    code:exports <#SPARQLHandler>, <#NewSPARQLHandler> ;
    code:tags "server", "sparql", "http", "handler" .
<!-- End LinkedDoc RDF -->
*/

package server

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/justin4957/graphfs/pkg/query"
)

// SPARQLHandler handles SPARQL query requests
type SPARQLHandler struct {
	executor   *query.Executor
	enableCORS bool
}

// NewSPARQLHandler creates a new SPARQL handler
func NewSPARQLHandler(executor *query.Executor, enableCORS bool) *SPARQLHandler {
	return &SPARQLHandler{
		executor:   executor,
		enableCORS: enableCORS,
	}
}

// ServeHTTP handles HTTP requests for SPARQL queries
func (h *SPARQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS if configured
	if h.enableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
	}

	// Handle OPTIONS for CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept GET and POST
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract query from request
	queryStr, err := h.extractQuery(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if queryStr == "" {
		h.writeError(w, http.StatusBadRequest, "Missing query parameter")
		return
	}

	// Execute query
	result, err := h.executor.ExecuteString(queryStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Query execution failed: %v", err))
		return
	}

	// Determine output format
	format := h.determineFormat(r)

	// Write response
	if err := h.writeResult(w, result, format); err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to write response: %v", err))
	}
}

// extractQuery extracts the SPARQL query from the request
func (h *SPARQLHandler) extractQuery(r *http.Request) (string, error) {
	if r.Method == http.MethodGet {
		return r.URL.Query().Get("query"), nil
	}

	// POST request
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/sparql-query") {
		// Direct SPARQL query in body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read request body: %w", err)
		}
		return string(body), nil
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Form-encoded query
		if err := r.ParseForm(); err != nil {
			return "", fmt.Errorf("failed to parse form: %w", err)
		}
		return r.FormValue("query"), nil
	}

	// Default: try to read as raw query
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}
	return string(body), nil
}

// determineFormat determines the output format from Accept header or query parameter
func (h *SPARQLHandler) determineFormat(r *http.Request) string {
	// Check query parameter first
	if format := r.URL.Query().Get("format"); format != "" {
		return format
	}

	// Check Accept header
	accept := r.Header.Get("Accept")

	if strings.Contains(accept, "application/sparql-results+json") || strings.Contains(accept, "application/json") {
		return "json"
	}
	if strings.Contains(accept, "application/sparql-results+xml") || strings.Contains(accept, "application/xml") {
		return "xml"
	}
	if strings.Contains(accept, "text/csv") {
		return "csv"
	}
	if strings.Contains(accept, "text/tab-separated-values") {
		return "tsv"
	}

	// Default to JSON
	return "json"
}

// writeResult writes the query result in the requested format
func (h *SPARQLHandler) writeResult(w http.ResponseWriter, result *query.QueryResult, format string) error {
	switch format {
	case "json":
		return h.writeJSON(w, result)
	case "csv":
		return h.writeCSV(w, result)
	case "tsv":
		return h.writeTSV(w, result)
	case "xml":
		return h.writeXML(w, result)
	default:
		return h.writeJSON(w, result)
	}
}

// writeJSON writes result as JSON
func (h *SPARQLHandler) writeJSON(w http.ResponseWriter, result *query.QueryResult) error {
	w.Header().Set("Content-Type", "application/sparql-results+json")

	response := map[string]interface{}{
		"head": map[string]interface{}{
			"vars": result.Variables,
		},
		"results": map[string]interface{}{
			"bindings": h.formatBindingsForJSON(result),
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatBindingsForJSON formats bindings for SPARQL JSON results format
func (h *SPARQLHandler) formatBindingsForJSON(result *query.QueryResult) []map[string]interface{} {
	var bindings []map[string]interface{}

	for _, binding := range result.Bindings {
		row := make(map[string]interface{})
		for _, varName := range result.Variables {
			if value, ok := binding[varName]; ok {
				row[varName] = map[string]string{
					"type":  "literal",
					"value": value,
				}
			}
		}
		bindings = append(bindings, row)
	}

	return bindings
}

// writeCSV writes result as CSV
func (h *SPARQLHandler) writeCSV(w http.ResponseWriter, result *query.QueryResult) error {
	w.Header().Set("Content-Type", "text/csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	if err := writer.Write(result.Variables); err != nil {
		return err
	}

	// Write rows
	for _, binding := range result.Bindings {
		row := make([]string, len(result.Variables))
		for i, varName := range result.Variables {
			row[i] = binding[varName]
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// writeTSV writes result as TSV
func (h *SPARQLHandler) writeTSV(w http.ResponseWriter, result *query.QueryResult) error {
	w.Header().Set("Content-Type", "text/tab-separated-values")

	// Write header
	fmt.Fprintln(w, strings.Join(result.Variables, "\t"))

	// Write rows
	for _, binding := range result.Bindings {
		row := make([]string, len(result.Variables))
		for i, varName := range result.Variables {
			row[i] = binding[varName]
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

// SPARQLResultsXML represents the XML structure for SPARQL results
type SPARQLResultsXML struct {
	XMLName xml.Name `xml:"sparql"`
	Xmlns   string   `xml:"xmlns,attr"`
	Head    struct {
		Variables []struct {
			Name string `xml:"name,attr"`
		} `xml:"variable"`
	} `xml:"head"`
	Results struct {
		Results []ResultXML `xml:"result"`
	} `xml:"results"`
}

// ResultXML represents a single result in XML
type ResultXML struct {
	Bindings []BindingXML `xml:"binding"`
}

// BindingXML represents a variable binding in XML
type BindingXML struct {
	Name    string `xml:"name,attr"`
	Literal string `xml:"literal"`
}

// writeXML writes result as SPARQL Results XML
func (h *SPARQLHandler) writeXML(w http.ResponseWriter, result *query.QueryResult) error {
	w.Header().Set("Content-Type", "application/sparql-results+xml")

	xmlResult := SPARQLResultsXML{
		Xmlns: "http://www.w3.org/2005/sparql-results#",
	}

	// Add variables
	for _, varName := range result.Variables {
		xmlResult.Head.Variables = append(xmlResult.Head.Variables, struct {
			Name string `xml:"name,attr"`
		}{Name: varName})
	}

	// Add results
	for _, binding := range result.Bindings {
		resultXML := ResultXML{}
		for _, varName := range result.Variables {
			if value, ok := binding[varName]; ok {
				resultXML.Bindings = append(resultXML.Bindings, BindingXML{
					Name:    varName,
					Literal: value,
				})
			}
		}
		xmlResult.Results.Results = append(xmlResult.Results.Results, resultXML)
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	w.Write([]byte(xml.Header))
	return encoder.Encode(xmlResult)
}

// writeError writes an error response
func (h *SPARQLHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"error": message,
	}

	json.NewEncoder(w).Encode(response)
}
