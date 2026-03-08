# go-agent 扩展 ToDoList

**项目**: Lattice - Go AI Agent 开发框架  
**创建日期**: 2026 年 3 月 6 日  
**最后更新**: 2026 年 3 月 6 日  
**维护**: 项目基础设施团队

---

## 📊 状态说明

| 状态 | 说明 |
|------|------|
| 📋 `pending` | 待处理，未开始 |
| 🔍 `researching` | 调研中，收集资料 |
| 📝 `designing` | 设计中，编写方案 |
| 💻 `implementing` | 开发中，编写代码 |
| 🧪 `testing` | 测试中，编写测试 |
| 📖 `documenting` | 文档中，补充说明 |
| ✅ `done` | 已完成，待验收 |
| 🚫 `blocked` | 已阻塞，等待外部依赖 |
| ❌ `cancelled` | 已取消，不再进行 |

---

## 🎯 优先级定义

| 优先级 | 说明 | 响应时间 |
|--------|------|----------|
| `P0` | 紧急且重要，阻塞核心功能 | 立即处理 |
| `P1` | 重要不紧急，影响用户体验 | 1 周内 |
| `P2` | 重要但可延后，功能增强 | 1 月内 |
| `P3` | 锦上添花，有空再做 | 季度内 |

---

## 📋 任务清单

---

### [EXT-017] 端到端流式传输支持

**状态**: 📋 `pending`
**优先级**: `P1`
**创建日期**: 2026-03-07
**负责人**: (待分配)
**预计工作量**: 8 小时

#### 描述
实现从 LLM 到前端的端到端流式传输，让用户实时看到 AI 响应逐字生成，提升用户体验。支持 SSE（Server-Sent Events）、WebSocket 和 HTTP Chunked 三种传输方式。

#### 核心功能

**1. SSE 流式传输（推荐）**
- 原生 HTTP 支持，无需额外库
- 自动重连机制
- 单向推送（服务器→客户端）
- 轻量级，适合对话场景

**2. WebSocket 双向通信**
- 支持客户端→服务器实时交互
- 适用于中途打断、修改参数等场景
- 全双工通信

**3. HTTP Chunked Transfer**
- 简单的流式需求
- 无需额外协议
- NDJSON 格式

**4. 前端流式界面**
- 打字机效果（逐字显示）
- 实时滚动到底部
- 连接状态指示
- 错误处理与重试

#### 实施内容

**后端实现**：
- [ ] **SSE Handler** (`handlers/stream_handler.go`)
  - [ ] SSE 响应头设置
  - [ ] StreamMessage 结构定义
  - [ ] StreamChat 主处理器
  - [ ] 错误处理与发送
- [ ] **WebSocket Handler** (`handlers/ws_handler.go`)
  - [ ] WebSocket Upgrader 配置
  - [ ] WSMessage 结构定义
  - [ ] WSChat 主处理器
  - [ ] 连接管理
- [ ] **路由集成**
  - [ ] `/api/stream` - SSE 端点
  - [ ] `/api/ws` - WebSocket 端点
  - [ ] CORS 配置

**前端实现**：
- [ ] **SSE 客户端** (`static/index.html`)
  - [ ] EventSource 连接
  - [ ] 消息处理与显示
  - [ ] 打字机光标效果
  - [ ] 自动滚动
  - [ ] 错误处理
- [ ] **WebSocket 客户端** (可选)
  - [ ] WebSocket 连接管理
  - [ ] 消息发送与接收
  - [ ] 重连机制

**性能优化**：
- [ ] **缓冲优化**
  - [ ] 通道缓冲区调整（默认 16→32）
  - [ ] 批量发送（每 5 个 token 发送）
- [ ] **超时控制**
  - [ ] 请求超时（60 秒）
  - [ ] 连接空闲超时
- [ ] **内存管理**
  - [ ] strings.Builder 预分配
  - [ ] 及时释放资源

**安全与稳定性**：
- [ ] **CORS 配置**
  - [ ] 允许的来源列表
  - [ ] 生产环境限制
- [ ] **认证机制**
  - [ ] Token 验证
  - [ ] 会话管理
- [ ] **错误处理**
  - [ ] 连接断开处理
  - [ ] 重试机制
  - [ ] 降级策略（流式失败转普通）

#### 技术架构

```
LLM Provider (Gemini/OpenAI)
    │
    │ <-chan StreamChunk
    ▼
Agent.GenerateStream()
    │
    │ SSE/WebSocket
    ▼
HTTP Server
    │
    │ text/event-stream
    ▼
前端 EventSource
    │
    │ 逐字显示
    ▼
用户看到实时输出
```

#### 代码结构

```
go-agent/
├── handlers/
│   ├── stream_handler.go      # SSE 处理器 ⭐
│   └── ws_handler.go          # WebSocket 处理器
├── static/
│   ├── index.html             # 流式对话界面 ⭐
│   └── js/
│       └── stream.js          # 流式客户端逻辑
└── docs/
    └── STREAMING_IMPLEMENTATION.md  # 本文档 ⭐
```

#### API 设计

**SSE 端点**：
```
GET /api/stream?session_id=xxx&message=yyy

响应格式（text/event-stream）：
data: {"type":"token","content":"你"}
data: {"type":"token","content":"好"}
data: {"type":"done","fullText":"你好，有什么可以帮助你的？"}
```

**WebSocket 端点**：
```
WS /api/ws

请求消息：
{
  "type": "message",
  "session_id": "session1",
  "message": "你好"
}

响应消息：
{
  "type": "token",
  "content": "你"
}
```

#### 验收标准

- [ ] **功能验收**
  - [ ] SSE 流式传输正常工作
  - [ ] 前端实时显示逐字输出
  - [ ] 错误处理完善
  - [ ] 支持并发连接（10+ 同时连接）
- [ ] **性能验收**
  - [ ] 首字延迟 < 500ms
  - [ ] 传输延迟 < 100ms/token
  - [ ] 内存占用 < 50MB/连接
- [ ] **兼容性验收**
  - [ ] 主流浏览器支持（Chrome/Firefox/Safari/Edge）
  - [ ] 移动端支持

#### 备注
- **高优先级**：显著提升用户体验的核心功能
- 预计 8 小时完成（约 1 个工作日）
- 可独立部署，不影响现有功能
- 建议作为 MareMind 产品的标准功能

---

### [EXT-016] 智能记忆与因果推理系统

**状态**: 📝 `designing`
**优先级**: `P0`
**创建日期**: 2026-03-07
**更新日期**: 2026-03-08
**负责人**: (待分配)
**预计工作量**: 分三阶段，总计 24-31 周（约 500+ 小时）
**设计文档**: [SMART_MEMORY_MODULE_DESIGN.md](./SMART_MEMORY_MODULE_DESIGN.md)

#### 描述（v2.0 - Module 架构）
构建集**自组织记忆、因果推理、软性因素建模**于一体的智能 Agent 记忆系统。基于 Lattice（go-agent）框架，采用 ADK Module 模块化架构，分三个子模块实现。

#### Module 架构

**三个子模块**：
1. **MemCellModule** - 记忆单元模块
   - MemCell（记忆单元）- 类型化记忆（情景/事实/偏好）
   - MemScene（记忆场景）- 语义聚类的记忆组
   - 情景痕迹形成 - LLM 从对话中提取 MemCell
   - 语义整合 - 后台聚类生成 MemScene

