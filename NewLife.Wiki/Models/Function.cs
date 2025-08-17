namespace NewLife.Wiki.Models;

/// <summary>函数/方法定义</summary>
public class Function
{
    /// <summary>函数名</summary>
    public String Name { get; set; } = String.Empty;

    /// <summary>所属模块</summary>
    public String Module { get; set; } = String.Empty;

    /// <summary>所在文件相对路径</summary>
    public String File { get; set; } = String.Empty;

    /// <summary>起始行（含）</summary>
    public Int32 StartLine { get; set; }

    /// <summary>结束行（含）</summary>
    public Int32 EndLine { get; set; }

    /// <summary>语言</summary>
    public String Language { get; set; } = String.Empty;

    /// <summary>签名（参数列表等）</summary>
    public String Signature { get; set; } = String.Empty;

    /// <summary>复杂度</summary>
    public Int32 Complexity { get; set; }

    /// <summary>是否公共可见</summary>
    public Boolean IsPublic { get; set; }
}
