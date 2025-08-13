package models

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestConfigSerialization_YAML(t *testing.T) {
	// Create a test config
	config := &Config{
		Sync: SyncConfig{
			EnabledSources:   []string{"google_calendar", "slack"},
			DefaultTarget:    "obsidian",
			DefaultSince:     "7d",
			DefaultOutputDir: "./exported",
			MergeSources:     true,
			SourceTags:       true,
			OnConflict:       "skip",
			SyncInterval:     time.Hour * 24,
		},
		Sources: map[string]SourceConfig{
			"google_calendar": {
				Enabled:  true,
				Type:     "google_calendar",
				Priority: 1,
				Google: GoogleSourceConfig{
					CalendarID:      "primary",
					IncludeDeclined: false,
					IncludePrivate:  true,
					DownloadDocs:    true,
					DocFormats:      []string{"markdown"},
					MaxDocSize:      "10MB",
				},
			},
		},
		Targets: map[string]TargetConfig{
			"obsidian": {
				Type: "obsidian",
				Obsidian: ObsidianTargetConfig{
					DefaultFolder:      "Calendar",
					FilenameTemplate:   "{{date}} - {{title}}",
					DateFormat:         "2006-01-02",
					IncludeFrontmatter: true,
				},
			},
		},
		Auth: AuthConfig{
			CredentialsPath: "/path/to/credentials.json",
			TokenPath:       "/path/to/token.json",
			EncryptTokens:   false,
			TokenExpiration: "30d",
		},
		App: AppConfig{
			LogLevel:        "info",
			CreateBackups:   true,
			MaxBackups:      5,
			CacheEnabled:    true,
			NotifyOnSuccess: false,
			NotifyOnError:   true,
		},
	}

	// Serialize to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config to YAML: %v", err)
	}

	// Deserialize from YAML
	var deserializedConfig Config

	err = yaml.Unmarshal(yamlData, &deserializedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from YAML: %v", err)
	}

	// Verify key fields
	if len(deserializedConfig.Sync.EnabledSources) != 2 {
		t.Errorf("Expected 2 enabled sources, got %d", len(deserializedConfig.Sync.EnabledSources))
	}

	if deserializedConfig.Sync.DefaultTarget != "obsidian" {
		t.Errorf("Expected default target obsidian, got %s", deserializedConfig.Sync.DefaultTarget)
	}

	google_calendarSource, exists := deserializedConfig.Sources["google_calendar"]
	if !exists {
		t.Fatal("Expected google_calendar source to exist")
	}

	if google_calendarSource.Google.CalendarID != "primary" {
		t.Errorf("Expected calendar ID primary, got %s", google_calendarSource.Google.CalendarID)
	}

	obsidianTarget, exists := deserializedConfig.Targets["obsidian"]
	if !exists {
		t.Fatal("Expected obsidian target to exist")
	}

	if obsidianTarget.Obsidian.DefaultFolder != "Calendar" {
		t.Errorf("Expected default folder Calendar, got %s", obsidianTarget.Obsidian.DefaultFolder)
	}
}

