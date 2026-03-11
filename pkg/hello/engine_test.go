package hello

import (
	"context"
	"testing"
)

// ============ 场景模板测试 ============

func TestGetPresetScenes(t *testing.T) {
	scenes := GetPresetScenes()

	// 验证场景数量（至少 5 种）
	if len(scenes) < 5 {
		t.Errorf("期望至少 5 种场景，实际：%d", len(scenes))
	}

	// 验证每种场景都有必要的字段
	for _, scene := range scenes {
		if scene.ID == "" {
			t.Errorf("场景 ID 不能为空：%v", scene)
		}
		if scene.Name == "" {
			t.Errorf("场景名称不能为空：%v", scene)
		}
		if scene.Type == "" {
			t.Errorf("场景类型不能为空：%v", scene)
		}
		if len(scene.Templates) == 0 {
			t.Errorf("场景模板不能为空：%s", scene.ID)
		}
	}
}

func TestGetSceneByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantErr   bool
		verifyFunc func(*testing.T, *SceneTemplate)
	}{
		{
			name:    "存在场景 - 商务会议",
			id:      "business_meeting",
			wantErr: false,
			verifyFunc: func(t *testing.T, scene *SceneTemplate) {
				if scene.Type != SceneBusiness {
					t.Errorf("期望场景类型为 business，实际：%s", scene.Type)
				}
			},
		},
		{
			name:    "存在场景 - 朋友聚会",
			id:      "casual_friend",
			wantErr: false,
			verifyFunc: func(t *testing.T, scene *SceneTemplate) {
				if scene.Type != SceneCasual {
					t.Errorf("期望场景类型为 casual，实际：%s", scene.Type)
				}
			},
		},
		{
			name:    "存在场景 - 生日庆祝",
			id:      "celebration_birthday",
			wantErr: false,
			verifyFunc: func(t *testing.T, scene *SceneTemplate) {
				if scene.Type != SceneCelebration {
					t.Errorf("期望场景类型为 celebration，实际：%s", scene.Type)
				}
			},
		},
		{
			name:    "存在场景 - 慰问关怀",
			id:      "condolence_sympathy",
			wantErr: false,
			verifyFunc: func(t *testing.T, scene *SceneTemplate) {
				if scene.Type != SceneCondolence {
					t.Errorf("期望场景类型为 condolence，实际：%s", scene.Type)
				}
			},
		},
		{
			name:    "存在场景 - 春节祝福",
			id:      "festival_spring",
			wantErr: false,
			verifyFunc: func(t *testing.T, scene *SceneTemplate) {
				if scene.Type != SceneFestival {
					t.Errorf("期望场景类型为 festival，实际：%s", scene.Type)
				}
			},
		},
		{
			name:    "不存在场景",
			id:      "nonexistent_scene",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scene, err := GetSceneByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSceneByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.verifyFunc != nil {
				tt.verifyFunc(t, scene)
			}
		})
	}
}

func TestGetScenesByType(t *testing.T) {
	tests := []struct {
		name     string
		sceneType SceneType
		minCount int
	}{
		{name: "商务场景", sceneType: SceneBusiness, minCount: 1},
		{name: "非正式场景", sceneType: SceneCasual, minCount: 1},
		{name: "正式场景", sceneType: SceneFormal, minCount: 1},
		{name: "庆祝场景", sceneType: SceneCelebration, minCount: 1},
		{name: "慰问场景", sceneType: SceneCondolence, minCount: 1},
		{name: "日常场景", sceneType: SceneDaily, minCount: 1},
		{name: "节日场景", sceneType: SceneFestival, minCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenes := GetScenesByType(tt.sceneType)
			if len(scenes) < tt.minCount {
				t.Errorf("期望至少 %d 个场景，实际：%d", tt.minCount, len(scenes))
			}

			// 验证返回的场景类型都匹配
			for _, scene := range scenes {
				if scene.Type != tt.sceneType {
					t.Errorf("场景类型不匹配：期望 %s，实际 %s", tt.sceneType, scene.Type)
				}
			}
		})
	}
}

func TestGetTimeBasedGreeting(t *testing.T) {
	recipient := "小明"
	greeting := GetTimeBasedGreeting(recipient)

	// 验证问候语包含收件人名称
	if greeting == "" {
		t.Error("问候语不能为空")
	}

	// 验证问候语包含收件人
	if !containsString(greeting, recipient) {
		t.Errorf("问候语应包含收件人名称：%s", recipient)
	}
}

// ============ 问候引擎测试 ============

func TestNewGreetingEngine(t *testing.T) {
	engine := NewGreetingEngine()

	if engine == nil {
		t.Error("引擎不应为 nil")
	}
	if engine.models == nil {
		t.Error("models map 应初始化")
	}
	if engine.rand == nil {
		t.Error("rand 应初始化")
	}
}

