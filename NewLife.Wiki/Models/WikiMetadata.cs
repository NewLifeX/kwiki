namespace NewLife.Wiki.Models;

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
    public List<String> Languages { get; set; } = [];

    /// <summary>复杂度指标</summary>
    public String Complexity { get; set; } = String.Empty;

    /// <summary>质量评分（0~1 或百分比）</summary>
    public Double Quality { get; set; }

    /// <summary>标签合集</summary>
    public List<String> Tags { get; set; } = [];

    /// <summary>分类合集</summary>
    public List<String> Categories { get; set; } = [];

    /// <summary>统计字典，不同维度计数</summary>
    public Dictionary<String, Int32> Statistics { get; set; } = [];

    /// <summary>打包输出路径</summary>
    public String PackagePath { get; set; } = String.Empty;

    /// <summary>仓库地址</summary>
    public String RepositoryUrl { get; set; } = String.Empty;
}