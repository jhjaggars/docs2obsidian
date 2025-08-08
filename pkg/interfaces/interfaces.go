package interfaces

import (
	"pkm-sync/pkg/models"
	"time"
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
}

// Syncer coordinates between sources and targets.
type Syncer interface {
	Sync(source Source, target Target, options SyncOptions) error
}

type SyncOptions struct {
	Since     time.Time
	OutputDir string
	DryRun    bool
	Overwrite bool
}
