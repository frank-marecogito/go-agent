# 5 SubAgents + 5 通信方式完整示例

## 项目结构

```
cmd/example/5ways_demo/
├── main.go                    # 主程序
├── agents/                    # 5 个专家 SubAgent
│   ├── researcher.go          # 研究员
│   ├── coder.go               # 程序员
│   ├── writer.go              # 作家
│   ├── reviewer.go            # 审核员
│   └── coordinator.go         # 协调员
├── README.md                  # 使用文档
└── test.sh                    # 测试脚本
```

---

## 5 个专家 SubAgent

### 1. Researcher（研究员）
- **职责**: 信息检索、事实核查、数据分析
- **能力**: 搜索、总结、引用来源

### 2. Coder（程序员）
- **职责**: 代码生成、代码审查、调试
- **能力**: Go/Python/JavaScript、单元测试

### 3. Writer（作家）
- **职责**: 文档撰写、报告生成、内容创作
- **能力**: 技术文档、博客文章、报告

### 4. Reviewer（审核员）
- **职责**: 质量审核、错误检查、改进建议
- **能力**: 语法检查、逻辑验证、最佳实践

### 5. Coordinator（协调员）
- **职责**: 任务分配、进度跟踪、结果整合
- **能力**: 工作流编排、多 Agent 协调

---

## 5 种通信方式调用示例

### 方式 1: SharedSession（共享记忆）

```go
// 场景：所有 Agent 共享项目记忆
package main

import (
    "context"
    "time"
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/memory/session"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 SpaceRegistry
    registry := session.NewSpaceRegistry(24 * time.Hour)
    
    // 2. 授予所有 Agent 访问权限
    agents := []string{"researcher", "coder", "writer", "reviewer", "coordinator"}
    for _, agent := range agents {
        registry.Grant("team:project-x", agent, session.SpaceRoleWriter, 0)
    }
    
    // 3. 每个 Agent 使用同一个 registry
    sessionMemResearcher := memory.NewSessionMemory(bank, 16)
    sessionMemResearcher.Spaces = registry
    sharedResearcher := session.NewSharedSession(sessionMemResearcher, "researcher", "team:project-x")
    
    sessionMemCoder := memory.NewSessionMemory(bank, 16)
    sessionMemCoder.Spaces = registry
    sharedCoder := session.NewSharedSession(sessionMemCoder, "coder", "team:project-x")
    
    // 4. Researcher 存储研究结果
    sharedResearcher.StoreLongTo(ctx, "team:project-x", 
        "Project requirement: Build a REST API with Go", nil)
    
    // 5. Coder 可以访问 Researcher 的研究结果
    recs, _ := sharedCoder.Retrieve(ctx, "project requirement", 5)
    // 找到："Project requirement: Build a REST API with Go"
    
    // 6. Coder 存储代码设计
    sharedCoder.StoreLongTo(ctx, "team:project-x",
        "API design: GET /users, POST /users, GET /users/:id", nil)
    
    // 7. Writer 可以访问所有记忆
    sharedWriter := session.NewSharedSession(sessionMemWriter, "writer", "team:project-x")
    allRecs, _ := sharedWriter.Retrieve(ctx, "project", 10)
    // 找到：研究结果 + 代码设计
}
```

**运行测试**:
```bash
go run cmd/example/shared_session_test/main.go
```

---

### 方式 2: SubAgent 委托

