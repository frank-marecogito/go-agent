// hello-agent 问候代理主程序
// 提供智能问候语生成功能，支持多种场景和语气
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Protocol-Lattice/go-agent/pkg/hello"
	"github.com/Protocol-Lattice/go-agent/src/models"
)

var (
	flagRecipient = flag.String("recipient", "", "收件人姓名（必填）")
	flagScene     = flag.String("scene", "daily_greeting", "场景 ID（可选，默认：daily_greeting）")
	flagTone      = flag.String("tone", "friendly", "语气类型（可选，默认：friendly）")
	flagLength    = flag.String("length", "medium", "长度类型（可选，默认：medium）")
	flagContext   = flag.String("context", "", "自定义上下文（可选）")
	flagModel     = flag.String("model", "", "LLM 模型名称（可选，为空则使用模板生成）")
	flagJSON      = flag.Bool("json", false, "以 JSON 格式输出")
	flagList      = flag.Bool("list", false, "列出所有可用场景")
	flagTimeout   = flag.Duration("timeout", 10*time.Second, "请求超时时间")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()

	// 列出场景模式
	if *flagList {
		listScenes()
		return
	}

	// 验证必填参数
	if *flagRecipient == "" {
		fail("收件人姓名不能为空，请使用 -recipient 参数指定")
	}

	// 创建问候语生成引擎
	engine := hello.NewGreetingEngine()

	// 如果指定了模型，尝试初始化 LLM
	if *flagModel != "" {
		model, err := initModel(ctx, *flagModel)
		if err != nil {
			log.Printf("警告：初始化 LLM 模型失败：%v，将使用模板生成", err)
		} else {
			engine = hello.NewGreetingEngine(
				hello.WithModel(*flagModel, model),
			)
		}
	}

	// 构建请求
	req := hello.GreetingRequest{
		RecipientName: *flagRecipient,
		SceneID:       *flagScene,
		Language:      "zh-CN",
		Tone:          hello.ToneType(*flagTone),
		Length:        hello.LengthType(*flagLength),
		CustomContext: *flagContext,
	}

	// 验证请求
	if err := engine.ValidateRequest(req); err != nil {
		fail(fmt.Sprintf("请求验证失败：%v", err))
	}

	// 生成问候语
	var resp *hello.GreetingResponse
	var err error

	if *flagModel != "" {
		resp, err = engine.GenerateWithLLM(ctx, req, *flagModel)
	} else {
		resp, err = engine.Generate(ctx, req)
	}

	if err != nil {
		fail(fmt.Sprintf("生成问候语失败：%v", err))
	}

	// 输出结果
	if *flagJSON {
		outputJSON(resp)
	} else {
		outputText(resp)
	}
}

// initModel 初始化 LLM 模型
func initModel(ctx context.Context, modelName string) (models.Agent, error) {
	// 根据模型名称选择合适的提供商
	// 支持：gemini, openai, anthropic, ollama, dummy
	switch {
	case modelName == "dummy" || modelName == "test":
		return models.NewDummyLLM("问候语生成助手"), nil
	case modelName == "ollama":
		return models.NewOllamaLLM("llama2", "问候语生成助手")
	case modelName == "gemini" || modelName == "gemini-pro":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("缺少 GOOGLE_API_KEY 或 GEMINI_API_KEY 环境变量")
		}
		return models.NewGeminiLLM(ctx, "gemini-pro", "问候语生成助手")
	case modelName == "openai" || modelName == "gpt-3.5" || modelName == "gpt-4":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("缺少 OPENAI_API_KEY 环境变量")
		}
		_ = apiKey // API key 由 NewOpenAILLM 内部读取环境变量
		return models.NewOpenAILLM(modelName, "问候语生成助手"), nil
	case modelName == "anthropic" || modelName == "claude":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("缺少 ANTHROPIC_API_KEY 环境变量")
		}
		_ = apiKey // API key 由 NewAnthropicLLM 内部读取环境变量
		return models.NewAnthropicLLM("claude-3-sonnet-20240229", "问候语生成助手"), nil
	default:
		return models.NewDummyLLM("问候语生成助手"), nil
	}
}

// listScenes 列出所有可用场景
func listScenes() {
	engine := hello.NewGreetingEngine()
	scenes := engine.GetAvailableScenes()
	tones := engine.GetAvailableTones()
	lengths := engine.GetAvailableLengths()

	fmt.Println("=== 可用场景模板 ===")
	fmt.Println()
	for _, scene := range scenes {
		fmt.Printf("📍 %s (%s)\n", scene.Name, scene.ID)
		fmt.Printf("   类型：%s\n", scene.Type)
		fmt.Printf("   描述：%s\n", scene.Description)
		if len(scene.Keywords) > 0 {
			fmt.Printf("   关键词：%v\n", scene.Keywords)
		}
		fmt.Println()
	}

	fmt.Println("=== 可用语气 ===")
	for _, tone := range tones {
		fmt.Printf("  • %s\n", tone)
	}
	fmt.Println()

	fmt.Println("=== 可用长度 ===")
	for _, length := range lengths {
		fmt.Printf("  • %s\n", length)
	}
	fmt.Println()

	fmt.Println("=== 使用示例 ===")
	fmt.Println("  # 生成日常问候")
	fmt.Println("  go run . -recipient 小明 -scene daily_greeting -tone friendly")
	fmt.Println()
	fmt.Println("  # 生成商务问候")
	fmt.Println("  go run . -recipient 张总 -scene business_meeting -tone professional -length medium")
	fmt.Println()
	fmt.Println("  # 生成生日祝福")
	fmt.Println("  go run . -recipient 小红 -scene celebration_birthday -tone enthusiastic")
	fmt.Println()
	fmt.Println("  # 使用 LLM 生成（需要配置 API Key）")
	fmt.Println("  export GOOGLE_API_KEY=your_key")
	fmt.Println("  go run . -recipient 小明 -scene daily_greeting -model gemini")
	fmt.Println()
}

// outputJSON 以 JSON 格式输出
func outputJSON(resp *hello.GreetingResponse) {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fail(fmt.Sprintf("JSON 编码失败：%v", err))
	}
	fmt.Println(string(data))
}

// outputText 以文本格式输出
func outputText(resp *hello.GreetingResponse) {
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("场景：%s\n", resp.Scene)
	fmt.Printf("语气：%s | 长度：%s\n", resp.Tone, resp.Length)
	fmt.Printf("生成时间：%dms\n", resp.GenerationTimeMs)
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("%s\n", resp.Content)
	fmt.Println("═══════════════════════════════════════")
}

// fail 输出错误信息并退出
func fail(message string) {
	if *flagJSON {
		errResp := map[string]interface{}{
			"error":   message,
			"success": false,
		}
		data, _ := json.Marshal(errResp)
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "错误：%s\n", message)
	}
	os.Exit(1)
}