2. **CausalModule** - 因果推理模块
   - 因果节点与边 - 事实/事件/规则/结果节点
   - 因果挖掘 - LLM 从 MemCell 中提取候选因果关系
   - 因果推理引擎 - 原因追溯、影响预测、反事实估计
   - DoWhy 集成 - 因果效应估计

3. **SoftFactorModule** - 软性因素模块
   - 企业文化节点 - 客户第一、创新驱动等
   - 哲学原则节点 - 系统论思维、辩证法等
   - 个人特质挖掘 - 拖延指数、风险偏好等
   - 软性因素调节 - 影响因果边强度

#### 实施阶段

**阶段一：MemCellModule**（8-9 周，约 120 小时）
- [ ] **M1.1**：MemCell/MemScene 定义及基础存储（1 周）
- [ ] **M1.2**：情景痕迹形成模块集成 LLM 提取（2 周）
- [ ] **M1.3**：语义整合后台任务（聚类 + 摘要）（2 周）
- [ ] **M1.4**：重构式检索与 UTCP 工具封装（1 周）
- [ ] **M1.5**：端到端测试与验证（1 周）
- **交付物**：
  - `src/memory/memcell/memcell_module.go` - MemCellModule ⭐
  - `src/memory/memcell/extractor.go` - 情景提取器
  - `src/memory/memcell/consolidator.go` - 语义整合器
  - `docs/SMART_MEMORY_MODULE_DESIGN.md` - 设计文档 ✅

**阶段二：CausalModule**（10-12 周，约 200 小时）
- [ ] **M2.1**：因果节点/边定义及存储（1 周）
- [ ] **M2.2**：因果挖掘模块集成 LLM 提取（2 周）
- [ ] **M2.3**：因果推理引擎基础功能（2 周）
- [ ] **M2.4**：集成 DoWhy 进行效应估计（2 周）
- [ ] **M2.5**：UTCP 工具封装与 Agent 集成（1 周）
- [ ] **M2.6**：因果图谱可视化与调试工具（2 周）
- **交付物**：
  - `src/memory/causal/causal_module.go` - CausalModule ⭐
  - `src/memory/causal/miner.go` - 因果挖掘器
  - `src/memory/causal/reasoner.go` - 因果推理引擎
  - `services/dowhy/` - DoWhy Python 微服务

**阶段三：SoftFactorModule**（8-10 周，约 180 小时）
- [ ] **M3.1**：软性因素节点定义与存储（1 周）
- [ ] **M3.2**：个人特质挖掘模块（2 周）
- [ ] **M3.3**：企业文化与哲学原则人工注入工具（1 周）
- [ ] **M3.4**：因果推理引擎整合软性因素调节（2 周）
- [ ] **M3.5**：端到端测试与可解释性增强（2 周）
- **交付物**：
  - `src/memory/causal/soft_factor_module.go` - SoftFactorModule ⭐
  - `configs/culture_nodes.json` - 企业文化预定义
  - `configs/philosophy_nodes.json` - 哲学原则预定义

#### 总体架构

**三层架构**：
- **记忆层** - 实现 EverMemOS 风格的 MemCell 和 MemScene，提供情景化记忆的存储与检索
- **因果层** - 在记忆层基础上构建因果图谱，支持因果挖掘与推理
- **软性因素层** - 将企业文化、哲学原则、个人特质建模为因果节点，影响推理结果

#### 核心功能

**1. EverMemOS 风格记忆系统**
- MemCell（记忆单元）- 类型化记忆（情景/事实/偏好）
- MemScene（记忆场景）- 语义聚类的记忆组
- 情景痕迹形成 - LLM 从对话中提取 MemCell
- 语义整合 - 后台聚类生成 MemScene
- 重构式回忆 - 场景引导的检索，提供必要且充分的上下文

**2. REMI 因果记忆框架**
- 因果节点与边 - 事实/事件/规则/结果节点
- 因果挖掘 - LLM 从 MemCell 中提取候选因果关系
- 因果推理引擎 - 原因追溯、影响预测、反事实估计
- DoWhy 集成 - 因果效应估计
- 可解释因果建议 - Agent 回答"为什么"和"如果...会怎样"

**3. 软性因素建模**
- 企业文化节点 - 客户第一、创新驱动等
- 哲学原则节点 - 系统论思维、辩证法等
- 个人特质挖掘 - 拖延指数、风险偏好等
- 软性因素调节 - 影响因果边强度
- 与 SOP 规则协同 - 软约束与硬规则共同作用

#### 实施阶段

**阶段一：EverMemOS 记忆系统**（8-9 周，约 120 小时）
- [ ] **M1.1**：MemCell/MemScene 定义及基础存储（1 周）
- [ ] **M1.2**：情景痕迹形成模块集成 LLM 提取（2 周）
- [ ] **M1.3**：语义整合后台任务（聚类 + 摘要）（2 周）
- [ ] **M1.4**：重构式检索与 UTCP 工具封装（1 周）
- [ ] **M1.5**：端到端测试与验证（1 周）
- **交付物**：
  - `src/memory/memcell/` - MemCell 模块
  - `src/memory/memscene/` - MemScene 模块
  - `src/memory/memcell/extractor.go` - 情景提取器
  - `src/memory/memcell/consolidator.go` - 语义整合器
  - `src/memory/memcell/retriever.go` - 重构检索器
  - UTCP 工具：`search_memory`

**阶段二：REMI 因果记忆框架**（10-12 周，约 200 小时）
- [ ] **M2.1**：因果节点/边定义及存储（1 周）
- [ ] **M2.2**：因果挖掘模块集成 LLM 提取（2 周）
- [ ] **M2.3**：因果推理引擎基础功能（2 周）
- [ ] **M2.4**：集成 DoWhy 进行效应估计（2 周）
- [ ] **M2.5**：UTCP 工具封装与 Agent 集成（1 周）
- [ ] **M2.6**：因果图谱可视化与调试工具（2 周）
- **交付物**：
  - `src/memory/causal/` - 因果模块
  - `src/memory/causal/miner.go` - 因果挖掘器
  - `src/memory/causal/reasoner.go` - 因果推理引擎
  - `services/dowhy/` - DoWhy Python 微服务
  - UTCP 工具：`causal_find_causes`, `causal_find_effects`, `causal_estimate_effect`

**阶段三：软性因素建模**（8-10 周，约 180 小时）
- [ ] **M3.1**：软性因素节点定义与存储（1 周）
- [ ] **M3.2**：个人特质挖掘模块（2 周）
- [ ] **M3.3**：企业文化与哲学原则人工注入工具（1 周）
- [ ] **M3.4**：因果推理引擎整合软性因素调节（2 周）
- [ ] **M3.5**：端到端测试与可解释性增强（2 周）
- **交付物**：
  - `src/memory/causal/soft_factor.go` - 软性因素模块
  - `src/memory/causal/personality.go` - 个人特质挖掘
  - `configs/culture_nodes.json` - 企业文化预定义
  - `configs/philosophy_nodes.json` - 哲学原则预定义

#### 技术选型

