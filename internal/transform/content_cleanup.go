package transform

import (
	"html"
	"regexp"
	"strings"

	nethtml "golang.org/x/net/html"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

const (
	transformerNameContentCleanup = "content_cleanup"
	htmlTagTh                     = "th"
	htmlTagTd                     = "td"
)

// ContentCleanupTransformer provides HTML→Markdown conversion and content cleanup.
// Extracted from Gmail's ContentProcessor to be universally available.
type ContentCleanupTransformer struct {
	config map[string]interface{}

	// Pre-compiled regular expressions for performance
	whitespaceCleanupRegex *regexp.Regexp
	consecutiveAsterisks   *regexp.Regexp
}

func NewContentCleanupTransformer() *ContentCleanupTransformer {
	return &ContentCleanupTransformer{
		config:                 make(map[string]interface{}),
		whitespaceCleanupRegex: regexp.MustCompile(`\n\s*\n\s*\n`),
		consecutiveAsterisks:   regexp.MustCompile(`\*{4,}`),
	}
}

func (t *ContentCleanupTransformer) Name() string {
	return transformerNameContentCleanup
}

func (t *ContentCleanupTransformer) Configure(config map[string]interface{}) error {
	t.config = config

	return nil
}

func (t *ContentCleanupTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	transformedItems := make([]*models.Item, len(items))

	for i, item := range items {
		transformed := false
		newItem := *item // Copy the item

		// Process content based on configuration
		if t.shouldProcessHTMLToMarkdown() && t.containsHTML(item.Content) {
			cleanedContent := t.ProcessHTMLContent(item.Content)
			if cleanedContent != item.Content {
				newItem.Content = cleanedContent
				transformed = true
			}
		}

		// Apply content cleanup
		if t.shouldRemoveExtraWhitespace() {
			cleanedContent := t.cleanupWhitespace(newItem.Content)
			if cleanedContent != newItem.Content {
				newItem.Content = cleanedContent
				transformed = true
			}
		}

		// Strip quoted text if enabled
		if t.shouldStripQuotedText() {
			cleanedContent := t.StripQuotedText(newItem.Content)
			if cleanedContent != newItem.Content {
				newItem.Content = cleanedContent
				transformed = true
			}
		}

		// Clean up title
		cleanedTitle := t.cleanupTitle(newItem.Title)
		if cleanedTitle != newItem.Title {
			newItem.Title = cleanedTitle
			transformed = true
		}

		if transformed {
			transformedItems[i] = &newItem
		} else {
			transformedItems[i] = item
		}
	}

	return transformedItems, nil
}

// ProcessHTMLContent converts HTML to markdown using proper HTML parsing.
// Extracted from Gmail's ContentProcessor.ProcessHTMLContent.
func (t *ContentCleanupTransformer) ProcessHTMLContent(htmlContent string) string {
	doc, err := nethtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback to the input if parsing fails
		return html.UnescapeString(htmlContent)
	}

	var markdown strings.Builder
	t.convertNodeToMarkdown(doc, &markdown)
	result := markdown.String()

	// Apply additional entity processing for any that weren't handled by the parser
	result = t.unescapeHTMLEntities(result)

	// Clean up whitespace and formatting issues
	result = t.whitespaceCleanupRegex.ReplaceAllString(result, "\n\n")

	// Fix consecutive asterisks that can occur from malformed HTML
	result = t.consecutiveAsterisks.ReplaceAllString(result, "***")

	return strings.TrimSpace(result)
}