```go
// 场景：主 Agent 委托任务给专家 SubAgent
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent"
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
    "github.com/Protocol-Lattice/go-agent/src/models"
)

// 1. 定义 SubAgent 接口实现
type ResearcherSubAgent struct {
    model models.Agent
}

func (r *ResearcherSubAgent) Name() string { return "researcher" }
func (r *ResearcherSubAgent) Description() string { 
    return "Researches topics and provides factual information" 
}
func (r *ResearcherSubAgent) Run(ctx context.Context, instruction string) (string, error) {
    resp, err := r.model.Generate(ctx, "research-session", instruction)
    return resp.(string), err
}

// 2. 创建其他 SubAgent
type CoderSubAgent struct{ model models.Agent }
func (c *CoderSubAgent) Name() string { return "coder" }
func (c *CoderSubAgent) Description() string { 
    return "Writes and reviews code in multiple languages" 
}
func (c *CoderSubAgent) Run(ctx context.Context, instruction string) (string, error) {
    // 代码生成逻辑
    return "func Hello() string { return \"world\" }", nil
}

type WriterSubAgent struct{ model models.Agent }
func (w *WriterSubAgent) Name() string { return "writer" }
func (w *WriterSubAgent) Description() string { 
    return "Creates professional documentation and reports" 
}
func (w *WriterSubAgent) Run(ctx context.Context, instruction string) (string, error) {
    // 文档撰写逻辑
    return "# API Documentation\n\n...", nil
}

type ReviewerSubAgent struct{ model models.Agent }
func (r *ReviewerSubAgent) Name() string { return "reviewer" }
func (r *ReviewerSubAgent) Description() string { 
    return "Reviews content for quality and accuracy" 
}
func (r *ReviewerSubAgent) Run(ctx context.Context, instruction string) (string, error) {
    // 审核逻辑
    return "Quality score: 95/100. Suggestions: ...", nil
}

type CoordinatorSubAgent struct{ model models.Agent }
func (c *CoordinatorSubAgent) Name() string { return "coordinator" }
func (c *CoordinatorSubAgent) Description() string { 
    return "Coordinates tasks between team members" 
}
func (c *CoordinatorSubAgent) Run(ctx context.Context, instruction string) (string, error) {
    // 协调逻辑
    return "Task assigned to: researcher", nil
}

func main() {
    ctx := context.Background()
    
    // 3. 创建模型
    model, _ := models.NewDeepSeekLLM(ctx, "deepseek-chat", "Expert analysis:")
    
    // 4. 创建 5 个 SubAgent
    researcher := &ResearcherSubAgent{model: model}
    coder := &CoderSubAgent{model: model}
    writer := &WriterSubAgent{model: model}
    reviewer := &ReviewerSubAgent{model: model}
    coordinator := &CoordinatorSubAgent{model: model}
    
    // 5. 注册到 ADK
    kit, _ := adk.New(ctx,
        adk.WithDefaultSystemPrompt("You coordinate a team of experts."),
        adk.WithSubAgents(
            researcher,
            coder,
            writer,
            reviewer,
            coordinator,
        ),
        adk.WithModules(
            modules.NewModelModule("llm", func(_ context.Context) (models.Agent, error) {
                return model, nil
            }),
        ),
    )
    
    // 6. 构建 Agent
    agent, _ := kit.BuildAgent(ctx)
    
    // 7. 用户请求自动委托给 SubAgent
    // 用户："Research quantum computing"
    // → Agent 自动生成："subagent:researcher Research quantum computing"
    // → 系统调用 researcher.Run(ctx, "Research quantum computing")
    resp, _ := agent.Generate(ctx, "session-1", "Research quantum computing")
    
    // 用户："Write code for a REST API"
    // → subagent:coder Write code for a REST API
    resp, _ = agent.Generate(ctx, "session-1", "Write code for a REST API")
}
```

**运行测试**:
```bash
go run cmd/example/5ways_demo/main.go
```

---

### 方式 3: Agent-as-Tool（UTCP 工具调用）

```go
// 场景：Agent 作为 UTCP 工具被其他 Agent 调用
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent"
    "github.com/universal-tool-calling-protocol/go-utcp"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 UTCP 客户端
    utcpClient, _ := utcp.NewUTCPClient(ctx, &utcp.UtcpClientConfig{}, nil, nil)
    
    // 2. 创建 5 个专家 Agent
    researcher, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a researcher. You find facts.",
    })
    
    coder, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a coder. You write code.",
    })
    
    writer, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a writer. You create documentation.",
    })
    
    reviewer, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a reviewer. You check quality.",
    })
    
    coordinator, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a coordinator. You manage workflows.",
    })
    
    // 3. 注册为 UTCP 工具
    researcher.RegisterAsUTCPProvider(ctx, utcpClient, 
        "agent.researcher", "Researches topics and provides facts")
    
    coder.RegisterAsUTCPProvider(ctx, utcpClient, 
        "agent.coder", "Writes and reviews code")
    
    writer.RegisterAsUTCPProvider(ctx, utcpClient, 
        "agent.writer", "Creates documentation and reports")
    
    reviewer.RegisterAsUTCPProvider(ctx, utcpClient, 
        "agent.reviewer", "Reviews content for quality")
    
    coordinator.RegisterAsUTCPProvider(ctx, utcpClient, 
        "agent.coordinator", "Coordinates workflows")
    
    // 4. Manager Agent 调用专家 Agent
    manager, _ := agent.New(agent.Options{
        Model:        model,
        SystemPrompt: "You are a manager. Delegate tasks to specialists.",
        Tools:        utcpClient.Tools(),  // ← 所有专家 Agent 作为工具
    })
    
    // 5. Manager 调用 Researcher
    // 用户："What is quantum computing?"
    // → Manager 调用：agent.researcher
    result, _ := utcpClient.CallTool(ctx, "agent.researcher", map[string]any{
        "instruction": "Explain quantum computing",
    })
    
    // 6. Manager 调用 Coder
    result, _ = utcpClient.CallTool(ctx, "agent.coder", map[string]any{
        "instruction": "Write a Go function to add two numbers",
    })
    
    // 7. Manager 调用 Writer
    result, _ = utcpClient.CallTool(ctx, "agent.writer", map[string]any{
        "instruction": "Write API documentation",
    })
}
```

