# go-agent 代码页清单

**项目**: Lattice - Go AI Agent 开发框架  
**生成日期**: 2026 年 3 月 6 日  
**Go 版本**: 1.25  

---

## 📁 目录结构总览

```
go-agent/
├── 根目录核心文件 (11 个)
├── src/                    # 核心功能模块
│   ├── adk/               # Agent Development Kit (5 个文件)
│   ├── memory/            # 记忆系统 (5 个子目录)
│   ├── models/            # LLM 提供商适配器 (10 个文件)
│   ├── subagents/         # 预构建专家代理 (2 个文件)
│   ├── swarm/             # 多代理协调 (4 个文件)
│   ├── cache/             # 缓存工具
│   ├── concurrent/        # 并发工具
│   └── helpers/           # 辅助函数
└── cmd/                   # 示例和演示程序
    ├── example/           # 功能示例 (15+ 个)
    ├── demo/              # 交互式演示
    └── ...
```

---

## 📄 第一部分：根目录核心文件

### 1. `agent.go` (1819 行)
**主要目的**: Agent 核心运行时实现

**核心功能**:
- `Agent` 结构体定义：协调模型调用、记忆系统、工具和子代理
- 构造函数 `New()` 和配置选项 `Options`
- 核心方法：`Generate()`, `GenerateWithFiles()`, `Save()`
- 工具调用基础设施：`CallTool()`, `executeTool()`
- 子代理委托：`handleCommand()`, `delegateToSubAgent()`
- UTCP 集成：`AsUTCPTool()`, `RegisterAsUTCPProvider()`
- 状态持久化：`Checkpoint()`, `Restore()`

**关键依赖**:
- `src/memory.SessionMemory` - 会话记忆
- `src/models.Agent` - LLM 接口
- `ToolCatalog` / `SubAgentDirectory` - 工具和代理注册表
- `utcp.UtcpClientInterface` - UTCP 客户端

**被引用**: 几乎所有模块都依赖此文件

---

### 2. `agent_tool.go` (318 行)
**主要目的**: Agent 作为工具的适配器实现

**核心功能**:
- `SubAgentTool` - 将子代理适配为 Tool 接口
- `AgentToolAdapter` - 将 Agent 适配为 UTCP 工具
- `agentCLITransport` - CLI 传输层适配器
- UTCP 工具注册和调用逻辑

**关键依赖**:
- `utcp` 包及其子模块
- `Tool` 接口定义

**关联**: 与 `agent.go` 的 UTCP 功能配合使用

---

### 3. `agent_stream.go` (168 行)
**主要目的**: 流式响应接口实现

**核心功能**:
- `GenerateStream()` - 流式生成方法
- 预取优化：并行执行上下文检索和工具发现
- 快速路径：直接工具调用、子代理命令、CodeMode
- 回退到 LLM 流式完成

**关键依赖**:
- `agent.go` 的 `retrieveContext()`, `executeTool()` 等
- `models.StreamChunk` - 流式数据块结构

**关联**: `agent.go` 的流式扩展

---

### 4. `agent_orchestrators.go` (260 行)
**主要目的**: LLM 驱动的工具编排引擎

**核心功能**:
- `codeChainOrchestrator()` - 多步骤 UTCP 链规划
- `toolOrchestrator()` - 单工具选择引擎
- TOON-Go 结构化推理层
- 链式执行：`[]ChainStep` 生成和执行

**关键依赖**:
- `utcp/src/plugins/chain` - UTCP 链插件
- `gotoon` - TOON 序列化

**关联**: 与 `CodeMode` 和 `CodeChain` 配合使用

---

### 5. `types.go` (100+ 行)
**主要目的**: 核心接口和类型定义

**核心功能**:
- `Tool` / `ToolSpec` / `ToolRequest` / `ToolResponse` - 工具系统接口
- `ToolCatalog` - 工具注册表接口
- `SubAgent` / `SubAgentDirectory` - 子代理接口
- `AgentState` - 可序列化状态
- `SafetyPolicy` / `FormatEnforcer` - 安全策略接口
- `OutputGuardrails` - 输出护栏

**被引用**: 整个项目的基础类型定义

---

### 6. `catalog.go` (150+ 行)
**主要目的**: 工具和子代理注册表实现

