# DeepSeek LLM 实现方案

**创建日期**: 2026 年 3 月 6 日  
**状态**: 设计方案（待实现）  
**兼容性**: OpenAI API 兼容  

---

## 📋 概述

DeepSeek（深度求索）提供 OpenAI 兼容的 API 接口，支持文本生成、多模态输入（图片）和流式输出。本文档提供完整的实现方案和使用示例。

---

## 🎯 特性支持

| 功能 | 状态 | 说明 |
|------|------|------|
| 基本文本生成 | ✅ | `Generate(ctx, prompt)` |
| 带文件生成 | ✅ | `GenerateWithFiles(ctx, prompt, files)` |
| 流式生成 | ✅ | `GenerateStream(ctx, prompt)` |
| 带文件流式 | ✅ | `GenerateStreamWithFiles(ctx, prompt, files)` |
| 图片理解 | ✅ | PNG/JPEG/GIF/WEBP |
| 视频理解 | ❌ | DeepSeek 暂不支持 |
| 文本文件内联 | ✅ | TXT/MD/JSON 等 |

---

## 🏗️ 架构设计

### 依赖关系

```
DeepSeekLLM
├── 基于：github.com/sashabaranov/go-openai
├── API 端点：https://api.deepseek.com（可配置）
└── 认证：DEEPSEEK_API_KEY
```

### 为什么使用 OpenAI 客户端？

DeepSeek API 与 OpenAI API **完全兼容**：
- 相同的请求/响应格式
- 相同的流式 API
- 相同的多模态消息结构（`MultiContent`）

因此可以直接复用 `go-openai` 库，只需修改 `BaseURL`。

---

## 📝 实现代码

### 文件位置

```
src/models/deepseek.go
```

### 完整实现

