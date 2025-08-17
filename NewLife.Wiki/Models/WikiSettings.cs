namespace NewLife.Wiki.Models;

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
    public Dictionary<String, String> CustomPrompts { get; set; } = [];

    /// <summary>排除文件/目录模式</summary>
    public List<String> ExcludePatterns { get; set; } = [];

    /// <summary>包含文件/目录模式（为空表示全部）</summary>
    public List<String> IncludePatterns { get; set; } = [];
}