package google

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"pkm-sync/internal/sources/google/auth"
	"pkm-sync/internal/sources/google/calendar"
	"pkm-sync/internal/sources/google/drive"
	"pkm-sync/internal/sources/google/gmail"
	"pkm-sync/pkg/models"
)

type GoogleSource struct {
	calendarService *calendar.Service
	driveService    *drive.Service
	gmailService    *gmail.Service
	httpClient      *http.Client
	config          models.SourceConfig
	sourceID        string
}

func NewGoogleSource() *GoogleSource {
	return &GoogleSource{}
}

func NewGoogleSourceWithConfig(sourceID string, config models.SourceConfig) *GoogleSource {
	return &GoogleSource{
		sourceID: sourceID,
		config:   config,
	}
}

func (g *GoogleSource) Name() string {
	return "google"
}

func (g *GoogleSource) Configure(config map[string]interface{}) error {
	// Use existing auth logic
	client, err := auth.GetClient()
	if err != nil {
		return err
	}
	g.httpClient = client

	// Initialize services using existing code
	g.calendarService, err = calendar.NewService(client)
	if err != nil {
		return err
	}

	// Configure attendee allow list if provided
	if allowListInterface, exists := config["attendee_allow_list"]; exists {
		if allowList, ok := allowListInterface.([]interface{}); ok {
			var stringAllowList []string
			for _, item := range allowList {
				if emailStr, ok := item.(string); ok {
					stringAllowList = append(stringAllowList, emailStr)
				}
			}
			g.calendarService.SetAttendeeAllowList(stringAllowList)
		}
	}

	// Configure attendee count filtering options
	if requireMultiple, exists := config["require_multiple_attendees"]; exists {
		if requireBool, ok := requireMultiple.(bool); ok {
			g.calendarService.SetRequireMultipleAttendees(requireBool)
		}
	}

	if includeSelfOnly, exists := config["include_self_only_events"]; exists {
		if includeBool, ok := includeSelfOnly.(bool); ok {
			g.calendarService.SetIncludeSelfOnlyEvents(includeBool)
		}
	}

	g.driveService, err = drive.NewService(client)
	if err != nil {
		return err
	}

	// Initialize Gmail service if this is a Gmail source type
	if g.config.Type == "gmail" {
		g.gmailService, err = gmail.NewService(client, g.config.Gmail, g.sourceID)
		if err != nil {
			return fmt.Errorf("failed to initialize Gmail service: %w", err)
		}
	}

	return nil
}

func (g *GoogleSource) Fetch(since time.Time, limit int) ([]*models.Item, error) {
	// Handle Gmail sources
	if g.config.Type == "gmail" {
		if g.gmailService == nil {
			return nil, fmt.Errorf("Gmail service not initialized")
		}

		// Fetch messages from Gmail
		messages, err := g.gmailService.GetMessages(since, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Gmail messages: %w", err)
		}

		// Convert to universal Item format
		var items []*models.Item
		for _, message := range messages {
			item, err := gmail.FromGmailMessageWithService(message, g.config.Gmail, g.gmailService)
			if err != nil {
				slog.Warn("Failed to convert message", "message_id", message.Id, "error", err)
				continue
			}
			items = append(items, item)
		}

		// Apply thread processing if enabled
		if g.config.Gmail.IncludeThreads {
			threadProcessor := gmail.NewThreadProcessor(g.config.Gmail)
			processedItems, err := threadProcessor.ProcessThreads(items)
			if err != nil {
				return nil, fmt.Errorf("failed to process threads: %w", err)
			}
			return processedItems, nil
		}

		return items, nil
	}

	// Default: Handle Google Calendar sources
	// Use existing calendar.GetEventsInRange
	end := time.Now().Add(24 * time.Hour) // Default to next 24 hours
	events, err := g.calendarService.GetEventsInRange(since, end, int64(limit))
	if err != nil {
		return nil, err
	}

	// Convert events to universal Item format
	var items []*models.Item
	for _, event := range events {
		// Convert to our model first
		calEvent := g.calendarService.ConvertToModelWithDrive(event)

		// Then convert to universal Item format
		item := models.FromCalendarEvent(calEvent)
		items = append(items, item)
	}

	return items, nil
}

func (g *GoogleSource) SupportsRealtime() bool {
	return false // Future: implement webhooks
}
