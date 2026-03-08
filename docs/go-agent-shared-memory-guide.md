# go-agent 共享记忆使用指南

## 概述

go-agent 通过 `SharedSession` + `SpaceRegistry` 机制实现**真正的跨 Agent 记忆共享**。

### 核心特性

| 特性 | 说明 |
|------|------|
| **跨 Agent 共享** | 不同 Agent 可以访问同一份记忆 |
| **权限控制** | Reader/Writer/Admin 三级权限 |
| **空间隔离** | 不同 space 之间完全隔离 |
| **多 space 支持** | 一个 Agent 可加入多个 space |
| **持久化** | PostgreSQL + pgvector 存储 |
| **语义搜索** | Ollama/ OpenAI Embedding |

---

## 架构设计

```
┌────────────────────────────────────────────────────────────┐
│                    SpaceRegistry                           │
│  ┌────────────────────────────────────────────────────┐    │
│  │ team:alpha                                         │    │
│  │   - agent-A: admin  |  agent-B: writer  |  ...     │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │ team:beta                                          │    │
│  │   - agent-C: admin  |  agent-D: writer  |  ...     │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────┬──────────────────────────────────────┘
                      │
         ┌────────────┼────────────┬──────────────┐
         │            │            │              │
         ▼            ▼            ▼              ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
   │ Agent-A  │ │ Agent-B  │ │ Agent-C  │ │ Agent-D  │
   │ session: │ │ session: │ │ session: │ │ session: │
   │ agent-A  │ │ agent-B  │ │ agent-C  │ │ agent-D  │
   │          │ │          │ │          │ │          │
   │ ✓ alpha  │ │ ✓ alpha  │ │ ✓ beta   │ │ ✓ alpha  │
   │          │ │          │ │          │ │ ✓ beta   │
   └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
        │            │            │            │
        └────────────┴────────────┴────────────┘
                         │
                         ▼
            ┌────────────────────────┐
            │   PostgreSQL Store     │
            │  ┌──────────────────┐  │
            │  │ session_id       │  │
            │  │ = space name     │  │
            │  │                  │  │
            │  │ team:alpha       │  │
            │  │ team:beta        │  │
            │  └──────────────────┘  │
            └────────────────────────┘
```

---

## 快速开始

### 1. 前置条件

```bash
# PostgreSQL + pgvector
docker run -d --name postgres-pgvector \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=ragdb \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Ollama Embedding
ollama pull nomic-embed-text
ollama serve &

# 环境变量
export DEEPSEEK_API_KEY="sk-xxx"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"
```

### 2. 基础示例

```go
package main

import (
    "context"
    "time"
    
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/memory/engine"
    "github.com/Protocol-Lattice/go-agent/src/memory/session"
    "github.com/Protocol-Lattice/go-agent/src/memory/store"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 PostgreSQL 存储
    pgStore, _ := store.NewPostgresStore(ctx, "postgres://admin:admin@localhost:5432/ragdb")
    
    // 2. 创建 SpaceRegistry 并配置 ACL
    registry := session.NewSpaceRegistry(24 * time.Hour)
    registry.Grant("team:alpha", "agent-A", session.SpaceRoleAdmin, 0)
    registry.Grant("team:alpha", "agent-B", session.SpaceRoleWriter, 0)
    
    // 3. 创建 SessionMemory（所有 Agent 共享同一 registry）
    bank := memory.NewMemoryBankWithStore(pgStore)
    embedder := memory.AutoEmbedder()
    opts := engine.DefaultOptions()
    
    // Agent-A
    sessionMemA := memory.NewSessionMemory(bank, 16)
    sessionMemA.WithEmbedder(embedder)
    sessionMemA.Spaces = registry  // ← 关键：设置 registry
    sessionMemA.WithEngine(engine.NewEngine(pgStore, opts))
    
    // Agent-B（使用同一个 registry）
    sessionMemB := memory.NewSessionMemory(bank, 16)
    sessionMemB.WithEmbedder(embedder)
    sessionMemB.Spaces = registry  // ← 关键：同一个 registry
    sessionMemB.WithEngine(engine.NewEngine(pgStore, opts))
    
    // 4. 创建 SharedSession
    sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
    sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")
    
    // 5. Agent-A 存储记忆到共享 space
    sharedA.StoreLongTo(ctx, "team:alpha", "Project codename is Phoenix", nil)
    sharedA.FlushSpace(ctx, "team:alpha")
    
    // 6. Agent-B 检索（可以访问 Agent-A 存储的记忆！）
    recs, _ := sharedB.Retrieve(ctx, "project codename", 5)
    for _, r := range recs {
        println(r.Content)  // 输出：Project codename is Phoenix
    }
}
```

