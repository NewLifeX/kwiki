package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
	usage      Usage
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes timeout for long generations
		},
		usage: Usage{},
	}
}

// GetName returns the provider name
func (o *OllamaProvider) GetName() string {
	return "ollama"
}

// GetModels returns available models from Ollama
func (o *OllamaProvider) GetModels() []string {
	log.Printf("[Ollama] 获取模型列表...")
	models, err := o.fetchModels()
	if err != nil {
		log.Printf("[Ollama] 获取模型失败: %v，使用默认模型列表", err)
		// Return common Ollama models as fallback
		fallbackModels := []string{
			"huihui_ai/deepseek-r1-abliterated:32b",
			"huihui_ai/deepseek-r1-abliterated:14b",
			"llama3.2:latest",
			"llama3.1:latest",
			"llama3:latest",
			"codellama:latest",
			"mistral:latest",
			"gemma2:latest",
			"qwen2.5:latest",
			"qwen3-coder:30b:latest",
			"deepseek-r1:1.5b",
			"deepseek-coder:6.7b",
		}
		log.Printf("[Ollama] 返回默认模型列表: %v", fallbackModels)
		return fallbackModels
	}
	log.Printf("[Ollama] 成功获取模型列表: %v", models)
	return models
}

// fetchModels fetches available models from Ollama API
func (o *OllamaProvider) fetchModels() ([]string, error) {
	resp, err := o.httpClient.Get(o.baseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models: %s", resp.Status)
	}

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	models := make([]string, len(response.Models))
	for i, model := range response.Models {
		models[i] = model.Name
	}

	return models, nil
}

// GenerateText generates text using Ollama
func (o *OllamaProvider) GenerateText(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error) {
	startTime := time.Now()

	// Prepare request
	reqBody := map[string]interface{}{
		"model":  options.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": options.Temperature,
			"num_predict": options.MaxTokens,
		},
	}

	if options.SystemPrompt != "" {
		reqBody["system"] = options.SystemPrompt
	}

	if options.TopP > 0 {
		reqBody["options"].(map[string]interface{})["top_p"] = options.TopP
	}

	if options.TopK > 0 {
		reqBody["options"].(map[string]interface{})["top_k"] = options.TopK
	}

	if len(options.Stop) > 0 {
		reqBody["options"].(map[string]interface{})["stop"] = options.Stop
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.usage.ErrorCount++
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		o.usage.ErrorCount++
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Context  []int  `json:"context,omitempty"`
		Model    string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	duration := time.Since(startTime)

	// Update usage statistics
	o.usage.TotalRequests++
	o.usage.LastUsed = time.Now().Unix()
	o.usage.AverageLatency = (o.usage.AverageLatency + duration.Milliseconds()) / 2

	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	tokensUsed := len(ollamaResp.Response) / 4
	o.usage.TotalTokens += int64(tokensUsed)

	return &GenerationResponse{
		Text:         ollamaResp.Response,
		TokensUsed:   tokensUsed,
		Model:        ollamaResp.Model,
		Provider:     "ollama",
		Duration:     duration.Milliseconds(),
		FinishReason: "stop",
		Metadata: map[string]string{
			"context_length": fmt.Sprintf("%d", len(ollamaResp.Context)),
		},
	}, nil
}

