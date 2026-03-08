#!/bin/bash
# go-agent SharedSession 快速测试脚本

set -e

export DEEPSEEK_API_KEY="sk-7398e54a08cf4c1ba86500a9bff10a18"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"

cd /Users/frank/MareCogito/go-agent

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║     go-agent SharedSession 共享记忆测试                      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 检查 PostgreSQL
if ! docker ps | grep -q postgres-pgvector; then
    echo "📦 启动 PostgreSQL..."
    docker run -d --name postgres-pgvector \
        -e POSTGRES_USER=admin \
        -e POSTGRES_PASSWORD=admin \
        -e POSTGRES_DB=ragdb \
        -p 5432:5432 \
        pgvector/pgvector:pg16
    sleep 5
fi
echo "✅ PostgreSQL 运行中"

# 检查 Ollama
if ! curl -s http://localhost:11434/api/tags | grep -q nomic-embed-text; then
    echo "⚠️  请运行：ollama pull nomic-embed-text"
fi
echo "✅ Ollama 就绪"

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "运行 SharedSession 测试..."
echo "══════════════════════════════════════════════════════════════"
echo ""

go run cmd/example/shared_session_test/main.go 2>&1 | grep -E "╔|║|╚|TEST|Step|✅|⚠️|📊|found"

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "查看 PostgreSQL 中的记忆..."
echo "══════════════════════════════════════════════════════════════"

docker exec postgres-pgvector psql -U admin -d ragdb \
    -c "SELECT id, session_id, substring(content, 1, 50) as content FROM memory_bank ORDER BY id DESC LIMIT 5;" 2>&1

echo ""
echo "✅ 测试完成！"