---

## 核心组件

### 1. SpaceRegistry（空间注册表）

管理所有共享空间及其 ACL 权限。

```go
// 创建注册表
registry := session.NewSpaceRegistry(24 * time.Hour)  // 默认 TTL 24 小时

// 授予权限
registry.Grant(spaceName, principal, role, ttl)

// 权限角色
session.SpaceRoleReader  // 只读
session.SpaceRoleWriter  // 读写
session.SpaceRoleAdmin   // 管理员
```

### 2. SharedSession（共享会话）

让多个 Agent 共享短期和长期记忆。

```go
// 创建共享会话
shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")

// 加入多个 space
shared.Join("team:beta")
shared.Join("guild:search")

// 离开 space
shared.Leave("team:beta")

// 获取所有 space
spaces := shared.Spaces()  // ["team:alpha", "team:beta"]
```

### 3. SessionMemory（会话记忆）

每个 Agent 的记忆实例，需要设置 `Spaces` 字段。

```go
sessionMem := memory.NewSessionMemory(bank, 16)
sessionMem.Spaces = registry  // ← 必须设置！
```

---

## 使用场景

### 场景 1: 团队协作 Agent

```go
// 配置：所有团队成员共享同一记忆空间
registry := session.NewSpaceRegistry(0)  // 0 = 永不过期
registry.Grant("team:project-x", "agent-1", session.SpaceRoleAdmin, 0)
registry.Grant("team:project-x", "agent-2", session.SpaceRoleWriter, 0)
registry.Grant("team:project-x", "agent-3", session.SpaceRoleWriter, 0)

// Agent-1 存储项目信息
shared1 := session.NewSharedSession(sessionMem1, "agent-1", "team:project-x")
shared1.StoreLongTo(ctx, "team:project-x", "PRD v1 approved", nil)

// Agent-2 可以访问
shared2 := session.NewSharedSession(sessionMem2, "agent-2", "team:project-x")
recs, _ := shared2.Retrieve(ctx, "PRD", 5)  // ✅ 可以找到
```

### 场景 2: 跨 Agent 知识共享

```go
// 专家 Agent 存储知识
registry.Grant("knowledge:domain", "expert-agent", session.SpaceRoleAdmin, 0)
registry.Grant("knowledge:domain", "novice-agent", session.SpaceRoleReader, 0)

// 专家存储
expertShared.StoreLongTo(ctx, "knowledge:domain", "API best practices...", nil)

// 新手学习（只读）
noviceShared := session.NewSharedSession(sessionMem, "novice-agent", "knowledge:domain")
recs, _ := noviceShared.Retrieve(ctx, "API practices", 5)  // ✅ 可以检索
```

### 场景 3: 多租户隔离

```go
// 租户 A
registry.Grant("tenant:a", "agent-a1", session.SpaceRoleAdmin, 0)
registry.Grant("tenant:a", "agent-a2", session.SpaceRoleWriter, 0)

// 租户 B（完全隔离）
registry.Grant("tenant:b", "agent-b1", session.SpaceRoleAdmin, 0)

// Agent-A1 无法访问 tenant:b 的记忆
agentA1Shared.Retrieve(ctx, "query", 5)  // 只能访问 tenant:a
```

### 场景 4: Agent 加入多个空间

```go
// Agent-D 同时参与两个项目
registry.Grant("team:alpha", "agent-D", session.SpaceRoleWriter, 0)
registry.Grant("team:beta", "agent-D", session.SpaceRoleWriter, 0)

sharedD := session.NewSharedSession(sessionMem, "agent-D", "team:alpha")
sharedD.Join("team:beta")  // 加入第二个 space

// 可以访问两个 space 的记忆
recs, _ := sharedD.Retrieve(ctx, "project info", 10)
// 返回 team:alpha 和 team:beta 的记忆
```

