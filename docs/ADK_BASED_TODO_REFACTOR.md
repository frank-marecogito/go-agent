# 基于 ADK 架构的 TODO 任务重构方案

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 7 日  
**状态**: 设计方案  
**相关任务**: EXT-015, EXT-016, EXT-017

---

## 📋 概述

本文档分析 TODO_EXTENSION.md 中的任务，识别哪些可以参考 ADK（Agent Development Kit）的**依赖注入 module**和**provider**方式来实现，以实现模块化、可扩展的架构设计。

---

## 🎯 适合 ADK 架构的任务分析

### 任务分析矩阵

| 任务 | 适合度 | 理由 | 推荐实现方式 |
|------|--------|------|-------------|
| **EXT-017** 流式传输 | ⭐⭐⭐ 非常适合 | HTTP Handler 模块化 | Module + Provider |
| **EXT-016** 智能记忆系统 | ⭐⭐⭐ 非常适合 | 记忆扩展、因果推理模块 | Module + Provider |
| **EXT-015** 可确认 Agent | ⭐⭐⭐ 非常适合 | 确认规则、授权管理模块 | Module + Provider |
| **EXT-014** executeTool 修复 | ⭐ 不适合 | 核心 Bug 修复 | 直接修改 agent.go |
| **EXT-013** 中文工具调用 | ⭐⭐ 部分适合 | 工具调用识别扩展 | Helper 函数 |
| **EXT-012** 角色系统扩展 | ⭐⭐⭐ 非常适合 | 角色管理器模块 | Module + Provider |
| **EXT-011** MemCell/MemScene | ⭐⭐⭐ 非常适合 | 记忆系统扩展 | Module + Provider |
| **EXT-006** 可观测性 | ⭐⭐⭐ 非常适合 | 日志、指标、追踪模块 | Module + Provider |
| **EXT-005** Agent 通信协议 | ⭐⭐⭐ 非常适合 | 通信协议模块 | Module + Provider |
| **EXT-004** 内置工具库 | ⭐⭐⭐ 非常适合 | 工具提供者 | ToolProvider |
| **EXT-003** 预构建子代理 | ⭐⭐⭐ 非常适合 | 子代理提供者 | SubAgentProvider |

---

## 📦 详细设计方案

### 方案 1：EXT-017 流式传输模块

#### 架构设计

```
ADK Kit
    │
    ├─ WithModule(StreamModule)
    │     │
    │     └─ Provision()
    │           └─ kit.UseHTTPHandlerProvider(streamHandlerProvider)
    │
    └─ BuildAgent()
          │
          ▼
    HTTPHandlerProvider() → []HTTPHandler
          │
          ├─ SSEHandler (/api/stream)
          └─ WSHandler (/api/ws)
```

#### 实现方式

**1. 扩展 providers.go**

```go
// src/adk/providers.go (扩展)

// HTTPHandlerProvider 返回 HTTP 处理器列表
type HTTPHandlerProvider func(ctx context.Context) ([]HTTPHandler, error)

// HTTPHandler 描述单个 HTTP 处理器
type HTTPHandler struct {
    Method  string        // "GET", "POST", "WS"
    Path    string        // "/api/stream"
    Handler http.HandlerFunc
}

// 添加到 ADK kit
type AgentDevelopmentKit struct {
    // ... 现有字段 ...
    httpHandlerProviders []HTTPHandlerProvider  // 新增
}

// UseHTTPHandlerProvider 注册 HTTP Handler 提供者
func (k *AgentDevelopmentKit) UseHTTPHandlerProvider(provider HTTPHandlerProvider) {
    k.mu.Lock()
    defer k.mu.Unlock()
    k.httpHandlerProviders = append(k.httpHandlerProviders, provider)
}

// HTTPHandlerProvider 获取所有 HTTP Handler 提供者
func (k *AgentDevelopmentKit) HTTPHandlerProvider() HTTPHandlerProvider {
    providers := k.httpHandlerProviders
    return func(ctx context.Context) ([]HTTPHandler, error) {
        var allHandlers []HTTPHandler
        for _, provider := range providers {
            handlers, err := provider(ctx)
            if err != nil {
                return nil, err
            }
            allHandlers = append(allHandlers, handlers...)
        }
        return allHandlers, nil
    }
}
```

**2. 创建流式传输模块**

