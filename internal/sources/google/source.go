package google

import (
	"fmt"
	"net/http"
	"time"

	"pkm-sync/internal/sources/google/auth"
	"pkm-sync/internal/sources/google/calendar"
	"pkm-sync/internal/sources/google/drive"
	"pkm-sync/internal/sources/google/gmail"
	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

const (
	SourceTypeGoogle   = "google"
	SourceTypeGmail    = "gmail"
	SourceTypeCalendar = "google_calendar"
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
	if g.sourceID != "" {
		return g.sourceID
	}

	if g.config.Type == SourceTypeGmail {
		return SourceTypeGmail
	}

	return SourceTypeCalendar
}

func (g *GoogleSource) Configure(config map[string]interface{}, client *http.Client) error {
	var err error
	if client == nil {
		// Use existing auth logic if no client is provided
		client, err = auth.GetClient()
		if err != nil {
			return fmt.Errorf("failed to get authenticated client: %w", err)
		}
	}

	g.httpClient = client

	// Initialize services based on source type
	if g.config.Type == SourceTypeGmail {
		return g.initializeGmailService(client)
	}

	// Default to calendar and drive services
	return g.initializeCalendarAndDriveServices(client, config)
}

// initializeGmailService initializes the Gmail service for Gmail sources.
func (g *GoogleSource) initializeGmailService(client *http.Client) error {
	var err error

	g.gmailService, err = gmail.NewService(client, g.config.Gmail, g.sourceID)
	if err != nil {
		return fmt.Errorf("failed to initialize Gmail service: %w", err)
	}

	return nil
}

// initializeCalendarAndDriveServices initializes calendar and drive services for non-Gmail sources.
func (g *GoogleSource) initializeCalendarAndDriveServices(client *http.Client, config map[string]interface{}) error {
	var err error

	// Initialize calendar service
	g.calendarService, err = calendar.NewService(client)
	if err != nil {
		return fmt.Errorf("failed to initialize Calendar service: %w", err)
	}

	// Configure calendar service options
	g.configureCalendarService(config)

	// Initialize drive service
	g.driveService, err = drive.NewService(client)
	if err != nil {
		return fmt.Errorf("failed to initialize Drive service: %w", err)
	}

	return nil
}

// configureCalendarService applies configuration settings to the calendar service.
func (g *GoogleSource) configureCalendarService(config map[string]interface{}) {
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
}

func (g *GoogleSource) Fetch(since time.Time, limit int) ([]models.ItemInterface, error) {
	if g.config.Type == SourceTypeGmail {
		return g.fetchGmail(since, limit)
	}

	// Default: Handle Google Calendar sources
	return g.fetchCalendar(since, limit)
}

func (g *GoogleSource) fetchGmail(since time.Time, limit int) ([]models.ItemInterface, error) {
	if g.gmailService == nil {
		return nil, fmt.Errorf("gmail service not initialized")
	}

	messages, err := g.gmailService.GetMessages(since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Gmail messages: %w", err)
	}

	items := make([]models.ItemInterface, 0, len(messages))

	for _, message := range messages {
		legacyItem, err := gmail.FromGmailMessageWithService(message, g.config.Gmail, g.gmailService)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Gmail message to item: %w", err)
		}

		// Convert legacy item to ItemInterface
		item := models.AsItemInterface(legacyItem)
		items = append(items, item)
	}

	if g.config.Gmail.IncludeThreads {
		threadProcessor := gmail.NewThreadProcessor(g.config.Gmail)

		// The thread processor still works with []*models.Item, so we need to convert
		legacyItems := make([]*models.Item, len(items))
		for i, item := range items {
			legacyItems[i] = models.AsItemStruct(item)
		}

		processedLegacyItems, err := threadProcessor.ProcessThreads(legacyItems)
		if err != nil {
			return nil, fmt.Errorf("failed to process threads: %w", err)
		}

		// Convert back to ItemInterface
		processedItems := make([]models.ItemInterface, len(processedLegacyItems))
		for i, legacyItem := range processedLegacyItems {
			processedItems[i] = models.AsItemInterface(legacyItem)
		}

		return processedItems, nil
	}

	return items, nil
}

func (g *GoogleSource) fetchCalendar(since time.Time, limit int) ([]models.ItemInterface, error) {
	if g.calendarService == nil {
		return nil, fmt.Errorf("calendar service not initialized")
	}

	events, err := g.calendarService.GetEventsInRange(since, time.Now().AddDate(0, 1, 0), int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar events: %w", err)
	}

	items := make([]models.ItemInterface, 0, len(events))

	for _, event := range events {
		// Convert API event to model, then to legacy item, then to interface
		calEvent := g.calendarService.ConvertToModelWithDrive(event)
		legacyItem := models.FromCalendarEvent(calEvent)
		item := models.AsItemInterface(legacyItem)
		items = append(items, item)
	}

	return items, nil
}

func (g *GoogleSource) SupportsRealtime() bool {
	return false // Future: implement webhooks
}

// Ensure GoogleSource implements Source interface.
var _ interfaces.Source = (*GoogleSource)(nil)
