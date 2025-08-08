package sync

import (
	"fmt"

	"pkm-sync/pkg/interfaces"
)

type DefaultSyncer struct{}

func NewSyncer() *DefaultSyncer {
	return &DefaultSyncer{}
}

func (s *DefaultSyncer) Sync(source interfaces.Source, target interfaces.Target, options interfaces.SyncOptions) error {
	fmt.Printf("Syncing from %s to %s...\n", source.Name(), target.Name())

	// Fetch items from source
	items, err := source.Fetch(options.Since, 100) // TODO: make limit configurable
	if err != nil {
		return fmt.Errorf("failed to fetch from source: %w", err)
	}

	fmt.Printf("Found %d items\n", len(items))

	if options.DryRun {
		fmt.Printf("DRY RUN: Would export %d items to %s\n", len(items), options.OutputDir)

		return nil
	}

	// Export to target
	if err := target.Export(items, options.OutputDir); err != nil {
		return fmt.Errorf("failed to export to target: %w", err)
	}

	fmt.Printf("Successfully exported %d items\n", len(items))

	return nil
}
