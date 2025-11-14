/*
# Module: cmd/graphfs/cmd_init.go
Init command implementation.

Initializes GraphFS in a directory by creating configuration and data directories.

## Linked Modules
- [root](./root.go) - Root command
- [config](./config.go) - Configuration handling

## Tags
cli, command, init

## Exports
initCmd

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#cmd_init.go> a code:Module ;

	code:name "cmd/graphfs/cmd_init.go" ;
	code:description "Init command implementation" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go>, <./config.go> ;
	code:exports <#initCmd> ;
	code:tags "cli", "command", "init" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize GraphFS in a directory",
	Long: `Initialize GraphFS in a directory by creating the .graphfs configuration
directory, config file, and .graphfsignore file.

Examples:
  graphfs init                  # Initialize in current directory
  graphfs init /path/to/project # Initialize in specific directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absPath)
	}

	// Create .graphfs directory
	graphfsDir := filepath.Join(absPath, ".graphfs")
	if err := os.MkdirAll(graphfsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .graphfs directory: %w", err)
	}

	// Create config file
	configPath := filepath.Join(graphfsDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := saveDefaultConfig(configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		fmt.Printf("✓ Created config file: %s\n", configPath)
	} else {
		fmt.Printf("⚠ Config file already exists: %s\n", configPath)
	}

	// Create .graphfsignore file
	ignorePath := filepath.Join(absPath, ".graphfsignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		if err := createDefaultIgnoreFile(ignorePath); err != nil {
			return fmt.Errorf("failed to create .graphfsignore file: %w", err)
		}
		fmt.Printf("✓ Created .graphfsignore file: %s\n", ignorePath)
	} else {
		fmt.Printf("⚠ .graphfsignore file already exists: %s\n", ignorePath)
	}

	// Create empty store directory
	storeDir := filepath.Join(graphfsDir, "store")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}
	fmt.Printf("✓ Created store directory: %s\n", storeDir)

	fmt.Printf("\n✓ GraphFS initialized successfully in %s\n", absPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize .graphfs/config.yaml")
	fmt.Println("  2. Add patterns to .graphfsignore if needed")
	fmt.Println("  3. Run 'graphfs scan' to build the knowledge graph")

	return nil
}

// createDefaultIgnoreFile creates a default .graphfsignore file
func createDefaultIgnoreFile(path string) error {
	content := `# GraphFS ignore patterns
# Add patterns for files and directories to ignore during scanning

# Dependencies
node_modules/
vendor/
*.pyc
__pycache__/

# Build outputs
dist/
build/
*.o
*.a
*.so

# IDE and editor files
.vscode/
.idea/
*.swp
*.swo
*~

# Version control
.git/
.svn/
.hg/

# OS files
.DS_Store
Thumbs.db

# Large files
*.log
*.zip
*.tar.gz
*.pdf
*.jpg
*.png
*.gif
`
	return os.WriteFile(path, []byte(content), 0644)
}
