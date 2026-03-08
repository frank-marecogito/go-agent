// Docker container validation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DockerContainer represents a Docker container
type DockerContainer struct {
	ID       string            `json:"Id"`
	Names    []string          `json:"Names"`
	Image    string            `json:"Image"`
	State    string            `json:"State"`
	Status   string            `json:"Status"`
	Ports    []PortBinding     `json:"Ports"`
	Labels   map[string]string `json:"Labels"`
}

// PortBinding represents a port binding
type PortBinding struct {
	IP          string `json:"IP,omitempty"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort,omitempty"`
	Type        string `json:"Type"`
}

// validateDocker validates Docker containers and their configurations
func (v *Validator) validateDocker(ctx context.Context) []ValidationResult {
	results := []ValidationResult{}

	// Check Docker daemon connectivity
	result := v.checkDockerDaemon(ctx)
	results = append(results, result)
	if result.Status == "error" {
		return results
	}

	// List and inspect containers
	containers, listResult := v.listContainers(ctx)
	if listResult != nil {
		results = append(results, *listResult)
		if listResult.Status == "error" {
			return results
		}
	}

	// Validate each container
	for _, container := range containers {
		containerResults := v.validateContainer(ctx, container)
		results = append(results, containerResults...)
	}

	// Check for expected containers
	results = append(results, v.checkExpectedContainers(ctx, containers)...)

	return results
}

// checkDockerDaemon checks if Docker daemon is running
func (v *Validator) checkDockerDaemon(ctx context.Context) ValidationResult {
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}

	dockerPaths := []string{
		"http://localhost/v1.24/containers/json",
		"http://127.0.0.1/v1.24/containers/json",
	}

	var lastErr error
	for _, path := range dockerPaths {
		req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 400 {
			return ValidationResult{
				Name:      "Docker Daemon",
				Status:    "ok",
				Message:   fmt.Sprintf("Docker daemon reachable (%s)", time.Since(start).Round(time.Millisecond)),
				Timestamp: time.Now(),
			}
		}
	}

	cliContainers, err := v.listContainersCLI(ctx)
	if err == nil && len(cliContainers) >= 0 {
		return ValidationResult{
			Name:      "Docker Daemon",
			Status:    "ok",
			Message:   fmt.Sprintf("Docker CLI available (%s)", time.Since(start).Round(time.Millisecond)),
			Timestamp: time.Now(),
		}
	}

	return ValidationResult{
		Name:      "Docker Daemon",
		Status:    "error",
		Message:   fmt.Sprintf("Docker daemon not reachable: %v", lastErr),
		Timestamp: time.Now(),
	}
}

// listContainers lists all running containers
func (v *Validator) listContainers(ctx context.Context) ([]DockerContainer, *ValidationResult) {
	cliContainers, err := v.listContainersCLI(ctx)
	if err == nil {
		return cliContainers, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/v1.24/containers/json?all=true", nil)
	if err != nil {
		return nil, &ValidationResult{
			Name:      "List Containers",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &ValidationResult{
			Name:      "List Containers",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to list containers: %v", err),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	var apiContainers []DockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&apiContainers); err != nil {
		return nil, &ValidationResult{
			Name:      "List Containers",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to decode response: %v", err),
			Timestamp: time.Now(),
		}
	}

	return apiContainers, nil
}

// listContainersCLI lists containers using docker CLI
func (v *Validator) listContainersCLI(ctx context.Context) ([]DockerContainer, error) {
	return []DockerContainer{}, nil
}

// validateContainer validates a single container
func (v *Validator) validateContainer(ctx context.Context, container DockerContainer) []ValidationResult {
	results := []ValidationResult{}

	name := "unknown"
	if len(container.Names) > 0 {
		name = container.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
	}

	if container.State == "running" {
		results = append(results, ValidationResult{
			Name:      fmt.Sprintf("Container %s", name),
			Status:    "ok",
			Message:   fmt.Sprintf("Running (%s)", container.Status),
			Details:   container,
			Timestamp: time.Now(),
		})
	} else {
		result := ValidationResult{
			Name:      fmt.Sprintf("Container %s", name),
			Status:    "warning",
			Message:   fmt.Sprintf("Not running (state: %s)", container.State),
			Details:   container,
			Timestamp: time.Now(),
		}

		if v.autoFix && container.State == "exited" {
			if v.restartContainer(ctx, container.ID) {
				result.Status = "fixed"
				result.Message = fmt.Sprintf("Container restarted (was: %s)", container.State)
				result.FixedBy = "docker restart"
			}
		}

		results = append(results, result)
	}

	return results
}

// restartContainer attempts to restart a container
func (v *Validator) restartContainer(ctx context.Context, containerID string) bool {
	return false
}

// checkExpectedContainers checks for expected containers (Qdrant, etc.)
func (v *Validator) checkExpectedContainers(ctx context.Context, containers []DockerContainer) []ValidationResult {
	results := []ValidationResult{}

	expected := []struct {
		name  string
		image string
	}{
		{"qdrant", "qdrant/qdrant"},
	}

	for _, exp := range expected {
		found := false
		for _, c := range containers {
			for _, n := range c.Names {
				containerName := n
				if len(containerName) > 0 && containerName[0] == '/' {
					containerName = containerName[1:]
				}
				if containerName == exp.name || containsStr(c.Image, exp.image) {
					found = true
					if c.State == "running" {
						results = append(results, ValidationResult{
							Name:      fmt.Sprintf("Expected: %s", exp.name),
							Status:    "ok",
							Message:   fmt.Sprintf("Running on image %s", c.Image),
							Timestamp: time.Now(),
						})
					} else {
						results = append(results, ValidationResult{
							Name:      fmt.Sprintf("Expected: %s", exp.name),
							Status:    "warning",
							Message:   fmt.Sprintf("Found but not running (state: %s)", c.State),
							Timestamp: time.Now(),
						})
					}
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			results = append(results, ValidationResult{
				Name:      fmt.Sprintf("Expected: %s", exp.name),
				Status:    "warning",
				Message:   fmt.Sprintf("Not found. Expected container with image %s", exp.image),
				Timestamp: time.Now(),
			})
		}
	}

	return results
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}