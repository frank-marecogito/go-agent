// Package orchestrate provides parallel orchestration for multi-service projects
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	StatusPending    AgentStatus = "pending"
	StatusRunning    AgentStatus = "running"
	StatusWaiting    AgentStatus = "waiting"    // Waiting for dependency
	StatusCompleted  AgentStatus = "completed"
	StatusFailed     AgentStatus = "failed"
	StatusBlocked    AgentStatus = "blocked"    // Blocked by another agent
)

// AgentID identifies an agent in the orchestration
type AgentID string

const (
	AgentGoSetup     AgentID = "go-agent-setup"
	AgentQdrant      AgentID = "qdrant-infra"
	AgentDeepSeek    AgentID = "deepseek-config"
)

// Dependency represents a dependency between agents
type Dependency struct {
	From      AgentID     `json:"from"`       // Agent that provides
	To        AgentID     `json:"to"`         // Agent that needs
	Resource  string      `json:"resource"`   // What is being provided
	Status    string      `json:"status"`     // "pending", "ready", "consumed"
	Timestamp time.Time   `json:"timestamp"`
}

// AgentProgress represents the progress of a single agent
type AgentProgress struct {
	ID           AgentID       `json:"id"`
	Name         string        `json:"name"`
	Status       AgentStatus   `json:"status"`
	StartedAt    *time.Time    `json:"started_at,omitempty"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	CurrentTask  string        `json:"current_task"`
	Tasks        []TaskResult  `json:"tasks"`
	Outputs      map[string]any `json:"outputs"`
	Dependencies []string      `json:"dependencies"`
	Blockers     []string      `json:"blockers"`
	Error        string        `json:"error,omitempty"`
}

// TaskResult represents the result of a single task
type TaskResult struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Duration    string    `json:"duration"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// OrchestrationState represents the shared state file
type OrchestrationState struct {
	mu sync.RWMutex

	SessionID      string                    `json:"session_id"`
	StartedAt      time.Time                 `json:"started_at"`
	CompletedAt    *time.Time                `json:"completed_at,omitempty"`
	Agents         map[AgentID]*AgentProgress `json:"agents"`
	Dependencies   []Dependency              `json:"dependencies"`
	Handoffs       []Handoff                 `json:"handoffs"`
	FinalTest      *TestResult               `json:"final_test,omitempty"`
	SequentialTime time.Duration             `json:"sequential_time"`
	ParallelTime   time.Duration             `json:"parallel_time"`
	TimeSaved      time.Duration             `json:"time_saved"`
}

// Handoff represents a resource handoff between agents
type Handoff struct {
	From      AgentID   `json:"from"`
	To        AgentID   `json:"to"`
	Resource  string    `json:"resource"`
	Value     any       `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// TestResult represents the result of the final integration test
type TestResult struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Duration  string    `json:"duration"`
	Details   map[string]any `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// StateManager manages the shared orchestration state
type StateManager struct {
	filePath string
	state    *OrchestrationState
	mu       sync.RWMutex
}

// NewStateManager creates a new state manager
func NewStateManager(filePath string) *StateManager {
	return &StateManager{
		filePath: filePath,
		state: &OrchestrationState{
			SessionID:    fmt.Sprintf("orch-%d", time.Now().Unix()),
			StartedAt:    time.Now(),
			Agents:       make(map[AgentID]*AgentProgress),
			Dependencies: []Dependency{},
			Handoffs:     []Handoff{},
		},
	}
}

// Initialize initializes all agents in the state
func (sm *StateManager) Initialize(agentIDs []AgentID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, id := range agentIDs {
		sm.state.Agents[id] = &AgentProgress{
			ID:           id,
			Name:         getAgentName(id),
			Status:       StatusPending,
			Tasks:        []TaskResult{},
			Outputs:      make(map[string]any),
			Dependencies: []string{},
			Blockers:     []string{},
		}
	}

	return sm.save()
}

// UpdateAgent updates an agent's progress
func (sm *StateManager) UpdateAgent(id AgentID, update func(*AgentProgress)) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	agent, ok := sm.state.Agents[id]
	if !ok {
		return fmt.Errorf("agent %s not found", id)
	}

	update(agent)
	return sm.save()
}

// SetAgentStatus sets an agent's status
func (sm *StateManager) SetAgentStatus(id AgentID, status AgentStatus, currentTask string) error {
	return sm.UpdateAgent(id, func(a *AgentProgress) {
		a.Status = status
		a.CurrentTask = currentTask
		
		if status == StatusRunning && a.StartedAt == nil {
			now := time.Now()
			a.StartedAt = &now
		}
		if status == StatusCompleted || status == StatusFailed {
			now := time.Now()
			a.CompletedAt = &now
		}
	})
}

// AddTaskResult adds a task result to an agent
func (sm *StateManager) AddTaskResult(id AgentID, task TaskResult) error {
	return sm.UpdateAgent(id, func(a *AgentProgress) {
		a.Tasks = append(a.Tasks, task)
	})
}

// SetAgentOutput sets an output value for an agent
func (sm *StateManager) SetAgentOutput(id AgentID, key string, value any) error {
	return sm.UpdateAgent(id, func(a *AgentProgress) {
		a.Outputs[key] = value
	})
}

// AddBlocker adds a blocker to an agent
func (sm *StateManager) AddBlocker(id AgentID, blocker string) error {
	return sm.UpdateAgent(id, func(a *AgentProgress) {
		for _, b := range a.Blockers {
			if b == blocker {
				return
			}
		}
		a.Blockers = append(a.Blockers, blocker)
	})
}

// ClearBlocker clears a blocker from an agent
func (sm *StateManager) ClearBlocker(id AgentID, blocker string) error {
	return sm.UpdateAgent(id, func(a *AgentProgress) {
		newBlockers := []string{}
		for _, b := range a.Blockers {
			if b != blocker {
				newBlockers = append(newBlockers, b)
			}
		}
		a.Blockers = newBlockers
	})
}

// AddDependency adds a dependency between agents
func (sm *StateManager) AddDependency(from, to AgentID, resource string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	dep := Dependency{
		From:      from,
		To:        to,
		Resource:  resource,
		Status:    "pending",
		Timestamp: time.Now(),
	}

	sm.state.Dependencies = append(sm.state.Dependencies, dep)
	return sm.save()
}

// SetDependencyReady marks a dependency as ready
func (sm *StateManager) SetDependencyReady(from AgentID, resource string, value any) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, dep := range sm.state.Dependencies {
		if dep.From == from && dep.Resource == resource {
			sm.state.Dependencies[i].Status = "ready"
			sm.state.Dependencies[i].Timestamp = time.Now()
		}
	}

	// Also add to handoffs
	sm.state.Handoffs = append(sm.state.Handoffs, Handoff{
		From:      from,
		Resource:  resource,
		Value:     value,
		Timestamp: time.Now(),
	})

	return sm.save()
}

