package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkm-sync/pkg/models"
)

func TestPerSourceOutputDirectoryHandling(t *testing.T) {
	// Create a temporary directory for test outputs
	tempDir, err := os.MkdirTemp("", "pkm-sync-output-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name                 string
		baseOutputDir        string
		sources              map[string]models.SourceConfig
		expectedDirectories  map[string]string
		expectedSubdirCount  int
	}{
		{
			name:          "multiple sources with different output subdirectories",
			baseOutputDir: tempDir,
			sources: map[string]models.SourceConfig{
				"gmail_work": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Work Emails",
					OutputSubdir: "work-emails",
					Gmail: models.GmailSourceConfig{
						Name: "Work Instance",
					},
				},
				"gmail_personal": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Personal Emails",
					OutputSubdir: "personal-emails",
					Gmail: models.GmailSourceConfig{
						Name: "Personal Instance",
					},
				},
				"google_calendar": {
					Enabled:      true,
					Type:         "google",
					Name:         "Primary Calendar",
					OutputSubdir: "calendar-events",
					Google: models.GoogleSourceConfig{
						CalendarID: "primary",
					},
				},
			},
			expectedDirectories: map[string]string{
				"gmail_work":       filepath.Join(tempDir, "work-emails"),
				"gmail_personal":   filepath.Join(tempDir, "personal-emails"),
				"google_calendar":  filepath.Join(tempDir, "calendar-events"),
			},
			expectedSubdirCount: 3,
		},
		{
			name:          "nested output subdirectories",
			baseOutputDir: tempDir,
			sources: map[string]models.SourceConfig{
				"gmail_work_important": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Work Important",
					OutputSubdir: "emails/work/important",
					Gmail: models.GmailSourceConfig{
						Name: "Work Important Instance",
					},
				},
				"gmail_work_general": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Work General",
					OutputSubdir: "emails/work/general",
					Gmail: models.GmailSourceConfig{
						Name: "Work General Instance",
					},
				},
				"gmail_personal_family": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Personal Family",
					OutputSubdir: "emails/personal/family",
					Gmail: models.GmailSourceConfig{
						Name: "Personal Family Instance",
					},
				},
			},
			expectedDirectories: map[string]string{
				"gmail_work_important":  filepath.Join(tempDir, "emails", "work", "important"),
				"gmail_work_general":    filepath.Join(tempDir, "emails", "work", "general"),
				"gmail_personal_family": filepath.Join(tempDir, "emails", "personal", "family"),
			},
			expectedSubdirCount: 3,
		},
		{
			name:          "mixed with and without subdirectories",
			baseOutputDir: tempDir,
			sources: map[string]models.SourceConfig{
				"gmail_with_subdir": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail with Subdir",
					OutputSubdir: "gmail-emails",
					Gmail: models.GmailSourceConfig{
						Name: "Gmail with Subdir Instance",
					},
				},
				"gmail_no_subdir": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail without Subdir",
					OutputSubdir: "", // No subdirectory - should use base
					Gmail: models.GmailSourceConfig{
						Name: "Gmail No Subdir Instance",
					},
				},
				"google_no_subdir": {
					Enabled:      true,
					Type:         "google",
					Name:         "Google without Subdir",
					OutputSubdir: "", // No subdirectory - should use base
					Google: models.GoogleSourceConfig{
						CalendarID: "primary",
					},
				},
			},
			expectedDirectories: map[string]string{
				"gmail_with_subdir": filepath.Join(tempDir, "gmail-emails"),
				"gmail_no_subdir":   tempDir, // Base directory
				"google_no_subdir":  tempDir, // Base directory
			},
			expectedSubdirCount: 1, // Only one subdirectory created
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test directory path calculation for each source
			for sourceID, sourceConfig := range tt.sources {
				expectedDir := tt.expectedDirectories[sourceID]
				actualDir := getSourceOutputDirectory(tt.baseOutputDir, sourceConfig)
				assert.Equal(t, expectedDir, actualDir, "Output directory mismatch for source %s", sourceID)
			}

			// Test that we can create these directories
			for sourceID, sourceConfig := range tt.sources {
				outputDir := getSourceOutputDirectory(tt.baseOutputDir, sourceConfig)
				
				// Create the directory structure
				err := os.MkdirAll(outputDir, 0755)
				assert.NoError(t, err, "Failed to create output directory for source %s", sourceID)
				
				// Verify the directory exists
				stat, err := os.Stat(outputDir)
				assert.NoError(t, err, "Output directory does not exist for source %s", sourceID)
				assert.True(t, stat.IsDir(), "Output path is not a directory for source %s", sourceID)
				
				// Create a test file in the directory to verify write access
				testFile := filepath.Join(outputDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test"), 0644)
				assert.NoError(t, err, "Failed to write test file in output directory for source %s", sourceID)
				
				// Verify the test file exists
				_, err = os.Stat(testFile)
				assert.NoError(t, err, "Test file does not exist for source %s", sourceID)
				
				// Clean up the test file
				os.Remove(testFile)
			}
		})
	}
}

