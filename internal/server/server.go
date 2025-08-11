package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/stcn52/kwiki/internal/ai"
	"github.com/stcn52/kwiki/internal/config"
	"github.com/stcn52/kwiki/internal/generator"
	"github.com/stcn52/kwiki/internal/storage"
	"github.com/stcn52/kwiki/pkg/models"
	"github.com/stcn52/kwiki/pkg/utils"
)

// Server represents the web server
type Server struct {
	config        *config.Config
	router        *gin.Engine
	aiManager     *ai.ProviderManager
	wikiGenerator *generator.WikiGenerator
	storage       storage.Storage
	activeWikis   map[string]*models.Wiki
	repoURLToWiki map[string]string   // Maps repository URL to wiki ID
	wikiLogs      map[string][]string // Maps wiki ID to generation logs
	wsUpgrader    websocket.Upgrader
	wsConnections map[string]*websocket.Conn
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Set Gin mode
	if cfg.Server.Host == "localhost" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize AI provider manager
	aiManager := ai.NewProviderManager()

	// Register AI providers
	if openaiKey := cfg.AI.Providers["openai"].APIKey; openaiKey != "" {
		openaiProvider := ai.NewOpenAIProvider(openaiKey, cfg.AI.Providers["openai"].BaseURL)
		aiManager.RegisterProvider("openai", openaiProvider)
	}

	if geminiKey := cfg.AI.Providers["gemini"].APIKey; geminiKey != "" {
		geminiProvider := ai.NewGeminiProvider(geminiKey)
		aiManager.RegisterProvider("gemini", geminiProvider)
	}

	log.Printf("Checking DeepSeek configuration...")
	log.Printf("Available providers in config: %+v", cfg.AI.Providers)
	if deepseekProvider, exists := cfg.AI.Providers["deepseek"]; exists {
		log.Printf("DeepSeek provider found in config: %+v", deepseekProvider)
		if deepseekProvider.APIKey != "" {
			log.Printf("Registering DeepSeek provider with API key: %s", deepseekProvider.APIKey[:4]+"****")
			deepseekAI := ai.NewDeepSeekProvider(deepseekProvider.APIKey)
			aiManager.RegisterProvider("deepseek", deepseekAI)
		} else {
			log.Printf("DeepSeek provider found but API key is empty")
		}
	} else {
		log.Printf("DeepSeek provider not found in config")
	}

	// Always register Ollama provider (it will check availability)
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		// Use base_url from config if OLLAMA_HOST is not set
		if ollamaConfig, exists := cfg.AI.Providers["ollama"]; exists && ollamaConfig.BaseURL != "" {
			ollamaHost = ollamaConfig.BaseURL
		}
	}
	ollamaProvider := ai.NewOllamaProvider(ollamaHost)
	aiManager.RegisterProvider("ollama", ollamaProvider)

	// Set default provider
	aiManager.SetDefaultProvider(cfg.AI.DefaultProvider)

	// Initialize storage
	dataDir := cfg.Server.DataDir
	if dataDir == "" {
		dataDir = "./data" // Default data directory
	}

	// 使用新的Markdown存储
	wikisDir := filepath.Join(dataDir, "wikis")
	markdownStorage := storage.NewMarkdownStorage(wikisDir)

	// Initialize wiki generator
	wikiGen := generator.New(cfg, aiManager)

	server := &Server{
		config:        cfg,
		aiManager:     aiManager,
		wikiGenerator: wikiGen,
		storage:       markdownStorage,
		activeWikis:   make(map[string]*models.Wiki),
		repoURLToWiki: make(map[string]string),
		wikiLogs:      make(map[string][]string),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		wsConnections: make(map[string]*websocket.Conn),
	}

	server.setupRoutes()

	// Load existing wikis from storage
	if err := server.loadWikisFromStorage(); err != nil {
		log.Printf("Warning: Failed to load wikis from storage: %v", err)
	}

	// Start progress monitoring
	go server.monitorProgress()

	return server, nil
}

// loadWikisFromStorage loads all wikis from persistent storage
func (s *Server) loadWikisFromStorage() error {
	wikis, err := s.storage.LoadAllWikis()
	if err != nil {
		return fmt.Errorf("failed to load wikis: %w", err)
	}

	// Load wikis into memory
	for id, wiki := range wikis {
		s.activeWikis[id] = wiki
		// Rebuild repository URL mapping
		if wiki.PackagePath != "" {
			repoURL := utils.PackagePathToURL(wiki.PackagePath)
			s.repoURLToWiki[repoURL] = id
		}

		// Load logs for this wiki
		logs, err := s.storage.LoadLogs(id)
		if err != nil {
			log.Printf("Warning: Failed to load logs for wiki %s: %v", id, err)
		} else {
			s.wikiLogs[id] = logs
		}
	}

	log.Printf("Loaded %d wikis from storage", len(wikis))
	return nil
}