**核心功能**:
- `StaticToolCatalog` - 内存工具注册表
  - 线程安全的工具注册和查找
  - 按注册顺序返回工具列表
- `StaticSubAgentDirectory` - 内存子代理目录
  - 子代理注册和查找
  - 保持插入顺序

**关键依赖**: `types.go` 的接口定义

---

### 7. `query.go` (50+ 行)
**主要目的**: 查询分类优化

**核心功能**:
- `classifyQuery()` - 将输入分类为：
  - `QueryComplex` - 复杂查询（使用完整记忆）
  - `QueryShortFactoid` - 简短事实（跳过记忆检索）
  - `QueryMath` - 数学表达式（直接计算）
- 正则表达式匹配和单字启发式

**关联**: 被 `agent.go` 用于优化记忆检索

---

### 8. `helpers.go` (20+ 行)
**主要目的**: 辅助函数

**核心功能**:
- `Save()` - 保存对话轮次到共享空间
- 批量持久化到长期记忆

**关联**: `agent.go` 的简化辅助方法

---

### 9. `safety_policies.go` (91 行)
**主要目的**: 输出安全策略实现

**核心功能**:
- `RegexBlocklistPolicy` - 正则表达式黑名单
- `LLMEvaluatorPolicy` - 使用 LLM 评估安全性
  - 检测仇恨言论、危险指令、PII 等
  - 可配置的评估提示模板

**关键依赖**: `types.go` 的 `SafetyPolicy` 接口

---

### 10. 测试文件
**主要目的**: 单元测试

**覆盖范围**:
- `agent_test.go` - Agent 核心功能测试
- `agent_tool_test.go` - 工具调用测试
- `agent_security_test.go` - 安全策略测试
- `agent_checkpoint_test.go` - 检查点/恢复测试

---

## 📦 第二部分：src/ 模块

### 2.1 ADK (Agent Development Kit)

**路径**: `src/adk/`

#### `kit.go` (467 行)
**主要目的**: ADK 核心编排器

**核心功能**:
- `AgentDevelopmentKit` 结构体
  - 模块系统管理
  - 依赖注入容器
- 构造函数 `New()` 和引导流程 `Bootstrap()`
- 提供者注册方法：
  - `UseModelProvider()`, `UseMemoryProvider()`
  - `UseToolProvider()`, `UseSubAgentProvider()`
- Agent 构建：`BuildAgent()`

**关键依赖**:
- `Module` 接口
- 各类 Provider 接口
- `agent.Agent`

**被引用**: 所有使用 ADK 的应用程序

---

#### `module.go` (20 行)
**主要目的**: 模块接口定义

**核心功能**:
- `Module` 接口：
  - `Name()` - 模块名称
  - `Provision()` - 附加功能到 ADK

**被引用**: 所有 ADK 模块实现

---

#### `options.go`
**主要目的**: ADK 配置选项

**核心功能**:
- `Option` 函数类型
- 配置函数：
  - `WithDefaultSystemPrompt()`
  - `WithSubAgents()`
  - `WithModules()`
  - `WithCodeModeUtcp()`
  - `WithCodeChain()`

---

#### `providers.go`
**主要目的**: 提供者接口定义

**核心功能**:
- `ModelProvider` - 模型提供者接口
- `MemoryProvider` - 记忆提供者接口
- `ToolProvider` - 工具提供者接口
- `SubAgentProvider` - 子代理提供者接口
- `AgentOption` - Agent 配置选项

---

#### `modules/` 子目录

##### `tool_module.go`
**目的**: 工具注册模块
- `ToolModule` 结构体
- 注册工具提供者到 ADK

##### `memory_module.go`
**目的**: 记忆注册模块
- `MemoryModule` 结构体
- 注册会话记忆提供者

##### `model_module.go`
**目的**: 模型注册模块
- `ModelModule` 结构体
- 注册 LLM 提供者

##### `subagent_module.go`
**目的**: 子代理注册模块
- 注册子代理提供者

##### `helpers.go`
**目的**: ADK 辅助函数
- 模块创建辅助函数：
  - `NewModelModule()`
  - `InQdrantMemory()`
  - 等便捷函数

---

### 2.2 Memory (记忆系统)

**路径**: `src/memory/`

