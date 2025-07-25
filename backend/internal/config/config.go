// ABOUTME: Configuration management for the GitHub analyzer backend
// ABOUTME: Loads environment variables and provides typed configuration structs
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	GitHub   GitHubConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	MongoURI             string
	QdrantURL            string
	DBName               string
	QdrantCollectionName string
	VectorDimension      int
	EnableQdrantRepoFilter bool // if true, attach repositoryId payload filter in Qdrant queries
}

type JWTConfig struct {
	Secret string
}

type GitHubConfig struct {
	ClientID      string
	ClientSecret  string
	EncryptionKey string
}

type AIConfig struct {
	// LLM Configuration (OpenAI-compatible API)
	LLMBaseURL        string
	LLMModel          string
	LLMAPIKey         string
	LLMRequestTimeout string

	// Universal embedding endpoint (OpenAI-compatible)
	EmbeddingBaseURL string

	// Provider agnostic model name
	// Generic embedding model name shared across providers (set via EMBEDDING_MODEL)
	EmbeddingModel string

	// Universal embedding API key (set via EMBEDDING_API_KEY). May be empty for local servers.
	EmbeddingAPIKey string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load() // Ignore error as .env file is optional

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			MongoURI:             getEnv("MONGODB_URI", "mongodb://localhost:27017/github-analyzer"),
			QdrantURL:            getEnv("QDRANT_URL", "http://localhost:6334"),
			DBName:               getEnv("DB_NAME", "github-analyzer"),
			QdrantCollectionName: getEnv("QDRANT_COLLECTION_NAME", "codechunks"),
			VectorDimension:      getEnvInt("VECTOR_DIMENSION", 1024),
			EnableQdrantRepoFilter: getEnv("ENABLE_QDRANT_REPO_FILTER", "true") != "false",
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		GitHub: GitHubConfig{
			ClientID:      getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
			EncryptionKey: getEnv("GITHUB_ENCRYPTION_KEY", ""),
		},
		AI: AIConfig{
			LLMBaseURL:        getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			LLMModel:          getEnv("LLM_MODEL", "gpt-4o-mini"),
			LLMAPIKey:         getEnv("LLM_API_KEY", ""),
			LLMRequestTimeout: getEnv("LLM_REQUEST_TIMEOUT", "30s"),
			EmbeddingBaseURL:  getEnv("EMBEDDING_BASE_URL", "https://api.openai.com/v1"),
			EmbeddingModel:    getEnv("EMBEDDING_MODEL", "text-embedding-nomic-embed-text-v1.5"),
			EmbeddingAPIKey:   getEnv("EMBEDDING_API_KEY", ""),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	// Basic validation – ensure base URL and model are set
	if c.AI.EmbeddingBaseURL == "" {
		return fmt.Errorf("EMBEDDING_BASE_URL is required")
	}
	if c.AI.EmbeddingModel == "" {
		return fmt.Errorf("EMBEDDING_MODEL is required")
	}

	// Validate LLM configuration
	if c.AI.LLMAPIKey == "" {
		return fmt.Errorf("LLM_API_KEY is required for AI chat functionality")
	}

	// Optional sanity check: common Voyage models expect 256/512/1024/2048 dims.
	if c.Database.VectorDimension != 256 && c.Database.VectorDimension != 512 && c.Database.VectorDimension != 1024 && c.Database.VectorDimension != 2048 {
		return fmt.Errorf("VECTOR_DIMENSION must be one of 256, 512, 1024, or 2048, got %d", c.Database.VectorDimension)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
