package calendar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"pkm-sync/pkg/models"
)

type Service struct {
	calendarService *calendar.Service
}

func NewService(client *http.Client) (*Service, error) {
	ctx := context.Background()

	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return &Service{
		calendarService: calendarService,
	}, nil
}

func (s *Service) GetUpcomingEvents(maxResults int64) ([]*calendar.Event, error) {
	t := time.Now().Format(time.RFC3339)

	events, err := s.calendarService.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(maxResults).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	return events.Items, nil
}

func (s *Service) GetEventsInRange(start, end time.Time, maxResults int64) ([]*calendar.Event, error) {
	startTime := start.Format(time.RFC3339)
	endTime := end.Format(time.RFC3339)

	events, err := s.calendarService.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startTime).
		TimeMax(endTime).
		MaxResults(maxResults).
		OrderBy("startTime").
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events in range: %w", err)
	}

	return events.Items, nil
}

func (s *Service) ConvertToModel(event *calendar.Event) *models.CalendarEvent {
	modelEvent := &models.CalendarEvent{
		ID:          event.Id,
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
	}

	if event.Start.DateTime != "" {
		if startTime, err := time.Parse(time.RFC3339, event.Start.DateTime); err == nil {
			modelEvent.Start = startTime
		}
	}

	if event.End.DateTime != "" {
		if endTime, err := time.Parse(time.RFC3339, event.End.DateTime); err == nil {
			modelEvent.End = endTime
		}
	}

	for _, attendee := range event.Attendees {
		if attendee.Email != "" {
			modelEvent.Attendees = append(modelEvent.Attendees, attendee.Email)
		}
	}

	if event.ConferenceData != nil && len(event.ConferenceData.EntryPoints) > 0 {
		for _, entryPoint := range event.ConferenceData.EntryPoints {
			if entryPoint.EntryPointType == "video" && entryPoint.Uri != "" {
				modelEvent.MeetingURL = entryPoint.Uri

				break
			}
		}
	}

	// Process native Calendar API attachments
	for _, attachment := range event.Attachments {
		calAttachment := models.CalendarAttachment{
			FileURL:  attachment.FileUrl,
			FileID:   attachment.FileId,
			Title:    attachment.Title,
			MimeType: attachment.MimeType,
			IconLink: attachment.IconLink,
		}
		modelEvent.Attachments = append(modelEvent.Attachments, calAttachment)
	}

	return modelEvent
}

// ConvertToModelWithDrive converts a calendar event to a model with drive file attachments populated.
func (s *Service) ConvertToModelWithDrive(event *calendar.Event) *models.CalendarEvent {
	// Now that we use native Calendar API attachments, just use the base conversion
	return s.ConvertToModel(event)
}