#### `memory.go` (118 行)
**主要目的**: 类型别名和公共 API 导出

**核心功能**:
- 重新导出所有子包的类型
- 便捷函数：
  - `NewEngine()`, `NewSessionMemory()`
  - `NewInMemoryStore()`, `NewQdrantStore()`, `NewPostgresStore()`
  - `AutoEmbedder()`, `NewOpenAIEmbedder()` 等
- `ChunkText()` - 文本分块函数

**被引用**: 所有使用记忆系统的代码

---

#### `memory_test.go`, `example_test.go`
**目的**: 记忆系统测试和示例

---

#### 2.2.1 `engine/` - 记忆引擎

##### `engine.go` (861 行)
**主要目的**: RAG 记忆引擎核心实现

**核心功能**:
- `Engine` 结构体
  - 记忆评分、聚类、剪枝、检索
- 核心方法：
  - `Store()` - 嵌入、评分、持久化新记忆
  - `Retrieve()` - MMR 检索（最大边际相关性）
  - `Prune()` - 自动剪枝低价值记忆
- 重要性评分算法
- 聚类摘要：`EnableSummaries`

**关键依赖**:
- `store.VectorStore` - 向量存储接口
- `embed.Embedder` - 嵌入提供者
- `model.MemoryRecord` - 记忆记录结构

**被引用**: `session.SessionMemory`

---

##### `options.go`
**目的**: 引擎配置选项
- `Options` 结构体
- `DefaultOptions()` - 默认配置
- 权重配置：`ScoreWeights`

---

##### `metrics.go`
**目的**: 运行时指标
- `Metrics` 结构体
- 指标快照：`MetricsSnapshot`

---

##### `summarizer.go`
**目的**: 记忆聚类摘要
- `Summarizer` 接口
- `HeuristicSummarizer` - 启发式摘要实现

---

#### 2.2.2 `session/` - 会话记忆

##### `memory_bank.go` (212 行)
**主要目的**: 记忆库和会话记忆封装

**核心功能**:
- `MemoryBank` - 向量存储的薄封装
- `SessionMemory` - 会话记忆管理器
  - 短期记忆缓存（可配置大小）
  - 长期记忆持久化
  - `AddShortTerm()`, `FlushToLongTerm()`
  - `Retrieve()` - 检索相关记忆
- 工厂函数：
  - `NewMemoryBank()` - PostgreSQL 后端
  - `NewMemoryBankWithStore()` - 自定义存储

**关键依赖**:
- `store.VectorStore`
- `engine.Engine`
- `embed.Embedder`

---

##### `shared_session.go`
**主要目的**: 共享会话实现

**核心功能**:
- `SharedSession` 结构体
  - 多代理共享的长期记忆
  - 空间（Space）管理
- 方法：
  - `AddShortTo()` - 添加短期记忆到空间
  - `FlushSpace()` - 持久化空间记忆
  - `Spaces()` - 获取所有空间列表

**关联**: 多代理协调场景

---

##### `spaces.go`, `spaces_test.go`
**主要目的**: 空间注册表

**核心功能**:
- `SpaceRegistry` - 空间注册和管理
- `Space` - 单个空间
  - 角色权限：Reader/Writer/Admin
  - 记忆添加和检索
- 访问控制实现

---

#### 2.2.3 `store/` - 向量存储

##### `vector_store.go`
**主要目的**: 向量存储接口定义

**核心功能**:
- `VectorStore` 接口
  - `CreateCollection()`
  - `StoreMemory()`, `RetrieveMemories()`
  - `DeleteMemories()`
- `SchemaInitializer` - 模式初始化
- `GraphStore` - 图存储接口（Neo4j）
- `Distance` - 距离度量（Cosine, Euclidean, DotProduct）

---

##### `in_memory_store.go`, `in_memory_store_test.go`
**主要目的**: 内存存储实现

**核心功能**:
- `InMemoryStore` - 用于测试和开发
- 简单的余弦相似度计算
- 无持久化

---

##### `postgres_store.go`, `postgres_store_test.go`
**主要目的**: PostgreSQL + pgvector 存储

**核心功能**:
- `PostgresStore` 结构体
- pgvector 集成
- 自动模式迁移
- SQL 查询优化

**依赖**: `github.com/jackc/pgx/v5`

