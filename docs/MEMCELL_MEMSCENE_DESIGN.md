# MemCell & MemScene 实现方案

**项目**: Lattice - Go AI Agent 开发框架  
**创建日期**: 2026 年 3 月 6 日  
**状态**: 设计方案（待实现）  
**参考**: EverMemOS 核心思想  
**优先级**: P1 (高)  

---

## 📋 概述

本方案旨在 Lattice 框架内实现 **MemCell（记忆细胞）** 和 **MemScene（记忆场景）** 系统，借鉴 EverMemOS 的核心思想，同时充分利用 Lattice 已有的 `memory.Engine`、图感知特性和 `Shared Spaces`，构建一个具备 SOTA（State-of-the-Art）潜力的记忆系统。

---

## 🎯 核心概念

### 1. MemCell（记忆细胞）

**定义**：记忆的最小功能单元，类似于生物神经元，具有独立的编码、存储、检索和更新能力。

```
┌─────────────────────────────────────────┐
│ MemCell                                 │
│ ├─ id: string              // 唯一标识  │
│ ├─ content: string         // 记忆内容  │
│ ├─ embedding: []float32    // 向量表示  │
│ ├─ metadata: map[string]any│ 元数据    │
│ ├─ connections: []Link     // 连接关系  │
│ ├─ activation: float32     // 激活强度  │
│ ├─ decay: float32          // 衰减系数  │
│ └─ timestamp: time.Time    // 时间戳    │
└─────────────────────────────────────────┘
```

**特性**：
- **自包含**：每个 MemCell 独立管理自己的生命周期
- **可组合**：多个 MemCell 可以组合成更复杂的结构
- **可进化**：根据使用频率和重要性动态调整权重
- **可追溯**：完整的创建、修改、访问历史

---

### 2. MemScene（记忆场景）

**定义**：MemCell 的组织单元，将相关的 MemCell 按场景/上下文分组，支持快速检索和批量操作。

```
┌─────────────────────────────────────────┐
│ MemScene                                │
│ ├─ id: string              // 场景 ID    │
│ ├─ name: string            // 场景名称  │
│ ├─ cells: []MemCell        // 记忆细胞  │
│ ├─ subScenes: []MemScene   // 子场景    │
│ ├─ context: SceneContext   // 上下文    │
│ └─ accessPattern: Pattern  // 访问模式  │
└─────────────────────────────────────────┘
```

**特性**：
- **层次化**：支持嵌套的子场景结构
- **上下文感知**：每个场景有独立的上下文环境
- **动态重组**：根据访问模式自动优化组织
- **跨场景引用**：支持 Cell 在多个 Scene 中的引用

---

### 3. 与 Lattice 现有组件的映射

| EverMemOS 概念 | Lattice 对应组件 | 扩展方式 |
|---------------|------------------|----------|
| Memory Cell | `memory.MemoryRecord` | 扩展为 `MemCell` |
| Scene | `memory.Space` | 扩展为 `MemScene` |
| Engine | `memory.Engine` | 增强引擎功能 |
| Graph Edge | `model.GraphEdge` | 直接使用 |
| Embedding | `embed.Embedder` | 直接使用 |

---

## 🏗️ 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  SubAgent1  │  │  SubAgent2  │  │  SubAgent3  │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
│         │                │                │                 │
│         └────────────────┼────────────────┘                 │
│                          │                                  │
│                  ┌───────▼────────┐                         │
│                  │  SharedSession │                         │
│                  └───────┬────────┘                         │
└──────────────────────────┼──────────────────────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────┐
│                  Memory Layer                                │
│                          │                                  │
│         ┌────────────────▼────────────────┐                 │
│         │        MemSceneManager          │                 │
│         │  ├─ scenes: map[string]MemScene │                 │
│         │  ├─ cellIndex: map[string]Cell  │                 │
│         │  └─ graphStore: GraphStore      │                 │
│         └────────────────┬────────────────┘                 │
│                          │                                  │
│    ┌─────────────────────┼─────────────────────┐            │
│    │                     │                     │            │
│ ┌──▼──────┐  ┌──────────▼──────────┐  ┌──────▼──────┐      │
│ │ Scene 1 │  │      Scene 2        │  │   Scene 3   │      │
│ │ ├─Cell1 │  │  ├─ Cell4           │  │  ├─ Cell7   │      │
│ │ ├─Cell2 │  │  ├─ Cell5           │  │  ├─ Cell8   │      │
│ │ └─Cell3 │  │  └─ Cell6           │  │  └─ Cell9   │      │
│ └─────────┘  └─────────────────────┘  └─────────────┘      │
│                                                             │
│         ┌──────────────────────────────────┐                │
│         │        memory.Engine             │                │
│         │  ├─ Store()                      │                │
│         │  ├─ Retrieve()                   │                │
│         │  ├─ Prune()                      │                │
│         │  └─ Score()                      │                │
│         └──────────────────────────────────┘                │
│                                                             │
│         ┌──────────────────────────────────┐                │
│         │        VectorStore               │                │
│         │  ├─ InMemory / Postgres / Qdrant │                │
│         │  └─ GraphStore (Neo4j)           │                │
│         └──────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

