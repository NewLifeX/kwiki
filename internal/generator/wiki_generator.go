package generator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stcn52/kwiki/internal/ai"
	"github.com/stcn52/kwiki/internal/config"
	"github.com/stcn52/kwiki/pkg/models"
)

// RepositoryInfo 仓库信息
type RepositoryInfo struct {
	Name        string
	URL         string
	Language    string
	Description string
	Topics      []string
	Framework   string
}

// RepositoryDocumentationData 仓库文档数据
type RepositoryDocumentationData struct {
	Repository *RepositoryInfo
	PageType   models.PageType
	Title      string
	Language   string
}

// WikiGenerator 负责生成wiki文档
type WikiGenerator struct {
	config          *config.Config
	aiManager       *ai.ProviderManager
	templateManager *TemplateManager
	progressChan    chan models.GenerationProgress
}

// New 创建新的WikiGenerator实例
func New(cfg *config.Config, aiManager *ai.ProviderManager) *WikiGenerator {
	// 加载生成器配置
	generatorConfig, err := LoadGeneratorConfig("config/generator.yaml")
	if err != nil {
		log.Printf("加载生成器配置失败，使用默认配置: %v", err)
		generatorConfig = nil // NewTemplateManager会使用默认配置
	}

	return &WikiGenerator{
		config:          cfg,
		aiManager:       aiManager,
		templateManager: NewTemplateManager(generatorConfig),
		progressChan:    make(chan models.GenerationProgress, 100),
	}
}

// GenerateWiki 生成wiki文档
func (wg *WikiGenerator) GenerateWiki(ctx context.Context, req models.GenerationRequest) (*models.Wiki, error) {
	log.Printf("开始生成wiki文档，仓库: %s", req.RepositoryURL)

	// 生成基于包路径的目录结构
	packagePath := generatePackagePath(req.RepositoryURL)

	// 自动生成title和description（如果未提供）
	title := req.Title
	if title == "" {
		title = generateWikiTitle(req.RepositoryURL, packagePath)
	}

	description := req.Description
	if description == "" {
		description = generateWikiDescription(req.RepositoryURL, packagePath)
	}

	// 创建wiki实例
	wiki := &models.Wiki{
		ID:           packagePath, // 使用包路径作为ID
		RepositoryID: req.RepositoryURL,
		PackagePath:  packagePath,
		Title:        title,
		Description:  description,
		Status:       models.WikiStatusGenerating,
		Progress:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GeneratedBy:  req.Settings.AIProvider,
		Model:        req.Settings.Model,
		Language:     req.PrimaryLanguage,
		Languages:    req.Languages,
		Pages:        []models.WikiPage{},
		Diagrams:     []models.WikiDiagram{},
		Settings:     req.Settings,
		Metadata: models.WikiMetadata{
			Languages:     req.Languages,
			PackagePath:   packagePath,
			RepositoryURL: req.RepositoryURL,
		},
	}

	// 创建独立的上下文用于异步生成，不依赖于HTTP请求的上下文
	backgroundCtx := context.Background()

	// 启动异步生成过程
	go wg.generateWikiAsync(backgroundCtx, wiki, req)

	return wiki, nil
}

// generateWikiAsync 异步生成wiki内容
func (wg *WikiGenerator) generateWikiAsync(ctx context.Context, wiki *models.Wiki, req models.GenerationRequest) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Wiki生成过程中发生panic: %v", r)
			wiki.Status = models.WikiStatusFailed
		}
	}()

	startTime := time.Now()

	// 如果是模板文档生成请求，使用特殊处理
	if req.RepositoryURL == "template-docs" {
		wg.generateTemplateDocumentation(ctx, wiki, req)
	} else {
		// 其他类型的文档生成
		wg.generateRepositoryDocumentation(ctx, wiki, req)
	}

	wiki.Metadata.GenerationTime = time.Since(startTime)
}

