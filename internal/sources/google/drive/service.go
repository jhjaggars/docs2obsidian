package drive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pkm-sync/pkg/models"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Service struct {
	client *drive.Service
}

func NewService(httpClient *http.Client) (*Service, error) {
	driveService, err := drive.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %w", err)
	}

	return &Service{client: driveService}, nil
}

// GetFileMetadata retrieves metadata for a Google Drive file.
func (s *Service) GetFileMetadata(fileID string) (*models.DriveFile, error) {
	file, err := s.client.Files.Get(fileID).Fields("id,name,mimeType,webViewLink,modifiedTime,owners").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve file metadata: %w", err)
	}

	driveFile := &models.DriveFile{
		ID:          file.Id,
		Name:        file.Name,
		MimeType:    file.MimeType,
		WebViewLink: file.WebViewLink,
	}

	for _, owner := range file.Owners {
		driveFile.Owners = append(driveFile.Owners, owner.DisplayName)
	}

	driveFile.Shared = len(file.Owners) > 1

	return driveFile, nil
}

// IsGoogleDoc checks if a file is a Google Doc that can be exported to markdown.
func (s *Service) IsGoogleDoc(mimeType string) bool {
	return mimeType == "application/vnd.google-apps.document"
}

// ExportDocAsMarkdown exports a Google Doc as markdown format.
func (s *Service) ExportDocAsMarkdown(fileID string, outputPath string) error {
	if !s.IsGoogleDocByID(fileID) {
		return fmt.Errorf("file %s is not a Google Doc", fileID)
	}

	// Export as plain text first (closest to markdown)
	resp, err := s.client.Files.Export(fileID, "text/plain").Download()
	if err != nil {
		return fmt.Errorf("unable to export document: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("unable to create output directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}

	defer func() {
		_ = outFile.Close()
	}()

	// Copy content to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to write file content: %w", err)
	}

	return nil
}

// IsGoogleDocByID checks if a file ID represents a Google Doc.
func (s *Service) IsGoogleDocByID(fileID string) bool {
	file, err := s.client.Files.Get(fileID).Fields("mimeType").Do()
	if err != nil {
		return false
	}

	return s.IsGoogleDoc(file.MimeType)
}

// GetAttachmentsFromEvent extracts Google Drive file attachments from a calendar event.
func (s *Service) GetAttachmentsFromEvent(eventDescription string) ([]string, error) {
	var fileIDs []string

	// Look for Google Drive links in the event description
	// Google Drive links typically follow patterns like:
	// https://docs.google.com/document/d/FILE_ID/edit
	// https://drive.google.com/file/d/FILE_ID/view

	lines := strings.Split(eventDescription, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "docs.google.com/document/d/") {
			// Extract file ID from Google Docs URL
			if fileID := extractFileIDFromDocsURL(line); fileID != "" {
				fileIDs = append(fileIDs, fileID)
			}
		} else if strings.Contains(line, "drive.google.com/file/d/") {
			// Extract file ID from Google Drive URL
			if fileID := extractFileIDFromDriveURL(line); fileID != "" {
				fileIDs = append(fileIDs, fileID)
			}
		}
	}

	return fileIDs, nil
}

// extractFileIDFromDocsURL extracts file ID from Google Docs URL.
func extractFileIDFromDocsURL(url string) string {
	// Pattern: https://docs.google.com/document/d/FILE_ID/edit
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "d" && i+1 < len(parts) {
			fileID := parts[i+1]
			// Remove any query parameters
			if idx := strings.Index(fileID, "?"); idx != -1 {
				fileID = fileID[:idx]
			}

			return fileID
		}
	}

	return ""
}

// extractFileIDFromDriveURL extracts file ID from Google Drive URL.
func extractFileIDFromDriveURL(url string) string {
	// Pattern: https://drive.google.com/file/d/FILE_ID/view
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "d" && i+1 < len(parts) {
			fileID := parts[i+1]
			// Remove any query parameters
			if idx := strings.Index(fileID, "?"); idx != -1 {
				fileID = fileID[:idx]
			}

			return fileID
		}
	}

	return ""
}

// ExportAttachedDocsFromEvent exports all Google Docs attached to an event.
func (s *Service) ExportAttachedDocsFromEvent(eventDescription, outputDir string) ([]string, error) {
	fileIDs, err := s.GetAttachmentsFromEvent(eventDescription)
	if err != nil {
		return nil, err
	}

	exportedFiles := make([]string, 0, len(fileIDs))

	for _, fileID := range fileIDs {
		// Get file metadata to determine name and type
		metadata, err := s.GetFileMetadata(fileID)
		if err != nil {
			fmt.Printf("Warning: Could not get metadata for file %s: %v\n", fileID, err)

			continue
		}

		// Only export Google Docs
		if !s.IsGoogleDoc(metadata.MimeType) {
			fmt.Printf("Skipping %s: not a Google Doc (type: %s)\n", metadata.Name, metadata.MimeType)

			continue
		}

		// Generate output filename
		filename := sanitizeFilename(metadata.Name)
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}

		outputPath := filepath.Join(outputDir, filename)

		// Export the document
		if err := s.ExportDocAsMarkdown(fileID, outputPath); err != nil {
			fmt.Printf("Warning: Could not export %s: %v\n", metadata.Name, err)

			continue
		}

		exportedFiles = append(exportedFiles, outputPath)
		fmt.Printf("Exported: %s -> %s\n", metadata.Name, outputPath)
	}

	return exportedFiles, nil
}

// sanitizeFilename removes or replaces characters that are invalid in filenames.
func sanitizeFilename(filename string) string {
	// Replace common problematic characters
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

	// Remove multiple consecutive spaces and trim
	filename = strings.TrimSpace(filename)
	for strings.Contains(filename, "  ") {
		filename = strings.ReplaceAll(filename, "  ", " ")
	}

	return filename
}
