# 记忆系统性能调优指南

**文档编号**: OPS-02  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐⭐ 核心文档  
**适用角色**: 后端工程师 / SRE 工程师 / 性能优化工程师  

---

## 📋 目录

1. [MMR 参数调优](#1-mmr 参数调优)
2. [存储性能优化](#2-存储性能优化)
3. [检索性能优化](#3-检索性能优化)
4. [性能基准测试](#4-性能基准测试)
5. [附录](#5-附录)

---

## 1. MMR 参数调优

### 1.1 参数详解

#### 1.1.1 Lambda (λ) 参数

**作用**：控制相关性与多样性的平衡

```
λ = 0.0          λ = 0.5          λ = 0.7          λ = 1.0
  │                │                │                │
  ▼                ▼                ▼                ▼
┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐
│ 只看   │    │ 平衡   │    │ 偏重   │    │ 只看   │
│ 多样性 │    │ 模式   │    │ 相关性 │    │ 相关性 │
│        │    │        │    │ (默认) │    │        │
└────────┘    └────────┘    └────────┘    └────────┘
  │                │                │                │
  ▼                ▼                ▼                ▼
结果非常       结果平衡       结果相关       结果最相关
多样化         适中           性高           但可能重复
```

**调优建议**：

| 场景 | 推荐值 | 说明 |
|------|--------|------|
| 知识库检索 | 0.8-0.9 | 准确性优先 |
| 客服对话 | 0.6-0.7 | 平衡准确性和覆盖度 |
| 创意灵感 | 0.3-0.5 | 多样性优先 |
| 通用搜索 | 0.7 | 默认推荐值 |
| 学术研究 | 0.5-0.6 | 需要全面覆盖 |

#### 1.1.2 权重参数

**ScoreWeights 结构**：

```go
type ScoreWeights struct {
    Similarity float64  // 向量相似度权重
    Keywords   float64  // 关键词匹配权重
    Importance float64  // 重要性权重
    Recency    float64  // 新鲜度权重
    Source     float64  // 来源权重
}
```

**权重分配原则**：
- 所有权重之和应为 1.0（系统会自动归一化）
- 根据业务场景调整权重分配

**场景化配置**：

| 场景 | Similarity | Keywords | Importance | Recency | Source |
|------|------------|----------|------------|---------|--------|
| **知识库** | 0.50 | 0.25 | 0.15 | 0.05 | 0.05 |
| **客服** | 0.35 | 0.30 | 0.20 | 0.10 | 0.05 |
| **创意** | 0.30 | 0.20 | 0.20 | 0.20 | 0.10 |
| **学术** | 0.35 | 0.25 | 0.25 | 0.10 | 0.05 |
| **通用** | 0.45 | 0.20 | 0.20 | 0.10 | 0.05 |

### 1.2 调优流程

#### 1.2.1 调优步骤

```
┌─────────────────────────────────────────┐
│  步骤 1: 明确业务场景                    │
│  - 用户需要什么类型的结果？              │
│  - 相关性重要还是多样性重要？            │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  步骤 2: 选择初始参数                    │
│  - 从默认值开始 (λ=0.7)                  │
│  - 根据场景调整权重                      │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  步骤 3: 构建测试集                      │
│  - 收集 20-50 个典型查询                  │
│  - 标注期望的返回结果                    │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  步骤 4: 批量测试                        │
│  - 遍历参数组合                          │
│  - 计算准确率/召回率/NDCG                │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  步骤 5: A/B 测试                         │
│  - 小流量测试最优参数                    │
│  - 监控业务指标                          │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  步骤 6: 持续监控                        │
│  - 用户满意度                            │
│  - 点击率/转化率                         │
└─────────────────────────────────────────┘
```

#### 1.2.2 测试集构建

**测试查询示例**：

```go
type TestQuery struct {
    Query        string   // 查询文本
    ExpectedIDs  []int64  // 期望返回的记忆 ID
    ExpectedRank map[int64]int // 期望排名
}

var testQueries = []TestQuery{
    {
        Query: "如何重置密码",
        ExpectedIDs: []int64{1, 2, 3},
        ExpectedRank: map[int64]int{
            1: 1,  // 期望第 1 条排第 1
            2: 2,  // 期望第 2 条排第 2
            3: 3,  // 期望第 3 条排第 3
        },
    },
    {
        Query: "产品定价",
        ExpectedIDs: []int64{10, 11, 12},
        ExpectedRank: map[int64]int{
            10: 1,
            11: 2,
            12: 3,
        },
    },
    // ... 更多测试查询
}
```

### 1.3 网格搜索方法

#### 1.3.1 参数范围定义

```go
// 定义参数搜索范围
var (
    lambdaValues = []float64{0.3, 0.5, 0.6, 0.7, 0.8, 0.9}
    
    similarityWeights = []float64{0.3, 0.4, 0.5, 0.6}
    
    // 其他权重按比例分配
    remainingWeights = map[string]float64{
        "keywords":   0.4,
        "importance": 0.3,
        "recency":    0.2,
        "source":     0.1,
    }
)
```

#### 1.3.2 网格搜索实现

```go
func GridSearch(ctx context.Context, eng *memory.Engine, 
                testQueries []TestQuery) []TestResult {
    
    var results []TestResult
    
    for _, lambda := range lambdaValues {
        for _, simW := range similarityWeights {
            // 计算其他权重
            remaining := 1.0 - simW
            weights := memory.ScoreWeights{
                Similarity: simW,
                Keywords:   remaining * 0.4,
                Importance: remaining * 0.3,
                Recency:    remaining * 0.2,
                Source:     remaining * 0.1,
            }
            
            // 创建临时引擎
            tempEng := memory.NewEngine(eng.Store, memory.Options{
                LambdaMMR: lambda,
                Weights:   weights,
            })
            
            // 运行测试
            result := runTests(ctx, tempEng, testQueries)
            result.Lambda = lambda
            result.Weights = weights
            results = append(results, result)
            
            fmt.Printf("λ=%.2f, sim=%.2f → P=%.3f, R=%.3f, NDCG=%.3f\n",
                lambda, simW, result.Precision, result.Recall, result.NDCG)
        }
    }
    
    // 排序找出最佳参数
    sort.Slice(results, func(i, j int) bool {
        scoreI := 0.5*results[i].NDCG + 0.3*results[i].Precision + 0.2*results[i].Recall
        scoreJ := 0.5*results[j].NDCG + 0.3*results[j].Precision + 0.2*results[j].Recall
        return scoreI > scoreJ
    })
    
    return results
}
```

#### 1.3.3 评估指标

**计算公式**：

```go
// 准确率 (Precision)
Precision = (返回的相关结果数) / (返回的总结果数)

// 召回率 (Recall)
Recall = (返回的相关结果数) / (所有相关结果数)

// NDCG (归一化折损累计增益)
DCG = Σ (rel_i / log2(i+1))
NDCG = DCG / IDCG
```

**指标说明**：

| 指标 | 说明 | 优点 | 缺点 |
|------|------|------|------|
| Precision | 返回结果中相关的比例 | 简单直观 | 不考虑排名 |
| Recall | 相关结果被返回的比例 | 考虑覆盖率 | 不考虑排名 |
| **NDCG** | 考虑排名的综合指标 | **考虑排名位置** | **计算复杂** |

**推荐**：使用 NDCG 作为主要评估指标

### 1.4 A/B 测试流程

#### 1.4.1 实验设计

```
┌─────────────────────────────────────────┐
│           A/B 测试设计                    │
├─────────────────────────────────────────┤
│                                         │
│  对照组 (A 组)                           │
│  - 使用当前生产参数                      │
│  - λ = 0.7                              │
│  - 流量占比：50%                        │
│                                         │
│  实验组 (B 组)                           │
│  - 使用新调优参数                        │
│  - λ = 0.8 (示例)                       │
│  - 流量占比：50%                        │
│                                         │
│  观察指标                                │
│  - 用户满意度 (1-5 分)                    │
│  - 点击率 (CTR)                         │
│  - 转化率                               │
│  - NDCG 得分                             │
│                                         │
└─────────────────────────────────────────┘
```

#### 1.4.2 流量分配

```go
// 流量分配示例
type TrafficSplit struct {
    ControlGroup   float64  // 对照组流量比例
    ExperimentGroup float64 // 实验组流量比例
}

func shouldUseExperiment(userID string, split TrafficSplit) bool {
    hash := md5.Sum([]byte(userID))
    hashInt := binary.BigEndian.Uint32(hash[:4])
    ratio := float64(hashInt) / math.MaxUint32
    
    return ratio < split.ExperimentGroup
}
```

#### 1.4.3 结果分析

**统计显著性检验**：

```go
// 使用 t 检验判断差异是否显著
func tTest(control, experiment []float64) (tStatistic, pValue float64) {
    n1 := float64(len(control))
    n2 := float64(len(experiment))
    
    mean1 := mean(control)
    mean2 := mean(experiment)
    
    var1 := variance(control)
    var2 := variance(experiment)
    
    se := math.Sqrt(var1/n1 + var2/n2)
    tStatistic = (mean1 - mean2) / se
    
    // 计算 p 值 (简化)
    df := n1 + n2 - 2
    pValue = calculatePValue(tStatistic, df)
    
    return
}

// 判断标准
if pValue < 0.05 {
    // 差异显著，可以采纳新参数
} else {
    // 差异不显著，需要更多数据或重新设计实验
}
```

---

## 2. 存储性能优化

### 2.1 向量数据库选型对比

| 数据库 | 优点 | 缺点 | 适用场景 |
|--------|------|------|----------|
| **Qdrant** | 性能优秀、支持过滤、易部署 | 相对较新、社区较小 | 生产环境首选 |
| **PostgreSQL+pgvector** | 成熟稳定、SQL 支持、易集成 | 性能略低、资源消耗大 | 中小规模、已有 PG |
| **Milvus** | 功能丰富、可扩展性强 | 部署复杂、资源消耗大 | 大规模、高并发 |
| **Weaviate** | 内置向量 + 图、易使用 | 性能一般、社区较小 | 知识图谱场景 |
| **In-Memory** | 性能最佳、零延迟 | 数据易失、容量有限 | 测试、缓存层 |

### 2.2 索引参数调优

#### 2.2.1 HNSW 参数（Qdrant）

```yaml
# Qdrant HNSW 配置
hnsw_config:
  m: 16              # 每个节点的连接数，越大越精确但占用更多内存
  ef_construct: 100  # 构建时的搜索深度，越大构建越慢但质量越高
  full_scan_threshold: 10000  # 全量扫描阈值
  
# 参数调优建议
# 小规模 (< 100 万): m=16, ef_construct=100
# 中规模 (100 万 -1000 万): m=32, ef_construct=200
# 大规模 (> 1000 万): m=64, ef_construct=400
```

#### 2.2.2 IVF 参数（Milvus）

```yaml
# Milvus IVF 配置
index_params:
  index_type: IVF_FLAT
  metric_type: COSINE
  params:
    nlist: 1024      # 聚类中心数量
    nprobe: 32       # 搜索时探测的聚类数量
    
# 参数调优建议
# nlist = 4 * sqrt(N) 其中 N 是向量数量
# nprobe = sqrt(nlist) 作为起点
```

### 2.3 批量操作优化

#### 2.3.1 批量存储

```go
// 批量存储优化
func BatchStore(ctx context.Context, eng *memory.Engine, 
                sessionID string, items []MemoryItem) error {
    
    const batchSize = 100  // 每批 100 条
    
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        
        // 并发处理批次
        err := processBatch(ctx, eng, sessionID, batch)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func processBatch(ctx context.Context, eng *memory.Engine, 
                  sessionID string, batch []MemoryItem) error {
    
    // 并发嵌入
    embeddings := make([][]float32, len(batch))
    errs := make([]error, len(batch))
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10)  // 限制并发数
    
    for i, item := range batch {
        wg.Add(1)
        go func(idx int, it MemoryItem) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            emb, err := eng.Embed(ctx, it.Content)
            embeddings[idx] = emb
            errs[idx] = err
        }(i, item)
    }
    
    wg.Wait()
    
    // 检查错误
    for _, err := range errs {
        if err != nil {
            return err
        }
    }
    
    // 批量存储
    // ...
    
    return nil
}
```

#### 2.3.2 批量检索

```go
// 批量检索优化
func BatchRetrieve(ctx context.Context, eng *memory.Engine,
                   sessionID string, queries []string, limit int) ([][]memory.MemoryRecord, error) {
    
    results := make([][]memory.MemoryRecord, len(queries))
    errs := make([]error, len(queries))
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 5)  // 限制并发检索数
    
    for i, query := range queries {
        wg.Add(1)
        go func(idx int, q string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            recs, err := eng.Retrieve(ctx, sessionID, q, limit)
            results[idx] = recs
            errs[idx] = err
        }(i, query)
    }
    
    wg.Wait()
    
    // 检查错误
    for _, err := range errs {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

---

## 3. 检索性能优化

### 3.1 缓存策略

#### 3.1.1 查询结果缓存

```go
type RetrieveCache struct {
    cache *ristretto.Cache
    ttl   time.Duration
}

func NewRetrieveCache() *RetrieveCache {
    cache, _ := ristretto.NewCache(&ristretto.Config{
        NumCounters: 1e7,     // 1000 万键值用于追踪频率
        MaxCost:     1 << 30, // 1GB
        BufferItems: 64,
    })
    
    return &RetrieveCache{
        cache: cache,
        ttl:   5 * time.Minute,
    }
}

func (c *RetrieveCache) Get(key string) ([]memory.MemoryRecord, bool) {
    val, found := c.cache.Get(key)
    if !found {
        return nil, false
    }
    
    recs, ok := val.([]memory.MemoryRecord)
    return recs, ok
}

func (c *RetrieveCache) Set(key string, recs []memory.MemoryRecord) {
    c.cache.Set(key, recs, int64(len(recs)*100)) // cost = 记录数 * 100
}

// 使用缓存优化检索
func RetrieveWithCache(ctx context.Context, eng *memory.Engine,
                       cache *RetrieveCache, sessionID, query string, 
                       limit int) ([]memory.MemoryRecord, error) {
    
    cacheKey := fmt.Sprintf("%s:%s:%d", sessionID, query, limit)
    
    // 尝试从缓存获取
    if recs, found := cache.Get(cacheKey); found {
        return recs, nil
    }
    
    // 从引擎检索
    recs, err := eng.Retrieve(ctx, sessionID, query, limit)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    cache.Set(cacheKey, recs)
    
    return recs, nil
}
```

#### 3.1.2 嵌入向量缓存

```go
type EmbeddingCache struct {
    cache *ristretto.Cache
}

func (c *EmbeddingCache) Get(text string) ([]float32, bool) {
    val, found := c.cache.Get(text)
    if !found {
        return nil, false
    }
    return val.([]float32), true
}

func (c *EmbeddingCache) Set(text string, embedding []float32) {
    cost := int64(len(embedding) * 4) // float32 占 4 字节
    c.cache.Set(text, embedding, cost)
}
```

### 3.2 并发控制

#### 3.2.1 限流器

```go
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(qps int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(qps), qps*2),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)
}

// 使用限流器
func RetrieveWithRateLimit(ctx context.Context, eng *memory.Engine,
                           limiter *RateLimiter, sessionID, query string,
                           limit int) ([]memory.MemoryRecord, error) {
    
    // 等待限流器允许
    if err := limiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    return eng.Retrieve(ctx, sessionID, query, limit)
}
```

### 3.3 查询优化

#### 3.3.1 查询预处理

```go
// 查询预处理优化
func preprocessQuery(query string) string {
    // 1. 去除多余空白
    query = strings.TrimSpace(query)
    
    // 2. 转小写
    query = strings.ToLower(query)
    
    // 3. 去除停用词
    stopWords := map[string]struct{}{
        "a": {}, "an": {}, "the": {}, "is": {}, "are": {},
    }
    
    words := strings.Fields(query)
    filtered := make([]string, 0, len(words))
    
    for _, word := range words {
        if _, stop := stopWords[word]; !stop {
            filtered = append(filtered, word)
        }
    }
    
    return strings.Join(filtered, " ")
}
```

#### 3.3.2 查询扩展

```go
// 查询扩展（同义词）
func expandQuery(query string) []string {
    synonyms := map[string][]string{
        "密码": {"口令", "pass", "password"},
        "重置": {"重设", "reset"},
        "登录": {"登陆", "signin", "login"},
    }
    
    expanded := []string{query}
    
    for original, syns := range synonyms {
        if strings.Contains(query, original) {
            for _, syn := range syns {
                expandedQuery := strings.Replace(query, original, syn, -1)
                expanded = append(expanded, expandedQuery)
            }
        }
    }
    
    return expanded
}
```

---

## 4. 性能基准测试

### 4.1 测试工具

#### 4.1.1 基准测试框架

```go
package memory_bench

import (
    "context"
    "fmt"
    "testing"
    "time"
    
    "github.com/Protocol-Lattice/go-agent/src/memory"
)

// 基准测试：存储性能
func BenchmarkStore(b *testing.B) {
    store := memory.NewInMemoryStore()
    eng := memory.NewEngine(store, memory.DefaultOptions())
    
    ctx := context.Background()
    sessionID := "bench-session"
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        content := fmt.Sprintf("测试记忆 %d", i)
        _, err := eng.Store(ctx, sessionID, content, nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// 基准测试：检索性能
func BenchmarkRetrieve(b *testing.B) {
    store := memory.NewInMemoryStore()
    eng := memory.NewEngine(store, memory.DefaultOptions())
    
    ctx := context.Background()
    sessionID := "bench-session"
    
    // 预加载数据
    for i := 0; i < 1000; i++ {
        content := fmt.Sprintf("测试记忆 %d", i)
        eng.Store(ctx, sessionID, content, nil)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := eng.Retrieve(ctx, sessionID, "测试", 10)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// 基准测试：MMR 选择
func BenchmarkMMRSelect(b *testing.B) {
    records := make([]memory.MemoryRecord, 100)
    for i := range records {
        records[i] = memory.MemoryRecord{
            ID:            int64(i),
            Content:       fmt.Sprintf("记忆 %d", i),
            WeightedScore: float64(100-i) / 100,
        }
    }
    
    query := make([]float32, 768)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = mmrSelect(records, query, 10, 0.7)
    }
}
```

### 4.2 性能基线

#### 4.2.1 存储性能基线

| 操作 | 目标延迟 | 警告延迟 | 严重延迟 |
|------|----------|----------|----------|
| Store (P50) | < 50ms | 50-100ms | > 100ms |
| Store (P95) | < 200ms | 200-500ms | > 500ms |
| Store (P99) | < 500ms | 500-1000ms | > 1000ms |
| Batch Store (100 条) | < 2s | 2-5s | > 5s |

#### 4.2.2 检索性能基线

| 操作 | 目标延迟 | 警告延迟 | 严重延迟 |
|------|----------|----------|----------|
| Retrieve (P50) | < 30ms | 30-100ms | > 100ms |
| Retrieve (P95) | < 100ms | 100-200ms | > 200ms |
| Retrieve (P99) | < 200ms | 200-500ms | > 500ms |
| Batch Retrieve (10 查询) | < 500ms | 500ms-1s | > 1s |

### 4.3 性能回归检测

#### 4.3.1 自动化检测脚本

```bash
#!/bin/bash
# performance_regression_test.sh

# 运行基准测试
go test -bench=. -benchmem -count=5 ./src/memory/... > benchmark_results.txt

# 与基线比较
python3 compare_with_baseline.py \
    --baseline baseline_results.txt \
    --current benchmark_results.txt \
    --threshold 0.1  # 10% 回归阈值

# 输出报告
if [ $? -ne 0 ]; then
    echo "性能回归检测失败！"
    exit 1
else
    echo "性能回归检测通过"
    exit 0
fi
```

#### 4.3.2 CI/CD 集成

```yaml
# .github/workflows/performance-test.yml
name: Performance Test

on:
  pull_request:
    branches: [main]

jobs:
  performance:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.25
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -count=5 ./... > benchmark_results.txt
      
      - name: Compare with baseline
        run: |
          python3 scripts/compare_benchmarks.py \
            --baseline scripts/baseline.txt \
            --current benchmark_results.txt \
            --threshold 0.1
      
      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: benchmark-results
          path: benchmark_results.txt
```

---

## 5. 附录

### 5.1 调优检查清单

#### 调优前

- [ ] 明确业务目标和优先级
- [ ] 准备 20-50 个典型测试查询
- [ ] 标注每个查询的期望结果
- [ ] 确定评估指标（Precision/Recall/NDCG）
- [ ] 建立性能基线

#### 调优中

- [ ] 从默认参数开始
- [ ] 每次只调整 1-2 个参数
- [ ] 记录每次测试的结果
- [ ] 关注 NDCG 而非单一指标
- [ ] 进行统计显著性检验

#### 调优后

- [ ] 在小流量环境 A/B 测试
- [ ] 监控用户满意度
- [ ] 定期检查参数是否仍然最优
- [ ] 建立参数变更审批流程
- [ ] 更新文档和基线

### 5.2 常用命令

```bash
# 运行基准测试
go test -bench=. -benchmem ./src/memory/...

# 运行特定基准测试
go test -bench=BenchmarkRetrieve -benchmem ./src/memory/...

# 生成性能分析报告
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./src/memory/...

# 查看 CPU 分析报告
go tool pprof cpu.prof

# 查看内存分析报告
go tool pprof mem.prof
```

### 5.3 参考资源

- [MMR 算法论文](https://www.cs.cmu.edu/~jgc/publication/The_Use_of_MMR_Diversity_Based_LM.pdf)
- [HNSW 算法论文](https://arxiv.org/abs/1603.09320)
- [Qdrant 性能调优指南](https://qdrant.tech/documentation/guides/performance/)
- [内部监控文档](./01_monitoring_guide.md)

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
