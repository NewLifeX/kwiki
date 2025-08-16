using System;
using System.Security.Cryptography;
using System.Text;

namespace NewLife.Wiki.Utils;

/// <summary>ID 与散列工具</summary>
public static class IdUtil
{
    /// <summary>生成随机 Id</summary>
    public static String NewId()
    {
#if NET5_0_OR_GREATER || NETSTANDARD2_1_OR_GREATER || NETCOREAPP
    Span<Byte> buf = stackalloc Byte[16];
    RandomNumberGenerator.Fill(buf);
    return Convert.ToHexString(buf).ToLowerInvariant();
#else
    var buf = new Byte[16];
    using (var rng = RandomNumberGenerator.Create()) rng.GetBytes(buf);
    var sb = new StringBuilder(buf.Length * 2);
    foreach (var b in buf) sb.Append(b.ToString("x2"));
    return sb.ToString();
#endif
    }

    /// <summary>计算 MD5</summary>
    public static String MD5(String text)
    {
    using var md5 = System.Security.Cryptography.MD5.Create();
    var bytes = Encoding.UTF8.GetBytes(text);
    var hash = md5.ComputeHash(bytes);
#if NET5_0_OR_GREATER || NETSTANDARD2_1_OR_GREATER || NETCOREAPP
    return Convert.ToHexString(hash).ToLowerInvariant();
#else
    var sb = new StringBuilder(hash.Length * 2);
    foreach (var b in hash) sb.Append(b.ToString("x2"));
    return sb.ToString();
#endif
    }
}
