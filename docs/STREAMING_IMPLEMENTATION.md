# 端到端流式传输实施方案

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 7 日  
**状态**: 设计方案  
**相关任务**: EXT-017

---

## 📋 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [SSE 实现方案](#sse-实现方案)
4. [WebSocket 实现方案](#websocket-实现方案)
5. [HTTP Chunked 方案](#http-chunked-方案)
6. [完整代码示例](#完整代码示例)
7. [性能优化](#性能优化)
8. [错误处理与重连](#错误处理与重连)
9. [安全考虑](#安全考虑)
10. [测试计划](#测试计划)

---

## 概述

### 背景

在 AI 对话应用中，用户希望实时看到 AI 的响应生成过程，而不是等待完整的响应。流式传输可以让用户看到文字逐字出现，类似 ChatGPT 的体验，显著提升用户体验。

### 目标

- ✅ **零等待** - 立即可见输出
- ✅ **打字机效果** - 逐字显示更自然
- ✅ **实时反馈** - 用户知道正在处理
- ✅ **低延迟** - 首字延迟 < 500ms
- ✅ **高并发** - 支持 10+ 同时连接

### 技术方案对比

| 方案 | 适用场景 | 复杂度 | 推荐度 |
|------|----------|--------|--------|
| **SSE** | 单向推送（服务器→客户端） | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **WebSocket** | 双向实时交互 | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **HTTP Chunked** | 简单流式需求 | ⭐ | ⭐⭐⭐ |

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    前端（浏览器）                            │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ EventSource │  │ WebSocket   │  │ 打字机效果 UI       │ │
│  │ (SSE)       │  │ Client      │  │ - 逐字显示          │ │
│  │             │  │             │  │ - 自动滚动          │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ SSE / WebSocket
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    HTTP Server (Go)                          │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ SSE Handler │  │ WS Handler  │  │ StreamMessage       │ │
│  │ /api/stream │  │ /api/ws     │  │ 序列化              │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ <-chan StreamChunk
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    Agent.GenerateStream()                    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ - 调用 LLM 流式 API                                   │   │
│  │ - 累积完整响应                                       │   │
│  │ - 缓存结果                                          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ <-chan StreamChunk
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    LLM Provider                              │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Gemini API  │  │ OpenAI API  │  │ Ollama API          │ │
│  │ (流式)      │  │ (流式)      │  │ (流式)              │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

```
用户输入消息
    │
    ▼
HTTP POST /api/stream
    │
    ▼
设置 SSE 响应头
    │
    ▼
Agent.GenerateStream()
    │
    ▼
LLM 流式返回 tokens
    │
    ▼
逐个发送到前端（SSE）
    │
    ▼
前端 EventSource 接收
    │
    ▼
逐字显示 + 自动滚动
    │
    ▼
完成/错误 → 关闭连接
```

---

## SSE 实现方案 ⭐ 推荐

### 什么是 SSE？

**Server-Sent Events (SSE)** 是一种服务器推送技术，允许服务器向浏览器推送实时更新。

**优点**：
- ✅ 原生支持，无需额外库
- ✅ 自动重连
- ✅ 简单轻量
- ✅ 基于 HTTP，防火墙友好

**缺点**：
- ❌ 单向通信（仅服务器→客户端）
- ❌ 不支持二进制数据

### 后端实现

#### 1. StreamMessage 结构定义

```go
// handlers/stream_handler.go
package handlers

import (
    "encoding/json"
)

// StreamMessage SSE 消息格式
type StreamMessage struct {
    Type     string `json:"type"`      // "token" | "done" | "error"
    Content  string `json:"content"`   // 增量文本
    FullText string `json:"fullText"`  // 完整文本（仅 done 时）
    Error    string `json:"error"`     // 错误信息（仅 error 时）
}

// MarshalSSE 将消息序列化为 SSE 格式
func (m *StreamMessage) MarshalSSE() string {
    data, _ := json.Marshal(m)
    return "data: " + string(data) + "\n\n"
}
```

#### 2. StreamHandler 结构

```go
type StreamHandler struct {
    agent *agent.Agent
}

func NewStreamHandler(agent *agent.Agent) *StreamHandler {
    return &StreamHandler{agent: agent}
}
```

#### 3. SSE 响应头设置

```go
func (h *StreamHandler) StreamChat(w http.ResponseWriter, r *http.Request) {
    // 1. 设置 SSE 响应头
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("X-Accel-Buffering", "no") // Nginx 禁用缓冲
    
    // 2. 获取 Flush 接口
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    
    // ... 继续处理
}
```

#### 4. 主处理器

```go
func (h *StreamHandler) StreamChat(w http.ResponseWriter, r *http.Request) {
    // 设置响应头（见上）
    
    // 3. 获取请求参数
    sessionID := r.URL.Query().Get("session_id")
    message := r.URL.Query().Get("message")
    
    if sessionID == "" || message == "" {
        h.sendError(w, flusher, "Missing session_id or message")
        return
    }
    
    // 4. 调用 Agent 流式生成
    ctx := r.Context()
    stream, err := h.agent.GenerateStream(ctx, sessionID, message)
    if err != nil {
        h.sendError(w, flusher, "Generate error: "+err.Error())
        return
    }
    
    // 5. 实时推送流式数据
    for chunk := range stream {
        msg := StreamMessage{
            Type:    "token",
            Content: chunk.Delta,
        }
        
        if chunk.Done {
            msg.Type = "done"
            msg.FullText = chunk.FullText
        }
        
        if chunk.Err != nil {
            msg.Type = "error"
            msg.Error = chunk.Err.Error()
        }
        
        // 发送 SSE 消息
        fmt.Fprint(w, msg.MarshalSSE())
        flusher.Flush()
        
        if chunk.Done || chunk.Err != nil {
            break
        }
    }
}
```

#### 5. 错误处理

```go
func (h *StreamHandler) sendError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
    msg := StreamMessage{
        Type:  "error",
        Error: errMsg,
    }
    fmt.Fprint(w, msg.MarshalSSE())
    flusher.Flush()
}
```

### 前端实现

#### 1. HTML 结构

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Agent 流式对话</title>
    <style>
        .response {
            white-space: pre-wrap;
            line-height: 1.6;
        }
        
        .typing-indicator {
            display: inline-block;
            width: 2px;
            height: 1.2em;
            background: #007bff;
            margin-left: 2px;
            animation: blink 1s infinite;
            vertical-align: text-bottom;
        }
        
        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0; }
        }
        
        .error {
            color: #dc3545;
            background: #f8d7da;
            padding: 10px;
            border-radius: 4px;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <input type="text" id="message" placeholder="输入消息...">
        <button onclick="sendMessage()" id="sendBtn">发送</button>
        
        <div id="response" class="response"></div>
    </div>
    
    <script src="stream.js"></script>
</body>
</html>
```

#### 2. JavaScript 客户端

```javascript
// static/js/stream.js
let eventSource = null;
let currentResponse = '';
let isStreaming = false;

const responseDiv = document.getElementById('response');
const messageInput = document.getElementById('message');
const sendBtn = document.getElementById('sendBtn');

async function sendMessage() {
    const message = messageInput.value.trim();
    if (!message || isStreaming) return;
    
    // 清空之前的响应
    currentResponse = '';
    responseDiv.innerHTML = '';
    isStreaming = true;
    sendBtn.disabled = true;
    messageInput.disabled = true;
    
    // 创建打字机效果的光标
    const cursor = document.createElement('span');
    cursor.className = 'typing-indicator';
    responseDiv.appendChild(cursor);
    
    // 创建 EventSource 连接 SSE
    const url = `/api/stream?session_id=${Date.now()}&message=${encodeURIComponent(message)}`;
    eventSource = new EventSource(url);
    
    eventSource.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        
        if (msg.type === 'token') {
            // 逐字显示（在光标前插入）
            currentResponse += msg.content;
            responseDiv.innerHTML = currentResponse;
            responseDiv.appendChild(cursor);
            
            // 滚动到底部
            window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
        }
        
        if (msg.type === 'done') {
            // 完成，移除光标
            cursor.remove();
            eventSource.close();
            cleanup();
        }
        
        if (msg.type === 'error') {
            cursor.remove();
            responseDiv.innerHTML += `<div class="error">❌ 错误：${msg.error}</div>`;
            eventSource.close();
            cleanup();
        }
    };
    
    eventSource.onerror = () => {
        cursor.remove();
        responseDiv.innerHTML += `<div class="error">❌ 连接错误</div>`;
        if (eventSource) {
            eventSource.close();
        }
        cleanup();
    };
    
    // 清空输入框
    messageInput.value = '';
}

function cleanup() {
    isStreaming = false;
    sendBtn.disabled = false;
    messageInput.disabled = false;
    messageInput.focus();
}

// 支持回车发送
messageInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && !isStreaming) {
        sendMessage();
    }
});
```

---

## WebSocket 实现方案

### 适用场景

- 需要客户端→服务器实时交互
- 如：用户中途打断、修改参数等
- 全双工通信需求

### 后端实现

#### 1. 引入 WebSocket 库

```bash
go get github.com/gorilla/websocket
```

#### 2. WebSocket Upgrader 配置

```go
// handlers/ws_handler.go
package handlers

import (
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // 生产环境应限制来源
        return true
    },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}
```

#### 3. WSMessage 结构

```go
type WSMessage struct {
    Type      string `json:"type"`
    Content   string `json:"content,omitempty"`
    FullText  string `json:"fullText,omitempty"`
    Error     string `json:"error,omitempty"`
    SessionID string `json:"session_id"`
    Message   string `json:"message"`
}
```

#### 4. WebSocket 处理器

```go
func (h *StreamHandler) WSChat(w http.ResponseWriter, r *http.Request) {
    // 1. 升级为 WebSocket 连接
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // 2. 消息循环
    for {
        // 接收客户端消息
        _, message, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
        
        // 解析请求
        var req WSMessage
        if err := json.Unmarshal(message, &req); err != nil {
            continue
        }
        
        // 调用 Agent 流式生成
        stream, err := h.agent.GenerateStream(r.Context(), req.SessionID, req.Message)
        if err != nil {
            conn.WriteJSON(WSMessage{Type: "error", Error: err.Error()})
            continue
        }
        
        // 实时推送
        for chunk := range stream {
            msg := WSMessage{
                Type:    "token",
                Content: chunk.Delta,
            }
            
            if chunk.Done {
                msg.Type = "done"
                msg.FullText = chunk.FullText
            }
            
            if err := conn.WriteJSON(msg); err != nil {
                break
            }
        }
    }
}
```

### 前端实现

```javascript
// static/js/ws_client.js
let ws = null;
let currentResponse = '';

function connectWebSocket() {
    ws = new WebSocket('ws://localhost:8080/api/ws');
    
    ws.onopen = () => {
        console.log('WebSocket 连接已建立');
    };
    
    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        
        if (msg.type === 'token') {
            currentResponse += msg.content;
            updateDisplay();
        }
        
        if (msg.type === 'done') {
            finalizeDisplay();
        }
        
        if (msg.type === 'error') {
            showError(msg.error);
        }
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket 错误:', error);
    };
    
    ws.onclose = () => {
        console.log('WebSocket 连接已关闭');
        // 自动重连
        setTimeout(connectWebSocket, 3000);
    };
}

