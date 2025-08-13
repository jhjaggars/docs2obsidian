package transform

import (
	"testing"
	"time"

	"pkm-sync/pkg/models"
)

func TestContentCleanupTransformer(t *testing.T) {
	transformer := NewContentCleanupTransformer()

	if transformer.Name() != "content_cleanup" {
		t.Errorf("Expected name 'content_cleanup', got '%s'", transformer.Name())
	}

	items := []*models.Item{
		{
			ID:      "1",
			Title:   "  Re: Test Subject  ",
			Content: "  Test content\n\n\n\nwith extra newlines\r\n  ",
		},
	}

	result, err := transformer.Transform(items)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	item := result[0]
	if item.Title != "Test Subject" {
		t.Errorf("Expected cleaned title 'Test Subject', got '%s'", item.Title)
	}

	expectedContent := "Test content\n\nwith extra newlines"
	if item.Content != expectedContent {
		t.Errorf("Expected cleaned content '%s', got '%s'", expectedContent, item.Content)
	}
}

func TestContentCleanupTransformerConfigure(t *testing.T) {
	transformer := NewContentCleanupTransformer()

	config := map[string]interface{}{
		"test_setting": "test_value",
	}

	err := transformer.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if transformer.config["test_setting"] != "test_value" {
		t.Error("Configuration not properly stored")
	}
}

func TestAutoTaggingTransformer(t *testing.T) {
	transformer := NewAutoTaggingTransformer()

	if transformer.Name() != "auto_tagging" {
		t.Errorf("Expected name 'auto_tagging', got '%s'", transformer.Name())
	}

	// Configure with tagging rules
	config := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"pattern": "meeting",
				"tags":    []interface{}{"work", "meeting"},
			},
			map[string]interface{}{
				"pattern": "urgent",
				"tags":    []interface{}{"priority", "urgent"},
			},
		},
	}

	err := transformer.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	items := []*models.Item{
		{
			ID:         "1",
			Title:      "Urgent meeting tomorrow",
			Content:    "Important meeting discussion",
			SourceType: "gmail",
			ItemType:   "email",
			Tags:       []string{"existing"},
		},
	}

	result, err := transformer.Transform(items)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	item := result[0]

	tagMap := make(map[string]bool)
	for _, tag := range item.Tags {
		tagMap[tag] = true
	}

	expectedTags := []string{"existing", "work", "meeting", "priority", "urgent", "source:gmail", "type:email"}
	for _, expectedTag := range expectedTags {
		if !tagMap[expectedTag] {
			t.Errorf("Missing expected tag: %s", expectedTag)
		}
	}
}

func TestAutoTaggingTransformerNoDuplicates(t *testing.T) {
	transformer := NewAutoTaggingTransformer()

	items := []*models.Item{
		{
			ID:         "1",
			Title:      "Test",
			Content:    "Test content",
			SourceType: "gmail",
			ItemType:   "email",
			Tags:       []string{"source:gmail"}, // Already has this tag
		},
	}

	result, err := transformer.Transform(items)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	item := result[0]

	// Count occurrences of "source:gmail"
	count := 0

	for _, tag := range item.Tags {
		if tag == "source:gmail" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 occurrence of 'source:gmail', got %d", count)
	}
}

func TestFilterTransformer(t *testing.T) {
	transformer := NewFilterTransformer()

	if transformer.Name() != "filter" {
		t.Errorf("Expected name 'filter', got '%s'", transformer.Name())
	}

	config := map[string]interface{}{
		"min_content_length":   10,
		"exclude_source_types": []interface{}{"spam"},
		"required_tags":        []interface{}{"important"},
	}

	err := transformer.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	items := []*models.Item{
		{
			ID:         "1",
			Title:      "Valid item",
			Content:    "This content is long enough",
			SourceType: "gmail",
			Tags:       []string{"important"},
		},
		{
			ID:         "2",
			Title:      "Too short",
			Content:    "Short",
			SourceType: "gmail",
			Tags:       []string{"important"},
		},
		{
			ID:         "3",
			Title:      "Spam item",
			Content:    "This content is long enough",
			SourceType: "spam",
			Tags:       []string{"important"},
		},
		{
			ID:         "4",
			Title:      "Missing tag",
			Content:    "This content is long enough",
			SourceType: "gmail",
			Tags:       []string{},
		},
	}

	result, err := transformer.Transform(items)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 filtered item, got %d", len(result))
	}

	if result[0].ID != "1" {
		t.Errorf("Expected item ID '1', got '%s'", result[0].ID)
	}
}

func TestFilterTransformerNoFilters(t *testing.T) {
	transformer := NewFilterTransformer()

	// No configuration, should pass all items
	items := []*models.Item{
		{ID: "1", Title: "Item 1", Content: "A"},
		{ID: "2", Title: "Item 2", Content: "B"},
	}

	result, err := transformer.Transform(items)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
}

func TestGetAllExampleTransformers(t *testing.T) {
	transformers := GetAllExampleTransformers()

	if len(transformers) != 3 {
		t.Errorf("Expected 3 example transformers, got %d", len(transformers))
	}

	expectedNames := map[string]bool{
		"content_cleanup": false,
		"auto_tagging":    false,
		"filter":          false,
	}

	for _, transformer := range transformers {
		name := transformer.Name()
		if _, exists := expectedNames[name]; exists {
			expectedNames[name] = true
		} else {
			t.Errorf("Unexpected transformer name: %s", name)
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Missing expected transformer: %s", name)
		}
	}
}

func createTestItemWithDetails(id, title, content, sourceType, itemType string, tags []string) *models.Item {
	return &models.Item{
		ID:         id,
		Title:      title,
		Content:    content,
		SourceType: sourceType,
		ItemType:   itemType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tags:       tags,
		Metadata:   make(map[string]interface{}),
	}
}
