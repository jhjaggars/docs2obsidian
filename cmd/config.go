package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pkm-sync/internal/config"
	"pkm-sync/pkg/models"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage pkm-sync configuration settings including default sources, targets, and sync options.

Examples:
  pkm-sync config init                    # Create default config file
  pkm-sync config show                    # Show current configuration  
  pkm-sync config path                    # Show config file location
  pkm-sync config edit                    # Open config in editor
  pkm-sync config validate               # Validate configuration`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default configuration file",
	Long:  "Creates a default configuration file with sensible defaults for pkm-sync.",
	RunE:  runConfigInitCommand,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration settings loaded from the config file.",
	RunE:  runConfigShowCommand,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  "Display the path to the configuration file that would be used or created.",
	RunE:  runConfigPathCommand,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  "Check if the configuration file is valid and can be loaded successfully.",
	RunE:  runConfigValidateCommand,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in editor",
	Long:  "Open the configuration file in your default editor (uses $EDITOR environment variable).",
	RunE:  runConfigEditCommand,
}

func init() {
	rootCmd.AddCommand(configCmd)
	
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configEditCmd)
	
	// Flags for config init
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing config file")
	configInitCmd.Flags().StringP("output", "o", "", "Output directory for default target")
	configInitCmd.Flags().String("target", "", "Default target (obsidian, logseq)")
	configInitCmd.Flags().String("source", "", "Default source (google)")
}

func runConfigInitCommand(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	output, _ := cmd.Flags().GetString("output")
	target, _ := cmd.Flags().GetString("target")
	source, _ := cmd.Flags().GetString("source")

	// Check if config already exists
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite", configPath)
	}

	// Create default config
	cfg := config.GetDefaultConfig()

	// Apply command line overrides
	if output != "" {
		cfg.Sync.DefaultOutputDir = output
	}
	
	if target != "" {
		cfg.Sync.DefaultTarget = target
	}
	
	if source != "" {
		// Add to enabled sources if not already present
		found := false
		for _, src := range cfg.Sync.EnabledSources {
			if src == source {
				found = true
				break
			}
		}
		if !found {
			cfg.Sync.EnabledSources = append(cfg.Sync.EnabledSources, source)
		}
		
		// Enable the source in the sources config
		if sourceConfig, exists := cfg.Sources[source]; exists {
			sourceConfig.Enabled = true
			cfg.Sources[source] = sourceConfig
		}
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Configuration file created at: %s\n", configPath)
	fmt.Println("\nYou can now:")
	fmt.Printf("  - Edit the config: pkm-sync config edit\n")
	fmt.Printf("  - View the config: pkm-sync config show\n")
	fmt.Printf("  - Use sync without flags: pkm-sync sync\n")
	
	return nil
}

func runConfigShowCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert to YAML for display
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

func runConfigPathCommand(cmd *cobra.Command, args []string) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	fmt.Println(configPath)
	
	// Show if file exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("(file exists)")
	} else {
		fmt.Println("(file does not exist - run 'pkm-sync config init' to create)")
	}
	
	return nil
}

func runConfigValidateCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	// Basic validation
	if cfg.Sync.DefaultTarget == "" {
		fmt.Println("❌ Default target not specified")
		return fmt.Errorf("invalid configuration")
	}

	// Check if default target exists
	if _, exists := cfg.Targets[cfg.Sync.DefaultTarget]; !exists {
		fmt.Printf("❌ Default target '%s' not configured\n", cfg.Sync.DefaultTarget)
		return fmt.Errorf("invalid configuration")
	}

	// Validate enabled sources
	enabledSources := getEnabledSourcesForValidation(cfg)
	if len(enabledSources) == 0 {
		fmt.Println("❌ No sources are enabled")
		return fmt.Errorf("invalid configuration")
	}

	// Check if all enabled sources exist and are configured
	for _, sourceName := range enabledSources {
		if sourceConfig, exists := cfg.Sources[sourceName]; !exists {
			fmt.Printf("❌ Enabled source '%s' not configured\n", sourceName)
			return fmt.Errorf("invalid configuration")
		} else if !sourceConfig.Enabled {
			fmt.Printf("❌ Source '%s' is listed as enabled but marked disabled\n", sourceName)
			return fmt.Errorf("invalid configuration")
		} else {
			// Source-specific validation
			if sourceConfig.Type == "" {
				fmt.Printf("❌ Source '%s' has no type specified\n", sourceName)
				return fmt.Errorf("invalid configuration")
			}
			if sourceConfig.Priority < 1 {
				fmt.Printf("❌ Source '%s' has invalid priority %d (must be >= 1)\n", sourceName, sourceConfig.Priority)
				return fmt.Errorf("invalid configuration")
			}
		}
	}

	// Validate output directory is writable
	if cfg.Sync.DefaultOutputDir != "" {
		if err := validateOutputDirectory(cfg.Sync.DefaultOutputDir); err != nil {
			fmt.Printf("❌ Default output directory '%s' is not writable: %v\n", cfg.Sync.DefaultOutputDir, err)
			return fmt.Errorf("invalid configuration")
		}
	}

	// Validate time duration settings
	if cfg.Sync.DefaultSince != "" {
		if _, err := parseSinceTime(cfg.Sync.DefaultSince); err != nil {
			fmt.Printf("❌ Invalid default_since time '%s': %v\n", cfg.Sync.DefaultSince, err)
			return fmt.Errorf("invalid configuration")
		}
	}

	fmt.Println("✅ Configuration is valid")
	fmt.Printf("   Enabled sources: [%s]\n", strings.Join(enabledSources, ", "))
	fmt.Printf("   Default target: %s\n", cfg.Sync.DefaultTarget)
	fmt.Printf("   Default output: %s\n", cfg.Sync.DefaultOutputDir)
	fmt.Printf("   Source tags: %t\n", cfg.Sync.SourceTags)
	fmt.Printf("   Merge sources: %t\n", cfg.Sync.MergeSources)
	
	return nil
}

func runConfigEditCommand(cmd *cobra.Command, args []string) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Check if config exists, create if not
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config file doesn't exist. Creating default config at %s\n", configPath)
		if err := config.CreateDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback editor
	}

	fmt.Printf("Opening config file in %s...\n", editor)
	fmt.Printf("Config file: %s\n", configPath)
	
	// Note: In a real implementation, you'd use exec.Command to launch the editor
	// For now, just show the path
	fmt.Println("Run the following command to edit:")
	fmt.Printf("  %s %s\n", editor, configPath)
	
	return nil
}

// Helper function to get config file path
func getConfigFilePath() (string, error) {
	if configDir != "" {
		return filepath.Join(configDir, config.ConfigFileName), nil
	}

	defaultConfigDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(defaultConfigDir, config.ConfigFileName), nil
}

// getEnabledSourcesForValidation returns enabled sources (same logic as sync command)
func getEnabledSourcesForValidation(cfg *models.Config) []string {
	var enabledSources []string
	
	// Use explicit enabled_sources list if provided
	if len(cfg.Sync.EnabledSources) > 0 {
		for _, srcName := range cfg.Sync.EnabledSources {
			if sourceConfig, exists := cfg.Sources[srcName]; exists && sourceConfig.Enabled {
				enabledSources = append(enabledSources, srcName)
			}
		}
		return enabledSources
	}
	
	// Fallback: find all enabled sources in config
	for srcName, sourceConfig := range cfg.Sources {
		if sourceConfig.Enabled {
			enabledSources = append(enabledSources, srcName)
		}
	}
	
	return enabledSources
}

// validateOutputDirectory checks if a directory path is writable
func validateOutputDirectory(dir string) error {
	// First check if directory exists or can be created
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	// Try to create a temporary file to test write permissions
	tempFile := filepath.Join(dir, ".pkm-sync-write-test")
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("no write permission: %w", err)
	}
	file.Close()
	
	// Clean up the test file
	if err := os.Remove(tempFile); err != nil {
		// Not critical if we can't remove it
	}
	
	return nil
}