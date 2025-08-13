package transform

import (
	"fmt"
	"testing"
	"time"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

// MockTransformer for testing.
type MockTransformer struct {
	name       string
	shouldFail bool
	config     map[string]interface{}
}

// Compile-time check to ensure MockTransformer implements interfaces.Transformer.
var _ interfaces.Transformer = (*MockTransformer)(nil)

func (m *MockTransformer) Name() string {
	return m.name
}

func (m *MockTransformer) Configure(config map[string]interface{}) error {
	m.config = config

	return nil
}

func (m *MockTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock transformer failed")
	}

	// Add a tag to indicate this transformer ran
	transformedItems := make([]*models.Item, len(items))
	for i, item := range items {
		transformedItem := *item
		transformedItem.Tags = append(transformedItem.Tags, "transformed_by_"+m.name)
		transformedItems[i] = &transformedItem
	}

	return transformedItems, nil
}

func TestNewPipeline(t *testing.T) {
	pipeline := NewPipeline()

	if pipeline == nil {
		t.Fatal("NewPipeline() returned nil")
	}

	if len(pipeline.transformers) != 0 {
		t.Errorf("Expected empty transformers slice, got %d transformers", len(pipeline.transformers))
	}

	if len(pipeline.transformerRegistry) != 0 {
		t.Errorf("Expected empty transformer registry, got %d transformers", len(pipeline.transformerRegistry))
	}
}

func TestAddTransformer(t *testing.T) {
	pipeline := NewPipeline()
	transformer := &MockTransformer{name: "test_transformer"}

	err := pipeline.AddTransformer(transformer)
	if err != nil {
		t.Fatalf("AddTransformer() failed: %v", err)
	}

	if len(pipeline.transformerRegistry) != 1 {
		t.Errorf("Expected 1 transformer in registry, got %d", len(pipeline.transformerRegistry))
	}

	if pipeline.transformerRegistry["test_transformer"] != transformer {
		t.Error("Transformer not properly registered")
	}
}

func TestAddTransformerNil(t *testing.T) {
	pipeline := NewPipeline()

	err := pipeline.AddTransformer(nil)
	if err == nil {
		t.Error("Expected error when adding nil transformer")
	}
}

func TestAddTransformerEmptyName(t *testing.T) {
	pipeline := NewPipeline()
	transformer := &MockTransformer{name: ""}

	err := pipeline.AddTransformer(transformer)
	if err == nil {
		t.Error("Expected error when adding transformer with empty name")
	}
}

func TestConfigure(t *testing.T) {
	pipeline := NewPipeline()
	transformer1 := &MockTransformer{name: "transformer1"}
	transformer2 := &MockTransformer{name: "transformer2"}

	pipeline.AddTransformer(transformer1)
	pipeline.AddTransformer(transformer2)

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"transformer1", "transformer2"},
		ErrorStrategy: "fail_fast",
		Transformers: map[string]map[string]interface{}{
			"transformer1": {"setting1": "value1"},
			"transformer2": {"setting2": "value2"},
		},
	}

	err := pipeline.Configure(config)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	if len(pipeline.transformers) != 2 {
		t.Errorf("Expected 2 transformers in pipeline, got %d", len(pipeline.transformers))
	}

	if pipeline.transformers[0].Name() != "transformer1" {
		t.Errorf("Expected first transformer to be 'transformer1', got '%s'", pipeline.transformers[0].Name())
	}

	if pipeline.transformers[1].Name() != "transformer2" {
		t.Errorf("Expected second transformer to be 'transformer2', got '%s'", pipeline.transformers[1].Name())
	}
}

func TestConfigureDisabled(t *testing.T) {
	pipeline := NewPipeline()
	transformer := &MockTransformer{name: "test_transformer"}
	pipeline.AddTransformer(transformer)

	config := models.TransformConfig{
		Enabled:       false,
		PipelineOrder: []string{"test_transformer"},
		ErrorStrategy: "fail_fast",
	}

	err := pipeline.Configure(config)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	if len(pipeline.transformers) != 0 {
		t.Errorf("Expected 0 transformers when disabled, got %d", len(pipeline.transformers))
	}
}

func TestConfigureUnknownTransformer(t *testing.T) {
	pipeline := NewPipeline()

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"unknown_transformer"},
		ErrorStrategy: "fail_fast",
	}

	err := pipeline.Configure(config)
	if err == nil {
		t.Error("Expected error when configuring unknown transformer")
	}
}

