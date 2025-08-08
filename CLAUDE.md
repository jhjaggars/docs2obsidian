# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the application
make build                   # Build the binary
go build -o pkm-sync ./cmd   # Alternative: direct go build

# Development commands
make test                    # Run all tests
make lint                    # Run comprehensive linting
make fmt                     # Format code with gofmt
make gofmt                   # Check gofmt formatting (CI-friendly)
make gofumpt                 # Format with stricter gofumpt
make imports                 # Fix imports with goimports
make security                # Run security checks (govulncheck + go vet)
make check                   # Run all checks (gofmt, imports, vet, lint, test, security)

# Container operations
make ko-build                # Build container image locally
make ko-push                 # Build and push container image
make ko-run                  # Run container image locally

# Development setup
make dev-setup               # Install all development tools
make help                    # Show all available commands

# Run the application (requires OAuth setup first)
./pkm-sync setup             # Verify authentication configuration
./pkm-sync calendar          # List upcoming calendar events (legacy)
./pkm-sync sync              # Main sync command with multi-source support

# Configuration management
./pkm-sync config init       # Create default configuration
./pkm-sync config show       # Display current configuration
./pkm-sync config validate   # Validate configuration

# Custom paths
./pkm-sync --credentials /path/to/credentials.json setup
./pkm-sync --config-dir /custom/config/dir setup
```

## Architecture Overview

This is a Go CLI application that provides universal Personal Knowledge Management (PKM) synchronization. It connects multiple data sources (Google Calendar, Slack, Gmail, Jira) to PKM targets (Obsidian, Logseq) using OAuth 2.0 authentication.

### CLI Framework
- Uses **Cobra** for command structure with persistent flags
- Root command (`cmd/root.go`) handles global flags: `--credentials`, `--config-dir`
- Main commands: `sync` (primary), `config` (management), legacy commands (`setup`, `calendar`)
- Global flags are processed in `PersistentPreRun` to configure paths

### Multi-Source Architecture
- **Universal interfaces** (`pkg/interfaces/`) for Source and Target abstractions
- **Universal data model** (`pkg/models/item.go`) for consistent data representation
- **Source implementations** in `internal/sources/` (currently: Google)
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
- **OAuth 2.0 only** (no ADC support) with Google Calendar and Drive APIs
- Enhanced copy/paste flow: supports pasting full callback URL or just auth code
- Automatic extraction of auth code from URLs with `extractAuthCode()` function
- Token and credentials stored in platform-specific config directories

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
- `golang.org/x/oauth2/google` - OAuth 2.0 client
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Code Quality Requirements

All code changes must pass the following checks before being merged:

### Required Checks (run with `make check`)
1. **Formatting**: Code must be gofmt'ed (`make gofmt`)
2. **Imports**: Imports must be properly organized (`make imports`)
3. **Linting**: Must pass golangci-lint with comprehensive rules (`make lint`)
4. **Testing**: All tests must pass (`make test`)
5. **Security**: Must pass vulnerability and security checks (`make security`)
6. **Go vet**: Must pass standard Go static analysis

### Linting Rules
The project uses golangci-lint with the following categories of linters:
- **Bug Detection**: errcheck, gosimple, govet, ineffassign, staticcheck, unused, unconvert
- **Code Style**: goimports, misspell, gofmt, gofumpt
- **Code Quality**: goconst, gocritic, gocyclo, revive, nakedret, unparam
- **Security**: gosec
- **Performance**: prealloc
- **Style**: lll (120 char limit), whitespace, nlreturn, nolintlint

### Import Requirements
- All Google API imports must use explicit aliases to avoid conflicts:
  ```go
  calendar "google.golang.org/api/calendar/v3"
  drive "google.golang.org/api/drive/v3"
  ```
- All package imports should be organized by goimports
- YAML package should be explicitly imported: `yaml "gopkg.in/yaml.v3"`

### CI/CD Pipeline
- GitHub Actions automatically run all checks on PRs and main branch
- ko is used for container image builds and multi-platform releases
- Security scanning with govulncheck and go vet
- Multi-version Go testing (1.22, 1.23)

## Current Implementation Status

### Sources
- ✅ **Google Calendar + Drive** - Fully implemented in `internal/sources/google/`
- 🔧 **Slack** - Configuration ready, implementation pending
- 🔧 **Gmail** - Configuration ready, implementation pending
- 🔧 **Jira** - Configuration ready, implementation pending

### Targets
- ✅ **Obsidian** - Implemented with YAML frontmatter and hierarchical structure
- ✅ **Logseq** - Implemented with property blocks and flat structure

### Configuration Features
- ✅ **Multi-source support** with `enabled_sources` array
- ✅ **Per-source configuration** (intervals, priorities, filtering)
- ✅ **Simplified output directory** structure
- ✅ **Local repository configuration** support
- ✅ **Comprehensive validation** and management commands

## OAuth Setup Requirements

Users must:
1. Create Google Cloud project with Calendar/Drive APIs enabled
2. Configure OAuth consent screen for "Desktop application"
3. Add `http://127.0.0.1:*` to authorized redirect URIs (enables automatic authorization flow)
4. Download credentials.json to config directory or use `--credentials` flag
5. Run `./pkm-sync setup` to verify configuration and complete OAuth flow

The application uses an automatic web server-based OAuth flow that opens the user's browser and captures the authorization code automatically. If the web server fails, it falls back to the manual copy/paste flow for compatibility.

## Development Notes

### Migration from docs2obsidian
This application evolved from `docs2obsidian` with full backward compatibility:
- Legacy commands (`calendar`, `export`) still work
- Old configuration paths are automatically detected
- Universal architecture allows easy addition of new sources and targets

### Adding New Sources
1. Implement `interfaces.Source` interface
2. Add source configuration to `models.SourceConfig`
3. Update `createSource()` function in `cmd/sync.go`
4. Add documentation to `CONFIGURATION.md`

### Adding New Targets
1. Implement `interfaces.Target` interface
2. Add target configuration to `models.TargetConfig`
3. Update `createTargetWithConfig()` function in `cmd/sync.go`
4. Add documentation to `CONFIGURATION.md`