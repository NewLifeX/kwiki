#!/bin/bash

# 测试模板文档生成API的脚本

echo "=== 测试模板文档生成API ==="

# 服务器地址
SERVER_URL="http://localhost:8080"

# 检查服务器是否运行
echo "检查服务器状态..."
if ! curl -s "$SERVER_URL/api/info" > /dev/null; then
    echo "错误: 服务器未运行，请先启动服务器"
    echo "运行: go run cmd/kwiki/main.go"
    exit 1
fi

echo "服务器运行正常"

# 获取可用的AI提供商
echo ""
echo "获取可用的AI提供商..."
curl -s "$SERVER_URL/api/providers" | jq '.'

# 发送模板文档生成请求
echo ""
echo "发送模板文档生成请求..."

RESPONSE=$(curl -s -X POST "$SERVER_URL/api/wiki/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "repository_url": "template-docs",
    "title": "KWiki模板系统文档",
    "description": "KWiki模板系统的完整文档",
    "languages": ["zh"],
    "primary_language": "zh",
    "settings": {
      "ai_provider": "deepseek",
      "model": "deepseek-chat",
      "temperature": 0.7,
      "max_tokens": 4000,
      "enable_rag": false,
      "enable_diagrams": false
    }
  }')

echo "响应: $RESPONSE"

# 提取wiki_id
WIKI_ID=$(echo "$RESPONSE" | jq -r '.wiki_id // empty')

if [ -z "$WIKI_ID" ]; then
    echo "错误: 无法获取wiki_id"
    echo "响应: $RESPONSE"
    exit 1
fi

echo ""
echo "Wiki ID: $WIKI_ID"
echo "开始监控生成进度..."

# 监控生成进度
for i in {1..60}; do
    echo ""
    echo "=== 第 $i 次检查 ($(date)) ==="
    
    # 获取进度
    PROGRESS=$(curl -s "$SERVER_URL/api/wiki/$WIKI_ID/progress")
    echo "进度: $PROGRESS"
    
    STATUS=$(echo "$PROGRESS" | jq -r '.status // empty')
    PROGRESS_PCT=$(echo "$PROGRESS" | jq -r '.progress // 0')
    
    echo "状态: $STATUS, 进度: $PROGRESS_PCT%"
    
    # 获取日志
    LOGS=$(curl -s "$SERVER_URL/api/wiki/$WIKI_ID/logs")
    LATEST_LOGS=$(echo "$LOGS" | jq -r '.logs[-3:] // [] | .[]' 2>/dev/null)
    if [ ! -z "$LATEST_LOGS" ]; then
        echo "最新日志:"
        echo "$LATEST_LOGS"
    fi
    
    # 检查是否完成
    if [ "$STATUS" = "completed" ]; then
        echo ""
        echo "🎉 生成完成!"
        break
    elif [ "$STATUS" = "failed" ]; then
        echo ""
        echo "❌ 生成失败"
        break
    fi
    
    # 等待10秒
    sleep 10
done

# 获取最终结果
echo ""
echo "=== 最终结果 ==="

WIKI=$(curl -s "$SERVER_URL/api/wiki/$WIKI_ID")
echo "Wiki信息:"
echo "$WIKI" | jq '.'

# 获取页面列表
echo ""
echo "生成的页面:"
PAGES=$(curl -s "$SERVER_URL/api/wiki/$WIKI_ID/pages")
echo "$PAGES" | jq '.[] | {id: .id, title: .title, word_count: .word_count, reading_time: .reading_time}'

# 获取完整日志
echo ""
echo "完整日志:"
curl -s "$SERVER_URL/api/wiki/$WIKI_ID/logs" | jq -r '.logs[] // empty'

echo ""
echo "=== 测试完成 ==="
echo "Wiki URL: $SERVER_URL/wiki/$WIKI_ID"
