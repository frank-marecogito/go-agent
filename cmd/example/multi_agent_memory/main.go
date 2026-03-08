// Multi-Agent Shared Memory Test (Simplified)
// Demonstrates multiple agents sharing memories through PostgreSQL + pgvector
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Protocol-Lattice/go-agent"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
	"github.com/Protocol-Lattice/go-agent/src/models"
)

var (
	flagProvider    = flag.String("provider", "deepseek", "LLM provider")
	flagModel       = flag.String("model", "deepseek-chat", "Model ID")
	flagPgConn      = flag.String("pg", "postgres://admin:admin@localhost:5432/ragdb?sslmode=disable", "PostgreSQL connection")
	flagSession     = flag.String("session", "shared-memory-test", "Shared session ID")
	flagTestMessage = flag.String("message", "My name is Alice and I love Go programming", "Test message to store")
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
	fmt.Println("║       Multi-Agent Shared Memory Test                         ║")
	fmt.Println("║       PostgreSQL + pgvector + Ollama Embedding               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Session: %-50s |\n", *flagSession)
	fmt.Printf("║  Provider: %-48s |\n", *flagProvider)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Setup PostgreSQL Store (shared across all agents)
	fmt.Println("📦 Setting up PostgreSQL store...")
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
	fmt.Println("   ✅ PostgreSQL connected and schema ready")

	// 2. Create Shared Memory Bank
	fmt.Println("\n🧠 Creating shared memory bank...")
	memOpts := engine.DefaultOptions()
	sharedBank := memory.NewMemoryBankWithStore(pgStore)
	embedder := memory.AutoEmbedder()
	fmt.Println("   ✅ Shared memory bank created")

	// 3. Create specialized agents with shared memory
	fmt.Println("\n🤖 Creating specialized agents with shared memory...")

	// Agent 1: MemoryAgent - stores facts
	memoryAgent := NewMemoryAgent(sharedBank, embedder, &memOpts)
	fmt.Println("   ✅ MemoryAgent created (stores facts)")

	// Agent 2: ChatAgent - conversational with memory access
	chatAgent := NewChatAgent(*flagProvider, *flagModel, sharedBank, embedder, &memOpts)
	fmt.Println("   ✅ ChatAgent created (conversational)")

	// Agent 3: ResearchAgent - analyzes memories
	researchAgent := NewResearchAgent(sharedBank, embedder, &memOpts)
	fmt.Println("   ✅ ResearchAgent created (analyzes patterns)")

	// 4. Run tests
	fmt.Println("\n══════════════════════════════════════════════════════════════")
	fmt.Println("                     TEST SEQUENCE                              ")
	fmt.Println("══════════════════════════════════════════════════════════════")

	// Test 1: MemoryAgent stores a fact
	fmt.Printf("\n📝 Test 1: MemoryAgent storing fact\n")
	fmt.Printf("   Input: \"%s\"\n", *flagTestMessage)

	result1, err := memoryAgent.Store(ctx, *flagSession, *flagTestMessage)
	if err != nil {
		log.Printf("   ⚠️ Store failed: %v", err)
	} else {
		fmt.Printf("   ✅ %s\n", result1)
	}

	// Test 2: ChatAgent retrieves the fact
	fmt.Printf("\n💬 Test 2: ChatAgent retrieving memory\n")
	fmt.Printf("   Query: \"What is my name?\"\n")

	result2, err := chatAgent.Chat(ctx, *flagSession, "What is my name?")
	if err != nil {
		log.Printf("   ⚠️ Chat failed: %v", err)
	} else {
		fmt.Printf("   ✅ ChatAgent: %s\n", truncate(result2, 200))
	}

	// Test 3: ResearchAgent analyzes all memories
	fmt.Printf("\n🔍 Test 3: ResearchAgent analyzing memories\n")

	result3, err := researchAgent.Analyze(ctx, *flagSession)
	if err != nil {
		log.Printf("   ⚠️ Analysis failed: %v", err)
	} else {
		fmt.Printf("   ✅ Analysis: %s\n", result3)
	}

	// Test 4: Verify in PostgreSQL directly
	fmt.Printf("\n💾 Test 4: Verifying shared memory in PostgreSQL\n")

	emb, _ := embedder.Embed(ctx, *flagTestMessage)
	memories, err := pgStore.SearchMemory(ctx, *flagSession, emb, 10)
	if err != nil {
		log.Printf("   ⚠️ Search failed: %v", err)
	} else {
		fmt.Printf("   ✅ Found %d memories in PostgreSQL:\n", len(memories))
		for i, m := range memories {
			fmt.Printf("      %d. [importance=%.2f] %s\n", i+1, m.Importance, truncate(m.Content, 60))
		}
	}

	// Test 5: Another agent accesses same memory
	fmt.Printf("\n🔄 Test 5: Cross-agent memory sharing\n")
	fmt.Printf("   ChatAgent asks: \"What does Alice love?\"\n")

	result5, err := chatAgent.Chat(ctx, *flagSession, "What does Alice love?")
	if err != nil {
		log.Printf("   ⚠️ Cross-agent failed: %v", err)
	} else {
		fmt.Printf("   ✅ ChatAgent: %s\n", truncate(result5, 200))
	}

	// Test 6: Store another fact and verify sharing
	fmt.Printf("\n📝 Test 6: Store another fact via ResearchAgent\n")

	result6, err := researchAgent.Store(ctx, *flagSession, "Alice also enjoys playing chess on weekends")
	if err != nil {
		log.Printf("   ⚠️ Store failed: %v", err)
	} else {
		fmt.Printf("   ✅ %s\n", result6)
	}

	// Verify both facts are accessible
	fmt.Printf("\n💬 Test 7: ChatAgent retrieves all facts\n")

	result7, err := chatAgent.Chat(ctx, *flagSession, "Tell me everything you know about Alice")
	if err != nil {
		log.Printf("   ⚠️ Chat failed: %v", err)
	} else {
		fmt.Printf("   ✅ ChatAgent: %s\n", truncate(result7, 300))
	}

	// Flush all
	fmt.Printf("\n💾 Flushing memories...\n")
	_ = memoryAgent.Flush(ctx, *flagSession)
	_ = chatAgent.Flush(ctx, *flagSession)
	_ = researchAgent.Flush(ctx, *flagSession)
	fmt.Println("   ✅ All memories flushed to PostgreSQL")

	// Final count
	fmt.Printf("\n📊 Final memory count in PostgreSQL:\n")
	count, _ := pgStore.Count(ctx)
	fmt.Printf("   Total memories: %d\n", count)

	fmt.Println("\n══════════════════════════════════════════════════════════════")
	fmt.Println("                     TEST COMPLETE                              ")
	fmt.Println("══════════════════════════════════════════════════════════════")
}

