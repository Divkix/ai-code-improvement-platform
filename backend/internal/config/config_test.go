// ABOUTME: Unit tests for configuration management system
// ABOUTME: Tests environment variable parsing, defaults, and validation rules

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithDefaults(t *testing.T) {
	t.Parallel()

	// Clear all environment variables that might affect the test
	envVars := []string{
		"PORT", "HOST", "GIN_MODE", "MONGODB_URI", "QDRANT_URL", "DB_NAME",
		"QDRANT_COLLECTION_NAME", "VECTOR_DIMENSION", "JWT_SECRET",
		"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "GITHUB_ENCRYPTION_KEY",
		"LLM_BASE_URL", "LLM_MODEL",
		"LLM_API_KEY", "LLM_REQUEST_TIMEOUT",
		"EMBEDDING_BASE_URL", "EMBEDDING_MODEL", "EMBEDDING_API_KEY",
		"ENABLE_QDRANT_REPO_FILTER", "LLM_CONTEXT_LENGTH", "CHUNK_SIZE", "CHUNK_OVERLAP_SIZE",
	}

	// Store original values
	originalValues := make(map[string]string)
	for _, env := range envVars {
		originalValues[env] = os.Getenv(env)
		require.NoError(t, os.Unsetenv(env))
	}

	// Set minimum required values
	require.NoError(t, os.Setenv("JWT_SECRET", "test-secret"))
	require.NoError(t, os.Setenv("LLM_API_KEY", "test-llm-key"))

	defer func() {
		// Restore original values
		for env, value := range originalValues {
			if value == "" {
				require.NoError(t, os.Unsetenv(env))
			} else {
				require.NoError(t, os.Setenv(env, value))
			}
		}
	}()

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test default values
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "debug", config.Server.Mode)

	assert.Equal(t, "mongodb://localhost:27017/acip.divkix.me", config.Database.MongoURI)
	assert.Equal(t, "http://localhost:6334", config.Database.QdrantURL)
	assert.Equal(t, "acip.divkix.me", config.Database.DBName)
	assert.Equal(t, "codechunks", config.Database.QdrantCollectionName)
	assert.Equal(t, 1024, config.Database.VectorDimension)
	assert.True(t, config.Database.EnableQdrantRepoFilter)

	assert.Equal(t, 30, config.CodeProcessing.ChunkSize)
	assert.Equal(t, 10, config.CodeProcessing.OverlapSize)

	assert.Equal(t, "test-secret", config.JWT.Secret)

	assert.Equal(t, "", config.GitHub.ClientID)
	assert.Equal(t, "", config.GitHub.ClientSecret)
	assert.Equal(t, "", config.GitHub.EncryptionKey)

	assert.Equal(t, "https://api.openai.com/v1", config.AI.LLMBaseURL)
	assert.Equal(t, "gpt-4o-mini", config.AI.LLMModel)
	assert.Equal(t, "test-llm-key", config.AI.LLMAPIKey)
	assert.Equal(t, "30s", config.AI.LLMRequestTimeout)
	assert.Equal(t, 2048, config.AI.LLMContextLength)
	assert.Equal(t, "https://api.openai.com/v1", config.AI.EmbeddingBaseURL)
	assert.Equal(t, "text-embedding-nomic-embed-text-v1.5", config.AI.EmbeddingModel)
	assert.Equal(t, "", config.AI.EmbeddingAPIKey)
}

func TestLoad_WithCustomValues(t *testing.T) {
	t.Parallel()

	// Set custom environment variables
	envVars := map[string]string{
		"PORT":                   "3000",
		"HOST":                   "127.0.0.1",
		"GIN_MODE":               "release",
		"MONGODB_URI":            "mongodb://custom:27017/test",
		"QDRANT_URL":             "http://custom:6334",
		"DB_NAME":                "test-db",
		"QDRANT_COLLECTION_NAME": "test-chunks",
		"VECTOR_DIMENSION":       "512",
		"JWT_SECRET":             "custom-secret",
		"GITHUB_CLIENT_ID":       "github-id",
		"GITHUB_CLIENT_SECRET":   "github-secret",
		"GITHUB_ENCRYPTION_KEY":  "encryption-key",
		"LLM_BASE_URL":           "https://custom-llm.com/v1",
		"LLM_MODEL":              "custom-model",
		"LLM_API_KEY":            "custom-llm-key",
		"LLM_REQUEST_TIMEOUT":    "60s",
		"EMBEDDING_BASE_URL":     "https://custom-embeddings.com/v1",
		"EMBEDDING_MODEL":        "custom-model",
		"EMBEDDING_API_KEY":      "emb-key",
	}

	// Store original values and set test values
	originalValues := make(map[string]string)
	for key, value := range envVars {
		originalValues[key] = os.Getenv(key)
		require.NoError(t, os.Setenv(key, value))
	}

	defer func() {
		// Restore original values
		for key, originalValue := range originalValues {
			if originalValue == "" {
				_ = os.Unsetenv(key) // Ignore error in cleanup
			} else {
				_ = os.Setenv(key, originalValue) // Ignore error in cleanup
			}
		}
	}()

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify custom values
	assert.Equal(t, "3000", config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, "release", config.Server.Mode)

	assert.Equal(t, "mongodb://custom:27017/test", config.Database.MongoURI)
	assert.Equal(t, "http://custom:6334", config.Database.QdrantURL)
	assert.Equal(t, "test-db", config.Database.DBName)
	assert.Equal(t, "test-chunks", config.Database.QdrantCollectionName)
	assert.Equal(t, 512, config.Database.VectorDimension)
	assert.True(t, config.Database.EnableQdrantRepoFilter)

	assert.Equal(t, "custom-secret", config.JWT.Secret)

	assert.Equal(t, "github-id", config.GitHub.ClientID)
	assert.Equal(t, "github-secret", config.GitHub.ClientSecret)
	assert.Equal(t, "encryption-key", config.GitHub.EncryptionKey)

	assert.Equal(t, "https://custom-llm.com/v1", config.AI.LLMBaseURL)
	assert.Equal(t, "custom-model", config.AI.LLMModel)
	assert.Equal(t, "custom-llm-key", config.AI.LLMAPIKey)
	assert.Equal(t, "60s", config.AI.LLMRequestTimeout)
	assert.Equal(t, "https://custom-embeddings.com/v1", config.AI.EmbeddingBaseURL)
	assert.Equal(t, "custom-model", config.AI.EmbeddingModel)
	assert.Equal(t, "emb-key", config.AI.EmbeddingAPIKey)
}

