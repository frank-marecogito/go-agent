# 角色系统扩展设计方案

**项目**: Lattice - Go AI Agent 开发框架  
**创建日期**: 2026 年 3 月 6 日  
**状态**: 设计方案（待实现）  
**优先级**: P2  
**关联任务**: EXT-012  

---

## 📋 概述

扩展 Lattice 现有的角色识别系统，从当前的 4 个基础角色（User、System、Assistant、Conversation memory）扩展到支持更多角色类型，包括工具响应、子代理、情感识别、错误处理等，以提升对话分析能力和系统可观测性。

---

## 🎯 当前状态

### 现有角色（4 个）

```go
var (
    roleUserRe      = regexp.MustCompile(`(?mi)^(?:User|User\s*\(quoted\))\s*:`)
    roleSystemRe    = regexp.MustCompile(`(?mi)^(?:System|System\s*\(quoted\))\s*:`)
    roleAssistantRe = regexp.MustCompile(`(?mi)^(?:Assistant|Assistant\s*\(quoted\))\s*:`)
    roleMemoryRe    = regexp.MustCompile(`(?mi)^Conversation memory`)
)
```

**局限性**：
- ❌ 无法区分工具响应和助手回答
- ❌ 无法追踪子代理的贡献
- ❌ 无法识别用户情感状态
- ❌ 无法分类系统错误
- ❌ 无法记录元数据和调试信息

---

## 🚀 扩展方向

### 方向 1：工具响应角色（Tool Response）

**优先级**: P1  
**预计工作量**: 4 小时  

#### 需求背景
当前工具调用结果与助手回答混在一起，无法：
- 统计工具调用成功率
- 分析工具响应质量
- 优化工具选择策略

#### 设计方案

**新增正则表达式**：
```go
var (
    roleToolRe       = regexp.MustCompile(`(?mi)^(?:Tool|Tool\s*\(response\))\s*:`)
    roleToolSuccessRe = regexp.MustCompile(`(?mi)^(?:Tool\s*\(success\))\s*:`)
    roleToolFailureRe = regexp.MustCompile(`(?mi)^(?:Tool\s*\(failed\))\s*:`)
)
```

**对话示例**：
```
User: 查询北京天气
Assistant: 让我调用天气工具...
Tool: {"temperature": 25, "condition": "sunny", "location": "Beijing"}
Assistant: 北京今天天气晴朗，气温 25 度
```

**使用场景**：
```go
// 统计工具调用
func countToolCalls(messages []Message) (success, failure int) {
    for _, msg := range messages {
        if msg.Role == "tool:success" {
            success++
        } else if msg.Role == "tool:failed" {
            failure++
        }
    }
    return
}

// 分析工具响应质量
func analyzeToolResponseQuality(messages []Message) map[string]float32 {
    quality := make(map[string]float32)
    // 根据后续用户满意度评估工具响应质量
    // ...
    return quality
}
```

#### 交付物
- [ ] `src/agent/roles/tool_role.go` - 工具角色定义
- [ ] `src/agent/roles/tool_role_test.go` - 单元测试
- [ ] 工具调用统计功能
- [ ] 工具响应质量分析示例

---

### 方向 2：子代理角色（SubAgent Roles）

**优先级**: P1  
**预计工作量**: 8 小时  

#### 需求背景
多代理协作时，无法区分不同子代理的贡献，导致：
- 无法评估子代理专业性
- 无法优化子代理选择
- 无法追踪任务分配效果

#### 设计方案

**动态注册子代理角色**：
```go
type SubAgentRoleManager struct {
    mu      sync.RWMutex
    agents  map[string]*regexp.Regexp
    display map[string]string  // name -> display name
}

func (m *SubAgentRoleManager) Register(name, displayName string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    pattern := regexp.MustCompile(`(?mi)^` + displayName + `(?:\s*\(subagent\))?\s*:`)
    m.agents[name] = pattern
    m.display[name] = displayName
}

func (m *SubAgentRoleManager) Identify(text string) (string, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    for name, pattern := range m.agents {
        if pattern.MatchString(text) {
            return name, true
        }
    }
    return "", false
}
```