| 组件 | 技术选型 | 备注 |
|------|----------|------|
| Agent 框架 | Lattice (go-agent) | 提供记忆、工具、多 Agent 协调 |
| 向量存储 | Lattice memory.Engine + PostgreSQL/pgvector | 支持向量检索与元数据过滤 |
| 图存储 | PostgreSQL (初期) / Neo4j (后期) | 因果图谱存储 |
| 嵌入模型 | DeepSeek Embedding | 统一向量维度 |
| LLM 服务 | DeepSeek API / 本地部署 | 用于提取、摘要、规划等 |
| 因果估计 | DoWhy (Python) | 封装为微服务供 Go 调用 |
| 聚类算法 | Go 实现 KMeans | 初期简单实现 |
| 任务调度 | 内置 cron | 用于语义整合等后台任务 |

#### 与 go-agent 集成

**直接利用的现有能力**：
- ✅ `memory.Engine` - 记忆存储与检索
- ✅ `memory.Embedder` - 向量嵌入
- ✅ `UTCP` - 工具封装
- ✅ `SharedSession` - 多 Agent 共享记忆

**需要扩展的部分**：
- ⚠️ 类型系统 - 在 Metadata 中增加类型识别
- ⚠️ 后台任务 - 新增语义整合器
- ⚠️ 因果图谱 - 全新功能模块
- ⚠️ DoWhy 集成 - Go-Python 跨语言调用

#### 验证指标

**阶段一**：
- 记忆检索准确率 > 85%（对比人工标注）
- 场景聚类主题一致性 > 80%（人工评估）

**阶段二**：
- 因果挖掘准确率 > 75%（对比专家标注）
- 推理结果合理性 > 80%（专家评估）

**阶段三**：
- 软性因素对推理的影响符合预期（专家评估）
- 用户对建议的满意度 > 85%

#### 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| LLM 提取准确性低 | 高 | 人工审核 + 主动学习 + 多模型投票 |
| 因果挖掘产生虚假因果 | 高 | 置信度阈值 + 专家审核 + 时间序列验证 |
| DoWhy 集成复杂度高 | 中 | 封装为独立微服务，定义清晰 API |
| 软性因素建模主观性强 | 中 | 专家输入 + 用户反馈调整 |
| 性能问题（大规模图谱） | 中 | 图索引优化 + 缓存策略 |

#### 备注
- **最高优先级**：这是 MareMind 项目的核心竞争力，决定 Agent 的智能水平
- 预计 24-31 周完成（约 6-8 个月）
- 必须分阶段发布，每阶段都有独立价值
- 需要领域专家参与（企业文化、哲学原则定义）
- 需要大量标注数据用于调优 LLM 提取

---

### [EXT-015] 可确认 Agent 扩展 - 异步确认与授权管理

**状态**: 📝 `designing`
**优先级**: `P1`
**创建日期**: 2026-03-07
**更新日期**: 2026-03-08
**负责人**: (待分配)
**预计工作量**: 16 小时
**设计文档**: [CONFIRMABLE_AGENT_DESIGN_V2.md](./CONFIRMABLE_AGENT_DESIGN_V2.md)

#### 描述（v2.0 - 装饰器模式）
实现基于**装饰器模式**的 Agent 确认机制，支持动态确认需求提取、异步确认、永久授权和撤销功能。通过 Tool 装饰器自动拦截需要确认的操作，在便利性和控制权之间取得平衡。

**核心改进**：
- ✅ **零侵入** - 不修改 Tool 接口或 ToolRequest
- ✅ **动态提取** - 分析用户提示词，动态判断是否需要确认
- ✅ **可组合** - 可叠加多个装饰器（确认/日志/缓存等）
- ✅ **模块化** - 通过 ADK Module 自动配置

#### 架构设计

**三层架构**：
1. **Layer 1: 用户提示词分析** - LLM 理解意图，生成 Tool 调用
2. **Layer 2: Guardrails 动态检查** - 检查用户提示词和预定义规则
3. **Layer 3: ConfirmableToolWrapper** - 装饰器执行确认逻辑

**装饰器模式**：
```
FileDeleteTool（原始工具）
    │
    └─ ConfirmableToolWrapper（确认装饰器）
        ├─ 额外功能：等待用户确认
        └─ 委托调用：FileDeleteTool.Invoke()
```

#### 核心功能

**1. 装饰器模式实现**
- `ConfirmableToolWrapper` - 确认装饰器
- `LoggingToolWrapper` - 日志装饰器（可选）
- `CachingToolWrapper` - 缓存装饰器（可选）

**2. 动态确认需求提取**
- 分析用户提示词中的确认要求（"先问我"、"需要确认"）
- 检查预定义规则（金额>1000、操作类型等）
- 无需在 System Prompt 中硬编码规则

**3. 三个确认选项**
- ✅ **本次可以** (ApproveOnce) - 仅本次执行有效
- ✅ **未来均可以** (ApprovePermanent) - 创建永久授权
- ❌ **拒绝** (Deny) - 拒绝执行

**4. 异步通知机制**
- Web 通知（SSE/WebSocket）
- Email 通知（SMTP）
- Slack/Teams 通知（Webhook）

**5. 授权管理**
- 永久授权存储和检查
- 一键撤销授权
- 会话级/时效性授权

#### 实施内容

**核心库** (`confirmable/`)：
- [ ] `confirmable_tool.go` - ConfirmableToolWrapper 装饰器 ⭐
- [ ] `guardrails.go` - ConfirmationGuardrails 动态检查 ⭐
- [ ] `confirmation_handler.go` - ConfirmationHandler
- [ ] `authorization_store.go` - AuthorizationStore
- [ ] `notify/` - 通知渠道
  - [ ] `web.go` - Web 通知
  - [ ] `email.go` - Email 通知
  - [ ] `slack.go` - Slack 通知
  - [ ] `teams.go` - Teams 通知

**ADK Module** (`src/adk/modules/`)：
- [ ] `confirmable_module.go` - ConfirmationModule ⭐

**示例与文档**：
- [ ] `examples/confirmable_demo/` - 使用示例
- [ ] `docs/CONFIRMABLE_AGENT_DESIGN_V2.md` - 设计文档 ✅
- [ ] `docs/DECORATOR_PATTERN_GUIDE.md` - 装饰器模式指南

#### 使用示例

```go
// 1. 定义确认规则
rules := []*confirmable.ConfirmationRule{
    {ToolName: "file.delete", Description: "文件删除"},
    {
        ToolName: "payment.process",
        Description: "大额支付需要确认",
        Condition: func(args map[string]any) bool {
            return args["amount"].(float64) > 1000
        },
    },
}

// 2. 创建确认处理器
handler := confirmable.NewConfirmationHandler(
    confirmable.WithWebNotification(),
    confirmable.WithEmailNotification("smtp.example.com"),
)

// 3. 创建 ADK，注册确认模块
kit, _ := adk.New(ctx,
    adk.WithModule(modules.NewConfirmationModule(
        "confirmation",
        rules,
        handler,
    )),
)

// 4. 构建 Agent（工具已自动包装）
agent, _ := kit.BuildAgent(ctx)

// 5. 使用
// 用户："帮我删除这个文件，但需要先问我确认"
resp, _ := agent.Generate(ctx, "session1", "帮我删除这个文件，但需要先问我确认")
```

#### 验收标准

