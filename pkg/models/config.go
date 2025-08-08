package models

import "time"

// Config represents the application configuration
type Config struct {
	// Default sync settings
	Sync SyncConfig `json:"sync" yaml:"sync"`
	
	// Source configurations
	Sources map[string]SourceConfig `json:"sources" yaml:"sources"`
	
	// Target configurations  
	Targets map[string]TargetConfig `json:"targets" yaml:"targets"`
	
	// Authentication settings
	Auth AuthConfig `json:"auth" yaml:"auth"`
	
	// General application settings
	App AppConfig `json:"app" yaml:"app"`
}

type SyncConfig struct {
	// Multi-source configuration
	EnabledSources []string `json:"enabled_sources" yaml:"enabled_sources"`     // ["google", "slack", "gmail"]
	DefaultTarget  string   `json:"default_target" yaml:"default_target"`
	
	// Default time range for syncing
	DefaultSince string `json:"default_since" yaml:"default_since"`
	
	// Default output directory
	DefaultOutputDir string `json:"default_output_dir" yaml:"default_output_dir"`
	
	// Source-specific scheduling
	SourceSchedules map[string]string `json:"source_schedules" yaml:"source_schedules"` // "google": "1h", "slack": "30m"
	
	// Global sync settings
	AutoSync     bool          `json:"auto_sync" yaml:"auto_sync"`
	SyncInterval time.Duration `json:"sync_interval" yaml:"sync_interval"` // Fallback interval
	
	// Data handling
	MergeSources    bool   `json:"merge_sources" yaml:"merge_sources"`       // Combine all sources into single export
	SourceTags      bool   `json:"source_tags" yaml:"source_tags"`           // Add source-specific tags
	OnConflict      string `json:"on_conflict" yaml:"on_conflict"`           // "skip", "overwrite", "prompt"
	DeduplicateBy   string `json:"deduplicate_by" yaml:"deduplicate_by"`     // "id", "title", "content", "none"
	
	// File management
	CreateSubdirs    bool   `json:"create_subdirs" yaml:"create_subdirs"`
	SubdirFormat     string `json:"subdir_format" yaml:"subdir_format"`       // "yyyy/mm", "yyyy-mm", "source", "flat"
	MaxFileAge       string `json:"max_file_age" yaml:"max_file_age"`          // "30d", "6m", "1y"
	ArchiveOldFiles  bool   `json:"archive_old_files" yaml:"archive_old_files"`
}

type SourceConfig struct {
	// Source enablement and type
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Type    string `json:"type" yaml:"type"`
	
	// Source-specific sync settings
	SyncInterval time.Duration `json:"sync_interval,omitempty" yaml:"sync_interval,omitempty"` // Override global interval
	Since        string        `json:"since,omitempty" yaml:"since,omitempty"`                 // Override global since
	Priority     int           `json:"priority,omitempty" yaml:"priority,omitempty"`           // Sync order (1=highest)
	
	// Source-specific configurations
	Google GoogleSourceConfig `json:"google,omitempty" yaml:"google,omitempty"`
	Slack  SlackSourceConfig  `json:"slack,omitempty" yaml:"slack,omitempty"`
	Gmail  GmailSourceConfig  `json:"gmail,omitempty" yaml:"gmail,omitempty"`
	Jira   JiraSourceConfig   `json:"jira,omitempty" yaml:"jira,omitempty"`
}

type GoogleSourceConfig struct {
	// Calendar settings
	CalendarID       string   `json:"calendar_id" yaml:"calendar_id"` // "primary" or specific calendar
	IncludeDeclined  bool     `json:"include_declined" yaml:"include_declined"`
	IncludePrivate   bool     `json:"include_private" yaml:"include_private"`
	EventTypes       []string `json:"event_types" yaml:"event_types"` // filter by event types
	
	// Drive settings
	DownloadDocs     bool     `json:"download_docs" yaml:"download_docs"`
	DocFormats       []string `json:"doc_formats" yaml:"doc_formats"` // "markdown", "pdf", "docx"
	MaxDocSize       string   `json:"max_doc_size" yaml:"max_doc_size"` // "10MB"
	IncludeShared    bool     `json:"include_shared" yaml:"include_shared"`
	
	// Rate limiting
	RequestDelay time.Duration `json:"request_delay" yaml:"request_delay"`
	MaxRequests  int           `json:"max_requests" yaml:"max_requests"`
}

type TargetConfig struct {
	// Target type (output directory comes from SyncConfig.DefaultOutputDir)
	Type string `json:"type" yaml:"type"`
	
	// Obsidian-specific settings
	Obsidian ObsidianTargetConfig `json:"obsidian,omitempty" yaml:"obsidian,omitempty"`
	
	// Logseq-specific settings
	Logseq LogseqTargetConfig `json:"logseq,omitempty" yaml:"logseq,omitempty"`
}

type ObsidianTargetConfig struct {
	// Vault organization (vault path is the output directory)
	DefaultFolder    string `json:"default_folder" yaml:"default_folder"` // "Calendar", "Inbox"
	
	// File naming and organization
	FilenameTemplate string `json:"filename_template" yaml:"filename_template"` // "{{date}} - {{title}}"
	DateFormat       string `json:"date_format" yaml:"date_format"`             // "2006-01-02"
	TagPrefix        string `json:"tag_prefix" yaml:"tag_prefix"`               // "calendar/"
	
	// Content formatting
	IncludeFrontmatter  bool     `json:"include_frontmatter" yaml:"include_frontmatter"`
	CustomFields        []string `json:"custom_fields" yaml:"custom_fields"`
	TemplateFile        string   `json:"template_file" yaml:"template_file"`
	
	// Linking and references
	CreateDailyNotes    bool   `json:"create_daily_notes" yaml:"create_daily_notes"`
	DailyNotesFolder    string `json:"daily_notes_folder" yaml:"daily_notes_folder"`
	LinkFormat          string `json:"link_format" yaml:"link_format"` // "wikilink", "markdown"
	
	// Attachments
	AttachmentFolder    string `json:"attachment_folder" yaml:"attachment_folder"`
	DownloadAttachments bool   `json:"download_attachments" yaml:"download_attachments"`
}