**对话示例**：
```
User: 帮我分析这个数据集并生成报告
Assistant: 让我协调专家团队...
Researcher: 数据集包含 10 万条记录，主要特征是...
Analyst: 统计结果显示：平均值 50.3，标准差 15.2
Writer: 根据分析结果，报告草稿如下...
Reviewer: 报告结构清晰，数据准确，建议通过
Assistant: 任务完成，请查看最终报告
```

**使用场景**：
```go
// 统计子代理贡献
func countSubAgentContributions(messages []Message) map[string]int {
    contributions := make(map[string]int)
    for _, msg := range messages {
        if strings.HasPrefix(msg.Role, "subagent:") {
            agentName := strings.TrimPrefix(msg.Role, "subagent:")
            contributions[agentName]++
        }
    }
    return contributions
}

// 评估子代理专业性
func evaluateSubAgentExpertise(messages []Message, agentName string) float32 {
    // 根据用户反馈和任务完成度评估
    // ...
    return expertiseScore
}
```

#### 交付物
- [ ] `src/agent/roles/subagent_role.go` - 子代理角色管理器
- [ ] `src/agent/roles/subagent_role_test.go` - 单元测试
- [ ] 子代理贡献统计功能
- [ ] 子代理专业性评估示例
- [ ] 预定义子代理（Researcher、Coder、Writer、Reviewer 等）

---

### 方向 3：情感识别角色（Emotion Roles）

**优先级**: P2  
**预计工作量**: 6 小时  

#### 需求背景
无法识别用户情感状态，导致：
- 无法及时调整回答策略
- 无法评估用户满意度
- 无法提供情感支持

#### 设计方案

**情感标记格式**：
```
[EMOTION: <EMOTION_NAME>]
```

**正则表达式**：
```go
type EmotionRoleManager struct {
    mu       sync.RWMutex
    emotions map[string]*regexp.Regexp
}

func (m *EmotionRoleManager) Register(name, displayName string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    pattern := regexp.MustCompile(`(?mi)^\[EMOTION:\s*` + displayName + `\]`)
    m.emotions[name] = pattern
}

func (m *EmotionRoleManager) Identify(text string) (string, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    for name, pattern := range m.emotions {
        if pattern.MatchString(text) {
            return name, true
        }
    }
    return "", false
}
```

**预定义情感**：
```go
var DefaultEmotions = map[string]string{
    "happy":       "HAPPY",
    "sad":         "SAD",
    "angry":       "ANGRY",
    "frustrated":  "FRUSTRATED",
    "confused":    "CONFUSED",
    "excited":     "EXCITED",
    "satisfied":   "SATISFIED",
    "disappointed": "DISAPPOINTED",
    "surprised":   "SURPRISED",
    "neutral":     "NEUTRAL",
}
```

**对话示例**：
```
User: 这个功能怎么用？
Assistant: 让我解释一下...
User: 我还是不明白
[EMOTION: FRUSTRATED]
Assistant: 抱歉让你困惑了，让我用更简单的方式解释...
User: 哦，现在我明白了！
[EMOTION: HAPPY]
Assistant: 太好了！还有什么可以帮你的吗？
User: 没有了，谢谢
[EMOTION: SATISFIED]
```

**使用场景**：
```go
// 追踪情感变化
func trackEmotionChanges(messages []Message) []EmotionEvent {
    var events []EmotionEvent
    for i, msg := range messages {
        if emotion, ok := msg.Emotion(); ok {
            events = append(events, EmotionEvent{
                Index:     i,
                Emotion:   emotion,
                Timestamp: msg.Timestamp,
            })
        }
    }
    return events
}

// 计算用户满意度
func calculateSatisfactionScore(messages []Message) float32 {
    positive := 0
    negative := 0
    
    for _, msg := range messages {
        switch msg.EmotionName() {
        case "happy", "satisfied", "excited":
            positive++
        case "frustrated", "angry", "disappointed":
            negative++
        }
    }
    
    if positive+negative == 0 {
        return 0.5 // 中性
    }
    
    return float32(positive) / float32(positive+negative)
}

// 根据情感调整回答策略
func adjustResponseStrategy(emotion string) ResponseStrategy {
    switch emotion {
    case "frustrated", "confused":
        return StrategySimpleAndClear
    case "angry":
        return StrategyEmpathetic
    case "happy", "excited":
        return StrategyEnthusiastic
    default:
        return StrategyStandard
    }
}
```