---

## API 参考

### 存储记忆

```go
// 存储到指定 space
rec, err := shared.StoreLongTo(ctx, "team:alpha", "content", metadata)

// 广播到所有 space
recs, err := shared.BroadcastLong(ctx, "content", metadata)

// 存储到短期记忆
shared.AddShortLocal("content", map[string]string{"key": "value"})
shared.AddShortTo("team:alpha", "content", metadata)
```

### 检索记忆

```go
// 从所有可访问的 space 检索
recs, err := shared.Retrieve(ctx, "query", limit)

// 只从共享 space 检索（不包括本地）
recs, err := shared.RetrieveShared(ctx, "query", limit)
```

### 持久化

```go
// 刷新本地短期记忆到长期存储
err := shared.FlushLocal(ctx)

// 刷新指定 space 的短期记忆
err := shared.FlushSpace(ctx, "team:alpha")
```

### 权限管理

```go
// 检查权限
canRead := registry.CanRead("team:alpha", "agent-A")
canWrite := registry.CanWrite("team:alpha", "agent-A")

// 授予/撤销权限
registry.Grant("team:alpha", "agent-X", session.SpaceRoleWriter, 0)
registry.Revoke("team:alpha", "agent-X")

// 列出可访问的 space
spaces := registry.List("agent-A")
```

---

## 权限说明

| 角色 | 读取 | 写入 | 说明 |
|------|------|------|------|
| `SpaceRoleReader` | ✅ | ❌ | 只能检索记忆 |
| `SpaceRoleWriter` | ✅ | ✅ | 可以存储和检索 |
| `SpaceRoleAdmin` | ✅ | ✅ | 管理员（语义相同，可扩展） |

---

## 内存结构

### PostgreSQL 表结构

```sql
-- memory_bank 表
CREATE TABLE memory_bank (
    id              SERIAL PRIMARY KEY,
    session_id      TEXT NOT NULL,      -- = space 名称
    content         TEXT NOT NULL,
    metadata        JSONB,
    embedding       VECTOR(768),
    importance      FLOAT8,
    source          TEXT,
    summary         TEXT,
    created_at      TIMESTAMPTZ,
    last_embedded   TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_memory_session ON memory_bank(session_id);
CREATE INDEX idx_memory_embedding ON memory_bank USING hnsw(embedding vector_cosine_ops);
```

### 记忆存储位置

```
session_id = space 名称

team:alpha → 存储所有 team:alpha 的记忆
team:beta  → 存储所有 team:beta 的记忆
```

---

## 注意事项

### 1. 必须设置 SpaceRegistry

```go
// ❌ 错误：没有设置 registry，访问会被拒绝
sessionMem := memory.NewSessionMemory(bank, 16)
shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")
shared.StoreLongTo(...)  // Error: space access denied

// ✅ 正确：设置 registry 并配置 ACL
registry.Grant("team:alpha", "agent-A", session.SpaceRoleWriter, 0)
sessionMem.Spaces = registry
```

### 2. 所有 Agent 必须共享同一 Registry 实例

```go
// ❌ 错误：每个 Agent 使用不同的 registry
registryA := session.NewSpaceRegistry(0)
registryB := session.NewSpaceRegistry(0)
sessionMemA.Spaces = registryA
sessionMemB.Spaces = registryB
// Agent-A 和 Agent-B 无法共享记忆！

// ✅ 正确：共享同一 registry 实例
registry := session.NewSpaceRegistry(0)
sessionMemA.Spaces = registry
sessionMemB.Spaces = registry
```

### 3. sessionID vs space

| 概念 | 说明 |
|------|------|
| **sessionID** | 单个 Agent 的会话标识（如 "agent-A"） |
| **space** | 多个 Agent 共享的记忆空间（如 "team:alpha"） |
| **存储位置** | 记忆存储在 `session_id = space 名称` 下 |

### 4. 权限检查时机

