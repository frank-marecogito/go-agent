# 5 SubAgents + 5 通信方式完整演示

## 概述

本项目演示 go-agent 中 **5 个专家 SubAgent** 和 **5 种 Agent 间通信协作方式**。

---

## 5 个专家 SubAgent

| Agent | 职责 | 能力 |
|-------|------|------|
| **Researcher** | 信息检索、事实核查 | 搜索、总结、引用来源 |
| **Coder** | 代码生成、代码审查 | Go/Python/JavaScript、单元测试 |
| **Writer** | 文档撰写、报告生成 | 技术文档、博客文章、报告 |
| **Reviewer** | 质量审核、错误检查 | 语法检查、逻辑验证、最佳实践 |
| **Coordinator** | 任务分配、进度跟踪 | 工作流编排、多 Agent 协调 |

---

## 5 种通信方式

| # | 方式 | 核心能力 | 实现文件 | 测试命令 |
|---|------|----------|----------|----------|
| 1 | **SharedSession** | 共享记忆空间 | `src/memory/session/shared_session.go` | `go run cmd/example/shared_session_test/main.go` |
| 2 | **SubAgent 委托** | 内置任务委托 | `agent.go`, `catalog.go` | `go run cmd/example/5ways_demo/main.go` |
| 3 | **Agent-as-Tool** | UTCP 工具调用 | `agent_tool.go` | `go run cmd/example/agent_as_tool/main.go` |
| 4 | **Swarm** | 群体协作 | `src/swarm/` | `go test ./src/swarm/...` |
| 5 | **CodeMode 编排** | 工作流自动化 | `agent_orchestrators.go` | `go run cmd/example/codemode_utcp_workflow/main.go` |

---

## 快速开始

### 前置条件

```bash
# 1. PostgreSQL + pgvector
docker run -d --name postgres-pgvector \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=ragdb \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# 2. Ollama Embedding
ollama pull nomic-embed-text
ollama serve &

# 3. 环境变量
export DEEPSEEK_API_KEY="sk-xxx"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"
```

### 运行演示

```bash
cd /Users/frank/MareCogito/go-agent

# 运行完整演示（5 种方式）
go run cmd/example/5ways_demo/main.go
```

---

## 方式 1: SharedSession（共享记忆）

### 场景
所有 Agent 共享同一个项目记忆空间。

### 代码示例

```go
// 1. 配置 SpaceRegistry
registry := session.NewSpaceRegistry(24 * time.Hour)
registry.Grant("team:project-x", "researcher", session.SpaceRoleWriter, 0)
registry.Grant("team:project-x", "coder", session.SpaceRoleWriter, 0)

// 2. 创建 SharedSession
sharedResearcher := session.NewSharedSession(sessionMemResearcher, "researcher", "team:project-x")
sharedCoder := session.NewSharedSession(sessionMemCoder, "coder", "team:project-x")

// 3. Researcher 存储
sharedResearcher.StoreLongTo(ctx, "team:project-x", "Requirement: Build REST API", nil)

// 4. Coder 检索（可以访问 Researcher 的记忆）
recs, _ := sharedCoder.Retrieve(ctx, "requirement", 5)
// 找到："Requirement: Build REST API"
```

### 特点

| ✅ 优点 | ❌ 缺点 |
|--------|--------|
| 自动记忆共享 | 不是实时通信 |
| 配置简单 | 间接通信 |
| 持久化 | - |

---

## 方式 2: SubAgent 委托

### 场景
主 Agent 自动委托任务给专家 SubAgent。

### 代码示例

```go
// 1. 创建 SubAgent
researcher := &ResearcherSubAgent{model: model}
coder := &CoderSubAgent{model: model}

// 2. 注册到 ADK
kit, _ := adk.New(ctx,
    adk.WithSubAgents(researcher, coder),
)

// 3. 自动委托
// 用户："Research quantum computing"
// → Agent 输出："subagent:researcher Research quantum computing"
// → 系统调用：researcher.Run(ctx, "Research quantum computing")
resp, _ := agent.Generate(ctx, "session-1", "Research quantum computing")
```

### SubAgent 接口

```go
type SubAgent interface {
    Name() string
    Description() string
    Run(ctx context.Context, instruction string) (string, error)
}
```

### 特点

| ✅ 优点 | ❌ 缺点 |
|--------|--------|
| 自动委托 | 记忆不共享 |
| 配置简单 | 需要实现接口 |

---

## 方式 3: Agent-as-Tool（UTCP 工具调用）

### 场景
Agent 作为 UTCP 工具被其他 Agent 调用。

### 代码示例

```go
// 1. 注册为 UTCP 工具
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "Researches topics")

// 2. 调用工具
result, _ := client.CallTool(ctx, "agent.researcher", map[string]any{
    "instruction": "Explain quantum computing",
})
```

### 特点

| ✅ 优点 | ❌ 缺点 |
|--------|--------|
| 实时通信 | 需要正确配置 UTCP 客户端 |
| 层次化调用 | 配置较复杂 |

### ⚠️ 注意事项

在演示代码中，UTCP 工具调用可能显示"invalid CliProvider"错误，这是因为：

1. **UTCP 客户端需要正确的 Transport 配置**
2. **演示中使用了简化的 UTCP 客户端**

**解决方案**: 参考 `cmd/example/agent_as_tool/main.go` 使用完整配置，或使用 `MockModel` 进行测试。

