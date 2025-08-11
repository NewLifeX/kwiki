package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// DeepSeekProvider implements the Provider interface for DeepSeek
type DeepSeekProvider struct {
	client *openai.Client
	apiKey string
	usage  Usage
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider(apiKey string) *DeepSeekProvider {
	log.Printf("[DeepSeek] 创建DeepSeek提供商，API密钥长度: %d", len(apiKey))

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"

	log.Printf("[DeepSeek] 配置API基础URL: %s", config.BaseURL)

	provider := &DeepSeekProvider{
		client: openai.NewClientWithConfig(config),
		apiKey: apiKey,
		usage:  Usage{},
	}

	log.Printf("[DeepSeek] DeepSeek提供商创建成功")
	return provider
}

// GetName returns the provider name
func (d *DeepSeekProvider) GetName() string {
	return "deepseek"
}

// GetModels returns available DeepSeek models
func (d *DeepSeekProvider) GetModels() []string {
	return []string{
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-reasoner",
		"deepseek-r1",
		"deepseek-r1-distill-llama-70b",
		"deepseek-r1-distill-qwen-32b",
		"deepseek-r1-distill-qwen-14b",
		"deepseek-r1-distill-qwen-7b",
		"deepseek-r1-distill-qwen-1.5b",
	}
}

// GenerateText generates text using DeepSeek
func (d *DeepSeekProvider) GenerateText(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error) {
	startTime := time.Now()
	log.Printf("[DeepSeek] 开始生成文本，提示词长度: %d", len(prompt))

	// Build messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Add system prompt if provided
	if options.SystemPrompt != "" {
		log.Printf("[DeepSeek] 添加系统提示词，长度: %d", len(options.SystemPrompt))
		messages = append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: options.SystemPrompt,
			},
		}, messages...)
	}

	// Set default model if not specified
	model := options.Model
	if model == "" {
		model = "deepseek-chat"
	}
	log.Printf("[DeepSeek] 使用模型: %s", model)

	// Create request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		TopP:        options.TopP,
		Stop:        options.Stop,
		Stream:      false,
	}

	log.Printf("[DeepSeek] 请求参数: 模型=%s, 温度=%.2f, 最大令牌=%d",
		req.Model, req.Temperature, req.MaxTokens)

	// Make API call
	log.Printf("[DeepSeek] 发送API请求到: %s", "https://api.deepseek.com/v1")

	// 添加超时控制（增加到2分钟）
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	resp, err := d.client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("[DeepSeek] API调用失败: %v", err)
		log.Printf("[DeepSeek] 错误类型: %T", err)
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[DeepSeek] API调用超时")
		}
		d.usage.ErrorCount++
		return nil, fmt.Errorf("deepseek API error: %w", err)
	}

	log.Printf("[DeepSeek] API调用成功，响应选择数: %d", len(resp.Choices))

	// Calculate duration
	duration := time.Since(startTime).Milliseconds()
	log.Printf("[DeepSeek] 请求完成，耗时: %dms", duration)

	// Log usage information
	log.Printf("[DeepSeek] 令牌使用: 提示=%d, 完成=%d, 总计=%d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	// Update usage statistics
	d.usage.TotalRequests++
	d.usage.TotalTokens += int64(resp.Usage.TotalTokens)
	d.usage.LastUsed = time.Now().Unix()
	d.usage.AverageLatency = (d.usage.AverageLatency + duration) / 2

	// Extract response text
	var text string
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		text = choice.Message.Content

		log.Printf("[DeepSeek] 响应选择详情:")
		log.Printf("  - 消息角色: %s", choice.Message.Role)
		log.Printf("  - 消息内容长度: %d", len(choice.Message.Content))
		log.Printf("  - 完成原因: %s", choice.FinishReason)

		// 对于reasoner模型，可能需要检查其他字段
		if choice.Message.Content == "" {
			log.Printf("[DeepSeek] 警告: 消息内容为空，检查原始响应")
			log.Printf("  - 原始消息: %+v", choice.Message)
		}

		if len(text) > 0 {
			log.Printf("[DeepSeek] 生成文本预览: %s...",
				text[:min(200, len(text))])
		} else {
			log.Printf("[DeepSeek] 生成文本为空")
		}
	} else {
		log.Printf("[DeepSeek] 警告: 响应中没有选择项")
	}

	log.Printf("[DeepSeek] 文本生成完成，总请求数: %d, 总令牌数: %d",
		d.usage.TotalRequests, d.usage.TotalTokens)

	return &GenerationResponse{
		Text:         text,
		TokensUsed:   resp.Usage.TotalTokens,
		Model:        model,
		Provider:     "deepseek",
		Duration:     duration,
		FinishReason: string(resp.Choices[0].FinishReason),
		Metadata: map[string]string{
			"prompt_tokens":     fmt.Sprintf("%d", resp.Usage.PromptTokens),
			"completion_tokens": fmt.Sprintf("%d", resp.Usage.CompletionTokens),
			"total_tokens":      fmt.Sprintf("%d", resp.Usage.TotalTokens),
		},
	}, nil
}

