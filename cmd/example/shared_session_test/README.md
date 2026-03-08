# SharedSession Test - go-agent 共享记忆测试

## 概述

测试 go-agent 的 `SharedSession` 机制是否实现了真正的跨 Agent 记忆共享。

## 运行测试

```bash
cd /Users/frank/MareCogito/go-agent

# 设置环境变量
export DEEPSEEK_API_KEY="sk-xxx"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"

# 运行测试
go run cmd/example/shared_session_test/main.go
```

## 测试内容

### Test 1: 同一 space 内的 Agent 共享记忆

```
Agent-A (space: team:alpha) → 存储 → "The project codename is Phoenix"
Agent-B (space: team:alpha) → 检索 → ✅ 找到记忆
```

### Test 2: 不同 space 之间的隔离

```
Agent-A (space: team:alpha) → 尝试检索 team:beta → ❌ 无法访问
```

### Test 3: Agent 加入多个 space

```
Agent-D 加入 [team:alpha, team:beta]
Agent-D 检索 → ✅ 找到两个 space 的记忆
```

### Test 4: PostgreSQL 持久化验证

```sql
SELECT COUNT(*) FROM memory_bank;
-- 验证记忆已持久化
```

## 核心代码

```go
// 1. 创建 SpaceRegistry 并配置 ACL
registry := session.NewSpaceRegistry(24 * time.Hour)
registry.Grant("team:alpha", "agent-A", session.SpaceRoleAdmin, 0)
registry.Grant("team:alpha", "agent-B", session.SpaceRoleWriter, 0)

// 2. 每个 Agent 的 SessionMemory 使用同一 registry
sessionMemA := memory.NewSessionMemory(bank, 16)
sessionMemA.Spaces = registry  // ← 关键

// 3. 创建 SharedSession
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

// 4. Agent-A 存储到共享 space
sharedA.StoreLongTo(ctx, "team:alpha", "fact", metadata)

// 5. Agent-B 可以从同一 space 检索
recs, _ := sharedB.Retrieve(ctx, "query", 5)
// ✅ 可以访问 Agent-A 存储的记忆！
```

## 预期输出

```
╔══════════════════════════════════════════════════════════════╗
║          SharedSession Memory Test                           ║
╚══════════════════════════════════════════════════════════════╝

TEST 1: Two agents sharing the same space
📝 Step 1: Agent-A stores memory to shared space
   ✅ Agent-A stored: 'The project codename is Phoenix'
🔍 Step 2: Agent-B retrieves from shared space
   ✅ Agent-B found 1 memories

TEST 2: Agents in different spaces (should NOT share)
🔍 Step 2: Agent-A tries to retrieve from team:beta
   ✅ Agent-A found NO memories (expected - different space)

TEST 3: Agent joins multiple spaces
🔍 Step 2: Agent-D retrieves from both spaces
   ✅ Agent-D found 2 memories from both spaces

📊 Total memories in PostgreSQL: 19
```

## 架构

```
┌─────────────────────────────────────────┐
│          SpaceRegistry (ACL)            │
│  team:alpha → [agent-A, agent-B, ...]   │
│  team:beta  → [agent-C, agent-D, ...]   │
└─────────────────┬───────────────────────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
     ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐
│Agent-A │  │Agent-B │  │Agent-D │
│✓ alpha │  │✓ alpha │  │✓ alpha │
│        │  │        │  │✓ beta  │
└────────┘  └────────┘  └────────┘
```

## 权限角色

| 角色 | 读取 | 写入 |
|------|------|------|
| `SpaceRoleReader` | ✅ | ❌ |
| `SpaceRoleWriter` | ✅ | ✅ |
| `SpaceRoleAdmin` | ✅ | ✅ |

## 相关文件

- `main.go` - 测试主程序
- `TEST_REPORT.md` - 详细测试报告

## 结论

✅ **go-agent 的 SharedSession 机制已实现真正的跨 Agent 记忆共享**

通过 `SpaceRegistry` 管理 ACL 权限，不同 Agent 可以：
- 加入同一个 space 共享记忆
- 加入多个 space 访问不同记忆
- 权限隔离保证安全性
