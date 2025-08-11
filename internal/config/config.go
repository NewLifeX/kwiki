package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	AI         AIConfig         `yaml:"ai"`
	Repository RepositoryConfig `yaml:"repository"`
	Generator  GeneratorConfig  `yaml:"generator"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port        string `yaml:"port"`
	Host        string `yaml:"host"`
	StaticDir   string `yaml:"static_dir"`
	TemplateDir string `yaml:"template_dir"`
	DataDir     string `yaml:"data_dir"`
	EnableCORS  bool   `yaml:"enable_cors"`
	MaxFileSize int64  `yaml:"max_file_size"`
}

// AIConfig contains AI provider configuration
type AIConfig struct {
	DefaultProvider string                `yaml:"default_provider"`
	Providers       map[string]AIProvider `yaml:"providers"`
}

// AIProvider represents configuration for an AI provider
type AIProvider struct {
	APIKey      string            `yaml:"api_key"`
	BaseURL     string            `yaml:"base_url,omitempty"`
	Model       string            `yaml:"model"`
	Temperature float32           `yaml:"temperature"`
	MaxTokens   int               `yaml:"max_tokens"`
	Extra       map[string]string `yaml:"extra,omitempty"`
}

// RepositoryConfig contains repository handling configuration
type RepositoryConfig struct {
	CloneDir        string   `yaml:"clone_dir"`
	MaxRepoSize     int64    `yaml:"max_repo_size"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	IncludePatterns []string `yaml:"include_patterns"`
	MaxFiles        int      `yaml:"max_files"`
}

// GeneratorConfig contains documentation generation configuration
type GeneratorConfig struct {
	OutputDir      string `yaml:"output_dir"`
	EnableDiagrams bool   `yaml:"enable_diagrams"`
	EnableRAG      bool   `yaml:"enable_rag"`
	ChunkSize      int    `yaml:"chunk_size"`
	ChunkOverlap   int    `yaml:"chunk_overlap"`
	MaxConcurrency int    `yaml:"max_concurrency"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Override with environment variables
	config.overrideWithEnv()

	return &config, nil
}

// Default returns a default configuration
func Default() *Config {
	config := &Config{
		Server: ServerConfig{
			Port:        "8080",
			Host:        "localhost",
			StaticDir:   "web/static",
			TemplateDir: "web/templates",
			DataDir:     "./data",
			EnableCORS:  true,
			MaxFileSize: 100 * 1024 * 1024, // 100MB
		},
		AI: AIConfig{
			DefaultProvider: "ollama",
			Providers: map[string]AIProvider{
				"ollama": {
					Model:       "huihui_ai/deepseek-r1-abliterated:32b",
					Temperature: 0.7,
					MaxTokens:   8000,
				},
				"openai": {
					Model:       "gpt-4o-mini",
					Temperature: 0.7,
					MaxTokens:   4000,
				},
				"gemini": {
					Model:       "gemini-2.0-flash-exp",
					Temperature: 0.7,
					MaxTokens:   4000,
				},
				"deepseek": {
					Model:       "deepseek-chat",
					Temperature: 0.7,
					MaxTokens:   8000,
				},
			},
		},
		Repository: RepositoryConfig{
			CloneDir:    "./repos",
			MaxRepoSize: 500 * 1024 * 1024, // 500MB
			ExcludePatterns: []string{
				"node_modules",
				".git",
				"vendor",
				"target",
				"build",
				"dist",
				"*.log",
				"*.tmp",
			},
			IncludePatterns: []string{
				"*.go",
				"*.py",
				"*.js",
				"*.ts",
				"*.java",
				"*.cpp",
				"*.c",
				"*.h",
				"*.rs",
				"*.rb",
				"*.php",
				"*.cs",
				"*.md",
				"*.txt",
				"*.yaml",
				"*.yml",
				"*.json",
				"*.toml",
			},
			MaxFiles: 10000,
		},
		Generator: GeneratorConfig{
			OutputDir:      "./output",
			EnableDiagrams: true,
			EnableRAG:      true,
			ChunkSize:      1000,
			ChunkOverlap:   200,
			MaxConcurrency: 5,
		},
	}

	// Override with environment variables
	config.overrideWithEnv()

	return config
}

// overrideWithEnv overrides configuration with environment variables
func (c *Config) overrideWithEnv() {
	if port := os.Getenv("PORT"); port != "" {
		c.Server.Port = port
	}

	if host := os.Getenv("HOST"); host != "" {
		c.Server.Host = host
	}

	// AI provider API keys
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		if provider, exists := c.AI.Providers["openai"]; exists {
			provider.APIKey = openaiKey
			c.AI.Providers["openai"] = provider
		}
	}

	if geminiKey := os.Getenv("GOOGLE_API_KEY"); geminiKey != "" {
		if provider, exists := c.AI.Providers["gemini"]; exists {
			provider.APIKey = geminiKey
			c.AI.Providers["gemini"] = provider
		}
	}

	if deepseekKey := os.Getenv("DEEPSEEK_API_KEY"); deepseekKey != "" {
		if provider, exists := c.AI.Providers["deepseek"]; exists {
			provider.APIKey = deepseekKey
			c.AI.Providers["deepseek"] = provider
		}
	}

	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		if provider, exists := c.AI.Providers["openai"]; exists {
			provider.BaseURL = baseURL
			c.AI.Providers["openai"] = provider
		}
	}
}

// Save saves the configuration to a YAML file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
