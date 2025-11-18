/*
# Module: cmd/graphfs/root.go
Root command for GraphFS CLI.

Defines the root command with global flags and version information.

## Linked Modules
- [main](./main.go) - CLI entry point
- [config](./config.go) - Configuration handling

## Tags
cli, root, cobra

## Exports
rootCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#root.go> a code:Module ;

	code:name "cmd/graphfs/root.go" ;
	code:description "Root command for GraphFS CLI" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./main.go>, <./config.go> ;
	code:exports <#rootCmd> ;
	code:tags "cli", "root", "cobra" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	// Version is the current version of GraphFS
	Version = "0.1.0"
	// Name is the application name
	Name = "GraphFS"
)

var (
	cfgFile string
	verbose bool
	noColor bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "graphfs",
	Short: "Semantic Code Filesystem Toolkit",
	Long: `GraphFS - Semantic Code Filesystem Toolkit

GraphFS transforms your codebase into a queryable knowledge graph using
LinkedDoc metadata. Build, query, and analyze your code architecture with
SPARQL and GraphQL.

For more information, visit: https://github.com/justin4957/graphfs`,
	Version: Version,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .graphfs/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "minimal output (for scripting)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(examplesCmd)
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s v%s\n", Name, Version)
	},
}
