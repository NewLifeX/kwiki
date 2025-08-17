namespace NewLife.Wiki.AI;

/// <summary>token用量</summary>
public class TokenUsage
{
    /// <summary>提示词部分消耗的 token 数。</summary>
    public Int32 PromptTokens { get; set; }
    /// <summary>模型补全部分消耗的 token 数。</summary>
    public Int32 CompletionTokens { get; set; }
    /// <summary>总 token 数（Prompt + Completion）。</summary>
    public Int32 TotalTokens => PromptTokens + CompletionTokens;
}
