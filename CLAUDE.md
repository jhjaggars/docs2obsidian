# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the application
go build -o pkm-sync ./cmd

# Run the application (requires OAuth setup first)
./pkm-sync setup             # Verify authentication configuration
./pkm-sync gmail             # Sync Gmail emails to PKM systems
./pkm-sync calendar          # List and sync Google Calendar events
./pkm-sync drive             # Export Google Drive documents to markdown

# Configuration management
./pkm-sync config init       # Create default configuration
./pkm-sync config show       # Display current configuration
./pkm-sync config validate   # Validate configuration

# Gmail-specific examples
./pkm-sync gmail --source gmail_work --output ./work-emails
./pkm-sync gmail --since 7d   # Sync last 7 days from all enabled Gmail sources
./pkm-sync gmail --dry-run    # Preview what would be synced

# Custom paths
./pkm-sync --credentials /path/to/credentials.json setup
./pkm-sync --config-dir /custom/config/dir setup

# Development setup
./scripts/install-hooks.sh   # Install Git hooks (pre-commit formatting)
```

## Architecture Overview

This is a Go CLI application that provides universal Personal Knowledge Management (PKM) synchronization. It connects multiple data sources (Google Calendar, Gmail, Drive) to PKM targets (Obsidian, Logseq) using OAuth 2.0 authentication.

### CLI Framework
- Uses **Cobra** for command structure with persistent flags
- Root command (`cmd/root.go`) handles global flags: `--credentials`, `--config-dir`
- Main commands: `gmail`, `calendar`, `drive`, `config`, `setup`
- Global flags are processed in `PersistentPreRun` to configure paths

### Multi-Source Architecture
- **Universal interfaces** (`pkg/interfaces/`) for Source, Target, and Transformer abstractions
- **Universal data model** (`pkg/models/item.go`) for consistent data representation
- **Source implementations** in `internal/sources/` (Google Calendar, Gmail, Drive)
- **Target implementations** in `internal/targets/` (Obsidian, Logseq)
- **Transformer pipeline** (`internal/transform/`) for configurable item processing
- **Sync engine** (`internal/sync/`) handles data pipeline with optional transformations

### Configuration System (`internal/config/config.go`)
- **Multi-source configuration** supporting enabled sources array
- **YAML-based configuration** with comprehensive options
- **Configuration search paths**:
  1. Custom directory (via `--config-dir` flag)
  2. Global config: `~/.config/pkm-sync/config.yaml`
  3. Local repository: `./config.yaml` (current directory)
- **Complete documentation** in `CONFIGURATION.md`

### Authentication Flow
- **OAuth 2.0 only** (no ADC support) with Google Calendar, Drive, and Gmail APIs
- Enhanced copy/paste flow: supports pasting full callback URL or just auth code
- Automatic extraction of auth code from URLs with `extractAuthCode()` function
- Token and credentials stored in platform-specific config directories
- **Gmail requires additional scope**: `gmail.readonly` for email access

### Data Flow
1. **Multi-source collection**: Sync engine iterates through enabled sources
2. **Universal data model**: Each source converts data to common `Item` format
3. **Transform pipeline**: Optional processing chain for item modification, filtering, and enhancement
4. **Source tagging**: Optional tags added to identify data source
5. **Target export**: Items formatted and exported according to target type
6. **Single output directory**: All targets use `sync.default_output_dir`

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/api/calendar/v3` - Google Calendar API
- `google.golang.org/api/drive/v3` - Google Drive API
- `google.golang.org/api/gmail/v1` - Gmail API
- `golang.org/x/oauth2/google` - OAuth 2.0 client
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Current Implementation Status

### Sources
- ✅ **Gmail** - Fully implemented with multi-instance support, advanced filtering, thread grouping, and performance optimizations
- ✅ **Google Calendar** - Fully implemented in `internal/sources/google/`
- ✅ **Google Drive** - Fully implemented for document export
- 🔧 **Slack** - Configuration ready, implementation pending
- 🔧 **Jira** - Configuration ready, implementation pending

### Targets
- ✅ **Obsidian** - Implemented with YAML frontmatter and hierarchical structure
- ✅ **Logseq** - Implemented with property blocks and flat structure

### Configuration Features
- ✅ **Multi-source support** with `enabled_sources` array
- ✅ **Per-source configuration** (intervals, priorities, filtering, output routing)
- ✅ **Multi-instance Gmail** (work, personal, newsletters) with independent configurations
- ✅ **Thread grouping** with configurable modes (individual, consolidated, summary)
- ✅ **Filename sanitization** (no spaces, command-line friendly)
- ✅ **Simplified output directory** structure with per-source subdirectories
- ✅ **Local repository configuration** support
- ✅ **Comprehensive validation** and management commands

## Command Structure

### Core Commands
- **`gmail`** - Sync Gmail emails to PKM systems
  - Supports multiple Gmail instances (work, personal, newsletters)
  - Gmail-specific configuration and filtering
  - Thread grouping: individual, consolidated, or summary modes
  - Example: `pkm-sync gmail --source gmail_work --target obsidian`

- **`calendar`** - List and sync Google Calendar events
  - Calendar-specific functionality
  - Example: `pkm-sync calendar --start 2025-01-01 --end 2025-01-31`

- **`drive`** - Export Google Drive documents to markdown
  - Drive-specific functionality for document export
  - Example: `pkm-sync drive --event-id 12345 --output ./docs`

