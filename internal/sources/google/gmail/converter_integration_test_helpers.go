package gmail

import (
	"testing"

	"pkm-sync/internal/sources/google/gmail/testdata"
	"pkm-sync/pkg/models"

	"google.golang.org/api/gmail/v1"
)

func setupConverterTest(t *testing.T, emailName string) (*gmail.Message, models.GmailSourceConfig) {
	testEmails, err := testdata.LoadTestEmails()
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	var message *gmail.Message

	switch emailName {
	case "simple_text":
		message = testEmails.SimpleTextEmail
	case "html_with_links":
		message = testEmails.HTMLEmailWithLinks
	case "with_attachments":
		message = testEmails.EmailWithAttachments
	case "complex_recipients":
		message = testEmails.ComplexRecipientsEmail
	case "quoted_reply":
		message = testEmails.QuotedReplyEmail
	default:
		t.Fatalf("Unknown email fixture name: %s", emailName)
	}

	config := models.GmailSourceConfig{
		ExtractRecipients:  true,
		ExtractLinks:       true,
		ProcessHTMLContent: true,
		StripQuotedText:    true,
	}

	return message, config
}
