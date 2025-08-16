using System;
using System.Collections.Generic;

namespace NewLife.Wiki.Models;

/// <summary>生成的 Wiki 实体，包含页面、图表及生成元数据。</summary>
public class Wiki
{
    /// <summary>Wiki 标识</summary>
    public String Id { get; set; } = String.Empty;
    /// <summary>关联的仓库 Id</summary>
    public String RepositoryId { get; set; } = String.Empty;
    /// <summary>打包输出路径</summary>
    public String PackagePath { get; set; } = String.Empty;
    /// <summary>标题</summary>
    public String Title { get; set; } = String.Empty;
    /// <summary>描述</summary>
    public String Description { get; set; } = String.Empty;
    /// <summary>当前状态</summary>
    public WikiStatus Status { get; set; } = WikiStatus.Pending;
    /// <summary>进度（0-100）</summary>
    public Int32 Progress { get; set; }
    /// <summary>标签集合</summary>
    public List<String> Tags { get; set; } = new();
    /// <summary>创建时间 UTC</summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>更新时间 UTC</summary>
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>生成工具或用户</summary>
    public String GeneratedBy { get; set; } = String.Empty;
    /// <summary>所用模型名称</summary>
    public String Model { get; set; } = String.Empty;
    /// <summary>主语言（默认 zh）</summary>
    public String Language { get; set; } = "zh";
    /// <summary>启用的语言列表</summary>
    public List<String> Languages { get; set; } = new();
    /// <summary>页面集合</summary>
    public List<WikiPage> Pages { get; set; } = new();
    /// <summary>图表集合</summary>
    public List<WikiDiagram> Diagrams { get; set; } = new();
    /// <summary>生成设置</summary>
    public WikiSettings Settings { get; set; } = new();
    /// <summary>元数据统计</summary>
    public WikiMetadata Metadata { get; set; } = new();
}

/// <summary>Wiki 生成阶段状态</summary>
public enum WikiStatus
{
    /// <summary>等待开始</summary>
    Pending,
    /// <summary>仓库分析中</summary>
    Analyzing,
    /// <summary>内容生成中</summary>
    Generating,
    /// <summary>完成</summary>
    Completed,
    /// <summary>失败</summary>
    Failed
}

/// <summary>页面类型</summary>
public enum PageType
{
    /// <summary>概览</summary>
    Overview,
    /// <summary>架构</summary>
    Architecture,
    /// <summary>API 参考</summary>
    Api,
    /// <summary>模块</summary>
    Module,
    /// <summary>函数</summary>
    Function,
    /// <summary>类</summary>
    Class,
    /// <summary>教程</summary>
    Tutorial,
    /// <summary>参考/索引</summary>
    Reference,
    /// <summary>变更日志</summary>
    Changelog,
    /// <summary>指南/使用</summary>
    Guide
}

/// <summary>图表类型</summary>
public enum DiagramType
{
    /// <summary>流程图</summary>
    Flowchart,
    /// <summary>时序图</summary>
    Sequence,
    /// <summary>类图</summary>
    Class,
    /// <summary>实体关系图</summary>
    ER,
    /// <summary>甘特图</summary>
    Gantt,
    /// <summary>Git 分支</summary>
    GitGraph,
    /// <summary>架构图</summary>
    Architecture,
    /// <summary>数据流图</summary>
    DataFlow
}

/// <summary>Wiki 单个页面</summary>
public class WikiPage
{
    /// <summary>页面 Id</summary>
    public String Id { get; set; } = String.Empty;
    /// <summary>标题</summary>
    public String Title { get; set; } = String.Empty;
    /// <summary>Markdown 内容</summary>
    public String Content { get; set; } = String.Empty;
    /// <summary>页面类型</summary>
    public PageType Type { get; set; }
    /// <summary>排序序号</summary>
    public Int32 Order { get; set; }
    /// <summary>父页面 Id</summary>
    public String? ParentId { get; set; }
    /// <summary>子页面 Id 列表</summary>
    public List<String> Children { get; set; } = new();
    /// <summary>标签</summary>
    public List<String> Tags { get; set; } = new();
    /// <summary>创建时间</summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>更新时间</summary>
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>字数统计</summary>
    public Int32 WordCount { get; set; }
    /// <summary>预计阅读时间（分钟）</summary>
    public Int32 ReadingTime { get; set; }
}