// saveWikiToStorage saves a wiki to persistent storage
func (s *Server) saveWikiToStorage(wiki *models.Wiki) error {
	if err := s.storage.SaveWiki(wiki); err != nil {
		return fmt.Errorf("failed to save wiki to storage: %w", err)
	}

	// Also save logs if they exist
	if logs, exists := s.wikiLogs[wiki.ID]; exists {
		if err := s.storage.SaveLogs(wiki.ID, logs); err != nil {
			log.Printf("Warning: Failed to save logs for wiki %s: %v", wiki.ID, err)
		}
	}

	return nil
}

// addWikiLog adds a log entry for a specific wiki
func (s *Server) addWikiLog(wikiID, message string) {
	if s.wikiLogs[wikiID] == nil {
		s.wikiLogs[wikiID] = make([]string, 0)
	}
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	s.wikiLogs[wikiID] = append(s.wikiLogs[wikiID], logEntry)

	// Keep only the last 100 log entries to prevent memory issues
	if len(s.wikiLogs[wikiID]) > 100 {
		s.wikiLogs[wikiID] = s.wikiLogs[wikiID][len(s.wikiLogs[wikiID])-100:]
	}
}

// monitorProgress monitors wiki generation progress and broadcasts updates
func (s *Server) monitorProgress() {
	progressCh := s.wikiGenerator.GetProgressChannel()
	log.Printf("Progress monitor started, listening for updates...")

	for progress := range progressCh {
		logMessage := fmt.Sprintf("Status: %s, Progress: %d%%, Step: %s",
			progress.Status, progress.Progress, progress.CurrentStep)
		if progress.Error != "" {
			logMessage += fmt.Sprintf(", Error: %s", progress.Error)
		}

		// Add to wiki logs
		s.addWikiLog(progress.WikiID, logMessage)

		log.Printf("Received progress update: WikiID=%s, Status=%s, Progress=%d, Step=%s, Error=%s",
			progress.WikiID, progress.Status, progress.Progress, progress.CurrentStep, progress.Error)

		// Update the wiki in activeWikis
		if wiki, exists := s.activeWikis[progress.WikiID]; exists {
			wiki.Status = progress.Status
			wiki.Progress = progress.Progress
			wiki.UpdatedAt = progress.UpdatedAt
			log.Printf("Updated wiki %s: Status=%s, Progress=%d", wiki.ID, wiki.Status, wiki.Progress)

			// Save updated wiki to storage
			if err := s.saveWikiToStorage(wiki); err != nil {
				log.Printf("Warning: Failed to save updated wiki to storage: %v", err)
			}
		} else {
			log.Printf("Warning: Wiki %s not found in activeWikis", progress.WikiID)
		}

		// Broadcast to WebSocket clients
		s.broadcastProgress(progress.WikiID, progress)
	}
}

// setupRoutes sets up the HTTP routes
func (s *Server) setupRoutes() {
	s.router = gin.Default()

	// Enable CORS if configured
	if s.config.Server.EnableCORS {
		s.router.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}

	// Serve static files
	s.router.Static("/static", s.config.Server.StaticDir)

	// Set up template functions
	s.router.SetFuncMap(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"jsonRaw": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	})
	s.router.LoadHTMLGlob(s.config.Server.TemplateDir + "/*")

	// Web interface routes
	s.router.GET("/", s.handleHome)
	s.router.GET("/debug", s.handleDebug)
	s.router.GET("/wiki/:id", s.handleWikiView)
	s.router.GET("/wiki/:id/page/:pageId", s.handlePageView)

	// Package path routes (e.g., /pkg/github.com/gorilla/websocket)
	s.router.GET("/pkg/*packagePath", s.handlePackageRoute)

	// API routes
	api := s.router.Group("/api")
	{
		// System info
		api.GET("/info", s.handleSystemInfo)
		api.GET("/providers", s.handleGetProviders)
		api.GET("/models", s.handleGetModels)

		// Wiki management
		api.POST("/wiki/generate", s.handleGenerateWiki)
		api.GET("/wiki/:id", s.handleGetWiki)
		api.GET("/wiki/:id/progress", s.handleGetProgress)
		api.GET("/wiki/:id/logs", s.handleGetLogs)
		api.GET("/wiki/logs", s.handleGetLogsQuery) // Alternative logs endpoint with query parameter
		api.DELETE("/wiki/:id", s.handleDeleteWiki)
		api.GET("/wikis", s.handleListWikis)
		api.GET("/tags", s.handleGetTags)
		api.GET("/wikis/by-tag/:tag", s.handleGetWikisByTag)

		// Wiki content
		api.GET("/wiki/:id/pages", s.handleGetPages)
		api.GET("/wiki/:id/page/:pageId", s.handleGetPage)
		api.GET("/wiki/:id/diagrams", s.handleGetDiagrams)
		api.GET("/wiki/:id/search", s.handleSearch)

		// Export
		api.GET("/wiki/:id/export/:format", s.handleExport)

		// RAG Chat
		api.POST("/wiki/:id/chat", s.handleChat)
		api.GET("/wiki/:id/chat/history", s.handleChatHistory)
	}

	// WebSocket for real-time updates
	s.router.GET("/ws/:wikiId", s.handleWebSocket)
}

