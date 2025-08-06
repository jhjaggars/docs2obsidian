package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"docs2obsidian/internal/config"
)

var (
	credentialsPath string
	configDir       string
)

var rootCmd = &cobra.Command{
	Use:   "docs2obsidian",
	Short: "A tool to integrate Google Calendar and Drive with Obsidian",
	Long: `docs2obsidian integrates Google Calendar and Google Drive with your Obsidian notes system.
It uses OAuth 2.0 for authentication and can fetch calendar events and shared documents.`,
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}