---

## 📝 代码结构草案

### 文件组织

```
src/memory/
├── cell/
│   ├── cell.go              # MemCell 核心定义
│   ├── cell_test.go
│   ├── builder.go           # MemCell 构建器
│   ├── lifecycle.go         # 生命周期管理
│   └── encoding.go          # 序列化/反序列化
│
├── scene/
│   ├── scene.go             # MemScene 核心定义
│   ├── scene_test.go
│   ├── manager.go           # MemSceneManager
│   ├── context.go           # SceneContext
│   └── hierarchy.go         # 层次化管理
│
├── graph/
│   ├── graph.go             # 图结构管理
│   ├── edge.go              # 边（连接）定义
│   ├── traversal.go         # 图遍历算法
│   └── clustering.go        # 聚类算法
│
├── activation/
│   ├── activation.go        # 激活函数
│   ├── decay.go             # 衰减模型
│   └─  spreading.go         # 激活扩散
│
└── integration/
    ├── engine_ext.go        # memory.Engine 扩展
    ├── session_ext.go       # SessionMemory 扩展
    └── adk_module.go        # ADK 模块集成
```

---

### 核心代码实现

#### 1. MemCell 定义 (`cell/cell.go`)

```go
package cell

import (
	"context"
	"sync"
	"time"

	"github.com/Protocol-Lattice/go-agent/src/memory/model"
)

// CellType defines the type of memory cell
type CellType string

const (
	CellTypeEpisodic    CellType = "episodic"    // 情景记忆
	CellTypeSemantic    CellType = "semantic"    // 语义记忆
	CellTypeProcedural  CellType = "procedural"  // 程序记忆
	CellTypeEmotional   CellType = "emotional"   // 情感记忆
)

// CellState defines the current state of a cell
type CellState int

const (
	StateActive CellState = iota
	StateDormant
	StateDecaying
	StateArchived
	StateDeleted
)

// MemCell represents a single memory unit
type MemCell struct {
	mu sync.RWMutex

	// Identity
	ID        string    `json:"id"`
	Type      CellType  `json:"type"`
	State     CellState `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Content
	Content   string                 `json:"content"`
	Embedding []float32              `json:"embedding,omitempty"`
	Metadata  map[string]any         `json:"metadata"`

	// Connections
	Incoming []string `json:"incoming,omitempty"` // IDs of cells pointing to this
	Outgoing []string `json:"outgoing,omitempty"` // IDs of cells this points to

	// Dynamics
	Activation    float32   `json:"activation"`     // Current activation level (0-1)
	BaseStrength  float32   `json:"base_strength"`  // Intrinsic strength (0-1)
	DecayRate     float32   `json:"decay_rate"`     // Decay per second
	LastAccessed  time.Time `json:"last_accessed"`
	AccessCount   int64     `json:"access_count"`

	// Context
	SessionID string   `json:"session_id"`
	SceneIDs  []string `json:"scene_ids"`
}

// NewMemCell creates a new memory cell with default values
func NewMemCell(content string, cellType CellType) *MemCell {
	now := time.Now()
	return &MemCell{
		ID:           generateCellID(),
		Type:         cellType,
		State:        StateActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Content:      content,
		Metadata:     make(map[string]any),
		Incoming:     make([]string, 0),
		Outgoing:     make([]string, 0),
		Activation:   1.0,
		BaseStrength: 0.5,
		DecayRate:    0.01, // 1% per second
		LastAccessed: now,
		AccessCount:  0,
		SceneIDs:     make([]string, 0),
	}
}

