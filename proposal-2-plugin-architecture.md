# Proposal 2: Plugin Architecture for Content Processing

## Summary

Implement a plugin system that allows external tools to process content between source fetching and target export, keeping AI analysis and other advanced processing separate from the core pkm-sync binary.

## Architecture

```
Source → [Plugin Chain] → Target
```

Content flows through a configurable chain of external plugins, each running as separate processes or services.

## Technical Implementation

### Core Plugin Interface

```go
// pkg/interfaces/plugin.go
type Plugin interface {
    Name() string
    Version() string
    Configure(config map[string]interface{}) error
    Process(items []*models.Item) ([]*models.Item, error)
    SupportsItemType(itemType string) bool
    Metadata() PluginMetadata
}

type PluginMetadata struct {
    Description  string
    Author       string
    Dependencies []string
    Categories   []string  // "ai", "filter", "transform"
}
```

### Plugin Communication Methods

#### Option 1: JSON over stdin/stdout
```bash
# Simple, language-agnostic communication
echo '{"items": [...], "config": {...}}' | ./plugins/ai-summarizer
```

#### Option 2: HTTP API
```go
// Plugin runs as microservice
POST /process
Content-Type: application/json

{
  "items": [...],
  "config": {...}
}
```

#### Option 3: Go Plugin System
```go
// For Go plugins only - more performant
//go:build plugin

func Process(items []*models.Item, config map[string]interface{}) ([]*models.Item, error) {
    // Plugin logic here
}
```

### Configuration

```yaml
sync:
  plugins:
    enabled: true
    
    # Plugin chain execution order
    chain:
      - name: "deduplicator"
        type: "builtin"
        config:
          strategy: "content_hash"
      
      - name: "ai-summarizer"
        type: "external"
        binary: "./plugins/pkm-ai-summarizer"
        config:
          model: "llama3.2:3b"
          endpoint: "http://localhost:11434"
          features:
            summarize: true
            extract_actions: true
        
      - name: "sentiment-analyzer"
        type: "http"
        endpoint: "http://localhost:8080"
        config:
          threshold: 0.7

# Per-source plugin overrides
sources:
  gmail_work:
    plugins:
      chain:
        - name: "work-filter"
          config:
            include_internal: true
        - name: "ai-summarizer"
          config:
            priority_keywords: ["urgent", "deadline", "action required"]
```

### Plugin Discovery & Management

```bash
# Plugin management commands
pkm-sync plugin list                     # Show installed plugins
pkm-sync plugin search ai                # Search plugin registry
pkm-sync plugin install ai-summarizer    # Install plugin
pkm-sync plugin verify                   # Check plugin dependencies
pkm-sync plugin enable ai-summarizer     # Enable/disable plugins
```

### Plugin Registry Structure

```
~/.config/pkm-sync/plugins/
├── registry.yaml                 # Available plugins metadata
├── installed/
│   ├── ai-summarizer/
│   │   ├── binary
│   │   ├── config.yaml
│   │   └── metadata.yaml
│   └── sentiment-analyzer/
│       ├── service.py
│       └── requirements.txt
└── cache/                        # Downloaded plugins
```

### Sync Flow with Plugins

```go
func (s *Syncer) Sync(source Source, target Target, options SyncOptions) error {
    // 1. Fetch data from source
    items, err := source.Fetch(options.Since, 0)
    if err != nil {
        return err
    }
    
    // 2. Run plugin chain
    processedItems, err := s.pluginRunner.RunChain(items, options.PluginConfig)
    if err != nil {
        return err
    }
    
    // 3. Export to target
    return target.Export(processedItems, options.OutputDir)
}
```

### Example Plugin Implementation

```go
// plugins/ai-summarizer/main.go
package main

import (
    "encoding/json"
    "os"
    "pkm-sync/pkg/models"
)

type Request struct {
    Items  []*models.Item            `json:"items"`
    Config map[string]interface{}    `json:"config"`
}

type Response struct {
    Items []*models.Item `json:"items"`
    Error string         `json:"error,omitempty"`
}

func main() {
    var req Request
    json.NewDecoder(os.Stdin).Decode(&req)
    
    // Process items with AI
    processedItems := processWithOllama(req.Items, req.Config)
    
    resp := Response{Items: processedItems}
    json.NewEncoder(os.Stdout).Encode(resp)
}
```

## Plugin Types

### Built-in Plugins (shipped with pkm-sync)
- **deduplicator**: Remove duplicate items
- **filter**: Basic content filtering by patterns
- **tagger**: Add tags based on rules
- **sorter**: Sort items by date, priority, etc.

### External Plugins (separate projects)
- **pkm-ai-summarizer**: AI summarization via Ollama
- **pkm-sentiment**: Sentiment analysis
- **pkm-translator**: Content translation
- **pkm-extractor**: Extract structured data (contacts, dates, etc.)
- **pkm-classifier**: Categorize content by topic

## Error Handling & Resilience

```yaml
plugins:
  error_handling:
    strategy: "continue_on_error"  # or "fail_fast"
    timeout: 30s
    retries: 2
    
  fallback:
    ai-summarizer:
      disable_on_failure: true
      fallback_to: "basic-tagger"
```

## Security Considerations

- **Sandboxing**: External plugins run in isolated processes
- **Permission model**: Plugins declare required capabilities
- **Validation**: Plugin input/output validation
- **Registry security**: Signed plugins, checksum verification