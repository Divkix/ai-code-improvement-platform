// ABOUTME: Voyage AI service for generating code embeddings
// ABOUTME: Handles batch processing, rate limiting, and error handling for vector embeddings
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type VoyageService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	rateLimiter *time.Ticker

	// simple in-memory cache to avoid re-embedding identical texts within a single
	// process lifetime. Key: text string (should be short for queries) or SHA256
	// for large chunks; Value: embedding vector. Not persisted across restarts.
	cache      map[string][]float32
	cacheMu    sync.RWMutex
}

type VoyageEmbeddingRequest struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type VoyageEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type VoyageError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`

	// Some errors are returned at the top‐level instead of under the "error" key
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}

func NewVoyageService(apiKey string) *VoyageService {
	return &VoyageService{
		apiKey:  apiKey,
		baseURL: "https://api.voyageai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(time.Second), // 60 RPM = 1 per second
		cache: make(map[string][]float32),
	}
}

func (vs *VoyageService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	const maxBatchSize = 128
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]

		// Split batch into cached vs uncached
		var toQuery []string
		cachedEmbeddings := make(map[int][]float32) // index in batch -> embedding

		vs.cacheMu.RLock()
		for idx, txt := range batch {
			if emb, ok := vs.cache[txt]; ok {
				cachedEmbeddings[idx] = emb
			} else {
				toQuery = append(toQuery, txt)
			}
		}
		vs.cacheMu.RUnlock()

		var batchEmbeddings [][]float32
		if len(toQuery) > 0 {
			// Call API with retry/backoff (3 attempts)
			const maxAttempts = 3
			var attempt int
			for {
				var err error
				batchEmbeddings, err = vs.generateBatchEmbeddings(ctx, toQuery)
				if err == nil {
					break
				}

				attempt++
				if attempt >= maxAttempts {
					return nil, fmt.Errorf("failed after %d attempts: %w", attempt, err)
				}

				backoff := time.Duration(500*attempt) * time.Millisecond
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			// Store in cache
			vs.cacheMu.Lock()
			qi := 0
			for _, txt := range toQuery {
				vs.cache[txt] = batchEmbeddings[qi]
				qi++
			}
			vs.cacheMu.Unlock()
		}

		// Merge cached + newly received embeddings preserving original order
		merged := make([][]float32, len(batch))
		qi := 0
		for idx := range batch {
			if emb, ok := cachedEmbeddings[idx]; ok {
				merged[idx] = emb
			} else {
				merged[idx] = batchEmbeddings[qi]
				qi++
			}
		}

		allEmbeddings = append(allEmbeddings, merged...)

		// Rate limiting - wait for next allowed request
		if i+maxBatchSize < len(texts) {
			select {
			case <-vs.rateLimiter.C:
				// Continue to next batch
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return allEmbeddings, nil
}

func (vs *VoyageService) generateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	requestBody := VoyageEmbeddingRequest{
		Input:     texts,
		Model:     "voyage-code-3",
		InputType: "document",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", vs.baseURL+"/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vs.apiKey)

	resp, err := vs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var voyageErr VoyageError
		if err := json.Unmarshal(body, &voyageErr); err == nil {
			// Prefer nested error fields, fall back to top-level ones, or raw body if still empty
			msg := voyageErr.Error.Message
			typ := voyageErr.Error.Type

			if msg == "" && voyageErr.Message != "" {
				msg = voyageErr.Message
			}
			if typ == "" && voyageErr.Type != "" {
				typ = voyageErr.Type
			}

			if msg != "" {
				return nil, fmt.Errorf("voyage API error: %s (type: %s)", msg, typ)
			}
		}

		// Could not decode structured error – return raw body
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response VoyageEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(response.Data))
	}

	embeddings := make([][]float32, len(response.Data))
	for _, item := range response.Data {
		if item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index %d", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

func (vs *VoyageService) Close() {
	if vs.rateLimiter != nil {
		vs.rateLimiter.Stop()
	}
}