using NewLife.Log;
using NewLife.Wiki;
using NewLife.Wiki.AI;
using NewLife.Wiki.Config;
using NewLife.Wiki.Generator;

XTrace.UseConsole();

if (args.Length == 0 || args[0].Equals("help", StringComparison.OrdinalIgnoreCase)) return ShowHelp();

var cmd = args.Length > 0 ? args[0].ToLowerInvariant() : "";
try
{
    switch (cmd)
    {
        case "analyze":
            await Analyze(args.Skip(1).ToArray());
            break;
        case "generate":
            Generate(args.Skip(1).ToArray());
            break;
        case "serve":
            await Serve(args.Skip(1).ToArray());
            break;
        case "ai-test":
            await AiTest(args.Skip(1).ToArray());
            break;
        default: return ShowHelp();
    }
    return 0;
}
catch (Exception ex)
{
    XTrace.WriteException(ex);
    Console.Error.WriteLine("错误: " + ex.Message);
    return -1;
}

async Task Analyze(String[] args)
{
    if (args.Length == 0)
    {
        Console.WriteLine("用法: analyze <path>");
        return;
    }

    var path = args[0];
    XTrace.WriteLine("Analyze {0}", path);
    var analyzer = new NewLife.Wiki.Analyzer.RepositoryAnalyzer { LoadContent = true };
    var cs = await analyzer.AnalyzeAsync(path, CancellationToken.None);
    XTrace.WriteLine("文件数={0} 行数={1}", cs.Files.Count, cs.Metrics.TotalLines);
    Console.WriteLine($"分析完成: Files={cs.Files.Count}, Lines={cs.Metrics.TotalLines}");
}

void Generate(String[] args)
{
    if (args.Length == 0)
    {
        Console.WriteLine("用法: generate <repoPath> [--out=outputDir] [--lang=zh|en] [--provider=name]");
        return;
    }

    var repoPath = args[0];
    var outputDir = args.FirstOrDefault(a => a.StartsWith("--out="))?.Substring("--out=".Length) ?? "_wiki";
    var language = args.FirstOrDefault(a => a.StartsWith("--lang="))?.Substring("--lang=".Length) ?? "zh";
    var providerName = args.FirstOrDefault(a => a.StartsWith("--provider="))?.Substring("--provider=".Length);
    var cfgPath = args.FirstOrDefault(a => a.StartsWith("--config="))?.Substring("--config=".Length) ?? "config.yaml";

    //var config = AppConfig.Load(cfgPath) ?? AppConfig.Default();
    var config = AppConfig.Current;

    IAIProvider? ai = null;
    if (!String.IsNullOrEmpty(providerName))
    {
        if (config.AI.Providers.TryGetValue(providerName, out var p))
        {
            switch (providerName.ToLowerInvariant())
            {
                case "openai":
                    {
                        var key = p.ApiKey ?? Environment.GetEnvironmentVariable("OPENAI_API_KEY");
                        if (!String.IsNullOrEmpty(key))
                            ai = new OpenAIProvider(providerName, key, p.BaseUrl, p.Model);
                        else
                            XTrace.WriteLine("缺少 OPENAI_API_KEY");
                        break;
                    }
                case "gemini":
                    {
                        var key = p.ApiKey ?? Environment.GetEnvironmentVariable("GOOGLE_API_KEY");
                        if (!String.IsNullOrEmpty(key))
                            ai = new GeminiProvider(providerName, key, p.Model ?? "gemini-2.0-flash-exp", p.BaseUrl);
                        else
                            XTrace.WriteLine("缺少 GOOGLE_API_KEY");
                        break;
                    }
                case "deepseek":
                    {
                        var key = p.ApiKey ?? Environment.GetEnvironmentVariable("DEEPSEEK_API_KEY");
                        if (!String.IsNullOrEmpty(key))
                            ai = new DeepSeekProvider(key, p.Model ?? "deepseek-chat", p.BaseUrl);
                        else
                            XTrace.WriteLine("缺少 DEEPSEEK_API_KEY");
                        break;
                    }
                case "ollama":
                    ai = new OllamaProvider(p.Model ?? "llama3", p.BaseUrl);
                    break;
            }
        }
    }

    var gen = new WikiGenerator { AI = ai };
    var files = gen.Generate(repoPath, outputDir, language);
    Console.WriteLine("生成完成: " + files.Count + " files");
}

