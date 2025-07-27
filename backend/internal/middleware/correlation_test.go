// ABOUTME: Tests for correlation ID middleware functionality
// ABOUTME: Ensures proper correlation ID generation and propagation
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                   string
		existingCorrelationID  string
		expectedHeaderPresent  bool
		expectedContextPresent bool
	}{
		{
			name:                   "generates new correlation ID when none provided",
			existingCorrelationID:  "",
			expectedHeaderPresent:  true,
			expectedContextPresent: true,
		},
		{
			name:                   "uses existing correlation ID when provided",
			existingCorrelationID:  "test-correlation-id-123",
			expectedHeaderPresent:  true,
			expectedContextPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create request with optional correlation ID header
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingCorrelationID != "" {
				req.Header.Set(CorrelationIDHeader, tt.existingCorrelationID)
			}
			c.Request = req

			// Test handler that checks correlation ID
			var correlationIDFromContext string
			testHandler := func(c *gin.Context) {
				correlationIDFromContext = GetCorrelationID(c)
				c.Status(http.StatusOK)
			}

			// Execute middleware
			CorrelationMiddleware()(c)
			testHandler(c)

			// Assertions
			if tt.expectedHeaderPresent {
				assert.NotEmpty(t, w.Header().Get(CorrelationIDHeader), "Response should have correlation ID header")
			}

			if tt.expectedContextPresent {
				assert.NotEmpty(t, correlationIDFromContext, "Context should contain correlation ID")
			}

			// If correlation ID was provided, it should be preserved
			if tt.existingCorrelationID != "" {
				assert.Equal(t, tt.existingCorrelationID, w.Header().Get(CorrelationIDHeader))
				assert.Equal(t, tt.existingCorrelationID, correlationIDFromContext)
			}
		})
	}
}

func TestGetCorrelationID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedResult string
	}{
		{
			name: "returns correlation ID when present",
			setupContext: func(c *gin.Context) {
				c.Set(CorrelationIDKey, "test-id-123")
			},
			expectedResult: "test-id-123",
		},
		{
			name: "returns empty string when not present",
			setupContext: func(c *gin.Context) {
				// Don't set correlation ID
			},
			expectedResult: "",
		},
		{
			name: "returns empty string when wrong type in context",
			setupContext: func(c *gin.Context) {
				c.Set(CorrelationIDKey, 12345) // Wrong type
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			// Execute
			result := GetCorrelationID(c)

			// Assert
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}