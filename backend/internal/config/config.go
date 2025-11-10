// ABOUTME: Configuration management for the ACIP backend
// ABOUTME: Loads environment variables and provides typed configuration structs
package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	Timeouts       TimeoutConfig
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

type TimeoutConfig struct {
	HTTPClient      time.Duration // HTTP client timeout for general requests
	GitHubAPI       time.Duration // GitHub API request timeout
	MongoOperation  time.Duration // MongoDB operation timeout
	QdrantSearch    time.Duration // Qdrant vector search timeout
	QdrantUpsert    time.Duration // Qdrant vector upsert timeout
	EmbeddingAPI    time.Duration // Embedding API request timeout
	LLMStreaming    time.Duration // LLM streaming response timeout
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
		Timeouts: TimeoutConfig{
			HTTPClient:      getEnvDuration("TIMEOUT_HTTP_CLIENT", 10*time.Second),
			GitHubAPI:       getEnvDuration("TIMEOUT_GITHUB_API", 10*time.Second),
			MongoOperation:  getEnvDuration("TIMEOUT_MONGO_OPERATION", 30*time.Second),
			QdrantSearch:    getEnvDuration("TIMEOUT_QDRANT_SEARCH", 5*time.Second),
			QdrantUpsert:    getEnvDuration("TIMEOUT_QDRANT_UPSERT", 30*time.Second),
			EmbeddingAPI:    getEnvDuration("TIMEOUT_EMBEDDING_API", 60*time.Second),
			LLMStreaming:    getEnvDuration("TIMEOUT_LLM_STREAMING", 5*time.Minute),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	// JWT Secret validation
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters for security (got %d)", len(c.JWT.Secret))
	}

	// Warn about development secrets
	if isDevelopmentSecret(c.JWT.Secret) {
		log.Printf("WARNING: JWT_SECRET appears to be a development/example value - use a strong random secret in production")
	}

	// GitHub encryption key validation
	if c.GitHub.EncryptionKey != "" {
		keyLen := len(c.GitHub.EncryptionKey)
		if keyLen != 16 && keyLen != 24 && keyLen != 32 {
			return fmt.Errorf("GITHUB_ENCRYPTION_KEY must be exactly 16, 24, or 32 bytes for AES-128/192/256 (got %d)", keyLen)
		}
		if isDevelopmentSecret(c.GitHub.EncryptionKey) {
			log.Printf("WARNING: GITHUB_ENCRYPTION_KEY appears to be a development/example value - use a strong random key in production")
		}
	}

	// MongoDB URI validation
	if c.Database.MongoURI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	if !isValidMongoURI(c.Database.MongoURI) {
		return fmt.Errorf("MONGODB_URI format is invalid - must start with mongodb:// or mongodb+srv://")
	}

	// API endpoint URL validation
	if c.AI.EmbeddingBaseURL == "" {
		return fmt.Errorf("EMBEDDING_BASE_URL is required")
	}
	if !isValidURL(c.AI.EmbeddingBaseURL) {
		return fmt.Errorf("EMBEDDING_BASE_URL is not a valid URL: %s", c.AI.EmbeddingBaseURL)
	}

	if c.AI.LLMBaseURL == "" {
		return fmt.Errorf("LLM_BASE_URL is required")
	}
	if !isValidURL(c.AI.LLMBaseURL) {
		return fmt.Errorf("LLM_BASE_URL is not a valid URL: %s", c.AI.LLMBaseURL)
	}

	if c.Database.QdrantURL == "" {
		return fmt.Errorf("QDRANT_URL is required")
	}
	if !isValidURL(c.Database.QdrantURL) {
		return fmt.Errorf("QDRANT_URL is not a valid URL: %s", c.Database.QdrantURL)
	}

	// Embedding model validation
	if c.AI.EmbeddingModel == "" {
		return fmt.Errorf("EMBEDDING_MODEL is required")
	}

	// LLM configuration validation
	if c.AI.LLMAPIKey == "" {
		return fmt.Errorf("LLM_API_KEY is required for AI chat functionality")
	}

	// Vector dimension validation
	validDimensions := []int{256, 512, 768, 1024, 2048}
	isValidDim := false
	for _, dim := range validDimensions {
		if c.Database.VectorDimension == dim {
			isValidDim = true
			break
		}
	}
	if !isValidDim {
		return fmt.Errorf("VECTOR_DIMENSION must be one of %v, got %d", validDimensions, c.Database.VectorDimension)
	}

	// Timeout validation - ensure all timeouts are positive
	if c.Timeouts.HTTPClient <= 0 {
		return fmt.Errorf("TIMEOUT_HTTP_CLIENT must be positive, got %v", c.Timeouts.HTTPClient)
	}
	if c.Timeouts.GitHubAPI <= 0 {
		return fmt.Errorf("TIMEOUT_GITHUB_API must be positive, got %v", c.Timeouts.GitHubAPI)
	}
	if c.Timeouts.MongoOperation <= 0 {
		return fmt.Errorf("TIMEOUT_MONGO_OPERATION must be positive, got %v", c.Timeouts.MongoOperation)
	}
	if c.Timeouts.QdrantSearch <= 0 {
		return fmt.Errorf("TIMEOUT_QDRANT_SEARCH must be positive, got %v", c.Timeouts.QdrantSearch)
	}
	if c.Timeouts.QdrantUpsert <= 0 {
		return fmt.Errorf("TIMEOUT_QDRANT_UPSERT must be positive, got %v", c.Timeouts.QdrantUpsert)
	}
	if c.Timeouts.EmbeddingAPI <= 0 {
		return fmt.Errorf("TIMEOUT_EMBEDDING_API must be positive, got %v", c.Timeouts.EmbeddingAPI)
	}
	if c.Timeouts.LLMStreaming <= 0 {
		return fmt.Errorf("TIMEOUT_LLM_STREAMING must be positive, got %v", c.Timeouts.LLMStreaming)
	}

	return nil
}

// isValidMongoURI checks if a MongoDB URI has valid format
func isValidMongoURI(uri string) bool {
	return strings.HasPrefix(uri, "mongodb://") || strings.HasPrefix(uri, "mongodb+srv://")
}

// isValidURL validates that a string is a valid HTTP/HTTPS URL
func isValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// isDevelopmentSecret detects common development/example secret patterns
func isDevelopmentSecret(secret string) bool {
	lowerSecret := strings.ToLower(secret)
	developmentPatterns := []string{
		"secret",
		"password",
		"changeme",
		"example",
		"test",
		"dev",
		"demo",
		"default",
		"12345",
		"asdf",
		"qwerty",
	}

	for _, pattern := range developmentPatterns {
		if strings.Contains(lowerSecret, pattern) {
			return true
		}
	}

	// Check for repeated characters (e.g., "aaaaaaaaaa")
	if len(secret) > 0 {
		firstChar := secret[0]
		allSame := true
		for _, char := range secret {
			if char != rune(firstChar) {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	}

	return false
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