**运行测试**:
```bash
go run cmd/example/agent_as_tool/main.go
```

---

### 方式 4: Swarm（群体协作）

```go
// 场景：5 个 Agent 组成 Swarm 协作完成项目
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent/src/swarm"
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/memory/session"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建共享记忆空间
    registry := session.NewSpaceRegistry(0)  // 永不过期
    registry.Grant("team:project-x", "researcher", session.SpaceRoleWriter, 0)
    registry.Grant("team:project-x", "coder", session.SpaceRoleWriter, 0)
    registry.Grant("team:project-x", "writer", session.SpaceRoleWriter, 0)
    registry.Grant("team:project-x", "reviewer", session.SpaceRoleWriter, 0)
    registry.Grant("team:project-x", "coordinator", session.SpaceRoleAdmin, 0)
    
    // 2. 创建 5 个 Participant
    participants := swarm.Participants{
        "researcher": &swarm.Participant{
            Alias:     "researcher",
            SessionID: "session-researcher",
            Agent:     researcherAgent,
            Shared:    session.NewSharedSession(sessionMemResearcher, "researcher", "team:project-x"),
        },
        "coder": &swarm.Participant{
            Alias:     "coder",
            SessionID: "session-coder",
            Agent:     coderAgent,
            Shared:    session.NewSharedSession(sessionMemCoder, "coder", "team:project-x"),
        },
        "writer": &swarm.Participant{
            Alias:     "writer",
            SessionID: "session-writer",
            Agent:     writerAgent,
            Shared:    session.NewSharedSession(sessionMemWriter, "writer", "team:project-x"),
        },
        "reviewer": &swarm.Participant{
            Alias:     "reviewer",
            SessionID: "session-reviewer",
            Agent:     reviewerAgent,
            Shared:    session.NewSharedSession(sessionMemReviewer, "reviewer", "team:project-x"),
        },
        "coordinator": &swarm.Participant{
            Alias:     "coordinator",
            SessionID: "session-coordinator",
            Agent:     coordinatorAgent,
            Shared:    session.NewSharedSession(sessionMemCoordinator, "coordinator", "team:project-x"),
        },
    }
    
    // 3. 创建 Swarm
    swarm := swarm.NewSwarm(&participants)
    
    // 4. 所有 Participant 加入同一个共享空间
    swarm.Join("researcher", "team:project-x")
    swarm.Join("coder", "team:project-x")
    swarm.Join("writer", "team:project-x")
    swarm.Join("reviewer", "team:project-x")
    swarm.Join("coordinator", "team:project-x")
    
    // 5. 协作流程
    // Step 1: Researcher 研究需求
    researcherAgent.Save(ctx, "user", "Project: Build REST API")
    researcherAgent.Generate(ctx, "session-researcher", "Research REST API best practices")
    swarm.Save(ctx)  // 持久化到共享记忆
    
    // Step 2: Coder 访问共享记忆，开始编码
    coderRecs, _ := swarm.Retrieve(ctx, "coder")
    // 找到 Researcher 的研究结果
    coderAgent.Generate(ctx, "session-coder", "Implement REST API based on research")
    swarm.Save(ctx)
    
    // Step 3: Writer 编写文档
    writerRecs, _ := swarm.Retrieve(ctx, "writer")
    // 找到研究结果 + 代码
    writerAgent.Generate(ctx, "session-writer", "Write API documentation")
    swarm.Save(ctx)
    
    // Step 4: Reviewer 审核
    reviewerRecs, _ := swarm.Retrieve(ctx, "reviewer")
    // 找到所有记忆
    reviewerAgent.Generate(ctx, "session-reviewer", "Review all deliverables")
    swarm.Save(ctx)
    
    // Step 5: Coordinator 整合结果
    coordinatorRecs, _ := swarm.Retrieve(ctx, "coordinator")
    // 找到所有记忆
    coordinatorAgent.Generate(ctx, "session-coordinator", "Summarize project status")
    swarm.Save(ctx)
}
```

**运行测试**:
```bash
go test ./src/swarm/...
```

---

### 方式 5: CodeMode 编排

