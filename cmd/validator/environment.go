// Environment variable validation and auto-fix
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// EnvironmentConfig represents expected environment configuration
type EnvironmentConfig struct {
	Key          string
	Required     bool
	Description  string
	ConflictWith []string // Keys that conflict with this one
	DependsOn    []string // Keys that this depends on
}

// validateEnvironment validates environment variables
func (v *Validator) validateEnvironment(ctx context.Context) []ValidationResult {
	results := []ValidationResult{}

	// LLM API Keys
	results = append(results, v.validateLLMKeys()...)

	// Embedding configuration
	results = append(results, v.validateEmbeddingConfig()...)

	// Cache configuration
	results = append(results, v.validateCacheConfig()...)

	// Proxy configuration
	results = append(results, v.validateProxyConfig()...)

	// Check for conflicts
	results = append(results, v.checkEnvConflicts()...)

	return results
}

// validateLLMKeys validates LLM API key configuration
func (v *Validator) validateLLMKeys() []ValidationResult {
	results := []ValidationResult{}

	llmKeys := []struct {
		key      string
		name     string
		required bool
	}{
		{"OPENAI_API_KEY", "OpenAI", false},
		{"OPENAI_KEY", "OpenAI (alt)", false},
		{"GOOGLE_API_KEY", "Google/Gemini", false},
		{"GEMINI_API_KEY", "Gemini (alt)", false},
		{"ANTHROPIC_API_KEY", "Anthropic/Claude", false},
		{"DEEPSEEK_API_KEY", "DeepSeek", false},
	}

	anyKeySet := false
	for _, k := range llmKeys {
		val := os.Getenv(k.key)
		if val != "" {
			anyKeySet = true
			// Validate key format (basic check)
			if len(val) < 10 {
				results = append(results, ValidationResult{
					Name:      fmt.Sprintf("API Key: %s", k.name),
					Status:    "warning",
					Message:   fmt.Sprintf("Key seems too short (%s)", k.key),
					Timestamp: time.Now(),
				})
			} else {
				// Mask the key for display
				masked := maskKey(val)
				results = append(results, ValidationResult{
					Name:      fmt.Sprintf("API Key: %s", k.name),
					Status:    "ok",
					Message:   fmt.Sprintf("Configured (%s=%s)", k.key, masked),
					Timestamp: time.Now(),
				})
			}
		}
	}

	if !anyKeySet {
		results = append(results, ValidationResult{
			Name:      "LLM API Keys",
			Status:    "warning",
			Message:   "No LLM API keys configured. Set at least one: OPENAI_API_KEY, GOOGLE_API_KEY, ANTHROPIC_API_KEY, or DEEPSEEK_API_KEY",
			Timestamp: time.Now(),
		})
	}

	return results
}

// validateEmbeddingConfig validates embedding provider configuration
func (v *Validator) validateEmbeddingConfig() []ValidationResult {
	results := []ValidationResult{}

	provider := os.Getenv("ADK_EMBED_PROVIDER")
	model := os.Getenv("ADK_EMBED_MODEL")

	if provider == "" {
		// Auto-detect based on available API keys
		detected := v.detectEmbeddingProvider()
		if detected != "" {
			result := ValidationResult{
				Name:      "Embedding Provider",
				Status:    "warning",
				Message:   fmt.Sprintf("ADK_EMBED_PROVIDER not set, auto-detected: %s", detected),
				Timestamp: time.Now(),
			}

			if v.autoFix {
				os.Setenv("ADK_EMBED_PROVIDER", detected)
				result.Status = "fixed"
				result.Message = fmt.Sprintf("Set ADK_EMBED_PROVIDER=%s", detected)
				result.FixedBy = "auto-detect from available API keys"
			}

			results = append(results, result)
		} else {
			results = append(results, ValidationResult{
				Name:      "Embedding Provider",
				Status:    "warning",
				Message:   "ADK_EMBED_PROVIDER not set and no API keys detected for auto-detection",
				Timestamp: time.Now(),
			})
		}
	} else {
		// Validate provider value
		validProviders := []string{"openai", "google", "gemini", "ollama", "claude", "anthropic", "fastembed", "deepseek"}
		isValid := false
		for _, vp := range validProviders {
			if strings.ToLower(provider) == vp {
				isValid = true
				break
			}
		}

		if isValid {
			// Check dimension compatibility with Qdrant
			expectedDim := v.getProviderDimension(provider)
			if expectedDim != v.expectedDim {
				results = append(results, ValidationResult{
					Name:      "Embedding Dimension",
					Status:    "warning",
					Message:   fmt.Sprintf("Provider %s produces %d-dim vectors, but Qdrant expects %d", provider, expectedDim, v.expectedDim),
					Details: map[string]any{
						"provider":     provider,
						"provider_dim": expectedDim,
						"qdrant_dim":   v.expectedDim,
						"suggestion":   fmt.Sprintf("Set ADK_EMBED_PROVIDER to a %d-dim compatible provider", v.expectedDim),
					},
					Timestamp: time.Now(),
				})
			} else {
				results = append(results, ValidationResult{
					Name:      "Embedding Provider",
					Status:    "ok",
					Message:   fmt.Sprintf("ADK_EMBED_PROVIDER=%s (%d dimensions)", provider, expectedDim),
					Timestamp: time.Now(),
				})
			}
		} else {
			results = append(results, ValidationResult{
				Name:      "Embedding Provider",
				Status:    "error",
				Message:   fmt.Sprintf("Invalid ADK_EMBED_PROVIDER: %s (expected: openai, google, ollama, claude, fastembed)", provider),
				Timestamp: time.Now(),
			})
		}
	}

	if model != "" {
		results = append(results, ValidationResult{
			Name:      "Embedding Model",
			Status:    "ok",
			Message:   fmt.Sprintf("ADK_EMBED_MODEL=%s", model),
			Timestamp: time.Now(),
		})
	}

	return results
}

