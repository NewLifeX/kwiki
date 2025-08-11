#!/bin/bash

# KWiki 启动脚本

set -e

echo "🚀 Starting KWiki - AI-Powered Wiki Generator"
echo "============================================="

# 检查 Go 版本
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24 or later."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# 检查并创建必要的目录
echo "📁 Creating necessary directories..."
mkdir -p repos output data/repos data/output web/static

# 加载环境变量 (在检查之前)
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "⚠️  .env file not found. You can copy .env.example to .env and configure your API keys."
fi

# 安装依赖
echo "📦 Installing dependencies..."
go mod tidy

# 检查环境变量
echo "🔧 Checking configuration..."

# 检查AI提供者API密钥
AI_PROVIDERS_FOUND=0

if [ -n "$OPENAI_API_KEY" ]; then
    echo "✅ OpenAI API key found"
    AI_PROVIDERS_FOUND=1
fi

if [ -n "$GOOGLE_API_KEY" ]; then
    echo "✅ Google (Gemini) API key found"
    AI_PROVIDERS_FOUND=1
fi

if [ -n "$DEEPSEEK_API_KEY" ]; then
    echo "✅ DeepSeek API key found"
    AI_PROVIDERS_FOUND=1
fi

if [ $AI_PROVIDERS_FOUND -eq 0 ]; then
    echo "⚠️  Warning: No cloud AI provider API keys found."
    echo "   You can still use Ollama for local AI generation."
    echo "   Set one of the following environment variables to use cloud AI providers:"
    echo "   - OPENAI_API_KEY for OpenAI GPT models"
    echo "   - GOOGLE_API_KEY for Google Gemini models"
    echo "   - DEEPSEEK_API_KEY for DeepSeek models"
fi

# 检查 Ollama 是否可用
if command -v ollama &> /dev/null; then
    echo "✅ Ollama found"
    if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "⚠️  Ollama service is not running. Starting Ollama..."
        ollama serve &
        sleep 3
        
        # 检查是否有模型
        if ! ollama list | grep -q "llama"; then
            echo "📥 Downloading default model (llama3.2:latest)..."
            ollama pull llama3.2:latest
        fi
    fi
else
    echo "⚠️  Ollama not found. You can install it from https://ollama.ai"
    echo "   Or use cloud AI providers with API keys."
fi

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "⚠️  config.yaml not found. Using default configuration."
fi

# 构建应用
echo "🔨 Building KWiki..."
go build -o kwiki ./cmd/kwiki

# 启动应用
echo "🌟 Starting KWiki server..."
echo ""
echo "📖 Open http://localhost:8080 in your browser"
echo "🛑 Press Ctrl+C to stop the server"
echo ""

# 设置默认环境变量
export PORT=${PORT:-8080}
export HOST=${HOST:-localhost}
export OLLAMA_HOST=${OLLAMA_HOST:-http://localhost:11434}

# 启动服务器
./kwiki
