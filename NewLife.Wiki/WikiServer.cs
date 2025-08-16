#if NET8_0_OR_GREATER
using System;
using System.IO;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.FileProviders;
using NewLife.Log;
using NewLife.Wiki.Generator;
using NewLife.Wiki.Config;
using NewLife.Wiki.AI;
namespace NewLife.Wiki;

/// <summary>最小 HTTP 服务骨架。仅在 net8.0 及以上可用。</summary>
public static class WikiServer
{
    /// <summary>启动 Web 服务器。</summary>
    /// <param name="args">命令行参数</param>
    /// <param name="config">应用配置</param>
    /// <param name="cancellationToken">取消令牌</param>
    public static async Task StartAsync(string[] args, AppConfig config, CancellationToken cancellationToken = default)
    {
        var builder = WebApplication.CreateBuilder(args);
        var app = builder.Build();

        app.MapGet("/health", () => Results.Ok(new { status = "ok", time = DateTime.UtcNow }));

        app.MapPost("/generate", async (HttpContext ctx) =>
        {
            try
            {
                var repoPath = ctx.Request.Query["path"].ToString();
                if (string.IsNullOrEmpty(repoPath) || !Directory.Exists(repoPath)) return Results.BadRequest("invalid path");
                var lang = ctx.Request.Query["lang"].ToString();
                if (string.IsNullOrEmpty(lang)) lang = "zh";
                var outDir = config.Generator?.OutputDir ?? "_wiki";

                IAIProvider? ai = null;
                var def = config.AI?.DefaultProvider;
                if (!string.IsNullOrEmpty(def) && config.AI!.Providers.TryGetValue(def, out var prov))
                {
                    if (def == "openai")
                    {
                        var key = prov.ApiKey ?? Environment.GetEnvironmentVariable("OPENAI_API_KEY") ?? string.Empty;
                        if (!string.IsNullOrEmpty(key)) ai = new OpenAIProvider(def, key, prov.BaseUrl, prov.Model);
                    }
                }

                var generator = new WikiGenerator { AI = ai };
                var files = await generator.GenerateAsync(repoPath, outDir, lang, cancellationToken);
                return Results.Ok(new { files, output = Path.GetFullPath(outDir) });
            }
            catch (Exception ex)
            {
                XTrace.WriteException(ex);
                return Results.Problem(ex.Message);
            }
        });

        // 简单静态文件（如果输出目录存在）
        var staticDir = config.Generator?.OutputDir ?? "_wiki";
        if (Directory.Exists(staticDir))
        {
            app.UseStaticFiles(new StaticFileOptions
            {
                FileProvider = new PhysicalFileProvider(Path.GetFullPath(staticDir)),
                RequestPath = ""
            });
        }

    XTrace.WriteLine("WikiServer starting. Endpoints: /health /generate");
    // 直接运行，不传 token（RunAsync 不支持命名 cancellationToken 参数）；上层可在需要时取消进程。
    await app.RunAsync();
    }
}
#endif
