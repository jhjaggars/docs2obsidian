package transform

import (
	"fmt"
	"log"
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
		cleanedContent := t.cleanupContent(item.Content)
		cleanedTitle := t.cleanupTitle(item.Title)

		if cleanedContent != item.Content || cleanedTitle != item.Title {
			// Only copy if changes were made
			transformedItem := *item
			transformedItem.Content = cleanedContent
			transformedItem.Title = cleanedTitle
			transformedItems[i] = &transformedItem
		} else {
			// No changes, so no copy
			transformedItems[i] = item
		}
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
		log.Printf("Warning: invalid tagging rule format, expected map[string]interface{}, got %T", ruleInterface)

		return nil
	}

	rule := TaggingRule{}

	pattern, hasPattern := ruleMap["pattern"]
	if !hasPattern {
		log.Printf("Warning: tagging rule missing required 'pattern' field")

		return nil
	}

	patternStr, ok := pattern.(string)
	if !ok {
		log.Printf("Warning: tagging rule 'pattern' must be a string, got %T", pattern)

		return nil
	}

	rule.Pattern = patternStr

	if tagsInterface, exists := ruleMap["tags"]; exists {
		tagsSlice, ok := tagsInterface.([]interface{})
		if !ok {
			log.Printf("Warning: tagging rule 'tags' must be an array, got %T", tagsInterface)

			return nil
		}

		rule.Tags = make([]string, 0, len(tagsSlice))
		for i, tagInterface := range tagsSlice {
			if tag, ok := tagInterface.(string); ok {
				rule.Tags = append(rule.Tags, tag)
			} else {
				log.Printf("Warning: tagging rule tag[%d] must be a string, got %T", i, tagInterface)
			}
		}
	}

	return &rule
}

func (t *AutoTaggingTransformer) Transform(items []*models.Item) ([]*models.Item, error) {
	transformedItems := make([]*models.Item, len(items))

	for i, item := range items {
		newTags := t.applyTaggingRules(item)

		if len(newTags) > 0 {
			// Copy-on-write: only copy if there are new tags to add
			transformedItem := *item

			existingTags := make(map[string]bool)
			for _, tag := range transformedItem.Tags {
				existingTags[tag] = true
			}

			for _, newTag := range newTags {
				if !existingTags[newTag] {
					transformedItem.Tags = append(transformedItem.Tags, newTag)
				}
			}

			transformedItems[i] = &transformedItem
		} else {
			// No new tags, so no copy
			transformedItems[i] = item
		}
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

	minContentLength, err := t.getMinContentLength()
	if err != nil {
		return nil, err
	}

	excludeSourceTypes, err := t.getExcludeSourceTypes()
	if err != nil {
		return nil, err
	}

	requiredTags, err := t.getRequiredTags()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if t.shouldIncludeItem(item, minContentLength, excludeSourceTypes, requiredTags) {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

func (t *FilterTransformer) getMinContentLength() (int, error) {
	if val, exists := t.config["min_content_length"]; exists {
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return 0, fmt.Errorf("invalid type for min_content_length: expected int, got %T", v)
		}
	}

	return 0, nil
}

func (t *FilterTransformer) getExcludeSourceTypes() ([]string, error) {
	val, exists := t.config["exclude_source_types"]
	if !exists {
		return nil, nil
	}

	types, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid type for exclude_source_types: expected array, got %T", val)
	}

	result := make([]string, 0, len(types))
	for i, typeInterface := range types {
		if sourceType, ok := typeInterface.(string); ok {
			result = append(result, sourceType)
		} else {
			return nil, fmt.Errorf("invalid type for exclude_source_types[%d]: expected string, got %T", i, typeInterface)
		}
	}

	return result, nil
}

func (t *FilterTransformer) getRequiredTags() ([]string, error) {
	val, exists := t.config["required_tags"]
	if !exists {
		return nil, nil
	}

	tags, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid type for required_tags: expected array, got %T", val)
	}

	result := make([]string, 0, len(tags))
	for i, tagInterface := range tags {
		if tag, ok := tagInterface.(string); ok {
			result = append(result, tag)
		} else {
			return nil, fmt.Errorf("invalid type for required_tags[%d]: expected string, got %T", i, tagInterface)
		}
	}

	return result, nil
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
