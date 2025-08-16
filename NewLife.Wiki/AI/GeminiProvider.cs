using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Text;
#if NET6_0_OR_GREATER || NET5_0_OR_GREATER || NET8_0_OR_GREATER || NETSTANDARD2_0
using System.Text.Json;
using System.Text.Json.Serialization;
#endif
using System.Threading;
using System.Threading.Tasks;

namespace NewLife.Wiki.AI;

/// <summary>Google Gemini 提供者实现。支持 generateContent 与流式 streamGenerateContent。</summary>
public class GeminiProvider : IAIProvider, IDisposable
{
    /// <summary>名称</summary>
    public String Name { get; }

    private readonly String _apiKey;
    private readonly String _model;
    private readonly HttpClient _http;
    private readonly String _baseUrl;

    /// <summary>实例化 Gemini 提供者</summary>
    /// <param name="name">注册名称</param>
    /// <param name="apiKey">Google Generative AI API Key</param>
    /// <param name="model">模型名称，如 gemini-2.0-flash-exp</param>
    /// <param name="baseUrl">可选基地址，默认 https://generativelanguage.googleapis.com</param>
    /// <param name="httpClient">可复用 HttpClient</param>
    public GeminiProvider(String name, String apiKey, String model, String? baseUrl = null, HttpClient? httpClient = null)
    {
        Name = name;
        _apiKey = apiKey ?? throw new ArgumentNullException(nameof(apiKey));
        _model = model;
        _baseUrl = (baseUrl ?? "https://generativelanguage.googleapis.com").TrimEnd('/');
        _http = httpClient ?? new HttpClient();
    }

