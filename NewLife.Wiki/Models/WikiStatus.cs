namespace NewLife.Wiki.Models;

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