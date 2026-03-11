// Package hello 提供智能问候语生成功能
package hello

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Protocol-Lattice/go-agent/src/models"
)

// GreetingRequest 问候语生成请求
type GreetingRequest struct {
	RecipientName string     `json:"recipient_name"` // 收件人姓名
	SceneID       string     `json:"scene_id"`       // 场景 ID
	Language      string     `json:"language"`       // 语言（目前支持中文）
	Tone          ToneType   `json:"tone"`           // 语气类型
	Length        LengthType `json:"length"`         // 长度类型
	CustomContext string     `json:"custom_context"` // 自定义上下文（可选）
}

// GreetingResponse 问候语生成响应
type GreetingResponse struct {
	ID              string    `json:"id"`               // 问候语 ID
	Content         string    `json:"content"`          // 生成的问候语内容
	Scene           string    `json:"scene"`            // 场景名称
	Tone            ToneType  `json:"tone"`             // 使用的语气
	Length          LengthType `json:"length"`          // 使用的长度
	GenerationTimeMs int64    `json:"generation_time_ms"` // 生成耗时（毫秒）
	CreatedAt       time.Time `json:"created_at"`       // 创建时间
}

// GreetingEngine 问候语生成引擎
type GreetingEngine struct {
	models map[string]models.Agent // LLM 模型映射
	rand   *rand.Rand              // 随机数生成器
}

// GreetingEngineOption 引擎配置选项
type GreetingEngineOption func(*GreetingEngine)

// WithModel 设置 LLM 模型
func WithModel(name string, model models.Agent) GreetingEngineOption {
	return func(e *GreetingEngine) {
		if e.models == nil {
			e.models = make(map[string]models.Agent)
		}
		e.models[name] = model
	}
}

