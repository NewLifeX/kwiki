using NewLife.Log;
using NewLife.Wiki.AI;
using NewLife.Wiki.Analyzer;
using NewLife.Wiki.Models;

namespace NewLife.Wiki.Generator;

/// <summary>Wiki 文档生成器。最小可用版本：读取仓库目录，统计基础指标，生成概要、架构与 API 参考占位文件。</summary>
public class WikiGenerator
{
    /// <summary>日志接口</summary>
    public ILog? Log { get; set; } = XTrace.Log;

    /// <summary>AI 提供器（可选，用于生成概要段落）</summary>
    public IAIProvider? AI { get; set; }

    /// <summary>代码分析器，可注入替换。默认使用 <see cref="RepositoryAnalyzer"/>。</summary>
    public ICodeAnalyzer Analyzer { get; set; } = new RepositoryAnalyzer();

    /// <summary>生成 Wiki 文档</summary>
    /// <param name="repoPath">源代码根目录</param>
    /// <param name="outputDir">输出目录，自动创建</param>
    /// <param name="language">目标语言，默认 zh</param>
    /// <returns>生成的文件路径集合</returns>
    public IList<String> Generate(String repoPath, String outputDir, String language = "zh")
    {
        if (String.IsNullOrEmpty(repoPath)) throw new ArgumentNullException(nameof(repoPath));
        if (!Directory.Exists(repoPath)) throw new DirectoryNotFoundException(repoPath);
        if (String.IsNullOrEmpty(outputDir)) throw new ArgumentNullException(nameof(outputDir));

        Directory.CreateDirectory(outputDir);
        if (Analyzer is RepositoryAnalyzer ra) ra.Log = Log;
        Log?.Info("Analyze repository: {0}", repoPath);
        var structure = (Analyzer as RepositoryAnalyzer)?.Analyze(repoPath) ?? Analyzer.AnalyzeAsync(repoPath).GetAwaiter().GetResult();

        // 尝试通过 AI 生成项目概要（可选）
        var overview = GenerateOverview(structure, language);

        var files = new List<String>
        {
            WriteReadme(structure, overview, outputDir, language),
            WriteArchitecture(structure, outputDir, language),
            WriteApiReference(structure, outputDir, language)
        };

        Log?.Info("Generated {0} files to {1}", files.Count, outputDir);
        return files;
    }

    /// <summary>异步生成</summary>
    public async Task<IList<String>> GenerateAsync(String repoPath, String outputDir, String language = "zh", CancellationToken cancellationToken = default)
    {
        if (String.IsNullOrEmpty(repoPath)) throw new ArgumentNullException(nameof(repoPath));
        if (!Directory.Exists(repoPath)) throw new DirectoryNotFoundException(repoPath);
        if (String.IsNullOrEmpty(outputDir)) throw new ArgumentNullException(nameof(outputDir));

        Directory.CreateDirectory(outputDir);
        if (Analyzer is RepositoryAnalyzer ra) ra.Log = Log;
        Log?.Info("Analyze repository (async): {0}", repoPath);
        var structure = await Analyzer.AnalyzeAsync(repoPath, cancellationToken);

        var overview = await GenerateOverviewAsync(structure, language, cancellationToken);

        var files = new List<String>
        {
            WriteReadme(structure, overview, outputDir, language),
            WriteArchitecture(structure, outputDir, language),
            WriteApiReference(structure, outputDir, language)
        };
        Log?.Info("Generated {0} files to {1}", files.Count, outputDir);
        return files;
    }

    private String GenerateOverview(CodeStructure structure, String language)
    {
        try
        {
            if (AI == null) return "(未配置 AI，跳过自动概要生成)";

            var prompt = BuildOverviewPrompt(structure, language);
            var result = AI.GenerateAsync(new GenerationOptions
            {
                Prompt = prompt,
                MaxTokens = 512,
                Temperature = 0.7,
                Model = null
            }).GetAwaiter().GetResult();
            return result.Text ?? String.Empty;
        }
        catch (Exception ex)
        {
            Log?.Warn("AI overview failed: {0}", ex.Message);
            return "(AI 概要生成失败)";
        }
    }

