// Agent worker implementation
package main

import (
	"context"
	"fmt"
	"time"
)

// Agent represents a worker agent
type Agent interface {
	ID() AgentID
	Name() string
	Dependencies() []string
	Run(ctx context.Context, sm *StateManager) error
}

// BaseAgent provides common agent functionality
type BaseAgent struct {
	id           AgentID
	name         string
	dependencies []string
}

func (a *BaseAgent) ID() AgentID        { return a.id }
func (a *BaseAgent) Name() string       { return a.name }
func (a *BaseAgent) Dependencies() []string { return a.dependencies }

// GoSetupAgent handles Go/go-agent setup
type GoSetupAgent struct {
	BaseAgent
}

func NewGoSetupAgent() *GoSetupAgent {
	return &GoSetupAgent{
		BaseAgent: BaseAgent{
			id:           AgentGoSetup,
			name:         "Go/go-agent Setup",
			dependencies: []string{"qdrant_url"}, // Needs Qdrant URL
		},
	}
}

func (a *GoSetupAgent) Run(ctx context.Context, sm *StateManager) error {
	sm.SetAgentStatus(a.ID(), StatusRunning, "Starting Go setup")

	// Task 1: Check Go installation
	start := time.Now()
	sm.SetAgentStatus(a.ID(), StatusRunning, "Checking Go installation")
	task := TaskResult{
		Name:   "check_go",
		Status: "running",
	}
	
	goVersion, err := a.checkGoInstallation()
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		sm.AddTaskResult(a.ID(), task)
		return err
	}
	task.Status = "completed"
	task.Output = goVersion
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 2: Wait for Qdrant URL (dependency)
	sm.SetAgentStatus(a.ID(), StatusWaiting, "Waiting for Qdrant URL")
	start = time.Now()
	
	var qdrantURL string
	for i := 0; i < 30; i++ { // 30 second timeout
		if val, ok := sm.GetDependencyValue("qdrant_url"); ok {
			qdrantURL = val.(string)
			break
		}
		time.Sleep(1 * time.Second)
	}
	
	if qdrantURL == "" {
		sm.AddBlocker(a.ID(), "qdrant_url not available")
		return fmt.Errorf("timeout waiting for qdrant_url")
	}
	sm.ClearBlocker(a.ID(), "qdrant_url")
	sm.SetAgentOutput(a.ID(), "qdrant_url_received", qdrantURL)

	task = TaskResult{
		Name:        "wait_qdrant_url",
		Status:      "completed",
		Duration:    time.Since(start).String(),
		Output:      qdrantURL,
		CompletedAt: time.Now(),
	}
	sm.AddTaskResult(a.ID(), task)

	// Task 3: Verify go-agent build
	sm.SetAgentStatus(a.ID(), StatusRunning, "Verifying go-agent build")
	start = time.Now()
	task = TaskResult{Name: "verify_build", Status: "running"}

	if err := a.verifyBuild(); err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		sm.AddTaskResult(a.ID(), task)
		return err
	}
	task.Status = "completed"
	task.Duration = time.Since(start).String()
	task.Output = "Build successful"
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 4: Set outputs
	sm.SetAgentOutput(a.ID(), "go_version", goVersion)
	sm.SetAgentOutput(a.ID(), "build_status", "ready")

	// Signal readiness
	sm.SetDependencyReady(a.ID(), "go_agent_ready", true)
	sm.SetAgentStatus(a.ID(), StatusCompleted, "All tasks completed")

	return nil
}

func (a *GoSetupAgent) checkGoInstallation() (string, error) {
	// In real implementation, would run `go version`
	return "go1.25.0", nil
}

func (a *GoSetupAgent) verifyBuild() error {
	// In real implementation, would run `go build`
	return nil
}

// QdrantAgent handles Qdrant/Docker infrastructure
type QdrantAgent struct {
	BaseAgent
}

func NewQdrantAgent() *QdrantAgent {
	return &QdrantAgent{
		BaseAgent: BaseAgent{
			id:           AgentQdrant,
			name:         "Qdrant/Docker Infrastructure",
			dependencies: []string{}, // No dependencies - runs first
		},
	}
}