// GetDependencyValue gets the value of a ready dependency
func (sm *StateManager) GetDependencyValue(resource string) (any, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, h := range sm.state.Handoffs {
		if h.Resource == resource {
			return h.Value, true
		}
	}
	return nil, false
}

// IsDependencyReady checks if a dependency is ready
func (sm *StateManager) IsDependencyReady(resource string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, dep := range sm.state.Dependencies {
		if dep.Resource == resource && dep.Status == "ready" {
			return true
		}
	}
	return false
}

// SetFinalTest sets the final test result
func (sm *StateManager) SetFinalTest(result *TestResult) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.FinalTest = result
	now := time.Now()
	sm.state.CompletedAt = &now
	sm.state.ParallelTime = now.Sub(sm.state.StartedAt)
	sm.state.TimeSaved = sm.state.SequentialTime - sm.state.ParallelTime

	return sm.save()
}

// GetState returns the current state
func (sm *StateManager) GetState() *OrchestrationState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// Complete marks the orchestration as complete
func (sm *StateManager) Complete() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	sm.state.CompletedAt = &now
	sm.state.ParallelTime = now.Sub(sm.state.StartedAt)
	sm.state.TimeSaved = sm.state.SequentialTime - sm.state.ParallelTime

	return sm.save()
}

// save saves the state to file
func (sm *StateManager) save() error {
	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.filePath, data, 0644)
}

// Load loads the state from file
func (sm *StateManager) Load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, sm.state)
}

func getAgentName(id AgentID) string {
	names := map[AgentID]string{
		AgentGoSetup:  "Go/go-agent Setup",
		AgentQdrant:   "Qdrant/Docker Infrastructure",
		AgentDeepSeek: "DeepSeek API Configuration",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return string(id)
}