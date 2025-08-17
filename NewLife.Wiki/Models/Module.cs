namespace NewLife.Wiki.Models;

/// <summary>逻辑模块（目录/包）</summary>
public class Module
{
    /// <summary>模块名</summary>
    public String Name { get; set; } = String.Empty;

    /// <summary>相对路径</summary>
    public String Path { get; set; } = String.Empty;

    /// <summary>主要语言</summary>
    public String Language { get; set; } = String.Empty;

    /// <summary>代码行数</summary>
    public Int32 LineCount { get; set; }
}
