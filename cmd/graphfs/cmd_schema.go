/*
# Module: cmd/graphfs/cmd_schema.go
CLI command to generate GraphQL schema.

Generates GraphQL Schema Definition Language (SDL) from knowledge graph.

## Linked Modules
- [../../pkg/schema/graphql](../../pkg/schema/graphql/generator.go) - Schema generator
- [../../pkg/graph](../../pkg/graph/graph.go) - Graph builder

## Tags
cli, schema, graphql, command

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_schema.go> a code:Module ;
    code:name "cmd/graphfs/cmd_schema.go" ;
    code:description "CLI command to generate GraphQL schema" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <../../pkg/schema/graphql/generator.go>, <../../pkg/graph/graph.go> ;
    code:tags "cli", "schema", "graphql", "command" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/scanner"
	graphqlschema "github.com/justin4957/graphfs/pkg/schema/graphql"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate GraphQL schema from knowledge graph",
	Long: `Generate GraphQL Schema Definition Language (SDL) from the knowledge graph.

The schema command scans your codebase, builds the knowledge graph, and generates
a GraphQL schema that can be used for querying via GraphQL.

Examples:
  # Generate schema and print to stdout
  graphfs schema generate

  # Generate schema and save to file
  graphfs schema generate --output schema.graphql

  # Generate with specific format
  graphfs schema generate --format graphql --output api/schema.graphql

  # List available types
  graphfs schema types
`,
}

var generateSchemaCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GraphQL schema",
	Long: `Generate GraphQL Schema Definition Language (SDL) from the knowledge graph.

The schema is generated from RDF triples in the knowledge graph and includes:
- Module type with all fields and relationships
- Export type for exported symbols
- Query root type with filtering and pagination
- Connection types for cursor-based pagination
- Statistics types for graph metrics

Examples:
  # Print schema to stdout
  graphfs schema generate

  # Save to file
  graphfs schema generate --output schema.graphql

  # Generate and validate
  graphfs schema generate --validate
`,
	RunE: runGenerateSchema,
}

var schemaTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available GraphQL types",
	Long:  `List all GraphQL types that will be generated from the knowledge graph.`,
	RunE:  runSchemaTypes,
}

var (
	schemaOutput   string
	schemaFormat   string
	schemaValidate bool
)

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.AddCommand(generateSchemaCmd)
	schemaCmd.AddCommand(schemaTypesCmd)

	generateSchemaCmd.Flags().StringVarP(&schemaOutput, "output", "o", "", "Output file path (default: stdout)")
	generateSchemaCmd.Flags().StringVar(&schemaFormat, "format", "graphql", "Output format (graphql)")
	generateSchemaCmd.Flags().BoolVar(&schemaValidate, "validate", false, "Validate generated schema")
}

func runGenerateSchema(cmd *cobra.Command, args []string) error {
	// Get root path
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration
	configPath := filepath.Join(rootPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Scanning codebase and building graph...")

	// Build graph
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			IncludePatterns: config.Scan.Include,
			ExcludePatterns: config.Scan.Exclude,
			MaxFileSize:     config.Scan.MaxFileSize,
			UseDefaults:     true,
			IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
			Concurrent:      true,
		},
		Validate: true,
	}

	g, err := builder.Build(rootPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	fmt.Printf("Built graph with %d modules\n", len(g.Modules))

	// Generate schema
	fmt.Println("Generating GraphQL schema...")
	schema, err := graphqlschema.GenerateSchemaFromGraph(g, graphqlschema.DefaultGenerateOptions())
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}

	// Validate if requested
	if schemaValidate {
		gen := graphqlschema.NewGenerator(g)
		if err := gen.ValidateSchema(schema); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		fmt.Println("✓ Schema validation passed")
	}

	// Write output
	if schemaOutput != "" {
		// Ensure directory exists
		dir := filepath.Dir(schemaOutput)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write to file
		if err := os.WriteFile(schemaOutput, []byte(schema), 0644); err != nil {
			return fmt.Errorf("failed to write schema file: %w", err)
		}

		fmt.Printf("✓ Schema written to %s\n", schemaOutput)
	} else {
		// Print to stdout
		fmt.Println()
		fmt.Println(schema)
	}

	return nil
}

func runSchemaTypes(cmd *cobra.Command, args []string) error {
	// Get root path
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration
	configPath := filepath.Join(rootPath, ".graphfs", "config.yaml")
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build graph (minimal build just to get structure)
	builder := graph.NewBuilder()
	buildOpts := graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			IncludePatterns: config.Scan.Include,
			ExcludePatterns: config.Scan.Exclude,
			MaxFileSize:     config.Scan.MaxFileSize,
			UseDefaults:     true,
			IgnoreFiles:     []string{".gitignore", ".graphfsignore"},
			Concurrent:      true,
		},
		Validate: true,
	}

	g, err := builder.Build(rootPath, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// Get available types
	gen := graphqlschema.NewGenerator(g)
	types := gen.GetAvailableTypes()

	fmt.Println("Available GraphQL Types:")
	fmt.Println()
	for _, typeName := range types {
		fmt.Printf("  - %s\n", typeName)
	}
	fmt.Println()
	fmt.Printf("Total: %d types\n", len(types))

	return nil
}
