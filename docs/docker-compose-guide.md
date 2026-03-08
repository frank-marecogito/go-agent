# go-agent Docker Compose 使用指南

快速启动 PostgreSQL (pgvector) 和 pgweb 管理工具。

## 🚀 快速开始

### 1. 配置环境变量

```bash
cd ~/MareCogito/go-agent

# 复制环境变量模板
cp .env.example .env

# 编辑配置文件（可选，使用默认值可直接跳过）
nano .env
```

### 2. 启动所有服务

```bash
# 启动 PostgreSQL + pgweb
docker compose up -d

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 3. 访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| **pgweb** | http://localhost:8081 | PostgreSQL Web 管理界面 |
| **PostgreSQL** | localhost:5432 | 数据库连接（go-agent 应用） |

### 4. 停止服务

```bash
# 停止所有服务
docker compose down

# 停止并删除数据卷（⚠️ 会删除所有数据）
docker compose down -v
```

## 📁 文件说明

| 文件 | 用途 |
|------|------|
| `docker-compose.yml` | 主配置文件（PostgreSQL + pgweb） |
| `docker-compose.pgweb.yml` | 仅 pgweb（连接外部 PostgreSQL） |
| `.env` | 环境变量配置（**不提交到 Git**） |
| `.env.example` | 环境变量模板（可提交） |

## 🔧 常用命令

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启某个服务
docker compose restart postgres
docker compose restart pgweb

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f postgres
docker compose logs -f pgweb

# 进入 PostgreSQL 容器
docker compose exec postgres psql -U admin -d ragdb

# 进入 pgweb 容器
docker compose exec pgweb sh

# 重新构建并启动
docker compose up -d --build
```

## 🔌 连接 go-agent 应用

### 方式 1: 使用 .env 文件

```bash
# .env 文件已包含 DATABASE_URL
# 注意：修改 POSTGRES_* 变量后，需同步更新 DATABASE_URL
source .env
go run ./cmd/example/main.go -message "Hello"
```

### 方式 2: 直接设置环境变量

```bash
export DATABASE_URL="postgres://admin:admin@localhost:5432/ragdb?sslmode=disable"
go run ./cmd/example/main.go -message "Hello"
```

### 方式 3: 命令行参数

```bash
go run ./cmd/example/main.go \
  -pg "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable" \
  -message "Hello"
```

## 🔒 安全建议

### ⚠️ 开发环境

当前配置使用默认密码 `admin/admin`，**仅用于本地开发**。

### ✅ 生产环境

1. **修改默认密码**
   ```bash
   # 生成强密码
   openssl rand -base64 32
   
   # 编辑 .env 文件
   POSTGRES_PASSWORD=你的强密码
   ```

2. **限制网络访问**
   ```yaml
   # docker-compose.yml
   ports:
     - "127.0.0.1:5432:5432"  # 仅本地访问
     - "127.0.0.1:8081:8081"  # 仅本地访问
   ```

3. **启用 SSL 连接**
   ```env
   DATABASE_URL=postgres://user:pass@localhost:5432/db?sslmode=require
   ```

## 🛠️ 故障排查

### PostgreSQL 无法启动

```bash
# 查看详细日志
docker compose logs postgres

# 检查端口是否被占用
lsof -i :5432

# 删除数据卷重新启动（⚠️ 会删除数据）
docker compose down -v
docker compose up -d
```

### pgweb 无法连接 PostgreSQL

```bash
# 检查 PostgreSQL 是否健康
docker compose ps

# 查看 pgweb 日志
docker compose logs pgweb

# 重启 pgweb
docker compose restart pgweb
```

### 密码错误

```bash
# 1. 停止服务
docker compose down

# 2. 删除数据卷
docker volume rm go-agent-postgres-data

# 3. 修改 .env 文件中的密码

# 4. 重新启动
docker compose up -d
```

## 📊 服务架构

```
┌─────────────────┐
│   go-agent 应用  │
│  (localhost)    │
└────────┬────────┘
         │ DATABASE_URL
         │ postgres://localhost:5432
         ▼
┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │◄────│     pgweb       │
│   (pgvector)    │     │  (Web UI:8081)  │
│   Port: 5432    │     │                 │
└─────────────────┘     └─────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
              Docker Network: go-agent-net
```

## 🔗 相关资源

- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [pgvector GitHub](https://github.com/pgvector/pgvector)
- [pgweb 官方文档](https://github.com/sosedoff/pgweb)
- [go-agent README](../README.md)

---

**最后更新**: 2026 年 3 月 2 日
