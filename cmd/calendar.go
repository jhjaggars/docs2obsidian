package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"

	"docs2obsidian/internal/auth"
	internalcalendar "docs2obsidian/internal/calendar"
	"docs2obsidian/pkg/models"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "List calendar events within a date range",
	Long: `Fetches and displays calendar events from your Google Calendar within a specified date range.

By default, shows events from the beginning of the current week to the end of today.
Supports flexible date formats including ISO 8601 dates and relative dates like 'today', 'tomorrow', 'yesterday'.

Examples:
  docs2obsidian calendar                           # Current week to today
  docs2obsidian calendar --start today            # Today only  
  docs2obsidian calendar --start 2025-01-01 --end 2025-01-31
  docs2obsidian calendar --include-details        # Show meeting URLs, attendees, etc.
  docs2obsidian calendar --format json            # Output as JSON`,
	RunE: runCalendarCommand,
}
// Calendar command flags
var (
	startDate      string
	endDate        string
	maxResults     int64
	outputFormat   string
	includeDetails bool
)

func init() {
	rootCmd.AddCommand(calendarCmd)
	
	// Add flags for date range and options
	calendarCmd.Flags().StringVar(&startDate, "start", "", "Start date (defaults to beginning of current week)")
	calendarCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (defaults to beginning of current week)")
	calendarCmd.Flags().StringVar(&endDate, "end", "", "End date (defaults to end of today)")
	calendarCmd.Flags().StringVar(&endDate, "end-date", "", "End date (defaults to end of today)")
	calendarCmd.Flags().Int64Var(&maxResults, "max-results", 100, "Maximum number of events to retrieve")
	calendarCmd.Flags().Int64Var(&maxResults, "limit", 100, "Maximum number of events to retrieve")
	calendarCmd.Flags().StringVar(&outputFormat, "format", "table", "Output format (table, json)")
	calendarCmd.Flags().BoolVar(&includeDetails, "include-details", false, "Include detailed meeting information (attendees, URLs, etc.)")
}

// getBeginningOfWeek returns the start of the current week (Monday at 00:00:00)
func getBeginningOfWeek() time.Time {
	now := time.Now()
	// Get the weekday (0 = Sunday, 1 = Monday, etc.)
	weekday := int(now.Weekday())
	// Convert Sunday (0) to 7 for easier calculation
	if weekday == 0 {
		weekday = 7
	}
	// Calculate days to subtract to get to Monday
	daysToSubtract := weekday - 1
	
	// Get Monday of this week at 00:00:00
	monday := now.AddDate(0, 0, -daysToSubtract)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// getEndOfDay returns the end of the specified day (23:59:59)
func getEndOfDay(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999999999, day.Location())
}

// getEndOfToday returns the end of today (23:59:59)
func getEndOfToday() time.Time {
	return getEndOfDay(time.Now())
}

// parseDate parses a date string with support for relative dates
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	
	// Handle relative dates
	now := time.Now()
	switch dateStr {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location()), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()), nil
	}
	
	// Try parsing ISO 8601 date formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s. Supported formats: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, 'today', 'tomorrow', 'yesterday'", dateStr)
}

// getDateRange returns the start and end dates for the calendar query
func getDateRange() (time.Time, time.Time, error) {
	var start, end time.Time
	var err error
	
	// Parse start date or use default (beginning of week)
	if startDate != "" {
		start, err = parseDate(startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
		}
	} else {
		start = getBeginningOfWeek()
	}
	
	// Parse end date or use default (end of today)
	if endDate != "" {
		end, err = parseDate(endDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
		}
		// If only date was provided (no time), set to end of day
		if end.Hour() == 0 && end.Minute() == 0 && end.Second() == 0 {
			end = getEndOfDay(end)
		}
	} else {
		end = getEndOfToday()
	}
	
	// Validate date range
	if start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date (%s) cannot be after end date (%s)", start.Format("2006-01-02"), end.Format("2006-01-02"))
	}
	
	return start, end, nil
}

