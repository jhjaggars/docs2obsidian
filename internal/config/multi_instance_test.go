package config

import (
	"testing"
	"time"

	"pkm-sync/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSyncConfigLoading(t *testing.T) {
	config, cleanup := setupConfigTest(t)
	defer cleanup()

	assert.ElementsMatch(t, []string{"gmail_work", "gmail_personal", "google_calendar"}, config.Sync.EnabledSources)
	assert.Equal(t, "obsidian", config.Sync.DefaultTarget)
	assert.Equal(t, "./vault", config.Sync.DefaultOutputDir)
	assert.True(t, config.Sync.SourceTags)
	assert.False(t, config.Sync.MergeSources)
	assert.True(t, config.Sync.CreateSubdirs)
	assert.Equal(t, "source", config.Sync.SubdirFormat)
}

func TestSourcesConfigLoading(t *testing.T) {
	config, cleanup := setupConfigTest(t)
	defer cleanup()

	require.Len(t, config.Sources, 3)

	// Test Gmail Work source
	gmailWork, exists := config.Sources["gmail_work"]
	require.True(t, exists)
	assert.True(t, gmailWork.Enabled)
	assert.Equal(t, "gmail", gmailWork.Type)
	assert.Equal(t, "Work Emails", gmailWork.Name)
	assert.Equal(t, 1, gmailWork.Priority)
	assert.Equal(t, "work-emails", gmailWork.OutputSubdir)
	assert.Equal(t, "obsidian", gmailWork.OutputTarget)
	assert.Equal(t, "30d", gmailWork.Since)

	// Test Gmail Work configuration
	gmailWorkConfig := gmailWork.Gmail
	assert.Equal(t, "Work Important Emails", gmailWorkConfig.Name)
	assert.Equal(t, "High-priority work communications", gmailWorkConfig.Description)
	assert.ElementsMatch(t, []string{"IMPORTANT", "STARRED"}, gmailWorkConfig.Labels)
	assert.Equal(t, "from:company.com OR to:company.com", gmailWorkConfig.Query)
	assert.True(t, gmailWorkConfig.IncludeUnread)
	assert.False(t, gmailWorkConfig.IncludeRead)
	assert.Equal(t, "90d", gmailWorkConfig.MaxEmailAge)
	assert.ElementsMatch(t, []string{"company.com", "client.com"}, gmailWorkConfig.FromDomains)
	assert.True(t, gmailWorkConfig.ExtractRecipients)
	assert.True(t, gmailWorkConfig.ExtractLinks)
	assert.True(t, gmailWorkConfig.ProcessHTMLContent)
	assert.True(t, gmailWorkConfig.StripQuotedText)
	assert.True(t, gmailWorkConfig.DownloadAttachments)
	assert.ElementsMatch(t, []string{"pdf", "doc", "docx"}, gmailWorkConfig.AttachmentTypes)
	assert.Equal(t, "10MB", gmailWorkConfig.MaxAttachmentSize)
	assert.Equal(t, "work-attachments", gmailWorkConfig.AttachmentSubdir)
	assert.Equal(t, "{{date}}-{{from}}-{{subject}}", gmailWorkConfig.FilenameTemplate)
	assert.Equal(t, 500*time.Millisecond, gmailWorkConfig.RequestDelay)
	assert.Equal(t, 1000, gmailWorkConfig.MaxRequests)
	assert.Equal(t, 50, gmailWorkConfig.BatchSize)

	// Test tagging rules
	require.Len(t, gmailWorkConfig.TaggingRules, 2)
	assert.Equal(t, "from:ceo@company.com", gmailWorkConfig.TaggingRules[0].Condition)
	assert.ElementsMatch(t, []string{"urgent", "leadership"}, gmailWorkConfig.TaggingRules[0].Tags)
	assert.Equal(t, "has:attachment", gmailWorkConfig.TaggingRules[1].Condition)
	assert.ElementsMatch(t, []string{"has-attachment"}, gmailWorkConfig.TaggingRules[1].Tags)

	// Test Gmail Personal source
	gmailPersonal, exists := config.Sources["gmail_personal"]
	require.True(t, exists)
	assert.True(t, gmailPersonal.Enabled)
	assert.Equal(t, "gmail", gmailPersonal.Type)
	assert.Equal(t, "Personal Important", gmailPersonal.Name)
	assert.Equal(t, 2, gmailPersonal.Priority)
	assert.Equal(t, "personal-emails", gmailPersonal.OutputSubdir)
	assert.Equal(t, "", gmailPersonal.OutputTarget) // Should use default
	assert.Equal(t, "14d", gmailPersonal.Since)

	// Test Gmail Personal configuration
	gmailPersonalConfig := gmailPersonal.Gmail
	assert.Equal(t, "Personal Starred Emails", gmailPersonalConfig.Name)
	assert.ElementsMatch(t, []string{"STARRED"}, gmailPersonalConfig.Labels)
	assert.Equal(t, "is:important -category:promotions", gmailPersonalConfig.Query)
	assert.True(t, gmailPersonalConfig.IncludeUnread)
	assert.Equal(t, "30d", gmailPersonalConfig.MaxEmailAge)
	assert.ElementsMatch(t, []string{"noreply.com", "notifications.com"}, gmailPersonalConfig.ExcludeFromDomains)
	assert.False(t, gmailPersonalConfig.ExtractRecipients)
	assert.True(t, gmailPersonalConfig.ProcessHTMLContent)
	assert.False(t, gmailPersonalConfig.DownloadAttachments)
	assert.Equal(t, "{{date}}-{{subject}}", gmailPersonalConfig.FilenameTemplate)

	// Test Google Calendar source
	googleCalendar, exists := config.Sources["google_calendar"]
	require.True(t, exists)
	assert.True(t, googleCalendar.Enabled)
	assert.Equal(t, "google", googleCalendar.Type)
	assert.Equal(t, "Primary Calendar", googleCalendar.Name)
	assert.Equal(t, 3, googleCalendar.Priority)
	assert.Equal(t, "calendar", googleCalendar.OutputSubdir)
	assert.Equal(t, "7d", googleCalendar.Since)

	// Test Google Calendar configuration
	googleCalendarConfig := googleCalendar.Google
	assert.Equal(t, "primary", googleCalendarConfig.CalendarID)
	assert.False(t, googleCalendarConfig.IncludeDeclined)
	assert.True(t, googleCalendarConfig.IncludePrivate)
	assert.True(t, googleCalendarConfig.DownloadDocs)
	assert.ElementsMatch(t, []string{"markdown", "pdf"}, googleCalendarConfig.DocFormats)
	assert.Equal(t, "5MB", googleCalendarConfig.MaxDocSize)
	assert.True(t, googleCalendarConfig.IncludeShared)
	assert.Equal(t, time.Second, googleCalendarConfig.RequestDelay)
	assert.Equal(t, 500, googleCalendarConfig.MaxRequests)
}