#### 交付物
- [ ] `src/agent/roles/emotion_role.go` - 情感角色管理器
- [ ] `src/agent/roles/emotion_role_test.go` - 单元测试
- [ ] 情感变化追踪功能
- [ ] 用户满意度计算功能
- [ ] 情感感知回答策略

---

### 方向 4：错误处理角色（Error Roles）

**优先级**: P2  
**预计工作量**: 4 小时  

#### 需求背景
系统错误与正常对话混合，导致：
- 无法统计错误类型和频率
- 无法分析错误原因
- 无法改进系统稳定性

#### 设计方案

**错误标记格式**：
```
[ERROR: <ERROR_TYPE>] <ERROR_MESSAGE>
```

**正则表达式**：
```go
var (
    roleErrorRe        = regexp.MustCompile(`(?mi)^\[ERROR:`)
    roleErrorWarningRe = regexp.MustCompile(`(?mi)^\[WARNING:`)
    roleErrorCriticalRe = regexp.MustCompile(`(?mi)^\[CRITICAL:`)
)

type ErrorType string

const (
    ErrorTool      ErrorType = "TOOL_ERROR"
    ErrorModel     ErrorType = "MODEL_ERROR"
    ErrorMemory    ErrorType = "MEMORY_ERROR"
    ErrorNetwork   ErrorType = "NETWORK_ERROR"
    ErrorValidation ErrorType = "VALIDATION_ERROR"
    ErrorSystem    ErrorType = "SYSTEM_ERROR"
)
```

**对话示例**：
```
User: 查询天气
Assistant: 让我调用天气工具...
[ERROR: TOOL_ERROR] 天气服务暂时不可用
Assistant: 抱歉，天气服务暂时不可用，请稍后再试
```

**使用场景**：
```go
// 统计错误类型
func countErrorTypes(messages []Message) map[ErrorType]int {
    counts := make(map[ErrorType]int)
    for _, msg := range messages {
        if msg.IsError() {
            counts[msg.ErrorType()]++
        }
    }
    return counts
}

// 分析错误趋势
func analyzeErrorTrend(messages []Message) ErrorTrend {
    // 分析错误是否随时间增加或减少
    // ...
    return trend
}

// 错误告警
func shouldAlert(errorCounts map[ErrorType]int) bool {
    // 关键错误超过阈值时告警
    if errorCounts[ErrorCritical] > 0 {
        return true
    }
    if errorCounts[ErrorSystem] > 5 {
        return true
    }
    return false
}
```

#### 交付物
- [ ] `src/agent/roles/error_role.go` - 错误角色定义
- [ ] `src/agent/roles/error_role_test.go` - 单元测试
- [ ] 错误统计和分类功能
- [ ] 错误告警机制

---

### 方向 5：元数据角色（Metadata Roles）

**优先级**: P3  
**预计工作量**: 4 小时  

#### 需求背景
系统调试和监控信息无处安放，导致：
- 调试信息污染对话历史
- 无法分离系统日志和对话内容
- 性能分析困难

#### 设计方案

**元数据标记格式**：
```
[META: <KEY>=<VALUE>]
```

**正则表达式**：
```go
var (
    roleMetaRe         = regexp.MustCompile(`(?mi)^\[META:`)
    roleMetaTimingRe   = regexp.MustCompile(`(?mi)^\[META:\s*TIMING`)
    roleMetaDebugRe    = regexp.MustCompile(`(?mi)^\[META:\s*DEBUG`)
    roleMetaPerformanceRe = regexp.MustCompile(`(?mi)^\[META:\s*PERFORMANCE`)
)
```

**对话示例**：
```
User: 帮我写个函数
[META: TIMING] LLM 响应时间：1.2s
Assistant: 好的，函数如下...
[META: PERFORMANCE] Token 使用：150 tokens
[META: DEBUG] 工具调用：code_generator
```

**使用场景**：
```go
// 提取性能指标
func extractPerformanceMetrics(messages []Message) PerformanceMetrics {
    var metrics PerformanceMetrics
    for _, msg := range messages {
        if msg.IsMetadata() && msg.MetadataType() == "TIMING" {
            // 解析时间数据
            // ...
        }
    }
    return metrics
}

// 分离调试信息和对话内容
func separateDebugInfo(messages []Message) (conversation, debug []Message) {
    for _, msg := range messages {
        if msg.IsMetadata() {
            debug = append(debug, msg)
        } else {
            conversation = append(conversation, msg)
        }
    }
    return
}
```

#### 交付物
- [ ] `src/agent/roles/metadata_role.go` - 元数据角色定义
- [ ] `src/agent/roles/metadata_role_test.go` - 单元测试
- [ ] 性能指标提取功能
- [ ] 调试信息分离功能

---

## 🏗️ 架构设计

### 统一角色管理系统

```go
package roles

// RoleType 定义角色类型
type RoleType string

const (
    // 核心角色
    RoleUser      RoleType = "user"
    RoleSystem    RoleType = "system"
    RoleAssistant RoleType = "assistant"
    RoleMemory    RoleType = "memory"
    
    // 扩展角色
    RoleTool      RoleType = "tool"
    RoleSubAgent  RoleType = "subagent"
    RoleEmotion   RoleType = "emotion"
    RoleError     RoleType = "error"
    RoleMetadata  RoleType = "metadata"
)

// RoleManager 统一管理所有角色识别
type RoleManager struct {
    mu            sync.RWMutex
    corePatterns  *CorePatterns
    toolPatterns  *ToolPatterns
    subAgentMgr   *SubAgentRoleManager
    emotionMgr    *EmotionRoleManager
    errorPatterns *ErrorPatterns
    metaPatterns  *MetaPatterns
    
    // 缓存
    cache map[string]RoleType
}

// CorePatterns 核心角色模式
type CorePatterns struct {
    User      *regexp.Regexp
    System    *regexp.Regexp
    Assistant *regexp.Regexp
    Memory    *regexp.Regexp
}

// IdentifyRole 识别文本的角色类型（带缓存）
func (m *RoleManager) IdentifyRole(text string) RoleType {
    m.mu.RLock()
    if role, ok := m.cache[text]; ok {
        m.mu.RUnlock()
        return role
    }
    m.mu.RUnlock()
    
    // 按优先级检查各类角色
    role := m.identifyCoreRole(text)
    if role != "" {
        m.mu.Lock()
        m.cache[text] = role
        m.mu.Unlock()
        return role
    }
    
    role = m.toolPatterns.Identify(text)
    if role != "" {
        m.mu.Lock()
        m.cache[text] = role
        m.mu.Unlock()
        return role
    }
    
    // ... 检查其他角色类型
    
    return "unknown"
}

// NewRoleManager 创建默认角色管理器
func NewRoleManager() *RoleManager {
    return &RoleManager{
        corePatterns:  NewCorePatterns(),
        toolPatterns:  NewToolPatterns(),
        subAgentMgr:   NewSubAgentRoleManager(),
        emotionMgr:    NewEmotionRoleManager(),
        errorPatterns: NewErrorPatterns(),
        metaPatterns:  NewMetaPatterns(),
        cache:         make(map[string]RoleType),
    }
}
```

---

### 与 Agent 集成

```go
type Agent struct {
    // ... 现有字段
    
    roleManager *roles.RoleManager
}

func New(opts Options) (*Agent, error) {
    // ... 现有初始化代码
    
    a := &Agent{
        // ...
        roleManager: roles.NewRoleManager(),
    }
    
    // 注册默认子代理
    if opts.SubAgents != nil {
        for _, sa := range opts.SubAgents {
            a.roleManager.RegisterSubAgent(sa.Name(), sa.Name())
        }
    }
    
    return a, nil
}

// 公开注册方法供用户使用
func (a *Agent) RegisterSubAgent(name, displayName string) {
    a.roleManager.RegisterSubAgent(name, displayName)
}

func (a *Agent) RegisterEmotion(name, displayName string) {
    a.roleManager.RegisterEmotion(name, displayName)
}
```

---

### Message 结构体扩展

```go
type Message struct {
    Role       RoleType
    Content    string
    Timestamp  time.Time
    Metadata   map[string]any
    
    // 扩展字段
    emotion    string
    errorType  ErrorType
    subAgent   string
}

// 访问方法
func (m *Message) Emotion() (string, bool) {
    if m.emotion != "" {
        return m.emotion, true
    }
    return "", false
}

func (m *Message) ErrorType() ErrorType {
    return m.errorType
}

func (m *Message) SubAgentName() (string, bool) {
    if m.subAgent != "" {
        return m.subAgent, true
    }
    return "", false
}
```

---

## 📊 使用示例

### 示例 1：完整的角色识别流程

```go
package main

import (
    "fmt"
    "github.com/Protocol-Lattice/go-agent/src/agent/roles"
)

func main() {
    // 创建角色管理器
    mgr := roles.NewRoleManager()
    
    // 注册自定义子代理
    mgr.RegisterSubAgent("analyst", "DataAnalyst")
    mgr.RegisterSubAgent("visualizer", "Visualizer")
    
    // 注册自定义情感
    mgr.RegisterEmotion("hopeful", "HOPEFUL")
    
    // 测试文本
    texts := []string{
        "User: 你好",
        "Assistant: 有什么可以帮你的？",
        "Tool: {\"result\": \"success\"}",
        "DataAnalyst: 数据分析完成",
        "[EMOTION: HAPPY]",
        "[ERROR: TOOL_ERROR] 服务不可用",
    }
    
    // 识别角色
    for _, text := range texts {
        role := mgr.IdentifyRole(text)
        fmt.Printf("%s -> %s\n", text, role)
    }
}

// 输出：
// User: 你好 -> user
// Assistant: 有什么可以帮你的？ -> assistant
// Tool: {"result": "success"} -> tool
// DataAnalyst: 数据分析完成 -> subagent:analyst
// [EMOTION: HAPPY] -> emotion:happy
// [ERROR: TOOL_ERROR] 服务不可用 -> error:TOOL_ERROR
```

---

### 示例 2：对话分析仪表板

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/Protocol-Lattice/go-agent"
)

func main() {
    // 创建 Agent
    agent, _ := agent.New(agent.Options{
        // ...
    })
    
    // 注册扩展角色
    agent.RegisterSubAgent("researcher", "Researcher")
    agent.RegisterSubAgent("coder", "Coder")
    agent.RegisterEmotion("frustrated", "FRUSTRATED")
    
    // 解析对话历史
    conversation := loadConversation("history.txt")
    messages := agent.ParseConversation(conversation)
    
    // 生成分析报告
    report := ConversationReport{
        TotalMessages:   len(messages),
        RoleDistribution: countRoles(messages),
        SubAgentContributions: countSubAgents(messages),
        EmotionTrend:    analyzeEmotions(messages),
        ErrorSummary:    analyzeErrors(messages),
    }
    
    // 输出 JSON 报告
    json, _ := json.MarshalIndent(report, "", "  ")
    fmt.Println(string(json))
}

type ConversationReport struct {
    TotalMessages        int                    `json:"total_messages"`
    RoleDistribution     map[string]int         `json:"role_distribution"`
    SubAgentContributions map[string]int        `json:"subagent_contributions"`
    EmotionTrend         []EmotionEvent        `json:"emotion_trend"`
    ErrorSummary         ErrorSummary          `json:"error_summary"`
}
```

---

### 示例 3：情感感知对话

```go
func (a *Agent) GenerateWithEmotionAwareness(ctx context.Context, sessionID, userInput string) (string, error) {
    // 获取当前会话的情感状态
    currentEmotion := a.getCurrentEmotion(sessionID)
    
    // 根据情感调整系统提示
    if currentEmotion == "frustrated" {
        a.systemPrompt += "\n注意：用户似乎有些困惑，请用简单清晰的方式解释。"
    } else if currentEmotion == "angry" {
        a.systemPrompt += "\n注意：用户似乎有些生气，请表达理解并提供帮助。"
    }
    
    // 正常生成回答
    response, err := a.Generate(ctx, sessionID, userInput)
    if err != nil {
        return "", err
    }
    
    // 记录情感变化
    a.trackEmotion(sessionID, currentEmotion)
    
    return response, nil
}
```

---

## 🧪 测试策略

### 单元测试

```go
package roles_test

import (
    "testing"
    "github.com/Protocol-Lattice/go-agent/src/agent/roles"
)

func TestRoleManager_IdentifyCoreRoles(t *testing.T) {
    mgr := roles.NewRoleManager()
    
    tests := []struct {
        input    string
        expected roles.RoleType
    }{
        {"User: hello", "user"},
        {"user: hello", "user"}, // 忽略大小写
        {"User (quoted): hello", "user"},
        {"Assistant: hi", "assistant"},
        {"System: be helpful", "system"},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            role := mgr.IdentifyRole(tt.input)
            if role != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, role)
            }
        })
    }
}

func TestRoleManager_IdentifyToolRoles(t *testing.T) {
    mgr := roles.NewRoleManager()
    
    tests := []struct {
        input    string
        expected roles.RoleType
    }{
        {"Tool: {\"result\": 1}", "tool"},
        {"Tool (response): ok", "tool"},
        {"Tool (success): done", "tool:success"},
        {"Tool (failed): error", "tool:failed"},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            role := mgr.IdentifyRole(tt.input)
            if role != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, role)
            }
        })
    }
}

