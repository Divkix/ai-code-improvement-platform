// ABOUTME: Main server entry point for the GitHub analyzer backend
// ABOUTME: Sets up Gin router, initializes database connections, and configures routes
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github-analyzer/internal/auth"
	"github-analyzer/internal/config"
	"github-analyzer/internal/database"
	"github-analyzer/internal/generated"
	"github-analyzer/internal/handlers"
	"github-analyzer/internal/middleware"
	"github-analyzer/internal/server"
	"github-analyzer/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize MongoDB
	mongoDB, err := database.NewMongoDB(cfg.Database.MongoURI, cfg.Database.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoDB.Close(); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	// Initialize Qdrant
	qdrant, err := database.NewQdrant(cfg.Database.QdrantURL)
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}
	defer func() {
		if err := qdrant.Close(); err != nil {
			log.Printf("Error closing Qdrant connection: %v", err)
		}
	}()

	// Test initial connections
	if err := mongoDB.Ping(); err != nil {
		log.Printf("Warning: MongoDB connection failed: %v", err)
	} else {
		log.Println("✅ MongoDB connected successfully")
	}

	if err := qdrant.Ping(); err != nil {
		log.Printf("Warning: Qdrant connection failed: %v", err)
	} else {
		log.Println("✅ Qdrant connected successfully")
	}

	// Initialize MongoDB collections and indexes
	if err := mongoDB.InitializeCollections(); err != nil {
		log.Printf("Warning: Failed to initialize MongoDB collections: %v", err)
	}
	if err := mongoDB.EnsureIndexes(); err != nil {
		log.Printf("Warning: Failed to ensure MongoDB indexes: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(mongoDB.Database())
	authService := auth.NewAuthService(cfg.JWT.Secret)
	dashboardService := services.NewDashboardService(mongoDB.Database())
	githubService := services.NewGitHubService(mongoDB.Database(), cfg.GitHub.ClientID, cfg.GitHub.ClientSecret, cfg.GitHub.EncryptionKey)

	// Initialize universal embedding provider
	embeddingProvider := services.NewOpenAIEmbeddingService(
		cfg.AI.EmbeddingBaseURL,
		cfg.AI.EmbeddingAPIKey,
		cfg.AI.EmbeddingModel,
		60*time.Second,
	)
	log.Println("🚀 Using embedding endpoint:", cfg.AI.EmbeddingBaseURL)

	embeddingService := services.NewEmbeddingService(embeddingProvider, qdrant, mongoDB, cfg)
	embeddingPipeline := services.NewEmbeddingPipeline(embeddingService, mongoDB, cfg)
	searchService := services.NewSearchService(mongoDB.Database(), embeddingProvider, qdrant, cfg)

	// Initialize LLM service
	llmService, err := services.NewLLMService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize LLM service: %v", err)
	}
	log.Printf("🤖 LLM service initialized with model: %s", llmService.GetModel())

	// Initialize chat RAG service
	chatRAGService := services.NewChatRAGService(mongoDB.Database(), searchService, llmService)

	// Initialize repository service with embedding pipeline
	repositoryService := services.NewRepositoryService(mongoDB.Database(), githubService, userService, embeddingPipeline)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mongoDB, qdrant)
	authHandler := handlers.NewAuthHandler(userService, authService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	repositoryHandler := handlers.NewRepositoryHandler(repositoryService)
	githubHandler := handlers.NewGitHubHandler(githubService, userService)
	searchHandler := handlers.NewSearchHandler(searchService)
	vectorSearchHandler := handlers.NewVectorSearchHandler(searchService, embeddingService, embeddingPipeline)
	chatHandler := handlers.NewChatHandler(mongoDB.Database(), chatRAGService)

	// Create unified server implementing ServerInterface
	unifiedServer := server.NewServer(
		healthHandler,
		authHandler,
		dashboardHandler,
		repositoryHandler,
		githubHandler,
		searchHandler,
		vectorSearchHandler,
		chatHandler,
		embeddingPipeline,
	)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Custom middleware configuration
	middlewareFunc := func(c *gin.Context) {
		// Apply authentication middleware only to protected routes
		path := c.FullPath()
		publicPaths := map[string]bool{
			"/health":           true,
			"/api/health":       true,
			"/api/auth/login":   true,
			"/docs/*any":        true,
			"/api/openapi.yaml": true,
			"/api/openapi.json": true,
		}

		if !publicPaths[path] {
			middleware.AuthMiddleware(authService)(c)
			if c.IsAborted() {
				return
			}
		}
		c.Next()
	}

	// Register all routes using generated function with custom middleware
	options := generated.GinServerOptions{
		BaseURL: "",
		Middlewares: []generated.MiddlewareFunc{
			middlewareFunc,
		},
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, generated.Error{
				Error:   "request_error",
				Message: err.Error(),
			})
		},
	}

	generated.RegisterHandlersWithOptions(router, unifiedServer, options)

	// Initialize Qdrant collection before starting embedding pipeline
	ctx := context.Background()
	if err := embeddingService.InitializeCollection(ctx); err != nil {
		log.Printf("Warning: Failed to initialize Qdrant collection: %v", err)
	} else {
		log.Println("✅ Qdrant collection initialized")
	}

	// Start background embedding pipeline
	if err := embeddingPipeline.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start embedding pipeline: %v", err)
	} else {
		log.Println("✅ Embedding pipeline started")

		// Queue existing repositories for embedding
		if err := embeddingPipeline.QueueAllRepositories(ctx); err != nil {
			log.Printf("Warning: Failed to queue existing repositories: %v", err)
		}
	}

	// Graceful shutdown handling
	defer func() {
		if err := embeddingPipeline.Stop(); err != nil {
			log.Printf("Error stopping embedding pipeline: %v", err)
		}
	}()

	// Add Swagger UI endpoint (serves the OpenAPI spec)
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve OpenAPI specification directly
	router.GET("/api/openapi.yaml", func(c *gin.Context) {
		c.File("api/openapi.yaml")
	})

	// Serve OpenAPI specification as JSON (if needed)
	router.GET("/api/openapi.json", func(c *gin.Context) {
		c.File("api/openapi.json")
	})

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", address)
	log.Printf("📚 API Documentation (Swagger UI): http://%s/docs/", address)
	log.Printf("📋 OpenAPI Specification: http://%s/api/openapi.yaml", address)
	log.Printf("🏥 Health Check: http://%s/health", address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
