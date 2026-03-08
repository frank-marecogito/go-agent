# go-agent Agent 间通信方式对比

## 问题：不使用 SharedSession，Agent A 可以发信息给 Agent B 吗？

**答案**：可以，但需要其他机制。SharedSession 是**共享记忆**，不是唯一的通信方式。

---

## 通信方式对比

| 方式 | 是否需要 SharedSession | 记忆共享 | 实时通信 | 适用场景 |
|------|----------------------|----------|----------|----------|
| **SharedSession** | ✅ 需要 | ✅ 共享记忆 | ❌ 间接 | 多 Agent 共享历史记忆 |
| **Agent-as-Tool** | ❌ 不需要 | ❌ 独立记忆 | ✅ 直接调用 | 任务委托、专业分工 |
| **Swarm** | ❌ 不需要 | ❌ 独立记忆 | ✅ 任务流转 | 复杂工作流 |
| **外部存储** | ❌ 不需要 | ⚠️ 手动同步 | ❌ 间接 | 持久化中转 |
| **直接访问记忆** | ❌ **不行** | ❌ 会话隔离 | ❌ 无法实现 | - |

---

## 方式 1: Agent-as-Tool（推荐）

### 原理

将 Agent B 注册为 UTCP 工具，Agent A 通过调用工具与 Agent B 通信。

### 架构图

```
┌─────────────┐      UTCP       ┌─────────────┐
│  Agent-A    │ ←─────────────→ │  Agent-B    │
│  (Manager)  │   CallTool()    │ (Researcher)│
│             │                 │             │
│ Memory: A   │                 │ Memory: B   │
│ (独立)      │                 │ (独立)      │
└─────────────┘                 └─────────────┘
```

### 代码示例

```go
// 1. 创建 Agent-B（作为工具提供者）
researcher, _ := agent.New(agent.Options{
    Model:        llm,
    Memory:       sessionMemB,  // 独立记忆
    SystemPrompt: "You are a researcher.",
})

// 2. 注册为 UTCP 工具
client, _ := utcp.NewUTCPClient(ctx, ...)
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "Performs research")

// 3. 创建 Agent-A（可以调用工具）
manager, _ := agent.New(agent.Options{
    Model:        llm,
    Memory:       sessionMemA,  // 独立记忆
    SystemPrompt: "You are a manager. Delegate tasks.",
    Tools:        client.Tools(),  // ← 可以调用 Agent-B
})

// 4. Agent-A 调用 Agent-B
// 用户问："Find facts about the sky"
// Agent-A 输出工具调用 → agent.researcher
// Agent-B 执行 → "The sky is blue because of Rayleigh scattering"
// 结果返回给 Agent-A
```

### 特点

| 优点 | 缺点 |
|------|------|
| ✅ Agent 间记忆隔离 | ❌ 需要 UTCP 客户端 |
| ✅ 实时通信 | ❌ 需要 LLM 支持工具调用 |
| ✅ 专业分工 | ❌ 配置较复杂 |

---

## 方式 2: Swarm（任务委托）

### 原理

Agent A 将任务委托给 Agent B，Agent B 完成后返回结果。

### 架构图

```
┌─────────────┐    Delegate    ┌─────────────┐
│  Agent-A    │ ─────────────→ │  Agent-B    │
│ (Coordinator)│   Task        │ (Specialist)│
│             │                │             │
│             │ ←───────────── │             │
│             │    Result      │             │
└─────────────┘                └─────────────┘
```

### 代码示例

```go
// 1. 创建多个 Agent
coordinator, _ := agent.New(agent.Options{
    Model: llm,
    SystemPrompt: "Coordinate tasks.",
})

specialist, _ := agent.New(agent.Options{
    Model: llm,
    SystemPrompt: "You are a coding specialist.",
})

// 2. 创建 Swarm
swarm := agent.NewSwarm(
    coordinator,
    specialist,
)

// 3. 执行任务
// 用户问："Write a function to sort numbers"
// coordinator → 委托给 specialist
// specialist → 生成代码
// 结果返回给用户
```

### 特点

| 优点 | 缺点 |
|------|------|
| ✅ 任务自动流转 | ❌ 需要 Swarm 支持 |
| ✅ 专业分工 | ❌ Agent 间记忆不共享 |

---

## 方式 3: 外部存储中转

### 原理

通过外部数据库/消息队列作为中介，Agent A 写入，Agent B 读取。

### 架构图

```
┌─────────────┐     Write      ┌─────────────┐
│  Agent-A    │ ─────────────→ │   External  │
│             │                │   Storage   │
│             │                │  (PostgreSQL│
│             │ ←───────────── │   /Redis/   │
│  Agent-B    │     Read       │    Kafka)   │
└─────────────┘                └─────────────┘
```

