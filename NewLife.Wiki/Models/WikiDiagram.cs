namespace NewLife.Wiki.Models;

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