# Multi-Agent Shared Memory Test

演示多个 Agent 通过 PostgreSQL + pgvector 共享记忆的功能。

## 架构

```
┌─────────────────────────────────────────────────────┐
│                  PostgreSQL + pgvector              │
│                  (共享记忆存储)                      │
└─────────────────────┬───────────────────────────────┘
                      │
         ┌────────────┼────────────┐
         │            │            │
         ▼            ▼            ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │ Memory   │ │  Chat    │ │ Research │
   │  Agent   │ │  Agent   │ │  Agent   │
   └──────────┘ └──────────┘ └──────────┘
   存储事实       对话检索       分析模式
```

## 前置条件

1. **PostgreSQL + pgvector**
   ```bash
   docker run -d --name postgres-pgvector \
     -e POSTGRES_USER=admin \
     -e POSTGRES_PASSWORD=admin \
     -e POSTGRES_DB=ragdb \
     -p 5432:5432 \
     pgvector/pgvector:pg16
   ```

2. **Ollama Embedding**
   ```bash
   ollama pull nomic-embed-text
   ollama serve &
   ```

3. **DeepSeek API Key**
   ```bash
   export DEEPSEEK_API_KEY="sk-xxx"
   ```

## 运行测试

```bash
cd /Users/frank/MareCogito/go-agent

go run cmd/example/multi_agent_memory/main.go \
  -provider deepseek \
  -model deepseek-chat \
  -session "test-session" \
  -message "My name is Alice and I love Go programming"
```

## 测试序列

| 测试 | 说明 | 预期结果 |
|------|------|----------|
| Test 1 | MemoryAgent 存储事实 | 记忆存入 PostgreSQL |
| Test 2 | ChatAgent 检索记忆 | 回答 "Your name is Alice" |
| Test 3 | ResearchAgent 分析 | 报告记忆数量 |
| Test 4 | PostgreSQL 验证 | 直接查询数据库 |
| Test 5 | 跨 Agent 共享 | ChatAgent 访问 MemoryAgent 的记忆 |
| Test 6 | 多事实存储 | ResearchAgent 存储新事实 |
| Test 7 | 综合检索 | 检索所有相关记忆 |

## 核心代码

```go
// 1. 创建共享 PostgreSQL 存储
pgStore, _ := store.NewPostgresStore(ctx, connStr)
sharedBank := memory.NewMemoryBankWithStore(pgStore)

// 2. 所有 Agent 使用同一个 sharedBank
memoryAgent := NewMemoryAgent(sharedBank, embedder, &opts)
chatAgent := NewChatAgent(provider, model, sharedBank, embedder, &opts)
researchAgent := NewResearchAgent(sharedBank, embedder, &opts)

// 3. 每个 Agent 都可以存储和检索记忆
memoryAgent.Store(ctx, sessionID, "fact")
chatAgent.Chat(ctx, sessionID, "query")  // 可以访问 memoryAgent 存储的记忆
```

## 验证记忆

```bash
# 直接查询 PostgreSQL
docker exec postgres-pgvector psql -U admin -d ragdb \
  -c "SELECT id, session_id, content, importance FROM memory_bank;"
```

## 关键配置

| 环境变量 | 值 | 说明 |
|----------|-----|------|
| `DEEPSEEK_API_KEY` | `sk-xxx` | DeepSeek API 密钥 |
| `ADK_EMBED_PROVIDER` | `ollama` | Embedding 提供商 |
| `ADK_EMBED_MODEL` | `nomic-embed-text` | Embedding 模型 |

## 输出示例

```
╔══════════════════════════════════════════════════════════════╗
║       Multi-Agent Shared Memory Test                         ║
║       PostgreSQL + pgvector + Ollama Embedding               ║
╚══════════════════════════════════════════════════════════════╝

📝 Test 1: MemoryAgent storing fact
   ✅ Memory stored: My name is Alice and I love Go programming

💬 Test 2: ChatAgent retrieving memory
   ✅ ChatAgent: Your name is Alice.

🔄 Test 5: Cross-agent memory sharing
   ✅ ChatAgent: Alice loves Go programming.

📊 Final memory count in PostgreSQL:
   Total memories: 17
```

## 结论

✅ **多 Agent 共享记忆有效**

- 所有 Agent 使用同一个 `sharedBank`
- 记忆持久化到 PostgreSQL + pgvector
- 语义搜索通过 Ollama embedding 实现
- 跨 Agent 可以访问彼此存储的记忆