async Task Serve(String[] args)
{
    var cfgPath = args.FirstOrDefault(a => a.StartsWith("--config="))?.Substring("--config=".Length) ?? "config.yaml";
    //var config = AppConfig.Load(cfgPath) ?? AppConfig.Default();
    var config = AppConfig.Current;
    await WikiServer.StartAsync(args, config, CancellationToken.None);
}

async Task AiTest(String[] args)
{
    var cfgPath = args.FirstOrDefault(a => a.StartsWith("--config="))?.Substring("--config=".Length) ?? "config.yaml";
    //var config = AppConfig.Load(cfgPath) ?? AppConfig.Default();
    var config = AppConfig.Current;
    var prompt = args.FirstOrDefault(a => a.StartsWith("--prompt="))?.Substring("--prompt=".Length) ?? "请用一句话介绍本项目";
    var providerName = args.FirstOrDefault(a => a.StartsWith("--provider="))?.Substring("--provider=".Length) ?? (config.AI.DefaultProvider ?? (config.AI.Providers.Keys.FirstOrDefault() ?? "openai"));

    var manager = new AIProviderManager();
    if (String.IsNullOrEmpty(providerName) || !config.AI.Providers.TryGetValue(providerName, out var p)) { XTrace.WriteLine("配置中未找到AI提供者: {0}", providerName); return; }

    switch (providerName.ToLowerInvariant())
    {
        case "openai":
            var openaiKey = p.ApiKey ?? Environment.GetEnvironmentVariable("OPENAI_API_KEY") ?? "";
            if (String.IsNullOrEmpty(openaiKey)) { XTrace.WriteLine("缺少OpenAI ApiKey，设置环境变量 OPENAI_API_KEY"); return; }
            manager.Register(new OpenAIProvider(providerName, openaiKey, p.BaseUrl, p.Model));
            break;
        case "gemini":
            var geminiKey = p.ApiKey ?? Environment.GetEnvironmentVariable("GOOGLE_API_KEY") ?? "";
            if (String.IsNullOrEmpty(geminiKey)) { XTrace.WriteLine("缺少Gemini ApiKey，设置环境变量 GOOGLE_API_KEY"); return; }
            manager.Register(new GeminiProvider(providerName, geminiKey, p.Model ?? "gemini-2.0-flash-exp", p.BaseUrl));
            break;
        case "deepseek":
            var deepseekKey = p.ApiKey ?? Environment.GetEnvironmentVariable("DEEPSEEK_API_KEY") ?? "";
            if (String.IsNullOrEmpty(deepseekKey)) { XTrace.WriteLine("缺少DeepSeek ApiKey，设置环境变量 DEEPSEEK_API_KEY"); return; }
            manager.Register(new DeepSeekProvider(deepseekKey, p.Model ?? "deepseek-chat", p.BaseUrl));
            break;
        case "ollama":
            manager.Register(new OllamaProvider(p.Model ?? "llama3", p.BaseUrl));
            break;
        default:
            XTrace.WriteLine("暂未实现该提供者: {0}", providerName); return;
    }

    var prov = manager.Get(providerName);
    XTrace.WriteLine("使用提供者 {0} 调用模型 {1} ...", providerName, p.Model ?? "(默认)");
    var res = await prov.GenerateAsync(new GenerationOptions { Prompt = prompt, Model = p.Model, Temperature = p.Temperature, MaxTokens = p.MaxTokens }, CancellationToken.None);
    XTrace.WriteLine("AI 输出长度={0}", res.Text?.Length ?? 0);
    Console.WriteLine("---- AI 输出 ----\n" + res.Text);
}

Int32 ShowHelp()
{
    Console.WriteLine("NewLife.Wiki - AI 驱动的代码仓库文档生成器 (.NET 版)");
    Console.WriteLine("命令:");
    Console.WriteLine("  analyze  <path>                                      分析本地代码目录");
    Console.WriteLine("  generate <repoPath> [--out=dir] [--lang=zh|en]       生成文档 (最小版本)");
    Console.WriteLine("           [--provider=name] [--config=config.yaml]");
    Console.WriteLine("  serve    [port]                                      启动 HTTP 服务 (待实现)");
    Console.WriteLine("  ai-test  [--provider=] [--prompt=]                   测试指定AI提供者");
    Console.WriteLine("  help                                                 显示帮助");
    return 0;
}