// Start starts the server
func (s *Server) Start() error {
	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port),
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited")
	return nil
}

// handleHome serves the main page
func (s *Server) handleHome(c *gin.Context) {
	// Build providers data structure similar to handleSystemInfo
	providers := make(map[string]interface{})
	for name, provider := range s.aiManager.GetAllProviders() {
		providers[name] = map[string]interface{}{
			"name":      provider.GetName(),
			"available": provider.IsAvailable(),
			"models":    provider.GetModels(),
		}
	}

	models := s.aiManager.GetAllModels()

	// Debug logging
	log.Printf("Available providers: %+v", providers)
	log.Printf("Available models: %+v", models)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":     "KWiki - AI-Powered Wiki Generator",
		"providers": providers,
		"models":    models,
		"config":    s.config,
	})
}

// handleDebug serves the debug page
func (s *Server) handleDebug(c *gin.Context) {
	// Build providers data structure similar to handleSystemInfo
	providers := make(map[string]interface{})
	for name, provider := range s.aiManager.GetAllProviders() {
		providers[name] = map[string]interface{}{
			"name":      provider.GetName(),
			"available": provider.IsAvailable(),
			"models":    provider.GetModels(),
		}
	}

	models := s.aiManager.GetAllModels()

	c.HTML(http.StatusOK, "debug.html", gin.H{
		"title":     "Debug - KWiki",
		"providers": providers,
		"models":    models,
		"config":    s.config,
	})
}

// handleSystemInfo returns system information
func (s *Server) handleSystemInfo(c *gin.Context) {
	providers := make(map[string]interface{})
	for name, provider := range s.aiManager.GetAllProviders() {
		providers[name] = map[string]interface{}{
			"name":      provider.GetName(),
			"available": provider.IsAvailable(),
			"models":    provider.GetModels(),
			"usage":     provider.GetUsage(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"version":   "1.0.0",
		"providers": providers,
		"config": gin.H{
			"max_repo_size":   s.config.Repository.MaxRepoSize,
			"max_files":       s.config.Repository.MaxFiles,
			"enable_rag":      s.config.Generator.EnableRAG,
			"enable_diagrams": s.config.Generator.EnableDiagrams,
		},
	})
}

// handleGetProviders returns available AI providers
func (s *Server) handleGetProviders(c *gin.Context) {
	providers := make(map[string]interface{})
	for name, provider := range s.aiManager.GetAvailableProviders() {
		providers[name] = map[string]interface{}{
			"name":      provider.GetName(),
			"available": provider.IsAvailable(),
			"models":    provider.GetModels(),
		}
	}

	c.JSON(http.StatusOK, providers)
}

// handleGetModels returns available models for all providers
func (s *Server) handleGetModels(c *gin.Context) {
	models := s.aiManager.GetAllModels()
	c.JSON(http.StatusOK, models)
}

// handleGenerateWiki handles wiki generation requests
func (s *Server) handleGenerateWiki(c *gin.Context) {
	var req models.GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.RepositoryURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "repository_url is required"})
		return
	}

	// Debug: 打印接收到的请求数据
	log.Printf("接收到的生成请求:")
	log.Printf("  - RepositoryURL: %s", req.RepositoryURL)
	log.Printf("  - Settings.AIProvider: '%s'", req.Settings.AIProvider)
	log.Printf("  - Settings.Model: '%s'", req.Settings.Model)
	log.Printf("  - 默认提供商: %s", s.config.AI.DefaultProvider)

	// Set default AI settings if not provided
	if req.Settings.AIProvider == "" {
		log.Printf("AIProvider为空，设置为默认提供商: %s", s.config.AI.DefaultProvider)
		req.Settings.AIProvider = s.config.AI.DefaultProvider
	} else {
		log.Printf("使用指定的AIProvider: %s", req.Settings.AIProvider)
	}
	if req.Settings.Model == "" {
		// Get default model from provider config
		if providerConfig, exists := s.config.AI.Providers[req.Settings.AIProvider]; exists {
			req.Settings.Model = providerConfig.Model
		} else {
			req.Settings.Model = "deepseek-chat" // fallback default
		}
	}

	// Normalize repository URL for comparison
	normalizedURL := strings.TrimSuffix(strings.ToLower(req.RepositoryURL), "/")

	// Special handling for template documentation generation
	if normalizedURL == "template-docs" {
		s.handleTemplateDocsGeneration(c, req)
		return
	}

	// Check if a wiki for this repository is already being generated
	if existingWikiID, exists := s.repoURLToWiki[normalizedURL]; exists {
		if existingWiki, wikiExists := s.activeWikis[existingWikiID]; wikiExists {
			// Only block if the existing wiki is still in progress
			if existingWiki.Status == models.WikiStatusPending ||
				existingWiki.Status == models.WikiStatusAnalyzing ||
				existingWiki.Status == models.WikiStatusGenerating {

				c.JSON(http.StatusConflict, gin.H{
					"error":            "A wiki for this repository is already being generated",
					"existing_wiki_id": existingWiki.ID,
					"status":           existingWiki.Status,
					"repository_url":   req.RepositoryURL,
				})
				return
			}
		}
	}

	// Start wiki generation
	wiki, err := s.wikiGenerator.GenerateWiki(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store active wiki and repository URL mapping
	s.activeWikis[wiki.ID] = wiki
	s.repoURLToWiki[normalizedURL] = wiki.ID

	// Save to persistent storage
	if err := s.saveWikiToStorage(wiki); err != nil {
		log.Printf("Warning: Failed to save wiki to storage: %v", err)
	}

	// Add initial log entry
	s.addWikiLog(wiki.ID, fmt.Sprintf("Wiki generation started for repository: %s", req.RepositoryURL))
	s.addWikiLog(wiki.ID, fmt.Sprintf("Using AI provider: %s, Model: %s", req.Settings.AIProvider, req.Settings.Model))

	c.JSON(http.StatusOK, gin.H{
		"wiki_id": wiki.ID,
		"status":  wiki.Status,
		"message": "Wiki generation started",
	})
}

