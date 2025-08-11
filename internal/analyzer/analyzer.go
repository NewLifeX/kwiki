package analyzer

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/stcn52/kwiki/internal/config"
	"github.com/stcn52/kwiki/pkg/models"
	"github.com/stcn52/kwiki/pkg/utils"
)

// CodeAnalyzer handles repository analysis
type CodeAnalyzer struct {
	config *config.Config
}

// New creates a new code analyzer
func New(cfg *config.Config) *CodeAnalyzer {
	return &CodeAnalyzer{
		config: cfg,
	}
}

// AnalyzeRepository clones and analyzes a repository
func (ca *CodeAnalyzer) AnalyzeRepository(ctx context.Context, repoURL, branch, accessToken string) (*models.Repository, error) {
	// Parse repository URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Extract owner and repo name
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid repository URL format")
	}

	owner := pathParts[0]
	repoName := strings.TrimSuffix(pathParts[1], ".git")
	provider := ca.detectProvider(parsedURL.Host)

	// Create repository model
	repo := &models.Repository{
		ID:        utils.GenerateID(),
		URL:       repoURL,
		Name:      repoName,
		Owner:     owner,
		Provider:  provider,
		Branch:    branch,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if branch == "" {
		repo.Branch = "main" // Default branch
	}

	// Clone repository
	localPath := filepath.Join(ca.config.Repository.CloneDir, repo.ID)
	repo.LocalPath = localPath

	err = ca.cloneRepository(repoURL, localPath, repo.Branch, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Analyze repository structure
	err = ca.analyzeRepoStructure(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repository structure: %w", err)
	}

	return repo, nil
}

// cloneRepository clones a Git repository
func (ca *CodeAnalyzer) cloneRepository(repoURL, localPath, branch, accessToken string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return err
	}

	// Prepare clone options
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	if branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + branch)
		cloneOptions.SingleBranch = true
	}

	// Add authentication if access token is provided
	if accessToken != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: "token", // Can be anything for token auth
			Password: accessToken,
		}
	}

	// Clone the repository
	_, err := git.PlainClone(localPath, false, cloneOptions)
	if err != nil {
		return err
	}

	return nil
}

