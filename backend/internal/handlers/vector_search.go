// ABOUTME: Vector search handlers for semantic code search and embedding operations
// ABOUTME: Implements REST endpoints for vector search, hybrid search, and similarity queries
package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"acip.divkix.me/internal/models"
	"acip.divkix.me/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VectorSearchHandler struct {
	searchService     *services.SearchService
	embeddingService  *services.EmbeddingService
	embeddingPipeline *services.EmbeddingPipeline
}

func NewVectorSearchHandler(searchService *services.SearchService, embeddingService *services.EmbeddingService, embeddingPipeline *services.EmbeddingPipeline) *VectorSearchHandler {
	return &VectorSearchHandler{
		searchService:     searchService,
		embeddingService:  embeddingService,
		embeddingPipeline: embeddingPipeline,
	}
}

// VectorSearch performs semantic vector search across code chunks
func (vsh *VectorSearchHandler) VectorSearch(c *gin.Context) {
	var req models.VectorSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "details": err.Error()})
		return
	}

	// Parse repository ID (optional)
	var repositoryID primitive.ObjectID
	if req.RepositoryID != "" {
		var err error
		repositoryID, err = primitive.ObjectIDFromHex(req.RepositoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
			return
		}
	}

	ctx := context.Background()

	// Perform vector search
	results, err := vsh.searchService.VectorSearch(ctx, repositoryID, req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed", "details": err.Error()})
		return
	}

	// Convert SimilarityResult to SearchResult for consistent API response
	searchResults := make([]models.SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = models.SearchResult{
			CodeChunk: result.CodeChunk,
			Score:     float64(result.Score), // Convert float32 to float64
			Highlight: result.Highlight,
		}
	}

	// Create proper SearchResponse structure
	response := models.SearchResponse{
		Results: searchResults,
		Total:   int64(len(results)),
		HasMore: len(results) == req.Limit, // If we got the full limit, there might be more
		Query:   req.Query,
	}

	c.JSON(http.StatusOK, response)
}

// HybridSearch combines text and vector search for better results
func (vsh *VectorSearchHandler) HybridSearch(c *gin.Context) {
	var req models.HybridSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "details": err.Error()})
		return
	}

	// Parse repository ID (optional)
	var repositoryID primitive.ObjectID
	if req.RepositoryID != "" {
		var err error
		repositoryID, err = primitive.ObjectIDFromHex(req.RepositoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
			return
		}
	}

	ctx := context.Background()

	// Perform hybrid search
	results, err := vsh.searchService.HybridSearch(ctx, repositoryID, req.Query, req.Limit, req.VectorWeight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hybrid search failed", "details": err.Error()})
		return
	}

	// Convert SimilarityResult to SearchResult for consistent API response
	searchResults := make([]models.SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = models.SearchResult{
			CodeChunk: result.CodeChunk,
			Score:     float64(result.Score), // Convert float32 to float64
			Highlight: result.Highlight,
		}
	}

	// Create proper SearchResponse structure
	response := models.SearchResponse{
		Results: searchResults,
		Total:   int64(len(results)),
		HasMore: len(results) == req.Limit, // If we got the full limit, there might be more
		Query:   req.Query,
	}

	c.JSON(http.StatusOK, response)
}

// FindSimilar finds code chunks similar to a specific chunk
func (vsh *VectorSearchHandler) FindSimilar(c *gin.Context) {
	chunkIDStr := c.Param("chunkId")
	chunkID, err := primitive.ObjectIDFromHex(chunkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk ID"})
		return
	}

	// Parse optional limit parameter
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	ctx := context.Background()

	// Find similar chunks
	results, err := vsh.searchService.FindSimilarChunks(ctx, chunkID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find similar chunks", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"total":   len(results),
		"chunkId": chunkIDStr,
	})
}

// EmbedRepository triggers embedding processing for a repository
func (vsh *VectorSearchHandler) EmbedRepository(c *gin.Context) {
	repositoryIDStr := c.Param("id")
	repositoryID, err := primitive.ObjectIDFromHex(repositoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	ctx := context.Background()

	// Check if repository exists and get current status
	status, progress, err := vsh.embeddingService.GetEmbeddingStatus(ctx, repositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check repository status", "details": err.Error()})
		return
	}

	// If already processing, return current status
	if status == services.EmbeddingStatusProcessing {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Repository is already being processed",
			"status":  string(status),
			"progress": progress,
		})
		return
	}

	// Queue repository for embedding processing using the pipeline
	if err := vsh.embeddingPipeline.QueueRepository(ctx, repositoryID, 2); err != nil {
		// If queueing fails, fall back to direct processing (but log the error)
		log.Printf("Failed to queue repository for embedding, falling back to direct processing: %v", err)
		go func() {
			backgroundCtx := context.Background()
			if err := vsh.embeddingService.ProcessRepository(backgroundCtx, repositoryID); err != nil {
				log.Printf("Background embedding processing failed: %v", err)
			}
		}()
	} else {
		log.Printf("Successfully queued repository %s for embedding processing", repositoryID.Hex())
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Embedding processing started",
		"repositoryId": repositoryIDStr,
		"status": "processing",
	})
}

// GetEmbeddingStatus returns the current embedding status for a repository
func (vsh *VectorSearchHandler) GetEmbeddingStatus(c *gin.Context) {
	repositoryIDStr := c.Param("id")
	repositoryID, err := primitive.ObjectIDFromHex(repositoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	ctx := context.Background()

	// Get embedding status and progress
	status, progress, err := vsh.embeddingService.GetEmbeddingStatus(ctx, repositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get embedding status", "details": err.Error()})
		return
	}

	// Get processing stats for more detailed information
	stats, err := vsh.embeddingService.GetProcessingStats(ctx, repositoryID)
	if err != nil {
		// Log error but don't fail the request
		println("Failed to get processing stats:", err.Error())
		stats = nil
	}

	response := models.EmbeddingStatusResponse{
		RepositoryID: repositoryIDStr,
		Status:       string(status),
		Progress:     progress,
	}

	if stats != nil {
		response.TotalChunks = &stats.TotalChunks
		response.ProcessedChunks = &stats.ProcessedChunks
		response.FailedChunks = &stats.FailedChunks
		response.StartedAt = &stats.StartedAt
		response.CompletedAt = stats.CompletedAt
		response.EstimatedTimeRemaining = stats.EstimatedTimeRemaining
	}

	c.JSON(http.StatusOK, response)
}