# 可确认 Agent 实施方案

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 7 日  
**状态**: 设计方案  
**相关任务**: EXT-015

---

## 📋 目录

1. [方案概述](#方案概述)
2. [架构设计](#架构设计)
3. [核心组件](#核心组件)
4. [三层确认机制](#三层确认机制)
5. [三个确认选项](#三个确认选项)
6. [异步通知与撤销](#异步通知与撤销)
7. [Checkpoint/Restore 集成](#checkpointrestore-集成)
8. [使用示例](#使用示例)
9. [通知渠道实现](#通知渠道实现)
10. [审计日志](#审计日志)
11. [测试计划](#测试计划)

---

## 方案概述

### 背景

在 AI Agent 实际应用中，用户需要在**便利性**和**控制权**之间取得平衡：
- 敏感操作（文件删除、数据库修改、大额支付）需要用户确认
- 重复性操作希望一次授权后自动执行
- 授权后仍希望收到通知并保留撤销权利

### 目标

实现完整的 Agent 确认机制，提供：
1. ✅ **三层确认** - System Prompt + Guardrails + Tool 拦截
2. ✅ **三个选项** - 本次可以/未来均可以/拒绝
3. ✅ **异步通知** - 自动执行后通知，不打扰用户
4. ✅ **一键撤销** - 随时收回授权
5. ✅ **完整审计** - 所有操作可追溯

### 设计原则

| 原则 | 说明 |
|------|------|
| **强制执行** | 不依赖 LLM 记忆，Tool 层强制检查 |
| **灵活授权** | 支持单次、会话、永久三种授权范围 |
| **透明执行** | 自动执行后发送通知，保持透明 |
| **用户控制** | 随时撤销授权，恢复确认流程 |
| **状态持久化** | 超时自动保存 Checkpoint，支持恢复 |

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    用户应用层                                    │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ 停止控制器  │  │ 任务管理器  │  │    确认通知适配器        │ │
│  │ StopCtrl    │  │ TaskManager │  │  - Email 适配器          │ │
│  │ - Stop      │  │ - Suspend   │  │  - Web 适配器            │ │
│  │ - Cancel    │  │ - Resume    │  │  - Slack 适配器          │ │
│  │ - Resume    │  │ - Checkpoint│  │  - 自定义适配器          │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│         │                │                    │                 │
│         └────────────────┼────────────────────┘                 │
│                          │                                      │
└──────────────────────────┼──────────────────────────────────────┘
                           │
                  ┌────────▼────────┐
                  │ ConfirmableTool │ ← 核心包装器
                  │                 │
                  │ - 检查 Stop/Cancel│
                  │ - 匹配确认规则   │
                  │ - 发起确认请求   │
                  │ - 等待确认结果   │
                  └────────┬────────┘
                           │
┌──────────────────────────┼──────────────────────────────────────┘
│                    go-agent 框架层                               │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  agent.Generate()                                       │   │
│  │      ↓                                                  │   │
│  │  ConfirmableTool.Invoke()                               │   │
│  │      ├─ Layer 1: 检查永久授权                           │   │
│  │      ├─ Layer 2: 匹配确认规则                           │   │
│  │      ├─ Layer 3: 发送确认通知                           │   │
│  │      ├─ TaskManager.Suspend()                           │   │
│  │      ├─ 等待确认（超时/取消/批准）                       │   │
│  │      └─ 执行原始 Tool                                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 模块结构

```
go-agent/
├── confirmable/                    # 新增模块
│   ├── types.go                    # 数据类型定义
│   ├── rule.go                     # 确认规则 + 授权管理
│   ├── approval.go                 # ApprovalManager
│   ├── notifier.go                 # 通知适配器接口
│   ├── task_manager.go             # TaskManager
│   ├── confirmable_tool.go         # ConfirmableTool
│   ├── guardrails.go               # Output Guardrails
│   ├── storage.go                  # 存储接口
│   ├── audit.go                    # 审计日志
│   └── notifiers/                  # 通知适配器实现
│       ├── web_notifier.go
│       ├── email_notifier.go
│       ├── slack_notifier.go
│       └── teams_notifier.go
├── examples/
│   └── confirmable_demo/           # 完整示例
│       ├── main.go
│       ├── handlers.go
│       └── frontend/
│           └── confirm.html
└── docs/
    └── CONFIRMABLE_AGENT_IMPLEMENTATION.md  # 本文档
```

---

## 核心组件

### 1. 数据类型 (types.go)

```go
package confirmable

import "time"

// TaskStatus 任务状态
type TaskStatus string

const (
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusSuspended TaskStatus = "suspended"    // 等待确认
    TaskStatusResumed   TaskStatus = "resumed"      // 已恢复
    TaskStatusCancelled TaskStatus = "cancelled"    // 已取消
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
)

// ActionType 用户操作类型
type ActionType string

const (
    ActionTypeApprove  ActionType = "approve"
    ActionTypeDeny     ActionType = "deny"
    ActionTypeCancel   ActionType = "cancel"
    ActionTypeResume   ActionType = "resume"
)

// ApprovalScope 授权范围
type ApprovalScope string

const (
    ApprovalScopeOnce      ApprovalScope = "once"      // 本次
    ApprovalScopeSession   ApprovalScope = "session"   // 当前会话
    ApprovalScopePermanent ApprovalScope = "permanent" // 永久
)

// ConfirmationRequest 确认请求
type ConfirmationRequest struct {
    ID          string                 `json:"id"`
    TaskID      string                 `json:"task_id"`
    SessionID   string                 `json:"session_id"`
    RuleID      string                 `json:"rule_id"`
    ToolName    string                 `json:"tool_name"`
    Arguments   map[string]any         `json:"arguments"`
    Description string                 `json:"description"`
    Context     map[string]string      `json:"context"`
    CreatedAt   time.Time              `json:"created_at"`
    ExpiresAt   time.Time              `json:"expires_at"`
    Status      TaskStatus             `json:"status"`
}

// ConfirmationResponse 确认响应
type ConfirmationResponse struct {
    RequestID   string                 `json:"request_id"`
    Action      ActionType             `json:"action"`
    Message     string                 `json:"message"`
    Modified    map[string]any         `json:"modified,omitempty"`
    Metadata    map[string]string      `json:"metadata,omitempty"` // scope 等
    RespondedAt time.Time              `json:"responded_at"`
}

// UserApproval 用户授权记录
type UserApproval struct {
    ID         string                 `json:"id"`
    UserID     string                 `json:"user_id"`
    RuleID     string                 `json:"rule_id"`
    ToolName   string                 `json:"tool_name"`
    Scope      ApprovalScope          `json:"scope"`
    GrantedAt  time.Time              `json:"granted_at"`
    ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
    Conditions map[string]any         `json:"conditions,omitempty"`
}

// TaskCheckpoint 任务检查点
type TaskCheckpoint struct {
    TaskID         string                 `json:"task_id"`
    SessionID      string                 `json:"session_id"`
    Status         TaskStatus             `json:"status"`
    AgentState     []byte                 `json:"agent_state"`
    PendingRequest *ConfirmationRequest   `json:"pending_request"`
    CreatedAt      time.Time              `json:"created_at"`
    Metadata       map[string]string      `json:"metadata"`
}

// AuditLog 审计日志
type AuditLog struct {
    ID        string                 `json:"id"`
    TaskID    string                 `json:"task_id"`
    SessionID string                 `json:"session_id"`
    EventType string                 `json:"event_type"`
    Timestamp time.Time              `json:"timestamp"`
    Details   map[string]any         `json:"details"`
    UserID    string                 `json:"user_id,omitempty"`
}
```

---

### 2. 确认规则 (rule.go)

```go
package confirmable

import "regexp"

// RuleType 确认规则类型
type RuleType int

const (
    RuleTypeToolName RuleType = iota // 基于工具名
    RuleTypeArgument                 // 基于参数
    RuleTypePermission               // 基于权限边界
    RuleTypePattern                  // 基于提示词模式
)

// ConfirmationRule 确认规则
type ConfirmationRule struct {
    ID          string                 `json:"id"`
    Type        RuleType               `json:"type"`
    ToolName    string                 `json:"tool_name,omitempty"`
    ArgumentKey string                 `json:"argument_key,omitempty"`
    Condition   func(value any) bool   `json:"-"`
    Pattern     *regexp.Regexp         `json:"pattern,omitempty"`
    Permission  string                 `json:"permission,omitempty"`
    Description string                 `json:"description"`
    Enabled     bool                   `json:"enabled"`
    AutoApprove bool                   `json:"auto_approve"`
    DefaultScope ApprovalScope         `json:"default_scope"`
}

// Match 检查是否匹配规则
func (r *ConfirmationRule) Match(toolName string, args map[string]any) bool {
    if !r.Enabled {
        return false
    }

    switch r.Type {
    case RuleTypeToolName:
        return r.ToolName == toolName
    
    case RuleTypeArgument:
        if r.ArgumentKey == "" || r.Condition == nil {
            return false
        }
        value, ok := args[r.ArgumentKey]
        if !ok {
            return false
        }
        return r.Condition(value)
    
    case RuleTypePattern:
        if r.Pattern == nil {
            return false
        }
        text := toolName
        for k, v := range args {
            text += fmt.Sprintf(" %s=%v", k, v)
        }
        return r.Pattern.MatchString(text)
    
    case RuleTypePermission:
        return CheckPermission(r.Permission, toolName, args)
    
    default:
        return false
    }
}

// CheckPermission 权限检查（可自定义实现）
var CheckPermission = func(permission string, toolName string, args map[string]any) bool {
    // 默认实现：简单权限匹配
    // 可以替换为复杂的权限系统
    return false
}
```

---

### 3. 授权管理器 (approval.go)

```go
package confirmable

import (
    "context"
    "sync"
    "time"
)

// ApprovalManager 授权管理器
type ApprovalManager struct {
    mu        sync.RWMutex
    approvals map[string][]*UserApproval // user_id -> approvals
    storage   ApprovalStorage
}

type ApprovalStorage interface {
    Save(ctx context.Context, approval *UserApproval) error
    Load(ctx context.Context, userID string) ([]*UserApproval, error)
    Revoke(ctx context.Context, userID, ruleID string) error
}

// NewApprovalManager 创建授权管理器
func NewApprovalManager(storage ApprovalStorage) *ApprovalManager {
    return &ApprovalManager{
        approvals: make(map[string][]*UserApproval),
        storage:   storage,
    }
}

// Grant 授予权限
func (m *ApprovalManager) Grant(userID, ruleID, toolName string, scope ApprovalScope) (*UserApproval, error) {
    approval := &UserApproval{
        ID:         uuid.New().String(),
        UserID:     userID,
        RuleID:     ruleID,
        ToolName:   toolName,
        Scope:      scope,
        GrantedAt:  time.Now(),
        Conditions: make(map[string]any),
    }

    // 设置过期时间
    switch scope {
    case ApprovalScopeOnce:
        expires := time.Now().Add(24 * time.Hour)
        approval.ExpiresAt = &expires
    case ApprovalScopeSession:
        // 会话级别，不设置过期时间，会话结束时清除
    case ApprovalScopePermanent:
        // 永久授权，ExpiresAt = nil
    }

    // 保存
    if err := m.storage.Save(context.Background(), approval); err != nil {
        return nil, err
    }

    // 加入内存缓存
    m.mu.Lock()
    m.approvals[userID] = append(m.approvals[userID], approval)
    m.mu.Unlock()

    return approval, nil
}

// CheckApproval 检查是否有授权
func (m *ApprovalManager) CheckApproval(userID, ruleID, toolName string) (*UserApproval, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    approvals, exists := m.approvals[userID]
    if !exists {
        return nil, false
    }

    // 查找匹配的授权
    for _, approval := range approvals {
        if approval.IsExpired() {
            continue
        }

        // 检查规则匹配
        if approval.RuleID == ruleID {
            return approval, true
        }

        // 检查工具名匹配（永久授权）
        if approval.Scope == ApprovalScopePermanent && approval.ToolName == toolName {
            return approval, true
        }
    }

    return nil, false
}

// IsExpired 检查授权是否过期
func (a *UserApproval) IsExpired() bool {
    if a.ExpiresAt == nil {
        return false // 永久授权
    }
    return time.Now().After(*a.ExpiresAt)
}

// Revoke 撤销授权
func (m *ApprovalManager) Revoke(ctx context.Context, userID, ruleID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 从内存中移除
    if approvals, exists := m.approvals[userID]; exists {
        filtered := make([]*UserApproval, 0)
        for _, approval := range approvals {
            if approval.RuleID != ruleID {
                filtered = append(filtered, approval)
            }
        }
        m.approvals[userID] = filtered
    }

    // 从存储中移除
    return m.storage.Revoke(ctx, userID, ruleID)
}

// ClearSession 清除会话级授权
func (m *ApprovalManager) ClearSession(userID string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    filtered := make([]*UserApproval, 0)
    for _, approval := range m.approvals[userID] {
        if approval.Scope != ApprovalScopeSession {
            filtered = append(filtered, approval)
        }
    }
    m.approvals[userID] = filtered
}
```

---

## 三层确认机制

### Layer 1: System Prompt

在 System Prompt 中说明确认规则，让 LLM 理解并告知用户：

```go
const confirmationPrompt = `
You are the primary coordinator for an AI agent team.

IMPORTANT CONFIRMATION RULES:
Before executing any of the following actions, you MUST request user confirmation:

1. File Operations:
   - Writing to files (file.write)
   - Deleting files (file.delete)
   - Modifying system files

2. Database Operations:
   - DELETE operations (database.delete)
   - DROP/TRUNCATE operations
   - Schema modifications

3. Financial Operations:
   - Payments over $1000 (payment.process)
   - Refunds over $500

When you need confirmation:
1. Explain what action you want to take
2. Explain why it's needed
3. Wait for explicit user approval before proceeding

Example:
"I need to delete the file /tmp/data.csv. This action is permanent and cannot be undone. Do you approve? (yes/no)"

USER AUTHORIZATION:
Users can authorize you with three options:
- "本次可以" (Approve once) - Only for this time
- "未来均可以" (Approve all) - For all future similar operations
- "拒绝" (Deny) - Reject the operation

If user chooses "未来均可以", you can execute similar operations automatically in the future.
`
```

---

### Layer 2: Output Guardrails

```go
// confirmable/guardrails.go
type ConfirmationGuardrails struct {
    manager  *TaskManager
    rules    []*ConfirmationRule
    patterns []*regexp.Regexp
}

func NewConfirmationGuardrails(manager *TaskManager, rules []*ConfirmationRule) *ConfirmationGuardrails {
    patterns := []string{
        `.*\b(delete|remove|drop|truncate|destroy)\b.*`,
        `.*\b(write|save|update|modify)\b.*file.*`,
        `.*\b(payment|transfer|refund)\b.*\d{4,}.*`,
    }

    compiled := make([]*regexp.Regexp, 0, len(patterns))
    for _, p := range patterns {
        compiled = append(compiled, regexp.MustCompile(p))
    }

    return &ConfirmationGuardrails{
        manager:  manager,
        rules:    rules,
        patterns: compiled,
    }
}

func (g *ConfirmationGuardrails) ValidateAndRepair(ctx context.Context, response string) (string, error) {
    // 检查是否包含敏感操作
    for _, pattern := range g.patterns {
        if pattern.MatchString(response) {
            g.logSensitiveOperation(ctx, response, pattern.String())
            break
        }
    }

    // 检查是否包含用户授权语句
    if grantInfo := g.extractGrantStatement(response); grantInfo != nil {
        g.processGrantStatement(ctx, grantInfo)
    }

    return response, nil
}
```

---

### Layer 3: Tool 拦截（核心）

详见下一节 ConfirmableTool 实现。

---

## 三个确认选项

### 通知动作定义

```go
// NotificationAction 通知动作
type NotificationAction struct {
    ID      string        `json:"id"`
    Label   string        `json:"label"`
    Type    ActionType    `json:"type"`
    Scope   ApprovalScope `json:"scope,omitempty"`
    Style   string        `json:"style,omitempty"`   // "primary" | "secondary" | "danger"
    Icon    string        `json:"icon,omitempty"`
    Confirm *string       `json:"confirm,omitempty"` // 二次确认提示
}

// 创建三个选项
actions := []NotificationAction{
    {
        ID:    "approve_once",
        Label: "本次可以",
        Type:  ActionTypeApprove,
        Scope: ApprovalScopeOnce,
        Style: "secondary",
        Icon:  "✓",
    },
    {
        ID:    "approve_all",
        Label: "未来均可以",
        Type:  ActionTypeApprove,
        Scope: ApprovalScopePermanent,
        Style: "primary",
        Icon:  "✓✓",
    },
    {
        ID:    "deny",
        Label: "拒绝",
        Type:  ActionTypeDeny,
        Style: "danger",
        Icon:  "✗",
    },
}
```

### 处理逻辑

```go
// 在 ConfirmableTool.Invoke() 中
case response := <-taskCtx.ResumeChan:
    switch response.Action {
    case ActionTypeDeny:
        // 用户拒绝
        return agent.ToolResponse{
            Content: fmt.Sprintf("Tool execution denied: %s", response.Message),
        }, nil

    case ActionTypeApprove:
        // 检查授权范围
        scope := response.Metadata["scope"]
        switch ApprovalScope(scope) {
        case ApprovalScopeOnce:
            // 本次可以 - 不保存授权
        case ApprovalScopePermanent:
            // 未来均可以 - 创建永久授权
            for _, rule := range c.rules {
                if rule.Match(toolName, req.Arguments) {
                    _, err := c.approvalMgr.Grant(c.userID, rule.ID, toolName, ApprovalScopePermanent)
                    if err != nil {
                        return agent.ToolResponse{}, fmt.Errorf("grant approval failed: %w", err)
                    }
                    break
                }
            }
        }

        // 用户可能修改了参数
        if response.Modified != nil {
            req.Arguments = response.Modified
        }
    }
```

---

## 异步通知与撤销

### 自动执行后通知

```go
// sendAutoExecutionNotification 发送自动执行通知
func (c *ConfirmableTool) sendAutoExecutionNotification(
    ctx context.Context,
    sessionID, toolName string,
    args map[string]any,
    approval *UserApproval,
    timing string,
    result ...agent.ToolResponse,
) {
    var title, message string
    if timing == "before" {
        title = fmt.Sprintf("自动执行：%s", toolName)
        message = fmt.Sprintf("根据您的预授权，将自动执行 %s 操作", toolName)
    } else {
        title = fmt.Sprintf("已自动执行：%s", toolName)
        message = fmt.Sprintf("根据您的预授权，已自动执行 %s 操作", toolName)
        if len(result) > 0 {
            message += fmt.Sprintf("\n结果：%s", result[0].Content)
        }
    }

    // 撤销授权的二次确认提示
    revokeConfirm := "确定要撤销这个预授权吗？撤销后，下次执行此操作将需要重新确认。"

    notif := &Notification{
        Type:      NotificationTypeAutoExecuted,
        Channel:   c.getNotificationChannel(req),
        Recipient: c.getRecipient(req),
        Title:     title,
        Message:   message,
        Priority:  "low", // 低优先级，不打扰
        Actions: []NotificationAction{
            {
                ID:    "view_details",
                Label: "查看详情",
                Type:  ActionTypeApprove,
                Style: "secondary",
                Icon:  "📄",
            },
            {
                ID:      "revoke_approval",
                Label:   "撤销预授权",
                Type:    ActionTypeCancel,
                Style:   "danger",
                Icon:    "⚠️",
                Confirm: &revokeConfirm,
            },
        },
        Metadata: map[string]string{
            "approval_id":    approval.ID,
            "rule_id":        approval.RuleID,
            "tool_name":      toolName,
            "session_id":     sessionID,
            "execution_time": time.Now().Format(time.RFC3339),
        },
    }

    // 发送通知
    if notifier, ok := c.manager.registry.Get(notif.Channel); ok {
        notifier.Send(ctx, notif)
    }
}
```

### 处理撤销请求

```go
// HandleNotificationResponse 处理通知响应
func (m *TaskManager) HandleNotificationResponse(
    ctx context.Context,
    notificationID string,
    actionID string,
    userID string,
) error {
    meta := m.parseNotificationID(notificationID)

    switch actionID {
    case "revoke_approval":
        // 用户点击"撤销预授权"
        ruleID := meta["rule_id"]
        toolName := meta["tool_name"]

        err := m.approvalMgr.RevokeWithCallback(ctx, userID, ruleID, func(err error) {
            if err != nil {
                return
            }
            // 发送撤销成功通知
            m.approvalMgr.sendRevocationNotification(ctx, userID, ruleID, toolName, m.registry)
        })

        if err != nil {
            return err
        }

        m.auditLogger.Log(ctx, &AuditLog{
            ID:        fmt.Sprintf("revoke_%s", ruleID),
            TaskID:    notificationID,
            EventType: "approval_revoked_by_user",
            Timestamp: time.Now(),
            Details: map[string]any{
                "user_id":   userID,
                "rule_id":   ruleID,
                "tool_name": toolName,
            },
        })
    }

    return nil
}
```

---

## Checkpoint/Restore 集成

### 挂起任务

```go
// Suspend 挂起任务（保存检查点）
func (m *TaskManager) Suspend(ctx context.Context, taskID string, pendingReq *ConfirmationRequest) error {
    m.mu.Lock()
    taskCtx, exists := m.tasks[taskID]
    m.mu.Unlock()

    if !exists {
        return fmt.Errorf("task not found: %s", taskID)
    }

    taskCtx.mu.Lock()
    defer taskCtx.mu.Unlock()

    // 1. 更新状态
    taskCtx.Status = TaskStatusSuspended

    // 2. 创建检查点
    agentState, err := m.agent.Checkpoint()
    if err != nil {
        return fmt.Errorf("checkpoint failed: %w", err)
    }

    checkpoint := &TaskCheckpoint{
        TaskID:         taskID,
        SessionID:      taskCtx.SessionID,
        Status:         TaskStatusSuspended,
        AgentState:     agentState,
        PendingRequest: pendingReq,
        CreatedAt:      time.Now(),
        Metadata: map[string]string{
            "suspend_reason": "waiting_confirmation",
        },
    }

    // 3. 持久化
    if err := m.storage.Save(ctx, checkpoint); err != nil {
        return fmt.Errorf("storage save failed: %w", err)
    }

    taskCtx.Checkpoint = checkpoint

    // 4. 记录审计日志
    m.auditLogger.Log(ctx, &AuditLog{
        ID:        fmt.Sprintf("suspend_%s", taskID),
        TaskID:    taskID,
        SessionID: taskCtx.SessionID,
        EventType: "task_suspended",
        Timestamp: time.Now(),
        Details: map[string]any{
            "reason":    "waiting_confirmation",
            "rule_id":   pendingReq.RuleID,
        },
    })

    return nil
}
```

### 恢复任务

```go
// Resume 恢复任务
func (m *TaskManager) Resume(ctx context.Context, taskID string, response *ConfirmationResponse) error {
    m.mu.Lock()
    taskCtx, exists := m.tasks[taskID]
    m.mu.Unlock()

    if !exists {
        // 尝试从存储中加载
        checkpoint, err := m.storage.Load(ctx, taskID)
        if err != nil {
            return fmt.Errorf("task not found and load failed: %w", err)
        }

        // 重建任务上下文
        taskCtx = &TaskContext{
            TaskID:     taskID,
            SessionID:  checkpoint.SessionID,
            Status:     TaskStatusResumed,
            Checkpoint: checkpoint,
            CancelChan: make(chan struct{}),
            ResumeChan: make(chan *ConfirmationResponse, 1),
        }
        m.mu.Lock()
        m.tasks[taskID] = taskCtx
        m.mu.Unlock()
    }

    taskCtx.mu.Lock()
    defer taskCtx.mu.Unlock()

    // 发送确认响应（非阻塞）
    select {
    case taskCtx.ResumeChan <- response:
    default:
    }

    // 记录审计日志
    m.auditLogger.Log(ctx, &AuditLog{
        ID:        fmt.Sprintf("resume_%s", taskID),
        TaskID:    taskID,
        SessionID: taskCtx.SessionID,
        EventType: "task_resumed",
        Timestamp: time.Now(),
        Details: map[string]any{
            "action": response.Action,
        },
    })

    return nil
}
```

---

## 使用示例

### 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Protocol-Lattice/go-agent"
    "your-module/confirmable"
    "your-module/confirmable/notifiers"
)

func main() {
    ctx := context.Background()

    // 1. 创建原始 Agent
    originalAgent, _ := agent.New(agent.Options{
        SystemPrompt: confirmationPrompt,
        // ...
    })

    // 2. 创建组件
    storage := &DatabaseStorage{db: db}
    auditLogger := &DatabaseAuditLogger{db: db}

    registry := confirmable.NewNotifierRegistry()
    registry.Register(confirmable.ChannelWeb, &notifiers.WebNotifier{hub: wsHub})
    registry.Register(confirmable.ChannelEmail, &notifiers.EmailNotifier{smtpClient: smtp})

    // 3. 创建管理器
    taskManager := confirmable.NewTaskManager(originalAgent, registry, storage, auditLogger)
    approvalMgr := confirmable.NewApprovalManager(storage)

    // 4. 创建可确认工具
    fileWriteTool := confirmable.NewConfirmableTool(
        &FileWriteTool{},
        confirmable.ConfirmableToolOptions{
            Manager:     taskManager,
            ApprovalMgr: approvalMgr,
            UserID:      "user123",
        },
    )

    // 5. 添加规则
    fileWriteTool.AddRule(&confirmable.ConfirmationRule{
        ID:          "file_write_all",
        Type:        confirmable.RuleTypeToolName,
        ToolName:    "file.write",
        Description: "文件写入操作需要您的确认",
    })

    // 6. 创建 Guardrails
    guardrails := confirmable.NewConfirmationGuardrails(taskManager, fileWriteTool.Rules())

    // 7. 创建最终 Agent
    agentWithConfirm, _ := agent.New(agent.Options{
        SystemPrompt: confirmationPrompt,
        Tools:        []agent.Tool{fileWriteTool},
        Guardrails:   guardrails,
    })

    // ========== 第一次运行：需要确认 ==========
    fmt.Println("=== 第一次运行：需要确认 ===")
    go func() {
        resp, err := agentWithConfirm.Generate(ctx, "session1", "帮我写一个文件")
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        } else {
            fmt.Printf("Response: %v\n", resp)
        }
    }()

    // 用户收到确认通知，选择"未来均可以"
    // ... 等待用户响应 ...

    // ========== 第二次运行：自动执行 + 发送异步通知 ==========
    fmt.Println("\n=== 第二次运行：自动执行 + 发送异步通知 ===")
    resp, err := agentWithConfirm.Generate(ctx, "session1", "再写一个文件")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Response: %v\n", resp)
        // 工具自动执行
        // 用户收到异步通知："已自动执行 file.write"
        // 通知包含"撤销预授权"按钮
    }

    // ========== 用户点击"撤销预授权" ==========
    // 用户在前端点击撤销按钮
    // HandleNotificationResponse 被调用
    // 授权被撤销

    // ========== 第三次运行：恢复确认流程 ==========
    fmt.Println("\n=== 第三次运行：恢复确认流程 ===")
    resp, err = agentWithConfirm.Generate(ctx, "session1", "再写一个文件")
    // 需要重新确认
}
```

---

## 通知渠道实现

### Web 通知（WebSocket/SSE）

```go
// confirmable/notifiers/web_notifier.go
type WebNotifier struct {
    hub *WebSocketHub
}

func (n *WebNotifier) Send(ctx context.Context, notif *Notification) (string, error) {
    notificationID := uuid.New().String()

    switch notif.Type {
    case NotificationTypeConfirmation:
        // 确认通知：高优先级，立即推送
        n.hub.Broadcast(notif.Recipient, &WebMessage{
            ID:       notificationID,
            Type:     "confirmation",
            Priority: "high",
            Data:     notif,
        })

    case NotificationTypeAutoExecuted:
        // 自动执行通知：低优先级
        n.hub.Send(notif.Recipient, &WebMessage{
            ID:       notificationID,
            Type:     "auto_executed",
            Priority: "low",
            Data:     notif,
        })
    }

    return notificationID, nil
}
```

### Email 通知

```go
// confirmable/notifiers/email_notifier.go
type EmailNotifier struct {
    smtpClient *smtp.Client
    from       string
}

func (n *EmailNotifier) Send(ctx context.Context, notif *Notification) (string, error) {
    notificationID := uuid.New().String()

    subject := notif.Title
    body := n.buildEmailBody(notif)

    err := n.smtpClient.Send(n.from, []string{notif.Recipient}, []byte(
        fmt.Sprintf("Subject: %s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
            subject, body)))

    return notificationID, err
}

func (n *EmailNotifier) buildEmailBody(notif *Notification) string {
    switch notif.Type {
    case NotificationTypeAutoExecuted:
        return fmt.Sprintf(`
<html>
<body>
  <h2>%s</h2>
  <p>%s</p>
  
  <div style="margin: 20px 0; padding: 15px; background: #f8fff9; border-left: 4px solid #28a745;">
    <h3>操作详情</h3>
    <ul>
      <li>工具：%s</li>
      <li>时间：%s</li>
    </ul>
  </div>
  
  <div style="margin: 20px 0;">
    <a href="%s/revoke?approval_id=%s" 
       style="background: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px;">
      ⚠️ 撤销预授权
    </a>
  </div>
</body>
</html>
`, notif.Title, notif.Message,
           notif.Metadata["tool_name"],
           notif.Metadata["execution_time"],
           os.Getenv("APP_URL"),
           notif.Metadata["approval_id"])
    }
    return ""
}
```

---

## 审计日志

### 审计日志接口

```go
// confirmable/audit.go
type AuditLogger interface {
    Log(ctx context.Context, log *AuditLog) error
    Query(ctx context.Context, taskID string) ([]*AuditLog, error)
}

// 事件类型
const (
    EventTypeToolCall          = "tool_call"
    EventTypeConfirmationSent  = "confirmation_sent"
    EventTypeUserResponded     = "user_responded"
    EventTypeAutoApproved      = "auto_approved"
    EventTypeApprovalGranted   = "approval_granted"
    EventTypeApprovalRevoked   = "approval_revoked"
    EventTypeTaskSuspended     = "task_suspended"
    EventTypeTaskResumed       = "task_resumed"
    EventTypeTaskCancelled     = "task_cancelled"
    EventTypeTaskCompleted     = "task_completed"
)
```

### 用户行为分析

```go
// 查询用户确认模式
func AnalyzeUserBehavior(userID string, startTime, endTime time.Time) (*BehaviorReport, error) {
    logs, err := auditLogger.QueryByUser(userID, startTime, endTime)
    if err != nil {
        return nil, err
    }

    report := &BehaviorReport{
        TotalConfirmations: 0,
        AutoApproveCount:  0,
        DenyCount:         0,
        RevokeCount:       0,
    }

    for _, log := range logs {
        switch log.EventType {
        case EventTypeUserResponded:
            report.TotalConfirmations++
            if log.Details["action"] == "approve_all" {
                report.AutoApproveCount++
            } else if log.Details["action"] == "deny" {
                report.DenyCount++
            }
        case EventTypeApprovalRevoked:
            report.RevokeCount++
        }
    }

    return report, nil
}
```

---

## 测试计划

### 单元测试

```go
// confirmable/confirmable_test.go

// 测试授权管理
func TestApprovalManager_GrantAndCheck(t *testing.T) {
    storage := &MemoryStorage{}
    mgr := NewApprovalManager(storage)

    // 授予永久权限
    approval, err := mgr.Grant("user1", "rule1", "file.write", ApprovalScopePermanent)
    if err != nil {
        t.Fatal(err)
    }

    // 检查权限
    foundApproval, exists := mgr.CheckApproval("user1", "rule1", "file.write")
    if !exists {
        t.Fatal("Expected approval to exist")
    }
    if foundApproval.ID != approval.ID {
        t.Fatal("Approval ID mismatch")
    }
}

// 测试 Tool 包装器
func TestConfirmableTool_Invoke(t *testing.T) {
    // 创建测试组件
    tool := &MockTool{}
    taskManager := NewTaskManager(...)
    approvalMgr := NewApprovalManager(...)

    confirmableTool := NewConfirmableTool(tool, ConfirmableToolOptions{
        Manager:     taskManager,
        ApprovalMgr: approvalMgr,
        UserID:      "user1",
    })

    // 测试无授权时需要确认
    // 测试有授权时自动执行
    // 测试三个选项的处理
}
```

### 集成测试

```go
// 完整流程测试
func TestFullConfirmationFlow(t *testing.T) {
    // 1. 第一次操作：需要确认
    // 2. 用户选择"未来均可以"
    // 3. 第二次操作：自动执行 + 发送通知
    // 4. 用户撤销授权
    // 5. 第三次操作：需要确认
}
```

---

## 总结

本方案提供了完整的 Agent 确认机制，在便利性和控制权之间取得平衡：

### 核心特性

| 特性 | 实现方式 | 优势 |
|------|----------|------|
| **三层确认** | Prompt + Guardrails + Tool 拦截 | 可靠强制执行 |
| **三个选项** | 本次/未来/拒绝 | 灵活授权 |
| **异步通知** | 自动执行后通知 | 透明不打扰 |
| **一键撤销** | 通知中直接撤销 | 保持控制 |
| **Checkpoint** | 超时自动保存 | 支持恢复 |
| **完整审计** | 所有操作记录 | 行为分析 |

### 实施建议

1. **分阶段发布**
   - 阶段 1-3：核心功能（确认 + 授权）
   - 阶段 4-6：增强功能（通知 + 审计）

2. **充分测试**
   - 状态管理正确性
   - 并发安全性
   - 超时处理

3. **文档完善**
   - API 文档
   - 使用示例
   - 最佳实践

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 7 日*
