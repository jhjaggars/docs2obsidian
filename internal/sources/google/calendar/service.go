package calendar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
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

	return modelEvent
}

// ConvertToModelWithDrive converts a calendar event to a model with drive file attachments populated
func (s *Service) ConvertToModelWithDrive(event *calendar.Event, driveService DriveServiceInterface) *models.CalendarEvent {
	modelEvent := s.ConvertToModel(event)
	
	if driveService != nil && event.Description != "" {
		// Extract Google Docs file IDs from event description
		if fileIDs, err := driveService.GetAttachmentsFromEvent(event.Description); err == nil {
			for _, fileID := range fileIDs {
				if metadata, err := driveService.GetFileMetadata(fileID); err == nil {
					// Only include Google Docs
					if driveService.IsGoogleDoc(metadata.MimeType) {
						modelEvent.AttachedDocs = append(modelEvent.AttachedDocs, *metadata)
					}
				}
			}
		}
	}
	
	return modelEvent
}

// DriveServiceInterface defines the interface for drive service operations needed by calendar
type DriveServiceInterface interface {
	GetAttachmentsFromEvent(eventDescription string) ([]string, error)
	GetFileMetadata(fileID string) (*models.DriveFile, error)
	IsGoogleDoc(mimeType string) bool
}