// GenerateStream generates text using DeepSeek with streaming
func (d *DeepSeekProvider) GenerateStream(ctx context.Context, prompt string, options GenerationOptions) (<-chan StreamResponse, error) {
	log.Printf("[DeepSeek] 开始流式生成文本，提示词长度: %d", len(prompt))

	// Build messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Add system prompt if provided
	if options.SystemPrompt != "" {
		log.Printf("[DeepSeek] 添加系统提示词，长度: %d", len(options.SystemPrompt))
		messages = append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: options.SystemPrompt,
			},
		}, messages...)
	}

	// Set default model if not specified
	model := options.Model
	if model == "" {
		model = "deepseek-chat"
	}
	log.Printf("[DeepSeek] 使用流式模型: %s", model)

	// Create request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		TopP:        options.TopP,
		Stop:        options.Stop,
		Stream:      true, // 启用流式
	}

	log.Printf("[DeepSeek] 流式请求参数: 模型=%s, 温度=%.2f, 最大令牌=%d",
		req.Model, req.Temperature, req.MaxTokens)

	// Create response channel
	responseChan := make(chan StreamResponse, 100)

	// Start streaming in goroutine
	go func() {
		defer close(responseChan)

		startTime := time.Now()
		log.Printf("[DeepSeek] 发送流式API请求")

		stream, err := d.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			log.Printf("[DeepSeek] 流式API调用失败: %v", err)
			d.usage.ErrorCount++
			responseChan <- StreamResponse{
				Error: fmt.Errorf("deepseek stream API error: %w", err),
				Done:  true,
			}
			return
		}
		defer stream.Close()

		log.Printf("[DeepSeek] 流式连接建立，开始接收数据")

		var fullText strings.Builder
		var totalTokens int
		var chunkCount int
		lastProgressTime := time.Now()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				log.Printf("[DeepSeek] 流式接收错误: %v", err)
				responseChan <- StreamResponse{
					Error: err,
					Done:  true,
				}
				return
			}

			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				delta := choice.Delta.Content

				if delta != "" {
					fullText.WriteString(delta)
					chunkCount++

					// 每2秒报告一次进度，避免日志过多
					if time.Since(lastProgressTime) > 2*time.Second {
						log.Printf("[DeepSeek] 流式进度: 已接收 %d 字符 (%d 片段)", fullText.Len(), chunkCount)
						lastProgressTime = time.Now()
					}

					responseChan <- StreamResponse{
						Text: delta,
						Done: false,
						Metadata: map[string]string{
							"model": model,
						},
					}
				}

				// 检查是否完成
				if choice.FinishReason != "" {
					duration := time.Since(startTime).Milliseconds()

					// 更新使用统计
					d.usage.TotalRequests++
					d.usage.TotalTokens += int64(totalTokens)
					d.usage.LastUsed = time.Now().Unix()
					d.usage.AverageLatency = (d.usage.AverageLatency + duration) / 2

					// 记录完成统计
					log.Printf("[DeepSeek] 流式生成完成统计:")
					log.Printf("  - 耗时: %dms", duration)
					log.Printf("  - 完成原因: %s", choice.FinishReason)
					log.Printf("  - 总文本长度: %d 字符", fullText.Len())
					log.Printf("  - 接收片段数: %d", chunkCount)
					log.Printf("  - 平均片段大小: %.1f 字符", float64(fullText.Len())/float64(chunkCount))
					log.Printf("  - 生成速度: %.1f 字符/秒", float64(fullText.Len())/float64(duration)*1000)

					// 发送最终响应
					responseChan <- StreamResponse{
						Text:       fullText.String(),
						Done:       true,
						TokensUsed: totalTokens,
						Metadata: map[string]string{
							"model":         model,
							"finish_reason": string(choice.FinishReason),
							"chunks":        fmt.Sprintf("%d", chunkCount),
							"duration_ms":   fmt.Sprintf("%d", duration),
						},
					}
					break
				}
			}
		}
	}()

	return responseChan, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsAvailable checks if DeepSeek is available
