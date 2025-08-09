package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

type ObsidianTarget struct {
	vaultPath        string
	templateDir      string
	dailyNotesFormat string
}

func NewObsidianTarget() *ObsidianTarget {
	return &ObsidianTarget{
		dailyNotesFormat: "2006-01-02", // Default: YYYY-MM-DD
	}
}

func (o *ObsidianTarget) Name() string {
	return "obsidian"
}

func (o *ObsidianTarget) Configure(config map[string]interface{}) error {
	if vaultPath, ok := config["vault_path"].(string); ok {
		o.vaultPath = vaultPath
	}
	if templateDir, ok := config["template_dir"].(string); ok {
		o.templateDir = templateDir
	}
	if format, ok := config["daily_notes_format"].(string); ok {
		o.dailyNotesFormat = format
	}
	return nil
}

func (o *ObsidianTarget) Export(items []*models.Item, outputDir string) error {
	for _, item := range items {
		if err := o.exportItem(item, outputDir); err != nil {
			return fmt.Errorf("failed to export item %s: %w", item.ID, err)
		}
	}
	return nil
}

func (o *ObsidianTarget) exportItem(item *models.Item, outputDir string) error {
	filename := o.FormatFilename(item.Title)
	filePath := filepath.Join(outputDir, filename)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	content := o.formatContent(item)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func (o *ObsidianTarget) formatContent(item *models.Item) string {
	var sb strings.Builder

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(o.FormatMetadata(item.Metadata))
	sb.WriteString(fmt.Sprintf("id: %s\n", item.ID))
	sb.WriteString(fmt.Sprintf("source: %s\n", item.SourceType))
	sb.WriteString(fmt.Sprintf("type: %s\n", item.ItemType))
	sb.WriteString(fmt.Sprintf("created: %s\n", item.CreatedAt.Format(time.RFC3339)))
	if len(item.Tags) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range item.Tags {
			sb.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}
	sb.WriteString("---\n\n")

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", item.Title))

	// Content
	if item.Content != "" {
		sb.WriteString(item.Content)
		sb.WriteString("\n\n")
	}

	// Attachments
	if len(item.Attachments) > 0 {
		sb.WriteString("## Attachments\n\n")
		for _, attachment := range item.Attachments {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", attachment.Name))
		}
		sb.WriteString("\n")
	}

	// Links
	if len(item.Links) > 0 {
		sb.WriteString("## Links\n\n")
		for _, link := range item.Links {
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", link.Title, link.URL))
		}
	}

	return sb.String()
}

func (o *ObsidianTarget) FormatFilename(title string) string {
	// Use sanitizeFilename logic similar to drive service
	filename := sanitizeFilename(title)
	return filename + ".md"
}

func (o *ObsidianTarget) GetFileExtension() string {
	return ".md"
}

func (o *ObsidianTarget) FormatMetadata(metadata map[string]interface{}) string {
	var sb strings.Builder
	for key, value := range metadata {
		if key == "attendees" {
			sb.WriteString(o.formatAttendees(value))
		} else {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
	}
	return sb.String()
}

// formatAttendees formats attendees as wikilink arrays for Obsidian
func (o *ObsidianTarget) formatAttendees(attendeesValue interface{}) string {
	var sb strings.Builder
	
	// Handle different types that attendees might be stored as
	switch attendees := attendeesValue.(type) {
	case []models.Attendee:
		if len(attendees) == 0 {
			return ""
		}
		sb.WriteString("attendees:\n")
		for _, attendee := range attendees {
			displayName := attendee.GetDisplayName()
			sb.WriteString(fmt.Sprintf("  - \"[[%s]]\"\n", displayName))
		}
	case []interface{}:
		// Handle case where attendees might be stored as generic interface slice
		if len(attendees) == 0 {
			return ""
		}
		sb.WriteString("attendees:\n")
		for _, attendee := range attendees {
			if attendeeMap, ok := attendee.(map[string]interface{}); ok {
				var displayName string
				if name, exists := attendeeMap["DisplayName"].(string); exists && name != "" {
					displayName = name
				} else if email, exists := attendeeMap["Email"].(string); exists {
					displayName = email
				} else {
					displayName = fmt.Sprintf("%v", attendee)
				}
				sb.WriteString(fmt.Sprintf("  - \"[[%s]]\"\n", displayName))
			} else {
				sb.WriteString(fmt.Sprintf("  - \"[[%v]]\"\n", attendee))
			}
		}
	default:
		// Fallback for other types
		sb.WriteString(fmt.Sprintf("attendees: %v\n", attendeesValue))
	}
	
	return sb.String()
}

// Preview generates a preview of what files would be created/modified without actually writing them
func (o *ObsidianTarget) Preview(items []*models.Item, outputDir string) ([]*interfaces.FilePreview, error) {
	var previews []*interfaces.FilePreview
	
	for _, item := range items {
		preview, err := o.previewItem(item, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to preview item %s: %w", item.ID, err)
		}
		previews = append(previews, preview)
	}
	
	return previews, nil
}

func (o *ObsidianTarget) previewItem(item *models.Item, outputDir string) (*interfaces.FilePreview, error) {
	filename := o.FormatFilename(item.Title)
	filePath := filepath.Join(outputDir, filename)
	
	// Generate content that would be written
	content := o.formatContent(item)
	
	// Check if file already exists
	var existingContent string
	var action string
	var conflict bool
	
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read current content
		if existingData, readErr := os.ReadFile(filePath); readErr == nil {
			existingContent = string(existingData)
			if existingContent == content {
				action = "skip"
			} else {
				action = "update"
				// Check for potential conflicts (basic check)
				conflict = o.detectConflict(existingContent, content)
			}
		} else {
			action = "update"
			existingContent = fmt.Sprintf("[Error reading file: %v]", readErr)
		}
	} else {
		// File doesn't exist
		action = "create"
	}
	
	return &interfaces.FilePreview{
		FilePath:        filePath,
		Action:          action,
		Content:         content,
		ExistingContent: existingContent,
		Conflict:        conflict,
	}, nil
}

// detectConflict performs basic conflict detection
func (o *ObsidianTarget) detectConflict(existing, new string) bool {
	// Simple conflict detection: check if existing file has been manually modified
	// by looking for content that wouldn't be generated by our sync process
	
	// If the existing file has different frontmatter structure, it might be manually edited
	existingLines := strings.Split(existing, "\n")
	newLines := strings.Split(new, "\n")
	
	// Check if frontmatter sections are very different
	existingFrontmatter := extractFrontmatter(existingLines)
	newFrontmatter := extractFrontmatter(newLines)
	
	// Basic heuristic: if existing has significantly more frontmatter fields, 
	// it might have been manually edited
	if len(existingFrontmatter) > len(newFrontmatter)+3 {
		return true
	}
	
	// Check for manual content additions (content after meeting details that we wouldn't generate)
	existingContent := extractContent(existingLines)
	newContent := extractContent(newLines)
	
	// If existing content is significantly longer, it might have manual additions
	if len(existingContent) > len(newContent)+100 {
		return true
	}
	
	return false
}

func extractFrontmatter(lines []string) []string {
	var frontmatter []string
	inFrontmatter := false
	frontmatterCount := 0
	
	for _, line := range lines {
		if line == "---" {
			frontmatterCount++
			if frontmatterCount == 2 {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			frontmatter = append(frontmatter, line)
		}
	}
	
	return frontmatter
}

func extractContent(lines []string) string {
	var content strings.Builder
	inFrontmatter := false
	frontmatterCount := 0
	
	for _, line := range lines {
		if line == "---" {
			frontmatterCount++
			if frontmatterCount == 2 {
				inFrontmatter = false
				continue
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter && frontmatterCount >= 2 {
			content.WriteString(line + "\n")
		}
	}
	
	return content.String()
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(filename string) string {
	replacements := map[string]string{
		"/":  "-",
		"\\": "-",
		":":  "-",
		"*":  "",
		"?":  "",
		"\"": "",
		"<":  "",
		">":  "",
		"|":  "-",
		" ":  "-",  // Replace spaces with hyphens
	}

	for old, new := range replacements {
		filename = strings.ReplaceAll(filename, old, new)
	}

	filename = strings.TrimSpace(filename)
	// Remove multiple consecutive hyphens
	for strings.Contains(filename, "--") {
		filename = strings.ReplaceAll(filename, "--", "-")
	}

	// Remove leading/trailing hyphens
	filename = strings.Trim(filename, "-")

	return filename
}