// generateTemplateDocumentation 生成模板系统文档
func (wg *WikiGenerator) generateTemplateDocumentation(ctx context.Context, wiki *models.Wiki, req models.GenerationRequest) {
	log.Printf("开始生成模板系统文档，目标语言: %v", req.Languages)

	// 发送初始进度
	wg.sendProgress(wiki.ID, models.WikiStatusGenerating, 10, "扫描模板目录", "正在扫描模板目录...", nil)

	// 扫描模板目录
	log.Printf("扫描模板目录: %s", wg.templateManager.config.TemplateDir)
	templateData, err := wg.templateManager.ScanTemplateDirectory()
	if err != nil {
		log.Printf("扫描模板目录失败: %v", err)
		wiki.Status = models.WikiStatusFailed
		wg.sendProgress(wiki.ID, models.WikiStatusFailed, 0, "扫描失败", "扫描模板目录失败", err)
		return
	}

	log.Printf("模板目录扫描完成，统计信息: 总模板数=%d, 支持语言数=%d",
		templateData.Statistics.TotalTemplates, templateData.Statistics.TotalLanguages)

	for langCode, count := range templateData.Statistics.TemplatesByLanguage {
		log.Printf("  语言 %s: %d 个模板", langCode, count)
	}

	for templateType, count := range templateData.Statistics.TemplatesByType {
		log.Printf("  类型 %s: %d 个模板", templateType, count)
	}

	wiki.Progress = 20
	wiki.UpdatedAt = time.Now()
	wg.sendProgress(wiki.ID, models.WikiStatusGenerating, 20, "开始生成", "模板扫描完成，开始生成页面", nil)

	// 为每种语言生成文档页面
	totalLanguages := len(req.Languages)
	for i, language := range req.Languages {
		if language == "" {
			language = "zh" // 默认中文
		}

		log.Printf("开始处理语言 %d/%d: %s", i+1, totalLanguages, language)
		wg.sendProgress(wiki.ID, models.WikiStatusGenerating, 20+(60*i/totalLanguages), "生成页面", fmt.Sprintf("正在生成%s语言的页面", language), nil)

		err := wg.generatePagesForLanguage(ctx, wiki, templateData, language, i, totalLanguages, req.Settings)
		if err != nil {
			log.Printf("生成%s语言模板文档失败: %v", language, err)
			wg.sendProgress(wiki.ID, models.WikiStatusGenerating, 20+(60*i/totalLanguages), "生成失败", fmt.Sprintf("生成%s语言失败", language), err)
			continue
		}

		log.Printf("语言 %s 处理完成", language)
	}

	// 完成生成
	wiki.Status = models.WikiStatusCompleted
	wiki.Progress = 100
	wiki.UpdatedAt = time.Now()
	wiki.Metadata.PagesGenerated = len(wiki.Pages)
	wg.sendProgress(wiki.ID, models.WikiStatusCompleted, 100, "完成", fmt.Sprintf("生成完成，共%d个页面", len(wiki.Pages)), nil)

	log.Printf("模板系统文档生成完成！")
	log.Printf("  总页面数: %d", len(wiki.Pages))
	log.Printf("  处理语言: %v", req.Languages)
	log.Printf("  生成状态: %s", wiki.Status)
}

// generateRepositoryDocumentation 生成仓库文档
func (wg *WikiGenerator) generateRepositoryDocumentation(ctx context.Context, wiki *models.Wiki, req models.GenerationRequest) {
	log.Printf("开始生成仓库文档: %s", req.RepositoryURL)

	// 发送初始进度
	wg.sendProgress(wiki.ID, models.WikiStatusAnalyzing, 10, "分析仓库", "正在分析仓库结构...", nil)

	// 分析仓库信息
	repoInfo, err := wg.analyzeRepository(req.RepositoryURL)
	if err != nil {
		log.Printf("分析仓库失败: %v", err)
		wiki.Status = models.WikiStatusFailed
		wg.sendProgress(wiki.ID, models.WikiStatusFailed, 0, "分析失败", "仓库分析失败", err)
		return
	}

	log.Printf("仓库分析完成: %s (%s)", repoInfo.Name, repoInfo.Language)
	wg.sendProgress(wiki.ID, models.WikiStatusGenerating, 30, "生成文档", "开始生成文档页面...", nil)

	// 为每种语言生成文档页面
	totalLanguages := len(req.Languages)
	for i, language := range req.Languages {
		if language == "" {
			language = "en" // 默认英文
		}

		log.Printf("开始处理语言 %d/%d: %s", i+1, totalLanguages, language)
		progress := 30 + (60 * i / totalLanguages)
		wg.sendProgress(wiki.ID, models.WikiStatusGenerating, progress, "生成页面", fmt.Sprintf("正在生成%s语言的页面", language), nil)

		err := wg.generateRepositoryPagesForLanguage(ctx, wiki, repoInfo, language, req.Settings)
		if err != nil {
			log.Printf("生成%s语言仓库文档失败: %v", language, err)
			wg.sendProgress(wiki.ID, models.WikiStatusGenerating, progress, "生成失败", fmt.Sprintf("生成%s语言失败", language), err)
			// 继续处理其他语言，不要因为一个语言失败就停止
			continue
		}

		log.Printf("语言 %s 处理完成", language)
	}

	// 检查是否有页面生成成功
	if len(wiki.Pages) == 0 {
		log.Printf("警告: 没有成功生成任何页面")
		wiki.Status = models.WikiStatusFailed
		wiki.Progress = 0
		wg.sendProgress(wiki.ID, models.WikiStatusFailed, 0, "失败", "没有成功生成任何页面", fmt.Errorf("所有页面生成都失败了"))
	} else {
		// 完成生成
		wiki.Status = models.WikiStatusCompleted
		wiki.Progress = 100
		wg.sendProgress(wiki.ID, models.WikiStatusCompleted, 100, "完成", fmt.Sprintf("生成完成，共%d个页面", len(wiki.Pages)), nil)
	}

	wiki.UpdatedAt = time.Now()
	wiki.Metadata.PagesGenerated = len(wiki.Pages)

	log.Printf("仓库文档生成完成: %s，生成页面数: %d", req.RepositoryURL, len(wiki.Pages))
}