```go
// 场景：LLM 自动生成代码编排 5 个 Agent 的工作流
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent"
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/universal-tool-calling-protocol/go-utcp"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 UTCP 客户端
    utcpClient, _ := utcp.NewUTCPClient(ctx, &utcp.UtcpClientConfig{}, nil, nil)
    
    // 2. 注册 5 个专家 Agent 为 UTCP 工具
    researcher.RegisterAsUTCPProvider(ctx, utcpClient, 
        "expert.researcher", "Researches topics")
    coder.RegisterAsUTCPProvider(ctx, utcpClient, 
        "expert.coder", "Writes code")
    writer.RegisterAsUTCPProvider(ctx, utcpClient, 
        "expert.writer", "Creates documentation")
    reviewer.RegisterAsUTCPProvider(ctx, utcpClient, 
        "expert.reviewer", "Reviews content")
    coordinator.RegisterAsUTCPProvider(ctx, utcpClient, 
        "expert.coordinator", "Coordinates workflows")
    
    // 3. 创建编排 Agent（启用 CodeMode）
    model := &DeepSeekModel{apiKey: "..."}
    
    orchestrator, _ := adk.New(ctx,
        adk.WithDefaultSystemPrompt("You orchestrate workflows using expert agents."),
        adk.WithCodeModeUtcp(utcpClient, model),  // ← 启用 CodeMode
    )
    
    agent, _ := orchestrator.BuildAgent(ctx)
    
    // 4. 用户请求自动编排工作流
    // 用户："Create a complete project: research, code, document, review"
    // LLM 自动生成 Go 代码：
    /*
    func workflow() {
        // Step 1: Research
        researchResult := codemode.CallTool("expert.researcher", map[string]any{
            "instruction": "Research REST API best practices",
        })
        
        // Step 2: Code
        codeResult := codemode.CallTool("expert.coder", map[string]any{
            "instruction": "Implement REST API: " + researchResult,
        })
        
        // Step 3: Document
        docResult := codemode.CallTool("expert.writer", map[string]any{
            "instruction": "Write documentation for: " + codeResult,
        })
        
        // Step 4: Review
        reviewResult := codemode.CallTool("expert.reviewer", map[string]any{
            "instruction": "Review: " + docResult,
        })
        
        // Step 5: Coordinate
        finalResult := codemode.CallTool("expert.coordinator", map[string]any{
            "instruction": "Summarize project: " + reviewResult,
        })
        
        return finalResult
    }
    */
    
    // 5. 执行工作流
    resp, _ := agent.Generate(ctx, "session-1", 
        "Create a complete project: research, code, document, review")
    
    // 输出：完整的项目成果（研究 + 代码 + 文档 + 审核 + 总结）
}
```

**运行测试**:
```bash
go run cmd/example/codemode_utcp_workflow/main.go
```

---

## 5 种方式对比

| 特性 | SharedSession | SubAgent | Agent-as-Tool | Swarm | CodeMode |
|------|---------------|----------|---------------|-------|----------|
| **代码行数** | ~20 | ~50 | ~40 | ~60 | ~30 |
| **配置复杂度** | ⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **记忆共享** | ✅ | ❌ | ❌ | ✅ | ❌ |
| **实时通信** | ❌ | ✅ | ✅ | ⚠️ | ✅ |
| **工作流编排** | ❌ | ⚠️ | ⚠️ | ❌ | ✅ |
| **适用场景** | 共享历史 | 简单委托 | 层次调用 | 团队协作 | 复杂工作流 |

---

## 组合使用（最佳实践）

```go
// 推荐：SharedSession + Agent-as-Tool + CodeMode
func main() {
    ctx := context.Background()
    
    // 1. SharedSession: 所有 Agent 共享记忆
    registry := session.NewSpaceRegistry(0)
    for _, agent := range []string{"researcher", "coder", "writer", "reviewer", "coordinator"} {
        registry.Grant("team:project-x", agent, session.SpaceRoleWriter, 0)
    }
    
    // 2. Agent-as-Tool: 注册为 UTCP 工具
    utcpClient, _ := utcp.NewUTCPClient(ctx, ...)
    researcher.RegisterAsUTCPProvider(ctx, utcpClient, "expert.researcher", "...")
    coder.RegisterAsUTCPProvider(ctx, utcpClient, "expert.coder", "...")
    // ...
    
    // 3. CodeMode: 编排工作流
    orchestrator, _ := adk.New(ctx,
        adk.WithCodeModeUtcp(utcpClient, model),
    )
    
    // 结果：
    // - 共享记忆（SharedSession）
    // - 实时调用（Agent-as-Tool）
    // - 自动编排（CodeMode）
}
```

---

## 运行所有测试

```bash
cd /Users/frank/MareCogito/go-agent

# 方式 1: SharedSession
go run cmd/example/shared_session_test/main.go

# 方式 2: SubAgent
go run cmd/example/5ways_demo/main.go

# 方式 3: Agent-as-Tool
go run cmd/example/agent_as_tool/main.go

# 方式 4: Swarm
go test ./src/swarm/...

# 方式 5: CodeMode
go run cmd/example/codemode_utcp_workflow/main.go

# 组合测试
go run cmd/example/5ways_demo/all_in_one.go
```
