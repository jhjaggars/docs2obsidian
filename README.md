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
# Sync Google Calendar to Obsidian (default)
pkm-sync sync --source google --target obsidian --output ./vault

# Sync to Logseq  
pkm-sync sync --source google --target logseq --output ./graph

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
- ✅ Google Calendar + Drive
- 📋 GMail (planned)
- 📋 Jira (planned)
- 📋 Slack (planned)

### Targets  
- ✅ Obsidian
- ✅ Logseq

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

### Sync last week's calendar to Logseq
```bash
pkm-sync sync --source google --target logseq --since 7d --output ~/Documents/Logseq
```

### Dry run to see what would be synced
```bash
pkm-sync sync --source google --target obsidian --dry-run
```

### Custom output location
```bash
pkm-sync sync --source google --target obsidian --output ~/MyVault/Calendar
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

The application will guide you through the OAuth flow on first run and save your authorization token in the same config directory.

## Command Reference

### New Sync Command
```bash
# General syntax
pkm-sync sync [flags]

# Available flags
--source string     Data source (google) (default "google")
--target string     PKM target (obsidian, logseq) (default "obsidian")
--output string     Output directory (default "./exported")  
--since string      Sync items since (7d, 2006-01-02, today) (default "7d")
--dry-run          Show what would be synced without making changes

# Time formats for --since
--since today      # Today only
--since 7d         # Last 7 days  
--since 2025-01-01 # Specific date
--since 24h        # Last 24 hours
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

## Future Plans

- GMail integration
- Slack integration  
- Jira integration
- Real-time sync with webhooks
- Configuration file support
- Template customization
- Automated scheduling