// generatePagesForLanguage 为指定语言生成页面
func (wg *WikiGenerator) generatePagesForLanguage(ctx context.Context, wiki *models.Wiki, templateData *TemplateDocumentationData, language string, languageIndex, totalLanguages int, settings models.WikiSettings) error {
	log.Printf("开始为语言 %s 生成页面", language)

	// 获取该语言下所有可用的模板
	log.Printf("获取语言 %s 的模板列表", language)
	templates, err := wg.templateManager.GetTemplatesWithMetadata(language)
	if err != nil {
		log.Printf("获取语言 %s 的模板列表失败: %v", language, err)
		return fmt.Errorf("获取模板列表失败: %w", err)
	}

	log.Printf("找到 %d 个模板用于语言 %s", len(templates), language)

	// 为每个模板生成对应的页面
	successCount := 0
	var allStats []*PageGenerationStats

	for i, tmpl := range templates {
		log.Printf("处理模板 %d/%d: %s (类型: %s)", i+1, len(templates), tmpl.Metadata.Title, tmpl.Metadata.Type)

		page, stats, err := wg.generatePageFromTemplate(ctx, tmpl, templateData, language, settings)
		if err != nil {
			log.Printf("使用模板 %s 生成页面失败: %v", tmpl.Metadata.Title, err)
			continue
		}

		wiki.Pages = append(wiki.Pages, *page)
		allStats = append(allStats, stats)
		successCount++

		// 更新页面级别的进度
		// 进度分配：20% 初始化，60% 生成页面，20% 完成
		// 每种语言占用60%的一部分，每个页面占用该语言部分的一部分
		languageProgressShare := 60.0 / float64(totalLanguages)
		pageProgressInLanguage := languageProgressShare / float64(len(templates))
		currentProgress := 20.0 + float64(languageIndex)*languageProgressShare + float64(i+1)*pageProgressInLanguage

		wiki.Progress = int(currentProgress)
		wiki.UpdatedAt = time.Now()

		log.Printf("成功生成页面: %s (%d/%d), 当前进度: %d%%", page.Title, successCount, len(templates), wiki.Progress)
	}

	// 记录语言级别的统计信息
	if len(allStats) > 0 {
		wg.logLanguageStats(language, allStats)
	}

	log.Printf("语言 %s 页面生成完成，成功: %d/%d", language, successCount, len(templates))
	return nil
}

// PageGenerationStats 页面生成统计
type PageGenerationStats struct {
	PageTitle       string
	PromptLength    int
	ContentLength   int
	TokensUsed      int
	GenerationTime  time.Duration
	GenerationSpeed float64 // 字符/秒
	Model           string
	FinishReason    string
}

// AIGenerationStats AI生成统计
type AIGenerationStats struct {
	TokensUsed      int
	GenerationTime  time.Duration
	GenerationSpeed float64 // 字符/秒
	Model           string
	FinishReason    string
}

