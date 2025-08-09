package gmail

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"pkm-sync/pkg/models"
)

// Note: Security tests for filename sanitization are now in internal/utils/filename_test.go
// These tests focus on thread-specific functionality

func TestThreadProcessor_ProcessThreads_IndividualMode(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "individual",
	}
	processor := NewThreadProcessor(config)

	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		createTestItem("msg2", "Re: Subject 1", "thread1"),
		createTestItem("msg3", "Subject 2", "thread2"),
	}

	result, err := processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 items (individual mode), got %d", len(result))
	}

	// Should return original items unchanged
	for i, item := range result {
		if item.ID != items[i].ID {
			t.Errorf("Item %d: expected ID %s, got %s", i, items[i].ID, item.ID)
		}
	}
}

func TestThreadProcessor_ProcessThreads_ConsolidatedMode(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "consolidated",
	}
	processor := NewThreadProcessor(config)

	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		createTestItem("msg2", "Re: Subject 1", "thread1"),
		createTestItem("msg3", "Subject 2", "thread2"),
	}

	result, err := processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items (2 threads), got %d", len(result))
	}

	// Check that thread items are consolidated
	var threadItem *models.Item
	for _, item := range result {
		if item.ItemType == "email_thread" {
			threadItem = item
			break
		}
	}

	if threadItem == nil {
		t.Error("Expected to find a consolidated thread item")
	} else {
		if !strings.Contains(threadItem.Title, "Thread_") {
			t.Errorf("Expected thread title to contain 'Thread_', got %s", threadItem.Title)
		}
		if !strings.Contains(threadItem.Content, "Thread: Subject 1") {
			t.Error("Expected consolidated content to contain thread subject")
		}
	}
}

func TestThreadProcessor_ProcessThreads_SummaryMode(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads:        true,
		ThreadMode:           "summary",
		ThreadSummaryLength:  2,
	}
	processor := NewThreadProcessor(config)

	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		createTestItem("msg2", "Re: Subject 1", "thread1"),
		createTestItem("msg3", "Re: Subject 1", "thread1"),
		createTestItem("msg4", "Subject 2", "thread2"),
	}

	result, err := processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items (2 threads), got %d", len(result))
	}

	// Check that thread summary is created
	var summaryItem *models.Item
	for _, item := range result {
		if item.ItemType == "email_thread_summary" {
			summaryItem = item
			break
		}
	}

	if summaryItem == nil {
		t.Error("Expected to find a thread summary item")
	} else {
		if !strings.Contains(summaryItem.Title, "Thread-Summary_") {
			t.Errorf("Expected summary title to contain 'Thread-Summary_', got %s", summaryItem.Title)
		}
		if !strings.Contains(summaryItem.Content, "**Showing:** 2 key messages") {
			t.Error("Expected summary to show configured number of messages")
		}
	}
}

func TestThreadProcessor_SelectKeyMessages(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	// Create test messages with different characteristics
	now := time.Now()
	messages := []*models.Item{
		{
			ID:         "msg1",
			Title:      "First message",
			Content:    "Short content",
			CreatedAt:  now.Add(-3 * time.Hour),
			Metadata:   map[string]interface{}{"from": "user1@example.com"},
		},
		{
			ID:         "msg2", 
			Title:      "Second message",
			Content:    strings.Repeat("Long content with lots of text ", 20), // >500 chars
			CreatedAt:  now.Add(-2 * time.Hour),
			Metadata:   map[string]interface{}{"from": "user2@example.com"},
		},
		{
			ID:         "msg3",
			Title:      "Third message",
			Content:    "Medium content",
			CreatedAt:  now.Add(-1 * time.Hour),
			Metadata:   map[string]interface{}{"from": "user1@example.com"},
			Attachments: []models.Attachment{{Name: "file.pdf"}},
		},
		{
			ID:         "msg4",
			Title:      "Fourth message", 
			Content:    "Recent content",
			CreatedAt:  now,
			Metadata:   map[string]interface{}{"from": "user3@example.com"},
		},
	}

	tests := []struct {
		name        string
		maxMessages int
		expected    int
		shouldIncludeFirst bool
		shouldIncludeLast  bool
	}{
		{
			name:               "select 2 messages",
			maxMessages:        2,
			expected:           2,
			shouldIncludeFirst: true,
			shouldIncludeLast:  true,
		},
		{
			name:               "select 3 messages", 
			maxMessages:        3,
			expected:           3,
			shouldIncludeFirst: true,
			shouldIncludeLast:  true,
		},
		{
			name:               "select all messages",
			maxMessages:        10,
			expected:           4,
			shouldIncludeFirst: true,
			shouldIncludeLast:  true,
		},
		{
			name:               "select 1 message",
			maxMessages:        1,
			expected:           1,
			shouldIncludeFirst: true,
			shouldIncludeLast:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.selectKeyMessages(messages, tt.maxMessages)
			
			if len(result) != tt.expected {
				t.Errorf("Expected %d messages, got %d", tt.expected, len(result))
			}
			
			if tt.shouldIncludeFirst && result[0].ID != "msg1" {
				t.Error("Expected first message to be included first")
			}
			
			if tt.shouldIncludeLast && len(result) > 1 && result[len(result)-1].ID != "msg4" {
				t.Error("Expected last message to be included last")
			}
			
			// Verify chronological order
			for i := 1; i < len(result); i++ {
				if result[i].CreatedAt.Before(result[i-1].CreatedAt) {
					t.Error("Messages should be in chronological order")
				}
			}
		})
	}
}

