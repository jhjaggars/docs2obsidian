# pkm-sync

A universal synchronization tool for Personal Knowledge Management (PKM) systems. Connect data sources like Google Calendar, Gmail, and Drive to PKM tools like Obsidian and Logseq.

## Migrated from docs2obsidian

This tool was formerly known as `docs2obsidian`. All existing functionality is preserved with full backward compatibility.

## Quick Start

### Install
```bash
go build -o pkm-sync ./cmd
```

### Basic Usage
```bash
# Quick start with configuration file (recommended)
pkm-sync config init                    # Create default config
pkm-sync gmail                          # Sync Gmail emails
pkm-sync calendar                       # Sync calendar events
pkm-sync drive                          # Export Google Drive documents

# Manual sync with flags (classic approach)
pkm-sync gmail --source gmail_work --target obsidian --output ./vault
pkm-sync calendar --start 2025-01-01 --end 2025-01-31
pkm-sync drive --event-id 12345 --output ./docs

# Multi-source sync (with configuration)
pkm-sync config init --source gmail_work --source gmail_personal
pkm-sync gmail                          # Syncs from all enabled Gmail sources

# Sync to Logseq  
pkm-sync gmail --target logseq --output ./graph
pkm-sync calendar --target logseq --output ./graph
```

### Configuration
Same OAuth setup as docs2obsidian - place `credentials.json` in:
- **Linux/Unix**: `~/.config/pkm-sync/credentials.json` (or old path: `~/.config/docs2obsidian/credentials.json`)
- **macOS**: `~/.config/pkm-sync/credentials.json` (or old paths: `~/.config/docs2obsidian/credentials.json`, `~/Library/Application Support/docs2obsidian/credentials.json`)
- **Windows**: `%APPDATA%\pkm-sync\credentials.json` (or old path: `%APPDATA%\docs2obsidian\credentials.json`)

## Supported Integrations

### Sources
- ✅ **Gmail** - Fully implemented with multi-instance support and thread grouping
- ✅ **Google Calendar** - Fully implemented
- ✅ **Google Drive** - Fully implemented for document export
- 📋 **Slack** - Configuration ready, implementation pending
- 📋 **Jira** - Configuration ready, implementation pending

### Targets  
- ✅ **Obsidian** - YAML frontmatter, hierarchical structure
- ✅ **Logseq** - Property blocks, flat structure

### Multi-Source Features
- ✅ **Simultaneous sync** from multiple sources
- ✅ **Source-specific tags** for data provenance
- ✅ **Priority-based sync order** (configurable)
- ✅ **Individual source scheduling** (different intervals)
- ✅ **Graceful error handling** (continues if one source fails)

## Migration from docs2obsidian

No changes needed! Your existing setup will continue to work:

```bash
# These still work exactly the same
pkm-sync setup
pkm-sync calendar  
pkm-sync drive

# New capabilities
pkm-sync gmail --target logseq
```

## Examples

### Gmail Sync Examples
```bash
# Sync work emails to Obsidian with thread grouping
pkm-sync gmail --source gmail_work --target obsidian --output ./vault

# Sync personal emails to Logseq
pkm-sync gmail --source gmail_personal --target logseq --since 7d

# Dry run to see what would be synced (includes thread grouping preview)
pkm-sync gmail --source gmail_work --dry-run

# Example output with thread grouping:
# "Found 62 emails from gmail_direct" → "Found 25 emails from gmail_direct"
# Creates files like: Thread-Summary_project-discussion_8-messages.md
```

### Calendar Sync Examples
```bash
# Sync last week's calendar to Obsidian
pkm-sync calendar --start 2025-01-01 --end 2025-01-31 --target obsidian

# List today's events
pkm-sync calendar --start today --end today

# Export calendar with details
pkm-sync calendar --include-details --format json
```

### Drive Export Examples
```bash
# Export docs from specific event
pkm-sync drive --event-id 12345 --output ./docs

# Export docs from date range
pkm-sync drive --start 2025-01-01 --end 2025-01-31 --output ./docs
```

### Multi-Source Configuration Examples
```bash
# Configure multiple Gmail sources
pkm-sync config init --source gmail_work --source gmail_personal
pkm-sync gmail  # Syncs from all enabled Gmail sources

# This will output: "Syncing Gmail from sources [gmail_work, gmail_personal] to obsidian"
```

