package main

import (
	"testing"

	"pkm-sync/pkg/models"
)

func TestGetEnabledSources_ExplicitList(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{"google", "slack"},
		},
		Sources: map[string]models.SourceConfig{
			"google": {
				Enabled: true,
				Type:    "google",
			},
			"slack": {
				Enabled: true,
				Type:    "slack",
			},
			"gmail": {
				Enabled: false, // Disabled, should not be included
				Type:    "gmail",
			},
		},
	}

	enabledSources := getEnabledSources(config)

	if len(enabledSources) != 2 {
		t.Errorf("Expected 2 enabled sources, got %d", len(enabledSources))
	}

	expectedSources := map[string]bool{
		"google": false,
		"slack":  false,
	}

	for _, source := range enabledSources {
		if _, exists := expectedSources[source]; !exists {
			t.Errorf("Unexpected source in enabled list: %s", source)
		}
		expectedSources[source] = true
	}

	for source, found := range expectedSources {
		if !found {
			t.Errorf("Expected source %s to be enabled", source)
		}
	}
}

func TestGetEnabledSources_ExplicitListButSourceDisabled(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{"google", "slack"},
		},
		Sources: map[string]models.SourceConfig{
			"google": {
				Enabled: true,
				Type:    "google",
			},
			"slack": {
				Enabled: false, // Listed in enabled_sources but disabled
				Type:    "slack",
			},
		},
	}

	enabledSources := getEnabledSources(config)

	// Should only include google since slack is disabled
	if len(enabledSources) != 1 {
		t.Errorf("Expected 1 enabled source, got %d", len(enabledSources))
	}

	if enabledSources[0] != "google" {
		t.Errorf("Expected google to be the only enabled source, got %s", enabledSources[0])
	}
}

func TestGetEnabledSources_FallbackToEnabled(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{}, // Empty list, should fallback
		},
		Sources: map[string]models.SourceConfig{
			"google": {
				Enabled: true,
				Type:    "google",
			},
			"slack": {
				Enabled: false,
				Type:    "slack",
			},
			"gmail": {
				Enabled: true,
				Type:    "gmail",
			},
		},
	}

	enabledSources := getEnabledSources(config)

	// Should include google and gmail based on enabled flags
	if len(enabledSources) != 2 {
		t.Errorf("Expected 2 enabled sources, got %d", len(enabledSources))
	}

	enabledMap := make(map[string]bool)
	for _, source := range enabledSources {
		enabledMap[source] = true
	}

	if !enabledMap["google"] {
		t.Error("Expected google to be enabled")
	}

	if !enabledMap["gmail"] {
		t.Error("Expected gmail to be enabled")
	}

	if enabledMap["slack"] {
		t.Error("Expected slack to be disabled")
	}
}

func TestGetEnabledSources_EmptyConfig(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{},
		},
		Sources: map[string]models.SourceConfig{},
	}

	enabledSources := getEnabledSources(config)

	if len(enabledSources) != 0 {
		t.Errorf("Expected 0 enabled sources for empty config, got %d", len(enabledSources))
	}
}

func TestGetEnabledSources_SourceNotInConfig(t *testing.T) {
	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources: []string{"google", "nonexistent"},
		},
		Sources: map[string]models.SourceConfig{
			"google": {
				Enabled: true,
				Type:    "google",
			},
			// "nonexistent" source is not defined
		},
	}

	enabledSources := getEnabledSources(config)

	// Should only include google, skip nonexistent source
	if len(enabledSources) != 1 {
		t.Errorf("Expected 1 enabled source, got %d", len(enabledSources))
	}

	if enabledSources[0] != "google" {
		t.Errorf("Expected google to be the only enabled source, got %s", enabledSources[0])
	}
}

func TestParseSinceTime_RelativeDays(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool // whether it should succeed
	}{
		{"7d", true},
		{"1d", true},
		{"30d", true},
		{"0d", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseSinceTime(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %s to parse successfully, got error: %v", tc.input, err)
			}
			if tc.expected && result.IsZero() {
				t.Errorf("Expected %s to return valid time, got zero time", tc.input)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected %s to fail parsing, but it succeeded", tc.input)
			}
		})
	}
}

