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
- **Universal interfaces** (`pkg/interfaces/`) for Source and Target abstractions
- **Universal data model** (`pkg/models/item.go`) for consistent data representation
- **Source implementations** in `internal/sources/` (Google Calendar, Gmail, Drive)
- **Target implementations** in `internal/targets/` (Obsidian, Logseq)
- **Sync engine** (`internal/sync/`) handles data pipeline

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
3. **Source tagging**: Optional tags added to identify data source
4. **Target export**: Items formatted and exported according to target type
5. **Single output directory**: All targets use `sync.default_output_dir`

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

# Install Git hooks for development
./scripts/install-hooks.sh
```

### Git Hooks
The repository includes development Git hooks to maintain code quality:

- **pre-commit**: Automatically runs `go fmt` on staged Go files before each commit
- Ensures consistent code formatting across all contributions
- Prevents commits with incorrectly formatted Go code

To install hooks after cloning:
```bash
./scripts/install-hooks.sh
```

The pre-commit hook will:
1. Detect staged Go files (`.go` extension)
2. Run `go fmt` on each staged file
3. Re-add formatted files to the staging area
4. Allow the commit to proceed

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