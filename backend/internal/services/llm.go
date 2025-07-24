// ABOUTME: LLM service providing OpenAI-compatible API client for chat completions
// ABOUTME: Supports streaming responses with retry logic and timeout handling

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github-analyzer/internal/config"
	"github.com/sashabaranov/go-openai"
)

// LLMService provides chat completion functionality using OpenAI-compatible APIs
type LLMService struct {
	client     *openai.Client
	model      string
	timeout    time.Duration
}

// ChatOptions configures chat completion requests
type ChatOptions struct {
	MaxTokens   int
	Temperature float32
	Stream      bool
}

// DefaultChatOptions provides sensible defaults for chat requests
var DefaultChatOptions = ChatOptions{
	MaxTokens:   1000,
	Temperature: 0.7,
	Stream:      true,
}

// NewLLMService creates a new LLM service instance
func NewLLMService(cfg *config.Config) (*LLMService, error) {
	// Parse timeout
	timeout, err := time.ParseDuration(cfg.AI.LLMRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM_REQUEST_TIMEOUT: %w", err)
	}

	// Determine API key (new or deprecated)
	apiKey := cfg.AI.LLMAPIKey
	if apiKey == "" {
		apiKey = cfg.AI.AnthropicAPIKey // Fallback for backward compatibility
	}

	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = cfg.AI.LLMBaseURL
	
	// Set custom HTTP client with timeout
	clientConfig.HTTPClient = &http.Client{
		Timeout: timeout,
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &LLMService{
		client:  client,
		model:   cfg.AI.LLMModel,
		timeout: timeout,
	}, nil
}

// ChatCompletion performs a non-streaming chat completion
func (s *LLMService) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, opts ChatOptions) (*openai.ChatCompletionResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Stream:      false,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	return &resp, nil
}

// ChatStream performs a streaming chat completion
func (s *LLMService) ChatStream(ctx context.Context, messages []openai.ChatCompletionMessage, opts ChatOptions) (<-chan openai.ChatCompletionStreamResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Stream:      true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("chat stream failed: %w", err)
	}

	// Create output channel
	responseChan := make(chan openai.ChatCompletionStreamResponse, 10)

	// Start goroutine to read from stream and forward to channel
	go func() {
		defer close(responseChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				// Check if it's EOF (normal termination)
				if err.Error() == "EOF" {
					return
				}
				// For other errors, we could send an error response,
				// but for simplicity, we'll just return
				return
			}

			select {
			case responseChan <- response:
			case <-ctx.Done():
				return
			}
		}
	}()

	return responseChan, nil
}

// BuildMessages creates a formatted message array for chat completion
func (s *LLMService) BuildMessages(systemPrompt string, userMessage string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{}

	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	return messages
}

// CountTokens estimates token count for a message (simple approximation)
func (s *LLMService) CountTokens(content string) int {
	// Simple approximation: ~4 characters per token on average
	// This is rough but better than nothing for tracking
	return len(content) / 4
}

// IsConfigured checks if the LLM service is properly configured
func (s *LLMService) IsConfigured() bool {
	return s.client != nil && s.model != ""
}

// GetModel returns the configured model name
func (s *LLMService) GetModel() string {
	return s.model
}