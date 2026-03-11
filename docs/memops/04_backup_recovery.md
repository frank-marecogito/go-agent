# 记忆系统备份恢复指南

**文档编号**: OPS-04  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐ 重要文档  
**适用角色**: SRE 工程师 / 运维工程师  

---

## 📋 目录

1. [备份策略](#1-备份策略)
2. [备份实施](#2-备份实施)
3. [恢复流程](#3-恢复流程)
4. [灾难恢复](#4-灾难恢复)
5. [附录](#5-附录)

---

## 1. 备份策略

### 1.1 备份类型

| 类型 | 说明 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|----------|
| **全量备份** | 备份所有数据 | 恢复简单 | 耗时、占用空间大 | 每周一次 |
| **增量备份** | 仅备份变更数据 | 快速、节省空间 | 恢复复杂 | 每日一次 |
| **差异备份** | 备份距上次全量后的变更 | 恢复较快 | 占用空间中等 | 每 3 天一次 |

### 1.2 备份频率规划

```
┌─────────────────────────────────────────────────────────┐
│                  备份频率规划                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  全量备份：每周日 02:00                                  │
│  ├── 保留 4 周                                          │
│  └── 存储：/backup/full/YYYYMMDD/                       │
│                                                         │
│  增量备份：周一至周六 02:00                              │
│  ├── 保留 7 天                                          │
│  └── 存储：/backup/incremental/YYYYMMDD/                │
│                                                         │
│  实时备份：WAL 归档                                      │
│  ├── 保留 7 天                                          │
│  └── 存储：/backup/wal/                                 │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 1.3 备份保留策略

| 备份类型 | 保留期限 | 存储位置 | 清理策略 |
|----------|----------|----------|----------|
| 全量备份 | 4 周 | 本地 + 云端 | 自动清理 4 周前 |
| 增量备份 | 7 天 | 本地 | 自动清理 7 天前 |
| WAL 归档 | 7 天 | 本地 | 自动清理 7 天前 |
| 月度归档 | 12 个月 | 云端冷存储 | 手动清理 |

### 1.4 RTO 与 RPO

| 指标 | 定义 | 目标值 | 说明 |
|------|------|--------|------|
| **RTO** (Recovery Time Objective) | 恢复时间目标 | < 30 分钟 | 从故障到恢复的时间 |
| **RPO** (Recovery Point Objective) | 恢复点目标 | < 5 分钟 | 允许丢失的数据量 |

---

## 2. 备份实施

### 2.1 PostgreSQL 备份

#### 2.1.1 全量备份脚本

创建 `scripts/backup_postgres.sh`：

```bash
#!/bin/bash
set -e

# 配置
BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30
DB_NAME="memory_db"
DB_USER="postgres"

# 创建备份目录
mkdir -p ${BACKUP_DIR}/full/${DATE}
mkdir -p ${BACKUP_DIR}/wal

echo "[$(date)] 开始 PostgreSQL 全量备份..."

# 使用 pg_dump 进行全量备份
pg_dump -U ${DB_USER} -h localhost -d ${DB_NAME} \
  --format=custom \
  --compress=9 \
  --verbose \
  --file=${BACKUP_DIR}/full/${DATE}/memory_db.dump

# 备份全局对象（角色、表空间等）
pg_dumpall -U ${DB_USER} -h localhost \
  --globals-only \
  --file=${BACKUP_DIR}/full/${DATE}/globals.sql

# 计算校验和
cd ${BACKUP_DIR}/full/${DATE}
md5sum *.dump *.sql > checksum.md5

# 清理过期备份
find ${BACKUP_DIR}/full -type d -mtime +${RETENTION_DAYS} -exec rm -rf {} \;

echo "[$(date)] PostgreSQL 全量备份完成"
echo "备份位置：${BACKUP_DIR}/full/${DATE}"
```

#### 2.1.2 增量备份（WAL 归档）

配置 `postgresql.conf`：

```conf
# WAL 归档配置
wal_level = replica
archive_mode = on
archive_command = 'cp %p /backup/postgres/wal/%f'
archive_timeout = 300  # 5 分钟强制切换
```

创建 WAL 归档脚本 `scripts/archive_wal.sh`：

```bash
#!/bin/bash

WAL_FILE=$1
ARCHIVE_DIR="/backup/postgres/wal"

mkdir -p ${ARCHIVE_DIR}

# 复制 WAL 文件
cp ${WAL_FILE} ${ARCHIVE_DIR}/$(basename ${WAL_FILE})

# 压缩旧 WAL 文件
find ${ARCHIVE_DIR} -name "*.gz" -mtime +7 -delete

exit 0
```

#### 2.1.3 备份验证

创建 `scripts/verify_backup.sh`：

```bash
#!/bin/bash

BACKUP_DIR="/backup/postgres"
LATEST_BACKUP=$(ls -td ${BACKUP_DIR}/full/*/ | head -1)

echo "验证备份：${LATEST_BACKUP}"

# 验证校验和
cd ${LATEST_BACKUP}
if md5sum -c checksum.md5 > /dev/null 2>&1; then
    echo "✓ 校验和验证通过"
else
    echo "✗ 校验和验证失败"
    exit 1
fi

# 测试恢复（到测试数据库）
TEST_DB="test_memory_db"
pg_restore -U postgres -h localhost -d ${TEST_DB} --clean --if-exists \
  ${LATEST_BACKUP}/memory_db.dump

if [ $? -eq 0 ]; then
    echo "✓ 恢复测试通过"
else
    echo "✗ 恢复测试失败"
    exit 1
fi

echo "备份验证完成"
```

### 2.2 Qdrant 备份

#### 2.2.1 快照备份

创建 `scripts/backup_qdrant.sh`：

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/qdrant"
DATE=$(date +%Y%m%d_%H%M%S)
QDRANT_URL="http://localhost:6333"
COLLECTION="memory"

echo "[$(date)] 开始 Qdrant 快照备份..."

# 创建快照目录
mkdir -p ${BACKUP_DIR}/${DATE}

# 创建快照
SNAPSHOT_RESPONSE=$(curl -X POST "${QDRANT_URL}/collections/${COLLECTION}/snapshots")
SNAPSHOT_NAME=$(echo ${SNAPSHOT_RESPONSE} | jq -r '.result.name')

# 下载快照
curl -o ${BACKUP_DIR}/${DATE}/snapshot.snap \
  "${QDRANT_URL}/collections/${COLLECTION}/snapshots/${SNAPSHOT_NAME}"

# 备份集合配置
curl -o ${BACKUP_DIR}/${DATE}/collection_config.json \
  "${QDRANT_URL}/collections/${COLLECTION}"

# 创建元数据文件
cat > ${BACKUP_DIR}/${DATE}/metadata.json << EOF
{
  "collection": "${COLLECTION}",
  "timestamp": "$(date -Iseconds)",
  "snapshot": "${SNAPSHOT_NAME}"
}
EOF

echo "[$(date)] Qdrant 快照备份完成"
echo "备份位置：${BACKUP_DIR}/${DATE}"
```

### 2.3 配置文件备份

创建 `scripts/backup_config.sh`：

```bash
#!/bin/bash

BACKUP_DIR="/backup/config"
DATE=$(date +%Y%m%d_%H%M%S)

echo "[$(date)] 开始配置文件备份..."

mkdir -p ${BACKUP_DIR}/${DATE}

# 备份应用配置
cp -r /etc/go-agent ${BACKUP_DIR}/${DATE}/ 2>/dev/null || true
cp -r /app/config ${BACKUP_DIR}/${DATE}/ 2>/dev/null || true

# 备份 Kubernetes 配置
kubectl get configmap,secret -n memory-system -o yaml \
  > ${BACKUP_DIR}/${DATE}/k8s-config.yaml 2>/dev/null || true

# 备份 Docker 配置
cp docker-compose.yml ${BACKUP_DIR}/${DATE}/ 2>/dev/null || true

echo "[$(date)] 配置文件备份完成"
```

### 2.4 备份自动化

#### 2.4.1 cron 任务

创建 `/etc/cron.d/memory-backup`：

```bash
# PostgreSQL 全量备份（每周日 02:00）
0 2 * * 0 root /opt/scripts/backup_postgres.sh >> /var/log/backup_postgres.log 2>&1

# Qdrant 快照备份（每天 03:00）
0 3 * * * root /opt/scripts/backup_qdrant.sh >> /var/log/backup_qdrant.log 2>&1

# 配置文件备份（每天 04:00）
0 4 * * * root /opt/scripts/backup_config.sh >> /var/log/backup_config.log 2>&1

# 备份验证（每天 06:00）
0 6 * * * root /opt/scripts/verify_backup.sh >> /var/log/verify_backup.log 2>&1
```

#### 2.4.2 Kubernetes CronJob

创建 `k8s/backup_cronjob.yaml`：

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: memory-system
spec:
  schedule: "0 2 * * *"  # 每天 02:00
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres -U postgres -d memory_db \
                --format=custom \
                --file=/backup/memory_db.dump
              
              # 上传到对象存储
              aws s3 cp /backup/memory_db.dump \
                s3://memory-backup/postgres/$(date +%Y%m%d)/memory_db.dump
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
            volumeMounts:
            - name: backup-volume
              mountPath: /backup
          volumes:
          - name: backup-volume
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

---

## 3. 恢复流程

### 3.1 PostgreSQL 恢复

#### 3.1.1 从全量备份恢复

```bash
#!/bin/bash
set -e

BACKUP_FILE="/backup/postgres/full/20260309_020000/memory_db.dump"
DB_NAME="memory_db"
DB_USER="postgres"

echo "[$(date)] 开始 PostgreSQL 恢复..."

# 删除现有数据库（可选）
dropdb -U ${DB_USER} -h localhost ${DB_NAME}

# 创建新数据库
createdb -U ${DB_USER} -h localhost ${DB_NAME}

# 恢复数据
pg_restore -U ${DB_USER} -h localhost -d ${DB_NAME} \
  --verbose \
  --clean \
  --if-exists \
  ${BACKUP_FILE}

echo "[$(date)] PostgreSQL 恢复完成"
```

#### 3.1.2 时间点恢复（PITR）

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/postgres"
TARGET_TIME="2026-03-09 15:00:00"
DATA_DIR="/var/lib/postgresql/data"

echo "[$(date)] 开始 PostgreSQL 时间点恢复..."

# 停止 PostgreSQL
systemctl stop postgresql

# 清空数据目录
rm -rf ${DATA_DIR}/*

# 恢复基础备份
pg_restore -U postgres -d memory_db \
  ${BACKUP_DIR}/full/latest/memory_db.dump

# 配置恢复
cat > ${DATA_DIR}/recovery.signal << EOF
EOF

cat > ${DATA_DIR}/postgresql.auto.conf << EOF
restore_command = 'cp ${BACKUP_DIR}/wal/%f %p'
recovery_target_time = '${TARGET_TIME}'
EOF

# 启动 PostgreSQL
systemctl start postgresql

echo "[$(date)] PostgreSQL 时间点恢复完成"
```

### 3.2 Qdrant 恢复

#### 3.2.1 从快照恢复

```bash
#!/bin/bash
set -e

BACKUP_FILE="/backup/qdrant/20260309_030000/snapshot.snap"
QDRANT_URL="http://localhost:6333"
COLLECTION="memory"

echo "[$(date)] 开始 Qdrant 快照恢复..."

# 删除现有集合（可选）
curl -X DELETE "${QDRANT_URL}/collections/${COLLECTION}"

# 创建新集合
curl -X PUT "${QDRANT_URL}/collections/${COLLECTION}" \
  -H "Content-Type: application/json" \
  -d @/backup/qdrant/20260309_030000/collection_config.json

# 上传并恢复快照
curl -X POST "${QDRANT_URL}/collections/${COLLECTION}/snapshots/upload" \
  -F "snapshot=@${BACKUP_FILE}"

echo "[$(date)] Qdrant 快照恢复完成"
```

### 3.3 完整恢复流程

```
1. 评估故障范围
   │
   ▼
2. 确定恢复点（RPO）
   │
   ▼
3. 停止相关服务
   │
   ▼
4. 恢复 PostgreSQL 数据
   │
   ▼
5. 恢复 Qdrant 数据
   │
   ▼
6. 恢复配置文件
   │
   ▼
7. 启动服务
   │
   ▼
8. 验证恢复效果
   │
   ▼
9. 通知用户恢复完成
```

---

## 4. 灾难恢复

### 4.1 灾难场景定义

| 场景 | 描述 | 恢复策略 | RTO |
|------|------|----------|-----|
| **单点故障** | 单个服务实例故障 | 自动重启/切换 | < 5 分钟 |
| **数据库故障** | PostgreSQL 不可用 | 主从切换 | < 15 分钟 |
| **向量数据库故障** | Qdrant 不可用 | 从备份恢复 | < 30 分钟 |
| **区域故障** | 整个区域不可用 | 跨区域切换 | < 1 小时 |
| **数据损坏** | 数据被错误修改 | 时间点恢复 | < 2 小时 |

### 4.2 灾难恢复计划

#### 4.2.1 应急联系人

| 角色 | 姓名 | 电话 | 邮箱 |
|------|------|------|------|
| 值班工程师 | | | |
| SRE 负责人 | | | |
| DBA | | | |
| 技术负责人 | | | |

#### 4.2.2 恢复优先级

| 优先级 | 服务 | 恢复时间目标 |
|--------|------|--------------|
| P0 | go-agent API | 15 分钟 |
| P0 | PostgreSQL | 30 分钟 |
| P1 | Qdrant | 30 分钟 |
| P2 | 监控系统 | 1 小时 |
| P3 | 日志系统 | 2 小时 |

### 4.3 灾难恢复演练

#### 4.3.1 演练计划

| 演练场景 | 频率 | 参与人员 | 时长 |
|----------|------|----------|------|
| PostgreSQL 主从切换 | 每季度 | SRE 团队 | 2 小时 |
| Qdrant 恢复 | 每半年 | SRE 团队 | 2 小时 |
| 跨区域切换 | 每年 | 全体团队 | 4 小时 |

#### 4.3.2 演练报告模板

```markdown
## 灾难恢复演练报告

**演练日期**: 2026-03-09  
**演练场景**: PostgreSQL 主从切换  
**参与人员**: XXX, YYY, ZZZ

### 演练目标

- [ ] 验证主从切换流程
- [ ] 测试 RTO < 30 分钟
- [ ] 验证数据完整性

### 演练过程

| 时间 | 操作 | 结果 |
|------|------|------|
| 10:00 | 开始演练 | - |
| 10:05 | 模拟主库故障 | - |
| 10:08 | 检测到故障 | - |
| 10:15 | 完成主从切换 | - |
| 10:20 | 验证服务恢复 | ✓ |
| 10:30 | 演练结束 | - |

### 演练结果

- 实际 RTO: 20 分钟
- 目标 RTO: 30 分钟
- 结果：✓ 通过

### 发现问题

(详细描述)

### 改进措施

(具体建议)
```

---

## 5. 附录

### 5.1 备份检查清单

#### 每日检查

- [ ] 昨日备份成功
- [ ] 备份文件大小正常
- [ ] 备份日志无错误

#### 每周检查

- [ ] 全量备份成功
- [ ] 备份验证通过
- [ ] 备份空间充足

#### 每月检查

- [ ] 恢复测试通过
- [ ] 备份保留策略执行
- [ ] 备份成本分析

### 5.2 常用命令

```bash
# PostgreSQL 备份
pg_dump -U postgres -h localhost -d memory_db --format=custom -f backup.dump

# PostgreSQL 恢复
pg_restore -U postgres -h localhost -d memory_db backup.dump

# Qdrant 快照
curl -X POST http://localhost:6333/collections/memory/snapshots

# 验证备份
md5sum -c checksum.md5

# 查看备份大小
du -sh /backup/*
```

### 5.3 备份监控指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| 备份成功率 | 备份成功的比例 | < 95% |
| 备份延迟 | 实际备份时间与计划时间的偏差 | > 1 小时 |
| 备份大小 | 备份文件的大小 | 异常变化 > 50% |
| 恢复测试成功率 | 恢复测试成功的比例 | < 100% |

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
