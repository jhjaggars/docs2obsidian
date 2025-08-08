package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pkm-sync/internal/config"
)

var (
	credentialsPath string
	configDir       string
)

var rootCmd = &cobra.Command{
	Use:   "pkm-sync",
	Short: "Synchronize data between various sources and PKM systems",
	Long: `pkm-sync integrates data sources (Google Calendar, Slack, etc.) 
with Personal Knowledge Management systems (Obsidian, Logseq, etc.).`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&credentialsPath, "credentials", "c", "", "Path to credentials.json file")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Custom configuration directory")
	
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if credentialsPath != "" {
			config.SetCustomCredentialsPath(credentialsPath)
		}
		if configDir != "" {
			config.SetCustomConfigDir(configDir)
		}
	}
	
	// Initialize legacy command compatibility
	initLegacyCommands()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}