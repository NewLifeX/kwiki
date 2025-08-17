namespace NewLife.Wiki.AI;

/// <summary>生成请求参数。用于向 AI 模型发送一次文本生成请求。</summary>
/// <remarks>
/// 可选字段 <see cref="System"/> 支持给模型一个系统级别角色，用于限定语气/身份。
/// <para>Extra 字典用于传递各模型特有参数（如 top_p、presence_penalty 等），保持核心结构精简。</para>
/// </remarks>
public class GenerationOptions
{
    /// <summary>模型名称（为空时使用提供者默认模型）</summary>
    public String? Model { get; set; }

    /// <summary>主提示词（用户输入）。不能为空。</summary>
    public String Prompt { get; set; } = String.Empty;

    /// <summary>系统指令 / 角色设定（可选）。</summary>
    public String? System { get; set; }

    /// <summary>采样温度，0~2。数值越大随机性越高。</summary>
    public Double Temperature { get; set; } = 0.7;

    /// <summary>回复最大 token 数。受模型上限限制。</summary>
    public Int32 MaxTokens { get; set; } = 4096;

    /// <summary>是否使用流式接口（onDelta）。</summary>
    public Boolean Stream { get; set; }

    /// <summary>模型特定的扩展参数键值集合。</summary>
    public IDictionary<String, Object?> Extra { get; set; } = new Dictionary<String, Object?>();
}
