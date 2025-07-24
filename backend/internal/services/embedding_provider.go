package services

import "context"

// EmbeddingProvider defines the common interface for services that can
// generate embedding vectors for a list of texts. Implementations can
// be backed by a remote API (e.g., Voyage AI) or a local model served
// via an OpenAI-compatible endpoint.
//
// The returned slice must have the same length as the input slice; each
// embedding must have a consistent dimension for a given provider.
// The provider should honour the context for cancellation / deadlines.
type EmbeddingProvider interface {
    GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
    Close()
} 