func TestGreetingEngine_Generate(t *testing.T) {
	ctx := context.Background()
	engine := NewGreetingEngine()

	tests := []struct {
		name      string
		req       GreetingRequest
		wantErr   bool
		verifyFunc func(*testing.T, *GreetingResponse)
	}{
		{
			name: "成功生成 - 商务场景",
			req: GreetingRequest{
				RecipientName: "张总",
				SceneID:       "business_meeting",
				Tone:          ToneProfessional,
				Length:        LengthMedium,
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
				if resp.Scene != "商务会议" {
					t.Errorf("期望场景为商务会议，实际：%s", resp.Scene)
				}
				if resp.GenerationTimeMs < 0 {
					t.Error("生成时间不能为负数")
				}
				// 验证响应时间 < 3 秒
				if resp.GenerationTimeMs > 3000 {
					t.Errorf("生成时间超过 3 秒：%dms", resp.GenerationTimeMs)
				}
			},
		},
		{
			name: "成功生成 - 生日场景",
			req: GreetingRequest{
				RecipientName: "小红",
				SceneID:       "celebration_birthday",
				Tone:          ToneEnthusiastic,
				Length:        LengthShort,
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
				if resp.Scene != "生日庆祝" {
					t.Errorf("期望场景为生日庆祝，实际：%s", resp.Scene)
				}
			},
		},
		{
			name: "成功生成 - 日常场景",
			req: GreetingRequest{
				RecipientName: "小明",
				SceneID:       "daily_greeting",
				Tone:          ToneFriendly,
				Length:        LengthMedium,
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
			},
		},
		{
			name: "成功生成 - 节日场景",
			req: GreetingRequest{
				RecipientName: "李叔叔",
				SceneID:       "festival_spring",
				Tone:          ToneEnthusiastic,
				Length:        LengthLong,
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
				if resp.Scene != "春节祝福" {
					t.Errorf("期望场景为春节祝福，实际：%s", resp.Scene)
				}
			},
		},
		{
			name: "成功生成 - 慰问场景",
			req: GreetingRequest{
				RecipientName: "王阿姨",
				SceneID:       "condolence_sympathy",
				Tone:          ToneWarm,
				Length:        LengthMedium,
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
				if resp.Scene != "慰问关怀" {
					t.Errorf("期望场景为慰问关怀，实际：%s", resp.Scene)
				}
			},
		},
		{
			name: "失败 - 场景不存在",
			req: GreetingRequest{
				RecipientName: "测试",
				SceneID:       "nonexistent",
			},
			wantErr: true,
		},
		{
			name: "成功生成 - 默认语气和长度",
			req: GreetingRequest{
				RecipientName: "小赵",
				SceneID:       "casual_friend",
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *GreetingResponse) {
				if resp.Content == "" {
					t.Error("问候语内容不能为空")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := engine.Generate(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.verifyFunc != nil {
				tt.verifyFunc(t, resp)
			}
		})
	}
}

func TestGreetingEngine_ValidateRequest(t *testing.T) {
	engine := NewGreetingEngine()

	tests := []struct {
		name    string
		req     GreetingRequest
		wantErr bool
	}{
		{
			name: "有效请求",
			req: GreetingRequest{
				RecipientName: "小明",
				SceneID:       "daily_greeting",
				Tone:          ToneFriendly,
				Length:        LengthMedium,
			},
			wantErr: false,
		},
		{
			name: "无效 - 收件人姓名为空",
			req: GreetingRequest{
				SceneID: "daily_greeting",
			},
			wantErr: true,
		},
		{
			name: "无效 - 场景 ID 为空",
			req: GreetingRequest{
				RecipientName: "小明",
			},
			wantErr: true,
		},
		{
			name: "无效 - 场景不存在",
			req: GreetingRequest{
				RecipientName: "小明",
				SceneID:       "nonexistent",
			},
			wantErr: true,
		},
		{
			name: "有效 - 空语气（使用默认）",
			req: GreetingRequest{
				RecipientName: "小明",
				SceneID:       "daily_greeting",
			},
			wantErr: false,
		},
		{
			name: "有效 - 空长度（使用默认）",
			req: GreetingRequest{
				RecipientName: "小明",
				SceneID:       "daily_greeting",
				Tone:          ToneFriendly,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGreetingEngine_GetAvailableScenes(t *testing.T) {
	engine := NewGreetingEngine()
	scenes := engine.GetAvailableScenes()

	if len(scenes) < 5 {
		t.Errorf("期望至少 5 种场景，实际：%d", len(scenes))
	}
}

func TestGreetingEngine_GetAvailableTones(t *testing.T) {
	engine := NewGreetingEngine()
	tones := engine.GetAvailableTones()

	// 验证至少 3 种语气
	if len(tones) < 3 {
		t.Errorf("期望至少 3 种语气，实际：%d", len(tones))
	}

	// 验证包含预期的语气
	expectedTones := []ToneType{
		ToneProfessional,
		ToneFriendly,
		ToneEnthusiastic,
	}

	for _, expected := range expectedTones {
		found := false
		for _, tone := range tones {
			if tone == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("期望包含语气：%s", expected)
		}
	}
}

func TestGreetingEngine_GetAvailableLengths(t *testing.T) {
	engine := NewGreetingEngine()
	lengths := engine.GetAvailableLengths()

	// 验证至少 3 种长度
	if len(lengths) < 3 {
		t.Errorf("期望至少 3 种长度，实际：%d", len(lengths))
	}

	expectedLengths := []LengthType{
		LengthShort,
		LengthMedium,
		LengthLong,
	}

	for _, expected := range expectedLengths {
		found := false
		for _, length := range lengths {
			if length == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("期望包含长度：%s", expected)
		}
	}
}

// ============ 性能测试 ============

func BenchmarkGreetingEngine_Generate(b *testing.B) {
	ctx := context.Background()
	engine := NewGreetingEngine()
	req := GreetingRequest{
		RecipientName: "小明",
		SceneID:       "daily_greeting",
		Tone:          ToneFriendly,
		Length:        LengthMedium,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGreetingEngine_Generate_Business(b *testing.B) {
	ctx := context.Background()
	engine := NewGreetingEngine()
	req := GreetingRequest{
		RecipientName: "张总",
		SceneID:       "business_meeting",
		Tone:          ToneProfessional,
		Length:        LengthLong,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Generate(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============ 辅助函数 ============

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstringInTest(s, substr)
}

func findSubstringInTest(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