- [ ] **功能验收**
  - [ ] 装饰器模式正确实现（零侵入）
  - [ ] 动态确认需求提取正常工作
  - [ ] 三个确认选项正常工作
  - [ ] 异步通知发送正常
  - [ ] 授权管理（创建/检查/撤销）正常
  - [ ] 支持并发确认请求（10+ 同时）
- [ ] **性能验收**
  - [ ] 确认检查延迟 < 10ms
  - [ ] 通知发送延迟 < 1s
  - [ ] 内存占用 < 10MB/装饰器
- [ ] **兼容性验收**
  - [ ] 现有工具无需修改
  - [ ] 支持所有 Tool 实现
  - [ ] 可与其他装饰器组合

#### 备注
- **高优先级**：显著提升用户体验和安全性
- 预计 16 小时完成（约 2 个工作日）
- 装饰器模式是 Go 标准实践
- 需要编写详细的装饰器模式指南
    ├── web_notifier.go   # Web 通知（WebSocket/SSE）
    ├── email_notifier.go # Email 通知
    ├── slack_notifier.go # Slack 通知
    └── teams_notifier.go # Teams 通知
```

#### 完整流程

```
第一次操作：
用户请求 → Layer 1/2/3 检查 → 无授权 → 发送确认通知（三个选项）
→ 用户选择"未来均可以" → 创建永久授权 → 执行工具

后续操作：
用户请求 → Layer 1/2/3 检查 → 有永久授权 → 自动执行
→ 发送异步通知（含"撤销预授权"按钮）

