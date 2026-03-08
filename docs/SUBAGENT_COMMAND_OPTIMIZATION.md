# 子代理命令识别优化方案

**项目**: Lattice - Go AI Agent 开发框架  
**创建日期**: 2026 年 3 月 6 日  
**状态**: 设计方案（待实现）  
**优先级**: P2  
**关联任务**: EXT-013 (中文工具调用识别改进) 的子任务  

---

## 📋 问题描述

当前 `handleCommand()` 函数仅支持英文 `subagent:` 前缀，无法识别中文自然语言表达，导致中文用户体验不佳。

### 当前实现

```go
func (a *Agent) handleCommand(ctx context.Context, sessionID, userInput string) (bool, string, error) {
    trimmed := strings.TrimSpace(userInput)
    lower := strings.ToLower(trimmed)

    switch {
    case strings.HasPrefix(lower, "subagent:"):
        // 处理子代理命令...
        return true, result, nil
        
    default:
        return false, "", nil, nil
    }
}
```

### 支持情况

| 输入 | 当前支持 | 期望支持 |
|------|---------|---------|
| `subagent: researcher 搜索 Go` | ✅ | ✅ |
| `调用子代理 researcher 搜索 Go` | ❌ | ✅ |
| `让研究员搜索信息` | ❌ | ✅ |
| `帮我调用 researcher` | ❌ | ✅ |
| `使用 coder 写个函数` | ❌ | ✅ |

---

## 🎯 优化目标

1. **支持中文前缀**：识别常见中文子代理调用表达
2. **支持自然语言**：识别模糊的自然语言表达
3. **保持性能**：不显著增加延迟（目标 < 10ms）
4. **向后兼容**：不影响现有 `subagent:` 格式

---

## 💡 解决方案

### 方向 1：多前缀支持（推荐短期）

**复杂度**: ⭐  
**工作量**: 2 小时  
**性能影响**: 无  

```go
func (a *Agent) handleCommand(ctx context.Context, sessionID, userInput string) (bool, string, map[string]string, error) {
    trimmed := strings.TrimSpace(userInput)
    lower := strings.ToLower(trimmed)

    // 支持多种自然语言前缀
    naturalPrefixes := []string{
        // 英文前缀
        "subagent:",
        "subagent:",
        "agent:",
        
        // 中文前缀
        "调用子代理",
        "调用代理",
        "让子代理",
        "让代理",
        "使用子代理",
        "使用代理",
        "帮我调用",
        "帮我使用",
        "帮我让",
        "帮我运行",
    }
    
    // 检查是否匹配任何前缀
    var matchedPrefix string
    for _, prefix := range naturalPrefixes {
        if strings.HasPrefix(lower, prefix) {
            matchedPrefix = prefix
            break
        }
    }
    
    if matchedPrefix != "" {
        // 提取前缀后的内容
        payload := strings.TrimSpace(trimmed[len(matchedPrefix):])
        if payload == "" {
            return true, "", nil, errors.New("subagent name is missing")
        }
        
        // 解析命令：name + args
        name, args := splitCommand(payload)
        
        // 查找子代理
        sa, ok := a.lookupSubAgent(name)
        if !ok {
            return true, "", nil, fmt.Errorf("unknown subagent: %s", name)
        }
        
        // 执行子代理
        result, err := sa.Run(ctx, args)
        if err != nil {
            return true, "", nil, err
        }
        
        // 记录记忆
        meta := map[string]string{
            "subagent": sa.Name(),
            "session":  sessionID,
            "prefix":   matchedPrefix,
        }
        a.storeMemory(sessionID, "subagent", 
            fmt.Sprintf("%s => %s", sa.Name(), strings.TrimSpace(result)), meta)
        
        return true, result, meta, nil
    }
    
    return false, "", nil, nil
}
```

**支持的新增格式**：
```
✅ "调用子代理 researcher 搜索 Go"
✅ "让研究员搜索信息"
✅ "使用 coder 写个函数"
✅ "帮我调用 researcher 分析数据"
✅ "帮我运行 coder 工具"
```

---

### 方向 2：LLM 意图分类（推荐中期/可选）

**复杂度**: ⭐⭐⭐  
**工作量**: 6 小时  
**性能影响**: 增加 1 次轻量级 LLM 调用（100-500ms）  

```go
// 意图分类器
type IntentClassifier struct {
    model models.Agent
}

type Intent struct {
    Type      string  // "subagent", "tool", "direct"
    AgentName string  // 子代理名称
    ToolName  string  // 工具名称
    Args      string  // 参数
    Confidence float32 // 置信度
}

func (c *IntentClassifier) Classify(userInput string) *Intent {
    prompt := fmt.Sprintf(`
