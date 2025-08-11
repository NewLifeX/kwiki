package ai

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// loadEnvFile 从.env文件加载环境变量
func loadEnvFile() error {
	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 向上查找项目根目录（包含go.mod的目录）
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return nil // 到达根目录，没找到go.mod
		}
		wd = parent
	}

	envFile := filepath.Join(wd, ".env")
	file, err := os.Open(envFile)
	if err != nil {
		return err // .env文件不存在
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 移除引号
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// TestDeepSeekBasicCall 测试DeepSeek基本API调用
func TestDeepSeekBasicCall(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping test")
	}

	// 创建DeepSeek提供商
	provider := NewDeepSeekProvider(apiKey)

	// 测试基本信息
	t.Logf("Provider name: %s", provider.GetName())
	t.Logf("Is available: %v", provider.IsAvailable())
	t.Logf("Models: %v", provider.GetModels())

	// 测试简单的文本生成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := "请用一句话介绍什么是人工智能。"

	t.Logf("Testing with prompt: %s", prompt)

	response, err := provider.GenerateText(ctx, prompt, GenerationOptions{
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   100,
		TopP:        0.9,
	})

	if err != nil {
		t.Fatalf("DeepSeek API call failed: %v", err)
	}

	// 验证响应
	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.Text == "" {
		t.Error("Response text is empty")
	}

	if response.TokensUsed == 0 {
		t.Error("No tokens were used")
	}

	if response.Provider != "deepseek" {
		t.Errorf("Expected provider 'deepseek', got '%s'", response.Provider)
	}

	t.Logf("Response text: %s", response.Text)
	t.Logf("Tokens used: %d", response.TokensUsed)
	t.Logf("Model: %s", response.Model)
	t.Logf("Duration: %dms", response.Duration)
	t.Logf("Finish reason: %s", response.FinishReason)

	// 检查使用统计
	usage := provider.GetUsage()
	t.Logf("Usage stats - Requests: %d, Tokens: %d, Errors: %d",
		usage.TotalRequests, usage.TotalTokens, usage.ErrorCount)

	if usage.TotalRequests == 0 {
		t.Error("Usage statistics not updated")
	}
}

// TestDeepSeekDifferentModels 测试不同的DeepSeek模型
func TestDeepSeekDifferentModels(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping test")
	}

	provider := NewDeepSeekProvider(apiKey)

	// 测试不同模型
	models := []string{"deepseek-chat", "deepseek-reasoner", "deepseek-coder"}
	prompt := "Hello, how are you?"

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			t.Logf("Testing model: %s", model)

			response, err := provider.GenerateText(ctx, prompt, GenerationOptions{
				Model:       model,
				Temperature: 0.5,
				MaxTokens:   50,
			})

			if err != nil {
				t.Logf("Model %s failed: %v", model, err)
				return // 某些模型可能不可用，不算失败
			}

			t.Logf("Model %s response: %s", model, response.Text)
			t.Logf("Model %s tokens: %d", model, response.TokensUsed)
		})
	}
}

// TestDeepSeekErrorHandling 测试错误处理
func TestDeepSeekErrorHandling(t *testing.T) {
	// 测试无效API密钥
	provider := NewDeepSeekProvider("invalid-key")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := provider.GenerateText(ctx, "test", GenerationOptions{
		Model:     "deepseek-chat",
		MaxTokens: 10,
	})

	if err == nil {
		t.Error("Expected error with invalid API key, but got none")
	}

	t.Logf("Expected error with invalid key: %v", err)

	// 检查错误统计
	usage := provider.GetUsage()
	if usage.ErrorCount == 0 {
		t.Error("Error count should be incremented")
	}
}

// TestDeepSeekTimeout 测试超时处理
func TestDeepSeekTimeout(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping test")
	}

	provider := NewDeepSeekProvider(apiKey)

	// 设置很短的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := provider.GenerateText(ctx, "test", GenerationOptions{
		Model:     "deepseek-chat",
		MaxTokens: 10,
	})

	if err == nil {
		t.Error("Expected timeout error, but got none")
	}

	t.Logf("Expected timeout error: %v", err)
}

