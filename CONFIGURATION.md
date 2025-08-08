# pkm-sync Configuration Guide

This document provides comprehensive configuration reference for pkm-sync.

## Configuration File Locations

pkm-sync looks for configuration files in this order:
1. **Custom directory** (via `--config-dir` flag)
2. **Global config**: `~/.config/pkm-sync/config.yaml`
3. **Local repository**: `./config.yaml` (current directory)

## Quick Start

### Create Configuration
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

### Configuration Commands
```bash
pkm-sync config init                    # Create default config
pkm-sync config show                    # Show current config
pkm-sync config path                    # Show config file location
pkm-sync config edit                    # Open config in editor
pkm-sync config validate               # Validate configuration
```

## Configuration File Structure

The config file allows you to set defaults so you don't need to specify flags every time:

```yaml
sync:
  enabled_sources: ["google"]  # Multiple sources can be enabled
  default_target: obsidian  
  default_since: 7d
  default_output_dir: ./exported  # Single output directory for all targets
  merge_sources: true      # Combine data from all sources
  source_tags: true        # Add source-specific tags
  on_conflict: skip        # How to handle conflicts

sources:
  google:
    enabled: true          # Must be true and in enabled_sources
    type: google
    priority: 1            # Sync order
    google:
      calendar_id: primary
      include_declined: false
      download_docs: true
      
  slack:
    enabled: false         # Future source (not implemented yet)
    type: slack
    priority: 2

targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: Calendar      # Folder within output directory
      date_format: "2006-01-02"
      include_frontmatter: true
      
  logseq:
    type: logseq
    logseq:
      default_page: Calendar
      use_properties: true
```

## Complete Configuration Reference

### Sync Settings (`sync:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled_sources` | array | `["google"]` | Array of active sources |
| `default_target` | string | `"obsidian"` | Default PKM target (obsidian, logseq) |
| `default_since` | string | `"7d"` | Default time range (7d, today, 2025-01-01) |
| `default_output_dir` | string | `"./exported"` | Single output directory for all targets |
| `source_schedules` | object | `{"google": "24h", "slack": "1h", "gmail_work": "4h", "gmail_personal": "6h", "jira": "2h"}` | Per-source sync intervals |
| `auto_sync` | boolean | `false` | Enable automatic syncing |
| `sync_interval` | duration | `24h` | Fallback sync interval |
| `merge_sources` | boolean | `true` | Combine data from all enabled sources |
| `source_tags` | boolean | `true` | Add source-specific tags to items |
| `on_conflict` | string | `"skip"` | How to handle conflicts (skip, overwrite, prompt) |
| `deduplicate_by` | string | `"id"` | Deduplication strategy (id, title, content, none) |
| `create_subdirs` | boolean | `true` | Create subdirectories for organization |
| `subdir_format` | string | `"source"` | Subdirectory naming (yyyy/mm, yyyy-mm, source, flat) |
| `max_file_age` | string | `"365d"` | Maximum age for keeping files |
| `archive_old_files` | boolean | `false` | Archive files exceeding max age |

### Source Configuration (`sources.{name}:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | boolean | `true` (google), `false` (others) | Enable this source |
| `type` | string | varies | Source type (google, slack, gmail, jira) |
| `priority` | integer | varies by source | Sync order priority (1=highest) |
| `sync_interval` | duration | inherited | Override global sync interval |
| `since` | string | inherited | Override global since parameter |

### Google Source Settings (`sources.google.google:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `calendar_id` | string | `"primary"` | Calendar to sync (primary or specific ID) |
| `include_declined` | boolean | `false` | Include declined events |
| `include_private` | boolean | `true` | Include private events |
| `event_types` | array | `[]` | Filter by event types |
| `download_docs` | boolean | `true` | Download attached Google Docs |
| `doc_formats` | array | `["markdown"]` | Export formats for docs |
| `max_doc_size` | string | `"10MB"` | Maximum document size |
| `include_shared` | boolean | `true` | Include shared documents |
| `request_delay` | duration | `100ms` | Delay between API requests |
| `max_requests` | integer | `100` | Maximum API requests |

