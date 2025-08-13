package transform

import (
	"strings"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

// ContentCleanupTransformer cleans up and normalizes item content.
type ContentCleanupTransformer struct {
	config map[string]interface{}
}

func NewContentCleanupTransformer() *ContentCleanupTransformer {
	return &ContentCleanupTransformer{
		config: make(map[string]interface{}),
	}
}

func (t *ContentCleanupTransformer) Name() string {
	return "content_cleanup"
}

func (t *ContentCleanupTransformer) Configure(config map[string]interface{}) error {
	t.config = config

	return nil
}

func (t *ContentCleanupTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	transformedItems := make([]*models.Item, len(items))

	for i, item := range items {
		// Create a copy of the item
		transformedItem := *item

		// Clean up content
		transformedItem.Content = t.cleanupContent(item.Content)
		transformedItem.Title = t.cleanupTitle(item.Title)

		transformedItems[i] = &transformedItem
	}

	return transformedItems, nil
}

func (t *ContentCleanupTransformer) cleanupContent(content string) string {
	// Remove excessive whitespace
	content = strings.TrimSpace(content)

	// Replace multiple newlines with double newlines
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	// Remove carriage returns
	content = strings.ReplaceAll(content, "\r", "")

	return content
}

func (t *ContentCleanupTransformer) cleanupTitle(title string) string {
	// Remove excessive whitespace and common prefixes
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "Re: ")
	title = strings.TrimPrefix(title, "Fwd: ")
	title = strings.TrimPrefix(title, "RE: ")
	title = strings.TrimPrefix(title, "FWD: ")

	return title
}

// AutoTaggingTransformer automatically adds tags based on content.
type AutoTaggingTransformer struct {
	config map[string]interface{}
	rules  []TaggingRule
}

type TaggingRule struct {
	Pattern string   `json:"pattern" yaml:"pattern"`
	Tags    []string `json:"tags"    yaml:"tags"`
}

func NewAutoTaggingTransformer() *AutoTaggingTransformer {
	return &AutoTaggingTransformer{
		config: make(map[string]interface{}),
		rules:  make([]TaggingRule, 0),
	}
}

func (t *AutoTaggingTransformer) Name() string {
	return "auto_tagging"
}

func (t *AutoTaggingTransformer) Configure(config map[string]interface{}) error {
	t.config = config

	// Parse tagging rules from config
	rulesInterface, exists := config["rules"]
	if !exists {
		return nil
	}

	rulesSlice, ok := rulesInterface.([]interface{})
	if !ok {
		return nil
	}

	for _, ruleInterface := range rulesSlice {
		rule := t.parseTaggingRule(ruleInterface)
		if rule != nil {
			t.rules = append(t.rules, *rule)
		}
	}

	return nil
}

func (t *AutoTaggingTransformer) parseTaggingRule(ruleInterface interface{}) *TaggingRule {
	ruleMap, ok := ruleInterface.(map[string]interface{})
	if !ok {
		return nil
	}

	rule := TaggingRule{}

	if pattern, ok := ruleMap["pattern"].(string); ok {
		rule.Pattern = pattern
	}

	if tagsInterface, ok := ruleMap["tags"].([]interface{}); ok {
		rule.Tags = make([]string, len(tagsInterface))
		for i, tagInterface := range tagsInterface {
			if tag, ok := tagInterface.(string); ok {
				rule.Tags[i] = tag
			}
		}
	}

	return &rule
}

func (t *AutoTaggingTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	transformedItems := make([]*models.Item, len(items))

	for i, item := range items {
		// Create a copy of the item
		transformedItem := *item

		// Copy existing tags
		existingTags := make([]string, len(item.Tags))
		copy(existingTags, item.Tags)

		// Apply tagging rules
		newTags := t.applyTaggingRules(item)

		// Merge tags (avoiding duplicates)
		tagMap := make(map[string]bool)
		for _, tag := range existingTags {
			tagMap[tag] = true
		}

		for _, tag := range newTags {
			tagMap[tag] = true
		}

		// Convert back to slice
		allTags := make([]string, 0, len(tagMap))
		for tag := range tagMap {
			allTags = append(allTags, tag)
		}

		transformedItem.Tags = allTags
		transformedItems[i] = &transformedItem
	}

	return transformedItems, nil
}

func (t *AutoTaggingTransformer) applyTaggingRules(item *models.Item) []string {
	var newTags []string

	searchText := strings.ToLower(item.Title + " " + item.Content)

	for _, rule := range t.rules {
		if strings.Contains(searchText, strings.ToLower(rule.Pattern)) {
			newTags = append(newTags, rule.Tags...)
		}
	}

	// Add source-based tags
	if item.SourceType != "" {
		newTags = append(newTags, "source:"+item.SourceType)
	}

	if item.ItemType != "" {
		newTags = append(newTags, "type:"+item.ItemType)
	}

	return newTags
}

// FilterTransformer filters items based on criteria.
type FilterTransformer struct {
	config map[string]interface{}
}

func NewFilterTransformer() *FilterTransformer {
	return &FilterTransformer{
		config: make(map[string]interface{}),
	}
}

func (t *FilterTransformer) Name() string {
	return "filter"
}

func (t *FilterTransformer) Configure(config map[string]interface{}) error {
	t.config = config

	return nil
}

func (t *FilterTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	var filteredItems []*models.Item

	minContentLength := t.getMinContentLength()
	excludeSourceTypes := t.getExcludeSourceTypes()
	requiredTags := t.getRequiredTags()

	for _, item := range items {
		if t.shouldIncludeItem(item, minContentLength, excludeSourceTypes, requiredTags) {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

func (t *FilterTransformer) getMinContentLength() int {
	if val, exists := t.config["min_content_length"]; exists {
		if length, ok := val.(int); ok {
			return length
		}

		if length, ok := val.(float64); ok {
			return int(length)
		}
	}

	return 0
}

func (t *FilterTransformer) getExcludeSourceTypes() []string {
	if val, exists := t.config["exclude_source_types"]; exists {
		if types, ok := val.([]interface{}); ok {
			result := make([]string, len(types))
			for i, typeInterface := range types {
				if sourceType, ok := typeInterface.(string); ok {
					result[i] = sourceType
				}
			}

			return result
		}
	}

	return nil
}

func (t *FilterTransformer) getRequiredTags() []string {
	if val, exists := t.config["required_tags"]; exists {
		if tags, ok := val.([]interface{}); ok {
			result := make([]string, len(tags))
			for i, tagInterface := range tags {
				if tag, ok := tagInterface.(string); ok {
					result[i] = tag
				}
			}

			return result
		}
	}

	return nil
}

func (t *FilterTransformer) shouldIncludeItem(
	item *models.Item,
	minContentLength int,
	excludeSourceTypes []string,
	requiredTags []string,
) bool {
	// Check minimum content length
	if len(item.Content) < minContentLength {
		return false
	}

	// Check excluded source types
	for _, excludeType := range excludeSourceTypes {
		if item.SourceType == excludeType {
			return false
		}
	}

	// Check required tags
	if len(requiredTags) > 0 {
		itemTagMap := make(map[string]bool)
		for _, tag := range item.Tags {
			itemTagMap[tag] = true
		}

		for _, requiredTag := range requiredTags {
			if !itemTagMap[requiredTag] {
				return false
			}
		}
	}

	return true
}

// GetAllExampleTransformers returns all example transformers for registration.
func GetAllExampleTransformers() []interfaces.Transformer {
	return []interfaces.Transformer{
		NewContentCleanupTransformer(),
		NewAutoTaggingTransformer(),
		NewFilterTransformer(),
	}
}
