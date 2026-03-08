# ADK Provider 参考手册

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: API 参考  
**相关文档**: [ADK_MODULE_DEVELOPMENT_GUIDE.md](./ADK_MODULE_DEVELOPMENT_GUIDE.md)

---

## 📋 目录

1. [Provider 概述](#provider 概述)
2. [ModelProvider](#modelprovider)
3. [MemoryProvider](#memoryprovider)
4. [ToolProvider](#toolprovider)
5. [SubAgentProvider](#subagentprovider)
6. [HTTPHandlerProvider](#httphandlerprovider)
7. [使用示例](#使用示例)
8. [最佳实践](#最佳实践)

---

## Provider 概述

### 什么是 Provider？

**Provider** 是**工厂函数**，用于延迟创建和配置组件。

```go
// Provider 本质
type Provider func(ctx context.Context) (Component, error)
```

### 为什么使用 Provider？

| 优势 | 说明 |
|------|------|
| **延迟初始化** | 需要时才创建组件 |
| **支持缓存** | 可以缓存实例 |
| **错误处理** | 可以返回错误 |
| **上下文支持** | 可以传递 context |

### Provider 类型

| Provider | 返回类型 | 用途 |
|----------|----------|------|
| **ModelProvider** | `models.Agent` | 创建 LLM |
| **MemoryProvider** | `MemoryBundle` | 创建记忆系统 |
| **ToolProvider** | `ToolBundle` | 创建工具列表 |
| **SubAgentProvider** | `SubAgentBundle` | 创建子代理列表 |
| **HTTPHandlerProvider** | `[]HTTPHandler` | 创建 HTTP 处理器 |

---

## ModelProvider

### 定义

```go
// src/adk/providers.go
type ModelProvider func(ctx context.Context) (models.Agent, error)
```

### 用途

创建协调器语言模型（LLM）。

### 示例

#### 1. 静态 Provider（始终返回相同实例）

```go
func StaticModelProvider(model models.Agent) ModelProvider {
    return func(ctx context.Context) (models.Agent, error) {
        return model, nil
    }
}

// 使用
gemini, _ := models.NewGeminiLLM(ctx, "gemini-2.5-pro", "")
kit.UseModelProvider(StaticModelProvider(gemini))
```

#### 2. 动态 Provider（每次创建新实例）

```go
func DynamicModelProvider(modelName string) ModelProvider {
    return func(ctx context.Context) (models.Agent, error) {
        return models.NewGeminiLLM(ctx, modelName, "")
    }
}

// 使用
kit.UseModelProvider(DynamicModelProvider("gemini-2.5-pro"))
```

#### 3. 带缓存的 Provider

```go
var cachedModel models.Agent

func CachedModelProvider(modelName string) ModelProvider {
    return func(ctx context.Context) (models.Agent, error) {
        if cachedModel != nil {
            return cachedModel, nil
        }
        
        model, err := models.NewGeminiLLM(ctx, modelName, "")
        if err != nil {
            return nil, err
        }
        
        cachedModel = model
        return model, nil
    }
}
```

### 在 Module 中使用

```go
type ModelModule struct {
    name      string
    modelName string
}

func (m *ModelModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    provider := func(ctx context.Context) (models.Agent, error) {
        return models.NewGeminiLLM(ctx, m.modelName, "")
    }
    
    kit.UseModelProvider(provider)
    return nil
}
```

---

## MemoryProvider

### 定义

```go
type MemoryProvider func(ctx context.Context) (MemoryBundle, error)

type MemoryBundle struct {
    Session *memory.SessionMemory
    Shared  SharedSessionFactory
}
```

### 用途

创建会话记忆和共享会话工厂。

### 示例

#### 1. 静态 Provider

```go
func StaticMemoryProvider(mem *memory.SessionMemory) MemoryProvider {
    shared := func(local string, spaces ...string) *memory.SharedSession {
        return memory.NewSharedSession(mem, local, spaces...)
    }
    
    return func(ctx context.Context) (MemoryBundle, error) {
        return MemoryBundle{
            Session: mem,
            Shared:  shared,
        }, nil
    }
}
```

#### 2. 内存存储 Provider

```go
func InMemoryMemoryProvider(windowSize int) MemoryProvider {
    return func(ctx context.Context) (MemoryBundle, error) {
        bank := memory.NewMemoryBankWithStore(memory.NewInMemoryStore())
        mem := memory.NewSessionMemory(bank, windowSize)
        mem.WithEmbedder(memory.AutoEmbedder())
        
        shared := func(local string, spaces ...string) *memory.SharedSession {
            return memory.NewSharedSession(mem, local, spaces...)
        }
        
        return MemoryBundle{
            Session: mem,
            Shared:  shared,
        }, nil
    }
}
```

#### 3. PostgreSQL 存储 Provider

```go
func PostgresMemoryProvider(ctx context.Context, connStr string, windowSize int) MemoryProvider {
    return func(ctx context.Context) (MemoryBundle, error) {
        store, err := memory.NewPostgresStore(ctx, connStr)
        if err != nil {
            return MemoryBundle{}, err
        }
        
        bank := memory.NewMemoryBankWithStore(store)
        mem := memory.NewSessionMemory(bank, windowSize)
        mem.WithEmbedder(memory.AutoEmbedder())
        
        shared := func(local string, spaces ...string) *memory.SharedSession {
            return memory.NewSharedSession(mem, local, spaces...)
        }
        
        return MemoryBundle{
            Session: mem,
            Shared:  shared,
        }, nil
    }
}
```

### 在 Module 中使用

```go
type MemoryModule struct {
    name       string
    windowSize int
    storeType  string  // "memory", "postgres", "qdrant"
    connStr    string
}

func (m *MemoryModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    provider := func(ctx context.Context) (kit.MemoryBundle, error) {
        var store memory.VectorStore
        var err error
        
        switch m.storeType {
        case "memory":
            store = memory.NewInMemoryStore()
        case "postgres":
            store, err = memory.NewPostgresStore(ctx, m.connStr)
        case "qdrant":
            store = memory.NewQdrantStore(m.connStr, "memories", "")
        }
        
        if err != nil {
            return kit.MemoryBundle{}, err
        }
        
        bank := memory.NewMemoryBankWithStore(store)
        mem := memory.NewSessionMemory(bank, m.windowSize)
        mem.WithEmbedder(memory.AutoEmbedder())
        
        shared := func(local string, spaces ...string) *memory.SharedSession {
            return memory.NewSharedSession(mem, local, spaces...)
        }
        
        return kit.MemoryBundle{
            Session: mem,
            Shared:  shared,
        }, nil
    }
    
    kit.UseMemoryProvider(provider)
    return nil
}
```

---

## ToolProvider

### 定义

```go
type ToolProvider func(ctx context.Context) (ToolBundle, error)

type ToolBundle struct {
    Catalog agent.ToolCatalog
    Tools   []agent.Tool
}
```

### 用途

创建工具目录和工具列表。

### 示例

#### 1. 静态工具列表

```go
func StaticToolProvider(tools []agent.Tool) ToolProvider {
    return func(ctx context.Context) (ToolBundle, error) {
        catalog := agent.NewStaticToolCatalog(tools)
        return ToolBundle{
            Catalog: catalog,
            Tools:   tools,
        }, nil
    }
}

// 使用
tools := []agent.Tool{
    &EchoTool{},
    &CalculatorTool{},
}
kit.UseToolProvider(StaticToolProvider(tools))
```

#### 2. 动态工具列表

```go
func DynamicToolProvider(toolFactory func() []agent.Tool) ToolProvider {
    return func(ctx context.Context) (ToolBundle, error) {
        tools := toolFactory()
        catalog := agent.NewStaticToolCatalog(tools)
        return ToolBundle{
            Catalog: catalog,
            Tools:   tools,
        }, nil
    }
}
```

#### 3. 装饰器工具 Provider

```go
func DecoratedToolProvider(
    baseProvider ToolProvider,
    decorator func(agent.Tool) agent.Tool,
) ToolProvider {
    return func(ctx context.Context) (ToolBundle, error) {
        bundle, err := baseProvider(ctx)
        if err != nil {
            return ToolBundle{}, err
        }
        
        // 装饰所有工具
        var decoratedTools []agent.Tool
        for _, tool := range bundle.Tools {
            decoratedTools = append(decoratedTools, decorator(tool))
        }
        
        return ToolBundle{
            Catalog: bundle.Catalog,
            Tools:   decoratedTools,
        }, nil
    }
}

// 使用：包装确认装饰器
kit.UseToolProvider(DecoratedToolProvider(
    baseProvider,
    func(tool agent.Tool) agent.Tool {
        return confirmable.NewConfirmableToolWrapper(tool, handler, guardrails)
    },
))
```

### 在 Module 中使用

```go
type ToolModule struct {
    name  string
    tools []agent.Tool
}

func (m *ToolModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    provider := func(ctx context.Context) (kit.ToolBundle, error) {
        catalog := agent.NewStaticToolCatalog(m.tools)
        return kit.ToolBundle{
            Catalog: catalog,
            Tools:   m.tools,
        }, nil
    }
    
    kit.UseToolProvider(provider)
    return nil
}
```

---

## SubAgentProvider

### 定义

```go
type SubAgentProvider func(ctx context.Context) (SubAgentBundle, error)

type SubAgentBundle struct {
    Directory agent.SubAgentDirectory
    SubAgents []agent.SubAgent
}
```

### 用途

创建子代理目录和子代理列表。

### 示例

#### 1. 静态子代理列表

```go
func StaticSubAgentProvider(subAgents []agent.SubAgent) SubAgentProvider {
    return func(ctx context.Context) (SubAgentBundle, error) {
        directory := agent.NewStaticSubAgentDirectory(subAgents)
        return SubAgentBundle{
            Directory: directory,
            SubAgents: subAgents,
        }, nil
    }
}
```

#### 2. 动态子代理 Provider

```go
func DynamicSubAgentProvider(factory func() []agent.SubAgent) SubAgentProvider {
    return func(ctx context.Context) (SubAgentBundle, error) {
        subAgents := factory()
        directory := agent.NewStaticSubAgentDirectory(subAgents)
        return SubAgentBundle{
            Directory: directory,
            SubAgents: subAgents,
        }, nil
    }
}
```

---

## HTTPHandlerProvider

### 定义

```go
type HTTPHandlerProvider func(ctx context.Context) ([]HTTPHandler, error)

type HTTPHandler struct {
    Method  string           // "GET", "POST", "WS"
    Path    string           // "/api/stream"
    Handler http.HandlerFunc
}
```

### 用途

创建 HTTP 处理器列表（用于流式传输等）。

### 示例

#### 1. SSE Handler Provider

```go
func SSEHandlerProvider(agent *agent.Agent) HTTPHandlerProvider {
    return func(ctx context.Context) ([]HTTPHandler, error) {
        handler := NewStreamHandler(agent)
        
        return []HTTPHandler{
            {
                Method:  "GET",
                Path:    "/api/stream",
                Handler: handler.StreamChat,
            },
        }, nil
    }
}
```

#### 2. 组合 Handler Provider

```go
func CombinedHandlerProvider(providers ...HTTPHandlerProvider) HTTPHandlerProvider {
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

---

## 使用示例

### 完整示例

```go
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 ADK
    kit, err := adk.New(ctx,
        // Model Provider
        adk.WithModule(modules.NewModelModule("model", func(ctx context.Context) (models.Agent, error) {
            return models.NewGeminiLLM(ctx, "gemini-2.5-pro", "")
        })),
        
        // Memory Provider
        adk.WithModule(modules.NewMemoryModule("memory", func(ctx context.Context) (kit.MemoryBundle, error) {
            bank := memory.NewMemoryBankWithStore(memory.NewInMemoryStore())
            mem := memory.NewSessionMemory(bank, 8)
            mem.WithEmbedder(memory.AutoEmbedder())
            
            shared := func(local string, spaces ...string) *memory.SharedSession {
                return memory.NewSharedSession(mem, local, spaces...)
            }
            
            return kit.MemoryBundle{
                Session: mem,
                Shared:  shared,
            }, nil
        })),
        
        // Tool Provider
        adk.WithModule(modules.NewToolModule("tools", func(ctx context.Context) (kit.ToolBundle, error) {
            tools := []agent.Tool{
                &EchoTool{},
                &CalculatorTool{},
            }
            catalog := agent.NewStaticToolCatalog(tools)
            
            return kit.ToolBundle{
                Catalog: catalog,
                Tools:   tools,
            }, nil
        })),
    )
    
    // 2. 构建 Agent
    agent, err := kit.BuildAgent(ctx)
    
    // 3. 使用 Agent
    resp, _ := agent.Generate(ctx, "session1", "Hello")
}
```

### Provider 链示例

```go
// 1. 基础 Provider
baseToolProvider := StaticToolProvider([]agent.Tool{
    &EchoTool{},
    &CalculatorTool{},
})

// 2. 装饰器 Provider（添加确认功能）
confirmableProvider := DecoratedToolProvider(
    baseToolProvider,
    func(tool agent.Tool) agent.Tool {
        return confirmable.NewConfirmableToolWrapper(tool, handler, guardrails)
    },
)

// 3. 再装饰（添加日志功能）
loggingProvider := DecoratedToolProvider(
    confirmableProvider,
    func(tool agent.Tool) agent.Tool {
        return logging.NewLoggingToolWrapper(tool, logger)
    },
)

// 4. 注册
kit.UseToolProvider(loggingProvider)

// 最终工具链：
// Logging(Confirmable(EchoTool))
// Logging(Confirmable(CalculatorTool))
```

---

## 最佳实践

### 1. Provider 命名

```go
// ✅ 好的命名
func StaticModelProvider(model models.Agent) ModelProvider
func InMemoryMemoryProvider(windowSize int) MemoryProvider
func DecoratedToolProvider(base ToolProvider, decorator func(Tool) Tool) ToolProvider

// ❌ 避免的命名
func Provider1() ModelProvider  // 无意义
func GetProvider() ToolProvider // 像 getter
```

### 2. 错误处理

```go
// ✅ 正确：详细的错误信息
provider := func(ctx context.Context) (MemoryBundle, error) {
    store, err := memory.NewPostgresStore(ctx, connStr)
    if err != nil {
        return MemoryBundle{}, fmt.Errorf("memory provider: create store: %w", err)
    }
    // ...
}

// ❌ 避免：错误信息不清晰
provider := func(ctx context.Context) (MemoryBundle, error) {
    store, err := memory.NewPostgresStore(ctx, connStr)
    if err != nil {
        return MemoryBundle{}, err  // ❌ 丢失上下文
    }
    // ...
}
```

### 3. 缓存策略

```go
// ✅ 正确：单例缓存
var cached *memory.SessionMemory

provider := func(ctx context.Context) (MemoryBundle, error) {
    if cached != nil {
        return MemoryBundle{Session: cached}, nil
    }
    
    // 创建实例
    cached = memory.NewSessionMemory(...)
    return MemoryBundle{Session: cached}, nil
}

// ❌ 避免：每次创建新实例（浪费资源）
provider := func(ctx context.Context) (MemoryBundle, error) {
    mem := memory.NewSessionMemory(...)  // 每次都创建
    return MemoryBundle{Session: mem}, nil
}
```

### 4. 上下文传递

```go
// ✅ 正确：传递 context
provider := func(ctx context.Context) (MemoryBundle, error) {
    store, err := memory.NewPostgresStore(ctx, connStr)  // ← 传递 ctx
    // ...
}

// ❌ 避免：忽略 context
provider := func(ctx context.Context) (MemoryBundle, error) {
    store, err := memory.NewPostgresStore(context.Background(), connStr)  // ❌
    // ...
}
```

### 5. Provider 组合

```go
// ✅ 正确：使用组合函数
func CombineToolProviders(providers ...ToolProvider) ToolProvider {
    return func(ctx context.Context) (ToolBundle, error) {
        var allTools []agent.Tool
        var catalog agent.ToolCatalog
        
        for _, provider := range providers {
            bundle, err := provider(ctx)
            if err != nil {
                return ToolBundle{}, err
            }
            allTools = append(allTools, bundle.Tools...)
            if catalog == nil {
                catalog = bundle.Catalog
            }
        }
        
        return ToolBundle{
            Catalog: catalog,
            Tools:   allTools,
        }, nil
    }
}
```

---

## 总结

### Provider 类型总览

| Provider | 返回类型 | 用途 | 示例 |
|----------|----------|------|------|
| **ModelProvider** | `models.Agent` | 创建 LLM | `StaticModelProvider(gemini)` |
| **MemoryProvider** | `MemoryBundle` | 创建记忆系统 | `InMemoryMemoryProvider(8)` |
| **ToolProvider** | `ToolBundle` | 创建工具列表 | `StaticToolProvider(tools)` |
| **SubAgentProvider** | `SubAgentBundle` | 创建子代理 | `StaticSubAgentProvider(agents)` |
| **HTTPHandlerProvider** | `[]HTTPHandler` | 创建 HTTP 处理器 | `SSEHandlerProvider(agent)` |

### 使用原则

| 原则 | 说明 |
|------|------|
| **延迟初始化** | 需要时才创建组件 |
| **支持缓存** | 避免重复创建 |
| **错误处理** | 返回详细的错误信息 |
| **上下文传递** | 传递 context 支持取消 |
| **可组合** | 支持 Provider 链式装饰 |

### 下一步

- 阅读 [ADK_MODULE_DEVELOPMENT_GUIDE.md](./ADK_MODULE_DEVELOPMENT_GUIDE.md) - 模块开发指南
- 阅读 [DECORATOR_PATTERN_GUIDE.md](./DECORATOR_PATTERN_GUIDE.md) - 装饰器模式指南
- 开始开发你的第一个 Provider！

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
