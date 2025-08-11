package logseq

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

type LogseqTarget struct {
	graphPath   string
	journalPath string
	pagesPath   string
}

func NewLogseqTarget() *LogseqTarget {
	return &LogseqTarget{}
}

func (l *LogseqTarget) Name() string {
	return "logseq"
}

func (l *LogseqTarget) Configure(config map[string]interface{}) error {
	if graphPath, ok := config["graph_path"].(string); ok {
		l.graphPath = graphPath
		l.journalPath = filepath.Join(graphPath, "journals")
		l.pagesPath = filepath.Join(graphPath, "pages")
	}
	return nil
}

func (l *LogseqTarget) Export(items []*models.Item, outputDir string) error {
	// Use flat structure - all files in outputDir
	for _, item := range items {
		if err := l.exportItem(item, outputDir); err != nil {
			return fmt.Errorf("failed to export item %s: %w", item.ID, err)
		}
	}
	return nil
}

func (l *LogseqTarget) exportItem(item *models.Item, outputDir string) error {
	filename := l.FormatFilename(item.Title)
	filePath := filepath.Join(outputDir, filename)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	content := l.formatContent(item)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func (l *LogseqTarget) formatContent(item *models.Item) string {
	var sb strings.Builder

	// Properties block (Logseq-specific)
	sb.WriteString("- id:: " + item.ID + "\n")
	sb.WriteString("- source:: " + item.SourceType + "\n")
	sb.WriteString("- type:: " + item.ItemType + "\n")
	sb.WriteString("- created:: [[" + item.CreatedAt.Format("Jan 2nd, 2006") + "]]\n")

	// Add custom metadata
	for key, value := range item.Metadata {
		sb.WriteString(fmt.Sprintf("- %s:: %v\n", key, value))
	}

	// Tags
	if len(item.Tags) > 0 {
		sb.WriteString("- tags:: ")
		for i, tag := range item.Tags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("#" + tag)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Title as heading
	sb.WriteString("# " + item.Title + "\n\n")

	// Content
	if item.Content != "" {
		sb.WriteString(item.Content)
		sb.WriteString("\n\n")
	}

	// Attachments as blocks
	if len(item.Attachments) > 0 {
		sb.WriteString("## Attachments\n")
		for _, attachment := range item.Attachments {
			sb.WriteString("- [[" + attachment.Name + "]]\n")
		}
		sb.WriteString("\n")
	}

	// Links as blocks
	if len(item.Links) > 0 {
		sb.WriteString("## Links\n")
		for _, link := range item.Links {
			sb.WriteString("- [" + link.Title + "](" + link.URL + ")\n")
		}
	}

	return sb.String()
}

func (l *LogseqTarget) FormatFilename(title string) string {
	// Logseq prefers page references format
	filename := sanitizeFilename(title)
	return filename + ".md"
}

func (l *LogseqTarget) GetFileExtension() string {
	return ".md"
}

func (l *LogseqTarget) FormatMetadata(metadata map[string]interface{}) string {
	var sb strings.Builder
	for key, value := range metadata {
		sb.WriteString(fmt.Sprintf("- %s:: %v\n", key, value))
	}
	return sb.String()
}

// Preview generates a preview of what files would be created/modified without actually writing them
func (l *LogseqTarget) Preview(items []*models.Item, outputDir string) ([]*interfaces.FilePreview, error) {
	var previews []*interfaces.FilePreview

	for _, item := range items {
		preview, err := l.previewItem(item, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to preview item %s: %w", item.ID, err)
		}
		previews = append(previews, preview)
	}

	return previews, nil
}

func (l *LogseqTarget) previewItem(item *models.Item, outputDir string) (*interfaces.FilePreview, error) {
	filename := l.FormatFilename(item.Title)
	filePath := filepath.Join(outputDir, filename)

	// Generate content that would be written
	content := l.formatContent(item)

	// Check if file already exists
	var existingContent string
	var action string

	if _, err := os.Stat(filePath); err == nil {
		// File exists, read current content
		if existingData, readErr := os.ReadFile(filePath); readErr == nil {
			existingContent = string(existingData)
			if existingContent == content {
				action = "skip"
			} else {
				action = "update"
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
		Conflict:        false, // Simplified for Logseq
	}, nil
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
	}

	for old, new := range replacements {
		filename = strings.ReplaceAll(filename, old, new)
	}

	filename = strings.TrimSpace(filename)
	for strings.Contains(filename, "  ") {
		filename = strings.ReplaceAll(filename, "  ", " ")
	}

	return filename
}