用户撤销：
点击"撤销预授权" → 撤销授权记录 → 发送确认通知
→ 下次操作恢复确认流程
```

#### 需求
- [ ] **阶段 1：核心框架**（4 小时）
  - [ ] 创建 `confirmable/` 目录结构
  - [ ] 实现 `types.go` - 数据类型定义
  - [ ] 实现 `rule.go` - 确认规则
  - [ ] 实现 `approval.go` - 授权管理器
- [ ] **阶段 2：任务管理**（4 小时）
  - [ ] 实现 `task_manager.go` - TaskManager
  - [ ] 实现 `storage.go` - 存储接口
  - [ ] 集成 Checkpoint/Restore
  - [ ] 实现 Stop/Cancel/Resume
- [ ] **阶段 3：Tool 包装器**（4 小时）
  - [ ] 实现 `confirmable_tool.go` - ConfirmableTool
  - [ ] 实现三层检查逻辑
  - [ ] 实现三个确认选项处理
  - [ ] 实现异步通知发送
- [ ] **阶段 4：通知系统**（2 小时）
  - [ ] 实现 `notifier.go` - 通知接口
  - [ ] 实现 `web_notifier.go` - Web 通知
  - [ ] 实现 `email_notifier.go` - Email 通知
  - [ ] 实现 `slack_notifier.go` - Slack 通知
- [ ] **阶段 5：Guardrails**（1 小时）
  - [ ] 实现 `guardrails.go` - Output Guardrails
  - [ ] 敏感操作检测
  - [ ] 授权语句提取
- [ ] **阶段 6：审计日志**（1 小时）
  - [ ] 实现 `audit.go` - 审计日志接口
  - [ ] 实现文件/数据库存储
- [ ] **阶段 7：测试和文档**（2 小时）
  - [ ] 编写单元测试
  - [ ] 编写集成测试
  - [ ] 编写使用文档
  - [ ] 编写示例代码

#### 交付物
- [ ] `confirmable/` - 完整实现模块
- [ ] `docs/CONFIRMABLE_AGENT_IMPLEMENTATION.md` - 详细实施文档 ✅
- [ ] `examples/confirmable_demo/` - 完整示例
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试用例

#### 技术亮点
- ✅ **三层确认** - 不依赖单一机制，可靠强制执行
- ✅ **三个选项** - 灵活授权，平衡便利与控制
- ✅ **异步通知** - 不打扰用户，保持透明
- ✅ **一键撤销** - 随时收回授权，保持控制
- ✅ **Checkpoint 集成** - 超时自动保存，支持恢复
- ✅ **完整审计** - 所有操作可追溯，支持行为分析
- ✅ **多渠道支持** - Web/Email/Slack 等统一接口

#### 与现有功能的关系
- **EXT-014**: 修复 executeTool 本地工具调用 → 本方案的基础
- **Checkpoint/Restore**: 已有的状态持久化 → 本方案用于超时保存
- **Guardrails**: 已有的输出护栏 → 扩展支持确认检查

#### 备注
- **高优先级**：提升 Agent 可控性和安全性的关键功能
- 预计 16 小时完成（约 2-3 个工作日）
- 可以分阶段发布（阶段 1-3 为核心功能）
- 需要充分的测试确保状态管理正确

---

### [EXT-014] 修复 executeTool() 本地工具调用缺陷

**状态**: 📋 `pending`
**优先级**: `P1`
**创建日期**: 2026-03-07
**负责人**: (待分配)
**预计工作量**: 4 小时

#### 描述
修复 `executeTool()` 函数无法调用本地工具的架构缺陷。当前实现只支持 UTCP 远程工具调用，导致通过 `toolCatalog` 注册的本地工具（如 `EchoTool`, `CalculatorTool` 等）无法通过 LLM 驱动的自动流程调用，只能通过手动调用 `tool.Invoke()` 使用。

#### 问题详情

**当前行为**：
```go
func (a *Agent) executeTool(...) (any, error) {
    // 1. REMOTE UTCP TOOL
    if a.UTCPClient != nil {
        return a.UTCPClient.CallTool(ctx, toolName, args)
    }
    
    // 2. Unknown tool ← 本地工具到这里返回错误
    return nil, fmt.Errorf("unknown tool: %s", toolName)
}
```

**影响范围**：
- ❌ `detectDirectToolCall` → `executeTool` 路径：本地工具调用失败
- ❌ `toolOrchestrator` → `executeTool` 路径：本地工具调用失败
- ✅ 手动调用 `tool.Invoke()`：本地工具可用（但失去 LLM 驱动能力）
- ✅ `AgentToolAdapter`：特殊处理，绕过 `executeTool`（仅限 Agent 作为工具）

**代码分析**：
1. `ToolSpecs()` 返回本地 + UTCP 工具混合列表 ✅
2. `detectDirectToolCall` 会匹配本地工具名 ✅
3. `toolOrchestrator` 验证工具存在时通过本地工具 ✅
4. `executeTool` 执行时只处理 UTCP 工具 ❌

#### 需求
- [ ] 修改 `executeTool()` 函数，添加本地工具调用逻辑
- [ ] 实现顺序：先本地工具 → 再 UTCP 工具 → 最后未知工具错误
- [ ] 更新代码注释，移除误导性说明（"Execute UTCP or local tool"）
- [ ] 编写单元测试覆盖本地工具调用场景
- [ ] 编写集成测试验证 LLM 驱动的工具选择
- [ ] 性能基准测试（确保不引入性能退化）
- [ ] 向后兼容性测试（确保不影响 UTCP 工具）

#### 修复方案

```go
func (a *Agent) executeTool(
    ctx context.Context,
    sessionID, toolName string,
    args map[string]any,
) (any, error) {

    if args == nil {
        args = map[string]any{}
    }

    // 1️⃣ 先尝试本地工具（新增）
    if tool, spec, ok := a.lookupTool(toolName); ok {
        req := ToolRequest{
            SessionID: sessionID,
            Arguments: args,
        }
        return tool.Invoke(ctx, req)
    }

    // 2. REMOTE UTCP TOOL
    if a.UTCPClient != nil {
        if streamFlag, ok := args["stream"].(bool); ok && streamFlag {
            stream, err := a.UTCPClient.CallToolStream(ctx, toolName, args)
            // ... 流式处理
        }
        return a.UTCPClient.CallTool(ctx, toolName, args)
    }

    // 3. Unknown tool
    return nil, fmt.Errorf("unknown tool: %s", toolName)
}
```

#### 交付物
- [ ] `agent.go` - 修复 `executeTool()` 函数
- [ ] `agent_tool_test.go` - 本地工具调用单元测试
- [ ] `agent_integration_test.go` - LLM 驱动工具选择集成测试
- [ ] `docs/EXECUTE_TOOL_FIX.md` - 修复方案文档
- [ ] 更新 TODO_EXTENSION.md 记录此问题

#### 测试场景
- [ ] 本地工具直接调用（`detectDirectToolCall` 路径）
- [ ] 本地工具 LLM 决策调用（`toolOrchestrator` 路径）
- [ ] UTCP 工具调用（确保不受影响）
- [ ] `AgentToolAdapter` 调用（确保不受影响）
- [ ] 未知工具错误处理
- [ ] 流式工具调用
- [ ] 并发工具调用

#### 技术亮点
- ✅ **本地工具完整支持**：LLM 驱动和直接调用两种方式
- ✅ **调用顺序优化**：本地优先，减少不必要的 UTCP 调用
- ✅ **向后兼容**：不影响现有 UTCP 工具调用
- ✅ **代码清晰**：移除误导性注释，逻辑清晰
- ✅ **测试覆盖**：完整的单元测试和集成测试

#### 与 EXT-013 的关系
- **EXT-013**：中文工具调用识别改进（识别层）
- **EXT-014**：本地工具调用修复（执行层）
- 两者互补，共同完善工具调用系统
- 建议一起完成，但 EXT-014 优先级更高（修复缺陷）

#### 备注
- **高优先级**：这是架构缺陷修复，不是功能增强
- 预计 4 小时完成（修复 + 测试 + 文档）
- 不影响现有功能，向后兼容
- 修复后本地工具生态系统才能完整工作

---

### [EXT-013] 中文工具调用识别改进

**状态**: 📝 `designing`
**优先级**: `P2`
**创建日期**: 2026-03-06
**负责人**: (待分配)
**预计工作量**: 8 小时

#### 描述
改进 `userLooksLikeToolCall()` 函数对中文工具调用识别的支持，添加中文前缀识别、正则表达式优化和可选的 LLM 智能判断，提升中文用户的使用体验。

#### 当前问题
- ❌ 不支持中文工具前缀（`工具：`、` 调用：` 等）
- ❌ 不支持中文自然语言表达
- ❌ 硬编码英文 `tool:` 前缀

#### 改进方案

**方案 1：中文前缀支持**（P2, 2 小时）
- [ ] 添加中文前缀列表（`工具：`、` 调用：`、` 运行：`、` 执行：`、` 使用：`）
- [ ] 修改 `userLooksLikeToolCall()` 函数
- [ ] 支持中文冒号和英文冒号

**方案 2：正则表达式优化**（P2, 4 小时）
- [ ] 定义预编译正则模式
- [ ] 支持多种中文前缀变体
- [ ] 支持工具名直接 +{} 格式
- [ ] 支持 JSON 格式

**方案 3：LLM 智能判断**（P3, 6 小时，可选）
- [ ] 实现 LLM 判断逻辑
- [ ] 添加配置开关
- [ ] 添加调用频率限制
- [ ] 错误降级处理

#### 子任务
- [ ] **EXT-013-A: 子代理命令识别优化**（P2, 4 小时）
  - [ ] 支持中文前缀（`调用子代理 `、` 让子代理 `、` 使用子代理` 等）
  - [ ] 支持自然语言表达（`帮我调用 `、` 找个懂代码的` 等）
  - [ ] 提取 `executeSubAgentCommand()` 独立函数
  - [ ] 实现 LLM 意图分类器（可选）
  - [ ] 文档：`docs/SUBAGENT_COMMAND_OPTIMIZATION.md` ✅ 已完成

#### 需求
- [ ] **阶段 1：中文前缀支持**（2 小时）
  - [ ] 修改 `userLooksLikeToolCall()` 函数
  - [ ] 添加中文前缀列表常量
  - [ ] 编写单元测试（中文场景）
  - [ ] 性能基准测试
- [ ] **阶段 2：正则表达式优化**（4 小时）
  - [ ] 定义预编译正则模式
  - [ ] 重构识别逻辑
  - [ ] 支持更多中文变体
  - [ ] 集成测试
- [ ] **阶段 3：LLM 智能判断**（可选，6 小时）
  - [ ] 实现 LLM 判断逻辑
  - [ ] 添加配置选项
  - [ ] 错误降级处理
- [ ] **阶段 4：文档和示例**（2 小时）
  - [ ] 更新使用文档
  - [ ] 添加中文示例
  - [ ] 编写最佳实践

#### 交付物
- [ ] `agent.go` - 改进 `userLooksLikeToolCall()` 函数
- [ ] `agent_chinese_test.go` - 中文场景单元测试
- [ ] `docs/CHINESE_TOOL_CALL_SUPPORT.md` - 设计方案文档 ✅ 已完成
- [ ] `docs/SUBAGENT_COMMAND_OPTIMIZATION.md` - 子代理命令优化文档 ✅ 已完成
- [ ] 中文工具调用示例

#### 技术亮点
- ✅ **中文前缀支持**：识别常见中文工具调用前缀
- ✅ **正则优化**：支持多种格式变体
- ✅ **智能降级**：LLM 判断作为可选增强
- ✅ **向后兼容**：不影响现有英文调用
- ✅ **性能优化**：预编译正则，快速匹配

#### 备注
- 中优先级，提升中文用户体验的关键功能
- 预计 8 小时完成基础功能
- 可以分阶段发布
- LLM 判断为可选功能

---

### [EXT-012] 扩展角色系统

**状态**: 📝 `designing`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 32 小时（4 周）  

#### 描述
扩展 Lattice 现有的角色识别系统，从当前的 4 个基础角色扩展到支持更多角色类型，包括工具响应、子代理、情感识别、错误处理、元数据等，以提升对话分析能力和系统可观测性。

#### 扩展方向

**方向 1：工具响应角色**（P1, 4 小时）
- [ ] 新增 `Tool`、`Tool(success)`、`Tool(failed)` 角色
- [ ] 实现工具调用成功率统计
- [ ] 实现工具响应质量分析
- [ ] 工具错误分类

**方向 2：子代理角色**（P1, 8 小时）
- [ ] 实现子代理角色动态注册机制
- [ ] 预定义子代理（Researcher、Coder、Writer、Reviewer、Analyzer 等）
- [ ] 实现子代理贡献统计
- [ ] 实现子代理专业性评估

**方向 3：情感识别角色**（P2, 6 小时）
- [ ] 实现情感标记注册机制
- [ ] 预定义情感（happy、sad、angry、frustrated、confused、excited、satisfied 等）
- [ ] 实现情感变化追踪
- [ ] 实现用户满意度计算
- [ ] 实现情感感知回答策略

**方向 4：错误处理角色**（P2, 4 小时）
- [ ] 实现错误角色定义（ERROR、WARNING、CRITICAL）
- [ ] 实现错误分类（TOOL_ERROR、MODEL_ERROR、NETWORK_ERROR 等）
- [ ] 实现错误统计和趋势分析
- [ ] 实现错误告警机制

**方向 5：元数据角色**（P3, 4 小时）
- [ ] 实现元数据标记（TIMING、DEBUG、PERFORMANCE）
- [ ] 实现性能指标提取
- [ ] 实现调试信息分离
- [ ] 实现系统日志集成

#### 需求
- [ ] **阶段 1：核心框架**（1 周）
  - [ ] 创建 `src/agent/roles/` 目录结构
  - [ ] 实现 `RoleManager` 核心类
  - [ ] 实现工具响应角色
  - [ ] 编写单元测试
- [ ] **阶段 2：子代理角色**（1 周）
  - [ ] 实现 `SubAgentRoleManager`
  - [ ] 预定义常用子代理
  - [ ] 实现子代理贡献统计
  - [ ] 集成测试
- [ ] **阶段 3：情感识别**（1 周）
  - [ ] 实现 `EmotionRoleManager`
  - [ ] 预定义情感类型
  - [ ] 实现情感变化追踪
  - [ ] 实现用户满意度计算
- [ ] **阶段 4：错误处理**（3 天）
  - [ ] 实现错误角色定义
  - [ ] 实现错误分类和统计
  - [ ] 实现错误告警机制
- [ ] **阶段 5：元数据角色**（3 天）
  - [ ] 实现元数据角色定义
  - [ ] 实现性能指标提取
  - [ ] 实现调试信息分离
- [ ] **阶段 6：文档和示例**（2 天）
  - [ ] 编写 API 文档
  - [ ] 编写使用示例
  - [ ] 性能基准测试
  - [ ] 最佳实践指南

#### 交付物
- [ ] `src/agent/roles/role_manager.go` - 统一角色管理器
- [ ] `src/agent/roles/tool_role.go` - 工具响应角色
- [ ] `src/agent/roles/subagent_role.go` - 子代理角色
- [ ] `src/agent/roles/emotion_role.go` - 情感识别角色
- [ ] `src/agent/roles/error_role.go` - 错误处理角色
- [ ] `src/agent/roles/metadata_role.go` - 元数据角色
- [ ] `docs/ROLE_SYSTEM_EXTENSION.md` - 设计方案文档 ✅ 已完成
- [ ] 完整的单元测试和集成测试
- [ ] 使用示例和文档

#### 技术亮点
- ✅ **统一角色管理**：一个管理器识别所有角色类型
- ✅ **动态注册**：运行时注册新角色
- ✅ **缓存优化**：识别结果缓存，提高性能
- ✅ **线程安全**：支持并发访问
- ✅ **向后兼容**：不影响现有功能

#### 备注
- 中优先级，提升系统可观测性的关键功能
- 预计 4 周完成全部功能
- 可以分阶段发布
- 需要充分的性能测试

---

### [EXT-011] MemCell & MemScene 记忆系统

**状态**: 📝 `designing`  
**优先级**: `P1`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 80 小时（8 周）  

#### 描述
基于 EverMemOS 核心思想，在 Lattice 框架内实现 MemCell（记忆细胞）和 MemScene（记忆场景）系统。充分利用 Lattice 已有的 `memory.Engine`、图感知特性和 `Shared Spaces`，构建具备 SOTA 潜力的记忆系统。

#### 核心概念
- **MemCell**: 记忆的最小功能单元，具有独立的编码、存储、检索和更新能力
- **MemScene**: MemCell 的组织单元，将相关的 MemCell 按场景/上下文分组
- **激活扩散**: 模拟人类记忆的激活扩散机制，实现智能检索
- **层次化场景**: 支持嵌套的子场景结构

#### 需求
- [ ] **阶段 1：核心数据结构**（2 周）
  - [ ] 实现 `MemCell` 基础结构（`src/memory/cell/cell.go`）
  - [ ] 实现 `MemScene` 基础结构（`src/memory/scene/scene.go`）
  - [ ] 实现 `MemSceneManager`（`src/memory/scene/manager.go`）
  - [ ] 编写单元测试
- [ ] **阶段 2：激活模型**（1 周）
  - [ ] 实现激活衰减算法（`src/memory/activation/decay.go`）
  - [ ] 实现激活扩散算法（`src/memory/activation/spreading.go`）
  - [ ] 实现优先级队列
  - [ ] 性能基准测试
- [ ] **阶段 3：图集成**（1 周）
  - [ ] 实现图遍历算法（`src/memory/graph/traversal.go`）
  - [ ] 实现聚类算法（`src/memory/graph/clustering.go`）
  - [ ] 与 Neo4j 集成
  - [ ] 图查询优化
- [ ] **阶段 4：Engine 集成**（1 周）
  - [ ] 实现 `ExtendedEngine`（`src/memory/integration/engine_ext.go`）
  - [ ] 与现有 `memory.Engine` 集成
  - [ ] 向后兼容性测试
  - [ ] 性能优化
- [ ] **阶段 5：ADK 集成**（1 周）
  - [ ] 实现 `MemCellModule`（`src/memory/integration/adk_module.go`）
  - [ ] ADK 模块测试
  - [ ] 示例代码
  - [ ] 文档编写
- [ ] **阶段 6：高级功能**（2 周）
  - [ ] 场景层次优化
  - [ ] 自动 pruning 策略
  - [ ] 访问模式学习
  - [ ] SOTA 特性实验

#### 交付物
- [ ] `src/memory/cell/` - MemCell 核心实现
- [ ] `src/memory/scene/` - MemScene 核心实现
- [ ] `src/memory/activation/` - 激活模型
- [ ] `src/memory/graph/` - 图算法
- [ ] `src/memory/integration/` - 集成模块
- [ ] `docs/MEMCELL_MEMSCENE_DESIGN.md` - 设计方案文档 ✅ 已完成
- [ ] 完整的单元测试和集成测试
- [ ] 使用示例和文档

#### 与现有组件的映射
| EverMemOS 概念 | Lattice 对应组件 | 扩展方式 |
|---------------|------------------|----------|
| Memory Cell | `memory.MemoryRecord` | 扩展为 `MemCell` |
| Scene | `memory.Space` | 扩展为 `MemScene` |
| Engine | `memory.Engine` | 增强引擎功能 |
| Graph Edge | `model.GraphEdge` | 直接使用 |
| Embedding | `embed.Embedder` | 直接使用 |

#### 技术亮点
- ✅ **自包含记忆细胞**：每个 MemCell 独立管理生命周期
- ✅ **激活扩散算法**：模拟人类记忆的联想检索
- ✅ **层次化场景**：支持嵌套的子场景结构
- ✅ **图感知存储**：利用 Neo4j/PostgreSQL 的图能力
- ✅ **向后兼容**：与现有 `memory.Engine` 完全兼容
- ✅ **ADK 集成**：通过模块系统无缝集成

#### 备注
- 高优先级，这是 Lattice 的核心竞争力
- 预计 8 周完成全部功能
- 可以分阶段发布
- 需要充分的性能测试

---

### [EXT-001] DeepSeek LLM 集成

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 4 小时  

#### 描述
集成 DeepSeek（深度求索）LLM 提供商，支持文本生成、多模态输入和流式输出。DeepSeek API 与 OpenAI 兼容，可复用现有代码。

#### 需求
- [ ] 创建 `src/models/deepseek.go` 文件
- [ ] 实现 `NewDeepSeekLLM()` 构造函数
- [ ] 实现 `Generate()` 方法
- [ ] 实现 `GenerateWithFiles()` 方法
- [ ] 实现 `GenerateStream()` 方法
- [ ] 实现 `GenerateStreamWithFiles()` 方法（新增功能）
- [ ] 添加接口检查 `var _ Agent = (*DeepSeekLLM)(nil)`
- [ ] 编写单元测试（3 个测试用例）
- [ ] 更新 README.md 添加 DeepSeek 示例
- [ ] 测试环境变量配置
- [ ] 测试不同模型（chat/coder/v2.5）
- [ ] 测试图片上传功能
- [ ] 测试流式输出

#### 交付物
- [ ] `src/models/deepseek.go` - 实现代码
- [ ] `src/models/deepseek_test.go` - 单元测试
- [ ] `docs/DEEPSEEK_IMPLEMENTATION.md` - 实现方案文档 ✅ 已完成

#### 参考资料
- [DeepSeek 开放平台](https://platform.deepseek.com/)
- [DeepSeek API 文档](https://platform.deepseek.com/api-docs/)
- [go-openai 库](https://github.com/sashabaranov/go-openai)

#### 备注
- DeepSeek API 与 OpenAI 完全兼容，可复用 `go-openai` 库
- 支持带文件的流式生成（独特优势）
- 人民币计价，价格更低
- 中文优化更好

---

### [EXT-002] 扩展 models.Agent 接口支持带文件流式

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 2 小时  

#### 描述
当前 `models.Agent` 接口缺少 `GenerateStreamWithFiles()` 方法，导致接口不完整。需要扩展接口定义并在所有实现中添加该方法。

#### 需求
- [ ] 修改 `src/models/interface.go` 添加第 4 个方法
- [ ] 更新 `GeminiLLM` 实现
- [ ] 更新 `AnthropicLLM` 实现
- [ ] 更新 `OllamaLLM` 实现
- [ ] 更新 `OpenAILLM` 实现
- [ ] 更新 `CachedLLM` 装饰器
- [ ] 更新 `DummyLLM` 测试实现

#### 交付物
- [ ] `src/models/interface.go` - 扩展接口
- [ ] 所有 LLM 实现的更新

#### 接口变更
```go
type Agent interface {
    Generate(context.Context, string) (any, error)
    GenerateWithFiles(context.Context, string, []File) (any, error)
    GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
    
    // 新增
    GenerateStreamWithFiles(ctx context.Context, prompt string, files []File) (<-chan StreamChunk, error)
}
```

#### 备注
- 这是一个破坏性变更，需要更新所有实现
- 可以提供默认回退实现减少工作量
- 建议与 EXT-001 一起完成

---

### [EXT-003] 添加更多预构建子代理

**状态**: 📋 `pending`  
**优先级**: `P3`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 8 小时  

#### 描述
当前只有 `Researcher` 一个预构建子代理，需要添加更多专家代理以丰富框架功能。

#### 需求
- [ ] `Coder` - 代码生成和审查代理
- [ ] `Writer` - 内容创作和编辑代理
- [ ] `Reviewer` - 质量评审代理
- [ ] `Translator` - 多语言翻译代理
- [ ] `Summarizer` - 文本摘要代理
- [ ] `Analyzer` - 数据分析代理
- [ ] `Planner` - 任务规划代理
- [ ] `Critic` - 批判性思维代理

#### 交付物
- [ ] `src/subagents/coder.go`
- [ ] `src/subagents/writer.go`
- [ ] `src/subagents/reviewer.go`
- [ ] `src/subagents/translator.go`
- [ ] `src/subagents/summarizer.go`
- [ ] `src/subagents/analyzer.go`
- [ ] `src/subagents/planner.go`
- [ ] `src/subagents/critic.go`
- [ ] 每个子代理的单元测试
- [ ] 使用示例文档

#### 备注
- 每个代理应有清晰的角色定位和提示词
- 支持自定义系统提示词
- 考虑多语言支持

---

### [EXT-004] 添加更多内置工具

**状态**: 📋 `pending`  
**优先级**: `P3`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 12 小时  

#### 描述
丰富内置工具库，让 Agent 可以直接使用常用功能而无需外部依赖。

#### 需求
- [ ] `FileReadTool` - 读取本地文件
- [ ] `FileWriteTool` - 写入本地文件
- [ ] `WebSearchTool` - 网络搜索
- [ ] `HTTPClientTool` - HTTP 请求
- [ ] `DatabaseQueryTool` - SQL 查询
- [ ] `CodeExecutorTool` - 代码执行
- [ ] `ShellCommandTool` -  shell 命令（需谨慎）
- [ ] `CalendarTool` - 日历和提醒
- [ ] `EmailTool` - 邮件发送
- [ ] `ImageAnalysisTool` - 图片分析

#### 交付物
- [ ] `src/tools/` 目录下的工具实现
- [ ] 每个工具的使用示例
- [ ] 安全策略文档（特别是 shell 命令执行）

#### 备注
- 需要实现权限控制机制
- 敏感工具需要用户确认
- 考虑沙箱执行环境

---

### [EXT-005] 实现 Agent 间通信协议

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 16 小时  

#### 描述
实现更丰富的 Agent 间通信机制，支持分布式 Agent 协作。

#### 需求
- [ ] 定义 Agent 间通信消息格式
- [ ] 实现基于 gRPC 的通信
- [ ] 实现基于 HTTP 的通信
- [ ] 实现基于消息队列的通信（Redis/RabbitMQ）
- [ ] 支持 Agent 发现和注册
- [ ] 支持 Agent 负载均衡
- [ ] 实现 Agent 通信日志和追踪

#### 交付物
- [ ] `src/communication/` 目录
- [ ] 通信协议规范文档
- [ ] 分布式 Agent 示例

#### 备注
- 可参考现有 Agent 通信协议（如 ACP）
- 需要考虑安全性和认证机制

---

### [EXT-006] 实现 Agent 可观测性

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 12 小时  

#### 描述
实现 Agent 运行的可观测性，包括日志、指标和追踪。

#### 需求
- [ ] 结构化日志记录
- [ ] Prometheus 指标导出
- [ ] OpenTelemetry 追踪集成
- [ ] Agent 行为审计日志
- [ ] Token 使用统计
- [ ] 成本计算和报告
- [ ] 性能分析工具

#### 交付物
- [ ] `src/observability/` 目录
- [ ] 监控仪表板示例（Grafana）
- [ ] 可观测性配置文档

#### 备注
- 需要支持多种后端（Prometheus, Jaeger, etc.）
- 考虑隐私和数据保护

---

### [EXT-007] 实现 Agent 持久化存储

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 8 小时  

#### 描述
增强 Agent 状态持久化功能，支持多种存储后端。

#### 需求
- [ ] Redis 存储后端
- [ ] S3 兼容存储后端
- [ ] 本地文件系统后端
- [ ] 数据库后端（PostgreSQL/MySQL）
- [ ] 状态版本控制
- [ ] 状态快照和恢复
- [ ] 状态迁移工具

#### 交付物
- [ ] `src/persistence/` 目录
- [ ] 存储适配器实现
- [ ] 持久化配置文档

#### 备注
- 需要支持加密存储
- 考虑状态压缩

---

### [EXT-008] 实现 Agent 安全沙箱

**状态**: 📋 `pending`  
**优先级**: `P1`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 20 小时  

#### 描述
实现 Agent 执行的安全沙箱，防止恶意或意外操作。

#### 需求
- [ ] 工具调用权限控制
- [ ] 资源使用限制（CPU/内存/时间）
- [ ] 网络访问控制
- [ ] 文件系统访问控制
- [ ] 环境变量隔离
- [ ] 安全策略引擎
- [ ] 审计和告警

#### 交付物
- [ ] `src/sandbox/` 目录
- [ ] 安全策略配置文档
- [ ] 沙箱使用示例

#### 备注
- 高优先级，涉及安全问题
- 可参考现有沙箱方案（gVisor, Firecracker）

---

### [EXT-009] 实现 Agent 工作流编排

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 16 小时  

#### 描述
实现更复杂的工作流编排能力，支持条件分支、循环、并行等。

#### 需求
- [ ] 可视化工作流编辑器
- [ ] 工作流 DSL 定义
- [ ] 条件分支支持
- [ ] 循环和迭代支持
- [ ] 并行执行支持
- [ ] 错误处理和重试
- [ ] 工作流状态管理
- [ ] 工作流版本控制

#### 交付物
- [ ] `src/workflow/` 目录
- [ ] 工作流 DSL 规范
- [ ] 可视化编辑器（可选）
- [ ] 工作流示例库

#### 备注
- 可参考 Argo Workflows, Airflow 等
- 考虑与 UTCP Chain 的集成

---

### [EXT-010] 实现 Agent 测试框架

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: 2026-03-06  
**负责人**: (待分配)  
**预计工作量**: 12 小时  

#### 描述
实现专门的 Agent 测试框架，支持 Mock LLM、回放测试等。

#### 需求
- [ ] Mock LLM 实现
- [ ] 响应录制和回放
- [ ] 确定性测试支持
- [ ] 性能基准测试
- [ ] 集成测试框架
- [ ] 测试覆盖率报告
- [ ] 测试用例库

#### 交付物
- [ ] `src/testing/` 目录
- [ ] 测试框架文档
- [ ] 测试用例示例

#### 备注
- 解决 LLM 非确定性带来的测试困难
- 支持离线测试

---

## 📈 统计信息

### 按状态统计

| 状态 | 数量 | 百分比 |
|------|------|--------|
| 📋 pending | 14 | 82% |
| 🔍 researching | 0 | 0% |
| 📝 designing | 3 | 18% |
| 💻 implementing | 0 | 0% |
| 🧪 testing | 0 | 0% |
| 📖 documenting | 0 | 0% |
| ✅ done | 1 | - |
| 🚫 blocked | 0 | 0% |
| ❌ cancelled | 0 | 0% |

### 按优先级统计

| 优先级 | 数量 | 百分比 |
|--------|------|--------|
| P0 | 1 | 6% |
| P1 | 5 | 29% |
| P2 | 9 | 53% |
| P3 | 2 | 12% |

### 预计工作量

| 优先级 | 预计小时数 |
|--------|-----------|
| P0 | 500+ |
| P1 | 128 |
| P2 | 118 |
| P3 | 20 |
| **总计** | **约 766+ 小时** |

---

## 📝 变更日志

| 日期 | 变更内容 | 操作人 |
|------|----------|--------|
| 2026-03-06 | 初始版本，创建 ToDoList 框架 | - |
| 2026-03-06 | 添加 EXT-001 DeepSeek 集成任务 | - |
| 2026-03-06 | 添加 EXT-002 接口扩展任务 | - |
| 2026-03-06 | 添加 EXT-003 到 EXT-010 任务 | - |
| 2026-03-06 | 添加 EXT-011 MemCell & MemScene 记忆系统任务 | - |
| 2026-03-06 | 添加 EXT-012 扩展角色系统任务 | - |
| 2026-03-06 | 添加 EXT-013 中文工具调用识别改进任务 | - |
| 2026-03-06 | 更新统计信息（总任务 13 项，总计 238 小时） | - |
| 2026-03-07 | 添加 EXT-014 修复 executeTool() 本地工具调用缺陷任务 | - |
| 2026-03-07 | 更新统计信息（总任务 14 项，P1 优先级 3 项，总计 242 小时） | - |
| 2026-03-07 | 添加 EXT-015 可确认 Agent 扩展 - 异步确认与授权管理任务 | - |
| 2026-03-07 | 创建详细实施方案文档 (CONFIRMABLE_AGENT_IMPLEMENTATION.md) | - |
| 2026-03-07 | 更新统计信息（总任务 15 项，P1 优先级 4 项，总计 258 小时） | - |
| 2026-03-07 | 添加 EXT-016 智能记忆与因果推理系统任务（P0 优先级） | - |
| 2026-03-07 | 创建智能记忆系统设计文档 (SMART_MEMORY_SYSTEM_DESIGN.md) | - |
| 2026-03-07 | 更新统计信息（总任务 16 项，P0 优先级 1 项，总计约 758+ 小时） | - |
| 2026-03-07 | 添加 EXT-017 端到端流式传输支持任务 | - |
| 2026-03-07 | 创建流式传输设计文档 (STREAMING_IMPLEMENTATION.md) | - |
| 2026-03-07 | 更新统计信息（总任务 17 项，P1 优先级 5 项，总计约 766+ 小时） | - |

---

## 🔗 相关链接

- [项目主 README](../README.md)
- [代码页清单](../CODE_INVENTORY.md)
- [DeepSeek 实现方案](./DEEPSEEK_IMPLEMENTATION.md)
- [MemCell & MemScene 设计方案](./MEMCELL_MEMSCENE_DESIGN.md)
- [角色系统扩展设计方案](./ROLE_SYSTEM_EXTENSION.md)
- [中文工具调用支持设计方案](./CHINESE_TOOL_CALL_SUPPORT.md)
- [子代理命令优化设计方案](./SUBAGENT_COMMAND_OPTIMIZATION.md)
- [项目规则](../PROJECT_GUIDELINES.md)

---

## 📌 使用说明

### 添加新任务

请在文档末尾添加新任务，格式如下：

```markdown
### [EXT-XXX] 任务名称

**状态**: 📋 `pending`  
**优先级**: `P2`  
**创建日期**: YYYY-MM-DD  
**负责人**: (待分配)  
**预计工作量**: X 小时  

#### 描述
任务描述...

#### 需求
- [ ] 需求 1
- [ ] 需求 2

#### 交付物
- [ ] 文件 1
- [ ] 文件 2

#### 备注
其他说明...
```

### 更新任务状态

使用以下命令更新任务状态：

```
/extend-todo update EXT-XXX status=<new-status>
```

### 完成任务

完成任务时，请确保：
- [ ] 所有代码已提交
- [ ] 所有测试已通过
- [ ] 文档已更新
- [ ] 变更日志已记录

---

*本文档由 go-agent 团队维护，欢迎贡献和建议。*