// NewGreetingEngine 创建新的问候语生成引擎
func NewGreetingEngine(opts ...GreetingEngineOption) *GreetingEngine {
	engine := &GreetingEngine{
		models: make(map[string]models.Agent),
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// Generate 生成问候语（基于模板）
// 这是快速路径，不依赖 LLM，响应时间 < 100ms
func (e *GreetingEngine) Generate(ctx context.Context, req GreetingRequest) (*GreetingResponse, error) {
	startTime := time.Now()

	// 获取场景模板
	scene, err := GetSceneByID(req.SceneID)
	if err != nil {
		return nil, fmt.Errorf("获取场景模板失败：%w", err)
	}

	// 查找匹配的模板变体
	variant := e.findBestVariant(scene.Templates, req.Tone, req.Length)
	if variant == nil {
		// 如果没有完全匹配的，使用第一个变体
		if len(scene.Templates) > 0 {
			variant = &scene.Templates[0]
		} else {
			return nil, fmt.Errorf("场景模板为空：%s", req.SceneID)
		}
	}

	// 从模板中选择一个模式
	pattern := e.selectPattern(variant.Patterns)

	// 替换变量生成问候语
	content := e.renderPattern(pattern, req)

	generationTime := time.Since(startTime).Milliseconds()

	return &GreetingResponse{
		ID:              fmt.Sprintf("greeting_%d", time.Now().UnixNano()),
		Content:         content,
		Scene:           scene.Name,
		Tone:            req.Tone,
		Length:          req.Length,
		GenerationTimeMs: generationTime,
		CreatedAt:       time.Now(),
	}, nil
}

// GenerateWithLLM 使用 LLM 生成个性化问候语
// 响应时间取决于 LLM，通常 < 3 秒
func (e *GreetingEngine) GenerateWithLLM(ctx context.Context, req GreetingRequest, modelName string) (*GreetingResponse, error) {
	startTime := time.Now()

	// 获取场景模板
	scene, err := GetSceneByID(req.SceneID)
	if err != nil {
		return nil, fmt.Errorf("获取场景模板失败：%w", err)
	}

	// 获取 LLM 模型
	model, ok := e.models[modelName]
	if !ok {
		// 如果没有指定模型，回退到模板生成
		return e.Generate(ctx, req)
	}

	// 构建提示词
	prompt := e.buildLLMPrompt(scene, req)

	// 调用 LLM 生成
	// 注意：这里使用简化的调用方式，实际项目中可能需要更复杂的处理
	resp, err := model.Generate(ctx, prompt)
	if err != nil {
		// LLM 调用失败，回退到模板生成
		return e.Generate(ctx, req)
	}

	// 类型断言
	content, ok := resp.(string)
	if !ok {
		content = fmt.Sprintf("%v", resp)
	}

	generationTime := time.Since(startTime).Milliseconds()

	return &GreetingResponse{
		ID:              fmt.Sprintf("greeting_%d", time.Now().UnixNano()),
		Content:         content,
		Scene:           scene.Name,
		Tone:            req.Tone,
		Length:          req.Length,
		GenerationTimeMs: generationTime,
		CreatedAt:       time.Now(),
	}, nil
}

// findBestVariant 查找最佳匹配的模板变体
func (e *GreetingEngine) findBestVariant(variants []TemplateVariant, tone ToneType, length LengthType) *TemplateVariant {
	// 优先查找完全匹配的
	for i := range variants {
		if variants[i].Tone == tone && variants[i].Length == length {
			return &variants[i]
		}
	}

	// 其次查找长度匹配的
	for i := range variants {
		if variants[i].Length == length {
			return &variants[i]
		}
	}

	// 最后查找语气匹配的
	for i := range variants {
		if variants[i].Tone == tone {
			return &variants[i]
		}
	}

	// 返回第一个
	if len(variants) > 0 {
		return &variants[0]
	}

	return nil
}

// selectPattern 从模式中选择一个
func (e *GreetingEngine) selectPattern(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}
	if len(patterns) == 1 {
		return patterns[0]
	}
	return patterns[e.rand.Intn(len(patterns))]
}

// renderPattern 渲染模板模式
func (e *GreetingEngine) renderPattern(pattern string, req GreetingRequest) string {
	result := pattern

	// 替换收件人名称
	if req.RecipientName != "" {
		result = strings.ReplaceAll(result, "{recipient}", req.RecipientName)
	} else {
		result = strings.ReplaceAll(result, "{recipient}", "您")
	}

	// 替换其他可能的变量
	result = strings.ReplaceAll(result, "{custom_context}", req.CustomContext)

	return result
}

// buildLLMPrompt 构建 LLM 提示词
func (e *GreetingEngine) buildLLMPrompt(scene *SceneTemplate, req GreetingRequest) string {
	var sb strings.Builder

	sb.WriteString("你是一位专业的问候语助手，请根据以下信息生成一段得体的问候语：\n\n")
	sb.WriteString(fmt.Sprintf("场景：%s (%s)\n", scene.Name, scene.Description))
	sb.WriteString(fmt.Sprintf("收件人：%s\n", req.RecipientName))
	sb.WriteString(fmt.Sprintf("语气：%s\n", req.Tone))
	sb.WriteString(fmt.Sprintf("长度：%s\n", req.Length))

	if req.CustomContext != "" {
		sb.WriteString(fmt.Sprintf("额外信息：%s\n", req.CustomContext))
	}

	sb.WriteString("\n请生成一段真诚、得体的问候语，直接输出问候语内容即可：\n")

	return sb.String()
}

// GetAvailableScenes 获取所有可用场景
func (e *GreetingEngine) GetAvailableScenes() []SceneTemplate {
	return GetPresetScenes()
}

// GetAvailableTones 获取所有可用语气
func (e *GreetingEngine) GetAvailableTones() []ToneType {
	return []ToneType{
		ToneProfessional,
		ToneFriendly,
		ToneEnthusiastic,
		ToneWarm,
		ToneHumorous,
	}
}

// GetAvailableLengths 获取所有可用长度
func (e *GreetingEngine) GetAvailableLengths() []LengthType {
	return []LengthType{
		LengthShort,
		LengthMedium,
		LengthLong,
	}
}

// ValidateRequest 验证请求参数
func (e *GreetingEngine) ValidateRequest(req GreetingRequest) error {
	if req.RecipientName == "" {
		return fmt.Errorf("收件人姓名不能为空")
	}

	if req.SceneID == "" {
		return fmt.Errorf("场景 ID 不能为空")
	}

	// 验证场景是否存在
	_, err := GetSceneByID(req.SceneID)
	if err != nil {
		return fmt.Errorf("无效的场景 ID：%s", req.SceneID)
	}

	// 验证语气
	validTones := e.GetAvailableTones()
	toneValid := false
	for _, tone := range validTones {
		if req.Tone == tone {
			toneValid = true
			break
		}
	}
	if !toneValid && req.Tone != "" {
		return fmt.Errorf("无效的语气类型：%s", req.Tone)
	}

	// 验证长度
	validLengths := e.GetAvailableLengths()
	lengthValid := false
	for _, length := range validLengths {
		if req.Length == length {
			lengthValid = true
			break
		}
	}
	if !lengthValid && req.Length != "" {
		return fmt.Errorf("无效的长度类型：%s", req.Length)
	}

	return nil
}
