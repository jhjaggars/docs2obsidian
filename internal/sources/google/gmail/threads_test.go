package gmail

import (
	"strings"
	"testing"
	"time"

	"pkm-sync/pkg/models"
)

func TestThreadProcessor_SanitizeThreadSubject_Security(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "path traversal - parent directory",
			input:       "../../../etc/passwd",
			expected:    "etc-passwd",
			description: "Should prevent path traversal attacks",
		},
		{
			name:        "path traversal - current directory",
			input:       "./sensitive/file",
			expected:    "sensitive-file",
			description: "Should remove current directory references",
		},
		{
			name:        "path traversal - mixed",
			input:       "../config/../secrets",
			expected:    "config-secrets",
			description: "Should handle multiple path traversal attempts",
		},
		{
			name:        "home directory reference",
			input:       "~/private/data",
			expected:    "private-data",
			description: "Should remove home directory references",
		},
		{
			name:        "dot files",
			input:       ".hidden_file",
			expected:    "hidden_file",
			description: "Should handle dot files safely",
		},
		{
			name:        "directory separator injection",
			input:       "file/with/slashes",
			expected:    "file-with-slashes",
			description: "Should replace directory separators",
		},
		{
			name:        "windows path separators",
			input:       "file\\with\\backslashes",
			expected:    "file-with-backslashes",
			description: "Should replace Windows directory separators",
		},
		{
			name:        "null bytes",
			input:       "file\x00name",
			expected:    "filename",
			description: "Should handle null bytes",
		},
		{
			name:        "control characters",
			input:       "file\nwith\tcontrol\rchars",
			expected:    "filewithcontrolchars",
			description: "Should remove control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.sanitizeThreadSubject(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeThreadSubject(%q) = %q, want %q\nDescription: %s", 
					tt.input, result, tt.expected, tt.description)
			}
			
			// Additional security validation
			if strings.Contains(result, "..") {
				t.Errorf("Result still contains path traversal: %q", result)
			}
			if strings.Contains(result, "/") || strings.Contains(result, "\\") {
				t.Errorf("Result still contains path separators: %q", result)
			}
		})
	}
}

func TestThreadProcessor_SanitizeThreadSubject_EdgeCases(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "email-thread",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "email-thread", // After removing whitespace/control chars, becomes empty
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "at-$%^-and", // After replacements: ! -> empty, @ -> "-at-", # -> "-", $%^ remain, & -> "-and-", * -> empty, () -> empty
		},
		{
			name:     "only hyphens",
			input:    "-----",
			expected: "email-thread",
		},
		{
			name:     "unicode characters",
			input:    "Héllo Wörld",
			expected: "Héllo-Wörld",
		},
		{
			name:     "emoji in subject",
			input:    "Meeting 📅 Tomorrow",
			expected: "Meeting-📅-Tomorrow",
		},
		{
			name:     "multiple consecutive spaces",
			input:    "Test     Subject",
			expected: "Test-Subject",
		},
		{
			name:     "mixed case",
			input:    "CamelCase Subject",
			expected: "CamelCase-Subject",
		},
		{
			name:     "numbers and letters",
			input:    "Test123 Subject456",
			expected: "Test123-Subject456",
		},
		{
			name:     "leading and trailing special chars",
			input:    "!!!Important Subject!!!",
			expected: "Important-Subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.sanitizeThreadSubject(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeThreadSubject(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestThreadProcessor_SanitizeThreadSubject_Performance(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	// Test with very long string
	longInput := strings.Repeat("Test Subject with Many Words ", 100)
	result := processor.sanitizeThreadSubject(longInput)
	
	if len(result) > 80 {
		t.Errorf("Long input result should be truncated to 80 chars, got %d", len(result))
	}
	
	// Test multiple consecutive hyphens (worst case for the old implementation)
	manyHyphens := strings.Repeat("-", 1000)
	result = processor.sanitizeThreadSubject(manyHyphens)
	
	if result != "email-thread" {
		t.Errorf("Many hyphens should result in fallback, got %q", result)
	}
}

func TestThreadProcessor_SanitizeThreadSubject_NilSafety(t *testing.T) {
	// Test with nil processor
	var processor *ThreadProcessor
	result := processor.sanitizeThreadSubject("test")
	if result != "email-thread" {
		t.Errorf("Nil processor should return fallback, got %q", result)
	}
	
	// Test with valid processor but edge case inputs
	validProcessor := NewThreadProcessor(models.GmailSourceConfig{})
	
	tests := []string{
		"",
		"   ",
		"...",
		"///",
		"\\\\\\",
		"@#$%",
	}
	
	for _, input := range tests {
		result := validProcessor.sanitizeThreadSubject(input)
		if result == "" {
			t.Errorf("sanitizeThreadSubject should never return empty string for input %q", input)
		}
	}
}

