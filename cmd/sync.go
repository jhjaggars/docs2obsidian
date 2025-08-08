package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"pkm-sync/internal/config"
	"pkm-sync/internal/sources/google"
	"pkm-sync/internal/targets/logseq"
	"pkm-sync/internal/targets/obsidian"
	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

var (
	sourceName string
	targetName string
	outputDir  string
	since      string
	dryRun     bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync data from source to target",
	Long: `Sync data from a source (google, slack, etc.) to a PKM target (obsidian, logseq, etc.)
    
Examples:
  pkm-sync sync --source google --target obsidian --output ./vault
  pkm-sync sync --source google --target logseq --output ./graph --since 7d
  pkm-sync sync --source google --target obsidian --dry-run`,
	RunE: runSyncCommand,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// These will be overridden by config defaults in runSyncCommand
	syncCmd.Flags().StringVar(&sourceName, "source", "", "Data source (google)")
	syncCmd.Flags().StringVar(&targetName, "target", "", "PKM target (obsidian, logseq)")
	syncCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	syncCmd.Flags().StringVar(&since, "since", "", "Sync items since (7d, 2006-01-02, today)")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without making changes")
}

func runSyncCommand(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// If no config exists, use defaults
		cfg = config.GetDefaultConfig()
	}

	// Determine which sources to sync
	var sourcesToSync []string
	if sourceName != "" {
		// CLI override: sync specific source
		sourcesToSync = []string{sourceName}
	} else {
		// Use enabled sources from config
		sourcesToSync = getEnabledSources(cfg)
	}

	// Apply config defaults, then CLI overrides
	finalTargetName := cfg.Sync.DefaultTarget
	if targetName != "" {
		finalTargetName = targetName
	}

	finalOutputDir := cfg.Sync.DefaultOutputDir
	if outputDir != "" {
		finalOutputDir = outputDir
	}

	finalSince := cfg.Sync.DefaultSince
	if since != "" {
		finalSince = since
	}

	// Parse since parameter
	sinceTime, err := parseSinceTime(finalSince)
	if err != nil {
		return fmt.Errorf("invalid since parameter: %w", err)
	}

	fmt.Printf("Syncing from sources [%s] to %s (output: %s, since: %s)\n",
		strings.Join(sourcesToSync, ", "), finalTargetName, finalOutputDir, finalSince)

	// Create target with config
	target, err := createTargetWithConfig(finalTargetName, cfg)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	// Collect items from all enabled sources
	var allItems []*models.Item
	for _, srcName := range sourcesToSync {
		// Get source-specific config
		sourceConfig, exists := cfg.Sources[srcName]
		if !exists {
			fmt.Printf("Warning: source '%s' not configured, skipping\n", srcName)

			continue
		}

		if !sourceConfig.Enabled {
			fmt.Printf("Source '%s' is disabled, skipping\n", srcName)

			continue
		}

		// Create source
		source, err := createSource(srcName)
		if err != nil {
			fmt.Printf("Warning: failed to create source '%s': %v, skipping\n", srcName, err)

			continue
		}

		// Use source-specific since time if configured
		sourceSince := finalSince
		if sourceConfig.Since != "" {
			sourceSince = sourceConfig.Since
		}

		sourceSinceTime, err := parseSinceTime(sourceSince)
		if err != nil {
			fmt.Printf("Warning: invalid since time for source '%s': %v, using default\n", srcName, err)
			sourceSinceTime = sinceTime
		}

		// Fetch items from this source
		fmt.Printf("Fetching from %s...\n", srcName)
		items, err := source.Fetch(sourceSinceTime, 1000) // TODO: make configurable
		if err != nil {
			fmt.Printf("Warning: failed to fetch from source '%s': %v, skipping\n", srcName, err)

			continue
		}

		// Add source tags if enabled
		if cfg.Sync.SourceTags {
			for _, item := range items {
				item.Tags = append(item.Tags, "source:"+srcName)
			}
		}

		fmt.Printf("Found %d items from %s\n", len(items), srcName)
		allItems = append(allItems, items...)
	}

	fmt.Printf("Total items collected: %d\n", len(allItems))

	if dryRun {
		fmt.Printf("DRY RUN: Would export %d items to %s\n", len(allItems), finalOutputDir)
		return nil
	}

	// Export all items to target
	if err := target.Export(allItems, finalOutputDir); err != nil {
		return fmt.Errorf("failed to export to target: %w", err)
	}

	fmt.Printf("Successfully exported %d items\n", len(allItems))
	return nil
}

func createSource(name string) (interfaces.Source, error) {
	switch name {
	case "google":
		source := google.NewGoogleSource()
		if err := source.Configure(nil); err != nil {
			return nil, err
		}
		return source, nil
	default:
		return nil, fmt.Errorf("unknown source '%s': supported sources are 'google' (others like slack, gmail, jira are planned for future releases)", name)
	}
}

func createTarget(name string) (interfaces.Target, error) {
	switch name {
	case "obsidian":
		target := obsidian.NewObsidianTarget()
		if err := target.Configure(nil); err != nil {
			return nil, err
		}
		return target, nil
	case "logseq":
		target := logseq.NewLogseqTarget()
		if err := target.Configure(nil); err != nil {
			return nil, err
		}
		return target, nil
	default:
		return nil, fmt.Errorf("unknown target '%s': supported targets are 'obsidian' and 'logseq'", name)
	}
}

func createTargetWithConfig(name string, cfg *models.Config) (interfaces.Target, error) {
	switch name {
	case "obsidian":
		target := obsidian.NewObsidianTarget()

		// Apply configuration
		configMap := make(map[string]interface{})
		if targetConfig, exists := cfg.Targets[name]; exists {
			configMap["template_dir"] = targetConfig.Obsidian.DefaultFolder
			configMap["daily_notes_format"] = targetConfig.Obsidian.DateFormat
		}

		if err := target.Configure(configMap); err != nil {
			return nil, err
		}
		return target, nil

	case "logseq":
		target := logseq.NewLogseqTarget()

		// Apply configuration
		configMap := make(map[string]interface{})
		if targetConfig, exists := cfg.Targets[name]; exists {
			configMap["default_page"] = targetConfig.Logseq.DefaultPage
		}

		if err := target.Configure(configMap); err != nil {
			return nil, err
		}
		return target, nil

	default:
		return nil, fmt.Errorf("unknown target '%s': supported targets are 'obsidian' and 'logseq'", name)
	}
}

func parseSinceTime(since string) (time.Time, error) {
	now := time.Now()

	switch since {
	case "today":
		return now.Truncate(24 * time.Hour), nil
	case "yesterday":
		return now.Add(-24 * time.Hour).Truncate(24 * time.Hour), nil
	}

	// Try relative duration (7d, 2h, etc.)
	// Go's ParseDuration doesn't handle "d" for days, so convert explicitly
	if strings.HasSuffix(since, "d") {
		daysStr := strings.TrimSuffix(since, "d")
		if daysInt, err := strconv.Atoi(daysStr); err == nil && daysInt >= 0 {
			daysDuration := time.Duration(daysInt) * 24 * time.Hour
			return now.Add(-daysDuration), nil
		}
	}

	if duration, err := time.ParseDuration(since); err == nil {
		return now.Add(-duration), nil
	}

	// Try absolute date
	if t, err := time.Parse("2006-01-02", since); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse since time '%s': supported formats are 'today', 'yesterday', relative durations (7d, 24h), or absolute dates (2006-01-02)", since)
}

// getEnabledSources returns list of sources that are enabled in the configuration
func getEnabledSources(cfg *models.Config) []string {
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