### 代码示例

```go
// Agent-A 写入消息
func (a *AgentA) SendMessage(to, content string) error {
    msg := Message{
        From: "agent-A",
        To: "agent-B",
        Content: content,
        Time: time.Now(),
    }
    return db.Insert("messages", msg)
}

// Agent-B 读取消息
func (b *AgentB) CheckMessages() ([]Message, error) {
    return db.Query("SELECT * FROM messages WHERE to = 'agent-B'")
}
```

### 特点

| 优点 | 缺点 |
|------|------|
| ✅ 完全解耦 | ❌ 需要额外基础设施 |
| ✅ 异步通信 | ❌ 需要手动同步 |
| ✅ 持久化 | ❌ 实时性差 |

---

## 方式 4: SharedSession（共享记忆）

### 原理

Agent A 和 Agent B 共享同一个记忆空间，通过记忆间接通信。

### 架构图

```
┌─────────────┐                ┌─────────────┐
│  Agent-A    │                │  Agent-B    │
│             │                │             │
│  ┌───────┐  │                │  ┌───────┐  │
│  │Shared │  │                │  │Shared │  │
│  │Session│  │                │  │Session│  │
│  │ α     │  │                │  │ α     │  │
│  └───┬───┘  │                │  └───┬───┘  │
└──────┼──────┘                └──────┼──────┘
       │                              │
       └──────────────┬───────────────┘
                      │
                      ▼
           ┌─────────────────────┐
           │  PostgreSQL Store   │
           │  session_id = α     │
           │  (共享记忆)         │
           └─────────────────────┘
```

### 代码示例

```go
// 1. 配置共享空间
registry := session.NewSpaceRegistry(0)
registry.Grant("team:alpha", "agent-A", SpaceRoleWriter, 0)
registry.Grant("team:alpha", "agent-B", SpaceRoleWriter, 0)

// 2. Agent-A 写入共享记忆
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedA.StoreLongTo(ctx, "team:alpha", "Meeting at 3pm", nil)

// 3. Agent-B 读取共享记忆
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")
recs, _ := sharedB.Retrieve(ctx, "meeting", 5)
// 找到："Meeting at 3pm"
```

### 特点

| 优点 | 缺点 |
|------|------|
| ✅ 记忆自动共享 | ❌ 不是实时通信 |
| ✅ 无需额外配置 | ❌ 间接通信 |
| ✅ 持久化 | ❌ 适合历史记忆，不适合即时消息 |

---

## 完整对比表

| 特性 | SharedSession | Agent-as-Tool | Swarm | 外部存储 |
|------|---------------|---------------|-------|----------|
| **记忆共享** | ✅ 自动 | ❌ 独立 | ❌ 独立 | ⚠️ 手动 |
| **实时通信** | ❌ 间接 | ✅ 直接 | ✅ 直接 | ⚠️ 异步 |
| **配置复杂度** | 低 | 中 | 中 | 高 |
| **持久化** | ✅ 自动 | ❌ 不持久 | ❌ 不持久 | ✅ 自动 |
| **适用场景** | 共享历史 | 任务委托 | 工作流 | 异步消息 |

---

## 使用建议

### 场景 1: 共享历史记忆
**推荐**: SharedSession

```
Agent-A 和 Agent-B 需要访问同一份用户历史记忆
→ 使用 SharedSession 共享 team:alpha space
```

### 场景 2: 任务委托
**推荐**: Agent-as-Tool 或 Swarm

```
Agent-A（Manager）需要 Agent-B（Researcher）做研究
→ 将 Agent-B 注册为工具，Agent-A 调用
```

### 场景 3: 异步消息
**推荐**: 外部存储

```
Agent-A 需要发送消息给 Agent-B（可能离线）
→ 写入消息队列，Agent-B 稍后读取
```

### 场景 4: 混合使用
**推荐**: SharedSession + Agent-as-Tool

```
- 共享记忆：SharedSession
- 实时通信：Agent-as-Tool
```

---

## 总结

| 问题 | 答案 |
|------|------|
| 不使用 SharedSession，Agent A 可以直接访问 Agent B 的记忆吗？ | ❌ **不可以**（会话隔离） |
| 不使用 SharedSession，Agent A 可以和 Agent B 通信吗？ | ✅ **可以**（通过其他方式） |
| 推荐方式 | Agent-as-Tool（实时）+ SharedSession（记忆共享） |

### 最佳实践

```go
// 1. 使用 SharedSession 共享记忆
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

// 2. 使用 Agent-as-Tool 实时通信
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "...")
manager.Tools = client.Tools()

// 3. Agent-A 可以：
//    - 访问共享记忆（通过 SharedSession）
//    - 调用 Agent-B（通过 UTCP）
```
