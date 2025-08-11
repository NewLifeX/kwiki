package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/stcn52/kwiki/pkg/models"

	"gopkg.in/yaml.v3"
)

// TemplateManager manages prompt templates
type TemplateManager struct {
	templates map[string]*template.Template
	config    *GeneratorConfig
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(config *GeneratorConfig) *TemplateManager {
	if config == nil {
		config = &GeneratorConfig{
			ReadingSpeed: 200,
			TemplateDir:  "templates/prompts",
		}
	}
	return &TemplateManager{
		templates: make(map[string]*template.Template),
		config:    config,
	}
}

// TemplateData represents data passed to templates
type TemplateData struct {
	ProjectName     string
	Description     string
	PrimaryLanguage string
	License         string
	Language        string
	Modules         []ModuleData
}

// ModuleData represents module data for templates
type ModuleData struct {
	Name        string
	Description string
	Functions   []FunctionData
}

// FunctionData represents function data for templates
type FunctionData struct {
	Name        string
	Description string
}

// TemplateMetadata represents metadata from template front matter
type TemplateMetadata struct {
	Title       string   `yaml:"title"`
	Type        string   `yaml:"type"`
	Order       int      `yaml:"order"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
	Tags        []string `yaml:"tags"`
	Language    string   `yaml:"language"`
	Variables   []string `yaml:"variables"`
}

// TemplateInfo combines template content with metadata
type TemplateInfo struct {
	Metadata TemplateMetadata
	Content  string
	Template *template.Template
}

// LoadTemplate loads a template for a specific language and type
func (tm *TemplateManager) LoadTemplate(language, templateType string) (*template.Template, error) {
	key := fmt.Sprintf("%s/%s", language, templateType)

	// Check if template is already loaded
	if tmpl, exists := tm.templates[key]; exists {
		return tmpl, nil
	}

	// Try to load from file system
	templatePath := filepath.Join(tm.config.TemplateDir, language, templateType+".md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// Fallback to English if language not found
		if language != "en" {
			return tm.LoadTemplate("en", templateType)
		}
		return nil, fmt.Errorf("template not found: %s (path: %s)", key, templatePath)
	}

	tmpl, err := template.New(key).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", key, err)
	}

	tm.templates[key] = tmpl
	return tmpl, nil
}

// RenderTemplate renders a template with the given data
func (tm *TemplateManager) RenderTemplate(language, templateType string, data TemplateData) (string, error) {
	tmpl, err := tm.LoadTemplate(language, templateType)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// PrepareTemplateData prepares template data from repository and structure
func (tm *TemplateManager) PrepareTemplateData(repo *models.Repository, structure *models.CodeStructure, language string) TemplateData {
	data := TemplateData{
		ProjectName:     repo.Name,
		Description:     repo.Description,
		PrimaryLanguage: tm.getPrimaryLanguage(repo),
		License:         repo.License,
		Language:        language,
		Modules:         make([]ModuleData, 0, len(structure.Modules)),
	}

	// Convert modules
	for _, module := range structure.Modules {
		moduleData := ModuleData{
			Name:        module.Name,
			Description: module.Description,
			Functions:   make([]FunctionData, 0, len(module.Functions)),
		}

		// Convert functions (limit to prevent template from being too long)
		for i, funcName := range module.Functions {
			if i >= 10 { // Limit to 10 functions per module
				break
			}

			// Find function details
			for _, fn := range structure.Functions {
				if fn.Name == funcName && fn.Module == module.Name {
					moduleData.Functions = append(moduleData.Functions, FunctionData{
						Name:        fn.Name,
						Description: fn.Description,
					})
					break
				}
			}
		}

		data.Modules = append(data.Modules, moduleData)
	}

	return data
}

// getPrimaryLanguage gets the primary language from repository
func (tm *TemplateManager) getPrimaryLanguage(repo *models.Repository) string {
	if len(repo.Languages) > 0 {
		return repo.Languages[0]
	}
	return "Unknown"
}

// GetAvailableTemplates returns available template types for a language
func (tm *TemplateManager) GetAvailableTemplates(language string) ([]string, error) {
	var templates []string

	// Check file system templates
	templateDir := filepath.Join(tm.config.TemplateDir, language)
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		// Fallback to English
		if language != "en" {
			return tm.GetAvailableTemplates("en")
		}
		return nil, fmt.Errorf("no templates found for language: %s (dir: %s)", language, templateDir)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			templateName := strings.TrimSuffix(entry.Name(), ".md")
			templates = append(templates, templateName)
		}
	}

	return templates, nil
}

// GetSupportedLanguages returns list of supported languages
func (tm *TemplateManager) GetSupportedLanguages() []string {
	return []string{"en", "zh", "ja", "ko", "es", "fr", "de", "ru", "pt", "it"}
}

// ValidateTemplate validates a template string
func (tm *TemplateManager) ValidateTemplate(content string) error {
	_, err := template.New("validation").Parse(content)
	return err
}

// LoadTemplateWithMetadata loads a template with its metadata
func (tm *TemplateManager) LoadTemplateWithMetadata(language, templateType string) (*TemplateInfo, error) {
	templatePath := filepath.Join(tm.config.TemplateDir, language, templateType+".md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// Fallback to English if language not found
		if language != "en" {
			return tm.LoadTemplateWithMetadata("en", templateType)
		}
		return nil, fmt.Errorf("template not found: %s (path: %s)", templateType, templatePath)
	}

	// Parse front matter and content
	metadata, templateContent, err := tm.parseFrontMatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter for %s: %w", templateType, err)
	}

	// Create template
	tmpl, err := template.New(templateType).Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateType, err)
	}

	return &TemplateInfo{
		Metadata: metadata,
		Content:  templateContent,
		Template: tmpl,
	}, nil
}

// parseFrontMatter parses YAML front matter from template content
func (tm *TemplateManager) parseFrontMatter(content string) (TemplateMetadata, string, error) {
	var metadata TemplateMetadata

	// Check if content starts with front matter
	if !strings.HasPrefix(content, "---\n") {
		// No front matter, return default metadata
		return TemplateMetadata{
			Title: "Untitled",
			Type:  "guide",
			Order: 999,
		}, content, nil
	}

	// Find the end of front matter
	parts := strings.SplitN(content, "\n---\n", 2)
	if len(parts) != 2 {
		return metadata, content, fmt.Errorf("invalid front matter format")
	}

	// Parse YAML front matter
	frontMatter := strings.TrimPrefix(parts[0], "---\n")
	if err := yaml.Unmarshal([]byte(frontMatter), &metadata); err != nil {
		return metadata, content, fmt.Errorf("failed to parse YAML front matter: %w", err)
	}

	return metadata, parts[1], nil
}

// GetTemplatesWithMetadata returns all templates for a language with their metadata, sorted by order
func (tm *TemplateManager) GetTemplatesWithMetadata(language string) ([]*TemplateInfo, error) {
	log.Printf("获取语言 %s 的模板列表", language)

	templateTypes, err := tm.GetAvailableTemplates(language)
	if err != nil {
		log.Printf("获取语言 %s 的可用模板类型失败: %v", language, err)
		return nil, err
	}

	log.Printf("语言 %s 找到 %d 个模板类型: %v", language, len(templateTypes), templateTypes)

	var templates []*TemplateInfo
	for _, templateType := range templateTypes {
		log.Printf("加载模板: %s/%s", language, templateType)
		templateInfo, err := tm.LoadTemplateWithMetadata(language, templateType)
		if err != nil {
			log.Printf("加载模板失败: %s/%s, 错误: %v", language, templateType, err)
			continue // Skip templates that fail to load
		}
		log.Printf("成功加载模板: %s/%s (标题: %s, 类型: %s, 顺序: %d)",
			language, templateType, templateInfo.Metadata.Title, templateInfo.Metadata.Type, templateInfo.Metadata.Order)
		templates = append(templates, templateInfo)
	}

	// Sort by order
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Metadata.Order < templates[j].Metadata.Order
	})

	log.Printf("语言 %s 模板加载完成，共 %d 个模板", language, len(templates))
	return templates, nil
}

// TemplateDocumentationData represents data for template documentation generation
type TemplateDocumentationData struct {
	ProjectName     string
	Description     string
	PrimaryLanguage string
	License         string
	Language        string
	Templates       []TemplateDocInfo
	Languages       []LanguageInfo
	Statistics      TemplateStatistics
}

// TemplateDocInfo represents information about a template
type TemplateDocInfo struct {
	Name        string
	Title       string
	Type        string
	Order       int
	Language    string
	Path        string
	Description string
	Variables   []string
	Examples    []string
}

// LanguageInfo represents information about a supported language
type LanguageInfo struct {
	Code          string
	Name          string
	TemplateCount int
	Available     bool
}

// TemplateStatistics represents statistics about templates
type TemplateStatistics struct {
	TotalTemplates      int
	TotalLanguages      int
	TemplatesByType     map[string]int
	TemplatesByLanguage map[string]int
}

// ScanTemplateDirectory 扫描模板目录并收集所有模板信息
func (tm *TemplateManager) ScanTemplateDirectory() (*TemplateDocumentationData, error) {
	supportedLanguages := tm.GetSupportedLanguages()

	// 初始化数据结构
	docData := &TemplateDocumentationData{
		ProjectName:     "KWiki Template System",
		Description:     "KWiki 提示词模板系统文档",
		PrimaryLanguage: "Go",
		License:         "MIT",
		Language:        "zh",
		Templates:       []TemplateDocInfo{},
		Languages:       []LanguageInfo{},
		Statistics: TemplateStatistics{
			TemplatesByType:     make(map[string]int),
			TemplatesByLanguage: make(map[string]int),
		},
	}

	// 扫描每种语言的模板
	for _, langCode := range supportedLanguages {
		langInfo := LanguageInfo{
			Code:          langCode,
			Name:          tm.getLanguageName(langCode),
			TemplateCount: 0,
			Available:     false,
		}

		// 获取该语言的所有模板
		templates, err := tm.GetTemplatesWithMetadata(langCode)
		if err != nil {
			// 如果语言目录不存在，跳过
			docData.Languages = append(docData.Languages, langInfo)
			continue
		}

		langInfo.Available = true
		langInfo.TemplateCount = len(templates)
		docData.Statistics.TemplatesByLanguage[langCode] = len(templates)

		// 处理每个模板，直接从模板元数据获取信息
		for _, tmplInfo := range templates {
			docInfo := TemplateDocInfo{
				Name:        tmplInfo.Metadata.Title,
				Title:       tmplInfo.Metadata.Title,
				Type:        tmplInfo.Metadata.Type,
				Order:       tmplInfo.Metadata.Order,
				Language:    langCode,
				Path:        fmt.Sprintf("%s/%s/%s.md", tm.config.TemplateDir, langCode, tmplInfo.Metadata.Title),
				Description: tmplInfo.Metadata.Description,
				Variables:   tmplInfo.Metadata.Variables,
				Examples:    []string{}, // 示例可以从模板内容中提取，但暂时简化
			}

			docData.Templates = append(docData.Templates, docInfo)
			docData.Statistics.TemplatesByType[tmplInfo.Metadata.Type]++
		}

		docData.Languages = append(docData.Languages, langInfo)
	}

	// 计算总体统计
	docData.Statistics.TotalTemplates = len(docData.Templates)
	docData.Statistics.TotalLanguages = len(docData.Languages)

	return docData, nil
}

// getLanguageName 获取语言代码对应的显示名称
func (tm *TemplateManager) getLanguageName(code string) string {
	// 简单映射，主要支持中英文
	switch code {
	case "zh":
		return "中文"
	case "en":
		return "English"
	default:
		return code
	}
}
