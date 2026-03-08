# go-agent Docker Compose 配置总结

**日期**: 2026 年 3 月 2 日

---

## 📁 配置文件

| 文件 | 用途 | 状态 |
|------|------|------|
| `docker-compose.yml` | PostgreSQL + pgweb 完整配置 | ✅ 已更新 |
| `.env` | 环境变量配置 | ✅ 已创建 |
| `.env.example` | 环境变量模板 | ✅ 已创建 |

## 🗑️ 已删除文件

| 文件 | 原因 |
|------|------|
| `docker-compose.pgweb.yml` | 功能已整合到 docker-compose.yml |

---

## 🚀 使用方法

### 启动所有服务

```bash
cd ~/MareCogito/go-agent

# 配置环境变量（首次使用）
cp .env.example .env

# 启动 PostgreSQL + pgweb
docker compose up -d

# 查看状态
docker compose ps
```

### 访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| **pgweb** | http://localhost:8081 | PostgreSQL Web 管理界面 |
| **PostgreSQL** | localhost:5432 | go-agent 应用连接 |

### 停止服务

```bash
docker compose down
```

---

## 🏗️ 服务架构

```
┌─────────────────┐
│  go-agent 应用   │
│   (本地运行)     │
└────────┬────────┘
         │ DATABASE_URL
         │ postgres://localhost:5432/ragdb
         ▼
┌─────────────────────────────────────┐
│         Docker Compose              │
│  ┌─────────────┐    ┌─────────────┐ │
│  │  PostgreSQL │◄───│   pgweb     │ │
│  │  (pgvector) │    │  (Web UI)   │ │
│  │  Port:5432  │    │  Port:8081  │ │
│  └─────────────┘    └─────────────┘ │
│         │                  │         │
│         └────────┬─────────┘         │
│                  │                   │
│       Docker Network: go-agent-net   │
└─────────────────────────────────────┘
```

---

## 📋 环境变量

### PostgreSQL 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `POSTGRES_USER` | admin | 数据库用户名 |
| `POSTGRES_PASSWORD` | admin | 数据库密码 |
| `POSTGRES_DB` | ragdb | 默认数据库名 |
| `POSTGRES_HOST` | localhost | 数据库主机（应用连接用） |
| `POSTGRES_PORT` | 5432 | 数据库端口 |

### 其他配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DATABASE_URL` | postgres://admin:admin@localhost:5432/ragdb | 完整连接字符串 |
| `QDRANT_URL` | http://localhost:6333 | Qdrant 向量数据库 |
| `ADK_EMBED_PROVIDER` | gemini | Embedding 模型提供商 |
| `GOOGLE_API_KEY` | - | Gemini API 密钥 |

---

## 🔒 安全改进

### 修改前

```yaml
# ❌ 明文密码硬编码
environment:
  - DATABASE_URL=postgres://admin:admin@localhost:5432/ragdb
  - POSTGRES_PASSWORD=admin
```

### 修改后

```yaml
# ✅ 使用环境变量
environment:
  - POSTGRES_USER=${POSTGRES_USER:-admin}
  - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-admin}
  - POSTGRES_DB=${POSTGRES_DB:-ragdb}
```

---

## ⚠️ 重要提示

### 默认密码

当前配置使用 `admin/admin` 作为默认密码，**仅用于本地开发环境**。

### 生产环境必须

1. 修改 `.env` 文件中的密码
2. 限制端口仅本地访问
3. 启用 SSL 连接
4. 定期轮换密码

---

## 📚 相关文档

- [Docker Compose 使用指南](./docker-compose-guide.md)
- [pgweb 使用指南](./pgweb-setup.md)
- [安全修复记录](./security-fix-postgres-password.md)
- [README.md](../README.md)

---

**最后更新**: 2026 年 3 月 2 日