### Utility Commands
- **`setup`** - Verify authentication configuration
  - Tests all Google services (Calendar, Drive, Gmail)
  - Provides clear error messages and instructions

- **`config`** - Manage configuration files
  - Configuration management and validation

## OAuth Setup Requirements

Users must:
1. Create Google Cloud project with Calendar/Drive/Gmail APIs enabled
2. Configure OAuth consent screen for "Desktop application"
3. Add `http://127.0.0.1:*` to authorized redirect URIs (enables automatic authorization flow)
4. Download credentials.json to config directory or use `--credentials` flag
5. Run `./pkm-sync setup` to verify configuration and complete OAuth flow

**Gmail-specific setup**:
- Enable Gmail API in Google Cloud Console
- Required scopes: `gmail.readonly` (automatically requested)
- Same OAuth credentials work for all Google services

The application uses an automatic web server-based OAuth flow that opens the user's browser and captures the authorization code automatically. If the web server fails, it falls back to the manual copy/paste flow for compatibility.

## Transformer Pipeline System

The transformer pipeline provides a configurable, chainable processing system for items between source fetch and target export. This enables content processing features like filtering, tagging, content cleanup, and future AI analysis.

### Core Architecture
- **Transformer Interface**: Simple `Transform(items) -> items` pattern
- **TransformPipeline**: Chains multiple transformers with configurable error handling
- **Configuration-driven**: Enable/disable transformers and configure processing order
- **Backward compatible**: Zero impact when disabled (default state)

### Configuration Example
```yaml
transformers:
  enabled: true
  pipeline_order: ["content_cleanup", "auto_tagging", "filter"]
  error_strategy: "log_and_continue"  # or "fail_fast", "skip_item"
  transformers:
    content_cleanup:
      strip_prefixes: true
    auto_tagging:
      rules:
        - pattern: "meeting"
          tags: ["work", "meeting"]
        - pattern: "urgent"
          tags: ["priority", "urgent"]
    filter:
      min_content_length: 50
      exclude_source_types: ["spam"]
      required_tags: ["important"]
```

### Built-in Transformers
- **`content_cleanup`**: Normalizes whitespace, removes email prefixes ("Re:", "Fwd:")
- **`auto_tagging`**: Adds tags based on content patterns and source metadata
- **`filter`**: Filters items by content length, source type, required tags

### Error Handling Strategies
- **`fail_fast`**: Stop processing on first transformer error
- **`log_and_continue`**: Log errors but continue with original items
- **`skip_item`**: Log errors and skip problematic items

### Integration Points
- **Sync Engine**: Automatically applies transformations between fetch and export
- **Configuration**: Transformers configured in main config.yaml
- **CLI**: Fully backward compatible - no CLI changes required

### Performance
- **Minimal overhead**: <5% performance impact when enabled
- **Memory efficient**: Processes items in-place where possible
- **Chainable**: Multiple transformers compose efficiently

## Gmail Thread Grouping

The Gmail source supports intelligent thread grouping to reduce email clutter and improve organization.

### Thread Modes
- **`individual`** (default) - Each email is treated as a separate item
- **`consolidated`** - All messages in a thread are combined into a single file
- **`summary`** - Creates summary files with key messages from each thread

### Configuration Example
```yaml
sources:
  gmail_work:
    type: gmail
    gmail:
      include_threads: true           # Enable thread processing
      thread_mode: "summary"          # Use summary mode
      thread_summary_length: 3        # Show 3 key messages per thread
      query: "in:inbox to:me"
```

### Thread Processing Features
- **Smart message selection** - Prioritizes different senders, longer content, attachments
- **Filename sanitization** - No spaces, command-line friendly filenames
- **Thread metadata** - Participants, duration, message count
- **Subject cleaning** - Removes "Re:", "Fwd:" prefixes

### Output Examples
- Consolidated: `Thread_PR-discussion-fix-security-issue_8-messages.md`
- Summary: `Thread-Summary_meeting-notes-weekly-sync_5-messages.md`
- Individual: `Re-Project-status-update.md`

## Development Workflow

### Initial Setup
```bash
# Clone the repository
git clone <repository-url>
cd pkm-sync

# REQUIRED: Install Git hooks for development
./scripts/install-hooks.sh
```

**⚠️ Important:** The pre-commit hook installation is **required** for all contributors. It ensures code quality by running comprehensive checks (formatting, linting, testing, building) before each commit. Commits will be blocked if quality checks fail.

### Git Hooks
The repository includes development Git hooks to maintain code quality:

- **pre-commit**: Comprehensive code quality enforcement before each commit
- Automatically runs `go fmt` on staged Go files
- Executes full CI pipeline (`make ci`) including linting, testing, and building
- Prevents commits if any quality checks fail
- Provides helpful diagnostic commands when failures occur

To install hooks after cloning:
```bash
./scripts/install-hooks.sh
```

The pre-commit hook will:
1. Format any staged Go files with `go fmt`
2. Run the complete CI pipeline (`make ci`):
   - **Lint**: Execute `golangci-lint` with comprehensive rules
   - **Test**: Run all unit tests with race detection
   - **Build**: Verify the project compiles successfully
3. Block the commit if any step fails
4. Provide clear feedback and diagnostic suggestions

**Without pre-commit hooks:** Pull requests may fail CI checks, requiring additional commits to fix formatting and quality issues. Installing hooks prevents this by catching issues locally before they're pushed.

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./cmd
go test ./internal/sources/google/gmail

# Run with verbose output
go test -v ./...
```