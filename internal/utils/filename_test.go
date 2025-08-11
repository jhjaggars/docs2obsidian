package utils

import (
	"strings"
	"testing"
)

func TestSanitizeFilename_Security(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "path traversal - parent directory",
			input:       "../../../etc/passwd",
			expected:    "etc-passwd",
			description: "Should prevent path traversal attacks",
		},
		{
			name:        "path traversal - current directory",
			input:       "./sensitive/file",
			expected:    "sensitive-file",
			description: "Should remove current directory references",
		},
		{
			name:        "path traversal - mixed",
			input:       "../config/../secrets",
			expected:    "config-secrets",
			description: "Should handle multiple path traversal attempts",
		},
		{
			name:        "home directory reference",
			input:       "~/private/data",
			expected:    "private-data",
			description: "Should remove home directory references",
		},
		{
			name:        "dot files",
			input:       ".hidden_file",
			expected:    "hidden_file",
			description: "Should handle dot files safely",
		},
		{
			name:        "directory separator injection",
			input:       "file/with/slashes",
			expected:    "file-with-slashes",
			description: "Should replace directory separators",
		},
		{
			name:        "windows path separators",
			input:       "file\\with\\backslashes",
			expected:    "file-with-backslashes",
			description: "Should replace Windows directory separators",
		},
		{
			name:        "null bytes",
			input:       "file\x00name",
			expected:    "filename",
			description: "Should handle null bytes",
		},
		{
			name:        "control characters",
			input:       "file\nwith\tcontrol\rchars",
			expected:    "filewithcontrolchars",
			description: "Should remove control characters",
		},
		{
			name:        "complex attack attempt",
			input:       "../../../etc/../home/user/.ssh/id_rsa",
			expected:    "etc-home-user-ssh-id_rsa",
			description: "Should handle complex path traversal attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q\nDescription: %s",
					tt.input, result, tt.expected, tt.description)
			}

			// Additional security validation
			if strings.Contains(result, "..") {
				t.Errorf("Result still contains path traversal: %q", result)
			}
			if strings.Contains(result, "/") || strings.Contains(result, "\\") {
				t.Errorf("Result still contains path separators: %q", result)
			}
		})
	}
}

func TestSanitizeFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "default-filename",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n  ",
			expected: "safe-filename",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "at-$%^-and",
		},
		{
			name:     "only hyphens",
			input:    "-----",
			expected: "safe-filename",
		},
		{
			name:     "unicode characters",
			input:    "Héllo Wörld",
			expected: "Héllo-Wörld",
		},
		{
			name:     "emoji in subject",
			input:    "Meeting 📅 Tomorrow",
			expected: "Meeting-📅-Tomorrow",
		},
		{
			name:     "multiple consecutive spaces",
			input:    "Test     Subject",
			expected: "Test-Subject",
		},
		{
			name:     "mixed case",
			input:    "CamelCase Subject",
			expected: "CamelCase-Subject",
		},
		{
			name:     "numbers and letters",
			input:    "Test123 Subject456",
			expected: "Test123-Subject456",
		},
		{
			name:     "leading and trailing special chars",
			input:    "!!!Important Subject!!!",
			expected: "Important-Subject",
		},
		{
			name:     "very long filename",
			input:    strings.Repeat("Test ", 50),          // 250 chars
			expected: strings.Repeat("Test-", 15) + "Test", // Should be trimmed to ~80 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Ensure result is within length limits
			if len(result) > 80 {
				t.Errorf("Result too long: %d chars, expected <= 80", len(result))
			}

			// Ensure result is never empty
			if result == "" {
				t.Errorf("Result should never be empty for input %q", tt.input)
			}
		})
	}
}

func TestSanitizeThreadSubject_WithThreadID(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		threadID string
		expected string
	}{
		{
			name:     "normal subject with thread ID",
			subject:  "Important Meeting",
			threadID: "abc123",
			expected: "Important-Meeting",
		},
		{
			name:     "empty subject with thread ID",
			subject:  "",
			threadID: "xyz789",
			expected: "email-thread-xyz789",
		},
		{
			name:     "subject with Re: prefix",
			subject:  "Re: Follow up",
			threadID: "def456",
			expected: "Follow-up",
		},
		{
			name:     "subject becomes unsafe after sanitization",
			subject:  "!!!@@@###",
			threadID: "ghi789",
			expected: "at-at-at",
		},
		{
			name:     "empty subject, empty thread ID",
			subject:  "",
			threadID: "",
			expected: "email-thread",
		},
		{
			name:     "thread ID with unsafe characters",
			subject:  "Test",
			threadID: "thread/../123",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeThreadSubject(tt.subject, tt.threadID)
			if result != tt.expected {
				t.Errorf("SanitizeThreadSubject(%q, %q) = %q, want %q",
					tt.subject, tt.threadID, result, tt.expected)
			}
		})
	}
}

func TestCleanEmailSubject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove Re: prefix",
			input:    "Re: Original Subject",
			expected: "Original Subject",
		},
		{
			name:     "remove Fwd: prefix",
			input:    "Fwd: Important Message",
			expected: "Important Message",
		},
		{
			name:     "remove multiple prefixes",
			input:    "RE: Fwd: Re: Final Subject",
			expected: "Final Subject",
		},
		{
			name:     "case insensitive removal",
			input:    "FWD: Test Subject",
			expected: "Test Subject",
		},
		{
			name:     "no prefix to remove",
			input:    "Clean Subject",
			expected: "Clean Subject",
		},
		{
			name:     "empty after prefix removal",
			input:    "Re:",
			expected: "",
		},
		{
			name:     "whitespace handling",
			input:    "  Re:   Fwd:   Subject  ",
			expected: "Subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanEmailSubject(tt.input)
			if result != tt.expected {
				t.Errorf("cleanEmailSubject(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename_Performance(t *testing.T) {
	// Test with very long string
	longInput := strings.Repeat("Test Subject with Many Words ", 100)
	result := SanitizeFilename(longInput)

	if len(result) > 80 {
		t.Errorf("Long input result should be truncated to 80 chars, got %d", len(result))
	}

	// Test multiple consecutive hyphens (worst case for performance)
	manyHyphens := strings.Repeat("-", 1000)
	result = SanitizeFilename(manyHyphens)

	if result != "safe-filename" {
		t.Errorf("Many hyphens should result in fallback, got %q", result)
	}
}

func TestSanitizeFilename_Consistency(t *testing.T) {
	// Test that the same input always produces the same output
	testCases := []string{
		"Test Subject",
		"../../../etc/passwd",
		"!!!@@@###",
		"",
		"Very Long " + strings.Repeat("Subject ", 20),
	}

	for _, input := range testCases {
		first := SanitizeFilename(input)
		for i := 0; i < 10; i++ {
			result := SanitizeFilename(input)
			if result != first {
				t.Errorf("Inconsistent results for input %q: first=%q, iteration %d=%q",
					input, first, i, result)
			}
		}
	}
}