### Single Source Examples  
```bash
# Sync last week's calendar to Logseq
pkm-sync calendar --start 2025-01-01 --end 2025-01-31 --target logseq --output ~/Documents/Logseq

# Dry run to see what would be synced from all enabled sources
pkm-sync gmail --dry-run

# Dry run for specific source only
pkm-sync gmail --source gmail_work --target obsidian --dry-run

# Custom output location
pkm-sync gmail --output ~/MyVault/Calendar
```

### Configuration-First Workflow
```bash
# Global configuration (persistent across all projects)
pkm-sync config init --target obsidian --output ~/MyVault
pkm-sync config show    # Verify settings

# OR: Local repository configuration (project-specific)
cat > config.yaml << EOF
sync:
  enabled_sources: ["google"]
  default_target: obsidian
  default_output_dir: ./vault
targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: Calendar
EOF

# Then use specific commands
pkm-sync gmail           # Sync Gmail emails
pkm-sync calendar        # Sync calendar events
pkm-sync drive           # Export Google Drive documents
pkm-sync gmail --since today  # Override just the time range
```

## Authentication Setup

### Prerequisites

1. Go 1.24.4 or later
2. A Google Cloud project that you control
3. Access to enable APIs in your Google Cloud project

### OAuth 2.0 Setup

This application uses OAuth 2.0 for Google API authentication:

1. **Create OAuth 2.0 credentials**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable the Google Calendar API and Google Drive API
   - Configure the OAuth consent screen
   - Create OAuth 2.0 Client ID credentials for a "Desktop application"
   - Add `http://127.0.0.1:*` to the authorized redirect URIs (enables automatic authorization)
   - Download the credentials file

2. **Place credentials file**:
   
   **Default locations (checked in order)**:
   - **Linux/Unix**: `~/.config/pkm-sync/credentials.json`
   - **macOS**: `~/.config/pkm-sync/credentials.json` OR `~/Library/Application Support/pkm-sync/credentials.json`
   - **Windows**: `%APPDATA%\pkm-sync\credentials.json`
   - **Fallback**: `./credentials.json` (current directory)
   
   **Backward compatibility**: Old `docs2obsidian` paths are still checked automatically.

3. **Verify setup**:
   ```bash
   pkm-sync setup
   ```

4. **Custom credential location** (optional):
   ```bash
   pkm-sync --credentials /path/to/credentials.json setup
   pkm-sync --config-dir /custom/config/dir setup
   ```

The application will automatically open your browser for authorization and handle the OAuth flow automatically. Your authorization token is saved in the same config directory. If the automatic flow fails, it falls back to manual copy/paste mode.

## Configuration

pkm-sync supports comprehensive configuration through YAML files. See **[CONFIGURATION.md](./CONFIGURATION.md)** for complete configuration documentation.

### Quick Start
```bash
# Create default config file
pkm-sync config init

# View current configuration
pkm-sync config show

# Edit configuration
pkm-sync config edit

# Validate configuration
pkm-sync config validate
```

### Configuration Files
pkm-sync looks for configuration files in this order:
1. Custom directory (`--config-dir` flag)
2. **Global config**: `~/.config/pkm-sync/config.yaml`
3. **Local repository**: `./config.yaml` (current directory)

### Example Configuration
```yaml
sync:
  enabled_sources: ["google"]
  default_target: obsidian
  default_output_dir: ./vault

sources:
  google:
    enabled: true
    type: google
    google:
      calendar_id: primary
      download_docs: true

targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: Calendar
      include_frontmatter: true
```

For complete configuration options including all sources (Google, Slack, Gmail, Jira), targets (Obsidian, Logseq), and advanced settings, see **[CONFIGURATION.md](./CONFIGURATION.md)**.

## Gmail Thread Grouping

Reduce email clutter with intelligent thread grouping:

### Thread Modes
- **`individual`** (default) - Each email as separate file
- **`consolidated`** - All thread messages in one file  
- **`summary`** - Key messages only per thread

### Quick Setup
```yaml
# Add to your config.yaml
sources:
  gmail_work:
    type: gmail
    gmail:
      include_threads: true
      thread_mode: "summary"          # or "consolidated"  
      thread_summary_length: 3        # for summary mode
      query: "in:inbox to:me"
```

