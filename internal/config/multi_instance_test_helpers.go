package config

import (
	"os"
	"path/filepath"
	"testing"

	"pkm-sync/pkg/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const multiInstanceConfig = `
sync:
  enabled_sources: ["gmail_work", "gmail_personal", "google_calendar"]
  default_target: obsidian
  default_output_dir: ./vault
  source_tags: true
  merge_sources: false
  create_subdirs: true
  subdir_format: source

sources:
  gmail_work:
    enabled: true
    type: gmail
    name: "Work Emails"
    priority: 1
    output_subdir: "work-emails"
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
      request_delay: 500ms
      max_requests: 1000
      batch_size: 50
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
    output_subdir: "personal-emails"
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

  google_calendar:
    enabled: true
    type: google
    name: "Primary Calendar"
    priority: 3
    output_subdir: "calendar"
    since: "7d"
    google:
      calendar_id: "primary"
      include_declined: false
      include_private: true
      download_docs: true
      doc_formats: ["markdown", "pdf"]
      max_doc_size: "5MB"
      include_shared: true
      request_delay: 1s
      max_requests: 500

targets:
  obsidian:
    type: obsidian
    obsidian:
      default_folder: "Synced"
      filename_template: "{{date}} - {{title}}"
      date_format: "2006-01-02"
      tag_prefix: "sync/"
      include_frontmatter: true
      create_daily_notes: false
      link_format: "wikilink"
      attachment_folder: "attachments"
      download_attachments: true

  logseq:
    type: logseq
    logseq:
      default_page: "Inbox"
      use_properties: true
      property_prefix: "sync::"
      block_indentation: 2
      create_journal_refs: true
      journal_date_format: "2006-01-02"

auth:
  credentials_path: "./credentials.json"
  token_path: "./token.json"
  encrypt_tokens: false

app:
  log_level: "info"
  quiet_mode: false
  verbose_mode: false
  create_backups: true
  backup_dir: "./backups"
  max_backups: 5
  cache_enabled: true
  cache_ttl: 24h
`

func setupConfigTest(t *testing.T) (*models.Config, func()) {
	tempDir, err := os.MkdirTemp("", "pkm-sync-config-test")
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(multiInstanceConfig), 0644)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config models.Config

	err = yaml.Unmarshal(data, &config)
	require.NoError(t, err)

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return &config, cleanup
}
