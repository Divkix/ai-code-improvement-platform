// ABOUTME: Authentication handlers for user registration, login, and profile management
// ABOUTME: Implements JWT-based auth with OpenAPI generated types and validation

package handlers

import (
	"net/http"

	"github-analyzer/internal/auth"
	"github-analyzer/internal/middleware"
	"github-analyzer/internal/models"
	"github-analyzer/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	userService *services.UserService
	authService *auth.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService, authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
	}
}


// LoginUser handles user login
func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req models.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_credentials",
				"message": "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Authentication failed",
		})
		return
	}

	// Check password
	if err := h.authService.CheckPassword(user.Password, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_credentials",
			"message": "Invalid email or password",
		})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate token",
		})
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusOK, response)
}

// GetCurrentUser handles getting current user information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get user information",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}