```go
package models

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// DeepSeekLLM implements the Agent interface using DeepSeek's API.
// DeepSeek uses OpenAI-compatible API endpoints.
type DeepSeekLLM struct {
	Client       *openai.Client
	Model        string
	PromptPrefix string
	BaseURL      string
}

// DeepSeekModels defines available DeepSeek models
type DeepSeekModels struct {
	// DeepSeek Chat models
	DeepSeekChat     = "deepseek-chat"      // 深度对话模型
	DeepSeekCoder    = "deepseek-coder"     // 代码专用模型
	DeepSeekV2       = "deepseek-v2"        // V2 版本
	DeepSeekV2_5     = "deepseek-v2.5"      // V2.5 版本
}

// NewDeepSeekLLM creates a new DeepSeek client.
// It reads DEEPSEEK_API_KEY from environment.
// Optional: DEEPSEEK_BASE_URL (defaults to https://api.deepseek.com)
func NewDeepSeekLLM(model string, promptPrefix string) (*DeepSeekLLM, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, errors.New("deepseek: missing DEEPSEEK_API_KEY environment variable")
	}

	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	// Create OpenAI client with custom base URL (DeepSeek is OpenAI-compatible)
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	client := openai.NewClientWithConfig(config)
	
	return &DeepSeekLLM{
		Client:       client,
		Model:        model,
		PromptPrefix: promptPrefix,
		BaseURL:      baseURL,
	}, nil
}

// Generate performs a single-turn completion.
func (d *DeepSeekLLM) Generate(ctx context.Context, prompt string) (any, error) {
	fullPrompt := prompt
	if d.PromptPrefix != "" {
		fullPrompt = d.PromptPrefix + "\n" + prompt
	}

	resp, err := d.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: d.Model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("deepseek generate: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("deepseek: no response")
	}
	return resp.Choices[0].Message.Content, nil
}

// GenerateWithFiles performs completion with file attachments (images, text, etc.).
func (d *DeepSeekLLM) GenerateWithFiles(ctx context.Context, prompt string, files []File) (any, error) {
	fullPrompt := prompt
	if d.PromptPrefix != "" {
		fullPrompt = d.PromptPrefix + "\n" + prompt
	}

	// Separate files by type
	var textFiles []File
	var mediaFiles []File

	for _, f := range files {
		mt := normalizeMIME(f.Name, f.MIME)

		if isImageOrVideoMIME(mt) && getOpenAIMimeType(mt) != "" {
			mediaFiles = append(mediaFiles, f)
		} else if isTextMIME(mt) {
			textFiles = append(textFiles, f)
		}
	}

	// If no media files, fall back to text-only approach
	if len(mediaFiles) == 0 {
		combined := combinePromptWithFiles(fullPrompt, textFiles)
		return d.Generate(ctx, combined)
	}

	// Build MultiContent message with text and media
	var contentParts []openai.ChatMessagePart

	// Add the text prompt (including inline text files)
	textPrompt := fullPrompt
	if len(textFiles) > 0 {
		textPrompt = combinePromptWithFiles(fullPrompt, textFiles)
	}

	contentParts = append(contentParts, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeText,
		Text: textPrompt,
	})

	// Add media files (images)
	for _, f := range mediaFiles {
		mt := normalizeMIME(f.Name, f.MIME)
		openaiMime := getOpenAIMimeType(mt)
		if openaiMime == "" {
			continue
		}

		// Encode as base64
		encoded := base64.StdEncoding.EncodeToString(f.Data)
		dataURL := fmt.Sprintf("data:%s;base64,%s", openaiMime, encoded)

		if strings.HasPrefix(openaiMime, "image/") {
			contentParts = append(contentParts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    dataURL,
					Detail: openai.ImageURLDetailAuto,
				},
			})
		}
	}

	resp, err := d.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: d.Model,
		Messages: []openai.ChatCompletionMessage{{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: contentParts,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("deepseek generateWithFiles: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("deepseek: no response")
	}
	return resp.Choices[0].Message.Content, nil
}

// GenerateStream uses DeepSeek's streaming chat completion API.
func (d *DeepSeekLLM) GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	fullPrompt := prompt
	if d.PromptPrefix != "" {
		fullPrompt = d.PromptPrefix + "\n" + prompt
	}

	stream, err := d.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: d.Model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		}},
		Stream: true,
	})
	if err != nil {
		return nil, fmt.Errorf("deepseek create stream: %w", err)
	}

	ch := make(chan StreamChunk, 16)
	go func() {
		defer close(ch)
		defer stream.Close()
		var sb strings.Builder
		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					ch <- StreamChunk{Done: true, FullText: sb.String()}
					return
				}
				ch <- StreamChunk{Done: true, FullText: sb.String(), Err: err}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					sb.WriteString(delta)
					ch <- StreamChunk{Delta: delta}
				}
			}
		}
	}()

	return ch, nil
}

// GenerateStreamWithFiles performs streaming completion with file attachments.
// This combines multi-modal input with streaming output.
func (d *DeepSeekLLM) GenerateStreamWithFiles(ctx context.Context, prompt string, files []File) (<-chan StreamChunk, error) {
	fullPrompt := prompt
	if d.PromptPrefix != "" {
		fullPrompt = d.PromptPrefix + "\n" + prompt
	}

	// Separate files by type
	var textFiles []File
	var mediaFiles []File

	for _, f := range files {
		mt := normalizeMIME(f.Name, f.MIME)

		if isImageOrVideoMIME(mt) && getOpenAIMimeType(mt) != "" {
			mediaFiles = append(mediaFiles, f)
		} else if isTextMIME(mt) {
			textFiles = append(textFiles, f)
		}
	}

	// Build MultiContent message
	var contentParts []openai.ChatMessagePart

	// Add the text prompt (including inline text files)
	textPrompt := fullPrompt
	if len(textFiles) > 0 {
		textPrompt = combinePromptWithFiles(fullPrompt, textFiles)
	}

	contentParts = append(contentParts, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeText,
		Text: textPrompt,
	})

	// Add media files (images)
	for _, f := range mediaFiles {
		mt := normalizeMIME(f.Name, f.MIME)
		openaiMime := getOpenAIMimeType(mt)
		if openaiMime == "" {
			continue
		}

		// Encode as base64
		encoded := base64.StdEncoding.EncodeToString(f.Data)
		dataURL := fmt.Sprintf("data:%s;base64,%s", openaiMime, encoded)

		if strings.HasPrefix(openaiMime, "image/") {
			contentParts = append(contentParts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    dataURL,
					Detail: openai.ImageURLDetailAuto,
				},
			})
		}
	}

	// Create streaming request with MultiContent
	stream, err := d.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: d.Model,
		Messages: []openai.ChatCompletionMessage{{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: contentParts,
		}},
		Stream: true,
	})
	if err != nil {
		return nil, fmt.Errorf("deepseek create stream with files: %w", err)
	}

	ch := make(chan StreamChunk, 16)
	go func() {
		defer close(ch)
		defer stream.Close()
		var sb strings.Builder
		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					ch <- StreamChunk{Done: true, FullText: sb.String()}
					return
				}
				ch <- StreamChunk{Done: true, FullText: sb.String(), Err: err}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					sb.WriteString(delta)
					ch <- StreamChunk{Delta: delta}
				}
			}
		}
	}()

	return ch, nil
}

// Ensure DeepSeekLLM implements the Agent interface
var _ Agent = (*DeepSeekLLM)(nil)
```