    private async Task<String> GenerateOverviewAsync(CodeStructure structure, String language, CancellationToken cancellationToken)
    {
        try
        {
            if (AI == null) return "(未配置 AI，跳过自动概要生成)";
            var prompt = BuildOverviewPrompt(structure, language);
            var result = await AI.GenerateAsync(new GenerationOptions { Prompt = prompt, MaxTokens = 512, Temperature = 0.7 }, cancellationToken);
            return result.Text ?? String.Empty;
        }
        catch (Exception ex)
        {
            Log?.Warn("AI overview failed: {0}", ex.Message);
            return "(AI 概要生成失败)";
        }
    }

    private String BuildOverviewPrompt(CodeStructure structure, String language)
    {
        var topFiles = structure.Files.OrderByDescending(e => e.LineCount).Take(10).ToList();
        var langs = String.Join(", ", structure.Files.Select(e => e.Language).Where(e => !String.IsNullOrEmpty(e)).GroupBy(e => e!).Select(g => g.Key + "(" + g.Count() + ")"));
        var lines = structure.Metrics != null ? structure.Metrics.TotalLines : 0;
        var repoId = structure.RepositoryId;
        var prompt = language == "en" ?
            "Generate a concise project overview (purpose, main languages, size) based on metrics:\n" :
            "根据以下指标生成一个简洁项目概要（用途/主要语言/规模）：\n";
        prompt += "Repo: " + repoId + "\n";
        prompt += "Lines: " + lines + "\n";
        prompt += "Languages: " + langs + "\n";
        prompt += "Top Files: " + String.Join(", ", topFiles.Select(e => e.Name + ":" + e.LineCount)) + "\n";
        prompt += (language == "en" ? "Use markdown. 120 words max." : "使用 Markdown，最多 200 字。") + "\n";
        return prompt;
    }

    private String WriteReadme(CodeStructure structure, String overview, String outputDir, String language)
    {
        var path = Path.Combine(outputDir, language == "en" ? "README.md" : "README.zh.md");
        var totalLines = structure.Metrics != null ? structure.Metrics.TotalLines : 0;
        var totalFiles = structure.Metrics != null ? structure.Metrics.TotalFiles : 0;
        var langs = String.Join(", ", structure.Files.Select(e => e.Language).Where(e => !String.IsNullOrEmpty(e)).GroupBy(e => e!).Select(g => g.Key + "(" + g.Count() + ")"));
        using (var sw = new StreamWriter(path, false))
        {
            sw.WriteLine("# " + (language == "en" ? "Project Overview" : "项目概览"));
            sw.WriteLine();
            sw.WriteLine(overview);
            sw.WriteLine();
            sw.WriteLine("## " + (language == "en" ? "Metrics" : "指标"));
            sw.WriteLine();
            sw.WriteLine("- Lines: " + totalLines);
            sw.WriteLine("- Files: " + totalFiles);
            sw.WriteLine("- Languages: " + langs);
            sw.WriteLine();
            sw.WriteLine("## " + (language == "en" ? "Structure" : "结构"));
            sw.WriteLine();
            foreach (var m in structure.Modules.Take(20)) sw.WriteLine("- `" + m.Name + "` (" + m.LineCount + " lines)");
        }
        return path;
    }

    private String WriteArchitecture(CodeStructure structure, String outputDir, String language)
    {
        var path = Path.Combine(outputDir, language == "en" ? "architecture.md" : "architecture.zh.md");
        using (var sw = new StreamWriter(path, false))
        {
            sw.WriteLine("# " + (language == "en" ? "Architecture" : "架构"));
            sw.WriteLine();
            sw.WriteLine("(" + (language == "en" ? "Placeholder for architecture description." : "架构说明占位，后续可由 AI 细化。") + ")");
            sw.WriteLine();
            sw.WriteLine("## Modules");
            foreach (var m in structure.Modules.Take(50))
                sw.WriteLine("- " + m.Name + " (" + m.LineCount + " lines)");
        }
        return path;
    }

    private String WriteApiReference(CodeStructure structure, String outputDir, String language)
    {
        var path = Path.Combine(outputDir, language == "en" ? "api-reference.md" : "api-reference.zh.md");
        using (var sw = new StreamWriter(path, false))
        {
            sw.WriteLine("# API Reference");
            sw.WriteLine();
            sw.WriteLine("(" + (language == "en" ? "Simplified file list; future version will include symbol parsing." : "简化文件列表；未来版本将包含符号解析。") + ")");
            sw.WriteLine();
            foreach (var f in structure.Files.OrderByDescending(e => e.LineCount).Take(200))
                sw.WriteLine("- " + f.Path + " (" + f.LineCount + " lines)");
        }
        return path;
    }
}
