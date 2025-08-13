package main

func initCommands() {
	// calendar.go
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.Flags().StringVar(&startDate, "start", "", "Start date (defaults to beginning of current week)")
	calendarCmd.Flags().StringVar(&endDate, "end", "", "End date (defaults to end of today)")
	calendarCmd.Flags().Int64Var(&maxResults, "limit", 100, "Maximum number of events to retrieve")
	calendarCmd.Flags().StringVar(&outputFormat, "format", "table", "Output format (table, json)")
	calendarCmd.Flags().BoolVar(&includeDetails, "include-details", false, "Include detailed meeting information (attendees, URLs, etc.)")
	calendarCmd.Flags().BoolVar(&exportDocs, "export-docs", false, "Export attached Google Docs to markdown")
	calendarCmd.Flags().StringVar(&exportDir, "export-dir", "./exported-docs", "Directory to export documents to")

	// config.go
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configEditCmd)
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing config file")
	configInitCmd.Flags().StringP("output", "o", "", "Output directory for default target")
	configInitCmd.Flags().String("target", "", "Default target (obsidian, logseq)")
	configInitCmd.Flags().String("source", "", "Default source (google)")

	// export.go
	rootCmd.AddCommand(driveCmd)
	driveCmd.Flags().StringVarP(&driveOutputDir, "output", "o", "./exported-docs", "Output directory for exported markdown files")
	driveCmd.Flags().StringVar(&driveEventID, "event-id", "", "Export docs from specific event ID")
	driveCmd.Flags().StringVar(&driveStartDate, "start", "", "Start date for range export (YYYY-MM-DD)")
	driveCmd.Flags().StringVar(&driveEndDate, "end", "", "End date for range export (YYYY-MM-DD)")

	// root.go
	rootCmd.PersistentFlags().StringVarP(&credentialsPath, "credentials", "c", "", "Path to credentials.json file")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Custom configuration directory")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging")

	// setup.go
	rootCmd.AddCommand(setupCmd)

	// sync.go
	rootCmd.AddCommand(gmailCmd)
	gmailCmd.Flags().StringVar(&gmailSourceName, "source", "", "Gmail source (gmail_work, gmail_personal, etc.)")
	gmailCmd.Flags().StringVar(&gmailTargetName, "target", "", "PKM target (obsidian, logseq)")
	gmailCmd.Flags().StringVarP(&gmailOutputDir, "output", "o", "", "Output directory")
	gmailCmd.Flags().StringVar(&gmailSince, "since", "", "Sync emails since (7d, 2006-01-02, today)")
	gmailCmd.Flags().BoolVar(&gmailDryRun, "dry-run", false, "Show what would be synced without making changes")
	gmailCmd.Flags().IntVar(&gmailLimit, "limit", 1000, "Maximum number of emails to fetch (default: 1000)")
	gmailCmd.Flags().StringVar(&gmailOutputFormat, "format", "summary", "Output format for dry-run (summary, json)")
}
