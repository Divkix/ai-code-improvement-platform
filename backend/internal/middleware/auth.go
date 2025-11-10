// ABOUTME: JWT authentication middleware for protected routes
// ABOUTME: Validates tokens and provides user context to handlers

package middleware

import (
	"log"
	"net/http"
	"strings"

	"acip.divkix.me/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from gin context with type safety
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	// Safe type assertion with validation
	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("SECURITY WARNING: user_id type mismatch - expected string, got %T for IP %s", userID, c.ClientIP())
		return "", false
	}

	// Validate ObjectID format (24 hex characters)
	if len(userIDStr) != 24 {
		log.Printf("SECURITY WARNING: invalid user_id length - expected 24 chars, got %d from IP %s", len(userIDStr), c.ClientIP())
		return "", false
	}

	// Validate hex characters
	for _, char := range userIDStr {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			log.Printf("SECURITY WARNING: invalid user_id format - contains non-hex character from IP %s", c.ClientIP())
			return "", false
		}
	}

	return userIDStr, true
}

// GetUserEmailFromContext extracts user email from gin context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	return email.(string), true
}