// Access updates the cell's activation and access statistics
func (c *MemCell) Access(ctx context.Context, boost float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Activation = min(1.0, c.Activation+boost)
	c.LastAccessed = time.Now()
	c.AccessCount++

	// Update metadata
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	c.Metadata["last_access"] = c.LastAccessed.Format(time.RFC3339)
	c.Metadata["access_count"] = c.AccessCount
}

// Decay applies the decay model to the cell
func (c *MemCell) Decay(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.LastAccessed).Seconds()
	decay := c.DecayRate * elapsed

	c.Activation = max(0.0, c.Activation-decay)

	// State transition based on activation
	if c.Activation < 0.1 && c.State == StateActive {
		c.State = StateDormant
	} else if c.Activation > 0.3 && c.State == StateDormant {
		c.State = StateActive
	}
}

// Connect creates a connection to another cell
func (c *MemCell) Connect(targetID string, edgeType model.EdgeType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if connection already exists
	for _, id := range c.Outgoing {
		if id == targetID {
			return
		}
	}

	c.Outgoing = append(c.Outgoing, targetID)
	c.Metadata["connections_updated"] = time.Now()
}

// Disconnect removes a connection from another cell
func (c *MemCell) Disconnect(targetID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove from outgoing
	for i, id := range c.Outgoing {
		if id == targetID {
			c.Outgoing = append(c.Outgoing[:i], c.Outgoing[i+1:]...)
			break
		}
	}
}

// ToMemoryRecord converts to legacy MemoryRecord for compatibility
func (c *MemCell) ToMemoryRecord() model.MemoryRecord {
	return model.MemoryRecord{
		ID:        c.ID,
		SessionID: c.SessionID,
		Content:   c.Content,
		Metadata:  model.EncodeMetadata(c.Metadata),
		Embedding: c.Embedding,
		Timestamp: c.CreatedAt,
	}
}

// Helper functions
func generateCellID() string {
	// Use UUID or similar
	return model.GenerateID()
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
```

---

#### 2. MemScene 定义 (`scene/scene.go`)

```go
package scene

import (
	"context"
	"sync"
	"time"

	"github.com/Protocol-Lattice/go-agent/src/memory/cell"
	"github.com/Protocol-Lattice/go-agent/src/memory/model"
)

// SceneContext defines the context for a scene
type SceneContext struct {
	SessionID   string            `json:"session_id"`
	UserID      string            `json:"user_id"`
	TaskType    string            `json:"task_type"`
	Domain      string            `json:"domain"`
	Priority    int               `json:"priority"`
	Tags        []string          `json:"tags"`
	CustomData  map[string]any    `json:"custom_data"`
}

// AccessPattern tracks how a scene is accessed
type AccessPattern struct {
	mu sync.RWMutex

	TotalAccesses   int64       `json:"total_accesses"`
	RecentAccesses  []time.Time `json:"recent_accesses"`
	PeakHours       []int       `json:"peak_hours"` // 0-23
	AvgAccessesPerDay float64   `json:"avg_accesses_per_day"`
}

// MemScene represents a collection of related memory cells
type MemScene struct {
	mu sync.RWMutex

	// Identity
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Hierarchy
	ParentID   string            `json:"parent_id,omitempty"`
	SubScenes  []string          `json:"sub_scenes"` // IDs of child scenes

	// Cells
	Cells       map[string]*cell.MemCell `json:"cells"`
	CellOrder   []string                 `json:"cell_order"` // Maintain order

	// Context
	Context     SceneContext         `json:"context"`
	AccessInfo  AccessPattern        `json:"access_info"`

	// Graph
	GraphStore  model.GraphStore   `json:"-"` // Reference to graph store
}

// NewMemScene creates a new memory scene
func NewMemScene(name string, ctx SceneContext) *MemScene {
	now := time.Now()
	return &MemScene{
		ID:          generateSceneID(),
		Name:        name,
		CreatedAt:   now,
		UpdatedAt:   now,
		SubScenes:   make([]string, 0),
		Cells:       make(map[string]*cell.MemCell),
		CellOrder:   make([]string, 0),
		Context:     ctx,
		AccessInfo: AccessPattern{
			RecentAccesses: make([]time.Time, 0),
			PeakHours:      make([]int, 0),
		},
	}
}

// AddCell adds a memory cell to the scene
func (s *MemScene) AddCell(ctx context.Context, c *cell.MemCell) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Cells[c.ID]; exists {
		return nil // Already exists
	}

	s.Cells[c.ID] = c
	s.CellOrder = append(s.CellOrder, c.ID)
	c.SceneIDs = append(c.SceneIDs, s.ID)

	s.UpdatedAt = time.Now()
	return nil
}

