# KWiki (.NET) - AI 代码仓库 Wiki 生成器

<div align="center">

![KWiki Logo](assets/logo.svg)

**🚀 基于 .NET 的智能代码文档生成工具（multi‑TFM：net8.0 + net462）**

[![.NET](https://img.shields.io/badge/.NET-8.0%20%2B%204.6.2-512BD4?logo=dotnet)](https://dotnet.microsoft.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/stcn52/kwiki?style=social)](https://github.com/stcn52/kwiki)
[![GitHub Issues](https://img.shields.io/github/issues/stcn52/kwiki)](https://github.com/stcn52/kwiki/issues)
[![Docker Pulls](https://img.shields.io/docker/pulls/stcn52/kwiki)](https://hub.docker.com/r/stcn52/kwiki)

[🚀 快速开始](#-快速开始) •
[🧪 命令行使用](#-命令行使用) •
[⚙️ 配置说明](#️-配置说明) •
[� AI 提供者](#-ai-提供者) •
[♻️ 多目标框架](#️-多目标框架策略) •
[�🤝 贡献指南](#-贡献指南) •
[📞 支持](#-支持与反馈)

</div>

---

## 🌟 项目简介

**KWiki (.NET 版)** 为单一控制台项目，自动分析本地或已克隆的代码仓库，生成结构化 Markdown Wiki（多语言可选）。该版本移植自原 Go 实现，聚焦：跨平台兼容（net462 + net8.0）、最小依赖、可扩展 AI Provider 抽象。

### 🎯 核心价值

- **🤖 AI 驱动**：利用多种 AI 模型深度理解代码逻辑
- **📊 可视化**：自动生成架构图和流程图
- **🌍 多语言**：支持中英日韩等多种语言文档
- **🔍 智能搜索**：基于 RAG 的文档问答系统
- **⚡ 高效率**：几分钟内生成完整项目文档

## ✨ 核心特性

### 🤖 多 AI 提供商支持
- **DeepSeek**：成本低廉，中文支持优秀
- **OpenAI GPT**：高质量输出，支持最新模型
- **Google Gemini**：免费额度高，多模态支持
- **Ollama**：本地部署，数据安全可控

### 📊 智能代码分析
- 自动识别项目架构和技术栈
- 分析依赖关系和模块结构
- 提取关键函数和 API 接口
- 生成代码调用关系图

### 🌍 多语言文档生成
- 支持中文、英文、日文、韩文等
- 可同时生成多语言版本
- 智能翻译技术术语
- 保持技术文档的准确性

### 📈 可视化图表
- **架构图**：系统整体架构展示
- **流程图**：业务流程和数据流
- **依赖图**：模块依赖关系
- **API 图**：接口调用关系

### 🔍 智能搜索与问答
- 基于 RAG 的语义搜索
- 自然语言问答系统
- 代码片段快速定位
- 上下文相关的智能推荐

### 📱 现代化 Web 界面
- 响应式设计，支持移动端
- 实时进度显示和日志查看
- 直观的文档浏览体验
- 支持主题定制

## 🎯 适用场景

| 场景 | 描述 | 收益 |
|------|------|------|
| **开源项目** | 为 GitHub 项目自动生成文档 | 提升项目可读性，吸引更多贡献者 |
| **企业内部** | 为内部代码仓库生成标准化文档 | 降低维护成本，提高开发效率 |
| **API 文档** | 自动分析接口并生成 API 文档 | 减少手工维护，保证文档同步 |
| **代码审查** | 快速了解项目结构和核心逻辑 | 提高审查效率，降低理解成本 |
| **新人培训** | 为新团队成员提供项目概览 | 加速上手过程，减少培训时间 |

## 🚀 快速开始

### 📋 环境要求 (.NET 版)

| 项目 | 要求 |
|------|------|
| .NET SDK | 8.0 (构建) / 4.6.2 运行时向下兼容 |
| 操作系统 | Windows / Linux / macOS |
| 内存 | ≥ 4GB 建议 |
| 网络 | 访问所选 AI API（Ollama 可本地） |

### ⏬ 获取源码
```bash
git clone https://github.com/stcn52/kwiki.git
cd kwiki/NewLife.Wiki
```

### 🔨 构建
```bash
dotnet build -c Release
```

生成产物：`bin/Release/net8.0/` 与 `bin/Release/net462/`

### 🏃 运行（示例）
```bash
# 分析指定目录
dotnet run -- analyze --path ../some-repo --output ./_analysis.json

# 生成 Wiki（使用默认配置与 openai 或 ollama）
dotnet run -- generate --path ../some-repo --lang zh --out ./_wiki

# 启动最小 HTTP 服务 (仅 net8.0)
dotnet run -- serve --path ../some-repo
```

> 注意：HTTP 服务仅在 net8.0 目标下启用（Minimal API）。

---

## 🧪 命令行使用

当前提供以下子命令（Program.cs 内实现）：

| 命令 | 说明 | 常用参数 |
|------|------|----------|
| analyze | 扫描仓库，输出结构与基本指标 | `--path`, `--include`, `--exclude` |
| generate | 生成 Wiki Markdown | `--path`, `--lang`, `--out`, `--provider` |
| serve | 启动 HTTP 服务 (net8.0) | `--path`, `--port` (配置覆盖) |
| ai-test | 直接测试 AI Provider 输出 | `--provider`, `--model`, `--prompt` |

### 示例
```bash
dotnet run -- analyze --path ../repo --output repo-structure.json
dotnet run -- generate --path ../repo --lang zh --out ./wiki
dotnet run -- ai-test --provider openai --model gpt-4o-mini --prompt "Summarize repo"
```

## 📖 使用指南

### 🎬 生成第一个 Wiki

1. **输入仓库信息**
   - 在首页输入 GitHub/GitLab 仓库 URL
   - 如果是私有仓库，提供访问令牌

2. **选择 AI 配置**
   - 选择 AI 提供商（DeepSeek/OpenAI/Gemini/Ollama）
   - 选择合适的模型
   - 调整生成参数

3. **配置生成选项**
   - 选择文档语言（支持多语言同时生成）
   - 启用图表生成
   - 配置其他高级选项

4. **开始生成**
   - 点击"Generate Wiki"
   - 实时查看生成进度
   - 查看详细日志

### 🔍 浏览和搜索文档

#### 访问方式
- **包路径访问**：`http://localhost:8080/pkg/github.com/owner/repo`
- **Wiki ID 访问**：`http://localhost:8080/wiki/wiki-id`
- **特定页面**：`http://localhost:8080/wiki/wiki-id/page/page-id`

#### 搜索功能
- **关键词搜索**：在搜索框输入关键词
- **智能问答**：使用自然语言提问
- **代码搜索**：搜索特定函数或类
- **标签过滤**：按标签筛选内容

### 🔄 管理 Wiki

#### 重新生成
- 点击 Wiki 卡片上的"Rebuild"按钮
- 确认重新生成操作
- 系统将使用最新代码重新生成文档

#### 删除 Wiki
- 点击"Delete"按钮
- 确认删除操作
- 系统将删除所有相关文件

#### 导出文档
- 支持 Markdown、HTML、PDF 格式
- 可选择导出特定页面或整个 Wiki
- 支持批量导出

## ⚙️ 配置说明

### 配置文件 (`config.yaml`) 结构（.NET 版）

与 AppConfig 映射：
```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  dataDir: "./data"
  staticDir: "./web/static"
  templateDir: "./web/templates"

ai:
  defaultProvider: "ollama"
  providers:
    openai:
      apiKey: ${OPENAI_API_KEY}
      model: gpt-4o-mini
      temperature: 0.7
      maxTokens: 4000
    gemini:
      apiKey: ${GOOGLE_API_KEY}
      model: gemini-2.0-flash-exp
    ollama:
      model: llama3

repository:
  cloneDir: ./repos
  excludePatterns: ["node_modules", ".git", "vendor", "dist", "*.log"]
  includePatterns: ["*.cs", "*.go", "*.md", "*.json", "*.yml", "*.yaml"]
  maxFiles: 10000

generator:
  outputDir: ./output
  enableDiagrams: true
  enableRag: false
  chunkSize: 1000
  chunkOverlap: 200
  maxConcurrency: 5
```

### 环境变量覆盖
| 变量 | 作用 |
|------|------|
| OPENAI_API_KEY | 自动注入到 `ai.providers.openai.apiKey` |
| GOOGLE_API_KEY | 自动注入到 `ai.providers.gemini.apiKey` |
| DEEPSEEK_API_KEY | 注入 `ai.providers.deepseek.apiKey` |
| OLLAMA_BASE_URL | 覆盖本地 Ollama 基地址 (默认 http://localhost:11434) |

> 所有列出的 Provider 均已具备基本 HTTP/流式实现；请确保相应密钥或本地服务可用。

### 最小运行无需配置文件
未提供 `config.yaml` 时，程序会使用 `AppConfig.Default()` 内置默认值。

## 🧠 AI 提供者

| Provider | 状态 | 说明 |
|----------|------|------|
| OpenAI | 已实现 | /v1/chat/completions + 流式 (SSE) |
| Gemini | 已实现 | generateContent / streamGenerateContent |
| DeepSeek | 已实现 | 兼容 OpenAI Chat Completions + 流式 |
| Ollama | 已实现 | 本地 /api/generate 与 /api/stream (JSON lines) |

统一抽象 `IAIProvider`：`GenerateAsync` 与 `StreamAsync`。注册通过 `AIProviderManager`。

> 可扩展：新增提供者时仅需实现接口并在启动处注册。

## ♻️ 多目标框架策略

项目 `NewLife.Wiki.csproj` 目标：`net8.0;net462`。

设计要点：
1. 低版本不使用 `default interface methods` 等新语法。
2. 条件编译启用 Minimal API 服务器：`#if NET8_0_OR_GREATER`。
3. 仅在高版本引用 ASP.NET Core `FrameworkReference`。
4. 序列化：高版本用 `System.Text.Json`，低版本使用手写最小 JSON（OpenAI 流式）。

后续可按需加入 `net6.0/net7.0/net9.0`。

### 配置文件

编辑 `config.yaml` 自定义配置：

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  data_dir: "./data"
  static_dir: "./web/static"
  template_dir: "./web/templates"

ai:
  default_provider: "deepseek"
  providers:
    deepseek:
      api_key: "${DEEPSEEK_API_KEY}"
      model: "deepseek-chat"
      temperature: 0.7
      max_tokens: 8000

    openai:
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o-mini"
      temperature: 0.7
      max_tokens: 4000

analysis:
  max_file_size: 1048576  # 1MB
  include_patterns:
    - "*.go"
    - "*.py"
    - "*.js"
    - "*.ts"
    - "*.java"
    - "*.cpp"
    - "*.c"
    - "*.h"
    - "*.md"
    - "*.yml"
    - "*.yaml"
    - "*.json"
  exclude_patterns:
    - "vendor/*"
    - "node_modules/*"
    - ".git/*"
    - "*.min.js"
    - "*.min.css"
```

## 🏗️ 项目结构（.NET 目录）

```
NewLife.Wiki/
├── NewLife.Wiki.csproj        # 单项目 multi-TFM
├── Program.cs                 # 入口 & 命令路由
├── AI/                        # IAIProvider + 各 Provider
├── Analyzer/                  # RepositoryAnalyzer
├── Generator/                 # WikiGenerator
├── Config/                    # AppConfig 及子配置
├── Models/                    # 数据模型 (Repository/Wiki/...)
├── Server/                    # Minimal API（仅 net8.0）
├── templates/                 # 提示词模板
├── web/                       # HTML 模板
└── ...
```

## 🔧 开发指南

### 环境要求

- Go 1.24+
- Node.js 16+ (用于前端资源处理)
- Git

### 本地开发 (.NET)
```bash
dotnet restore
dotnet build
dotnet run -- analyze --path ../repo
```

### 添加新的 AI Provider (.NET)
1. 新建类实现 `IAIProvider`。
2. 在启动或命令执行前调用 `AIProviderManager.Register(new XxxProvider(...))`。
3. 可支持流式：实现 `StreamAsync` 并逐步回调。

### 扩展代码分析

1. 在 `internal/analyzer/analyzer.go` 中添加新的语言支持
2. 扩展 `detectLanguage` 和相关解析函数
3. 更新配置文件中的 `include_patterns`

## 📊 后续优化方向

- 并行语法级解析，提取函数/类复杂度
- AI 概览摘要多模型回退
- 真正的 RAG：向量化 + 相似度检索
- 流式 SSE 输出对接前端
- 更多 Provider：Azure OpenAI / Qwen / Claude

## 🔒 安全考虑

- **API 密钥保护**：环境变量存储，不记录日志
- **输入验证**：严格验证用户输入
- **文件访问控制**：限制文件访问范围
- **速率限制**：防止 API 滥用

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 贡献方式

1. **报告问题**：在 [Issues](https://github.com/stcn52/kwiki/issues) 中报告 bug
2. **功能建议**：提出新功能想法
3. **代码贡献**：提交 Pull Request
4. **文档改进**：完善文档和示例

### 开发流程

1. Fork 项目
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 创建 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 添加必要的注释和文档
- 编写单元测试
- 确保所有测试通过

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

参考与致谢：OpenAI / DeepSeek / Gemini / Ollama 等模型与生态工具。

## 📞 支持与反馈

如果你遇到问题或有建议，请：

1. 查看 [FAQ](https://github.com/stcn52/kwiki/wiki/FAQ)
2. 搜索现有 [Issues](https://github.com/stcn52/kwiki/issues)
3. 创建新的 Issue
4. 加入我们的讨论群

## 🗺️ 路线图

### Roadmap (.NET 版阶段性)
| 版本 | 目标 |
|------|------|
| v0.1 | 移植基础：分析 + 生成 + OpenAI Provider ✅ |
| v0.2 | 其余 Provider 实现 (Gemini/DeepSeek/Ollama HTTP) |
| v0.3 | 语法级解析 + 图表生成 | 
| v0.4 | RAG & 多语言并行生成 |
| v1.0 | 插件/扩展、打包发布 |

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给我们一个 Star！**

[🏠 首页](https://github.com/stcn52/kwiki) •
[📖 文档](https://github.com/stcn52/kwiki/wiki) •
[🐛 报告问题](https://github.com/stcn52/kwiki/issues) •
[💡 功能建议](https://github.com/stcn52/kwiki/discussions)

</div>