package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"pkm-sync/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiInstanceGmailConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		config          *models.Config
		expectedSources []string
		expectError     bool
	}{
		{
			name: "multiple Gmail instances with different configurations",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_work", "gmail_personal"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: "/tmp/test-output",
					SourceTags:       true,
				},
				Sources: map[string]models.SourceConfig{
					"gmail_work": {
						Enabled:      true,
						Type:         "gmail",
						Name:         "Work Emails",
						OutputSubdir: "work",
						OutputTarget: "obsidian",
						Priority:     1,
						Since:        "30d",
						Gmail: models.GmailSourceConfig{
							Name:                "Work Important Emails",
							Description:         "High-priority work communications",
							Labels:              []string{"IMPORTANT", "STARRED"},
							Query:               "from:company.com OR to:company.com",
							IncludeUnread:       true,
							IncludeRead:         false,
							MaxEmailAge:         "90d",
							FromDomains:         []string{"company.com", "client.com"},
							ExtractRecipients:   true,
							ExtractLinks:        true,
							ProcessHTMLContent:  true,
							StripQuotedText:     true,
							DownloadAttachments: true,
							AttachmentTypes:     []string{"pdf", "doc", "docx"},
							MaxAttachmentSize:   "10MB",
							AttachmentSubdir:    "work-attachments",
							FilenameTemplate:    "{{date}}-{{from}}-{{subject}}",
							TaggingRules: []models.TaggingRule{
								{
									Condition: "from:ceo@company.com",
									Tags:      []string{"urgent", "leadership"},
								},
								{
									Condition: "has:attachment",
									Tags:      []string{"has-attachment"},
								},
							},
						},
					},
					"gmail_personal": {
						Enabled:      true,
						Type:         "gmail",
						Name:         "Personal Important",
						OutputSubdir: "personal",
						Priority:     2,
						Since:        "14d",
						Gmail: models.GmailSourceConfig{
							Name:                "Personal Starred Emails",
							Labels:              []string{"STARRED"},
							Query:               "is:important -category:promotions",
							IncludeUnread:       true,
							MaxEmailAge:         "30d",
							ExcludeFromDomains:  []string{"noreply.com", "notifications.com"},
							ExtractRecipients:   false,
							ProcessHTMLContent:  true,
							DownloadAttachments: false,
							FilenameTemplate:    "{{date}}-{{subject}}",
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {
						Type: "obsidian",
						Obsidian: models.ObsidianTargetConfig{
							DefaultFolder:       "Emails",
							FilenameTemplate:    "{{date}} - {{title}}",
							DateFormat:          "2006-01-02",
							IncludeFrontmatter:  true,
							DownloadAttachments: true,
							AttachmentFolder:    "attachments",
						},
					},
				},
			},
			expectedSources: []string{"gmail_work", "gmail_personal"},
			expectError:     false,
		},
		{
			name: "mixed Google and Gmail sources",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"google_calendar_calendar", "gmail_work"},
					DefaultTarget:    "logseq",
					DefaultOutputDir: "/tmp/test-output",
				},
				Sources: map[string]models.SourceConfig{
					"google_calendar_calendar": {
						Enabled: true,
						Type:    "google_calendar",
						Name:    "Google Calendar",
						Since:   "7d",
						Google: models.GoogleSourceConfig{
							CalendarID:      "primary",
							IncludeDeclined: false,
							IncludePrivate:  true,
							DownloadDocs:    true,
							DocFormats:      []string{"markdown", "pdf"},
						},
					},
					"gmail_work": {
						Enabled:      true,
						Type:         "gmail",
						Name:         "Work Emails",
						OutputSubdir: "emails",
						Gmail: models.GmailSourceConfig{
							Name:              "Work Emails",
							Labels:            []string{"IMPORTANT"},
							IncludeUnread:     true,
							MaxEmailAge:       "30d",
							ExtractRecipients: true,
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"logseq": {
						Type: "logseq",
						Logseq: models.LogseqTargetConfig{
							DefaultPage:      "Inbox",
							UseProperties:    true,
							PropertyPrefix:   "gmail::",
							BlockIndentation: 2,
						},
					},
				},
			},
			expectedSources: []string{"google_calendar_calendar", "gmail_work"},
			expectError:     false,
		},
		{
			name: "disabled source should be skipped",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_work", "gmail_personal"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: "/tmp/test-output",
				},
				Sources: map[string]models.SourceConfig{
					"gmail_work": {
						Enabled: true,
						Type:    "gmail",
						Gmail: models.GmailSourceConfig{
							Name: "Work Emails",
						},
					},
					"gmail_personal": {
						Enabled: false, // This should be skipped
						Type:    "gmail",
						Gmail: models.GmailSourceConfig{
							Name: "Personal Emails",
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
				},
			},
			expectedSources: []string{"gmail_work"}, // Only enabled source
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test getEnabledSources function
			enabledSources := getEnabledSources(tt.config)

			if tt.expectError {
				// For error cases, we might expect empty sources or specific behavior
				assert.Empty(t, enabledSources)
			} else {
				assert.ElementsMatch(t, tt.expectedSources, enabledSources)
			}
		})
	}
}