func TestConfigSerialization_JSON(t *testing.T) {
	// Create a minimal test config
	config := &Config{
		Sync: SyncConfig{
			EnabledSources: []string{"google_calendar"},
			DefaultTarget:  "obsidian",
		},
		Sources: map[string]SourceConfig{
			"google_calendar": {
				Enabled: true,
				Type:    "google_calendar",
			},
		},
		Targets: map[string]TargetConfig{
			"obsidian": {
				Type: "obsidian",
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	// Deserialize from JSON
	var deserializedConfig Config

	err = json.Unmarshal(jsonData, &deserializedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config from JSON: %v", err)
	}

	// Verify structure
	if deserializedConfig.Sync.DefaultTarget != "obsidian" {
		t.Errorf("Expected default target obsidian, got %s", deserializedConfig.Sync.DefaultTarget)
	}

	if len(deserializedConfig.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(deserializedConfig.Sources))
	}

	if len(deserializedConfig.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(deserializedConfig.Targets))
	}
}

func TestSyncConfigDefaults(t *testing.T) {
	config := SyncConfig{}

	// Test zero values
	if len(config.EnabledSources) != 0 {
		t.Error("Expected empty enabled sources by default")
	}

	if config.MergeSources {
		t.Error("Expected merge_sources to be false by default")
	}

	if config.SourceTags {
		t.Error("Expected source_tags to be false by default")
	}

	if config.AutoSync {
		t.Error("Expected auto_sync to be false by default")
	}
}

func TestSourceConfigValidation(t *testing.T) {
	testCases := []struct {
		name   string
		config SourceConfig
		valid  bool
	}{
		{
			name: "valid google_calendar source",
			config: SourceConfig{
				Enabled:  true,
				Type:     "google_calendar",
				Priority: 1,
				Google: GoogleSourceConfig{
					CalendarID: "primary",
				},
			},
			valid: true,
		},
		{
			name: "disabled source",
			config: SourceConfig{
				Enabled: false,
				Type:    "google_calendar",
			},
			valid: true, // Disabled sources are valid
		},
		{
			name: "missing type",
			config: SourceConfig{
				Enabled: true,
				Type:    "", // Missing type
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation - check required fields
			if tc.config.Enabled && tc.config.Type == "" {
				if tc.valid {
					t.Error("Expected config to be invalid due to missing type")
				}
			} else {
				if !tc.valid {
					t.Error("Expected config to be valid")
				}
			}
		})
	}
}

func TestTargetConfigValidation(t *testing.T) {
	testCases := []struct {
		name   string
		config TargetConfig
		valid  bool
	}{
		{
			name: "valid obsidian target",
			config: TargetConfig{
				Type: "obsidian",
				Obsidian: ObsidianTargetConfig{
					DefaultFolder: "Calendar",
					DateFormat:    "2006-01-02",
				},
			},
			valid: true,
		},
		{
			name: "valid logseq target",
			config: TargetConfig{
				Type: "logseq",
				Logseq: LogseqTargetConfig{
					DefaultPage:   "Calendar",
					UseProperties: true,
				},
			},
			valid: true,
		},
		{
			name: "missing type",
			config: TargetConfig{
				Type: "", // Missing type
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation
			if tc.config.Type == "" {
				if tc.valid {
					t.Error("Expected config to be invalid due to missing type")
				}
			} else {
				if !tc.valid {
					t.Error("Expected config to be valid")
				}
			}
		})
	}
}

func TestGoogleSourceConfigDefaults(t *testing.T) {
	config := GoogleSourceConfig{}

	// Test zero values
	if config.CalendarID != "" {
		t.Error("Expected empty calendar ID by default")
	}

	if config.IncludeDeclined {
		t.Error("Expected include_declined to be false by default")
	}

	if config.IncludePrivate {
		t.Error("Expected include_private to be false by default")
	}

	if config.DownloadDocs {
		t.Error("Expected download_docs to be false by default")
	}

	if len(config.DocFormats) != 0 {
		t.Error("Expected empty doc_formats by default")
	}

	if len(config.EventTypes) != 0 {
		t.Error("Expected empty event_types by default")
	}

	if config.RequestDelay != 0 {
		t.Error("Expected request_delay to be 0 by default")
	}

	if config.MaxRequests != 0 {
		t.Error("Expected max_requests to be 0 by default")
	}
}

func TestObsidianTargetConfigDefaults(t *testing.T) {
	config := ObsidianTargetConfig{}

	// Test zero values
	if config.DefaultFolder != "" {
		t.Error("Expected empty default_folder by default")
	}

	if config.FilenameTemplate != "" {
		t.Error("Expected empty filename_template by default")
	}

	if config.DateFormat != "" {
		t.Error("Expected empty date_format by default")
	}

	if config.IncludeFrontmatter {
		t.Error("Expected include_frontmatter to be false by default")
	}

	if config.CreateDailyNotes {
		t.Error("Expected create_daily_notes to be false by default")
	}

	if config.DownloadAttachments {
		t.Error("Expected download_attachments to be false by default")
	}

	if len(config.CustomFields) != 0 {
		t.Error("Expected empty custom_fields by default")
	}
}

func TestLogseqTargetConfigDefaults(t *testing.T) {
	config := LogseqTargetConfig{}

	// Test zero values
	if config.DefaultPage != "" {
		t.Error("Expected empty default_page by default")
	}

	if config.UseProperties {
		t.Error("Expected use_properties to be false by default")
	}

	if config.PropertyPrefix != "" {
		t.Error("Expected empty property_prefix by default")
	}

	if config.BlockIndentation != 0 {
		t.Error("Expected block_indentation to be 0 by default")
	}

	if config.CreateJournalRefs {
		t.Error("Expected create_journal_refs to be false by default")
	}

	if config.JournalDateFormat != "" {
		t.Error("Expected empty journal_date_format by default")
	}
}

func TestAuthConfigDefaults(t *testing.T) {
	config := AuthConfig{}

	// Test zero values
	if config.CredentialsPath != "" {
		t.Error("Expected empty credentials_path by default")
	}

	if config.TokenPath != "" {
		t.Error("Expected empty token_path by default")
	}

	if config.EncryptTokens {
		t.Error("Expected encrypt_tokens to be false by default")
	}

	if config.TokenExpiration != "" {
		t.Error("Expected empty token_expiration by default")
	}
}

func TestAppConfigDefaults(t *testing.T) {
	config := AppConfig{}

	// Test zero values
	if config.LogLevel != "" {
		t.Error("Expected empty log_level by default")
	}

	if config.QuietMode {
		t.Error("Expected quiet_mode to be false by default")
	}

	if config.VerboseMode {
		t.Error("Expected verbose_mode to be false by default")
	}

	if config.CreateBackups {
		t.Error("Expected create_backups to be false by default")
	}

	if config.MaxBackups != 0 {
		t.Error("Expected max_backups to be 0 by default")
	}

	if config.CacheEnabled {
		t.Error("Expected cache_enabled to be false by default")
	}

	if config.NotifyOnSuccess {
		t.Error("Expected notify_on_success to be false by default")
	}

	if config.NotifyOnError {
		t.Error("Expected notify_on_error to be false by default")
	}
}
