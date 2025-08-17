namespace NewLife.Wiki.AI;

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