// generatePageFromTemplate 使用模板生成页面
func (wg *WikiGenerator) generatePageFromTemplate(ctx context.Context, tmpl *TemplateInfo, data *TemplateDocumentationData, language string, settings models.WikiSettings) (*models.WikiPage, *PageGenerationStats, error) {
	log.Printf("开始生成页面: %s (类型: %s, 语言: %s)", tmpl.Metadata.Title, tmpl.Metadata.Type, language)

	// 准备模板数据
	templateData := TemplateData{
		ProjectName:     data.ProjectName,
		Description:     data.Description,
		PrimaryLanguage: "Go",
		License:         "MIT",
		Language:        language,
		Modules:         []ModuleData{}, // 模板文档不需要代码模块信息
	}

	log.Printf("模板数据准备完成: 项目名=%s, 描述=%s, 语言=%s",
		templateData.ProjectName, templateData.Description, templateData.Language)

	// 渲染模板生成AI提示词
	log.Printf("开始渲染模板: %s", tmpl.Metadata.Title)
	var promptBuilder strings.Builder
	err := tmpl.Template.Execute(&promptBuilder, templateData)
	if err != nil {
		log.Printf("渲染模板失败: %s, 错误: %v", tmpl.Metadata.Title, err)
		return nil, nil, fmt.Errorf("渲染模板失败: %w", err)
	}

	prompt := promptBuilder.String()
	log.Printf("模板渲染完成: %s, 提示词长度: %d", tmpl.Metadata.Title, len(prompt))

	// 使用AI生成内容并记录统计
	log.Printf("开始AI生成内容: %s", tmpl.Metadata.Title)
	content, stats, err := wg.generateContentWithAIStats(ctx, prompt, settings)
	if err != nil {
		log.Printf("AI生成内容失败: %s, 错误: %v", tmpl.Metadata.Title, err)
		return nil, nil, fmt.Errorf("AI生成内容失败: %w", err)
	}

	// 创建页面
	page := &models.WikiPage{
		ID:          fmt.Sprintf("%s_%s", tmpl.Metadata.Type, language),
		Title:       tmpl.Metadata.Title,
		Content:     content,
		Type:        wg.getPageType(tmpl.Metadata.Type),
		Order:       tmpl.Metadata.Order,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		WordCount:   len(content),
		ReadingTime: wg.calculateReadingTime(content),
	}

	// 完善统计信息
	pageStats := &PageGenerationStats{
		PageTitle:       tmpl.Metadata.Title,
		PromptLength:    len(prompt),
		ContentLength:   len(content),
		TokensUsed:      stats.TokensUsed,
		GenerationTime:  stats.GenerationTime,
		GenerationSpeed: stats.GenerationSpeed,
		Model:           stats.Model,
		FinishReason:    stats.FinishReason,
	}

	log.Printf("页面生成完成: %s", tmpl.Metadata.Title)
	log.Printf("  - 字数: %d, 阅读时间: %d分钟", page.WordCount, page.ReadingTime)
	log.Printf("  - Token消耗: %d, 生成速度: %.1f字符/秒", pageStats.TokensUsed, pageStats.GenerationSpeed)
	log.Printf("  - 模型: %s, 完成原因: %s", pageStats.Model, pageStats.FinishReason)

	return page, pageStats, nil
}