// handleTemplateDocsGeneration handles template documentation generation requests
func (s *Server) handleTemplateDocsGeneration(c *gin.Context, req models.GenerationRequest) {
	log.Printf("Starting template documentation generation")

	// Set default values for template docs generation
	if req.Title == "" {
		req.Title = "KWiki Template System Documentation"
	}
	if req.Description == "" {
		req.Description = "Comprehensive documentation for the KWiki template system"
	}
	if len(req.Languages) == 0 {
		req.Languages = []string{"zh"} // Default to Chinese
	}
	if req.PrimaryLanguage == "" {
		req.PrimaryLanguage = "zh"
	}

	// Validate AI settings
	if req.Settings.AIProvider == "" {
		req.Settings.AIProvider = s.config.AI.DefaultProvider
	}
	if req.Settings.Model == "" {
		// Get default model from provider config
		if providerConfig, exists := s.config.AI.Providers[req.Settings.AIProvider]; exists {
			req.Settings.Model = providerConfig.Model
		} else {
			req.Settings.Model = "deepseek-chat" // fallback default
		}
	}

	// Check if AI provider is available
	provider, exists := s.aiManager.GetProvider(req.Settings.AIProvider)
	if !exists || provider == nil || !provider.IsAvailable() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("AI provider '%s' is not available", req.Settings.AIProvider),
		})
		return
	}

	// Create a unique key for template docs generation
	templateDocsKey := fmt.Sprintf("template-docs-%s-%s", req.Settings.AIProvider, req.Settings.Model)

	// Check if template docs are already being generated with same settings
	if existingWikiID, exists := s.repoURLToWiki[templateDocsKey]; exists {
		if existingWiki, wikiExists := s.activeWikis[existingWikiID]; wikiExists {
			// Only block if the existing wiki is still in progress
			if existingWiki.Status == models.WikiStatusPending ||
				existingWiki.Status == models.WikiStatusAnalyzing ||
				existingWiki.Status == models.WikiStatusGenerating {

				c.JSON(http.StatusConflict, gin.H{
					"error":            "Template documentation is already being generated with these settings",
					"existing_wiki_id": existingWiki.ID,
					"status":           existingWiki.Status,
					"settings":         req.Settings,
				})
				return
			}
		}
	}

	// Start template documentation generation
	wiki, err := s.wikiGenerator.GenerateWiki(c.Request.Context(), req)
	if err != nil {
		log.Printf("Template docs generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store active wiki and mapping
	s.activeWikis[wiki.ID] = wiki
	s.repoURLToWiki[templateDocsKey] = wiki.ID

	// Save to persistent storage
	if err := s.saveWikiToStorage(wiki); err != nil {
		log.Printf("Warning: Failed to save template docs wiki to storage: %v", err)
	}

	// Add initial log entries
	s.addWikiLog(wiki.ID, "Template documentation generation started")
	s.addWikiLog(wiki.ID, fmt.Sprintf("Target languages: %v", req.Languages))
	s.addWikiLog(wiki.ID, fmt.Sprintf("Using AI provider: %s, Model: %s", req.Settings.AIProvider, req.Settings.Model))
	s.addWikiLog(wiki.ID, fmt.Sprintf("Title: %s", req.Title))

	log.Printf("Template docs generation started successfully, Wiki ID: %s", wiki.ID)

	c.JSON(http.StatusOK, gin.H{
		"wiki_id":   wiki.ID,
		"status":    wiki.Status,
		"message":   "Template documentation generation started",
		"title":     wiki.Title,
		"languages": req.Languages,
	})
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(c *gin.Context) {
	wikiID := c.Param("wikiId")

	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Store connection
	s.wsConnections[wikiID] = conn

	// Send initial status
	if wiki, exists := s.activeWikis[wikiID]; exists {
		conn.WriteJSON(map[string]interface{}{
			"type":     "status",
			"wiki_id":  wikiID,
			"status":   wiki.Status,
			"progress": wiki.Progress,
		})
	}

	// Keep connection alive and handle messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
	}

	// Clean up connection
	delete(s.wsConnections, wikiID)
}

// broadcastProgress broadcasts progress updates to WebSocket clients
func (s *Server) broadcastProgress(wikiID string, progress models.GenerationProgress) {
	if conn, exists := s.wsConnections[wikiID]; exists {
		conn.WriteJSON(map[string]interface{}{
			"type":         "progress",
			"wiki_id":      progress.WikiID,
			"status":       progress.Status,
			"progress":     progress.Progress,
			"current_step": progress.CurrentStep,
			"message":      progress.Message,
			"error":        progress.Error,
		})
	}
}

// handleGetWiki returns wiki information
func (s *Server) handleGetWiki(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	c.JSON(http.StatusOK, wiki)
}

// handleGetProgress returns generation progress
func (s *Server) handleGetProgress(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	progress := models.GenerationProgress{
		WikiID:      wikiID,
		Status:      wiki.Status,
		Progress:    wiki.Progress,
		CurrentStep: "Generation in progress",
		UpdatedAt:   wiki.UpdatedAt,
	}

	c.JSON(http.StatusOK, progress)
}

// getWikiIDFromParam extracts and decodes wiki ID from URL parameter
func getWikiIDFromParam(c *gin.Context, paramName string) string {
	param := c.Param(paramName)
	// URL decode the parameter to handle encoded slashes
	decoded, err := url.QueryUnescape(param)
	if err != nil {
		// If decoding fails, return the original parameter
		return param
	}
	return decoded
}

// handleGetLogs returns the generation logs for a specific wiki
func (s *Server) handleGetLogs(c *gin.Context) {
	rawParam := c.Param("id")
	wikiID := getWikiIDFromParam(c, "id")
	log.Printf("handleGetLogs: rawParam='%s', decoded wikiID='%s'", rawParam, wikiID)

	// First try to get logs from memory (for active generation)
	logs, exists := s.wikiLogs[wikiID]

	// If not in memory, try to read from storage
	if !exists {
		storageLogs, err := s.storage.LoadLogs(wikiID)
		if err != nil {
			log.Printf("Failed to read logs for wiki %s: %v", wikiID, err)
			logs = []string{} // Return empty array if no logs exist
		} else {
			logs = storageLogs
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"wiki_id": wikiID,
		"logs":    logs,
	})
}

// handleGetLogsQuery returns the generation logs using query parameter
func (s *Server) handleGetLogsQuery(c *gin.Context) {
	wikiID := c.Query("id")
	if wikiID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wiki ID is required"})
		return
	}

	log.Printf("handleGetLogsQuery: wikiID='%s'", wikiID)

	// First try to get logs from memory (for active generation)
	logs, exists := s.wikiLogs[wikiID]

	// If not in memory, try to read from storage
	if !exists {
		storageLogs, err := s.storage.LoadLogs(wikiID)
		if err != nil {
			log.Printf("Failed to read logs for wiki %s: %v", wikiID, err)
			logs = []string{} // Return empty array if no logs exist
		} else {
			logs = storageLogs
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"wiki_id": wikiID,
		"logs":    logs,
	})
}

