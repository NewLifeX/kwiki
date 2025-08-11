# KWiki Makefile

.PHONY: help build run clean test docker docker-build docker-run deps install dev

# 默认目标
help: ## 显示帮助信息
	@echo "KWiki - AI-Powered Wiki Generator"
	@echo "================================="
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 应用信息
APP_NAME := kwiki
VERSION := 1.0.0
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 依赖管理
deps: ## 安装 Go 依赖
	@echo "📦 Installing Go dependencies..."
	go mod download
	go mod tidy

# 构建应用
build: deps ## 构建应用
	@echo "🔨 Building $(APP_NAME)..."
	go build $(LDFLAGS) -o $(APP_NAME) ./cmd/kwiki
	@echo "✅ Build complete: ./$(APP_NAME)"

# 运行应用
run: build ## 构建并运行应用
	@echo "🚀 Starting $(APP_NAME)..."
	./$(APP_NAME)

# 开发模式运行
dev: ## 开发模式运行 (使用 go run)
	@echo "🔧 Running in development mode..."
	go run ./cmd/kwiki

# 安装到系统
install: build ## 安装到 $GOPATH/bin
	@echo "📥 Installing $(APP_NAME) to $$GOPATH/bin..."
	go install $(LDFLAGS) ./cmd/kwiki
	@echo "✅ Installed successfully"

# 测试
test: ## 运行测试
	@echo "🧪 Running tests..."
	go test -v ./...

# 测试覆盖率
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# 代码检查
lint: ## 运行代码检查
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install it with:"; \
		echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 格式化代码
fmt: ## 格式化代码
	@echo "🎨 Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# 清理构建文件
clean: ## 清理构建文件
	@echo "🧹 Cleaning up..."
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	rm -rf repos output data
	@echo "✅ Cleanup complete"

# Docker 相关命令
docker-build: ## 构建 Docker 镜像
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "✅ Docker image built: $(APP_NAME):$(VERSION)"

docker-run: ## 运行 Docker 容器
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 \
		-e OPENAI_API_KEY="$(OPENAI_API_KEY)" \
		-e GOOGLE_API_KEY="$(GOOGLE_API_KEY)" \
		-e OLLAMA_HOST="$(OLLAMA_HOST)" \
		-v $(PWD)/data:/root/data \
		$(APP_NAME):latest

docker: docker-build docker-run ## 构建并运行 Docker 容器

# Docker Compose 命令
compose-up: ## 使用 Docker Compose 启动服务
	@echo "🐳 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started. Access KWiki at http://localhost:8080"

compose-down: ## 停止 Docker Compose 服务
	@echo "🐳 Stopping Docker Compose services..."
	docker-compose down

compose-logs: ## 查看 Docker Compose 日志
	docker-compose logs -f

# 环境设置
setup-env: ## 创建环境配置文件
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file from template..."; \
		cp .env.example .env; \
		echo "✅ Created .env file. Please edit it with your configuration."; \
	else \
		echo "⚠️  .env file already exists."; \
	fi

# Ollama 相关命令
ollama-install: ## 安装 Ollama (macOS/Linux)
	@echo "📥 Installing Ollama..."
	@if command -v ollama >/dev/null 2>&1; then \
		echo "✅ Ollama is already installed"; \
	else \
		curl -fsSL https://ollama.ai/install.sh | sh; \
	fi

ollama-start: ## 启动 Ollama 服务
	@echo "🚀 Starting Ollama service..."
	ollama serve &

ollama-pull: ## 下载推荐的 AI 模型
	@echo "📥 Downloading recommended models..."
	ollama pull llama3.2:latest
	ollama pull codellama:latest
	@echo "✅ Models downloaded"

ollama-setup: ollama-install ollama-start ollama-pull ## 完整设置 Ollama

# 快速开始
quickstart: setup-env ollama-setup build ## 快速开始 (设置环境 + Ollama + 构建)
	@echo ""
	@echo "🎉 KWiki is ready!"
	@echo "Run 'make run' to start the server"
	@echo "Or run './start.sh' for guided startup"

# 发布相关
release: clean test build ## 准备发布版本
	@echo "📦 Preparing release $(VERSION)..."
	mkdir -p dist
	tar -czf dist/$(APP_NAME)-$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz $(APP_NAME) config.yaml README.md
	@echo "✅ Release package created: dist/$(APP_NAME)-$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz"

# 显示版本信息
version: ## 显示版本信息
	@echo "$(APP_NAME) version $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# 显示状态
status: ## 显示项目状态
	@echo "📊 Project Status"
	@echo "================"
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(shell go version)"
	@echo "Git Status: $(shell git status --porcelain | wc -l) files changed"
	@echo "Dependencies: $(shell go list -m all | wc -l) modules"
	@echo ""
	@echo "🔧 Environment:"
	@echo "OPENAI_API_KEY: $(if $(OPENAI_API_KEY),✅ Set,❌ Not set)"
	@echo "GOOGLE_API_KEY: $(if $(GOOGLE_API_KEY),✅ Set,❌ Not set)"
	@echo "OLLAMA_HOST: $(if $(OLLAMA_HOST),$(OLLAMA_HOST),http://localhost:11434 (default))"