// generateContentWithAIStats 使用AI生成内容并返回统计信息
func (wg *WikiGenerator) generateContentWithAIStats(ctx context.Context, prompt string, settings models.WikiSettings) (string, *AIGenerationStats, error) {
	log.Printf("请求使用AI提供商: %s, 模型: %s", settings.AIProvider, settings.Model)

	// 获取指定的AI提供商
	provider, exists := wg.aiManager.GetProvider(settings.AIProvider)
	if !exists {
		log.Printf("错误: AI提供商 '%s' 不存在", settings.AIProvider)
		return "", nil, fmt.Errorf("AI提供商 '%s' 不存在", settings.AIProvider)
	}

	if provider == nil {
		log.Printf("错误: AI提供商 '%s' 为空", settings.AIProvider)
		return "", nil, fmt.Errorf("AI提供商 '%s' 为空", settings.AIProvider)
	}

	if !provider.IsAvailable() {
		log.Printf("错误: AI提供商 '%s' 不可用", settings.AIProvider)
		return "", nil, fmt.Errorf("AI提供商 '%s' 不可用", settings.AIProvider)
	}

	log.Printf("成功获取AI提供商: %s", provider.GetName())
	log.Printf("开始AI流式生成内容，提供商: %s, 模型: %s", provider.GetName(), settings.Model)
	log.Printf("提示词长度: %d 字符", len(prompt))

	startTime := time.Now()

	// 使用用户指定的模型
	modelName := settings.Model
	if modelName == "" {
		modelName = "deepseek-chat" // 默认模型
	}

	log.Printf("使用指定模型: %s", modelName)

	// 使用流式生成，不设置额外的超时（让AI提供商自己处理超时）
	streamChan, streamErr := provider.GenerateStream(ctx, prompt, ai.GenerationOptions{
		Model:       modelName,
		Temperature: float32(settings.Temperature),
		MaxTokens:   settings.MaxTokens,
		TopP:        0.9,
	})

	if streamErr != nil {
		log.Printf("流式生成启动失败: %v", streamErr)
		return "", nil, fmt.Errorf("流式生成启动失败: %w", streamErr)
	}

	log.Printf("流式生成启动成功，开始接收数据...")

	var fullText string
	var tokensUsed int
	var finishReason string
	var err error

	var textBuilder strings.Builder
	var lastLogTime time.Time
	var chunkCount int

	// 设置接收超时
	timeout := 5 * time.Minute
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case streamResp, ok := <-streamChan:
			if !ok {
				// 通道关闭，生成完成
				fullText = textBuilder.String()
				if fullText == "" && err == nil {
					err = fmt.Errorf("流式生成完成但没有收到任何内容")
				}
				goto done
			}

			// 重置超时计时器
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}
			timeoutTimer.Reset(timeout)

			if streamResp.Error != nil {
				log.Printf("流式生成错误: %v", streamResp.Error)
				err = streamResp.Error
				goto done
			}

			if streamResp.Text != "" {
				textBuilder.WriteString(streamResp.Text)
				chunkCount++

				// 每3秒最多记录一次进度，减少日志噪音
				if time.Since(lastLogTime) > 3*time.Second {
					log.Printf("流式生成进度: 已接收 %d 字符 (%d 片段)", textBuilder.Len(), chunkCount)
					lastLogTime = time.Now()
				}
			}

			if streamResp.Done {
				fullText = textBuilder.String()
				tokensUsed = streamResp.TokensUsed
				err = nil

				// 从元数据中获取完成原因
				if streamResp.Metadata != nil {
					if reason, ok := streamResp.Metadata["finish_reason"]; ok {
						finishReason = reason
					}
				}

				goto done
			}

		case <-timeoutTimer.C:
			log.Printf("流式生成超时 (%v)", timeout)
			err = fmt.Errorf("流式生成超时")
			goto done

		case <-ctx.Done():
			log.Printf("上下文取消: %v", ctx.Err())
			err = ctx.Err()
			goto done
		}
	}

done:

	duration := time.Since(startTime)

	if err != nil {
		log.Printf("AI流式生成失败，耗时: %v, 错误: %v", duration, err)
		return "", nil, fmt.Errorf("AI生成内容失败: %w", err)
	}

	// 计算生成速度
	generationSpeed := float64(len(fullText)) / duration.Seconds()

	// 创建统计信息
	stats := &AIGenerationStats{
		TokensUsed:      tokensUsed,
		GenerationTime:  duration,
		GenerationSpeed: generationSpeed,
		Model:           modelName,
		FinishReason:    finishReason,
	}

	log.Printf("AI流式生成成功统计:")
	log.Printf("  - 耗时: %v", duration)
	log.Printf("  - 生成内容长度: %d 字符", len(fullText))
	log.Printf("  - Token消耗: %d", tokensUsed)
	log.Printf("  - 生成速度: %.1f 字符/秒", generationSpeed)
	log.Printf("  - 模型: %s", modelName)
	log.Printf("  - 完成原因: %s", finishReason)

	return fullText, stats, nil
}

