// ABOUTME: Dashboard handlers for statistics, activity feed, and trend data
// ABOUTME: Implements compelling demo metrics with cost savings calculations

package handlers

import (
	"net/http"
	"strconv"

	"github-analyzer/internal/middleware"
	"github-analyzer/internal/services"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard operations
type DashboardHandler struct {
	dashboardService *services.DashboardService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats handles getting dashboard statistics
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	stats, err := h.dashboardService.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get dashboard statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboardActivity handles getting recent activity items
func (h *DashboardHandler) GetDashboardActivity(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	// Parse limit parameter
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit >= 1 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}
	}

	activities, err := h.dashboardService.GetActivity(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get dashboard activity",
		})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetDashboardTrends handles getting trend data for charts
func (h *DashboardHandler) GetDashboardTrends(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	// Parse days parameter
	days := 30 // default
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			if parsedDays >= 7 && parsedDays <= 90 {
				days = parsedDays
			}
		}
	}

	trends, err := h.dashboardService.GetTrends(c.Request.Context(), userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get dashboard trends",
		})
		return
	}

	c.JSON(http.StatusOK, trends)
}