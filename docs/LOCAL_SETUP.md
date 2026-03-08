# go-agent 本地配置指南

**日期**: 2026 年 3 月 2 日

---

## 📋 背景

本项目是 GitHub 开源库（github.com/Protocol-Lattice/go-agent）的本地 Fork。
为了保持 Git 状态整洁，同时保留个性化配置，使用了以下策略：

---

## 🔧 配置策略

### 1. Skip-worktree 标记的文件

以下文件已标记为 `skip-worktree`，本地修改不会被 Git 追踪：

| 文件 | 说明 |
|------|------|
| `.gitignore` | 添加了本地忽略规则 |
| `README.md` | 更新了环境变量说明 |
| `cmd/example/main.go` | 改用环境变量读取数据库密码 |

**查看标记**：
```bash
git ls-files -v | grep "^S"
```

**恢复追踪**：
```bash
git update-index --no-skip-worktree <文件>
```

### 2. 本地忽略文件（.git/info/exclude）

以下文件通过 `.git/info/exclude` 忽略，仅本地生效：

| 文件 | 说明 |
|------|------|
| `.env` | 环境变量配置（敏感信息） |
| `.env.example` | 环境变量模板 |
| `docker-compose.yml` | Docker Compose 配置 |
| `docker-compose.*.yml` | Docker Compose 变体 |
| `.gitignore.local` | 本地忽略规则 |
| `/main` | 编译产物 |

**查看忽略规则**：
```bash
cat .git/info/exclude
```

**验证忽略**：
```bash
git check-ignore -v .env docker-compose.yml
```

---

## 🚀 快速启动

### 配置环境变量

```bash
# 复制模板（如果不存在）
cp .env.example .env

# 编辑配置
nano .env
```

### 启动 Docker 服务

```bash
# 启动 PostgreSQL + pgweb
docker compose up -d

# 查看状态
docker compose ps

# 访问 pgweb
open http://localhost:8081
```

### 运行应用

```bash
# 使用环境变量
source .env
go run ./cmd/example/main.go -message "Hello"
```

---

## 📁 文件结构

```
go-agent/
├── .git/
│   └── info/
│       └── exclude          # 本地忽略规则（不提交）
├── .env                     # 环境变量配置（已忽略）
├── .env.example             # 环境变量模板（已忽略）
├── .gitignore               # Git 忽略规则（skip-worktree）
├── .gitignore.local         # 本地忽略规则（已忽略）
├── docker-compose.yml       # Docker Compose 配置（已忽略）
└── cmd/example/
    └── main.go              # 示例代码（skip-worktree）
```

---

## 🔄 同步上游更新

```bash
# 添加上游远程（如果未添加）
git remote add upstream https://github.com/Protocol-Lattice/go-agent.git

# 获取上游更新
git fetch upstream

# 合并到本地分支
git merge upstream/main

# 或使用 rebase
git rebase upstream/main
```

**注意**: skip-worktree 标记的文件不会受上游更新影响。

---

## ⚠️ 注意事项

### 提交代码前

1. **检查 Git 状态**：
   ```bash
   git status
   ```

2. **确认 skip-worktree 文件**：
   ```bash
   git ls-files -v | grep "^S"
   ```

3. **如需提交更改**：
   ```bash
   # 临时恢复追踪
   git update-index --no-skip-worktree <文件>
   
   # 提交更改
   git add <文件>
   git commit -m "描述"
   
   # 重新标记为 skip-worktree（如果需要）
   git update-index --skip-worktree <文件>
   ```

### 清理本地配置

如需恢复原始状态：

```bash
# 恢复所有 skip-worktree 文件
git ls-files -v | grep "^S" | cut -d' ' -f2 | xargs git update-index --no-skip-worktree

# 恢复文件到上游状态
git checkout HEAD -- .gitignore README.md cmd/example/main.go

# 清理本地忽略
git checkout HEAD -- .git/info/exclude
```

---

## 📚 相关文档

- [Docker Compose 使用指南](./docs/docker-compose-guide.md)
- [pgweb 使用指南](./docs/pgweb-setup.md)
- [安全修复记录](./docs/security-fix-postgres-password.md)

---

**最后更新**: 2026 年 3 月 2 日
