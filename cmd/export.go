package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"docs2obsidian/internal/auth"
	"docs2obsidian/internal/calendar"
	"docs2obsidian/internal/drive"
)

var (
	exportOutputDir string
	exportEventID   string
	exportStartDate string
	exportEndDate   string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Google Docs from calendar events to markdown",
	Long: `Export Google Docs attached to calendar events as markdown files.
	
You can export docs from:
- A specific event by ID
- All events in a date range
- Today's events (default)`,
	RunE: runExportCommand,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	
	exportCmd.Flags().StringVarP(&exportOutputDir, "output", "o", "./exported-docs", "Output directory for exported markdown files")
	exportCmd.Flags().StringVar(&exportEventID, "event-id", "", "Export docs from specific event ID")
	exportCmd.Flags().StringVar(&exportStartDate, "start", "", "Start date for range export (YYYY-MM-DD)")
	exportCmd.Flags().StringVar(&exportEndDate, "end", "", "End date for range export (YYYY-MM-DD)")
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	// Get authenticated client
	client, err := auth.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Create services
	calendarService, err := calendar.NewService(client)
	if err != nil {
		return fmt.Errorf("failed to create calendar service: %w", err)
	}

	driveService, err := drive.NewService(client)
	if err != nil {
		return fmt.Errorf("failed to create drive service: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(exportOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var totalExported int

	if exportEventID != "" {
		// Export from specific event
		count, err := exportFromEventID(calendarService, driveService, exportEventID)
		if err != nil {
			return err
		}
		totalExported = count
	} else {
		// Export from date range
		start, end, err := getExportDateRange()
		if err != nil {
			return err
		}
		
		count, err := exportFromDateRange(calendarService, driveService, start, end)
		if err != nil {
			return err
		}
		totalExported = count
	}

	fmt.Printf("\nExport complete! %d documents exported to %s\n", totalExported, exportOutputDir)
	return nil
}

func exportFromEventID(calendarService *calendar.Service, driveService *drive.Service, eventID string) (int, error) {
	fmt.Printf("Exporting docs from event ID: %s\n", eventID)
	
	// Note: We'd need to add a GetEvent method to calendar service
	// For now, we'll search in today's events
	events, err := calendarService.GetEventsInRange(
		time.Now().Add(-24*time.Hour), 
		time.Now().Add(24*time.Hour), 
		100,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get events: %w", err)
	}

	for _, event := range events {
		if event.Id == eventID {
			return exportFromSingleEvent(driveService, event.Summary, event.Description)
		}
	}

	return 0, fmt.Errorf("event with ID %s not found", eventID)
}

func exportFromDateRange(calendarService *calendar.Service, driveService *drive.Service, start, end time.Time) (int, error) {
	fmt.Printf("Exporting docs from events between %s and %s\n", 
		start.Format("2006-01-02"), end.Format("2006-01-02"))

	events, err := calendarService.GetEventsInRange(start, end, 100)
	if err != nil {
		return 0, fmt.Errorf("failed to get events: %w", err)
	}

	var totalExported int
	for _, event := range events {
		if event.Description == "" {
			continue
		}

		fmt.Printf("\nProcessing event: %s\n", event.Summary)
		count, err := exportFromSingleEvent(driveService, event.Summary, event.Description)
		if err != nil {
			fmt.Printf("Warning: Error processing event %s: %v\n", event.Summary, err)
			continue
		}
		totalExported += count
	}

	return totalExported, nil
}

func exportFromSingleEvent(driveService *drive.Service, eventSummary, eventDescription string) (int, error) {
	if eventDescription == "" {
		return 0, nil
	}

	// Create subdirectory for this event
	eventDir := filepath.Join(exportOutputDir, sanitizeEventName(eventSummary))
	
	exportedFiles, err := driveService.ExportAttachedDocsFromEvent(eventDescription, eventDir)
	if err != nil {
		return 0, err
	}

	return len(exportedFiles), nil
}

func getExportDateRange() (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if exportStartDate != "" {
		start, err = time.Parse("2006-01-02", exportStartDate)
		if err != nil {
			return start, end, fmt.Errorf("invalid start date format: %w", err)
		}
	} else {
		// Default to today
		start = time.Now().Truncate(24 * time.Hour)
	}

	if exportEndDate != "" {
		end, err = time.Parse("2006-01-02", exportEndDate)
		if err != nil {
			return start, end, fmt.Errorf("invalid end date format: %w", err)
		}
		// Set to end of day
		end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		// Default to end of today
		end = start.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	return start, end, nil
}