---

## 🔧 环境配置

### 环境变量

```bash
# 必需：DeepSeek API 密钥
export DEEPSEEK_API_KEY="your-api-key-here"

# 可选：自定义基础 URL（默认：https://api.deepseek.com）
export DEEPSEEK_BASE_URL="https://api.deepseek.com"
```

### 获取 API 密钥

1. 访问 [DeepSeek 开放平台](https://platform.deepseek.com/)
2. 注册/登录账号
3. 进入 API 密钥管理页面
4. 创建新的 API 密钥
5. 复制密钥并设置到环境变量

---

## 💻 使用示例

### 示例 1：基本文本生成

```go
package main

import (
	"context"
	"log"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
	ctx := context.Background()

	// 创建 DeepSeek 客户端
	llm, err := models.NewDeepSeekLLM("deepseek-chat", "你是一个有帮助的助手")
	if err != nil {
		log.Fatalf("create deepseek client: %v", err)
	}

	// 基本生成
	resp, err := llm.Generate(ctx, "你好，请介绍一下自己")
	if err != nil {
		log.Fatalf("generate: %v", err)
	}

	log.Printf("回复：%s", resp)
}
```

---

### 示例 2：带图片的生成

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
	ctx := context.Background()

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		log.Fatalf("create deepseek client: %v", err)
	}

	// 读取图片文件
	imageData, err := os.ReadFile("chart.png")
	if err != nil {
		log.Fatalf("read image: %v", err)
	}

	files := []models.File{
		{
			Name: "chart.png",
			MIME: "image/png",
			Data: imageData,
		},
	}

	// 带图片生成
	resp, err := llm.GenerateWithFiles(ctx, "请分析这张图表的内容", files)
	if err != nil {
		log.Fatalf("generate with files: %v", err)
	}

	log.Printf("分析结果：%s", resp)
}
```

---

### 示例 3：流式生成

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
	ctx := context.Background()

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		log.Fatalf("create deepseek client: %v", err)
	}

	// 流式生成
	stream, err := llm.GenerateStream(ctx, "请写一首关于春天的诗")
	if err != nil {
		log.Fatalf("create stream: %v", err)
	}

	fmt.Println("正在生成...")
	for chunk := range stream {
		if chunk.Err != nil {
			log.Fatalf("stream error: %v", chunk.Err)
		}
		if chunk.Delta != "" {
			fmt.Print(chunk.Delta)
		}
		if chunk.Done {
			fmt.Println("\n生成完成")
		}
	}
}
```

---

### 示例 4：带图片的流式生成（完整功能）

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
	ctx := context.Background()

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		log.Fatalf("create deepseek client: %v", err)
	}

	// 读取图片
	imageData, err := os.ReadFile("screenshot.png")
	if err != nil {
		log.Fatalf("read image: %v", err)
	}

	files := []models.File{
		{
			Name: "screenshot.png",
			MIME: "image/png",
			Data: imageData,
		},
	}

	// 带图片的流式生成
	stream, err := llm.GenerateStreamWithFiles(ctx, "请描述这张截图中的内容", files)
	if err != nil {
		log.Fatalf("create stream with files: %v", err)
	}

	fmt.Println("正在分析图片...")
	for chunk := range stream {
		if chunk.Err != nil {
			log.Fatalf("stream error: %v", chunk.Err)
		}
		if chunk.Delta != "" {
			fmt.Print(chunk.Delta)
		}
		if chunk.Done {
			fmt.Printf("\n分析完成，总长度：%d 字符\n", len(chunk.FullText))
		}
	}
}
```

---

### 示例 5：在 ADK 中使用 DeepSeek

```go
package main

