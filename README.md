# docs2obsidian

A Go application that integrates Google Calendar and Google Drive with Obsidian notes.

## Setup

### Prerequisites

1. Go 1.24.4 or later
2. A Google Cloud project that you control
3. Access to enable APIs in your Google Cloud project

### Authentication Setup

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
   - **Linux/Unix**: `~/.config/docs2obsidian/credentials.json`
   - **macOS**: `~/.config/docs2obsidian/credentials.json` OR `~/Library/Application Support/docs2obsidian/credentials.json`
   - **Windows**: `%APPDATA%\docs2obsidian\credentials.json`
   - **Fallback**: `./credentials.json` (current directory)

3. **Verify setup**:
   ```bash
   docs2obsidian setup
   ```

4. **Custom credential location** (optional):
   ```bash
   docs2obsidian --credentials /path/to/credentials.json setup
   docs2obsidian --config-dir /custom/config/dir setup
   ```

The application will guide you through the OAuth flow on first run and save your authorization token in the same config directory.

## Usage

### Build the application
```bash
go build -o docs2obsidian ./cmd
```

### Verify setup
```bash
docs2obsidian setup
```

### List upcoming calendar events
```bash
docs2obsidian calendar
```

### Using custom credential paths
```bash
# Use custom credentials file
docs2obsidian --credentials /path/to/my-credentials.json calendar

# Use custom config directory
docs2obsidian --config-dir /my/config/dir calendar
```

### Command help
```bash
# Get general help
docs2obsidian --help

# Get help for specific commands
docs2obsidian setup --help
docs2obsidian calendar --help
```

## Project Structure

- `cmd/` - Main application entry point
- `internal/auth/` - Authentication handling (OAuth2 and ADC)
- `internal/calendar/` - Google Calendar API integration
- `internal/drive/` - Google Drive API integration (planned)
- `pkg/models/` - Data models for events and files

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
- Use `docs2obsidian setup` to see which paths are being checked

#### "token refresh failed" or authentication errors
- Your OAuth token may have expired
- Delete the token file from your config directory and re-authenticate
- Token locations: same as credentials.json but named `token.json`
- Run `docs2obsidian calendar` to start the OAuth flow again

#### "Calendar API has not been used in project"
- Enable the Google Calendar API in your Google Cloud project
- Go to APIs & Services > Library in Google Cloud Console
- Search for "Calendar API" and enable it

### Getting Help
Run `docs2obsidian setup` to diagnose authentication issues and get specific guidance.

## Next Steps

- Google Drive integration for shared documents
- Obsidian note generation with templates
- Automated synchronization
- Configuration file support