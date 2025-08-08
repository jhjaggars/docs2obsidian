package models

import "time"

// Item represents a universal data item from any source
type Item struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	SourceType  string                 `json:"source_type"`  // "google_calendar", "slack", etc.
	ItemType    string                 `json:"item_type"`    // "event", "message", "document", etc.
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Attachments []Attachment           `json:"attachments"`
	Metadata    map[string]interface{} `json:"metadata"`
	Links       []Link                 `json:"links"` // URLs, references
}

type Attachment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	URL       string `json:"url"`
	LocalPath string `json:"local_path,omitempty"`
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Type  string `json:"type"` // "meeting_url", "document", "external"
}

// Migrate from existing CalendarEvent model
func FromCalendarEvent(event *CalendarEvent) *Item {
	item := &Item{
		ID:         event.ID,
		Title:      event.Summary,
		Content:    event.Description,
		SourceType: "google_calendar",
		ItemType:   "event",
		CreatedAt:  event.Start, // Using start time as creation time for events
		UpdatedAt:  event.Start, // Using start time since we don't have modified time in CalendarEvent
		Metadata: map[string]interface{}{
			"start_time": event.Start,
			"end_time":   event.End,
			"location":   event.Location,
			"attendees":  event.Attendees,
		},
	}

	// Convert Drive attachments
	for _, doc := range event.AttachedDocs {
		item.Attachments = append(item.Attachments, Attachment{
			ID:       doc.ID,
			Name:     doc.Name,
			MimeType: doc.MimeType,
			URL:      doc.WebViewLink,
		})
	}

	// Add meeting URL as a link
	if event.MeetingURL != "" {
		item.Links = append(item.Links, Link{
			URL:   event.MeetingURL,
			Title: "Meeting URL",
			Type:  "meeting_url",
		})
	}

	return item
}