// detectEmbeddingProvider auto-detects embedding provider from available API keys
func (v *Validator) detectEmbeddingProvider() string {
	if os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("OPENAI_KEY") != "" {
		return "openai"
	}
	if os.Getenv("GOOGLE_API_KEY") != "" || os.Getenv("GEMINI_API_KEY") != "" {
		return "google"
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "claude"
	}
	if os.Getenv("DEEPSEEK_API_KEY") != "" {
		return "deepseek"
	}
	// Check for Ollama
	if _, err := os.Stat("/usr/local/bin/ollama"); err == nil {
		return "ollama"
	}
	return ""
}

// getProviderDimension returns the default dimension for an embedding provider
func (v *Validator) getProviderDimension(provider string) int {
	dims := map[string]int{
		"openai":    1536, // text-embedding-3-small
		"google":    768,  // text-embedding-004
		"gemini":    768,
		"ollama":    768,  // nomic-embed-text
		"claude":    1024, // voyage-2
		"anthropic": 1024,
		"fastembed": 384,
		"deepseek":  768,
	}

	if dim, ok := dims[strings.ToLower(provider)]; ok {
		return dim
	}
	return 0
}

// validateCacheConfig validates cache configuration
func (v *Validator) validateCacheConfig() []ValidationResult {
	results := []ValidationResult{}

	cacheSize := os.Getenv("AGENT_LLM_CACHE_SIZE")
	cacheTTL := os.Getenv("AGENT_LLM_CACHE_TTL")
	cachePath := os.Getenv("AGENT_LLM_CACHE_PATH")

	if cacheSize != "" {
		results = append(results, ValidationResult{
			Name:      "LLM Cache Size",
			Status:    "ok",
			Message:   fmt.Sprintf("AGENT_LLM_CACHE_SIZE=%s", cacheSize),
			Timestamp: time.Now(),
		})
	}

	if cacheTTL != "" {
		results = append(results, ValidationResult{
			Name:      "LLM Cache TTL",
			Status:    "ok",
			Message:   fmt.Sprintf("AGENT_LLM_CACHE_TTL=%s seconds", cacheTTL),
			Timestamp: time.Now(),
		})
	}

	if cachePath != "" {
		results = append(results, ValidationResult{
			Name:      "LLM Cache Path",
			Status:    "ok",
			Message:   fmt.Sprintf("AGENT_LLM_CACHE_PATH=%s", cachePath),
			Timestamp: time.Now(),
		})
	}

	return results
}

// validateProxyConfig validates proxy configuration
func (v *Validator) validateProxyConfig() []ValidationResult {
	results := []ValidationResult{}

	proxyVars := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"http_proxy",
		"https_proxy",
		"NO_PROXY",
		"no_proxy",
	}

	proxySet := false
	for _, pv := range proxyVars {
		if os.Getenv(pv) != "" {
			proxySet = true
			results = append(results, ValidationResult{
				Name:      fmt.Sprintf("Proxy: %s", pv),
				Status:    "ok",
				Message:   fmt.Sprintf("%s=%s", pv, os.Getenv(pv)),
				Timestamp: time.Now(),
			})
		}
	}

	if !proxySet {
		results = append(results, ValidationResult{
			Name:      "Proxy Configuration",
			Status:    "ok",
			Message:   "No proxy configured (direct connection)",
			Timestamp: time.Now(),
		})
	}

	return results
}

// checkEnvConflicts checks for environment variable conflicts
func (v *Validator) checkEnvConflicts() []ValidationResult {
	results := []ValidationResult{}

	// Check for conflicting API key variations
	if os.Getenv("OPENAI_API_KEY") != "" && os.Getenv("OPENAI_KEY") != "" {
		results = append(results, ValidationResult{
			Name:      "Env Conflict: OpenAI Keys",
			Status:    "warning",
			Message:   "Both OPENAI_API_KEY and OPENAI_KEY are set. OPENAI_API_KEY takes precedence.",
			Timestamp: time.Now(),
		})
	}

	if os.Getenv("GOOGLE_API_KEY") != "" && os.Getenv("GEMINI_API_KEY") != "" {
		results = append(results, ValidationResult{
			Name:      "Env Conflict: Google Keys",
			Status:    "warning",
			Message:   "Both GOOGLE_API_KEY and GEMINI_API_KEY are set. GOOGLE_API_KEY takes precedence.",
			Timestamp: time.Now(),
		})
	}

	// Check for dimension mismatch between embedding provider and Qdrant
	provider := os.Getenv("ADK_EMBED_PROVIDER")
	if provider != "" {
		providerDim := v.getProviderDimension(provider)
		if providerDim > 0 && providerDim != v.expectedDim {
			result := ValidationResult{
				Name:      "Dimension Mismatch",
				Status:    "warning",
				Message:   fmt.Sprintf("Embedding provider %s (%d dim) doesn't match Qdrant expected dimension (%d)", provider, providerDim, v.expectedDim),
				Timestamp: time.Now(),
			}

			if v.autoFix {
				// Suggest compatible providers
				compatibleProviders := []string{}
				for p, d := range map[string]int{
					"deepseek": 768,
					"google":   768,
					"ollama":   768,
				} {
					if d == v.expectedDim {
						compatibleProviders = append(compatibleProviders, p)
					}
				}

				if len(compatibleProviders) > 0 {
					result.Details = map[string]any{
						"suggested_providers": compatibleProviders,
					}
				}
			}

			results = append(results, result)
		}
	}

	return results
}

// maskKey masks an API key for display
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}