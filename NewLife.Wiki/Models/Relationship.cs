namespace NewLife.Wiki.Models;

/// <summary>实体间关系（调用/依赖/继承等）</summary>
public class Relationship
{
    /// <summary>源实体名称</summary>
    public String From { get; set; } = String.Empty;

    /// <summary>目标实体名称</summary>
    public String To { get; set; } = String.Empty;

    /// <summary>关系类型（call / import / inherit 等）</summary>
    public String Type { get; set; } = String.Empty;

    /// <summary>关系出现的文件</summary>
    public String File { get; set; } = String.Empty;

    /// <summary>出现的行号</summary>
    public Int32 Line { get; set; }
}