---

##### `qdrant_store.go`
**主要目的**: Qdrant 向量数据库适配器

**核心功能**:
- `QdrantStore` 结构体
- gRPC 客户端连接
- 批量操作支持
- 768 维向量优化（Gemini 嵌入）

---

##### `neo4j_store.go`, `neo4j_store_test.go`, `neo4j_driver_adapter.go`
**主要目的**: Neo4j 图数据库存储

**核心功能**:
- `Neo4jStore` - 图记忆存储
- 记忆关系：Follows/Explains/Contradicts/DerivedFrom
- Cypher 查询优化

**依赖**: `github.com/neo4j/neo4j-go-driver/v5`

---

##### `mongodb_store.go`, `mongodb_store_test.go`
**主要目的**: MongoDB 存储

**核心功能**:
- `MongoStore` 结构体
- MongoDB 向量搜索
- 索引管理

**依赖**: `go.mongodb.org/mongo-driver`

---

#### 2.2.4 `embed/` - 嵌入提供者

##### `embed.go`
**主要目的**: 嵌入接口和自动选择

**核心功能**:
- `Embedder` 接口
  - `Embed()` - 生成向量
  - `Dimension()` - 向量维度
- `AutoEmbedder()` - 自动选择提供者
  - 环境变量检测
  - 回退到 DummyEmbedder

---

##### `openai.go`
**目的**: OpenAI 嵌入实现
- text-embedding-3-small/large
- 可配置维度

##### `ollama.go`
**目的**: Ollama 本地嵌入
- 支持本地模型
- 无 API 密钥需求

##### `vertex.go`
**目的**: Google Vertex AI 嵌入
- 企业级部署
- 自定义维度

##### `claude.go`
**目的**: Anthropic Claude 嵌入
- 使用 Anthropic API

##### `fast_embed_fast.go`, `fast_embed_stub.go`, `fast_embed_opts.go`
**目的**: fastembed-go 封装
- 基于 ONNX 的本地嵌入
- 高性能
- 多模型支持

**依赖**: `github.com/anush008/fastembed-go`

---

#### 2.2.5 `model/` - 数据模型

**目的**: 记忆记录数据结构

**核心功能**:
- `MemoryRecord` 结构体
  - ID, SessionID, Content, Metadata
  - Embedding, Timestamp
  - Importance 评分
- `GraphEdge` - 记忆关系
- `EdgeType` - 关系类型枚举
- 元数据编码/解码函数

---

### 2.3 Models (LLM 提供商适配器)

**路径**: `src/models/`

#### `interface.go` (35 行)
**主要目的**: LLM 接口定义

**核心功能**:
- `Agent` 接口（注意：与主 agent 包不同）
  - `Generate()` - 同步生成
  - `GenerateWithFiles()` - 带文件附件
  - `GenerateStream()` - 流式响应
- `File` 结构体 - 文件附件
- `StreamChunk` 结构体 - 流式数据块

**被引用**: 所有 LLM 实现

---

#### `gemini.go` (150 行)
**主要目的**: Google Gemini 实现

**核心功能**:
- `GeminiLLM` 结构体
- 构造函数 `NewGeminiLLM()`
  - 环境变量：`GOOGLE_API_KEY` / `GEMINI_API_KEY`
- 方法：
  - `Generate()` - 同步调用
  - `GenerateStream()` - 流式迭代
  - `GenerateWithFiles()` - 多模态支持（图片/视频）

**依赖**: `github.com/google/generative-ai-go`

---

#### `anthropics.go`
**主要目的**: Anthropic Claude 实现

**核心功能**:
- `AnthropicLLM` 结构体
- 构造函数 `NewAnthropicLLM()`
  - 环境变量：`ANTHROPIC_API_KEY`
- 流式支持
- 工具调用格式

**依赖**: `github.com/anthropics/anthropic-sdk-go`

---

#### `ollama.go`
**主要目的**: Ollama 本地模型实现

**核心功能**:
- `OllamaLLM` 结构体
- 构造函数 `NewOllamaLLM()`
  - 可配置基础 URL
- 本地模型支持
- 无 API 密钥需求

**依赖**: `github.com/ollama/ollama`

---

#### `openai.go`
**主要目的**: OpenAI GPT 实现