function sendMessage(message) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: 'message',
            session_id: 'session1',
            message: message
        }));
    }
}

// 初始化连接
connectWebSocket();
```

---

## HTTP Chunked 方案

### 适用场景

- 简单的流式需求
- 不需要 SSE 的额外功能
- 兼容性要求高

### 后端实现

```go
func (h *StreamHandler) ChunkedChat(w http.ResponseWriter, r *http.Request) {
    // 设置 Chunked 传输
    w.Header().Set("Content-Type", "application/x-ndjson")
    w.Header().Set("Transfer-Encoding", "chunked")
    w.Header().Set("Cache-Control", "no-cache")
    
    flusher, _ := w.(http.Flusher)
    
    sessionID := r.URL.Query().Get("session_id")
    message := r.URL.Query().Get("message")
    
    // 调用流式生成
    stream, _ := h.agent.GenerateStream(r.Context(), sessionID, message)
    
    for chunk := range stream {
        // 发送 NDJSON 格式
        json.NewEncoder(w).Encode(map[string]string{
            "type":    "token",
            "content": chunk.Delta,
        })
        flusher.Flush() // 立即发送
        
        if chunk.Done {
            break
        }
    }
}
```

### 前端实现

```javascript
async function fetchStream(message) {
    const response = await fetch(`/api/chunked?session_id=1&message=${encodeURIComponent(message)}`);
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        
        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');
        
        for (const line of lines) {
            if (line.trim()) {
                const data = JSON.parse(line);
                if (data.type === 'token') {
                    currentResponse += data.content;
                    updateDisplay();
                }
            }
        }
    }
}
```

---

## 完整代码示例

### 完整后端代码

```go
// main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    
    "github.com/Protocol-Lattice/go-agent"
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/models"
    "your-module/handlers"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 LLM
    llm, err := models.NewGeminiLLM(ctx, "gemini-2.5-pro", "You are a helpful assistant.")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建记忆系统
    mem := memory.NewSessionMemory(
        memory.NewMemoryBankWithStore(memory.NewInMemoryStore()),
        8,
    )
    
    // 3. 创建 Agent
    agent, err := agent.New(agent.Options{
        Model:  llm,
        Memory: mem,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 创建流式处理器
    streamHandler := handlers.NewStreamHandler(agent)
    
    // 5. 设置路由
    http.HandleFunc("/api/stream", streamHandler.StreamChat) // SSE
    http.HandleFunc("/api/ws", streamHandler.WSChat)         // WebSocket
    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    // 6. 启动服务器
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("服务器启动在 http://localhost:%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

### 完整前端代码

详见前面 SSE 前端实现部分。

---

## 性能优化

### 1. 缓冲优化

```go
// 使用带缓冲的通道，减少阻塞
ch := make(chan StreamChunk, 32) // 增加缓冲区（默认 16）
```

### 2. 批量发送

```go
// 累积多个 token 后发送（减少网络往返）
var batch strings.Builder
count := 0

for chunk := range stream {
    batch.WriteString(chunk.Delta)
    count++
    
    if count >= 5 { // 每 5 个 token 发送一次
        msg := StreamMessage{
            Type:    "token",
            Content: batch.String(),
        }
        fmt.Fprint(w, msg.MarshalSSE())
        flusher.Flush()
        
        batch.Reset()
        count = 0
    }
}

// 发送剩余内容
if batch.Len() > 0 {
    msg := StreamMessage{
        Type:    "token",
        Content: batch.String(),
    }
    fmt.Fprint(w, msg.MarshalSSE())
    flusher.Flush()
}
```

### 3. 超时控制

```go
// 设置请求超时
ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
defer cancel()

stream, err := h.agent.GenerateStream(ctx, sessionID, message)
```

### 4. 内存管理

```go
// strings.Builder 预分配
var b strings.Builder
b.Grow(1024) // 预分配 1KB
```

### 5. Nginx 配置

```nginx
# 禁用响应缓冲
location /api/stream {
    proxy_pass http://localhost:8080;
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 60s;
    
    # SSE 要求
    proxy_set_header Content-Type text/event-stream;
    proxy_set_header Cache-Control no-cache;
    proxy_set_header Connection keep-alive;
}
```

---

## 错误处理与重连

### SSE 自动重连

```javascript
eventSource.onerror = () => {
    console.log('连接断开，3 秒后重连...');
    
    // EventSource 会自动重连，但我们可以手动控制
    setTimeout(() => {
        eventSource = new EventSource(url);
        // 重新绑定事件处理器
        eventSource.onmessage = ...;
    }, 3000);
};
```

### WebSocket 重连

```javascript
function connectWebSocket() {
    ws = new WebSocket('ws://localhost:8080/api/ws');
    
    ws.onclose = () => {
        console.log('连接关闭，3 秒后重连...');
        setTimeout(connectWebSocket, 3000);
    };
    
    ws.onerror = (error) => {
        console.error('连接错误:', error);
        ws.close();
    };
}
```

### 降级策略

```javascript
async function sendMessage() {
    try {
        // 尝试 SSE 流式
        await sendViaSSE(message);
    } catch (error) {
        console.warn('SSE 失败，降级为普通请求');
        // 降级为普通 HTTP 请求
        await sendViaHTTP(message);
    }
}
```

---

## 安全考虑

### 1. CORS 配置

```go
// 生产环境应限制来源
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        allowed := []string{
            "https://yourdomain.com",
            "https://www.yourdomain.com",
        }
        origin := r.Header.Get("Origin")
        for _, a := range allowed {
            if origin == a {
                return true
            }
        }
        return false
    },
}
```

### 2. 认证机制

```go
func (h *StreamHandler) StreamChat(w http.ResponseWriter, r *http.Request) {
    // Token 验证
    token := r.URL.Query().Get("token")
    if !validateToken(token) {
        h.sendError(w, flusher, "Unauthorized")
        return
    }
    
    // ... 继续处理
}
```

### 3. 速率限制

```go
// 使用中间件限制请求频率
func rateLimit(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 实现速率限制逻辑
        if !allowRequest(r.RemoteAddr) {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next(w, r)
    }
}
```

---

## 测试计划

### 单元测试

```go
// handlers/stream_handler_test.go
func TestStreamHandler_SSE(t *testing.T) {
    // 创建测试 Agent
    agent := createTestAgent()
    handler := NewStreamHandler(agent)
    
    // 创建测试请求
    req := httptest.NewRequest("GET", "/api/stream?session_id=test&message=hello", nil)
    w := httptest.NewRecorder()
    
    // 执行请求
    handler.StreamChat(w, req)
    
    // 验证响应
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    // 验证 SSE 格式
    contentType := w.Header().Get("Content-Type")
    if contentType != "text/event-stream" {
        t.Errorf("Expected text/event-stream, got %s", contentType)
    }
}
```

### 性能测试

```bash
# 使用 ab 进行压力测试
ab -n 1000 -c 10 http://localhost:8080/api/stream?session_id=test&message=hello

# 验证指标
# - 首字延迟 < 500ms
# - 传输延迟 < 100ms/token
# - 内存占用 < 50MB/连接
```

### 端到端测试

```javascript
// tests/e2e/streaming.test.js
describe('流式传输', () => {
    it('应该实时显示响应', async () => {
        await page.goto('http://localhost:8080');
        await page.type('#message', '你好');
        await page.click('#sendBtn');
        
        // 等待第一个 token 出现
        await page.waitForFunction(() => {
            const response = document.getElementById('response');
            return response.textContent.length > 0;
        });
        
        // 验证响应内容
        const response = await page.$eval('#response', el => el.textContent);
        expect(response).toBeTruthy();
    });
});
```

---

## 总结

### 推荐方案

**SSE (Server-Sent Events)** 是最适合 AI 对话场景的方案：
- ✅ 简单易用
- ✅ 自动重连
- ✅ 轻量级
- ✅ 浏览器原生支持

### 实施步骤

1. **后端**：实现 SSE Handler（2 小时）
2. **前端**：实现 EventSource 客户端（2 小时）
3. **优化**：缓冲、批量发送（2 小时）
4. **测试**：功能测试 + 性能测试（2 小时）

### 预期效果

- **首字延迟** < 500ms
- **传输延迟** < 100ms/token
- **用户体验** 显著提升

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 7 日*  
*维护：MareMind 项目基础设施团队*