### Gmail Source Settings (`sources.{gmail_instance}.gmail:`)

Gmail integration supports multiple instances (e.g., `gmail_work`, `gmail_personal`) with independent configurations:

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `name` | string | **required** | Human-readable instance name |
| `description` | string | `""` | Optional description of this Gmail instance |
| `labels` | array | `["IMPORTANT", "STARRED"]` | Gmail labels to sync |
| `query` | string | `""` | Custom Gmail search query |
| `include_unread` | boolean | `true` | Include unread emails |
| `include_read` | boolean | `false` | Include read emails |
| `include_threads` | boolean | `false` | Include full email threads |
| `max_email_age` | string | `"30d"` | Maximum email age (30d, 1y, etc.) |
| `min_email_age` | string | `""` | Minimum email age (exclude very recent) |
| `from_domains` | array | `[]` | Filter by sender domains (["company.com"]) |
| `to_domains` | array | `[]` | Filter by recipient domains |
| `exclude_from_domains` | array | `[]` | Exclude sender domains (["noreply.com"]) |
| `require_attachments` | boolean | `false` | Only emails with attachments |
| `extract_links` | boolean | `true` | Extract URLs from email content |
| `extract_recipients` | boolean | `true` | Extract to/cc/bcc details |
| `include_full_headers` | boolean | `false` | Include all email headers |
| `process_html_content` | boolean | `true` | Convert HTML to markdown |
| `include_original_html` | boolean | `false` | Keep original HTML version |
| `strip_quoted_text` | boolean | `false` | Remove quoted reply text |
| `extract_signatures` | boolean | `false` | Extract email signatures |
| `download_attachments` | boolean | `false` | Download email attachments |
| `attachment_types` | array | `["pdf", "doc", "docx"]` | Allowed attachment types |
| `max_attachment_size` | string | `"5MB"` | Maximum attachment size |
| `attachment_subdir` | string | `""` | Custom attachment folder |
| `request_delay` | duration | `0` | Delay between API requests for rate limiting |
| `max_requests` | integer | `0` | Maximum requests per sync (0=unlimited) |
| `batch_size` | integer | `0` | Messages per API call for large mailboxes (0=auto) |
| `filename_template` | string | `""` | Custom filename template |
| `include_thread_context` | boolean | `false` | Link to thread messages |
| `group_by_thread` | boolean | `false` | One file per thread |
| `tagging_rules` | array | `[]` | Custom tagging rules |

#### Gmail Filename Templates

Available template variables for `filename_template`:
- `{{date}}` - Email date (YYYY-MM-DD)
- `{{time}}` - Email time (HH:MM:SS)
- `{{from}}` - Sender email or name
- `{{to}}` - Primary recipient
- `{{subject}}` - Email subject
- `{{id}}` - Gmail message ID
- `{{thread}}` - Gmail thread ID
- `{{year}}`, `{{month}}`, `{{day}}` - Date components

#### Gmail Tagging Rules

Tagging rules allow automatic tag assignment based on email properties:

```yaml
tagging_rules:
  - condition: "from:ceo@company.com"
    tags: ["urgent", "leadership"]
  - condition: "has:attachment"
    tags: ["has-attachment"]
  - condition: "to:team@company.com"
    tags: ["team-email"]
```

Supported conditions:
- `from:domain.com` - Sender domain
- `to:domain.com` - Recipient domain
- `has:attachment` - Has attachments
- `subject:keyword` - Subject contains keyword
- `label:LABELNAME` - Has specific Gmail label

### Enhanced Source Configuration (`sources.{name}:`)

Enhanced source settings support per-instance customization:

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | boolean | varies | Enable this source |
| `type` | string | varies | Source type (google, gmail, slack, jira) |
| `name` | string | `""` | Human-readable instance name |
| `output_subdir` | string | `""` | Custom subdirectory for this source |
| `output_target` | string | `""` | Override default target for this source |
| `priority` | integer | varies | Sync order priority (1=highest) |
| `sync_interval` | duration | inherited | Override global sync interval |
| `since` | string | inherited | Override global since parameter |

