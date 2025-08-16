using System;
using System.Collections.Generic;

namespace NewLife.Wiki.Models;

/// <summary>代码仓库信息</summary>
public class Repository
{
    /// <summary>本地唯一标识（目录名或派生）</summary>
    public String Id { get; set; } = String.Empty;
    /// <summary>远程仓库地址（HTTPS / SSH）</summary>
    public String Url { get; set; } = String.Empty;
    /// <summary>仓库名称</summary>
    public String Name { get; set; } = String.Empty;
    /// <summary>拥有者 / 组织</summary>
    public String Owner { get; set; } = String.Empty;
    /// <summary>托管平台（GitHub/GitLab/Gitee 等）</summary>
    public String Provider { get; set; } = String.Empty;
    /// <summary>默认分支</summary>
    public String Branch { get; set; } = String.Empty;
    /// <summary>本地克隆路径</summary>
    public String LocalPath { get; set; } = String.Empty;
    /// <summary>估算大小（字节）</summary>
    public Int64 Size { get; set; }
    /// <summary>文件数量</summary>
    public Int32 FileCount { get; set; }
    /// <summary>主要使用的语言集合</summary>
    public List<String> Languages { get; set; } = new();
    /// <summary>创建时间（UTC）</summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>最近更新时间（UTC）</summary>
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    /// <summary>简介</summary>
    public String Description { get; set; } = String.Empty;
    /// <summary>主题标签</summary>
    public List<String> Topics { get; set; } = new();
    /// <summary>许可证</summary>
    public String License { get; set; } = String.Empty;
    /// <summary>Star 数</summary>
    public Int32 Stars { get; set; }
    /// <summary>Fork 数</summary>
    public Int32 Forks { get; set; }
}