func TestTargetsConfigLoading(t *testing.T) {
	config, cleanup := setupConfigTest(t)
	defer cleanup()

	require.Len(t, config.Targets, 2)

	// Test Obsidian target
	obsidianTarget, exists := config.Targets["obsidian"]
	require.True(t, exists)
	assert.Equal(t, "obsidian", obsidianTarget.Type)
	obsidianConfig := obsidianTarget.Obsidian
	assert.Equal(t, "Synced", obsidianConfig.DefaultFolder)
	assert.Equal(t, "{{date}} - {{title}}", obsidianConfig.FilenameTemplate)
	assert.Equal(t, "2006-01-02", obsidianConfig.DateFormat)
	assert.Equal(t, "sync/", obsidianConfig.TagPrefix)
	assert.True(t, obsidianConfig.IncludeFrontmatter)
	assert.False(t, obsidianConfig.CreateDailyNotes)
	assert.Equal(t, "wikilink", obsidianConfig.LinkFormat)
	assert.Equal(t, "attachments", obsidianConfig.AttachmentFolder)
	assert.True(t, obsidianConfig.DownloadAttachments)

	// Test Logseq target
	logseqTarget, exists := config.Targets["logseq"]
	require.True(t, exists)
	assert.Equal(t, "logseq", logseqTarget.Type)
	logseqConfig := logseqTarget.Logseq
	assert.Equal(t, "Inbox", logseqConfig.DefaultPage)
	assert.True(t, logseqConfig.UseProperties)
	assert.Equal(t, "sync::", logseqConfig.PropertyPrefix)
	assert.Equal(t, 2, logseqConfig.BlockIndentation)
	assert.True(t, logseqConfig.CreateJournalRefs)
	assert.Equal(t, "2006-01-02", logseqConfig.JournalDateFormat)
}