### Target Configuration (`targets.{name}:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `type` | string | varies | Target type (obsidian, logseq) |

### Obsidian Target Settings (`targets.obsidian.obsidian:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `default_folder` | string | `"Calendar"` | Folder within output directory |
| `filename_template` | string | `"{{date}} - {{title}}"` | File naming pattern |
| `date_format` | string | `"2006-01-02"` | Date format for filenames |
| `tag_prefix` | string | `"calendar/"` | Prefix for tags |
| `include_frontmatter` | boolean | `true` | Add YAML frontmatter |
| `custom_fields` | array | `[]` | Additional frontmatter fields |
| `template_file` | string | `""` | Custom template file path |
| `create_daily_notes` | boolean | `false` | Create daily note entries |
| `daily_notes_folder` | string | `"Daily Notes"` | Folder for daily notes |
| `link_format` | string | `"wikilink"` | Link style (wikilink, markdown) |
| `attachment_folder` | string | `"Attachments"` | Folder for attachments |
| `download_attachments` | boolean | `true` | Download file attachments |

### Logseq Target Settings (`targets.logseq.logseq:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `default_page` | string | `"Calendar"` | Default page for entries |
| `use_properties` | boolean | `true` | Use property blocks |
| `property_prefix` | string | `""` | Prefix for properties |
| `block_indentation` | integer | `2` | Indentation level |
| `create_journal_refs` | boolean | `true` | Link to journal pages |
| `journal_date_format` | string | `"Jan 2nd, 2006"` | Date format for journal refs |

### Authentication Settings (`auth:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `credentials_path` | string | `~/.config/pkm-sync/credentials.json` | Path to OAuth credentials file |
| `token_path` | string | `~/.config/pkm-sync/token.json` | Path to stored tokens |
| `encrypt_tokens` | boolean | `false` | Encrypt stored tokens |
| `token_expiration` | string | `"30d"` | Token refresh period |

### Application Settings (`app:`)

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `log_level` | string | `"info"` | Logging level (debug, info, warn, error) |
| `log_file` | string | `""` | Log file path (empty = stdout) |
| `quiet_mode` | boolean | `false` | Suppress non-error output |
| `verbose_mode` | boolean | `false` | Enable verbose output |
| `create_backups` | boolean | `true` | Create backups before sync |
| `backup_dir` | string | `~/.config/pkm-sync/backups` | Backup directory path |
| `max_backups` | integer | `5` | Maximum backup files to keep |
| `cache_enabled` | boolean | `true` | Enable local caching |
| `cache_dir` | string | `~/.config/pkm-sync/cache` | Cache directory path |
| `cache_ttl` | duration | `24h` | Cache expiration time |
| `notify_on_success` | boolean | `false` | Show success notifications |
| `notify_on_error` | boolean | `true` | Show error notifications |

<!--
Future Source Configurations (Hidden for Development Planning)

### Slack Source Settings (`sources.slack.slack:`)
- `workspace_id` - Slack workspace ID
- `channels` - Channels to sync (["#general", "#dev"])
- `include_threads` - Include threaded messages (true/false)
- `include_dms` - Include direct messages (true/false)
- `min_importance` - Message importance filter ("starred", "mentions", "all")
- `exclude_bots` - Exclude bot messages (true/false)
- `min_length` - Minimum message length (10)
- `include_files` - Include file attachments (true/false)
- `file_types` - Allowed file types (["pdf", "doc", "img"])

