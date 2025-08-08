package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}
	return sb.String()
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