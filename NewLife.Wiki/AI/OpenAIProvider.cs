using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NewLife.Wiki.AI;

/// <summary>简单OpenAI兼容实现。支持 /v1/chat/completions</summary>
public class OpenAIProvider : IAIProvider, IDisposable
{
    private readonly HttpClient _http;

    /// <summary>名称</summary>
    public String Name { get; }

    private readonly String _apiKey;
    private readonly String _baseUrl;
    private readonly String _defaultModel;

    /// <summary>实例化 OpenAI 提供者</summary>
    /// <param name="name">提供者名称注册键</param>
    /// <param name="apiKey">API 密钥</param>
    /// <param name="baseUrl">可选基地址，默认官方 https://api.openai.com</param>
    /// <param name="defaultModel">默认模型名称</param>
    /// <param name="httpClient">自定义 HttpClient（可复用，提高并发）</param>
    public OpenAIProvider(String name, String apiKey, String? baseUrl = null, String? defaultModel = null, HttpClient? httpClient = null)
    {
        Name = name;
        _apiKey = apiKey ?? throw new ArgumentNullException(nameof(apiKey));
        _baseUrl = baseUrl?.TrimEnd('/') ?? "https://api.openai.com";
        _defaultModel = defaultModel ?? "gpt-3.5-turbo";
        _http = httpClient ?? new HttpClient();
        if (!_http.DefaultRequestHeaders.Contains("Authorization")) _http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", _apiKey);
    }

    /// <summary>生成完整结果</summary>
    /// <param name="options">生成参数</param>
    /// <param name="cancellationToken">取消令牌</param>
    public async Task<GenerationResult> GenerateAsync(GenerationOptions options, CancellationToken cancellationToken = default)
    {
        options.Stream = false;
        var res = await SendAsync(options, null, cancellationToken).ConfigureAwait(false);
        return res;
    }

    /// <summary>流式生成，逐段回调</summary>
    /// <param name="options">生成参数</param>
    /// <param name="onDelta">增量回调，返回 false 中断</param>
    /// <param name="cancellationToken">取消令牌</param>
    public async Task<GenerationResult> StreamAsync(GenerationOptions options, Func<StreamDelta, Task<Boolean>> onDelta, CancellationToken cancellationToken = default)
    {
        options.Stream = true;
        var final = await SendAsync(options, onDelta, cancellationToken).ConfigureAwait(false);
        return final;
    }

    private async Task<GenerationResult> SendAsync(GenerationOptions options, Func<StreamDelta, Task<Boolean>>? onDelta, CancellationToken ct)
    {
        var url = _baseUrl + "/v1/chat/completions";
        var req = new ChatRequest
        {
            Model = options.Model ?? _defaultModel,
            Temperature = options.Temperature,
            MaxTokens = options.MaxTokens,
            Stream = options.Stream,
            Messages = BuildMessages(options)
        };

        if (options.Extra != null)
        {
            foreach (var kv in options.Extra)
            {
                if (kv.Value == null) continue;
                req.Extra[kv.Key] = kv.Value;
            }
        }

        var payload = Serialize(req);
        using var msg = new HttpRequestMessage(HttpMethod.Post, url)
        {
            Content = new StringContent(payload, Encoding.UTF8, "application/json")
        };

        using var resp = await _http.SendAsync(msg, HttpCompletionOption.ResponseHeadersRead, ct).ConfigureAwait(false);
        resp.EnsureSuccessStatusCode();

        if (!req.Stream)
        {
            var json = await resp.Content.ReadAsStringAsync().ConfigureAwait(false);
            var cr = Deserialize(json) ?? new ChatResponse();
            var text = cr.Choices.Count > 0 ? (cr.Choices[0].Message?.Content ?? String.Empty) : String.Empty;
            return new GenerationResult
            {
                Text = text,
                Usage = cr.Usage != null ? new TokenUsage { PromptTokens = cr.Usage.PromptTokens, CompletionTokens = cr.Usage.CompletionTokens } : null,
                Raw = cr
            };
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
            if (onDelta != null) await onDelta(new StreamDelta { Text = String.Empty, Completed = true }).ConfigureAwait(false);
            return new GenerationResult { Text = sb.ToString() };
        }
    }

    private static List<ChatMessage> BuildMessages(GenerationOptions opt)
    {
        var list = new List<ChatMessage>();
        if (!String.IsNullOrEmpty(opt.System)) list.Add(new ChatMessage { Role = "system", Content = opt.System });
        list.Add(new ChatMessage { Role = "user", Content = opt.Prompt ?? String.Empty });
        return list;
    }

    private static String Serialize(ChatRequest req)
    {
        return JsonSerializer.Serialize(req, JsonOptions);
    }

    private static ChatResponse? Deserialize(String json)
    {
        return JsonSerializer.Deserialize<ChatResponse>(json, JsonOptions);
    }

    private static readonly JsonSerializerOptions JsonOptions = new(JsonSerializerDefaults.Web)
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
    };

    /// <summary>释放底层 HttpClient</summary>
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
        public String? Id { get; set; }
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
