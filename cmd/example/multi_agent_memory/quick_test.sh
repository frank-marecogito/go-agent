#!/bin/bash
# 快速测试命令集合
# 用法：复制粘贴到终端运行

# ═══════════════════════════════════════════════════════════════
# 环境变量配置
# ═══════════════════════════════════════════════════════════════
export DEEPSEEK_API_KEY="sk-7398e54a08cf4c1ba86500a9bff10a18"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"
SESSION="quick-test"

cd /Users/frank/MareCogito/go-agent

# ═══════════════════════════════════════════════════════════════
# 1. 运行完整测试
# ═══════════════════════════════════════════════════════════════
echo "🧪 运行多 Agent 共享记忆测试..."
go run cmd/example/multi_agent_memory/main.go \
    -session "$SESSION" \
    -message "My name is Bob and I work as a software engineer"

# ═══════════════════════════════════════════════════════════════
# 2. 使用简单示例测试记忆存储
# ═══════════════════════════════════════════════════════════════
echo "📝 存储记忆..."
go run cmd/example/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
    -session "$SESSION" \
    -message "I live in Beijing and love hiking"

# ═══════════════════════════════════════════════════════════════
# 3. 测试记忆检索
# ═══════════════════════════════════════════════════════════════
echo "💬 检索记忆..."
go run cmd/example/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
    -session "$SESSION" \
    -message "Where do I live?"

# ═══════════════════════════════════════════════════════════════
# 4. 查看 PostgreSQL 中的记忆
# ═══════════════════════════════════════════════════════════════
echo "💾 查看数据库中的记忆..."
docker exec postgres-pgvector psql -U admin -d ragdb \
    -c "SELECT id, substring(content, 1, 60) as content, importance FROM memory_bank WHERE session_id = '$SESSION' ORDER BY id DESC LIMIT 5;"

# ═══════════════════════════════════════════════════════════════
# 5. 清空测试数据（可选）
# ═══════════════════════════════════════════════════════════════
# docker exec postgres-pgvector psql -U admin -d ragdb \
#     -c "DELETE FROM memory_bank WHERE session_id = '$SESSION';"
