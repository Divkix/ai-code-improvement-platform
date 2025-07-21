// ABOUTME: Main server entry point for the GitHub analyzer backend
// ABOUTME: Sets up Gin router, initializes database connections, and configures routes
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github-analyzer/internal/auth"
	"github-analyzer/internal/config"
	"github-analyzer/internal/database"
	"github-analyzer/internal/handlers"
	"github-analyzer/internal/middleware"
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
	qdrant := database.NewQdrant(cfg.Database.QdrantURL)

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

	// Initialize services
	userService := services.NewUserService(mongoDB.Database())
	authService := auth.NewAuthService(cfg.JWT.Secret)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mongoDB, qdrant)
	authHandler := handlers.NewAuthHandler(userService, authService)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Health check routes (as defined in OpenAPI spec)
	router.GET("/health", healthHandler.GetHealth)
	router.GET("/api/health", healthHandler.GetApiHealth)

	// API group
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.LoginUser)
			auth.GET("/me", middleware.AuthMiddleware(authService), authHandler.GetCurrentUser)
		}

		// Protected API routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Future protected routes will be added here
			protected.GET("/ping", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "pong",
					"service": "github-analyzer-api",
				})
			})
		}
	}

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", address)
	log.Printf("📚 API Documentation: http://%s/docs (coming soon)", address)
	log.Printf("🏥 Health Check: http://%s/health", address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}