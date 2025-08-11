package gmail

import (
	"testing"
	"time"

	"pkm-sync/pkg/models"
)

// BenchmarkBuildQuery tests the performance of query building
func BenchmarkBuildQuery(b *testing.B) {
	config := models.GmailSourceConfig{
		Labels:             []string{"IMPORTANT", "STARRED", "WORK"},
		Query:              "has:attachment subject:urgent OR subject:meeting",
		FromDomains:        []string{"company.com", "client.com", "partner.com"},
		ToDomains:          []string{"work.com", "business.com"},
		ExcludeFromDomains: []string{"noreply.com", "spam.com", "notifications.com"},
		IncludeUnread:      true,
		IncludeRead:        false,
		RequireAttachments: true,
		MaxEmailAge:        "30d",
		MinEmailAge:        "1d",
	}
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildQuery(config, since)
	}
}

// BenchmarkBuildComplexQuery tests the performance of complex query building
func BenchmarkBuildComplexQuery(b *testing.B) {
	config := models.GmailSourceConfig{
		Query:  "base:query has:attachment",
		Labels: []string{"IMPORTANT", "STARRED"},
	}
	criteria := map[string]interface{}{
		"from":         "example.com",
		"to":           "recipient.com",
		"subject":      "urgent meeting",
		"is_starred":   true,
		"is_important": true,
		"newer_than":   "1d",
		"older_than":   "7d",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildComplexQuery(config, criteria)
	}
}

// BenchmarkParseDuration tests the performance of duration parsing
func BenchmarkParseDuration(b *testing.B) {
	durations := []string{"30d", "1y", "2w", "12h", "45m"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range durations {
			_, _ = parseDuration(d)
		}
	}
}

// BenchmarkQueryValidation tests the performance of query validation
func BenchmarkQueryValidation(b *testing.B) {
	queries := []string{
		"from:example.com",
		"(from:example.com OR to:test.com)",
		"((from:example.com AND to:test.com) OR subject:urgent)",
		"has:attachment subject:meeting from:company.com",
		"label:IMPORTANT (subject:urgent OR subject:meeting) -from:noreply.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, query := range queries {
			_ = ValidateQuery(query)
		}
	}
}