// handleDeleteWiki deletes a wiki
func (s *Server) handleDeleteWiki(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	log.Printf("删除Wiki请求: %s", wikiID)

	// 首先尝试从持久化存储中删除
	if err := s.storage.DeleteWiki(wikiID); err != nil {
		log.Printf("从存储中删除Wiki失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete wiki from storage: " + err.Error()})
		return
	}

	// 从内存中删除
	delete(s.activeWikis, wikiID)

	// 清理仓库URL映射
	for repoURL, id := range s.repoURLToWiki {
		if id == wikiID {
			delete(s.repoURLToWiki, repoURL)
			break
		}
	}

	// 清理日志
	delete(s.wikiLogs, wikiID)

	// 关闭WebSocket连接
	if conn, exists := s.wsConnections[wikiID]; exists {
		conn.Close()
		delete(s.wsConnections, wikiID)
	}

	log.Printf("Wiki %s 删除成功", wikiID)
	c.JSON(http.StatusOK, gin.H{"message": "Wiki deleted successfully"})
}

// handleListWikis returns all wikis
func (s *Server) handleListWikis(c *gin.Context) {
	wikis := make([]*models.Wiki, 0, len(s.activeWikis))
	for _, wiki := range s.activeWikis {
		wikis = append(wikis, wiki)
	}

	c.JSON(http.StatusOK, wikis)
}

