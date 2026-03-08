// Qdrant vector database validation
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// QdrantCollection represents a Qdrant collection
type QdrantCollection struct {
	Name string `json:"name"`
}

// QdrantCollectionsResponse is the response from listing collections
type QdrantCollectionsResponse struct {
	Result struct {
		Collections []QdrantCollection `json:"collections"`
	} `json:"result"`
	Status string `json:"status"`
}

// QdrantCollectionInfo represents detailed collection info
type QdrantCollectionInfo struct {
	Result struct {
		Status       string `json:"status"`
		PointsCount  int    `json:"points_count"`
		CollectionName string `json:"collection_name,omitempty"`
		Config       struct {
			Params struct {
				Vectors interface{} `json:"vectors"`
			} `json:"params"`
		} `json:"config"`
	} `json:"result"`
	Status string `json:"status"`
}

// VectorParams represents vector parameters
type VectorParams struct {
	Size     int    `json:"size"`
	Distance string `json:"distance"`
}

// validateQdrant validates Qdrant collections and dimensions
func (v *Validator) validateQdrant(ctx context.Context) []ValidationResult {
	results := []ValidationResult{}

	// 1. Check Qdrant connectivity
	result := v.checkQdrantConnectivity(ctx)
	results = append(results, result)
	if result.Status == "error" {
		return results
	}

	// 2. List collections
	collections, listResult := v.listQdrantCollections(ctx)
	if listResult != nil {
		results = append(results, *listResult)
		if listResult.Status == "error" {
			return results
		}
	}

	// 3. Validate each collection's dimensions
	for _, col := range collections {
		colResults := v.validateQdrantCollection(ctx, col.Name)
		results = append(results, colResults...)
	}

	// 4. Check expected collections exist
	results = append(results, v.checkExpectedQdrantCollections(ctx, collections)...)

	return results
}

// checkQdrantConnectivity checks if Qdrant is reachable
func (v *Validator) checkQdrantConnectivity(ctx context.Context) ValidationResult {
	start := time.Now()

	client := &http.Client{Timeout: 10 * time.Second}
	url := v.qdrantURL + "/collections"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ValidationResult{
			Name:      "Qdrant Connectivity",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		}
	}

	if v.qdrantAPIKey != "" {
		req.Header.Set("api-key", v.qdrantAPIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		result := ValidationResult{
			Name:      "Qdrant Connectivity",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to connect to Qdrant at %s: %v", v.qdrantURL, err),
			Timestamp: time.Now(),
		}

		// Try to auto-fix: check common issues
		if v.autoFix {
			// Check if it's a proxy issue
			if fixResult := v.tryFixQdrantProxy(ctx); fixResult != nil {
				result.Status = "fixed"
				result.Message = fmt.Sprintf("Connected after fixing proxy settings (%s)", time.Since(start).Round(time.Millisecond))
				result.FixedBy = "proxy configuration"
				return result
			}
		}

		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ValidationResult{
			Name:      "Qdrant Connectivity",
			Status:    "error",
			Message:   fmt.Sprintf("Qdrant returned status %d", resp.StatusCode),
			Timestamp: time.Now(),
		}
	}

	return ValidationResult{
		Name:      "Qdrant Connectivity",
		Status:    "ok",
		Message:   fmt.Sprintf("Connected to %s (%s)", v.qdrantURL, time.Since(start).Round(time.Millisecond)),
		Timestamp: time.Now(),
	}
}

// tryFixQdrantProxy attempts to fix proxy-related connection issues
func (v *Validator) tryFixQdrantProxy(ctx context.Context) *ValidationResult {
	// This would check and fix proxy settings
	// For now, return nil to indicate no fix was applied
	return nil
}

// listQdrantCollections lists all Qdrant collections
func (v *Validator) listQdrantCollections(ctx context.Context) ([]QdrantCollection, *ValidationResult) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := v.qdrantURL + "/collections"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, &ValidationResult{
			Name:      "List Qdrant Collections",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		}
	}

	if v.qdrantAPIKey != "" {
		req.Header.Set("api-key", v.qdrantAPIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &ValidationResult{
			Name:      "List Qdrant Collections",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to list collections: %v", err),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	var response QdrantCollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &ValidationResult{
			Name:      "List Qdrant Collections",
			Status:    "error",
			Message:   fmt.Sprintf("Failed to decode response: %v", err),
			Timestamp: time.Now(),
		}
	}

	return response.Result.Collections, nil
}

// validateQdrantCollection validates a single Qdrant collection
func (v *Validator) validateQdrantCollection(ctx context.Context, name string) []ValidationResult {
	results := []ValidationResult{}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/collections/%s", v.qdrantURL, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		results = append(results, ValidationResult{
			Name:      fmt.Sprintf("Collection %s", name),
			Status:    "error",
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Timestamp: time.Now(),
		})
		return results
	}

	if v.qdrantAPIKey != "" {
		req.Header.Set("api-key", v.qdrantAPIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		results = append(results, ValidationResult{
			Name:      fmt.Sprintf("Collection %s", name),
			Status:    "error",
			Message:   fmt.Sprintf("Failed to get collection info: %v", err),
			Timestamp: time.Now(),
		})
		return results
	}
	defer resp.Body.Close()

	var info QdrantCollectionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		results = append(results, ValidationResult{
			Name:      fmt.Sprintf("Collection %s", name),
			Status:    "error",
			Message:   fmt.Sprintf("Failed to decode collection info: %v", err),
			Timestamp: time.Now(),
		})
		return results
	}

	// Extract vector dimensions
	vectorParams := extractVectorParams(info.Result.Config.Params.Vectors)
	
	for _, vp := range vectorParams {
		result := v.validateVectorDimensions(name, vp, info.Result.PointsCount)
		results = append(results, result)
	}

	return results
}