type LogseqTargetConfig struct {
	// Graph settings (graph path is the output directory)
	DefaultPage   string `json:"default_page" yaml:"default_page"`
	
	// Content formatting
	UseProperties     bool   `json:"use_properties" yaml:"use_properties"`
	PropertyPrefix    string `json:"property_prefix" yaml:"property_prefix"`
	BlockIndentation  int    `json:"block_indentation" yaml:"block_indentation"`
	
	// Journal integration
	CreateJournalRefs bool   `json:"create_journal_refs" yaml:"create_journal_refs"`
	JournalDateFormat string `json:"journal_date_format" yaml:"journal_date_format"`
}

type AuthConfig struct {
	// OAuth settings
	CredentialsPath string `json:"credentials_path" yaml:"credentials_path"`
	TokenPath       string `json:"token_path" yaml:"token_path"`
	
	// Security settings
	EncryptTokens   bool   `json:"encrypt_tokens" yaml:"encrypt_tokens"`
	TokenExpiration string `json:"token_expiration" yaml:"token_expiration"` // "30d"
}

type AppConfig struct {
	// Logging and output
	LogLevel     string `json:"log_level" yaml:"log_level"`         // "debug", "info", "warn", "error"
	LogFile      string `json:"log_file" yaml:"log_file"`
	QuietMode    bool   `json:"quiet_mode" yaml:"quiet_mode"`
	VerboseMode  bool   `json:"verbose_mode" yaml:"verbose_mode"`
	
	// Backup and recovery
	CreateBackups bool   `json:"create_backups" yaml:"create_backups"`
	BackupDir     string `json:"backup_dir" yaml:"backup_dir"`
	MaxBackups    int    `json:"max_backups" yaml:"max_backups"`
	
	// Performance
	CacheEnabled bool          `json:"cache_enabled" yaml:"cache_enabled"`
	CacheDir     string        `json:"cache_dir" yaml:"cache_dir"`
	CacheTTL     time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
	
	// Notifications
	NotifyOnSuccess bool `json:"notify_on_success" yaml:"notify_on_success"`
	NotifyOnError   bool `json:"notify_on_error" yaml:"notify_on_error"`
}

// Future source configurations (placeholders for planned integrations)

type SlackSourceConfig struct {
	// Workspace and channel settings
	WorkspaceID     string   `json:"workspace_id" yaml:"workspace_id"`
	Channels        []string `json:"channels" yaml:"channels"`           // ["#general", "#dev"]
	IncludeThreads  bool     `json:"include_threads" yaml:"include_threads"`
	IncludeDMs      bool     `json:"include_dms" yaml:"include_dms"`
	MinImportance   string   `json:"min_importance" yaml:"min_importance"` // "starred", "mentions", "all"
	
	// Content filtering
	ExcludeBots     bool     `json:"exclude_bots" yaml:"exclude_bots"`
	MinLength       int      `json:"min_length" yaml:"min_length"`       // Minimum message length
	IncludeFiles    bool     `json:"include_files" yaml:"include_files"`
	FileTypes       []string `json:"file_types" yaml:"file_types"`       // ["pdf", "doc", "img"]
}

type GmailSourceConfig struct {
	// Email filtering
	Labels          []string `json:"labels" yaml:"labels"`               // ["IMPORTANT", "STARRED"]
	Query           string   `json:"query" yaml:"query"`                 // Gmail search query
	IncludeUnread   bool     `json:"include_unread" yaml:"include_unread"`
	IncludeThreads  bool     `json:"include_threads" yaml:"include_threads"`
	MaxEmailAge     string   `json:"max_email_age" yaml:"max_email_age"`  // "30d"
	
	// Content processing
	ExtractLinks    bool     `json:"extract_links" yaml:"extract_links"`
	DownloadAttachments bool `json:"download_attachments" yaml:"download_attachments"`
	AttachmentTypes []string `json:"attachment_types" yaml:"attachment_types"` // ["pdf", "doc"]
	MaxAttachmentSize string `json:"max_attachment_size" yaml:"max_attachment_size"` // "5MB"
}

type JiraSourceConfig struct {
	// Instance and authentication
	InstanceURL     string   `json:"instance_url" yaml:"instance_url"`   // "https://company.atlassian.net"
	ProjectKeys     []string `json:"project_keys" yaml:"project_keys"`   // ["PROJ", "TEAM"]
	
	// Issue filtering  
	JQL             string   `json:"jql" yaml:"jql"`                     // Custom JQL query
	IssueTypes      []string `json:"issue_types" yaml:"issue_types"`     // ["Bug", "Story", "Task"]
	Statuses        []string `json:"statuses" yaml:"statuses"`           // ["In Progress", "Done"]
	AssigneeFilter  string   `json:"assignee_filter" yaml:"assignee_filter"` // "me", "team", "all"
	
	// Content inclusion
	IncludeComments bool     `json:"include_comments" yaml:"include_comments"`
	IncludeHistory  bool     `json:"include_history" yaml:"include_history"`
	IncludeAttachments bool  `json:"include_attachments" yaml:"include_attachments"`
}