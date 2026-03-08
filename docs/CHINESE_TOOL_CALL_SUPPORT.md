# 中文工具调用识别改进方案

**项目**: Lattice - Go AI Agent 开发框架  
**创建日期**: 2026 年 3 月 6 日  
**状态**: 设计方案（待实现）  
**优先级**: P2  
**关联任务**: EXT-012 (角色系统扩展) 的子任务  

---

## 📋 问题描述

当前 `userLooksLikeToolCall()` 函数对中文工具调用识别支持有限，无法识别中文前缀和自然语言表达。

### 当前实现

```go
func (a *Agent) userLooksLikeToolCall(s string) bool {
    s = strings.TrimSpace(strings.ToLower(s))

    // 1. 检查 { } 格式
    if strings.Contains(s, "{") && strings.Contains(s, "}") {
        parts := strings.Fields(s)
        tool := parts[0]
        // 匹配工具名...
    }

    // 2. 检查 tool: 前缀（仅英文）
    if strings.HasPrefix(s, "tool:") {
        return true
    }

    // 3. 检查 JSON 格式
    if strings.HasPrefix(s, "{") && strings.Contains(s, "\"tool\"") {
        return true
    }

    return false
}
```

### 支持情况

| 输入 | 当前支持 | 期望支持 |
|------|---------|---------|
| `echo {"input": "你好"}` | ✅ | ✅ |
| `tool: echo {...}` | ✅ | ✅ |
| `工具：echo {...}` | ❌ | ✅ |
| `调用：echo {...}` | ❌ | ✅ |
| `帮我运行 echo 工具` | ❌ | ✅ |
| `执行 echo 命令` | ❌ | ✅ |

---

## 🎯 改进目标

1. **支持中文工具前缀**：`工具：`、`调用：`、`运行：`、`执行：`
2. **支持中英文混合**：`echo 工具 {...}`
3. **支持自然语言表达**：`帮我运行计算器`
4. **保持性能**：不显著增加延迟
5. **向后兼容**：不影响现有英文调用

---

## 💡 解决方案

### 方案 1：添加中文前缀支持（推荐短期）

**复杂度**: ⭐  
**工作量**: 2 小时  
**性能影响**: 无  

```go
func (a *Agent) userLooksLikeToolCall(s string) bool {
    s = strings.TrimSpace(strings.ToLower(s))

    // 原有逻辑：tool: 前缀
    if strings.HasPrefix(s, "tool:") {
        return true
    }
    
    // 新增：中文前缀支持
    chinesePrefixes := []string{
        "工具：",
        "工具:",      // 中文冒号
        "调用：",
        "调用:",
        "运行：",
        "运行:",
        "执行：",
        "执行:",
        "使用：",
        "使用:",
    }
    
    for _, prefix := range chinesePrefixes {
        if strings.HasPrefix(s, prefix) {
            return true
        }
    }

    // 原有逻辑：{...} 格式
    if strings.Contains(s, "{") && strings.Contains(s, "}") {
        parts := strings.Fields(s)
        if len(parts) > 0 {
            tool := parts[0]
            for _, t := range a.ToolSpecs() {
                if strings.ToLower(t.Name) == tool {
                    return true
                }
            }
        }
    }

    // 原有逻辑：JSON 格式
    if strings.HasPrefix(s, "{") && strings.Contains(s, "\"tool\"") {
        return true
    }

    return false
}
```

**支持的新增格式**：
```
✅ "工具：echo {...}"
✅ "调用：echo {...}"
✅ "运行：echo {...}"
✅ "执行：echo {...}"
✅ "使用：echo {...}"
```

---

### 方案 2：正则表达式识别（推荐中期）

**复杂度**: ⭐⭐  
**工作量**: 4 小时  
**性能影响**: 轻微  