func (d *DeepSeekProvider) IsAvailable() bool {
	return d.apiKey != ""
}

// GetUsage returns usage statistics
func (d *DeepSeekProvider) GetUsage() Usage {
	return d.usage
}

// SetAPIKey updates the API key
func (d *DeepSeekProvider) SetAPIKey(apiKey string) {
	d.apiKey = apiKey
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"
	d.client = openai.NewClientWithConfig(config)
}

// GetAPIKey returns the current API key (masked)
func (d *DeepSeekProvider) GetAPIKey() string {
	if len(d.apiKey) < 8 {
		return "****"
	}
	return d.apiKey[:4] + "****" + d.apiKey[len(d.apiKey)-4:]
}

// ValidateAPIKey validates the API key by making a test request
func (d *DeepSeekProvider) ValidateAPIKey(ctx context.Context) error {
	if d.apiKey == "" {
		return ErrAPIKeyMissing
	}

	// Make a simple test request
	_, err := d.GenerateText(ctx, "Hello", GenerationOptions{
		Model:       "deepseek-chat",
		Temperature: 0.1,
		MaxTokens:   10,
	})

	return err
}

// GetModelInfo returns detailed information about DeepSeek models
func (d *DeepSeekProvider) GetModelInfo() []ModelInfo {
	return []ModelInfo{
		{
			Name:        "deepseek-chat",
			Description: "DeepSeek's flagship conversational AI model with strong reasoning capabilities",
			Size:        "67B parameters",
			Tags:        []string{"chat", "reasoning", "general"},
		},
		{
			Name:        "deepseek-coder",
			Description: "Specialized model for code generation and programming tasks",
			Size:        "33B parameters",
			Tags:        []string{"code", "programming", "development"},
		},
		{
			Name:        "deepseek-reasoner",
			Description: "Advanced reasoning model for complex problem solving",
			Size:        "67B parameters",
			Tags:        []string{"reasoning", "analysis", "problem-solving"},
		},
		{
			Name:        "deepseek-r1",
			Description: "Latest reasoning model with enhanced capabilities",
			Size:        "671B parameters",
			Tags:        []string{"reasoning", "latest", "advanced"},
		},
		{
			Name:        "deepseek-r1-distill-llama-70b",
			Description: "Distilled version of R1 based on Llama architecture",
			Size:        "70B parameters",
			Tags:        []string{"distilled", "llama", "efficient"},
		},
		{
			Name:        "deepseek-r1-distill-qwen-32b",
			Description: "Distilled version of R1 based on Qwen architecture (32B)",
			Size:        "32B parameters",
			Tags:        []string{"distilled", "qwen", "medium"},
		},
		{
			Name:        "deepseek-r1-distill-qwen-14b",
			Description: "Distilled version of R1 based on Qwen architecture (14B)",
			Size:        "14B parameters",
			Tags:        []string{"distilled", "qwen", "compact"},
		},
		{
			Name:        "deepseek-r1-distill-qwen-7b",
			Description: "Distilled version of R1 based on Qwen architecture (7B)",
			Size:        "7B parameters",
			Tags:        []string{"distilled", "qwen", "small"},
		},
		{
			Name:        "deepseek-r1-distill-qwen-1.5b",
			Description: "Distilled version of R1 based on Qwen architecture (1.5B)",
			Size:        "1.5B parameters",
			Tags:        []string{"distilled", "qwen", "tiny"},
		},
	}
}
