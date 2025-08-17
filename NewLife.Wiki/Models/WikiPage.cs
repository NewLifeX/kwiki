namespace NewLife.Wiki.Models;

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
    public List<String> Children { get; set; } = [];

    /// <summary>标签</summary>
    public List<String> Tags { get; set; } = [];

    /// <summary>创建时间</summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>更新时间</summary>
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>字数统计</summary>
    public Int32 WordCount { get; set; }

    /// <summary>预计阅读时间（分钟）</summary>
    public Int32 ReadingTime { get; set; }
}