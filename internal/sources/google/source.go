package google

import (
	"net/http"
	"time"

	"pkm-sync/internal/sources/google/auth"
	"pkm-sync/internal/sources/google/calendar"
	"pkm-sync/internal/sources/google/drive"
	"pkm-sync/pkg/models"
)

type GoogleSource struct {
	calendarService *calendar.Service
	driveService    *drive.Service
	httpClient      *http.Client
}

func NewGoogleSource() *GoogleSource {
	return &GoogleSource{}
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

	return nil
}

func (g *GoogleSource) Fetch(since time.Time, limit int) ([]*models.Item, error) {
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