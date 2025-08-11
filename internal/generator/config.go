package generator

// GeneratorConfig 生成器配置
type GeneratorConfig struct {
	ReadingSpeed int    `yaml:"reading_speed"` // 每分钟阅读单词数
	TemplateDir  string `yaml:"template_dir"`  // 模板目录路径
}

// LoadGeneratorConfig 加载生成器配置
func LoadGeneratorConfig(configPath string) (*GeneratorConfig, error) {
	// 直接返回默认配置，简化逻辑
	return getDefaultConfig(), nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *GeneratorConfig {
	return &GeneratorConfig{
		ReadingSpeed: 200,
		TemplateDir:  "templates/prompts",
	}
}
