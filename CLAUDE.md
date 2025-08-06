# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the application
go build -o docs2obsidian ./cmd

# Run the application (requires OAuth setup first)
docs2obsidian setup           # Verify authentication configuration
docs2obsidian calendar        # List upcoming calendar events

# Custom credential paths
docs2obsidian --credentials /path/to/credentials.json setup
docs2obsidian --config-dir /custom/config/dir setup
```

## Architecture Overview

This is a Go CLI application that integrates Google Calendar and Google Drive with Obsidian notes using OAuth 2.0 authentication.

### CLI Framework
- Uses **Cobra** for command structure with persistent flags
- Root command (`cmd/root.go`) handles global flags: `--credentials`, `--config-dir`
- Subcommands: `setup` (auth verification) and `calendar` (event listing)
- Global flags are processed in `PersistentPreRun` to configure paths

### Authentication Flow
- **OAuth 2.0 only** (no ADC support) with Google Calendar and Drive APIs
- Enhanced copy/paste flow: supports pasting full callback URL or just auth code
- Automatic extraction of auth code from URLs with `extractAuthCode()` function
- Token and credentials stored in platform-specific config directories

### Configuration Management (`internal/config/paths.go`)
- Cross-platform config directory resolution:
  - **Linux/Unix**: `~/.config/docs2obsidian/`
  - **macOS**: `~/.config/docs2obsidian/` (preferred) OR `~/Library/Application Support/docs2obsidian/` (fallback)
  - **Windows**: `%APPDATA%\docs2obsidian\`
- Multi-path credential search with `FindCredentialsFile()` checking custom path → config dirs → current directory
- Global state management for custom paths via `SetCustomCredentialsPath()` and `SetCustomConfigDir()`

### Google API Integration
- Calendar service wrapper in `internal/calendar/service.go`
- Structured data models in `pkg/models/event.go` for type-safe API responses
- Scopes: `calendar.CalendarReadonlyScope` and `drive.DriveMetadataReadonlyScope`

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/api/calendar/v3` - Google Calendar API
- `google.golang.org/api/drive/v3` - Google Drive API (planned)
- `golang.org/x/oauth2/google` - OAuth 2.0 client

## OAuth Setup Requirements

Users must:
1. Create Google Cloud project with Calendar/Drive APIs enabled
2. Configure OAuth consent screen for "Desktop application"
3. Download credentials.json to config directory or use `--credentials` flag
4. Run `docs2obsidian setup` to verify configuration and complete OAuth flow

The application will guide users through browser authorization and handle token storage automatically.