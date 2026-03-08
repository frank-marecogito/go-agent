# go-agent SharedSession 测试报告

## 测试日期
2026 年 3 月 2 日

## 测试目标
验证 go-agent 的 `SharedSession` 机制是否实现了真正的跨 Agent 记忆共享。

---

## 核心发现

### ✅ go-agent **已实现共享记忆**

通过 `SharedSession` + `SpaceRegistry` 实现：

```go
// 1. 创建 SpaceRegistry 并配置 ACL
registry := session.NewSpaceRegistry(24 * time.Hour)
registry.Grant("team:alpha", "agent-A", session.SpaceRoleAdmin, 0)
registry.Grant("team:alpha", "agent-B", session.SpaceRoleWriter, 0)

// 2. 每个 Agent 的 SessionMemory 使用同一 registry
sessionMemA.Spaces = registry
sessionMemB.Spaces = registry

// 3. Agent 加入共享 space
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

// 4. Agent-A 存储到共享 space
sharedA.StoreLongTo(ctx, "team:alpha", "fact", metadata)

// 5. Agent-B 可以从同一 space 检索
recs, _ := sharedB.Retrieve(ctx, "fact", 5)
// ✅ 可以访问 Agent-A 存储的记忆！
```

---

## 测试结果

### Test 1: 同一 space 内的 Agent 共享记忆

```
Agent-A (session: agent-A, space: team:alpha) → 存储 → team:alpha
Agent-B (session: agent-B, space: team:alpha) → 检索 → team:alpha
```

**结果**: ✅ **通过**
- Agent-A 存储："The project codename is Phoenix"
- Agent-B 检索到：1 条记忆
- 记忆内容匹配

---

### Test 2: 不同 space 之间的隔离

```
Agent-A (space: team:alpha) → 尝试检索 → team:beta (应该失败)
```

**结果**: ✅ **通过**
- Agent-A 无法访问 team:beta 的记忆
- 只能访问自己的 team:alpha 记忆
- 空间隔离有效

---

### Test 3: Agent 加入多个 space

```
Agent-D 加入：[team:alpha, team:beta]
Agent-D 检索 → 应该找到两个 space 的记忆
```

**结果**: ✅ **通过**
- Agent-D 找到 2 条记忆：
  1. "The project codename is Phoenix" (team:alpha)
  2. "Secret password is BlueDragon" (team:beta)
- 多 space 访问有效

---

### Test 4: PostgreSQL 持久化验证

```sql
SELECT COUNT(*) FROM memory_bank;
-- 结果：19 条记忆
```

**结果**: ✅ **通过**
- 所有记忆正确持久化
- session_id = space 名称（如 "team:alpha"）

---

## 架构图

```
┌────────────────────────────────────────────────────────────┐
│                    SpaceRegistry                           │
│  ┌────────────────────────────────────────────────────┐    │
│  │ team:alpha                                         │    │
│  │   - agent-A: admin                                 │    │
│  │   - agent-B: writer                                │    │
│  │   - agent-D: writer                                │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │ team:beta                                          │    │
│  │   - agent-C: admin                                 │    │
│  │   - agent-D: writer                                │    │
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
        │  │ team:alpha       │  │
        │  │ - Phoenix        │  │
        │  └──────────────────┘  │
        │  ┌──────────────────┐  │
        │  │ team:beta        │  │
        │  │ - BlueDragon     │  │
        │  └──────────────────┘  │
        └────────────────────────┘
```

---

## 关键 API

### 1. 创建 SpaceRegistry

```go
registry := session.NewSpaceRegistry(24 * time.Hour)
registry.Grant("team:alpha", "agent-A", session.SpaceRoleAdmin, 0)
```

### 2. 配置 SessionMemory

```go
sessionMem := memory.NewSessionMemory(bank, 16)
sessionMem.Spaces = registry  // ← 关键：设置 registry
```

### 3. 创建 SharedSession

```go
shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")
```

### 4. 存储到共享 space

```go
_, err := shared.StoreLongTo(ctx, "team:alpha", "fact", metadata)
```

### 5. 从共享 space 检索

```go
recs, err := shared.Retrieve(ctx, "query", 5)
```

### 6. 加入多个 space

```go
shared.Join("team:beta")  // Agent 可以访问多个 space
```

---

## 权限角色

| 角色 | 读取 | 写入 | 说明 |
|------|------|------|------|
| `SpaceRoleReader` | ✅ | ❌ | 只读 |
| `SpaceRoleWriter` | ✅ | ✅ | 读写 |
| `SpaceRoleAdmin` | ✅ | ✅ | 管理员 |

---

## 使用场景

### 1. 团队协作 Agent

```go
// 所有团队成员共享同一记忆空间
registry.Grant("team:project-x", "agent-1", SpaceRoleWriter, 0)
registry.Grant("team:project-x", "agent-2", SpaceRoleWriter, 0)
registry.Grant("team:project-x", "agent-3", SpaceRoleWriter, 0)
```

### 2. 跨 Agent 知识共享

```go
// 专家 Agent 存储知识到共享 space
expert.StoreLongTo(ctx, "knowledge:domain", "expert fact", nil)

// 新手 Agent 可以从同一 space 学习
novice.Retrieve(ctx, "domain knowledge", 5)
```

### 3. 多租户隔离

```go
// 租户 A 的记忆
registry.Grant("tenant:a", "agent-a1", SpaceRoleAdmin, 0)

// 租户 B 的记忆（完全隔离）
registry.Grant("tenant:b", "agent-b1", SpaceRoleAdmin, 0)
```

---

## 注意事项

### 1. 必须配置 SpaceRegistry

```go
// ❌ 错误：没有配置 registry，访问会被拒绝
shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")

// ✅ 正确：配置 registry 并设置 ACL
registry.Grant("team:alpha", "agent-A", SpaceRoleWriter, 0)
sessionMem.Spaces = registry
```

### 2. sessionID vs space

- **sessionID**: 单个 agent 的会话标识
- **space**: 多个 agent 共享的记忆空间
- 记忆存储在 `session_id = space 名称` 下

### 3. 权限检查

```go
// 写入前检查权限
if !shared.canWrite("team:alpha") {
    return ErrSpaceForbidden
}
```

---

## 结论

### ✅ go-agent 的共享记忆已完整实现

| 特性 | 状态 | 说明 |
|------|------|------|
| 跨 Agent 记忆共享 | ✅ | 通过 SharedSession + SpaceRegistry |
| 权限控制 | ✅ | Reader/Writer/Admin 角色 |
| 空间隔离 | ✅ | 不同 space 互相隔离 |
| 多 space 访问 | ✅ | Agent 可加入多个 space |
| 持久化 | ✅ | PostgreSQL + pgvector |
| 语义搜索 | ✅ | Ollama embedding |

### 与之前测试的区别

| 测试类型 | 机制 | 共享效果 |
|----------|------|----------|
| 相同 session ID | 会话内连续 | ❌ 不是真正的共享 |
| SharedSession + space | 跨 Agent 共享 | ✅ **真正的共享记忆** |

---

*测试完成时间：2026 年 3 月 2 日*