// handleGetTags returns all unique tags from all wikis
func (s *Server) handleGetTags(c *gin.Context) {
	tagSet := make(map[string]int) // tag -> count

	// Collect tags from all wikis
	for _, wiki := range s.activeWikis {
		// Wiki-level tags
		for _, tag := range wiki.Tags {
			tagSet[tag]++
		}

		// Page-level tags
		for _, page := range wiki.Pages {
			for _, tag := range page.Tags {
				tagSet[tag]++
			}
		}
	}

	// Convert to response format
	type TagInfo struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	tags := make([]TagInfo, 0, len(tagSet))
	for tag, count := range tagSet {
		tags = append(tags, TagInfo{
			Name:  tag,
			Count: count,
		})
	}

	c.JSON(http.StatusOK, tags)
}

// handleGetWikisByTag returns wikis filtered by tag
func (s *Server) handleGetWikisByTag(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag parameter is required"})
		return
	}

	var filteredWikis []*models.Wiki

	for _, wiki := range s.activeWikis {
		// Check wiki-level tags
		hasTag := false
		for _, wikiTag := range wiki.Tags {
			if wikiTag == tag {
				hasTag = true
				break
			}
		}

		// Check page-level tags if not found at wiki level
		if !hasTag {
			for _, page := range wiki.Pages {
				for _, pageTag := range page.Tags {
					if pageTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
		}

		if hasTag {
			filteredWikis = append(filteredWikis, wiki)
		}
	}

	c.JSON(http.StatusOK, filteredWikis)
}

// handleGetPages returns all pages for a wiki
func (s *Server) handleGetPages(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	c.JSON(http.StatusOK, wiki.Pages)
}

// handleGetPage returns a specific page
func (s *Server) handleGetPage(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")
	pageID := getWikiIDFromParam(c, "pageId")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	for _, page := range wiki.Pages {
		if page.ID == pageID {
			c.JSON(http.StatusOK, page)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
}

// handleGetDiagrams returns all diagrams for a wiki
func (s *Server) handleGetDiagrams(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	c.JSON(http.StatusOK, wiki.Diagrams)
}

// handleSearch performs search within a wiki
func (s *Server) handleSearch(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	results := s.performSearch(wiki, query)
	c.JSON(http.StatusOK, results)
}

// performSearch performs simple text search in wiki content
func (s *Server) performSearch(wiki *models.Wiki, query string) []models.SearchResult {
	var results []models.SearchResult
	query = strings.ToLower(query)

	// Search in pages
	for _, page := range wiki.Pages {
		if strings.Contains(strings.ToLower(page.Title), query) ||
			strings.Contains(strings.ToLower(page.Content), query) {

			// Extract excerpt
			content := strings.ToLower(page.Content)
			index := strings.Index(content, query)
			excerpt := ""
			if index >= 0 {
				start := index - 50
				if start < 0 {
					start = 0
				}
				end := index + len(query) + 50
				if end > len(page.Content) {
					end = len(page.Content)
				}
				excerpt = page.Content[start:end]
			}

			result := models.SearchResult{
				Type:    "page",
				ID:      page.ID,
				Title:   page.Title,
				Content: page.Content,
				Excerpt: excerpt,
				Score:   s.calculateSearchScore(page.Title, page.Content, query),
				URL:     fmt.Sprintf("/wiki/%s/page/%s", wiki.ID, page.ID),
				PageID:  page.ID,
			}
			results = append(results, result)
		}
	}

	// Search in diagrams
	for _, diagram := range wiki.Diagrams {
		if strings.Contains(strings.ToLower(diagram.Title), query) ||
			strings.Contains(strings.ToLower(diagram.Description), query) {

			result := models.SearchResult{
				Type:      "diagram",
				ID:        diagram.ID,
				Title:     diagram.Title,
				Content:   diagram.Description,
				Excerpt:   diagram.Description,
				Score:     s.calculateSearchScore(diagram.Title, diagram.Description, query),
				URL:       fmt.Sprintf("/wiki/%s#diagram-%s", wiki.ID, diagram.ID),
				DiagramID: diagram.ID,
			}
			results = append(results, result)
		}
	}

	return results
}

// calculateSearchScore calculates a simple search relevance score
func (s *Server) calculateSearchScore(title, content, query string) float64 {
	score := 0.0
	query = strings.ToLower(query)
	title = strings.ToLower(title)
	content = strings.ToLower(content)

	// Title matches are more important
	if strings.Contains(title, query) {
		score += 10.0
	}

	// Count occurrences in content
	occurrences := strings.Count(content, query)
	score += float64(occurrences)

	// Normalize by content length
	if len(content) > 0 {
		score = score / float64(len(content)) * 1000
	}

	return score
}

// handleWikiView serves the wiki view page
func (s *Server) handleWikiView(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Wiki Not Found",
			"error": "The requested wiki was not found",
		})
		return
	}

	c.HTML(http.StatusOK, "wiki.html", gin.H{
		"title": wiki.Title,
		"wiki":  wiki,
	})
}

