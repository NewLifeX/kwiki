package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	apiKey string
	usage  Usage
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		usage:  Usage{},
	}
}

// GetName returns the provider name
func (g *GeminiProvider) GetName() string {
	return "gemini"
}

// GetModels returns available Gemini models
func (g *GeminiProvider) GetModels() []string {
	return []string{
		"gemini-2.0-flash-exp",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-1.0-pro",
	}
}

// initClient initializes the Gemini client if not already initialized
func (g *GeminiProvider) initClient(ctx context.Context) error {
	if g.client != nil {
		return nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	g.client = client
	return nil
}

// GenerateText generates text using Gemini
func (g *GeminiProvider) GenerateText(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error) {
	startTime := time.Now()

	if err := g.initClient(ctx); err != nil {
		return nil, err
	}

	// Get the model
	model := g.client.GenerativeModel(options.Model)

	// Configure generation parameters
	model.SetTemperature(options.Temperature)
	if options.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(options.MaxTokens))
	}
	if options.TopP > 0 {
		model.SetTopP(options.TopP)
	}
	if options.TopK > 0 {
		model.SetTopK(int32(options.TopK))
	}
	if len(options.Stop) > 0 {
		model.StopSequences = options.Stop
	}

	// Set system instruction if provided
	if options.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(options.SystemPrompt)},
		}
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		g.usage.ErrorCount++
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	duration := time.Since(startTime)

	// Update usage statistics
	g.usage.TotalRequests++
	g.usage.LastUsed = time.Now().Unix()
	g.usage.AverageLatency = (g.usage.AverageLatency + duration.Milliseconds()) / 2

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates returned")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response content")
	}

	// Extract text from parts
	var text string
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			text += string(textPart)
		}
	}

	// Estimate tokens (rough approximation)
	tokensUsed := len(text) / 4
	g.usage.TotalTokens += int64(tokensUsed)

	// Get finish reason
	finishReason := "stop"
	if candidate.FinishReason != 0 {
		finishReason = candidate.FinishReason.String()
	}

	metadata := map[string]string{
		"finish_reason": finishReason,
	}

	// Add usage metadata if available
	if resp.UsageMetadata != nil {
		if resp.UsageMetadata.PromptTokenCount > 0 {
			metadata["prompt_tokens"] = fmt.Sprintf("%d", resp.UsageMetadata.PromptTokenCount)
		}
		if resp.UsageMetadata.CandidatesTokenCount > 0 {
			metadata["completion_tokens"] = fmt.Sprintf("%d", resp.UsageMetadata.CandidatesTokenCount)
		}
		if resp.UsageMetadata.TotalTokenCount > 0 {
			metadata["total_tokens"] = fmt.Sprintf("%d", resp.UsageMetadata.TotalTokenCount)
			tokensUsed = int(resp.UsageMetadata.TotalTokenCount)
		}
	}

	return &GenerationResponse{
		Text:         text,
		TokensUsed:   tokensUsed,
		Model:        options.Model,
		Provider:     "gemini",
		Duration:     duration.Milliseconds(),
		FinishReason: finishReason,
		Metadata:     metadata,
	}, nil
}

// GenerateStream generates text with streaming response
func (g *GeminiProvider) GenerateStream(ctx context.Context, prompt string, options GenerationOptions) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 10)

	go func() {
		defer close(responseChan)

		if err := g.initClient(ctx); err != nil {
			responseChan <- StreamResponse{Error: err}
			return
		}

		// Get the model
		model := g.client.GenerativeModel(options.Model)

		// Configure generation parameters
		model.SetTemperature(options.Temperature)
		if options.MaxTokens > 0 {
			model.SetMaxOutputTokens(int32(options.MaxTokens))
		}
		if options.TopP > 0 {
			model.SetTopP(options.TopP)
		}
		if options.TopK > 0 {
			model.SetTopK(int32(options.TopK))
		}
		if len(options.Stop) > 0 {
			model.StopSequences = options.Stop
		}

		// Set system instruction if provided
		if options.SystemPrompt != "" {
			model.SystemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(options.SystemPrompt)},
			}
		}

		// Generate streaming content
		iter := model.GenerateContentStream(ctx, genai.Text(prompt))

		totalTokens := 0

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() == "iterator done" {
					// Stream finished
					g.usage.TotalRequests++
					g.usage.TotalTokens += int64(totalTokens)
					g.usage.LastUsed = time.Now().Unix()

					responseChan <- StreamResponse{Done: true, TokensUsed: totalTokens}
					break
				}
				responseChan <- StreamResponse{Error: fmt.Errorf("stream error: %w", err)}
				return
			}

			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				if candidate.Content != nil && len(candidate.Content.Parts) > 0 {
					// Extract text from parts
					var text string
					for _, part := range candidate.Content.Parts {
						if textPart, ok := part.(genai.Text); ok {
							text += string(textPart)
						}
					}

					if text != "" {
						// Estimate tokens
						tokens := len(text) / 4
						totalTokens += tokens

						responseChan <- StreamResponse{
							Text:       text,
							Done:       false,
							TokensUsed: tokens,
							Metadata: map[string]string{
								"model": options.Model,
							},
						}
					}
				}

				// Check if finished
				if candidate.FinishReason != 0 {
					g.usage.TotalRequests++
					g.usage.TotalTokens += int64(totalTokens)
					g.usage.LastUsed = time.Now().Unix()

					responseChan <- StreamResponse{
						Done:       true,
						TokensUsed: totalTokens,
						Metadata: map[string]string{
							"finish_reason": candidate.FinishReason.String(),
						},
					}
					break
				}
			}
		}
	}()

	return responseChan, nil
}

// IsAvailable checks if Gemini is available
func (g *GeminiProvider) IsAvailable() bool {
	if g.apiKey == "" {
		return false
	}

	// Test with a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := g.initClient(ctx); err != nil {
		return false
	}

	// Try to list models to test connectivity
	model := g.client.GenerativeModel("gemini-1.5-flash")
	_, err := model.GenerateContent(ctx, genai.Text("test"))

	return err == nil
}

// GetUsage returns usage statistics
func (g *GeminiProvider) GetUsage() Usage {
	return g.usage
}

// Close closes the Gemini client
func (g *GeminiProvider) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
