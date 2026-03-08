// Parallel Orchestration System for Multi-Service Projects
// Launches concurrent agent sessions that coordinate through shared state
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	flagOutput    = flag.String("output", "", "Path to save orchestration.json")
	flagVerbose   = flag.Bool("v", false, "Verbose output")
	flagTimeout   = flag.Duration("timeout", 60*time.Second, "Total orchestration timeout")
	flagSequential = flag.Bool("sequential", false, "Run sequentially instead of parallel (for comparison)")
)

func main() {
	flag.Parse()

	// Set up context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	// Determine output path
	outputPath := *flagOutput
	if outputPath == "" {
		outputPath = filepath.Join(os.TempDir(), "orchestration.json")
	}

	// Create state manager
	sm := NewStateManager(outputPath)

	// Initialize agents
	agentIDs := []AgentID{AgentGoSetup, AgentQdrant, AgentDeepSeek}
	if err := sm.Initialize(agentIDs); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	// Create coordinator
	coord := NewCoordinator(sm)

	// Register agents
	coord.RegisterAgent(NewGoSetupAgent())
	coord.RegisterAgent(NewQdrantAgent())
	coord.RegisterAgent(NewDeepSeekAgent())

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║        PARALLEL ORCHESTRATION SYSTEM                          ║")
	fmt.Println("║        Multi-Service Project Setup                             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Agents:                                                      ║")
	fmt.Println("║    1. Go/go-agent Setup (needs: qdrant_url)                   ║")
	fmt.Println("║    2. Qdrant/Docker Infrastructure (no dependencies)          ║")
	fmt.Println("║    3. DeepSeek API Configuration (no dependencies)            ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Output: %s                          \n", outputPath)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	// Run orchestration
	var err error
	if *flagSequential {
		err = runSequential(ctx, coord, sm)
	} else {
		err = coord.RunAll(ctx)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Orchestration failed: %v\n", err)
		os.Exit(1)
	}

	// Print report
	coord.PrintReport()

	// Print output file location
	fmt.Printf("\n📄 Orchestration state saved to: %s\n", outputPath)

	// Print time saved
	state := sm.GetState()
	if state.TimeSaved > 0 {
		fmt.Printf("⏱️  Time saved: %v (parallel vs sequential)\n", state.TimeSaved.Round(time.Millisecond))
	}
}

// runSequential runs agents one by one (for timing comparison)
func runSequential(ctx context.Context, coord *Coordinator, sm *StateManager) error {
	fmt.Println("\n🔄 Running sequentially (comparison mode)...")

	agents := []Agent{
		NewQdrantAgent(),      // Run Qdrant first (provides dependency)
		NewDeepSeekAgent(),    // Then DeepSeek
		NewGoSetupAgent(),     // Finally Go setup (needs qdrant_url)
	}

	for _, agent := range agents {
		fmt.Printf("   Running %s...\n", agent.Name())
		if err := agent.Run(ctx, sm); err != nil {
			return err
		}
	}

	return coord.runFinalTest(ctx)
}