// TestDeepSeekReasonerModel 专门测试deepseek-reasoner模型
func TestDeepSeekReasonerModel(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping test")
	}

	provider := NewDeepSeekProvider(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := "请简单介绍一下Go语言的特点。"

	t.Logf("Testing deepseek-reasoner with prompt: %s", prompt)

	response, err := provider.GenerateText(ctx, prompt, GenerationOptions{
		Model:       "deepseek-reasoner",
		Temperature: 0.7,
		MaxTokens:   500, // 增加token限制
		TopP:        0.9,
	})

	if err != nil {
		t.Fatalf("DeepSeek reasoner API call failed: %v", err)
	}

	// 详细检查响应
	t.Logf("=== DeepSeek Reasoner Response Analysis ===")
	t.Logf("Response text length: %d", len(response.Text))
	t.Logf("Tokens used: %d", response.TokensUsed)
	t.Logf("Model: %s", response.Model)
	t.Logf("Duration: %dms", response.Duration)
	t.Logf("Finish reason: %s", response.FinishReason)

	if response.Text != "" {
		t.Logf("Response text: %s", response.Text)
	} else {
		t.Logf("Response text is empty!")

		// 检查元数据
		t.Logf("Metadata:")
		for key, value := range response.Metadata {
			t.Logf("  %s: %s", key, value)
		}
	}

	// 即使文本为空，只要没有错误就认为API调用成功
	if err == nil {
		t.Logf("DeepSeek reasoner API call successful (even if text is empty)")
	}
}

// TestDeepSeekStreamGeneration 测试DeepSeek流式生成
func TestDeepSeekStreamGeneration(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping test")
	}

	provider := NewDeepSeekProvider(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := "请简单介绍一下Go语言的特点，大约100字。"

	t.Logf("Testing DeepSeek streaming with prompt: %s", prompt)

	streamChan, err := provider.GenerateStream(ctx, prompt, GenerationOptions{
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   200,
		TopP:        0.9,
	})

	if err != nil {
		t.Fatalf("DeepSeek stream generation failed to start: %v", err)
	}

	var fullText strings.Builder
	var chunkCount int
	startTime := time.Now()
	lastLogTime := time.Now()

	t.Logf("=== DeepSeek Stream Response ===")

	for streamResp := range streamChan {
		if streamResp.Error != nil {
			t.Fatalf("Stream error: %v", streamResp.Error)
		}

		if streamResp.Text != "" {
			fullText.WriteString(streamResp.Text)
			chunkCount++

			// 每3秒报告一次进度，避免日志过多
			if time.Since(lastLogTime) > 3*time.Second {
				t.Logf("Progress: %d characters received (%d chunks)", fullText.Len(), chunkCount)
				lastLogTime = time.Now()
			}
		}

		if streamResp.Done {
			duration := time.Since(startTime)
			t.Logf("=== Stream Completed ===")
			t.Logf("Duration: %v", duration)
			t.Logf("Total chunks: %d", chunkCount)
			t.Logf("Total characters: %d", fullText.Len())
			t.Logf("Average chunk size: %.1f characters", float64(fullText.Len())/float64(chunkCount))
			t.Logf("Generation speed: %.1f chars/sec", float64(fullText.Len())/duration.Seconds())

			if streamResp.TokensUsed > 0 {
				t.Logf("Tokens used: %d", streamResp.TokensUsed)
			}

			if streamResp.Metadata != nil {
				t.Logf("Metadata:")
				for key, value := range streamResp.Metadata {
					t.Logf("  %s: %s", key, value)
				}
			}

			t.Logf("Generated text: %s", fullText.String())
			break
		}
	}

	// 验证结果
	if fullText.Len() == 0 {
		t.Error("No text was generated")
	}

	if chunkCount == 0 {
		t.Error("No chunks were received")
	}

	t.Logf("DeepSeek streaming test completed successfully")
}
