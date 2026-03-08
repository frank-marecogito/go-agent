// API connectivity validation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// APIProvider represents an API provider configuration
type APIProvider struct {
	Name        string
	EnvKey      string
	EnvKeyAlt   string   // Alternative env key
	TestURL     string
	TestMethod  string
	Headers     map[string]string
	ExpectedStatusRange [2]int // [min, max] expected status codes
}

// validateAPIs validates API connectivity for all providers
func (v *Validator) validateAPIs(ctx context.Context) []ValidationResult {
	results := []ValidationResult{}

	// Define API providers to check
	providers := []APIProvider{
		{
			Name:        "OpenAI",
			EnvKey:      "OPENAI_API_KEY",
			EnvKeyAlt:   "OPENAI_KEY",
			TestURL:     "https://api.openai.com/v1/models",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
		{
			Name:        "DeepSeek",
			EnvKey:      "DEEPSEEK_API_KEY",
			TestURL:     "https://api.deepseek.com/v1/models",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
		{
			Name:        "Google/Gemini",
			EnvKey:      "GOOGLE_API_KEY",
			EnvKeyAlt:   "GEMINI_API_KEY",
			TestURL:     "https://generativelanguage.googleapis.com/v1/models",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
		{
			Name:        "Anthropic",
			EnvKey:      "ANTHROPIC_API_KEY",
			TestURL:     "https://api.anthropic.com/v1/models",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
		{
			Name:        "Ollama",
			EnvKey:      "OLLAMA_HOST",
			TestURL:     "http://localhost:11434/api/tags",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
		{
			Name:        "Voyage AI",
			EnvKey:      "VOYAGE_API_KEY",
			TestURL:     "https://api.voyageai.com/v1/models",
			TestMethod:  "GET",
			ExpectedStatusRange: [2]int{200, 299},
		},
	}

	for _, provider := range providers {
		result := v.validateAPIProvider(ctx, provider)
		results = append(results, result)
	}

	return results
}

// validateAPIProvider validates a single API provider
func (v *Validator) validateAPIProvider(ctx context.Context, provider APIProvider) ValidationResult {
	start := time.Now()

	// Check if API key exists (if required)
	apiKey := os.Getenv(provider.EnvKey)
	if apiKey == "" && provider.EnvKeyAlt != "" {
		apiKey = os.Getenv(provider.EnvKeyAlt)
	}

	// Special handling for Ollama (uses host URL instead of API key)
	if provider.Name == "Ollama" {
		host := os.Getenv("OLLAMA_HOST")
		if host != "" {
			// Update test URL with custom host
			provider.TestURL = host + "/api/tags"
		}
	}

	// Build request
	client := &http.Client{Timeout: 15 * time.Second}

	var req *http.Request
	var err error

	if provider.TestMethod == "GET" {
		req, err = http.NewRequestWithContext(ctx, "GET", provider.TestURL, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, provider.TestMethod, provider.TestURL, nil)
	}

	if err != nil {
		return ValidationResult{
			Name:      provider.Name,
			Status:    "error",
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		}
	}

	// Set headers
	for key, value := range provider.Headers {
		req.Header.Set(key, value)
	}

	// Set authorization header if API key exists
	if apiKey != "" {
		switch provider.Name {
		case "OpenAI", "DeepSeek", "Voyage AI":
			req.Header.Set("Authorization", "Bearer "+apiKey)
		case "Anthropic":
			req.Header.Set("x-api-key", apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		case "Google/Gemini":
			// Google uses query parameter
			q := req.URL.Query()
			q.Add("key", apiKey)
			req.URL.RawQuery = q.Encode()
		}
	}

	// Check if API key is configured
	if apiKey == "" && provider.Name != "Ollama" {
		return ValidationResult{
			Name:      provider.Name,
			Status:    "warning",
			Message:   fmt.Sprintf("API key not configured (%s)", provider.EnvKey),
			Details: map[string]any{
				"env_key": provider.EnvKey,
			},
			Timestamp: time.Now(),
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		result := ValidationResult{
			Name:      provider.Name,
			Status:    "error",
			Message:   fmt.Sprintf("Connection failed: %v", err),
			Timestamp: time.Now(),
		}

		// Check for common issues
		if v.autoFix {
			// Check for proxy issues
			if isProxyError(err) {
				if v.tryFixProxy(ctx, provider.Name) {
					result.Status = "fixed"
					result.Message = "Fixed proxy configuration"
					result.FixedBy = "proxy fix"
				}
			}
		}

		return result
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= provider.ExpectedStatusRange[0] && resp.StatusCode <= provider.ExpectedStatusRange[1] {
		// Try to parse response for model list
		models := v.parseModelsResponse(provider.Name, resp)
		
		return ValidationResult{
			Name:      provider.Name,
			Status:    "ok",
			Message:   fmt.Sprintf("Connected (%s, %d models available)", time.Since(start).Round(time.Millisecond), len(models)),
			Details: map[string]any{
				"models": models,
				"status": resp.StatusCode,
			},
			Timestamp: time.Now(),
		}
	}

	// Handle error responses
	errorMsg := fmt.Sprintf("API returned status %d", resp.StatusCode)
	
	switch resp.StatusCode {
	case 401:
		errorMsg = "Invalid API key"
	case 403:
		errorMsg = "Access forbidden - check API key permissions"
	case 429:
		errorMsg = "Rate limited - try again later"
	case 500, 502, 503:
		errorMsg = fmt.Sprintf("Server error (status %d)", resp.StatusCode)
	}

	return ValidationResult{
		Name:      provider.Name,
		Status:    "error",
		Message:   errorMsg,
		Details: map[string]any{
			"status_code": resp.StatusCode,
		},
		Timestamp: time.Now(),
	}
}

// parseModelsResponse parses the models from an API response
func (v *Validator) parseModelsResponse(providerName string, resp *http.Response) []string {
	models := []string{}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models
	}

	// Extract models based on provider format
	switch providerName {
	case "OpenAI", "DeepSeek":
		if dataVal, ok := data["data"].([]interface{}); ok {
			for _, m := range dataVal {
				if model, ok := m.(map[string]interface{}); ok {
					if id, ok := model["id"].(string); ok {
						models = append(models, id)
					}
				}
			}
		}
	case "Google/Gemini":
		if modelsVal, ok := data["models"].([]interface{}); ok {
			for _, m := range modelsVal {
				if model, ok := m.(map[string]interface{}); ok {
					if name, ok := model["name"].(string); ok {
						models = append(models, name)
					}
				}
			}
		}
	case "Anthropic":
		if dataVal, ok := data["data"].([]interface{}); ok {
			for _, m := range dataVal {
				if model, ok := m.(map[string]interface{}); ok {
					if id, ok := model["id"].(string); ok {
						models = append(models, id)
					}
				}
			}
		}
	case "Ollama":
		if modelsVal, ok := data["models"].([]interface{}); ok {
			for _, m := range modelsVal {
				if model, ok := m.(map[string]interface{}); ok {
					if name, ok := model["name"].(string); ok {
						models = append(models, name)
					}
				}
			}
		}
	}

	// Limit to first 10 models for display
	if len(models) > 10 {
		models = models[:10]
	}

	return models
}

// isProxyError checks if an error is related to proxy configuration
func isProxyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	proxyKeywords := []string{
		"proxy",
		"connect: connection refused",
		"no such host",
		"timeout",
	}
	for _, kw := range proxyKeywords {
		if containsStr(errStr, kw) {
			return true
		}
	}
	return false
}

// tryFixProxy attempts to fix proxy-related issues
func (v *Validator) tryFixProxy(ctx context.Context, providerName string) bool {
	// This would check and adjust proxy settings
	// For now, return false to indicate no fix was applied
	if v.verbose {
		fmt.Printf("  Attempting proxy fix for %s...\n", providerName)
	}
	return false
}