# Proposal 1: Integrated AI Content Analysis

## Summary

Add AI-powered content analysis directly into pkm-sync as a built-in processing step that runs between data fetching and export, using local language models (Ollama/RamaLama) to summarize, prioritize, and filter content.

## Architecture

```
Source → AI Processor → Target
```

Content flows through an integrated AI processing pipeline within the main pkm-sync binary.

## Technical Implementation

### Core Components

1. **ContentProcessor Interface** (built into main binary)
   ```go
   type ContentProcessor interface {
       Configure(config map[string]interface{}) error
       Process(items []*models.Item) ([]*models.Item, error)
   }
   ```

2. **Enhanced Item Model** with AI fields
   ```go
   type Item struct {
       // Existing fields...
       Summary       string      `json:"summary,omitempty"`
       PriorityScore float64     `json:"priority_score,omitempty"`
       ActionItems   []ActionItem `json:"action_items,omitempty"`
       AITags        []string    `json:"ai_tags,omitempty"`
   }
   ```

3. **Ollama/RamaLama Integration** (internal packages)
   - HTTP client for Ollama API
   - Containerized RamaLama execution
   - Prompt templates for different content types
   - Model management utilities

### Configuration

```yaml
ai_analysis:
  enabled: true
  provider: "ollama"
  model: "llama3.2:3b"
  
  features:
    summarize: true
    extract_actions: true
    priority_scoring: true
    noise_filtering: true
  
  per_source:
    gmail_work:
      max_thread_messages: 10
      ignore_patterns: ["thanks", "automated"]
```

### Sync Flow

```go
func (s *Syncer) Sync(source Source, target Target, options SyncOptions) error {
    // 1. Fetch data from source
    items, err := source.Fetch(options.Since, 0)
    
    // 2. Process with AI (if enabled)
    if s.aiProcessor != nil {
        items, err = s.aiProcessor.Process(items)
    }
    
    // 3. Export to target
    return target.Export(items, options.OutputDir)
}
```

## Key Features

- **Unified experience**: Single binary, single config, single command
- **Deep integration**: AI processor has access to all item metadata and context
- **Optimized performance**: Shared memory, batch processing, connection pooling
- **Consistent error handling**: Unified logging and error reporting
- **Atomic operations**: AI processing success/failure affects entire sync

## Dependencies

- Ollama (required if AI features enabled)
- Local language models (downloaded via Ollama)
- Additional Go dependencies for model communication

## Target Implementation

AI-enhanced items export with rich metadata:

```markdown
---
title: "Meeting: Project Kickoff"
ai_summary: "Team alignment meeting covering timeline, roles, and deliverables"
priority_score: 0.8
action_items:
  - "John to create project timeline by Friday"
  - "Sarah to set up development environment"
ai_tags: ["project-management", "kickoff", "high-priority"]
---

[Enhanced content with AI insights...]
```