package models

import (
	"time"
)

// Wiki represents a generated wiki
type Wiki struct {
	ID           string                `json:"id"`
	RepositoryID string                `json:"repository_id"`
	PackagePath  string                `json:"package_path"` // e.g., github.com/gorilla/websocket
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Status       WikiStatus            `json:"status"`
	Progress     int                   `json:"progress"` // 0-100
	Tags         []string              `json:"tags"`     // Wiki-level tags
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	GeneratedBy  string                `json:"generated_by"` // AI provider used
	Model        string                `json:"model"`        // AI model used
	Language     string                `json:"language"`     // Primary language of this wiki
	Languages    []string              `json:"languages"`    // All available languages
	Pages        []WikiPage            `json:"pages"`
	Diagrams     []WikiDiagram         `json:"diagrams"`
	Settings     WikiSettings          `json:"settings"`
	Metadata     WikiMetadata          `json:"metadata"`
	Translations map[string]*WikiTrans `json:"translations,omitempty"` // Language code -> translation
}

// WikiStatus represents the status of wiki generation
type WikiStatus string

const (
	WikiStatusPending    WikiStatus = "pending"
	WikiStatusAnalyzing  WikiStatus = "analyzing"
	WikiStatusGenerating WikiStatus = "generating"
	WikiStatusCompleted  WikiStatus = "completed"
	WikiStatusFailed     WikiStatus = "failed"
)

// LogLevel represents the level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelSuccess LogLevel = "success"
	LogLevelDebug   LogLevel = "debug"
)