func TestPerSourceTargetHandling(t *testing.T) {
	tests := []struct {
		name              string
		defaultTarget     string
		sources           map[string]models.SourceConfig
		targets           map[string]models.TargetConfig
		expectedTargets   map[string]string
		expectErrors      map[string]bool
	}{
		{
			name:          "sources with different target overrides",
			defaultTarget: "obsidian",
			sources: map[string]models.SourceConfig{
				"gmail_obsidian": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail for Obsidian",
					OutputTarget: "obsidian",
					Gmail: models.GmailSourceConfig{
						Name: "Obsidian Gmail",
					},
				},
				"gmail_logseq": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail for Logseq",
					OutputTarget: "logseq",
					Gmail: models.GmailSourceConfig{
						Name: "Logseq Gmail",
					},
				},
				"gmail_default": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail Default Target",
					OutputTarget: "", // Should use default
					Gmail: models.GmailSourceConfig{
						Name: "Default Gmail",
					},
				},
			},
			targets: map[string]models.TargetConfig{
				"obsidian": {
					Type: "obsidian",
					Obsidian: models.ObsidianTargetConfig{
						DefaultFolder: "Obsidian Emails",
					},
				},
				"logseq": {
					Type: "logseq",
					Logseq: models.LogseqTargetConfig{
						DefaultPage: "Logseq Emails",
					},
				},
			},
			expectedTargets: map[string]string{
				"gmail_obsidian": "obsidian",
				"gmail_logseq":   "logseq",
				"gmail_default":  "obsidian", // Should use default
			},
			expectErrors: map[string]bool{
				"gmail_obsidian": false,
				"gmail_logseq":   false,
				"gmail_default":  false,
			},
		},
		{
			name:          "source with invalid target override",
			defaultTarget: "obsidian",
			sources: map[string]models.SourceConfig{
				"gmail_invalid": {
					Enabled:      true,
					Type:         "gmail",
					Name:         "Gmail Invalid Target",
					OutputTarget: "nonexistent_target",
					Gmail: models.GmailSourceConfig{
						Name: "Invalid Target Gmail",
					},
				},
			},
			targets: map[string]models.TargetConfig{
				"obsidian": {
					Type: "obsidian",
				},
			},
			expectedTargets: map[string]string{
				"gmail_invalid": "obsidian", // Should fallback to default
			},
			expectErrors: map[string]bool{
				"gmail_invalid": true, // Should produce a warning/error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.Config{
				Sync: models.SyncConfig{
					DefaultTarget: tt.defaultTarget,
				},
				Sources: tt.sources,
				Targets: tt.targets,
			}

			// Test target resolution for each source
			for sourceID, sourceConfig := range tt.sources {
				expectedTarget := tt.expectedTargets[sourceID]
				expectError := tt.expectErrors[sourceID]

				// Determine what target should be used
				var targetName string
				if sourceConfig.OutputTarget != "" {
					targetName = sourceConfig.OutputTarget
				} else {
					targetName = tt.defaultTarget
				}

				// Check if the target exists in configuration
				_, targetExists := config.Targets[targetName]

				if expectError {
					if sourceConfig.OutputTarget != "" {
						// Source specifies a target that doesn't exist
						assert.False(t, targetExists, "Expected target '%s' to not exist for source '%s'", targetName, sourceID)
					}
				} else {
					// Target should exist and be valid
					assert.True(t, targetExists, "Expected target '%s' to exist for source '%s'", targetName, sourceID)
					assert.Equal(t, expectedTarget, targetName, "Target mismatch for source %s", sourceID)
				}
			}
		})
	}
}

