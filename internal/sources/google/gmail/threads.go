package gmail

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pkm-sync/pkg/models"
)

// ThreadGroup represents a group of emails that belong to the same thread
type ThreadGroup struct {
	ThreadID     string          `json:"thread_id"`
	Subject      string          `json:"subject"`
	Messages     []*models.Item  `json:"messages"`
	Participants []string        `json:"participants"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	MessageCount int             `json:"message_count"`
}

// ThreadProcessor handles thread grouping and consolidation
type ThreadProcessor struct {
	config models.GmailSourceConfig
}

// NewThreadProcessor creates a new thread processor with the given configuration
func NewThreadProcessor(config models.GmailSourceConfig) *ThreadProcessor {
	return &ThreadProcessor{
		config: config,
	}
}

// ProcessThreads groups messages by thread and applies the configured thread mode
func (tp *ThreadProcessor) ProcessThreads(items []*models.Item) ([]*models.Item, error) {
	// Ensure we always return a non-nil slice
	if items == nil {
		return []*models.Item{}, nil
	}
	
	if !tp.config.IncludeThreads {
		// No threading - return individual messages as-is
		return items, nil
	}

	// Group messages by thread ID
	threadGroups := tp.groupMessagesByThread(items)

	// Apply the configured thread processing mode
	switch strings.ToLower(tp.config.ThreadMode) {
	case "consolidated":
		return tp.consolidateThreads(threadGroups), nil
	case "summary":
		return tp.summarizeThreads(threadGroups), nil
	case "individual", "":
		// Default: return individual messages
		return items, nil
	default:
		return nil, fmt.Errorf("unknown thread mode: %s (supported: individual, consolidated, summary)", tp.config.ThreadMode)
	}
}

// groupMessagesByThread groups messages by their thread ID
func (tp *ThreadProcessor) groupMessagesByThread(items []*models.Item) map[string]*ThreadGroup {
	threadGroups := make(map[string]*ThreadGroup)

	for _, item := range items {
		threadID := tp.extractThreadID(item)
		if threadID == "" {
			// No thread ID - treat as individual message
			threadID = item.ID
		}

		if group, exists := threadGroups[threadID]; exists {
			group.Messages = append(group.Messages, item)
			// MessageCount is calculated from len(Messages) - no separate counter needed

			// Update time range
			if item.CreatedAt.Before(group.StartTime) {
				group.StartTime = item.CreatedAt
			}
			if item.CreatedAt.After(group.EndTime) {
				group.EndTime = item.CreatedAt
			}

			// Update participants
			tp.updateParticipants(group, item)
		} else {
			// Create new thread group
			threadGroups[threadID] = &ThreadGroup{
				ThreadID:     threadID,
				Subject:      tp.extractThreadSubject(item),
				Messages:     []*models.Item{item},
				Participants: tp.extractParticipants(item),
				StartTime:    item.CreatedAt,
				EndTime:      item.CreatedAt,
				MessageCount: 1, // Will be updated after processing
			}
		}
	}

	// Sort messages within each thread by creation time and update message count
	for _, group := range threadGroups {
		sort.Slice(group.Messages, func(i, j int) bool {
			return group.Messages[i].CreatedAt.Before(group.Messages[j].CreatedAt)
		})
		// Update message count to be thread-safe
		group.MessageCount = len(group.Messages)
	}

	return threadGroups
}

// consolidateThreads creates one item per thread containing all messages (Option 2A)
func (tp *ThreadProcessor) consolidateThreads(threadGroups map[string]*ThreadGroup) []*models.Item {
	var consolidatedItems []*models.Item

	for _, group := range threadGroups {
		if len(group.Messages) == 1 {
			// Single message - keep as individual
			consolidatedItems = append(consolidatedItems, group.Messages[0])
			continue
		}

		// Create consolidated thread item
		consolidated := &models.Item{
			ID:         fmt.Sprintf("thread_%s", group.ThreadID),
			Title:      fmt.Sprintf("Thread_%s_%d-messages", tp.sanitizeThreadSubject(group.Subject), group.MessageCount),
			Content:    tp.buildConsolidatedContent(group),
			SourceType: "gmail",
			ItemType:   "email_thread",
			CreatedAt:  group.StartTime,
			UpdatedAt:  group.EndTime,
			Metadata:   tp.buildThreadMetadata(group),
			Tags:       tp.buildThreadTags(group),
		}

		consolidatedItems = append(consolidatedItems, consolidated)
	}

	return consolidatedItems
}

// summarizeThreads creates summary items for threads with key messages (Option 2B)
func (tp *ThreadProcessor) summarizeThreads(threadGroups map[string]*ThreadGroup) []*models.Item {
	var summarizedItems []*models.Item

	for _, group := range threadGroups {
		if len(group.Messages) == 1 {
			// Single message - keep as individual
			summarizedItems = append(summarizedItems, group.Messages[0])
			continue
		}

		// Create thread summary
		maxMessages := tp.config.ThreadSummaryLength
		if maxMessages <= 0 {
			maxMessages = 5 // Default
		}

		summary := &models.Item{
			ID:         fmt.Sprintf("thread_summary_%s", group.ThreadID),
			Title:      fmt.Sprintf("Thread-Summary_%s_%d-messages", tp.sanitizeThreadSubject(group.Subject), group.MessageCount),
			Content:    tp.buildThreadSummary(group, maxMessages),
			SourceType: "gmail",
			ItemType:   "email_thread_summary",
			CreatedAt:  group.StartTime,
			UpdatedAt:  group.EndTime,
			Metadata:   tp.buildThreadMetadata(group),
			Tags:       tp.buildThreadTags(group),
		}

		summarizedItems = append(summarizedItems, summary)
	}

	return summarizedItems
}

// buildConsolidatedContent builds content for consolidated thread (all messages)
func (tp *ThreadProcessor) buildConsolidatedContent(group *ThreadGroup) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Thread: %s\n\n", group.Subject))
	content.WriteString(fmt.Sprintf("**Thread ID:** %s  \n", group.ThreadID))
	content.WriteString(fmt.Sprintf("**Messages:** %d  \n", group.MessageCount))
	content.WriteString(fmt.Sprintf("**Participants:** %s  \n", strings.Join(group.Participants, ", ")))
	content.WriteString(fmt.Sprintf("**Duration:** %s to %s  \n\n", 
		group.StartTime.Format("2006-01-02 15:04"), 
		group.EndTime.Format("2006-01-02 15:04")))

	content.WriteString("---\n\n")

	for i, message := range group.Messages {
		content.WriteString(fmt.Sprintf("## Message %d: %s\n\n", i+1, message.Title))
		content.WriteString(fmt.Sprintf("**Date:** %s  \n", message.CreatedAt.Format("2006-01-02 15:04:05")))
		
		// Add sender information if available
		if sender := tp.extractSender(message); sender != "" {
			content.WriteString(fmt.Sprintf("**From:** %s  \n", sender))
		}

		content.WriteString("\n")
		content.WriteString(message.Content)
		content.WriteString("\n\n---\n\n")
	}

	return content.String()
}

// buildThreadSummary builds content for thread summary (key messages only)
func (tp *ThreadProcessor) buildThreadSummary(group *ThreadGroup, maxMessages int) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Thread Summary: %s\n\n", group.Subject))
	content.WriteString(fmt.Sprintf("**Thread ID:** %s  \n", group.ThreadID))
	content.WriteString(fmt.Sprintf("**Total Messages:** %d  \n", group.MessageCount))
	content.WriteString(fmt.Sprintf("**Showing:** %d key messages  \n", min(maxMessages, len(group.Messages))))
	content.WriteString(fmt.Sprintf("**Participants:** %s  \n", strings.Join(group.Participants, ", ")))
	content.WriteString(fmt.Sprintf("**Duration:** %s to %s  \n\n", 
		group.StartTime.Format("2006-01-02 15:04"), 
		group.EndTime.Format("2006-01-02 15:04")))

	content.WriteString("---\n\n")

	// Select key messages to include in summary
	keyMessages := tp.selectKeyMessages(group.Messages, maxMessages)

	for i, message := range keyMessages {
		content.WriteString(fmt.Sprintf("## Key Message %d: %s\n\n", i+1, message.Title))
		content.WriteString(fmt.Sprintf("**Date:** %s  \n", message.CreatedAt.Format("2006-01-02 15:04:05")))
		
		if sender := tp.extractSender(message); sender != "" {
			content.WriteString(fmt.Sprintf("**From:** %s  \n", sender))
		}

		content.WriteString("\n")
		content.WriteString(message.Content)
		content.WriteString("\n\n---\n\n")
	}

	// Add summary of remaining messages if any
	if len(group.Messages) > maxMessages {
		remaining := len(group.Messages) - maxMessages
		content.WriteString(fmt.Sprintf("*%d additional messages not shown in summary*\n", remaining))
	}

	return content.String()
}

// selectKeyMessages selects the most important messages from a thread
func (tp *ThreadProcessor) selectKeyMessages(messages []*models.Item, maxMessages int) []*models.Item {
	if len(messages) <= maxMessages {
		return messages
	}

	var keyMessages []*models.Item

	// Always include first message (thread starter)
	keyMessages = append(keyMessages, messages[0])
	maxMessages--

	// Always include last message (most recent)
	if maxMessages > 0 && len(messages) > 1 {
		keyMessages = append(keyMessages, messages[len(messages)-1])
		maxMessages--
	}

	// Select additional messages from the middle, prioritizing:
	// 1. Messages with different senders
	// 2. Longer messages (more content)
	// 3. Messages with attachments
	if maxMessages > 0 && len(messages) > 2 {
		candidates := messages[1 : len(messages)-1] // Exclude first and last
		
		// Score messages based on importance criteria
		type scoredMessage struct {
			item  *models.Item
			score int
		}

		var scored []scoredMessage
		seenSenders := make(map[string]bool)
		
		// Track senders from first and last message
		if sender := tp.extractSender(messages[0]); sender != "" {
			seenSenders[sender] = true
		}
		if sender := tp.extractSender(messages[len(messages)-1]); sender != "" {
			seenSenders[sender] = true
		}

		for _, msg := range candidates {
			score := 0
			
			// Different sender bonus
			if sender := tp.extractSender(msg); sender != "" && !seenSenders[sender] {
				score += 3
			}
			
			// Content length bonus
			if len(msg.Content) > 500 {
				score += 2
			}
			
			// Attachment bonus
			if len(msg.Attachments) > 0 {
				score += 1
			}

			scored = append(scored, scoredMessage{msg, score})
		}

		// Sort by score (descending)
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		// Add top-scored messages
		for i := 0; i < min(maxMessages, len(scored)); i++ {
			keyMessages = append(keyMessages, scored[i].item)
		}

		// Sort key messages by creation time
		sort.Slice(keyMessages, func(i, j int) bool {
			return keyMessages[i].CreatedAt.Before(keyMessages[j].CreatedAt)
		})
	}

	return keyMessages
}

// Helper functions

func (tp *ThreadProcessor) extractThreadID(item *models.Item) string {
	if threadID, exists := item.Metadata["thread_id"].(string); exists {
		return threadID
	}
	return ""
}

func (tp *ThreadProcessor) extractThreadSubject(item *models.Item) string {
	// Clean up subject line (remove Re:, Fwd:, etc.)
	subject := item.Title
	subject = strings.TrimSpace(subject)
	
	// Remove common prefixes iteratively to handle multiple prefixes
	prefixes := []string{"Re:", "RE:", "Fwd:", "FWD:", "Fw:", "FW:"}
	maxIterations := 10 // Prevent infinite loops
	iterations := 0
	
	for iterations < maxIterations {
		original := subject
		for _, prefix := range prefixes {
			if strings.HasPrefix(subject, prefix) {
				subject = strings.TrimSpace(subject[len(prefix):])
			}
		}
		// If no change was made, we're done
		if subject == original {
			break
		}
		iterations++
	}
	
	return subject
}

func (tp *ThreadProcessor) extractParticipants(item *models.Item) []string {
	var participants []string
	
	// Extract from metadata if available
	if from, exists := item.Metadata["from"]; exists {
		if sender := tp.extractEmailFromRecipient(from); sender != "" {
			participants = append(participants, sender)
		}
	}
	
	return participants
}

func (tp *ThreadProcessor) updateParticipants(group *ThreadGroup, item *models.Item) {
	if from, exists := item.Metadata["from"]; exists {
		if sender := tp.extractEmailFromRecipient(from); sender != "" {
			// Add sender if not already in participants
			found := false
			for _, p := range group.Participants {
				if p == sender {
					found = true
					break
				}
			}
			if !found {
				group.Participants = append(group.Participants, sender)
			}
		}
	}
}

func (tp *ThreadProcessor) extractSender(item *models.Item) string {
	if from, exists := item.Metadata["from"]; exists {
		return tp.extractEmailFromRecipient(from)
	}
	return ""
}

func (tp *ThreadProcessor) extractEmailFromRecipient(recipient interface{}) string {
	if recipient == nil {
		return ""
	}
	
	switch r := recipient.(type) {
	case string:
		return r
	case map[string]interface{}:
		if r == nil {
			return ""
		}
		// Safe type assertion with validation
		if email, exists := r["email"]; exists {
			if emailStr, ok := email.(string); ok && emailStr != "" {
				return emailStr
			}
		}
		if name, exists := r["name"]; exists {
			if nameStr, ok := name.(string); ok && nameStr != "" {
				if email, exists := r["email"]; exists {
					if emailStr, ok := email.(string); ok && emailStr != "" {
						return fmt.Sprintf("%s <%s>", nameStr, emailStr)
					}
				}
				return nameStr
			}
		}
	}
	return ""
}

func (tp *ThreadProcessor) buildThreadMetadata(group *ThreadGroup) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	if group == nil {
		return metadata
	}
	
	metadata["thread_id"] = group.ThreadID
	metadata["message_count"] = group.MessageCount
	metadata["participants"] = group.Participants
	metadata["start_time"] = group.StartTime
	metadata["end_time"] = group.EndTime
	
	// Safe duration calculation
	if !group.StartTime.IsZero() && !group.EndTime.IsZero() {
		metadata["duration_hours"] = group.EndTime.Sub(group.StartTime).Hours()
	} else {
		metadata["duration_hours"] = 0.0
	}
	
	return metadata
}

func (tp *ThreadProcessor) buildThreadTags(group *ThreadGroup) []string {
	var tags []string
	tags = append(tags, "gmail", "thread")
	
	if group.MessageCount > 5 {
		tags = append(tags, "long-thread")
	}
	
	if len(group.Participants) > 2 {
		tags = append(tags, "multi-participant")
	}
	
	return tags
}

// sanitizeThreadSubject cleans up thread subject for use in filenames with security validation
func (tp *ThreadProcessor) sanitizeThreadSubject(subject string) string {
	// Input validation
	if subject == "" {
		return "email-thread"
	}
	
	// Validate thread processor is not nil
	if tp == nil {
		return "email-thread"
	}
	
	// Remove common email prefixes and clean up
	cleaned := tp.extractThreadSubject(&models.Item{Title: subject})
	if cleaned == "" {
		cleaned = subject // Fallback to original if extraction fails
	}
	
	// Optimized string replacements using strings.Replacer for better performance
	// Create replacer with all replacement patterns including security ones
	replacer := strings.NewReplacer(
		// Security: Remove path traversal sequences (order matters - longer patterns first)
		"../", "",
		"./", "",
		"..", "",
		"~", "",
		// Control characters
		"\n", "",
		"\r", "",
		"\t", "",
		"\x00", "",
		// Filename-friendly replacements
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
		"[", "",
		"]", "",
		"(", "",
		")", "",
		"@", "-at-",
		"#", "-",
		"!", "",
		"&", "-and-",
		".", "", // Remove dots to handle .hidden files
	)
	
	// Apply all replacements in one pass
	cleaned = replacer.Replace(cleaned)

	// Remove multiple consecutive hyphens using regex-like approach
	// Use a more efficient approach to collapse multiple hyphens
	var result strings.Builder
	result.Grow(len(cleaned)) // Pre-allocate capacity
	
	prevWasHyphen := false
	for _, char := range cleaned {
		if char == '-' {
			if !prevWasHyphen {
				result.WriteRune(char)
				prevWasHyphen = true
			}
			// Skip additional consecutive hyphens
		} else {
			result.WriteRune(char)
			prevWasHyphen = false
		}
	}
	cleaned = result.String()

	// Remove leading/trailing hyphens and limit length
	cleaned = strings.Trim(cleaned, "-")
	
	// Limit length to avoid very long filenames
	if len(cleaned) > 80 {
		// Ensure we don't slice beyond string length
		if len(cleaned) >= 80 {
			cleaned = cleaned[:80]
		}
		cleaned = strings.Trim(cleaned, "-")
	}

	// Security: Use filepath.Clean to prevent path traversal and validate result
	cleaned = filepath.Base(filepath.Clean(cleaned))
	
	// Additional security validation: ensure it's a safe filename
	if cleaned == "." || cleaned == ".." || strings.Contains(cleaned, string(filepath.Separator)) {
		cleaned = "email-thread"
	}

	// Final validation - ensure we have a valid filename
	if cleaned == "" || cleaned == "-" {
		cleaned = "email-thread"
	}

	return cleaned
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}