import (
	"context"
	"flag"
	"log"

	"github.com/Protocol-Lattice/go-agent/src/adk"
	adkmodules "github.com/Protocol-Lattice/go-agent/src/adk/modules"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	"github.com/Protocol-Lattice/go-agent/src/models"
)

func main() {
	qdrantURL := flag.String("qdrant-url", "http://localhost:6333", "Qdrant base URL")
	qdrantCollection := flag.String("qdrant-collection", "adk_memories", "Qdrant collection name")
	flag.Parse()
	ctx := context.Background()

	// 创建 DeepSeek 模型
	deepseekModel, err := models.NewDeepSeekLLM("deepseek-chat", "Swarm orchestration:")
	if err != nil {
		log.Fatalf("create deepseek model: %v", err)
	}

	// 创建 ADK
	memOpts := engine.DefaultOptions()
	adkAgent, err := adk.New(ctx,
		adk.WithDefaultSystemPrompt("你是一个使用 DeepSeek 的 AI 助手"),
		adk.WithModules(
			adkmodules.NewModelModule("deepseek-model", func(_ context.Context) (models.Agent, error) {
				return deepseekModel, nil
			}),
			adkmodules.InQdrantMemory(100000, *qdrantURL, *qdrantCollection, memory.AutoEmbedder(), &memOpts),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	agent, err := adkAgent.BuildAgent(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 使用 agent
	resp, err := agent.Generate(ctx, "SessionID", "你好")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
```

---

## 📊 支持的模型

| 模型名称 | 用途 | 上下文窗口 | 价格（约） |
|---------|------|-----------|-----------|
| `deepseek-chat` | 通用对话 | 32K | ¥0.002/1K tokens |
| `deepseek-coder` | 代码生成 | 16K | ¥0.002/1K tokens |
| `deepseek-v2` | V2 版本 | 32K | ¥0.002/1K tokens |
| `deepseek-v2.5` | V2.5 增强版 | 32K | ¥0.002/1K tokens |

**使用方式**：
```go
// 对话模型
models.NewDeepSeekLLM("deepseek-chat", "")

// 代码专用模型
models.NewDeepSeekLLM("deepseek-coder", "你是一个专业的程序员助手")

// 最新版本
models.NewDeepSeekLLM("deepseek-v2.5", "")
```

---

## 🔗 与其他 LLM 实现的对比

| 功能 | OpenAI | Anthropic | DeepSeek | Gemini |
|------|--------|-----------|----------|--------|
| 基本生成 | ✅ | ✅ | ✅ | ✅ |
| 带文件生成 | ✅ | ✅ | ✅ | ✅ |
| 流式生成 | ✅ | ✅ | ✅ | ✅ |
| 带文件流式 | ❌ | ❌ | ✅ | ❌ |
| 图片支持 | ✅ | ✅ | ✅ | ✅ |
| 视频支持 | ✅ | ❌ | ❌ | ✅ |
| API 兼容性 | 原生 | 独立 | OpenAI 兼容 | 独立 |
| 中文支持 | ✅ | ✅ | ✅✅ | ✅ |
| 价格 | $ | $ | ¥ | ¥ |

**DeepSeek 优势**：
- ✅ 价格更低（人民币计价）
- ✅ 中文优化更好
- ✅ OpenAI API 兼容（代码复用）
- ✅ 支持带文件的流式生成

---

## 🛠️ 实现要点

### 1. 客户端初始化

```go
// 关键：使用 OpenAI 兼容配置
config := openai.DefaultConfig(apiKey)
config.BaseURL = "https://api.deepseek.com"  // 修改端点
client := openai.NewClientWithConfig(config)
```

### 2. 多模态内容构建

```go
// DeepSeek 使用与 OpenAI 相同的 MultiContent 格式
contentParts := []openai.ChatMessagePart{
    {Type: "text", Text: prompt},
    {
        Type: "image_url",
        ImageURL: &openai.ChatMessageImageURL{
            URL:    "data:image/png;base64,...",
            Detail: openai.ImageURLDetailAuto,
        },
    },
}
```

### 3. 流式处理

```go
// 使用 CreateChatCompletionStream
stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
    Model:    model,
    Messages: messages,
    Stream:   true,  // 启用流式
})

// 迭代接收
for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break
    }
    delta := resp.Choices[0].Delta.Content
    // 处理增量
}
```

### 4. 带文件的流式

```go
// 结合 MultiContent 和 Stream
stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
    Model: model,
    Messages: []openai.ChatCompletionMessage{{
        Role:         "user",
        MultiContent: contentParts,  // 包含文本 + 图片
    }},
    Stream: true,
})
```

---

## 🧪 测试建议

### 单元测试

```go
package models_test