// logLanguageStats 记录语言级别的统计信息
func (wg *WikiGenerator) logLanguageStats(language string, stats []*PageGenerationStats) {
	if len(stats) == 0 {
		return
	}

	var totalTokens int
	var totalTime time.Duration
	var totalChars int
	var totalPromptChars int
	modelCount := make(map[string]int)
	finishReasonCount := make(map[string]int)

	for _, stat := range stats {
		totalTokens += stat.TokensUsed
		totalTime += stat.GenerationTime
		totalChars += stat.ContentLength
		totalPromptChars += stat.PromptLength
		modelCount[stat.Model]++
		finishReasonCount[stat.FinishReason]++
	}

	avgTokensPerPage := float64(totalTokens) / float64(len(stats))
	avgTimePerPage := totalTime / time.Duration(len(stats))
	avgCharsPerPage := float64(totalChars) / float64(len(stats))
	avgPromptCharsPerPage := float64(totalPromptChars) / float64(len(stats))
	overallSpeed := float64(totalChars) / totalTime.Seconds()

	log.Printf("=== 语言 %s 生成统计 ===", language)
	log.Printf("页面数量: %d", len(stats))
	log.Printf("Token消耗:")
	log.Printf("  - 总计: %d tokens", totalTokens)
	log.Printf("  - 平均每页: %.1f tokens", avgTokensPerPage)
	log.Printf("生成时间:")
	log.Printf("  - 总计: %v", totalTime)
	log.Printf("  - 平均每页: %v", avgTimePerPage)
	log.Printf("内容统计:")
	log.Printf("  - 总字符数: %d", totalChars)
	log.Printf("  - 平均每页: %.1f 字符", avgCharsPerPage)
	log.Printf("  - 提示词总字符: %d", totalPromptChars)
	log.Printf("  - 平均提示词: %.1f 字符", avgPromptCharsPerPage)
	log.Printf("生成速度:")
	log.Printf("  - 整体速度: %.1f 字符/秒", overallSpeed)
	log.Printf("模型使用:")
	for model, count := range modelCount {
		log.Printf("  - %s: %d 次", model, count)
	}
	log.Printf("完成原因:")
	for reason, count := range finishReasonCount {
		log.Printf("  - %s: %d 次", reason, count)
	}
	log.Printf("=== 统计结束 ===")
}

// GetProgressChannel 返回进度通道
func (wg *WikiGenerator) GetProgressChannel() <-chan models.GenerationProgress {
	return wg.progressChan
}

// sendProgress 发送进度更新
func (wg *WikiGenerator) sendProgress(wikiID string, status models.WikiStatus, progress int, step string, message string, err error) {
	progressUpdate := models.GenerationProgress{
		WikiID:      wikiID,
		Status:      status,
		Progress:    progress,
		CurrentStep: step,
		Message:     message,
		UpdatedAt:   time.Now(),
	}

	if err != nil {
		progressUpdate.Error = err.Error()
	}

	// 非阻塞发送
	select {
	case wg.progressChan <- progressUpdate:
	default:
		log.Printf("Warning: Progress channel is full, dropping update")
	}
}

// getPageType 根据模板类型获取页面类型
func (wg *WikiGenerator) getPageType(templateType string) models.PageType {
	switch templateType {
	case "overview":
		return models.PageTypeOverview
	case "guide":
		return models.PageTypeGuide
	case "reference":
		return models.PageTypeReference
	case "architecture":
		return models.PageTypeArchitecture
	default:
		return models.PageTypeGuide
	}
}

// calculateReadingTime 计算阅读时间（分钟）
func (wg *WikiGenerator) calculateReadingTime(content string) int {
	wordCount := len(strings.Fields(content))

	readingSpeed := 200
	if wg.templateManager.config != nil && wg.templateManager.config.ReadingSpeed > 0 {
		readingSpeed = wg.templateManager.config.ReadingSpeed
	}

	readingTime := wordCount / readingSpeed
	if readingTime < 1 {
		readingTime = 1
	}
	return readingTime
}

// generateWikiID 生成唯一的wiki ID
func generateWikiID() string {
	return fmt.Sprintf("wiki_%d", time.Now().UnixNano())
}

// generatePackagePath 从仓库URL生成包路径
func generatePackagePath(repositoryURL string) string {
	if repositoryURL == "" {
		return "unknown"
	}

	// 处理特殊的template-docs类型
	if repositoryURL == "template-docs" {
		return "template-docs/example"
	}

	// 解析GitHub/GitLab URL
	if strings.Contains(repositoryURL, "github.com") {
		// https://github.com/gin-gonic/gin -> github.com/gin-gonic/gin
		parts := strings.Split(repositoryURL, "/")
		if len(parts) >= 5 {
			owner := parts[len(parts)-2]
			repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
			return fmt.Sprintf("github.com/%s/%s", owner, repo)
		}
	} else if strings.Contains(repositoryURL, "gitlab.com") {
		// https://gitlab.com/owner/repo -> gitlab.com/owner/repo
		parts := strings.Split(repositoryURL, "/")
		if len(parts) >= 5 {
			owner := parts[len(parts)-2]
			repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
			return fmt.Sprintf("gitlab.com/%s/%s", owner, repo)
		}
	}

	// 如果无法解析，使用URL的hash作为目录名
	return fmt.Sprintf("unknown/%x", repositoryURL)
}

