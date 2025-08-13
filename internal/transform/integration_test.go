package transform

import (
	"net/http"
	"testing"
	"time"

	"pkm-sync/internal/sync"
	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

// MockSource implements interfaces.Source for testing pipeline integration.
type MockSource struct {
	items []*models.Item
}

func (m *MockSource) Name() string {
	return "mock_source"
}

func (m *MockSource) Configure(config map[string]interface{}, client *http.Client) error {
	return nil
}

func (m *MockSource) Fetch(since time.Time, limit int) ([]*models.Item, error) {
	return m.items, nil
}

func (m *MockSource) SupportsRealtime() bool {
	return false
}

// MockTarget implements interfaces.Target for testing pipeline integration.
type MockTarget struct {
	exportedItems []*models.Item
}

func (m *MockTarget) Name() string {
	return "mock_target"
}

func (m *MockTarget) Configure(config map[string]interface{}) error {
	return nil
}

func (m *MockTarget) Export(items []*models.Item, outputDir string) error {
	m.exportedItems = items

	return nil
}

func (m *MockTarget) FormatFilename(title string) string {
	return title + ".md"
}

func (m *MockTarget) GetFileExtension() string {
	return ".md"
}

func (m *MockTarget) FormatMetadata(metadata map[string]interface{}) string {
	return ""
}

func (m *MockTarget) Preview(items []*models.Item, outputDir string) ([]*interfaces.FilePreview, error) {
	return nil, nil
}

// TestPipelineIntegrationWithSyncEngine tests the complete flow from source -> pipeline -> target.
func TestPipelineIntegrationWithSyncEngine(t *testing.T) {
	// Create test items with content that will trigger transformations
	testItems := []*models.Item{
		{
			ID:         "1",
			Title:      "  Re: Important Meeting  ",
			Content:    "  This is about a meeting\n\n\n\nwith urgent details  ",
			SourceType: "test_source",
			ItemType:   "email",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Tags:       []string{"existing"},
			Metadata:   make(map[string]interface{}),
		},
		{
			ID:         "2",
			Title:      "Short note",
			Content:    "Too short",
			SourceType: "test_source",
			ItemType:   "note",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Tags:       []string{},
			Metadata:   make(map[string]interface{}),
		},
	}

	// Create mock source and target
	source := &MockSource{items: testItems}
	target := &MockTarget{}

	// Create and configure the transform pipeline
	pipeline := NewPipeline()

	// Register transformers
	contentCleanup := NewContentCleanupTransformer()
	autoTagging := NewAutoTaggingTransformer()
	filter := NewFilterTransformer()

	pipeline.AddTransformer(contentCleanup)
	pipeline.AddTransformer(autoTagging)
	pipeline.AddTransformer(filter)

	// Configure the pipeline
	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"content_cleanup", "auto_tagging", "filter"},
		ErrorStrategy: "log_and_continue",
		Transformers: map[string]map[string]interface{}{
			"auto_tagging": {
				"rules": []interface{}{
					map[string]interface{}{
						"pattern": "meeting",
						"tags":    []interface{}{"work", "meeting"},
					},
					map[string]interface{}{
						"pattern": "urgent",
						"tags":    []interface{}{"priority"},
					},
				},
			},
			"filter": {
				"min_content_length": 15,
			},
		},
	}

	err := pipeline.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure pipeline: %v", err)
	}

	// Create syncer with pipeline
	syncer := sync.NewSyncerWithPipeline(pipeline)

	// Execute sync
	syncOptions := interfaces.SyncOptions{
		Since:     time.Now().Add(-24 * time.Hour),
		OutputDir: "/tmp/test",
		DryRun:    false,
		Overwrite: true,
	}

	err = syncer.Sync(source, target, syncOptions)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify the results
	if len(target.exportedItems) != 1 {
		t.Fatalf("Expected 1 item to be exported after filtering, got %d", len(target.exportedItems))
	}

	exportedItem := target.exportedItems[0]

	// Verify content cleanup worked
	if exportedItem.Title != "Important Meeting" {
		t.Errorf("Expected cleaned title 'Important Meeting', got '%s'", exportedItem.Title)
	}

	expectedContent := "This is about a meeting\n\nwith urgent details"
	if exportedItem.Content != expectedContent {
		t.Errorf("Expected cleaned content '%s', got '%s'", expectedContent, exportedItem.Content)
	}

	// Verify auto-tagging worked
	tagMap := make(map[string]bool)
	for _, tag := range exportedItem.Tags {
		tagMap[tag] = true
	}

	expectedTags := []string{"existing", "work", "meeting", "source:test_source", "type:email"}
	for _, expectedTag := range expectedTags {
		if !tagMap[expectedTag] {
			t.Errorf("Missing expected tag: %s", expectedTag)
		}
	}

	// Verify filter worked (second item should be filtered out due to short content)
	if len(target.exportedItems) > 1 {
		t.Error("Filter should have removed the short content item")
	}
}

// TestPipelineIntegrationErrorHandling tests that error handling works correctly in the sync engine.
func TestPipelineIntegrationErrorHandling(t *testing.T) {
	testItems := []*models.Item{
		{
			ID:         "1",
			Title:      "Test Item",
			Content:    "Test content",
			SourceType: "test_source",
			ItemType:   "email",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Tags:       []string{},
			Metadata:   make(map[string]interface{}),
		},
	}

	source := &MockSource{items: testItems}
	target := &MockTarget{}

	// Create pipeline with a failing transformer
	pipeline := NewPipeline()
	failingTransformer := &MockTransformer{name: "failing_transformer", shouldFail: true}
	workingTransformer := &MockTransformer{name: "working_transformer"}

	pipeline.AddTransformer(failingTransformer)
	pipeline.AddTransformer(workingTransformer)

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"failing_transformer", "working_transformer"},
		ErrorStrategy: "log_and_continue",
		Transformers:  make(map[string]map[string]interface{}),
	}

	err := pipeline.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure pipeline: %v", err)
	}

	syncer := sync.NewSyncerWithPipeline(pipeline)

	syncOptions := interfaces.SyncOptions{
		Since:     time.Now().Add(-24 * time.Hour),
		OutputDir: "/tmp/test",
		DryRun:    false,
		Overwrite: true,
	}

	// This should not fail despite the failing transformer
	err = syncer.Sync(source, target, syncOptions)
	if err != nil {
		t.Fatalf("Sync should not fail with log_and_continue strategy: %v", err)
	}

	// Verify items were still exported
	if len(target.exportedItems) != 1 {
		t.Fatalf("Expected 1 item to be exported, got %d", len(target.exportedItems))
	}

	// Verify the working transformer processed the item
	exportedItem := target.exportedItems[0]
	hasWorkingTag := false
	hasFailingTag := false

	for _, tag := range exportedItem.Tags {
		if tag == "transformed_by_working_transformer" {
			hasWorkingTag = true
		}

		if tag == "transformed_by_failing_transformer" {
			hasFailingTag = true
		}
	}

	if !hasWorkingTag {
		t.Error("Working transformer should have processed the item")
	}

	if hasFailingTag {
		t.Error("Failing transformer should not have tagged the item")
	}
}