```go
var toolCallPatterns = []*regexp.Regexp{
    // 英文 tool: 前缀
    regexp.MustCompile(`(?i)^\s*tool\s*:\s*(\w+)\s*\{`),
    
    // 中文前缀
    regexp.MustCompile(`(?i)^\s*(工具 | 调用 | 运行 | 执行 | 使用)\s*[:：]\s*(\w+)\s*\{`),
    
    // 工具名 + { 直接跟随
    regexp.MustCompile(`(?i)^\s*(\w+)\s*\{`),
    
    // JSON 格式
    regexp.MustCompile(`^\s*\{\s*["']tool["']\s*:`),
}

func (a *Agent) userLooksLikeToolCall(s string) bool {
    for _, pattern := range toolCallPatterns {
        matches := pattern.FindStringSubmatch(s)
        if len(matches) > 0 {
            // 提取工具名并验证
            toolName := matches[len(matches)-1]
            for _, t := range a.ToolSpecs() {
                if strings.EqualFold(t.Name, toolName) {
                    return true
                }
            }
        }
    }
    return false
}
```

**支持的新增格式**：
```
✅ "tool: echo {\"input\": \"hello\"}"
✅ "工具：echo {\"input\": \"你好\"}"
✅ "调用：echo {\"input\": \"你好\"}"
✅ "echo {\"input\": \"你好\"}"
✅ "{\"tool\": \"echo\", \"input\": \"你好\"}"
✅ "工具：calculator {\"expr\": \"1+1\"}"
```

---

### 方案 3：LLM 智能判断（推荐长期/可选）

**复杂度**: ⭐⭐⭐  
**工作量**: 6 小时  
**性能影响**: 增加 1 次 LLM 调用  

```go
func (a *Agent) userLooksLikeToolCall(s string) bool {
    // 快速路径：先检查简单模式
    if strings.Contains(s, "tool:") || 
       strings.HasPrefix(s, "{") ||
       hasChineseToolPrefix(s) {
        return true
    }
    
    // 复杂情况：用 LLM 判断（可配置阈值）
    prompt := fmt.Sprintf(`
判断以下用户输入是否意图调用工具。

用户输入:
%s

工具调用示例:
- "echo {\"input\": \"hello\"}"
- "帮我运行计算器"
- "调用天气工具查询北京"
- "用翻译工具翻译这句话"

如果用户想要调用工具，回答 true，否则回答 false。
`, s)
    
    result, err := a.model.Generate(context.Background(), prompt)
    if err != nil {
        return false  // 错误时返回 false，降级处理
    }
    
    return strings.Contains(strings.ToLower(fmt.Sprint(result)), "true")
}

// 辅助函数：检查中文前缀
func hasChineseToolPrefix(s string) bool {
    prefixes := []string{"工具：", "调用：", "运行：", "执行：", "使用："}
    for _, prefix := range prefixes {
        if strings.HasPrefix(s, prefix) {
            return true
        }
    }
    return false
}
```

**支持的新增格式**：
```
✅ "帮我运行计算器"
✅ "调用天气工具查询北京"
✅ "我想用 echo 工具说句话"
✅ "用翻译工具翻译这句话"
✅ "echo {\"input\": \"你好\"}"
```

---

## 📊 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **方案 1** | 简单快速，无性能影响 | 需要维护前缀列表 | 短期快速改进 |
| **方案 2** | 灵活，支持多种格式 | 正则复杂度高 | 中期稳定方案 |
| **方案 3** | 最智能，支持自然语言 | 慢，有成本 | 高端场景/可选 |

**推荐策略**：
1. **短期**：实施方案 1（2 小时，立即见效）
2. **中期**：实施方案 2（4 小时，更稳定）
3. **长期**：方案 2 + 方案 3 结合（智能降级）

---

## 📝 实现需求

### 阶段 1：中文前缀支持（2 小时）
- [ ] 修改 `userLooksLikeToolCall()` 函数
- [ ] 添加中文前缀列表
- [ ] 编写单元测试（中文场景）
- [ ] 性能基准测试

### 阶段 2：正则表达式优化（4 小时）
- [ ] 定义预编译正则模式
- [ ] 重构识别逻辑
- [ ] 支持更多中文变体
- [ ] 集成测试

### 阶段 3：LLM 智能判断（可选，6 小时）
- [ ] 实现 LLM 判断逻辑
- [ ] 添加配置开关
- [ ] 添加调用频率限制
- [ ] 错误降级处理

### 阶段 4：文档和示例（2 小时）
- [ ] 更新使用文档
- [ ] 添加中文示例
- [ ] 编写最佳实践

---

## 🧪 测试用例

### 单元测试

```go
func TestUserLooksLikeToolCall_ChineseSupport(t *testing.T) {
    agent := createTestAgentWithTools()
    
    tests := []struct {
        input    string
        expected bool
        desc     string
    }{
        // === 现有英文支持（向后兼容）===
        {"echo {\"input\": \"hello\"}", true, "英文工具调用"},
        {"tool: echo {...}", true, "tool: 前缀"},
        {"{\"tool\": \"echo\"}", true, "JSON 格式"},
        {"TOOL: ECHO {...}", true, "忽略大小写"},
        
        // === 新增中文前缀支持 ===
        {"工具：echo {...}", true, "中文工具前缀"},
        {"工具：echo {...}", true, "中文冒号"},
        {"调用：echo {...}", true, "调用前缀"},
        {"运行：echo {...}", true, "运行前缀"},
        {"执行：echo {...}", true, "执行前缀"},
        {"使用：echo {...}", true, "使用前缀"},
        
        // === 中文参数支持（已支持）===
        {"echo {\"input\": \"你好\"}", true, "中文参数"},
        {"tool: echo {\"msg\": \"你好世界\"}", true, "中文参数 +tool 前缀"},
        {"工具：echo {\"msg\": \"你好\"}", true, "中文参数 + 中文前缀"},
        
        // === 自然语言（方案 3 支持）===
        {"帮我运行计算器", false, "自然语言 - 待方案 3 支持"},
        {"调用天气工具", false, "自然语言 - 待方案 3 支持"},
        
        // === 不应该识别的情况 ===
        {"今天天气不错", false, "普通对话"},
        {"echo 这个词怎么用", false, "提及工具名但不是调用"},
        {"", false, "空字符串"},
    }
    
    for _, tt := range tests {
        t.Run(tt.desc, func(t *testing.T) {
            result := agent.userLooksLikeToolCall(tt.input)
            if result != tt.expected {
                t.Errorf("输入 %q: 期望 %v, 得到 %v", tt.input, tt.expected, result)
            }
        })
    }
}
```

### 性能基准测试

```go
func BenchmarkUserLooksLikeToolCall_English(b *testing.B) {
    agent := createTestAgentWithTools()
    input := "tool: echo {\"input\": \"hello\"}"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.userLooksLikeToolCall(input)
    }
}

func BenchmarkUserLooksLikeToolCall_Chinese(b *testing.B) {
    agent := createTestAgentWithTools()
    input := "工具：echo {\"input\": \"你好\"}"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.userLooksLikeToolCall(input)
    }
}

func BenchmarkUserLooksLikeToolCall_LLM(b *testing.B) {
    agent := createTestAgentWithTools()
    input := "帮我运行计算器"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.userLooksLikeToolCall(input)
    }
}
```

---

## 🔗 与现有组件集成

### 1. 与角色系统扩展集成

```go
// 在角色系统中添加工具调用角色
var (
    roleToolCallRe = regexp.MustCompile(`(?mi)^(?:工具 | 调用 | 运行 | 执行)\s*[:：]`)
)

func (a *Agent) userLooksLikeToolCall(s string) bool {
    // 使用角色系统的识别能力
    if roleToolCallRe.MatchString(s) {
        return true
    }
    // ... 原有逻辑
}
```

### 2. 配置化支持

```go
type Agent struct {
    // ... 现有字段
    
    // 中文工具调用支持配置
    chineseToolPrefixes []string
    enableLLMJudgment   bool
}

func New(opts Options) (*Agent, error) {
    a := &Agent{
        // ...
        chineseToolPrefixes: getDefaultChinesePrefixes(),
        enableLLMJudgment:   false,  // 默认关闭
    }
    return a, nil
}

// 用户可自定义前缀
func (a *Agent) SetChineseToolPrefixes(prefixes []string) {
    a.chineseToolPrefixes = prefixes
}
```

---

## 📊 预期效果

### 改进前 vs 改进后

| 指标 | 改进前 | 改进后（方案 1+2） | 改进后（方案 1+2+3） |
|------|-------|------------------|-------------------|
| 中文前缀识别 | 0% | 95% | 98% |
| 自然语言识别 | 0% | 20% | 90% |
| 平均延迟 | 0ms | 0ms | +50-200ms |
| 代码复杂度 | 低 | 中 | 高 |

### 用户体验提升

**改进前**：
```
用户：工具：echo {"input": "你好"}
Agent: （未识别，当作普通对话）

用户：帮我运行计算器
Agent: （未识别，当作普通对话）
```

**改进后（方案 1+2）**：
```
用户：工具：echo {"input": "你好"}
Agent: （识别为工具调用）→ 执行 echo 工具

用户：调用：calculator {"expr": "1+1"}
Agent: （识别为工具调用）→ 执行计算器工具
```

**改进后（方案 1+2+3）**：
```
用户：帮我运行计算器
Agent: （LLM 判断为工具调用）→ 执行计算器工具

用户：用翻译工具翻译这句话
Agent: （LLM 判断为工具调用）→ 执行翻译工具
```

---

## ✅ 实现检查清单

### 代码实现
- [ ] 修改 `userLooksLikeToolCall()` 函数
- [ ] 添加中文前缀列表常量
- [ ] 实现正则表达式模式（方案 2）
- [ ] 实现 LLM 判断逻辑（方案 3，可选）
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
- [`agent.go` userLooksLikeToolCall()](../agent.go#L157-L183)
- [角色正则表达式](../agent.go#L31-L35)

### 相关任务
- [EXT-012 角色系统扩展](./TODO_EXTENSION.md)
- [角色系统扩展设计方案](./ROLE_SYSTEM_EXTENSION.md)

---

*本文档为设计方案，具体实现时可能需要调整。*
