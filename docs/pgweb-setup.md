# pgweb - PostgreSQL 网页管理工具

为 go-agent 项目的 PostgreSQL 实例提供可视化 Web 管理界面。

## 🚀 快速开始

### 1. 配置环境变量

```bash
cd ~/MareCogito/go-agent

# 复制示例配置
cp .env.example .env

# 编辑 .env 文件，修改数据库连接信息（可选）
nano .env
```

### 2. 启动服务

```bash
# 启动 PostgreSQL + pgweb
docker compose up -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f pgweb
```

### 3. 访问 Web 界面

打开浏览器访问：**http://localhost:8081**

pgweb 会自动连接到 PostgreSQL 容器（通过 Docker 内部网络）。

## 📁 文件说明

| 文件 | 用途 |
|------|------|
| `docker-compose.yml` | PostgreSQL + pgweb 完整配置 |
| `.env` | 环境变量（包含敏感信息，**不提交到 Git**） |
| `.env.example` | 环境变量模板（可提交到 Git） |

## 🔧 常用命令

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启 pgweb
docker compose restart pgweb

# 查看日志
docker logs -f go-agent-pgweb

# 进入容器
docker exec -it go-agent-pgweb sh
```

## 🔌 连接信息

pgweb 通过 Docker 内部网络自动连接 PostgreSQL：

| 字段 | 值 |
|------|-----|
| **Host** | `postgres:5432`（Docker 网络内） |
| **User** | `${POSTGRES_USER}`（默认：admin） |
| **Password** | `${POSTGRES_PASSWORD}`（默认：admin） |
| **Database** | `${POSTGRES_DB}`（默认：ragdb） |
| **SSL Mode** | `disable` |

## 🔒 安全建议

### ⚠️ 当前配置仅用于本地开发

- **默认密码**: `admin/admin` 仅用于开发环境
- **端口暴露**: `5432` 和 `8081` 仅绑定到 `localhost`
- **SSL 模式**: `disable` 仅用于本地连接

### ✅ 生产环境配置

1. **使用强密码**
   ```bash
   # 生成随机密码
   openssl rand -base64 32
   ```

2. **修改 .env 文件**
   ```env
   POSTGRES_PASSWORD=你的强密码
   ```

3. **限制网络访问**
   ```yaml
   # docker-compose.yml
   ports:
     - "127.0.0.1:8081:8081"  # 仅允许本地访问
     - "127.0.0.1:5432:5432"  # 仅允许本地访问
   ```

4. **启用 SSL 连接**
   ```env
   DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
   ```

## 🛠️ 故障排查

### pgweb 无法启动

```bash
# 查看详细日志
docker compose logs pgweb

# 检查容器状态
docker compose ps

# 重启容器
docker compose restart pgweb
```

### 无法连接到 PostgreSQL

```bash
# 1. 检查 PostgreSQL 容器是否运行
docker compose ps postgres

# 2. 检查 PostgreSQL 健康状态
docker compose exec postgres pg_isready -U admin -d ragdb

# 3. 查看 PostgreSQL 日志
docker compose logs postgres
```

### 端口冲突

如果 `8081` 端口被占用：

```bash
# 查找占用端口的进程
lsof -i :8081

# 修改 docker-compose.yml 中的端口映射
# pgweb 服务：
ports:
  - "8082:8081"  # 改为其他端口
```

## 📚 pgweb 功能

- ✅ 浏览数据库和表结构
- ✅ 执行 SQL 查询（支持语法高亮）
- ✅ 导入/导出数据（CSV、JSON、SQL）
- ✅ 查看查询执行计划
- ✅ 管理表、索引、视图、函数
- ✅ 查看服务器状态和进程
- ✅ 多数据库连接管理

## 🔗 相关资源

- [pgweb 官方文档](https://github.com/sosedoff/pgweb)
- [Docker Compose 使用指南](./docker-compose-guide.md)
- [go-agent README](../README.md)

---

**最后更新**: 2026 年 3 月 2 日
