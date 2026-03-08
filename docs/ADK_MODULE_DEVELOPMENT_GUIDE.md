# ADK 模块开发指南

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: 开发指南  
**相关文档**: [DECORATOR_PATTERN_GUIDE.md](./DECORATOR_PATTERN_GUIDE.md), [ADK_PROVIDER_REFERENCE.md](./ADK_PROVIDER_REFERENCE.md)

---

## 📋 目录

1. [什么是 ADK Module](#什么是 adk-module)
2. [Module 接口](#module-接口)
3. [开发步骤](#开发步骤)
4. [完整示例](#完整示例)
5. [最佳实践](#最佳实践)
6. [测试方法](#测试方法)
7. [常见问题](#常见问题)

---

## 什么是 ADK Module

### 概念

**ADK Module**是 Agent Development Kit 的**可插拔功能单元**，允许开发者以模块化方式扩展 Agent 功能。

### 类比

```
ADK Kit = 电脑主板
Module  = 扩展卡（显卡/声卡/网卡）
Provider = 接口（PCIe/USB）
```

### 优势

| 优势 | 说明 |
|------|------|
| **可插拔** | 按需启用/禁用模块 |
| **解耦** | 模块之间独立开发 |
| **可测试** | 单独测试每个模块 |
| **可复用** | 模块可在不同项目复用 |

---

## Module 接口

### 接口定义

```go
// src/adk/module.go
type Module interface {
    // Name 返回人类友好的名称（用于调试）
    Name() string

    // Provision 将功能附加到 kit
    Provision(ctx context.Context, kit *AgentDevelopmentKit) error
}
```

### 生命周期

```
创建 Module
    │
    ▼
adk.New(ctx, WithModule(module))
    │
    ▼
Bootstrap()
    │
    └─ module.Provision(kit)  ← 注册功能
        │
        ├─ kit.UseModelProvider(...)
        ├─ kit.UseMemoryProvider(...)
        ├─ kit.UseToolProvider(...)
        └─ kit.UseAgentOption(...)
    │
    ▼
BuildAgent()
    │
    └─ 使用已注册的功能
```

---

## 开发步骤

### 步骤 1：定义 Module 结构

```go
package modules

type MyModule struct {
    name    string
    enabled bool
    config  MyModuleConfig
}

type MyModuleConfig struct {
    // 模块配置
    Option1 string
    Option2 int
}

func NewMyModule(name string, config MyModuleConfig) *MyModule {
    return &MyModule{
        name:    name,
        enabled: true,
        config:  config,
    }
}
```

### 步骤 2：实现 Name() 方法

```go
func (m *MyModule) Name() string {
    return m.name
}
```

### 步骤 3：实现 Provision() 方法

```go
func (m *MyModule) Provision(ctx context.Context, kit *AgentDevelopmentKit) error {
    if !m.enabled {
        return nil
    }
    
    // 1. 创建 Provider
    provider := func(ctx context.Context) (SomeBundle, error) {
        // 创建组件
        component := NewComponent(m.config)
        
        return SomeBundle{
            Component: component,
        }, nil
    }
    
    // 2. 注册 Provider
    kit.UseSomeProvider(provider)
    
    return nil
}
```

### 步骤 4：编写测试

```go
func TestMyModule(t *testing.T) {
    ctx := context.Background()
    
    // 创建模块
    module := NewMyModule("test", MyModuleConfig{})
    
    // 创建 Kit
    kit, _ := adk.New(ctx, adk.WithModule(module))
    
    // 验证 Provision 正确执行
    err := kit.Bootstrap(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    // 验证 Provider 已注册
    // ...
}
```

---

## 完整示例

### 示例 1：流式传输模块

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
    config  StreamConfig
}

type StreamConfig struct {
    EnableSSE       bool
    EnableWebSocket bool
    Port            int
}

func NewStreamModule(name string, config StreamConfig) *StreamModule {
    return &StreamModule{
        name:    name,
        enabled: true,
        config:  config,
    }
}

func (m *StreamModule) Name() string {
    return m.name
}

// Provision 注册 HTTP Handler Provider
func (m *StreamModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    if !m.enabled {
        return nil
    }
    
    // 创建 HTTP Handler Provider
    provider := func(ctx context.Context) ([]kit.HTTPHandler, error) {
        // 创建 Agent（用于流式生成）
        agent, err := kit.BuildAgent(ctx)
        if err != nil {
            return nil, err
        }
        
        // 创建流式处理器
        handler := NewStreamHandler(agent)
        
        var handlers []kit.HTTPHandler
        
        // 添加 SSE 端点
        if m.config.EnableSSE {
            handlers = append(handlers, kit.HTTPHandler{
                Method:  "GET",
                Path:    "/api/stream",
                Handler: handler.StreamChat,
            })
        }
        
        // 添加 WebSocket 端点
        if m.config.EnableWebSocket {
            handlers = append(handlers, kit.HTTPHandler{
                Method:  "WS",
                Path:    "/api/ws",
                Handler: handler.WSChat,
            })
        }
        
        return handlers, nil
    }
    
    // 注册 Provider
    kit.UseHTTPHandlerProvider(provider)
    
    return nil
}
```

### 示例 2：确认功能模块

```go
// src/adk/modules/confirmable_module.go
package modules

import (
    "context"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    agent "github.com/Protocol-Lattice/go-agent"
    "your-module/confirmable"
)

// ConfirmationModule 确认功能模块
type ConfirmationModule struct {
    name     string
    rules    []*confirmable.ConfirmationRule
    handler  confirmable.ConfirmationHandler
    enabled  bool
}

func NewConfirmationModule(
    name string,
    rules []*confirmable.ConfirmationRule,
    handler confirmable.ConfirmationHandler,
) *ConfirmationModule {
    return &ConfirmationModule{
        name:    name,
        rules:   rules,
        handler: handler,
        enabled: true,
    }
}

func (m *ConfirmationModule) Name() string {
    return m.name
}

// Provision 注册确认功能
func (m *ConfirmationModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    if !m.enabled {
        return nil
    }
    
    // 1. 创建 Guardrails
    guardrails := confirmable.NewConfirmationGuardrails(m.rules)
    
    // 2. 创建 Tool Provider（包装所有工具）
    toolProvider := func(ctx context.Context) (kit.ToolBundle, error) {
        // 获取原始工具
        bundle, err := kit.ToolProvider()(ctx)
        if err != nil {
            return kit.ToolBundle{}, err
        }
        
        // 包装所有需要确认的工具
        var wrappedTools []agent.Tool
        for _, tool := range bundle.Tools {
            if m.shouldDecorate(tool) {
                wrapped := confirmable.NewConfirmableToolWrapper(
                    tool,
                    m.handler,
                    guardrails,
                )
                wrappedTools = append(wrappedTools, wrapped)
            } else {
                wrappedTools = append(wrappedTools, tool)
            }
        }
        
        return kit.ToolBundle{
            Catalog: bundle.Catalog,
            Tools:   wrappedTools,
        }, nil
    }
    
    kit.UseToolProvider(toolProvider)
    
    return nil
}

func (m *ConfirmationModule) shouldDecorate(tool agent.Tool) bool {
    // 检查工具是否在规则列表中
    for _, rule := range m.rules {
        if rule.ToolName == tool.Spec().Name {
            return true
        }
    }
    return false
}
```

### 示例 3：记忆增强模块

```go
// src/adk/modules/memory_enhance_module.go
package modules

import (
    "context"
    "time"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/memory"
)

// MemoryEnhanceModule 记忆增强模块
type MemoryEnhanceModule struct {
    name              string
    enableSummaries   bool
    clusterSize       int
    consolidateInterval time.Duration
}

func NewMemoryEnhanceModule(
    name string,
    enableSummaries bool,
    clusterSize int,
    interval time.Duration,
) *MemoryEnhanceModule {
    return &MemoryEnhanceModule{
        name:              name,
        enableSummaries:   enableSummaries,
        clusterSize:       clusterSize,
        consolidateInterval: interval,
    }
}

func (m *MemoryEnhanceModule) Name() string {
    return m.name
}

// Provision 注册增强的记忆功能
func (m *MemoryEnhanceModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 获取基础记忆提供者
    baseProvider := kit.MemoryProvider()
    if baseProvider == nil {
        return nil
    }
    
    // 创建增强的记忆提供者
    enhancedProvider := func(ctx context.Context) (kit.MemoryBundle, error) {
        // 1. 获取基础记忆
        bundle, err := baseProvider(ctx)
        if err != nil {
            return kit.MemoryBundle{}, err
        }
        
        // 2. 创建语义整合器
        if m.enableSummaries {
            consolidator := NewSemanticConsolidator(
                bundle.Session.Engine,
                m.clusterSize,
            )
            
            // 3. 启动后台任务
            go consolidator.RunPeriodically(ctx, m.consolidateInterval)
        }
        
        return bundle, nil
    }
    
    kit.UseMemoryProvider(enhancedProvider)
    return nil
}
```

---

## 最佳实践

### 1. Module 命名

```go
// ✅ 好的命名
type StreamModule struct{}
type ConfirmationModule struct{}
type MemoryEnhanceModule struct{}

// ❌ 避免的命名
type Module struct{}              // 太泛
type MyModule struct{}            // 无意义
type StreamConfirmationModule struct{}  // 职责不清
```

### 2. 单一职责

```go
// ✅ 正确：一个模块一个职责
type StreamModule struct{}      // 只负责流式传输
type ConfirmationModule struct{} // 只负责确认功能

// ❌ 错误：职责混杂
type AllInOneModule struct{}     // 什么都做
```

### 3. 配置结构

```go
// ✅ 好的配置
type StreamConfig struct {
    EnableSSE       bool
    EnableWebSocket bool
    Port            int
}

// ❌ 避免的配置
type Config struct {  // 太泛
    Option1 string
    Option2 int
}
```

### 4. 错误处理

```go
// ✅ 正确：详细的错误信息
func (m *MyModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    provider := func(ctx context.Context) (Bundle, error) {
        component, err := NewComponent()
        if err != nil {
            return Bundle{}, fmt.Errorf("my module: create component: %w", err)
        }
        return Bundle{Component: component}, nil
    }
    
    kit.UseProvider(provider)
    return nil
}

// ❌ 避免：错误信息不清晰
func (m *MyModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    provider := func(ctx context.Context) (Bundle, error) {
        component, err := NewComponent()
        if err != nil {
            return Bundle{}, err  // ❌ 丢失上下文
        }
        return Bundle{Component: component}, nil
    }
    
    kit.UseProvider(provider)
    return nil
}
```

### 5. 依赖检查

```go
// ✅ 正确：检查依赖
func (m *MyModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    baseProvider := kit.MemoryProvider()
    if baseProvider == nil {
        return fmt.Errorf("my module requires memory provider")
    }
    
    // 继续...
    return nil
}

// ❌ 避免：不检查依赖
func (m *MyModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    baseProvider := kit.MemoryProvider()
    // 直接使用，可能为 nil
    
    return nil
}
```

---

## 测试方法

### 单元测试

```go
// modules/my_module_test.go
package modules

import (
    "context"
    "testing"
    
    "github.com/Protocol-Lattice/go-agent/src/adk"
)

func TestMyModule_Name(t *testing.T) {
    module := NewMyModule("test", MyModuleConfig{})
    
    if module.Name() != "test" {
        t.Errorf("Expected name 'test', got '%s'", module.Name())
    }
}

func TestMyModule_Provision(t *testing.T) {
    ctx := context.Background()
    
    // 创建模块
    module := NewMyModule("test", MyModuleConfig{})
    
    // 创建 Kit
    kit, err := adk.New(ctx, adk.WithModule(module))
    if err != nil {
        t.Fatal(err)
    }
    
    // 执行 Bootstrap
    err = kit.Bootstrap(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    // 验证 Provider 已注册
    // ...
}
```

### 集成测试

```go
// modules/integration_test.go
package modules

import (
    "context"
    "testing"
    
    "github.com/Protocol-Lattice/go-agent/src/adk"
)

func TestStreamModule_Integration(t *testing.T) {
    ctx := context.Background()
    
    // 创建多个模块
    streamModule := NewStreamModule("stream", StreamConfig{
        EnableSSE: true,
        Port:      8080,
    })
    
    memoryModule := modules.NewMemoryModule(...)
    
    // 创建 Kit
    kit, err := adk.New(ctx,
        adk.WithModule(streamModule),
        adk.WithModule(memoryModule),
    )
    if err != nil {
        t.Fatal(err)
    }
    
    // 构建 Agent
    agent, err := kit.BuildAgent(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    // 测试流式功能
    stream, err := agent.GenerateStream(ctx, "session1", "Hello")
    if err != nil {
        t.Fatal(err)
    }
    
    // 验证流式输出
    // ...
}
```

---

## 常见问题

### Q1: Module 和 Provider 的区别？

**Module** = 功能单元（包含配置 + 逻辑）  
**Provider** = 工厂函数（创建组件）

```go
// Module 包含配置
type StreamModule struct {
    config StreamConfig
}

// Provider 只是工厂
type StreamProvider func(ctx) (StreamBundle, error)
```

### Q2: 如何决定创建新 Module？

**适合创建 Module 的场景**：
- ✅ 功能相对独立
- ✅ 可能需要启用/禁用
- ✅ 有独立的配置
- ✅ 可复用

**不适合**：
- ❌ 功能太小（几个函数）
- ❌ 强依赖其他模块
- ❌ 只使用一次

### Q3: Module 之间如何通信？

**推荐方式**：通过 Kit 提供的 Provider

```go
// Module A 注册 Provider
kit.UseProviderA(providerA)

// Module B 使用 Provider A
providerA := kit.ProviderA()
```

**避免**：Module 之间直接依赖

```go
// ❌ 避免
type ModuleB struct {
    moduleA *ModuleA  // 直接依赖
}
```

### Q4: 如何处理 Module 加载顺序？

**ADK 按注册顺序执行 Module.Provision()**

```go
adk.New(ctx,
    adk.WithModule(&ModuleA{}),  // 先执行
    adk.WithModule(&ModuleB{}),  // 后执行
)
```

**如果 Module B 依赖 Module A**：
- 确保 Module A 先注册
- 或者在 Provision() 中检查依赖

---

## 总结

### Module 开发流程

```
1. 定义 Module 结构
    │
2. 实现 Name() 方法
    │
3. 实现 Provision() 方法
    │   ├─ 创建 Provider
    │   └─ 注册 Provider
    │
4. 编写测试
    ├─ 单元测试
    └─ 集成测试
    │
5. 编写文档
```

### 最佳实践总结

| 实践 | 说明 |
|------|------|
| **单一职责** | 一个模块一个职责 |
| **明确命名** | 名称反映功能 |
| **配置分离** | 使用配置结构 |
| **错误处理** | 详细的错误信息 |
| **依赖检查** | 检查必需依赖 |
| **充分测试** | 单元 + 集成测试 |

### 下一步

- 阅读 [ADK_PROVIDER_REFERENCE.md](./ADK_PROVIDER_REFERENCE.md) - Provider 参考手册
- 阅读 [DECORATOR_PATTERN_GUIDE.md](./DECORATOR_PATTERN_GUIDE.md) - 装饰器模式指南
- 开始开发你的第一个 Module！

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
