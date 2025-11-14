/*
# Module: cmd/graphfs/config.go
Configuration handling for GraphFS CLI.

Manages loading and validation of configuration from files and environment.

## Linked Modules
- [root](./root.go) - Root command

## Tags
cli, config, viper

## Exports
Config, initConfig, loadConfig, saveDefaultConfig

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#config.go> a code:Module ;

	code:name "cmd/graphfs/config.go" ;
	code:description "Configuration handling for GraphFS CLI" ;
	code:language "go" ;
	code:layer "cli" ;
	code:linksTo <./root.go> ;
	code:exports <#Config>, <#initConfig>, <#loadConfig>, <#saveDefaultConfig> ;
	code:tags "cli", "config", "viper" .

<!-- End LinkedDoc RDF -->
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents GraphFS configuration
type Config struct {
	Version int         `yaml:"version"`
	Scan    ScanConfig  `yaml:"scan"`
	Query   QueryConfig `yaml:"query"`
}

// ScanConfig configures scanning behavior
type ScanConfig struct {
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
	MaxFileSize int64    `yaml:"max_file_size"`
}

// QueryConfig configures query behavior
type QueryConfig struct {
	DefaultLimit int           `yaml:"default_limit"`
	Timeout      time.Duration `yaml:"timeout"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Scan: ScanConfig{
			Include:     []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts"},
			Exclude:     []string{"**/node_modules/**", "**/vendor/**", "**/.git/**"},
			MaxFileSize: 1048576, // 1MB
		},
		Query: QueryConfig{
			DefaultLimit: 100,
			Timeout:      30 * time.Second,
		},
	}
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in .graphfs directory
		viper.AddConfigPath(".graphfs")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// loadConfig loads configuration from file or returns default
func loadConfig(configPath string) (*Config, error) {
	// Try to read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// saveDefaultConfig saves default configuration to file
func saveDefaultConfig(configPath string) error {
	config := DefaultConfig()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
