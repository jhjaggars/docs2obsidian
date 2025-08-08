package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Legacy command compatibility note
func addLegacyNote(cmd *cobra.Command) {
	originalRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		fmt.Println("NOTE: This is a legacy command. Consider using:")
		fmt.Printf("  pkm-sync sync --source google --target obsidian --output %s\n", outputDir)
		fmt.Println()
		
		if originalRunE != nil {
			return originalRunE(c, args)
		}
		return nil
	}
}

// initLegacyCommands wraps existing commands with legacy compatibility notes
func initLegacyCommands() {
	// Add legacy note to existing calendar command
	if calendarCmd != nil {
		addLegacyNote(calendarCmd)
	}
	
	// Add legacy note to existing export command
	if exportCmd != nil {
		addLegacyNote(exportCmd)
	}
}