```go
// src/adk/modules/stream_module.go
package modules

import (
    "context"
    "net/http"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
)

// StreamModule 流式传输模块
type StreamModule struct {
    name    string
    enabled bool
}

func NewStreamModule(name string) *StreamModule {
    return &StreamModule{name: name, enabled: true}
}

func (m *StreamModule) Name() string {
    return m.name
}

func (m *StreamModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 注册 HTTP Handler Provider
    provider := func(ctx context.Context) ([]kit.HTTPHandler, error) {
        if !m.enabled {
            return nil, nil
        }
        
        // 创建流式处理器
        agent, err := kit.BuildAgent(ctx)
        if err != nil {
            return nil, err
        }
        
        streamHandler := NewStreamHandler(agent)
        
        return []kit.HTTPHandler{
            {
                Method:  "GET",
                Path:    "/api/stream",
                Handler: streamHandler.StreamChat,  // SSE
            },
            {
                Method:  "WS",
                Path:    "/api/ws",
                Handler: streamHandler.WSChat,  // WebSocket
            },
        }, nil
    }
    
    kit.UseHTTPHandlerProvider(provider)
    return nil
}
```

**3. 使用方式**

```go
package main

import (
    "context"
    "net/http"
    "log"
    
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 ADK，注册流式传输模块
    kit, err := adk.New(ctx,
        adk.WithModule(modules.NewStreamModule("stream")),
        adk.WithModule(modules.InMemoryMemoryModule(8, memory.AutoEmbedder(), memory.DefaultOptions())),
        // ... 其他模块
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 获取 HTTP Handlers
    handlers, err := kit.HTTPHandlerProvider()(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 注册路由
    for _, h := range handlers {
        switch h.Method {
        case "GET":
            http.HandleFunc(h.Path, h.Handler)
        case "WS":
            http.HandleFunc(h.Path, h.Handler)
        }
    }
    
    // 4. 启动服务器
    http.ListenAndServe(":8080", nil)
}
```

**优势**：
- ✅ **模块化** - 流式传输作为独立模块
- ✅ **可插拔** - 可以启用/禁用
- ✅ **可测试** - 可以单独测试模块
- ✅ **可组合** - 可以与其他模块组合

---

### 方案 2：EXT-016 智能记忆系统（分模块实现）

#### 架构设计

```
ADK Kit
    │
    ├─ WithModule(MemCellModule)     ← 记忆单元模块
    ├─ WithModule(CausalModule)      ← 因果推理模块
    └─ WithModule(SoftFactorModule)  ← 软性因素模块
    
Bootstrap()
    │
    ├─ MemCellModule.Provision()
    │   └─ kit.UseMemoryProvider(memCellProvider)
    │
    ├─ CausalModule.Provision()
    │   └─ kit.UseToolProvider(causalToolProvider)
    │
    └─ SoftFactorModule.Provision()
        └─ kit.UseAgentOption(withSoftFactors)
```

#### 模块实现

**1. MemCell 模块**

```go
// src/memory/memcell/memcell_module.go
package memcell

import (
    "context"
    "time"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/memory"
)

// MemCellModule 记忆单元模块
type MemCellModule struct {
    name string
    opts MemCellOptions
}

type MemCellOptions struct {
    EnableSummaries bool
    ClusterSize     int
}

func NewMemCellModule(name string, opts MemCellOptions) *MemCellModule {
    return &MemCellModule{name: name, opts: opts}
}

func (m *MemCellModule) Name() string {
    return m.name
}

func (m *MemCellModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 注册增强的记忆提供者
    provider := func(ctx context.Context) (kit.MemoryBundle, error) {
        // 1. 获取基础记忆
        baseProvider := kit.MemoryProvider()
        if baseProvider == nil {
            return kit.MemoryBundle{}, nil
        }
        
        bundle, err := baseProvider(ctx)
        if err != nil {
            return kit.MemoryBundle{}, err
        }
        
        // 2. 创建 MemCell 提取器
        extractor := NewEpisodicTraceFormator(bundle.Session)
        
        // 3. 创建语义整合器
        consolidator := NewSemanticConsolidator(
            bundle.Session.Engine,
            extractor,
            m.opts,
        )
        
        // 4. 启动后台任务
        go consolidator.RunPeriodically(ctx, 1*time.Hour)
        
        return bundle, nil
    }
    
    kit.UseMemoryProvider(provider)
    return nil
}
```

