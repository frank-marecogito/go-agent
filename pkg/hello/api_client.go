// Package hello 提供智能问候语生成功能
package hello

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient API 客户端配置
type APIClient struct {
	BaseURL    string        // API 基础 URL
	HTTPClient *http.Client  // HTTP 客户端
	APIKey     string        // API 密钥（可选）
	Timeout    time.Duration // 请求超时
}

// APIClientOption 客户端配置选项
type APIClientOption func(*APIClient)

// WithBaseURL 设置 API 基础 URL
func WithBaseURL(url string) APIClientOption {
	return func(c *APIClient) {
		c.BaseURL = url
	}
}

// WithAPIKey 设置 API 密钥
func WithAPIKey(key string) APIClientOption {
	return func(c *APIClient) {
		c.APIKey = key
	}
}

// WithTimeout 设置请求超时
func WithTimeout(timeout time.Duration) APIClientOption {
	return func(c *APIClient) {
		c.Timeout = timeout
	}
}

// WithHTTPClient 设置 HTTP 客户端
func WithHTTPClient(client *http.Client) APIClientOption {
	return func(c *APIClient) {
		c.HTTPClient = client
	}
}

// NewAPIClient 创建新的 API 客户端
func NewAPIClient(opts ...APIClientOption) *APIClient {
	client := &APIClient{
		BaseURL: "http://localhost:8080", // 默认本地地址
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.Timeout > 0 {
		client.HTTPClient.Timeout = client.Timeout
	}

	return client
}

// APIResponse API 响应结构
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// GreetingGenerateRequest API 问候语生成请求
type GreetingGenerateRequest struct {
	RecipientName string     `json:"recipient_name"`
	SceneID       string     `json:"scene_id"`
	Language      string     `json:"language,omitempty"`
	Tone          ToneType   `json:"tone,omitempty"`
	Length        LengthType `json:"length,omitempty"`
	CustomContext string     `json:"custom_context,omitempty"`
}

// GreetingGenerateResponse API 问候语生成响应
type GreetingGenerateResponse struct {
	ID               string    `json:"id"`
	Content          string    `json:"content"`
	Scene            string    `json:"scene"`
	Tone             ToneType  `json:"tone,omitempty"`
	Length           LengthType `json:"length,omitempty"`
	CreationTimeMs   int64     `json:"creation_time_ms"`
	CreatedAt        time.Time `json:"created_at"`
}

// SceneInfo 场景信息
type SceneInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        SceneType `json:"type"`
	Description string    `json:"description"`
}

// SceneListResponse 场景列表响应
type SceneListResponse struct {
	PresetScenes  []SceneInfo `json:"preset_scenes"`
	CustomScenes  []SceneInfo `json:"custom_scenes"`
}

// GenerateViaAPI 通过 API 生成问候语（带重试机制）
func (c *APIClient) GenerateViaAPI(ctx context.Context, req GreetingGenerateRequest) (*GreetingGenerateResponse, error) {
	return c.GenerateViaAPIWithRetry(ctx, req, 3, 100*time.Millisecond)
}

// GenerateViaAPIWithRetry 通过 API 生成问候语（带重试机制）
func (c *APIClient) GenerateViaAPIWithRetry(ctx context.Context, req GreetingGenerateRequest, maxRetries int, retryDelay time.Duration) (*GreetingGenerateResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.doGenerateRequest(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 检查是否为可重试错误
		if !c.isRetryableError(err) {
			return nil, err
		}

		// 等待后重试（指数退避）
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)):
				// 继续重试
			}
		}
	}

	return nil, fmt.Errorf("请求失败，已重试 %d 次：%w", maxRetries, lastErr)
}

// doGenerateRequest 执行生成请求
func (c *APIClient) doGenerateRequest(ctx context.Context, req GreetingGenerateRequest) (*GreetingGenerateResponse, error) {
	url := c.BaseURL + "/api/v1/greetings/generate"

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败：%w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API 返回错误：%s", apiResp.Error)
	}

	// 解析数据
	var greeting GreetingGenerateResponse
	if err := json.Unmarshal(apiResp.Data, &greeting); err != nil {
		return nil, fmt.Errorf("解析问候语数据失败：%w", err)
	}

	return &greeting, nil
}

// GetScenesViaAPI 获取场景列表
func (c *APIClient) GetScenesViaAPI(ctx context.Context) (*SceneListResponse, error) {
	url := c.BaseURL + "/api/v1/scenes"

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Accept", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var sceneResp SceneListResponse
	if err := json.NewDecoder(resp.Body).Decode(&sceneResp); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	return &sceneResp, nil
}

// HealthCheck 健康检查
func (c *APIClient) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	url := c.BaseURL + "/api/v1/health"

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("解析响应失败：%w", err)
	}

	return health, nil
}

// isRetryableError 判断是否为可重试错误
func (c *APIClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 网络错误可重试
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"i/o timeout",
		"EOF",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring 查找子串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
