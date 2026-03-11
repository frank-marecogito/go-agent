# 记忆系统部署指南

**文档编号**: OPS-03  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 9 日  
**优先级**: ⭐⭐ 重要文档  
**适用角色**: 运维工程师 / DevOps 工程师 / SRE 工程师  

---

## 📋 目录

1. [部署架构](#1-部署架构)
2. [环境准备](#2-环境准备)
3. [容器化部署](#3-容器化部署)
4. [Kubernetes 部署](#4-kubernetes 部署)
5. [配置管理](#5-配置管理)

---

## 1. 部署架构

### 1.1 单机部署架构

**适用场景**：开发、测试、小规模生产

```
┌─────────────────────────────────────────┐
│           单机部署架构                   │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────┐                        │
│  │  go-agent   │                        │
│  │  (1 实例)    │                        │
│  └──────┬──────┘                        │
│         │                               │
│  ┌──────┴──────┐                        │
│  │  PostgreSQL │                        │
│  │  + pgvector │                        │
│  └──────┬──────┘                        │
│         │                               │
│  ┌──────┴──────┐                        │
│  │   Qdrant    │                        │
│  │  (向量数据库)│                        │
│  └─────────────┘                        │
│                                         │
└─────────────────────────────────────────┘
```

**资源配置**：
| 组件 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| go-agent | 2 核 | 4GB | 10GB |
| PostgreSQL | 2 核 | 4GB | 50GB |
| Qdrant | 2 核 | 4GB | 50GB |
| **合计** | **6 核** | **12GB** | **110GB** |

### 1.2 高可用架构

**适用场景**：中大规模生产环境

```
┌─────────────────────────────────────────────────────────┐
│                  高可用架构                              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│                    ┌─────────────┐                      │
│                    │   Nginx     │                      │
│                    │  (负载均衡)  │                      │
│                    └──────┬──────┘                      │
│                           │                             │
│         ┌─────────────────┼─────────────────┐           │
│         │                 │                 │           │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐     │
│  │  go-agent   │  │  go-agent   │  │  go-agent   │     │
│  │  实例 1      │  │  实例 2      │  │  实例 3      │     │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘     │
│         │                │                │             │
│         └────────────────┼────────────────┘             │
│                          │                              │
│         ┌────────────────┼────────────────┐             │
│         │                │                │             │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐     │
│  │  PostgreSQL │  │  PostgreSQL │  │  PostgreSQL │     │
│  │  主节点      │──▶│  从节点 1    │  │  从节点 2    │     │
│  └──────┬──────┘  └─────────────┘  └─────────────┘     │
│         │                                               │
│  ┌──────▼──────┐                                        │
│  │   Qdrant    │                                        │
│  │  集群 (3 节点) │                                        │
│  └─────────────┘                                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**资源配置**：
| 组件 | 实例数 | CPU/实例 | 内存/实例 | 磁盘/实例 |
|------|--------|----------|-----------|----------|
| go-agent | 3 | 2 核 | 4GB | 10GB |
| PostgreSQL | 3 | 4 核 | 8GB | 100GB |
| Qdrant | 3 | 4 核 | 8GB | 100GB |
| Nginx | 2 | 1 核 | 1GB | 5GB |
| **合计** | **11** | **38 核** | **77GB** | **845GB** |

### 1.3 混合云架构

**适用场景**：大规模、跨地域部署

```
┌─────────────────────────────────────────────────────────────┐
│                     混合云架构                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  阿里云 (北京)                        阿里云 (上海)           │
│  ┌─────────────┐                    ┌─────────────┐         │
│  │   Nginx     │                    │   Nginx     │         │
│  └──────┬──────┘                    └──────┬──────┘         │
│         │                                  │                │
│  ┌──────┴──────┐                    ┌──────┴──────┐         │
│  │  go-agent   │                    │  go-agent   │         │
│  │  ×3         │                    │  ×3         │         │
│  └──────┬──────┘                    └──────┬──────┘         │
│         │                                  │                │
│         └────────────┬─────────────────────┘                │
│                      │                                      │
│         ┌────────────┴─────────────┐                        │
│         │                          │                        │
│  ┌──────▼──────┐            ┌──────▼──────┐                 │
│  │  PostgreSQL │            │   Qdrant    │                 │
│  │  主集群      │            │   集群      │                 │
│  └─────────────┘            └─────────────┘                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 环境准备

### 2.1 系统要求

| 组件 | 操作系统 | 最低配置 | 推荐配置 |
|------|----------|----------|----------|
| go-agent | Linux/macOS | 2 核 4GB | 4 核 8GB |
| PostgreSQL | Linux | 2 核 4GB | 4 核 8GB |
| Qdrant | Linux | 2 核 4GB | 4 核 8GB |

### 2.2 依赖服务

#### 2.2.1 PostgreSQL 15+ with pgvector

```bash
# Docker 安装
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=memory_db \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  pgvector/pgvector:pg16

# 验证安装
docker exec -it postgres psql -U postgres -d memory_db -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

#### 2.2.2 Qdrant

```bash
# Docker 安装
docker run -d \
  --name qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v qdrant_data:/qdrant/storage \
  qdrant/qdrant:latest

# 验证安装
curl http://localhost:6333/
```

#### 2.2.3 Ollama (可选，本地嵌入)

```bash
# 安装 Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# 拉取嵌入模型
ollama pull nomic-embed-text

# 验证安装
curl http://localhost:11434/api/embeddings -d '{"model": "nomic-embed-text", "prompt": "test"}'
```

### 2.3 网络配置

#### 2.3.1 端口规划

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| go-agent | 8080 | HTTP | API 服务 |
| go-agent | 8081 | HTTP | Metrics |
| PostgreSQL | 5432 | TCP | 数据库 |
| Qdrant | 6333 | HTTP | REST API |
| Qdrant | 6334 | gRPC | gRPC API |
| Ollama | 11434 | HTTP | 嵌入服务 |

#### 2.3.2 防火墙配置

```bash
# 允许 API 访问
sudo ufw allow 8080/tcp

# 允许 Metrics 访问（仅内网）
sudo ufw allow from 10.0.0.0/8 to any port 8081

# 允许数据库访问（仅应用服务器）
sudo ufw allow from 10.0.0.0/8 to any port 5432

# 允许 Qdrant 访问（仅应用服务器）
sudo ufw allow from 10.0.0.0/8 to any port 6333
```

---

## 3. 容器化部署

### 3.1 Dockerfile

创建 `Dockerfile`：

```dockerfile
# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git

# 复制代码
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 CA 证书
RUN apk --no-cache add ca-certificates

# 复制构建产物
COPY --from=builder /build/main .

# 暴露端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# 运行应用
CMD ["./main"]
```

### 3.2 docker-compose.yml

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  go-agent:
    build: .
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/memory_db?sslmode=disable
      - QDRANT_URL=http://qdrant:6333
      - QDRANT_COLLECTION=memory
      - EMBED_PROVIDER=ollama
      - EMBED_MODEL=nomic-embed-text
      - OLLAMA_HOST=http://ollama:11434
    depends_on:
      postgres:
        condition: service_healthy
      qdrant:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - memory-network

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=memory_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - memory-network

  qdrant:
    image: qdrant/qdrant:latest
    volumes:
      - qdrant_data:/qdrant/storage
    ports:
      - "6333:6333"
      - "6334:6334"
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:6333/ || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - memory-network

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"
    restart: unless-stopped
    networks:
      - memory-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - go-agent
    restart: unless-stopped
    networks:
      - memory-network

volumes:
  postgres_data:
  qdrant_data:
  ollama_data:

networks:
  memory-network:
    driver: bridge
```

### 3.3 nginx.conf

创建 `nginx.conf`：

```nginx
events {
    worker_connections 1024;
}

http {
    upstream go_agent {
        server go-agent:8080;
    }

    server {
        listen 80;
        server_name _;

        location / {
            proxy_pass http://go_agent;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        location /metrics {
            proxy_pass http://go_agent:8081;
            proxy_set_header Host $host;
        }

        location /health {
            proxy_pass http://go_agent;
            access_log off;
        }
    }
}
```

### 3.4 部署命令

```bash
# 构建并启动
docker-compose up -d --build

# 查看日志
docker-compose logs -f go-agent

# 检查服务状态
docker-compose ps

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

---

## 4. Kubernetes 部署

### 4.1 Namespace 配置

创建 `k8s/namespace.yaml`：

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: memory-system
  labels:
    name: memory-system
```

### 4.2 ConfigMap 配置

创建 `k8s/configmap.yaml`：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-agent-config
  namespace: memory-system
data:
  DATABASE_URL: "postgres://postgres:postgres@postgres.memory-system.svc.cluster.local:5432/memory_db?sslmode=disable"
  QDRANT_URL: "http://qdrant.memory-system.svc.cluster.local:6333"
  QDRANT_COLLECTION: "memory"
  EMBED_PROVIDER: "ollama"
  EMBED_MODEL: "nomic-embed-text"
  OLLAMA_HOST: "http://ollama.memory-system.svc.cluster.local:11434"
  LOG_LEVEL: "info"
  METRICS_PORT: "8081"
```

### 4.3 Secret 配置

创建 `k8s/secret.yaml`：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: go-agent-secret
  namespace: memory-system
type: Opaque
stringData:
  POSTGRES_PASSWORD: "your-secure-password"
  API_KEY: "your-api-key"
```

### 4.4 Deployment 配置

创建 `k8s/deployment.yaml`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-agent
  namespace: memory-system
  labels:
    app: go-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-agent
  template:
    metadata:
      labels:
        app: go-agent
    spec:
      containers:
      - name: go-agent
        image: registry.marecogito.ai/go-agent:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: metrics
        envFrom:
        - configMapRef:
            name: go-agent-config
        - secretRef:
            name: go-agent-secret
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/config
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: go-agent-config
```

### 4.5 Service 配置

创建 `k8s/service.yaml`：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-agent
  namespace: memory-system
  labels:
    app: go-agent
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 8081
    targetPort: 8081
    name: metrics
  selector:
    app: go-agent
---
apiVersion: v1
kind: Service
metadata:
  name: go-agent-external
  namespace: memory-system
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    name: http
  selector:
    app: go-agent
```

### 4.6 HPA 自动扩缩容

创建 `k8s/hpa.yaml`：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: go-agent-hpa
  namespace: memory-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: go-agent
  minReplicas: 3
  maxReplicas: 10
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
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
```

### 4.7 部署命令

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 创建配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 创建依赖服务（PostgreSQL, Qdrant 等）
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/qdrant.yaml

# 部署应用
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml

# 查看部署状态
kubectl get pods -n memory-system
kubectl get svc -n memory-system
kubectl get hpa -n memory-system

# 查看日志
kubectl logs -l app=go-agent -n memory-system -f

# 扩缩容
kubectl scale deployment go-agent --replicas=5 -n memory-system
```

---

## 5. 配置管理

### 5.1 配置文件结构

```
config/
├── default.yaml          # 默认配置
├── development.yaml      # 开发环境配置
├── staging.yaml          # 测试环境配置
└── production.yaml       # 生产环境配置
```

### 5.2 配置示例

创建 `config/production.yaml`：

```yaml
server:
  port: 8080
  metrics_port: 8081
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: postgres
  url: ${DATABASE_URL}
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 5m

qdrant:
  url: ${QDRANT_URL}
  collection: ${QDRANT_COLLECTION}
  timeout: 10s

embed:
  provider: ${EMBED_PROVIDER}
  model: ${EMBED_MODEL}
  ollama_host: ${OLLAMA_HOST}
  timeout: 30s

memory:
  max_size: 200000
  ttl: 720h
  duplicate_similarity: 0.97
  drift_threshold: 0.90

log:
  level: ${LOG_LEVEL}
  format: json
  output: stdout

metrics:
  enabled: true
  port: ${METRICS_PORT}
```

### 5.3 密钥管理

#### 5.3.1 使用环境变量

```bash
# .env 文件（不要提交到版本控制）
DATABASE_URL=postgres://user:password@host:5432/db
API_KEY=your-api-key
SECRET_KEY=your-secret-key
```

#### 5.3.2 使用 Vault

```bash
# 写入密钥
vault kv put secret/go-agent \
  database_url="postgres://..." \
  api_key="your-api-key"

# 读取密钥
vault kv get secret/go-agent

# 在 Kubernetes 中使用
kubectl exec -it go-agent-pod -- env | grep VAULT
```

### 5.4 配置热更新

```go
// 使用 viper 实现配置热更新
import "github.com/spf13/viper"

func loadConfig() *viper.Viper {
    v := viper.New()
    v.SetConfigFile("config/production.yaml")
    v.AutomaticEnv()
    
    v.WatchConfig()
    v.OnConfigChange(func(e fsnotify.Event) {
        log.Println("配置文件已更新")
        // 重新加载配置
    })
    
    return v
}
```

---

## 6. 附录

### 6.1 部署检查清单

#### 部署前

- [ ] 环境准备完成
- [ ] 依赖服务就绪
- [ ] 配置文件已准备
- [ ] 密钥已生成
- [ ] 网络配置完成

#### 部署中

- [ ] 镜像构建成功
- [ ] 容器启动正常
- [ ] 健康检查通过
- [ ] 日志无错误
- [ ] 监控指标正常

#### 部署后

- [ ] 功能测试通过
- [ ] 性能测试通过
- [ ] 备份配置完成
- [ ] 监控告警配置
- [ ] 文档已更新

### 6.2 常用命令

```bash
# Docker
docker-compose up -d
docker-compose logs -f
docker-compose ps

# Kubernetes
kubectl apply -f k8s/
kubectl get pods -n memory-system
kubectl logs -l app=go-agent -n memory-system
kubectl scale deployment go-agent --replicas=5 -n memory-system

# 配置
envsubst < config.template > config.yaml
vault kv put secret/go-agent key=value
```

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 9 日*  
*维护：MareMind 项目基础设施团队*
