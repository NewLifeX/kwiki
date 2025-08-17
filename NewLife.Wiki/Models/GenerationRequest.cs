namespace NewLife.Wiki.Models;

/// <summary>Wiki 生成请求</summary>
public class GenerationRequest
{
    /// <summary>目标仓库地址</summary>
    public String RepositoryUrl { get; set; } = String.Empty;

    /// <summary>分支</summary>
    public String Branch { get; set; } = String.Empty;

    /// <summary>访问令牌（私有仓库可选）</summary>
    public String AccessToken { get; set; } = String.Empty;

    /// <summary>生成设置</summary>
    public WikiSettings Settings { get; set; } = new();

    /// <summary>自定义标题</summary>
    public String Title { get; set; } = String.Empty;

    /// <summary>自定义描述</summary>
    public String Description { get; set; } = String.Empty;

    /// <summary>额外自定义提示词（按顺序追加）</summary>
    public List<String> CustomPrompts { get; set; } = [];

    /// <summary>需生成的语言列表</summary>
    public List<String> Languages { get; set; } = [];

    /// <summary>主语言</summary>
    public String PrimaryLanguage { get; set; } = "zh";

    /// <summary>是否对全部检测到语言生成</summary>
    public Boolean GenerateAllLangs { get; set; }
}