    /// <summary>生成完整结果</summary>
    public async Task<GenerationResult> GenerateAsync(GenerationOptions options, CancellationToken cancellationToken = default)
    {
        options.Stream = false;
        return await SendAsync(options, null, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>流式生成</summary>
    public async Task<GenerationResult> StreamAsync(GenerationOptions options, Func<StreamDelta, Task<Boolean>> onDelta, CancellationToken cancellationToken = default)
    {
        options.Stream = true;
        return await SendAsync(options, onDelta, cancellationToken).ConfigureAwait(false);
    }

    private async Task<GenerationResult> SendAsync(GenerationOptions opt, Func<StreamDelta, Task<Boolean>>? onDelta, CancellationToken ct)
    {
        var path = opt.Stream ? $"/v1beta/models/{_model}:streamGenerateContent" : $"/v1beta/models/{_model}:generateContent";
        var url = _baseUrl + path + "?key=" + _apiKey;
        var req = BuildRequest(opt);
        var json = Serialize(req);
        using var msg = new HttpRequestMessage(HttpMethod.Post, url)
        {
            Content = new StringContent(json, Encoding.UTF8, "application/json")
        };
        using var resp = await _http.SendAsync(msg, opt.Stream ? HttpCompletionOption.ResponseHeadersRead : HttpCompletionOption.ResponseContentRead, ct).ConfigureAwait(false);
        resp.EnsureSuccessStatusCode();

        if (!opt.Stream)
        {
            var body = await resp.Content.ReadAsStringAsync().ConfigureAwait(false);
            var gr = Deserialize(body);
            var text = ExtractText(gr);
            return new GenerationResult { Text = text, Raw = gr };
        }
        else
        {
            using var stream = await resp.Content.ReadAsStreamAsync().ConfigureAwait(false);
            using var reader = new StreamReader(stream, Encoding.UTF8, true);
            var sb = new StringBuilder();
            while (!reader.EndOfStream && !ct.IsCancellationRequested)
            {
                var line = await reader.ReadLineAsync().ConfigureAwait(false);
                if (line == null) break;
                if (line.Length == 0) continue;
                if (line.StartsWith("data:")) line = line.Substring(5).Trim();
                if (line == "[DONE]") break;
                try
                {
                    var partial = Deserialize(line);
                    var delta = ExtractText(partial);
                    if (!String.IsNullOrEmpty(delta))
                    {
                        sb.Append(delta);
                        if (onDelta != null)
                        {
                            if (!await onDelta(new StreamDelta { Text = delta }).ConfigureAwait(false)) break;
                        }
                    }
                }
                catch { }
            }
            if (onDelta != null) await onDelta(new StreamDelta { Completed = true }).ConfigureAwait(false);
            return new GenerationResult { Text = sb.ToString() };
        }
    }

    private static GeminiRequest BuildRequest(GenerationOptions opt)
    {
        var req = new GeminiRequest
        {
            Contents = new List<GeminiRequest.Content>
            {
                new()
                {
                    Role = "user",
                    Parts = new List<GeminiRequest.Part> { new() { Text = opt.Prompt } }
                }
            }
        };
        var cfg = new GeminiRequest.GenerationConfig
        {
            Temperature = opt.Temperature,
            MaxOutputTokens = opt.MaxTokens > 0 ? opt.MaxTokens : (Int32?)null
        };
    req.Generation = cfg;
        return req;
    }

    private static String ExtractText(GeminiResponse? resp)
    {
    if (resp?.Candidates == null || resp.Candidates.Count == 0) return String.Empty;
    var cand = resp.Candidates[0];
    if (cand?.Content?.Parts == null || cand.Content.Parts.Count == 0) return String.Empty;
        var sb = new StringBuilder();
        foreach (var p in cand.Content.Parts)
        {
            if (!String.IsNullOrEmpty(p.Text)) sb.Append(p.Text);
        }
        return sb.ToString();
    }

    private static String Serialize(GeminiRequest req)
    {
#if NET6_0_OR_GREATER || NET5_0_OR_GREATER || NET8_0_OR_GREATER || NETSTANDARD2_0
        return JsonSerializer.Serialize(req, JsonOpts);
#else
        // 极简序列化，仅包含 prompt
        var sb = new StringBuilder();
        sb.Append("{\"contents\":[{\"role\":\"user\",\"parts\":[{\"text\":\"")
          .Append(req.Contents[0].Parts[0].Text.Replace("\"", "\\\""))
          .Append("\"}]}]}");
        return sb.ToString();
#endif
    }

    private static GeminiResponse? Deserialize(String json)
    {
#if NET6_0_OR_GREATER || NET5_0_OR_GREATER || NET8_0_OR_GREATER || NETSTANDARD2_0
        try { return JsonSerializer.Deserialize<GeminiResponse>(json, JsonOpts); } catch { return null; }
#else
        return null;
#endif
    }

#if NET6_0_OR_GREATER || NET5_0_OR_GREATER || NET8_0_OR_GREATER || NETSTANDARD2_0
    private static readonly JsonSerializerOptions JsonOpts = new(JsonSerializerDefaults.Web)
    {
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase
    };
#endif

    /// <summary>释放 HttpClient</summary>
    public void Dispose() => _http.Dispose();

    #region DTO
    private sealed class GeminiRequest
    {
        public List<Content> Contents { get; set; } = new();
        public GenerationConfig? Generation { get; set; }

        public sealed class Content
        {
            public String? Role { get; set; }
            public List<Part> Parts { get; set; } = new();
        }

        public sealed class Part
        {
            public String? Text { get; set; }
        }

        public sealed class GenerationConfig
        {
            public Double Temperature { get; set; }
            public Int32? MaxOutputTokens { get; set; }
        }
    }

    private sealed class GeminiResponse
    {
        public List<Candidate> Candidates { get; set; } = new();

        public sealed class Candidate
        {
            public Content? Content { get; set; }
            public Single? SafetyRatings { get; set; }
        }

        public sealed class Content
        {
            public List<Part> Parts { get; set; } = new();
        }

        public sealed class Part
        {
            public String? Text { get; set; }
        }
    }
    #endregion
}
