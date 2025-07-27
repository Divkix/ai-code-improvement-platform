// ABOUTME: Rate limiting middleware using token bucket algorithm
// ABOUTME: Protects API endpoints from abuse by limiting requests per client
package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github-analyzer/internal/logger"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	RequestsPerSecond float64 // requests per second allowed
	BurstSize         int     // burst capacity
	Enabled           bool    // whether rate limiting is enabled
}

// IPRateLimiter holds rate limiters for different IP addresses
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(config RateLimiterConfig) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}
}

// GetLimiter returns the rate limiter for the given IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(i.config.RequestsPerSecond), i.config.BurstSize)
		i.limiters[ip] = limiter
	}

	return limiter
}

// CleanupStaleEntries removes inactive rate limiters (called periodically)
func (i *IPRateLimiter) CleanupStaleEntries() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for ip, limiter := range i.limiters {
		// Remove limiters that haven't been used recently
		if limiter.Tokens() == float64(i.config.BurstSize) {
			delete(i.limiters, ip)
		}
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimiterConfig, logger *logger.StructuredLogger) gin.HandlerFunc {
	if !config.Enabled {
		// Return a no-op middleware if rate limiting is disabled
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rateLimiter := NewIPRateLimiter(config)

	return func(c *gin.Context) {
		correlationID := GetCorrelationID(c)
		clientIP := c.ClientIP()
		
		limiter := rateLimiter.GetLimiter(clientIP)
		
		if !limiter.Allow() {
			// Log rate limit exceeded
			logger.WithCorrelation(correlationID).WithFields(map[string]interface{}{
				"client_ip":    clientIP,
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"user_agent":   c.Request.UserAgent(),
			}).Warn("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
				"details": map[string]interface{}{
					"limit":            config.RequestsPerSecond,
					"window":           "1 second",
					"burst_capacity":   config.BurstSize,
					"correlation_id":   correlationID,
				},
			})
			c.Abort()
			return
		}

		// Log successful rate limit check for debugging
		if logger.GetLevel() <= logrus.DebugLevel {
			logger.WithCorrelation(correlationID).WithFields(map[string]interface{}{
				"client_ip":        clientIP,
				"tokens_remaining": limiter.Tokens(),
			}).Debug("Rate limit check passed")
		}

		c.Next()
	}
}