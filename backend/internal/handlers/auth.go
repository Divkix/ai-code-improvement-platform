// ABOUTME: Authentication handlers for user registration, login, and profile management
// ABOUTME: Implements JWT-based auth with OpenAPI generated types and validation

package handlers

import (
	"net/http"

	"acip.divkix.me/internal/auth"
	"acip.divkix.me/internal/generated"
	"acip.divkix.me/internal/middleware"
	"acip.divkix.me/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	userService *services.UserService
	authService *auth.AuthService
	validate    *validator.Validate
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService, authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
		validate:    validator.New(),
	}
}

// LoginRequest represents validated login credentials
type LoginRequest struct {
	Email    string `validate:"required,email,max=255"`
	Password string `validate:"required,min=8,max=72"`
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(c *gin.Context) {
	// TODO: Implement rate limiting to prevent brute force attacks
	// Consider using middleware or a dedicated rate limiting service
	// Example: 5 failed attempts per IP per 15 minutes

	var req generated.LoginUserJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{
			Error:   "invalid_request",
			Message: "Invalid request format",
		})
		return
	}

	// Validate input with structured validation
	loginReq := LoginRequest{
		Email:    string(req.Email),
		Password: req.Password,
	}

	if err := h.validate.Struct(loginReq); err != nil {
		// Use generic error message to prevent information leakage
		c.JSON(http.StatusBadRequest, generated.Error{
			Error:   "invalid_request",
			Message: "Invalid email or password format",
		})
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(c.Request.Context(), loginReq.Email)
	if err != nil {
		if err == services.ErrUserNotFound {
			// Use same error as password check to prevent user enumeration
			c.JSON(http.StatusUnauthorized, generated.Error{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Authentication failed",
		})
		return
	}

	// Check password
	if err := h.authService.CheckPassword(user.Password, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, generated.Error{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	response := generated.AuthResponse{
		Token: token,
		User:  user.ToGeneratedUser(),
	}

	c.JSON(http.StatusOK, response)
}

// GetCurrentUser handles getting current user information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, generated.Error{
			Error:   "unauthorized",
			Message: "User not found in context",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, generated.Error{
				Error:   "unauthorized",
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Failed to get user information",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToGeneratedUser())
}
