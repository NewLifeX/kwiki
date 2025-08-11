package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stcn52/kwiki/pkg/models"
)

// MarkdownStorage 基于Markdown文件的存储实现
type MarkdownStorage struct {
	baseDir string
}

// WikiMetadata Wiki的元数据结构
type WikiMetadata struct {
	ID           string              `json:"id"`
	RepositoryID string              `json:"repository_id"`
	PackagePath  string              `json:"package_path"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	Status       models.WikiStatus   `json:"status"`
	Progress     int                 `json:"progress"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Languages    []string            `json:"languages"`
	Settings     models.WikiSettings `json:"settings"`
	Metadata     models.WikiMetadata `json:"metadata"`
	Tags         []string            `json:"tags"`
}

// NewMarkdownStorage 创建新的Markdown存储实例
func NewMarkdownStorage(baseDir string) *MarkdownStorage {
	return &MarkdownStorage{
		baseDir: baseDir,
	}
}

// getWikiPath 根据Wiki信息生成存储路径
func (ms *MarkdownStorage) getWikiPath(wiki *models.Wiki) string {
	// 如果有PackagePath，使用包路径
	if wiki.PackagePath != "" {
		return ms.sanitizePath(wiki.PackagePath)
	}

	// 如果有RepositoryID，尝试从中提取路径
	if wiki.RepositoryID != "" {
		return ms.extractPathFromRepository(wiki.RepositoryID)
	}

	// 特殊处理模板文档
	if wiki.ID == "template-docs" || strings.Contains(wiki.ID, "template") {
		return "template-docs"
	}

	// 默认使用Wiki ID
	return ms.sanitizePath(wiki.ID)
}

// extractPathFromRepository 从仓库URL中提取路径
func (ms *MarkdownStorage) extractPathFromRepository(repoURL string) string {
	// 移除协议前缀
	path := strings.TrimPrefix(repoURL, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "git@")

	// 处理GitHub URL
	if strings.Contains(path, "github.com") {
		// github.com/user/repo.git -> github.com/user/repo
		path = strings.TrimSuffix(path, ".git")
		// 移除tree/branch部分
		if idx := strings.Index(path, "/tree/"); idx != -1 {
			path = path[:idx]
		}
		return ms.sanitizePath(path)
	}

	// 处理GitLab URL
	if strings.Contains(path, "gitlab.com") {
		path = strings.TrimSuffix(path, ".git")
		if idx := strings.Index(path, "/-/"); idx != -1 {
			path = path[:idx]
		}
		return ms.sanitizePath(path)
	}

	// 其他情况，使用域名/路径结构
	return ms.sanitizePath(path)
}

// sanitizePath 清理路径，确保适合文件系统
func (ms *MarkdownStorage) sanitizePath(path string) string {
	// 替换不安全的字符
	path = strings.ReplaceAll(path, ":", "_")
	path = strings.ReplaceAll(path, "?", "_")
	path = strings.ReplaceAll(path, "*", "_")
	path = strings.ReplaceAll(path, "<", "_")
	path = strings.ReplaceAll(path, ">", "_")
	path = strings.ReplaceAll(path, "|", "_")
	path = strings.ReplaceAll(path, "\"", "_")

	// 移除多余的斜杠
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.Trim(path, "/")

	return path
}

// findWikiPath 查找Wiki的实际存储路径
func (ms *MarkdownStorage) findWikiPath(wikiID string) string {
	var foundPath string

	// 递归搜索所有目录，查找包含指定wikiID的meta.json
	filepath.WalkDir(ms.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 只检查meta.json文件
		if !d.IsDir() && d.Name() == "meta.json" {
			// 读取meta.json检查ID
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var metadata WikiMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				return nil
			}

			// 如果ID匹配，记录路径
			if metadata.ID == wikiID {
				// 获取相对于baseDir的路径
				relPath, err := filepath.Rel(ms.baseDir, filepath.Dir(path))
				if err == nil {
					foundPath = relPath
					return filepath.SkipAll // 找到后停止搜索
				}
			}
		}

		return nil
	})

	return foundPath
}