// handlePackageRoute handles both wiki and page requests by package path
func (s *Server) handlePackageRoute(c *gin.Context) {
	packagePath := strings.TrimPrefix(c.Param("packagePath"), "/")
	log.Printf("handlePackageRoute: 请求包路径: %s", packagePath)
	log.Printf("handlePackageRoute: 当前活跃wikis数量: %d", len(s.activeWikis))

	for id, wiki := range s.activeWikis {
		log.Printf("handlePackageRoute: Wiki ID=%s, PackagePath=%s", id, wiki.PackagePath)
	}

	// Check if this is a page request (ends with /page/pageId)
	if strings.Contains(packagePath, "/page/") {
		// Extract package path and page ID
		parts := strings.Split(packagePath, "/page/")
		if len(parts) == 2 {
			actualPackagePath := parts[0]
			pageID := parts[1]

			// Find wiki by package path
			var foundWiki *models.Wiki
			for _, wiki := range s.activeWikis {
				if wiki.PackagePath == actualPackagePath {
					foundWiki = wiki
					break
				}
			}

			if foundWiki == nil {
				c.HTML(http.StatusNotFound, "error.html", gin.H{
					"title": "Wiki Not Found",
					"error": fmt.Sprintf("Wiki not found for package: %s", actualPackagePath),
				})
				return
			}

			// Find the specific page
			var foundPage *models.WikiPage
			for _, page := range foundWiki.Pages {
				if page.ID == pageID {
					foundPage = &page
					break
				}
			}

			if foundPage == nil {
				c.HTML(http.StatusNotFound, "error.html", gin.H{
					"title": "Page Not Found",
					"error": fmt.Sprintf("Page not found: %s", pageID),
				})
				return
			}

			c.HTML(http.StatusOK, "wiki.html", gin.H{
				"title": foundPage.Title,
				"wiki":  foundWiki,
				"page":  foundPage,
			})
			return
		}
	}

	// This is a wiki request
	// Find wiki by package path
	var foundWiki *models.Wiki
	for _, wiki := range s.activeWikis {
		if wiki.PackagePath == packagePath {
			foundWiki = wiki
			break
		}
	}

	if foundWiki == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Wiki Not Found",
			"error": fmt.Sprintf("Wiki not found for package: %s", packagePath),
		})
		return
	}

	c.HTML(http.StatusOK, "wiki.html", gin.H{
		"title": foundWiki.Title,
		"wiki":  foundWiki,
	})
}

// handlePageView serves a specific wiki page
func (s *Server) handlePageView(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")
	pageID := getWikiIDFromParam(c, "pageId")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Wiki Not Found",
			"error": "The requested wiki was not found",
		})
		return
	}

	var page *models.WikiPage
	for _, p := range wiki.Pages {
		if p.ID == pageID {
			page = &p
			break
		}
	}

	if page == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Page Not Found",
			"error": "The requested page was not found",
		})
		return
	}

	c.HTML(http.StatusOK, "page.html", gin.H{
		"title": page.Title,
		"wiki":  wiki,
		"page":  page,
	})
}

// handleExport handles wiki export requests
func (s *Server) handleExport(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")
	format := c.Param("format")

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	switch format {
	case "markdown":
		s.exportMarkdown(c, wiki)
	case "json":
		s.exportJSON(c, wiki)
	case "html":
		s.exportHTML(c, wiki)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format"})
	}
}