```go
// StoreLongTo 会检查写权限
shared.StoreLongTo(ctx, "team:alpha", "content", nil)
// ↓
// registry.Check("team:alpha", "agent-A", requireWrite=true)

// Retrieve 会检查读权限
shared.Retrieve(ctx, "query", 5)
// ↓
// registry.CanRead(space, principal)
```

---

## 测试

```bash
# 运行共享记忆测试
cd /Users/frank/MareCogito/go-agent
./cmd/example/shared_session_test/test.sh

# 或手动运行
go run cmd/example/shared_session_test/main.go
```

### 测试内容

| 测试 | 说明 |
|------|------|
| Test 1 | 同一 space 内的 Agent 共享记忆 |
| Test 2 | 不同 space 之间的隔离 |
| Test 3 | Agent 加入多个 space |
| Test 4 | PostgreSQL 持久化验证 |

---

## 故障排查

### 问题 1: `space access denied`

**原因**: 没有配置 ACL 或 principal 不在 ACL 中

```go
// 解决方案
registry.Grant("team:alpha", "agent-A", session.SpaceRoleWriter, 0)
sessionMem.Spaces = registry
```

### 问题 2: Agent 无法访问其他 Agent 的记忆

**检查清单**:
1. 所有 Agent 使用同一个 `registry` 实例？
2. 所有 Agent 加入同一个 space？
3. `sessionMem.Spaces = registry` 已设置？
4. 记忆已 `FlushSpace` 到 PostgreSQL？

### 问题 3: 检索结果为空

**可能原因**:
1. 记忆未持久化（调用 `FlushSpace`）
2. Embedding 失败（检查 Ollama 服务）
3. 权限不足（检查 ACL 配置）

---

## 相关文件

```
go-agent/src/memory/
├── session/
│   ├── shared_session.go    # SharedSession 实现
│   ├── spaces.go            # SpaceRegistry 实现
│   └── memory_bank.go       # MemoryBank 实现
├── engine/
│   └── engine.go            # 记忆引擎
└── store/
    └── postgres_store.go    # PostgreSQL 存储

cmd/example/
├── shared_session_test/     # 共享记忆测试
└── multi_agent_memory/      # 多 Agent 测试
```

---

## 总结

go-agent 的共享记忆机制通过以下组件实现：

1. **SpaceRegistry** - 管理空间和 ACL 权限
2. **SharedSession** - 让多个 Agent 共享记忆
3. **SessionMemory** - 每个 Agent 的记忆实例
4. **PostgreSQL Store** - 持久化存储

**关键点**:
- 所有 Agent 必须共享同一个 `SpaceRegistry` 实例
- 必须设置 `sessionMem.Spaces = registry`
- 记忆存储在 `session_id = space 名称` 下
- 权限控制通过 Reader/Writer/Admin 实现

---

## 个人记忆 vs 共享记忆

### Agent 可以同时拥有两种记忆

```go
shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")

// 个人记忆（私有）
shared.StoreLongTo(ctx, "agent-A", "My birthday is March 15th", nil)
shared.AddShortLocal("I prefer coffee", metadata)

// 共享记忆（team:alpha 成员可访问）
shared.StoreLongTo(ctx, "team:alpha", "Project deadline", nil)

// 检索：同时返回个人和共享记忆
recs, _ := shared.Retrieve(ctx, "query", 10)
```

### 记忆访问权限

| 记忆类型 | 存储位置 | Agent-A | Agent-B（同 team） |
|----------|----------|---------|-------------------|
| 个人记忆 | `session_id = "agent-A"` | ✅ 可访问 | ❌ 无法访问 |
| 共享记忆 | `session_id = "team:alpha"` | ✅ 可访问 | ✅ 可访问 |

### 测试结果

```
🔍 Found 4 memories:
   🌐 [SHARED] Project deadline (session: team:alpha)
   👤 [PERSONAL] My birthday is March 15th (session: agent-A)
   🌐 [SHARED] Phoenix (session: team:alpha)
   👤 [PERSONAL] I prefer coffee (session: agent-A)
   → Personal: 2, Shared: 2

✅ Agent-B CANNOT access Agent-A's personal memory
```
