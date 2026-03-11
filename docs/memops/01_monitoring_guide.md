# 记忆系统监控指南

**文档编号**: OPS-01  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐⭐ 核心文档  
**适用角色**: SRE 工程师 / 运维工程师 / 值班工程师  
**所属目录**: docs/memops/  

---

## 📋 目录

1. [监控指标体系](#1-监控指标体系)
2. [监控工具配置](#2-监控工具配置)
3. [内存系统专项监控](#3-内存系统专项监控)
4. [告警分级与通知](#4-告警分级与通知)
5. [附录](#5-附录)

---

## 1. 监控指标体系

### 1.1 核心指标概览

记忆系统监控分为四个层面：

```
┌─────────────────────────────────────────────────────────┐
│              记忆系统健康度指标                          │
├─────────────────────────────────────────────────────────┤
│  存储层  │  检索层  │  质量层  │  资源层                │
│  - 总量  │  - 延迟  │  - 去重  │  - CPU                 │
│  - 增长  │  - 成功率 │  - 重嵌入 │  - 内存               │
│  - 修剪  │  - 空结果 │  - 摘要  │  - 磁盘               │
│  - 延迟  │  - 数量  │  - 满意度 │  - 网络               │
└─────────────────────────────────────────────────────────┘
```

### 1.2 存储层指标

| 指标名称 | 英文名称 | 类型 | 单位 | 说明 | 告警阈值 |
|----------|----------|------|------|------|----------|
| 存储总量 | `memory_total_count` | Gauge | 条 | 当前存储的记忆总数 | - |
| 存储大小 | `memory_total_size_gb` | Gauge | GB | 占用存储空间 | > 80% 容量 |
| 存储增长率 | `memory_growth_rate` | Gauge | 条/天 | 每日新增记忆数 | 突增>200% |
| 修剪率 | `memory_prune_rate` | Gauge | % | 被修剪的记忆占比 | > 30% |
| 存储延迟 P50 | `memory_store_latency_p50` | Histogram | ms | 存储操作延迟中位数 | > 100ms |
| 存储延迟 P95 | `memory_store_latency_p95` | Histogram | ms | 存储操作延迟 95 分位 | > 500ms |
| 存储延迟 P99 | `memory_store_latency_p99` | Histogram | ms | 存储操作延迟 99 分位 | > 1000ms |

**计算公式**：
```
修剪率 = (修剪数量 / 存储总量) × 100%
存储增长率 = (今日新增 - 昨日新增) / 昨日新增 × 100%
```

### 1.3 检索层指标

| 指标名称 | 英文名称 | 类型 | 单位 | 说明 | 告警阈值 |
|----------|----------|------|------|------|----------|
| 检索延迟 P50 | `memory_retrieve_latency_p50` | Histogram | ms | 检索操作延迟中位数 | > 50ms |
| 检索延迟 P95 | `memory_retrieve_latency_p95` | Histogram | ms | 检索操作延迟 95 分位 | > 200ms |
| 检索延迟 P99 | `memory_retrieve_latency_p99` | Histogram | ms | 检索操作延迟 99 分位 | > 500ms |
| 检索成功率 | `memory_retrieve_success_rate` | Gauge | % | 成功检索的占比 | < 95% |
| 平均返回数量 | `memory_avg_results_count` | Gauge | 条 | 平均每次检索返回数 | < 3 条 |
| 空结果率 | `memory_empty_result_rate` | Gauge | % | 返回空结果的占比 | > 50% |

**计算公式**：
```
检索成功率 = (成功检索次数 / 总检索次数) × 100%
空结果率 = (空结果次数 / 总检索次数) × 100%
```

### 1.4 质量层指标

| 指标名称 | 英文名称 | 类型 | 单位 | 说明 | 告警阈值 |
|----------|----------|------|------|------|----------|
| 去重率 | `memory_dedup_rate` | Gauge | % | 被去重的记忆占比 | > 20% |
| 重嵌入率 | `memory_reembed_rate` | Gauge | % | 被重新嵌入的记忆占比 | > 10% |
| 摘要覆盖率 | `memory_summary_coverage` | Gauge | % | 有摘要的记忆占比 | < 50% |
| 平均重要性 | `memory_avg_importance` | Gauge | 0-1 | 记忆平均重要性评分 | - |
| 用户满意度 | `memory_user_satisfaction` | Gauge | 1-5 分 | 用户对检索结果的评分 | < 3.5 |
| 健康评分 | `memory_health_score` | Gauge | 0-100 | 综合健康度评分 | < 60 |

**计算公式**：
```
去重率 = (去重数量 / 存储总量) × 100%
重嵌入率 = (重嵌入数量 / 存储总量) × 100%
摘要覆盖率 = (有摘要的记忆数 / 存储总量) × 100%
```

### 1.5 资源层指标

| 指标名称 | 英文名称 | 类型 | 单位 | 说明 | 告警阈值 |
|----------|----------|------|------|------|----------|
| CPU 使用率 | `node_cpu_usage_percent` | Gauge | % | 服务器 CPU 使用率 | > 80% |
| 内存使用率 | `node_memory_usage_percent` | Gauge | % | 服务器内存使用率 | > 85% |
| 磁盘使用率 | `node_disk_usage_percent` | Gauge | % | 服务器磁盘使用率 | > 80% |
| 网络带宽 | `node_network_bandwidth_mb` | Gauge | MB/s | 网络吞吐量 | > 80% 带宽 |
| 数据库连接数 | `postgres_connections` | Gauge | 个 | PostgreSQL 连接数 | > 80% 上限 |
| Qdrant 连接数 | `qdrant_connections` | Gauge | 个 | Qdrant 连接数 | > 80% 上限 |

---

## 2. 监控工具配置

### 2.1 Prometheus 指标导出

#### 2.1.1 安装 Prometheus 客户端

```bash
cd /Users/frank/MareCogito/go-agent
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
```

#### 2.1.2 指标导出代码

创建 `src/memory/monitor/metrics.go`：

```go
package monitor

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // ========== 存储层指标 ==========
    
    // memory_total_count - 存储总量
    MemoryTotalCount = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_total_count",
        Help: "Total number of memories stored",
    })
    
    // memory_total_size_gb - 存储大小
    MemoryTotalSizeGB = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_total_size_gb",
        Help: "Total storage size in GB",
    })
    
    // memory_store_latency_seconds - 存储延迟
    MemoryStoreLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "memory_store_latency_seconds",
        Help:    "Latency of store operations in seconds",
        Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
    })
    
    // memory_prune_rate - 修剪率
    MemoryPruneRate = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_prune_rate",
        Help: "Percentage of pruned memories",
    })
    
    // ========== 检索层指标 ==========
    
    // memory_retrieve_latency_seconds - 检索延迟
    MemoryRetrieveLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "memory_retrieve_latency_seconds",
        Help:    "Latency of retrieve operations in seconds",
        Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
    })
    
    // memory_retrieve_success_rate - 检索成功率
    MemoryRetrieveSuccessRate = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_retrieve_success_rate",
        Help: "Success rate of retrieve operations (percentage)",
    })
    
    // memory_empty_result_rate - 空结果率
    MemoryEmptyResultRate = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_empty_result_rate",
        Help: "Rate of empty result retrievals (percentage)",
    })
    
    // memory_avg_results_count - 平均返回数量
    MemoryAvgResultsCount = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_avg_results_count",
        Help: "Average number of results per retrieve operation",
    })
    
    // ========== 质量层指标 ==========
    
    // memory_dedup_rate - 去重率
    MemoryDedupRate = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_dedup_rate",
        Help: "Rate of deduplicated memories (percentage)",
    })
    
    // memory_reembed_rate - 重嵌入率
    MemoryReembedRate = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_reembed_rate",
        Help: "Rate of re-embedded memories (percentage)",
    })
    
    // memory_summary_coverage - 摘要覆盖率
    MemorySummaryCoverage = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_summary_coverage",
        Help: "Coverage of summarized memories (percentage)",
    })
    
    // memory_health_score - 健康评分
    MemoryHealthScore = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_health_score",
        Help: "Overall health score (0-100)",
    })
    
    // ========== MMR 参数指标 ==========
    
    // memory_mmr_lambda - MMR Lambda 参数
    MemoryMMRLambda = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_mmr_lambda",
        Help: "Current MMR lambda parameter value",
    })
    
    // memory_mmr_similarity_weight - 相似度权重
    MemoryMMRSimilarityWeight = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_mmr_similarity_weight",
        Help: "Current MMR similarity weight",
    })
    
    // memory_mmr_keywords_weight - 关键词权重
    MemoryMMRKeywordsWeight = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "memory_mmr_keywords_weight",
        Help: "Current MMR keywords weight",
    })
)
```

#### 2.1.3 指标采集器

创建 `src/memory/monitor/collector.go`：

```go
package monitor

import (
    "context"
    "time"
    
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/memory/engine"
)

// Collector 负责采集记忆系统指标
type Collector struct {
    engine      *memory.Engine
    interval    time.Duration
    stopChan    chan struct{}
    latencyHist *LatencyHistogram
}

// LatencyHistogram 延迟直方图
type LatencyHistogram struct {
    storeLatencies   []float64
    retrieveLatencies []float64
}

// NewCollector 创建指标采集器
func NewCollector(eng *memory.Engine, interval time.Duration) *Collector {
    return &Collector{
        engine:   eng,
        interval: interval,
        stopChan: make(chan struct{}),
        latencyHist: &LatencyHistogram{
            storeLatencies:   make([]float64, 0, 1000),
            retrieveLatencies: make([]float64, 0, 1000),
        },
    }
}

// Start 启动指标采集
func (c *Collector) Start(ctx context.Context) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            c.collectMetrics(ctx)
        case <-ctx.Done():
            return
        case <-c.stopChan:
            return
        }
    }
}

// Stop 停止指标采集
func (c *Collector) Stop() {
    close(c.stopChan)
}

// collectMetrics 采集指标
func (c *Collector) collectMetrics(ctx context.Context) {
    metrics := c.engine.MetricsSnapshot()
    
    // 存储层指标
    MemoryTotalCount.Set(float64(metrics.Stored - metrics.Pruned))
    MemoryPruneRate.Set(float64(metrics.Pruned) / float64(metrics.Stored) * 100)
    
    // 质量层指标
    MemoryDedupRate.Set(float64(metrics.Deduplicated) / float64(metrics.Stored) * 100)
    MemoryReembedRate.Set(float64(metrics.Reembedded) / float64(metrics.Stored) * 100)
    MemorySummaryCoverage.Set(float64(metrics.ClustersSummarized) / float64(metrics.Stored) * 100)
    
    // 计算健康评分
    healthScore := c.calculateHealthScore(metrics)
    MemoryHealthScore.Set(healthScore)
}

// calculateHealthScore 计算健康评分
func (c *Collector) calculateHealthScore(metrics engine.MetricsSnapshot) float64 {
    score := 100.0
    
    // 修剪率过高扣分
    pruneRate := float64(metrics.Pruned) / float64(metrics.Stored) * 100
    if pruneRate > 30 {
        score -= 20
    } else if pruneRate > 20 {
        score -= 10
    }
    
    // 去重率过高扣分
    dedupRate := float64(metrics.Deduplicated) / float64(metrics.Stored) * 100
    if dedupRate > 20 {
        score -= 15
    } else if dedupRate > 10 {
        score -= 5
    }
    
    // 重嵌入率异常扣分
    reembedRate := float64(metrics.Reembedded) / float64(metrics.Stored) * 100
    if reembedRate > 15 {
        score -= 10
    }
    
    return score
}

// RecordStoreLatency 记录存储延迟
func (c *Collector) RecordStoreLatency(latencySeconds float64) {
    MemoryStoreLatency.Observe(latencySeconds)
}

// RecordRetrieveLatency 记录检索延迟
func (c *Collector) RecordRetrieveLatency(latencySeconds float64) {
    MemoryRetrieveLatency.Observe(latencySeconds)
}
```

#### 2.1.4 HTTP 暴露端点

创建 `src/memory/monitor/http_handler.go`：

```go
package monitor

import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterMetricsHandler 注册 Prometheus 指标处理函数
func RegisterMetricsHandler(mux *http.ServeMux) {
    // Prometheus 指标端点
    mux.Handle("/metrics", promhttp.Handler())
    
    // 健康检查端点
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    // 就绪检查端点
    mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    })
}
```

### 2.2 Prometheus 配置

#### 2.2.1 prometheus.yml

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts.yml"

scrape_configs:
  - job_name: 'go-agent-memory'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 10s
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['localhost:9187']
      
  - job_name: 'qdrant'
    static_configs:
      - targets: ['localhost:6333']

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

#### 2.2.2 alerts.yml 告警规则

```yaml
groups:
  - name: memory_system_alerts
    rules:
      # ========== P0 严重告警 ==========
      
      - alert: MemoryHealthCritical
        expr: memory_health_score < 50
        for: 5m
        labels:
          severity: critical
          team: sre
        annotations:
          summary: "记忆系统健康度严重下降"
          description: "健康度评分 {{ $value }}，低于 50 分"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-health-critical"
      
      - alert: MemoryRetrieveFailureRateHigh
        expr: 100 - memory_retrieve_success_rate > 20
        for: 5m
        labels:
          severity: critical
          team: sre
        annotations:
          summary: "记忆检索失败率过高"
          description: "检索失败率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-retrieve-failure"
      
      # ========== P1 高优先级告警 ==========
      
      - alert: MemoryHighEmptyRate
        expr: memory_empty_result_rate > 50
        for: 10m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "记忆检索空结果率过高"
          description: "空结果率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-empty-rate"
      
      - alert: MemoryStoreLatencyHigh
        expr: histogram_quantile(0.95, memory_store_latency_seconds_bucket) > 1
        for: 5m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "记忆存储延迟过高"
          description: "P95 延迟 {{ $value }}秒"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-latency"
      
      - alert: MemoryHighDedupRate
        expr: memory_dedup_rate > 20
        for: 30m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "记忆重复率过高"
          description: "重复率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-dedup"
      
      # ========== P2 中优先级告警 ==========
      
      - alert: MemoryPruneRateHigh
        expr: memory_prune_rate > 30
        for: 1h
        labels:
          severity: warning
          team: sre
        annotations:
          summary: "记忆修剪率过高"
          description: "修剪率 {{ $value }}%，可能需要增加存储容量"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-prune"
      
      - alert: MemoryLowSummaryCoverage
        expr: memory_summary_coverage < 50
        for: 2h
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "记忆摘要覆盖率低"
          description: "摘要覆盖率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/memory-summary"
      
      # ========== 资源告警 ==========
      
      - alert: MemoryDiskUsageHigh
        expr: node_disk_usage_percent > 80
        for: 10m
        labels:
          severity: warning
          team: sre
        annotations:
          summary: "磁盘使用率过高"
          description: "磁盘使用率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/disk-usage"
      
      - alert: MemoryCPUUsageHigh
        expr: node_cpu_usage_percent > 80
        for: 10m
        labels:
          severity: warning
          team: sre
        annotations:
          summary: "CPU 使用率过高"
          description: "CPU 使用率 {{ $value }}%"
          runbook_url: "https://wiki.marecogito.ai/runbooks/cpu-usage"
```

### 2.3 Grafana 仪表板

#### 2.3.1 仪表板 JSON 配置

创建 `monitoring/grafana/dashboards/memory-system.json`：

```json
{
  "dashboard": {
    "id": null,
    "title": "记忆系统健康度",
    "tags": ["memory", "go-agent"],
    "timezone": "browser",
    "schemaVersion": 16,
    "version": 0,
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "健康评分",
        "type": "gauge",
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0},
        "targets": [
          {
            "expr": "memory_health_score",
            "legendFormat": "Health Score"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "min": 0,
            "max": 100,
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {"value": 0, "color": "red"},
                {"value": 60, "color": "yellow"},
                {"value": 80, "color": "green"}
              ]
            }
          }
        }
      },
      {
        "id": 2,
        "title": "存储总量",
        "type": "stat",
        "gridPos": {"h": 8, "w": 6, "x": 6, "y": 0},
        "targets": [
          {
            "expr": "memory_total_count",
            "legendFormat": "Total Memories"
          }
        ]
      },
      {
        "id": 3,
        "title": "检索成功率",
        "type": "gauge",
        "gridPos": {"h": 8, "w": 6, "x": 12, "y": 0},
        "targets": [
          {
            "expr": "memory_retrieve_success_rate",
            "legendFormat": "Success Rate"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "min": 0,
            "max": 100,
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {"value": 0, "color": "red"},
                {"value": 90, "color": "yellow"},
                {"value": 95, "color": "green"}
              ]
            }
          }
        }
      },
      {
        "id": 4,
        "title": "空结果率",
        "type": "gauge",
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0},
        "targets": [
          {
            "expr": "memory_empty_result_rate",
            "legendFormat": "Empty Rate"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "min": 0,
            "max": 100,
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {"value": 0, "color": "green"},
                {"value": 30, "color": "yellow"},
                {"value": 50, "color": "red"}
              ]
            }
          }
        }
      },
      {
        "id": 5,
        "title": "检索延迟 (P50/P95/P99)",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(memory_retrieve_latency_seconds_bucket[5m]))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(memory_retrieve_latency_seconds_bucket[5m]))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(memory_retrieve_latency_seconds_bucket[5m]))",
            "legendFormat": "P99"
          }
        ],
        "yaxes": [
          {"format": "s", "label": "Latency"},
          {"show": false}
        ]
      },
      {
        "id": 6,
        "title": "存储延迟 (P50/P95/P99)",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8},
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(memory_store_latency_seconds_bucket[5m]))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(memory_store_latency_seconds_bucket[5m]))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(memory_store_latency_seconds_bucket[5m]))",
            "legendFormat": "P99"
          }
        ],
        "yaxes": [
          {"format": "s", "label": "Latency"},
          {"show": false}
        ]
      },
      {
        "id": 7,
        "title": "去重率 & 修剪率",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 16},
        "targets": [
          {
            "expr": "memory_dedup_rate",
            "legendFormat": "Dedup Rate"
          },
          {
            "expr": "memory_prune_rate",
            "legendFormat": "Prune Rate"
          }
        ],
        "yaxes": [
          {"format": "percent", "label": "Rate"},
          {"show": false}
        ]
      },
      {
        "id": 8,
        "title": "摘要覆盖率 & 重嵌入率",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 16},
        "targets": [
          {
            "expr": "memory_summary_coverage",
            "legendFormat": "Summary Coverage"
          },
          {
            "expr": "memory_reembed_rate",
            "legendFormat": "Re-embed Rate"
          }
        ],
        "yaxes": [
          {"format": "percent", "label": "Rate"},
          {"show": false}
        ]
      }
    ]
  }
}
```

---

## 3. 内存系统专项监控

### 3.1 MMR 参数监控

#### 3.1.1 参数监控指标

| 参数 | 指标名称 | 说明 | 推荐范围 |
|------|----------|------|----------|
| Lambda | `memory_mmr_lambda` | MMR 平衡参数 | 0.5-0.9 |
| 相似度权重 | `memory_mmr_similarity_weight` | 向量相似度权重 | 0.3-0.6 |
| 关键词权重 | `memory_mmr_keywords_weight` | 关键词匹配权重 | 0.1-0.3 |
| 重要性权重 | `memory_mmr_importance_weight` | 重要性权重 | 0.1-0.3 |
| 新鲜度权重 | `memory_mmr_recency_weight` | 新鲜度权重 | 0.05-0.2 |

#### 3.1.2 参数变更记录表

创建 `monitoring/mmr_params_log.sql`：

```sql
CREATE TABLE mmr_params_log (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lambda FLOAT NOT NULL,
    similarity_weight FLOAT NOT NULL,
    keywords_weight FLOAT NOT NULL,
    importance_weight FLOAT NOT NULL,
    recency_weight FLOAT NOT NULL,
    changed_by VARCHAR(100),
    reason TEXT,
    before_ndcg FLOAT,
    after_ndcg FLOAT
);

-- 查询参数变更历史
SELECT 
    timestamp,
    lambda,
    similarity_weight,
    keywords_weight,
    changed_by,
    reason,
    ROUND((after_ndcg - before_ndcg)::numeric, 4) as ndcg_delta
FROM mmr_params_log
ORDER BY timestamp DESC
LIMIT 50;
```

### 3.2 健康度评分算法

#### 3.2.1 评分公式

```go
// 健康评分计算（满分 100 分）
func calculateHealthScore(metrics MetricsSnapshot) float64 {
    score := 100.0
    
    // 存储健康 (25 分)
    pruneRate := float64(metrics.Pruned) / float64(metrics.Stored) * 100
    if pruneRate > 30 {
        score -= 20
    } else if pruneRate > 20 {
        score -= 10
    }
    
    // 检索健康 (25 分)
    emptyRate := metrics.EmptyResultRate
    if emptyRate > 50 {
        score -= 15
    } else if emptyRate > 30 {
        score -= 10
    }
    
    successRate := metrics.SuccessRate
    if successRate < 95 {
        score -= 10
    }
    
    // 质量健康 (25 分)
    dedupRate := float64(metrics.Deduplicated) / float64(metrics.Stored) * 100
    if dedupRate > 20 {
        score -= 15
    } else if dedupRate > 10 {
        score -= 5
    }
    
    // 资源健康 (25 分)
    cpuUsage := getCPUUsage()
    if cpuUsage > 90 {
        score -= 15
    } else if cpuUsage > 80 {
        score -= 10
    }
    
    diskUsage := getDiskUsage()
    if diskUsage > 90 {
        score -= 10
    } else if diskUsage > 80 {
        score -= 5
    }
    
    return score
}
```

#### 3.2.2 健康度等级

| 分数范围 | 等级 | 颜色 | 说明 | 行动建议 |
|----------|------|------|------|----------|
| 90-100 | 优秀 | 🟢 绿色 | 系统运行良好 | 持续监控 |
| 80-89 | 良好 | 🟢 绿色 | 系统正常 | 定期巡检 |
| 60-79 | 一般 | 🟡 黄色 | 需要注意 | 分析原因 |
| 40-59 | 较差 | 🟠 橙色 | 需要干预 | 制定优化计划 |
| 0-39 | 严重 | 🔴 红色 | 紧急状态 | 立即处理 |

### 3.3 漂移检测监控

#### 3.3.1 漂移检测指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| 平均漂移率 | 所有记忆的平均漂移程度 | > 0.15 |
| 高漂移记忆数 | 漂移率 > 0.2 的记忆数量 | > 1000 |
| 重嵌入队列长度 | 等待重嵌入的记忆数量 | > 500 |
| 重嵌入失败率 | 重嵌入失败的比例 | > 5% |

#### 3.3.2 漂移监控查询

```sql
-- 查询漂移严重的记忆
SELECT 
    id,
    content,
    last_embedded,
    drift_score
FROM memories
WHERE drift_score > 0.2
ORDER BY drift_score DESC
LIMIT 100;

-- 查询漂移趋势
SELECT 
    DATE(last_embedded) as date,
    COUNT(*) as count,
    AVG(drift_score) as avg_drift
FROM memories
WHERE last_embedded > NOW() - INTERVAL '30 days'
GROUP BY DATE(last_embedded)
ORDER BY date DESC;
```

---

## 4. 告警分级与通知

### 4.1 告警分级标准

| 级别 | 名称 | 响应时间 | 升级时间 | 通知渠道 |
|------|------|----------|----------|----------|
| **P0** | 严重 | 5 分钟 | 15 分钟 | 电话 + 短信 + 邮件 + IM |
| **P1** | 高 | 15 分钟 | 30 分钟 | 短信 + 邮件 + IM |
| **P2** | 中 | 30 分钟 | 1 小时 | 邮件 + IM |
| **P3** | 低 | 2 小时 | 4 小时 | IM |

### 4.2 通知渠道配置

#### 4.2.1 Alertmanager 配置

创建 `monitoring/alertmanager/alertmanager.yml`：

```yaml
global:
  smtp_smarthost: 'smtp.marecogito.ai:587'
  smtp_from: 'alertmanager@marecogito.ai'
  smtp_auth_username: 'alertmanager@marecogito.ai'
  smtp_auth_password: '${SMTP_PASSWORD}'

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default-receiver'
  routes:
    - match:
        severity: critical
      receiver: 'critical-receiver'
      group_wait: 10s
      repeat_interval: 1h
    - match:
        severity: warning
      receiver: 'warning-receiver'
      group_wait: 30s
      repeat_interval: 4h

receivers:
  - name: 'default-receiver'
    email_configs:
      - to: 'team@marecogito.ai'
        send_resolved: true

  - name: 'critical-receiver'
    email_configs:
      - to: 'sre-team@marecogito.ai'
        send_resolved: true
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts-critical'
        send_resolved: true
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        send_resolved: true

  - name: 'warning-receiver'
    email_configs:
      - to: 'backend-team@marecogito.ai'
        send_resolved: true
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts-warning'
        send_resolved: true

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

### 4.3 告警升级流程

```
告警触发
    │
    ▼
通知值班工程师 (P0:5 分钟 / P1:15 分钟 / P2:30 分钟)
    │
    ├─ 已响应 → 处理告警
    │
    └─ 未响应 (超时)
        │
        ▼
    通知备份负责人 (P0:15 分钟 / P1:30 分钟 / P2:1 小时)
        │
        ├─ 已响应 → 处理告警
        │
        └─ 未响应 (超时)
            │
            ▼
        通知技术负责人
            │
            ▼
        启动应急会议
```

### 4.4 告警静默规则

以下情况可设置告警静默：

| 场景 | 静默时长 | 审批人 |
|------|----------|--------|
| 计划内维护 | 维护窗口期 + 30 分钟 | 技术负责人 |
| 已知问题 | 至问题解决 | 技术负责人 |
| 测试环境 | 长期 | 团队负责人 |
| 误报 | 至规则修复 | SRE 负责人 |

---

## 5. 附录

### 5.1 监控检查清单

#### 日常检查

- [ ] 健康评分 > 80
- [ ] 无 P0/P1 告警
- [ ] 检索成功率 > 95%
- [ ] 空结果率 < 30%
- [ ] P95 延迟 < 200ms

#### 周检查

- [ ] 查看告警历史
- [ ] 分析性能趋势
- [ ] 检查容量使用
- [ ] 审查 MMR 参数

#### 月检查

- [ ] 健康度趋势分析
- [ ] 参数调优效果评估
- [ ] 监控规则优化
- [ ] 仪表板更新

### 5.2 常用查询

#### Prometheus 查询示例

```promql
# 查询健康评分趋势
memory_health_score

# 查询检索延迟 P95
histogram_quantile(0.95, rate(memory_retrieve_latency_seconds_bucket[5m]))

# 查询存储增长率
rate(memory_total_count[1d]) * 86400

# 查询空结果率趋势
memory_empty_result_rate

# 查询 MMR 参数变更历史
memory_mmr_lambda
```

#### SQL 查询示例

```sql
-- 查询最近 24 小时存储量
SELECT COUNT(*) as count
FROM memories
WHERE created_at > NOW() - INTERVAL '24 hours';

-- 查询检索失败 TOP10 会话
SELECT session_id, COUNT(*) as fail_count
FROM retrieve_logs
WHERE success = false
  AND timestamp > NOW() - INTERVAL '7 days'
GROUP BY session_id
ORDER BY fail_count DESC
LIMIT 10;

-- 查询平均延迟趋势
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    AVG(latency_ms) as avg_latency
FROM retrieve_logs
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', timestamp)
ORDER BY hour;
```

### 5.3 故障排查流程图

```
收到告警
    │
    ▼
确认告警级别
    │
    ├─ P0 → 立即响应，通知团队
    │
    ├─ P1 → 15 分钟内响应
    │
    └─ P2/P3 → 按计划处理
    
查看 Grafana 仪表板
    │
    ▼
定位问题组件
    │
    ├─ 存储层 → 检查数据库连接、磁盘空间
    │
    ├─ 检索层 → 检查嵌入服务、向量数据库
    │
    ├─ 质量层 → 检查去重逻辑、参数配置
    │
    └─ 资源层 → 检查 CPU、内存、网络
    
查看详细日志
    │
    ▼
执行诊断命令
    │
    ▼
实施修复方案
    │
    ▼
验证修复效果
    │
    ▼
记录故障报告
```

### 5.4 参考资源

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [Alertmanager 配置指南](https://prometheus.io/docs/alerting/latest/configuration/)
- [内部故障排查文档](./05_troubleshooting.md)

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
