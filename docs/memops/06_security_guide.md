# 记忆系统安全指南

**文档编号**: OPS-06  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐ 重要文档  
**适用角色**: 安全工程师 / SRE 工程师 / 运维工程师  

---

## 📋 目录

1. [安全配置](#1-安全配置)
2. [访问控制](#2-访问控制)
3. [数据安全](#3-数据安全)
4. [安全审计](#4-安全审计)
5. [附录](#5-附录)

---

## 1. 安全配置

### 1.1 网络安全

#### 1.1.1 防火墙配置

```bash
# UFW 防火墙配置

# 允许 SSH
sudo ufw allow 22/tcp

# 允许 HTTP/HTTPS（仅 Nginx）
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 允许应用端口（仅内网）
sudo ufw allow from 10.0.0.0/8 to any port 8080

# 允许 Metrics（仅监控服务器）
sudo ufw allow from 10.0.1.0/24 to any port 8081

# 允许数据库（仅应用服务器）
sudo ufw allow from 10.0.0.0/8 to any port 5432
sudo ufw allow from 10.0.0.0/8 to any port 6333

# 启用防火墙
sudo ufw enable

# 查看状态
sudo ufw status verbose
```

#### 1.1.2 SSL/TLS 配置

创建 Nginx SSL 配置 `/etc/nginx/conf.d/ssl.conf`：

```nginx
# SSL 配置
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
ssl_prefer_server_ciphers on;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;

# HSTS
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
```

生成自签名证书（开发环境）：

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/go-agent.key \
  -out /etc/ssl/certs/go-agent.crt \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=MareCogito/CN=marecogito.ai"
```

### 1.2 认证配置

#### 1.2.1 API Key 认证

```go
// API Key 中间件
func APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        
        if apiKey == "" {
            http.Error(w, "Missing API key", http.StatusUnauthorized)
            return
        }
        
        // 验证 API Key
        if !validateAPIKey(apiKey) {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func validateAPIKey(key string) bool {
    // 从数据库或缓存验证
    // ...
    return true
}
```

#### 1.2.2 JWT 认证

```go
// JWT 中间件
func JWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 验证 JWT
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 1.3 密钥管理

#### 1.3.1 密钥生成

```bash
# 生成 API Key
openssl rand -hex 32

# 生成 JWT Secret
openssl rand -base64 64

# 生成数据库密码
openssl rand -base64 32
```

#### 1.3.2 密钥存储

**使用环境变量**（不推荐用于生产）：

```bash
# .env 文件（权限 600）
chmod 600 .env

# 内容
API_KEY=your-api-key
JWT_SECRET=your-jwt-secret
DATABASE_PASSWORD=your-db-password
```

**使用 HashiCorp Vault**（推荐）：

```bash
# 启动 Vault
vault server -dev -dev-root-token-id=root

# 写入密钥
vault kv put secret/go-agent \
  api_key="your-api-key" \
  jwt_secret="your-jwt-secret" \
  db_password="your-db-password"

# 读取密钥
vault kv get secret/go-agent

# 在应用中使用
# 使用 Vault Agent 自动注入
```

#### 1.3.3 密钥轮换

```bash
#!/bin/bash
# 密钥轮换脚本

# 生成新密钥
NEW_API_KEY=$(openssl rand -hex 32)

# 更新 Vault
vault kv put secret/go-agent api_key="${NEW_API_KEY}"

# 通知应用重新加载配置
kubectl rollout restart deployment/go-agent

# 记录轮换日志
echo "$(date): API Key rotated" >> /var/log/key-rotation.log
```

---

## 2. 访问控制

### 2.1 用户认证

#### 2.1.1 用户注册流程

```
用户提交注册信息
    │
    ▼
验证邮箱/手机
    │
    ▼
密码强度检查
    │
    ▼
创建用户记录
    │
    ▼
发送欢迎邮件
```

#### 2.1.2 密码策略

```go
// 密码强度检查
func validatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }
    
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        return errors.New("password must contain uppercase letter")
    }
    
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        return errors.New("password must contain lowercase letter")
    }
    
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        return errors.New("password must contain digit")
    }
    
    if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) {
        return errors.New("password must contain special character")
    }
    
    return nil
}
```

### 2.2 权限管理

#### 2.2.1 RBAC 模型

```
┌─────────────────────────────────────────┐
│              RBAC 模型                    │
├─────────────────────────────────────────┤
│                                         │
│  用户 → 角色 → 权限                      │
│                                         │
│  角色定义：                              │
│  ├── admin（管理员）                     │
│  │   └── 所有权限                        │
│  ├── developer（开发者）                 │
│  │   ├── 读取记忆                        │
│  │   ├── 存储记忆                        │
│  │   └── 删除自己的记忆                  │
│  ├── reader（只读用户）                  │
│  │   └── 读取记忆                        │
│  └── service（服务账号）                 │
│      └── 特定 API 权限                    │
│                                         │
└─────────────────────────────────────────┘
```

#### 2.2.2 权限检查中间件

```go
// RBAC 中间件
func RBACMiddleware(requiredPermissions ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 获取用户信息
            user := getUserFromContext(r.Context())
            
            // 检查权限
            for _, perm := range requiredPermissions {
                if !user.HasPermission(perm) {
                    http.Error(w, "Forbidden", http.StatusForbidden)
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// 使用示例
mux.Handle("/memories",
    JWTAuth(
        RBACMiddleware("memory:read", "memory:write")(
            http.HandlerFunc(handleMemories))))
```

### 2.3 会话管理

#### 2.3.1 会话配置

```go
// 会话配置
type SessionConfig struct {
    Timeout        time.Duration `json:"timeout"`         // 会话超时
    MaxSessions    int           `json:"max_sessions"`    // 最大会话数
    SecureCookie   bool          `json:"secure_cookie"`   // 安全 Cookie
    HTTPOnly       bool          `json:"http_only"`       // HTTPOnly Cookie
    SameSite       string        `json:"same_site"`       // SameSite 策略
}

var DefaultSessionConfig = SessionConfig{
    Timeout:      24 * time.Hour,
    MaxSessions:  5,
    SecureCookie: true,
    HTTPOnly:     true,
    SameSite:     "Strict",
}
```

#### 2.3.2 会话清理

```bash
#!/bin/bash
# 清理过期会话

# PostgreSQL 清理
psql -U postgres -d memory_db -c "
  DELETE FROM sessions
  WHERE expires_at < NOW();
"

# Redis 清理（如果使用）
redis-cli KEYS "session:*" | xargs redis-cli DEL
```

---

## 3. 数据安全

### 3.1 传输加密

#### 3.1.1 HTTPS 强制

```nginx
# 强制 HTTPS
server {
    listen 80;
    server_name _;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    # ... SSL 配置
}
```

#### 3.1.2 数据库连接加密

```yaml
# PostgreSQL SSL 配置
database:
  url: "postgres://user:password@host:5432/db?sslmode=require"
  ssl_cert: "/etc/ssl/certs/postgresql.crt"
  ssl_key: "/etc/ssl/private/postgresql.key"
  ssl_rootcert: "/etc/ssl/certs/ca-certificates.crt"
```

### 3.2 存储加密

#### 3.2.1 数据库加密

```sql
-- 启用 pgcrypto 扩展
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 加密敏感数据
UPDATE users
SET encrypted_ssn = pgp_sym_encrypt(ssn, 'encryption-key');

-- 解密数据
SELECT pgp_sym_decrypt(encrypted_ssn, 'encryption-key') FROM users;
```

#### 3.2.2 文件加密

```bash
# 使用 GPG 加密备份文件
gpg --symmetric --cipher-algo AES256 backup.dump

# 解密
gpg --decrypt backup.dump.gpg
```

### 3.3 敏感数据脱敏

#### 3.3.1 日志脱敏

```go
// 日志脱敏过滤器
func sanitizeLog(data map[string]interface{}) map[string]interface{} {
    sensitiveFields := []string{"password", "token", "api_key", "secret"}
    
    for _, field := range sensitiveFields {
        if val, ok := data[field]; ok {
            if str, ok := val.(string); ok {
                // 只显示前 4 个字符
                if len(str) > 4 {
                    data[field] = str[:4] + "***"
                }
            }
        }
    }
    
    return data
}
```

#### 3.3.2 查询结果脱敏

```go
// 脱敏用户数据
func sanitizeUserData(user *User) *User {
    // 隐藏邮箱中间部分
    parts := strings.Split(user.Email, "@")
    if len(parts[0]) > 2 {
        parts[0] = parts[0][:2] + "***"
    }
    user.Email = strings.Join(parts, "@")
    
    // 隐藏手机号中间部分
    if len(user.Phone) == 11 {
        user.Phone = user.Phone[:3] + "****" + user.Phone[7:]
    }
    
    return user
}
```

---

## 4. 安全审计

### 4.1 审计日志

#### 4.1.1 审计事件类型

| 事件类型 | 说明 | 记录内容 |
|----------|------|----------|
| LOGIN | 用户登录 | 用户 ID、IP、时间、结果 |
| LOGOUT | 用户登出 | 用户 ID、时间 |
| ACCESS | 资源访问 | 用户 ID、资源、操作、结果 |
| MODIFY | 数据修改 | 用户 ID、数据、变更内容 |
| DELETE | 数据删除 | 用户 ID、数据、时间 |
| EXPORT | 数据导出 | 用户 ID、数据范围、时间 |

#### 4.1.2 审计日志表结构

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id BIGINT,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100),
    resource_id BIGINT,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    result VARCHAR(20) NOT NULL
);

-- 创建索引
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

### 4.2 合规检查

#### 4.2.1 安全检查清单

- [ ] 所有 API 端点都有认证
- [ ] 敏感操作有权限检查
- [ ] 密码符合强度要求
- [ ] 会话有超时机制
- [ ] 日志无敏感信息
- [ ] 数据传输使用 HTTPS
- [ ] 数据存储有加密
- [ ] 定期密钥轮换
- [ ] 审计日志完整

#### 4.2.2 自动化安全检查

```bash
#!/bin/bash
# 自动化安全检查脚本

echo "=== 安全检查 ==="

# 检查防火墙状态
echo "检查防火墙..."
if sudo ufw status | grep -q "Status: active"; then
    echo "✓ 防火墙已启用"
else
    echo "✗ 防火墙未启用"
fi

# 检查 SSL 证书
echo "检查 SSL 证书..."
if [ -f /etc/ssl/certs/go-agent.crt ]; then
    echo "✓ SSL 证书存在"
else
    echo "✗ SSL 证书缺失"
fi

# 检查密钥权限
echo "检查密钥权限..."
KEY_PERMS=$(stat -c %a /etc/ssl/private/go-agent.key 2>/dev/null)
if [ "$KEY_PERMS" = "600" ]; then
    echo "✓ 密钥权限正确"
else
    echo "✗ 密钥权限不正确 (当前：$KEY_PERMS)"
fi

# 检查过期密码
echo "检查密码过期..."
# ...

echo "安全检查完成"
```

### 4.3 安全扫描

#### 4.3.1 依赖漏洞扫描

```bash
# 使用 govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 使用 nancy
go list -json -m all | nancy sleuth
```

#### 4.3.2 容器安全扫描

```bash
# 使用 trivy
trivy image registry.marecogito.ai/go-agent:v1.0.0

# 使用 grype
grype registry.marecogito.ai/go-agent:v1.0.0
```

---

## 5. 附录

### 5.1 安全事件响应流程

```
发现安全事件
    │
    ▼
确认事件级别
    │
    ├─ 严重 → 立即通知，启动应急响应
    │
    ├─ 高 → 15 分钟内响应
    │
    └─ 中/低 → 按计划处理

隔离受影响系统
    │
    ▼
收集证据
    │
    ▼
分析根因
    │
    ▼
实施修复
    │
    ▼
验证修复
    │
    ▼
编写事件报告
```

### 5.2 安全配置检查清单

#### 部署前

- [ ] 修改默认密码
- [ ] 禁用不必要的服务
- [ ] 配置防火墙
- [ ] 启用 SSL/TLS
- [ ] 配置审计日志

#### 运营中

- [ ] 定期更新依赖
- [ ] 定期轮换密钥
- [ ] 定期审查权限
- [ ] 定期检查日志
- [ ] 定期进行安全扫描

### 5.3 参考资源

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)
- [Go 安全最佳实践](https://go.dev/doc/security/)

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
