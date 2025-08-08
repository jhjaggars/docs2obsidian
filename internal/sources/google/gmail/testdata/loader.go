package testdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"google.golang.org/api/gmail/v1"
)

// EmailTestData contains sample email data for testing
type EmailTestData struct {
	SimpleTextEmail      *gmail.Message `json:"simple_text_email"`
	HTMLEmailWithLinks   *gmail.Message `json:"html_email_with_links"`
	EmailWithAttachments *gmail.Message `json:"email_with_attachments"`
	ComplexRecipientsEmail *gmail.Message `json:"complex_recipients_email"`
	QuotedReplyEmail     *gmail.Message `json:"quoted_reply_email"`
}

// LoadTestEmails loads sample email data from JSON fixtures
func LoadTestEmails() (*EmailTestData, error) {
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get current file path")
	}
	
	dir := filepath.Dir(filename)
	dataFile := filepath.Join(dir, "sample_emails.json")
	
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read test data file: %w", err)
	}
	
	var testData EmailTestData
	if err := json.Unmarshal(data, &testData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test data: %w", err)
	}
	
	return &testData, nil
}

// LoadTestEmail loads a specific test email by name
func LoadTestEmail(name string) (*gmail.Message, error) {
	testData, err := LoadTestEmails()
	if err != nil {
		return nil, err
	}
	
	switch name {
	case "simple_text":
		return testData.SimpleTextEmail, nil
	case "html_with_links":
		return testData.HTMLEmailWithLinks, nil
	case "with_attachments":
		return testData.EmailWithAttachments, nil
	case "complex_recipients":
		return testData.ComplexRecipientsEmail, nil
	case "quoted_reply":
		return testData.QuotedReplyEmail, nil
	default:
		return nil, fmt.Errorf("unknown test email: %s", name)
	}
}