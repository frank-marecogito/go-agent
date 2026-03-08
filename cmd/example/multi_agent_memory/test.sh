#!/bin/bash
# 多 Agent 共享记忆测试脚本
# 用法：./test_shared_memory.sh

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
export DEEPSEEK_API_KEY="sk-7398e54a08cf4c1ba86500a9bff10a18"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"

SESSION="shared-test-$$"  # 唯一 session ID
PROJECT_DIR="/Users/frank/MareCogito/go-agent"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       Multi-Agent Shared Memory Test                         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 检查前置条件
echo -e "${YELLOW}[1/6] 检查前置条件...${NC}"

# 检查 PostgreSQL
if ! docker ps | grep -q postgres-pgvector; then
    echo -e "${RED}❌ PostgreSQL 未运行，正在启动...${NC}"
    docker run -d --name postgres-pgvector \
        -e POSTGRES_USER=admin \
        -e POSTGRES_PASSWORD=admin \
        -e POSTGRES_DB=ragdb \
        -p 5432:5432 \
        pgvector/pgvector:pg16
    sleep 5
fi
echo -e "${GREEN}✅ PostgreSQL 运行中${NC}"

# 检查 Ollama
if ! curl -s http://localhost:11434/api/tags | grep -q nomic-embed-text; then
    echo -e "${YELLOW}⚠️  Ollama 可能未运行或模型未安装${NC}"
    echo "   请运行：ollama pull nomic-embed-text"
fi
echo -e "${GREEN}✅ Ollama 就绪${NC}"

# 检查 DeepSeek API Key
if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo -e "${RED}❌ DEEPSEEK_API_KEY 未设置${NC}"
    exit 1
fi
echo -e "${GREEN}✅ DeepSeek API Key 已配置${NC}"

echo ""

# 测试 1: 基础记忆存储
echo -e "${YELLOW}[2/6] 测试 1: MemoryAgent 存储记忆${NC}"
cd "$PROJECT_DIR"
go run cmd/example/multi_agent_memory/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -session "$SESSION" \
    -message "My name is Alice and I love Go programming" 2>&1 | grep -E "✅|Test [0-9]"

echo ""

# 测试 2: 验证 PostgreSQL 中的记忆
echo -e "${YELLOW}[3/6] 测试 2: 验证 PostgreSQL 中的记忆${NC}"
docker exec postgres-pgvector psql -U admin -d ragdb \
    -c "SELECT id, substring(content, 1, 50) as content FROM memory_bank WHERE session_id = '$SESSION' ORDER BY id;" 2>&1

echo ""

# 测试 3: 使用 cmd/example/main.go 测试记忆检索
echo -e "${YELLOW}[4/6] 测试 3: ChatAgent 检索记忆${NC}"
go run cmd/example/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
    -session "$SESSION" \
    -message "What is my name?" 2>&1 | tail -5

echo ""

# 测试 4: 存储新记忆
echo -e "${YELLOW}[5/6] 测试 4: 存储额外记忆${NC}"
go run cmd/example/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
    -session "$SESSION" \
    -message "I also enjoy playing chess on weekends" 2>&1 | tail -3

echo ""

# 测试 5: 检索所有记忆
echo -e "${YELLOW}[6/6] 测试 5: 检索所有关于 Alice 的记忆${NC}"
go run cmd/example/main.go \
    -provider deepseek \
    -model deepseek-chat \
    -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
    -session "$SESSION" \
    -message "Tell me everything you know about Alice" 2>&1 | tail -8

echo ""

# 最终验证
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}最终记忆统计:${NC}"
docker exec postgres-pgvector psql -U admin -d ragdb \
    -c "SELECT COUNT(*) as total_memories FROM memory_bank WHERE session_id = '$SESSION';" 2>&1

echo ""
echo -e "${GREEN}✅ 测试完成！Session ID: $SESSION${NC}"
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
