package transform

import (
	"regexp"
	"strings"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

// LinkExtractionTransformer extracts URLs from content and populates the Links field.
// Extracted from Gmail's ContentProcessor.ExtractLinks to be universally available.
type LinkExtractionTransformer struct {
	config map[string]interface{}

	// Pre-compiled regular expressions for performance
	urlRegex          *regexp.Regexp
	markdownLinkRegex *regexp.Regexp
}

func NewLinkExtractionTransformer() *LinkExtractionTransformer {
	return &LinkExtractionTransformer{
		config:            make(map[string]interface{}),
		urlRegex:          regexp.MustCompile(`https?://[^\s<>"\\]+`),
		markdownLinkRegex: regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`),
	}
}

func (t *LinkExtractionTransformer) Name() string {
	return "link_extraction"
}

func (t *LinkExtractionTransformer) Configure(config map[string]interface{}) error {
	t.config = config

	return nil
}

func (t *LinkExtractionTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	transformedItems := make([]*models.Item, len(items))

	for i, item := range items {
		extractedLinks := t.ExtractLinks(item.Content)

		// Check if we found any new links
		if len(extractedLinks) > 0 || t.shouldAlwaysProcessLinks() {
			// Copy-on-write: only copy if there are links to add or we need to merge
			transformedItem := *item

			// Merge with existing links if any
			if len(item.Links) > 0 {
				transformedItem.Links = t.mergeLinks(item.Links, extractedLinks)
			} else {
				transformedItem.Links = extractedLinks
			}

			transformedItems[i] = &transformedItem
		} else {
			// No new links found, keep original
			transformedItems[i] = item
		}
	}

	return transformedItems, nil
}

// ExtractLinks extracts URLs from content with enhanced detection.
// Extracted from Gmail's ContentProcessor.ExtractLinks.
func (t *LinkExtractionTransformer) ExtractLinks(content string) []models.Link {
	// Collect all URL matches with their positions to maintain order
	type urlMatch struct {
		url   string
		title string
		pos   int
	}

	allMatches := make([]urlMatch, 0)
	seenURL := make(map[string]bool)

	// Find markdown URLs first to prioritize them if enabled
	if t.shouldExtractMarkdownLinks() {
		markdownMatches := t.markdownLinkRegex.FindAllStringSubmatchIndex(content, -1)
		for _, match := range markdownMatches {
			if len(match) >= 6 {
				title := content[match[2]:match[3]]
				url := content[match[4]:match[5]]
				url = strings.TrimLeft(strings.TrimRight(url, ".,!?;:)"), "(")

				if !seenURL[url] {
					allMatches = append(allMatches, urlMatch{
						url:   url,
						title: title,
						pos:   match[0],
					})
					seenURL[url] = true
				}
			}
		}
	}

	// Find standalone URLs and add them if they haven't been seen in markdown links
	if t.shouldExtractPlainURLs() {
		urlMatches := t.urlRegex.FindAllStringIndex(content, -1)
		markdownMatches := t.markdownLinkRegex.FindAllStringSubmatchIndex(content, -1)

		for _, match := range urlMatches {
			url := content[match[0]:match[1]]
			url = strings.TrimLeft(strings.TrimRight(url, ".,!?;:)"), "(")

			// Check if this match is inside a markdown link
			isInsideMarkdown := false

			for _, mdMatch := range markdownMatches {
				if match[0] >= mdMatch[0] && match[1] <= mdMatch[1] {
					isInsideMarkdown = true

					break
				}
			}

			if !isInsideMarkdown && !seenURL[url] {
				allMatches = append(allMatches, urlMatch{
					url:   url,
					title: "",
					pos:   match[0],
				})
				seenURL[url] = true
			}
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

	// Convert to Link objects
	links := make([]models.Link, 0, len(allMatches))
	for _, match := range allMatches {
		linkType := "external"

		// Determine link type based on URL
		if t.isInternalLink(match.url) {
			linkType = "internal"
		} else if t.isDocumentLink(match.url) {
			linkType = "document"
		}

		links = append(links, models.Link{
			URL:   match.url,
			Title: match.title,
			Type:  linkType,
		})
	}

	// Deduplicate if enabled
	if t.shouldDeduplicateLinks() {
		links = t.deduplicateLinks(links)
	}

	return links
}

// mergeLinks combines existing links with newly extracted links.
func (t *LinkExtractionTransformer) mergeLinks(existing []models.Link, extracted []models.Link) []models.Link {
	// Create a map of existing URLs for fast lookup
	existingURLs := make(map[string]bool)
	for _, link := range existing {
		existingURLs[link.URL] = true
	}

	// Start with existing links
	merged := make([]models.Link, 0, len(existing)+len(extracted))
	merged = append(merged, existing...)

	// Add extracted links that don't already exist
	for _, link := range extracted {
		if !existingURLs[link.URL] {
			merged = append(merged, link)
		}
	}

	return merged
}

// deduplicateLinks removes duplicate URLs while preserving order.
func (t *LinkExtractionTransformer) deduplicateLinks(links []models.Link) []models.Link {
	seen := make(map[string]bool)
	deduplicated := make([]models.Link, 0, len(links))

	for _, link := range links {
		if !seen[link.URL] {
			deduplicated = append(deduplicated, link)
			seen[link.URL] = true
		}
	}

	return deduplicated
}

// isInternalLink checks if a URL appears to be an internal/relative link.
func (t *LinkExtractionTransformer) isInternalLink(url string) bool {
	// Simple heuristics for internal links
	return strings.HasPrefix(url, "/") ||
		strings.HasPrefix(url, "#") ||
		strings.HasPrefix(url, "./") ||
		strings.HasPrefix(url, "../")
}

// isDocumentLink checks if a URL points to a document.
func (t *LinkExtractionTransformer) isDocumentLink(url string) bool {
	// Common document extensions
	documentExts := []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".txt", ".md", ".jpg", ".png", ".gif",
	}

	lowerURL := strings.ToLower(url)
	for _, ext := range documentExts {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}

	// Check for common document hosting domains
	documentDomains := []string{"docs.google.com", "drive.google.com", "dropbox.com", "onedrive.com"}
	for _, domain := range documentDomains {
		if strings.Contains(lowerURL, domain) {
			return true
		}
	}

	return false
}

// Configuration helper methods

func (t *LinkExtractionTransformer) shouldExtractMarkdownLinks() bool {
	if val, exists := t.config["extract_markdown_links"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *LinkExtractionTransformer) shouldExtractPlainURLs() bool {
	if val, exists := t.config["extract_plain_urls"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *LinkExtractionTransformer) shouldDeduplicateLinks() bool {
	if val, exists := t.config["deduplicate_links"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return true // Default: enabled
}

func (t *LinkExtractionTransformer) shouldAlwaysProcessLinks() bool {
	if val, exists := t.config["always_process"]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return false // Default: only process when links found
}

// Ensure interface compliance.
var _ interfaces.Transformer = (*LinkExtractionTransformer)(nil)
