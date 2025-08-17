namespace NewLife.Wiki.Models;

/// <summary>源代码或目录的基本信息</summary>
public class FileInfoModel
{
    /// <summary>相对路径（或逻辑路径）</summary>
    public String Path { get; set; } = String.Empty;

    /// <summary>文件名（不含路径）</summary>
    public String Name { get; set; } = String.Empty;

    /// <summary>扩展名（小写，含点）</summary>
    public String Extension { get; set; } = String.Empty;

    /// <summary>字节大小。目录为 0</summary>
    public Int64 Size { get; set; }

    /// <summary>推断的主要语言（如 CSharp / Go / Markdown）</summary>
    public String Language { get; set; } = String.Empty;

    /// <summary>文件内容。较大文件或未要求加载时为空</summary>
    public String? Content { get; set; }

    /// <summary>内容哈希（SHA1 或其他）</summary>
    public String Hash { get; set; } = String.Empty;

    /// <summary>最后修改时间（UTC）</summary>
    public DateTime ModifiedAt { get; set; } = DateTime.UtcNow;

    /// <summary>是否为目录</summary>
    public Boolean IsDirectory { get; set; }

    /// <summary>行数。目录为 0</summary>
    public Int32 LineCount { get; set; }

    /// <summary>简易复杂度指标（占位，后续可由解析器计算）</summary>
    public Int32 Complexity { get; set; }
}