// RemoveCell removes a memory cell from the scene
func (s *MemScene) RemoveCell(ctx context.Context, cellID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Cells[cellID]; !exists {
		return nil
	}

	delete(s.Cells, cellID)

	// Remove from order
	for i, id := range s.CellOrder {
		if id == cellID {
			s.CellOrder = append(s.CellOrder[:i], s.CellOrder[i+1:]...)
			break
		}
	}

	s.UpdatedAt = time.Now()
	return nil
}

// GetCell retrieves a cell by ID
func (s *MemScene) GetCell(cellID string) (*cell.MemCell, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cell, exists := s.Cells[cellID]
	return cell, exists
}

// ListCells returns all cells in the scene
func (s *MemScene) ListCells() []*cell.MemCell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cells := make([]*cell.MemCell, 0, len(s.Cells))
	for _, id := range s.CellOrder {
		if cell, exists := s.Cells[id]; exists {
			cells = append(cells, cell)
		}
	}
	return cells
}

// AddSubScene adds a child scene
func (s *MemScene) AddSubScene(subSceneID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range s.SubScenes {
		if id == subSceneID {
			return
		}
	}

	s.SubScenes = append(s.SubScenes, subSceneID)
}

// Access records an access event
func (s *MemScene) Access(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AccessInfo.mu.Lock()
	defer s.AccessInfo.mu.Unlock()

	s.AccessInfo.TotalAccesses++
	s.AccessInfo.RecentAccesses = append(s.AccessInfo.RecentAccesses, time.Now())

	// Keep only recent accesses (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	recent := make([]time.Time, 0)
	for _, t := range s.AccessInfo.RecentAccesses {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	s.AccessInfo.RecentAccesses = recent

	// Update peak hours
	hourCounts := make(map[int]int)
	for _, t := range recent {
		hourCounts[int(t.Hour())]++
	}

	// Find top 3 peak hours
	// ... (implementation)
}

// GetActiveCells returns cells with activation > threshold
func (s *MemScene) GetActiveCells(threshold float32) []*cell.MemCell {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]*cell.MemCell, 0)
	for _, c := range s.Cells {
		if c.Activation > threshold {
			active = append(active, c)
		}
	}
	return active
}

// Helper functions
func generateSceneID() string {
	return model.GenerateID()
}
```

---

#### 3. MemSceneManager (`scene/manager.go`)

```go
package scene

import (
	"context"
	"sync"

	"github.com/Protocol-Lattice/go-agent/src/memory/cell"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
)

// MemSceneManager manages all scenes and cells
type MemSceneManager struct {
	mu sync.RWMutex

	scenes     map[string]*MemScene
	cellIndex  map[string]*cell.MemCell // Global cell index
	rootScenes []string                 // Top-level scenes

	graphStore store.GraphStore
	vectorStore store.VectorStore

	config *ManagerConfig
}

// ManagerConfig holds configuration for the manager
type ManagerConfig struct {
	MaxCellsPerScene     int
	MaxScenes            int
	EnableAutoPruning    bool
	ActivationThreshold float32
	GlobalDecayRate      float32
}

// DefaultConfig returns default configuration
func DefaultConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxCellsPerScene:     1000,
		MaxScenes:            100,
		EnableAutoPruning:    true,
		ActivationThreshold:  0.2,
		GlobalDecayRate:      0.001,
	}
}

