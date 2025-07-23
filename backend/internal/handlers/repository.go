// ABOUTME: Repository handlers for CRUD operations with user ownership validation
// ABOUTME: Implements repository management endpoints with proper error handling and OpenAPI compliance

package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github-analyzer/internal/middleware"
	"github-analyzer/internal/models"
	"github-analyzer/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RepositoryHandler handles repository operations
type RepositoryHandler struct {
	repositoryService *services.RepositoryService
}

// NewRepositoryHandler creates a new repository handler
func NewRepositoryHandler(repositoryService *services.RepositoryService) *RepositoryHandler {
	return &RepositoryHandler{
		repositoryService: repositoryService,
	}
}

// GetRepositories handles getting user repositories with pagination and filtering
func (h *RepositoryHandler) GetRepositories(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	statusFilter := c.Query("status")

	// Validate limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	repositories, err := h.repositoryService.GetRepositories(c.Request.Context(), objectID, limit, offset, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve repositories",
		})
		return
	}

	c.JSON(http.StatusOK, repositories)
}

// CreateRepository handles creating a new repository
func (h *RepositoryHandler) CreateRepository(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	var req models.CreateRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	repository, err := h.repositoryService.CreateRepository(c.Request.Context(), objectID, req)
	if err != nil {
		if err == services.ErrRepositoryExists {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "repository_exists",
				"message": "Repository already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to create repository",
		})
		return
	}

	c.JSON(http.StatusCreated, repository)
}

// GetRepository handles getting a specific repository by ID
func (h *RepositoryHandler) GetRepository(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	repoID := c.Param("id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Repository ID is required",
		})
		return
	}

	repository, err := h.repositoryService.GetRepository(c.Request.Context(), objectID, repoID)
	if err != nil {
		if err == services.ErrRepositoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "repository_not_found",
				"message": "Repository not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve repository",
		})
		return
	}

	c.JSON(http.StatusOK, repository)
}

// UpdateRepository handles updating repository information
func (h *RepositoryHandler) UpdateRepository(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	repoID := c.Param("id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Repository ID is required",
		})
		return
	}

	var req models.UpdateRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	repository, err := h.repositoryService.UpdateRepository(c.Request.Context(), objectID, repoID, req)
	if err != nil {
		if err == services.ErrRepositoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "repository_not_found",
				"message": "Repository not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to update repository",
		})
		return
	}

	c.JSON(http.StatusOK, repository)
}

// DeleteRepository handles deleting a repository
func (h *RepositoryHandler) DeleteRepository(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	repoID := c.Param("id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Repository ID is required",
		})
		return
	}

	err = h.repositoryService.DeleteRepository(c.Request.Context(), objectID, repoID)
	if err != nil {
		if err == services.ErrRepositoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "repository_not_found",
				"message": "Repository not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to delete repository",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetRepositoryStats handles getting repository statistics
func (h *RepositoryHandler) GetRepositoryStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	repoID := c.Param("id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Repository ID is required",
		})
		return
	}

	stats, err := h.repositoryService.GetRepositoryStats(c.Request.Context(), objectID, repoID)
	if err != nil {
		if err == services.ErrRepositoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "repository_not_found",
				"message": "Repository not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve repository statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Additional utility methods for repository status and progress updates
// These are typically called internally by import/processing services

// UpdateRepositoryStatus handles updating repository status (internal use)
func (h *RepositoryHandler) UpdateRepositoryStatus(ctx context.Context, userID primitive.ObjectID, repoID string, status string) error {
	return h.repositoryService.UpdateRepositoryStatus(ctx, userID, repoID, status)
}

// UpdateRepositoryProgress handles updating repository import progress (internal use)
func (h *RepositoryHandler) UpdateRepositoryProgress(ctx context.Context, userID primitive.ObjectID, repoID string, progress int) error {
	return h.repositoryService.UpdateRepositoryProgress(ctx, userID, repoID, progress)
}

// TriggerRepositoryImport handles manually triggering repository import for stuck repositories
func (h *RepositoryHandler) TriggerRepositoryImport(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	repoID := c.Param("id")
	if repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Repository ID is required",
		})
		return
	}

	// Get repository to check current status
	repo, err := h.repositoryService.GetRepository(c.Request.Context(), objectID, repoID)
	if err != nil {
		if err == services.ErrRepositoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "repository_not_found",
				"message": "Repository not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve repository",
		})
		return
	}

	// Only allow triggering import for pending or error status repositories
	if repo.Status != models.StatusPending && repo.Status != models.StatusError {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_status",
			"message": "Repository import can only be triggered for repositories with 'pending' or 'error' status",
			"current_status": repo.Status,
		})
		return
	}

	// Trigger the import process
	err = h.repositoryService.TriggerRepositoryImport(c.Request.Context(), objectID, repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "import_failed",
			"message": "Failed to trigger repository import: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Repository import triggered successfully",
		"repository_id": repoID,
		"status": "importing",
	})
}