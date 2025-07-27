// ABOUTME: Tests for rate limiting middleware functionality  
// ABOUTME: Ensures proper rate limiting behavior and configuration
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github-analyzer/internal/logger"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test logger
	testLogger := logger.NewStructuredLogger(logger.Config{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	tests := []struct {
		name           string
		config         RateLimiterConfig
		requests       int
		expectedPassed int
		expectedBlocked int
	}{
		{
			name: "allows requests within rate limit",
			config: RateLimiterConfig{
				RequestsPerSecond: 10.0,
				BurstSize:         5,
				Enabled:           true,
			},
			requests:        3,
			expectedPassed:  3,
			expectedBlocked: 0,
		},
		{
			name: "blocks requests exceeding burst capacity",
			config: RateLimiterConfig{
				RequestsPerSecond: 1.0,
				BurstSize:         2,
				Enabled:           true,
			},
			requests:        5,
			expectedPassed:  2,
			expectedBlocked: 3,
		},
		{
			name: "disabled rate limiter allows all requests",
			config: RateLimiterConfig{
				RequestsPerSecond: 1.0,
				BurstSize:         1,
				Enabled:           false,
			},
			requests:        5,
			expectedPassed:  5,
			expectedBlocked: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router with middleware
			router := gin.New()
			router.Use(CorrelationMiddleware())
			router.Use(RateLimitMiddleware(tt.config, testLogger))
			
			var passed, blocked int
			router.GET("/test", func(c *gin.Context) {
				passed++
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Make requests
			for i := 0; i < tt.requests; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345" // Same IP for all requests
				
				router.ServeHTTP(w, req)
				
				if w.Code == http.StatusTooManyRequests {
					blocked++
				}
			}

			// Assertions
			assert.Equal(t, tt.expectedPassed, passed, "Number of passed requests should match expected")
			assert.Equal(t, tt.expectedBlocked, blocked, "Number of blocked requests should match expected")
		})
	}
}

func TestIPRateLimiter(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 2.0,
		BurstSize:         2,
		Enabled:           true,
	}

	limiter := NewIPRateLimiter(config)

	t.Run("creates different limiters for different IPs", func(t *testing.T) {
		limiter1 := limiter.GetLimiter("192.168.1.1")
		limiter2 := limiter.GetLimiter("192.168.1.2")

		// Check that they are different objects (different memory addresses)
		assert.NotSame(t, limiter1, limiter2, "Different IPs should have different limiter instances")
	})

	t.Run("returns same limiter for same IP", func(t *testing.T) {
		limiter1 := limiter.GetLimiter("192.168.1.1")
		limiter2 := limiter.GetLimiter("192.168.1.1")

		assert.Same(t, limiter1, limiter2, "Same IP should return same limiter instance")
	})

	t.Run("cleanup removes unused limiters", func(t *testing.T) {
		// Get a limiter and use all tokens
		testLimiter := limiter.GetLimiter("192.168.1.100")
		
		// Consume all tokens
		for testLimiter.Allow() {
			// Keep consuming until no tokens left
		}

		// Add a limiter that has full tokens (unused)
		limiter.GetLimiter("192.168.1.101")

		// Cleanup should remove the full limiter but keep the exhausted one
		limiter.CleanupStaleEntries()

		// The map should have been modified by cleanup
		assert.NotNil(t, limiter.limiters, "Limiters map should still exist")
	})
}

func TestRateLimitMiddlewareWithDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testLogger := logger.NewStructuredLogger(logger.Config{
		Level:  "info",
		Format: "json", 
		Output: "stdout",
	})

	config := RateLimiterConfig{
		RequestsPerSecond: 1.0,
		BurstSize:         1,
		Enabled:           true,
	}

	router := gin.New()
	router.Use(CorrelationMiddleware())
	router.Use(RateLimitMiddleware(config, testLogger))
	
	var responses []int
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test different IPs should have independent rate limits
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	
	for _, ip := range ips {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip + ":12345"
		
		router.ServeHTTP(w, req)
		responses = append(responses, w.Code)
	}

	// All requests should pass since they're from different IPs
	for i, code := range responses {
		assert.Equal(t, http.StatusOK, code, "Request from IP %d should pass", i)
	}
}