func TestValidation_JWTSecretRequired(t *testing.T) {
	// Don't run in parallel due to environment variable manipulation

	// Store all relevant environment variables
	originalJWT := os.Getenv("JWT_SECRET")
	originalLLM := os.Getenv("LLM_API_KEY")
	originalDimension := os.Getenv("VECTOR_DIMENSION")
	originalLocalURL := os.Getenv("LOCAL_EMBEDDING_URL")

	defer func() {
		restoreEnv("JWT_SECRET", originalJWT)
		restoreEnv("LLM_API_KEY", originalLLM)
		restoreEnv("VECTOR_DIMENSION", originalDimension)
		restoreEnv("LOCAL_EMBEDDING_URL", originalLocalURL)
	}()

	// Clear JWT_SECRET but set other required values to ensure JWT validation is tested
	require.NoError(t, os.Unsetenv("JWT_SECRET"))
	require.NoError(t, os.Setenv("LLM_API_KEY", "test-llm-key"))
	require.NoError(t, os.Setenv("VECTOR_DIMENSION", "1024"))

	config, err := Load()
	assert.Error(t, err)
	assert.Nil(t, config)
	if err != nil {
		assert.Contains(t, err.Error(), "JWT_SECRET is required")
	}
}

func TestValidation_LLMAPIKey(t *testing.T) {
	// Don't run in parallel due to environment variable manipulation

	tests := []struct {
		name          string
		llmKey        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "LLM key provided",
			llmKey:      "llm-key",
			expectError: false,
		},
		{
			name:          "No API keys provided",
			expectError:   true,
			errorContains: "LLM_API_KEY is required for AI chat functionality",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Don't run subtests in parallel due to environment variable manipulation

			// Store original values
			originalLLM := os.Getenv("LLM_API_KEY")
			originalJWT := os.Getenv("JWT_SECRET")

			defer func() {
				restoreEnv("LLM_API_KEY", originalLLM)
				restoreEnv("JWT_SECRET", originalJWT)
			}()

			// Set required values
			require.NoError(t, os.Setenv("JWT_SECRET", "test-secret"))
			require.NoError(t, os.Setenv("LLM_API_KEY", tt.llmKey))

			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestValidation_VectorDimension(t *testing.T) {
	// Don't run in parallel due to environment variable manipulation

	tests := []struct {
		name        string
		dimension   string
		expectError bool
	}{
		{
			name:        "valid dimension 256",
			dimension:   "256",
			expectError: false,
		},
		{
			name:        "valid dimension 512",
			dimension:   "512",
			expectError: false,
		},
		{
			name:        "valid dimension 1024",
			dimension:   "1024",
			expectError: false,
		},
		{
			name:        "valid dimension 2048",
			dimension:   "2048",
			expectError: false,
		},
		{
			name:        "invalid dimension 128",
			dimension:   "128",
			expectError: true,
		},
		{
			name:        "invalid dimension 4096",
			dimension:   "4096",
			expectError: true,
		},
		{
			name:        "valid dimension 768",
			dimension:   "768",
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Don't run subtests in parallel due to environment variable manipulation

			// Store original values
			originalJWT := os.Getenv("JWT_SECRET")
			originalLLM := os.Getenv("LLM_API_KEY")
			originalDimension := os.Getenv("VECTOR_DIMENSION")
			originalLocalURL := os.Getenv("LOCAL_EMBEDDING_URL")

			defer func() {
				restoreEnv("JWT_SECRET", originalJWT)
				restoreEnv("LLM_API_KEY", originalLLM)
				restoreEnv("VECTOR_DIMENSION", originalDimension)
				restoreEnv("LOCAL_EMBEDDING_URL", originalLocalURL)
			}()

			// Set required values
			require.NoError(t, os.Setenv("JWT_SECRET", "test-secret"))
			require.NoError(t, os.Setenv("LLM_API_KEY", "test-llm-key"))
			require.NoError(t, os.Setenv("VECTOR_DIMENSION", tt.dimension))

			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				if err != nil {
					assert.Contains(t, err.Error(), "VECTOR_DIMENSION must be one of")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "invalid integer",
			envValue:     "not-a-number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "negative integer",
			envValue:     "-42",
			defaultValue: 10,
			expected:     -42,
		},
		{
			name:         "zero",
			envValue:     "0",
			defaultValue: 10,
			expected:     0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testEnvKey := "TEST_INT_VAR"
			originalValue := os.Getenv(testEnvKey)

			defer restoreEnv(testEnvKey, originalValue)

			if tt.envValue == "" {
				require.NoError(t, os.Unsetenv(testEnvKey))
			} else {
				require.NoError(t, os.Setenv(testEnvKey, tt.envValue))
			}

			result := getEnvInt(testEnvKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to restore environment variable
func restoreEnv(key, value string) {
	if value == "" {
		_ = os.Unsetenv(key) // Ignore error in cleanup
	} else {
		_ = os.Setenv(key, value) // Ignore error in cleanup
	}
}