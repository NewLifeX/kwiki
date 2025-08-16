using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

namespace NewLife.Wiki.AI;

/// <summary>生成请求参数</summary>
/// <summary>生成请求参数。用于向 AI 模型发送一次文本生成请求。</summary>
/// <remarks>
/// 可选字段 <see cref="System"/> 支持给模型一个系统级别角色，用于限定语气/身份。
/// <para>Extra 字典用于传递各模型特有参数（如 top_p、presence_penalty 等），保持核心结构精简。</para>
/// </remarks>
public class GenerationOptions
{
    /// <summary>模型名称</summary>
    /// <summary>模型名称（为空时使用提供者默认模型）</summary>
    public String? Model { get; set; }

    /// <summary>提示词</summary>
    /// <summary>主提示词（用户输入）。不能为空。</summary>
    public String Prompt { get; set; } = String.Empty;

    /// <summary>系统角色（可选）</summary>
    /// <summary>系统指令 / 角色设定（可选）。</summary>
    public String? System { get; set; }

    /// <summary>温度</summary>
    /// <summary>采样温度，0~2。数值越大随机性越高。</summary>
    public Double Temperature { get; set; } = 0.7;

    /// <summary>最大token</summary>
    /// <summary>回复最大 token 数。受模型上限限制。</summary>
    public Int32 MaxTokens { get; set; } = 4096;

    /// <summary>是否流式</summary>
    /// <summary>是否使用流式接口（onDelta）。</summary>
    public Boolean Stream { get; set; }

    /// <summary>附加参数</summary>
    /// <summary>模型特定的扩展参数键值集合。</summary>
    public IDictionary<String, Object?> Extra { get; set; } = new Dictionary<String, Object?>();
}

/// <summary>生成结果</summary>
public class GenerationResult
{
    /// <summary>完整文本</summary>
    /// <summary>模型返回的完整文本（流式时为合并结果）。</summary>
    public String Text { get; set; } = String.Empty;

    /// <summary>使用的token统计</summary>
    /// <summary>Token 使用统计信息（若提供者返回）。</summary>
    public TokenUsage? Usage { get; set; }

    /// <summary>原始响应（可选）</summary>
    /// <summary>底层提供者的原始响应对象，调试/扩展用。</summary>
    public Object? Raw { get; set; }
}

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

/// <summary>流式片段</summary>
public class StreamDelta
{
    /// <summary>当前增量文本内容。</summary>
    public String Text { get; set; } = String.Empty;
    /// <summary>是否标记为完成（最后一片段）。</summary>
    public Boolean Completed { get; set; }
}

/// <summary>AI提供者接口</summary>
public interface IAIProvider
{
    /// <summary>名称</summary>
    String Name { get; }

    /// <summary>同步/一次性生成</summary>
    Task<GenerationResult> GenerateAsync(GenerationOptions options, CancellationToken cancellationToken = default);

    /// <summary>流式生成。回调返回false可中断。</summary>
    Task<GenerationResult> StreamAsync(GenerationOptions options, Func<StreamDelta, Task<Boolean>> onDelta, CancellationToken cancellationToken = default);
}

/// <summary>AI提供者管理器</summary>
public class AIProviderManager
{
    private readonly Dictionary<String, IAIProvider> _providers = new(StringComparer.OrdinalIgnoreCase);

    /// <summary>注册提供者</summary>
    public void Register(IAIProvider provider)
    {
        if (provider == null) throw new ArgumentNullException(nameof(provider));
        _providers[provider.Name] = provider;
    }

    /// <summary>获取指定名称的提供者</summary>
    public IAIProvider Get(String name)
    {
        if (String.IsNullOrEmpty(name)) throw new ArgumentNullException(nameof(name));
        if (_providers.TryGetValue(name, out var p)) return p;
        throw new InvalidOperationException("Provider not found: " + name);
    }

    /// <summary>尝试获取</summary>
    public Boolean TryGet(String name, out IAIProvider provider) => _providers.TryGetValue(name, out provider!);

    /// <summary>全部名称</summary>
    public ICollection<String> Names => _providers.Keys;
}