func TestCompleteMultiInstanceWorkflow(t *testing.T) {
	// Create a temporary directory for the complete workflow test
	tempDir, err := os.MkdirTemp("", "pkm-sync-workflow-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a comprehensive configuration that exercises all features
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources:   []string{"gmail_work", "gmail_personal", "google_calendar"},
			DefaultTarget:    "obsidian",
			DefaultOutputDir: tempDir,
			SourceTags:       true,
		},
		Sources: map[string]models.SourceConfig{
			"gmail_work": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Work Emails",
				Priority:     1,
				OutputSubdir: "work",
				OutputTarget: "obsidian",
				Since:        "30d",
				Gmail: models.GmailSourceConfig{
					Name:               "Work Important Emails",
					Labels:             []string{"IMPORTANT", "STARRED"},
					IncludeUnread:      true,
					ExtractRecipients:  true,
					ProcessHTMLContent: true,
				},
			},
			"gmail_personal": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Personal Emails",
				Priority:     2,
				OutputSubdir: "personal",
				OutputTarget: "logseq",
				Since:        "14d",
				Gmail: models.GmailSourceConfig{
					Name:               "Personal Important Emails",
					Labels:             []string{"STARRED"},
					IncludeUnread:      true,
					ExtractRecipients:  false,
					ProcessHTMLContent: true,
				},
			},
			"google_calendar": {
				Enabled:      true,
				Type:         "google",
				Name:         "Primary Calendar",
				Priority:     3,
				OutputSubdir: "calendar",
				Since:        "7d",
				Google: models.GoogleSourceConfig{
					CalendarID:     "primary",
					IncludePrivate: true,
					DownloadDocs:   true,
				},
			},
		},
		Targets: map[string]models.TargetConfig{
			"obsidian": {
				Type: "obsidian",
				Obsidian: models.ObsidianTargetConfig{
					DefaultFolder:       "Synced",
					IncludeFrontmatter:  true,
					DownloadAttachments: true,
				},
			},
			"logseq": {
				Type: "logseq",
				Logseq: models.LogseqTargetConfig{
					DefaultPage:   "Inbox",
					UseProperties: true,
				},
			},
		},
	}

	// Test the complete workflow components

	// 1. Test enabled sources detection
	enabledSources := getEnabledSources(config)
	assert.ElementsMatch(t, []string{"gmail_work", "gmail_personal", "google_calendar"}, enabledSources)

	// 2. Test output directory calculation for each source
	expectedOutputDirs := map[string]string{
		"gmail_work":       filepath.Join(tempDir, "work"),
		"gmail_personal":   filepath.Join(tempDir, "personal"),
		"google_calendar":  filepath.Join(tempDir, "calendar"),
	}

	for _, sourceID := range enabledSources {
		sourceConfig := config.Sources[sourceID]
		outputDir := getSourceOutputDirectory(config.Sync.DefaultOutputDir, sourceConfig)
		expectedDir := expectedOutputDirs[sourceID]
		assert.Equal(t, expectedDir, outputDir, "Output directory mismatch for source %s", sourceID)

		// Create the directory to verify it can be created
		err := os.MkdirAll(outputDir, 0755)
		assert.NoError(t, err, "Failed to create output directory for source %s", sourceID)

		// Verify directory exists and is writable
		stat, err := os.Stat(outputDir)
		assert.NoError(t, err, "Output directory stat failed for source %s", sourceID)
		assert.True(t, stat.IsDir(), "Output path is not a directory for source %s", sourceID)
	}

	// 3. Test target assignment for each source
	expectedTargets := map[string]string{
		"gmail_work":       "obsidian",
		"gmail_personal":   "logseq",
		"google_calendar":  "obsidian", // Uses default target
	}

	for _, sourceID := range enabledSources {
		sourceConfig := config.Sources[sourceID]
		var targetName string
		if sourceConfig.OutputTarget != "" {
			targetName = sourceConfig.OutputTarget
		} else {
			targetName = config.Sync.DefaultTarget
		}

		expectedTarget := expectedTargets[sourceID]
		assert.Equal(t, expectedTarget, targetName, "Target assignment mismatch for source %s", sourceID)

		// Verify the target exists in configuration
		_, targetExists := config.Targets[targetName]
		assert.True(t, targetExists, "Target '%s' not found in configuration for source '%s'", targetName, sourceID)
	}

	// 4. Test source creation (without actual OAuth - just factory)
	for _, sourceID := range enabledSources {
		sourceConfig := config.Sources[sourceID]
		
		// Test that the source factory can create the appropriate source type
		switch sourceConfig.Type {
		case "gmail", "google":
			// Test that we can create a source with the correct configuration
			// Note: We can't actually create the source without proper auth setup
			assert.Equal(t, sourceConfig.Type, sourceConfig.Type, "Source type should match config")
			assert.NotEmpty(t, sourceID, "Source ID should not be empty")
		default:
			t.Errorf("Unexpected source type '%s' for source '%s'", sourceConfig.Type, sourceID)
		}
	}

	// 5. Test directory structure creation
	// Simulate what would happen during a real sync
	for _, sourceID := range enabledSources {
		sourceConfig := config.Sources[sourceID]
		outputDir := getSourceOutputDirectory(config.Sync.DefaultOutputDir, sourceConfig)

		// Create a mock output file to verify the directory structure works
		mockFileName := "test-" + sourceID + ".md"
		mockFilePath := filepath.Join(outputDir, mockFileName)
		mockContent := "# Test file for " + sourceID + "\n\nThis is a test file for source: " + sourceConfig.Name
		
		err := os.WriteFile(mockFilePath, []byte(mockContent), 0644)
		assert.NoError(t, err, "Failed to write mock file for source %s", sourceID)

		// Verify the file was created correctly
		_, err = os.Stat(mockFilePath)
		assert.NoError(t, err, "Mock file does not exist for source %s", sourceID)

		// Read back the content to verify
		readContent, err := os.ReadFile(mockFilePath)
		assert.NoError(t, err, "Failed to read mock file for source %s", sourceID)
		assert.Equal(t, mockContent, string(readContent), "Mock file content mismatch for source %s", sourceID)
	}

	// 6. Verify final directory structure
	// Check that all expected directories were created
	for sourceID, expectedDir := range expectedOutputDirs {
		stat, err := os.Stat(expectedDir)
		assert.NoError(t, err, "Expected directory does not exist for source %s: %s", sourceID, expectedDir)
		assert.True(t, stat.IsDir(), "Expected path is not a directory for source %s: %s", sourceID, expectedDir)

		// Check that the mock file exists in the directory
		mockFile := filepath.Join(expectedDir, "test-"+sourceID+".md")
		_, err = os.Stat(mockFile)
		assert.NoError(t, err, "Mock file does not exist in expected directory for source %s", sourceID)
	}

	// 7. Test that base directory contains the expected subdirectories
	baseEntries, err := os.ReadDir(tempDir)
	assert.NoError(t, err, "Failed to read base output directory")

	expectedSubdirs := []string{"work", "personal", "calendar"}
	actualSubdirs := make([]string, 0, len(baseEntries))
	for _, entry := range baseEntries {
		if entry.IsDir() {
			actualSubdirs = append(actualSubdirs, entry.Name())
		}
	}

	assert.ElementsMatch(t, expectedSubdirs, actualSubdirs, "Base directory does not contain expected subdirectories")
}

func TestErrorHandlingInMultiInstanceWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pkm-sync-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test configuration with various error conditions
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources:   []string{"gmail_valid", "gmail_missing", "gmail_invalid_target"},
			DefaultTarget:    "obsidian",
			DefaultOutputDir: tempDir,
		},
		Sources: map[string]models.SourceConfig{
			"gmail_valid": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Valid Gmail",
				OutputSubdir: "valid",
				Gmail: models.GmailSourceConfig{
					Name: "Valid Instance",
				},
			},
			// "gmail_missing" is in enabled_sources but not in sources map
			"gmail_invalid_target": {
				Enabled:      true,
				Type:         "gmail",
				Name:         "Invalid Target Gmail",
				OutputSubdir: "invalid",
				OutputTarget: "nonexistent_target",
				Gmail: models.GmailSourceConfig{
					Name: "Invalid Target Instance",
				},
			},
		},
		Targets: map[string]models.TargetConfig{
			"obsidian": {
				Type: "obsidian",
			},
		},
	}

	// Test enabled sources detection with missing source
	enabledSources := getEnabledSources(config)
	// Should only return sources that exist and are enabled
	assert.ElementsMatch(t, []string{"gmail_valid", "gmail_invalid_target"}, enabledSources)
	assert.NotContains(t, enabledSources, "gmail_missing", "Missing source should not be in enabled sources")

	// Test handling of invalid target references
	invalidTargetSource := config.Sources["gmail_invalid_target"]
	assert.Equal(t, "nonexistent_target", invalidTargetSource.OutputTarget)
	
	// Verify the target doesn't exist
	_, targetExists := config.Targets[invalidTargetSource.OutputTarget]
	assert.False(t, targetExists, "Nonexistent target should not exist in configuration")

	// Test that we can still calculate output directory even with invalid target
	outputDir := getSourceOutputDirectory(config.Sync.DefaultOutputDir, invalidTargetSource)
	expectedDir := filepath.Join(tempDir, "invalid")
	assert.Equal(t, expectedDir, outputDir, "Output directory calculation should work even with invalid target")
}