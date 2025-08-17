namespace NewLife.Wiki.Models;

/// <summary>类/结构定义</summary>
public class Class
{
    /// <summary>类名</summary>
    public String Name { get; set; } = String.Empty;

    /// <summary>所属模块</summary>
    public String Module { get; set; } = String.Empty;

    /// <summary>所在文件</summary>
    public String File { get; set; } = String.Empty;

    /// <summary>起始行</summary>
    public Int32 StartLine { get; set; }

    /// <summary>结束行</summary>
    public Int32 EndLine { get; set; }

    /// <summary>语言</summary>
    public String Language { get; set; } = String.Empty;

    /// <summary>是否公共可见</summary>
    public Boolean IsPublic { get; set; }
}
