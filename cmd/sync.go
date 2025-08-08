package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"pkm-sync/internal/sources/google"
	"pkm-sync/internal/sync"
	"pkm-sync/internal/targets/logseq"
	"pkm-sync/internal/targets/obsidian"
	"pkm-sync/pkg/interfaces"
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

	syncCmd.Flags().StringVar(&sourceName, "source", "google", "Data source (google)")
	syncCmd.Flags().StringVar(&targetName, "target", "obsidian", "PKM target (obsidian, logseq)")
	syncCmd.Flags().StringVarP(&outputDir, "output", "o", "./exported", "Output directory")
	syncCmd.Flags().StringVar(&since, "since", "7d", "Sync items since (7d, 2006-01-02, today)")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without making changes")
}

func runSyncCommand(cmd *cobra.Command, args []string) error {
	// Parse since parameter
	sinceTime, err := parseSinceTime(since)
	if err != nil {
		return fmt.Errorf("invalid since parameter: %w", err)
	}

	// Create source
	source, err := createSource(sourceName)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	// Create target
	target, err := createTarget(targetName)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	// Create syncer
	syncer := sync.NewSyncer()

	// Sync
	options := interfaces.SyncOptions{
		Since:     sinceTime,
		OutputDir: outputDir,
		DryRun:    dryRun,
	}

	return syncer.Sync(source, target, options)
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
		return nil, fmt.Errorf("unknown source: %s", name)
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
		return nil, fmt.Errorf("unknown target: %s", name)
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
	// Go's ParseDuration doesn't handle "d" for days, so convert first
	if strings.HasSuffix(since, "d") {
		daysStr := strings.TrimSuffix(since, "d")
		if days, err := time.ParseDuration(daysStr + "h"); err == nil {
			return now.Add(-days * 24), nil
		}
	}
	
	if duration, err := time.ParseDuration(since); err == nil {
		return now.Add(-duration), nil
	}

	// Try absolute date
	if t, err := time.Parse("2006-01-02", since); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse since time: %s", since)
}