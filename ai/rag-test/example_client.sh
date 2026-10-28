#!/bin/bash

# RAG 系统客户端示例脚本

SERVER_URL="http://localhost:8080"

echo "=== RAG 系统客户端示例 ==="
echo ""

# 健康检查
echo "1. 健康检查..."
curl -s "$SERVER_URL/api/health" | jq .
echo ""

# # 上传文档（需要替换为实际文件路径）
# echo "2. 上传文档..."
# echo "正在上传并处理文档（这可能需要一些时间，请查看服务器日志了解进度）..."
# echo ""

# response=$(curl -s -w "\n%{http_code}" -X POST "$SERVER_URL/api/upload" \
#   -F "file=@/Users/hmyy/rag-test/book.pdf")

# http_code=$(echo "$response" | tail -n1)
# body=$(echo "$response" | sed '$d')

# if [ "$http_code" = "200" ]; then
#   if command -v jq &> /dev/null; then
#     echo "$body" | jq .
#   else
#     echo "$body"
#   fi
#   echo ""
#   echo "✓ 文档上传和处理成功！"
# else
#   echo "✗ 上传失败 (HTTP $http_code)"
#   echo "$body"
# fi
# echo ""

# 查询示例
echo "3. 查询示例..."
response=$(curl -s -X POST "$SERVER_URL/api/query" \
  -H "Content-Type: application/json" \
  -d '{"question": "文档的主要内容是什么？"}')

if command -v jq &> /dev/null; then
  echo "$response" | jq .
else
  echo "$response"
fi
echo ""

echo "完成！"
