namespace NewLife.Wiki.Models;

/// <summary>外部依赖（包/模块）信息</summary>
public class Dependency
{
    /// <summary>名称</summary>
    public String Name { get; set; } = String.Empty;

    /// <summary>版本号</summary>
    public String Version { get; set; } = String.Empty;

    /// <summary>类型（NuGet / Npm / GoModule 等）</summary>
    public String Type { get; set; } = String.Empty;

    /// <summary>来源（仓库地址或文件）</summary>
    public String Source { get; set; } = String.Empty;
}
