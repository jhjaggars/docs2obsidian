package models

import "testing"

func TestAttendee_GetDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		attendee    Attendee
		expected    string
		description string
	}{
		{
			name: "both display name and email present",
			attendee: Attendee{
				Email:       "john.doe@example.com",
				DisplayName: "John Doe",
			},
			expected:    "John Doe",
			description: "Should return display name when both display name and email are present",
		},
		{
			name: "only email present",
			attendee: Attendee{
				Email:       "jane.smith@example.com",
				DisplayName: "",
			},
			expected:    "jane.smith@example.com",
			description: "Should return email when display name is empty",
		},
		{
			name: "display name is whitespace only",
			attendee: Attendee{
				Email:       "user@example.com",
				DisplayName: "   ",
			},
			expected:    "   ",
			description: "Should return whitespace display name as-is (no trimming logic)",
		},
		{
			name: "both fields empty",
			attendee: Attendee{
				Email:       "",
				DisplayName: "",
			},
			expected:    "",
			description: "Should return empty string when both fields are empty",
		},
		{
			name: "display name with special characters",
			attendee: Attendee{
				Email:       "test@example.com",
				DisplayName: "Test User (External)",
			},
			expected:    "Test User (External)",
			description: "Should handle display names with special characters",
		},
		{
			name: "display name with unicode characters",
			attendee: Attendee{
				Email:       "unicode@example.com",
				DisplayName: "José María",
			},
			expected:    "José María",
			description: "Should handle display names with unicode characters",
		},
		{
			name: "very long display name",
			attendee: Attendee{
				Email:       "long@example.com",
				DisplayName: "This is a very long display name that might be used in some systems",
			},
			expected:    "This is a very long display name that might be used in some systems",
			description: "Should handle very long display names",
		},
		{
			name: "email with plus addressing",
			attendee: Attendee{
				Email:       "user+tag@example.com",
				DisplayName: "",
			},
			expected:    "user+tag@example.com",
			description: "Should handle emails with plus addressing when display name is empty",
		},
		{
			name: "email with subdomain",
			attendee: Attendee{
				Email:       "user@subdomain.example.com",
				DisplayName: "",
			},
			expected:    "user@subdomain.example.com",
			description: "Should handle emails with subdomains when display name is empty",
		},
		{
			name: "display name same as email local part",
			attendee: Attendee{
				Email:       "john.doe@example.com",
				DisplayName: "john.doe",
			},
			expected:    "john.doe",
			description: "Should prefer display name even when it matches email local part",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attendee.GetDisplayName()
			if result != tt.expected {
				t.Errorf("GetDisplayName() = %q, expected %q. %s", result, tt.expected, tt.description)
			}
		})
	}
}

func TestAttendee_GetDisplayName_EdgeCases(t *testing.T) {
	// Test zero value attendee
	zeroAttendee := Attendee{}
	result := zeroAttendee.GetDisplayName()
	if result != "" {
		t.Errorf("GetDisplayName() on zero value = %q, expected empty string", result)
	}
}

func TestAttendee_StructFields(t *testing.T) {
	// Test that the struct fields are correctly accessible
	attendee := Attendee{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	if attendee.Email != "test@example.com" {
		t.Errorf("Email field = %q, expected %q", attendee.Email, "test@example.com")
	}

	if attendee.DisplayName != "Test User" {
		t.Errorf("DisplayName field = %q, expected %q", attendee.DisplayName, "Test User")
	}
}
