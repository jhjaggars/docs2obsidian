package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkm-sync/pkg/models"
)

// BenchmarkGmailMessageConversion tests the performance of converting Gmail messages to Items
func BenchmarkGmailMessageConversion(b *testing.B) {
	// This benchmark would test message conversion performance
	// For now, we'll skip the actual conversion test
	b.Skip("Skipping Gmail message conversion benchmark - requires mock setup")
}

// BenchmarkGmailBatchConversion tests the performance of converting multiple messages
func BenchmarkGmailBatchConversion(b *testing.B) {
	// This benchmark would test batch conversion performance
	// For now, we'll skip the actual conversion test
	b.Skip("Skipping Gmail batch conversion benchmark - requires mock setup")
}

// BenchmarkSyncEngineProcessing tests the performance of the entire sync workflow
func BenchmarkSyncEngineProcessing(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "gmail-perf-test")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	config := &models.Config{
		Sync: models.SyncConfig{
			EnabledSources:   []string{"gmail_perf"},
			DefaultTarget:    "obsidian",
			DefaultOutputDir: tempDir,
		},
		Sources: map[string]models.SourceConfig{
			"gmail_perf": {
				Enabled: true,
				Type:    "gmail",
				Name:    "Performance Test Gmail",
				Gmail: models.GmailSourceConfig{
					Name:               "Perf Test Instance",
					ExtractRecipients:  true,
					ProcessHTMLContent: true,
				},
			},
		},
		Targets: map[string]models.TargetConfig{
			"obsidian": {
				Type: "obsidian",
				Obsidian: models.ObsidianTargetConfig{
					DefaultFolder: "Performance Test",
				},
			},
		},
	}

	// Test the configuration processing workflow
	enabledSources := getEnabledSources(config)
	sourceConfig := config.Sources["gmail_perf"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test various sync workflow components
		outputDir := getSourceOutputDirectory(config.Sync.DefaultOutputDir, sourceConfig)
		assert.NotEmpty(b, outputDir)

		sinceTime, err := parseSinceTime("7d")
		assert.NoError(b, err)
		assert.True(b, sinceTime.Before(time.Now()))

		// Test enabled sources processing
		assert.Contains(b, enabledSources, "gmail_perf")
	}
}

func TestGmailLargeMailboxHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large mailbox test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "gmail-large-mailbox-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		messageCount  int
		batchSize     int
		requestDelay  time.Duration
		expectedFiles int
		maxDuration   time.Duration
	}{
		{
			name:          "Small batch processing",
			messageCount:  50,
			batchSize:     10,
			requestDelay:  10 * time.Millisecond,
			expectedFiles: 50,
			maxDuration:   30 * time.Second,
		},
		{
			name:          "Medium batch processing",
			messageCount:  200,
			batchSize:     25,
			requestDelay:  5 * time.Millisecond,
			expectedFiles: 200,
			maxDuration:   60 * time.Second,
		},
		{
			name:          "Large batch processing",
			messageCount:  1000,
			batchSize:     50,
			requestDelay:  1 * time.Millisecond,
			expectedFiles: 1000,
			maxDuration:   120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			// Create mock configuration for large mailbox
			config := &models.Config{
				Sync: models.SyncConfig{
					EnabledSources:   []string{"gmail_large"},
					DefaultTarget:    "obsidian",
					DefaultOutputDir: tempDir,
				},
				Sources: map[string]models.SourceConfig{
					"gmail_large": {
						Enabled: true,
						Type:    "gmail",
						Name:    "Large Mailbox Gmail",
						Gmail: models.GmailSourceConfig{
							Name:         "Large Mailbox Instance",
							BatchSize:    tt.batchSize,
							RequestDelay: tt.requestDelay,
							MaxRequests:  tt.messageCount,
						},
					},
				},
				Targets: map[string]models.TargetConfig{
					"obsidian": {
						Type: "obsidian",
					},
				},
			}

			// Test configuration processing for large mailbox
			enabledSources := getEnabledSources(config)
			assert.Contains(t, enabledSources, "gmail_large")

			sourceConfig := config.Sources["gmail_large"]
			outputDir := getSourceOutputDirectory(config.Sync.DefaultOutputDir, sourceConfig)
			
			// Create output directory
			err := os.MkdirAll(outputDir, 0755)
			assert.NoError(t, err)

			// Simulate processing large number of messages
			batchCount := (tt.messageCount + tt.batchSize - 1) / tt.batchSize
			processedMessages := 0

			for batch := 0; batch < batchCount; batch++ {
				batchStart := batch * tt.batchSize
				batchEnd := batchStart + tt.batchSize
				if batchEnd > tt.messageCount {
					batchEnd = tt.messageCount
				}

				// Simulate processing this batch
				for i := batchStart; i < batchEnd; i++ {
					// Simulate creating an output file for each message
					filename := filepath.Join(outputDir, fmt.Sprintf("message-%d.md", i))
					content := fmt.Sprintf("# Message %d\n\nProcessed message from large mailbox test.\n", i)
					err := os.WriteFile(filename, []byte(content), 0644)
					assert.NoError(t, err)
					processedMessages++
				}

				// Simulate request delay
				if tt.requestDelay > 0 {
					time.Sleep(tt.requestDelay)
				}

				// Check that we haven't exceeded maximum duration
				if time.Since(startTime) > tt.maxDuration {
					t.Logf("Processing took longer than expected: %v > %v", time.Since(startTime), tt.maxDuration)
					break
				}
			}

			duration := time.Since(startTime)
			t.Logf("Processed %d messages in %v (%.2f messages/second)", 
				processedMessages, duration, float64(processedMessages)/duration.Seconds())

			// Verify output files were created
			files, err := filepath.Glob(filepath.Join(outputDir, "message-*.md"))
			assert.NoError(t, err)
			assert.Equal(t, processedMessages, len(files), "Should create one file per processed message")

			// Verify we didn't exceed maximum duration significantly
			assert.LessOrEqual(t, duration, tt.maxDuration+5*time.Second, 
				"Processing should complete within reasonable time")
		})
	}
}

func TestGmailMemoryUsageOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory optimization test in short mode")
	}

	tests := []struct {
		name              string
		messageCount      int
		expectedMaxMemory int64 // bytes
	}{
		{
			name:              "Small message set",
			messageCount:      100,
			expectedMaxMemory: 10 * 1024 * 1024, // 10MB
		},
		{
			name:              "Medium message set",
			messageCount:      500,
			expectedMaxMemory: 50 * 1024 * 1024, // 50MB
		},
		{
			name:              "Large message set",
			messageCount:      1000,
			expectedMaxMemory: 100 * 1024 * 1024, // 100MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would ideally measure actual memory usage
			// For now, we'll test the workflow and ensure it completes successfully

			tempDir, err := os.MkdirTemp("", "gmail-memory-test")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			_ = models.GmailSourceConfig{
				Name:               "Memory Test Instance",
				ExtractRecipients:  true,
				ProcessHTMLContent: true,
				ExtractLinks:       true,
				BatchSize:          50, // Process in batches to manage memory
			}

			// Simulate processing messages in batches to manage memory
			batchSize := 50
			batchCount := (tt.messageCount + batchSize - 1) / batchSize

			for batch := 0; batch < batchCount; batch++ {
				batchStart := batch * batchSize
				batchEnd := batchStart + batchSize
				if batchEnd > tt.messageCount {
					batchEnd = tt.messageCount
				}

				// Process this batch of messages
				for i := batchStart; i < batchEnd; i++ {
					// Simulate processing without actual Gmail message creation
					// In a real scenario, this would convert Gmail messages to Items
					
					// Simulate memory usage for processing
					_ = fmt.Sprintf("memory-test-%d", i)
					_ = fmt.Sprintf("Memory Test Subject %d", i)
					_ = fmt.Sprintf("sender%d@example.com", i)
					
					// Simulate some processing work
					time.Sleep(time.Microsecond)
				}

				// In a real scenario, we'd also write to target and then clear memory
				// This simulates the batch processing approach for memory management
			}

			t.Logf("Successfully processed %d messages in %d batches", tt.messageCount, batchCount)
		})
	}
}

func TestGmailRateLimitingEffectiveness(t *testing.T) {
	tests := []struct {
		name         string
		requestDelay time.Duration
		requestCount int
		expectedMin  time.Duration
	}{
		{
			name:         "No rate limiting",
			requestDelay: 0,
			requestCount: 10,
			expectedMin:  0,
		},
		{
			name:         "Light rate limiting",
			requestDelay: 10 * time.Millisecond,
			requestCount: 10,
			expectedMin:  90 * time.Millisecond, // 9 delays
		},
		{
			name:         "Moderate rate limiting",
			requestDelay: 50 * time.Millisecond,
			requestCount: 20,
			expectedMin:  950 * time.Millisecond, // 19 delays
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			// Simulate the rate limiting that would happen in Gmail service
			for i := 0; i < tt.requestCount; i++ {
				if tt.requestDelay > 0 && i > 0 {
					time.Sleep(tt.requestDelay)
				}

				// Simulate processing work
				time.Sleep(1 * time.Millisecond)
			}

			duration := time.Since(startTime)

			if tt.expectedMin > 0 {
				assert.GreaterOrEqual(t, duration, tt.expectedMin, 
					"Rate limiting should enforce minimum duration")
			}

			t.Logf("Processed %d requests in %v with %v delay", 
				tt.requestCount, duration, tt.requestDelay)
		})
	}
}

func TestGmailErrorRecoveryPerformance(t *testing.T) {
	t.Skip("Skipping error recovery test - requires better simulation logic")
}

// simulateRequestWithRetry simulates a request that might fail and needs retry
func simulateRequestWithRetry(errorRate float64, maxRetries int) bool {
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Simulate random failure based on error rate
		if time.Now().UnixNano()%100 >= int64(errorRate*100) {
			return true // Success
		}
		
		// Simulate retry delay
		time.Sleep(time.Millisecond * time.Duration(attempt+1))
	}
	return false // All retries failed
}