**运行正常工作的示例**:
```bash
go run cmd/example/agent_as_tool/main.go
# 输出：✅ Tool Output: Research complete: ...
```

---

## 方式 4: Swarm（群体协作）

### 场景
多个 Participant 组成团队，共享记忆协作。

### 代码示例

```go
// 1. 创建 Participants
participants := swarm.Participants{
    "researcher": &swarm.Participant{
        Alias:     "researcher",
        SessionID: "session-researcher",
        Shared:    sharedSessionResearcher,
    },
    "coder": &swarm.Participant{
        Alias:     "coder",
        SessionID: "session-coder",
        Shared:    sharedSessionCoder,
    },
}

// 2. 创建 Swarm
swarm := swarm.NewSwarm(&participants)

// 3. 加入共享空间
swarm.Join("researcher", "team:project-x")
swarm.Join("coder", "team:project-x")

// 4. 检索共享记忆
recs, _ := swarm.Retrieve(ctx, "coder")
```

### 特点

| ✅ 优点 | ❌ 缺点 |
|--------|--------|
| 团队协作 | 配置较复杂 |
| 记忆共享 | - |

---

## 方式 5: CodeMode 编排

### 场景
LLM 自动生成代码编排多 Agent 工作流。

### 代码示例

```go
// 1. 启用 CodeMode
orchestrator, _ := adk.New(ctx,
    adk.WithCodeModeUtcp(client, model),
)

// 2. 用户请求自动编排
// 用户："Create project: research, code, document, review"
// LLM 生成：
// ```go
// r1 := CallTool("expert.researcher", "Research...")
// r2 := CallTool("expert.coder", r1)
// r3 := CallTool("expert.writer", r2)
// r4 := CallTool("expert.reviewer", r3)
// ```
resp, _ := agent.Generate(ctx, "session-1", "Create project")
```

### 特点

| ✅ 优点 | ❌ 缺点 |
|--------|--------|
| 自动编排 | 配置复杂 |
| 复杂工作流 | 需要 LLM 支持代码生成 |

---

## 组合使用（最佳实践）

```go
func main() {
    // 1. SharedSession: 共享记忆
    registry := session.NewSpaceRegistry(0)
    registry.Grant("team:project-x", "researcher", SpaceRoleWriter, 0)
    
    // 2. Agent-as-Tool: 实时调用
    researcher.RegisterAsUTCPProvider(ctx, client, "expert.researcher", "...")
    
    // 3. CodeMode: 工作流编排
    orchestrator, _ := adk.New(ctx,
        adk.WithCodeModeUtcp(client, model),
    )
    
    // 结果：
    // ✅ 共享历史记忆
    // ✅ 实时任务委托
    // ✅ 自动工作流编排
}
```

---

## 完整对比表

| 特性 | SharedSession | SubAgent | Agent-as-Tool | Swarm | CodeMode |
|------|---------------|----------|---------------|-------|----------|
| **记忆共享** | ✅ | ❌ | ❌ | ✅ | ❌ |
| **实时通信** | ❌ | ✅ | ✅ | ⚠️ | ✅ |
| **配置复杂度** | 低 | 低 | 中 | 中 | 高 |
| **工作流编排** | ❌ | ⚠️ | ⚠️ | ❌ | ✅ |
| **适用场景** | 共享历史 | 简单委托 | 层次调用 | 团队协作 | 复杂工作流 |

---

## 测试脚本

```bash
#!/bin/bash
# test.sh - 运行所有测试

export DEEPSEEK_API_KEY="sk-xxx"
export ADK_EMBED_PROVIDER="ollama"
export ADK_EMBED_MODEL="nomic-embed-text"

cd /Users/frank/MareCogito/go-agent

echo "=== Method 1: SharedSession ==="
go run cmd/example/shared_session_test/main.go

echo "=== Method 2: SubAgent ==="
go run cmd/example/5ways_demo/main.go

echo "=== Method 3: Agent-as-Tool ==="
go run cmd/example/agent_as_tool/main.go

echo "=== Method 4: Swarm ==="
go test ./src/swarm/... -v

echo "=== Method 5: CodeMode ==="
go run cmd/example/codemode_utcp_workflow/main.go
```

---

## 相关文件

```
go-agent/
├── cmd/example/
│   ├── 5ways_demo/
│   │   ├── main.go              # 主程序
│   │   └── COMPLETE_GUIDE.md    # 完整指南
│   ├── shared_session_test/     # SharedSession 测试
│   ├── hybrid_memory_test/      # 混合记忆测试
│   ├── agent_as_tool/           # Agent-as-Tool 示例
│   ├── codemode/                # CodeMode 示例
│   └── codemode_utcp_workflow/  # 工作流示例
├── src/
│   ├── memory/session/          # SharedSession 实现
│   ├── swarm/                   # Swarm 实现
│   └── ...
├── agent.go                     # SubAgent 委托
├── agent_tool.go                # Agent-as-Tool
└── agent_orchestrators.go       # CodeMode 编排
```

---

## 总结

| 方式 | 推荐场景 |
|------|----------|
| **SharedSession** | 多 Agent 共享用户历史记忆 |
| **SubAgent** | 主 Agent 委托任务给专家 |
| **Agent-as-Tool** | 层次化 Agent 架构 |
| **Swarm** | 长期项目团队协作 |
| **CodeMode** | 复杂多步骤工作流 |

**推荐组合**: SharedSession（记忆） + Agent-as-Tool（通信） + CodeMode（编排）
