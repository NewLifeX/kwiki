namespace NewLife.Wiki.Models;

/// <summary>聚合统计指标</summary>
public class CodeMetrics
{
    /// <summary>总代码行（含空行/注释）</summary>
    public Int32 TotalLines { get; set; }

    /// <summary>代码有效行数</summary>
    public Int32 CodeLines { get; set; }

    /// <summary>注释行</summary>
    public Int32 CommentLines { get; set; }

    /// <summary>文件总数</summary>
    public Int32 TotalFiles { get; set; }

    /// <summary>函数总数</summary>
    public Int32 TotalFunctions { get; set; }

    /// <summary>类总数</summary>
    public Int32 TotalClasses { get; set; }

    /// <summary>平均复杂度</summary>
    public Double AverageComplexity { get; set; }

    /// <summary>最大复杂度</summary>
    public Int32 MaxComplexity { get; set; }
}