**核心功能**:
- `OpenAILLM` 结构体
- 构造函数 `NewOpenAILLM()`
  - 环境变量：`OPENAI_API_KEY`
- 工具调用格式
- 流式支持

**依赖**: `github.com/sashabaranov/go-openai`

---

#### `cached.go`, `cached_test.go`
**主要目的**: LLM 响应缓存

**核心功能**:
- `CachedLLM` 装饰器
- LRU 缓存实现
- 缓存键生成
- TTL 过期

**优化**: 减少 API 调用成本

---

#### `dummy.go`
**主要目的**: 虚拟/测试 LLM

**核心功能**:
- `DummyLLM` 结构体
- 固定响应或回显
- 用于单元测试

---

#### `helper.go`, `helper_test.go`, `helper_bench_test.go`
**主要目的**: 辅助函数和基准测试

**核心功能**:
- MIME 类型检测
- 文件预处理
- 性能基准测试

---

#### `models_test.go`
**主要目的**: 模型集成测试

---

### 2.4 SubAgents (预构建专家代理)

**路径**: `src/subagents/`

#### `researcher.go` (50 行)
**主要目的**: 研究员代理实现

**核心功能**:
- `Researcher` 结构体
- 构造函数 `NewResearcher()`
  - 需要 `models.Agent` 作为底层模型
- 方法：
  - `Name()` - "researcher"
  - `Description()` - 功能描述
  - `Run()` - 执行研究任务
- 角色提示：结构化研究简报

**关键依赖**: `agent.SubAgent` 接口

**被引用**: ADK 示例和演示

---

#### `researcher_test.go`
**目的**: 研究员代理测试

---

### 2.5 Swarm (多代理协调)

**路径**: `src/swarm/`

#### `swarm.go`
**主要目的**: 多代理协调器

**核心功能**:
- `Swarm` 结构体
  - 代理注册和发现
  - 任务分配策略
- 协调方法：
  - 广播
  - 选择性委托
  - 结果聚合

**关键依赖**:
- `agent.SubAgent`
- `memory.SharedSession`

---

#### `swarm_test.go`
**目的**: 多代理协调测试

---

#### `participant.go`
**主要目的**: 参与者管理

**核心功能**:
- `Participant` 结构体
- 角色和权限
- 状态跟踪

---

#### `participant_test.go`
**目的**: 参与者测试

---

### 2.6 Cache (缓存工具)

**路径**: `src/cache/`

**功能**: LRU 缓存、TTL 缓存实现
**被引用**: `models/cached.go`, 工具系统

---

### 2.7 Concurrent (并发工具)

**路径**: `src/concurrent/`

**功能**:
- 工作池
- 并行处理
- 限流器

**被引用**: 高性能场景

---

### 2.8 Helpers (辅助函数)

**路径**: `src/helpers/`

**功能**:
- 字符串处理
- 映射操作
- 通用工具函数

---

## 🎯 第三部分：示例程序

### cmd/example/ (15+ 个示例)

#### 核心示例

##### `codemode/main.go`
**目的**: CodeMode + UTCP 工具调用演示

**展示**:
- Agent 暴露为 UTCP 工具
- `RegisterAsUTCPProvider()`
- `client.CallTool()` 直接调用
- CodeMode 启用自然语言 → 工具编排

**流程**: 用户输入 → LLM 生成 `codemode.CallTool()` → UTCP 执行

---

##### `codemode_utcp_workflow/main.go`
**目的**: 多代理工作流编排

**展示**:
- 多个专家代理（分析师、作家、评审员）
- 每个代理注册为 UTCP 工具
- CodeMode 协调多步骤工作流
- 链式调用：分析 → 写作 → 评审 → 定稿

---

##### `agent_as_tool/main.go`
**目的**: 代理间通信

**展示**:
- 层级代理架构
- Manager-Specialist 模式
- `agent.researcher` 工具命名
- 直接工具调用验证

---

##### `agent_as_utcp_codemode/main.go`
**目的**: Agent 作为 UTCP 工具 + CodeMode

**展示**:
- 代理暴露为 UTCP 工具
- CodeMode orchestration
- 自然语言 → 工具调用生成

---

##### `checkpoint/main.go`
**目的**: Agent 状态持久化

