package gmail

import (
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"

	"pkm-sync/internal/sources/google/gmail/testdata"
	"pkm-sync/pkg/models"
)

func TestFromGmailMessageWithFixtures(t *testing.T) {
	testEmails, err := testdata.LoadTestEmails()
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tests := []struct {
		name    string
		message func() interface{}
		config  models.GmailSourceConfig
		checks  func(*testing.T, *models.Item)
	}{
		{
			name:    "simple text email fixture",
			message: func() interface{} { return testEmails.SimpleTextEmail },
			config: models.GmailSourceConfig{
				ExtractRecipients: true,
				ExtractLinks:      false,
			},
			checks: func(t *testing.T, item *models.Item) {
				if item.Title != "Test Email - Simple Text" {
					t.Errorf("Expected subject 'Test Email - Simple Text', got '%s'", item.Title)
				}
				
				if item.SourceType != "gmail" {
					t.Errorf("Expected source type 'gmail', got '%s'", item.SourceType)
				}
				
				if item.ItemType != "email" {
					t.Errorf("Expected item type 'email', got '%s'", item.ItemType)
				}
				
				// Check metadata for recipient extraction
				from, ok := item.Metadata["from"].(EmailRecipient)
				if !ok {
					t.Error("Expected 'from' metadata to be EmailRecipient")
				} else {
					if from.Email != "john.doe@example.com" {
						t.Errorf("Expected from email 'john.doe@example.com', got '%s'", from.Email)
					}
					if from.Name != "John Doe" {
						t.Errorf("Expected from name 'John Doe', got '%s'", from.Name)
					}
				}
				
				to, ok := item.Metadata["to"].([]EmailRecipient)
				if !ok {
					t.Error("Expected 'to' metadata to be []EmailRecipient")
				} else if len(to) != 1 {
					t.Errorf("Expected 1 recipient, got %d", len(to))
				} else {
					if to[0].Email != "jane.smith@example.com" {
						t.Errorf("Expected to email 'jane.smith@example.com', got '%s'", to[0].Email)
					}
				}
			},
		},
		{
			name:    "HTML email with links fixture",
			message: func() interface{} { return testEmails.HTMLEmailWithLinks },
			config: models.GmailSourceConfig{
				ProcessHTMLContent: true,
				ExtractLinks:       true,
				ExtractRecipients:  true,
			},
			checks: func(t *testing.T, item *models.Item) {
				if item.Title != "Weekly Newsletter - HTML Format" {
					t.Errorf("Expected subject 'Weekly Newsletter - HTML Format', got '%s'", item.Title)
				}
				
				// Should have extracted links
				if len(item.Links) == 0 {
					t.Error("Expected links to be extracted from HTML content")
				}
				
				// Content should be processed (no HTML tags)
				if len(item.Content) == 0 {
					t.Error("Expected content to be processed")
				}
				
				// Check CC recipients
				cc, ok := item.Metadata["cc"].([]EmailRecipient)
				if !ok {
					t.Error("Expected 'cc' metadata to be []EmailRecipient")
				} else if len(cc) != 2 {
					t.Errorf("Expected 2 CC recipients, got %d", len(cc))
				}
			},
		},
		{
			name:    "email with attachments fixture",
			message: func() interface{} { return testEmails.EmailWithAttachments },
			config: models.GmailSourceConfig{
				ExtractRecipients: true,
				DownloadAttachments: true,
				TaggingRules: []models.TaggingRule{
					{
						Condition: "has:attachment",
						Tags:      []string{"has-files"},
					},
				},
			},
			checks: func(t *testing.T, item *models.Item) {
				if item.Title != "Project Documents - Q1 Report" {
					t.Errorf("Expected subject 'Project Documents - Q1 Report', got '%s'", item.Title)
				}
				
				// Should have 3 attachments
				if len(item.Attachments) != 3 {
					t.Errorf("Expected 3 attachments, got %d", len(item.Attachments))
				}
				
				// Check attachment details
				if len(item.Attachments) > 0 {
					pdf := item.Attachments[0]
					if pdf.Name != "Q1_Report_2024.pdf" {
						t.Errorf("Expected PDF name 'Q1_Report_2024.pdf', got '%s'", pdf.Name)
					}
					if pdf.MimeType != "application/pdf" {
						t.Errorf("Expected PDF MIME type 'application/pdf', got '%s'", pdf.MimeType)
					}
				}
				
				// Check tagging rule applied
				if !containsString(item.Tags, "has-files") {
					t.Error("Expected 'has-files' tag from tagging rule")
				}
			},
		},
		{
			name:    "complex recipients email fixture", 
			message: func() interface{} { return testEmails.ComplexRecipientsEmail },
			config: models.GmailSourceConfig{
				ExtractRecipients: true,
				TaggingRules: []models.TaggingRule{
					{
						Condition: "from:ceo@company.com",
						Tags:      []string{"executive", "priority"},
					},
				},
			},
			checks: func(t *testing.T, item *models.Item) {
				if item.Title != "Meeting Invitation - All Hands" {
					t.Errorf("Expected subject 'Meeting Invitation - All Hands', got '%s'", item.Title)
				}
				
				// Check complex from parsing
				from, ok := item.Metadata["from"].(EmailRecipient)
				if !ok {
					t.Error("Expected 'from' metadata to be EmailRecipient")
				} else {
					if from.Email != "ceo@company.com" {
						t.Errorf("Expected from email 'ceo@company.com', got '%s'", from.Email)
					}
					if from.Name != "CEO, Company Inc." {
						t.Errorf("Expected from name 'CEO, Company Inc.', got '%s'", from.Name)
					}
				}
				
				// Check multiple TO recipients with quoted names
				to, ok := item.Metadata["to"].([]EmailRecipient)
				if !ok {
					t.Error("Expected 'to' metadata to be []EmailRecipient")
				} else if len(to) != 3 {
					t.Errorf("Expected 3 TO recipients, got %d", len(to))
				} else {
					// Check parsing of quoted names with commas
					if to[0].Name != "Smith, John" {
						t.Errorf("Expected first recipient name 'Smith, John', got '%s'", to[0].Name)
					}
					if to[1].Name != "Doe, Jane" {
						t.Errorf("Expected second recipient name 'Doe, Jane', got '%s'", to[1].Name)
					}
				}
				
				// Check CC recipients
				cc, ok := item.Metadata["cc"].([]EmailRecipient)
				if !ok {
					t.Error("Expected 'cc' metadata to be []EmailRecipient")
				} else if len(cc) != 2 {
					t.Errorf("Expected 2 CC recipients, got %d", len(cc))
				}
				
				// Check BCC recipients 
				bcc, ok := item.Metadata["bcc"].([]EmailRecipient)
				if !ok {
					t.Error("Expected 'bcc' metadata to be []EmailRecipient")
				} else if len(bcc) != 1 {
					t.Errorf("Expected 1 BCC recipient, got %d", len(bcc))
				}
				
				// Check tagging rule applied
				if !containsString(item.Tags, "executive") || !containsString(item.Tags, "priority") {
					t.Error("Expected 'executive' and 'priority' tags from tagging rule")
				}
			},
		},
		{
			name:    "quoted reply email fixture",
			message: func() interface{} { return testEmails.QuotedReplyEmail },
			config: models.GmailSourceConfig{
				StripQuotedText:   true,
				ExtractRecipients: true,
			},
			checks: func(t *testing.T, item *models.Item) {
				if item.Title != "Re: Project Update" {
					t.Errorf("Expected subject 'Re: Project Update', got '%s'", item.Title)
				}
				
				// Content should have quoted text stripped
				if len(item.Content) == 0 {
					t.Error("Expected content after processing")
				}
				
				// Should not contain quoted lines (starting with >)
				if strings.Contains(item.Content, ">") {
					t.Error("Expected quoted text to be stripped from content")
				}
				
				// Check reply metadata
				messageID := item.Metadata["message_id"]
				if messageID != "<reply005@company.com>" {
					t.Errorf("Expected message ID '<reply005@company.com>', got '%v'", messageID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the test message to gmail.Message type
			var msg *gmail.Message
			switch m := tt.message().(type) {
			case *gmail.Message:
				msg = m
			default:
				t.Fatalf("Invalid message type for test %s", tt.name)
			}
			
			if msg == nil {
				t.Fatalf("Test message is nil for %s", tt.name)
			}
			
			item, err := FromGmailMessage(msg, tt.config)
			if err != nil {
				t.Fatalf("FromGmailMessage() error = %v", err)
			}
			
			if item == nil {
				t.Fatal("FromGmailMessage() returned nil item")
			}
			
			tt.checks(t, item)
		})
	}
}

func TestGmailConverterPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}
	
	testEmail, err := testdata.LoadTestEmail("with_attachments")
	if err != nil {
		t.Fatalf("Failed to load test email: %v", err)
	}
	
	config := models.GmailSourceConfig{
		ProcessHTMLContent: true,
		ExtractLinks:       true,
		ExtractRecipients:  true,
		StripQuotedText:    true,
		TaggingRules: []models.TaggingRule{
			{Condition: "has:attachment", Tags: []string{"files"}},
			{Condition: "from:company.com", Tags: []string{"internal"}},
		},
	}
	
	// Benchmark the conversion
	const iterations = 1000
	
	for i := 0; i < iterations; i++ {
		_, err := FromGmailMessage(testEmail, config)
		if err != nil {
			t.Fatalf("FromGmailMessage() failed on iteration %d: %v", i, err)
		}
	}
}

// Helper function to check if slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