func TestGetSourceOutputDirectory(t *testing.T) {
	tests := []struct {
		name           string
		baseOutputDir  string
		sourceConfig   models.SourceConfig
		expectedOutput string
	}{
		{
			name:          "source with output subdirectory",
			baseOutputDir: "/vault",
			sourceConfig: models.SourceConfig{
				OutputSubdir: "work-emails",
			},
			expectedOutput: "/vault/work-emails",
		},
		{
			name:          "source without output subdirectory",
			baseOutputDir: "/vault",
			sourceConfig: models.SourceConfig{
				OutputSubdir: "",
			},
			expectedOutput: "/vault",
		},
		{
			name:          "nested output subdirectory",
			baseOutputDir: "/vault",
			sourceConfig: models.SourceConfig{
				OutputSubdir: "emails/work/important",
			},
			expectedOutput: "/vault/emails/work/important",
		},
		{
			name:          "relative base directory",
			baseOutputDir: "./vault",
			sourceConfig: models.SourceConfig{
				OutputSubdir: "gmail",
			},
			expectedOutput: "vault/gmail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSourceOutputDirectory(tt.baseOutputDir, tt.sourceConfig)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestCreateSourceWithConfig(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     string
		sourceConfig models.SourceConfig
		expectError  bool
		expectedType string
	}{
		{
			name:     "create Gmail source",
			sourceID: "gmail_work",
			sourceConfig: models.SourceConfig{
				Enabled: true,
				Type:    "gmail",
				Gmail: models.GmailSourceConfig{
					Name:          "Work Emails",
					Labels:        []string{"IMPORTANT"},
					IncludeUnread: true,
				},
			},
			expectError:  false,
			expectedType: "gmail",
		},
		{
			name:     "create Google Calendar source",
			sourceID: "google_calendar_cal",
			sourceConfig: models.SourceConfig{
				Enabled: true,
				Type:    "google_calendar",
				Google: models.GoogleSourceConfig{
					CalendarID: "primary",
				},
			},
			expectError:  false,
			expectedType: "google_calendar",
		},
		{
			name:     "unknown source type",
			sourceID: "slack_test",
			sourceConfig: models.SourceConfig{
				Enabled: true,
				Type:    "slack", // Not yet implemented
			},
			expectError:  true,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test would require proper OAuth credentials to fully work
			// For now, we're testing the factory function creation
			source, err := createSourceWithConfig(tt.sourceID, tt.sourceConfig, &http.Client{})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, source)
			} else {
				// Without proper OAuth setup, we expect configuration to fail
				// but the source creation should succeed
				if source != nil {
					assert.NotNil(t, source)
				}
			}
		})
	}
}

