package gmail

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"

	"pkm-sync/pkg/models"
)

func TestProcessHTMLContent_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "nested formatting",
			html: "<p><strong>Bold <em>and italic</em></strong> text</p>",
			expected: "**Bold *and italic*** text",
		},
		{
			name: "table conversion",
			html: `<table>
				<tr><th>Header 1</th><th>Header 2</th></tr>
				<tr><td>Cell 1</td><td>Cell 2</td></tr>
			</table>`,
			expected: "| Header 1 | Header 2 | \n| Cell 1 | Cell 2 |",
		},
		{
			name: "mixed content types",
			html: `<h1>Main Title</h1>
				<p>Some <strong>bold</strong> text with <a href="http://example.com">a link</a>.</p>
				<ul>
					<li>First item</li>
					<li>Second item with <em>emphasis</em></li>
				</ul>
				<blockquote>This is a quote</blockquote>`,
			expected: "# Main Title\nSome **bold** text with [a link](http://example.com).\n- First item\n- Second item with *emphasis*\n> This is a quote",
		},
		{
			name: "email with images and links",
			html: `<div>
				<p>Check out this image: <img src="https://example.com/image.jpg" alt="Sample Image"></p>
				<p>And visit <a href="https://example.com">our website</a> for more!</p>
			</div>`,
			expected: "Check out this image: ![Sample Image](https://example.com/image.jpg)\nAnd visit [our website](https://example.com) for more!",
		},
		{
			name: "code blocks and inline code",
			html: `<p>Use <code>git commit</code> to save changes:</p>
				<pre>git add .
git commit -m "Update files"
git push origin main</pre>`,
			expected: "Use `git commit` to save changes:\n```\ngit add .\ngit commit -m \"Update files\"\ngit push origin main\n```",
		},
		{
			name: "heavily nested structure",
			html: `<div class="email-content">
				<div class="header">
					<h2>Weekly Update</h2>
				</div>
				<div class="body">
					<p>Hello team,</p>
					<div class="section">
						<h3>Progress</h3>
						<ul>
							<li>Feature A: <strong>Completed</strong></li>
							<li>Feature B: <em>In Progress</em></li>
						</ul>
					</div>
				</div>
			</div>`,
			expected: "## Weekly Update\nHello team,\n### Progress\n- Feature A: **Completed**\n- Feature B: *In Progress*",
		},
		{
			name: "email with signature and quoted content",
			html: `<div>
				<p>Thanks for the update!</p>
				<br>
				<div class="signature">
					<p>--<br>John Doe<br>Software Engineer</p>
				</div>
				<div class="quoted">
					<blockquote>
						<p>On Mon, Jan 1, 2024, Jane wrote:</p>
						<p>Here's the original message...</p>
					</blockquote>
				</div>
			</div>`,
			expected: "Thanks for the update!\n--\nJohn Doe\nSoftware Engineer\n> On Mon, Jan 1, 2024, Jane wrote:\n> Here's the original message...",
		},
		{
			name: "malformed HTML handling",
			html: `<p>Start paragraph<div>Nested div</div>Missing close tag
				<strong>Bold text<em>Italic inside bold</strong>Unclosed italic</em>`,
			expected: "Start paragraph\nNested div\nMissing close tag\n**Bold text*Italic inside bold***Unclosed italic*",
		},
		{
			name: "HTML with style and script removal",
			html: `<html>
				<head>
					<style>
						.header { color: blue; font-size: 18px; }
						.content { margin: 10px; }
					</style>
					<script>
						function trackClick() { analytics.track('click'); }
					</script>
				</head>
				<body>
					<div class="header">Important Message</div>
					<div class="content">This is the actual content.</div>
				</body>
			</html>`,
			expected: "Important Message\nThis is the actual content.",
		},
		{
			name: "complex list structures",
			html: `<div>
				<h3>Project Tasks</h3>
				<ol>
					<li>Setup environment
						<ul>
							<li>Install dependencies</li>
							<li>Configure database</li>
						</ul>
					</li>
					<li>Development
						<ul>
							<li>Write tests</li>
							<li>Implement features</li>
						</ul>
					</li>
				</ol>
			</div>`,
			expected: "### Project Tasks\n- Setup environment\n- Install dependencies\n- Configure database\n- Development\n- Write tests\n- Implement features",
		},
		{
			name: "special characters and entities",
			html: `<p>&ldquo;Hello&rdquo; &amp; &lt;goodbye&gt; &mdash; with &hellip; and &nbsp; spaces</p>`,
			expected: `"Hello" & <goodbye> — with ... and   spaces`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(models.GmailSourceConfig{ProcessHTMLContent: true})
			result := processor.ProcessHTMLContent(tt.html)
			
			// Normalize whitespace for comparison
			result = normalizeWhitespace(result)
			expected := normalizeWhitespace(tt.expected)
			
			if result != expected {
				t.Errorf("ProcessHTMLContent() = %q, want %q", result, expected)
			}
		})
	}
}