/// <summary>图表定义</summary>
public class WikiDiagram
{
    /// <summary>图表 Id</summary>
    public String Id { get; set; } = String.Empty;
    /// <summary>标题</summary>
    public String Title { get; set; } = String.Empty;
    /// <summary>类型</summary>
    public DiagramType Type { get; set; }
    /// <summary>源码（Mermaid/PlantUML 等）</summary>
    public String Content { get; set; } = String.Empty;
    /// <summary>描述说明</summary>
    public String Description { get; set; } = String.Empty;
    /// <summary>关联页面 Id</summary>
    public String? PageId { get; set; }
    /// <summary>创建时间</summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>更新时间</summary>
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}

/// <summary>生成设置</summary>
public class WikiSettings
{
    /// <summary>AI 提供者标识</summary>
    public String AiProvider { get; set; } = String.Empty;
    /// <summary>模型名称</summary>
    public String Model { get; set; } = String.Empty;
    /// <summary>温度（0-1）</summary>
    public Single Temperature { get; set; } = 0.7f;
    /// <summary>最大生成 token 数</summary>
    public Int32 MaxTokens { get; set; } = 4000;
    /// <summary>是否生成图表</summary>
    public Boolean EnableDiagrams { get; set; } = true;
    /// <summary>是否启用 RAG（后续扩展）</summary>
    public Boolean EnableRag { get; set; } = true;
    /// <summary>主语言</summary>
    public String Language { get; set; } = "zh";
    /// <summary>主题</summary>
    public String Theme { get; set; } = "default";
    /// <summary>自定义提示词（键为用途，如 overview）</summary>
    public Dictionary<String,String> CustomPrompts { get; set; } = new();
    /// <summary>排除文件/目录模式</summary>
    public List<String> ExcludePatterns { get; set; } = new();
    /// <summary>包含文件/目录模式（为空表示全部）</summary>
    public List<String> IncludePatterns { get; set; } = new();
}

/// <summary>生成过程及结果统计</summary>
public class WikiMetadata
{
    /// <summary>生成耗时</summary>
    public TimeSpan GenerationTime { get; set; }
    /// <summary>使用 token 数</summary>
    public Int32 TokensUsed { get; set; }
    /// <summary>处理的文件数</summary>
    public Int32 FilesProcessed { get; set; }
    /// <summary>生成的页面数</summary>
    public Int32 PagesGenerated { get; set; }
    /// <summary>生成的图表数</summary>
    public Int32 DiagramsGenerated { get; set; }
    /// <summary>覆盖的语言列表</summary>
    public List<String> Languages { get; set; } = new();
    /// <summary>复杂度指标</summary>
    public String Complexity { get; set; } = String.Empty;
    /// <summary>质量评分（0~1 或百分比）</summary>
    public Double Quality { get; set; }
    /// <summary>标签合集</summary>
    public List<String> Tags { get; set; } = new();
    /// <summary>分类合集</summary>
    public List<String> Categories { get; set; } = new();
    /// <summary>统计字典，不同维度计数</summary>
    public Dictionary<String,Int32> Statistics { get; set; } = new();
    /// <summary>打包输出路径</summary>
    public String PackagePath { get; set; } = String.Empty;
    /// <summary>仓库地址</summary>
    public String RepositoryUrl { get; set; } = String.Empty;
}

/// <summary>Wiki 生成请求</summary>
public class GenerationRequest
{
    /// <summary>目标仓库地址</summary>
    public String RepositoryUrl { get; set; } = String.Empty;
    /// <summary>分支</summary>
    public String Branch { get; set; } = String.Empty;
    /// <summary>访问令牌（私有仓库可选）</summary>
    public String AccessToken { get; set; } = String.Empty;
    /// <summary>生成设置</summary>
    public WikiSettings Settings { get; set; } = new();
    /// <summary>自定义标题</summary>
    public String Title { get; set; } = String.Empty;
    /// <summary>自定义描述</summary>
    public String Description { get; set; } = String.Empty;
    /// <summary>额外自定义提示词（按顺序追加）</summary>
    public List<String> CustomPrompts { get; set; } = new();
    /// <summary>需生成的语言列表</summary>
    public List<String> Languages { get; set; } = new();
    /// <summary>主语言</summary>
    public String PrimaryLanguage { get; set; } = "zh";
    /// <summary>是否对全部检测到语言生成</summary>
    public Boolean GenerateAllLangs { get; set; }
}
