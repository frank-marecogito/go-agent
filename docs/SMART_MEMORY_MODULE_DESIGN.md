# 智能记忆系统 Module 架构设计

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 2.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: 设计方案  
**相关任务**: EXT-016

---

## 📋 目录

1. [概述](#概述)
2. [Module 架构](#module-架构)
3. [MemCellModule](#memcellmodule)
4. [CausalModule](#causalmodule)
5. [SoftFactorModule](#softfactormodule)
6. [使用示例](#使用示例)
7. [实施计划](#实施计划)

---

## 概述

### 背景

智能记忆系统包含三个子模块：
- **MemCell** - 记忆单元（情景化记忆）
- **Causal** - 因果推理（因果图谱）
- **SoftFactor** - 软性因素（企业文化/哲学/个性）

### 设计原则

| 原则 | 说明 |
|------|------|
| **模块化** | 每个子模块独立实现，可单独启用 |
| **依赖注入** | 通过 ADK Module 自动配置 |
| **向后兼容** | 不破坏现有记忆系统 |
| **可扩展** | 支持新增记忆类型和推理规则 |

---

## Module 架构

### 整体架构

```
ADK Kit
    │
    ├─ WithModule(MemCellModule)      ← 记忆单元模块
    ├─ WithModule(CausalModule)       ← 因果推理模块
    └─ WithModule(SoftFactorModule)   ← 软性因素模块
    
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

### 模块依赖关系

```
MemCellModule（基础）
    │
    ├─ 提供：增强的记忆存储和检索
    ├─ 依赖：基础记忆系统
    │
    ▼
CausalModule（增强）
    │
    ├─ 提供：因果推理工具
    ├─ 依赖：MemCellModule
    │
    ▼
SoftFactorModule（增强）
    │
    ├─ 提供：软性因素调节
    ├─ 依赖：CausalModule
    │
    ▼
完整的智能记忆系统
```

---

## MemCellModule

### 功能

- MemCell（记忆单元）- 类型化记忆（情景/事实/偏好）
- MemScene（记忆场景）- 语义聚类的记忆组
- 情景痕迹形成 - LLM 从对话中提取 MemCell
- 语义整合 - 后台聚类生成 MemScene
- 重构式回忆 - 场景引导的检索

### Module 实现

```go
// src/memory/memcell/memcell_module.go
package memcell

import (
    "context"
    "time"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/memory"
)

// MemCellModule 配置
type MemCellModule struct {
    name string
    opts MemCellOptions
}

type MemCellOptions struct {
    EnableSummaries bool
    ClusterSize     int
    ConsolidateInterval time.Duration  // 语义整合间隔
}

func NewMemCellModule(name string, opts MemCellOptions) *MemCellModule {
    return &MemCellModule{name: name, opts: opts}
}

func (m *MemCellModule) Name() string {
    return m.name
}

// Provision 注册增强的记忆提供者
func (m *MemCellModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
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
        
        // 2. 创建 MemCell 提取器
        extractor := NewEpisodicTraceFormator(bundle.Session)
        
        // 3. 创建语义整合器
        consolidator := NewSemanticConsolidator(
            bundle.Session.Engine,
            extractor,
            m.opts,
        )
        
        // 4. 启动后台任务（定期语义整合）
        go consolidator.RunPeriodically(ctx, m.opts.ConsolidateInterval)
        
        return bundle, nil
    }
    
    kit.UseMemoryProvider(enhancedProvider)
    return nil
}
```

### 核心组件

```go
// 1. MemCell 提取器
type EpisodicTraceFormator struct {
    llm models.Agent
}

func (f *EpisodicTraceFormator) Extract(ctx context.Context, conversation string) ([]*MemCell, error) {
    // 调用 LLM 从对话中提取记忆单元
    prompt := `从对话中提取记忆单元：
    - 关键事实（Fact）
    - 用户偏好（Preference）
    - 时间敏感信息
    
    对话：` + conversation
    
    // 解析 LLM 输出为 MemCell 列表
}

// 2. 语义整合器
type SemanticConsolidator struct {
    engine *memory.Engine
    extractor *EpisodicTraceFormator
    opts MemCellOptions
}

func (c *SemanticConsolidator) RunPeriodically(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.Run(ctx)
        }
    }
}

func (c *SemanticConsolidator) Run(ctx context.Context) error {
    // 1. 获取未聚类的 MemCell
    cells := c.getUnclusteredCells(ctx)
    
    // 2. 向量聚类（KMeans）
    clusters := kmeans.Cluster(cells, c.opts.ClusterSize)
    
    // 3. 为每个聚类生成 MemScene
    for _, cluster := range clusters {
        summary := c.llm.Generate(ctx, "为这些记忆生成场景摘要："+concatenateContents(cluster))
        
        scene := &MemScene{
            Name: extractSceneName(summary),
            Summary: summary,
            MemberIDs: extractIDs(cluster),
        }
        
        // 4. 更新 MemCell 的 scene_id
        for _, cell := range cluster {
            cell.SceneID = scene.ID
        }
    }
    
    return nil
}
```

---

## CausalModule

### 功能

- 因果节点与边 - 事实/事件/规则/结果节点
- 因果挖掘 - LLM 从 MemCell 中提取候选因果关系
- 因果推理引擎 - 原因追溯、影响预测、反事实估计
- DoWhy 集成 - 因果效应估计
- 可解释因果建议

### Module 实现

```go
// src/memory/causal/causal_module.go
package causal

import (
    "context"
    "time"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
    agent "github.com/Protocol-Lattice/go-agent"
)

// CausalModule 配置
type CausalModule struct {
    name         string
    enableMining bool
    dowhyURL     string
    miningInterval time.Duration
}

func NewCausalModule(name string, enableMining bool, dowhyURL string) *CausalModule {
    return &CausalModule{
        name:         name,
        enableMining: enableMining,
        dowhyURL:     dowhyURL,
        miningInterval: 24 * time.Hour,
    }
}

func (m *CausalModule) Name() string {
    return m.name
}

// Provision 注册因果推理功能
func (m *CausalModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    // 1. 注册因果挖掘器（后台任务）
    if m.enableMining {
        miner := NewCausalMiner(kit.ModelProvider(), kit.MemoryProvider())
        
        // 启动后台挖掘任务
        go miner.RunPeriodically(ctx, m.miningInterval)
    }
    
    // 2. 注册因果推理引擎
    reasoner := NewCausalReasoner(kit.MemoryProvider())
    
    // 3. 注册因果推理工具提供者
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

### 核心组件

```go
// 1. 因果挖掘器
type CausalMiner struct {
    llm    models.Agent
    memory kit.MemoryProvider
}

func (m *CausalMiner) RunPeriodically(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.Mine(ctx)
        }
    }
}

func (m *CausalMiner) Mine(ctx context.Context) error {
    // 1. 获取近期新增 MemCell
    cells := m.getRecentCells(ctx)
    
    // 2. 对每个 MemCell 调用 LLM 提取因果对
    for _, cell := range cells {
        prompt := `从以下记忆中提取因果关系：
        格式：原因 → 结果（置信度 0-1）
        
        记忆：` + cell.Content
        
        // 解析 LLM 输出，创建因果边
    }
    
    return nil
}

// 2. 因果推理引擎
type CausalReasoner struct {
    memory kit.MemoryProvider
}

func (r *CausalReasoner) FindCauses(ctx context.Context, nodeID string, depth int) ([]*CausalPath, error) {
    // 图遍历：反向追溯原因
}

func (r *CausalReasoner) FindEffects(ctx context.Context, nodeID string, depth int) ([]*CausalPath, error) {
    // 图遍历：正向追踪影响
}

func (r *CausalReasoner) EstimateEffect(ctx context.Context, treatment, outcome string) (float64, error) {
    // 调用 DoWhy 微服务
    client := NewDoWhyClient(m.dowhyURL)
    return client.EstimateEffect(ctx, treatment, outcome)
}
```

---

## SoftFactorModule

### 功能

- 企业文化节点 - 客户第一、创新驱动等
- 哲学原则节点 - 系统论思维、辩证法等
- 个人特质挖掘 - 拖延指数、风险偏好等
- 软性因素调节 - 影响因果边强度
- 与 SOP 规则协同 - 软约束与硬规则共同作用

### Module 实现

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

// SoftFactorModule 配置
type SoftFactorModule struct {
    name            string
    cultureFile     string
    philosophyFile  string
    enablePersonalityMining bool
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

// Provision 注册软性因素
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
        
        for _, node := range philosophyNodes {
            opts.SystemPrompt += "- " + node.Description + "\n"
        }
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
            ConsolidateInterval: 1 * time.Hour,
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
    // 现在 Agent 可以：
    // 1. 自动提取 MemCell
    // 2. 后台语义整合生成 MemScene
    // 3. 后台因果挖掘
    // 4. 调用因果推理工具
    // 5. 考虑软性因素
    
    resp, err := agent.Generate(ctx, "session1", "帮我分析一下这个现象的原因")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println(resp)
}
```

### 配置文件示例

```json
// configs/culture_nodes.json
{
  "nodes": [
    {
      "id": "culture_customer_first",
      "name": "客户第一",
      "type": "rule",
      "description": "客户需求优先于其他考虑",
      "strength": 0.9,
      "applicable_to": ["all"]
    },
    {
      "id": "culture_innovation",
      "name": "创新驱动",
      "type": "rule",
      "description": "鼓励创新和尝试新方法",
      "strength": 0.7,
      "applicable_to": ["all"]
    }
  ]
}
```

```json
// configs/philosophy_nodes.json
{
  "nodes": [
    {
      "id": "philosophy_systems_thinking",
      "name": "系统论思维",
      "type": "rule",
      "description": "从整体和关联的角度看问题",
      "strength": 0.8,
      "applicable_to": ["all"]
    },
    {
      "id": "philosophy_dialectical",
      "name": "辩证法",
      "type": "rule",
      "description": "一分为二看问题，矛盾统一",
      "strength": 0.6,
      "applicable_to": ["all"]
    }
  ]
}
```

---

## 实施计划

### 阶段 1：MemCellModule（2 周）

- [ ] **M1.1**：MemCell/MemScene 定义及基础存储
- [ ] **M1.2**：情景痕迹形成模块集成 LLM 提取
- [ ] **M1.3**：语义整合后台任务（聚类 + 摘要）
- [ ] **M1.4**：重构式检索与 UTCP 工具封装
- [ ] **M1.5**：端到端测试与验证

### 阶段 2：CausalModule（4 周）

- [ ] **M2.1**：因果节点/边定义及存储
- [ ] **M2.2**：因果挖掘模块集成 LLM 提取
- [ ] **M2.3**：因果推理引擎基础功能
- [ ] **M2.4**：集成 DoWhy 进行效应估计
- [ ] **M2.5**：UTCP 工具封装与 Agent 集成
- [ ] **M2.6**：因果图谱可视化与调试工具

### 阶段 3：SoftFactorModule（3 周）

- [ ] **M3.1**：软性因素节点定义与存储
- [ ] **M3.2**：个人特质挖掘模块
- [ ] **M3.3**：企业文化与哲学原则人工注入工具
- [ ] **M3.4**：因果推理引擎整合软性因素调节
- [ ] **M3.5**：端到端测试与可解释性增强

---

## 总结

### 模块优势

| 优势 | 说明 |
|------|------|
| **模块化** | 每个子模块独立实现，可单独启用 |
| **依赖注入** | 通过 ADK Module 自动配置 |
| **向后兼容** | 不破坏现有记忆系统 |
| **可扩展** | 支持新增记忆类型和推理规则 |

### 下一步

- 实施 MemCellModule
- 实施 CausalModule
- 实施 SoftFactorModule
- 编写测试和文档

---

*文档版本：2.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