**2. 因果推理模块**

```go
// src/memory/causal/causal_module.go
package causal

import (
    "context"
    "time"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    agent "github.com/Protocol-Lattice/go-agent"
)

// CausalModule 因果推理模块
type CausalModule struct {
    name         string
    enableMining bool
    dowhyURL     string
}

func NewCausalModule(name string, enableMining bool, dowhyURL string) *CausalModule {
    return &CausalModule{
        name:         name,
        enableMining: enableMining,
        dowhyURL:     dowhyURL,
    }
}

func (m *CausalModule) Name() string {
    return m.name
}

func (m *CausalModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 1. 注册因果挖掘器
    if m.enableMining {
        miner := NewCausalMiner(kit.ModelProvider(), kit.MemoryProvider())
        
        // 启动后台挖掘任务
        go miner.RunPeriodically(ctx, 24*time.Hour)
    }
    
    // 2. 注册因果推理引擎
    reasoner := NewCausalReasoner(kit.MemoryProvider())
    
    // 3. 注册工具提供者
    toolProvider := func(ctx context.Context) (kit.ToolBundle, error) {
        if !m.enableMining {
            return kit.ToolBundle{}, nil
        }
        
        // 创建因果推理工具
        tools := []agent.Tool{
            NewCausalFindCausesTool(reasoner),
            NewCausalFindEffectsTool(reasoner),
            NewCausalEstimateEffectTool(reasoner, m.dowhyURL),
        }
        
        return kit.ToolBundle{
            Tools: tools,
        }, nil
    }
    
    kit.UseToolProvider(toolProvider)
    return nil
}
```

**3. 软性因素模块**

```go
// src/memory/causal/soft_factor_module.go
package causal

import (
    "context"
    "encoding/json"
    "os"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    agent "github.com/Protocol-Lattice/go-agent"
)

// SoftFactorModule 软性因素模块
type SoftFactorModule struct {
    name            string
    cultureFile     string
    philosophyFile  string
}

func NewSoftFactorModule(name string, cultureFile, philosophyFile string) *SoftFactorModule {
    return &SoftFactorModule{
        name:           name,
        cultureFile:    cultureFile,
        philosophyFile: philosophyFile,
    }
}

func (m *SoftFactorModule) Name() string {
    return m.name
}

func (m *SoftFactorModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 1. 加载企业文化节点
    var cultureNodes []*SoftFactor
    if m.cultureFile != "" {
        data, err := os.ReadFile(m.cultureFile)
        if err == nil {
            json.Unmarshal(data, &cultureNodes)
        }
    }
    
    // 2. 加载哲学原则节点
    var philosophyNodes []*SoftFactor
    if m.philosophyFile != "" {
        data, err := os.ReadFile(m.philosophyFile)
        if err == nil {
            json.Unmarshal(data, &philosophyNodes)
        }
    }
    
    // 3. 注册 Agent Option，在构建时注入软性因素
    kit.UseAgentOption(func(opts *agent.Options) {
        // 在系统提示词中加入软性因素
        opts.SystemPrompt += "\n\nConsider the following cultural factors:\n"
        for _, node := range cultureNodes {
            opts.SystemPrompt += "- " + node.Description + "\n"
        }
    })
    
    return nil
}
```

**4. 使用方式**

```go
package main

import (
    "context"
    "log"
    
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
    "github.com/Protocol-Lattice/go-agent/src/memory/memcell"
    "github.com/Protocol-Lattice/go-agent/src/memory/causal"
)

func main() {
    ctx := context.Background()
    
    // 创建 ADK，注册所有模块
    kit, err := adk.New(ctx,
        // 基础模块
        adk.WithModule(modules.NewModelModule("model", geminiProvider)),
        adk.WithModule(modules.InMemoryMemoryModule(8, memory.AutoEmbedder(), memory.DefaultOptions())),
        
        // 智能记忆模块
        adk.WithModule(memcell.NewMemCellModule("memcell", memcell.MemCellOptions{
            EnableSummaries: true,
            ClusterSize:     5,
        })),
        
        // 因果推理模块
        adk.WithModule(causal.NewCausalModule("causal", true, "http://localhost:8000")),
        
        // 软性因素模块
        adk.WithModule(causal.NewSoftFactorModule("soft_factors",
            "configs/culture_nodes.json",
            "configs/philosophy_nodes.json",
        )),
        
        // UTCP 集成
        adk.WithUTCP(utcpClient),
        adk.WithCodeModeUtcp(utcpClient, geminiModel),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 构建 Agent（自动包含所有模块功能）
    agent, err := kit.BuildAgent(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用 Agent
    resp, err := agent.Generate(ctx, "session1", "帮我分析一下这个现象的原因")
    // Agent 现在可以调用因果推理工具
}
```

