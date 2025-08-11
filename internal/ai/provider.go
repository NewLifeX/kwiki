package ai

import (
	"context"
)

// Provider represents an AI provider interface
type Provider interface {
	// GetName returns the provider name
	GetName() string

	// GetModels returns available models for this provider
	GetModels() []string

	// GenerateText generates text based on the prompt
	GenerateText(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error)

	// GenerateStream generates text with streaming response
	GenerateStream(ctx context.Context, prompt string, options GenerationOptions) (<-chan StreamResponse, error)

	// IsAvailable checks if the provider is available and configured
	IsAvailable() bool

	// GetUsage returns usage statistics
	GetUsage() Usage
}

// GenerationOptions represents options for text generation
type GenerationOptions struct {
	Model        string            `json:"model"`
	Temperature  float32           `json:"temperature"`
	MaxTokens    int               `json:"max_tokens"`
	TopP         float32           `json:"top_p,omitempty"`
	TopK         int               `json:"top_k,omitempty"`
	Stop         []string          `json:"stop,omitempty"`
	Stream       bool              `json:"stream"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Context      []string          `json:"context,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// GenerationResponse represents the response from text generation
type GenerationResponse struct {
	Text         string            `json:"text"`
	TokensUsed   int               `json:"tokens_used"`
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	Duration     int64             `json:"duration"` // milliseconds
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	Text       string            `json:"text"`
	Done       bool              `json:"done"`
	TokensUsed int               `json:"tokens_used,omitempty"`
	Error      error             `json:"error,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Usage represents usage statistics for a provider
type Usage struct {
	TotalRequests  int64   `json:"total_requests"`
	TotalTokens    int64   `json:"total_tokens"`
	TotalCost      float64 `json:"total_cost"`
	LastUsed       int64   `json:"last_used"`
	ErrorCount     int64   `json:"error_count"`
	AverageLatency int64   `json:"average_latency"` // milliseconds
}

// ProviderManager manages multiple AI providers
type ProviderManager struct {
	providers       map[string]Provider
	defaultProvider string
	usage           map[string]Usage
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]Provider),
		usage:     make(map[string]Usage),
	}
}

// RegisterProvider registers a new AI provider
func (pm *ProviderManager) RegisterProvider(name string, provider Provider) {
	pm.providers[name] = provider
}

// GetProvider returns a provider by name
func (pm *ProviderManager) GetProvider(name string) (Provider, bool) {
	provider, exists := pm.providers[name]
	return provider, exists
}

// GetDefaultProvider returns the default provider
func (pm *ProviderManager) GetDefaultProvider() Provider {
	if pm.defaultProvider != "" {
		if provider, exists := pm.providers[pm.defaultProvider]; exists {
			return provider
		}
	}

	// Return first available provider
	for _, provider := range pm.providers {
		if provider.IsAvailable() {
			return provider
		}
	}

	return nil
}

// SetDefaultProvider sets the default provider
func (pm *ProviderManager) SetDefaultProvider(name string) {
	pm.defaultProvider = name
}

// GetAvailableProviders returns all available providers
func (pm *ProviderManager) GetAvailableProviders() map[string]Provider {
	available := make(map[string]Provider)
	for name, provider := range pm.providers {
		if provider.IsAvailable() {
			available[name] = provider
		}
	}
	return available
}

// GetAllProviders returns all registered providers
func (pm *ProviderManager) GetAllProviders() map[string]Provider {
	return pm.providers
}

// GenerateText generates text using the specified or default provider
func (pm *ProviderManager) GenerateText(ctx context.Context, providerName string, prompt string, options GenerationOptions) (*GenerationResponse, error) {
	var provider Provider
	var exists bool

	if providerName != "" {
		provider, exists = pm.providers[providerName]
		if !exists {
			return nil, ErrProviderNotFound
		}
	} else {
		provider = pm.GetDefaultProvider()
		if provider == nil {
			return nil, ErrNoAvailableProvider
		}
	}

	return provider.GenerateText(ctx, prompt, options)
}

// GenerateStream generates streaming text using the specified or default provider
func (pm *ProviderManager) GenerateStream(ctx context.Context, providerName string, prompt string, options GenerationOptions) (<-chan StreamResponse, error) {
	var provider Provider
	var exists bool

	if providerName != "" {
		provider, exists = pm.providers[providerName]
		if !exists {
			return nil, ErrProviderNotFound
		}
	} else {
		provider = pm.GetDefaultProvider()
		if provider == nil {
			return nil, ErrNoAvailableProvider
		}
	}

	return provider.GenerateStream(ctx, prompt, options)
}

// GetProviderModels returns available models for a provider
func (pm *ProviderManager) GetProviderModels(providerName string) ([]string, error) {
	provider, exists := pm.providers[providerName]
	if !exists {
		return nil, ErrProviderNotFound
	}

	return provider.GetModels(), nil
}

// GetAllModels returns all available models from all providers
func (pm *ProviderManager) GetAllModels() map[string][]string {
	models := make(map[string][]string)
	for name, provider := range pm.providers {
		if provider.IsAvailable() {
			models[name] = provider.GetModels()
		}
	}
	return models
}

// UpdateUsage updates usage statistics for a provider
func (pm *ProviderManager) UpdateUsage(providerName string, usage Usage) {
	pm.usage[providerName] = usage
}

// GetUsage returns usage statistics for a provider
func (pm *ProviderManager) GetUsage(providerName string) Usage {
	return pm.usage[providerName]
}

// GetAllUsage returns usage statistics for all providers
func (pm *ProviderManager) GetAllUsage() map[string]Usage {
	return pm.usage
}

// PromptTemplate represents a template for generating prompts
type PromptTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Template    string            `json:"template"`
	Variables   []string          `json:"variables"`
	Type        PromptType        `json:"type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PromptType represents the type of prompt template
type PromptType string

const (
	PromptTypeOverview     PromptType = "overview"
	PromptTypeArchitecture PromptType = "architecture"
	PromptTypeAPI          PromptType = "api"
	PromptTypeFunction     PromptType = "function"
	PromptTypeClass        PromptType = "class"
	PromptTypeDiagram      PromptType = "diagram"
	PromptTypeRAG          PromptType = "rag"
)

// Common errors
var (
	ErrProviderNotFound    = NewAIError("provider not found", "PROVIDER_NOT_FOUND")
	ErrNoAvailableProvider = NewAIError("no available provider", "NO_AVAILABLE_PROVIDER")
	ErrInvalidModel        = NewAIError("invalid model", "INVALID_MODEL")
	ErrAPIKeyMissing       = NewAIError("API key missing", "API_KEY_MISSING")
	ErrRateLimitExceeded   = NewAIError("rate limit exceeded", "RATE_LIMIT_EXCEEDED")
	ErrInsufficientCredits = NewAIError("insufficient credits", "INSUFFICIENT_CREDITS")
)

// AIError represents an AI-related error
type AIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *AIError) Error() string {
	return e.Message
}

// NewAIError creates a new AI error
func NewAIError(message, code string) *AIError {
	return &AIError{
		Message: message,
		Code:    code,
	}
}

// PullProgress represents the progress of pulling a model
type PullProgress struct {
	Status     string  `json:"status"`
	Digest     string  `json:"digest,omitempty"`
	Total      int64   `json:"total,omitempty"`
	Completed  int64   `json:"completed,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
	Error      error   `json:"error,omitempty"`
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Size        string   `json:"size"`
	Tags        []string `json:"tags"`
}