// exportMarkdown exports wiki as markdown
func (s *Server) exportMarkdown(c *gin.Context, wiki *models.Wiki) {
	var content strings.Builder

	// Write title and description
	content.WriteString(fmt.Sprintf("# %s\n\n", wiki.Title))
	if wiki.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", wiki.Description))
	}

	// Write table of contents
	content.WriteString("## Table of Contents\n\n")
	for _, page := range wiki.Pages {
		content.WriteString(fmt.Sprintf("- [%s](#%s)\n", page.Title, strings.ToLower(strings.ReplaceAll(page.Title, " ", "-"))))
	}
	content.WriteString("\n")

	// Write pages
	for _, page := range wiki.Pages {
		content.WriteString(fmt.Sprintf("## %s\n\n", page.Title))
		content.WriteString(fmt.Sprintf("%s\n\n", page.Content))
	}

	// Write diagrams
	if len(wiki.Diagrams) > 0 {
		content.WriteString("## Diagrams\n\n")
		for _, diagram := range wiki.Diagrams {
			content.WriteString(fmt.Sprintf("### %s\n\n", diagram.Title))
			if diagram.Description != "" {
				content.WriteString(fmt.Sprintf("%s\n\n", diagram.Description))
			}
			content.WriteString("```mermaid\n")
			content.WriteString(diagram.Content)
			content.WriteString("\n```\n\n")
		}
	}

	filename := fmt.Sprintf("%s-wiki.md", strings.ReplaceAll(wiki.Title, " ", "-"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/markdown")
	c.String(http.StatusOK, content.String())
}

// exportJSON exports wiki as JSON
func (s *Server) exportJSON(c *gin.Context, wiki *models.Wiki) {
	filename := fmt.Sprintf("%s-wiki.json", strings.ReplaceAll(wiki.Title, " ", "-"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.JSON(http.StatusOK, wiki)
}

// exportHTML exports wiki as HTML
func (s *Server) exportHTML(c *gin.Context, wiki *models.Wiki) {
	var content strings.Builder

	// HTML header
	content.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + wiki.Title + `</title>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2, h3 { color: #333; }
        pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
        code { background: #f5f5f5; padding: 2px 4px; border-radius: 3px; }
        .mermaid { text-align: center; }
    </style>
</head>
<body>`)

	// Title and description
	content.WriteString(fmt.Sprintf("<h1>%s</h1>", wiki.Title))
	if wiki.Description != "" {
		content.WriteString(fmt.Sprintf("<p>%s</p>", wiki.Description))
	}

	// Table of contents
	content.WriteString("<h2>Table of Contents</h2><ul>")
	for _, page := range wiki.Pages {
		anchor := strings.ToLower(strings.ReplaceAll(page.Title, " ", "-"))
		content.WriteString(fmt.Sprintf(`<li><a href="#%s">%s</a></li>`, anchor, page.Title))
	}
	content.WriteString("</ul>")

	// Pages (convert markdown to HTML - simplified)
	for _, page := range wiki.Pages {
		anchor := strings.ToLower(strings.ReplaceAll(page.Title, " ", "-"))
		content.WriteString(fmt.Sprintf(`<h2 id="%s">%s</h2>`, anchor, page.Title))

		// Simple markdown to HTML conversion
		htmlContent := strings.ReplaceAll(page.Content, "\n", "<br>")
		htmlContent = strings.ReplaceAll(htmlContent, "**", "<strong>")
		htmlContent = strings.ReplaceAll(htmlContent, "**", "</strong>")
		content.WriteString(fmt.Sprintf("<div>%s</div>", htmlContent))
	}

	// Diagrams
	if len(wiki.Diagrams) > 0 {
		content.WriteString("<h2>Diagrams</h2>")
		for _, diagram := range wiki.Diagrams {
			content.WriteString(fmt.Sprintf("<h3>%s</h3>", diagram.Title))
			if diagram.Description != "" {
				content.WriteString(fmt.Sprintf("<p>%s</p>", diagram.Description))
			}
			content.WriteString(fmt.Sprintf(`<div class="mermaid">%s</div>`, diagram.Content))
		}
	}

	// HTML footer
	content.WriteString(`
    <script>
        mermaid.initialize({ startOnLoad: true });
    </script>
</body>
</html>`)

	filename := fmt.Sprintf("%s-wiki.html", strings.ReplaceAll(wiki.Title, " ", "-"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, content.String())
}

// handleChat handles RAG chat requests
func (s *Server) handleChat(c *gin.Context) {
	wikiID := getWikiIDFromParam(c, "id")

	var req struct {
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wiki, exists := s.activeWikis[wikiID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wiki not found"})
		return
	}

	if !wiki.Settings.EnableRAG {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RAG chat is not enabled for this wiki"})
		return
	}

	// Simple RAG implementation - find relevant content
	context := s.findRelevantContent(wiki, req.Message)

	// Generate response using AI
	prompt := fmt.Sprintf(`Based on the following code documentation, answer the user's question.

Context:
%s

User Question: %s

Please provide a helpful and accurate answer based on the documentation provided.`,
		strings.Join(context, "\n\n"), req.Message)

	response, err := s.aiManager.GenerateText(c.Request.Context(), wiki.Settings.AIProvider, prompt, ai.GenerationOptions{
		Model:        wiki.Settings.Model,
		Temperature:  0.7,
		MaxTokens:    1000,
		SystemPrompt: "You are a helpful assistant that answers questions about code documentation. Be concise and accurate.",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
		return
	}

	chatMessage := models.ChatMessage{
		ID:         utils.GenerateID(),
		WikiID:     wikiID,
		Role:       models.MessageRoleAssistant,
		Content:    response.Text,
		Context:    context,
		Timestamp:  time.Now(),
		TokensUsed: response.TokensUsed,
	}

	c.JSON(http.StatusOK, chatMessage)
}

// findRelevantContent finds relevant content for RAG
func (s *Server) findRelevantContent(wiki *models.Wiki, query string) []string {
	var relevantContent []string
	query = strings.ToLower(query)

	// Search in pages for relevant content
	for _, page := range wiki.Pages {
		if strings.Contains(strings.ToLower(page.Title), query) ||
			strings.Contains(strings.ToLower(page.Content), query) {

			// Add page content (truncated if too long)
			content := page.Content
			if len(content) > 1000 {
				content = content[:1000] + "..."
			}
			relevantContent = append(relevantContent, fmt.Sprintf("From %s:\n%s", page.Title, content))

			if len(relevantContent) >= 3 { // Limit to prevent context overflow
				break
			}
		}
	}

	return relevantContent
}

// handleChatHistory returns chat history (placeholder)
func (s *Server) handleChatHistory(c *gin.Context) {
	_ = getWikiIDFromParam(c, "id") // wikiID for future use

	// In a real implementation, this would retrieve chat history from storage
	// For now, return empty history
	c.JSON(http.StatusOK, []models.ChatMessage{})
}