func TestTransformDisabled(t *testing.T) {
	pipeline := NewPipeline()

	config := models.TransformConfig{
		Enabled: false,
	}
	pipeline.Configure(config)

	items := []*models.Item{
		{ID: "1", Title: "Test Item"},
	}

	result, err := pipeline.Transform(items)
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}

	if result[0] != items[0] {
		t.Error("Expected same item reference when disabled")
	}
}

func TestTransformSuccess(t *testing.T) {
	pipeline := NewPipeline()
	transformer1 := &MockTransformer{name: "transformer1"}
	transformer2 := &MockTransformer{name: "transformer2"}

	pipeline.AddTransformer(transformer1)
	pipeline.AddTransformer(transformer2)

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"transformer1", "transformer2"},
		ErrorStrategy: "fail_fast",
	}
	pipeline.Configure(config)

	items := []*models.Item{
		{ID: "1", Title: "Test Item", Tags: []string{}},
	}

	result, err := pipeline.Transform(items)
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}

	item := result[0]
	if len(item.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item.Tags))
	}

	expectedTags := map[string]bool{
		"transformed_by_transformer1": true,
		"transformed_by_transformer2": true,
	}

	for _, tag := range item.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestTransformFailFast(t *testing.T) {
	pipeline := NewPipeline()
	transformer1 := &MockTransformer{name: "transformer1"}
	transformer2 := &MockTransformer{name: "transformer2", shouldFail: true}

	pipeline.AddTransformer(transformer1)
	pipeline.AddTransformer(transformer2)

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"transformer1", "transformer2"},
		ErrorStrategy: "fail_fast",
	}
	pipeline.Configure(config)

	items := []*models.Item{
		{ID: "1", Title: "Test Item", Tags: []string{}},
	}

	_, err := pipeline.Transform(items)
	if err == nil {
		t.Error("Expected error with fail_fast strategy")
	}
}

func TestTransformLogAndContinue(t *testing.T) {
	pipeline := NewPipeline()
	transformer1 := &MockTransformer{name: "transformer1"}
	transformer2 := &MockTransformer{name: "transformer2", shouldFail: true}
	transformer3 := &MockTransformer{name: "transformer3"}

	pipeline.AddTransformer(transformer1)
	pipeline.AddTransformer(transformer2)
	pipeline.AddTransformer(transformer3)

	config := models.TransformConfig{
		Enabled:       true,
		PipelineOrder: []string{"transformer1", "transformer2", "transformer3"},
		ErrorStrategy: "log_and_continue",
	}
	pipeline.Configure(config)

	items := []*models.Item{
		{ID: "1", Title: "Test Item", Tags: []string{}},
	}

	result, err := pipeline.Transform(items)
	if err != nil {
		t.Fatalf("Transform() failed with log_and_continue: %v", err)
	}

	// Should have tags from transformer1 and transformer3, but not transformer2
	item := result[0]
	expectedTags := map[string]bool{
		"transformed_by_transformer1": false,
		"transformed_by_transformer3": false,
	}

	for _, tag := range item.Tags {
		if tag == "transformed_by_transformer1" || tag == "transformed_by_transformer3" {
			expectedTags[tag] = true
		}

		if tag == "transformed_by_transformer2" {
			t.Errorf("Should not have tag from failed transformer: %s", tag)
		}
	}

	for tag, found := range expectedTags {
		if !found {
			t.Errorf("Missing expected tag: %s", tag)
		}
	}
}

func TestGetRegisteredTransformers(t *testing.T) {
	pipeline := NewPipeline()
	transformer1 := &MockTransformer{name: "transformer1"}
	transformer2 := &MockTransformer{name: "transformer2"}

	pipeline.AddTransformer(transformer1)
	pipeline.AddTransformer(transformer2)

	names := pipeline.GetRegisteredTransformers()

	if len(names) != 2 {
		t.Errorf("Expected 2 transformer names, got %d", len(names))
	}

	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["transformer1"] || !nameMap["transformer2"] {
		t.Error("Missing expected transformer names")
	}
}

func createTestItem(id, title, content string) *models.Item {
	return &models.Item{
		ID:         id,
		Title:      title,
		Content:    content,
		SourceType: "test",
		ItemType:   "test_item",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tags:       []string{},
		Metadata:   make(map[string]interface{}),
	}
}
