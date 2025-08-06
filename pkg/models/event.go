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
	AttachedDocs []DriveFile
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