---

### 方案 3：EXT-015 可确认 Agent 模块

#### 模块实现

```go
// src/adk/modules/confirmable_module.go
package modules

import (
    "context"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent"
    "your-module/confirmable"
)

// ConfirmableModule 可确认 Agent 模块
type ConfirmableModule struct {
    name          string
    rules         []*confirmable.ConfirmationRule
    handler       confirmable.ConfirmationHandler
    enableLogging bool
}

func NewConfirmableModule(name string, rules []*confirmable.ConfirmationRule, handler confirmable.ConfirmationHandler) *ConfirmableModule {
    return &ConfirmableModule{
        name:    name,
        rules:   rules,
        handler: handler,
    }
}

func (m *ConfirmableModule) Name() string {
    return m.name
}

func (m *ConfirmableModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 1. 创建确认管理器
    manager := confirmable.NewConfirmationManager(m.handler)
    
    // 添加规则
    for _, rule := range m.rules {
        manager.AddRule(rule)
    }
    
    // 2. 创建工具包装器 Provider
    toolProvider := func(ctx context.Context) (kit.ToolBundle, error) {
        // 获取原始工具
        baseProvider := kit.ToolProvider()
        if baseProvider == nil {
            return kit.ToolBundle{}, nil
        }
        
        bundle, err := baseProvider(ctx)
        if err != nil {
            return kit.ToolBundle{}, err
        }
        
        // 包装工具为可确认工具
        var wrappedTools []agent.Tool
        for _, tool := range bundle.Tools {
            wrapped := confirmable.NewConfirmableTool(tool, confirmable.ConfirmableToolOptions{
                Manager:  manager,
                StopCtrl: confirmable.NewStopController(),
            })
            wrappedTools = append(wrappedTools, wrapped)
        }
        
        return kit.ToolBundle{
            Catalog: bundle.Catalog,
            Tools:   wrappedTools,
        }, nil
    }
    
    kit.UseToolProvider(toolProvider)
    
    // 3. 注册 Guardrails
    if m.enableLogging {
        kit.UseAgentOption(func(opts *agent.Options) {
            opts.Guardrails = append(opts.Guardrails, 
                confirmable.NewConfirmationGuardrails(manager))
        })
    }
    
    return nil
}
```

#### 使用方式

```go
// 定义确认规则
rules := []*confirmable.ConfirmationRule{
    {
        ID:          "file_write_all",
        Type:        confirmable.RuleTypeToolName,
        ToolName:    "file.write",
        Description: "文件写入操作需要您的确认",
        Enabled:     true,
    },
    {
        ID:          "large_payment",
        Type:        confirmable.RuleTypeArgument,
        ToolName:    "payment.process",
        ArgumentKey: "amount",
        Condition: func(value any) bool {
            if amount, ok := value.(float64); ok {
                return amount > 1000
            }
            return false
        },
        Description: "大额支付需要确认",
        Enabled:     true,
    },
}

// 创建确认处理器
handler := func(ctx context.Context, req confirmable.ConfirmationRequest) (confirmable.ConfirmationResponse, error) {
    // 发送邮件/Slack/Web 通知
    // 等待用户响应
    // 返回用户决定
}

// 创建模块
confirmableModule := modules.NewConfirmableModule(
    "confirmable",
    rules,
    handler,
)

// 注册到 ADK
kit, err := adk.New(ctx,
    adk.WithModule(confirmableModule),
    // ... 其他模块
)
```

---

## 📊 重构后的 TODO 任务拆分

### 任务拆分表