import (
	"context"
	"os"
	"testing"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

func TestDeepSeek_Generate(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	resp, err := llm.Generate(context.Background(), "你好")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if resp.(string) == "" {
		t.Error("empty response")
	}
}

func TestDeepSeek_GenerateWithFiles(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	// 创建测试图片
	files := []models.File{
		{
			Name: "test.png",
			MIME: "image/png",
			Data: []byte("fake image data"),
		},
	}

	_, err = llm.GenerateWithFiles(context.Background(), "这是什么？", files)
	if err != nil {
		t.Fatalf("generate with files: %v", err)
	}
}

func TestDeepSeek_GenerateStream(t *testing.T) {
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	llm, err := models.NewDeepSeekLLM("deepseek-chat", "")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	stream, err := llm.GenerateStream(context.Background(), "数到 10")
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}

	var fullText string
	for chunk := range stream {
		if chunk.Err != nil {
			t.Fatalf("stream error: %v", chunk.Err)
		}
		fullText = chunk.FullText
	}

	if fullText == "" {
		t.Error("empty stream")
	}
}
```

---

## 📚 参考资料

- [DeepSeek 开放平台](https://platform.deepseek.com/)
- [DeepSeek API 文档](https://platform.deepseek.com/api-docs/)
- [go-openai 库](https://github.com/sashabaranov/go-openai)
- [OpenAI API 参考](https://platform.openai.com/docs/api-reference)

---

## ✅ 实现检查清单

在实现完成后，请确认：

- [ ] 创建 `src/models/deepseek.go` 文件
- [ ] 实现 `NewDeepSeekLLM()` 构造函数
- [ ] 实现 `Generate()` 方法
- [ ] 实现 `GenerateWithFiles()` 方法
- [ ] 实现 `GenerateStream()` 方法
- [ ] 实现 `GenerateStreamWithFiles()` 方法（新增功能）
- [ ] 添加 `var _ Agent = (*DeepSeekLLM)(nil)` 接口检查
- [ ] 编写单元测试
- [ ] 更新 README.md 添加 DeepSeek 示例
- [ ] 测试环境变量配置
- [ ] 测试不同模型（chat/coder/v2.5）
- [ ] 测试图片上传功能
- [ ] 测试流式输出

---

## 🔄 后续扩展

### 可能的优化方向

1. **批量处理**：支持多图片批量上传
2. **缓存机制**：复用 `models.CachedLLM` 模式
3. **重试逻辑**：添加自动重试和退避策略
4. **速率限制**：实现请求限流
5. **日志记录**：添加详细的请求/响应日志
6. **错误分类**：区分 API 错误、网络错误、认证错误

### 接口扩展建议

考虑在 `src/models/interface.go` 中添加第 4 个方法：

```go
type Agent interface {
    Generate(context.Context, string) (any, error)
    GenerateWithFiles(context.Context, string, []File) (any, error)
    GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
    
    // 新增：带文件的流式生成
    GenerateStreamWithFiles(ctx context.Context, prompt string, files []File) (<-chan StreamChunk, error)
}
```

这将使接口更加完整（2×2 矩阵：有/无文件 × 同步/流式）。

---

**文档结束**

---

*本文档为设计方案，实现时请参考最新 DeepSeek API 文档。*