// SaveWiki 保存Wiki到Markdown文件结构
func (ms *MarkdownStorage) SaveWiki(wiki *models.Wiki) error {
	// 使用包路径或仓库路径作为目录结构
	wikiPath := ms.getWikiPath(wiki)
	wikiDir := filepath.Join(ms.baseDir, wikiPath)

	// 创建Wiki目录
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return fmt.Errorf("创建Wiki目录失败: %w", err)
	}

	// 保存元数据
	if err := ms.saveMetadata(wikiDir, wiki); err != nil {
		return fmt.Errorf("保存元数据失败: %w", err)
	}

	// 按语言组织页面
	pagesByLanguage := ms.groupPagesByLanguage(wiki.Pages)

	// 为每种语言创建目录并保存页面
	for language, pages := range pagesByLanguage {
		langDir := filepath.Join(wikiDir, language)
		if err := os.MkdirAll(langDir, 0755); err != nil {
			return fmt.Errorf("创建语言目录 %s 失败: %w", language, err)
		}

		for _, page := range pages {
			if err := ms.savePage(langDir, &page); err != nil {
				log.Printf("保存页面失败: %s/%s, 错误: %v", language, page.Title, err)
				continue
			}
		}
	}

	// 创建assets目录
	assetsDir := filepath.Join(wikiDir, "assets")
	if err := os.MkdirAll(filepath.Join(assetsDir, "images"), 0755); err != nil {
		return fmt.Errorf("创建assets目录失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(assetsDir, "diagrams"), 0755); err != nil {
		return fmt.Errorf("创建diagrams目录失败: %w", err)
	}

	log.Printf("Wiki %s 保存成功，目录: %s", wiki.ID, wikiDir)
	return nil
}

// saveMetadata 保存Wiki元数据
func (ms *MarkdownStorage) saveMetadata(wikiDir string, wiki *models.Wiki) error {
	metadata := WikiMetadata{
		ID:           wiki.ID,
		RepositoryID: wiki.RepositoryID,
		PackagePath:  wiki.PackagePath,
		Title:        wiki.Title,
		Description:  wiki.Description,
		Status:       wiki.Status,
		Progress:     wiki.Progress,
		CreatedAt:    wiki.CreatedAt,
		UpdatedAt:    wiki.UpdatedAt,
		Languages:    ms.extractLanguages(wiki.Pages),
		Settings:     wiki.Settings,
		Metadata:     wiki.Metadata,
		Tags:         wiki.Tags,
	}

	metaFile := filepath.Join(wikiDir, "meta.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	if err := os.WriteFile(metaFile, data, 0644); err != nil {
		return fmt.Errorf("写入元数据文件失败: %w", err)
	}

	return nil
}

// groupPagesByLanguage 按语言分组页面
func (ms *MarkdownStorage) groupPagesByLanguage(pages []models.WikiPage) map[string][]models.WikiPage {
	result := make(map[string][]models.WikiPage)

	for _, page := range pages {
		// 从页面ID中提取语言信息，格式通常是 "type_language"
		language := ms.extractLanguageFromPageID(page.ID)
		if language == "" {
			language = "zh" // 默认中文
		}

		result[language] = append(result[language], page)
	}

	return result
}

// extractLanguageFromPageID 从页面ID中提取语言
func (ms *MarkdownStorage) extractLanguageFromPageID(pageID string) string {
	parts := strings.Split(pageID, "_")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // 取最后一部分作为语言
	}
	return ""
}

// extractLanguages 从页面中提取所有语言
func (ms *MarkdownStorage) extractLanguages(pages []models.WikiPage) []string {
	languageSet := make(map[string]bool)

	for _, page := range pages {
		language := ms.extractLanguageFromPageID(page.ID)
		if language == "" {
			language = "zh"
		}
		languageSet[language] = true
	}

	var languages []string
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	return languages
}

// savePage 保存单个页面为Markdown文件
func (ms *MarkdownStorage) savePage(langDir string, page *models.WikiPage) error {
	// 生成文件名
	filename := ms.generateFilename(page)
	filepath := filepath.Join(langDir, filename)

	// 生成Markdown内容
	content := ms.generateMarkdownContent(page)

	// 写入文件
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入页面文件失败: %w", err)
	}

	log.Printf("页面保存成功: %s", filepath)
	return nil
}

