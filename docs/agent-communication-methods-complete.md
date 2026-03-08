# go-agent Agent 间通信与协作方式完整指南

## 概述

go-agent 实现了 **5 种** Agent 间通信与协作方式，每种适用于不同场景。

---

## 方式总览

| 方式 | 实现位置 | 记忆共享 | 实时通信 | 复杂度 |
|------|----------|----------|----------|--------|
| **1. SharedSession** | `src/memory/session/` | ✅ 自动共享 | ❌ 间接 | 低 |
| **2. SubAgent 委托** | `agent.go`, `catalog.go` | ❌ 独立 | ✅ 直接 | 低 |
| **3. Agent-as-Tool** | `agent_tool.go` | ❌ 独立 | ✅ 直接 | 中 |
| **4. Swarm** | `src/swarm/` | ⚠️ 可选 | ✅ 任务流转 | 中 |
| **5. CodeMode 编排** | `agent_orchestrators.go` | ❌ 独立 | ✅ 工作流 | 高 |

---

## 方式 1: SharedSession（共享记忆）

### 原理

多个 Agent 通过 `SpaceRegistry` 加入同一个记忆空间，实现记忆共享。

### 架构图

```
┌─────────────┐                ┌─────────────┐
│  Agent-A    │                │  Agent-B    │
│  session: A │                │  session: B │
│      ✓ α    │                │      ✓ α    │
└──────┬──────┘                └──────┬──────┘
       │                              │
       └──────────────┬───────────────┘
                      │
                      ▼
           ┌─────────────────────┐
           │  PostgreSQL Store   │
           │  session_id = α     │
           └─────────────────────┘
```

### 核心代码

```go
// 1. 配置 SpaceRegistry
registry := session.NewSpaceRegistry(24 * time.Hour)
registry.Grant("team:alpha", "agent-A", session.SpaceRoleWriter, 0)
registry.Grant("team:alpha", "agent-B", session.SpaceRoleWriter, 0)

// 2. 创建 SessionMemory（共享同一 registry）
sessionMemA := memory.NewSessionMemory(bank, 16)
sessionMemA.Spaces = registry
sessionMemB := memory.NewSessionMemory(bank, 16)
sessionMemB.Spaces = registry

// 3. 创建 SharedSession
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

// 4. 存储和检索
sharedA.StoreLongTo(ctx, "team:alpha", "Fact", nil)
recs, _ := sharedB.Retrieve(ctx, "query", 5)  // ✅ 可访问 sharedA 的记忆
```

### 关键 API

| API | 说明 |
|-----|------|
| `session.NewSharedSession()` | 创建共享会话 |
| `shared.Join(space)` | 加入共享空间 |
| `shared.StoreLongTo(ctx, space, content, meta)` | 存储到共享空间 |
| `shared.Retrieve(ctx, query, limit)` | 检索（自动包含个人 + 共享） |
| `shared.FlushSpace(ctx, space)` | 持久化到数据库 |

### 适用场景

- ✅ 多 Agent 共享用户历史记忆
- ✅ 团队协作 Agent
- ✅ 跨会话知识传承

### 测试示例

```bash
go run cmd/example/shared_session_test/main.go
go run cmd/example/hybrid_memory_test/main.go
```

---

## 方式 2: SubAgent 委托（内置委托机制）

### 原理

主 Agent 通过 `SubAgentDirectory` 管理专家 Agent，使用 `subagent:` 命令委托任务。

### 架构图

```
┌─────────────────┐
│   Main Agent    │
│  (Coordinator)  │
│                 │
│  SubAgents:     │
│  - Researcher   │
│  - Coder        │
│  - Writer       │
└────────┬────────┘
         │ subagent:researcher
         ▼
┌─────────────────┐
│  Researcher     │
│  (Specialist)   │
│                 │
│  Run(task)      │
└─────────────────┘
```

### 核心代码

