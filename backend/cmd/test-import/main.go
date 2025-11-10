// ABOUTME: Test utility to manually trigger repository import for stuck repositories
// ABOUTME: Used for debugging import issues and manually restarting failed imports

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"acip.divkix.me/internal/config"
	"acip.divkix.me/internal/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/test-import/main.go <userID> <repoID>")
		fmt.Println("Example: go run cmd/test-import/main.go 60d5ec74b5f9c001f5e4b8a 68812d0c49ade8ebc2bfb9c9")
		os.Exit(1)
	}

	userIDStr := os.Args[1]
	repoIDStr := os.Args[2]

	// Parse ObjectIDs
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Fatalf("Invalid user ID: %v", err)
	}

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/acip.divkix.me"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	db := client.Database("acip.divkix.me")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create services
	githubService, err := services.NewGitHubService(db,
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
		os.Getenv("JWT_SECRET"),
		50,        // Default batch size
		1024*1024) // Default max file size (1MB)
	if err != nil {
		log.Fatalf("Failed to create GitHub service: %v", err)
	}

	userService := services.NewUserService(db)

	// Use nil for embedding pipeline since we only need file import
	repositoryService := services.NewRepositoryService(db, githubService, userService, nil, cfg)

	// Get repository info first
	repo, err := repositoryService.GetRepository(context.Background(), userID, repoIDStr)
	if err != nil {
		log.Fatalf("Failed to get repository: %v", err)
	}

	fmt.Printf("Repository: %s (%s)\n", repo.FullName, repo.Status)
	fmt.Printf("Progress: %d%%\n", repo.ImportProgress)

	if repo.Status != "pending" && repo.Status != "error" {
		fmt.Printf("Repository status is '%s', not pending or error. Import may already be in progress.\n", repo.Status)
		fmt.Println("Proceeding anyway...")
	}

	// Trigger the import
	fmt.Println("Triggering repository import...")
	err = repositoryService.TriggerRepositoryImport(context.Background(), userID, repoIDStr)
	if err != nil {
		log.Fatalf("Failed to trigger repository import: %v", err)
	}

	fmt.Println("Repository import triggered successfully!")
	fmt.Println("Check the Docker logs for progress:")
	fmt.Println("docker logs ai-code-improvement-platform-backend-1 --follow")
}