// generateFilename 根据页面标题生成文件名
func (ms *MarkdownStorage) generateFilename(page *models.WikiPage) string {
	// 直接使用标题生成文件名，不硬编码
	filename := ms.slugify(page.Title)

	// 如果是overview类型且标题包含"说明"、"readme"等，使用README
	if page.Type == models.PageTypeOverview &&
		(strings.Contains(strings.ToLower(page.Title), "readme") ||
			strings.Contains(page.Title, "说明") ||
			strings.Contains(page.Title, "项目")) {
		filename = "README"
	}

	return filename + ".md"
}

// slugify 将标题转换为适合文件名的格式
func (ms *MarkdownStorage) slugify(title string) string {
	// 简单的slugify实现
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "　", "-") // 中文空格
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	slug = strings.ReplaceAll(slug, ":", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "*", "")
	slug = strings.ReplaceAll(slug, "<", "")
	slug = strings.ReplaceAll(slug, ">", "")
	slug = strings.ReplaceAll(slug, "|", "-")
	slug = strings.ReplaceAll(slug, "\"", "")

	// 移除连续的连字符
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// 移除首尾的连字符
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "untitled"
	}

	return slug
}

// generateMarkdownContent 生成Markdown内容
func (ms *MarkdownStorage) generateMarkdownContent(page *models.WikiPage) string {
	var content strings.Builder

	// 添加前置元数据
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("title: %s\n", page.Title))
	content.WriteString(fmt.Sprintf("type: %s\n", page.Type))
	content.WriteString(fmt.Sprintf("order: %d\n", page.Order))
	content.WriteString(fmt.Sprintf("word_count: %d\n", page.WordCount))
	content.WriteString(fmt.Sprintf("reading_time: %d\n", page.ReadingTime))
	content.WriteString(fmt.Sprintf("created_at: %s\n", page.CreatedAt.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("updated_at: %s\n", page.UpdatedAt.Format(time.RFC3339)))
	content.WriteString("---\n\n")

	// 添加页面内容
	content.WriteString(page.Content)

	return content.String()
}

// LoadWiki 从Markdown文件结构加载Wiki
func (ms *MarkdownStorage) LoadWiki(wikiID string) (*models.Wiki, error) {
	log.Printf("LoadWiki: 尝试加载wiki ID: %s", wikiID)

	// 首先尝试直接使用wikiID作为路径
	wikiDir := filepath.Join(ms.baseDir, wikiID)
	log.Printf("LoadWiki: 尝试路径: %s", wikiDir)

	// 如果直接路径不存在，尝试查找
	if _, err := os.Stat(wikiDir); os.IsNotExist(err) {
		log.Printf("LoadWiki: 直接路径不存在，尝试查找: %s", wikiID)
		foundPath := ms.findWikiPath(wikiID)
		if foundPath == "" {
			log.Printf("LoadWiki: 未找到wiki路径: %s", wikiID)
			return nil, fmt.Errorf("wiki %s 不存在", wikiID)
		}
		log.Printf("LoadWiki: 找到路径: %s", foundPath)
		wikiDir = filepath.Join(ms.baseDir, foundPath)
	} else {
		log.Printf("LoadWiki: 直接路径存在: %s", wikiDir)
	}

	// 加载元数据
	metadata, err := ms.loadMetadata(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("加载元数据失败: %w", err)
	}

	// 加载所有页面
	pages, err := ms.loadAllPages(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("加载页面失败: %w", err)
	}

	// 自动修复空的title和description
	title := metadata.Title
	if title == "" && metadata.PackagePath != "" {
		title = generateWikiTitleFromPackagePath(metadata.PackagePath)
	}

	description := metadata.Description
	if description == "" && metadata.PackagePath != "" {
		description = generateWikiDescriptionFromPackagePath(metadata.PackagePath, metadata.Metadata.RepositoryURL)
	}

	// 构建Wiki对象
	wiki := &models.Wiki{
		ID:          metadata.ID,
		Title:       title,
		Description: description,
		PackagePath: metadata.PackagePath, // 设置包路径
		Status:      metadata.Status,
		Progress:    metadata.Progress,
		CreatedAt:   metadata.CreatedAt,
		UpdatedAt:   metadata.UpdatedAt,
		Settings:    metadata.Settings,
		Metadata:    metadata.Metadata,
		Tags:        metadata.Tags,
		Pages:       pages,
	}

	return wiki, nil
}

