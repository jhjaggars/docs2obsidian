package models

import "time"

type CalendarEvent struct {
	ID          string
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
	Location    string
	Attendees   []string
	MeetingURL  string
	Attachments []CalendarAttachment
}

type CalendarAttachment struct {
	FileURL  string
	FileID   string
	Title    string
	MimeType string
	IconLink string
}

type DriveFile struct {
	ID           string
	Name         string
	MimeType     string
	WebViewLink  string
	ModifiedTime time.Time
	Owners       []string
	Shared       bool
}