// analyzeRepository 分析仓库信息
func (wg *WikiGenerator) analyzeRepository(repoURL string) (*RepositoryInfo, error) {
	log.Printf("分析仓库: %s", repoURL)

	// 从URL中提取仓库信息
	repoInfo := &RepositoryInfo{
		URL: repoURL,
	}

	// 解析GitHub/GitLab URL
	if strings.Contains(repoURL, "github.com") {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 5 {
			owner := parts[len(parts)-2]
			repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
			repoInfo.Name = fmt.Sprintf("%s/%s", owner, repo)
		}
		repoInfo.Language = "Go" // 假设是Go项目
		repoInfo.Framework = "Web Framework"
		repoInfo.Description = fmt.Sprintf("Documentation for %s", repoInfo.Name)
	} else if strings.Contains(repoURL, "gitlab.com") {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 5 {
			owner := parts[len(parts)-2]
			repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
			repoInfo.Name = fmt.Sprintf("%s/%s", owner, repo)
		}
		repoInfo.Language = "Go"
		repoInfo.Framework = "Application"
		repoInfo.Description = fmt.Sprintf("Documentation for %s", repoInfo.Name)
	} else {
		return nil, fmt.Errorf("不支持的仓库类型: %s", repoURL)
	}

	if repoInfo.Name == "" {
		repoInfo.Name = "Unknown Repository"
	}

	log.Printf("仓库分析完成: %s", repoInfo.Name)
	return repoInfo, nil
}

// generateRepositoryPagesForLanguage 为指定语言生成仓库页面
func (wg *WikiGenerator) generateRepositoryPagesForLanguage(ctx context.Context, wiki *models.Wiki, repoInfo *RepositoryInfo, language string, settings models.WikiSettings) error {
	log.Printf("开始为语言 %s 生成仓库页面", language)

	// 定义要生成的页面类型
	pageTypes := []struct {
		Type  models.PageType
		Title string
		Order int
	}{
		{models.PageTypeOverview, getLocalizedTitle("项目概述", "Project Overview", language), 1},
		{models.PageTypeGuide, getLocalizedTitle("快速开始", "Getting Started", language), 2},
		{models.PageTypeGuide, getLocalizedTitle("安装指南", "Installation Guide", language), 3},
		{models.PageTypeAPI, getLocalizedTitle("API参考", "API Reference", language), 4},
		{models.PageTypeArchitecture, getLocalizedTitle("架构设计", "Architecture", language), 5},
	}

	// 为每种页面类型生成内容
	successCount := 0
	for _, pageType := range pageTypes {
		page, err := wg.generateRepositoryPage(ctx, repoInfo, pageType.Type, pageType.Title, language, pageType.Order, settings)
		if err != nil {
			log.Printf("生成页面失败: %s, 错误: %v", pageType.Title, err)
			continue
		}

		wiki.Pages = append(wiki.Pages, *page)
		successCount++
		log.Printf("页面生成成功: %s (%s)", page.Title, page.ID)
	}

	if successCount == 0 {
		return fmt.Errorf("语言 %s 的所有页面生成都失败了", language)
	}

	log.Printf("语言 %s 成功生成 %d/%d 个页面", language, successCount, len(pageTypes))

	return nil
}

// generateRepositoryPage 生成单个仓库页面
func (wg *WikiGenerator) generateRepositoryPage(ctx context.Context, repoInfo *RepositoryInfo, pageType models.PageType, title, language string, order int, settings models.WikiSettings) (*models.WikiPage, error) {
	log.Printf("生成页面: %s (%s)", title, pageType)

	// 生成页面ID
	pageID := fmt.Sprintf("%s_%s", strings.ToLower(strings.ReplaceAll(title, " ", "-")), language)

	// 根据页面类型生成提示词
	prompt := wg.generateRepositoryPagePrompt(repoInfo, pageType, title, language)

	// 使用AI生成内容，带重试机制
	var content string
	var err error
	maxRetries := 3

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("第 %d 次重试生成页面: %s", retry+1, title)
			time.Sleep(time.Duration(retry) * time.Second) // 递增延迟
		}

		content, _, err = wg.generateContentWithAIStats(ctx, prompt, settings)
		if err == nil {
			break
		}

		log.Printf("AI生成内容失败 (尝试 %d/%d): %v", retry+1, maxRetries, err)
	}

	if err != nil {
		return nil, fmt.Errorf("AI生成内容失败 (已重试 %d 次): %w", maxRetries, err)
	}

	// 创建页面对象
	page := &models.WikiPage{
		ID:          pageID,
		Title:       title,
		Content:     content,
		Type:        pageType,
		Order:       order,
		WordCount:   len(content),
		ReadingTime: wg.calculateReadingTime(content),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return page, nil
}

