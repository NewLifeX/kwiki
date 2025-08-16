using NewLife.Configuration;

namespace NewLife.Wiki.Config;

/// <summary>应用程序总体配置</summary>
[Config("Wiki")]
public class AppConfig : Config<AppConfig>
{
    /// <summary>HTTP 服务配置</summary>
    public ServerConfig Server { get; set; } = new();
    /// <summary>AI 提供方与默认模型配置</summary>
    public AIConfig AI { get; set; } = new();
    /// <summary>仓库扫描与过滤配置</summary>
    public RepositoryConfig Repository { get; set; } = new();
    /// <summary>文档生成行为配置</summary>
    public GeneratorConfig Generator { get; set; } = new();

    ///// <summary>从指定 yaml 路径加载配置</summary>
    ///// <param name="path">配置文件完整路径</param>
    ///// <returns>配置实例</returns>
    ///// <exception cref="FileNotFoundException">文件不存在时抛出</exception>
    //public static AppConfig Load(String path)
    //{
    //    if (!File.Exists(path)) throw new FileNotFoundException("配置文件不存在", path);
    //    var yaml = File.ReadAllText(path);
    //    var deserializer = new DeserializerBuilder().WithNamingConvention(CamelCaseNamingConvention.Instance).Build();
    //    var cfg = deserializer.Deserialize<AppConfig>(yaml) ?? new AppConfig();
    //    cfg.ApplyEnvironment();
    //    return cfg;
    //}

    ///// <summary>创建默认配置</summary>
    //public static AppConfig Default()
    //{
    //    var cfg = new AppConfig();
    //    cfg.AI.DefaultProvider = "ollama";
    //    cfg.Repository.ExcludePatterns = ["node_modules", ".git", "vendor", "target", "build", "dist", "*.log", "*.tmp"];
    //    cfg.Repository.IncludePatterns = ["*.go", "*.py", "*.js", "*.ts", "*.java", "*.cpp", "*.c", "*.h", "*.rs", "*.rb", "*.php", "*.cs", "*.md", "*.txt", "*.yaml", "*.yml", "*.json", "*.toml"];
    //    return cfg;
    //}

    protected override void OnLoaded()
    {
        if (AI.DefaultProvider.IsNullOrEmpty()) AI.DefaultProvider = "ollama";

        if (Repository.ExcludePatterns.Length == 0)
            Repository.ExcludePatterns = ["node_modules", ".git", "vendor", "target", "build", "dist", "*.log", "*.tmp"];
        if (Repository.IncludePatterns.Length == 0)
            Repository.IncludePatterns = ["*.go", "*.py", "*.js", "*.ts", "*.java", "*.cpp", "*.c", "*.h", "*.rs", "*.rb", "*.php", "*.cs", "*.md", "*.txt", "*.yaml", "*.yml", "*.json", "*.toml"];

        ApplyEnvironment();

        base.OnLoaded();
    }

    /// <summary>应用环境变量覆盖（如 OpenAI / Gemini key）</summary>
    public void ApplyEnvironment()
    {
        var openai = Environment.GetEnvironmentVariable("OPENAI_API_KEY");
        if (!String.IsNullOrEmpty(openai))
        {
            AI.Providers["openai"] = new AIProvider
            {
                ApiKey = openai,
                Model = "gpt-4o-mini",
                Temperature = 0.7f,
                MaxTokens = 4000
            };
        }
        var gemini = Environment.GetEnvironmentVariable("GOOGLE_API_KEY");
        if (!String.IsNullOrEmpty(gemini))
        {
            AI.Providers["gemini"] = new AIProvider
            {
                ApiKey = gemini,
                Model = "gemini-2.0-flash-exp",
                Temperature = 0.7f,
                MaxTokens = 4000
            };
        }
    }
}

/// <summary>HTTP 服务设置</summary>
public class ServerConfig
{
    /// <summary>监听端口（字符串形式便于环境覆盖）</summary>
    public String Port { get; set; } = "8080";
    /// <summary>绑定主机名/地址</summary>
    public String Host { get; set; } = "localhost";
    /// <summary>静态资源目录</summary>
    public String StaticDir { get; set; } = "web/static";
    /// <summary>HTML 模板目录</summary>
    public String TemplateDir { get; set; } = "web/templates";
    /// <summary>数据存储根目录</summary>
    public String DataDir { get; set; } = "./data";
    /// <summary>是否启用 CORS</summary>
    public Boolean EnableCors { get; set; } = true;
    /// <summary>上传或处理的最大文件尺寸（字节）</summary>
    public Int64 MaxFileSize { get; set; } = 100 * 1024 * 1024;
}

/// <summary>AI 提供者集合配置</summary>
public class AIConfig
{
    /// <summary>默认使用的 AI 提供方 key</summary>
    public String DefaultProvider { get; set; } = "ollama";
    /// <summary>所有可用 AI 提供方配置字典</summary>
    public Dictionary<String, AIProvider> Providers { get; set; } = new(StringComparer.OrdinalIgnoreCase);
}

/// <summary>单个 AI 提供者配置</summary>
public class AIProvider
{
    /// <summary>访问密钥（可来自环境变量）</summary>
    public String? ApiKey { get; set; }
    /// <summary>自定义基地址（为空则用默认官方地址）</summary>
    public String? BaseUrl { get; set; }
    /// <summary>模型名称</summary>
    public String Model { get; set; } = String.Empty;
    /// <summary>采样温度 0~1</summary>
    public Single Temperature { get; set; } = 0.7f;
    /// <summary>最大 tokens 限制</summary>
    public Int32 MaxTokens { get; set; } = 4000;
    /// <summary>扩展字段（特定提供方自定义参数）</summary>
    public Dictionary<String, String>? Extra { get; set; }
}

/// <summary>仓库扫描配置</summary>
public class RepositoryConfig
{
    /// <summary>仓库克隆/缓存目录</summary>
    public String CloneDir { get; set; } = "./repos";
    /// <summary>允许的最大仓库体积（字节，0 表示不限制）</summary>
    public Int64 MaxRepoSize { get; set; }
    /// <summary>排除匹配模式（glob 或通配）</summary>
    public String[] ExcludePatterns { get; set; } = [];
    /// <summary>包含匹配模式（为空表示不过滤）</summary>
    public String[] IncludePatterns { get; set; } = [];
    /// <summary>最大处理文件数量</summary>
    public Int32 MaxFiles { get; set; } = 10000;
}

/// <summary>文档生成配置</summary>
public class GeneratorConfig
{
    /// <summary>输出目录</summary>
    public String OutputDir { get; set; } = "./output";
    /// <summary>是否生成图表（Mermaid 等）</summary>
    public Boolean EnableDiagrams { get; set; } = true;
    /// <summary>是否启用 RAG（检索增强生成）</summary>
    public Boolean EnableRag { get; set; } = true;
    /// <summary>切分块大小（字符）</summary>
    public Int32 ChunkSize { get; set; } = 1000;
    /// <summary>相邻块重叠字符数</summary>
    public Int32 ChunkOverlap { get; set; } = 200;
    /// <summary>最大并发工作数</summary>
    public Int32 MaxConcurrency { get; set; } = 5;
}