// MemoryAgent stores and retrieves facts
type MemoryAgent struct {
	agent    *agent.Agent
	store    store.VectorStore
	embedder memory.Embedder
	engine   *engine.Engine
}

func NewMemoryAgent(bank *memory.MemoryBank, embedder memory.Embedder, opts *engine.Options) *MemoryAgent {
	sessionMem := memory.NewSessionMemory(bank, 16)
	sessionMem.WithEmbedder(embedder)
	eng := engine.NewEngine(bank.Store, *opts)
	eng.WithEmbedder(embedder)
	sessionMem.WithEngine(eng)

	ag, _ := agent.New(agent.Options{
		Model:        &DummyModel{name: "MemoryAgent"},
		Memory:       sessionMem,
		SystemPrompt: "You are a memory management agent.",
	})

	return &MemoryAgent{
		agent:    ag,
		store:    bank.Store,
		embedder: embedder,
		engine:   eng,
	}
}

func (m *MemoryAgent) Store(ctx context.Context, sessionID, content string) (string, error) {
	meta := map[string]any{
		"source": "memory_agent",
		"space":  sessionID,
	}

	if _, err := m.engine.Store(ctx, sessionID, content, meta); err != nil {
		return "", err
	}
	return fmt.Sprintf("Memory stored: %s", truncate(content, 50)), nil
}

