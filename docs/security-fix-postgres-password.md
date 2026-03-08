# 安全修复记录 - PostgreSQL 密码管理

**日期**: 2026 年 3 月 2 日  
**问题**: 明文密码硬编码在代码和配置文件中  
**修复**: 改用环境变量管理敏感信息

---

## 🔍 受影响的文件

| 文件 | 修改内容 | 状态 |
|------|----------|------|
| `cmd/example/main.go` | 添加 `getEnv()` 函数，从环境变量读取数据库连接 | ✅ 已修复 |
| `docker-compose.yml` | 使用 `${POSTGRES_USER:-admin}` 变量语法 | ✅ 已修复 |
| `docker-compose.pgweb.yml` | 移除自动连接配置，通过 Web 界面手动连接 | ✅ 已修复 |
| `.env` | 新建实际环境变量配置文件 | ✅ 已创建 |
| `.env.example` | 新建环境变量模板文件 | ✅ 已创建 |
| `README.md` | 更新环境变量说明 | ✅ 已更新 |
| `docs/pgweb-setup.md` | 新建 pgweb 使用文档 | ✅ 已创建 |

---

## 🔧 修改详情

### 1. cmd/example/main.go

**修改前**:
```go
pgConnStr = flag.String("pg", "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable", ...)
```

**修改后**:
```go
defaultPgConnStr = getEnv("DATABASE_URL",
    fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
        getEnv("POSTGRES_USER", "admin"),
        getEnv("POSTGRES_PASSWORD", "admin"),
        getEnv("POSTGRES_HOST", "localhost"),
        getEnv("POSTGRES_PORT", "5432"),
        getEnv("POSTGRES_DB", "ragdb"),
    ),
)
pgConnStr = flag.String("pg", defaultPgConnStr, ...)
```

**使用方式**:
```bash
# 方式 1: 使用 .env 文件（推荐）
cp .env.example .env
# 编辑 .env 文件设置密码

# 方式 2: 直接设置环境变量
export DATABASE_URL="postgres://user:pass@localhost:5432/ragdb?sslmode=disable"
export POSTGRES_PASSWORD="your_secure_password"

# 方式 3: 命令行参数覆盖
go run ./cmd/example/main.go -pg "postgres://user:pass@localhost:5432/mydb"
```

### 2. docker-compose.yml / docker-compose.pgweb.yml

**修改前**:
```yaml
environment:
  - DATABASE_URL=postgres://admin:admin@host.docker.internal:5432/postgres?sslmode=disable
```

**修改后**:
```yaml
environment:
  - PGWEB_DATABASE_URL=postgres://${POSTGRES_USER:-admin}:${POSTGRES_PASSWORD:-admin}@${POSTGRES_HOST:-host.docker.internal}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-postgres}?sslmode=disable
```

---

## 📋 使用指南

### 快速开始

```bash
cd ~/MareCogito/go-agent

# 1. 复制环境变量模板
cp .env.example .env

# 2. 编辑 .env 文件，设置安全密码
nano .env

# 3. 启动 pgweb
docker compose -f docker-compose.pgweb.yml up -d

# 4. 访问 Web 界面
open http://localhost:8081
```

### 运行示例代码

```bash
# 使用 .env 文件中的配置
source .env
go run ./cmd/example/main.go -message "Hello"

# 或直接使用环境变量
DATABASE_URL="postgres://admin:secure_password@localhost:5432/ragdb" \
  go run ./cmd/example/main.go -message "Hello"
```

---

## 🔒 安全建议

### ⚠️ 当前状态

- ✅ 代码中不再硬编码密码
- ✅ 使用环境变量管理敏感信息
- ✅ `.env` 文件已在 `.gitignore` 中
- ⚠️ 默认密码仍是 `admin/admin`（仅用于开发）

### ✅ 生产环境必须

1. **修改默认密码**
   ```bash
   # 生成强密码
   openssl rand -base64 32
   
   # 在 PostgreSQL 中修改
   docker exec -it postgres-pgvector psql -U admin -d postgres \
     -c "ALTER USER admin WITH PASSWORD '强密码';"
   
   # 更新 .env 文件
   POSTGRES_PASSWORD=强密码
   ```

2. **限制网络访问**
   ```yaml
   # docker-compose.pgweb.yml
   ports:
     - "127.0.0.1:8081:8081"  # 仅本地访问
   ```

3. **启用 SSL**
   ```env
   DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
   ```

4. **定期轮换密码**
   - 建议每 90 天更换一次
   - 使用密码管理工具记录

---

## 🧪 验证

```bash
# 1. 检查是否还有明文密码
grep -r "admin:admin" --include="*.go" --include="*.yml" .
# 应该无输出（除了注释中的示例）

# 2. 验证代码编译
go build ./...

# 3. 验证 docker-compose 配置
docker compose config

# 4. 测试连接
docker exec marecogito-pgweb pgweb --version
```

---

## 📚 相关文档

- [pgweb 使用指南](./docs/pgweb-setup.md)
- [环境变量模板](./.env.example)
- [README.md](./README.md) - 环境变量说明

---

**最后更新**: 2026 年 3 月 2 日
