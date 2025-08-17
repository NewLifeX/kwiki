namespace NewLife.Wiki.Models;

/// <summary>图表类型</summary>
public enum DiagramType
{
    /// <summary>流程图</summary>
    Flowchart,
    /// <summary>时序图</summary>
    Sequence,
    /// <summary>类图</summary>
    Class,
    /// <summary>实体关系图</summary>
    ER,
    /// <summary>甘特图</summary>
    Gantt,
    /// <summary>Git 分支</summary>
    GitGraph,
    /// <summary>架构图</summary>
    Architecture,
    /// <summary>数据流图</summary>
    DataFlow
}