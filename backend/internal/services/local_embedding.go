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

// LocalEmbeddingService talks to a self-hosted embedding model (e.g. LM-Studio
// running nomic-embed-text-v2) that exposes an OpenAI-compatible `/v1/embeddings`
// endpoint.
type LocalEmbeddingService struct {
    model       string
    baseURL     string // e.g. http://localhost:1234
    httpClient  *http.Client
    rateLimiter *time.Ticker // rudimentary RPM limiting (optional)
}

func NewLocalEmbeddingService(baseURL, model string) *LocalEmbeddingService {
    return &LocalEmbeddingService{
        model:   model,
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 60 * time.Second,
        },
        // LM-Studio is local; we can afford higher RPS, but keep a small limit to
        // avoid CPU overload.
        rateLimiter: time.NewTicker(100 * time.Millisecond), // 10 req/s
    }
}

type openAIEmbeddingRequest struct {
    Model string        `json:"model"`
    Input interface{}   `json:"input"` // string | []string
}

type openAIEmbeddingResponse struct {
    Data []struct {
        Embedding []float32 `json:"embedding"`
        Index     int       `json:"index"`
    } `json:"data"`
}

func (ls *LocalEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
    if len(texts) == 0 {
        return nil, fmt.Errorf("no texts provided")
    }

    reqBody := openAIEmbeddingRequest{
        Model: ls.model,
        Input: texts,
    }
    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // basic rate-limit
    select {
    case <-ls.rateLimiter.C:
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    req, err := http.NewRequestWithContext(ctx, "POST", ls.baseURL+"/v1/embeddings", bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := ls.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("failed to close response body: %v", err)
        }
    }()

    if resp.StatusCode != http.StatusOK {
        raw, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("local embedding API error %d: %s", resp.StatusCode, string(raw))
    }

    var parsed openAIEmbeddingResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    if len(parsed.Data) != len(texts) {
        return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(texts), len(parsed.Data))
    }

    out := make([][]float32, len(texts))
    for _, item := range parsed.Data {
        if item.Index >= len(out) {
            return nil, fmt.Errorf("invalid embedding index %d", item.Index)
        }
        out[item.Index] = item.Embedding
    }

    return out, nil
}

func (ls *LocalEmbeddingService) Close() {
    if ls.rateLimiter != nil {
        ls.rateLimiter.Stop()
    }
} 