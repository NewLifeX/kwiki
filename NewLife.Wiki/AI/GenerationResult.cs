namespace NewLife.Wiki.AI;

/// <summary>生成结果</summary>
public class GenerationResult
{
    /// <summary>模型返回的完整文本（流式时为合并结果）。</summary>
    public String Text { get; set; } = String.Empty;

    /// <summary>Token 使用统计信息（若提供者返回）。</summary>
    public TokenUsage? Usage { get; set; }

    /// <summary>底层提供者的原始响应对象，调试/扩展用。</summary>
    public Object? Raw { get; set; }
}
