# 装饰器模式指南

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: 设计指南  
**相关文档**: [CONFIRMABLE_AGENT_DESIGN_V2.md](./CONFIRMABLE_AGENT_DESIGN_V2.md), [ADK_MODULE_DEVELOPMENT_GUIDE.md](./ADK_MODULE_DEVELOPMENT_GUIDE.md)

---

## 📋 目录

1. [什么是装饰器模式](#什么是装饰器模式)
2. [为什么使用装饰器](#为什么使用装饰器)
3. [Go 语言实现](#go 语言实现)
4. [在 go-agent 中的应用](#在 go-agent 中的应用)
5. [完整示例](#完整示例)
6. [最佳实践](#最佳实践)
7. [常见误区](#常见误区)

---

## 什么是装饰器模式

### 通俗解释

**装饰器模式**就像**给礼物包装**：

```
原始礼物（Tool）
    │
    ├─ 包装纸层（装饰器 1）- 功能：美观
    ├─ 丝带层（装饰器 2）- 功能：装饰
    └─ 卡片层（装饰器 3）- 功能：祝福

最终：包装好的礼物（功能增强，但本质还是原来的礼物）
```

**关键特点**：
- ✅ 不改变礼物本身
- ✅ 可以叠加多层包装
- ✅ 每层包装都有额外功能
- ✅ 可以随时拆除包装

### 正式定义

**装饰器模式（Decorator Pattern）**是一种结构型设计模式，允许在不修改原有对象的情况下，通过包装来动态地给对象添加新功能。

**UML 结构图**：
```
┌─────────────────────────────────────────────────────────────┐
│  Component (接口)                                           │
│  - operation()                                              │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
         ▼                       ▼
┌─────────────────┐     ┌─────────────────┐
│ ConcreteComponent│     │   Decorator    │
│ (原始实现)       │     │ (抽象装饰器)   │
│                 │     │ - component     │
│ + operation()   │     │ + operation()   │
└─────────────────┘     └────────┬────────┘
                                 │
                                 │ 继承
                                 ▼
                    ┌─────────────────────────┐
                    │ ConcreteDecorator A/B   │
                    │ (具体装饰器)            │
                    │                         │
                    │ + operation()           │
                    │   ├─ 添加功能           │
                    │   └─ component.operation() │
                    └─────────────────────────┘
```

---

## 为什么使用装饰器

### 传统继承的问题

```go
// ❌ 继承方式：类爆炸
type Tool interface {
    Invoke()
}

type FileTool struct{}
func (t *FileTool) Invoke() {}

// 需要确认功能？创建子类
type ConfirmableFileTool struct {
    FileTool  // 继承
}
func (t *ConfirmableFileTool) Invoke() {
    // 添加确认逻辑
    t.FileTool.Invoke()
}

// 需要日志功能？再创建子类
type LoggingFileTool struct {
    FileTool
}

// 需要确认 + 日志？再创建子类
type ConfirmableLoggingFileTool struct {
    ConfirmableFileTool
}
// 类爆炸：每增加一个功能，就需要 N 个子类
```

**问题**：
- ❌ 类数量爆炸（2^n 增长）
- ❌ 编译时静态绑定，无法动态添加功能
- ❌ 难以维护

### 装饰器模式的优势

```go
// ✅ 装饰器方式：灵活组合
type Tool interface {
    Invoke()
}

// 原始工具
type FileTool struct{}

// 装饰器 1：确认
type ConfirmableDecorator struct {
    tool Tool  // ← 持有接口引用，不是具体实现
}

// 装饰器 2：日志
type LoggingDecorator struct {
    tool Tool
}

// 装饰器 3：缓存
type CachingDecorator struct {
    tool Tool
}

// 使用时灵活组合
tool := &FileTool{}
tool = &ConfirmableDecorator{tool: tool}     // 添加确认
tool = &LoggingDecorator{tool: tool}         // 添加日志
tool = &CachingDecorator{tool: tool}         // 添加缓存

// 最终：Caching(Logging(Confirmable(FileTool)))
```

**优势**：
- ✅ 功能组合灵活（运行时动态添加）
- ✅ 类数量线性增长（n 个装饰器）
- ✅ 符合开闭原则（对扩展开放，对修改关闭）

---

## Go 语言实现

### 基础结构

```go
package decorator

// 1. 定义组件接口
type Component interface {
    Operation() string
}

// 2. 具体组件（原始实现）
type ConcreteComponent struct{}

func (c *ConcreteComponent) Operation() string {
    return "原始功能"
}

// 3. 抽象装饰器（持有组件引用）
type Decorator struct {
    component Component
}

func (d *Decorator) Operation() string {
    return d.component.Operation()
}

// 4. 具体装饰器 A
type ConcreteDecoratorA struct {
    Decorator
}

func (d *ConcreteDecoratorA) Operation() string {
    // 添加额外功能
    result := d.Decorator.Operation()
    return "装饰器 A + " + result
}

// 5. 具体装饰器 B
type ConcreteDecoratorB struct {
    Decorator
}

func (d *ConcreteDecoratorB) Operation() string {
    // 添加额外功能
    result := d.Decorator.Operation()
    return "装饰器 B + " + result
}
```

### 使用方式

```go
func main() {
    // 1. 创建原始组件
    component := &ConcreteComponent{}
    
    // 2. 逐层包装
    component = &ConcreteDecoratorA{Decorator{component}}
    component = &ConcreteDecoratorB{Decorator{component}}
    
    // 3. 使用（和原始组件一样的接口）
    result := component.Operation()
    // → "装饰器 B + 装饰器 A + 原始功能"
}
```

---

## 在 go-agent 中的应用

### 场景：Tool 确认功能

#### 需求

在 Tool 执行前，自动检查是否需要用户确认：
- 分析用户提示词（"先问我"、"需要确认"）
- 检查预定义规则（金额>1000）
- 等待用户确认
- 执行或拒绝

#### 装饰器方案

```go
// confirmable/confirmable_tool.go
package confirmable

import (
    "context"
    "github.com/Protocol-Lattice/go-agent"
)

// 1. 装饰器结构（实现 Tool 接口）
type ConfirmableToolWrapper struct {
    tool       agent.Tool              // ← 持有原始 Tool
    handler    ConfirmationHandler     // ← 确认处理器
    guardrails *ConfirmationGuardrails // ← 规则检查
}

// 2. 构造函数
func NewConfirmableToolWrapper(
    tool agent.Tool,
    handler ConfirmationHandler,
    guardrails *ConfirmationGuardrails,
) *ConfirmableToolWrapper {
    return &ConfirmableToolWrapper{
        tool:       tool,
        handler:    handler,
        guardrails: guardrails,
    }
}

// 3. 实现 Tool 接口（委托给原始工具）
func (w *ConfirmableToolWrapper) Spec() agent.ToolSpec {
    return w.tool.Spec()  // ← 委托
}

// 4. 添加额外功能（确认逻辑）
func (w *ConfirmableToolWrapper) Invoke(
    ctx context.Context,
    req agent.ToolRequest,
) (agent.ToolResponse, error) {
    // 额外功能：检查确认
    confirmationReq := w.guardrails.ShouldConfirm(req)
    
    if confirmationReq != nil {
        // 等待用户确认
        confirmed, err := w.handler(ctx, *confirmationReq)
        if err != nil {
            return agent.ToolResponse{}, err
        }
        
        if !confirmed {
            return agent.ToolResponse{
                Content: "Operation denied by user",
            }, nil
        }
    }
    
    // 委托调用：原始工具
    return w.tool.Invoke(ctx, req)
}
```

### 装饰器链

```go
// 可以叠加多个装饰器
tool := &FileDeleteTool{}  // 原始工具

// 包装确认装饰器
tool = &ConfirmableToolWrapper{
    tool: tool,
    handler: confirmHandler,
    guardrails: guardrails,
}

// 包装日志装饰器
tool = &LoggingToolWrapper{
    tool: tool,
    logger: logger,
}

// 包装缓存装饰器
tool = &CachingToolWrapper{
    tool: tool,
    cache: cache,
}

// 最终：Caching(Logging(Confirmable(FileDeleteTool)))
```

### 调用流程

```
用户调用：tool.Invoke(ctx, req)
    │
    ▼
┌─────────────────────────────────────┐
│  CachingToolWrapper.Invoke()        │
│  ┌───────────────────────────────┐  │
│  │ 1. 检查缓存                   │  │
│  │    [命中] → 返回缓存结果      │  │
│  │    [未命中] → 继续            │  │
│  └──────────────┬────────────────┘  │
│                 ▼                   │
│  ┌───────────────────────────────┐  │
│  │  LoggingToolWrapper.Invoke()  │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │ 2. 记录日志             │  │  │
│  │  │   "Calling file.delete" │  │  │
│  │  └──────────┬──────────────┘  │  │
│  │             ▼                 │  │
│  │  ┌─────────────────────────┐  │  │
│  │  │ ConfirmableToolWrapper  │  │  │
│  │  │ ┌─────────────────────┐ │  │  │
│  │  │ │ 3. 检查确认         │ │  │  │
│  │  │ │   [需要] → 等待确认 │ │  │  │
│  │  │ │   [不需要] → 继续   │ │  │  │
│  │  │ └──────────┬──────────┘ │  │  │
│  │  │            ▼            │  │  │
│  │  │ ┌─────────────────────┐ │  │  │
│  │  │ │ FileDeleteTool      │ │  │  │
│  │  │ │ 4. 执行删除         │ │  │  │
│  │  │ └─────────────────────┘ │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
    │
    ▼
返回结果（逐层返回）
    │
┌─────────────────────────────────────┐
│  Caching: 写入缓存                  │
│  Logging: 记录结果                  │
│  Confirmable: 无                    │
└─────────────────────────────────────┘
```

---

## 完整示例

### 示例 1：基础装饰器

```go
package main

import "fmt"

// 1. 组件接口
type Messenger interface {
    Send(message string) string
}

// 2. 具体组件
type SMSMessenger struct{}

func (s *SMSMessenger) Send(message string) string {
    return "发送短信：" + message
}

// 3. 抽象装饰器
type MessengerDecorator struct {
    messenger Messenger
}

func (d *MessengerDecorator) Send(message string) string {
    return d.messenger.Send(message)
}

// 4. 具体装饰器 A：邮件
type EmailDecorator struct {
    MessengerDecorator
}

func (d *EmailDecorator) Send(message string) string {
    // 添加额外功能
    emailResult := "发送邮件：" + message
    // 委托调用
    smsResult := d.MessengerDecorator.Send(message)
    return emailResult + " + " + smsResult
}

// 5. 具体装饰器 B：推送
type PushDecorator struct {
    MessengerDecorator
}

func (d *PushDecorator) Send(message string) string {
    pushResult := "发送推送：" + message
    smsResult := d.MessengerDecorator.Send(message)
    return pushResult + " + " + smsResult
}

// 6. 使用
func main() {
    // 创建原始组件
    messenger := &SMSMessenger{}
    
    // 逐层包装
    messenger = &EmailDecorator{MessengerDecorator{messenger}}
    messenger = &PushDecorator{MessengerDecorator{messenger}}
    
    // 使用（接口不变）
    result := messenger.Send("你好")
    fmt.Println(result)
    // → "发送推送：你好 + 发送邮件：你好 + 发送短信：你好"
}
```

### 示例 2：go-agent 确认装饰器

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/Protocol-Lattice/go-agent"
)

// 1. 原始工具
type FileDeleteTool struct{}

func (t *FileDeleteTool) Spec() agent.ToolSpec {
    return agent.ToolSpec{
        Name:        "file.delete",
        Description: "删除文件",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "path": map[string]any{
                    "type": "string",
                    "description": "文件路径",
                },
            },
            "required": []string{"path"},
        },
    }
}

func (t *FileDeleteTool) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
    path := req.Arguments["path"].(string)
    fmt.Printf("删除文件：%s\n", path)
    return agent.ToolResponse{Content: "File deleted"}, nil
}

// 2. 确认装饰器
type ConfirmableToolWrapper struct {
    tool    agent.Tool
    handler func(ctx context.Context, msg string) bool
}

func NewConfirmableToolWrapper(
    tool agent.Tool,
    handler func(ctx context.Context, msg string) bool,
) *ConfirmableToolWrapper {
    return &ConfirmableToolWrapper{
        tool:    tool,
        handler: handler,
    }
}

func (w *ConfirmableToolWrapper) Spec() agent.ToolSpec {
    return w.tool.Spec()
}

func (w *ConfirmableToolWrapper) Invoke(ctx context.Context, req agent.ToolRequest) (agent.ToolResponse, error) {
    // 检查用户提示词
    userPrompt := getUserPrompt(req.SessionID)
    if strings.Contains(userPrompt, "先问我") || strings.Contains(userPrompt, "需要确认") {
        // 等待确认
        if !w.handler(ctx, "确定要执行此操作吗？") {
            return agent.ToolResponse{Content: "Denied by user"}, nil
        }
    }
    
    // 执行原始工具
    return w.tool.Invoke(ctx, req)
}

// 辅助函数（模拟）
func getUserPrompt(sessionID string) string {
    return "帮我删除这个文件，但需要先问我确认"
}

// 3. 使用
func main() {
    ctx := context.Background()
    
    // 创建原始工具
    tool := &FileDeleteTool{}
    
    // 包装确认装饰器
    wrapped := NewConfirmableToolWrapper(tool, func(ctx context.Context, msg string) bool {
        fmt.Printf("确认：%s\n", msg)
        fmt.Print("是否确认？(y/n): ")
        var input string
        fmt.Scanln(&input)
        return input == "y"
    })
    
    // 使用（接口不变）
    resp, err := wrapped.Invoke(ctx, agent.ToolRequest{
        SessionID: "session1",
        Arguments: map[string]any{
            "path": "/path/to/file",
        },
    })
    
    if err != nil {
        fmt.Printf("错误：%v\n", err)
    } else {
        fmt.Printf("结果：%s\n", resp.Content)
    }
}
```

---

## 最佳实践

### 1. 装饰器命名

```go
// ✅ 好的命名
type ConfirmableToolWrapper struct{}
type LoggingToolWrapper struct{}
type CachingToolWrapper struct{}

// ❌ 避免的命名
type ToolWithConfirmation struct{}  // 像继承
type ConfirmedTool struct{}         // 含义不清
```

### 2. 保持接口一致

```go
// ✅ 正确：装饰器实现相同接口
type Decorator struct {
    component Component
}

func (d *Decorator) Operation() string {
    return d.component.Operation()  // ← 委托
}

// ❌ 错误：改变接口
type Decorator struct {
    component Component
}

func (d *Decorator) OperationWithLogging() string {  // ← 新接口
    log.Println("Calling")
    return d.component.Operation()
}
```

### 3. 装饰器顺序

```go
// 装饰器顺序可能影响结果
tool := &FileTool{}
tool = &ConfirmableDecorator{tool: tool}   // 先确认
tool = &LoggingDecorator{tool: tool}       // 后日志

// vs
tool := &FileTool{}
tool = &LoggingDecorator{tool: tool}       // 先日志
tool = &ConfirmableDecorator{tool: tool}   // 后确认

// 选择：根据业务需求决定顺序
```

### 4. 装饰器工厂

```go
// ✅ 使用工厂函数简化创建
func NewConfirmableDecorator(tool agent.Tool, handler ConfirmationHandler) agent.Tool {
    return &ConfirmableToolWrapper{
        tool:    tool,
        handler: handler,
    }
}

// 使用
tool := NewConfirmableDecorator(originalTool, handler)
```

### 5. 装饰器组合

```go
// ✅ 提供组合函数
func WithDecorators(tool agent.Tool, decorators ...func(agent.Tool) agent.Tool) agent.Tool {
    for _, decorator := range decorators {
        tool = decorator(tool)
    }
    return tool
}

// 使用
tool := WithDecorators(
    originalTool,
    WithConfirmation(handler),
    WithLogging(logger),
    WithCaching(cache),
)
```

---

## 常见误区

### 误区 1：装饰器 vs 继承

```go
// ❌ 错误：使用继承
type ConfirmableFileTool struct {
    FileTool  // 继承
}

// ✅ 正确：使用装饰器
type ConfirmableToolWrapper struct {
    tool agent.Tool  // ← 持有接口引用
}
```

### 误区 2：忘记委托

```go
// ❌ 错误：忘记委托调用
func (w *ConfirmableToolWrapper) Invoke() (ToolResponse, error) {
    if !confirm() {
        return ToolResponse{Content: "Denied"}, nil
    }
    // 忘记调用原始工具
    return ToolResponse{}, nil
}

// ✅ 正确：委托调用
func (w *ConfirmableToolWrapper) Invoke() (ToolResponse, error) {
    if !confirm() {
        return ToolResponse{Content: "Denied"}, nil
    }
    return w.tool.Invoke()  // ← 委托
}
```

### 误区 3：过度装饰

```go
// ❌ 避免：过度装饰（性能问题）
tool := originalTool
for i := 0; i < 100; i++ {
    tool = &LoggingDecorator{tool: tool}
}

// ✅ 合理：适度装饰
tool := WithDecorators(
    originalTool,
    WithLogging(logger),
    WithCaching(cache),
)
```

---

## 总结

### 装饰器模式核心

| 特点 | 说明 |
|------|------|
| **结构型模式** | 通过组合实现功能增强 |
| **接口一致** | 装饰器和原始对象实现相同接口 |
| **动态添加** | 运行时动态添加功能 |
| **可组合** | 可以叠加多层装饰器 |
| **开闭原则** | 对扩展开放，对修改关闭 |

### 在 go-agent 中的应用

| 装饰器 | 功能 | 位置 |
|--------|------|------|
| **ConfirmableToolWrapper** | 确认功能 | `confirmable/` |
| **LoggingToolWrapper** | 日志功能 | （可选） |
| **CachingToolWrapper** | 缓存功能 | （可选） |

### 下一步

- 阅读 [CONFIRMABLE_AGENT_DESIGN_V2.md](./CONFIRMABLE_AGENT_DESIGN_V2.md) - 确认功能详细设计
- 阅读 [ADK_MODULE_DEVELOPMENT_GUIDE.md](./ADK_MODULE_DEVELOPMENT_GUIDE.md) - 如何开发 ADK 模块
- 阅读 [ADK_PROVIDER_REFERENCE.md](./ADK_PROVIDER_REFERENCE.md) - Provider 参考手册

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
