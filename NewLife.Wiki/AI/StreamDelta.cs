namespace NewLife.Wiki.AI;

/// <summary>流式片段</summary>
public class StreamDelta
{
    /// <summary>当前增量文本内容。</summary>
    public String Text { get; set; } = String.Empty;
    /// <summary>是否标记为完成（最后一片段）。</summary>
    public Boolean Completed { get; set; }
}
