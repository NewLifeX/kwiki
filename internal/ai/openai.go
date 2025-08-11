package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	apiKey string
	usage  Usage
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	
	return &OpenAIProvider{
		client: openai.NewClientWithConfig(config),
		apiKey: apiKey,
		usage:  Usage{},
	}
}

// GetName returns the provider name
func (o *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels returns available OpenAI models
func (o *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"o1-preview",
		"o1-mini",
	}
}

// GenerateText generates text using OpenAI
func (o *OpenAIProvider) GenerateText(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error) {
	startTime := time.Now()
	
	// Prepare messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}
	
	if options.SystemPrompt != "" {
		messages = append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: options.SystemPrompt,
			},
		}, messages...)
	}
	
	// Prepare request
	req := openai.ChatCompletionRequest{
		Model:       options.Model,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		TopP:        options.TopP,
		Stop:        options.Stop,
		Stream:      false,
	}
	
	// Send request
	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		o.usage.ErrorCount++
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}
	
	duration := time.Since(startTime)
	
	// Update usage statistics
	o.usage.TotalRequests++
	o.usage.TotalTokens += int64(resp.Usage.TotalTokens)
	o.usage.LastUsed = time.Now().Unix()
	o.usage.AverageLatency = (o.usage.AverageLatency + duration.Milliseconds()) / 2
	
	// Calculate cost (approximate)
	cost := o.calculateCost(options.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	o.usage.TotalCost += cost
	
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}
	
	return &GenerationResponse{
		Text:         resp.Choices[0].Message.Content,
		TokensUsed:   resp.Usage.TotalTokens,
		Model:        resp.Model,
		Provider:     "openai",
		Duration:     duration.Milliseconds(),
		FinishReason: string(resp.Choices[0].FinishReason),
		Metadata: map[string]string{
			"prompt_tokens":     fmt.Sprintf("%d", resp.Usage.PromptTokens),
			"completion_tokens": fmt.Sprintf("%d", resp.Usage.CompletionTokens),
			"cost":              fmt.Sprintf("%.6f", cost),
		},
	}, nil
}

// GenerateStream generates text with streaming response
func (o *OpenAIProvider) GenerateStream(ctx context.Context, prompt string, options GenerationOptions) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 10)
	
	go func() {
		defer close(responseChan)
		
		// Prepare messages
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
		
		if options.SystemPrompt != "" {
			messages = append([]openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: options.SystemPrompt,
				},
			}, messages...)
		}
		
		// Prepare request
		req := openai.ChatCompletionRequest{
			Model:       options.Model,
			Messages:    messages,
			Temperature: options.Temperature,
			MaxTokens:   options.MaxTokens,
			TopP:        options.TopP,
			Stop:        options.Stop,
			Stream:      true,
		}
		
		// Create stream
		stream, err := o.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			o.usage.ErrorCount++
			responseChan <- StreamResponse{Error: fmt.Errorf("OpenAI stream error: %w", err)}
			return
		}
		defer stream.Close()
		
		totalTokens := 0
		
		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					// Stream finished
					responseChan <- StreamResponse{Done: true, TokensUsed: totalTokens}
					break
				}
				responseChan <- StreamResponse{Error: fmt.Errorf("stream receive error: %w", err)}
				return
			}
			
			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					// Estimate tokens (rough approximation)
					tokens := len(content) / 4
					totalTokens += tokens
					
					responseChan <- StreamResponse{
						Text:       content,
						Done:       false,
						TokensUsed: tokens,
						Metadata: map[string]string{
							"model": response.Model,
						},
					}
				}
				
				if response.Choices[0].FinishReason != "" {
					// Stream finished
					o.usage.TotalRequests++
					o.usage.TotalTokens += int64(totalTokens)
					o.usage.LastUsed = time.Now().Unix()
					
					responseChan <- StreamResponse{
						Done:       true,
						TokensUsed: totalTokens,
						Metadata: map[string]string{
							"finish_reason": string(response.Choices[0].FinishReason),
						},
					}
					break
				}
			}
		}
	}()
	
	return responseChan, nil
}

// IsAvailable checks if OpenAI is available
func (o *OpenAIProvider) IsAvailable() bool {
	if o.apiKey == "" {
		return false
	}
	
	// Test with a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := o.client.ListModels(ctx)
	return err == nil
}

// GetUsage returns usage statistics
func (o *OpenAIProvider) GetUsage() Usage {
	return o.usage
}

// calculateCost calculates the approximate cost for OpenAI API usage
func (o *OpenAIProvider) calculateCost(model string, promptTokens, completionTokens int) float64 {
	// Pricing as of 2024 (per 1K tokens)
	pricing := map[string]struct {
		input  float64
		output float64
	}{
		"gpt-4o":         {0.0025, 0.01},
		"gpt-4o-mini":    {0.00015, 0.0006},
		"gpt-4-turbo":    {0.01, 0.03},
		"gpt-4":          {0.03, 0.06},
		"gpt-3.5-turbo":  {0.0005, 0.0015},
		"o1-preview":     {0.015, 0.06},
		"o1-mini":        {0.003, 0.012},
	}
	
	if price, exists := pricing[model]; exists {
		inputCost := (float64(promptTokens) / 1000) * price.input
		outputCost := (float64(completionTokens) / 1000) * price.output
		return inputCost + outputCost
	}
	
	// Default pricing if model not found
	return (float64(promptTokens+completionTokens) / 1000) * 0.002
}
