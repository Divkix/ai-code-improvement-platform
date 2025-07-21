// ABOUTME: Health check handlers for monitoring service dependencies
// ABOUTME: Implements OpenAPI-defined health endpoints with database connectivity checks
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github-analyzer/internal/database"
)

type HealthHandler struct {
	mongoDB *database.MongoDB
	qdrant  *database.Qdrant
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Services  map[string]string      `json:"services"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewHealthHandler(mongoDB *database.MongoDB, qdrant *database.Qdrant) *HealthHandler {
	return &HealthHandler{
		mongoDB: mongoDB,
		qdrant:  qdrant,
	}
}

// GetHealth implements the /health endpoint
func (h *HealthHandler) GetHealth(c *gin.Context) {
	services := make(map[string]string)
	overall := "healthy"

	// Check MongoDB
	if err := h.mongoDB.Ping(); err != nil {
		services["mongodb"] = "disconnected"
		overall = "degraded"
	} else {
		services["mongodb"] = "connected"
	}

	// Check Qdrant
	if err := h.qdrant.Ping(); err != nil {
		services["qdrant"] = "disconnected"
		overall = "degraded"
	} else {
		services["qdrant"] = "connected"
	}

	// If both services are down, mark as unhealthy
	if services["mongodb"] == "disconnected" && services["qdrant"] == "disconnected" {
		overall = "unhealthy"
	}

	response := HealthResponse{
		Status:    overall,
		Services:  services,
		Timestamp: time.Now(),
	}

	statusCode := http.StatusOK
	if overall == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// GetApiHealth implements the /api/health endpoint
func (h *HealthHandler) GetApiHealth(c *gin.Context) {
	// For now, API health is the same as general health
	h.GetHealth(c)
}