// NewMemSceneManager creates a new scene manager
func NewMemSceneManager(graphStore store.GraphStore, vectorStore store.VectorStore, config *ManagerConfig) *MemSceneManager {
	if config == nil {
		config = DefaultConfig()
	}

	return &MemSceneManager{
		scenes:      make(map[string]*MemScene),
		cellIndex:   make(map[string]*cell.MemCell),
		rootScenes:  make([]string, 0),
		graphStore:  graphStore,
		vectorStore: vectorStore,
		config:      config,
	}
}

// CreateScene creates a new scene
func (m *MemSceneManager) CreateScene(ctx context.Context, name string, context SceneContext) (*MemScene, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.scenes) >= m.config.MaxScenes {
		return nil, ErrMaxScenesReached
	}

	scene := NewMemScene(name, context)
	m.scenes[scene.ID] = scene

	if context.ParentID == "" {
		m.rootScenes = append(m.rootScenes, scene.ID)
	}

	return scene, nil
}

// GetScene retrieves a scene by ID
func (m *MemSceneManager) GetScene(sceneID string) (*MemScene, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scene, exists := m.scenes[sceneID]
	return scene, exists
}

// AddCellToScene adds a cell to a scene
func (m *MemSceneManager) AddCellToScene(ctx context.Context, sceneID string, c *cell.MemCell) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.scenes[sceneID]
	if !exists {
		return ErrSceneNotFound
	}

	if len(scene.Cells) >= m.config.MaxCellsPerScene {
		return ErrMaxCellsReached
	}

	// Add to scene
	if err := scene.AddCell(ctx, c); err != nil {
		return err
	}

	// Update global index
	m.cellIndex[c.ID] = c

	// Persist to vector store
	if m.vectorStore != nil {
		record := c.ToMemoryRecord()
		if err := m.vectorStore.StoreMemory(ctx, c.SessionID, c.Content, c.Metadata, c.Embedding); err != nil {
			return err
		}
	}

	// Persist graph connections
	if m.graphStore != nil {
		// Store cell connections
		for _, targetID := range c.Outgoing {
			edge := model.GraphEdge{
				From: c.ID,
				To:   targetID,
				Type: model.EdgeFollows, // Default type
			}
			if err := m.graphStore.AddEdge(ctx, edge); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCell retrieves a cell by ID from global index
func (m *MemSceneManager) GetCell(cellID string) (*cell.MemCell, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cell, exists := m.cellIndex[cellID]
	return cell, exists
}

// SearchCells searches for cells across all scenes
func (m *MemSceneManager) SearchCells(ctx context.Context, query string, limit int) ([]*cell.MemCell, error) {
	// Use vector search if available
	if m.vectorStore != nil {
		// Embed query
		// Search vector store
		// Return matching cells
	}

	// Fallback to linear search
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*cell.MemCell, 0)
	for _, c := range m.cellIndex {
		if contains(c.Content, query) {
			results = append(results, c)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// ApplyDecay applies decay to all cells
func (m *MemSceneManager) ApplyDecay(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.cellIndex {
		c.Decay(ctx)
	}
}

// PruneInactive removes cells below activation threshold
func (m *MemSceneManager) PruneInactive(ctx context.Context) {
	if !m.config.EnableAutoPruning {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := make([]string, 0)
	for id, c := range m.cellIndex {
		if c.Activation < m.config.ActivationThreshold && c.State == cell.StateDecaying {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(m.cellIndex, id)
		// Also remove from scenes
		// ...
	}
}

// Error definitions
var (
	ErrSceneNotFound     = &Error{"scene not found"}
	ErrMaxScenesReached  = &Error{"maximum scenes reached"}
	ErrMaxCellsReached   = &Error{"maximum cells per scene reached"}
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
```

---

#### 4. 激活扩散算法 (`activation/spreading.go`)

```go
package activation

import (
	"context"
	"container/heap"

	"github.com/Protocol-Lattice/go-agent/src/memory/cell"
)

// SpreadingActivation implements the activation spreading algorithm
type SpreadingActivation struct {
	decayFactor    float32 // Decay per hop
	threshold      float32 // Minimum activation to continue spreading
	maxIterations  int     // Maximum iterations
}

// NewSpreadingActivation creates a new spreading activation engine
func NewSpreadingActivation(decay, threshold float32, maxIter int) *SpreadingActivation {
	return &SpreadingActivation{
		decayFactor:   decay,
		threshold:     threshold,
		maxIterations: maxIter,
	}
}

// ActivationQueueItem for priority queue
type ActivationQueueItem struct {
	cellID     string
	activation float32
	index      int
}

// ActivationQueue implements heap.Interface
type ActivationQueue []*ActivationQueueItem

func (pq ActivationQueue) Len() int { return len(pq) }

func (pq ActivationQueue) Less(i, j int) bool {
	return pq[i].activation > pq[j].activation // Higher activation first
}

func (pq ActivationQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *ActivationQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*ActivationQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *ActivationQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// Spread applies activation spreading from seed cells
func (sa *SpreadingActivation) Spread(
	ctx context.Context,
	seedCells []*cell.MemCell,
	getNeighbors func(cellID string) []string,
	getCell func(cellID string) *cell.MemCell,
) map[string]float32 {
	
	// Result: cellID -> final activation
	activations := make(map[string]float32)

	// Initialize priority queue
	pq := make(ActivationQueue, 0)
	heap.Init(&pq)

	// Add seed cells
	for _, c := range seedCells {
		heap.Push(&pq, &ActivationQueueItem{
			cellID:     c.ID,
			activation: c.Activation,
		})
		activations[c.ID] = c.Activation
	}

	iterations := 0
	for pq.Len() > 0 && iterations < sa.maxIterations {
		item := heap.Pop(&pq).(*ActivationQueueItem)

		// Skip if activation is below threshold
		if item.activation < sa.threshold {
			continue
		}

		// Get neighbors
		neighbors := getNeighbors(item.cellID)

		// Spread activation to neighbors
		spreadActivation := item.activation * sa.decayFactor

		for _, neighborID := range neighbors {
			neighbor := getCell(neighborID)
			if neighbor == nil {
				continue
			}

			// Calculate new activation
			newActivation := neighbor.Activation + spreadActivation
			newActivation = min(1.0, newActivation)

			// Update if significant change
			if newActivation-activations[neighborID] > sa.threshold {
				activations[neighborID] = newActivation

				heap.Push(&pq, &ActivationQueueItem{
					cellID:     neighborID,
					activation: newActivation,
				})
			}
		}

		iterations++
	}

	return activations
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
```

---

#### 5. 与 memory.Engine 集成 (`integration/engine_ext.go`)

```go
package integration

import (
	"context"

	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/cell"
	"github.com/Protocol-Lattice/go-agent/src/memory/scene"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
)

// ExtendedEngine wraps memory.Engine with MemCell/MemScene support
type ExtendedEngine struct {
	baseEngine  *memory.Engine
	sceneManager *scene.MemSceneManager
}

// NewExtendedEngine creates a new extended engine
func NewExtendedEngine(
	baseEngine *memory.Engine,
	graphStore store.GraphStore,
	vectorStore store.VectorStore,
) *ExtendedEngine {
	sceneManager := scene.NewMemSceneManager(graphStore, vectorStore, scene.DefaultConfig())

	return &ExtendedEngine{
		baseEngine:   baseEngine,
		sceneManager: sceneManager,
	}
}

// StoreCell stores a memory cell
func (e *ExtendedEngine) StoreCell(ctx context.Context, c *cell.MemCell, sceneID string) error {
	// Store using base engine
	record := c.ToMemoryRecord()
	_, err := e.baseEngine.Store(ctx, c.SessionID, c.Content, c.Metadata)
	if err != nil {
		return err
	}

	// Add to scene
	return e.sceneManager.AddCellToScene(ctx, sceneID, c)
}

// RetrieveCells retrieves cells with activation spreading
func (e *ExtendedEngine) RetrieveCells(
	ctx context.Context,
	sessionID string,
	query string,
	limit int,
) ([]*cell.MemCell, error) {
	// 1. Search for initial cells
	initialCells, err := e.sceneManager.SearchCells(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}

	// 2. Apply activation spreading
	// (Implementation depends on graph structure)

	// 3. Return top activated cells
	activeCells := e.sceneManager.GetActiveCells(0.3)
	if len(activeCells) > limit {
		activeCells = activeCells[:limit]
	}

	return activeCells, nil
}

// GetSceneManager returns the scene manager
func (e *ExtendedEngine) GetSceneManager() *scene.MemSceneManager {
	return e.sceneManager
}

// CreateScene creates a new scene
func (e *ExtendedEngine) CreateScene(
	ctx context.Context,
	name string,
	context scene.SceneContext,
) (*scene.MemScene, error) {
	return e.sceneManager.CreateScene(ctx, name, context)
}
```

---

#### 6. ADK 模块集成 (`integration/adk_module.go`)

```go
package integration

import (
	"context"

	"github.com/Protocol-Lattice/go-agent/src/adk"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
)

// MemCellModule integrates MemCell/MemScene with ADK
type MemCellModule struct {
	name        string
	graphStore  store.GraphStore
	vectorStore store.VectorStore
	config      *ModuleConfig
}

// ModuleConfig holds module configuration
type ModuleConfig struct {
	EnableMemCell       bool
	EnableSceneHierarchy bool
	EnableActivation    bool
	DecayRate           float32
}

// NewMemCellModule creates a new MemCell module
func NewMemCellModule(
	name string,
	graphStore store.GraphStore,
	vectorStore store.VectorStore,
	config *ModuleConfig,
) *MemCellModule {
	if config == nil {
		config = DefaultModuleConfig()
	}

	return &MemCellModule{
		name:        name,
		graphStore:  graphStore,
		vectorStore: vectorStore,
		config:      config,
	}
}

// Name returns the module name
func (m *MemCellModule) Name() string {
	return m.name
}

// Provision attaches the module to the ADK
func (m *MemCellModule) Provision(ctx context.Context, kit *adk.AgentDevelopmentKit) error {
	if !m.config.EnableMemCell {
		return nil
	}

	// Get or create memory provider
	// Wrap with ExtendedEngine
	// Register with kit

	return nil
}

// DefaultModuleConfig returns default configuration
func DefaultModuleConfig() *ModuleConfig {
	return &ModuleConfig{
		EnableMemCell:        true,
		EnableSceneHierarchy: true,
		EnableActivation:     true,
		DecayRate:            0.001,
	}
}
```

---

## 🔗 与 Lattice 现有功能集成

### 1. 与 memory.Engine 集成

```go
// 现有 Engine 保持不变，通过扩展方式添加功能
engine := memory.NewEngine(store, opts)
extendedEngine := integration.NewExtendedEngine(engine, graphStore, vectorStore)

// 使用扩展功能
scene, _ := extendedEngine.CreateScene(ctx, "工作项目", scene.SceneContext{
    SessionID: "session-123",
    TaskType:  "coding",
})

cell := cell.NewMemCell("实现了用户认证功能", cell.CellTypeEpisodic)
extendedEngine.StoreCell(ctx, cell, scene.ID)
```

### 2. 与 SharedSession 集成

```go
// SharedSession 可以管理多个 MemScene
shared := memory.NewSharedSession(sessionID)

// 创建场景
workScene := scene.NewMemScene("工作", scene.SceneContext{
    Domain: "professional",
})
personalScene := scene.NewMemScene("个人", scene.SceneContext{
    Domain: "personal",
})

// 添加到共享会话
shared.AddScene(workScene)
shared.AddScene(personalScene)
```

### 3. 与 ADK 集成

```go
kit, _ := adk.New(ctx,
    adk.WithModules(
        integration.NewMemCellModule("memcell", graphStore, vectorStore, nil),
        adkmodules.InQdrantMemory(100000, qdrantURL, collection, embedder, &memOpts),
    ),
)
```

---

## 📊 实现路线图

### 阶段 1：核心数据结构（2 周）
- [ ] 实现 `MemCell` 基础结构
- [ ] 实现 `MemScene` 基础结构
- [ ] 实现 `MemSceneManager`
- [ ] 编写单元测试

### 阶段 2：激活模型（1 周）
- [ ] 实现激活衰减算法
- [ ] 实现激活扩散算法
- [ ] 实现优先级队列
- [ ] 性能基准测试

### 阶段 3：图集成（1 周）
- [ ] 实现图遍历算法
- [ ] 实现聚类算法
- [ ] 与 Neo4j 集成
- [ ] 图查询优化

### 阶段 4：Engine 集成（1 周）
- [ ] 实现 `ExtendedEngine`
- [ ] 与现有 `memory.Engine` 集成
- [ ] 向后兼容性测试
- [ ] 性能优化

### 阶段 5：ADK 集成（1 周）
- [ ] 实现 `MemCellModule`
- [ ] ADK 模块测试
- [ ] 示例代码
- [ ] 文档编写

### 阶段 6：高级功能（2 周）
- [ ] 场景层次优化
- [ ] 自动 pruning 策略
- [ ] 访问模式学习
- [ ] SOTA 特性实验

**总预计时间**: 8 周

---

## 🧪 测试策略

### 单元测试

```go
func TestMemCell_Access(t *testing.T) {
    cell := cell.NewMemCell("test content", cell.CellTypeEpisodic)
    
    // Initial activation
    if cell.Activation != 1.0 {
        t.Errorf("expected initial activation 1.0, got %f", cell.Activation)
    }
    
    // Access with boost
    cell.Access(context.Background(), 0.5)
    
    // Activation should be capped at 1.0
    if cell.Activation != 1.0 {
        t.Errorf("expected activation 1.0 after boost, got %f", cell.Activation)
    }
}

func TestMemScene_AddCell(t *testing.T) {
    scene := scene.NewMemScene("test", scene.SceneContext{})
    c := cell.NewMemCell("test", cell.CellTypeEpisodic)
    
    err := scene.AddCell(context.Background(), c)
    if err != nil {
        t.Fatalf("AddCell failed: %v", err)
    }
    
    // Verify cell was added
    retrieved, exists := scene.GetCell(c.ID)
    if !exists {
        t.Error("cell not found after adding")
    }
    if retrieved.ID != c.ID {
        t.Error("retrieved cell ID mismatch")
    }
}
```

### 集成测试

```go
func TestExtendedEngine_Integration(t *testing.T) {
    // Setup
    vectorStore := store.NewInMemoryStore()
    graphStore := store.NewInMemoryGraph()
    engine := memory.NewEngine(vectorStore, memory.DefaultOptions())
    extEngine := integration.NewExtendedEngine(engine, graphStore, vectorStore)
    
    // Create scene
    scene, _ := extEngine.CreateScene(context.Background(), "test", scene.SceneContext{})
    
    // Create and store cell
    c := cell.NewMemCell("test content", cell.CellTypeEpisodic)
    extEngine.StoreCell(context.Background(), c, scene.ID)
    
    // Retrieve
    cells, _ := extEngine.RetrieveCells(context.Background(), "session", "test", 10)
    
    if len(cells) != 1 {
        t.Errorf("expected 1 cell, got %d", len(cells))
    }
}
```

---

## 📚 参考资料

### EverMemOS
- [EverMemOS GitHub](https://github.com/evermemos/evermemos)
- [EverMemOS 论文](待补充)

### 记忆系统研究
- [Human Memory Models](https://en.wikipedia.org/wiki/Memory_model)
- [Spreading Activation](https://en.wikipedia.org/wiki/Spreading_activation)
- [ACT-R Cognitive Architecture](https://act-r.psy.cmu.edu/)

### Lattice 现有文档
- [memory.Engine 文档](../src/memory/engine/engine.go)
- [SharedSession 文档](../src/memory/session/shared_session.go)
- [GraphStore 接口](../src/memory/store/vector_store.go)

---

## ✅ 实现检查清单

### 核心功能
- [ ] `MemCell` 结构体实现
- [ ] `MemScene` 结构体实现
- [ ] `MemSceneManager` 实现
- [ ] 激活衰减算法
- [ ] 激活扩散算法
- [ ] 图遍历算法

### 集成
- [ ] 与 `memory.Engine` 集成
- [ ] 与 `SharedSession` 集成
- [ ] 与 ADK 模块集成
- [ ] 与 `GraphStore` 集成
- [ ] 与 `VectorStore` 集成

### 测试
- [ ] 单元测试（覆盖率 > 80%）
- [ ] 集成测试
- [ ] 性能基准测试
- [ ] 压力测试

### 文档
- [ ] API 文档
- [ ] 使用示例
- [ ] 最佳实践
- [ ] 性能调优指南

---

*本文档为设计方案，具体实现时可能需要调整。*
