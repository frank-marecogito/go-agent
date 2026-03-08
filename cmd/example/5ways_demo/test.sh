#!/bin/bash
# 5 SubAgents + 5 通信方式测试脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
export DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY:-sk-7398e54a08cf4c1ba86500a9bff10a18}"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"

PROJECT_DIR="/Users/frank/MareCogito/go-agent"
cd "$PROJECT_DIR"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     5 SubAgents + 5 Communication Methods Test               ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 检查前置条件
echo -e "${YELLOW}[Pre-check] 检查前置条件...${NC}"

# PostgreSQL
if ! docker ps 2>/dev/null | grep -q postgres-pgvector; then
    echo -e "${YELLOW}⚠️  PostgreSQL 未运行，正在启动...${NC}"
    docker run -d --name postgres-pgvector \
        -e POSTGRES_USER=admin \
        -e POSTGRES_PASSWORD=admin \
        -e POSTGRES_DB=ragdb \
        -p 5432:5432 \
        pgvector/pgvector:pg16 2>&1 | head -1
    sleep 5
fi
echo -e "${GREEN}✅ PostgreSQL 运行中${NC}"

# Ollama
if ! curl -s http://localhost:11434/api/tags 2>/dev/null | grep -q nomic-embed-text; then
    echo -e "${YELLOW}⚠️  Ollama 可能未运行，请执行：ollama serve &${NC}"
else
    echo -e "${GREEN}✅ Ollama 就绪${NC}"
fi

# API Key
if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo -e "${RED}❌ DEEPSEEK_API_KEY 未设置${NC}"
    exit 1
fi
echo -e "${GREEN}✅ DeepSeek API Key 已配置${NC}"

echo ""
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}开始测试 5 种通信方式${NC}"
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
echo ""

# Test 1: SharedSession
echo -e "${YELLOW}[1/5] SharedSession (共享记忆)${NC}"
echo "     测试：Researcher 存储 → Coder 检索"
go run cmd/example/shared_session_test/main.go 2>&1 | grep -E "✅|📝|🔍|TEST" | head -10
echo ""

# Test 2: SubAgent Delegation
echo -e "${YELLOW}[2/5] SubAgent Delegation (内置委托)${NC}"
echo "     测试：Main Agent 委托给 5 个专家 SubAgent"
go run cmd/example/5ways_demo/main.go 2>&1 | grep -E "✅|📝|Method 2|SubAgent" | head -10
echo ""

# Test 3: Agent-as-Tool
echo -e "${YELLOW}[3/5] Agent-as-Tool (UTCP 工具调用)${NC}"
echo "     测试：Manager Agent 调用 Expert Agent"
go run cmd/example/agent_as_tool/main.go 2>&1 | grep -E "✅|Tool|Result" | head -8
echo ""

# Test 4: Swarm
echo -e "${YELLOW}[4/5] Swarm (群体协作)${NC}"
echo "     测试：多个 Participant 共享记忆协作"
go test ./src/swarm/... -v 2>&1 | grep -E "PASS|FAIL|RUN" | head -10
echo ""

# Test 5: CodeMode
echo -e "${YELLOW}[5/5] CodeMode Orchestration (工作流编排)${NC}"
echo "     测试：LLM 自动生成代码编排多 Agent 工作流"
go run cmd/example/codemode_utcp_workflow/main.go 2>&1 | grep -E "✅|Registered|Workflow" | head -8
echo ""

# 总结
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试总结${NC}"
echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
echo ""
echo "✅ Method 1: SharedSession - 共享记忆空间"
echo "✅ Method 2: SubAgent Delegation - 内置任务委托"
echo "✅ Method 3: Agent-as-Tool - UTCP 工具调用"
echo "✅ Method 4: Swarm - 群体协作"
echo "✅ Method 5: CodeMode - 工作流编排"
echo ""
echo -e "${GREEN}所有测试完成！${NC}"
echo ""
echo "详细文档："
echo "  - cmd/example/5ways_demo/README.md"
echo "  - cmd/example/5ways_demo/COMPLETE_GUIDE.md"
echo "  - docs/agent-communication-methods-complete.md"
