namespace NewLife.Wiki.AI;

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