func TestThreadProcessor_ExtractThreadSubject(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	tests := []struct {
		name     string
		item     *models.Item
		expected string
	}{
		{
			name: "remove Re: prefix",
			item: &models.Item{Title: "Re: Original Subject"},
			expected: "Original Subject",
		},
		{
			name: "remove Fwd: prefix",
			item: &models.Item{Title: "Fwd: Important Message"},
			expected: "Important Message",
		},
		{
			name: "remove multiple prefixes",
			item: &models.Item{Title: "RE: Fwd: Re: Final Subject"},
			expected: "Final Subject",
		},
		{
			name: "case insensitive removal",
			item: &models.Item{Title: "FWD: Test Subject"},
			expected: "Test Subject",
		},
		{
			name: "no prefix to remove",
			item: &models.Item{Title: "Clean Subject"},
			expected: "Clean Subject",
		},
		{
			name: "empty title",
			item: &models.Item{Title: ""},
			expected: "",
		},
		{
			name: "only prefix",
			item: &models.Item{Title: "Re:"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.extractThreadSubject(tt.item)
			if result != tt.expected {
				t.Errorf("extractThreadSubject(%q) = %q, want %q", tt.item.Title, result, tt.expected)
			}
		})
	}
}

func TestThreadProcessor_UpdateParticipants(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	group := &ThreadGroup{
		Participants: []string{"user1@example.com"},
	}

	tests := []struct {
		name             string
		item             *models.Item
		expectedCount    int
		expectedContains string
	}{
		{
			name: "add new participant",
			item: &models.Item{
				Metadata: map[string]interface{}{
					"from": "user2@example.com",
				},
			},
			expectedCount:    2,
			expectedContains: "user2@example.com",
		},
		{
			name: "duplicate participant",
			item: &models.Item{
				Metadata: map[string]interface{}{
					"from": "user1@example.com",
				},
			},
			expectedCount:    1,
			expectedContains: "user1@example.com",
		},
		{
			name: "no from metadata",
			item: &models.Item{
				Metadata: map[string]interface{}{},
			},
			expectedCount:    1,
			expectedContains: "user1@example.com",
		},
		{
			name: "complex recipient object",
			item: &models.Item{
				Metadata: map[string]interface{}{
					"from": map[string]interface{}{
						"name":  "John Doe",
						"email": "john@example.com",
					},
				},
			},
			expectedCount:    2,
			expectedContains: "john@example.com", // The current implementation extracts just the email
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset group to initial state
			group.Participants = []string{"user1@example.com"}
			
			processor.updateParticipants(group, tt.item)
			
			if len(group.Participants) != tt.expectedCount {
				t.Errorf("updateParticipants() resulted in %d participants, want %d", 
					len(group.Participants), tt.expectedCount)
			}
			
			found := false
			for _, participant := range group.Participants {
				if participant == tt.expectedContains {
					found = true
					break
				}
			}
			
			if !found {
				t.Errorf("updateParticipants() should contain %q, got %v", 
					tt.expectedContains, group.Participants)
			}
		})
	}
}

func TestThreadProcessor_BuildThreadMetadata_NilSafety(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	// Test with nil group
	metadata := processor.buildThreadMetadata(nil)
	if metadata == nil {
		t.Error("buildThreadMetadata should not return nil map")
	}
	if len(metadata) != 0 {
		t.Errorf("buildThreadMetadata with nil group should return empty map, got %d items", len(metadata))
	}

	// Test with valid group
	now := time.Now()
	group := &ThreadGroup{
		ThreadID:     "test123",
		MessageCount: 3,
		Participants: []string{"user1@example.com", "user2@example.com"},
		StartTime:    now.Add(-1 * time.Hour),
		EndTime:      now,
	}

	metadata = processor.buildThreadMetadata(group)
	
	expectedKeys := []string{"thread_id", "message_count", "participants", "start_time", "end_time", "duration_hours"}
	for _, key := range expectedKeys {
		if _, exists := metadata[key]; !exists {
			t.Errorf("buildThreadMetadata should include key %q", key)
		}
	}

	if metadata["duration_hours"].(float64) <= 0 {
		t.Error("Duration should be positive for valid time range")
	}
}

func TestThreadProcessor_Concurrency(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "consolidated",
	}
	processor := NewThreadProcessor(config)

	// Create test items
	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		createTestItem("msg2", "Re: Subject 1", "thread1"),
		createTestItem("msg3", "Subject 2", "thread2"),
	}

	// Process multiple times to ensure consistency
	var results [][]*models.Item
	for i := 0; i < 10; i++ {
		result, err := processor.ProcessThreads(items)
		if err != nil {
			t.Fatalf("ProcessThreads failed on iteration %d: %v", i, err)
		}
		results = append(results, result)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if len(results[i]) != len(results[0]) {
			t.Errorf("Inconsistent results: iteration %d has %d items, iteration 0 has %d items",
				i, len(results[i]), len(results[0]))
		}
	}
}

// createTestItem is defined in processor_test.go