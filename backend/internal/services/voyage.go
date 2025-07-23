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
	"time"
)

type VoyageService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	rateLimiter *time.Ticker
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
}

func NewVoyageService(apiKey string) *VoyageService {
	return &VoyageService{
		apiKey:  apiKey,
		baseURL: "https://api.voyageai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(time.Second), // 60 RPM = 1 per second
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
		batchEmbeddings, err := vs.generateBatchEmbeddings(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings for batch %d-%d: %w", i, end-1, err)
		}

		allEmbeddings = append(allEmbeddings, batchEmbeddings...)

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
		if err := json.Unmarshal(body, &voyageErr); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("voyage API error: %s (type: %s)", voyageErr.Error.Message, voyageErr.Error.Type)
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