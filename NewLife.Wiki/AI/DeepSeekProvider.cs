using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NewLife.Wiki.AI;

/// <summary>DeepSeek 提供者实现。兼容 OpenAI Chat Completions 风格。</summary>
public sealed class DeepSeekProvider : IAIProvider, IDisposable
{
    #region 属性
    /// <summary>名称</summary>
    public String Name { get; } = "deepseek";
    #endregion

    private readonly String _apiKey;
    private readonly String _model;
    private readonly String _baseUrl;
    private readonly HttpClient _http;

    /// <summary>实例化 DeepSeek 提供者</summary>
    /// <param name="apiKey">API Key</param>
    /// <param name="model">模型，默认 deepseek-chat</param>
    /// <param name="baseUrl">基地址，默认 https://api.deepseek.com</param>
    /// <param name="httpClient">可注入 HttpClient 复用</param>
    public DeepSeekProvider(String apiKey, String model = "deepseek-chat", String? baseUrl = null, HttpClient? httpClient = null)
    {
        _apiKey = apiKey ?? throw new ArgumentNullException(nameof(apiKey));
        _model = model;
        _baseUrl = (baseUrl ?? "https://api.deepseek.com").TrimEnd('/');
        _http = httpClient ?? new HttpClient();
        if (!_http.DefaultRequestHeaders.Contains("Authorization")) _http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", _apiKey);
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
        var url = _baseUrl + "/v1/chat/completions";
        var req = new ChatRequest
        {
            Model = opt.Model ?? _model,
            Temperature = opt.Temperature,
            MaxTokens = opt.MaxTokens,
            Stream = opt.Stream,
            Messages = BuildMessages(opt)
        };
        if (opt.Extra != null)
        {
            foreach (var kv in opt.Extra)
            {
                if (kv.Value == null) continue; req.Extra[kv.Key] = kv.Value;
            }
        }
        var payload = Serialize(req);
        using var msg = new HttpRequestMessage(HttpMethod.Post, url)
        {
            Content = new StringContent(payload, Encoding.UTF8, "application/json")
        };
        using var resp = await _http.SendAsync(msg, opt.Stream ? HttpCompletionOption.ResponseHeadersRead : HttpCompletionOption.ResponseContentRead, ct).ConfigureAwait(false);
        resp.EnsureSuccessStatusCode();
        if (!opt.Stream)
        {
            var body = await resp.Content.ReadAsStringAsync().ConfigureAwait(false);
            var cr = Deserialize(body) ?? new ChatResponse();
            var text = cr.Choices.Count > 0 ? (cr.Choices[0].Message?.Content ?? String.Empty) : String.Empty;
            return new GenerationResult { Text = text, Raw = cr, Usage = cr.Usage != null ? new TokenUsage { PromptTokens = cr.Usage.PromptTokens, CompletionTokens = cr.Usage.CompletionTokens } : null };
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
                if (line.StartsWith("data: ")) line = line.Substring(6);
                if (line == "[DONE]") break;
                try
                {
                    var part = Deserialize(line);
                    if (part != null && part.Choices != null && part.Choices.Count > 0)
                    {
                        var first = part.Choices[0];
                        var delta = first?.Delta?.Content;
                        if (!String.IsNullOrEmpty(delta))
                        {
                            sb.Append(delta);
                            if (onDelta != null)
                            {
                                if (!await onDelta(new StreamDelta { Text = delta }).ConfigureAwait(false)) break;
                            }
                        }
                    }
                }
                catch { }
            }
            if (onDelta != null) await onDelta(new StreamDelta { Completed = true }).ConfigureAwait(false);
            return new GenerationResult { Text = sb.ToString() };
        }
    }

    private static List<ChatMessage> BuildMessages(GenerationOptions opt)
    {
        var list = new List<ChatMessage>();
        if (!String.IsNullOrEmpty(opt.System)) list.Add(new ChatMessage { Role = "system", Content = opt.System });
        list.Add(new ChatMessage { Role = "user", Content = opt.Prompt });
        return list;
    }

    private static String Serialize(ChatRequest req) => JsonSerializer.Serialize(req, JsonOptions);

    private static ChatResponse? Deserialize(String json) { try { return JsonSerializer.Deserialize<ChatResponse>(json, JsonOptions); } catch { return null; } }

    private static readonly JsonSerializerOptions JsonOptions = new(JsonSerializerDefaults.Web)
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
    };

    /// <summary>释放 HttpClient</summary>
    public void Dispose() => _http.Dispose();

    #region DTO
    private sealed class ChatRequest
    {
        public String Model { get; set; } = String.Empty;

        public List<ChatMessage> Messages { get; set; } = [];

        public Double Temperature { get; set; }

        public Int32 MaxTokens { get; set; }

        public Boolean Stream { get; set; }

        public Dictionary<String, Object?> Extra { get; set; } = [];
    }

    private sealed class ChatMessage
    {
        public String Role { get; set; } = String.Empty;

        public String? Content { get; set; }
    }

    private sealed class ChatResponse
    {
        public List<Choice> Choices { get; set; } = [];

        public UsageInfo? Usage { get; set; }

        public sealed class Choice
        {
            public Int32 Index { get; set; }

            public ChatMessage? Message { get; set; }

            public DeltaMessage? Delta { get; set; }

            public String? FinishReason { get; set; }
        }

        public sealed class DeltaMessage
        {
            public String? Content { get; set; }
        }

        public sealed class UsageInfo
        {
            public Int32 PromptTokens { get; set; }

            public Int32 CompletionTokens { get; set; }
        }
    }
    #endregion
}