// StripQuotedText removes quoted text from email content with enhanced detection.
// Extracted from Gmail's ContentProcessor.StripQuotedText.
func (t *ContentCleanupTransformer) StripQuotedText(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

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

// convertNodeToMarkdown recursively converts HTML nodes to markdown.
// Extracted from Gmail's ContentProcessor.convertNodeToMarkdown.
func (t *ContentCleanupTransformer) convertNodeToMarkdown(n *nethtml.Node, markdown *strings.Builder) {
	switch n.Type {
	case nethtml.TextNode:
		text := t.unescapeHTMLEntities(n.Data)
		markdown.WriteString(text)

	case nethtml.ElementNode:
		switch n.Data {
		case "h1":
			markdown.WriteString("# ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h2":
			markdown.WriteString("## ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h3":
			markdown.WriteString("### ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h4":
			markdown.WriteString("#### ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h5":
			markdown.WriteString("##### ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "h6":
			markdown.WriteString("###### ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "p":
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n\n")
		case "br":
			markdown.WriteString("\n")
		case "strong", "b":
			markdown.WriteString("**")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("**")
		case "em", "i":
			markdown.WriteString("*")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("*")
		case "code":
			markdown.WriteString("`")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("`")
		case "pre":
			markdown.WriteString("```\n")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n```\n")
		case "blockquote":
			// Process blockquote content and add > prefix to each line
			var blockquoteContent strings.Builder
			t.convertChildNodes(n, &blockquoteContent)

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
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "ol":
			markdown.WriteString("\n")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "li":
			markdown.WriteString("- ")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "a":
			href := t.getAttributeValue(n, "href")
			if href != "" {
				markdown.WriteString("[")
				t.convertChildNodes(n, markdown)
				markdown.WriteString("](")
				markdown.WriteString(href)
				markdown.WriteString(")")
			} else {
				t.convertChildNodes(n, markdown)
			}
		case "img":
			src := t.getAttributeValue(n, "src")
			alt := t.getAttributeValue(n, "alt")

			if src != "" {
				markdown.WriteString("![")
				markdown.WriteString(alt)
				markdown.WriteString("](")
				markdown.WriteString(src)
				markdown.WriteString(")")
			}
		case "div":
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "table":
			markdown.WriteString("\n")
			t.convertChildNodes(n, markdown)
			markdown.WriteString("\n")
		case "tr":
			t.convertTableRow(n, markdown)
		case htmlTagTd, htmlTagTh:
			t.convertChildNodes(n, markdown)
		case "style", "script":
			// Skip style and script tags completely
			return
		default:
			// For other elements, just process children
			t.convertChildNodes(n, markdown)
		}

	default:
		// For document and other node types, process children
		t.convertChildNodes(n, markdown)
	}
}

// convertChildNodes processes all child nodes.
func (t *ContentCleanupTransformer) convertChildNodes(n *nethtml.Node, markdown *strings.Builder) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		t.convertNodeToMarkdown(child, markdown)
	}
}

// convertTableRow processes a table row with proper cell separation.
func (t *ContentCleanupTransformer) convertTableRow(n *nethtml.Node, markdown *strings.Builder) {
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
		t.convertChildNodes(cell, markdown)

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

// getAttributeValue gets the value of an HTML attribute.
func (t *ContentCleanupTransformer) getAttributeValue(n *nethtml.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

// unescapeHTMLEntities handles HTML entities including common ones like &hellip;, &ldquo;, etc.
// Extracted from Gmail's ContentProcessor.unescapeHTMLEntities.
func (t *ContentCleanupTransformer) unescapeHTMLEntities(text string) string {
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

// cleanupWhitespace removes excessive whitespace.
func (t *ContentCleanupTransformer) cleanupWhitespace(content string) string {
	content = strings.TrimSpace(content)

	// Replace multiple newlines with double newlines
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	// Remove carriage returns
	content = strings.ReplaceAll(content, "\r", "")

	return content
}

// cleanupTitle removes common email prefixes and cleans up title.
func (t *ContentCleanupTransformer) cleanupTitle(title string) string {
	title = strings.TrimSpace(title)

	// Remove common prefixes iteratively to handle multiple prefixes
	prefixes := []string{"Re:", "RE:", "Fwd:", "FWD:", "Fw:", "FW:"}
	maxIterations := 10 // Prevent infinite loops
	iterations := 0

	for iterations < maxIterations {
		original := title
		for _, prefix := range prefixes {
			if strings.HasPrefix(title, prefix) {
				title = strings.TrimSpace(title[len(prefix):])
			}
		}
		// If no change was made, we're done
		if title == original {
			break
		}

		iterations++
	}

	return title
}

// Configuration helper methods

func (t *ContentCleanupTransformer) shouldProcessHTMLToMarkdown() bool {
	if val, exists := t.config["html_to_markdown"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *ContentCleanupTransformer) shouldStripQuotedText() bool {
	if val, exists := t.config["strip_quoted_text"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *ContentCleanupTransformer) shouldRemoveExtraWhitespace() bool {
	if val, exists := t.config["remove_extra_whitespace"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *ContentCleanupTransformer) shouldPreserveFormatting() bool {
	if val, exists := t.config["preserve_formatting"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

// containsHTML checks if content appears to contain HTML.
func (t *ContentCleanupTransformer) containsHTML(content string) bool {
	return strings.Contains(content, "<") && strings.Contains(content, ">")
}

// Ensure interface compliance.
var _ interfaces.Transformer = (*ContentCleanupTransformer)(nil)