// loadMetadata 加载Wiki元数据
func (ms *MarkdownStorage) loadMetadata(wikiDir string) (*WikiMetadata, error) {
	metaFile := filepath.Join(wikiDir, "meta.json")

	data, err := os.ReadFile(metaFile)
	if err != nil {
		return nil, fmt.Errorf("读取元数据文件失败: %w", err)
	}

	var metadata WikiMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("解析元数据失败: %w", err)
	}

	return &metadata, nil
}

// loadAllPages 加载所有页面
func (ms *MarkdownStorage) loadAllPages(wikiDir string) ([]models.WikiPage, error) {
	var pages []models.WikiPage

	// 遍历所有语言目录
	entries, err := os.ReadDir(wikiDir)
	if err != nil {
		return nil, fmt.Errorf("读取Wiki目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "assets" {
			continue // 跳过文件和assets目录
		}

		language := entry.Name()
		langDir := filepath.Join(wikiDir, language)

		// 加载该语言的所有页面
		langPages, err := ms.loadLanguagePages(langDir, language)
		if err != nil {
			log.Printf("加载语言 %s 的页面失败: %v", language, err)
			continue
		}

		pages = append(pages, langPages...)
	}

	return pages, nil
}

// loadLanguagePages 加载指定语言的所有页面
func (ms *MarkdownStorage) loadLanguagePages(langDir, language string) ([]models.WikiPage, error) {
	var pages []models.WikiPage

	err := filepath.WalkDir(langDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理.md文件
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		page, err := ms.loadPage(path, language)
		if err != nil {
			log.Printf("加载页面失败: %s, 错误: %v", path, err)
			return nil // 继续处理其他页面
		}

		pages = append(pages, *page)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历语言目录失败: %w", err)
	}

	return pages, nil
}

// loadPage 加载单个页面
func (ms *MarkdownStorage) loadPage(filePath, language string) (*models.WikiPage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取页面文件失败: %w", err)
	}

	// 解析Markdown内容
	page, err := ms.parseMarkdownContent(string(content), filePath, language)
	if err != nil {
		return nil, fmt.Errorf("解析Markdown内容失败: %w", err)
	}

	return page, nil
}

// parseMarkdownContent 解析Markdown内容
func (ms *MarkdownStorage) parseMarkdownContent(content, filePath, language string) (*models.WikiPage, error) {
	// 分离前置元数据和内容
	frontMatter, markdownContent := ms.splitFrontMatter(content)

	// 解析前置元数据
	metadata := ms.parseFrontMatter(frontMatter)

	// 生成页面ID
	filename := filepath.Base(filePath)
	pageID := ms.generatePageID(filename, language, metadata)

	// 创建页面对象
	page := &models.WikiPage{
		ID:          pageID,
		Title:       metadata["title"],
		Content:     markdownContent,
		Type:        ms.parsePageType(metadata["type"]),
		Order:       ms.parseInt(metadata["order"], 999),
		WordCount:   len(markdownContent),
		ReadingTime: ms.parseInt(metadata["reading_time"], ms.calculateReadingTime(markdownContent)),
		CreatedAt:   ms.parseTime(metadata["created_at"]),
		UpdatedAt:   ms.parseTime(metadata["updated_at"]),
	}

	// 如果没有标题，从文件名生成
	if page.Title == "" {
		page.Title = ms.generateTitleFromFilename(filename)
	}

	return page, nil
}

// splitFrontMatter 分离前置元数据和内容
func (ms *MarkdownStorage) splitFrontMatter(content string) (map[string]string, string) {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		// 没有前置元数据
		return make(map[string]string), content
	}

	// 查找结束的 ---
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		// 没有找到结束标记
		return make(map[string]string), content
	}

	// 解析前置元数据
	frontMatter := make(map[string]string)
	for i := 1; i < endIndex; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			frontMatter[key] = value
		}
	}

	// 返回内容部分
	contentLines := lines[endIndex+1:]
	markdownContent := strings.Join(contentLines, "\n")
	markdownContent = strings.TrimSpace(markdownContent)

	return frontMatter, markdownContent
}