**展示**:
- `agent.Checkpoint()` - 序列化到 `[]byte`
- `agent.Restore()` - 从检查点恢复
- 保留对话历史和共享空间成员
- 跨进程重启恢复

---

##### `composability/main.go`
**目的**: Agent 可组合性

**展示**:
- Agent 作为一等工具
- 递归能力：子代理可以有子代理
- 上下文隔离

---

##### `multi_agent_memory/main.go`
**目的**: 多代理共享记忆

**展示**:
- `memory.SharedSession`
- 共享空间（Space）
- 跨代理上下文共享

---

##### `shared_session_test/main.go`
**目的**: 共享会话测试

---

##### `context/main.go`
**目的**: 上下文管理

---

##### `guardrails/main.go`
**目的**: 输出护栏

**展示**:
- `OutputGuardrails`
- 安全策略验证
- 格式强制

---

##### `hybrid_memory_test/main.go`
**目的**: 混合记忆测试

---

##### `autonomous_agent/main.go`
**目的**: 自主代理

---

##### `autonomous_cron/main.go`
**目的**: 定时任务代理

---

##### `5ways_demo/main.go`
**目的**: 五种使用方式演示

---

##### `main.go`
**目的**: 示例入口和索引

---

### cmd/demo/
**目的**: 交互式 CLI 演示

**功能**:
- 完整功能演示
- 工具调用
- 子代理委托
- 记忆检索

---

### cmd/team/
**目的**: 多代理协调演示

---

### cmd/app/
**目的**: 应用程序示例

---

### cmd/validator/
**目的**: 验证器工具

---

### cmd/orchestrate/
**目的**: 编排器示例

---

## 🔗 第四部分：依赖关系图

### 核心依赖层级

```
Level 0 (基础类型):
  - types.go
  - src/models/interface.go
  - src/memory/model/*.go
  - src/memory/store/vector_store.go
  - src/adk/module.go
  - src/adk/providers.go

Level 1 (核心实现):
  - agent.go (依赖 Level 0)
  - catalog.go (依赖 types.go)
  - src/memory/memory.go (类型导出)
  - src/adk/kit.go (依赖 providers.go)

Level 2 (扩展功能):
  - agent_tool.go (依赖 agent.go)
  - agent_stream.go (依赖 agent.go)
  - agent_orchestrators.go (依赖 agent.go)
  - src/adk/modules/*.go (依赖 kit.go)

Level 3 (具体实现):
  - src/memory/engine/engine.go
  - src/memory/session/*.go
  - src/memory/store/*.go (各存储后端)
  - src/models/*.go (各 LLM 提供商)
  - src/subagents/*.go
  - src/swarm/*.go

Level 4 (应用层):
  - cmd/example/*
  - cmd/demo/*
  - 测试文件
```

### 模块间依赖

```
agent.go
├── src/memory.SessionMemory
│   ├── src/memory/engine.Engine
│   │   ├── src/memory/store.VectorStore
│   │   └── src/memory/embed.Embedder
│   └── src/memory/session.*
├── src/models.Agent
│   ├── GeminiLLM
│   ├── AnthropicLLM
│   ├── OllamaLLM
│   └── OpenAILLM
├── ToolCatalog
│   └── StaticToolCatalog (catalog.go)
├── SubAgentDirectory
│   └── StaticSubAgentDirectory (catalog.go)
└── utcp.UtcpClientInterface

src/adk/kit.go
├── Module (module.go)
├── ModelProvider / MemoryProvider / ToolProvider / SubAgentProvider (providers.go)
├── agent.Agent
└── src/adk/modules.*
```

---

## 📊 第五部分：代码统计

### 按目录统计

| 目录 | 文件数 | 估算行数 | 主要功能 |
|------|--------|----------|----------|
| 根目录 | 11 | ~3000 | Agent 核心运行时 |
| src/adk/ | 5 + 5 | ~700 | ADK 框架 |
| src/memory/ | 20+ | ~2000 | 记忆系统 |
| src/models/ | 10 | ~800 | LLM 适配器 |
| src/subagents/ | 2 | ~100 | 预构建代理 |
| src/swarm/ | 4 | ~300 | 多代理协调 |
| cmd/example/ | 15+ | ~2000 | 示例代码 |
| 测试文件 | 20+ | ~1500 | 单元测试 |