Classify the user's intent into one of these categories:
1. "subagent" - User wants to invoke a sub-agent
2. "tool" - User wants to call a tool
3. "direct" - General conversation

Available subagents: researcher, coder, writer, reviewer
Available tools: echo, calculator, translator

User Input: %s

Respond in JSON:
{
  "type": "subagent|tool|direct",
  "agent_name": "researcher",  // if type is subagent
  "tool_name": "echo",         // if type is tool
  "args": "search Go language",
  "confidence": 0.95
}
`, userInput)

    result, err := c.model.Generate(context.Background(), prompt)
    if err != nil {
        return &Intent{Type: "direct", Confidence: 0.5}
    }

    var intent Intent
    json.Unmarshal([]byte(fmt.Sprint(result)), &intent)
    return &intent
}

// 在 handleCommand 中使用
func (a *Agent) handleCommand(ctx context.Context, sessionID, userInput string) (bool, string, map[string]string, error) {
    trimmed := strings.TrimSpace(userInput)
    lower := strings.ToLower(trimmed)

    // 步骤 1: 快速路径 - 检查标准前缀（< 5ms）
    standardPrefixes := []string{"subagent:", "agent:"}
    for _, prefix := range standardPrefixes {
        if strings.HasPrefix(lower, prefix) {
            return a.executeSubAgentCommand(ctx, sessionID, trimmed[len(prefix):])
        }
    }

    // 步骤 2: 中速路径 - 检查中文前缀（< 5ms）
    chinesePrefixes := []string{
        "调用子代理", "调用代理", "让子代理", "让代理",
        "使用子代理", "使用代理", "帮我调用", "帮我使用",
    }
    for _, prefix := range chinesePrefixes {
        if strings.HasPrefix(lower, prefix) {
            return a.executeSubAgentCommand(ctx, sessionID, trimmed[len(prefix):])
        }
    }

    // 步骤 3: 慢速路径 - LLM 意图判断（100-500ms，可选配置）
    if a.enableIntentClassifier {
        intent := a.intentClassifier.Classify(userInput)
        
        // 只在高置信度时使用
        if intent.Type == "subagent" && intent.Confidence > 0.8 {
            return a.executeSubAgentCommand(ctx, sessionID, intent.Args)
        }
    }

    // 不匹配任何模式
    return false, "", nil, nil
}

// 提取子代理命令执行逻辑到独立函数
func (a *Agent) executeSubAgentCommand(ctx context.Context, sessionID, payload string) (bool, string, map[string]string, error) {
    payload = strings.TrimSpace(payload)
    if payload == "" {
        return true, "", nil, errors.New("subagent name is missing")
    }

    name, args := splitCommand(payload)
    sa, ok := a.lookupSubAgent(name)
    if !ok {
        return true, "", nil, fmt.Errorf("unknown subagent: %s", name)
    }

    result, err := sa.Run(ctx, args)
    if err != nil {
        return true, "", nil, err
    }

    meta := map[string]string{
        "subagent": sa.Name(),
        "session":  sessionID,
    }
    a.storeMemory(sessionID, "subagent", 
        fmt.Sprintf("%s => %s", sa.Name(), strings.TrimSpace(result)), meta)

    return true, result, meta, nil
}
```

**支持的新增格式**：
```
✅ "我觉得应该让研究员分析一下"
✅ "这种情况用 coder 比较合适"
✅ "找个懂代码的帮我看看"
✅ "有没有擅长写作的代理？"
```

---

### 方向 3：混合模式（推荐长期）

**复杂度**: ⭐⭐  
**工作量**: 4 小时  
**性能影响**: 可配置  