func TestStripQuotedText_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "multiple quote styles mixed",
			content: `New message here

> Quoted with angle bracket
>> Nested quote
> More quoted text

On 2024-01-01, someone wrote:
Even more original content`,
			expected: "New message here",
		},
		{
			name: "false positive prevention",
			content: `Meeting agenda:
1. Project update
2. On time delivery discussion
3. Next steps

Please review.`,
			expected: `Meeting agenda:
1. Project update
2. On time delivery discussion
3. Next steps

Please review.`,
		},
		{
			name: "complex forwarded message",
			content: `Please see below for details.

---------- Forwarded message ---------
From: Project Manager <pm@company.com>
Date: Mon, Jan 1, 2024 at 3:00 PM
Subject: Project Update
To: Team <team@company.com>

Original forwarded content here...`,
			expected: "Please see below for details.",
		},
		{
			name: "outlook style original message",
			content: `Thanks for the info!

From: sender@example.com
Sent: Monday, January 1, 2024 9:00 AM
To: recipient@example.com
Subject: RE: Project

Original message content`,
			expected: "Thanks for the info!",
		},
		{
			name: "gmail style quoted text",
			content: `My response to your question.

On Mon, Jan 1, 2024 at 9:00 AM John Doe <john@example.com> wrote:

Previous message content here
with multiple lines of quoted text.`,
			expected: "My response to your question.",
		},
		{
			name: "signature vs quoted text disambiguation",
			content: `Message content here.

-- 
John Doe
Senior Developer
john@company.com

> This is actually quoted text
> not part of the signature`,
			expected: "Message content here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewContentProcessor(models.GmailSourceConfig{StripQuotedText: true})
			result := processor.StripQuotedText(tt.content)
			
			if result != tt.expected {
				t.Errorf("StripQuotedText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractSignatures_ComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "corporate signature with contact info",
			content: `Thanks for reaching out!

Best regards,
John Doe
Senior Software Engineer
Acme Corporation
Phone: +1 (555) 123-4567
Email: john.doe@acme.com
Website: https://acme.com`,
			expected: "Thanks for reaching out!",
		},
		{
			name: "mobile signature variations",
			content: `Quick update on the project.

Sent from my iPhone
Please excuse any typos.`,
			expected: "Quick update on the project.",
		},
		{
			name: "signature with legal disclaimer",
			content: `Please review the attached document.

Sincerely,
Jane Smith
Legal Department

CONFIDENTIALITY NOTICE: This email may contain confidential information.`,
			expected: "Please review the attached document.",
		},
		{
			name: "multiple signature patterns",
			content: `Email body content here.

Thanks,
Bob
--
Robert Johnson
Director of Engineering
bob.johnson@company.com
This email is confidential.`,
			expected: "Email body content here.",
		},
		{
			name: "signature with social media links",
			content: `Great meeting today!

Cheers,
Alice Cooper
Marketing Manager
alice@company.com
LinkedIn: linkedin.com/in/alicecooper
Twitter: @alice_cooper`,
			expected: "Great meeting today!",
		},
		{
			name: "no clear signature boundary",
			content: `Long email content with multiple paragraphs.

This is still part of the main content.
Even though it's near the end.

And this too.`,
			expected: `Long email content with multiple paragraphs.

This is still part of the main content.
Even though it's near the end.

And this too.`,
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

func TestProcessEmailBody_IntegrationScenarios(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		config   models.GmailSourceConfig
		expected string
	}{
		{
			name: "full processing pipeline",
			html: `<div>
				<p>Hi team,</p>
				<p>Please review the <a href="https://docs.company.com">documentation</a>.</p>
				<blockquote>
					<p>On Mon, someone wrote:</p>
					<p>Original message here...</p>
				</blockquote>
				<div class="signature">
					<p>--<br>John Doe</p>
				</div>
			</div>`,
			config: models.GmailSourceConfig{
				ProcessHTMLContent: true,
				StripQuotedText:    true,
				ExtractSignatures:  true,
				ExtractLinks:       true,
			},
			expected: "Hi team,\nPlease review the [documentation](https://docs.company.com).",
		},
		{
			name: "HTML processing disabled",
			html: `<p>Hello <strong>world</strong>!</p>`,
			config: models.GmailSourceConfig{
				ProcessHTMLContent: false,
			},
			expected: `<p>Hello <strong>world</strong>!</p>`,
		},
		{
			name: "selective processing",
			html: `<div>
				<p>Message content</p>
				<blockquote>Quoted text here</blockquote>
			</div>`,
			config: models.GmailSourceConfig{
				ProcessHTMLContent: true,
				StripQuotedText:    false, // Keep quoted text
			},
			expected: "Message content\n> Quoted text here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Gmail message with HTML content
			msg := createMockGmailMessage(tt.html)
			
			processor := NewContentProcessor(tt.config)
			result, err := processor.ProcessEmailBody(msg)
			
			if err != nil {
				t.Errorf("ProcessEmailBody() error = %v", err)
				return
			}
			
			result = normalizeWhitespace(result)
			expected := normalizeWhitespace(tt.expected)
			
			if result != expected {
				t.Errorf("ProcessEmailBody() = %q, want %q", result, expected)
			}
		})
	}
}

// Helper functions for tests

// normalizeWhitespace normalizes whitespace for test comparisons
func normalizeWhitespace(s string) string {
	// Replace multiple consecutive whitespace with single spaces
	// and trim leading/trailing whitespace
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
}

// createMockGmailMessage creates a mock Gmail message with HTML content for testing
func createMockGmailMessage(htmlContent string) *gmail.Message {
	// Encode HTML content as base64
	encodedContent := base64.URLEncoding.EncodeToString([]byte(htmlContent))
	
	return &gmail.Message{
		Id:      "test-message-123",
		Snippet: "Test message snippet",
		Payload: &gmail.MessagePart{
			MimeType: "multipart/alternative",
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/html",
					Body: &gmail.MessagePartBody{
						Data: encodedContent,
					},
				},
			},
		},
	}
}