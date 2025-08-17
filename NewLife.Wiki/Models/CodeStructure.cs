namespace NewLife.Wiki.Models;

/// <summary>仓库结构化分析结果</summary>
public class CodeStructure
{
    /// <summary>仓库标识。一般为目录名或外部传入 Id</summary>
    public String RepositoryId { get; set; } = String.Empty;

    /// <summary>所有文件及目录信息</summary>
    public List<FileInfoModel> Files { get; set; } = [];

    /// <summary>依赖列表（包、模块等）</summary>
    public List<Dependency> Dependencies { get; set; } = [];

    /// <summary>模块列表（按目录或语言分组）</summary>
    public List<Module> Modules { get; set; } = [];

    /// <summary>函数/方法集合（后续语法解析产生）</summary>
    public List<Function> Functions { get; set; } = [];

    /// <summary>类/类型集合</summary>
    public List<Class> Classes { get; set; } = [];

    /// <summary>实体关系（调用、依赖、继承等）</summary>
    public List<Relationship> Relationships { get; set; } = [];

    /// <summary>总体统计指标</summary>
    public CodeMetrics Metrics { get; set; } = new();
}