```go
// 1. 创建专家 SubAgent
researcher := &ResearcherSubAgent{model: llm}

// 2. 注册到 SubAgentDirectory
kit, _ := adk.New(ctx,
    adk.WithSubAgents(researcher),
    // ...
)

// 3. 主 Agent 自动生成 subagent: 命令
// 用户："Research quantum computing"
// Agent 输出："subagent:researcher Research quantum computing"
// 系统自动调用 researcher.Run()
```

### 关键 API

| API | 说明 |
|-----|------|
| `adk.WithSubAgents(...)` | 注册 SubAgent |
| `SubAgent.Run(ctx, instruction)` | 执行任务 |
| `SubAgent.Name()` | SubAgent 名称 |
| `SubAgent.Description()` | 描述（用于 LLM 决策） |

### SubAgent 接口

```go
type SubAgent interface {
    Name() string
    Description() string
    Run(ctx context.Context, instruction string) (string, error)
}
```

### 适用场景

- ✅ 专业分工（研究/编码/写作）
- ✅ 任务自动委托
- ✅ 简化主 Agent 逻辑

### 示例

```bash
go run cmd/quickstart/main.go  # 使用内置 SubAgents
```

---

## 方式 3: Agent-as-Tool（UTCP 工具调用）

### 原理

将 Agent 注册为 UTCP 工具，其他 Agent 通过 `CallTool()` 调用。

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

### 核心代码

```go
// 1. 创建 Agent-B（作为工具提供者）
researcher, _ := agent.New(agent.Options{
    Model:        llm,
    Memory:       sessionMemB,
    SystemPrompt: "You are a researcher.",
})

// 2. 注册为 UTCP 工具
client, _ := utcp.NewUTCPClient(ctx, ...)
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "Performs research")

// 3. Agent-A 调用 Agent-B
// LLM 决定调用工具 → agent.researcher
result, _ := client.CallTool(ctx, "agent.researcher", {
    "instruction": "Find facts about quantum computing",
})
```

### 关键 API

| API | 说明 |
|-----|------|
| `agent.RegisterAsUTCPProvider(ctx, client, name, desc)` | 注册为工具 |
| `client.CallTool(ctx, toolName, args)` | 调用工具 |
| `agent.AsUTCPTool(name, desc)` | 转换为 Tool 接口 |

### 适用场景

- ✅ Agent 间实时通信
- ✅ 层次化 Agent 架构（Manager → Worker）
- ✅ 跨服务 Agent 调用

### 示例

```bash
go run cmd/example/agent_as_tool/main.go
go run cmd/example/agent_as_utcp_codemode/main.go
go run cmd/example/codemode/main.go
```

---

## 方式 4: Swarm（群体协作）

### 原理

多个 Participant（参与者）通过共享记忆空间协作，每个 Participant 有自己的 Agent。

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│                         Swarm                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │Participant A│  │Participant B│  │Participant C│      │
│  │  Alias: A   │  │  Alias: B   │  │  Alias: C   │      │
│  │  Shared: α  │  │  Shared: α  │  │  Shared: α  │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
│         │                │                │             │
│         └────────────────┼────────────────┘             │
│                          │                              │
│                          ▼                              │
│              ┌─────────────────────┐                    │
│              │  Shared Memory (α)  │                    │
│              └─────────────────────┘                    │
└─────────────────────────────────────────────────────────┘
```

### 核心代码

```go
// 1. 创建 Participants
participants := swarm.Participants{
    "alice": &swarm.Participant{
        Alias:     "alice",
        SessionID: "session-alice",
        Agent:     agentAlice,
        Shared:    sharedSessionAlice,
    },
    "bob": &swarm.Participant{
        Alias:     "bob",
        SessionID: "session-bob",
        Agent:     agentBob,
        Shared:    sharedSessionBob,
    },
}

// 2. 创建 Swarm
swarm := swarm.NewSwarm(&participants)