func (m *MemoryAgent) Retrieve(ctx context.Context, sessionID, query string) (string, error) {
	records, err := m.engine.Retrieve(ctx, sessionID, query, 5)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "No memories found", nil
	}

	var results []string
	for _, r := range records {
		results = append(results, r.Content)
	}
	return strings.Join(results, "; "), nil
}

func (m *MemoryAgent) Flush(ctx context.Context, sessionID string) error {
	return m.agent.Flush(ctx, sessionID)
}

// ChatAgent is a conversational agent with shared memory
type ChatAgent struct {
	agent *agent.Agent
}

func NewChatAgent(provider, model string, bank *memory.MemoryBank, embedder memory.Embedder, opts *engine.Options) *ChatAgent {
	sessionMem := memory.NewSessionMemory(bank, 16)
	sessionMem.WithEmbedder(embedder)
	eng := engine.NewEngine(bank.Store, *opts)
	eng.WithEmbedder(embedder)
	sessionMem.WithEngine(eng)

	llm, _ := models.NewLLMProvider(context.Background(), provider, model, "You have access to conversation memory.")

	ag, _ := agent.New(agent.Options{
		Model:        llm,
		Memory:       sessionMem,
		SystemPrompt: "You are a helpful assistant. Use your memory to remember facts about the user.",
	})

	return &ChatAgent{agent: ag}
}

func (c *ChatAgent) Chat(ctx context.Context, sessionID, message string) (string, error) {
	resp, err := c.agent.Generate(ctx, sessionID, message)
	if err != nil {
		return "", err
	}

	// Flush after each interaction
	_ = c.agent.Flush(ctx, sessionID)

	return fmt.Sprint(resp), nil
}

func (c *ChatAgent) Flush(ctx context.Context, sessionID string) error {
	return c.agent.Flush(ctx, sessionID)
}

// ResearchAgent analyzes memory patterns
type ResearchAgent struct {
	agent  *agent.Agent
	store  store.VectorStore
	engine *engine.Engine
}

func NewResearchAgent(bank *memory.MemoryBank, embedder memory.Embedder, opts *engine.Options) *ResearchAgent {
	sessionMem := memory.NewSessionMemory(bank, 16)
	sessionMem.WithEmbedder(embedder)
	eng := engine.NewEngine(bank.Store, *opts)
	eng.WithEmbedder(embedder)
	sessionMem.WithEngine(eng)

	ag, _ := agent.New(agent.Options{
		Model:        &DummyModel{name: "ResearchAgent"},
		Memory:       sessionMem,
		SystemPrompt: "You analyze memory patterns.",
	})

	return &ResearchAgent{
		agent:  ag,
		store:  bank.Store,
		engine: eng,
	}
}

func (r *ResearchAgent) Analyze(ctx context.Context, sessionID string) (string, error) {
	// Get all memories
	count, err := r.store.Count(ctx)
	if err != nil {
		return "", err
	}

	if count == 0 {
		return "No memories to analyze", nil
	}

	return fmt.Sprintf("Found %d total memories in shared store", count), nil
}

func (r *ResearchAgent) Store(ctx context.Context, sessionID, content string) (string, error) {
	meta := map[string]any{
		"source": "research_agent",
		"space":  sessionID,
	}

	if _, err := r.engine.Store(ctx, sessionID, content, meta); err != nil {
		return "", err
	}
	return fmt.Sprintf("Research stored: %s", truncate(content, 50)), nil
}

func (r *ResearchAgent) Flush(ctx context.Context, sessionID string) error {
	return r.agent.Flush(ctx, sessionID)
}

// DummyModel for internal agents
type DummyModel struct {
	name string
}

func (d *DummyModel) Generate(ctx context.Context, prompt string) (any, error) {
	return fmt.Sprintf("[%s] Processed", d.name), nil
}

func (d *DummyModel) GenerateWithFiles(ctx context.Context, prompt string, files []models.File) (any, error) {
	return d.Generate(ctx, prompt)
}

func (d *DummyModel) GenerateStream(ctx context.Context, prompt string) (<-chan models.StreamChunk, error) {
	ch := make(chan models.StreamChunk, 1)
	ch <- models.StreamChunk{Delta: "done", FullText: "done", Done: true}
	close(ch)
	return ch, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}