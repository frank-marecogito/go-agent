# 记忆系统故障排查指南

**文档编号**: OPS-05  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐⭐ 核心文档  
**适用角色**: 值班工程师 / SRE 工程师 / 后端工程师  

---

## 📋 目录

1. [故障诊断流程](#1-故障诊断流程)
2. [常见问题排查](#2-常见问题排查)
3. [日志分析](#3-日志分析)
4. [故障案例库](#4-故障案例库)
5. [附录](#5-附录)

---

## 1. 故障诊断流程

### 1.1 故障分级标准

| 级别 | 名称 | 定义 | 响应时间 | 升级时间 |
|------|------|------|----------|----------|
| **P0** | 严重 | 系统完全不可用，数据丢失 | 5 分钟 | 15 分钟 |
| **P1** | 高 | 核心功能受损，性能严重下降 | 15 分钟 | 30 分钟 |
| **P2** | 中 | 部分功能受损，性能下降 | 30 分钟 | 1 小时 |
| **P3** | 低 | 轻微问题，不影响核心功能 | 2 小时 | 4 小时 |

### 1.2 故障诊断流程图

```
收到故障报告
    │
    ▼
确认故障级别
    │
    ├─ P0 → 立即响应，通知团队，启动应急会议
    │
    ├─ P1 → 15 分钟内响应，通知备份负责人
    │
    ├─ P2 → 30 分钟内响应，记录故障
    │
    └─ P3 → 2 小时内响应，加入待办

查看 Grafana 仪表板
    │
    ▼
定位问题组件
    │
    ├─ 存储层 → 检查数据库、向量数据库
    │
    ├─ 检索层 → 检查嵌入服务、检索逻辑
    │
    ├─ 质量层 → 检查去重、修剪逻辑
    │
    └─ 资源层 → 检查 CPU、内存、磁盘

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

### 1.3 信息收集清单

#### 1.3.1 故障基本信息

- [ ] 故障发生时间
- [ ] 故障发现方式（监控告警/用户报告/主动发现）
- [ ] 影响范围（全部用户/部分用户/特定功能）
- [ ] 故障现象（错误信息、截图、日志）

#### 1.3.2 技术信息

- [ ] 相关服务状态
- [ ] 监控指标异常
- [ ] 错误日志内容
- [ ] 最近的变更（代码/配置/基础设施）

#### 1.3.3 时间线记录

| 时间 | 事件 | 负责人 |
|------|------|--------|
| HH:MM | 故障发生 | - |
| HH:MM | 故障发现 | XXX |
| HH:MM | 开始排查 | XXX |
| HH:MM | 定位原因 | XXX |
| HH:MM | 实施修复 | XXX |
| HH:MM | 验证通过 | XXX |
| HH:MM | 故障恢复 | XXX |

---

## 2. 常见问题排查

### 2.1 记忆检索失败

#### 2.1.1 故障现象

- 用户报告检索返回空结果
- 监控显示空结果率飙升
- 检索延迟异常高

#### 2.1.2 排查步骤

```
1. 检查检索服务状态
   │
   ▼
2. 检查嵌入服务是否正常
   │
   ▼
3. 检查向量数据库连接
   │
   ▼
4. 检查检索参数配置
   │
   ▼
5. 检查索引状态
```

#### 2.1.3 诊断命令

```bash
# 1. 检查服务状态
kubectl get pods -l app=go-agent
kubectl get svc go-agent

# 2. 检查嵌入服务
curl -X POST http://localhost:8080/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "test"}'

# 3. 检查向量数据库连接
kubectl exec -it go-agent-pod -- nc -zv qdrant 6333

# 4. 检查检索日志
kubectl logs -l app=go-agent | grep "retrieve" | tail -50

# 5. 检查 Qdrant 索引
curl http://qdrant:6333/collections/memory
```

#### 2.1.4 常见原因及解决方案

| 原因 | 症状 | 解决方案 |
|------|------|----------|
| 嵌入服务超时 | 检索延迟高，超时错误 | 重启嵌入服务，增加超时时间 |
| 向量数据库连接失败 | 连接错误，检索失败 | 检查网络，重启 Qdrant |
| 索引损坏 | 检索结果为空 | 重建索引 |
| 参数配置错误 | 检索结果异常 | 检查 MMR 参数配置 |

### 2.2 记忆存储失败

#### 2.2.1 故障现象

- 存储操作返回错误
- 存储延迟异常高
- 存储成功率下降

#### 2.2.2 排查步骤

```
1. 检查数据库连接
   │
   ▼
2. 检查磁盘空间
   │
   ▼
3. 检查嵌入服务
   │
   ▼
4. 检查存储逻辑
```

#### 2.2.3 诊断命令

```bash
# 1. 检查 PostgreSQL 连接
kubectl exec -it go-agent-pod -- psql -h postgres -U user -d memory -c "SELECT 1"

# 2. 检查磁盘空间
kubectl exec -it postgres-pod -- df -h

# 3. 检查嵌入服务
curl -X POST http://localhost:8080/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "test"}'

# 4. 检查存储日志
kubectl logs -l app=go-agent | grep "store" | grep "ERROR" | tail -50
```

#### 2.2.4 常见原因及解决方案

| 原因 | 症状 | 解决方案 |
|------|------|----------|
| 数据库连接池耗尽 | 连接超时错误 | 增加连接池大小，检查连接泄漏 |
| 磁盘空间不足 | 写入失败错误 | 清理空间，扩容磁盘 |
| 嵌入服务失败 | 嵌入错误 | 重启嵌入服务，检查 API 密钥 |
| 事务锁等待 | 存储延迟高 | 分析慢查询，优化事务 |

### 2.3 性能下降

#### 2.3.1 故障现象

- 检索延迟升高
- 存储延迟升高
- 用户报告系统变慢

#### 2.3.2 排查步骤

```
1. 查看监控指标
   │
   ▼
2. 检查资源使用率
   │
   ▼
3. 检查慢查询
   │
   ▼
4. 检查并发量
   │
   ▼
5. 检查最近变更
```

#### 2.3.3 诊断命令

```bash
# 1. 查看监控指标
# Grafana: memory-system 仪表板

# 2. 检查资源使用率
kubectl top pods -l app=go-agent
kubectl top nodes

# 3. 检查慢查询
psql -h postgres -U user -d memory -c "
  SELECT query, calls, total_time, mean_time
  FROM pg_stat_statements
  ORDER BY mean_time DESC
  LIMIT 10;
"

# 4. 检查并发量
kubectl get hpa go-agent

# 5. 检查最近变更
kubectl rollout history deployment/go-agent
```

#### 2.3.4 常见原因及解决方案

| 原因 | 症状 | 解决方案 |
|------|------|----------|
| 资源不足 | CPU/内存使用率高 | 扩容，优化资源使用 |
| 慢查询 | 数据库响应慢 | 优化查询，添加索引 |
| 并发过高 | 请求队列堆积 | 限流，增加实例 |
| 索引效率低 | 检索延迟高 | 优化索引参数 |

### 2.4 数据不一致

#### 2.4.1 故障现象

- 检索结果与预期不符
- 记忆内容错乱
- 元数据丢失

#### 2.4.2 排查步骤

```
1. 确认数据不一致范围
   │
   ▼
2. 检查最近写入操作
   │
   ▼
3. 检查备份数据
   │
   ▼
4. 检查数据同步状态
```

#### 2.4.3 诊断命令

```bash
# 1. 检查数据一致性
psql -h postgres -U user -d memory -c "
  SELECT COUNT(*) FROM memories;
"

# 2. 检查最近写入
psql -h postgres -U user -d memory -c "
  SELECT id, content, created_at
  FROM memories
  ORDER BY created_at DESC
  LIMIT 10;
"

# 3. 检查备份
ls -lh /backup/memory/

# 4. 对比备份数据
pg_restore --list /backup/memory/latest.dump
```

#### 2.4.4 常见原因及解决方案

| 原因 | 症状 | 解决方案 |
|------|------|----------|
| 写入失败 | 数据丢失 | 检查写入逻辑，重试机制 |
| 同步延迟 | 数据不一致 | 等待同步，检查同步状态 |
| 缓存污染 | 检索结果错误 | 清除缓存，检查缓存逻辑 |
| 人为错误 | 数据错乱 | 从备份恢复 |

---

## 3. 日志分析

### 3.1 日志级别说明

| 级别 | 说明 | 处理建议 |
|------|------|----------|
| ERROR | 错误，需要立即处理 | 立即响应 |
| WARN | 警告，可能有问题 | 关注趋势 |
| INFO | 信息，正常运行日志 | 定期审查 |
| DEBUG | 调试，详细日志 | 排查问题时使用 |

### 3.2 日志收集

#### 3.2.1 Kubernetes 环境

```bash
# 查看最近 100 行日志
kubectl logs -l app=go-agent --tail=100

# 查看最近 1 小时日志
kubectl logs -l app=go-agent --since=1h

# 查看错误日志
kubectl logs -l app=go-agent | grep ERROR

# 实时查看日志
kubectl logs -l app=go-agent -f

# 查看上一个实例的日志（崩溃后）
kubectl logs -l app=go-agent --previous
```

#### 3.2.2 本地环境

```bash
# 查看日志文件
tail -f /var/log/go-agent/*.log

# 搜索错误
grep ERROR /var/log/go-agent/*.log | tail -50

# 按时间过滤
grep "2026-03-09 10:" /var/log/go-agent/*.log
```

### 3.3 常见错误模式

#### 3.3.1 数据库连接错误

```
ERROR: connection to database failed: FATAL: too many connections
```

**原因**：连接池耗尽  
**解决方案**：
```bash
# 检查当前连接数
psql -c "SELECT count(*) FROM pg_stat_activity;"

# 增加最大连接数
psql -c "ALTER SYSTEM SET max_connections = 200;"
psql -c "SELECT pg_reload_conf();"
```

#### 3.3.2 嵌入服务超时

```
ERROR: embed service timeout: context deadline exceeded
```

**原因**：嵌入服务响应慢或不可用  
**解决方案**：
```bash
# 检查嵌入服务状态
curl http://embed-service:8080/health

# 增加超时时间
# 修改配置：embed_timeout: 30s → 60s
```

#### 3.3.3 向量数据库错误

```
ERROR: qdrant request failed: connection refused
```

**原因**：Qdrant 服务不可用  
**解决方案**：
```bash
# 检查 Qdrant 状态
kubectl get pods -l app=qdrant
kubectl logs -l app=qdrant

# 重启 Qdrant
kubectl rollout restart statefulset/qdrant
```

### 3.4 日志分析工具

#### 3.4.1 Loki + Grafana

```promql
# 查询错误日志
{app="go-agent"} |= "ERROR"

# 统计错误数量
sum(rate({app="go-agent"} |= "ERROR"[5m]))

# 查看错误趋势
sum_over_time({app="go-agent"} |= "ERROR"[1h])
```

#### 3.4.2 ELK Stack

```
# 查询错误日志
service: go-agent AND level: ERROR

# 统计错误类型
level: ERROR AND message: * | stats count() by message
```

---

## 4. 故障案例库

### 4.1 案例 1：检索延迟飙升

**故障信息**：
- **日期**: 2026-03-05
- **级别**: P1
- **影响时长**: 30 分钟
- **现象**: 检索延迟 P95 从 100ms 升至 2s

**排查过程**：
1. 查看监控发现检索延迟飙升
2. 检查资源使用率正常
3. 检查慢查询发现大量全表扫描
4. 发现 Qdrant 索引损坏

**根因**：
Qdrant HNSW 索引参数配置不当，导致索引损坏

**解决方案**：
1. 重建 Qdrant 索引
2. 优化 HNSW 参数

**改进措施**：
- 添加索引健康监控
- 定期验证索引完整性

### 4.2 案例 2：备份失败

**故障信息**：
- **日期**: 2026-03-10
- **级别**: P2
- **影响时长**: 2 小时
- **现象**: 备份任务失败，告警通知

**排查过程**：
1. 检查备份日志发现磁盘空间不足
2. 检查磁盘使用率发现 /backup 分区 100%
3. 检查备份文件发现未清理过期备份

**根因**：
备份清理任务未执行，导致磁盘空间耗尽

**解决方案**：
1. 手动清理过期备份
2. 修复备份清理任务

**改进措施**：
- 添加磁盘空间监控告警
- 备份清理任务增加执行确认

### 4.3 案例 3：记忆丢失

**故障信息**：
- **日期**: 2026-03-15
- **级别**: P0
- **影响时长**: 1 小时
- **现象**: 用户报告记忆数据丢失

**排查过程**：
1. 确认数据丢失范围（特定会话）
2. 检查数据库记录正常
3. 检查缓存发现缓存污染
4. 检查缓存逻辑发现 bug

**根因**：
缓存键生成逻辑 bug，导致不同会话共享缓存

**解决方案**：
1. 清除污染缓存
2. 修复缓存键生成逻辑
3. 从数据库重新加载数据

**改进措施**：
- 增加缓存键验证
- 添加数据一致性检查

---

## 5. 附录

### 5.1 故障报告模板

```markdown
## 故障报告

### 基本信息

- **故障编号**: INC-YYYYMMDD-XXX
- **故障日期**: 2026-03-09
- **故障级别**: P0/P1/P2/P3
- **发现时间**: HH:MM
- **恢复时间**: HH:MM
- **影响时长**: X 小时 X 分钟

### 故障描述

(详细描述故障现象)

### 影响范围

- 影响用户：全部/部分/特定用户
- 影响功能：检索/存储/全部
- 影响程度：完全不可用/性能下降

### 时间线

| 时间 | 事件 | 负责人 |
|------|------|--------|
| HH:MM | 故障发生 | - |
| HH:MM | 故障发现 | XXX |
| HH:MM | 开始排查 | XXX |
| HH:MM | 定位原因 | XXX |
| HH:MM | 实施修复 | XXX |
| HH:MM | 验证通过 | XXX |
| HH:MM | 故障恢复 | XXX |

### 根因分析

(详细描述根本原因)

### 解决方案

(已采取的修复措施)

### 改进措施

| 措施 | 负责人 | 截止日期 | 状态 |
|------|--------|----------|------|
| 添加监控 | XXX | 2026-03-15 | 待办 |
| 修复 bug | XXX | 2026-03-12 | 进行中 |

### 经验教训

(从本次故障中学到的经验)
```

### 5.2 诊断命令速查

```bash
# 服务状态
kubectl get pods -l app=go-agent
kubectl get svc go-agent

# 日志查看
kubectl logs -l app=go-agent --tail=100
kubectl logs -l app=go-agent | grep ERROR

# 资源检查
kubectl top pods -l app=go-agent
kubectl top nodes

# 数据库检查
psql -h postgres -U user -d memory -c "SELECT 1"
psql -c "SELECT count(*) FROM pg_stat_activity;"

# 向量数据库检查
curl http://qdrant:6333/collections/memory
curl http://qdrant:6333/cluster/status

# 网络检查
kubectl exec -it go-agent-pod -- nc -zv qdrant 6333
kubectl exec -it go-agent-pod -- nc -zv postgres 5432
```

### 5.3 联系人列表

| 角色 | 姓名 | 电话 | 邮箱 |
|------|------|------|------|
| 值班工程师 | | | |
| SRE 负责人 | | | |
| 后端负责人 | | | |
| 技术负责人 | | | |

### 5.4 参考资源

- [监控指南](./01_monitoring_guide.md)
- [性能调优](./02_performance_tuning.md)
- [维护检查清单](./08_maintenance_checklist.md)
- [内部 Wiki](https://wiki.marecogito.ai/)

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
