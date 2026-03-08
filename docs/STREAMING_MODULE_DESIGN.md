# 流式传输 Module 设计

**项目**: Lattice - Go AI Agent 开发框架  
**版本**: 1.0.0  
**创建日期**: 2026 年 3 月 8 日  
**状态**: 设计方案  
**相关任务**: EXT-017

---

## 📋 目录

1. [概述](#概述)
2. [Module 架构](#module-架构)
3. [StreamModule 实现](#streammodule-实现)
4. [StreamHandler 实现](#streamhandler-实现)
5. [前端实现](#前端实现)
6. [使用示例](#使用示例)
7. [实施计划](#实施计划)

---

## 概述

### 背景

流式传输功能允许用户实时看到 AI 响应逐字生成，显著提升用户体验。

### 设计原则

| 原则 | 说明 |
|------|------|
| **模块化** | 通过 ADK Module 自动配置 |
| **多协议支持** | SSE + WebSocket |
| **向后兼容** | 不影响现有功能 |
| **高性能** | 支持并发连接 |

---

## Module 架构

### 整体架构

```
ADK Kit
    │
    └─ WithModule(StreamModule)
          │
          ▼
    StreamModule.Provision()
          │
          └─ kit.UseHTTPHandlerProvider(streamHandlerProvider)
                │
                ▼
          BuildAgent()
                │
                ▼
          HTTPHandlerProvider()
                │
                ├─ SSEHandler (/api/stream)
                └─ WSHandler (/api/ws)
```

### 组件关系

```
StreamModule
    │
    ├─ StreamConfig（配置）
    │   ├─ EnableSSE
    │   ├─ EnableWebSocket
    │   └─ Port
    │
    └─ StreamHandler（处理器）
        ├─ SSE Chat Handler
        └─ WebSocket Chat Handler
```

---

## StreamModule 实现

### Module 结构

```go
// src/adk/modules/stream_module.go
package modules

import (
    "context"
    "net/http"
    
    kit "github.com/Protocol-Lattice/go-agent/src/adk"
)

// StreamModule 流式传输模块
type StreamModule struct {
    name    string
    enabled bool
    config  StreamConfig
}

// StreamConfig 流式传输配置
type StreamConfig struct {
    EnableSSE       bool   // 启用 SSE
    EnableWebSocket bool   // 启用 WebSocket
    Port            int    // 服务器端口
    CorsOrigins     []string // 允许的 CORS 来源
}

// 构造函数
func NewStreamModule(name string, config StreamConfig) *StreamModule {
    return &StreamModule{
        name:    name,
        enabled: true,
        config:  config,
    }
}

// 实现 Module 接口
func (m *StreamModule) Name() string {
    return m.name
}

// Provision 注册 HTTP Handler Provider
func (m *StreamModule) Provision(ctx context.Context, kit *kit.AgentDevelopmentKit) error {
    if !m.enabled {
        return nil
    }
    
    // 创建 HTTP Handler Provider
    provider := func(ctx context.Context) ([]kit.HTTPHandler, error) {
        // 创建 Agent（用于流式生成）
        agent, err := kit.BuildAgent(ctx)
        if err != nil {
            return nil, err
        }
        
        // 创建流式处理器
        handler := NewStreamHandler(agent, StreamHandlerConfig{
            CorsOrigins: m.config.CorsOrigins,
        })
        
        var handlers []kit.HTTPHandler
        
        // 添加 SSE 端点
        if m.config.EnableSSE {
            handlers = append(handlers, kit.HTTPHandler{
                Method:  "GET",
                Path:    "/api/stream",
                Handler: handler.StreamChat,
            })
        }
        
        // 添加 WebSocket 端点
        if m.config.EnableWebSocket {
            handlers = append(handlers, kit.HTTPHandler{
                Method:  "WS",
                Path:    "/api/ws",
                Handler: handler.WSChat,
            })
        }
        
        return handlers, nil
    }
    
    // 注册 Provider
    kit.UseHTTPHandlerProvider(provider)
    
    return nil
}
```

---

## StreamHandler 实现

### SSE Handler

```go
// handlers/stream_handler.go
package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    
    "github.com/Protocol-Lattice/go-agent"
)

// StreamMessage SSE/WebSocket 消息格式
type StreamMessage struct {
    Type     string `json:"type"`      // "token" | "done" | "error"
    Content  string `json:"content"`   // 增量文本
    FullText string `json:"fullText"`  // 完整文本（仅 done 时）
    Error    string `json:"error"`     // 错误信息（仅 error 时）
}

// StreamHandler 流式处理器
type StreamHandler struct {
    agent  *agent.Agent
    config StreamHandlerConfig
}

type StreamHandlerConfig struct {
    CorsOrigins []string
}

func NewStreamHandler(agent *agent.Agent, config StreamHandlerConfig) *StreamHandler {
    return &StreamHandler{
        agent:  agent,
        config: config,
    }
}

// StreamChat SSE 流式聊天处理器
func (h *StreamHandler) StreamChat(w http.ResponseWriter, r *http.Request) {
    // 1. 设置 SSE 响应头
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("X-Accel-Buffering", "no") // Nginx 禁用缓冲
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    
    // 2. 获取请求参数
    sessionID := r.URL.Query().Get("session_id")
    message := r.URL.Query().Get("message")
    
    if sessionID == "" || message == "" {
        h.sendError(w, flusher, "Missing session_id or message")
        return
    }
    
    // 3. 调用 Agent 流式生成
    ctx := r.Context()
    stream, err := h.agent.GenerateStream(ctx, sessionID, message)
    if err != nil {
        h.sendError(w, flusher, "Generate error: "+err.Error())
        return
    }
    
    // 4. 实时推送流式数据
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
        fmt.Fprintf(w, "data: %s\n\n", h.toJSON(msg))
        flusher.Flush()
        
        if chunk.Done || chunk.Err != nil {
            break
        }
    }
}

func (h *StreamHandler) sendError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
    msg := StreamMessage{
        Type:  "error",
        Error: errMsg,
    }
    fmt.Fprintf(w, "data: %s\n\n", h.toJSON(msg))
    flusher.Flush()
}

func (h *StreamHandler) toJSON(v interface{}) string {
    data, _ := json.Marshal(v)
    return string(data)
}
```

### WebSocket Handler

```go
// handlers/ws_handler.go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/gorilla/websocket"
    "github.com/Protocol-Lattice/go-agent"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境应限制来源
    },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// WSMessage WebSocket 消息格式
type WSMessage struct {
    Type      string                 `json:"type"`
    Content   string                 `json:"content,omitempty"`
    FullText  string                 `json:"fullText,omitempty"`
    Error     string                 `json:"error,omitempty"`
    SessionID string                 `json:"session_id"`
    Message   string                 `json:"message"`
}

// WSChat WebSocket 聊天处理器
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
                return
            }
            break
        }
        
        // 解析请求
        var req WSMessage
        if err := json.Unmarshal(message, &req); err != nil {
            continue
        }
        
        // 调用 Agent 流式生成
        ctx := r.Context()
        stream, err := h.agent.GenerateStream(ctx, req.SessionID, req.Message)
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

---

## 前端实现

### SSE 客户端

```html
<!-- static/stream.html -->
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
            font-family: monospace;
            background: #f5f5f5;
            padding: 20px;
            border-radius: 8px;
            min-height: 200px;
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
        <h1>🤖 Agent 流式对话</h1>
        
        <div class="input-area">
            <input 
                type="text" 
                id="message" 
                placeholder="输入消息..." 
                autocomplete="off"
            >
            <button id="sendBtn" onclick="sendMessage()">发送</button>
        </div>
        
        <div id="response" class="response">
            <p style="color: #999;">请输入消息开始对话...</p>
        </div>
    </div>

    <script>
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
    </script>
</body>
</html>
```

---

## 使用示例

### 完整示例

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/Protocol-Lattice/go-agent/src/adk"
    "github.com/Protocol-Lattice/go-agent/src/adk/modules"
    "github.com/Protocol-Lattice/go-agent/src/memory"
    "github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建 ADK，注册流式传输模块
    kit, err := adk.New(ctx,
        // 基础模块
        adk.WithModule(modules.NewModelModule("model", func(ctx context.Context) (models.Agent, error) {
            return models.NewGeminiLLM(ctx, "gemini-2.5-pro", "")
        })),
        
        adk.WithModule(modules.NewMemoryModule("memory", func(ctx context.Context) (kit.MemoryBundle, error) {
            bank := memory.NewMemoryBankWithStore(memory.NewInMemoryStore())
            mem := memory.NewSessionMemory(bank, 8)
            mem.WithEmbedder(memory.AutoEmbedder())
            
            shared := func(local string, spaces ...string) *memory.SharedSession {
                return memory.NewSharedSession(mem, local, spaces...)
            }
            
            return kit.MemoryBundle{Session: mem, Shared: shared}, nil
        })),
        
        // 流式传输模块
        adk.WithModule(modules.NewStreamModule("stream", modules.StreamConfig{
            EnableSSE:       true,
            EnableWebSocket: true,
            Port:            8080,
            CorsOrigins:     []string{"http://localhost:3000"},
        })),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 获取 HTTP Handlers
    handlers, err := kit.HTTPHandlerProvider()(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 注册路由
    for _, h := range handlers {
        switch h.Method {
        case "GET":
            http.HandleFunc(h.Path, h.Handler)
        case "WS":
            http.HandleFunc(h.Path, h.Handler)
        }
    }
    
    // 4. 提供静态文件（前端界面）
    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    // 5. 启动服务器
    log.Printf("服务器启动在 http://localhost:%d", 8080)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 前端调用示例

```javascript
// 使用 SSE
const eventSource = new EventSource('/api/stream?session_id=1&message=' + encodeURIComponent('你好'));

eventSource.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    
    if (msg.type === 'token') {
        console.log('收到 token:', msg.content);
    }
    
    if (msg.type === 'done') {
        console.log('完成:', msg.fullText);
        eventSource.close();
    }
};
```

---

## 实施计划

### 阶段 1：核心实现（4 小时）

- [ ] **M1.1**：StreamModule 结构定义（30 分钟）
- [ ] **M1.2**：StreamHandler 实现（2 小时）
  - [ ] SSE Handler
  - [ ] WebSocket Handler
- [ ] **M1.3**：消息格式定义（30 分钟）
- [ ] **M1.4**：错误处理（30 分钟）
- [ ] **M1.5**：单元测试（1 小时）

### 阶段 2：前端实现（2 小时）

- [ ] **M2.1**：SSE 客户端（1 小时）
- [ ] **M2.2**：WebSocket 客户端（可选）（1 小时）
- [ ] **M2.3**：UI 样式和交互（1 小时）

### 阶段 3：集成测试（1 小时）

- [ ] **M3.1**：端到端测试（30 分钟）
- [ ] **M3.2**：性能测试（30 分钟）
  - [ ] 并发连接测试（10+ 同时连接）
  - [ ] 延迟测试（首字 < 500ms）

### 阶段 4：文档和示例（1 小时）

- [ ] **M4.1**：使用文档（30 分钟）
- [ ] **M4.2**：示例代码（30 分钟）

**总计**：8 小时（1 个工作日）

---

## 总结

### 模块优势

| 优势 | 说明 |
|------|------|
| **模块化** | 通过 ADK Module 自动配置 |
| **多协议** | SSE + WebSocket 支持 |
| **零侵入** | 不影响现有功能 |
| **高性能** | 支持并发连接 |

### 下一步

- 实施 StreamModule
- 编写测试
- 编写文档

---

*文档版本：1.0.0*  
*最后更新：2026 年 3 月 8 日*  
*维护：MareMind 项目基础设施团队*