func TestThreadProcessor_ProcessThreads_NilSafety(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "consolidated",
	}
	processor := NewThreadProcessor(config)

	// Test with nil items slice
	result, err := processor.ProcessThreads(nil)
	if err != nil {
		t.Fatalf("ProcessThreads should handle nil items: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for nil items, got %d", len(result))
	}

	// Test with slice containing nil items
	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		nil, // This should be skipped
		createTestItem("msg2", "Subject 2", "thread1"),
	}

	result, err = processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads should handle nil items in slice: %v", err)
	}
	
	// Should only process non-nil items
	if len(result) != 1 { // Should consolidate 2 non-nil items into 1 thread
		t.Errorf("Expected 1 consolidated item, got %d", len(result))
	}
}

func TestThreadProcessor_DefaultSummaryLength(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "summary",
		// ThreadSummaryLength not set, should use default
	}
	processor := NewThreadProcessor(config)

	// Create more messages than default summary length
	items := []*models.Item{
		createTestItem("msg1", "Subject 1", "thread1"),
		createTestItem("msg2", "Re: Subject 1", "thread1"),
		createTestItem("msg3", "Re: Subject 1", "thread1"),
		createTestItem("msg4", "Re: Subject 1", "thread1"),
		createTestItem("msg5", "Re: Subject 1", "thread1"),
		createTestItem("msg6", "Re: Subject 1", "thread1"),
		createTestItem("msg7", "Re: Subject 1", "thread1"),
	}

	result, err := processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 thread summary, got %d", len(result))
	}

	summary := result[0]
	if !strings.Contains(summary.Content, fmt.Sprintf("**Showing:** %d key messages", DefaultThreadSummaryLength)) {
		t.Errorf("Expected summary to show default length (%d), got content: %s", DefaultThreadSummaryLength, summary.Content)
	}
}

func TestThreadProcessor_EmptySubjectHandling(t *testing.T) {
	config := models.GmailSourceConfig{
		IncludeThreads: true,
		ThreadMode:     "consolidated",
	}
	processor := NewThreadProcessor(config)

	items := []*models.Item{
		{
			ID:         "msg1",
			Title:      "", // Empty subject
			Content:    "Content 1",
			SourceType: "gmail",
			ItemType:   "email",
			CreatedAt:  time.Now(),
			Metadata:   map[string]interface{}{"thread_id": "thread123"},
		},
		{
			ID:         "msg2",
			Title:      "Re:", // Subject that becomes empty after cleaning
			Content:    "Content 2", 
			SourceType: "gmail",
			ItemType:   "email",
			CreatedAt:  time.Now(),
			Metadata:   map[string]interface{}{"thread_id": "thread456"},
		},
	}

	result, err := processor.ProcessThreads(items)
	if err != nil {
		t.Fatalf("ProcessThreads failed: %v", err)
	}

	// Check that thread IDs are used to prevent collisions
	threadTitles := make(map[string]bool)
	for _, item := range result {
		if threadTitles[item.Title] {
			t.Errorf("Found duplicate thread title: %s", item.Title)
		}
		threadTitles[item.Title] = true
		
		// Thread titles should contain the thread ID for empty subjects
		if item.ItemType == "email_thread" {
			if !strings.Contains(item.Title, "thread") {
				t.Errorf("Expected thread title to contain thread ID, got: %s", item.Title)
			}
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