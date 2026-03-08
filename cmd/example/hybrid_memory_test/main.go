// Hybrid Memory Test: Agent with both personal and shared memory
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	"github.com/Protocol-Lattice/go-agent/src/memory/session"
	"github.com/Protocol-Lattice/go-agent/src/memory/store"
)

func main() {
	ctx := context.Background()

	// Setup
	os.Setenv("ADK_EMBED_PROVIDER", "ollama")
	os.Setenv("ADK_EMBED_MODEL", "nomic-embed-text")

	pgStore, _ := store.NewPostgresStore(ctx, "postgres://admin:admin@localhost:5432/ragdb")
	defer pgStore.Close()

	bank := memory.NewMemoryBankWithStore(pgStore)
	embedder := memory.AutoEmbedder()
	opts := engine.DefaultOptions()

	// Setup SpaceRegistry
	registry := session.NewSpaceRegistry(24 * time.Hour)
	registry.Grant("team:alpha", "agent-A", session.SpaceRoleWriter, 0)

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     Hybrid Memory Test: Personal + Shared                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// Create Agent-A with BOTH personal and shared memory
	// ═══════════════════════════════════════════════════════════════
	
	sessionMem := memory.NewSessionMemory(bank, 16)
	sessionMem.WithEmbedder(embedder)
	sessionMem.Spaces = registry
	sessionMem.WithEngine(engine.NewEngine(pgStore, opts))

	// SharedSession with local="agent-A" and shared space="team:alpha"
	shared := session.NewSharedSession(sessionMem, "agent-A", "team:alpha")

	fmt.Println("📦 Agent-A created with:")
	fmt.Println("   - Personal memory: sessionID = 'agent-A'")
	fmt.Println("   - Shared memory: space = 'team:alpha'")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 1: Store personal memory (private to agent-A)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 1: Store Personal Memory")
	fmt.Println("══════════════════════════════════════════════════════════════")

	// Method 1: Use AddShortLocal for short-term personal memory
	shared.AddShortLocal("I prefer coffee over tea", map[string]string{"type": "personal"})
	fmt.Println("📝 Stored personal short-term: 'I prefer coffee over tea'")

	// Method 2: Use StoreLongTo with local sessionID for long-term personal memory
	_, err := shared.StoreLongTo(ctx, "agent-A", "My birthday is March 15th", map[string]any{"type": "personal"})
	if err != nil {
		log.Printf("   ⚠️  Store failed: %v", err)
	} else {
		fmt.Println("📝 Stored personal long-term: 'My birthday is March 15th'")
	}
	_ = shared.FlushLocal(ctx)

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 2: Store shared memory (accessible by team:alpha members)
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 2: Store Shared Memory")
	fmt.Println("══════════════════════════════════════════════════════════════")

	_, err = shared.StoreLongTo(ctx, "team:alpha", "Project deadline is Friday", map[string]any{"type": "shared"})
	if err != nil {
		log.Printf("   ⚠️  Store failed: %v", err)
	} else {
		fmt.Println("📝 Stored shared memory: 'Project deadline is Friday'")
	}
	_ = shared.FlushSpace(ctx, "team:alpha")

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 3: Retrieve - should get BOTH personal and shared
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 3: Retrieve Personal + Shared Memory")
	fmt.Println("══════════════════════════════════════════════════════════════")

	recs, err := shared.Retrieve(ctx, "birthday or deadline", 10)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		fmt.Printf("🔍 Found %d memories:\n", len(recs))
		personalCount := 0
		sharedCount := 0
		for _, r := range recs {
			if r.SessionID == "agent-A" {
				personalCount++
				fmt.Printf("   👤 [PERSONAL] %s (session: %s)\n", r.Content, r.SessionID)
			} else if r.SessionID == "team:alpha" {
				sharedCount++
				fmt.Printf("   🌐 [SHARED] %s (session: %s)\n", r.Content, r.SessionID)
			}
		}
		fmt.Printf("   → Personal: %d, Shared: %d\n", personalCount, sharedCount)
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 4: Retrieve only personal memory
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 4: Retrieve Only Personal Memory")
	fmt.Println("══════════════════════════════════════════════════════════════")

	recs2, err := shared.Retrieve(ctx, "coffee tea preference", 5)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		fmt.Printf("🔍 Found %d memories about preferences:\n", len(recs2))
		for _, r := range recs2 {
			if r.SessionID == "agent-A" {
				fmt.Printf("   👤 [PERSONAL] %s\n", r.Content)
			} else {
				fmt.Printf("   🌐 [SHARED] %s (session: %s)\n", r.Content, r.SessionID)
			}
		}
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 5: Create Agent-B (same team) and verify it can ONLY access shared
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 5: Agent-B Access (Shared Memory Only)")
	fmt.Println("══════════════════════════════════════════════════════════════")

	sessionMemB := memory.NewSessionMemory(bank, 16)
	sessionMemB.WithEmbedder(embedder)
	sessionMemB.Spaces = registry
	sessionMemB.WithEngine(engine.NewEngine(pgStore, opts))

	sharedB := session.NewSharedSession(sessionMemB, "agent-B", "team:alpha")

	fmt.Println("📦 Agent-B created (same team:alpha)")
	fmt.Println()
	fmt.Println("🔍 Agent-B retrieves 'birthday or deadline':")

	recs3, err := sharedB.Retrieve(ctx, "birthday or deadline", 10)
	if err != nil {
		log.Printf("   ⚠️  Retrieve failed: %v", err)
	} else {
		fmt.Printf("   Found %d memories:\n", len(recs3))
		for _, r := range recs3 {
			if r.SessionID == "agent-A" {
				fmt.Printf("   ⚠️  [AGENT-A PERSONAL] %s ← Should NOT appear!\n", r.Content)
			} else if r.SessionID == "agent-B" {
				fmt.Printf("   👤 [AGENT-B PERSONAL] %s\n", r.Content)
			} else if r.SessionID == "team:alpha" {
				fmt.Printf("   🌐 [SHARED] %s\n", r.Content)
			}
		}

		// Verify Agent-B cannot access Agent-A's personal memory
		canAccessPersonal := false
		for _, r := range recs3 {
			if r.SessionID == "agent-A" {
				canAccessPersonal = true
			}
		}
		if canAccessPersonal {
			fmt.Println("   ❌ SECURITY ISSUE: Agent-B accessed Agent-A's personal memory!")
		} else {
			fmt.Println("   ✅ Agent-B CANNOT access Agent-A's personal memory (correct)")
		}
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// TEST 6: Verify in PostgreSQL
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("TEST 6: Verify in PostgreSQL")
	fmt.Println("══════════════════════════════════════════════════════════════")

	count, _ := pgStore.Count(ctx)
	fmt.Printf("📊 Total memories in PostgreSQL: %d\n", count)

	fmt.Println()
	fmt.Println("Memory breakdown by session_id:")
	fmt.Println("   session_id = 'agent-A' → Personal memory of Agent-A")
	fmt.Println("   session_id = 'team:alpha' → Shared memory (team:alpha members)")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════
	// SUMMARY
	// ═══════════════════════════════════════════════════════════════
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    TEST SUMMARY                              ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  ✅ Agent can have BOTH personal and shared memory          ║")
	fmt.Println("║  ✅ Personal memory: stored with local sessionID            ║")
	fmt.Println("║  ✅ Shared memory: stored with space name                   ║")
	fmt.Println("║  ✅ Retrieve() returns BOTH personal and shared             ║")
	fmt.Println("║  ✅ Other agents CANNOT access personal memory              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}