```go
type SubAgentCommandHandler struct {
    // 标准前缀（最快）
    standardPrefixes []string
    
    // 中文前缀（快）
    chinesePrefixes []string
    
    // 正则模式（中等）
    patterns []*regexp.Regexp
    
    // LLM 分类器（慢，可选）
    intentClassifier *IntentClassifier
    
    // 配置
    enableLLM bool
    llmThreshold float32
}

func NewSubAgentCommandHandler() *SubAgentCommandHandler {
    return &SubAgentCommandHandler{
        standardPrefixes: []string{"subagent:", "agent:"},
        chinesePrefixes: []string{
            "调用子代理", "调用代理", "让子代理", "让代理",
            "使用子代理", "使用代理", "帮我调用", "帮我使用",
        },
        patterns: []*regexp.Regexp{
            // "用 researcher 工具"
            regexp.MustCompile(`(?i)用\s*(\w+)\s*(工具 | 代理 | 助手)?`),
            // "让研究员帮忙"
            regexp.MustCompile(`(?i)让\s*(\w+)\s*(帮忙 | 处理 | 分析)?`),
            // "找个懂代码的"
            regexp.MustCompile(`(?i)找个\s*(\w+)\s*的`),
        },
        intentClassifier: NewIntentClassifier(),
        enableLLM: true,
        llmThreshold: 0.8,
    }
}

func (h *SubAgentCommandHandler) Handle(ctx context.Context, sessionID, userInput string) (bool, string, map[string]string, error) {
    lower := strings.ToLower(strings.TrimSpace(userInput))

    // 层级 1: 标准前缀（< 1ms）
    for _, prefix := range h.standardPrefixes {
        if strings.HasPrefix(lower, prefix) {
            return true, "standard", nil
        }
    }

    // 层级 2: 中文前缀（< 5ms）
    for _, prefix := range h.chinesePrefixes {
        if strings.HasPrefix(lower, prefix) {
            return true, "chinese", nil
        }
    }

    // 层级 3: 正则模式（< 10ms）
    for _, pattern := range h.patterns {
        if pattern.MatchString(userInput) {
            return true, "regex", nil
        }
    }

    // 层级 4: LLM 分类（100-500ms，可选）
    if h.enableLLM {
        intent := h.intentClassifier.Classify(userInput)
        if intent.Type == "subagent" && intent.Confidence > h.llmThreshold {
            return true, "llm", nil
        }
    }

    return false, "", nil, nil
}
```

**性能分层**：
```
层级 1 (标准前缀): < 1ms, 覆盖率 30%
层级 2 (中文前缀): < 5ms, 覆盖率 50%
层级 3 (正则模式): < 10ms, 覆盖率 70%
层级 4 (LLM 分类): 100-500ms, 覆盖率 90%
```

---

## 📊 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **方向 1** | 简单快速，无性能影响 | 需要维护前缀列表 | 短期快速改进 |
| **方向 2** | 智能，支持自然语言 | 慢，有成本 | 高端场景/可选 |
| **方向 3** | 分层处理，性能/智能平衡 | 复杂度较高 | 长期稳定方案 |

**推荐策略**：
1. **短期**：实施方向 1（2 小时，立即见效）
2. **中期**：实施方向 3 的前 3 层（4 小时）
3. **长期**：完整实施方向 3（含 LLM 层）

---

## 📝 实现需求

### 阶段 1：多前缀支持（2 小时）
- [ ] 修改 `handleCommand()` 函数
- [ ] 添加中文前缀列表常量
- [ ] 提取子代理执行逻辑到独立函数
- [ ] 编写单元测试（中文场景）
- [ ] 性能基准测试

### 阶段 2：正则模式支持（4 小时）
- [ ] 定义预编译正则模式
- [ ] 支持常见自然语言表达
- [ ] 集成到处理流程
- [ ] 集成测试

### 阶段 3：LLM 意图分类（可选，6 小时）
- [ ] 实现 `IntentClassifier` 类
- [ ] 添加配置开关
- [ ] 添加置信度阈值配置
- [ ] 错误降级处理
- [ ] 性能优化（缓存分类结果）

### 阶段 4：文档和示例（2 小时）
- [ ] 更新使用文档
- [ ] 添加中文示例
- [ ] 编写最佳实践

---

## 🧪 测试用例

### 单元测试

```go
func TestHandleCommand_ChinesePrefixes(t *testing.T) {
    agent := createTestAgentWithSubAgents()
    
    tests := []struct {
        input    string
        expected bool
        desc     string
    }{
        // === 现有英文支持（向后兼容）===
        {"subagent: researcher 搜索 Go", true, "英文 subagent: 前缀"},
        {"agent: coder 写函数", true, "英文 agent: 前缀"},
        
        // === 新增中文前缀支持 ===
        {"调用子代理 researcher 搜索 Go", true, "调用子代理"},
        {"调用代理 researcher 搜索 Go", true, "调用代理"},
        {"让子代理 researcher 搜索 Go", true, "让子代理"},
        {"让代理 researcher 搜索 Go", true, "让代理"},
        {"使用子代理 researcher 搜索 Go", true, "使用子代理"},
        {"使用代理 researcher 搜索 Go", true, "使用代理"},
        {"帮我调用 researcher 搜索 Go", true, "帮我调用"},
        {"帮我使用 researcher 搜索 Go", true, "帮我使用"},
        {"帮我运行 researcher 搜索 Go", true, "帮我运行"},
        
        // === 中文参数支持 ===
        {"调用子代理 researcher 分析一下这个数据", true, "中文参数"},
        {"帮我调用 coder 写个排序函数", true, "中文参数"},
        
        // === 不应该识别的情况 ===
        {"researcher 搜索 Go", false, "缺少前缀"},
        {"今天天气不错", false, "普通对话"},
        {"", false, "空字符串"},
    }
    
    for _, tt := range tests {
        t.Run(tt.desc, func(t *testing.T) {
            handled, _, _, _ := agent.handleCommand(context.Background(), "session-1", tt.input)
            if handled != tt.expected {
                t.Errorf("输入 %q: 期望 %v, 得到 %v", tt.input, tt.expected, handled)
            }
        })
    }
}
```