// 3. 加入共享空间
swarm.Join("alice", "team:alpha")
swarm.Join("bob", "team:alpha")

// 4. 检索共享记忆
recs, _ := swarm.Retrieve(ctx, "alice")
```

### 关键 API

| API | 说明 |
|-----|------|
| `swarm.NewSwarm(participants)` | 创建 Swarm |
| `swarm.Join(id, space)` | Participant 加入空间 |
| `swarm.Retrieve(ctx, id)` | 检索共享记忆 |
| `participant.Save(ctx)` | 持久化记忆 |

### 适用场景

- ✅ 多角色协作（如：产品/开发/测试）
- ✅ 长期项目协作
- ✅ 需要共享上下文的团队

### 示例

```bash
# 查看 swarm 测试
go test ./src/swarm/...
```

---

## 方式 5: CodeMode 工作流编排

### 原理

通过 UTCP 工具链和 CodeMode，LLM 生成 Go 代码编排多个 Agent 的工作流。

### 架构图

```
┌──────────────────────────────────────────────────────────┐
│                    CodeMode Orchestrator                 │
│                                                          │
│  User: "Analyze data, write report, review it"          │
│         ↓                                                │
│  LLM generates Go code:                                 │
│  ```go                                                   │
│  result1 := CallTool("analyst", "Analyze data")         │
│  result2 := CallTool("writer", result1)                 │
│  result3 := CallTool("reviewer", result2)               │
│  ```                                                     │
│         ↓                                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                  │
│  │Analyst  │→ │Writer   │→ │Reviewer │                  │
│  └─────────┘  └─────────┘  └─────────┘                  │
└──────────────────────────────────────────────────────────┘
```

### 核心代码

```go
// 1. 注册多个专家 Agent 为 UTCP 工具
analyst.RegisterAsUTCPProvider(ctx, client, "analyst", "Analyzes data")
writer.RegisterAsUTCPProvider(ctx, client, "writer", "Writes reports")
reviewer.RegisterAsUTCPProvider(ctx, client, "reviewer", "Reviews content")

// 2. 创建带 CodeMode 的编排 Agent
orchestrator, _ := adk.New(ctx,
    adk.WithCodeModeUtcp(client, model),  // ← 关键：启用 CodeMode
)

// 3. 用户请求自动编排工作流
// 用户："Analyze Q4 sales and write a report"
// LLM 生成：
// ```go
// data := codemode.CallTool("analyst", "Analyze Q4 sales")
// report := codemode.CallTool("writer", data)
// ```
```

### 关键 API

| API | 说明 |
|-----|------|
| `adk.WithCodeModeUtcp(client, model)` | 启用 CodeMode |
| `codemode.CallTool(name, args)` | LLM 生成的工具调用 |
| `codemode.CallToolStream(name, args)` | 流式调用 |

### 适用场景

- ✅ 复杂多步骤工作流
- ✅ 需要条件判断/循环的任务
- ✅ 动态编排专家 Agent

### 示例

```bash
go run cmd/example/codemode/main.go
go run cmd/example/codemode_utcp_workflow/main.go
```

---

## 完整对比表

| 特性 | SharedSession | SubAgent | Agent-as-Tool | Swarm | CodeMode |
|------|---------------|----------|---------------|-------|----------|
| **记忆共享** | ✅ 自动 | ❌ 独立 | ❌ 独立 | ⚠️ 可选 | ❌ 独立 |
| **实时通信** | ❌ 间接 | ✅ 直接 | ✅ 直接 | ✅ 间接 | ✅ 直接 |
| **配置复杂度** | 低 | 低 | 中 | 中 | 高 |
| **持久化** | ✅ 自动 | ❌ 不持久 | ❌ 不持久 | ✅ 自动 | ❌ 不持久 |
| **任务委托** | ❌ | ✅ 自动 | ⚠️ 手动 | ⚠️ 手动 | ✅ 自动编排 |
| **工作流** | ❌ | ❌ | ⚠️ 简单 | ❌ | ✅ 复杂 |
| **适用场景** | 共享历史 | 专业分工 | 层次调用 | 团队协作 | 工作流编排 |

---

## 组合使用模式

### 模式 1: SharedSession + Agent-as-Tool（推荐）

```go
// 1. 共享记忆
sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")
sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

