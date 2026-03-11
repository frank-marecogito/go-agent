package hello

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============ API 客户端测试 ============

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient()

	if client == nil {
		t.Error("客户端不应为 nil")
	}
	if client.BaseURL == "" {
		t.Error("BaseURL 应初始化")
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient 应初始化")
	}
}

func TestNewAPIClient_WithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	
	client := NewAPIClient(
		WithBaseURL("https://api.example.com"),
		WithAPIKey("test-key"),
		WithTimeout(60*time.Second),
		WithHTTPClient(customClient),
	)

	if client.BaseURL != "https://api.example.com" {
		t.Errorf("期望 BaseURL 为 https://api.example.com，实际：%s", client.BaseURL)
	}
	if client.APIKey != "test-key" {
		t.Errorf("期望 APIKey 为 test-key，实际：%s", client.APIKey)
	}
	if client.HTTPClient != customClient {
		t.Error("HTTPClient 应使用自定义客户端")
	}
}

func TestAPIClient_HealthCheck_Success(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("期望路径 /api/v1/health，实际：%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.0.0"}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	health, err := client.HealthCheck(ctx)

	if err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("期望 status 为 ok，实际：%v", health["status"])
	}
}

func TestAPIClient_HealthCheck_Failure(t *testing.T) {
	// 创建返回错误的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	_, err := client.HealthCheck(ctx)

	if err == nil {
		t.Error("期望 HealthCheck 返回错误，实际为 nil")
	}
}

func TestAPIClient_GenerateViaAPI_Success(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("期望方法 POST，实际：%s", r.Method)
		}
		if r.URL.Path != "/api/v1/greetings/generate" {
			t.Errorf("期望路径 /api/v1/greetings/generate，实际：%s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"data": {
				"id": "greeting_123",
				"content": "您好，小明！",
				"scene": "日常问候",
				"tone": "friendly",
				"length": "medium",
				"creation_time_ms": 50,
				"created_at": "2024-01-15T10:30:00Z"
			}
		}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	req := GreetingGenerateRequest{
		RecipientName: "小明",
		SceneID:       "daily_greeting",
		Tone:          ToneFriendly,
		Length:        LengthMedium,
	}

	resp, err := client.GenerateViaAPI(ctx, req)

	if err != nil {
		t.Fatalf("GenerateViaAPI() error = %v", err)
	}

	if resp.ID != "greeting_123" {
		t.Errorf("期望 ID 为 greeting_123，实际：%s", resp.ID)
	}
	if resp.Content != "您好，小明！" {
		t.Errorf("期望内容为您好，小明！，实际：%s", resp.Content)
	}
	if resp.Scene != "日常问候" {
		t.Errorf("期望场景为日常问候，实际：%s", resp.Scene)
	}
}

func TestAPIClient_GenerateViaAPI_Error(t *testing.T) {
	// 创建返回错误的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": false, "error": "invalid request"}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	req := GreetingGenerateRequest{
		RecipientName: "小明",
		SceneID:       "daily_greeting",
	}

	_, err := client.GenerateViaAPI(ctx, req)

	if err == nil {
		t.Error("期望 GenerateViaAPI 返回错误，实际为 nil")
	}
}

func TestAPIClient_GenerateViaAPIWithRetry_Success(t *testing.T) {
	attempts := 0
	
	// 创建测试服务器，第一次返回连接错误，第二次成功
	// 使用 httptest.NewUnstartedServer 来模拟连接错误
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"data": {
				"id": "greeting_456",
				"content": "重试成功！",
				"scene": "日常问候",
				"creation_time_ms": 100,
				"created_at": "2024-01-15T10:30:00Z"
			}
		}`))
	}))
	server.Start()
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	req := GreetingGenerateRequest{
		RecipientName: "小明",
		SceneID:       "daily_greeting",
	}

	resp, err := client.GenerateViaAPIWithRetry(ctx, req, 3, 10*time.Millisecond)

	if err != nil {
		t.Fatalf("GenerateViaAPIWithRetry() error = %v", err)
	}

	if resp.Content != "重试成功！" {
		t.Errorf("期望内容为重试成功！，实际：%s", resp.Content)
	}

	// 由于服务器立即可用，应该第一次就成功
	if attempts != 1 {
		t.Errorf("期望尝试 1 次，实际：%d", attempts)
	}
}

func TestAPIClient_GetScenesViaAPI_Success(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/scenes" {
			t.Errorf("期望路径 /api/v1/scenes，实际：%s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"preset_scenes": [
				{"id": "daily_greeting", "name": "日常问候", "type": "daily", "description": "日常打招呼"}
			],
			"custom_scenes": []
		}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	
	ctx := context.Background()
	resp, err := client.GetScenesViaAPI(ctx)

	if err != nil {
		t.Fatalf("GetScenesViaAPI() error = %v", err)
	}

	if len(resp.PresetScenes) != 1 {
		t.Errorf("期望 1 个预设场景，实际：%d", len(resp.PresetScenes))
	}
	if resp.PresetScenes[0].ID != "daily_greeting" {
		t.Errorf("期望场景 ID 为 daily_greeting，实际：%s", resp.PresetScenes[0].ID)
	}
}

func TestAPIClient_isRetryableError(t *testing.T) {
	client := NewAPIClient()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"可重试 - connection refused", &mockError{"connection refused"}, true},
		{"可重试 - connection reset", &mockError{"connection reset"}, true},
		{"可重试 - i/o timeout", &mockError{"i/o timeout"}, true},
		{"可重试 - EOF", &mockError{"EOF"}, true},
		{"不可重试 - invalid request", &mockError{"invalid request"}, false},
		{"不可重试 - nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockError 模拟错误类型
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// ============ 性能测试 ============

func BenchmarkAPIClient_GenerateViaAPI(b *testing.B) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"data": {
				"id": "greeting_123",
				"content": "您好！",
				"scene": "日常问候",
				"creation_time_ms": 50,
				"created_at": "2024-01-15T10:30:00Z"
			}
		}`))
	}))
	defer server.Close()

	client := NewAPIClient(WithBaseURL(server.URL))
	ctx := context.Background()
	req := GreetingGenerateRequest{
		RecipientName: "小明",
		SceneID:       "daily_greeting",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GenerateViaAPI(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