// analyzeRepoStructure analyzes the structure of the cloned repository
func (ca *CodeAnalyzer) analyzeRepoStructure(repo *models.Repository) error {
	var totalSize int64
	var fileCount int
	languageCount := make(map[string]int)

	err := filepath.WalkDir(repo.LocalPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory and other excluded patterns
		relPath, _ := filepath.Rel(repo.LocalPath, path)
		if ca.shouldExclude(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}

			totalSize += info.Size()
			fileCount++

			// Detect language by file extension
			ext := strings.ToLower(filepath.Ext(path))
			if lang := ca.detectLanguage(ext); lang != "" {
				languageCount[lang]++
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	repo.Size = totalSize
	repo.FileCount = fileCount

	// Extract top languages
	var languages []string
	for lang := range languageCount {
		languages = append(languages, lang)
	}
	repo.Languages = languages

	// Try to read README for description
	repo.Description = ca.extractDescription(repo.LocalPath)

	return nil
}

// AnalyzeCodeStructure analyzes the code structure of a repository
func (ca *CodeAnalyzer) AnalyzeCodeStructure(ctx context.Context, repo *models.Repository) (*models.CodeStructure, error) {
	structure := &models.CodeStructure{
		RepositoryID:  repo.ID,
		Files:         []models.FileInfo{},
		Dependencies:  []models.Dependency{},
		Modules:       []models.Module{},
		Functions:     []models.Function{},
		Classes:       []models.Class{},
		Relationships: []models.Relationship{},
		Metrics:       models.CodeMetrics{},
	}

	// Analyze files
	err := filepath.WalkDir(repo.LocalPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(repo.LocalPath, path)
		if ca.shouldExclude(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && ca.shouldInclude(relPath) {
			fileInfo, err := ca.analyzeFile(path, relPath)
			if err != nil {
				return err
			}
			structure.Files = append(structure.Files, *fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Analyze dependencies
	structure.Dependencies = ca.analyzeDependencies(repo.LocalPath)

	// Analyze modules and extract functions/classes
	structure.Modules, structure.Functions, structure.Classes = ca.analyzeModules(structure.Files)

	// Calculate metrics
	structure.Metrics = ca.calculateMetrics(structure)

	return structure, nil
}

// analyzeFile analyzes a single file
func (ca *CodeAnalyzer) analyzeFile(fullPath, relPath string) (*models.FileInfo, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(relPath))
	language := ca.detectLanguage(ext)

	fileInfo := &models.FileInfo{
		Path:        relPath,
		Name:        filepath.Base(relPath),
		Extension:   ext,
		Size:        info.Size(),
		Language:    language,
		ModifiedAt:  info.ModTime(),
		IsDirectory: info.IsDir(),
	}

	// Read file content for analysis (limit size to prevent memory issues)
	if info.Size() < 1024*1024 { // 1MB limit
		content, err := os.ReadFile(fullPath)
		if err == nil {
			fileInfo.Content = string(content)
			fileInfo.LineCount = strings.Count(string(content), "\n") + 1
			fileInfo.Hash = utils.HashString(string(content))
		}
	}

	return fileInfo, nil
}

// shouldExclude checks if a path should be excluded
func (ca *CodeAnalyzer) shouldExclude(path string) bool {
	for _, pattern := range ca.config.Repository.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// shouldInclude checks if a file should be included
func (ca *CodeAnalyzer) shouldInclude(path string) bool {
	if len(ca.config.Repository.IncludePatterns) == 0 {
		return true
	}

	for _, pattern := range ca.config.Repository.IncludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// detectProvider detects the Git provider from hostname
func (ca *CodeAnalyzer) detectProvider(host string) string {
	switch {
	case strings.Contains(host, "github"):
		return "github"
	case strings.Contains(host, "gitlab"):
		return "gitlab"
	case strings.Contains(host, "bitbucket"):
		return "bitbucket"
	default:
		return "unknown"
	}
}

// detectLanguage detects programming language from file extension
func (ca *CodeAnalyzer) detectLanguage(ext string) string {
	languageMap := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".java":  "Java",
		".cpp":   "C++",
		".c":     "C",
		".h":     "C/C++",
		".rs":    "Rust",
		".rb":    "Ruby",
		".php":   "PHP",
		".cs":    "C#",
		".kt":    "Kotlin",
		".swift": "Swift",
		".scala": "Scala",
		".r":     "R",
		".m":     "Objective-C",
		".sh":    "Shell",
		".ps1":   "PowerShell",
		".sql":   "SQL",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".less":  "LESS",
		".vue":   "Vue",
		".jsx":   "JSX",
		".tsx":   "TSX",
		".dart":  "Dart",
		".lua":   "Lua",
		".pl":    "Perl",
		".clj":   "Clojure",
		".ex":    "Elixir",
		".erl":   "Erlang",
		".hs":    "Haskell",
		".ml":    "OCaml",
		".fs":    "F#",
		".jl":    "Julia",
		".nim":   "Nim",
		".zig":   "Zig",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return ""
}

// extractDescription tries to extract description from README files
func (ca *CodeAnalyzer) extractDescription(repoPath string) string {
	readmeFiles := []string{"README.md", "README.txt", "README.rst", "README"}

	for _, filename := range readmeFiles {
		readmePath := filepath.Join(repoPath, filename)
		if content, err := os.ReadFile(readmePath); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "!") {
					if len(line) > 10 && len(line) < 200 {
						return line
					}
				}
			}
		}
	}

	return ""
}

// analyzeDependencies analyzes project dependencies
func (ca *CodeAnalyzer) analyzeDependencies(repoPath string) []models.Dependency {
	var dependencies []models.Dependency

	// Check for different dependency files
	depFiles := map[string]string{
		"package.json":     "npm",
		"requirements.txt": "pip",
		"Cargo.toml":       "cargo",
		"go.mod":           "go",
		"pom.xml":          "maven",
		"build.gradle":     "gradle",
		"composer.json":    "composer",
		"Gemfile":          "bundler",
	}

	for filename, source := range depFiles {
		depPath := filepath.Join(repoPath, filename)
		if _, err := os.Stat(depPath); err == nil {
			deps := ca.parseDependencyFile(depPath, source)
			dependencies = append(dependencies, deps...)
		}
	}

	return dependencies
}

// parseDependencyFile parses a dependency file (simplified implementation)
func (ca *CodeAnalyzer) parseDependencyFile(filePath, source string) []models.Dependency {
	var dependencies []models.Dependency

	content, err := os.ReadFile(filePath)
	if err != nil {
		return dependencies
	}

	// This is a simplified parser - in a real implementation,
	// you would use proper parsers for each file type
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			// Extract dependency name (very basic parsing)
			if strings.Contains(line, "\"") {
				parts := strings.Split(line, "\"")
				if len(parts) >= 2 {
					name := parts[1]
					if name != "" && !strings.Contains(name, "/") {
						dep := models.Dependency{
							Name:   name,
							Source: source,
							Type:   "direct",
						}
						dependencies = append(dependencies, dep)
					}
				}
			}
		}
	}

	return dependencies
}

// analyzeModules analyzes modules and extracts functions/classes
func (ca *CodeAnalyzer) analyzeModules(files []models.FileInfo) ([]models.Module, []models.Function, []models.Class) {
	var modules []models.Module
	var functions []models.Function
	var classes []models.Class

	// Group files by directory to create modules
	moduleMap := make(map[string][]models.FileInfo)
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if dir == "." {
			dir = "root"
		}
		moduleMap[dir] = append(moduleMap[dir], file)
	}

	// Create modules
	for dirPath, moduleFiles := range moduleMap {
		module := models.Module{
			Name:     filepath.Base(dirPath),
			Path:     dirPath,
			Language: ca.detectModuleLanguage(moduleFiles),
		}

		var totalLines int
		for _, file := range moduleFiles {
			totalLines += file.LineCount

			// Extract functions and classes (simplified)
			fileFunctions, fileClasses := ca.extractCodeElements(file)
			functions = append(functions, fileFunctions...)
			classes = append(classes, fileClasses...)
		}

		module.LineCount = totalLines
		modules = append(modules, module)
	}

	return modules, functions, classes
}

// detectModuleLanguage detects the primary language of a module
func (ca *CodeAnalyzer) detectModuleLanguage(files []models.FileInfo) string {
	langCount := make(map[string]int)
	for _, file := range files {
		if file.Language != "" {
			langCount[file.Language]++
		}
	}

	var maxLang string
	var maxCount int
	for lang, count := range langCount {
		if count > maxCount {
			maxCount = count
			maxLang = lang
		}
	}

	return maxLang
}

// extractCodeElements extracts functions and classes from a file (simplified)
func (ca *CodeAnalyzer) extractCodeElements(file models.FileInfo) ([]models.Function, []models.Class) {
	var functions []models.Function
	var classes []models.Class

	if file.Content == "" {
		return functions, classes
	}

	lines := strings.Split(file.Content, "\n")

	// Simple pattern matching for different languages
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract functions (very basic pattern matching)
		if ca.isFunctionDeclaration(line, file.Language) {
			fn := models.Function{
				Name:      ca.extractFunctionName(line, file.Language),
				File:      file.Path,
				StartLine: i + 1,
				Language:  file.Language,
				Signature: line,
				IsPublic:  ca.isPublicFunction(line, file.Language),
			}
			functions = append(functions, fn)
		}

		// Extract classes (very basic pattern matching)
		if ca.isClassDeclaration(line, file.Language) {
			class := models.Class{
				Name:      ca.extractClassName(line, file.Language),
				File:      file.Path,
				StartLine: i + 1,
				Language:  file.Language,
				IsPublic:  ca.isPublicClass(line, file.Language),
			}
			classes = append(classes, class)
		}
	}

	return functions, classes
}

