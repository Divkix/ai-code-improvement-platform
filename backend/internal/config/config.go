// ABOUTME: Configuration management for the ACIP backend
// ABOUTME: Loads environment variables and provides typed configuration structs
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	CodeProcessing CodeProcessingConfig
	JWT            JWTConfig
	GitHub         GitHubConfig
	AI             AIConfig
	Logging        LoggingConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	MongoURI               string
	QdrantURL              string
	QdrantAPIKey           string
	DBName                 string
	QdrantCollectionName   string
	VectorDimension        int
	EnableQdrantRepoFilter bool // if true, attach repositoryId payload filter in Qdrant queries

	// MongoDB Connection Pooling Configuration
	MaxPoolSize            uint64        // Maximum number of connections in the pool
	MinPoolSize            uint64        // Minimum number of connections in the pool
	MaxIdleTime            time.Duration // Maximum time a connection can be idle
	ConnectTimeout         time.Duration // Timeout for establishing connections
	ServerSelectionTimeout time.Duration // Timeout for server selection
}

type CodeProcessingConfig struct {
	ChunkSize           int // Lines per chunk
	OverlapSize         int // Lines to overlap between chunks
	EmbeddingBatchSize  int // Number of chunks to process per embedding batch
	EmbeddingWorkersNum int // Number of concurrent workers for embedding pipeline
}

type JWTConfig struct {
	Secret string
}

type GitHubConfig struct {
	ClientID      string
	ClientSecret  string
	EncryptionKey string
	BatchSize     int // Number of files to process per batch
	MaxFileSize   int // Maximum file size in bytes to process
}

type AIConfig struct {
	// LLM Configuration (OpenAI-compatible API)
	LLMBaseURL        string
	LLMModel          string
	LLMAPIKey         string
	LLMRequestTimeout string
	LLMContextLength  int

	// Universal embedding endpoint (OpenAI-compatible)
	EmbeddingBaseURL string

	// Provider agnostic model name
	// Generic embedding model name shared across providers (set via EMBEDDING_MODEL)
	EmbeddingModel string

	// Universal embedding API key (set via EMBEDDING_API_KEY). May be empty for local servers.
	EmbeddingAPIKey string

	// Chat and retrieval configuration
	MaxPromptLength   int     // Soft cut-off where a prompt is truncated
	ChatContextChunks int     // How many code chunks are pulled from vector DB for a single question
	ChatVectorWeight  float64 // Weight given to ANN similarity vs. BM-25 text search in hybrid retrieval
}

type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, file
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
			MongoURI:               getEnv("MONGODB_URI", "mongodb://localhost:27017/acip.divkix.me"),
			QdrantURL:              getEnv("QDRANT_URL", "http://localhost:6334"),
			QdrantAPIKey:           getEnv("QDRANT_API_KEY", ""),
			DBName:                 getEnv("DB_NAME", "acip.divkix.me"),
			QdrantCollectionName:   getEnv("QDRANT_COLLECTION_NAME", "codechunks"),
			VectorDimension:        getEnvInt("VECTOR_DIMENSION", 1024),
			EnableQdrantRepoFilter: getEnv("ENABLE_QDRANT_REPO_FILTER", "true") != "false",

			// MongoDB Connection Pooling
			MaxPoolSize:            uint64(getEnvInt("MONGODB_MAX_POOL_SIZE", 100)),
			MinPoolSize:            uint64(getEnvInt("MONGODB_MIN_POOL_SIZE", 5)),
			MaxIdleTime:            getEnvDuration("MONGODB_MAX_IDLE_TIME", 30*time.Second),
			ConnectTimeout:         getEnvDuration("MONGODB_CONNECT_TIMEOUT", 10*time.Second),
			ServerSelectionTimeout: getEnvDuration("MONGODB_SERVER_SELECTION_TIMEOUT", 5*time.Second),
		},
		CodeProcessing: CodeProcessingConfig{
			ChunkSize:           getEnvInt("CHUNK_SIZE", 30),
			OverlapSize:         getEnvInt("CHUNK_OVERLAP_SIZE", 10),
			EmbeddingBatchSize:  getEnvInt("EMBEDDING_BATCH_SIZE", 50),
			EmbeddingWorkersNum: getEnvInt("EMBEDDING_WORKERS_NUM", 3),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		GitHub: GitHubConfig{
			ClientID:      getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret:  getEnv("GITHUB_CLIENT_SECRET", ""),
			EncryptionKey: getEnv("GITHUB_ENCRYPTION_KEY", ""),
			BatchSize:     getEnvInt("GITHUB_BATCH_SIZE", 50),
			MaxFileSize:   getEnvInt("GITHUB_MAX_FILE_SIZE", 1024*1024),
		},
		AI: AIConfig{
			LLMBaseURL:        getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			LLMModel:          getEnv("LLM_MODEL", "gpt-4o-mini"),
			LLMAPIKey:         getEnv("LLM_API_KEY", ""),
			LLMRequestTimeout: getEnv("LLM_REQUEST_TIMEOUT", "30s"),
			LLMContextLength:  getEnvInt("LLM_CONTEXT_LENGTH", 32000),
			EmbeddingBaseURL:  getEnv("EMBEDDING_BASE_URL", "https://api.openai.com/v1"),
			EmbeddingModel:    getEnv("EMBEDDING_MODEL", "text-embedding-nomic-embed-text-v1.5"),
			EmbeddingAPIKey:   getEnv("EMBEDDING_API_KEY", ""),
			MaxPromptLength:   getEnvInt("MAX_PROMPT_LENGTH", 12000),
			ChatContextChunks: getEnvInt("CHAT_CONTEXT_CHUNKS", 8),
			ChatVectorWeight:  getEnvFloat("CHAT_VECTOR_WEIGHT", 0.7),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
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

	// Optional sanity check: common Voyage models expect 256/512/768/1024/2048 dims.
	if c.Database.VectorDimension != 256 && c.Database.VectorDimension != 512 && c.Database.VectorDimension != 768 && c.Database.VectorDimension != 1024 && c.Database.VectorDimension != 2048 {
		return fmt.Errorf("VECTOR_DIMENSION must be one of 256, 512, 768, 1024, or 2048, got %d", c.Database.VectorDimension)
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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
