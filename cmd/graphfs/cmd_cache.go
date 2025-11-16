/*
# Module: cmd/graphfs/cmd_cache.go
CLI commands for cache management.

Implements cache stats and management CLI commands.

## Linked Modules
- [main](./main.go) - CLI entry point

## Tags
cli, cache, commands

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_cache.go> a code:Module ;
    code:name "cmd/graphfs/cmd_cache.go" ;
    code:description "CLI commands for cache management" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <./main.go> ;
    code:tags "cli", "cache", "commands" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management commands",
	Long:  "Commands for managing GraphFS query result cache",
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Long:  "Display statistics about the query result cache from a running GraphFS server",
	RunE:  runCacheStats,
}

var cacheStatsPort int

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatsCmd)

	cacheStatsCmd.Flags().IntVarP(&cacheStatsPort, "port", "p", 8080, "Server port")
}

func runCacheStats(cmd *cobra.Command, args []string) error {
	// Construct URL
	url := fmt.Sprintf("http://localhost:%d/cache/stats", cacheStatsPort)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	// Read response
	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Display stats
	fmt.Println("Cache Statistics:")
	fmt.Println("=================")
	fmt.Printf("Hits:        %v\n", stats["hits"])
	fmt.Printf("Misses:      %v\n", stats["misses"])
	fmt.Printf("Hit Rate:    %.2f%%\n", stats["hitRate"].(float64)*100)
	fmt.Printf("Evictions:   %v\n", stats["evictions"])
	fmt.Printf("Size:        %v / %v entries\n", stats["size"], stats["maxSize"])
	fmt.Printf("Total Bytes: %v bytes\n", stats["totalBytes"])

	return nil
}