// GenerateStream generates text with streaming response
func (o *OllamaProvider) GenerateStream(ctx context.Context, prompt string, options GenerationOptions) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 10)

	go func() {
		defer close(responseChan)

		// Prepare request
		reqBody := map[string]interface{}{
			"model":  options.Model,
			"prompt": prompt,
			"stream": true,
			"options": map[string]interface{}{
				"temperature": options.Temperature,
				"num_predict": options.MaxTokens,
			},
		}

		if options.SystemPrompt != "" {
			reqBody["system"] = options.SystemPrompt
		}

		if options.TopP > 0 {
			reqBody["options"].(map[string]interface{})["top_p"] = options.TopP
		}

		if options.TopK > 0 {
			reqBody["options"].(map[string]interface{})["top_k"] = options.TopK
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			responseChan <- StreamResponse{Error: fmt.Errorf("failed to marshal request: %w", err)}
			return
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
		if err != nil {
			responseChan <- StreamResponse{Error: fmt.Errorf("failed to create request: %w", err)}
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := o.httpClient.Do(req)
		if err != nil {
			o.usage.ErrorCount++
			responseChan <- StreamResponse{Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			o.usage.ErrorCount++
			body, _ := io.ReadAll(resp.Body)
			responseChan <- StreamResponse{Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		// Read streaming response
		decoder := json.NewDecoder(resp.Body)
		totalTokens := 0

		for {
			var chunk struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
				Model    string `json:"model"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				responseChan <- StreamResponse{Error: fmt.Errorf("failed to decode chunk: %w", err)}
				return
			}

			// Estimate tokens
			tokens := len(chunk.Response) / 4
			totalTokens += tokens

			responseChan <- StreamResponse{
				Text:       chunk.Response,
				Done:       chunk.Done,
				TokensUsed: tokens,
				Metadata: map[string]string{
					"model": chunk.Model,
				},
			}

			if chunk.Done {
				// Update usage statistics
				o.usage.TotalRequests++
				o.usage.TotalTokens += int64(totalTokens)
				o.usage.LastUsed = time.Now().Unix()
				break
			}
		}
	}()

	return responseChan, nil
}

// IsAvailable checks if Ollama is available
func (o *OllamaProvider) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("[Ollama] 检查可用性，URL: %s", o.baseURL+"/api/tags")

	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		log.Printf("[Ollama] 创建请求失败: %v", err)
		return false
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		log.Printf("[Ollama] 请求失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	available := resp.StatusCode == http.StatusOK
	log.Printf("[Ollama] 可用性检查结果: %v (状态码: %d)", available, resp.StatusCode)
	return available
}

// GetUsage returns usage statistics
func (o *OllamaProvider) GetUsage() Usage {
	return o.usage
}

// PullModel pulls a model from Ollama registry
func (o *OllamaProvider) PullModel(ctx context.Context, modelName string) (<-chan PullProgress, error) {
	progressChan := make(chan PullProgress, 10)

	go func() {
		defer close(progressChan)

		// Prepare request
		reqBody := map[string]interface{}{
			"name":   modelName,
			"stream": true,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			progressChan <- PullProgress{Error: fmt.Errorf("failed to marshal request: %w", err)}
			return
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/pull", bytes.NewBuffer(jsonData))
		if err != nil {
			progressChan <- PullProgress{Error: fmt.Errorf("failed to create request: %w", err)}
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := o.httpClient.Do(req)
		if err != nil {
			progressChan <- PullProgress{Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			progressChan <- PullProgress{Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		// Read streaming response
		decoder := json.NewDecoder(resp.Body)

		for {
			var chunk struct {
				Status    string `json:"status"`
				Digest    string `json:"digest,omitempty"`
				Total     int64  `json:"total,omitempty"`
				Completed int64  `json:"completed,omitempty"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				progressChan <- PullProgress{Error: fmt.Errorf("failed to decode chunk: %w", err)}
				return
			}

			progress := PullProgress{
				Status: chunk.Status,
				Digest: chunk.Digest,
			}

			// Calculate percentage if we have total and completed
			if chunk.Total > 0 {
				progress.Total = chunk.Total
				progress.Completed = chunk.Completed
				progress.Percentage = float64(chunk.Completed) / float64(chunk.Total) * 100
			}

			progressChan <- progress

			// Check if pull is complete
			if chunk.Status == "success" || strings.Contains(chunk.Status, "already exists") {
				break
			}
		}
	}()

	return progressChan, nil
}

// DeleteModel deletes a model from Ollama
func (o *OllamaProvider) DeleteModel(ctx context.Context, modelName string) error {
	reqBody := map[string]interface{}{
		"name": modelName,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", o.baseURL+"/api/delete", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAvailableModels returns a list of popular models that can be pulled
func (o *OllamaProvider) GetAvailableModels() []ModelInfo {
	return []ModelInfo{
		{Name: "huihui_ai/deepseek-r1-abliterated:32b", Description: "DeepSeek R1 Abliterated 32B model (recommended)", Size: "20GB", Tags: []string{"chat", "reasoning", "large"}},
		{Name: "huihui_ai/deepseek-r1-abliterated:14b", Description: "DeepSeek R1 Abliterated 14B model", Size: "8.7GB", Tags: []string{"chat", "reasoning", "medium"}},
		{Name: "deepseek-r1:1.5b", Description: "DeepSeek R1 1.5B model", Size: "1.0GB", Tags: []string{"chat", "reasoning", "small"}},
		{Name: "llama3.2:latest", Description: "Meta's Llama 3.2 model (latest)", Size: "2.0GB", Tags: []string{"chat", "general"}},
		{Name: "llama3.2:1b", Description: "Meta's Llama 3.2 1B model", Size: "1.3GB", Tags: []string{"chat", "small"}},
		{Name: "llama3.2:3b", Description: "Meta's Llama 3.2 3B model", Size: "2.0GB", Tags: []string{"chat", "medium"}},
		{Name: "llama3.1:latest", Description: "Meta's Llama 3.1 model (latest)", Size: "4.7GB", Tags: []string{"chat", "general"}},
		{Name: "llama3.1:8b", Description: "Meta's Llama 3.1 8B model", Size: "4.7GB", Tags: []string{"chat", "medium"}},
		{Name: "llama3.1:70b", Description: "Meta's Llama 3.1 70B model", Size: "40GB", Tags: []string{"chat", "large"}},
		{Name: "codellama:latest", Description: "Meta's Code Llama model", Size: "3.8GB", Tags: []string{"code", "programming"}},
		{Name: "codellama:7b", Description: "Meta's Code Llama 7B model", Size: "3.8GB", Tags: []string{"code", "programming"}},
		{Name: "codellama:13b", Description: "Meta's Code Llama 13B model", Size: "7.3GB", Tags: []string{"code", "programming"}},
		{Name: "mistral:latest", Description: "Mistral AI's Mistral model", Size: "4.1GB", Tags: []string{"chat", "general"}},
		{Name: "mistral:7b", Description: "Mistral AI's Mistral 7B model", Size: "4.1GB", Tags: []string{"chat", "general"}},
		{Name: "mixtral:latest", Description: "Mistral AI's Mixtral model", Size: "26GB", Tags: []string{"chat", "large"}},
		{Name: "gemma2:latest", Description: "Google's Gemma 2 model", Size: "5.4GB", Tags: []string{"chat", "general"}},
		{Name: "gemma2:2b", Description: "Google's Gemma 2 2B model", Size: "1.6GB", Tags: []string{"chat", "small"}},
		{Name: "gemma2:9b", Description: "Google's Gemma 2 9B model", Size: "5.4GB", Tags: []string{"chat", "medium"}},
		{Name: "qwen2.5:latest", Description: "Alibaba's Qwen 2.5 model", Size: "4.4GB", Tags: []string{"chat", "general"}},
		{Name: "qwen2.5:7b", Description: "Alibaba's Qwen 2.5 7B model", Size: "4.4GB", Tags: []string{"chat", "general"}},
		{Name: "qwen2.5:14b", Description: "Alibaba's Qwen 2.5 14B model", Size: "8.7GB", Tags: []string{"chat", "medium"}},
		{Name: "qwen2.5-coder:latest", Description: "Alibaba's Qwen 2.5 Coder model", Size: "4.4GB", Tags: []string{"code", "programming"}},
		{Name: "deepseek-coder:latest", Description: "DeepSeek's Coder model", Size: "3.8GB", Tags: []string{"code", "programming"}},
		{Name: "deepseek-coder:6.7b", Description: "DeepSeek's Coder 6.7B model", Size: "3.8GB", Tags: []string{"code", "programming"}},
		{Name: "phi3:latest", Description: "Microsoft's Phi-3 model", Size: "2.3GB", Tags: []string{"chat", "small"}},
		{Name: "phi3:mini", Description: "Microsoft's Phi-3 Mini model", Size: "2.3GB", Tags: []string{"chat", "small"}},
		{Name: "nomic-embed-text:latest", Description: "Nomic's text embedding model", Size: "274MB", Tags: []string{"embedding", "text"}},
	}
}
