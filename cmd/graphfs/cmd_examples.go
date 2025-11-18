/*
# Module: cmd/graphfs/cmd_examples.go
Examples command implementation for query templates.

Provides commands to list, show, run, save, and export query templates.

## Linked Modules
- [root](./root.go) - Root command
- [../../pkg/query](../../pkg/query/templates.go) - Query templates
- [../../pkg/cli](../../pkg/cli/output.go) - Output formatting

## Tags
cli, command, examples, templates

## Exports
examplesCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_examples.go> a code:Module ;
    code:name "cmd/graphfs/cmd_examples.go" ;
    code:description "Examples command implementation for query templates" ;
    code:language "go" ;
    code:layer "cli" ;
    code:linksTo <./root.go>, <../../pkg/query/templates.go>, <../../pkg/cli/output.go> ;
    code:exports <#examplesCmd> ;
    code:tags "cli", "command", "examples", "templates" .
<!-- End LinkedDoc RDF -->
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/justin4957/graphfs/pkg/cli"
	"github.com/justin4957/graphfs/pkg/graph"
	"github.com/justin4957/graphfs/pkg/query"
	"github.com/justin4957/graphfs/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	examplesCategory string
	examplesOutput   string
)

// examplesCmd represents the examples command
var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Browse and run query templates",
	Long: `Browse and run pre-built query templates and examples.

Query templates provide common patterns to help you query the knowledge graph
without memorizing SPARQL syntax.

Examples:
  # List all templates
  graphfs examples list

  # List templates by category
  graphfs examples list --category dependencies

  # Show template details
  graphfs examples show find-dependencies

  # Run a template
  graphfs examples run find-dependencies --module=api/handlers.go

  # Save custom template
  graphfs examples save my-query --query="SELECT * WHERE {...}"

  # Export template to file
  graphfs examples export find-dependencies > my-query.sparql`,
}

// examplesListCmd represents the list subcommand
var examplesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available query templates",
	Long: `List all available query templates, optionally filtered by category.

Available categories:
  - dependencies: Dependency analysis queries
  - security: Security zone and boundary queries
  - analysis: Code quality and complexity queries
  - layers: Architectural layer queries
  - impact: Change impact analysis queries
  - documentation: Documentation coverage queries`,
	RunE: runExamplesList,
}

// examplesShowCmd represents the show subcommand
var examplesShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a query template",
	Long: `Show detailed information about a specific query template,
including its description, variables, query, and example usage.`,
	Args: cobra.ExactArgs(1),
	RunE: runExamplesShow,
}

// examplesRunCmd represents the run subcommand
var examplesRunCmd = &cobra.Command{
	Use:   "run <name> [flags]",
	Short: "Execute a query template",
	Long: `Execute a query template with the given variables.

Variables can be passed as flags (e.g., --module=api/handlers.go).
If a required variable is missing, the command will fail.

Use 'graphfs examples show <name>' to see available variables.`,
	Args: cobra.ExactArgs(1),
	RunE: runExamplesRun,
}

// examplesSaveCmd represents the save subcommand
var examplesSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a custom query template",
	Long: `Save a custom query template to the templates directory.

Custom templates are stored in .graphfs/templates/ and can be used
just like built-in templates.`,
	Args: cobra.ExactArgs(1),
	RunE: runExamplesSave,
}

// examplesExportCmd represents the export subcommand
var examplesExportCmd = &cobra.Command{
	Use:   "export <name> [flags]",
	Short: "Export a template to a .sparql file",
	Long: `Export a rendered query template to a .sparql file.

Variables can be passed as flags. The rendered query will be written
to the specified output file or stdout.

Use 'graphfs examples show <name>' to see available variables.`,
	Args: cobra.ExactArgs(1),
	RunE: runExamplesExport,
}

func init() {
	examplesCmd.AddCommand(examplesListCmd)
	examplesCmd.AddCommand(examplesShowCmd)
	examplesCmd.AddCommand(examplesRunCmd)
	examplesCmd.AddCommand(examplesSaveCmd)
	examplesCmd.AddCommand(examplesExportCmd)

	examplesListCmd.Flags().StringVar(&examplesCategory, "category", "", "Filter by category")
	examplesExportCmd.Flags().StringVarP(&examplesOutput, "output", "o", "", "Output file (default: stdout)")

	// Register dynamic template variable flags
	// This allows any --variable=value flags to be accepted
	examplesRunCmd.Flags().SetInterspersed(true)
	examplesExportCmd.Flags().SetInterspersed(true)

	// Allow unknown flags for template variables
	examplesRunCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
	examplesExportCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
}

// parseTemplateVariables extracts template variables from command flags and os.Args
func parseTemplateVariables(cmd *cobra.Command, tmpl *query.QueryTemplate) map[string]string {
	variables := make(map[string]string)

	// Parse variables from os.Args since cobra doesn't handle unknown flags well
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--") {
			// Parse --key=value format
			parts := strings.SplitN(arg[2:], "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				// Skip known flags
				if key != "output" && key != "category" && key != "config" &&
				   key != "verbose" && key != "quiet" && key != "no-color" {
					variables[key] = value
				}
			}
		}
	}

	// Also try registered flags
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip known flags
		if flag.Name != "output" && flag.Name != "category" {
			variables[flag.Name] = flag.Value.String()
		}
	})

	// Apply defaults for missing variables
	for _, v := range tmpl.Variables {
		if _, ok := variables[v.Name]; !ok && v.Default != "" {
			variables[v.Name] = v.Default
		}
	}

	return variables
}

func runExamplesList(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize template manager
	templatesDir := filepath.Join(currentDir, ".graphfs", "templates")
	tm := query.NewTemplateManager(templatesDir)

	// Get templates
	templates := tm.ListTemplates(examplesCategory)
	if len(templates) == 0 {
		if examplesCategory != "" {
			out.Info("No templates found in category: %s", examplesCategory)
		} else {
			out.Info("No templates found")
		}
		return nil
	}

	// Group templates by category
	categoryMap := make(map[string][]*query.QueryTemplate)
	for _, tmpl := range templates {
		categoryMap[tmpl.Category] = append(categoryMap[tmpl.Category], tmpl)
	}

	// Print header
	if noColor {
		fmt.Println("\nQuery Templates")
		fmt.Println()
	} else {
		color.New(color.FgCyan, color.Bold).Println("\n📚 Query Templates")
		fmt.Println()
	}

	// Sort categories
	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	// Print templates by category
	for _, cat := range categories {
		// Category header
		categoryTitle := strings.Title(cat)
		if noColor {
			fmt.Printf("%s:\n", categoryTitle)
		} else {
			color.New(color.FgYellow, color.Bold).Printf("%s:\n", categoryTitle)
		}

		// Sort templates in category
		tmplList := categoryMap[cat]
		sort.Slice(tmplList, func(i, j int) bool {
			return tmplList[i].Name < tmplList[j].Name
		})

		// Print templates
		for _, tmpl := range tmplList {
			if noColor {
				fmt.Printf("  • %-25s - %s\n", tmpl.Name, tmpl.Description)
			} else {
				color.New(color.FgWhite).Printf("  • ")
				color.New(color.FgGreen).Printf("%-25s", tmpl.Name)
				color.New(color.FgWhite).Printf(" - %s\n", tmpl.Description)
			}
		}
		fmt.Println()
	}

	// Print footer
	if noColor {
		fmt.Println("Use 'graphfs examples show <name>' for details")
		fmt.Println("Use 'graphfs examples run <name>' to execute")
	} else {
		color.New(color.FgHiBlack).Println("Use 'graphfs examples show <name>' for details")
		color.New(color.FgHiBlack).Println("Use 'graphfs examples run <name>' to execute")
	}

	return nil
}

func runExamplesShow(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize template manager
	templatesDir := filepath.Join(currentDir, ".graphfs", "templates")
	tm := query.NewTemplateManager(templatesDir)

	// Get template
	tmpl, err := tm.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Print template details
	if noColor {
		fmt.Printf("\n%s\n\n", tmpl.Name)
		fmt.Printf("%s\n\n", tmpl.Description)
		fmt.Printf("Category: %s\n\n", tmpl.Category)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("\n📄 %s\n\n", tmpl.Name)
		color.New(color.FgWhite).Printf("%s\n\n", tmpl.Description)
		color.New(color.FgYellow).Printf("Category: ")
		color.New(color.FgWhite).Printf("%s\n\n", tmpl.Category)
	}

	// Print variables
	if len(tmpl.Variables) > 0 {
		if noColor {
			fmt.Println("Variables:")
		} else {
			color.New(color.FgYellow).Println("Variables:")
		}

		for _, v := range tmpl.Variables {
			if v.Default != "" {
				if noColor {
					fmt.Printf("  • %s - %s (default: %s)\n", v.Name, v.Description, v.Default)
				} else {
					color.New(color.FgWhite).Printf("  • ")
					color.New(color.FgGreen).Printf("%s", v.Name)
					color.New(color.FgWhite).Printf(" - %s ", v.Description)
					color.New(color.FgHiBlack).Printf("(default: %s)\n", v.Default)
				}
			} else {
				if noColor {
					fmt.Printf("  • %s - %s\n", v.Name, v.Description)
				} else {
					color.New(color.FgWhite).Printf("  • ")
					color.New(color.FgGreen).Printf("%s", v.Name)
					color.New(color.FgWhite).Printf(" - %s\n", v.Description)
				}
			}
		}
		fmt.Println()
	}

	// Print query
	if noColor {
		fmt.Println("Query:")
		fmt.Println(tmpl.Query)
		fmt.Println()
	} else {
		color.New(color.FgYellow).Println("Query:")
		color.New(color.FgHiBlack).Println(tmpl.Query)
		fmt.Println()
	}

	// Print example
	if tmpl.Example != "" {
		if noColor {
			fmt.Printf("Example:\n%s\n", tmpl.Example)
		} else {
			color.New(color.FgYellow).Println("Example:")
			color.New(color.FgCyan).Println(tmpl.Example)
		}
	}

	return nil
}

func runExamplesRun(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)
	templateName := args[0]

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if GraphFS is initialized
	graphfsDir := filepath.Join(currentDir, ".graphfs")
	if _, err := os.Stat(graphfsDir); os.IsNotExist(err) {
		return fmt.Errorf("GraphFS not initialized. Run 'graphfs init' first")
	}

	// Initialize template manager
	templatesDir := filepath.Join(currentDir, ".graphfs", "templates")
	tm := query.NewTemplateManager(templatesDir)

	// Get template
	tmpl, err := tm.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Parse variables from flags
	variables := parseTemplateVariables(cmd, tmpl)

	// Render template
	queryString, err := tm.Render(tmpl, variables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	out.Debug("Rendered query:\n%s", queryString)

	// Build graph
	out.Debug("Building knowledge graph...")
	builder := graph.NewBuilder()
	graphObj, err := builder.Build(currentDir, graph.BuildOptions{
		ScanOptions: scanner.ScanOptions{
			UseDefaults: true,
			IgnoreFiles: []string{".gitignore", ".graphfsignore"},
			Concurrent:  true,
		},
		ReportProgress: verbose,
	})
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	out.Debug("Graph loaded: %d modules, %d triples",
		graphObj.Statistics.TotalModules,
		graphObj.Statistics.TotalTriples)

	// Execute query
	executor := query.NewExecutor(graphObj.Store)
	result, err := executor.ExecuteString(queryString)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Format and output results
	output, err := formatTable(result)
	if err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	fmt.Println(output)

	return nil
}

func runExamplesSave(cmd *cobra.Command, args []string) error {
	out := cli.NewOutputFormatter(quiet, verbose, noColor)
	templateName := args[0]

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize template manager
	templatesDir := filepath.Join(currentDir, ".graphfs", "templates")
	_ = query.NewTemplateManager(templatesDir)

	// For now, we'll prompt the user to create the template JSON manually
	// In a real implementation, we could read from stdin or flags
	out.Info("To save a custom template, create a JSON file in:")
	out.Info("  %s", templatesDir)
	out.Info("")
	out.Info("Example format:")
	out.Info(`  {
    "name": "%s",
    "description": "Your template description",
    "category": "custom",
    "query": "SELECT * WHERE { ... }",
    "variables": [
      {"name": "var1", "description": "Variable description"}
    ],
    "example": "graphfs examples run %s --var1=value"
  }`, templateName, templateName)

	return nil
}

func runExamplesExport(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize template manager
	templatesDir := filepath.Join(currentDir, ".graphfs", "templates")
	tm := query.NewTemplateManager(templatesDir)

	// Get template
	tmpl, err := tm.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Parse variables from flags
	variables := parseTemplateVariables(cmd, tmpl)

	// Render template
	queryString, err := tm.Render(tmpl, variables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Output to file or stdout
	if examplesOutput != "" {
		if err := os.WriteFile(examplesOutput, []byte(queryString), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		fmt.Println(queryString)
	}

	return nil
}