### Gmail Source Settings (`sources.gmail.gmail:`)
- `labels` - Gmail labels to sync (["IMPORTANT", "STARRED"])
- `query` - Custom Gmail search query
- `include_unread` - Include unread emails (true/false)
- `include_threads` - Include email threads (true/false)
- `max_email_age` - Maximum email age ("30d")
- `extract_links` - Extract links from emails (true/false)
- `download_attachments` - Download email attachments (true/false)
- `attachment_types` - Allowed attachment types (["pdf", "doc"])
- `max_attachment_size` - Maximum attachment size ("5MB")

### Jira Source Settings (`sources.jira.jira:`)
- `instance_url` - Jira instance URL ("https://company.atlassian.net")
- `project_keys` - Project keys to sync (["PROJ", "TEAM"])
- `jql` - Custom JQL query for filtering
- `issue_types` - Issue types to include (["Bug", "Story", "Task"])
- `statuses` - Issue statuses to include (["In Progress", "Done"])
- `assignee_filter` - Assignee filter ("me", "team", "all")
- `include_comments` - Include issue comments (true/false)
- `include_history` - Include issue history (true/false)
- `include_attachments` - Include issue attachments (true/false)
-->

## Configuration Examples

### Repository-Specific Configuration
```bash
# Create a project-specific config in your repository
cat > config.yaml << EOF
sync:
  enabled_sources: ["google"]
  default_target: obsidian
  default_output_dir: ./docs/calendar
  
targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: calendar
EOF

# Add to .gitignore (output directory)
echo "docs/calendar/" >> .gitignore

# Now sync uses repository configuration
pkm-sync sync  # Uses ./config.yaml instead of ~/.config/pkm-sync/config.yaml
```

### Enable Multiple Sources
```bash
# Create config with multiple sources
pkm-sync config init --source google
# Edit config to enable additional sources
pkm-sync config edit
# Add slack and gmail to enabled_sources array
```

### Per-Source Configuration
```yaml
sources:
  google:
    enabled: true
    priority: 1
    since: "7d"              # Google-specific time range
    sync_interval: 24h       # Sync Google daily
    
  slack:
    enabled: true  
    priority: 2
    since: "1d"              # Slack-specific time range
    sync_interval: 1h        # Sync Slack hourly
```

### Gmail Multi-Instance Configuration

Gmail integration supports multiple independent instances for different email workflows:

```yaml
sync:
  enabled_sources: ["gmail_work", "gmail_personal", "gmail_newsletters"]
  default_target: obsidian
  default_output_dir: ./email-vault

sources:
  gmail_work:
    enabled: true
    type: gmail
    name: "Work Emails"
    priority: 1
    output_subdir: "work"
    output_target: obsidian
    since: "30d"
    gmail:
      name: "Work Important Emails"
      description: "High-priority work communications"
      labels: ["IMPORTANT", "STARRED"]
      query: "from:company.com OR to:company.com"
      include_unread: true
      include_read: false
      max_email_age: "90d"
      from_domains: ["company.com", "client.com"]
      extract_recipients: true
      extract_links: true
      process_html_content: true
      strip_quoted_text: true
      download_attachments: true
      attachment_types: ["pdf", "doc", "docx"]
      max_attachment_size: "10MB"
      attachment_subdir: "work-attachments"
      filename_template: "{{date}}-{{from}}-{{subject}}"
      tagging_rules:
        - condition: "from:ceo@company.com"
          tags: ["urgent", "leadership"]
        - condition: "has:attachment"
          tags: ["has-attachment"]

  gmail_personal:
    enabled: true
    type: gmail
    name: "Personal Important"
    priority: 2
    output_subdir: "personal"
    since: "14d"
    gmail:
      name: "Personal Starred Emails"
      labels: ["STARRED"]
      query: "is:important -category:promotions"
      include_unread: true
      max_email_age: "30d"
      exclude_from_domains: ["noreply.com", "notifications.com"]
      extract_recipients: false
      process_html_content: true
      download_attachments: false
      filename_template: "{{date}}-{{subject}}"
      
  gmail_newsletters:
    enabled: false
    type: gmail
    name: "Newsletters & Updates"
    priority: 3
    output_subdir: "newsletters"
    gmail:
      name: "Newsletter Archive"
      query: "category:promotions OR category:updates"
      include_unread: false
      include_read: true
      max_email_age: "7d"
      min_email_age: "1d"  # Skip very recent to avoid duplicates
      process_html_content: false
      extract_links: true
      filename_template: "{{date}}-newsletter-{{from}}"
      tagging_rules:
        - condition: "category:promotions"
          tags: ["newsletter", "promotion"]
```

