# KWiki 模板文档生成器

## 概述

KWiki 模板文档生成器是一个完全基于模板驱动的文档生成系统。它通过扫描模板目录中的文件，自动生成对应的wiki文档页面。

## 核心特性

- **完全模板驱动**: 所有文档内容和结构都由模板文件决定
- **多语言支持**: 支持多种语言的模板和文档生成
- **零硬编码**: 无需修改代码即可添加新的文档类型
- **AI集成**: 使用AI模型根据模板生成高质量文档内容
- **元数据驱动**: 通过YAML前置元数据控制页面属性

## 工作原理

### 1. 模板发现
系统自动扫描 `templates/prompts/` 目录下的所有语言子目录，发现可用的模板文件。

### 2. 元数据解析
每个模板文件包含YAML前置元数据，定义页面的基本属性：

```yaml
---
title: "系统概览"
type: "overview"
order: 1
description: "KWiki模板系统概览文档"
variables: ["ProjectName", "Description", "Statistics"]
---
```

### 3. 模板渲染
使用Go的 `text/template` 引擎将项目数据填入模板，生成AI提示词。

### 4. AI生成
将渲染后的提示词发送给AI模型，生成最终的文档内容。

### 5. 页面创建
根据模板元数据创建WikiPage对象，包含标题、类型、顺序等信息。

## 使用方法

### API调用

发送POST请求到 `/api/wiki/generate`：

```json
{
  "repository_url": "template-docs",
  "title": "KWiki Template System Documentation",
  "description": "Documentation for KWiki template system",
  "primary_language": "zh",
  "languages": ["zh", "en"],
  "settings": {
    "ai_provider": "openai",
    "model": "gpt-4"
  }
}
```

### Web界面

1. 访问KWiki主页
2. 在仓库URL字段输入 `template-docs`
3. 选择目标语言
4. 点击"生成Wiki"按钮

## 模板结构

### 目录结构
```
templates/prompts/
├── zh/                  # 中文模板
│   ├── readme.md
│   ├── getting-started.md
│   └── architecture.md
├── en/                  # 英文模板
│   ├── readme.md
│   ├── getting-started.md
│   └── architecture.md
└── config/
    └── generator.yaml   # 生成器配置
```

### 模板格式

每个模板文件包含两部分：

1. **YAML前置元数据**（必需）
2. **模板内容**（AI提示词）

```markdown
---
title: "README"
type: "overview"
order: 1
description: "项目README文档生成"
variables: ["ProjectName", "Description", "PrimaryLanguage"]
---

# README文档生成提示词

为以下项目生成README文档：

**项目信息：**
- 项目名称: {{.ProjectName}}
- 描述: {{.Description}}
- 主要语言: {{.PrimaryLanguage}}

**要求：**
生成包含以下部分的README文档：
1. 项目概述
2. 主要特性
3. 安装说明
4. 使用方法
5. 贡献指南

使用Markdown格式，语言简洁明了。
```

## 配置选项

### 生成器配置 (config/generator.yaml)

```yaml
# 阅读速度配置（每分钟阅读单词数）
reading_speed: 200

# 模板目录路径
template_dir: "templates/prompts"
```

### 模板元数据字段

- `title`: 页面标题
- `type`: 页面类型 (overview, guide, reference, architecture)
- `order`: 页面排序（数字越小越靠前）
- `description`: 页面描述
- `variables`: 模板中使用的变量列表

## 扩展指南

### 添加新的文档类型

1. 在对应语言目录下创建新的模板文件
2. 添加YAML前置元数据
3. 编写模板内容（AI提示词）
4. 无需修改任何代码

### 添加新语言支持

1. 在 `templates/prompts/` 下创建新的语言目录
2. 复制现有模板并翻译内容
3. 系统会自动发现并支持新语言

### 自定义页面类型

在模板元数据中使用 `type` 字段定义页面类型：

- `overview`: 概览页面
- `guide`: 指南页面  
- `reference`: 参考页面
- `architecture`: 架构页面
- 或任何自定义类型

## 测试

运行测试：

```bash
go test ./internal/generator/...
```

测试覆盖：
- 模板发现和解析
- 元数据验证
- 文档生成流程
- 多语言支持
- 错误处理

## 故障排除

### 常见问题

1. **模板未被发现**
   - 检查模板文件路径是否正确
   - 确认YAML前置元数据格式正确

2. **生成失败**
   - 检查AI提供商配置
   - 验证模板语法
   - 查看生成日志

3. **页面顺序错误**
   - 检查模板元数据中的 `order` 字段
   - 确保数值设置正确

### 调试模式

启用详细日志：

```bash
export LOG_LEVEL=debug
./kwiki
```

## 性能优化

- 模板缓存：已解析的模板会被缓存
- 并发生成：支持多页面并发生成
- 增量更新：只重新生成变更的页面

## 安全考虑

- 模板内容经过安全验证
- AI提示词过滤敏感信息
- 生成内容自动清理

## 贡献

欢迎贡献新的模板和功能改进！请参考项目的贡献指南。