// 2. 实时通信
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "...")
result, _ := client.CallTool(ctx, "agent.researcher", {...})

// 优势：
// - 共享历史记忆（SharedSession）
// - 实时任务委托（Agent-as-Tool）
```

### 模式 2: Swarm + SharedSession

```go
// Swarm 内部使用 SharedSession 实现记忆共享
participant := &swarm.Participant{
    Alias:     "alice",
    SessionID: "session-alice",
    Agent:     agentAlice,
    Shared:    sharedSession,  // ← SharedSession
}
```

### 模式 3: SubAgent + CodeMode

```go
// SubAgent 作为 UTCP 工具，由 CodeMode 编排
subagentTool := agent.NewSubAgentTool(researcher)
client.RegisterTool(subagentTool)

// CodeMode 自动生成调用代码
```

---

## 选择指南

### 场景 1: 多 Agent 共享用户记忆
**推荐**: SharedSession

```
客服 Agent A 和客服 Agent B 需要访问同一用户的历史对话
→ 使用 SharedSession 共享 team:user-123 space
```

### 场景 2: 经理分配任务给专家
**推荐**: SubAgent 或 Agent-as-Tool

```
经理 Agent 需要研究员 Agent 做研究
→ SubAgent: 简单委托
→ Agent-as-Tool: 需要 UTCP 生态
```

### 场景 3: 多步骤工作流
**推荐**: CodeMode

```
分析数据 → 写报告 → 审核 → 发布
→ CodeMode 自动编排 UTCP 工具链
```

### 场景 4: 长期项目协作
**推荐**: Swarm

```
产品/开发/测试组成项目团队，需要共享项目记忆
→ Swarm 管理多个 Participant
```

### 场景 5: 混合场景（最佳实践）
**推荐**: SharedSession + Agent-as-Tool + CodeMode

```go
// 1. 共享记忆
shared := session.NewSharedSession(...)

// 2. 专家 Agent 注册为工具
expert.RegisterAsUTCPProvider(...)

// 3. CodeMode 编排工作流
adk.WithCodeModeUtcp(client, model)
```

---

## 文件位置

```
go-agent/
├── src/memory/session/
│   ├── shared_session.go    # SharedSession 实现
│   └── spaces.go            # SpaceRegistry 实现
├── src/swarm/
│   ├── swarm.go             # Swarm 实现
│   └── participant.go       # Participant 实现
├── agent.go                 # SubAgent 委托
├── agent_tool.go            # Agent-as-Tool 适配器
├── agent_orchestrators.go   # CodeMode 编排
├── catalog.go               # SubAgentDirectory 实现
└── cmd/example/
    ├── shared_session_test/     # SharedSession 测试
    ├── hybrid_memory_test/      # 混合记忆测试
    ├── agent_as_tool/           # Agent-as-Tool 示例
    ├── agent_as_utcp_codemode/  # CodeMode 示例
    ├── codemode/                # CodeMode 示例
    └── codemode_utcp_workflow/  # 工作流示例
```

---

## 总结

| 方式 | 核心优势 | 最佳场景 |
|------|----------|----------|
| **SharedSession** | 自动记忆共享 | 多 Agent 共享历史 |
| **SubAgent** | 简单委托 | 专业分工 |
| **Agent-as-Tool** | UTCP 生态 | 层次化调用 |
| **Swarm** | 团队协作 | 长期项目 |
| **CodeMode** | 自动编排 | 复杂工作流 |

**推荐组合**: SharedSession（记忆） + Agent-as-Tool（通信） + CodeMode（编排）
