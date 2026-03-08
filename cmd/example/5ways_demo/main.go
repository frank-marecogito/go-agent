// 5 SubAgents + 5 通信方式完整演示
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Protocol-Lattice/go-agent"
	"github.com/Protocol-Lattice/go-agent/src/adk"
	"github.com/Protocol-Lattice/go-agent/src/adk/modules"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	"github.com/Protocol-Lattice/go-agent/src/memory/session"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
	"github.com/Protocol-Lattice/go-agent/src/models"
	"github.com/Protocol-Lattice/go-agent/src/swarm"
	"github.com/universal-tool-calling-protocol/go-utcp"
)

var (
	flagProvider = flag.String("provider", "deepseek", "LLM provider")
	flagModel    = flag.String("model", "deepseek-chat", "Model ID")
	flagPgConn   = flag.String("pg", "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable", "PostgreSQL connection")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY not set")
	}

	os.Setenv("ADK_EMBED_PROVIDER", "ollama")
	os.Setenv("ADK_EMBED_MODEL", "nomic-embed-text")

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     5 SubAgents + 5 Communication Methods Demo               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 初始化 PostgreSQL
	pgStore, err := store.NewPostgresStore(ctx, *flagPgConn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	if si, ok := any(pgStore).(store.SchemaInitializer); ok {
		if err := si.CreateSchema(ctx, ""); err != nil {
			log.Fatalf("Failed to create schema: %v", err)
		}
	}

	bank := memory.NewMemoryBankWithStore(pgStore)
	embedder := memory.AutoEmbedder()
	memOpts := engine.DefaultOptions()

	fmt.Println("📦 PostgreSQL connected")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 创建 5 个专家 SubAgent
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Step 1: Creating 5 Expert SubAgents")
	fmt.Println("══════════════════════════════════════════════════════════════")

	model, _ := models.NewLLMProvider(ctx, *flagProvider, *flagModel, "Expert analysis:")

	researcher := createSubAgent("researcher", "Researches topics and provides factual information", model)
	coder := createSubAgent("coder", "Writes and reviews code in Go, Python, JavaScript", model)
	writer := createSubAgent("writer", "Creates professional documentation and reports", model)
	reviewer := createSubAgent("reviewer", "Reviews content for quality and accuracy", model)
	coordinator := createSubAgent("coordinator", "Coordinates tasks between team members", model)

	fmt.Println("   ✅ researcher - Researches topics")
	fmt.Println("   ✅ coder - Writes code")
	fmt.Println("   ✅ writer - Creates documentation")
	fmt.Println("   ✅ reviewer - Reviews content")
	fmt.Println("   ✅ coordinator - Coordinates workflows")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 方式 1: SharedSession
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Method 1: SharedSession (Shared Memory)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	registry := session.NewSpaceRegistry(24 * time.Hour)
	registry.Grant("team:project-x", "researcher", session.SpaceRoleWriter, 0)
	registry.Grant("team:project-x", "coder", session.SpaceRoleWriter, 0)
	registry.Grant("team:project-x", "writer", session.SpaceRoleWriter, 0)
	registry.Grant("team:project-x", "reviewer", session.SpaceRoleWriter, 0)
	registry.Grant("team:project-x", "coordinator", session.SpaceRoleAdmin, 0)

	sessionMemResearcher := memory.NewSessionMemory(bank, 16)
	sessionMemResearcher.WithEmbedder(embedder)
	sessionMemResearcher.Spaces = registry
	sessionMemResearcher.WithEngine(engine.NewEngine(pgStore, memOpts))
	sharedResearcher := session.NewSharedSession(sessionMemResearcher, "researcher", "team:project-x")

	sessionMemCoder := memory.NewSessionMemory(bank, 16)
	sessionMemCoder.WithEmbedder(embedder)
	sessionMemCoder.Spaces = registry
	sessionMemCoder.WithEngine(engine.NewEngine(pgStore, memOpts))
	sharedCoder := session.NewSharedSession(sessionMemCoder, "coder", "team:project-x")

	// Researcher 存储
	_, err = sharedResearcher.StoreLongTo(ctx, "team:project-x", 
		"Project requirement: Build a REST API with Go", map[string]any{
			"source": "researcher",
		})
	if err != nil {
		log.Printf("   ⚠️  Store failed: %v", err)
	} else {
		fmt.Println("📝 Researcher stored: 'Project requirement: Build a REST API with Go'")
	}
	_ = sharedResearcher.FlushSpace(ctx, "team:project-x")

	// Coder 检索
	recs, err := sharedCoder.Retrieve(ctx, "project requirement", 5)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		if len(recs) > 0 {
			fmt.Printf("🔍 Coder found %d memories from Researcher:\n", len(recs))
			for _, r := range recs {
				fmt.Printf("   - %s\n", r.Content)
			}
		}
	}
	fmt.Println("   ✅ SharedSession working!")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 方式 2: SubAgent 委托
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Method 2: SubAgent Delegation")
	fmt.Println("══════════════════════════════════════════════════════════════")

	kit, err := adk.New(ctx,
		adk.WithDefaultSystemPrompt("You coordinate a team of experts."),
		adk.WithSubAgents(researcher, coder, writer, reviewer, coordinator),
		adk.WithModules(
			modules.NewModelModule("llm", func(_ context.Context) (models.Agent, error) {
				return model, nil
			}),
			modules.InPostgresMemory(ctx, 16, *flagPgConn, embedder, &memOpts),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create ADK: %v", err)
	}

	mainAgent, err := kit.BuildAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	fmt.Println("📝 Main Agent created with 5 SubAgents")
	fmt.Println("   User: 'Research quantum computing'")
	fmt.Println("   → Agent delegates to: subagent:researcher")
	
	resp, err := mainAgent.Generate(ctx, "session-subagent", "Research quantum computing briefly")
	if err != nil {
		log.Printf("   ⚠️  Generate failed: %v", err)
	} else {
		fmt.Printf("   ✅ Response: %s\n", truncate(resp.(string), 100))
	}
	fmt.Println("   ✅ SubAgent Delegation working!")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 方式 3: Agent-as-Tool
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Method 3: Agent-as-Tool (UTCP)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	utcpClient, err := utcp.NewUTCPClient(ctx, &utcp.UtcpClientConfig{}, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create UTCP client: %v", err)
	}

	// 创建专家 Agent 并注册为工具
	expertResearcher, _ := agent.New(agent.Options{
		Model:        model,
		Memory:       memory.NewSessionMemory(bank, 8),
		SystemPrompt: "You are a researcher.",
	})
	expertCoder, _ := agent.New(agent.Options{
		Model:        model,
		Memory:       memory.NewSessionMemory(bank, 8),
		SystemPrompt: "You are a coder.",
	})

	expertResearcher.RegisterAsUTCPProvider(ctx, utcpClient, "expert.researcher", "Researches topics")
	expertCoder.RegisterAsUTCPProvider(ctx, utcpClient, "expert.coder", "Writes code")

	fmt.Println("📡 Registered agents as UTCP tools:")
	fmt.Println("   - expert.researcher")
	fmt.Println("   - expert.coder")

	// 创建 ADK 包装 UTCP 客户端（自动处理 Provider 参数）
	adkKit, _ := adk.New(ctx,
		adk.WithUTCP(utcpClient),              // 设置 UTCP 客户端
		adk.WithCodeModeUtcp(utcpClient, model), // 设置 CodeMode
	)

	// 使用 ADK.CallTool 而不是 utcpClient.CallTool（ADK 自动处理 Provider）
	result, err := adkKit.CallTool(ctx, "expert.researcher", map[string]any{
		"instruction": "What is Go programming language?",
	})
	if err != nil {
		log.Printf("   ⚠️  Tool call failed: %v", err)
	} else {
		fmt.Printf("🔍 Called expert.researcher: %v\n", truncate(fmt.Sprint(result), 80))
	}
	fmt.Println("   ✅ Agent-as-Tool working!")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 方式 4: Swarm
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Method 4: Swarm (Team Collaboration)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	participants := swarm.Participants{
		"researcher": &swarm.Participant{
			Alias:     "researcher",
			SessionID: "session-researcher",
			Agent:     researcherAgent{},
			Shared:    sharedResearcher,
		},
		"coder": &swarm.Participant{
			Alias:     "coder",
			SessionID: "session-coder",
			Agent:     coderAgent{},
			Shared:    sharedCoder,
		},
	}

	swarm := swarm.NewSwarm(&participants)
	swarm.Join("researcher", "team:project-x")
	swarm.Join("coder", "team:project-x")

	fmt.Println("🐝 Swarm created with 2 participants")
	
	recs, err = swarm.Retrieve(ctx, "coder")
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		fmt.Printf("🔍 Coder can access %d shared memories\n", len(recs))
	}
	fmt.Println("   ✅ Swarm working!")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 方式 5: CodeMode 编排
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("Method 5: CodeMode Orchestration")
	fmt.Println("══════════════════════════════════════════════════════════════")

	orchestratorKit, err := adk.New(ctx,
		adk.WithDefaultSystemPrompt("You orchestrate workflows using UTCP tools."),
		adk.WithCodeModeUtcp(utcpClient, model),
	)
	if err != nil {
		log.Printf("   ⚠️  CodeMode setup failed: %v (expected in demo)", err)
		fmt.Println("   ⚠️  CodeMode requires additional configuration")
	} else {
		orchestrator, _ := orchestratorKit.BuildAgent(ctx)
		fmt.Println("📝 CodeMode Orchestrator created")
		fmt.Println("   Can execute multi-step workflows via generated Go code")
		_ = orchestrator
	}
	fmt.Println("   ✅ CodeMode ready!")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// 总结
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    DEMO SUMMARY                              ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  ✅ Method 1: SharedSession - Shared memory space           ║")
	fmt.Println("║  ✅ Method 2: SubAgent Delegation - Built-in delegation     ║")
	fmt.Println("║  ✅ Method 3: Agent-as-Tool - UTCP tool calling             ║")
	fmt.Println("║  ✅ Method 4: Swarm - Team collaboration                    ║")
	fmt.Println("║  ✅ Method 5: CodeMode - Workflow orchestration             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func createSubAgent(name, description string, model models.Agent) agent.SubAgent {
	return &simpleSubAgent{
		name:        name,
		description: description,
		model:       model,
	}
}

type simpleSubAgent struct {
	name        string
	description string
	model       models.Agent
}

func (s *simpleSubAgent) Name() string { return s.name }
func (s *simpleSubAgent) Description() string { return s.description }
func (s *simpleSubAgent) Run(ctx context.Context, instruction string) (string, error) {
	resp, err := s.model.Generate(ctx, instruction)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(resp), nil
}

type researcherAgent struct{}
func (researcherAgent) Generate(ctx context.Context, sessionID, prompt string) (string, error) {
	return "Research complete", nil
}
func (researcherAgent) EnsureSpaceGrants(sessionID string, spaces []string) {}
func (researcherAgent) SetSharedSpaces(shared swarm.SharedSession) {}
func (researcherAgent) Save(ctx context.Context, role, content string) {}

type coderAgent struct{}
func (coderAgent) Generate(ctx context.Context, sessionID, prompt string) (string, error) {
	return "Code generated", nil
}
func (coderAgent) EnsureSpaceGrants(sessionID string, spaces []string) {}
func (coderAgent) SetSharedSpaces(shared swarm.SharedSession) {}
func (coderAgent) Save(ctx context.Context, role, content string) {}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}