// extractVectorParams extracts vector parameters from the vectors config
func extractVectorParams(vectors interface{}) []VectorParams {
	params := []VectorParams{}

	switch v := vectors.(type) {
	case map[string]interface{}:
		// Single vector config
		if size, ok := v["size"].(float64); ok {
			distance := "Cosine"
			if d, ok := v["distance"].(string); ok {
				distance = d
			}
			params = append(params, VectorParams{
				Size:     int(size),
				Distance: distance,
			})
		}
		// Named vectors
		for key, val := range v {
			if key == "size" || key == "distance" {
				continue
			}
			if nested, ok := val.(map[string]interface{}); ok {
				if size, ok := nested["size"].(float64); ok {
					distance := "Cosine"
					if d, ok := nested["distance"].(string); ok {
						distance = d
					}
					params = append(params, VectorParams{
						Size:     int(size),
						Distance: distance,
					})
				}
			}
		}
	}

	return params
}

// validateVectorDimensions validates vector dimensions against expected
func (v *Validator) validateVectorDimensions(collection string, params VectorParams, pointsCount int) ValidationResult {
	name := fmt.Sprintf("Collection %s (dim=%d)", collection, params.Size)

	if params.Size == v.expectedDim {
		return ValidationResult{
			Name:    name,
			Status:  "ok",
			Message: fmt.Sprintf("Dimensions match expected %d (%s, %d points)", v.expectedDim, params.Distance, pointsCount),
			Details: map[string]any{
				"collection":    collection,
				"dimension":     params.Size,
				"expected_dim":  v.expectedDim,
				"distance":      params.Distance,
				"points_count":  pointsCount,
			},
			Timestamp: time.Now(),
		}
	}

	result := ValidationResult{
		Name:    name,
		Status:  "warning",
		Message: fmt.Sprintf("Dimension mismatch: got %d, expected %d (%s, %d points)", params.Size, v.expectedDim, params.Distance, pointsCount),
		Details: map[string]any{
			"collection":    collection,
			"dimension":     params.Size,
			"expected_dim":  v.expectedDim,
			"distance":      params.Distance,
			"points_count":  pointsCount,
		},
		Timestamp: time.Now(),
	}

	// Check embedding model compatibility
	compatibleModels := getCompatibleModels(params.Size)
	result.Details.(map[string]any)["compatible_models"] = compatibleModels

	return result
}

// checkExpectedQdrantCollections checks for expected Qdrant collections
func (v *Validator) checkExpectedQdrantCollections(ctx context.Context, collections []QdrantCollection) []ValidationResult {
	results := []ValidationResult{}

	// Expected collections for go-agent
	expectedCollections := []string{"adk_memories"}

	existingMap := make(map[string]bool)
	for _, c := range collections {
		existingMap[c.Name] = true
	}

	for _, expected := range expectedCollections {
		if existingMap[expected] {
			results = append(results, ValidationResult{
				Name:      fmt.Sprintf("Expected collection: %s", expected),
				Status:    "ok",
				Message:   "Collection exists",
				Timestamp: time.Now(),
			})
		} else {
			result := ValidationResult{
				Name:      fmt.Sprintf("Expected collection: %s", expected),
				Status:    "warning",
				Message:   "Collection not found",
				Timestamp: time.Now(),
			}

			// Try to auto-fix: create collection
			if v.autoFix {
				if v.createQdrantCollection(ctx, expected, v.expectedDim) {
					result.Status = "fixed"
					result.Message = fmt.Sprintf("Created collection with %d dimensions", v.expectedDim)
					result.FixedBy = "auto-create collection"
				}
			}

			results = append(results, result)
		}
	}

	return results
}

// createQdrantCollection creates a new Qdrant collection
func (v *Validator) createQdrantCollection(ctx context.Context, name string, dimension int) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/collections/%s", v.qdrantURL, name)

	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dimension,
			"distance": "Cosine",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	if v.qdrantAPIKey != "" {
		req.Header.Set("api-key", v.qdrantAPIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// getCompatibleModels returns embedding models compatible with a dimension
func getCompatibleModels(dimension int) []string {
	modelMap := map[int][]string{
		384:  {"sentence-transformers/all-MiniLM-L6-v2"},
		768: {
			"DeepSeek embeddings",
			"Google text-embedding-004",
			"Ollama nomic-embed-text",
			"sentence-transformers/all-mpnet-base-v2",
		},
		1024: {
			"Voyage voyage-2",
			"Ollama mxbai-embed-large",
		},
		1536: {
			"OpenAI text-embedding-ada-002",
			"OpenAI text-embedding-3-small",
		},
		3072: {
			"OpenAI text-embedding-3-large",
		},
	}

	if models, ok := modelMap[dimension]; ok {
		return models
	}
	return []string{"Unknown compatible models"}
}