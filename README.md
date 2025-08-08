# pkm-sync

A universal synchronization tool for Personal Knowledge Management (PKM) systems. Connect data sources like Google Calendar to PKM tools like Obsidian and Logseq.

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
pkm-sync sync                           # Sync using config defaults

# Manual sync with flags (classic approach)
pkm-sync sync --source google --target obsidian --output ./vault

# Multi-source sync (with configuration)
pkm-sync config init --source slack    # Enable additional sources
pkm-sync sync                           # Syncs from all enabled sources

# Sync to Logseq  
pkm-sync sync --target logseq --output ./graph

# Legacy commands still work
pkm-sync calendar  # equivalent to sync with obsidian target
pkm-sync export    # equivalent to sync with obsidian target
```

### Configuration
Same OAuth setup as docs2obsidian - place `credentials.json` in:
- **Linux/Unix**: `~/.config/pkm-sync/credentials.json` (or old path: `~/.config/docs2obsidian/credentials.json`)
- **macOS**: `~/.config/pkm-sync/credentials.json` (or old paths: `~/.config/docs2obsidian/credentials.json`, `~/Library/Application Support/docs2obsidian/credentials.json`)
- **Windows**: `%APPDATA%\pkm-sync\credentials.json` (or old path: `%APPDATA%\docs2obsidian\credentials.json`)

## Supported Integrations

### Sources
- ✅ **Google Calendar + Drive** - Fully implemented
- 📋 **Slack** - Configuration ready, implementation pending
- 📋 **Gmail** - Configuration ready, implementation pending  
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
pkm-sync export

# New capabilities
pkm-sync sync --target logseq
```

## Examples

### Multi-Source Sync Examples
```bash
# Sync from all enabled sources (Google + Slack + Gmail) to Obsidian
pkm-sync config init --source google --source slack
pkm-sync sync

# This will output: "Syncing from sources [google, slack] to obsidian"
```

### Single Source Examples  
```bash
# Sync last week's calendar to Logseq
pkm-sync sync --source google --target logseq --since 7d --output ~/Documents/Logseq

# Dry run to see what would be synced from all enabled sources
pkm-sync sync --dry-run

# Dry run for specific source only
pkm-sync sync --source google --target obsidian --dry-run

# Custom output location
pkm-sync sync --output ~/MyVault/Calendar
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

# Then just sync without flags
pkm-sync sync           # Uses your configured defaults
pkm-sync sync --since today  # Override just the time range
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

## Command Reference

### Sync Command (with Multi-Source Support)
```bash
# Use config defaults - syncs from ALL enabled sources
pkm-sync sync

# Override to sync from specific source only
pkm-sync sync --source google

# Override other settings
pkm-sync sync --target logseq --since today
pkm-sync sync --output ./custom-output --dry-run

# Multi-source example output:
# "Syncing from sources [google, slack] to obsidian"

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