func TestRoleManager_IdentifyEmotions(t *testing.T) {
    mgr := roles.NewRoleManager()
    
    tests := []struct {
        input    string
        expected roles.RoleType
    }{
        {"[EMOTION: HAPPY]", "emotion:happy"},
        {"[EMOTION: FRUSTRATED]", "emotion:frustrated"},
        {"[emotion: happy]", "emotion:happy"}, // 忽略大小写
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            role := mgr.IdentifyRole(tt.input)
            if role != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, role)
            }
        })
    }
}

func TestRoleManager_Caching(t *testing.T) {
    mgr := roles.NewRoleManager()
    
    text := "User: hello"
    
    // 第一次识别
    role1 := mgr.IdentifyRole(text)
    
    // 第二次识别（应该使用缓存）
    role2 := mgr.IdentifyRole(text)
    
    if role1 != role2 {
        t.Error("cached role should match")
    }
}
```

---

### 集成测试

```go
func TestAgent_ConversationAnalysis(t *testing.T) {
    agent, _ := agent.New(agent.Options{
        // ...
    })
    
    // 注册扩展角色
    agent.RegisterSubAgent("researcher", "Researcher")
    agent.RegisterEmotion("happy", "HAPPY")
    
    // 解析对话
    conversation := `
User: 帮我研究一下 Go 语言
Assistant: 让我委托给研究员...
Researcher: Go 是一门静态类型语言
User: 太好了，我明白了！
[EMOTION: HAPPY]
`
    
    messages := agent.ParseConversation(conversation)
    
    // 验证角色识别
    if len(messages) != 5 {
        t.Errorf("expected 5 messages, got %d", len(messages))
    }
    
    // 验证子代理识别
    subAgentCount := 0
    for _, msg := range messages {
        if msg.Role == "subagent:researcher" {
            subAgentCount++
        }
    }
    if subAgentCount != 1 {
        t.Error("should identify 1 subagent message")
    }
    
    // 验证情感识别
    emotionCount := 0
    for _, msg := range messages {
        if strings.HasPrefix(string(msg.Role), "emotion:") {
            emotionCount++
        }
    }
    if emotionCount != 1 {
        t.Error("should identify 1 emotion message")
    }
}
```

---

## 📊 实现路线图

### 阶段 1：核心框架（1 周）
- [ ] 创建 `src/agent/roles/` 目录结构
- [ ] 实现 `RoleManager` 核心类
- [ ] 实现工具响应角色
- [ ] 编写单元测试

### 阶段 2：子代理角色（1 周）
- [ ] 实现 `SubAgentRoleManager`
- [ ] 预定义常用子代理（Researcher、Coder、Writer、Reviewer）
- [ ] 实现子代理贡献统计
- [ ] 集成测试

### 阶段 3：情感识别（1 周）
- [ ] 实现 `EmotionRoleManager`
- [ ] 预定义情感类型
- [ ] 实现情感变化追踪
- [ ] 实现用户满意度计算

### 阶段 4：错误处理（3 天）
- [ ] 实现错误角色定义
- [ ] 实现错误分类和统计
- [ ] 实现错误告警机制

### 阶段 5：元数据角色（3 天）
- [ ] 实现元数据角色定义
- [ ] 实现性能指标提取
- [ ] 实现调试信息分离

### 阶段 6：文档和示例（2 天）
- [ ] 编写 API 文档
- [ ] 编写使用示例
- [ ] 性能基准测试
- [ ] 最佳实践指南

**总预计时间**: 4 周

---

## 🔗 与现有组件集成

### 1. 与 `memory.Engine` 集成

```go
// 在存储记忆时记录角色信息
func (e *Engine) StoreWithRole(ctx context.Context, sessionID, content string, role roles.RoleType) (MemoryRecord, error) {
    metadata := map[string]any{
        "role": string(role),
        "timestamp": time.Now(),
    }
    
    return e.Store(ctx, sessionID, content, metadata)
}
```

### 2. 与 `SharedSession` 集成

```go
// 在共享会话中追踪角色分布
func (s *SharedSession) GetRoleDistribution() map[string]int {
    distribution := make(map[string]int)
    
    for _, space := range s.Spaces() {
        messages := s.GetMessages(space)
        for _, msg := range messages {
            distribution[string(msg.Role)]++
        }
    }
    
    return distribution
}
```

### 3. 与 ADK 集成

```go
// ADK 模块自动注册子代理角色
func (m *SubAgentModule) Provision(ctx context.Context, kit *adk.AgentDevelopmentKit) error {
    for _, sa := range m.subAgents {
        kit.RegisterSubAgent(sa.Name(), sa.Name())
    }
    return nil
}
```

---

## ✅ 实现检查清单

### 核心框架
- [ ] `RoleManager` 实现
- [ ] 角色类型定义
- [ ] 缓存机制
- [ ] 线程安全

### 工具响应角色
- [ ] 正则表达式定义
- [ ] 成功/失败分类
- [ ] 统计功能

### 子代理角色
- [ ] 动态注册机制
- [ ] 预定义子代理
- [ ] 贡献统计

### 情感识别
- [ ] 情感注册
- [ ] 情感变化追踪
- [ ] 满意度计算
- [ ] 策略调整

### 错误处理
- [ ] 错误分类
- [ ] 错误统计
- [ ] 告警机制

### 元数据角色
- [ ] 元数据解析
- [ ] 性能指标提取
- [ ] 调试分离

### 集成测试
- [ ] 单元测试（覆盖率 > 80%）
- [ ] 集成测试
- [ ] 性能测试

### 文档
- [ ] API 文档
- [ ] 使用示例
- [ ] 最佳实践

---

## 📚 参考资料

### 对话系统角色理论
- [Conversation Analysis](https://en.wikipedia.org/wiki/Conversation_analysis)
- [Speech Act Theory](https://en.wikipedia.org/wiki/Speech_act)

### 情感计算
- [Affective Computing](https://en.wikipedia.org/wiki/Affective_computing)
- [Emotion Recognition in Text](https://www.ibm.com/cloud/watson-natural-language-understanding)

### Lattice 现有文档
- [角色正则表达式](../agent.go#L31-L35)
- [记忆系统](../src/memory/)
- [Agent 结构](../agent.go)

---

*本文档为设计方案，具体实现时可能需要调整。*
