using System.Security.Cryptography;
using System.Text;
using NewLife.Log;
using NewLife.Wiki.Models;

namespace NewLife.Wiki.Analyzer;

/// <summary>代码仓库分析器接口</summary>
/// <remarks>
/// 抽象对源代码目录的结构化分析，返回文件、依赖、模块、函数/类及统计指标等概要信息。
/// 实现可根据需要扩展复杂度计算、AST 解析、依赖图构建等能力。
/// </remarks>
public interface ICodeAnalyzer
{
    /// <summary>异步分析指定根目录</summary>
    /// <param name="root">仓库根目录的本地路径</param>
    /// <param name="cancellationToken">取消令牌，用于在长时间分析过程中提前终止</param>
    /// <returns>代码结构结果，包含文件列表与基础度量</returns>
    Task<CodeStructure> AnalyzeAsync(String root, CancellationToken cancellationToken = default);
}

/// <summary>默认仓库分析器</summary>
/// <remarks>
/// 递归扫描目录，收集文件基础信息（名称、扩展名、大小、修改时间、哈希、行数）。
/// 可选载入文本内容（受 <see cref="MaxFileBytes"/> 限制）以便后续生成摘要或进一步解析。
/// 当前实现不做 AST / 复杂度 / 依赖关系解析，仅提供最小结构数据。
/// </remarks>
public sealed class RepositoryAnalyzer : ICodeAnalyzer
{
    /// <summary>日志接口</summary>
    public ILog? Log { get; set; } = XTrace.Log;
    /// <summary>是否加载文件内容</summary>
    /// <remarks>启用后会读取文本并计算行数，超过 <see cref="MaxFileBytes"/> 的文件自动跳过内容载入。</remarks>
    public Boolean LoadContent { get; set; }
    /// <summary>单文件内容读取的最大字节数限制</summary>
    public Int64 MaxFileBytes { get; set; } = 512 * 1024;
    /// <summary>排除的文件扩展名集合（不区分大小写）</summary>
    public HashSet<String> ExcludeExtensions { get; } = new(StringComparer.OrdinalIgnoreCase)
    {
        ".png",".jpg",".jpeg",".gif",".bmp",".svg",".ico",".dll",".exe",".so",".dylib",".zip",".tar",".gz",".7z"
    };

    /// <summary>同步分析（包装异步实现）</summary>
    /// <param name="root">仓库根目录路径</param>
    /// <returns>代码结构</returns>
    public CodeStructure Analyze(String root) => AnalyzeAsync(root, CancellationToken.None).GetAwaiter().GetResult();

    /// <inheritdoc />
    public Task<CodeStructure> AnalyzeAsync(String root, CancellationToken cancellationToken = default)
    {
        if (String.IsNullOrEmpty(root)) throw new ArgumentNullException(nameof(root));
        if (!Directory.Exists(root)) throw new DirectoryNotFoundException(root);

        Log?.Info("开始分析: {0}", root);
        var cs = new CodeStructure { RepositoryId = Path.GetFileName(root.TrimEnd(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar)) };

        var files = Directory.GetFiles(root, "*", SearchOption.AllDirectories).ToList();
        foreach (var file in files)
        {
            cancellationToken.ThrowIfCancellationRequested();
            try
            {
                var fi = new FileInfo(file);
                var rel = GetRelativePath(root, file);
                var ext = fi.Extension;
                if (ExcludeExtensions.Contains(ext)) continue;

                var model = new FileInfoModel
                {
                    Path = rel,
                    Name = fi.Name,
                    Extension = ext,
                    Size = fi.Length,
                    Language = GuessLanguage(ext),
                    ModifiedAt = fi.LastWriteTimeUtc,
                    Hash = HashFile(fi)
                };

                if (LoadContent && fi.Length <= MaxFileBytes)
                {
                    var text = File.ReadAllText(file, Encoding.UTF8);
                    model.Content = text;
                    model.LineCount = CountLines(text);
                }
                else
                {
                    model.LineCount = 0;
                }

                cs.Files.Add(model);
            }
            catch (Exception ex)
            {
                Log?.Warn("跳过文件 {0}: {1}", file, ex.Message);
            }
        }

        cs.Metrics.TotalFiles = cs.Files.Count;
        cs.Metrics.TotalLines = 0;
        foreach (var f in cs.Files) cs.Metrics.TotalLines += f.LineCount;
        cs.Metrics.CodeLines = cs.Metrics.TotalLines;
        cs.Metrics.CommentLines = 0;

        return Task.FromResult(cs);
    }

    private static Int32 CountLines(String text)
    {
        if (String.IsNullOrEmpty(text)) return 0;
        var count = 1;
        for (var i = 0; i < text.Length; i++) if (text[i] == '\n') count++;
        return count;
    }

    private static String GuessLanguage(String ext)
    {
        if (String.IsNullOrEmpty(ext)) return String.Empty;

        return ext.ToLowerInvariant() switch
        {
            ".cs" => "C#",
            ".go" => "Go",
            ".js" => "JavaScript",
            ".ts" => "TypeScript",
            ".py" => "Python",
            ".java" => "Java",
            ".rs" => "Rust",
            ".cpp" or ".cc" or ".cxx" or ".hpp" or ".h" or ".c" => "C/C++",
            ".md" => "Markdown",
            ".json" => "JSON",
            ".yml" or ".yaml" => "YAML",
            _ => ext.TrimStart('.'),
        };
    }

    private static String HashFile(FileInfo fi)
    {
        using var sha = SHA1.Create();
        using var fs = fi.OpenRead();
        var hash = sha.ComputeHash(fs);
        var sb = new StringBuilder(hash.Length * 2);
        for (var i = 0; i < hash.Length; i++) sb.Append(hash[i].ToString("X2"));
        return sb.ToString();
    }

    private static String GetRelativePath(String root, String full)
    {
        if (full.StartsWith(root, StringComparison.OrdinalIgnoreCase))
        {
            var rel = full.Substring(root.Length).TrimStart(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar);
            return rel.Replace(Path.DirectorySeparatorChar, '/');
        }
        return full;
    }
}
