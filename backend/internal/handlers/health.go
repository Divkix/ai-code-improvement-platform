// ABOUTME: Health check handlers for monitoring service dependencies
// ABOUTME: Implements OpenAPI-defined health endpoints with database connectivity checks
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github-analyzer/internal/database"
	"github-analyzer/internal/generated"
)

type HealthHandler struct {
	mongoDB *database.MongoDB
	qdrant  *database.Qdrant
}

// HealthHandler implements health check methods from generated.ServerInterface

func NewHealthHandler(mongoDB *database.MongoDB, qdrant *database.Qdrant) *HealthHandler {
	return &HealthHandler{
		mongoDB: mongoDB,
		qdrant:  qdrant,
	}
}

// GetHealth implements the /health endpoint
func (h *HealthHandler) GetHealth(c *gin.Context) {
	var mongodbStatus generated.HealthCheckServicesMongodb
	var qdrantStatus generated.HealthCheckServicesQdrant
	overall := generated.Healthy

	// Check MongoDB
	if err := h.mongoDB.Ping(); err != nil {
		mongodbStatus = generated.HealthCheckServicesMongodbDisconnected
		overall = generated.Degraded
	} else {
		mongodbStatus = generated.HealthCheckServicesMongodbConnected
	}

	// Check Qdrant
	if err := h.qdrant.Ping(); err != nil {
		qdrantStatus = generated.HealthCheckServicesQdrantDisconnected
		overall = generated.Degraded
	} else {
		qdrantStatus = generated.HealthCheckServicesQdrantConnected
	}

	// If both services are down, mark as unhealthy
	if mongodbStatus == generated.HealthCheckServicesMongodbDisconnected && 
		qdrantStatus == generated.HealthCheckServicesQdrantDisconnected {
		overall = generated.Unhealthy
	}

	timestamp := time.Now()
	response := generated.HealthCheck{
		Status: overall,
		Services: struct {
			Mongodb *generated.HealthCheckServicesMongodb `json:"mongodb,omitempty"`
			Qdrant  *generated.HealthCheckServicesQdrant  `json:"qdrant,omitempty"`
		}{
			Mongodb: &mongodbStatus,
			Qdrant:  &qdrantStatus,
		},
		Timestamp: &timestamp,
	}

	statusCode := http.StatusOK
	if overall == generated.Unhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// GetApiHealth implements the /api/health endpoint
func (h *HealthHandler) GetApiHealth(c *gin.Context) {
	// For now, API health is the same as general health
	h.GetHealth(c)
}