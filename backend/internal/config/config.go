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
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	GitHub     GitHubConfig
	AI         AIConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	MongoURI              string
	QdrantURL             string
	DBName                string
	QdrantCollectionName  string
	VectorDimension       int
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
	VoyageAPIKey   string
	AnthropicAPIKey string
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
			MongoURI:              getEnv("MONGODB_URI", "mongodb://localhost:27017/github-analyzer"),
			QdrantURL:             getEnv("QDRANT_URL", "http://localhost:6333"),
			DBName:                getEnv("DB_NAME", "github-analyzer"),
			QdrantCollectionName:  getEnv("QDRANT_COLLECTION_NAME", "code_chunks"),
			VectorDimension:       getEnvInt("VECTOR_DIMENSION", 1536),
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
			VoyageAPIKey:   getEnv("VOYAGE_API_KEY", ""),
			AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
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
	
	// Validate Vector RAG requirements
	if c.AI.VoyageAPIKey == "" {
		return fmt.Errorf("VOYAGE_API_KEY is required for vector search functionality")
	}
	
	if c.AI.AnthropicAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required for AI chat functionality")
	}
	
	// Validate vector dimension for Voyage AI compatibility
	if c.Database.VectorDimension != 1536 {
		return fmt.Errorf("VECTOR_DIMENSION must be 1536 for Voyage AI voyage-code-3 model, got %d", c.Database.VectorDimension)
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