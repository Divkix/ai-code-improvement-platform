// ABOUTME: Unit tests for LLM service functionality
// ABOUTME: Tests client creation, message building, and configuration validation

package services

import (
	"testing"
	"time"

	"github-analyzer/internal/config"

	"github.com/sashabaranov/go-openai"
)

func TestNewLLMService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		cfg := &config.Config{
			AI: config.AIConfig{
				LLMBaseURL:        "https://api.openai.com/v1",
				LLMModel:          "gpt-4o-mini",
				LLMAPIKey:         "test-key",
				LLMRequestTimeout: "30s",
			},
		}

		service, err := NewLLMService(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if service == nil {
			t.Fatal("Expected service to be created")
		}

		if service.model != "gpt-4o-mini" {
			t.Errorf("Expected model 'gpt-4o-mini', got %v", service.model)
		}

		if service.timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", service.timeout)
		}
	})

	t.Run("fails with invalid timeout", func(t *testing.T) {
		cfg := &config.Config{
			AI: config.AIConfig{
				LLMBaseURL:        "https://api.openai.com/v1",
				LLMModel:          "gpt-4o-mini",
				LLMAPIKey:         "test-key",
				LLMRequestTimeout: "invalid-timeout",
			},
		}

		service, err := NewLLMService(cfg)
		if err == nil {
			t.Fatal("Expected error for invalid timeout")
		}

		if service != nil {
			t.Fatal("Expected service to be nil on error")
		}
	})
}

func TestLLMService_BuildMessages(t *testing.T) {
	cfg := &config.Config{
		AI: config.AIConfig{
			LLMBaseURL:        "https://api.openai.com/v1",
			LLMModel:          "gpt-4o-mini",
			LLMAPIKey:         "test-key",
			LLMRequestTimeout: "30s",
		},
	}

	service, err := NewLLMService(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	t.Run("builds messages with system prompt", func(t *testing.T) {
		systemPrompt := "You are a helpful assistant."
		userMessage := "Hello, how are you?"

		messages := service.BuildMessages(systemPrompt, userMessage)

		if len(messages) != 2 {
			t.Errorf("Expected 2 messages, got %v", len(messages))
		}

		if messages[0].Role != openai.ChatMessageRoleSystem {
			t.Errorf("Expected system role, got %v", messages[0].Role)
		}

		if messages[0].Content != systemPrompt {
			t.Errorf("Expected system content %v, got %v", systemPrompt, messages[0].Content)
		}

		if messages[1].Role != openai.ChatMessageRoleUser {
			t.Errorf("Expected user role, got %v", messages[1].Role)
		}

		if messages[1].Content != userMessage {
			t.Errorf("Expected user content %v, got %v", userMessage, messages[1].Content)
		}
	})

	t.Run("builds messages without system prompt", func(t *testing.T) {
		userMessage := "Hello, how are you?"

		messages := service.BuildMessages("", userMessage)

		if len(messages) != 1 {
			t.Errorf("Expected 1 message, got %v", len(messages))
		}

		if messages[0].Role != openai.ChatMessageRoleUser {
			t.Errorf("Expected user role, got %v", messages[0].Role)
		}

		if messages[0].Content != userMessage {
			t.Errorf("Expected user content %v, got %v", userMessage, messages[0].Content)
		}
	})
}

func TestLLMService_CountTokens(t *testing.T) {
	cfg := &config.Config{
		AI: config.AIConfig{
			LLMBaseURL:        "https://api.openai.com/v1",
			LLMModel:          "gpt-4o-mini",
			LLMAPIKey:         "test-key",
			LLMRequestTimeout: "30s",
		},
	}

	service, err := NewLLMService(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty string",
			content:  "",
			expected: 0,
		},
		{
			name:     "short message",
			content:  "test",
			expected: 1,
		},
		{
			name:     "longer message",
			content:  "This is a test message",
			expected: 5, // 22 chars / 4 = 5.5 -> 5
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := service.CountTokens(tc.content)
			if tokens != tc.expected {
				t.Errorf("Expected %v tokens, got %v", tc.expected, tokens)
			}
		})
	}
}

func TestLLMService_IsConfigured(t *testing.T) {
	t.Run("returns true for properly configured service", func(t *testing.T) {
		cfg := &config.Config{
			AI: config.AIConfig{
				LLMBaseURL:        "https://api.openai.com/v1",
				LLMModel:          "gpt-4o-mini",
				LLMAPIKey:         "test-key",
				LLMRequestTimeout: "30s",
			},
		}

		service, err := NewLLMService(cfg)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		if !service.IsConfigured() {
			t.Error("Expected service to be configured")
		}
	})
}

func TestLLMService_GetModel(t *testing.T) {
	cfg := &config.Config{
		AI: config.AIConfig{
			LLMBaseURL:        "https://api.openai.com/v1",
			LLMModel:          "gpt-4o-mini",
			LLMAPIKey:         "test-key",
			LLMRequestTimeout: "30s",
		},
	}

	service, err := NewLLMService(cfg)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	model := service.GetModel()
	if model != "gpt-4o-mini" {
		t.Errorf("Expected model 'gpt-4o-mini', got %v", model)
	}
}

func TestDefaultChatOptions(t *testing.T) {
	opts := DefaultChatOptions

	if opts.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens 1000, got %v", opts.MaxTokens)
	}

	if opts.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %v", opts.Temperature)
	}

	if !opts.Stream {
		t.Error("Expected Stream to be true")
	}
}