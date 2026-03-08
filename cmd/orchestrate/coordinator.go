// Coordinator manages agent orchestration and dependency resolution
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Coordinator manages parallel agent execution
type Coordinator struct {
	sm        *StateManager
	agents    map[AgentID]Agent
	handoffs  chan Handoff
	errors    chan error
	wg        sync.WaitGroup
}

// NewCoordinator creates a new coordinator
func NewCoordinator(sm *StateManager) *Coordinator {
	return &Coordinator{
		sm:       sm,
		agents:   make(map[AgentID]Agent),
		handoffs: make(chan Handoff, 100),
		errors:   make(chan error, 10),
	}
}

// RegisterAgent registers an agent with the coordinator
func (c *Coordinator) RegisterAgent(agent Agent) {
	c.agents[agent.ID()] = agent

	// Register dependencies in state
	for _, dep := range agent.Dependencies() {
		// Find which agent provides this dependency
		for _, other := range c.agents {
			if other.ID() == agent.ID() {
				continue
			}
			// For now, we'll map resources to providers
			if dep == "qdrant_url" && other.ID() == AgentQdrant {
				c.sm.AddDependency(other.ID(), agent.ID(), dep)
			}
		}
	}
}

// RunAll runs all agents in parallel
func (c *Coordinator) RunAll(ctx context.Context) error {
	// Calculate theoretical sequential time
	sequentialTime := c.calculateSequentialTime()
	c.sm.state.SequentialTime = sequentialTime

	// Start all agents
	fmt.Println("\n🚀 Starting parallel orchestration...")
	fmt.Printf("   Sequential estimate: %v\n", sequentialTime.Round(time.Millisecond))
	fmt.Println()

	// Create child context for cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start agent goroutines
	for id, agent := range c.agents {
		c.wg.Add(1)
		go func(a Agent) {
			defer c.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					c.errors <- fmt.Errorf("agent %s panicked: %v", a.ID(), r)
				}
			}()

			if err := a.Run(ctx, c.sm); err != nil {
				c.errors <- fmt.Errorf("agent %s failed: %w", a.ID(), err)
				cancel()
			}
		}(agent)
		c.sm.SetAgentStatus(id, StatusRunning, "Starting")
	}

	// Wait for all agents with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All agents completed
		return c.runFinalTest(ctx)
	case err := <-c.errors:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// calculateSequentialTime estimates time for sequential execution
func (c *Coordinator) calculateSequentialTime() time.Duration {
	// Estimated times for each agent
	estimates := map[AgentID]time.Duration{
		AgentGoSetup:  5 * time.Second,
		AgentQdrant:   8 * time.Second,
		AgentDeepSeek: 3 * time.Second,
	}

	total := time.Duration(0)
	for id := range c.agents {
		if d, ok := estimates[id]; ok {
			total += d
		}
	}
	return total
}

// runFinalTest runs the integration test
func (c *Coordinator) runFinalTest(ctx context.Context) error {
	fmt.Println("\n🔬 Running final integration test...")

	start := time.Now()
	sm := c.sm.GetState()

	// Check all agents completed
	allCompleted := true
	for _, agent := range sm.Agents {
		if agent.Status != StatusCompleted {
			allCompleted = false
			fmt.Printf("   ⚠️  Agent %s not completed: %s\n", agent.Name, agent.Status)
		}
	}

	// Build test result
	result := &TestResult{
		Status:    "passed",
		Message:   "All systems integrated successfully",
		Timestamp: time.Now(),
		Details:   make(map[string]any),
	}

	// Collect outputs from all agents
	for id, agent := range sm.Agents {
		result.Details[string(id)] = map[string]any{
			"status":  agent.Status,
			"outputs": agent.Outputs,
			"tasks":   len(agent.Tasks),
		}
	}

	if !allCompleted {
		result.Status = "failed"
		result.Message = "Some agents did not complete"
	}

	result.Duration = time.Since(start).String()
	c.sm.SetFinalTest(result)

	return nil
}

// PrintReport prints the final orchestration report
func (c *Coordinator) PrintReport() {
	sm := c.sm.GetState()

	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          PARALLEL ORCHESTRATION REPORT                        ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	fmt.Printf("║  Session: %s                                     \n", sm.SessionID)
	fmt.Printf("║  Started: %s                           \n", sm.StartedAt.Format("2006-01-02 15:04:05"))
	if sm.CompletedAt != nil {
		fmt.Printf("║  Completed: %s                          \n", sm.CompletedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Agent summary
	fmt.Println("║  AGENTS                                                       ║")
	for _, agent := range sm.Agents {
		statusIcon := "✅"
		if agent.Status != StatusCompleted {
			statusIcon = "⚠️"
		}
		fmt.Printf("║    %s %-25s %-15s (%d tasks)   \n",
			statusIcon, agent.Name, agent.Status, len(agent.Tasks))
		for _, task := range agent.Tasks {
			fmt.Printf("║        └─ %s: %s (%s)              \n",
				task.Name, task.Status, task.Duration)
		}
	}

	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Handoffs
	if len(sm.Handoffs) > 0 {
		fmt.Println("║  HANDOFFS                                                     ║")
		for _, h := range sm.Handoffs {
			fmt.Printf("║    %s → %s: %s                      \n",
				h.From, h.Resource, h.Value)
		}
		fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	}

	// Time savings
	fmt.Println("║  TIMING                                                       ║")
	fmt.Printf("║    Sequential estimate: %v                              \n",
		sm.SequentialTime.Round(time.Millisecond))
	fmt.Printf("║    Parallel actual:     %v                              \n",
		sm.ParallelTime.Round(time.Millisecond))
	if sm.TimeSaved > 0 {
		fmt.Printf("║    Time saved:          %v 🚀                          \n",
			sm.TimeSaved.Round(time.Millisecond))
		savedPercent := float64(sm.TimeSaved) / float64(sm.SequentialTime) * 100
		fmt.Printf("║    Efficiency gain:     %.1f%%                               \n",
			savedPercent)
	}

	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	// Final test
	if sm.FinalTest != nil {
		statusIcon := "✅"
		if sm.FinalTest.Status != "passed" {
			statusIcon = "❌"
		}
		fmt.Printf("║  FINAL TEST: %s %s (%s)                     \n",
			statusIcon, sm.FinalTest.Status, sm.FinalTest.Duration)
		fmt.Printf("║    %s                                          \n", sm.FinalTest.Message)
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

// DetectWaitingAgents detects agents that are waiting on dependencies
func (c *Coordinator) DetectWaitingAgents() []AgentID {
	waiting := []AgentID{}
	sm := c.sm.GetState()

	for id, agent := range sm.Agents {
		if agent.Status == StatusWaiting || agent.Status == StatusBlocked {
			waiting = append(waiting, id)
		}
	}

	return waiting
}

// FacilitateHandoff facilitates a handoff between agents
func (c *Coordinator) FacilitateHandoff(from, to AgentID, resource string) error {
	sm := c.sm.GetState()

	// Find the handoff
	for _, h := range sm.Handoffs {
		if h.From == from && h.Resource == resource {
			// Notify waiting agent
			c.handoffs <- h
			return nil
		}
	}

	return fmt.Errorf("handoff %s from %s to %s not found", resource, from, to)
}