// Helper functions for code element extraction (simplified implementations)
func (ca *CodeAnalyzer) isFunctionDeclaration(line, language string) bool {
	switch language {
	case "Go":
		return strings.Contains(line, "func ")
	case "Python":
		return strings.Contains(line, "def ")
	case "JavaScript", "TypeScript":
		return strings.Contains(line, "function ") || strings.Contains(line, "=> ")
	case "Java", "C#":
		return strings.Contains(line, "(") && strings.Contains(line, ")") &&
			(strings.Contains(line, "public ") || strings.Contains(line, "private ") || strings.Contains(line, "protected "))
	default:
		return false
	}
}

func (ca *CodeAnalyzer) isClassDeclaration(line, language string) bool {
	switch language {
	case "Go":
		return strings.Contains(line, "type ") && strings.Contains(line, "struct")
	case "Python":
		return strings.Contains(line, "class ")
	case "JavaScript", "TypeScript":
		return strings.Contains(line, "class ")
	case "Java", "C#":
		return strings.Contains(line, "class ") || strings.Contains(line, "interface ")
	default:
		return false
	}
}

func (ca *CodeAnalyzer) extractFunctionName(line, language string) string {
	// Simplified name extraction
	words := strings.Fields(line)
	for i, word := range words {
		if word == "func" || word == "def" || word == "function" {
			if i+1 < len(words) {
				name := words[i+1]
				if idx := strings.Index(name, "("); idx > 0 {
					return name[:idx]
				}
				return name
			}
		}
	}
	return "unknown"
}

