package interfaces

import (
	"time"
	"pkm-sync/pkg/models"
)

// Source represents any data source (Google Calendar, Slack, etc.)
type Source interface {
	Name() string
	Configure(config map[string]interface{}) error
	Fetch(since time.Time, limit int) ([]*models.Item, error)
	SupportsRealtime() bool
}

// Target represents any PKM system (Obsidian, Logseq, etc.)
type Target interface {
	Name() string
	Configure(config map[string]interface{}) error
	Export(items []*models.Item, outputDir string) error
	FormatFilename(title string) string
	GetFileExtension() string
	FormatMetadata(metadata map[string]interface{}) string
	Preview(items []*models.Item, outputDir string) ([]*FilePreview, error)
}

// FilePreview represents what would happen to a file during sync
type FilePreview struct {
	FilePath    string    // Full path where file would be created
	Action      string    // "create", "update", "skip"
	Content     string    // Full content that would be written
	ExistingContent string // Current content if file exists
	Conflict    bool      // True if there would be a conflict
}

// Syncer coordinates between sources and targets
type Syncer interface {
	Sync(source Source, target Target, options SyncOptions) error
}

type SyncOptions struct {
	Since     time.Time
	OutputDir string
	DryRun    bool
	Overwrite bool
}