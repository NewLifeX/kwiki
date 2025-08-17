using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NewLife.Wiki.AI;

/// <summary>Ollama 本地模型提供者。调用本地 HTTP (默认 http://localhost:11434)。</summary>
public sealed class OllamaProvider : IAIProvider, IDisposable
{
    /// <summary>名称</summary>
    public String Name { get; } = "ollama";

    private readonly String _model;
    private readonly String _baseUrl;
    private readonly HttpClient _http;

    /// <summary>实例化 Ollama 提供者</summary>
    /// <param name="model">模型名称（例如 llama3）</param>
    /// <param name="baseUrl">服务地址，默认 http://localhost:11434</param>
    /// <param name="httpClient">可复用 HttpClient</param>
    public OllamaProvider(String model = "llama3", String? baseUrl = null, HttpClient? httpClient = null)
    {
        _model = model;
        _baseUrl = (baseUrl ?? "http://localhost:11434").TrimEnd('/');
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
        var url = _baseUrl + (opt.Stream ? "/api/stream" : "/api/generate");
        var req = new OllamaRequest
        {
            Model = _model,
            Prompt = opt.Prompt ?? String.Empty,
            Stream = opt.Stream,
            Options = []
        };
        if (opt.Temperature > 0) req.Options["temperature"] = opt.Temperature;
        if (opt.MaxTokens > 0) req.Options["num_predict"] = opt.MaxTokens;
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
            var r = Deserialize(body);
            var text = r?.Response ?? String.Empty;
            return new GenerationResult { Text = text, Raw = r };
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
                try
                {
                    var part = Deserialize(line);
                    var delta = part?.Response;
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

    private static String Serialize(OllamaRequest req)
    {
        return JsonSerializer.Serialize(req, JsonOptions);
    }

    private static OllamaResponse? Deserialize(String json)
    {
        try { return JsonSerializer.Deserialize<OllamaResponse>(json, JsonOptions); } catch { return null; }
    }

    private static readonly JsonSerializerOptions JsonOptions = new(JsonSerializerDefaults.Web)
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
    };

    /// <summary>释放 HttpClient</summary>
    public void Dispose() => _http.Dispose();

    #region DTO
    private sealed class OllamaRequest
    {
        public String Model { get; set; } = String.Empty;
        public String Prompt { get; set; } = String.Empty;
        public Boolean Stream { get; set; }
        public Dictionary<String, Object?> Options { get; set; } = [];
    }

    private sealed class OllamaResponse
    {
        public String? Model { get; set; }
        public String? CreatedAt { get; set; }
        public String? Response { get; set; }
        public Boolean Done { get; set; }
    }
    #endregion
}