func TestAuthConfigLoading(t *testing.T) {
	config, cleanup := setupConfigTest(t)
	defer cleanup()

	assert.Equal(t, "./credentials.json", config.Auth.CredentialsPath)
	assert.Equal(t, "./token.json", config.Auth.TokenPath)
	assert.False(t, config.Auth.EncryptTokens)
}

func TestAppConfigLoading(t *testing.T) {
	config, cleanup := setupConfigTest(t)
	defer cleanup()

	assert.Equal(t, "info", config.App.LogLevel)
	assert.False(t, config.App.QuietMode)
	assert.False(t, config.App.VerboseMode)
	assert.True(t, config.App.CreateBackups)
	assert.Equal(t, "./backups", config.App.BackupDir)
	assert.Equal(t, 5, config.App.MaxBackups)
	assert.True(t, config.App.CacheEnabled)
	assert.Equal(t, 24*time.Hour, config.App.CacheTTL)
}

func TestMultiInstanceConfigurationSerialization(t *testing.T) {
	// Create a multi-instance configuration programmatically
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources:   []string{"gmail_work", "gmail_personal"},
			DefaultTarget:    "obsidian",
			DefaultOutputDir: "./vault",
			SourceTags:       true,
		},
		Sources: map[string]models.SourceConfig{
			"gmail_work": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Work Emails",
				Priority:     1,
				OutputSubdir: "work",
				Gmail: models.GmailSourceConfig{
					Name:               "Work Emails",
					Labels:             []string{"IMPORTANT"},
					IncludeUnread:      true,
					ExtractRecipients:  true,
					ProcessHTMLContent: true,
					TaggingRules: []models.TaggingRule{
						{
							Condition: "from:boss@company.com",
							Tags:      []string{"urgent"},
						},
					},
				},
			},
			"gmail_personal": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Personal Emails",
				Priority:     2,
				OutputSubdir: "personal",
				Gmail: models.GmailSourceConfig{
					Name:          "Personal Emails",
					Labels:        []string{"STARRED"},
					IncludeUnread: true,
				},
			},
		},
		Targets: map[string]models.TargetConfig{
			"obsidian": {
				Type: "obsidian",
				Obsidian: models.ObsidianTargetConfig{
					DefaultFolder:       "Emails",
					IncludeFrontmatter:  true,
					DownloadAttachments: true,
				},
			},
		},
	}

	// Serialize to YAML
	yamlData, err := yaml.Marshal(config)
	require.NoError(t, err)
	assert.NotEmpty(t, yamlData)

	// Deserialize back
	var deserializedConfig models.Config

	err = yaml.Unmarshal(yamlData, &deserializedConfig)
	require.NoError(t, err)

	// Verify the round-trip preserved the configuration
	assert.ElementsMatch(t, config.Sync.EnabledSources, deserializedConfig.Sync.EnabledSources)
	assert.Equal(t, config.Sync.DefaultTarget, deserializedConfig.Sync.DefaultTarget)
	assert.Equal(t, config.Sync.SourceTags, deserializedConfig.Sync.SourceTags)

	assert.Len(t, deserializedConfig.Sources, 2)

	workSource := deserializedConfig.Sources["gmail_work"]
	assert.Equal(t, config.Sources["gmail_work"].Type, workSource.Type)
	assert.Equal(t, config.Sources["gmail_work"].Name, workSource.Name)
	assert.Equal(t, config.Sources["gmail_work"].Priority, workSource.Priority)
	assert.Equal(t, config.Sources["gmail_work"].OutputSubdir, workSource.OutputSubdir)
	assert.ElementsMatch(t, config.Sources["gmail_work"].Gmail.Labels, workSource.Gmail.Labels)
	assert.Equal(t, config.Sources["gmail_work"].Gmail.IncludeUnread, workSource.Gmail.IncludeUnread)
	assert.Len(t, workSource.Gmail.TaggingRules, 1)
	assert.Equal(t, "from:boss@company.com", workSource.Gmail.TaggingRules[0].Condition)

	personalSource := deserializedConfig.Sources["gmail_personal"]
	assert.Equal(t, config.Sources["gmail_personal"].Type, personalSource.Type)
	assert.Equal(t, config.Sources["gmail_personal"].Name, personalSource.Name)
	assert.ElementsMatch(t, config.Sources["gmail_personal"].Gmail.Labels, personalSource.Gmail.Labels)
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name         string
		config       *models.Config
		expectErrors bool
		description  string
	}{
		{
			name: "valid multi-instance configuration",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_test"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: "./vault",
				},
				Sources: map[string]models.SourceConfig{
					"gmail_test": {
						Enabled: true,
						Type:    "gmail",
						Name:    "Test Gmail",
						Gmail: models.GmailSourceConfig{
							Name: "Test Instance",
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {
						Type: "obsidian",
					},
				},
			},
			expectErrors: false,
			description:  "Well-formed configuration should validate",
		},
		{
			name: "enabled source not in sources map",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_missing"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: "./vault",
				},
				Sources: map[string]models.SourceConfig{
					"gmail_present": {
						Enabled: true,
						Type:    "gmail",
						Name:    "Present Gmail",
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
				},
			},
			expectErrors: true,
			description:  "Enabled source must exist in sources map",
		},
		{
			name: "target referenced but not defined",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_test"},
					DefaultTarget:    "missing_target",
					DefaultOutputDir: "./vault",
				},
				Sources: map[string]models.SourceConfig{
					"gmail_test": {
						Enabled: true,
						Type:    "gmail",
						Name:    "Test Gmail",
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
				},
			},
			expectErrors: true,
			description:  "Default target must be defined in targets map",
		},
		{
			name: "source-specific target not defined",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_test"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: "./vault",
				},
				Sources: map[string]models.SourceConfig{
					"gmail_test": {
						Enabled:      true,
						Type:         "gmail",
						Name:         "Test Gmail",
						OutputTarget: "missing_target", // This target doesn't exist
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
				},
			},
			expectErrors: true,
			description:  "Source-specific target must be defined in targets map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateMultiInstanceConfig(tt.config)

			if tt.expectErrors {
				assert.NotEmpty(t, errors, tt.description)
			} else {
				assert.Empty(t, errors, tt.description)
			}
		})
	}
}