// WikiTrans represents a translation of a wiki
type WikiTrans struct {
	Language    string        `json:"language"`    // Language code (e.g., "en", "zh", "ja")
	Title       string        `json:"title"`       // Translated title
	Description string        `json:"description"` // Translated description
	Pages       []WikiPage    `json:"pages"`       // Translated pages
	Diagrams    []WikiDiagram `json:"diagrams"`    // Translated diagrams
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// SupportedLanguages defines the supported languages for wiki generation
var SupportedLanguages = map[string]string{
	"en": "English",
	"zh": "中文",
	"ja": "日本語",
	"ko": "한국어",
	"es": "Español",
	"fr": "Français",
	"de": "Deutsch",
	"ru": "Русский",
	"pt": "Português",
	"it": "Italiano",
}

// WikiLogEntry represents a detailed log entry
type WikiLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      LogLevel  `json:"level"`
	Step       string    `json:"step"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Duration   string    `json:"duration,omitempty"`
	Progress   int       `json:"progress,omitempty"`
	FilesCount int       `json:"files_count,omitempty"`
	TokensUsed int       `json:"tokens_used,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// WikiPage represents a page in the wiki
type WikiPage struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        PageType  `json:"type"`
	Order       int       `json:"order"`
	ParentID    string    `json:"parent_id,omitempty"`
	Children    []string  `json:"children,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WordCount   int       `json:"word_count"`
	ReadingTime int       `json:"reading_time"` // in minutes
}

// PageType represents the type of wiki page
type PageType string

const (
	PageTypeOverview     PageType = "overview"
	PageTypeArchitecture PageType = "architecture"
	PageTypeAPI          PageType = "api"
	PageTypeModule       PageType = "module"
	PageTypeFunction     PageType = "function"
	PageTypeClass        PageType = "class"
	PageTypeTutorial     PageType = "tutorial"
	PageTypeReference    PageType = "reference"
	PageTypeChangelog    PageType = "changelog"
	PageTypeGuide        PageType = "guide"
)

// WikiDiagram represents a diagram in the wiki
type WikiDiagram struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        DiagramType `json:"type"`
	Content     string      `json:"content"` // Mermaid syntax
	Description string      `json:"description"`
	PageID      string      `json:"page_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// DiagramType represents the type of diagram
type DiagramType string

const (
	DiagramTypeFlowchart    DiagramType = "flowchart"
	DiagramTypeSequence     DiagramType = "sequence"
	DiagramTypeClass        DiagramType = "class"
	DiagramTypeER           DiagramType = "er"
	DiagramTypeGantt        DiagramType = "gantt"
	DiagramTypeGitGraph     DiagramType = "gitgraph"
	DiagramTypeArchitecture DiagramType = "architecture"
	DiagramTypeDataFlow     DiagramType = "dataflow"
)

// WikiSettings represents wiki generation settings
type WikiSettings struct {
	AIProvider      string            `json:"ai_provider"`
	Model           string            `json:"model"`
	Temperature     float32           `json:"temperature"`
	MaxTokens       int               `json:"max_tokens"`
	EnableDiagrams  bool              `json:"enable_diagrams"`
	EnableRAG       bool              `json:"enable_rag"`
	Language        string            `json:"language"`
	Theme           string            `json:"theme"`
	CustomPrompts   map[string]string `json:"custom_prompts,omitempty"`
	ExcludePatterns []string          `json:"exclude_patterns,omitempty"`
	IncludePatterns []string          `json:"include_patterns,omitempty"`
}

// WikiMetadata represents additional metadata about the wiki
type WikiMetadata struct {
	GenerationTime    time.Duration  `json:"generation_time"`
	TokensUsed        int            `json:"tokens_used"`
	FilesProcessed    int            `json:"files_processed"`
	PagesGenerated    int            `json:"pages_generated"`
	DiagramsGenerated int            `json:"diagrams_generated"`
	Languages         []string       `json:"languages"`
	Complexity        string         `json:"complexity"` // low, medium, high
	Quality           float64        `json:"quality"`    // 0-1
	Tags              []string       `json:"tags"`
	Categories        []string       `json:"categories"`
	Statistics        map[string]int `json:"statistics"`
	PackagePath       string         `json:"package_path"`   // 包路径，如 github.com/gin-gonic/gin
	RepositoryURL     string         `json:"repository_url"` // 原始仓库URL
}

// GenerationRequest represents a request to generate a wiki
type GenerationRequest struct {
	RepositoryURL    string       `json:"repository_url"`
	Branch           string       `json:"branch,omitempty"`
	AccessToken      string       `json:"access_token,omitempty"`
	Settings         WikiSettings `json:"settings"`
	Title            string       `json:"title,omitempty"`
	Description      string       `json:"description,omitempty"`
	CustomPrompts    []string     `json:"custom_prompts,omitempty"`
	Languages        []string     `json:"languages,omitempty"`          // Languages to generate (e.g., ["en", "zh"])
	PrimaryLanguage  string       `json:"primary_language,omitempty"`   // Primary language (default: "en")
	GenerateAllLangs bool         `json:"generate_all_langs,omitempty"` // Generate all supported languages
}

// GenerationProgress represents the progress of wiki generation
type GenerationProgress struct {
	WikiID      string     `json:"wiki_id"`
	Status      WikiStatus `json:"status"`
	Progress    int        `json:"progress"`
	CurrentStep string     `json:"current_step"`
	Message     string     `json:"message"`
	Error       string     `json:"error,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ETA         *time.Time `json:"eta,omitempty"`
}

// ChatMessage represents a message in the RAG chat system
type ChatMessage struct {
	ID         string      `json:"id"`
	WikiID     string      `json:"wiki_id"`
	Role       MessageRole `json:"role"`
	Content    string      `json:"content"`
	Context    []string    `json:"context,omitempty"` // Retrieved context
	Sources    []string    `json:"sources,omitempty"` // Source files
	Timestamp  time.Time   `json:"timestamp"`
	TokensUsed int         `json:"tokens_used,omitempty"`
}

// MessageRole represents the role of a chat message
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// SearchResult represents a search result in the wiki
type SearchResult struct {
	Type       string   `json:"type"` // page, diagram, code
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Excerpt    string   `json:"excerpt"`
	Score      float64  `json:"score"`
	Highlights []string `json:"highlights"`
	URL        string   `json:"url"`
	PageID     string   `json:"page_id,omitempty"`
	DiagramID  string   `json:"diagram_id,omitempty"`
}

// WikiExport represents exported wiki data
type WikiExport struct {
	Format    ExportFormat `json:"format"`
	Content   string       `json:"content"`
	Filename  string       `json:"filename"`
	Size      int64        `json:"size"`
	CreatedAt time.Time    `json:"created_at"`
}

// ExportFormat represents the format for wiki export
type ExportFormat string

const (
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatZIP      ExportFormat = "zip"
)
