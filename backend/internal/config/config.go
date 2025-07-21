// ABOUTME: Configuration management for the GitHub analyzer backend
// ABOUTME: Loads environment variables and provides typed configuration structs
package config

import (
	"fmt"
	"os"

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
	MongoURI   string
	QdrantURL  string
	DBName     string
}

type JWTConfig struct {
	Secret string
}

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
}

type AIConfig struct {
	VoyageAPIKey   string
	AnthropicAPIKey string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			MongoURI:  getEnv("MONGODB_URI", "mongodb://localhost:27017/github-analyzer"),
			QdrantURL: getEnv("QDRANT_URL", "http://localhost:6333"),
			DBName:    getEnv("DB_NAME", "github-analyzer"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		GitHub: GitHubConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
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
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}