func runCalendarCommand(cmd *cobra.Command, args []string) error {
	client, err := auth.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	calendarService, err := internalcalendar.NewService(client)
	if err != nil {
		return fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Get date range using smart defaults
	start, end, err := getDateRange()
	if err != nil {
		return err
	}

	// Fetch events in the specified range
	events, err := calendarService.GetEventsInRange(start, end, maxResults)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Format and display results
	return formatAndDisplayEvents(events, start, end, calendarService)
}

// formatAndDisplayEvents formats and displays the calendar events
// formatAndDisplayEvents formats and displays the calendar events
// formatAndDisplayEvents formats and displays the calendar events
func formatAndDisplayEvents(events []*calendar.Event, start, end time.Time, calendarService *internalcalendar.Service) error {
	if len(events) == 0 {
		fmt.Printf("No events found between %s and %s\n", 
			start.Format("2006-01-02"), end.Format("2006-01-02"))
		return nil
	}

	switch outputFormat {
	case "json":
		return displayEventsAsJSON(events, calendarService)
	case "table":
		fallthrough
	default:
		return displayEventsAsTable(events, start, end, calendarService)
	}
}

// displayEventsAsTable displays events in a human-readable table format
// displayEventsAsTable displays events in a human-readable table format
// displayEventsAsTable displays events in a human-readable table format
func displayEventsAsTable(events []*calendar.Event, start, end time.Time, calendarService *internalcalendar.Service) error {
	fmt.Printf("Events from %s to %s (%d events):\n\n", 
		start.Format("2006-01-02"), end.Format("2006-01-02"), len(events))

	for _, event := range events {
		// Convert to model for rich data access
		modelEvent := calendarService.ConvertToModel(event)
		
		// Display event summary and time
		eventTime := ""
		if event.Start.DateTime != "" {
			if startTime, err := time.Parse(time.RFC3339, event.Start.DateTime); err == nil {
				eventTime = startTime.Format("Mon Jan 2 15:04")
			}
		} else if event.Start.Date != "" {
			eventTime = event.Start.Date + " (All day)"
		}
		
		fmt.Printf("• %s\n", event.Summary)
		if eventTime != "" {
			fmt.Printf("  %s\n", eventTime)
		}
		
		// Show additional details if requested
		if includeDetails {
			if event.Location != "" {
				fmt.Printf("  📍 %s\n", event.Location)
			}
			
			if modelEvent.MeetingURL != "" {
				fmt.Printf("  🔗 %s\n", modelEvent.MeetingURL)
			}
			
			if len(modelEvent.Attendees) > 0 && len(modelEvent.Attendees) <= 5 {
				fmt.Printf("  👥 %s\n", strings.Join(modelEvent.Attendees, ", "))
			} else if len(modelEvent.Attendees) > 5 {
				fmt.Printf("  👥 %s and %d others\n", 
					strings.Join(modelEvent.Attendees[:3], ", "), 
					len(modelEvent.Attendees)-3)
			}
			
			if event.Description != "" && len(event.Description) > 0 {
				description := event.Description
				if len(description) > 100 {
					description = description[:97] + "..."
				}
				fmt.Printf("  📝 %s\n", description)
			}
		}
		fmt.Println()
	}
	
	return nil
}

// displayEventsAsJSON displays events in JSON format
// displayEventsAsJSON displays events in JSON format
// displayEventsAsJSON displays events in JSON format
func displayEventsAsJSON(events []*calendar.Event, calendarService *internalcalendar.Service) error {
	// Convert to model events for rich data
	modelEvents := make([]*models.CalendarEvent, len(events))
	for i, event := range events {
		modelEvents[i] = calendarService.ConvertToModel(event)
	}
	
	jsonData, err := json.MarshalIndent(modelEvents, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events to JSON: %w", err)
	}
	
	fmt.Println(string(jsonData))
	return nil
}