func (ca *CodeAnalyzer) extractClassName(line, language string) string {
	// Simplified name extraction
	words := strings.Fields(line)
	for i, word := range words {
		if word == "class" || word == "type" {
			if i+1 < len(words) {
				name := words[i+1]
				if idx := strings.Index(name, " "); idx > 0 {
					return name[:idx]
				}
				return name
			}
		}
	}
	return "unknown"
}

func (ca *CodeAnalyzer) isPublicFunction(line, language string) bool {
	switch language {
	case "Go":
		// In Go, functions starting with uppercase are public
		name := ca.extractFunctionName(line, language)
		return len(name) > 0 && strings.ToUpper(name[:1]) == name[:1]
	case "Java", "C#":
		return strings.Contains(line, "public ")
	default:
		return true // Default to public
	}
}

func (ca *CodeAnalyzer) isPublicClass(line, language string) bool {
	switch language {
	case "Go":
		// In Go, types starting with uppercase are public
		name := ca.extractClassName(line, language)
		return len(name) > 0 && strings.ToUpper(name[:1]) == name[:1]
	case "Java", "C#":
		return strings.Contains(line, "public ")
	default:
		return true // Default to public
	}
}

// calculateMetrics calculates code metrics
func (ca *CodeAnalyzer) calculateMetrics(structure *models.CodeStructure) models.CodeMetrics {
	var totalLines, codeLines, commentLines, blankLines int
	var totalComplexity int

	for _, file := range structure.Files {
		totalLines += file.LineCount

		// Simple heuristic for code vs comments vs blank lines
		if file.Content != "" {
			lines := strings.Split(file.Content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					blankLines++
				} else if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
					commentLines++
				} else {
					codeLines++
				}
			}
		}
	}

	// Calculate average complexity
	var avgComplexity float64
	if len(structure.Functions) > 0 {
		for _, fn := range structure.Functions {
			totalComplexity += fn.Complexity
		}
		avgComplexity = float64(totalComplexity) / float64(len(structure.Functions))
	}

	// Find max complexity
	var maxComplexity int
	for _, fn := range structure.Functions {
		if fn.Complexity > maxComplexity {
			maxComplexity = fn.Complexity
		}
	}

	return models.CodeMetrics{
		TotalLines:        totalLines,
		CodeLines:         codeLines,
		CommentLines:      commentLines,
		TotalFiles:        len(structure.Files),
		TotalFunctions:    len(structure.Functions),
		TotalClasses:      len(structure.Classes),
		AverageComplexity: avgComplexity,
		MaxComplexity:     maxComplexity,
	}
}
