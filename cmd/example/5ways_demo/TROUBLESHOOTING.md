# Agent-as-Tool 故障排除

## 问题描述

在 `cmd/example/5ways_demo/main.go` 中，UTCP 工具调用显示错误：

```
⚠️  Tool call failed: invalid CliProvider or missing CommandName
```

## 原因分析

### 1. UTCP 客户端 Transport 配置

`RegisterAsUTCPProvider` 会在 UTCP 客户端中注册一个 `agentCLITransport`，并存储工具映射：

```go
// agent_tool.go:289
shim.tools[tp.Name] = []tools.Tool{tool}
```

### 2. CallTool 调用链

```
client.CallTool(toolName, args)
  ↓
agentCLITransport.CallTool(toolName, args, prov, _)
  ↓
需要正确的 Provider 参数
```

### 3. 问题所在

在 `5ways_demo/main.go` 中：
- UTCP 客户端使用简化配置创建
- `CallTool` 时没有传递正确的 `Provider` 参数
- 导致 `agentCLITransport` 无法找到对应的工具

## 解决方案

### 方案 1: 使用完整配置（推荐）

参考 `cmd/example/agent_as_tool/main.go`：

```go
// 使用 MockModel 或正确配置的 LLM
type MockModel struct {
    Name string
}

func (m *MockModel) Generate(ctx context.Context, prompt string) (any, error) {
    // 模拟 LLM 决策逻辑
    if strings.Contains(prompt, "fact") {
        return "I need to ask the researcher about this.", nil
    }
    return fmt.Sprintf("[%s] I received: %s", m.Name, prompt), nil
}

// 创建 Agent
researcher, _ := agent.New(agent.Options{
    Model:        &MockModel{Name: "Researcher"},
    Memory:       memory.NewSessionMemory(...),
    SystemPrompt: "You are a researcher.",
})

// 注册为 UTCP 工具
researcher.RegisterAsUTCPProvider(ctx, client, "agent.researcher", "...")

// 调用工具
result, _ := client.CallTool(ctx, "agent.researcher", map[string]any{
    "instruction": "Find facts about the sky.",
})
// ✅ 输出：Research complete: The sky is blue because of Rayleigh scattering.
```

### 方案 2: 使用 ADK 的 CallTool

如果使用 ADK（Agent Development Kit），可以直接调用：

```go
kit, _ := adk.New(ctx,
    adk.WithCodeModeUtcp(client, model),
)

// ADK 会正确处理 Provider
result, _ := kit.CallTool(ctx, "agent.researcher", map[string]any{
    "instruction": "Research topic",
})
```

### 方案 3: 修复 5ways_demo

修改 `cmd/example/5ways_demo/main.go` 中的 UTCP 调用：

```go
// 当前代码（可能失败）
result, _ := utcpClient.CallTool(ctx, "expert.researcher", map[string]any{
    "instruction": "What is Go programming language?",
})

// 修复：确保 Provider 正确
// 1. 获取注册的 Provider
transports := utcpClient.GetTransports()
cliTransport := transports[string(base.ProviderCLI)].(*agentCLITransport)

// 2. 调用时传递正确的 Provider
provider := &cli.CliProvider{
    BaseProvider: base.BaseProvider{
        Name:         "expert",
        ProviderType: base.ProviderCLI,
    },
}
result, _ := cliTransport.CallTool(ctx, "expert.researcher", args, provider, nil)
```

## 验证正常工作

运行已配置的示例：

```bash
cd /Users/frank/MareCogito/go-agent

# 这个示例可以正常工作
go run cmd/example/agent_as_tool/main.go

# 输出：
# ✅ Researcher agent registered as UTCP tool: 'agent.researcher'
# --- Direct Tool Call Test ---
# Tool Output: Research complete: The sky is blue because of Rayleigh scattering.
```

## 总结

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `invalid CliProvider` | UTCP Transport 配置不完整 | 使用 `agent_as_tool/main.go` 的完整配置 |
| `tool not found` | Provider 名称不匹配 | 确保 `RegisterAsUTCPProvider` 和 `CallTool` 使用相同的 provider 名称 |
| `missing CommandName` | Provider 参数缺失 | 传递正确的 `cli.CliProvider` 参数 |

## 相关文件

- `agent_tool.go` - Agent-as-Tool 实现
- `agent_tool_test.go` - 单元测试（参考正确用法）
- `cmd/example/agent_as_tool/main.go` - 完整示例
- `cmd/example/5ways_demo/main.go` - 演示（需要修复）