| 原任务 | 拆分为 | 说明 | 预计工作量 |
|--------|--------|------|-----------|
| **EXT-017** 流式传输 | EXT-017A: StreamModule<br>EXT-017B: HTTPHandlerProvider | 模块化实现 | 8h → 6h+4h |
| **EXT-016** 智能记忆 | EXT-016A: MemCellModule<br>EXT-016B: CausalModule<br>EXT-016C: SoftFactorModule | 分模块实现 | 500h → 120h+200h+180h |
| **EXT-015** 可确认 Agent | EXT-015A: ConfirmableModule<br>EXT-015B: ConfirmationManager | 模块化实现 | 16h → 10h+8h |
| **EXT-012** 角色系统 | EXT-012A: RoleModule<br>EXT-012B: RoleProvider | 模块化实现 | - |
| **EXT-006** 可观测性 | EXT-006A: ObservabilityModule<br>EXT-006B: MetricsProvider<br>EXT-006C: TracingProvider | 模块化实现 | - |

### 实施顺序

**阶段 1：基础架构扩展**（1 周）
- [ ] 扩展 `providers.go` - 添加 `HTTPHandlerProvider`
- [ ] 扩展 `kit.go` - 添加 `UseHTTPHandlerProvider()` 方法
- [ ] 创建 `modules/` 目录结构

**阶段 2：流式传输模块**（1 周）
- [ ] 实现 `StreamModule`
- [ ] 实现 `StreamHandler`
- [ ] 编写测试

**阶段 3：智能记忆模块**（分阶段）
- [ ] EXT-016A: MemCellModule（2 周）
- [ ] EXT-016B: CausalModule（4 周）
- [ ] EXT-016C: SoftFactorModule（3 周）

**阶段 4：可确认 Agent 模块**（1 周）
- [ ] 实现 `ConfirmableModule`
- [ ] 实现 `ConfirmationManager`
- [ ] 编写测试

---

## 🎯 架构优势

### 1. 模块化

```
每个功能作为独立模块
    │
    ├─ 独立开发
    ├─ 独立测试
    └─ 独立部署
```

### 2. 可插拔

```
adk.New(ctx,
    WithModule(ModuleA),  ← 可以启用/禁用
    WithModule(ModuleB),  ← 可以启用/禁用
    WithModule(ModuleC),  ← 可以启用/禁用
)
```

### 3. 可组合

```
ModuleA + ModuleB → 功能 AB
ModuleA + ModuleC → 功能 AC
ModuleA + ModuleB + ModuleC → 功能 ABC
```

### 4. 可测试

```go
// 单独测试模块
func TestStreamModule(t *testing.T) {
    module := NewStreamModule("stream")
    kit := &AgentDevelopmentKit{}
    
    err := module.Provision(ctx, kit)
    if err != nil {
        t.Fatal(err)
    }
    
    // 验证模块正确注册
    if len(kit.httpHandlerProviders) != 1 {
        t.Fatal("Expected 1 handler provider")
    }
}
```

---

## 📋 总结

### 适合 ADK 架构的任务

| 任务 | 适合度 | 理由 |
|------|--------|------|
| EXT-017 流式传输 | ⭐⭐⭐ | HTTP Handler 模块化 |
| EXT-016 智能记忆 | ⭐⭐⭐ | 分模块实现（MemCell/Causal/SoftFactor） |
| EXT-015 可确认 Agent | ⭐⭐⭐ | 确认规则、授权管理模块化 |
| EXT-012 角色系统 | ⭐⭐⭐ | 角色管理器模块 |
| EXT-011 MemCell/MemScene | ⭐⭐⭐ | 记忆系统扩展模块 |
| EXT-006 可观测性 | ⭐⭐⭐ | 日志、指标、追踪模块 |
| EXT-005 Agent 通信 | ⭐⭐⭐ | 通信协议模块 |
| EXT-004 内置工具库 | ⭐⭐⭐ | ToolProvider |
| EXT-003 预构建子代理 | ⭐⭐⭐ | SubAgentProvider |

### 实施建议

1. **先扩展基础架构** - 添加 `HTTPHandlerProvider` 等
2. **分模块实施** - 每个任务拆分为独立模块
3. **逐步迁移** - 现有功能逐步重构为模块
4. **文档同步** - 每个模块配使用文档

### 预期收益

- ✅ **代码复用率提升** - 模块可在不同项目复用
- ✅ **开发效率提升** - 并行开发多个模块
- ✅ **测试覆盖率提升** - 模块可单独测试
- ✅ **维护成本降低** - 模块边界清晰

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 7 日*  
*维护：MareMind 项目基础设施团队*
