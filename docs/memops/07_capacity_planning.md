# 记忆系统容量规划指南

**文档编号**: OPS-07  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐ 补充文档  
**适用角色**: 架构师 / SRE 工程师 / 技术负责人  

---

## 📋 目录

1. [容量评估](#1-容量评估)
2. [扩容策略](#2-扩容策略)
3. [成本优化](#3-成本优化)
4. [容量监控](#4-容量监控)
5. [附录](#5-附录)

---

## 1. 容量评估

### 1.1 存储容量评估

#### 1.1.1 存储需求计算

```
单条记忆存储大小 ≈ 内容大小 + 向量大小 + 元数据

假设:
- 平均内容大小：500 字节
- 向量大小：768 * 4 字节 (float32) = 3KB
- 元数据：1KB

单条记忆 ≈ 500 + 3072 + 1024 = 4.5KB
```

**存储容量公式**：

```
总存储容量 = 记忆数量 × 单条记忆大小 × (1 + 冗余系数)

其中:
- 冗余系数 = 0.3 (30% 用于索引和冗余)
```

**示例计算**：

| 记忆数量 | 存储容量 | 建议配置 |
|----------|----------|----------|
| 100 万 | 4.5GB × 1.3 ≈ 6GB | 10GB |
| 1000 万 | 45GB × 1.3 ≈ 59GB | 100GB |
| 1 亿 | 450GB × 1.3 ≈ 585GB | 1TB |

#### 1.1.2 存储增长预测

```sql
-- 查询存储增长趋势
SELECT 
    DATE(created_at) as date,
    COUNT(*) as daily_count,
    SUM(pg_column_size(content) + pg_column_size(embedding)) as daily_bytes
FROM memories
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

**增长预测公式**：

```
预计可用天数 = (总容量 - 已用容量) / 日均增长

示例:
- 总容量：100GB
- 已用容量：60GB
- 日均增长：500MB

预计可用天数 = (100 - 60)GB / 0.5GB/天 = 80 天
```

### 1.2 计算资源评估

#### 1.2.1 CPU 需求评估

**嵌入计算**：

```
单条嵌入计算 CPU 时间 ≈ 50ms (768 维向量)

并发嵌入需求 = QPS × 单条计算时间

示例:
- QPS: 100
- 单条计算时间：50ms

并发嵌入需求 = 100 × 0.05s = 5 CPU 核心
```

**检索计算**：

```
单次检索 CPU 时间 ≈ 10ms (向量相似度计算)

并发检索需求 = QPS × 单条计算时间

示例:
- QPS: 500
- 单条计算时间：10ms

并发检索需求 = 500 × 0.01s = 5 CPU 核心
```

#### 1.2.2 内存需求评估

**应用内存**：

```
单实例内存需求 = 基础内存 + 缓存内存 + 并发内存

假设:
- 基础内存：512MB
- 缓存内存：1GB
- 并发内存：100MB × 并发数

示例 (并发数=10):
单实例内存 = 512MB + 1GB + 100MB×10 = 2.5GB
```

**数据库内存**：

```
PostgreSQL 内存需求 = shared_buffers + work_mem × 连接数

推荐配置:
- shared_buffers: 总内存的 25%
- work_mem: 4-8MB
- 连接数：50-100

示例 (总内存 16GB):
shared_buffers = 16GB × 0.25 = 4GB
work_mem = 8MB
总需求 = 4GB + 8MB × 50 = 4.4GB
```

### 1.3 网络带宽评估

#### 1.3.1 带宽需求计算

```
带宽需求 = (请求大小 + 响应大小) × QPS

假设:
- 平均请求大小：1KB
- 平均响应大小：10KB
- QPS: 1000

带宽需求 = (1KB + 10KB) × 1000 = 11MB/s = 88Mbps
```

#### 1.3.2 跨区带宽

```
跨区域复制带宽 = 数据变更量 / 复制窗口

假设:
- 日均数据变更：1GB
- 复制窗口：1 小时

跨区带宽 = 1GB / 3600s = 289KB/s = 2.3Mbps
```

---

## 2. 扩容策略

### 2.1 垂直扩容

#### 2.1.1 扩容时机

| 指标 | 阈值 | 建议操作 |
|------|------|----------|
| CPU 使用率 | > 80% (持续 1 小时) | 增加 CPU 核心 |
| 内存使用率 | > 85% (持续 1 小时) | 增加内存 |
| 磁盘使用率 | > 80% | 增加磁盘空间 |
| 连接数 | > 80% 上限 | 增加连接池 |

#### 2.1.2 扩容方案

| 当前配置 | 扩容后配置 | 适用场景 |
|----------|------------|----------|
| 2 核 4GB | 4 核 8GB | 小规模增长 |
| 4 核 8GB | 8 核 16GB | 中等规模增长 |
| 8 核 16GB | 16 核 32GB | 大规模增长 |

### 2.2 水平扩容

#### 2.2.1 自动扩缩容配置

创建 `k8s/hpa.yaml`：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: go-agent-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: go-agent
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 20
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
```

#### 2.2.2 扩容触发条件

| 触发条件 | 扩容动作 | 扩容幅度 |
|----------|----------|----------|
| CPU > 70% | 增加实例 | +50% |
| 内存 > 80% | 增加实例 | +50% |
| QPS > 1000 | 增加实例 | +100% |
| 延迟 P95 > 200ms | 增加实例 | +50% |

### 2.3 数据库扩容

#### 2.3.1 PostgreSQL 扩容

**读扩容（只读副本）**：

```sql
-- 创建只读副本
SELECT pg_create_physical_replication_slot('replica1');

-- 配置从库
-- postgresql.conf
hot_standby = on
```

**写扩容（分区表）**：

```sql
-- 创建分区表
CREATE TABLE memories_partitioned (
    id BIGSERIAL,
    session_id VARCHAR(100),
    content TEXT,
    created_at TIMESTAMPTZ
) PARTITION BY RANGE (created_at);

-- 创建分区
CREATE TABLE memories_2026_q1 PARTITION OF memories_partitioned
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');
```

#### 2.3.2 Qdrant 扩容

**分片扩容**：

```bash
# 添加新节点到集群
curl -X POST http://qdrant-leader:6333/cluster/peer \
  -d '{"peer_id": 3}'

# 重新平衡分片
curl -X POST http://qdrant-leader:6333/collections/memory/cluster \
  -d '{"replica_shard": {"shard_id": 1, "peer_id": 3}}'
```

---

## 3. 成本优化

### 3.1 资源利用率优化

#### 3.1.1 资源请求与限制

```yaml
# Kubernetes 资源配置
resources:
  requests:
    cpu: "500m"      # 保证资源
    memory: "512Mi"
  limits:
    cpu: "1000m"     # 最大资源
    memory: "1Gi"
```

**优化建议**：
- 根据实际使用情况调整 requests
- 设置合理的 limits 防止资源耗尽
- 使用 VPA（Vertical Pod Autoscaler）自动调整

#### 3.1.2 闲时缩容

```yaml
# CronJob 定时缩容（夜间）
apiVersion: batch/v1
kind: CronJob
metadata:
  name: scale-down-night
spec:
  schedule: "0 22 * * *"  # 每晚 22:00
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: kubectl
            image: bitnami/kubectl
            command:
            - kubectl
            - scale
            - deployment/go-agent
            - --replicas=2
            - -n
            - memory-system
          restartPolicy: OnFailure
```

### 3.2 存储成本优化

#### 3.2.1 数据分层存储

```
┌─────────────────────────────────────────┐
│           数据分层存储                   │
├─────────────────────────────────────────┤
│                                         │
│  热数据 (最近 7 天)                      │
│  ├── 存储：SSD                          │
│  ├── 成本：高                           │
│  └── 访问：频繁                         │
│                                         │
│  温数据 (7-30 天)                        │
│  ├── 存储：HDD                          │
│  ├── 成本：中                           │
│  └── 访问：偶尔                         │
│                                         │
│  冷数据 (> 30 天)                        │
│  ├── 存储：对象存储                     │
│  ├── 成本：低                           │
│  └── 访问：很少                         │
│                                         │
└─────────────────────────────────────────┘
```

#### 3.2.2 数据压缩

```sql
-- 启用数据压缩
ALTER TABLE memories SET (fillfactor = 90);

-- 定期清理
VACUUM ANALYZE memories;

-- 删除过期数据
DELETE FROM memories
WHERE created_at < NOW() - INTERVAL '365 days';
```

### 3.3 成本分析

#### 3.3.1 成本构成

| 成本项 | 占比 | 优化空间 |
|--------|------|----------|
| 计算资源 | 40% | 中 |
| 存储资源 | 30% | 高 |
| 网络带宽 | 15% | 低 |
| 备份存储 | 10% | 中 |
| 其他 | 5% | - |

#### 3.3.2 成本优化建议

| 优化项 | 预计节省 | 实施难度 |
|--------|----------|----------|
| 闲时缩容 | 20% | 低 |
| 数据分层 | 15% | 中 |
| 预留实例 | 30% | 低 |
| 数据压缩 | 10% | 低 |
| 总计 | ~50% | - |

---

## 4. 容量监控

### 4.1 容量指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| 存储使用率 | 已用存储/总存储 | > 80% |
| CPU 使用率 | 已用 CPU/总 CPU | > 80% |
| 内存使用率 | 已用内存/总内存 | > 85% |
| 连接数使用率 | 已用连接/最大连接 | > 80% |
| 容量增长率 | 容量增长速度 | 突增 > 50% |

### 4.2 容量预测

```promql
# 预测 7 天后存储使用量
predict_linear(memory_total_size_gb[7d], 7*24*3600)

# 预测磁盘满的时间
(disk_total_bytes - disk_used_bytes) / rate(disk_used_bytes[7d])
```

### 4.3 容量报告

#### 4.3.1 日报

```markdown
## 容量日报

**日期**: 2026-03-09

### 存储容量

| 组件 | 已用 | 总量 | 使用率 | 日增长 |
|------|------|------|--------|--------|
| PostgreSQL | 50GB | 100GB | 50% | +0.5GB |
| Qdrant | 30GB | 100GB | 30% | +0.3GB |

### 计算资源

| 组件 | CPU | 内存 | 实例数 |
|------|-----|------|--------|
| go-agent | 45% | 60% | 3 |

### 容量预警

- 无
```

#### 4.3.2 周报

```markdown
## 容量周报

**周次**: 2026-W10

### 容量趋势

| 指标 | 周初 | 周末 | 变化 |
|------|------|------|------|
| 存储总量 | 75GB | 80GB | +6.7% |
| 平均 CPU | 40% | 45% | +12.5% |
| 平均内存 | 55% | 58% | +5.5% |

### 扩容计划

- 下周需要扩容：是/否
- 预计扩容时间：YYYY-MM-DD
```

---

## 5. 附录

### 5.1 容量规划计算器

```python
#!/usr/bin/env python3
# capacity_calculator.py

def calculate_storage(memory_count, avg_content_size=500, vector_dim=768):
    """计算存储容量需求"""
    vector_size = vector_dim * 4  # float32
    metadata_size = 1024  # 1KB
    single_size = avg_content_size + vector_size + metadata_size
    total_size = single_size * memory_count * 1.3  # 30% 冗余
    return total_size / (1024**3)  # 转换为 GB

def calculate_bandwidth(qps, req_size=1024, resp_size=10240):
    """计算带宽需求"""
    bandwidth = (req_size + resp_size) * qps
    return bandwidth / (1024*1024) * 8  # 转换为 Mbps

# 使用示例
storage_gb = calculate_storage(10_000_000)  # 1000 万条记忆
bandwidth_mbps = calculate_bandwidth(1000)  # 1000 QPS

print(f"存储需求：{storage_gb:.2f} GB")
print(f"带宽需求：{bandwidth_mbps:.2f} Mbps")
```

### 5.2 容量规划检查清单

#### 规划前

- [ ] 收集历史数据
- [ ] 分析增长趋势
- [ ] 评估业务需求
- [ ] 确定容量目标

#### 规划中

- [ ] 计算存储需求
- [ ] 计算计算资源
- [ ] 计算网络带宽
- [ ] 评估扩容方案
- [ ] 估算成本

#### 规划后

- [ ] 制定扩容计划
- [ ] 配置容量监控
- [ ] 建立预警机制
- [ ] 定期审查更新

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
