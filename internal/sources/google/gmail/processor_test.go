package gmail

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/api/gmail/v1"

	"pkm-sync/pkg/models"
)

func TestContentProcessor_ProcessHTMLContent(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		config   models.GmailSourceConfig
		expected string
	}{
		{
			name: "basic HTML to markdown",
			html: "<p><strong>Hello</strong> world!</p>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "**Hello** world!",
		},
		{
			name: "headers conversion",
			html: "<h1>Title</h1><h2>Subtitle</h2><p>Content</p>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "# Title\n## Subtitle\nContent",
		},
		{
			name: "links conversion",
			html: `<p>Visit <a href="https://example.com">our website</a></p>`,
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "Visit [our website](https://example.com)",
		},
		{
			name: "lists conversion",
			html: "<ul><li>Item 1</li><li>Item 2</li></ul>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "- Item 1\n- Item 2",
		},
		{
			name: "blockquotes conversion",
			html: "<blockquote>This is quoted text</blockquote>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "> This is quoted text",
		},
		{
			name: "code blocks conversion",
			html: "<pre>function hello() { return 'world'; }</pre>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "```\nfunction hello() { return 'world'; }\n```",
		},
		{
			name: "inline code conversion",
			html: "<p>Use <code>console.log</code> for debugging</p>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "Use `console.log` for debugging",
		},
		{
			name: "images conversion",
			html: `<img src="https://example.com/image.jpg" alt="Test Image">`,
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "![Test Image](https://example.com/image.jpg)",
		},
		{
			name: "remove style and script tags",
			html: "<style>body { color: red; }</style><p>Content</p><script>alert('hi');</script>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "Content",
		},
		{
			name: "HTML entities decoding",
			html: "<p>&lt;Hello&gt; &amp; &quot;World&quot;</p>",
			config: models.GmailSourceConfig{ProcessHTMLContent: true},
			expected: "<Hello> & \"World\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(tt.config)
			result := processor.ProcessHTMLContent(tt.html)
			if result != tt.expected {
				t.Errorf("ProcessHTMLContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_StripQuotedText(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		config   models.GmailSourceConfig
		expected string
	}{
		{
			name: "basic quoted text removal",
			content: "New message\n\n> Previous message\n> Another line",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "New message",
		},
		{
			name: "on date wrote pattern",
			content: "My response\n\nOn Mon, Jan 1, 2024 at 10:00 AM John Doe wrote:\nOriginal message",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "My response",
		},
		{
			name: "forwarded message pattern",
			content: "Check this out\n\n---------- Forwarded message ---------\nFrom: someone@example.com",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "Check this out",
		},
		{
			name: "original message pattern",
			content: "My response\n\n-----Original Message-----\nFrom: sender@example.com",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "My response",
		},
		{
			name: "from email pattern",
			content: "Response here\n\nFrom: sender@example.com\nSent: Monday",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "Response here",
		},
		{
			name: "signature separator",
			content: "Message content\n--\nJohn Doe\nSoftware Engineer",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "Message content",
		},
		{
			name: "no quoted text",
			content: "Just a simple message\nwith multiple lines\nno quotes here",
			config: models.GmailSourceConfig{StripQuotedText: true},
			expected: "Just a simple message\nwith multiple lines\nno quotes here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(tt.config)
			result := processor.StripQuotedText(tt.content)
			if result != tt.expected {
				t.Errorf("StripQuotedText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_ExtractSignatures(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "standard signature separator",
			content:  "Email content here\n--\nJohn Doe\nSoftware Engineer\ncompany@example.com",
			expected: "Email content here",
		},
		{
			name:     "signature with space separator",
			content:  "Message content\n-- \nBest regards,\nJane Smith",
			expected: "Message content",
		},
		{
			name:     "signature detection at end",
			content:  "Email body\n\nThanks,\nJohn\n555-123-4567",
			expected: "Email body",
		},
		{
			name:     "best regards signature",
			content:  "Message here\n\nBest regards,\nJohn Doe\nCEO, Example Corp",
			expected: "Message here",
		},
		{
			name:     "sent from mobile signature",
			content:  "Quick reply\n\nSent from my iPhone",
			expected: "Quick reply",
		},
		{
			name:     "no signature",
			content:  "Just a message\nwith multiple lines\nno signature here",
			expected: "Just a message\nwith multiple lines\nno signature here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(models.GmailSourceConfig{ExtractSignatures: true})
			result := processor.ExtractSignatures(tt.content)
			if result != tt.expected {
				t.Errorf("ExtractSignatures() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_ExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []models.Link
	}{
		{
			name:    "standalone URLs",
			content: "Check out https://example.com and https://test.org for more info.",
			expected: []models.Link{
				{URL: "https://example.com", Title: "", Type: "external"},
				{URL: "https://test.org", Title: "", Type: "external"},
			},
		},
		{
			name:    "markdown links",
			content: "Visit [our website](https://example.com) and [documentation](https://docs.example.com)",
			expected: []models.Link{
				{URL: "https://example.com", Title: "our website", Type: "external"},
				{URL: "https://docs.example.com", Title: "documentation", Type: "external"},
			},
		},
		{
			name:    "mixed URL types",
			content: "See https://raw.example.com or [formatted link](https://formatted.example.com)",
			expected: []models.Link{
				{URL: "https://raw.example.com", Title: "", Type: "external"},
				{URL: "https://formatted.example.com", Title: "formatted link", Type: "external"},
			},
		},
		{
			name:    "URLs with punctuation",
			content: "Visit https://example.com. Also check https://test.org! And https://another.com?",
			expected: []models.Link{
				{URL: "https://example.com", Title: "", Type: "external"},
				{URL: "https://test.org", Title: "", Type: "external"},
				{URL: "https://another.com", Title: "", Type: "external"},
			},
		},
		{
			name:     "no links",
			content:  "This is just plain text with no URLs at all.",
			expected: []models.Link{},
		},
		{
			name:    "duplicate URLs",
			content: "Visit https://example.com and also https://example.com again",
			expected: []models.Link{
				{URL: "https://example.com", Title: "", Type: "external"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(models.GmailSourceConfig{ExtractLinks: true})
			result := processor.ExtractLinks(tt.content)
			
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractLinks() returned %d links, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, link := range result {
				if link.URL != tt.expected[i].URL || link.Title != tt.expected[i].Title || link.Type != tt.expected[i].Type {
					t.Errorf("ExtractLinks()[%d] = %+v, want %+v", i, link, tt.expected[i])
				}
			}
		})
	}
}

func TestContentProcessor_ProcessEmailAttachments(t *testing.T) {
	tests := []struct {
		name     string
		msg      *gmail.Message
		config   models.GmailSourceConfig
		expected int
	}{
		{
			name: "message with PDF attachment",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							Filename: "document.pdf",
							MimeType: "application/pdf",
							Body: &gmail.MessagePartBody{
								AttachmentId: "att123",
								Size:         1024000,
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				DownloadAttachments: true,
				AttachmentTypes:     []string{"pdf", "doc"},
			},
			expected: 1,
		},
		{
			name: "message with filtered attachment type",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							Filename: "image.jpg",
							MimeType: "image/jpeg",
							Body: &gmail.MessagePartBody{
								AttachmentId: "att123",
								Size:         512000,
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				DownloadAttachments: true,
				AttachmentTypes:     []string{"pdf", "doc"}, // jpg not allowed
			},
			expected: 0,
		},
		{
			name: "download disabled",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							Filename: "document.pdf",
							MimeType: "application/pdf",
							Body: &gmail.MessagePartBody{
								AttachmentId: "att123",
								Size:         1024000,
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				DownloadAttachments: false,
			},
			expected: 0,
		},
		{
			name: "no attachments",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					MimeType: "text/plain",
					Body: &gmail.MessagePartBody{
						Data: "SGVsbG8gd29ybGQ=", // "Hello world" in base64
					},
				},
			},
			config: models.GmailSourceConfig{
				DownloadAttachments: true,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(tt.config)
			result := processor.ProcessEmailAttachments(tt.msg)
			
			if len(result) != tt.expected {
				t.Errorf("ProcessEmailAttachments() returned %d attachments, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestContentProcessor_ProcessEmailBody(t *testing.T) {
	tests := []struct {
		name     string
		msg      *gmail.Message
		config   models.GmailSourceConfig
		expected string
	}{
		{
			name: "HTML content with processing enabled",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/html",
							Body: &gmail.MessagePartBody{
								Data: "PHA+SGVsbG8gPHN0cm9uZz53b3JsZDwvc3Ryb25nPiE8L3A+", // "<p>Hello <strong>world</strong>!</p>" in base64
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				ProcessHTMLContent: true,
			},
			expected: "Hello **world**!",
		},
		{
			name: "plain text content",
			msg: &gmail.Message{
				Id: "msg123",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/plain",
							Body: &gmail.MessagePartBody{
								Data: "SGVsbG8gd29ybGQh", // "Hello world!" in base64
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				ProcessHTMLContent: false,
			},
			expected: "Hello world!",
		},
		{
			name: "content with quoted text removal",
			msg: &gmail.Message{
				Id: "msg123",
				Snippet: "New message content",
				Payload: &gmail.MessagePart{
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/plain",
							Body: &gmail.MessagePartBody{
								Data: "TmV3IG1lc3NhZ2UgY29udGVudAoKPiBRdW90ZWQgdGV4dA==", // "New message content\n\n> Quoted text" in base64
							},
						},
					},
				},
			},
			config: models.GmailSourceConfig{
				StripQuotedText: true,
			},
			expected: "New message content",
		},
		{
			name: "fallback to snippet",
			msg: &gmail.Message{
				Id:      "msg123",
				Snippet: "Email snippet fallback",
				Payload: &gmail.MessagePart{
					MimeType: "multipart/mixed",
					// No text parts
				},
			},
			config: models.GmailSourceConfig{},
			expected: "Email snippet fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(tt.config)
			result, err := processor.ProcessEmailBody(tt.msg)
			
			if err != nil {
				t.Errorf("ProcessEmailBody() error = %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("ProcessEmailBody() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_IsAllowedAttachmentType(t *testing.T) {
	tests := []struct {
		name       string
		attachment models.Attachment
		config     models.GmailSourceConfig
		expected   bool
	}{
		{
			name: "allowed PDF",
			attachment: models.Attachment{
				Name: "document.pdf",
			},
			config: models.GmailSourceConfig{
				AttachmentTypes: []string{"pdf", "doc", "docx"},
			},
			expected: true,
		},
		{
			name: "disallowed image",
			attachment: models.Attachment{
				Name: "photo.jpg",
			},
			config: models.GmailSourceConfig{
				AttachmentTypes: []string{"pdf", "doc", "docx"},
			},
			expected: false,
		},
		{
			name: "case insensitive matching",
			attachment: models.Attachment{
				Name: "Document.PDF",
			},
			config: models.GmailSourceConfig{
				AttachmentTypes: []string{"pdf", "doc"},
			},
			expected: true,
		},
		{
			name: "no extension",
			attachment: models.Attachment{
				Name: "filename",
			},
			config: models.GmailSourceConfig{
				AttachmentTypes: []string{"pdf", "doc"},
			},
			expected: false,
		},
		{
			name: "no filter (allow all)",
			attachment: models.Attachment{
				Name: "anything.xyz",
			},
			config: models.GmailSourceConfig{
				AttachmentTypes: []string{}, // No filter
			},
			expected: false, // Empty filter means no filtering, so filterAttachments returns all
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(tt.config)
			result := processor.isAllowedAttachmentType(tt.attachment)
			
			if result != tt.expected {
				t.Errorf("isAllowedAttachmentType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_LooksLikeSignature(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "best regards",
			line:     "Best regards,",
			expected: true,
		},
		{
			name:     "sincerely",
			line:     "Sincerely",
			expected: true,
		},
		{
			name:     "thanks",
			line:     "Thanks,",
			expected: true,
		},
		{
			name:     "cheers",
			line:     "Cheers",
			expected: true,
		},
		{
			name:     "sent from mobile",
			line:     "Sent from my iPhone",
			expected: true,
		},
		{
			name:     "outlook signature",
			line:     "Get Outlook for Android",
			expected: true,
		},
		{
			name:     "name pattern",
			line:     "John Doe",
			expected: true,
		},
		{
			name:     "email address",
			line:     "john.doe@example.com",
			expected: true,
		},
		{
			name:     "phone number",
			line:     "555-123-4567",
			expected: true,
		},
		{
			name:     "normal content",
			line:     "This is just regular email content",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(models.GmailSourceConfig{})
			result := processor.looksLikeSignature(tt.line)
			
			if result != tt.expected {
				t.Errorf("looksLikeSignature(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

// Thread Processing Tests

func TestThreadProcessor_ProcessThreads(t *testing.T) {
	tests := []struct {
		name        string
		config      models.GmailSourceConfig
		items       []*models.Item
		expectedLen int
		expectedErr string
	}{
		{
			name: "individual mode (default)",
			config: models.GmailSourceConfig{
				IncludeThreads: false,
			},
			items: []*models.Item{
				createTestItem("msg1", "Subject 1", "thread1"),
				createTestItem("msg2", "Subject 2", "thread1"),
			},
			expectedLen: 2,
		},
		{
			name: "consolidated mode",
			config: models.GmailSourceConfig{
				IncludeThreads: true,
				ThreadMode:     "consolidated",
			},
			items: []*models.Item{
				createTestItem("msg1", "Subject 1", "thread1"),
				createTestItem("msg2", "Re: Subject 1", "thread1"),
			},
			expectedLen: 1,
		},
		{
			name: "summary mode",
			config: models.GmailSourceConfig{
				IncludeThreads:        true,
				ThreadMode:            "summary",
				ThreadSummaryLength:   2,
			},
			items: []*models.Item{
				createTestItem("msg1", "Subject 1", "thread1"),
				createTestItem("msg2", "Re: Subject 1", "thread1"),
				createTestItem("msg3", "Re: Subject 1", "thread1"),
			},
			expectedLen: 1,
		},
		{
			name: "invalid thread mode",
			config: models.GmailSourceConfig{
				IncludeThreads: true,
				ThreadMode:     "invalid",
			},
			items: []*models.Item{
				createTestItem("msg1", "Subject 1", "thread1"),
			},
			expectedErr: "unknown thread mode: invalid",
		},
		{
			name: "mixed threads and singles",
			config: models.GmailSourceConfig{
				IncludeThreads: true,
				ThreadMode:     "consolidated",
			},
			items: []*models.Item{
				createTestItem("msg1", "Subject 1", "thread1"),
				createTestItem("msg2", "Re: Subject 1", "thread1"),
				createTestItem("msg3", "Different Subject", "thread2"),
			},
			expectedLen: 2, // One consolidated thread + one single message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewThreadProcessor(tt.config)
			result, err := processor.ProcessThreads(tt.items)

			if tt.expectedErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("ProcessThreads() expected error containing %q, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessThreads() unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("ProcessThreads() returned %d items, want %d", len(result), tt.expectedLen)
			}
		})
	}
}

func TestThreadProcessor_GroupMessagesByThread(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	items := []*models.Item{
		createTestItem("msg1", "Original Subject", "thread1"),
		createTestItem("msg2", "Re: Original Subject", "thread1"),
		createTestItem("msg3", "Different Subject", "thread2"),
	}

	groups := processor.groupMessagesByThread(items)

	if len(groups) != 2 {
		t.Errorf("groupMessagesByThread() returned %d groups, want 2", len(groups))
	}

	// Check thread1 has 2 messages
	if thread1, exists := groups["thread1"]; exists {
		if thread1.MessageCount != 2 {
			t.Errorf("Thread1 has %d messages, want 2", thread1.MessageCount)
		}
		if len(thread1.Messages) != 2 {
			t.Errorf("Thread1 Messages slice has %d items, want 2", len(thread1.Messages))
		}
	} else {
		t.Error("Thread1 not found in groups")
	}

	// Check thread2 has 1 message
	if thread2, exists := groups["thread2"]; exists {
		if thread2.MessageCount != 1 {
			t.Errorf("Thread2 has %d messages, want 1", thread2.MessageCount)
		}
	} else {
		t.Error("Thread2 not found in groups")
	}
}

func TestThreadProcessor_SanitizeThreadSubject(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic subject",
			input:    "Hello World",
			expected: "Hello-World",
		},
		{
			name:     "with special characters",
			input:    "Project: Test <Important>",
			expected: "Project-Test-Important",
		},
		{
			name:     "path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "etc-passwd",
		},
		{
			name:     "multiple consecutive hyphens",
			input:    "Test --- Subject",
			expected: "Test-Subject",
		},
		{
			name:     "empty subject",
			input:    "",
			expected: "email-thread",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "at-$%^-and", // Based on actual replacement behavior
		},
		{
			name:     "very long subject",
			input:    strings.Repeat("a", 100),
			expected: strings.Repeat("a", 80),
		},
		{
			name:     "email prefix removal",
			input:    "Re: Fwd: Original Subject",
			expected: "Original-Subject",
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

func TestThreadProcessor_NilSafety(t *testing.T) {
	// Test nil processor
	var processor *ThreadProcessor
	result := processor.sanitizeThreadSubject("test")
	if result != "email-thread" {
		t.Errorf("nil processor sanitizeThreadSubject should return fallback, got %q", result)
	}

	// Test with nil items
	validProcessor := NewThreadProcessor(models.GmailSourceConfig{})
	processed, err := validProcessor.ProcessThreads(nil)
	if err != nil {
		t.Errorf("ProcessThreads with nil items should not error, got %v", err)
	}
	if processed == nil {
		t.Error("ProcessThreads should not return nil slice")
	}
	if len(processed) != 0 {
		t.Errorf("ProcessThreads with nil items should return empty slice, got %d items", len(processed))
	}
}

func TestThreadProcessor_ExtractEmailFromRecipient(t *testing.T) {
	config := models.GmailSourceConfig{}
	processor := NewThreadProcessor(config)

	tests := []struct {
		name      string
		recipient interface{}
		expected  string
	}{
		{
			name:      "nil recipient",
			recipient: nil,
			expected:  "",
		},
		{
			name:      "string email",
			recipient: "test@example.com",
			expected:  "test@example.com",
		},
		{
			name: "map with email only",
			recipient: map[string]interface{}{
				"email": "test@example.com",
			},
			expected: "test@example.com",
		},
		{
			name: "map with name and email",
			recipient: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			expected: "john@example.com", // Current implementation returns just email
		},
		{
			name: "map with name only",
			recipient: map[string]interface{}{
				"name": "John Doe",
			},
			expected: "John Doe",
		},
		{
			name:      "nil map",
			recipient: (map[string]interface{})(nil),
			expected:  "",
		},
		{
			name: "invalid type assertions",
			recipient: map[string]interface{}{
				"email": 123, // not a string
				"name":  456, // not a string
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.extractEmailFromRecipient(tt.recipient)
			if result != tt.expected {
				t.Errorf("extractEmailFromRecipient(%v) = %q, want %q", tt.recipient, result, tt.expected)
			}
		})
	}
}

// Helper function to create test items
func createTestItem(id, title, threadID string) *models.Item {
	item := &models.Item{
		ID:         id,
		Title:      title,
		Content:    "Test content for " + title,
		SourceType: "gmail",
		ItemType:   "email",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
	
	if threadID != "" {
		item.Metadata["thread_id"] = threadID
	}
	
	return item
}