// generateRepositoryPagePrompt 生成仓库页面的提示词
func (wg *WikiGenerator) generateRepositoryPagePrompt(repoInfo *RepositoryInfo, pageType models.PageType, title, language string) string {
	var prompt strings.Builder

	// 根据语言设置基础提示
	if language == "zh" {
		prompt.WriteString("你是一个专业的技术文档编写专家。请为以下项目生成高质量的中文技术文档。\n\n")
	} else {
		prompt.WriteString("You are a professional technical documentation expert. Please generate high-quality English technical documentation for the following project.\n\n")
	}

	// 项目信息
	prompt.WriteString("项目信息 / Project Information:\n")
	prompt.WriteString(fmt.Sprintf("- 名称 / Name: %s\n", repoInfo.Name))
	prompt.WriteString(fmt.Sprintf("- URL: %s\n", repoInfo.URL))
	prompt.WriteString(fmt.Sprintf("- 语言 / Language: %s\n", repoInfo.Language))
	prompt.WriteString(fmt.Sprintf("- 框架 / Framework: %s\n", repoInfo.Framework))
	prompt.WriteString(fmt.Sprintf("- 描述 / Description: %s\n\n", repoInfo.Description))

	// 页面类型提示
	if language == "zh" {
		prompt.WriteString(fmt.Sprintf("请生成一个详细的%s文档，标题为\"%s\"。\n", getPageTypeNameZh(pageType), title))
	} else {
		prompt.WriteString(fmt.Sprintf("Please generate a detailed %s document titled \"%s\".\n", getPageTypeNameEn(pageType), title))
	}

	prompt.WriteString("\n请使用Markdown格式，包含适当的标题、代码块、列表等格式。确保内容专业、准确、易于理解。")
	prompt.WriteString("\nPlease use Markdown format with appropriate headings, code blocks, lists, etc. Ensure the content is professional, accurate, and easy to understand.")

	return prompt.String()
}

// getPageTypeNameZh 获取页面类型的中文名称
func getPageTypeNameZh(pageType models.PageType) string {
	nameMap := map[models.PageType]string{
		models.PageTypeOverview:     "项目概述",
		models.PageTypeGuide:        "使用指南",
		models.PageTypeAPI:          "API参考",
		models.PageTypeArchitecture: "架构设计",
	}

	if name, exists := nameMap[pageType]; exists {
		return name
	}
	return "技术文档"
}

// getPageTypeNameEn 获取页面类型的英文名称
func getPageTypeNameEn(pageType models.PageType) string {
	nameMap := map[models.PageType]string{
		models.PageTypeOverview:     "project overview",
		models.PageTypeGuide:        "user guide",
		models.PageTypeAPI:          "API reference",
		models.PageTypeArchitecture: "architecture design",
	}

	if name, exists := nameMap[pageType]; exists {
		return name
	}
	return "technical documentation"
}

// generateWikiTitle 从仓库URL和包路径生成wiki标题
func generateWikiTitle(repositoryURL, packagePath string) string {
	// 从包路径提取项目名称
	parts := strings.Split(packagePath, "/")
	if len(parts) > 0 {
		projectName := parts[len(parts)-1]
		// 将项目名称转换为标题格式
		return strings.Title(strings.ReplaceAll(projectName, "-", " "))
	}
	return "Unknown Project"
}

// generateWikiDescription 从仓库URL和包路径生成wiki描述
func generateWikiDescription(repositoryURL, packagePath string) string {
	parts := strings.Split(packagePath, "/")
	if len(parts) >= 2 {
		owner := parts[len(parts)-2]
		project := parts[len(parts)-1]
		return fmt.Sprintf("Documentation for %s/%s - A Go package from %s", owner, project, repositoryURL)
	}
	return fmt.Sprintf("Documentation for %s", packagePath)
}

// getLocalizedTitle 获取本地化标题
func getLocalizedTitle(zhTitle, enTitle, language string) string {
	if language == "zh" {
		return zhTitle
	}
	return enTitle
}