### Gmail OAuth Setup

Gmail integration requires OAuth 2.0 setup with Google Cloud:

1. **Create Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable Gmail API and Google Drive API

2. **Configure OAuth Consent Screen**:
   - Set application type to "Desktop application"
   - Add your email as a test user

3. **Create OAuth Credentials**:
   - Create "Desktop application" credentials
   - Add `http://127.0.0.1:*` to authorized redirect URIs
   - Download `credentials.json`

4. **Place Credentials**:
   ```bash
   # Global config directory
   mv ~/Downloads/credentials.json ~/.config/pkm-sync/credentials.json
   
   # OR: Use custom path with flag
   pkm-sync --credentials /path/to/credentials.json setup
   ```

5. **Complete OAuth Flow**:
   ```bash
   pkm-sync setup  # Opens browser for OAuth authorization
   ```

### Gmail Performance Optimization

For large mailboxes (1000+ emails), pkm-sync automatically optimizes performance:

```yaml
sources:
  gmail_large_mailbox:
    enabled: true
    type: gmail
    gmail:
      name: "Large Mailbox Optimization"
      
      # Performance settings for large mailboxes
      batch_size: 100                    # Process emails in batches of 100
      request_delay: 50ms                # Add delay to avoid rate limits
      max_requests: 5000                 # Limit total requests per sync
      
      # Memory management
      max_email_age: "90d"               # Limit scope to reduce memory usage
      process_html_content: false        # Disable heavy processing if not needed
      download_attachments: false        # Skip attachments for faster processing
      
      # Error handling
      strip_quoted_text: true            # Reduce content size
      extract_links: false               # Skip link extraction for speed
      
      # Progress reporting automatically enabled for large batches
```

**Automatic Optimizations:**
- Batch processing for requests > 1000 messages
- Progress reporting every 500 messages
- Memory management for batches > 1000 messages  
- Exponential backoff retry for rate limits
- Streaming interface for very large mailboxes

### Advanced Gmail Filtering

Gmail supports powerful search operators for precise email filtering:

```yaml
sources:
  gmail_advanced:
    enabled: true
    type: gmail
    gmail:
      name: "Advanced Gmail Filtering"
      
      # Combine multiple filter approaches
      labels: ["IMPORTANT"]                    # Gmail labels
      query: "(has:attachment OR is:starred) AND from:company.com"  # Custom search
      from_domains: ["company.com", "client.com"]  # Domain filtering
      max_email_age: "60d"                     # Time-based filtering
      require_attachments: false               # Attachment filtering
      
      # Content processing
      process_html_content: true
      strip_quoted_text: true
      extract_links: true
      extract_recipients: true
      
      # Smart tagging
      tagging_rules:
        - condition: "from:support@company.com"
          tags: ["support", "internal"]
        - condition: "subject:invoice"
          tags: ["finance", "invoice"]
        - condition: "has:attachment"
          tags: ["has-file"]
        - condition: "to:team@company.com"
          tags: ["team-communication"]
```

## Migration from Previous Versions

### Simplified Output Directory Structure
If you have an older configuration with separate `output_dir` fields per target, you can migrate to the simplified structure:

**Old Structure (no longer supported):**
```yaml
targets:
  obsidian:
    output_dir: ./vault        # ❌ Removed
    obsidian:
      vault_path: ./vault      # ❌ Removed
```

**New Structure:**
```yaml
sync:
  default_output_dir: ./vault  # ✅ Single output directory

targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: Calendar  # ✅ Folder within output directory
```

This change simplifies configuration and works better with multi-source synchronization.