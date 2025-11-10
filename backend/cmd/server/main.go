// ABOUTME: Main server entry point for the GitHub ACIP backend
// ABOUTME: Sets up Gin router, initializes database connections, and configures routes
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"acip.divkix.me/internal/auth"
	"acip.divkix.me/internal/config"
	"acip.divkix.me/internal/database"
	"acip.divkix.me/internal/generated"
	"acip.divkix.me/internal/handlers"
	"acip.divkix.me/internal/logger"
	"acip.divkix.me/internal/middleware"
	"acip.divkix.me/internal/server"
	"acip.divkix.me/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	structuredLogger := logger.NewStructuredLogger(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})

	// Log startup with correlation ID
	startupCorrelationID := "startup"
	structuredLogger.WithCorrelation(startupCorrelationID).WithFields(map[string]interface{}{
		"version": "1.0.0",
		"mode":    cfg.Server.Mode,
	}).Info("Starting ACIP backend server")

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize MongoDB with connection pooling
	mongoDB, err := database.NewMongoDBWithConfig(cfg.Database.MongoURI, cfg.Database.DBName, cfg.Database)
	if err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Fatal("Failed to connect to MongoDB")
	}
	defer func() {
		if err := mongoDB.Close(); err != nil {
			structuredLogger.WithError(startupCorrelationID, err).Error("Error closing MongoDB connection")
		}
	}()

	// Initialize Qdrant
	qdrant, err := database.NewQdrant(cfg.Database.QdrantURL, cfg.Database.QdrantAPIKey)
	if err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Fatal("Failed to connect to Qdrant")
	}
	defer func() {
		if err := qdrant.Close(); err != nil {
			structuredLogger.WithError(startupCorrelationID, err).Error("Error closing Qdrant connection")
		}
	}()

	// Test initial connections
	if err := mongoDB.Ping(); err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Warn("MongoDB connection failed")
	} else {
		structuredLogger.WithCorrelation(startupCorrelationID).Info("MongoDB connected successfully")
	}

	if err := qdrant.Ping(); err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Warn("Qdrant connection failed")
	} else {
		structuredLogger.WithCorrelation(startupCorrelationID).Info("Qdrant connected successfully")
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
	githubService, err := services.NewGitHubService(mongoDB.Database(), cfg.GitHub.ClientID, cfg.GitHub.ClientSecret, cfg.GitHub.EncryptionKey, cfg.GitHub.BatchSize, cfg.GitHub.MaxFileSize)
	if err != nil {
		log.Fatalf("Failed to initialize GitHub service: %v", err)
	}

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
	chatRAGService := services.NewChatRAGService(mongoDB.Database(), searchService, llmService, cfg)

	// Initialize repository service with embedding pipeline
	repositoryService := services.NewRepositoryService(mongoDB.Database(), githubService, userService, embeddingPipeline, cfg)

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

	// Add structured middleware in order
	router.Use(middleware.CorrelationMiddleware())
	router.Use(middleware.StructuredLoggingMiddleware(structuredLogger))
	router.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Correlation-ID"}
	router.Use(cors.New(corsConfig))

	structuredLogger.WithCorrelation(startupCorrelationID).Info("Middleware configured successfully")

	// Custom middleware configuration
	middlewareFunc := func(c *gin.Context) {
		// Apply authentication middleware only to protected routes
		path := c.FullPath()
		publicPaths := map[string]bool{
			"/health":           true,
			"/api/health":       true,
			"/api/auth/login":   true,
			"/docs":             true,
			"/docs/*any":        true,
			"/api/docs":         true,
			"/api/docs/*any":    true,
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

	// Swagger UI will fetch the latest OpenAPI spec directly from /api/openapi.yaml every time the page loads.
	router.GET("/docs/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api/openapi.yaml"), // use YAML spec served by the backend
	))

	// Convenience redirects so /docs and /api/docs open the UI without needing /index.html
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/docs/index.html")
	})

	// Serve OpenAPI specification directly
	router.GET("/api/openapi.yaml", func(c *gin.Context) {
		c.File("api/openapi.yaml")
	})

	// Serve OpenAPI specification as JSON (if needed)
	router.GET("/api/openapi.json", func(c *gin.Context) {
		c.File("api/openapi.json")
	})

	// Create HTTP server with graceful shutdown support
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		structuredLogger.WithCorrelation(startupCorrelationID).WithFields(map[string]interface{}{
			"address":          address,
			"docs_url":         fmt.Sprintf("http://%s/docs/", address),
			"openapi_url":      fmt.Sprintf("http://%s/api/openapi.yaml", address),
			"health_check_url": fmt.Sprintf("http://%s/health", address),
		}).Info("Server starting")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			structuredLogger.WithError(startupCorrelationID, err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	structuredLogger.WithCorrelation(startupCorrelationID).Info("Shutting down server...")

	// Create shutdown context with 30 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: Stop embedding pipeline first
	structuredLogger.WithCorrelation(startupCorrelationID).Info("Stopping embedding pipeline...")
	if err := embeddingPipeline.Stop(); err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Error("Error stopping embedding pipeline")
	} else {
		structuredLogger.WithCorrelation(startupCorrelationID).Info("Embedding pipeline stopped successfully")
	}

	// Step 2: Shutdown HTTP server
	structuredLogger.WithCorrelation(startupCorrelationID).Info("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		structuredLogger.WithError(startupCorrelationID, err).Error("Server forced to shutdown")
	} else {
		structuredLogger.WithCorrelation(startupCorrelationID).Info("HTTP server shutdown successfully")
	}

	// Step 3: Close database connections (deferred close will handle this)
	structuredLogger.WithCorrelation(startupCorrelationID).Info("Server exited gracefully")
}