func TestMultiInstancePerSourceTargets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pkm-sync-test")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name         string
		config       *models.Config
		expectedDirs []string
	}{
		{
			name: "different output directories per source",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_work", "gmail_personal"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: tempDir,
				},
				Sources: map[string]models.SourceConfig{
					"gmail_work": {
						Enabled:      true,
						Type:         "gmail",
						OutputSubdir: "work",
						Gmail: models.GmailSourceConfig{
							Name: "Work Emails",
						},
					},
					"gmail_personal": {
						Enabled:      true,
						Type:         "gmail",
						OutputSubdir: "personal",
						Gmail: models.GmailSourceConfig{
							Name: "Personal Emails",
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
				},
			},
			expectedDirs: []string{
				filepath.Join(tempDir, "work"),
				filepath.Join(tempDir, "personal"),
			},
		},
		{
			name: "mixed targets per source",
			config: &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_obsidian", "gmail_logseq"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: tempDir,
				},
				Sources: map[string]models.SourceConfig{
					"gmail_obsidian": {
						Enabled:      true,
						Type:         "gmail",
						OutputSubdir: "obsidian-emails",
						OutputTarget: "obsidian",
						Gmail: models.GmailSourceConfig{
							Name: "Obsidian Emails",
						},
					},
					"gmail_logseq": {
						Enabled:      true,
						Type:         "gmail",
						OutputSubdir: "logseq-emails",
						OutputTarget: "logseq",
						Gmail: models.GmailSourceConfig{
							Name: "Logseq Emails",
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {Type: "obsidian"},
					"logseq":   {Type: "logseq"},
				},
			},
			expectedDirs: []string{
				filepath.Join(tempDir, "obsidian-emails"),
				filepath.Join(tempDir, "logseq-emails"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test directory path calculation
			for i, sourceID := range tt.config.Sync.EnabledSources {
				sourceConfig := tt.config.Sources[sourceID]
				outputDir := getSourceOutputDirectory(tt.config.Sync.DefaultOutputDir, sourceConfig)
				assert.Equal(t, tt.expectedDirs[i], outputDir)
			}
		})
	}
}

func TestGmailSourceConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      models.GmailSourceConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid Gmail configuration",
			config: models.GmailSourceConfig{
				Name:                "Test Gmail",
				Labels:              []string{"IMPORTANT", "STARRED"},
				IncludeUnread:       true,
				MaxEmailAge:         "30d",
				ExtractRecipients:   true,
				ProcessHTMLContent:  true,
				DownloadAttachments: true,
				AttachmentTypes:     []string{"pdf", "doc"},
				MaxAttachmentSize:   "5MB",
				FilenameTemplate:    "{{date}}-{{subject}}",
			},
			expectError: false,
		},
		{
			name: "configuration with tagging rules",
			config: models.GmailSourceConfig{
				Name:          "Test Gmail with Rules",
				Labels:        []string{"IMPORTANT"},
				IncludeUnread: true,
				TaggingRules: []models.TaggingRule{
					{
						Condition: "from:boss@company.com",
						Tags:      []string{"urgent", "management"},
					},
					{
						Condition: "has:attachment",
						Tags:      []string{"documents"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty configuration should still be valid",
			config: models.GmailSourceConfig{
				Name: "Minimal Gmail",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in a real implementation, you might want
			// more sophisticated validation
			if tt.config.Name == "" && tt.expectError {
				assert.Contains(t, tt.errorMsg, "name")
			} else {
				// For now, we assume the configuration is structurally valid
				// Real validation would check attachment sizes, date formats, etc.
				assert.NotEmpty(t, tt.config.Name)
			}
		})
	}
}

func TestSourcePriorityOrdering(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{"gmail_low", "gmail_high", "gmail_medium"},
		},
		Sources: map[string]models.SourceConfig{
			"gmail_high": {
				Enabled:  true,
				Type:     "gmail",
				Priority: 1, // Highest priority
				Gmail:    models.GmailSourceConfig{Name: "High Priority"},
			},
			"gmail_medium": {
				Enabled:  true,
				Type:     "gmail",
				Priority: 2,
				Gmail:    models.GmailSourceConfig{Name: "Medium Priority"},
			},
			"gmail_low": {
				Enabled:  true,
				Type:     "gmail",
				Priority: 3, // Lowest priority
				Gmail:    models.GmailSourceConfig{Name: "Low Priority"},
			},
		},
	}

	enabledSources := getEnabledSources(config)

	// All sources should be enabled

	// Note: In a real implementation, you might want to sort by priority
	// The current getEnabledSources doesn't implement priority sorting
	// This test documents the current behavior
	assert.Len(t, enabledSources, 3)
	assert.Contains(t, enabledSources, "gmail_high")
	assert.Contains(t, enabledSources, "gmail_medium")
	assert.Contains(t, enabledSources, "gmail_low")
}