**总计**: 约 10,000+ 行 Go 代码

---

## 🚀 第六部分：使用指南

### 逐步引入代码的顺序

#### 阶段 1: 理解核心接口 (1-2 小时)
1. `types.go` - 工具和子代理接口
2. `src/models/interface.go` - LLM 接口
3. `catalog.go` - 注册表实现

#### 阶段 2: Agent 核心 (3-4 小时)
4. `agent.go` (前 200 行) - 结构体和构造函数
5. `agent.go` (工具调用部分) - `CallTool()`, `executeTool()`
6. `agent_tool.go` - Agent 作为工具

#### 阶段 3: 记忆系统 (3-4 小时)
7. `src/memory/memory.go` - 公共 API
8. `src/memory/session/memory_bank.go` - 会话记忆
9. `src/memory/engine/engine.go` (前 200 行) - RAG 引擎

#### 阶段 4: ADK 框架 (2-3 小时)
10. `src/adk/kit.go` (前 200 行) - ADK 核心
11. `src/adk/module.go`, `providers.go` - 模块系统
12. `src/adk/modules/*.go` - 具体模块

#### 阶段 5: 高级功能 (4-6 小时)
13. `agent_stream.go` - 流式接口
14. `agent_orchestrators.go` - 编排引擎
15. `src/swarm/*.go` - 多代理协调

#### 阶段 6: 示例学习 (2-3 小时)
16. `cmd/example/codemode/main.go` - CodeMode
17. `cmd/example/agent_as_tool/main.go` - 代理作为工具
18. `cmd/example/checkpoint/main.go` - 状态持久化

---

## 📝 第七部分：关键设计模式

### 1. 依赖注入 (ADK)
```go
kit, _ := adk.New(ctx,
    adk.WithModules(
        adkmodules.NewModelModule(...),
        adkmodules.InQdrantMemory(...),
    ),
)
```

### 2. 适配器模式
- `AgentToolAdapter` - Agent → Tool
- `SubAgentTool` - SubAgent → Tool
- `CachedLLM` - LLM 装饰器

### 3. 策略模式
- `SafetyPolicy` - 多种安全策略
- `VectorStore` - 多种存储后端
- `Embedder` - 多种嵌入提供者

### 4. 观察者模式
- `SharedSession` - 多代理共享状态
- `SpaceRegistry` - 空间事件

---

## 🔍 第八部分：调试和分析建议

### 使用 AI 分析时的引入顺序

1. **先引入类型定义**
   ```
   请分析 types.go 中的接口定义
   ```

2. **再引入核心实现**
   ```
   现在结合 agent.go 分析 Agent 如何调用工具
   ```

3. **逐步添加依赖**
   ```
   请查看 catalog.go 了解工具注册表实现
   ```

4. **最后分析示例**
   ```
   参考 cmd/example/codemode/main.go 看实际用法
   ```

### 常见分析场景

| 场景 | 需要读取的文件 |
|------|---------------|
| 工具调用流程 | `agent.go` → `catalog.go` → `agent_tool.go` |
| 记忆检索 | `agent.go` → `session/memory_bank.go` → `engine/engine.go` |
| LLM 集成 | `models/interface.go` → `models/gemini.go` |
| ADK 使用 | `adk/kit.go` → `adk/modules/*.go` |
| 多代理 | `swarm/swarm.go` → `session/shared_session.go` |

---

## 📌 总结

go-agent (Lattice) 是一个结构清晰、模块化设计的 AI Agent 框架：

**核心优势**:
- ✅ 清晰的接口分离
- ✅ 可插拔的模块系统
- ✅ 多种 LLM 和存储后端支持
- ✅ 完整的测试覆盖
- ✅ 丰富的示例代码

**学习曲线**:
- 基础使用：2-4 小时
- 深入理解：1-2 天
- 贡献代码：1 周

**推荐学习路径**:
1. 从 `cmd/example` 的简单示例开始
2. 阅读 `types.go` 了解接口
3. 深入 `agent.go` 理解核心逻辑
4. 探索 `src/memory` 和 `src/adk` 的高级功能

---

*文档生成完成。使用此清单逐步引入代码到 AI 分析会话中。*