### 性能基准测试

```go
func BenchmarkHandleCommand_EnglishPrefix(b *testing.B) {
    agent := createTestAgentWithSubAgents()
    input := "subagent: researcher 搜索 Go"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.handleCommand(context.Background(), "session-1", input)
    }
}

func BenchmarkHandleCommand_ChinesePrefix(b *testing.B) {
    agent := createTestAgentWithSubAgents()
    input := "调用子代理 researcher 搜索 Go"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.handleCommand(context.Background(), "session-1", input)
    }
}

func BenchmarkHandleCommand_LLMClassifier(b *testing.B) {
    agent := createTestAgentWithSubAgents()
    agent.enableIntentClassifier = true
    input := "我觉得应该让研究员分析一下"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.handleCommand(context.Background(), "session-1", input)
    }
}
```

---

## 🔗 与现有组件集成

### 1. 与角色系统扩展集成

```go
// 在角色系统中添加子代理命令角色
var (
    roleSubAgentCommandRe = regexp.MustCompile(`(?mi)^(?:调用 | 让 | 使用 | 帮我)\s*(?:子代理 | 代理 | 研究员 | 程序员)`)
)

func (a *Agent) handleCommand(ctx context.Context, sessionID, userInput string) (bool, string, error) {
    // 使用角色系统的识别能力
    if roleSubAgentCommandRe.MatchString(userInput) {
        // 处理子代理命令...
    }
    // ... 原有逻辑
}
```

### 2. 配置化支持

```go
type Agent struct {
    // ... 现有字段
    
    // 子代理命令处理配置
    subAgentPrefixes []string
    enableIntentClassifier bool
    intentClassifier *IntentClassifier
}

func New(opts Options) (*Agent, error) {
    a := &Agent{
        // ...
        subAgentPrefixes: getDefaultSubAgentPrefixes(),
        enableIntentClassifier: false,  // 默认关闭
    }
    return a, nil
}

// 用户可自定义前缀
func (a *Agent) SetSubAgentPrefixes(prefixes []string) {
    a.subAgentPrefixes = prefixes
}
```

---

## 📊 预期效果

### 改进前 vs 改进后

| 指标 | 改进前 | 改进后（方向 1） | 改进后（方向 3） |
|------|-------|----------------|----------------|
| 中文前缀识别 | 0% | 90% | 95% |
| 自然语言识别 | 0% | 20% | 85% |
| 平均延迟 | < 5ms | < 10ms | 50-200ms |
| 覆盖率 | 30% | 80% | 95% |

### 用户体验提升

**改进前**：
```
用户：调用子代理 researcher 搜索 Go
Agent: （未识别，交给 LLM 处理，2000ms）
```

**改进后（方向 1）**：
```
用户：调用子代理 researcher 搜索 Go
Agent: （< 5ms 识别，直接执行子代理）
```

**改进后（方向 3）**：
```
用户：我觉得应该让研究员分析一下
Agent: （LLM 分类器识别，100ms，执行子代理）
```

---

## ✅ 实现检查清单

### 代码实现
- [ ] 修改 `handleCommand()` 函数
- [ ] 添加中文前缀列表常量
- [ ] 提取 `executeSubAgentCommand()` 函数
- [ ] 实现正则模式（方向 2）
- [ ] 实现 LLM 意图分类器（方向 3，可选）
- [ ] 添加配置选项

### 测试
- [ ] 单元测试（中文场景覆盖）
- [ ] 性能基准测试
- [ ] 集成测试
- [ ] 向后兼容性测试

### 文档
- [ ] 更新 API 文档
- [ ] 添加中文使用示例
- [ ] 编写最佳实践指南

---

## 📚 参考资料

### 相关代码
- [`agent.go` handleCommand()](../agent.go#L600-L650)
- [子代理目录](../catalog.go#L81-L140)

### 相关任务
- [EXT-013 中文工具调用识别改进](./TODO_EXTENSION.md)
- [中文工具调用支持设计方案](./CHINESE_TOOL_CALL_SUPPORT.md)

---

*本文档为设计方案，具体实现时可能需要调整。*
