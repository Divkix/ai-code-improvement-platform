// ABOUTME: Correlation ID middleware for request tracking and tracing
// ABOUTME: Generates unique request IDs to trace requests across the entire system
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CorrelationIDKey = "correlation_id"
const CorrelationIDHeader = "X-Correlation-ID"

// CorrelationMiddleware adds a unique correlation ID to each request for tracing
func CorrelationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if correlation ID is already provided in request header
		correlationID := c.GetHeader(CorrelationIDHeader)
		if correlationID == "" {
			// Generate a new UUID if no correlation ID provided
			correlationID = uuid.New().String()
		}

		// Store correlation ID in gin context for use throughout request lifecycle
		c.Set(CorrelationIDKey, correlationID)
		
		// Add correlation ID to response headers for client tracking
		c.Header(CorrelationIDHeader, correlationID)
		
		// Continue to next middleware/handler
		c.Next()
	})
}

// GetCorrelationID retrieves the correlation ID from gin context
func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get(CorrelationIDKey); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}