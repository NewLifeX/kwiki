package generator

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stcn52/kwiki/internal/ai"
	"github.com/stcn52/kwiki/internal/config"
	"github.com/stcn52/kwiki/pkg/models"
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

// TestTemplateDocumentationGeneration 测试模板文档生成
func TestTemplateDocumentationGeneration(t *testing.T) {
	// 加载.env文件
	loadEnvFile()

	// 检查是否有DeepSeek API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set in .env file, skipping integration test")
	}

	// 获取项目根目录
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// 向上查找项目根目录（包含go.mod的目录）
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("Could not find project root directory")
		}
		wd = parent
	}

	templateDir := filepath.Join(wd, "templates", "prompts")

	// 创建配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Generator: config.GeneratorConfig{
			EnableRAG:      false,
			EnableDiagrams: false,
		},
	}

	// 创建AI管理器并手动注册DeepSeek提供商
	aiManager := ai.NewProviderManager()

	// 手动创建DeepSeek提供商
	deepseekProvider := ai.NewDeepSeekProvider(apiKey)
	aiManager.RegisterProvider("deepseek", deepseekProvider)
	aiManager.SetDefaultProvider("deepseek")

	// 创建生成器配置，使用正确的模板目录
	generatorConfig := &GeneratorConfig{
		ReadingSpeed: 200,
		TemplateDir:  templateDir,
	}

	// 创建WikiGenerator，传递正确的模板目录配置
	wikiGen := &WikiGenerator{
		config:          cfg,
		aiManager:       aiManager,
		templateManager: NewTemplateManager(generatorConfig),
	}

	// 创建模板文档生成请求
	req := models.GenerationRequest{
		RepositoryURL:   "template-docs",
		Title:           "KWiki Template System Documentation",
		Description:     "Documentation for KWiki template system",
		PrimaryLanguage: "zh",
		Languages:       []string{"zh"},
		Settings: models.WikiSettings{
			AIProvider: "deepseek",
			Model:      "deepseek-chat", // 使用稳定的chat模型
		},
	}

	// 生成wiki
	wiki, err := wikiGen.GenerateWiki(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to generate wiki: %v", err)
	}

	// 验证基本属性
	if wiki.ID == "" {
		t.Error("Wiki ID should not be empty")
	}

	if wiki.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, wiki.Title)
	}

	if wiki.Status != models.WikiStatusGenerating {
		t.Errorf("Expected status %s, got %s", models.WikiStatusGenerating, wiki.Status)
	}

	// 等待生成完成（增加到5分钟，因为需要生成多个页面）
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Wiki generation timed out. Status: %s, Progress: %d%%, Pages: %d",
				wiki.Status, wiki.Progress, len(wiki.Pages))
		case <-ticker.C:
			t.Logf("Wiki status: %s, Progress: %d%%, Pages: %d",
				wiki.Status, wiki.Progress, len(wiki.Pages))

			if wiki.Status == models.WikiStatusCompleted {
				// 验证生成结果
				if len(wiki.Pages) == 0 {
					t.Error("Expected at least one page to be generated")
				}

				// 验证页面内容
				for _, page := range wiki.Pages {
					if page.Title == "" {
						t.Error("Page title should not be empty")
					}
					if page.Content == "" {
						t.Error("Page content should not be empty")
					}
					if page.WordCount == 0 {
						t.Error("Page word count should be greater than 0")
					}
					t.Logf("Generated page: %s (%d words)", page.Title, page.WordCount)
				}

				t.Logf("Successfully generated %d pages", len(wiki.Pages))
				return
			}
			if wiki.Status == models.WikiStatusFailed {
				t.Fatal("Wiki generation failed")
			}
		}
	}
}

// TestTemplateManager 测试模板管理器
func TestTemplateManager(t *testing.T) {
	config := &GeneratorConfig{
		ReadingSpeed: 200,
		TemplateDir:  "../../templates/prompts",
	}

	tm := NewTemplateManager(config)

	// 测试获取支持的语言
	languages := tm.GetSupportedLanguages()
	if len(languages) == 0 {
		t.Error("Expected at least one supported language")
	}

	// 测试扫描模板目录
	data, err := tm.ScanTemplateDirectory()
	if err != nil {
		t.Fatalf("Failed to scan template directory: %v", err)
	}

	if data.ProjectName == "" {
		t.Error("Project name should not be empty")
	}

	if data.Statistics.TotalLanguages == 0 {
		t.Error("Expected at least one language")
	}

	t.Logf("Found %d templates across %d languages",
		data.Statistics.TotalTemplates, data.Statistics.TotalLanguages)
}

// TestGeneratorConfig 测试生成器配置
func TestGeneratorConfig(t *testing.T) {
	// 测试加载默认配置
	config, err := LoadGeneratorConfig("nonexistent.yaml")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.ReadingSpeed != 200 {
		t.Errorf("Expected reading speed 200, got %d", config.ReadingSpeed)
	}

	if config.TemplateDir != "templates/prompts" {
		t.Errorf("Expected template dir 'templates/prompts', got %s", config.TemplateDir)
	}
}