func TestParseSinceTime_RelativeHours(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"24h", true},
		{"1h", true},
		{"168h", true}, // 7 days in hours
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseSinceTime(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %s to parse successfully, got error: %v", tc.input, err)
			}
			if tc.expected && result.IsZero() {
				t.Errorf("Expected %s to return valid time, got zero time", tc.input)
			}
		})
	}
}

func TestParseSinceTime_SpecialValues(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"today", true},
		{"yesterday", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseSinceTime(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %s to parse successfully, got error: %v", tc.input, err)
			}
			if tc.expected && result.IsZero() {
				t.Errorf("Expected %s to return valid time, got zero time", tc.input)
			}
		})
	}
}

func TestParseSinceTime_AbsoluteDates(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"2025-01-01", true},
		{"2024-12-31", true},
		{"2025-02-29", false}, // Invalid date (2025 is not a leap year)
		{"invalid-date", false},
		{"2025/01/01", false}, // Wrong format
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseSinceTime(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %s to parse successfully, got error: %v", tc.input, err)
			}
			if tc.expected && result.IsZero() {
				t.Errorf("Expected %s to return valid time, got zero time", tc.input)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected %s to fail parsing, but it succeeded", tc.input)
			}
		})
	}
}

func TestParseSinceTime_InvalidInputs(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"7days", // Should be "7d"
		"1week",
		"abc",
		"-1d",    // Negative days should be invalid
		"-5d",    // Negative days should be invalid
		"d",      // Missing number
		"3.5d",   // Float days should be invalid
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			_, err := parseSinceTime(input)
			if err == nil {
				t.Errorf("Expected %s to fail parsing, but it succeeded", input)
			}
		})
	}
}

func TestParseSinceTime_EdgeCases(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"0d", true, "zero days should be valid"},
		{"365d", true, "large number of days should be valid"},
		{"1000d", true, "very large number of days should be valid"},
		{"24h", true, "24 hours should equal 1 day"},
		{"168h", true, "168 hours should equal 7 days"},
	}

	for _, tc := range testCases {
		t.Run(tc.input+"_"+tc.desc, func(t *testing.T) {
			result, err := parseSinceTime(tc.input)
			if tc.expected && err != nil {
				t.Errorf("Expected %s to parse successfully (%s), got error: %v", tc.input, tc.desc, err)
			}
			if tc.expected && result.IsZero() {
				t.Errorf("Expected %s to return valid time (%s), got zero time", tc.input, tc.desc)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected %s to fail parsing (%s), but it succeeded", tc.input, tc.desc)
			}
		})
	}
}

func TestCreateSource_Google(t *testing.T) {
	source, err := createSource("google")
	if err != nil {
		t.Fatalf("Failed to create google source: %v", err)
	}

	if source == nil {
		t.Error("Expected non-nil source")
	}
}

func TestCreateSource_Unknown(t *testing.T) {
	_, err := createSource("unknown")
	if err == nil {
		t.Error("Expected error for unknown source")
	}

	expectedError := "unknown source 'unknown': supported sources are 'google' (others like slack, gmail, jira are planned for future releases)"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestCreateSource_FutureSources(t *testing.T) {
	futureSources := []string{"slack", "gmail", "jira"}

	for _, sourceName := range futureSources {
		t.Run(sourceName, func(t *testing.T) {
			_, err := createSource(sourceName)
			if err == nil {
				t.Errorf("Expected error for unimplemented source %s", sourceName)
			}
		})
	}
}

func TestCreateTarget_Obsidian(t *testing.T) {
	target, err := createTarget("obsidian")
	if err != nil {
		t.Fatalf("Failed to create obsidian target: %v", err)
	}

	if target == nil {
		t.Error("Expected non-nil target")
	}
}

func TestCreateTarget_Logseq(t *testing.T) {
	target, err := createTarget("logseq")
	if err != nil {
		t.Fatalf("Failed to create logseq target: %v", err)
	}

	if target == nil {
		t.Error("Expected non-nil target")
	}
}

func TestCreateTarget_Unknown(t *testing.T) {
	_, err := createTarget("unknown")
	if err == nil {
		t.Error("Expected error for unknown target")
	}

	expectedError := "unknown target 'unknown': supported targets are 'obsidian' and 'logseq'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}