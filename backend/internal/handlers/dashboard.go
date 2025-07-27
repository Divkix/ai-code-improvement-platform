// ABOUTME: Dashboard handlers for statistics, activity feed, and trend data
// ABOUTME: Implements compelling demo metrics with cost savings calculations

package handlers

import (
	"net/http"

	"acip.divkix.me/internal/generated"
	"acip.divkix.me/internal/middleware"
	"acip.divkix.me/internal/services"
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
		c.JSON(http.StatusUnauthorized, generated.Error{
			Error:   "unauthorized",
			Message: "User not found in context",
		})
		return
	}

	stats, err := h.dashboardService.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Failed to get dashboard statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboardActivity handles getting recent activity items
func (h *DashboardHandler) GetDashboardActivity(c *gin.Context, params generated.GetDashboardActivityParams) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, generated.Error{
			Error:   "unauthorized",
			Message: "User not found in context",
		})
		return
	}

	// Use limit from params
	limit := 10 // default
	if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 100 {
		limit = *params.Limit
	}

	activities, err := h.dashboardService.GetActivity(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Failed to get dashboard activity",
		})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetDashboardTrends handles getting trend data for charts
func (h *DashboardHandler) GetDashboardTrends(c *gin.Context, params generated.GetDashboardTrendsParams) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, generated.Error{
			Error:   "unauthorized",
			Message: "User not found in context",
		})
		return
	}

	// Use days from params
	days := 30 // default
	if params.Days != nil && *params.Days >= 7 && *params.Days <= 90 {
		days = *params.Days
	}

	trends, err := h.dashboardService.GetTrends(c.Request.Context(), userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{
			Error:   "internal_error",
			Message: "Failed to get dashboard trends",
		})
		return
	}

	c.JSON(http.StatusOK, trends)
}