### Results
- **Before**: 62 individual email files  
- **After**: 25 grouped items (60% reduction!)
- **Filenames**: No spaces, command-line friendly
  - `Thread-Summary_project-update_5-messages.md`
  - `Thread_meeting-notes-weekly-sync_8-messages.md`

See **[CONFIGURATION.md](./CONFIGURATION.md#gmail-thread-grouping)** for complete thread grouping documentation.

## Command Reference

### Sync Command (with Multi-Source Support)
```bash
# Use config defaults - syncs from ALL enabled sources
pkm-sync gmail

# Override to sync from specific source only
pkm-sync gmail --source gmail_work

# Override other settings
pkm-sync gmail --target logseq --since today
pkm-sync gmail --output ./custom-output --dry-run

# Multi-source example output:
# "Syncing Gmail from sources [gmail_work, gmail_personal] to obsidian"

# Time formats for --since
--since today      # Today only
--since 7d         # Last 7 days  
--since 2025-01-01 # Specific date
--since 24h        # Last 24 hours
```

### Configuration Commands
```bash
pkm-sync config init                    # Create default config
pkm-sync config show                    # Show current config
pkm-sync config path                    # Show config file location
pkm-sync config edit                    # Open config in editor
pkm-sync config validate               # Validate configuration
```

### Legacy Commands (Still Supported)
```bash
pkm-sync setup      # Verify authentication
pkm-sync calendar   # List calendar events  
pkm-sync export     # Export Google Docs from calendar events
```

### Global Flags
```bash
--credentials string    Path to credentials.json file
--config-dir string     Custom configuration directory
```

## Target Differences

### Obsidian Output
- YAML frontmatter with metadata
- Hierarchical file structure support
- Standard markdown format
- Attachments as `[[filename]]` links

### Logseq Output
- Property blocks instead of YAML frontmatter
- Flat file structure (all in output directory)
- Block-based content structure
- Date format: `[[Jan 2nd, 2006]]`
- Tags as `#tagname`

## Troubleshooting

### Common Issues

#### "Error 403: Access denied" or "insufficient authentication scopes"
- Calendar API or Drive API may not be enabled in your Google Cloud project
- Check that both APIs are enabled in the Google Cloud Console
- Verify your OAuth consent screen includes the necessary scopes

#### "credentials.json not found"
- Make sure you've downloaded the OAuth 2.0 credentials from Google Cloud Console
- Save the file in the default config directory (see Authentication Setup section for paths)
- Alternatively, place in current directory as `./credentials.json`
- Verify the file is named exactly `credentials.json` (not `client_secret_*.json`)
- Use `pkm-sync setup` to see which paths are being checked

#### "token refresh failed" or authentication errors
- Your OAuth token may have expired
- Delete the token file from your config directory and re-authenticate
- Token locations: same as credentials.json but named `token.json`
- Run `pkm-sync calendar` to start the OAuth flow again

#### Migration Issues
- Old `docs2obsidian` credentials are automatically found and used
- No manual migration required
- Both old and new config directory paths are checked

### Getting Help
Run `pkm-sync setup` to diagnose authentication issues and get specific guidance.

## Architecture

### Project Structure
```
pkm-sync/
├── cmd/                 # CLI entry points
├── internal/
│   ├── sources/         # Data source interfaces and implementations
│   │   └── google/      # Google Calendar + Drive (migrated code)
│   ├── targets/         # PKM output interfaces and implementations
│   │   ├── obsidian/    # Obsidian-specific formatting
│   │   └── logseq/      # Logseq-specific formatting
│   ├── sync/           # Core synchronization logic
│   └── config/         # Configuration management (enhanced)
├── pkg/
│   ├── models/         # Universal data models
│   └── interfaces/     # Core interfaces
```

### Extensibility
The new architecture makes it easy to add:
- **New Sources**: Implement the `Source` interface
- **New Targets**: Implement the `Target` interface  
- **Custom Sync Logic**: Implement the `Syncer` interface

## Documentation

- **[README.md](./README.md)** - Main documentation and quick start guide
- **[CONFIGURATION.md](./CONFIGURATION.md)** - Complete configuration reference
- **[CLAUDE.md](./CLAUDE.md)** - Development guide for Claude Code
