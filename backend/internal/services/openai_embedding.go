package services

import (
	"context"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIEmbeddingService implements EmbeddingProvider using an OpenAI-compatible /v1/embeddings endpoint.
// It can talk to OpenAI, Voyage AI, LM Studio, Groq, Together AI, etc. — anything that follows the spec.
type OpenAIEmbeddingService struct {
    client *openai.Client
    model  string
    rateLimiter *time.Ticker
}

// NewOpenAIEmbeddingService returns an EmbeddingProvider backed by the go-openai client.
//   baseURL – e.g. "https://api.openai.com/v1" or "https://api.voyageai.com/v1" or "http://localhost:1234/v1"
//   apiKey  – token for Authorization header; can be blank for local servers that don’t require auth.
//   model   – model name to pass in the request payload.
//   timeout – HTTP timeout for the underlying client.
func NewOpenAIEmbeddingService(baseURL, apiKey, model string, timeout time.Duration) *OpenAIEmbeddingService {
    cfg := openai.DefaultConfig(apiKey)
    cfg.BaseURL = baseURL
    cfg.HTTPClient = &http.Client{Timeout: timeout}

    return &OpenAIEmbeddingService{
        client: openai.NewClientWithConfig(cfg),
        model:  model,
        rateLimiter: time.NewTicker(100 * time.Millisecond), // 10 req/s
    }
}

// GenerateEmbeddings implements EmbeddingProvider.
func (s *OpenAIEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
    if len(texts) == 0 {
        return nil, nil
    }

    // rudimentary rate limiting to avoid overloading local CPU or remote quota
    select {
    case <-s.rateLimiter.C:
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    req := openai.EmbeddingRequest{
        Model: openai.EmbeddingModel(s.model),
        Input: texts,
    }

    resp, err := s.client.CreateEmbeddings(ctx, req)
    if err != nil {
        return nil, err
    }

    // The API may return the items out of order; use the index field to rebuild order.
    out := make([][]float32, len(resp.Data))
    for _, item := range resp.Data {
        // Ensure slice bounds; skip if unexpected.
        if item.Index >= 0 && item.Index < len(out) {
            out[item.Index] = item.Embedding
        }
    }
    return out, nil
}

// Close satisfies EmbeddingProvider. There’s nothing to clean up but keeps interface parity.
func (s *OpenAIEmbeddingService) Close() {
    if s.rateLimiter != nil {
        s.rateLimiter.Stop()
    }
} 