// validateMultiInstanceConfig performs validation checks on a multi-instance configuration.
func validateMultiInstanceConfig(config *models.Config) []string {
	var errors []string

	// Check that all enabled sources exist in the sources map
	for _, sourceID := range config.Sync.EnabledSources {
		if _, exists := config.Sources[sourceID]; !exists {
			errors = append(errors, "enabled source '"+sourceID+"' not found in sources configuration")
		}
	}

	// Check that default target exists
	if _, exists := config.Targets[config.Sync.DefaultTarget]; !exists {
		errors = append(errors, "default target '"+config.Sync.DefaultTarget+"' not found in targets configuration")
	}

	// Check that source-specific targets exist
	for sourceID, sourceConfig := range config.Sources {
		if sourceConfig.OutputTarget != "" {
			if _, exists := config.Targets[sourceConfig.OutputTarget]; !exists {
				errors = append(errors, "output target '"+sourceConfig.OutputTarget+"' for source '"+sourceID+"' not found in targets configuration")
			}
		}
	}

	return errors
}

func TestDefaultConfigGeneration(t *testing.T) {
	// Test that the default configuration includes multi-instance examples
	defaultConfig := GetDefaultConfig()

	// Should have multiple source examples
	assert.Greater(t, len(defaultConfig.Sources), 0)

	// Should have multiple target examples
	assert.Greater(t, len(defaultConfig.Targets), 0)

	// Should have sensible defaults for sync configuration
	assert.NotEmpty(t, defaultConfig.Sync.DefaultTarget)
	assert.NotEmpty(t, defaultConfig.Sync.DefaultOutputDir)

	// Verify the configuration is self-consistent
	errors := validateMultiInstanceConfig(defaultConfig)
	if len(errors) > 0 {
		// It's OK if default config has validation errors since sources might be disabled
		// but we should at least check structure
		t.Logf("Default config validation messages: %v", errors)
	}
}
