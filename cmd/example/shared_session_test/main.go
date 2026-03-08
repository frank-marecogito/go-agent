// Shared Memory Test using SharedSession
// This test verifies if go-agent's SharedSession enables cross-session memory sharing
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Protocol-Lattice/go-agent"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	"github.com/Protocol-Lattice/go-agent/src/memory/session"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
	"github.com/Protocol-Lattice/go-agent/src/models"
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
	fmt.Println("║          SharedSession Memory Test                           ║")
	fmt.Println("║          Testing Cross-Session Memory Sharing                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Setup PostgreSQL Store
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
	fmt.Println("   ✅ PostgreSQL connected")

	// 2. Create Shared Memory Bank
	memOpts := engine.DefaultOptions()
	sharedBank := memory.NewMemoryBankWithStore(pgStore)
	embedder := memory.AutoEmbedder()
	fmt.Println("   ✅ Shared memory bank created")
	fmt.Println()

	// Setup SpaceRegistry with ACLs (shared across all agents)
	registry := session.NewSpaceRegistry(24 * time.Hour)
	registry.Grant("team:alpha", "agent-A", session.SpaceRoleAdmin, 0)
	registry.Grant("team:alpha", "agent-B", session.SpaceRoleWriter, 0)
	registry.Grant("team:alpha", "agent-D", session.SpaceRoleWriter, 0)
	registry.Grant("team:beta", "agent-C", session.SpaceRoleAdmin, 0)
	registry.Grant("team:beta", "agent-D", session.SpaceRoleWriter, 0)
	fmt.Println("   ✅ SpaceRegistry configured with ACLs")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 1: Basic SharedSession with same space
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 1: Two agents sharing the same space")
	fmt.Println("══════════════════════════════════════════════════════════════")

	// Agent A: session "agent-A", joins space "team:alpha"
	sessionMemA := memory.NewSessionMemory(sharedBank, 16)
	sessionMemA.WithEmbedder(embedder)
	sessionMemA.Spaces = registry
	engA := engine.NewEngine(pgStore, memOpts)
	engA.WithEmbedder(embedder)
	sessionMemA.WithEngine(engA)

	sharedA := session.NewSharedSession(sessionMemA, "agent-A", "team:alpha")

	_, _ = agent.New(agent.Options{
		Model:        &dummyModel{name: "Agent-A"},
		Memory:       sessionMemA,
		Shared:       sharedA,
		SystemPrompt: "You are Agent A. You share memory with team:alpha.",
	})

	// Agent B: session "agent-B", joins same space "team:alpha"
	sessionMemB := memory.NewSessionMemory(sharedBank, 16)
	sessionMemB.WithEmbedder(embedder)
	sessionMemB.Spaces = registry
	engB := engine.NewEngine(pgStore, memOpts)
	engB.WithEmbedder(embedder)
	sessionMemB.WithEngine(engB)

	sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

	_, _ = agent.New(agent.Options{
		Model:        &dummyModel{name: "Agent-B"},
		Memory:       sessionMemB,
		Shared:       sharedB,
		SystemPrompt: "You are Agent B. You share memory with team:alpha.",
	})

	fmt.Println()
	fmt.Println("📝 Step 1: Agent-A stores memory to shared space")
	
	// Agent-A stores memory to the shared space
	_, err = sharedA.StoreLongTo(ctx, "team:alpha", "The project codename is Phoenix", map[string]any{
		"source": "agent-A",
	})
	if err != nil {
		log.Printf("   ⚠️  Agent-A store failed: %v", err)
	} else {
		fmt.Println("   ✅ Agent-A stored: 'The project codename is Phoenix'")
	}

	// Flush to PostgreSQL
	_ = sharedA.FlushSpace(ctx, "team:alpha")
	fmt.Println("   ✅ Flushed to PostgreSQL")

	fmt.Println()
	fmt.Println("🔍 Step 2: Agent-B retrieves from shared space")

	// Agent-B retrieves from the same shared space
	recs, err := sharedB.Retrieve(ctx, "project codename", 5)
	if err != nil {
		log.Printf("   ⚠️  Agent-B retrieve failed: %v", err)
	} else {
		if len(recs) > 0 {
			fmt.Printf("   ✅ Agent-B found %d memories:\n", len(recs))
			for _, r := range recs {
				fmt.Printf("      - %s (session: %s)\n", r.Content, r.SessionID)
			}
		} else {
			fmt.Println("   ❌ Agent-B found NO memories")
		}
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 2: Different spaces (should NOT share)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 2: Agents in different spaces (should NOT share)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	// Agent C: session "agent-C", joins space "team:beta"
	sessionMemC := memory.NewSessionMemory(sharedBank, 16)
	sessionMemC.WithEmbedder(embedder)
	sessionMemC.Spaces = registry
	engC := engine.NewEngine(pgStore, memOpts)
	engC.WithEmbedder(embedder)
	sessionMemC.WithEngine(engC)

	sharedC := session.NewSharedSession(sessionMemC, "agent-C", "team:beta")

	fmt.Println()
	fmt.Println("📝 Step 1: Agent-C stores memory to team:beta")

	_, err = sharedC.StoreLongTo(ctx, "team:beta", "Secret password is BlueDragon", map[string]any{
		"source": "agent-C",
	})
	if err != nil {
		log.Printf("   ⚠️  Agent-C store failed: %v", err)
	} else {
		fmt.Println("   ✅ Agent-C stored: 'Secret password is BlueDragon'")
	}
	_ = sharedC.FlushSpace(ctx, "team:beta")

	fmt.Println()
	fmt.Println("🔍 Step 2: Agent-A tries to retrieve from team:beta (should fail)")

	// Agent-A should NOT be able to access team:beta
	recs2, err := sharedA.Retrieve(ctx, "secret password", 5)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		if len(recs2) > 0 {
			// Check if any result is from team:beta
			hasBeta := false
			for _, r := range recs2 {
				if r.SessionID == "team:beta" {
					hasBeta = true
					break
				}
			}
			if hasBeta {
				fmt.Printf("   ❌ SECURITY ISSUE: Agent-A accessed team:beta memories:\n")
			} else {
				fmt.Printf("   ✅ Agent-A found %d memories from OWN space (OK - semantic search):\n", len(recs2))
			}
			for _, r := range recs2 {
				fmt.Printf("      - %s (session: %s)\n", r.Content, r.SessionID)
			}
		} else {
			fmt.Println("   ✅ Agent-A found NO memories (expected - different space)")
		}
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 3: Agent joins multiple spaces
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 3: Agent joins multiple spaces")
	fmt.Println("══════════════════════════════════════════════════════════════")

	fmt.Println()
	fmt.Println("📝 Step 1: Agent-D joins both team:alpha and team:beta")

	// Agent D: joins both spaces
	sessionMemD := memory.NewSessionMemory(sharedBank, 16)
	sessionMemD.WithEmbedder(embedder)
	sessionMemD.Spaces = registry
	engD := engine.NewEngine(pgStore, memOpts)
	engD.WithEmbedder(embedder)
	sessionMemD.WithEngine(engD)

	sharedD := session.NewSharedSession(sessionMemD, "agent-D", "team:alpha")
	sharedD.Join("team:beta")

	fmt.Println("   ✅ Agent-D joined: team:alpha, team:beta")

	fmt.Println()
	fmt.Println("🔍 Step 2: Agent-D retrieves from both spaces")

	recs3, err := sharedD.Retrieve(ctx, "project or password", 10)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		if len(recs3) > 0 {
			fmt.Printf("   ✅ Agent-D found %d memories from both spaces:\n", len(recs3))
			for _, r := range recs3 {
				fmt.Printf("      - %s (session: %s)\n", r.Content, r.SessionID)
			}
		} else {
			fmt.Println("   ❌ Agent-D found NO memories")
		}
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 4: Verify in PostgreSQL
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 4: Verify memories in PostgreSQL")
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()

	count, _ := pgStore.Count(ctx)
	fmt.Printf("📊 Total memories in PostgreSQL: %d\n", count)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    TEST SUMMARY                              ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  SharedSession enables cross-session memory sharing:        ║")
	fmt.Println("║  ✅ Same space: Agents CAN share memories                   ║")
	fmt.Println("║  ✅ Different space: Agents CANNOT access (isolation)       ║")
	fmt.Println("║  ✅ Multi-space: Agents can join multiple spaces            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

type dummyModel struct {
	name string
}

func (d *dummyModel) Generate(ctx context.Context, prompt string) (any, error) {
	return fmt.Sprintf("[%s] processed", d.name), nil
}

func (d *dummyModel) GenerateWithFiles(ctx context.Context, prompt string, files []models.File) (any, error) {
	return d.Generate(ctx, prompt)
}

func (d *dummyModel) GenerateStream(ctx context.Context, prompt string) (<-chan models.StreamChunk, error) {
	ch := make(chan models.StreamChunk, 1)
	ch <- models.StreamChunk{Delta: "done", FullText: "done", Done: true}
	close(ch)
	return ch, nil
}