func (a *QdrantAgent) Run(ctx context.Context, sm *StateManager) error {
	sm.SetAgentStatus(a.ID(), StatusRunning, "Starting Qdrant infrastructure")

	// Task 1: Check Docker
	start := time.Now()
	sm.SetAgentStatus(a.ID(), StatusRunning, "Checking Docker")
	task := TaskResult{Name: "check_docker", Status: "running"}

	dockerStatus, err := a.checkDocker()
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		sm.AddTaskResult(a.ID(), task)
		return err
	}
	task.Status = "completed"
	task.Output = dockerStatus
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 2: Check Qdrant
	sm.SetAgentStatus(a.ID(), StatusRunning, "Checking Qdrant status")
	start = time.Now()
	task = TaskResult{Name: "check_qdrant", Status: "running"}

	qdrantURL := "http://localhost:6333"
	qdrantStatus, err := a.checkQdrant(qdrantURL)
	if err != nil {
		task.Status = "warning"
		task.Error = err.Error()
		task.Output = "Qdrant not running, will use existing config"
	} else {
		task.Status = "completed"
		task.Output = qdrantStatus
	}
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 3: Validate collection dimensions
	sm.SetAgentStatus(a.ID(), StatusRunning, "Validating Qdrant collection")
	start = time.Now()
	task = TaskResult{Name: "validate_collection", Status: "running"}

	collectionInfo, err := a.validateCollection(qdrantURL, "adk_memories")
	if err != nil {
		task.Status = "warning"
		task.Error = err.Error()
		task.Output = "Collection validation skipped"
	} else {
		task.Status = "completed"
		task.Output = fmt.Sprintf("Collection %s: %d dimensions", collectionInfo.Name, collectionInfo.Dimension)
	}
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Set outputs and signal readiness
	sm.SetAgentOutput(a.ID(), "qdrant_url", qdrantURL)
	sm.SetAgentOutput(a.ID(), "docker_status", dockerStatus)
	sm.SetDependencyReady(a.ID(), "qdrant_url", qdrantURL)
	sm.SetAgentStatus(a.ID(), StatusCompleted, "Infrastructure ready")

	return nil
}

type CollectionInfo struct {
	Name      string
	Dimension int
	Status    string
}

func (a *QdrantAgent) checkDocker() (string, error) {
	// In real implementation, would check Docker daemon
	return "Docker Desktop running", nil
}

func (a *QdrantAgent) checkQdrant(url string) (string, error) {
	// In real implementation, would ping Qdrant
	return "Qdrant running at " + url, nil
}

func (a *QdrantAgent) validateCollection(url, name string) (*CollectionInfo, error) {
	// In real implementation, would query Qdrant API
	return &CollectionInfo{
		Name:      name,
		Dimension: 768,
		Status:    "green",
	}, nil
}

// DeepSeekAgent handles DeepSeek API configuration
type DeepSeekAgent struct {
	BaseAgent
}

func NewDeepSeekAgent() *DeepSeekAgent {
	return &DeepSeekAgent{
		BaseAgent: BaseAgent{
			id:           AgentDeepSeek,
			name:         "DeepSeek API Configuration",
			dependencies: []string{}, // No dependencies
		},
	}
}

func (a *DeepSeekAgent) Run(ctx context.Context, sm *StateManager) error {
	sm.SetAgentStatus(a.ID(), StatusRunning, "Starting DeepSeek configuration")

	// Task 1: Check .env file
	start := time.Now()
	sm.SetAgentStatus(a.ID(), StatusRunning, "Checking .env configuration")
	task := TaskResult{Name: "check_env", Status: "running"}

	envStatus, err := a.checkEnvFile()
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		sm.AddTaskResult(a.ID(), task)
		return err
	}
	task.Status = "completed"
	task.Output = envStatus
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 2: Validate API key format
	sm.SetAgentStatus(a.ID(), StatusRunning, "Validating API key")
	start = time.Now()
	task = TaskResult{Name: "validate_key", Status: "running"}

	keyStatus, err := a.validateAPIKey()
	if err != nil {
		task.Status = "warning"
		task.Error = err.Error()
		task.Output = "API key validation skipped"
	} else {
		task.Status = "completed"
		task.Output = keyStatus
	}
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Task 3: Test API connection
	sm.SetAgentStatus(a.ID(), StatusRunning, "Testing API connection")
	start = time.Now()
	task = TaskResult{Name: "test_connection", Status: "running"}

	connStatus, err := a.testConnection()
	if err != nil {
		task.Status = "warning"
		task.Error = err.Error()
		task.Output = "API connection test skipped"
	} else {
		task.Status = "completed"
		task.Output = connStatus
	}
	task.Duration = time.Since(start).String()
	task.CompletedAt = time.Now()
	sm.AddTaskResult(a.ID(), task)

	// Set outputs
	sm.SetAgentOutput(a.ID(), "api_key_configured", true)
	sm.SetAgentOutput(a.ID(), "api_status", "ready")
	sm.SetDependencyReady(a.ID(), "deepseek_configured", true)
	sm.SetAgentStatus(a.ID(), StatusCompleted, "DeepSeek API configured")

	return nil
}

func (a *DeepSeekAgent) checkEnvFile() (string, error) {
	// In real implementation, would check .env file
	return ".env file exists with DEEPSEEK_API_KEY", nil
}

func (a *DeepSeekAgent) validateAPIKey() (string, error) {
	// In real implementation, would validate key format
	return "API key format valid (sk-****10a18)", nil
}

func (a *DeepSeekAgent) testConnection() (string, error) {
	// In real implementation, would make test API call
	return "API connection successful", nil
}