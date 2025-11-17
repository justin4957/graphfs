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

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management commands",
	Long: `Commands for managing GraphFS cache.

Available subcommands:
  stats - Show local persistent cache statistics
  clear - Clear local persistent cache
  server-stats - Show query result cache statistics from running server`,
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show local cache statistics",
	Long:  "Display statistics about the persistent module cache",
	RunE:  runLocalCacheStats,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear local cache",
	Long:  "Clear all cached modules from the persistent cache",
	RunE:  runCacheClear,
}

var cacheServerStatsCmd = &cobra.Command{
	Use:   "server-stats",
	Short: "Show server cache statistics",
	Long:  "Display statistics about the query result cache from a running GraphFS server",
	RunE:  runCacheStats,
}

var (
	cacheStatsPort int
	cacheTarget    string
)

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheServerStatsCmd)

	cacheStatsCmd.Flags().StringVarP(&cacheTarget, "target", "t", ".", "Target directory")
	cacheClearCmd.Flags().StringVarP(&cacheTarget, "target", "t", ".", "Target directory")
	cacheServerStatsCmd.Flags().IntVarP(&cacheStatsPort, "port", "p", 8080, "Server port")
}

func runLocalCacheStats(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("📊 Persistent Cache Statistics")
	fmt.Println()

	// Open cache manager
	cacheManager, err := cache.NewManager(cacheTarget)
	if err != nil {
		return fmt.Errorf("failed to open cache: %w", err)
	}
	defer cacheManager.Close()

	// Get statistics
	stats, err := cacheManager.Stats()
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	// Display statistics
	fmt.Printf("Modules Cached:  %d\n", stats.ModuleCount)
	fmt.Printf("Cache Hits:      %d\n", stats.CacheHits)
	fmt.Printf("Cache Misses:    %d\n", stats.CacheMisses)

	total := stats.CacheHits + stats.CacheMisses
	if total > 0 {
		green.Printf("Hit Rate:        %.1f%%\n", stats.HitRate*100)
	} else {
		fmt.Println("Hit Rate:        N/A (no queries yet)")
	}

	// Format cache size
	sizeKB := float64(stats.CacheSize) / 1024
	sizeMB := sizeKB / 1024
	if sizeMB >= 1 {
		fmt.Printf("Cache Size:      %.2f MB\n", sizeMB)
	} else {
		fmt.Printf("Cache Size:      %.2f KB\n", sizeKB)
	}

	fmt.Printf("Last Updated:    %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nCache Location:  %s/.graphfs/cache/\n\n", cacheTarget)

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("🗑️  Clearing Persistent Cache")
	fmt.Println()

	// Open cache manager
	cacheManager, err := cache.NewManager(cacheTarget)
	if err != nil {
		return fmt.Errorf("failed to open cache: %w", err)
	}
	defer cacheManager.Close()

	// Get stats before clearing
	statsBefore, _ := cacheManager.Stats()

	// Clear cache
	if err := cacheManager.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	green.Printf("✓ Cache cleared successfully\n")
	fmt.Printf("  Removed %d cached modules\n", statsBefore.ModuleCount)

	sizeKB := float64(statsBefore.CacheSize) / 1024
	if sizeKB >= 1024 {
		fmt.Printf("  Freed %.2f MB\n\n", sizeKB/1024)
	} else {
		fmt.Printf("  Freed %.2f KB\n\n", sizeKB)
	}

	return nil
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
	cyan := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	cyan.Println("📊 Server Query Cache Statistics")
	fmt.Println()

	fmt.Printf("Hits:        %v\n", stats["hits"])
	fmt.Printf("Misses:      %v\n", stats["misses"])
	fmt.Printf("Hit Rate:    %.2f%%\n", stats["hitRate"].(float64)*100)
	fmt.Printf("Evictions:   %v\n", stats["evictions"])
	fmt.Printf("Size:        %v / %v entries\n", stats["size"], stats["maxSize"])
	fmt.Printf("Total Bytes: %v bytes\n\n", stats["totalBytes"])

	return nil
}
