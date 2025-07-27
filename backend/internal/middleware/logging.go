// ABOUTME: Request/response logging middleware with structured logging support
// ABOUTME: Logs HTTP requests and responses with correlation IDs for tracing
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"acip.divkix.me/internal/logger"
)

// StructuredLoggingMiddleware logs HTTP requests and responses with correlation IDs
func StructuredLoggingMiddleware(logger *logger.StructuredLogger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Get correlation ID from context
		correlationID := ""
		if param.Keys != nil {
			if id, exists := param.Keys[CorrelationIDKey]; exists {
				if idStr, ok := id.(string); ok {
					correlationID = idStr
				}
			}
		}

		// Log the request with structured fields
		entry := logger.WithRequest(
			correlationID,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ClientIP,
		).WithFields(map[string]interface{}{
			"status_code":     param.StatusCode,
			"latency":         param.Latency.String(),
			"response_size":   param.BodySize,
			"timestamp":       param.TimeStamp.Format(time.RFC3339),
		})

		// Log at different levels based on status code
		switch {
		case param.StatusCode >= 500:
			entry.Error("HTTP request completed with server error")
		case param.StatusCode >= 400:
			entry.Warn("HTTP request completed with client error")
		case param.StatusCode >= 300:
			entry.Info("HTTP request completed with redirect")
		default:
			entry.Info("HTTP request completed successfully")
		}

		// Return empty string since we're doing structured logging
		return ""
	})
}