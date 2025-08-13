package transform

import (
	"fmt"
	"log"

	"pkm-sync/pkg/interfaces"
	"pkm-sync/pkg/models"
)

// DefaultTransformPipeline implements the TransformPipeline interface.
type DefaultTransformPipeline struct {
	transformers        []interfaces.Transformer
	config              models.TransformConfig
	transformerRegistry map[string]interfaces.Transformer
}

// NewPipeline creates a new transform pipeline.
func NewPipeline() *DefaultTransformPipeline {
	return &DefaultTransformPipeline{
		transformers:        make([]interfaces.Transformer, 0),
		transformerRegistry: make(map[string]interfaces.Transformer),
	}
}

// Configure sets up the pipeline based on configuration.
func (p *DefaultTransformPipeline) Configure(config models.TransformConfig) error {
	p.config = config

	if !config.Enabled {
		return nil
	}

	// Clear existing transformers
	p.transformers = make([]interfaces.Transformer, 0)

	// Add transformers in the specified order
	for _, name := range config.PipelineOrder {
		transformer, exists := p.transformerRegistry[name]
		if !exists {
			return fmt.Errorf("transformer '%s' not found in registry", name)
		}

		// Configure the transformer if config exists
		if transformerConfig, hasConfig := config.Transformers[name]; hasConfig {
			if err := transformer.Configure(transformerConfig); err != nil {
				return fmt.Errorf("failed to configure transformer '%s': %w", name, err)
			}
		}

		p.transformers = append(p.transformers, transformer)
	}

	return nil
}

// AddTransformer adds a transformer to the registry.
func (p *DefaultTransformPipeline) AddTransformer(transformer interfaces.Transformer) error {
	if transformer == nil {
		return fmt.Errorf("transformer cannot be nil")
	}

	name := transformer.Name()
	if name == "" {
		return fmt.Errorf("transformer name cannot be empty")
	}

	p.transformerRegistry[name] = transformer

	return nil
}

// Transform processes items through the configured pipeline.
func (p *DefaultTransformPipeline) Transform(items []*models.Item) ([]*models.Item, error) {
	if !p.config.Enabled || len(p.transformers) == 0 {
		return items, nil
	}

	currentItems := items

	for _, transformer := range p.transformers {
		transformedItems, err := p.processWithErrorHandling(transformer, currentItems)
		if err != nil {
			switch p.config.ErrorStrategy {
			case "fail_fast":
				return nil, fmt.Errorf("transformer '%s' failed: %w", transformer.Name(), err)
			case "log_and_continue":
				log.Printf("Warning: transformer '%s' failed: %v. Continuing with previous items.", transformer.Name(), err)
				// Do not update currentItems, effectively skipping the failed transformer
			case "skip_item":
				log.Printf("Warning: transformer '%s' failed, skipping this batch of items: %v", transformer.Name(), err)

				currentItems = []*models.Item{} // Skip the batch
			default:
				return nil, fmt.Errorf("unknown error strategy '%s'", p.config.ErrorStrategy)
			}
		} else {
			currentItems = transformedItems
		}
	}

	return currentItems, nil
}

// processWithErrorHandling wraps transformer execution with error handling.
func (p *DefaultTransformPipeline) processWithErrorHandling(
	transformer interfaces.Transformer,
	items []*models.Item,
) (processedItems []*models.Item, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Transformer '%s' panicked: %v", transformer.Name(), r)
			err = fmt.Errorf("panic in transformer '%s': %v", transformer.Name(), r)
		}
	}()

	processedItems, err = transformer.Transform(items)

	return
}

// RegisterTransformer is a helper function to register transformers.
func (p *DefaultTransformPipeline) RegisterTransformer(transformer interfaces.Transformer) error {
	return p.AddTransformer(transformer)
}

// GetRegisteredTransformers returns a list of registered transformer names.
func (p *DefaultTransformPipeline) GetRegisteredTransformers() []string {
	names := make([]string, 0, len(p.transformerRegistry))
	for name := range p.transformerRegistry {
		names = append(names, name)
	}

	return names
}
