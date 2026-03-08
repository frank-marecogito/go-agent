# 可确认 Agent 设计文档（装饰器模式）

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 2.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: 设计方案  
**相关任务**: EXT-015

---

## 📋 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [装饰器模式详解](#装饰器模式详解)
4. [动态确认需求提取](#动态确认需求提取)
5. [核心组件实现](#核心组件实现)
6. [ADK Module 配置](#adk-module-配置)
7. [使用示例](#使用示例)
8. [通知渠道](#通知渠道)
9. [授权管理](#授权管理)
10. [测试计划](#测试计划)

---

## 概述

### 背景

在 AI Agent 实际应用中，用户需要在**便利性**和**控制权**之间取得平衡：
- 敏感操作（文件删除、数据库修改、大额支付）需要用户确认
- 重复性操作希望一次授权后自动执行
- 授权后仍希望收到通知并保留撤销权利

### 设计原则

| 原则 | 说明 |
|------|------|
| **零侵入** | 不修改 Tool 接口或 ToolRequest |
| **动态提取** | 分析用户提示词，动态判断是否需要确认 |
| **可组合** | 可叠加多个装饰器（确认/日志/缓存等） |
| **模块化** | 通过 ADK Module 自动配置 |

### 与原有方案对比

| 特性 | 原有方案 | 装饰器方案（新） | 改进 |
|------|----------|----------------|------|
| **侵入性** | 修改 ToolRequest | 零侵入 | ✅ 完全兼容 |
| **配置方式** | System Prompt 硬编码 | 动态提取 + 规则配置 | ✅ 灵活 |
| **确认触发** | 固定规则 | 动态分析用户提示词 | ✅ 智能 |
| **模块化** | 紧耦合 | ADK Module 自动配置 | ✅ 可插拔 |
| **可组合** | 否 | 可叠加多个装饰器 | ✅ 灵活 |

---

## 架构设计

### 三层架构

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: 用户提示词分析                                     │
│                                                             │
│  用户："帮我删除这个文件，但需要先问我确认"                  │
│                                                             │
│  LLM 理解意图，生成 Tool 调用（无需特殊参数）                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Guardrails 动态检查                                │
│                                                             │
│  1. 获取用户原始提示词                                       │
│  2. 检查是否包含"先问我"、"需要确认"等                       │
│  3. 检查预定义规则（如金额>1000）                            │
│  4. 返回是否需要确认                                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: ConfirmableToolWrapper 装饰器                      │
│                                                             │
│  1. 检查是否需要确认                                         │
│  2. 需要 → 等待用户确认                                      │
│  3. 不需要/已确认 → 执行原始工具                             │
└─────────────────────────────────────────────────────────────┘
```

### 装饰器模式

```
原始工具（FileDeleteTool）
    │
    └─ ConfirmableToolWrapper（确认装饰器）
        ├─ 额外功能：确认逻辑
        └─ 委托调用：原始工具.Invoke()
```

**关键点**：
- ✅ 不改变原始工具
- ✅ 可以叠加多层装饰器
- ✅ 每层都有额外功能
- ✅ 随时可以移除装饰器

---

## 装饰器模式详解

### 什么是装饰器？

**装饰器** = **包装器** = **功能增强层**

```
┌─────────────────────────────────────────────────────────────┐
│  原始工具：FileDeleteTool                                   │
│                                                             │
│  func (t *FileDeleteTool) Invoke(ctx, req) (ToolResponse, error) {
│      os.Remove(req.Arguments["path"].(string))             │
│      return ToolResponse{Content: "Deleted"}, nil          │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
```

```
┌─────────────────────────────────────────────────────────────┐
│  装饰器：ConfirmableToolWrapper                             │
│                                                             │
│  type ConfirmableToolWrapper struct {                       │
│      tool    Tool              // ← 持有原始工具引用         │
│      handler ConfirmationHandler                            │
│      guardrails *ConfirmationGuardrails                     │
│  }                                                          │
│                                                             │
│  func (w *ConfirmableToolWrapper) Invoke(...) {            │
│      // 1. 额外功能：检查确认                               │
│      if w.guardrails.ShouldConfirm(req) {                  │
│          if !w.waitForConfirmation(ctx, req) {             │
│              return ToolResponse{Content: "Denied"}, nil   │
│          }                                                  │
│      }                                                      │
│      // 2. 委托调用：原始工具                               │
│      return w.tool.Invoke(ctx, req)                        │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
```

### 装饰器链（多层包装）

```
原始工具：FileDeleteTool
    │
    ├─ ConfirmableToolWrapper（确认装饰器）
    │   └─ 功能：等待用户确认
    │
    ├─ LoggingToolWrapper（日志装饰器）
    │   └─ 功能：记录调用日志
    │
    └─ CachingToolWrapper（缓存装饰器）
        └─ 功能：缓存结果

调用流程：
CachingToolWrapper.Invoke()
  ├─ 检查缓存
  └─ LoggingToolWrapper.Invoke()
      ├─ 记录日志
      └─ ConfirmableToolWrapper.Invoke()
          ├─ 等待确认
          └─ FileDeleteTool.Invoke() ← 原始工具
```

---

## 动态确认需求提取

### Guardrails 动态检查

```go
// confirmable/guardrails.go
type ConfirmationGuardrails struct {
    rules []*DynamicConfirmationRule
}

type DynamicConfirmationRule struct {
    ToolName  string
    Condition func(req ToolRequest) bool  // 动态条件
    Message   string                       // 确认消息
}

func (g *ConfirmationGuardrails) ShouldConfirm(req ToolRequest) *ConfirmationRequest {
    // 1. 获取用户原始提示词
    userPrompt := g.getUserPrompt(req.SessionID)
    
    // 2. 检查用户是否要求确认
    if strings.Contains(userPrompt, "先问我") || 
       strings.Contains(userPrompt, "需要确认") ||
       strings.Contains(userPrompt, "ask me first") {
        return &ConfirmationRequest{
            ToolName:  req.Arguments["tool"].(string),
            Arguments: req.Arguments,
            Message:   "用户要求确认此操作",
        }
    }
    
    // 3. 检查预定义规则
    for _, rule := range g.rules {
        if rule.ToolName == req.Arguments["tool"] {
            if rule.Condition == nil || rule.Condition(req) {
                return &ConfirmationRequest{
                    ToolName:  rule.ToolName,
                    Arguments: req.Arguments,
                    Message:   rule.Message,
                }
            }
        }
    }
    
    return nil
}
```

### 规则配置示例

```yaml
# rules.yaml
rules:
  - tool: file.delete
    message: "确定要删除这个文件吗？"
    
  - tool: file.write
    message: "确定要写入这个文件吗？"
    
  - tool: payment.process
    message: "大额支付需要确认"
    condition: "amount > 1000"
    
  - tool: database.execute
    message: "SQL 执行需要确认"
    condition: |
      sql.match(/^(DELETE|DROP|TRUNCATE)/i)
```

---

## 核心组件实现

### 1. ConfirmableToolWrapper

```go
// confirmable/confirmable_tool.go
package confirmable

import (
    "context"
    "github.com/Protocol-Lattice/go-agent"
)

type ConfirmableToolWrapper struct {
    tool       agent.Tool
    handler    ConfirmationHandler
    guardrails *ConfirmationGuardrails
}

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

// 实现 Tool 接口
func (w *ConfirmableToolWrapper) Spec() agent.ToolSpec {
    return w.tool.Spec()  // ← 委托给原始工具
}

func (w *ConfirmableToolWrapper) Invoke(
    ctx context.Context,
    req agent.ToolRequest,
) (agent.ToolResponse, error) {
    // 1. Guardrails 动态检查
    confirmationReq := w.guardrails.ShouldConfirm(req)
    
    if confirmationReq != nil {
        // 2. 等待确认
        confirmed, err := w.handler(ctx, *confirmationReq)
        if err != nil {
            return agent.ToolResponse{}, err
        }
        
        if !confirmed {
            return agent.ToolResponse{
                Content: "Operation denied by user",
                Metadata: map[string]string{
                    "reason": "user_denied",
                },
            }, nil
        }
    }
    
    // 3. 执行原始工具（零侵入）
    return w.tool.Invoke(ctx, req)
}
```

### 2. ConfirmationHandler

```go
// confirmable/confirmation_handler.go
type ConfirmationHandler func(
    ctx context.Context,
    req ConfirmationRequest,
) (bool, error)

type ConfirmationRequest struct {
    ToolName  string
    Arguments map[string]any
    Message   string
    Options   []ConfirmationOption
}

type ConfirmationOption struct {
    ID    string `json:"id"`
    Label string `json:"label"`
    Type  string `json:"type"` // "approve_once" | "approve_all" | "deny"
}

// 创建确认处理器
func NewConfirmationHandler(
    notifiers ...NotificationChannel,
) ConfirmationHandler {
    return func(ctx context.Context, req ConfirmationRequest) (bool, error) {
        // 1. 发送通知到所有渠道
        for _, notifier := range notifiers {
            notifier.Send(ctx, req)
        }
        
        // 2. 等待用户响应
        response := <-waitForUserResponse(req.ID)
        
        // 3. 处理响应
        switch response.OptionID {
        case "approve_once":
            return true, nil
        case "approve_all":
            // 创建永久授权
            createAuthorization(req)
            return true, nil
        case "deny":
            return false, nil
        }
        
        return false, nil
    }
}
```

### 3. AuthorizationStore

```go
// confirmable/authorization_store.go
type AuthorizationStore interface {
    Create(ctx context.Context, auth *Authorization) error
    Check(ctx context.Context, req ConfirmationRequest) (bool, error)
    Revoke(ctx context.Context, id string) error
    List(ctx context.Context, userID string) ([]*Authorization, error)
}

type Authorization struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    ToolName  string    `json:"tool_name"`
    Condition string    `json:"condition,omitempty"` // SQL-like condition
    ExpiresAt time.Time `json:"expires_at,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

func (s *AuthorizationStore) Check(ctx context.Context, req ConfirmationRequest) (bool, error) {
    // 检查永久授权
    auths, err := s.List(ctx, req.UserID)
    if err != nil {
        return false, err
    }
    
    for _, auth := range auths {
        if auth.ToolName == req.ToolName {
            if auth.ExpiresAt.IsZero() || auth.ExpiresAt.After(time.Now()) {
                if auth.Condition == "" || evalCondition(auth.Condition, req.Arguments) {
                    return true, nil
                }
            }
        }
    }
    
    return false, nil
}
```

---

## ADK Module 配置

### ConfirmationModule

```go
// src/adk/modules/confirmable_module.go
package modules

import (
    "context"
    "github.com/Protocol-Lattice/go-agent/src/adk"
    agent "github.com/Protocol-Lattice/go-agent"
    "your-module/confirmable"
)

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

func (m *ConfirmationModule) Provision(
    ctx context.Context,
    kit *adk.AgentDevelopmentKit,
) error {
    if !m.enabled {
        return nil
    }
    
    // 创建 Guardrails
    guardrails := confirmable.NewConfirmationGuardrails(m.rules)
    
    // 注册 ToolProvider，自动包装所有工具
    kit.UseToolProvider(func(ctx context.Context) (adk.ToolBundle, error) {
        // 获取原始工具
        bundle, err := kit.ToolProvider()(ctx)
        if err != nil {
            return adk.ToolBundle{}, err
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
        
        return adk.ToolBundle{
            Catalog: bundle.Catalog,
            Tools:   wrappedTools,
        }, nil
    })
    
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

### 使用方式

```go
package main

import (
    "context"
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
    "your-module/confirmable"
)

func main() {
    ctx := context.Background()
    
    // 1. 定义确认规则
    rules := []*confirmable.ConfirmationRule{
        {
            ToolName:    "file.delete",
            Description: "文件删除操作",
        },
        {
            ToolName:    "payment.process",
            Description: "大额支付需要确认",
            Condition: func(args map[string]any) bool {
                if amount, ok := args["amount"].(float64); ok {
                    return amount > 1000
                }
                return false
            },
        },
    }
    
    // 2. 创建确认处理器
    handler := confirmable.NewConfirmationHandler(
        confirmable.WithWebNotification(),
        confirmable.WithEmailNotification("smtp.example.com"),
    )
    
    // 3. 创建 ADK，注册确认模块
    kit, err := adk.New(ctx,
        adk.WithModule(modules.NewConfirmationModule(
            "confirmation",
            rules,
            handler,
        )),
        // ... 其他模块
    )
    
    // 4. 构建 Agent（工具已自动包装）
    agent, err := kit.BuildAgent(ctx)
    
    // 5. 使用 Agent
    // 用户："帮我删除这个文件，但需要先问我确认"
    resp, err := agent.Generate(ctx, "session1", "帮我删除这个文件，但需要先问我确认")
}
```

---

## 通知渠道

### Web 通知（SSE/WebSocket）

```go
// confirmable/notify/web.go
type WebNotifier struct {
    hub *WebSocketHub
}

func (n *WebNotifier) Send(ctx context.Context, req ConfirmationRequest) error {
    msg := WebMessage{
        Type:    "confirmation_request",
        Request: req,
        Options: []ConfirmationOption{
            {ID: "approve_once", Label: "本次可以"},
            {ID: "approve_all", Label: "未来均可以"},
            {ID: "deny", Label: "拒绝"},
        },
    }
    
    return n.hub.Broadcast(req.UserID, msg)
}
```

### Email 通知

```go
// confirmable/notify/email.go
type EmailNotifier struct {
    smtpClient *smtp.Client
    from       string
}

func (n *EmailNotifier) Send(ctx context.Context, req ConfirmationRequest) error {
    // 构建 HTML 邮件
    html := buildConfirmationEmail(req)
    
    // 发送邮件
    return n.smtpClient.Send(n.from, []string{req.UserID}, []byte(html))
}
```

### Slack 通知

```go
// confirmable/notify/slack.go
type SlackNotifier struct {
    webhookURL string
}

func (n *SlackNotifier) Send(ctx context.Context, req ConfirmationRequest) error {
    msg := slack.WebhookMessage{
        Attachments: []slack.Attachment{
            {
                Color: "#FFA500",
                Title: "确认请求",
                Text:  req.Message,
                Actions: []slack.AttachmentAction{
                    {
                        Type:  "button",
                        Text:  "本次可以",
                        Style: "primary",
                        URL:   buildApprovalURL(req.ID, "approve_once"),
                    },
                    {
                        Type:  "button",
                        Text:  "未来均可以",
                        Style: "primary",
                        URL:   buildApprovalURL(req.ID, "approve_all"),
                    },
                    {
                        Type:  "button",
                        Text:  "拒绝",
                        Style: "danger",
                        URL:   buildApprovalURL(req.ID, "deny"),
                    },
                },
            },
        },
    }
    
    return slack.PostWebhook(n.webhookURL, &msg)
}
```

---

## 授权管理

### 授权类型

| 类型 | 说明 | 过期时间 |
|------|------|----------|
| **本次有效** | 仅本次执行有效 | 执行后立即失效 |
| **会话有效** | 当前会话期间有效 | 会话结束时失效 |
| **永久有效** | 永久有效（可撤销） | 永不过期 |
| **时效授权** | 指定时间内有效 | 指定时间后失效 |

### 授权 CRUD

```go
// 创建授权
func CreateAuthorization(
    ctx context.Context,
    userID string,
    toolName string,
    scope AuthorizationScope,
    condition string,
    duration time.Duration,
) (*Authorization, error) {
    auth := &Authorization{
        ID:        uuid.New().String(),
        UserID:    userID,
        ToolName:  toolName,
        Condition: condition,
        CreatedAt: time.Now(),
    }
    
    switch scope {
    case ScopeSession:
        auth.ExpiresAt = getSessionEnd(userID)
    case ScopePermanent:
        auth.ExpiresAt = time.Time{} // 永不过期
    case ScopeTimed:
        auth.ExpiresAt = time.Now().Add(duration)
    }
    
    return auth, store.Create(ctx, auth)
}

// 检查授权
func CheckAuthorization(
    ctx context.Context,
    userID string,
    req ConfirmationRequest,
) (bool, error) {
    auths, err := store.List(ctx, userID)
    if err != nil {
        return false, err
    }
    
    for _, auth := range auths {
        if auth.ToolName == req.ToolName {
            // 检查过期
            if !auth.ExpiresAt.IsZero() && auth.ExpiresAt.Before(time.Now()) {
                continue
            }
            
            // 检查条件
            if auth.Condition == "" || evalCondition(auth.Condition, req.Arguments) {
                return true, nil
            }
        }
    }
    
    return false, nil
}

// 撤销授权
func RevokeAuthorization(ctx context.Context, id string) error {
    return store.Revoke(ctx, id)
}
```

---

## 测试计划

### 单元测试

```go
// confirmable/confirmable_tool_test.go
func TestConfirmableToolWrapper_ShouldConfirm(t *testing.T) {
    // 创建 Mock 工具
    mockTool := &MockTool{}
    
    // 创建 Guardrails
    guardrails := NewConfirmationGuardrails([]*ConfirmationRule{
        {ToolName: "file.delete"},
    })
    
    // 创建装饰器
    wrapper := NewConfirmableToolWrapper(mockTool, nil, guardrails)
    
    // 测试：用户提示词包含"先问我"
    req := agent.ToolRequest{
        SessionID: "session1",
        Arguments: map[string]any{
            "tool": "file.delete",
        },
    }
    
    setSessionPrompt("session1", "帮我删除这个文件，但需要先问我确认")
    
    confirmationReq := wrapper.guardrails.ShouldConfirm(req)
    if confirmationReq == nil {
        t.Fatal("Expected confirmation request")
    }
}
```

### 集成测试

```go
// confirmable/integration_test.go
func TestConfirmationModule_Integration(t *testing.T) {
    ctx := context.Background()
    
    // 创建 ADK
    kit, err := adk.New(ctx,
        adk.WithModule(modules.NewConfirmationModule(
            "confirmation",
            []*confirmable.ConfirmationRule{
                {ToolName: "file.delete"},
            },
            mockConfirmationHandler,
        )),
    )
    
    // 构建 Agent
    agent, err := kit.BuildAgent(ctx)
    
    // 测试：调用需要确认的工具
    resp, err := agent.Generate(ctx, "session1", "帮我删除这个文件，但需要先问我确认")
    
    // 验证：等待了确认
    // 验证：用户确认后执行了工具
}
```

---

## 总结

### 核心优势

| 优势 | 说明 |
|------|------|
| **零侵入** | 不修改 Tool 接口或 ToolRequest |
| **动态提取** | 分析用户提示词，动态判断是否需要确认 |
| **可组合** | 可叠加多个装饰器（确认/日志/缓存等） |
| **模块化** | 通过 ADK Module 自动配置 |
| **向后兼容** | 现有工具无需修改 |

### 实施建议

1. **先实现核心库** (`confirmable/`)
2. **再实现 ADK Module** (`adk/modules/confirmable_module.go`)
3. **编写测试**（单元测试 + 集成测试）
4. **编写文档**（使用指南 + 最佳实践）

---

*文档版本：2.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
