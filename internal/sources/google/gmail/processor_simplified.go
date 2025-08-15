package gmail

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"pkm-sync/pkg/models"

	"google.golang.org/api/gmail/v1"
)

// SimplifiedContentProcessor handles minimal email content extraction.
// Processing logic has been moved to universal transformers.
type SimplifiedContentProcessor struct {
	config  models.GmailSourceConfig
	service *Service
}

// NewSimplifiedContentProcessor creates a new simplified content processor.
func NewSimplifiedContentProcessor(config models.GmailSourceConfig) *SimplifiedContentProcessor {
	return &SimplifiedContentProcessor{
		config: config,
	}
}

// NewSimplifiedContentProcessorWithService creates a new simplified content processor with service.
func NewSimplifiedContentProcessorWithService(
	config models.GmailSourceConfig,
	service *Service,
) *SimplifiedContentProcessor {
	return &SimplifiedContentProcessor{
		config:  config,
		service: service,
	}
}

// ProcessEmailBody extracts raw email body without processing.
// Content processing is now handled by transformers.
func (p *SimplifiedContentProcessor) ProcessEmailBody(msg *gmail.Message) (string, error) {
	if msg.Payload == nil {
		return "", nil
	}

	// Try to get HTML content first, then plain text
	htmlContent := p.extractBodyPart(msg.Payload, "text/html")
	textContent := p.extractBodyPart(msg.Payload, "text/plain")

	var content string

	// Return raw content - transformers will handle conversion
	if htmlContent != "" {
		content = htmlContent
	} else if textContent != "" {
		content = textContent
	} else {
		// Fallback to snippet
		content = msg.Snippet
	}

	return content, nil
}

// extractBodyPart recursively extracts body content of specified mime type.
func (p *SimplifiedContentProcessor) extractBodyPart(part *gmail.MessagePart, mimeType string) string {
	if part == nil {
		return ""
	}

	// Check if this part matches the desired MIME type
	if part.MimeType == mimeType && part.Body != nil && part.Body.Data != "" {
		// Try URL-safe base64 first, then standard base64
		decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			// Try standard base64 if URL-safe fails
			decoded, err = base64.StdEncoding.DecodeString(part.Body.Data)
		}

		if err == nil {
			return string(decoded)
		}
	}

	// Recursively check parts
	for _, subPart := range part.Parts {
		if content := p.extractBodyPart(subPart, mimeType); content != "" {
			return content
		}
	}

	return ""
}

// ProcessEmailAttachments processes email attachments (unchanged functionality).
func (p *SimplifiedContentProcessor) ProcessEmailAttachments(msg *gmail.Message) []models.Attachment {
	if msg.Payload == nil || !p.config.DownloadAttachments {
		return []models.Attachment{}
	}

	var attachments []models.Attachment

	p.extractAttachmentsFromPart(msg.Payload, msg.Id, &attachments)

	filtered := p.filterAttachments(attachments)

	// If we have a service, fetch the actual attachment data
	if p.service != nil {
		for i := range filtered {
			if err := p.fetchAttachmentData(msg.Id, &filtered[i]); err != nil {
				// Log error but continue with other attachments
				slog.Warn("Failed to fetch attachment data", "attachment_name", filtered[i].Name, "error", err)
			}
		}
	}

	return filtered
}

// extractAttachmentsFromPart recursively extracts attachments from message parts.
func (p *SimplifiedContentProcessor) extractAttachmentsFromPart(
	part *gmail.MessagePart,
	messageID string,
	attachments *[]models.Attachment,
) {
	if part == nil {
		return
	}

	// Check if this part is an attachment
	if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
		attachment := models.Attachment{
			ID:       part.Body.AttachmentId,
			Name:     part.Filename,
			MimeType: part.MimeType,
			Size:     part.Body.Size,
		}

		*attachments = append(*attachments, attachment)
	}

	// Recursively check parts
	for _, subPart := range part.Parts {
		p.extractAttachmentsFromPart(subPart, messageID, attachments)
	}
}

// fetchAttachmentData fetches the actual attachment data from Gmail API.
func (p *SimplifiedContentProcessor) fetchAttachmentData(messageID string, attachment *models.Attachment) error {
	if p.service == nil {
		return fmt.Errorf("service not available for attachment download")
	}

	attachmentData, err := p.service.GetAttachment(messageID, attachment.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch attachment data: %w", err)
	}

	// Decode the base64 encoded data
	if attachmentData.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(attachmentData.Data)
		if err != nil {
			// Try standard base64 if URL-safe fails
			decoded, err = base64.StdEncoding.DecodeString(attachmentData.Data)
			if err != nil {
				return fmt.Errorf("failed to decode attachment data: %w", err)
			}
		}

		// Store the decoded data as base64 string for embedding in targets
		attachment.Data = base64.StdEncoding.EncodeToString(decoded)
		attachment.Size = int64(len(decoded))
	}

	return nil
}

// filterAttachments filters attachments based on configuration.
func (p *SimplifiedContentProcessor) filterAttachments(attachments []models.Attachment) []models.Attachment {
	if len(p.config.AttachmentTypes) == 0 {
		return attachments // No filtering
	}

	var filtered []models.Attachment

	for _, attachment := range attachments {
		if p.isAllowedAttachmentType(attachment) {
			filtered = append(filtered, attachment)
		}
	}

	return filtered
}

// isAllowedAttachmentType checks if an attachment type is allowed based on configuration.
func (p *SimplifiedContentProcessor) isAllowedAttachmentType(attachment models.Attachment) bool {
	// Extract extension from filename
	parts := strings.Split(attachment.Name, ".")
	if len(parts) < 2 {
		return false // No extension
	}

	extension := strings.ToLower(parts[len(parts)-1])

	for _, allowedType := range p.config.AttachmentTypes {
		if strings.ToLower(allowedType) == extension {
			return true
		}
	}

	return false
}