// parseFrontMatter 解析前置元数据
func (ms *MarkdownStorage) parseFrontMatter(frontMatter map[string]string) map[string]string {
	// 清理值中的引号
	for key, value := range frontMatter {
		value = strings.Trim(value, "\"'")
		frontMatter[key] = value
	}
	return frontMatter
}

// generatePageID 生成页面ID
func (ms *MarkdownStorage) generatePageID(filename, language string, metadata map[string]string) string {
	// 移除.md扩展名
	name := strings.TrimSuffix(filename, ".md")

	// 将文件名转换为适合ID的格式
	pageID := strings.ToLower(name)
	pageID = strings.ReplaceAll(pageID, " ", "-")
	pageID = strings.ReplaceAll(pageID, "_", "-")

	// README特殊处理
	if name == "README" {
		pageID = "overview"
	}

	return fmt.Sprintf("%s_%s", pageID, language)
}

// parsePageType 解析页面类型
func (ms *MarkdownStorage) parsePageType(typeStr string) models.PageType {
	switch typeStr {
	case "overview":
		return models.PageTypeOverview
	case "guide":
		return models.PageTypeGuide
	case "api":
		return models.PageTypeAPI
	case "architecture":
		return models.PageTypeArchitecture
	default:
		return models.PageTypeGuide
	}
}

// parseInt 解析整数
func (ms *MarkdownStorage) parseInt(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(str, "%d", &value); err != nil {
		return defaultValue
	}

	return value
}

// parseTime 解析时间
func (ms *MarkdownStorage) parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now()
	}

	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t
	}

	return time.Now()
}

// generateTitleFromFilename 从文件名生成标题
func (ms *MarkdownStorage) generateTitleFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, ".md")

	// README特殊处理
	if name == "README" {
		return "项目说明"
	}

	// 将连字符和下划线替换为空格
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// 首字母大写（简单实现，避免使用已废弃的strings.Title）
	if len(title) > 0 {
		words := strings.Fields(title)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		title = strings.Join(words, " ")
	}

	return title
}

// calculateReadingTime 计算阅读时间（分钟）
func (ms *MarkdownStorage) calculateReadingTime(content string) int {
	wordCount := len(strings.Fields(content))
	readingSpeed := 200 // 每分钟200字
	readingTime := wordCount / readingSpeed
	if readingTime < 1 {
		readingTime = 1
	}
	return readingTime
}

// ListWikis 列出所有Wiki
func (ms *MarkdownStorage) ListWikis() ([]models.Wiki, error) {
	var wikis []models.Wiki

	// 递归查找所有meta.json文件
	err := filepath.WalkDir(ms.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 只处理meta.json文件
		if !d.IsDir() && d.Name() == "meta.json" {
			// 读取并解析meta.json
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("读取meta.json失败: %s, 错误: %v", path, err)
				return nil
			}

			var metadata WikiMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				log.Printf("解析meta.json失败: %s, 错误: %v", path, err)
				return nil
			}

			// 加载完整的Wiki
			wiki, err := ms.LoadWiki(metadata.ID)
			if err != nil {
				log.Printf("加载Wiki %s 失败: %v", metadata.ID, err)
				return nil
			}

			wikis = append(wikis, *wiki)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历存储目录失败: %w", err)
	}

	return wikis, nil
}

// DeleteWiki 删除Wiki
func (ms *MarkdownStorage) DeleteWiki(wikiID string) error {
	// 查找Wiki的实际路径
	wikiPath := ms.findWikiPath(wikiID)
	if wikiPath == "" {
		return fmt.Errorf("wiki %s 不存在", wikiID)
	}

	wikiDir := filepath.Join(ms.baseDir, wikiPath)

	if err := os.RemoveAll(wikiDir); err != nil {
		return fmt.Errorf("删除Wiki目录失败: %w", err)
	}

	log.Printf("Wiki %s 删除成功", wikiID)
	return nil
}

// UpdateWiki 更新Wiki
func (ms *MarkdownStorage) UpdateWiki(wiki *models.Wiki) error {
	// 更新就是重新保存
	return ms.SaveWiki(wiki)
}

