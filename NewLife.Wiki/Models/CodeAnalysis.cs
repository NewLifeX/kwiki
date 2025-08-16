using System;
using System.Collections.Generic;

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

/// <summary>仓库结构化分析结果</summary>
public class CodeStructure
{
    /// <summary>仓库标识。一般为目录名或外部传入 Id</summary>
    public String RepositoryId { get; set; } = String.Empty;
    /// <summary>所有文件及目录信息</summary>
    public List<FileInfoModel> Files { get; set; } = new();
    /// <summary>依赖列表（包、模块等）</summary>
    public List<Dependency> Dependencies { get; set; } = new();
    /// <summary>模块列表（按目录或语言分组）</summary>
    public List<Module> Modules { get; set; } = new();
    /// <summary>函数/方法集合（后续语法解析产生）</summary>
    public List<Function> Functions { get; set; } = new();
    /// <summary>类/类型集合</summary>
    public List<Class> Classes { get; set; } = new();
    /// <summary>实体关系（调用、依赖、继承等）</summary>
    public List<Relationship> Relationships { get; set; } = new();
    /// <summary>总体统计指标</summary>
    public CodeMetrics Metrics { get; set; } = new();
}

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

/// <summary>聚合统计指标</summary>
public class CodeMetrics
{
    /// <summary>总代码行（含空行/注释）</summary>
    public Int32 TotalLines { get; set; }
    /// <summary>代码有效行数</summary>
    public Int32 CodeLines { get; set; }
    /// <summary>注释行</summary>
    public Int32 CommentLines { get; set; }
    /// <summary>文件总数</summary>
    public Int32 TotalFiles { get; set; }
    /// <summary>函数总数</summary>
    public Int32 TotalFunctions { get; set; }
    /// <summary>类总数</summary>
    public Int32 TotalClasses { get; set; }
    /// <summary>平均复杂度</summary>
    public Double AverageComplexity { get; set; }
    /// <summary>最大复杂度</summary>
    public Int32 MaxComplexity { get; set; }
}
