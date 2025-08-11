package gmail

import (
	"encoding/base64"
	"fmt"
	"html"
	"log/slog"
	"regexp"
	"strings"

	nethtml "golang.org/x/net/html"
	"google.golang.org/api/gmail/v1"

	"pkm-sync/pkg/models"
)

// Pre-compiled regular expressions for performance
var (
	whitespaceCleanupRegex = regexp.MustCompile(`\n\s*\n\s*\n`)
	urlRegex               = regexp.MustCompile(`https?://[^\s<>")\]]+`)
	markdownLinkRegex      = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

	// Signature patterns compiled once
	signatureRegexPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^Best regards?,?`),
		regexp.MustCompile(`^Sincerely,?`),
		regexp.MustCompile(`^Thanks?,?`),
		regexp.MustCompile(`^Cheers,?`),
		regexp.MustCompile(`^Sent from my`),
		regexp.MustCompile(`^Get Outlook for`),
		regexp.MustCompile(`@\w+\.\w+`),                     // Email address
		regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`), // Phone number
		regexp.MustCompile(`^[A-Z][a-z]+ [A-Z][a-z]+$`),     // Name pattern (two capitalized words)
	}
)

// ContentProcessor handles advanced email content processing
type ContentProcessor struct {
	config  models.GmailSourceConfig
	service *Service
}

// NewContentProcessor creates a new content processor with the given configuration
func NewContentProcessor(config models.GmailSourceConfig) *ContentProcessor {
	return &ContentProcessor{
		config: config,
	}
}

// NewContentProcessorWithService creates a new content processor with service for attachment downloads
func NewContentProcessorWithService(config models.GmailSourceConfig, service *Service) *ContentProcessor {
	return &ContentProcessor{
		config:  config,
		service: service,
	}
}

// ProcessEmailBody extracts and processes the email body based on configuration
func (p *ContentProcessor) ProcessEmailBody(msg *gmail.Message) (string, error) {
	if msg.Payload == nil {
		return "", nil
	}

	// Try to get HTML content first, then plain text
	htmlContent := p.extractBodyPart(msg.Payload, "text/html")
	textContent := p.extractBodyPart(msg.Payload, "text/plain")

	var content string

	// Prefer HTML if available and processing is enabled
	if p.config.ProcessHTMLContent && htmlContent != "" {
		// Convert HTML to markdown
		content = p.ProcessHTMLContent(htmlContent)
	} else if textContent != "" {
		content = textContent
	} else if htmlContent != "" {
		// Use HTML as-is if no processing enabled
		content = htmlContent
	} else {
		// Fallback to snippet
		content = msg.Snippet
	}

	// Apply additional processing
	if p.config.StripQuotedText {
		content = p.StripQuotedText(content)
	}

	// Extract signatures if enabled
	if p.config.ExtractSignatures {
		content = p.ExtractSignatures(content)
	}

	return content, nil
}

// extractBodyPart recursively extracts body content of specified MIME type
func (p *ContentProcessor) extractBodyPart(part *gmail.MessagePart, mimeType string) string {
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

// ProcessHTMLContent converts HTML to markdown using proper HTML parsing
func (p *ContentProcessor) ProcessHTMLContent(htmlContent string) string {
	doc, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback to the input if parsing fails
		return html.UnescapeString(htmlContent)
	}

	var markdown strings.Builder
	p.convertNodeToMarkdown(doc, &markdown)

	result := markdown.String()

	// Apply additional entity processing for any that weren't handled by the parser
	result = p.unescapeHTMLEntities(result)

	// Clean up whitespace and formatting issues
	result = whitespaceCleanupRegex.ReplaceAllString(result, "\n\n")

	// Fix consecutive asterisks that can occur from malformed HTML
	consecutiveAsterisks := regexp.MustCompile(`\*{4,}`)
	result = consecutiveAsterisks.ReplaceAllString(result, "***")

	result = strings.TrimSpace(result)

	return result
}