// LoadAllWikis 加载所有Wiki（实现Storage接口）
func (ms *MarkdownStorage) LoadAllWikis() (map[string]*models.Wiki, error) {
	wikis, err := ms.ListWikis()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.Wiki)
	for _, wiki := range wikis {
		// 创建副本以获取指针
		wikiCopy := wiki
		result[wiki.ID] = &wikiCopy
	}

	return result, nil
}

// SaveLogs 保存Wiki日志
func (ms *MarkdownStorage) SaveLogs(wikiID string, logs []string) error {
	// 查找Wiki的实际路径
	wikiPath := ms.findWikiPath(wikiID)
	if wikiPath == "" {
		// 如果找不到，使用wikiID作为路径（可能是新创建的）
		wikiPath = wikiID
	}

	wikiDir := filepath.Join(ms.baseDir, wikiPath)
	logsFile := filepath.Join(wikiDir, "generation.log")

	// 确保目录存在
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return fmt.Errorf("创建Wiki目录失败: %w", err)
	}

	// 将日志转换为文本格式
	var content strings.Builder
	for _, logEntry := range logs {
		content.WriteString(logEntry)
		content.WriteString("\n")
	}

	if err := os.WriteFile(logsFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("写入日志文件失败: %w", err)
	}

	return nil
}

// LoadLogs 加载Wiki日志
func (ms *MarkdownStorage) LoadLogs(wikiID string) ([]string, error) {
	// 查找Wiki的实际路径
	wikiPath := ms.findWikiPath(wikiID)
	if wikiPath == "" {
		wikiPath = wikiID // 回退到直接使用wikiID
	}

	logsFile := filepath.Join(ms.baseDir, wikiPath, "generation.log")

	// 检查文件是否存在
	if _, err := os.Stat(logsFile); os.IsNotExist(err) {
		return []string{}, nil // 返回空日志
	}

	data, err := os.ReadFile(logsFile)
	if err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %w", err)
	}

	// 按行分割日志内容
	content := strings.TrimSpace(string(data))
	if content == "" {
		return []string{}, nil
	}

	logs := strings.Split(content, "\n")

	// 过滤空行
	var filteredLogs []string
	for _, log := range logs {
		if strings.TrimSpace(log) != "" {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs, nil
}

// AppendLog 追加单条日志到文件
func (ms *MarkdownStorage) AppendLog(wikiID, logEntry string) error {
	// 查找Wiki的实际路径
	wikiPath := ms.findWikiPath(wikiID)
	if wikiPath == "" {
		wikiPath = wikiID // 回退到直接使用wikiID
	}

	wikiDir := filepath.Join(ms.baseDir, wikiPath)
	logsFile := filepath.Join(wikiDir, "generation.log")

	// 确保目录存在
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		return fmt.Errorf("创建Wiki目录失败: %w", err)
	}

	// 追加日志到文件
	file, err := os.OpenFile(logsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// 添加时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, logEntry)

	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("写入日志失败: %w", err)
	}

	return nil
}

// generateWikiTitleFromPackagePath 从包路径生成wiki标题
func generateWikiTitleFromPackagePath(packagePath string) string {
	// 从包路径提取项目名称
	parts := strings.Split(packagePath, "/")
	if len(parts) > 0 {
		projectName := parts[len(parts)-1]
		// 将项目名称转换为标题格式
		return strings.Title(strings.ReplaceAll(projectName, "-", " "))
	}
	return "Unknown Project"
}

// generateWikiDescriptionFromPackagePath 从包路径生成wiki描述
func generateWikiDescriptionFromPackagePath(packagePath, repositoryURL string) string {
	parts := strings.Split(packagePath, "/")
	if len(parts) >= 2 {
		owner := parts[len(parts)-2]
		project := parts[len(parts)-1]
		if repositoryURL != "" {
			return fmt.Sprintf("Documentation for %s/%s - A Go package from %s", owner, project, repositoryURL)
		}
		return fmt.Sprintf("Documentation for %s/%s", owner, project)
	}
	return fmt.Sprintf("Documentation for %s", packagePath)
}
