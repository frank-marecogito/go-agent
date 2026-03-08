# 基于 Lattice 的智能记忆与因果推理系统

**项目**: MareMind - 智能 Agent 记忆系统  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 7 日  
**状态**: 设计方案  
**相关任务**: EXT-016

---

## 📋 目录

1. [引言与总体目标](#引言与总体目标)
2. [总体架构](#总体架构)
3. [阶段一：EverMemOS 风格记忆系统](#阶段一 evermemos-风格记忆系统)
4. [阶段二：REMI 因果记忆框架](#阶段二 remi-因果记忆框架)
5. [阶段三：软性因素建模](#阶段三软性因素建模)
6. [技术选型与集成要点](#技术选型与集成要点)
7. [与 go-agent 现有能力对比](#与 go-agent 现有能力对比)
8. [实施路线图](#实施路线图)
9. [风险评估与缓解](#风险评估与缓解)
10. [附录](#附录)

---

## 引言与总体目标

### 背景

在 AI Agent 实际应用中，记忆系统和推理能力是决定 Agent 智能水平的核心因素。当前 go-agent 框架提供了基础的记忆存储和检索功能，但缺乏：
- **情景化记忆** - 无法像人类一样按场景组织记忆
- **因果推理** - 无法回答"为什么"和"如果...会怎样"
- **软性因素** - 无法融入企业文化、哲学原则等软性约束

### 目标

构建集**自组织记忆、因果推理、软性因素建模**于一体的智能 Agent 记忆系统：
- **记住** - 情景化记忆，场景引导回忆
- **理解** - 因果推理，解释现象，预测未来
- **共情** - 融入企业文化与个人特质，提供人性化建议

### 设计原则

| 原则 | 说明 |
|------|------|
| **渐进式** | 分三阶段实施，每阶段交付可验证成果 |
| **兼容性** | 基于 go-agent 现有记忆系统扩展 |
| **可解释** | 因果推理过程透明，支持审计 |
| **可配置** | 软性因素可人工定义和调整 |
| **高性能** | 支持大规模记忆和图谱查询 |

---

## 总体架构

### 三层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层 (Agent/UTCP)                       │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ 记忆检索    │  │ 因果推理    │  │ 软性因素调节        │ │
│  │ search_     │  │ causal_     │  │ culture_adjust_     │ │
│  │ memory      │  │ find_causes │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    因果层 (Causal Layer)                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 因果图谱 (Causal Graph)                             │   │
│  │  - CausalNode (事实/事件/规则/结果)                 │   │
│  │  - CausalEdge (causes/leads_to/inhibits)            │   │
│  │  - 软性因素节点 (文化/哲学/个性)                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ CausalMiner │  │ Causal      │  │ DoWhy               │ │
│  │ (因果挖掘)  │  │ Reasoner    │  │ Service             │ │
│  │             │  │ (推理引擎)  │  │ (效应估计)          │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    记忆层 (Memory Layer)                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ MemCell (记忆单元)                                   │   │
│  │  - Episodic (情景记忆)                               │   │
│  │  - Fact (事实记忆)                                   │   │
│  │  - Preference (偏好记忆)                             │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ MemScene (记忆场景)                                  │   │
│  │  - 语义聚类的记忆组                                  │   │
│  │  - 场景名称和摘要                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Episodic    │  │ Semantic    │  │ Reconstructive      │ │
│  │ Trace       │  │ Consolida-  │  │ Retriever           │ │
│  │ Formator    │  │ tor         │  │ (重构检索)          │ │
│  │ (情景提取)  │  │ (语义整合)  │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
用户对话
    │
    ▼
EpisodicTraceFormator (LLM 提取)
    │
    ▼
MemCell (类型化记忆)
    │
    ├─→ SemanticConsolidator (后台聚类)
    │       │
    │       ▼
    │   MemScene (场景)
    │
    ├─→ CausalMiner (因果挖掘)
    │       │
    │       ▼
    │   CausalGraph (因果图谱)
    │
    └─→ ReconstructiveRetriever (场景引导检索)
            │
            ▼
        Agent 上下文
```

---

## 阶段一：EverMemOS 风格记忆系统

### 目标

- 实现 MemCell（记忆单元）和 MemScene（记忆场景）的数据结构
- 完成情景痕迹形成：将对话流转化为 MemCell 并存储
- 实现语义整合：后台聚类 MemCell 生成 MemScene，并更新用户画像
- 实现重构式回忆：场景引导的检索，提供必要且充分的上下文

### 技术实现

#### 1.1 MemCell 与 MemScene 数据结构

```go
// src/memory/memcell/types.go
package memcell

import (
    "time"
    "github.com/Protocol-Lattice/go-agent/src/memory/model"
)

// MemCellType 记忆单元类型
type MemCellType string

const (
    TypeEpisodic   MemCellType = "episodic"    // 情景记忆
    TypeFact       MemCellType = "fact"        // 事实记忆
    TypePreference MemCellType = "preference"  // 偏好记忆
)

// MemCell 记忆单元（封装 MemoryRecord）
type MemCell struct {
    Record     model.MemoryRecord
    Type       MemCellType
    SceneID    string      // 所属场景 ID
    Confidence float64     // 置信度 0-1
}

// MemCellMetadata 元数据结构
type MemCellMetadata struct {
    Type        string    `json:"type"`
    SceneID     string    `json:"scene_id,omitempty"`
    Importance  float64   `json:"importance"`
    Confidence  float64   `json:"confidence"`
    AccessCount int       `json:"access_count"`
    LastUsed    time.Time `json:"last_used"`
}

// MemScene 记忆场景
type MemScene struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Summary     string    `json:"summary"`
    Embedding   []float32 `json:"embedding"`
    MemberIDs   []string  `json:"member_ids"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Importance  float64   `json:"importance"`
}
```

**与 go-agent 集成**：
- MemCell 作为 `Source="memcell"` 的 MemoryRecord 存储
- MemScene 作为 `Source="memscene"` 的 MemoryRecord 存储
- 元数据嵌入到 `MemoryRecord.Metadata`（JSON 格式）

#### 1.2 情景痕迹形成模块

```go
// src/memory/memcell/extractor.go
package memcell

type EpisodicTraceFormator struct {
    llm models.Agent
}

func NewEpisodicTraceFormator(llm models.Agent) *EpisodicTraceFormator {
    return &EpisodicTraceFormator{llm: llm}
}

// Extract 从对话中提取 MemCell
func (f *EpisodicTraceFormator) Extract(ctx context.Context, conversation string) ([]*MemCell, error) {
    prompt := `从以下对话中提取记忆单元：

要求：
1. 识别关键事实（Fact）
2. 识别用户偏好（Preference）
3. 识别时间敏感信息
4. 为每个记忆单元分配类型

对话：
` + conversation

    response, err := f.llm.Generate(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 解析 LLM 输出为 MemCell 列表
    cells, err := parseMemCells(response)
    if err != nil {
        return nil, err
    }

    // 生成向量嵌入
    for _, cell := range cells {
        embedding, _ := embed(ctx, cell.Record.Content)
        cell.Record.Embedding = embedding
    }

    return cells, nil
}
```

**LLM 提示词示例**：
```
从以下对话中提取记忆单元：

对话：
用户：我下周要去北京出差，帮我记得带上充电器
助手：好的，我已经记下了。您还需要准备什么吗？

提取结果：
[
  {
    "type": "fact",
    "content": "用户下周要去北京出差",
    "confidence": 0.95
  },
  {
    "type": "preference",
    "content": "用户需要带充电器",
    "confidence": 0.9
  }
]
```

#### 1.3 语义整合模块（后台任务）

```go
// src/memory/memcell/consolidator.go
package memcell

type SemanticConsolidator struct {
    engine *memory.Engine
    llm    models.Agent
}

func NewSemanticConsolidator(engine *memory.Engine, llm models.Agent) *SemanticConsolidator {
    return &SemanticConsolidator{engine: engine, llm: llm}
}

// Run 执行语义整合（每日运行）
func (c *SemanticConsolidator) Run(ctx context.Context) error {
    // 1. 获取未聚类的 MemCell
    cells, err := c.getUnclusteredCells(ctx)
    if err != nil {
        return err
    }

    if len(cells) == 0 {
        return nil
    }

    // 2. 向量聚类（KMeans）
    clusters := kmeans.Cluster(cells, k=5)

    // 3. 为每个聚类生成场景
    for _, cluster := range clusters {
        // 调用 LLM 生成场景名称和摘要
        summary, err := c.llm.Generate(ctx, 
            "为这些记忆生成场景名称和摘要："+concatenateContents(cluster))
        if err != nil {
            continue
        }

        // 4. 创建或更新 MemScene
        scene := &MemScene{
            Name:      extractSceneName(summary),
            Summary:   summary,
            MemberIDs: extractIDs(cluster),
        }

        // 5. 更新 MemCell 的 scene_id
        for _, cell := range cluster {
            cell.SceneID = scene.ID
            c.engine.Store(ctx, cell)
        }
    }

    return nil
}
```

**调度配置**：
```go
// 每日凌晨 2 点运行
cron.AddFunc("0 2 * * *", func() {
    consolidator.Run(context.Background())
})
```

#### 1.4 重构式回忆检索

```go
// src/memory/memcell/retriever.go
package memcell

type ReconstructiveRetriever struct {
    engine *memory.Engine
}

func NewReconstructiveRetriever(engine *memory.Engine) *ReconstructiveRetriever {
    return &ReconstructiveRetriever{engine: engine}
}

// Retrieve 场景引导的检索
func (r *ReconstructiveRetriever) Retrieve(ctx context.Context, query string, limit int) ([]*MemCell, error) {
    // 1. 检索最相关 MemScene
    scenes, err := r.engine.RetrieveByType(ctx, "memscene", query, 3)
    if err != nil {
        return nil, err
    }

    // 2. 在场景内检索 MemCell
    var cells []*MemCell
    for _, scene := range scenes {
        sceneCells, err := r.engine.RetrieveByScene(ctx, scene.ID, query, limit)
        if err != nil {
            continue
        }
        cells = append(cells, sceneCells...)
    }

    // 3. 如果不足，全局检索
    if len(cells) < limit {
        globalCells, err := r.engine.RetrieveByType(ctx, "memcell", query, limit-len(cells))
        if err == nil {
            cells = append(cells, globalCells...)
        }
    }

    // 4. 可选：LLM 过滤"必要且充分"
    cells = r.filterNecessaryAndSufficient(ctx, cells, query, limit)

    return cells, nil
}

// filterNecessaryAndSufficient 使用 LLM 过滤冗余记忆
func (r *ReconstructiveRetriever) filterNecessaryAndSufficient(
    ctx context.Context, cells []*MemCell, query string, limit int,
) []*MemCell {
    if len(cells) <= limit {
        return cells
    }

    // 调用 LLM 选择最相关的记忆
    prompt := `从以下记忆中选择最必要且充分的 ` + fmt.Sprintf("%d", limit) + ` 条来回答用户问题：

问题：` + query + `

记忆列表：
` + concatenateCells(cells)

    // 解析 LLM 输出，返回选中的记忆
    selectedIDs, _ := parseSelectedIDs(r.llm.Generate(ctx, prompt))
    return filterByID(cells, selectedIDs)
}
```

#### 1.5 UTCP 工具封装

```go
// src/memory/memcell/utcp_tools.go
package memcell

import (
    "github.com/universal-tool-calling-protocol/go-utcp"
)

// RegisterUTCPTools 注册记忆检索工具
func RegisterUTCPTools(client utcp.UtcpClientInterface, retriever *ReconstructiveRetriever) error {
    // search_memory 工具
    tool := tools.Tool{
        Name:        "search_memory",
        Description: "从记忆系统中检索相关记忆",
        Inputs: tools.ToolInputOutputSchema{
            Type: "object",
            Properties: map[string]any{
                "query": map[string]any{
                    "type":        "string",
                    "description": "检索查询",
                },
                "limit": map[string]any{
                    "type":        "integer",
                    "description": "返回数量限制",
                    "default":     5,
                },
            },
            Required: []string{"query"},
        },
        Handler: func(ctx context.Context, inputs map[string]any) (any, error) {
            query, _ := inputs["query"].(string)
            limit, _ := inputs["limit"].(int)
            if limit <= 0 {
                limit = 5
            }

            cells, err := retriever.Retrieve(ctx, query, limit)
            if err != nil {
                return nil, err
            }

            // 返回记忆内容列表
            var results []string
            for _, cell := range cells {
                results = append(results, cell.Record.Content)
            }
            return results, nil
        },
    }

    _, err := client.RegisterToolProvider(ctx, tool)
    return err
}
```

### 里程碑

| 里程碑 | 内容 | 周数 | 交付物 |
|--------|------|------|--------|
| **M1.1** | MemCell/MemScene 定义及基础存储 | 1 | `types.go`, 存储集成 |
| **M1.2** | 情景痕迹形成模块集成 LLM 提取 | 2 | `extractor.go`, LLM 提示词 |
| **M1.3** | 语义整合后台任务（聚类 + 摘要） | 2 | `consolidator.go`, KMeans 实现 |
| **M1.4** | 重构式检索与 UTCP 工具封装 | 1 | `retriever.go`, UTCP 工具 |
| **M1.5** | 端到端测试与验证 | 1 | 测试用例，验证报告 |

### 验证指标

- **记忆检索准确率** > 85%（对比人工标注）
- **场景聚类主题一致性** > 80%（人工评估）
- **检索响应时间** < 500ms（P95）

---

## 阶段二：REMI 因果记忆框架

### 目标

- 在记忆系统基础上构建因果图谱，支持因果节点和边
- 实现因果挖掘：从 MemCell 中提取候选因果关系
- 实现因果推理引擎：支持原因追溯、影响预测、反事实估计
- 与 Agent 集成，提供可解释的因果建议

### 技术实现

#### 2.1 因果节点与边数据结构

```go
// src/memory/causal/types.go
package causal

// NodeType 因果节点类型
type NodeType string

const (
    TypeFact    NodeType = "fact"     // 事实
    TypeEvent   NodeType = "event"    // 事件
    TypeRule    NodeType = "rule"     // 规则
    TypeOutcome NodeType = "outcome"  // 结果
)

// CausalNode 因果节点
type CausalNode struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        NodeType               `json:"type"`
    Description string                 `json:"description"`
    Embedding   []float32              `json:"embedding"`
    Metadata    map[string]interface{} `json:"metadata"`
    MemCellIDs  []string               `json:"mem_cell_ids"` // 关联的 MemCell
}

// EdgeType 因果边类型
type EdgeType string

const (
    TypeCauses   EdgeType = "causes"    // 导致
    TypeLeadsTo  EdgeType = "leads_to"  // 引发
    TypeInhibits EdgeType = "inhibits"  // 抑制
)

// CausalEdge 因果边
type CausalEdge struct {
    ID         string                 `json:"id"`
    From       string                 `json:"from_node_id"`
    To         string                 `json:"to_node_id"`
    Type       EdgeType               `json:"type"`
    Strength   float64                `json:"strength"`      // 强度 0-1
    Confidence float64                `json:"confidence"`    // 置信度 0-1
    Evidence   []string               `json:"evidence"`      // MemCell IDs
    Metadata   map[string]interface{} `json:"metadata"`
}
```

**存储方式**：
- CausalNode 作为 `Source="causal_node"` 的 MemoryRecord
- CausalEdge 作为 `Source="causal_edge"` 的 MemoryRecord
- 元数据中包含 `from_node_id`、`to_node_id` 便于过滤

#### 2.2 因果挖掘模块

```go
// src/memory/causal/miner.go
package causal

type CausalMiner struct {
    llm    models.Agent
    engine *memory.Engine
}

func NewCausalMiner(llm models.Agent, engine *memory.Engine) *CausalMiner {
    return &CausalMiner{llm: llm, engine: engine}
}

// Mine 从 MemCell 中提取因果关系
func (m *CausalMiner) Mine(ctx context.Context, cells []*memcell.MemCell) ([]*CausalEdge, error) {
    var edges []*CausalEdge

    for _, cell := range cells {
        // 调用 LLM 提取因果对
        prompt := `从以下记忆中提取因果关系：

格式：原因 → 结果（置信度 0-1）
示例：
- "用户熬夜" → "第二天工作效率低" (0.8)
- "经常运动" → "睡眠质量好" (0.7)

记忆：
` + cell.Record.Content

        response, err := m.llm.Generate(ctx, prompt)
        if err != nil {
            continue
        }

        // 解析 LLM 输出
        pairs, err := parseCausalPairs(response)
        if err != nil {
            continue
        }

        // 节点归一化（通过向量相似度匹配现有节点或新建）
        for _, pair := range pairs {
            fromNode, err := m.normalizeNode(ctx, pair.Cause)
            if err != nil {
                continue
            }

            toNode, err := m.normalizeNode(ctx, pair.Effect)
            if err != nil {
                continue
            }

            // 创建或更新因果边
            edge, err := m.createOrUpdateEdge(ctx, fromNode, toNode, pair.Confidence, cell.ID)
            if err != nil {
                continue
            }

            edges = append(edges, edge)
        }
    }

    return edges, nil
}

// normalizeNode 节点归一化
func (m *CausalMiner) normalizeNode(ctx context.Context, name string) (*CausalNode, error) {
    // 1. 向量相似度匹配现有节点
    embedding, _ := embed(ctx, name)
    existingNodes, _ := m.engine.RetrieveSimilarNodes(ctx, embedding, 5)

    // 2. 如果相似度 > 0.9，返回现有节点
    for _, node := range existingNodes {
        if similarity(node.Embedding, embedding) > 0.9 {
            return &node, nil
        }
    }

    // 3. 否则创建新节点
    newNode := &CausalNode{
        ID:          uuid.New().String(),
        Name:        name,
        Type:        TypeFact,
        Description: name,
        Embedding:   embedding,
    }

    m.engine.StoreNode(ctx, newNode)
    return newNode, nil
}
```

#### 2.3 因果推理引擎

```go
// src/memory/causal/reasoner.go
package causal

type CausalReasoner struct {
    engine *memory.Engine
}

func NewCausalReasoner(engine *memory.Engine) *CausalReasoner {
    return &CausalReasoner{engine: engine}
}

// FindCauses 反向追溯原因
func (r *CausalReasoner) FindCauses(ctx context.Context, nodeID string, depth int) ([]*CausalPath, error) {
    var paths []*CausalPath

    // 图遍历：反向追踪
    visited := make(map[string]bool)
    var traverse func(nodeID string, currentPath []*CausalEdge, currentDepth int)
    
    traverse = func(nodeID string, currentPath []*CausalEdge, currentDepth int) {
        if currentDepth > depth {
            return
        }

        // 检索指向该节点的边
        edges, _ := r.engine.RetrieveEdgesTo(ctx, nodeID)
        
        for _, edge := range edges {
            if visited[edge.From] {
                continue
            }
            visited[edge.From] = true

            newPath := append(currentPath, edge)
            paths = append(paths, &CausalPath{Edges: newPath})

            // 递归
            traverse(edge.From, newPath, currentDepth+1)
        }
    }

    traverse(nodeID, []*CausalEdge{}, 0)
    return paths, nil
}

// FindEffects 正向追踪影响
func (r *CausalReasoner) FindEffects(ctx context.Context, nodeID string, depth int) ([]*CausalPath, error) {
    // 类似 FindCauses，方向相反
}

// EstimateEffect 因果效应估计（集成 DoWhy）
func (r *CausalReasoner) EstimateEffect(
    ctx context.Context,
    treatment string,
    outcome string,
    context map[string]any,
) (float64, error) {
    // 1. 提取因果子图
    subgraph, err := r.GetSubgraph(ctx, []string{treatment, outcome}, 3)
    if err != nil {
        return 0, err
    }

    // 2. 调用 DoWhy（Python 微服务）
    client := NewDoWhyClient("http://localhost:8000")
    effect, err := client.EstimateEffect(ctx, subgraph, treatment, outcome, context)
    if err != nil {
        return 0, err
    }

    return effect, nil
}

// GetSubgraph 提取因果子图
func (r *CausalReasoner) GetSubgraph(ctx context.Context, seedIDs []string, depth int) (*CausalGraph, error) {
    // 从种子节点开始，BFS 遍历获取子图
}
```

#### 2.4 DoWhy 集成

```python
# services/dowhy/main.py
from fastapi import FastAPI
import dowhy.causalmodel as cm

app = FastAPI()

@app.post("/estimate_effect")
async def estimate_effect(request: EffectRequest):
    """
    估计因果效应
    """
    # 1. 构建因果图
    causal_graph = build_graph_from_request(request)
    
    # 2. 创建因果模型
    model = cm.CausalModel(
        data=request.data,
        treatment=request.treatment,
        outcome=request.outcome,
        graph=causal_graph
    )
    
    # 3. 识别效应
    identified_estimand = model.identify_effect()
    
    # 4. 估计效应
    estimate = model.estimate_effect(
        identified_estimand,
        method_name="backdoor.propensity_score_matching"
    )
    
    return {"effect": estimate.value, "confidence": estimate.confidence_interval}
```

**Go 客户端**：
```go
// src/memory/causal/dowhy_client.go
package causal

type DoWhyClient struct {
    endpoint string
}

func (c *DoWhyClient) EstimateEffect(
    ctx context.Context,
    graph *CausalGraph,
    treatment, outcome string,
    context map[string]any,
) (float64, error) {
    req := EffectRequest{
        Graph:     graph,
        Treatment: treatment,
        Outcome:   outcome,
        Context:   context,
    }

    resp, err := httpPost(ctx, c.endpoint+"/estimate_effect", req)
    if err != nil {
        return 0, err
    }

    return resp.Effect, nil
}
```

#### 2.5 UTCP 工具封装

```go
// src/memory/causal/utcp_tools.go
package causal

func RegisterUTCPTools(client utcp.UtcpClientInterface, reasoner *CausalReasoner) error {
    // causal_find_causes 工具
    findCausesTool := tools.Tool{
        Name:        "causal_find_causes",
        Description: "反向追溯事件的原因",
        Inputs: tools.ToolInputOutputSchema{
            Type: "object",
            Properties: map[string]any{
                "node_id": map[string]any{"type": "string"},
                "depth":   map[string]any{"type": "integer", "default": 3},
            },
            Required: []string{"node_id"},
        },
        Handler: func(ctx context.Context, inputs map[string]any) (any, error) {
            nodeID, _ := inputs["node_id"].(string)
            depth, _ := inputs["depth"].(int)
            paths, err := reasoner.FindCauses(ctx, nodeID, depth)
            return formatPaths(paths), err
        },
    }

    // causal_estimate_effect 工具
    estimateEffectTool := tools.Tool{
        Name:        "causal_estimate_effect",
        Description: "估计干预的因果效应",
        Inputs: tools.ToolInputOutputSchema{
            Type: "object",
            Properties: map[string]any{
                "treatment": map[string]any{"type": "string"},
                "outcome":   map[string]any{"type": "string"},
            },
            Required: []string{"treatment", "outcome"},
        },
        Handler: func(ctx context.Context, inputs map[string]any) (any, error) {
            treatment, _ := inputs["treatment"].(string)
            outcome, _ := inputs["outcome"].(string)
            effect, err := reasoner.EstimateEffect(ctx, treatment, outcome, nil)
            return map[string]float64{"effect": effect}, err
        },
    }

    // 注册工具
    _, err := client.RegisterToolProvider(ctx, findCausesTool)
    if err != nil {
        return err
    }

    _, err = client.RegisterToolProvider(ctx, estimateEffectTool)
    return err
}
```

### 里程碑

| 里程碑 | 内容 | 周数 | 交付物 |
|--------|------|------|--------|
| **M2.1** | 因果节点/边定义及存储 | 1 | `types.go`, 存储集成 |
| **M2.2** | 因果挖掘模块集成 LLM 提取 | 2 | `miner.go`, LLM 提示词 |
| **M2.3** | 因果推理引擎基础功能 | 2 | `reasoner.go`, 图遍历算法 |
| **M2.4** | 集成 DoWhy 进行效应估计 | 2 | `dowhy_client.go`, Python 服务 |
| **M2.5** | UTCP 工具封装与 Agent 集成 | 1 | UTCP 工具 |
| **M2.6** | 因果图谱可视化与调试工具 | 2 | Web 界面，调试工具 |

### 验证指标

- **因果挖掘准确率** > 75%（对比专家标注）
- **推理结果合理性** > 80%（专家评估）
- **效应估计误差** < 20%（对比真实数据）

---

## 阶段三：软性因素建模

### 目标

- 将企业文化、哲学原则、个人特质作为因果节点纳入系统
- 实现软性因素的识别与动态更新
- 在因果推理中考虑软性因素对因果强度的影响
- 提供融合软性因素的可解释输出

### 技术实现

#### 3.1 软性因素节点定义

```go
// src/memory/causal/soft_factor.go
package causal

// FactorType 软性因素类型
type FactorType string

const (
    TypeCulture    FactorType = "culture"    // 企业文化
    TypePhilosophy FactorType = "philosophy" // 哲学原则
    TypePersonality FactorType = "personality" // 个人特质
)

// SoftFactor 软性因素节点
type SoftFactor struct {
    CausalNode
    FactorType   FactorType `json:"factor_type"`
    Strength     float64    `json:"strength"`      // 影响强度 0-1
    ApplicableTo []string   `json:"applicable_to"` // 适用的用户/场景
}

// 预定义的企业文化节点
var CultureNodes = map[string]*SoftFactor{
    "customer_first": {
        CausalNode: CausalNode{
            ID:          "culture_customer_first",
            Name:        "客户第一",
            Type:        TypeRule,
            Description: "客户需求优先于其他考虑",
        },
        FactorType:   TypeCulture,
        Strength:     0.9,
        ApplicableTo: []string{"all"},
    },
    "innovation": {
        CausalNode: CausalNode{
            ID:          "culture_innovation",
            Name:        "创新驱动",
            Type:        TypeRule,
            Description: "鼓励创新和尝试新方法",
        },
        FactorType:   TypeCulture,
        Strength:     0.7,
        ApplicableTo: []string{"all"},
    },
}

// 预定义的哲学原则
var PhilosophyNodes = map[string]*SoftFactor{
    "systems_thinking": {
        CausalNode: CausalNode{
            ID:          "philosophy_systems_thinking",
            Name:        "系统论思维",
            Type:        TypeRule,
            Description: "从整体和关联的角度看问题",
        },
        FactorType:   TypePhilosophy,
        Strength:     0.8,
        ApplicableTo: []string{"all"},
    },
}
```

#### 3.2 个人特质挖掘

```go
// src/memory/causal/personality.go
package causal

type PersonalityMiner struct {
    llm models.Agent
}

type PersonalityProfile struct {
    ProcrastinationIndex float64 `json:"procrastination_index"` // 拖延指数 0-1
    RiskPreference       float64 `json:"risk_preference"`       // 风险偏好 0-1
    DecisionStyle        string  `json:"decision_style"`        // "rational" or "emotional"
    // ... 更多特质
}

func NewPersonalityMiner(llm models.Agent) *PersonalityMiner {
    return &PersonalityMiner{llm: llm}
}

// Infer 从 MemCell 推断用户特质
func (m *PersonalityMiner) Infer(ctx context.Context, cells []*memcell.MemCell) (*PersonalityProfile, error) {
    // 拼接用户交互历史
    history := concatenateCells(cells)

    prompt := `基于以下用户交互历史，推断用户特质：

要求：
1. 拖延指数（0-1）：1 表示严重拖延
2. 风险偏好（0-1）：1 表示高风险偏好
3. 决策风格："rational"（理性）或 "emotional"（感性）

交互历史：
` + history

    response, err := m.llm.Generate(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 解析 LLM 输出
    profile, err := parsePersonalityProfile(response)
    if err != nil {
        return nil, err
    }

    return profile, nil
}
```

#### 3.3 软性因素调节因果推理

```go
// 修改 CausalReasoner 的边强度计算
func (r *CausalReasoner) CalculateEdgeStrength(
    baseEdge *CausalEdge,
    softFactors []*SoftFactor,
    userProfile *PersonalityProfile,
) float64 {
    strength := baseEdge.Strength

    for _, factor := range softFactors {
        // 企业文化调节
        if factor.FactorType == TypeCulture {
            if factor.Name == "customer_first" {
                // 客户第一文化抑制高频推送
                if baseEdge.To == "push_frequency_high" {
                    strength *= (1 - factor.Strength) // 抑制
                }
            }
        }

        // 哲学原则调节
        if factor.FactorType == TypePhilosophy {
            if factor.Name == "systems_thinking" {
                // 系统论思维增强长期影响权重
                if baseEdge.Metadata["time_horizon"] == "long_term" {
                    strength *= (1 + factor.Strength*0.5) // 增强
                }
            }
        }
    }

    // 个人特质调节
    if userProfile != nil {
        if userProfile.ProcrastinationIndex > 0.7 {
            // 高拖延指数，降低即时行动建议的强度
            if baseEdge.To == "immediate_action" {
                strength *= 0.7
            }
        }
    }

    return strength
}
```

#### 3.4 与 SOP 规则协同

```go
// src/memory/causal/sop_integration.go
package causal

// SOPRule 标准操作程序规则
type SOPRule struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Condition   string    `json:"condition"`   // 触发条件
    Action      string    `json:"action"`      // 执行动作
    Priority    int       `json:"priority"`    // 优先级
    CultureLink []string  `json:"culture_link"` // 关联的文化节点
}

// CheckConflict 检查软性因素与 SOP 规则的冲突
func CheckConflict(softFactors []*SoftFactor, sopRules []*SOPRule) []Conflict {
    var conflicts []Conflict

    for _, rule := range sopRules {
        for _, factor := range softFactors {
            // 例如：SOP 要求高频推送，但文化要求客户第一（可能抑制推送）
            if rule.Action == "push_high_frequency" && factor.Name == "customer_first" {
                conflicts = append(conflicts, Conflict{
                    Rule:   rule,
                    Factor: factor,
                    Type:   "inhibit",
                })
            }
        }
    }

    return conflicts
}
```

### 里程碑

| 里程碑 | 内容 | 周数 | 交付物 |
|--------|------|------|--------|
| **M3.1** | 软性因素节点定义与存储 | 1 | `soft_factor.go` |
| **M3.2** | 个人特质挖掘模块 | 2 | `personality.go`, LLM 提示词 |
| **M3.3** | 企业文化与哲学原则人工注入工具 | 1 | 配置文件，注入工具 |
| **M3.4** | 因果推理引擎整合软性因素调节 | 2 | 调节算法，测试用例 |
| **M3.5** | 端到端测试与可解释性增强 | 2 | 测试报告，解释模板 |

### 验证指标

- **软性因素对推理的影响符合预期**（专家评估）
- **用户对建议的满意度** > 85%
- **冲突检测准确率** > 90%

---

## 技术选型与集成要点

### 技术选型

| 组件 | 技术选型 | 备注 |
|------|----------|------|
| Agent 框架 | Lattice (go-agent) | 提供记忆、工具、多 Agent 协调 |
| 向量存储 | Lattice memory.Engine + PostgreSQL/pgvector | 支持向量检索与元数据过滤 |
| 图存储 | PostgreSQL (初期) / Neo4j (后期) | 因果图谱存储 |
| 嵌入模型 | DeepSeek Embedding | 统一向量维度（1024） |
| LLM 服务 | DeepSeek API / 本地部署 | 用于提取、摘要、规划等 |
| 因果估计 | DoWhy (Python) | 封装为微服务供 Go 调用 |
| 聚类算法 | Go 实现 KMeans | 初期简单实现，后期可优化 |
| 任务调度 | 内置 cron | 用于语义整合等后台任务 |
| API 框架 | FastAPI (Python) | DoWhy 微服务 |

### 关键集成点

**1. 与 go-agent 记忆系统集成**
```go
// 直接使用 memory.Engine
engine := memory.NewEngine(store, memory.DefaultOptions())

// MemCell 作为特殊类型的 MemoryRecord 存储
cell := &memcell.MemCell{...}
engine.Store(ctx, sessionID, cell.Record.Content, cell.Metadata)
```

**2. UTCP 工具封装**
```go
// 注册记忆检索工具
memcell.RegisterUTCPTools(utcpClient, retriever)

// 注册因果推理工具
causal.RegisterUTCPTools(utcpClient, reasoner)
```

**3. DoWhy 微服务调用**
```go
// HTTP/gRPC 调用 Python 服务
client := causal.NewDoWhyClient("http://localhost:8000")
effect, err := client.EstimateEffect(ctx, graph, treatment, outcome, context)
```

---

## 与 go-agent 现有能力对比

### 直接利用的现有能力

| go-agent 组件 | 用途 | 集成方式 |
|--------------|------|----------|
| `memory.Engine` | 记忆存储与检索 | 直接调用 |
| `memory.Embedder` | 向量嵌入 | 直接调用 |
| `UTCP` | 工具封装 | 直接注册 |
| `SharedSession` | 多 Agent 共享记忆 | 扩展使用 |
| `models.Agent` | LLM 调用 | 直接调用 |

### 需要扩展的部分

| 扩展点 | 说明 | 实现方式 |
|--------|------|----------|
| **类型系统** | 在 Metadata 中增加类型识别 | 新增 `memcell`/`causal` 包 |
| **后台任务** | 语义整合器 | 新增 `consolidator.go` |
| **因果图谱** | 因果节点和边 | 新增 `causal/` 模块 |
| **DoWhy 集成** | Go-Python 跨语言调用 | 新增微服务 |
| **软性因素** | 文化/哲学/个性建模 | 新增 `soft_factor.go` |

---

## 实施路线图

### 总体时间线

```
2026 Q2 (4-6 月)          2026 Q3 (7-9 月)          2026 Q4 (10-12 月)
│                        │                        │
├─ 阶段一：EverMemOS       ├─ 阶段二：REMI           ├─ 阶段三：软性因素
│  M1.1  M1.2  M1.3       │  M2.1  M2.2  M2.3       │  M3.1  M3.2  M3.3
│     │     │     │       │     │     │     │       │     │     │     │
│  1w  2w  2w  1w  1w     │  1w  2w  2w  2w  1w 2w  │  1w  2w  1w  2w  2w
│                        │                        │
└────────────────────────┴────────────────────────┴────────────────────
  8-9 周，约 120 小时        10-12 周，约 200 小时       8-10 周，约 180 小时
```

### 关键依赖

```
阶段一 (M1.1-M1.5)
    │
    ├─ 依赖：go-agent memory.Engine
    │
    └─ 产出：MemCell/MemScene 模块
            │
            ▼
阶段二 (M2.1-M2.6)
    │
    ├─ 依赖：MemCell 模块
    │
    └─ 产出：Causal 模块，DoWhy 集成
            │
            ▼
阶段三 (M3.1-M3.5)
    │
    ├─ 依赖：Causal 模块
    │
    └─ 产出：SoftFactor 模块，个人特质挖掘
```

---

## 风险评估与缓解

### 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| LLM 提取准确性低 | 高 | 中 | 人工审核 + 主动学习 + 多模型投票 |
| 因果挖掘产生虚假因果 | 高 | 中 | 置信度阈值 + 专家审核 + 时间序列验证 |
| DoWhy 集成复杂度高 | 中 | 高 | 封装为独立微服务，定义清晰 API |
| 图查询性能问题 | 中 | 中 | 图索引优化 + 缓存策略 + 分页查询 |
| 向量维度不统一 | 低 | 低 | 统一使用 DeepSeek Embedding（1024 维） |

### 业务风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 软性因素建模主观性强 | 中 | 高 | 专家输入 + 用户反馈调整 + A/B 测试 |
| 用户对因果推理不信任 | 中 | 中 | 可解释性增强 + 透明度提升 |
| 文化冲突（跨国企业） | 高 | 低 | 可配置文化节点 + 区域化适配 |

### 缓解策略

**LLM 提取准确性**：
1. **人工审核** - 初期 100% 人工审核，建立标注数据集
2. **主动学习** - 选择不确定性高的样本让人工标注
3. **多模型投票** - 使用 2-3 个 LLM，投票决定最终结果

**因果挖掘质量**：
1. **置信度阈值** - 只保留置信度 > 0.7 的因果
2. **专家审核** - 领域专家定期审核因果图谱
3. **时间序列验证** - 因果必须有时间先后顺序

**DoWhy 集成**：
1. **微服务架构** - Python 服务独立部署
2. **清晰 API** - 定义 Request/Response 格式
3. **超时与降级** - 超时返回近似估计

---

## 附录

### A. 配置示例

**企业文化配置** (`configs/culture_nodes.json`):
```json
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

**哲学原则配置** (`configs/philosophy_nodes.json`):
```json
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

### B. LLM 提示词模板

**情景提取提示词**:
```
从以下对话中提取记忆单元：

要求：
1. 识别关键事实（Fact）
2. 识别用户偏好（Preference）
3. 识别时间敏感信息
4. 为每个记忆单元分配类型（episodic/fact/preference）
5. 评估置信度（0-1）

对话：
{{conversation}}

提取结果（JSON 格式）：
[
  {
    "type": "fact",
    "content": "...",
    "confidence": 0.95
  }
]
```

**因果挖掘提示词**:
```
从以下记忆中提取因果关系：

格式：原因 → 结果（置信度 0-1）
示例：
- "用户熬夜" → "第二天工作效率低" (0.8)
- "经常运动" → "睡眠质量好" (0.7)

记忆：
{{memory_content}}

提取结果（JSON 格式）：
[
  {
    "cause": "...",
    "effect": "...",
    "confidence": 0.8
  }
]
```

### C. 测试计划

**单元测试**:
- MemCell 类型转换测试
- 因果边强度计算测试
- 软性因素调节算法测试

**集成测试**:
- 对话→MemCell→MemScene 完整流程
- 因果挖掘→推理→建议完整流程
- UTCP 工具调用测试

**端到端测试**:
- 用户使用场景测试
- 性能测试（大规模记忆和图谱）
- 准确性测试（对比人工标注）

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 7 日*  
*维护：MareMind 项目基础设施团队*