// convertNodeToMarkdown recursively converts HTML nodes to markdown
func (p *ContentProcessor) convertNodeToMarkdown(n *nethtml.Node, markdown *strings.Builder) {
	switch n.Type {
	case nethtml.TextNode:
		text := p.unescapeHTMLEntities(n.Data)
		markdown.WriteString(text)

	case nethtml.ElementNode:
		switch n.Data {
		case "h1":
			markdown.WriteString("# ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h2":
			markdown.WriteString("## ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h3":
			markdown.WriteString("### ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h4":
			markdown.WriteString("#### ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h5":
			markdown.WriteString("##### ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h6":
			markdown.WriteString("###### ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "p":
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n\n")
		case "br":
			markdown.WriteString("\n")
		case "strong", "b":
			markdown.WriteString("**")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("**")
		case "em", "i":
			markdown.WriteString("*")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("*")
		case "code":
			markdown.WriteString("`")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("`")
		case "pre":
			markdown.WriteString("```\n")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n```\n")
		case "blockquote":
			// Process blockquote content and add > prefix to each line
			var blockquoteContent strings.Builder
			p.convertChildNodes(n, &blockquoteContent)
			content := strings.TrimSpace(blockquoteContent.String())
			if content != "" {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						markdown.WriteString("> ")
						markdown.WriteString(line)
						markdown.WriteString("\n")
					}
				}
			}
		case "ul":
			markdown.WriteString("\n")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "ol":
			markdown.WriteString("\n")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "li":
			markdown.WriteString("- ")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "a":
			href := p.getAttributeValue(n, "href")
			if href != "" {
				markdown.WriteString("[")
				p.convertChildNodes(n, markdown)
				markdown.WriteString("](")
				markdown.WriteString(href)
				markdown.WriteString(")")
			} else {
				p.convertChildNodes(n, markdown)
			}
		case "img":
			src := p.getAttributeValue(n, "src")
			alt := p.getAttributeValue(n, "alt")
			if src != "" {
				markdown.WriteString("![")
				markdown.WriteString(alt)
				markdown.WriteString("](")
				markdown.WriteString(src)
				markdown.WriteString(")")
			}
		case "div":
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "table":
			markdown.WriteString("\n")
			p.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "tr":
			p.convertTableRow(n, markdown)
		case "td", "th":
			p.convertChildNodes(n, markdown)
		case "style", "script":
			// Skip style and script tags completely
			return
		default:
			// For other elements, just process children
			p.convertChildNodes(n, markdown)
		}

	default:
		// For document and other node types, process children
		p.convertChildNodes(n, markdown)
	}
}

// convertChildNodes processes all child nodes
func (p *ContentProcessor) convertChildNodes(n *nethtml.Node, markdown *strings.Builder) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		p.convertNodeToMarkdown(child, markdown)
	}
}

// convertTableRow processes a table row with proper cell separation
func (p *ContentProcessor) convertTableRow(n *nethtml.Node, markdown *strings.Builder) {
	markdown.WriteString("| ")

	// Count cells first
	var cells []*nethtml.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == nethtml.ElementNode && (child.Data == "td" || child.Data == "th") {
			cells = append(cells, child)
		}
	}

	// Process each cell
	for i, cell := range cells {
		p.convertChildNodes(cell, markdown)
		if i < len(cells)-1 {
			markdown.WriteString(" | ")
		} else {
			// Check if this row has header cells - if so, add trailing space
			hasHeaders := false
			for _, c := range cells {
				if c.Data == "th" {
					hasHeaders = true
					break
				}
			}
			if hasHeaders {
				markdown.WriteString(" | ")
			} else {
				markdown.WriteString(" |")
			}
		}
	}

	markdown.WriteString("\n")
}

// getAttributeValue gets the value of an HTML attribute
func (p *ContentProcessor) getAttributeValue(n *nethtml.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// unescapeHTMLEntities handles HTML entities including common ones like &hellip;, &ldquo;, etc.
func (p *ContentProcessor) unescapeHTMLEntities(text string) string {
	// First apply the standard html.UnescapeString
	text = html.UnescapeString(text)

	// Handle additional entities and Unicode characters that html.UnescapeString might not handle
	replacer := strings.NewReplacer(
		"&hellip;", "...",
		"&ldquo;", "\"",
		"&rdquo;", "\"",
		"&mdash;", "—",
		"&ndash;", "–",
		"&nbsp;", " ",
		"\u00a0", " ", // non-breaking space
		"&rsquo;", "'",
		"&lsquo;", "'",
		"&quot;", "\"",
		// Unicode characters that might come from HTML parsing
		"\u201c", "\"", // left double quotation mark
		"\u201d", "\"", // right double quotation mark
		"\u2018", "'", // left single quotation mark
		"\u2019", "'", // right single quotation mark
		"\u2026", "...", // horizontal ellipsis
		"\u2014", "—", // em dash
		"\u2013", "–", // en dash
	)

	return replacer.Replace(text)
}

// StripQuotedText removes quoted text from email content with enhanced detection
func (p *ContentProcessor) StripQuotedText(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip lines that start with common quote indicators
		if strings.HasPrefix(trimmed, ">") {
			break // Stop processing at first quoted line
		}

		// Check for "On [date] [person] wrote:" patterns
		if strings.HasPrefix(trimmed, "On ") && strings.Contains(trimmed, " wrote:") {
			break
		}

		// Check for "From: [email]" patterns (often indicates forwarded content)
		if strings.HasPrefix(trimmed, "From: ") && strings.Contains(trimmed, "@") {
			break
		}

		// Check for "-----Original Message-----" patterns
		if strings.Contains(trimmed, "Original Message") || strings.Contains(trimmed, "original message") {
			break
		}

		// Check for forwarding indicators
		if strings.HasPrefix(trimmed, "---------- Forwarded message") {
			break
		}

		// Check for signature separators
		if trimmed == "--" || strings.HasPrefix(trimmed, "-- ") {
			// This might be a signature, check if this is near the end
			remainingLines := len(lines) - i
			if remainingLines <= 10 { // Likely a signature if less than 10 lines remain
				break
			}
		}

		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// ExtractSignatures extracts email signatures from content
func (p *ContentProcessor) ExtractSignatures(content string) string {
	lines := strings.Split(content, "\n")
	var contentLines []string
	var signatureLines []string
	var inSignature bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Common signature indicators
		if trimmed == "--" || strings.HasPrefix(trimmed, "-- ") {
			inSignature = true
			signatureLines = append(signatureLines, line)
			continue
		}

		// Look for patterns that might indicate signatures
		if !inSignature {
			// Check if we're near the end and this looks like signature content
			remainingLines := len(lines) - i
			if remainingLines <= 8 {
				// Safely call looksLikeSignature with nil check
				if p != nil && p.looksLikeSignature(trimmed) {
					inSignature = true
				}
			}
		}

		if inSignature {
			signatureLines = append(signatureLines, line)
		} else {
			contentLines = append(contentLines, line)
		}
	}

	// Return just the content without signatures
	return strings.TrimSpace(strings.Join(contentLines, "\n"))
}

// looksLikeSignature checks if a line looks like it could be part of a signature
func (p *ContentProcessor) looksLikeSignature(line string) bool {
	for _, pattern := range signatureRegexPatterns {
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}

// ExtractLinks extracts URLs from email content with enhanced detection
func (p *ContentProcessor) ExtractLinks(content string) []models.Link {
	// Collect all URL matches with their positions to maintain order
	type urlMatch struct {
		url   string
		title string
		pos   int
	}

	var allMatches []urlMatch

	// Find standalone URLs with positions
	urlMatches := urlRegex.FindAllStringIndex(content, -1)
	for _, match := range urlMatches {
		url := content[match[0]:match[1]]
		url = strings.TrimRight(url, ".,!?;:")
		allMatches = append(allMatches, urlMatch{
			url:   url,
			title: "",
			pos:   match[0],
		})
	}

	// Find markdown URLs with positions and titles
	markdownMatches := markdownLinkRegex.FindAllStringSubmatchIndex(content, -1)
	for _, match := range markdownMatches {
		if len(match) >= 6 {
			title := content[match[2]:match[3]]
			url := content[match[4]:match[5]]
			url = strings.TrimRight(url, ".,!?;:")

			allMatches = append(allMatches, urlMatch{
				url:   url,
				title: title,
				pos:   match[0],
			})
		}
	}

	// Sort by position to maintain order of appearance
	for i := 0; i < len(allMatches); i++ {
		for j := i + 1; j < len(allMatches); j++ {
			if allMatches[i].pos > allMatches[j].pos {
				allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
			}
		}
	}

	// Convert to Link objects and deduplicate
	var links []models.Link
	seen := make(map[string]bool)

	for _, match := range allMatches {
		if !seen[match.url] {
			seen[match.url] = true
			links = append(links, models.Link{
				URL:   match.url,
				Title: match.title,
				Type:  "external",
			})
		}
	}

	return links
}

// extractUrlTitle attempts to extract a meaningful title from a URL
func (p *ContentProcessor) extractUrlTitle(url string) string {
	// For now, just return empty - could be enhanced to fetch actual page titles
	// or extract meaningful parts from the URL structure
	return ""
}

// deduplicateLinks removes duplicate links from the slice
func (p *ContentProcessor) deduplicateLinks(links []models.Link) []models.Link {
	seen := make(map[string]bool)
	var result []models.Link

	for _, link := range links {
		if !seen[link.URL] {
			seen[link.URL] = true
			result = append(result, link)
		}
	}

	return result
}

// ProcessEmailAttachments processes email attachments based on configuration
func (p *ContentProcessor) ProcessEmailAttachments(msg *gmail.Message) []models.Attachment {
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

// extractAttachmentsFromPart recursively extracts attachments from message parts
func (p *ContentProcessor) extractAttachmentsFromPart(part *gmail.MessagePart, messageID string, attachments *[]models.Attachment) {
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

// fetchAttachmentData fetches the actual attachment data from Gmail API
func (p *ContentProcessor) fetchAttachmentData(messageID string, attachment *models.Attachment) error {
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

// filterAttachments filters attachments based on configuration
func (p *ContentProcessor) filterAttachments(attachments []models.Attachment) []models.Attachment {
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

// isAllowedAttachmentType checks if an attachment type is allowed based on configuration
func (p *ContentProcessor) isAllowedAttachmentType(attachment models.Attachment) bool {
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
