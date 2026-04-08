package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealth serves GET /api/health — basic health check.
func (h *Handlers) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleHealthz serves GET /api/healthz — liveness probe (always 200 if process is running).
func (h *Handlers) HandleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// HandleReadyz serves GET /api/readyz — readiness probe checking Go + Python AI Service.
func (h *Handlers) HandleReadyz(c *gin.Context) {
	status := gin.H{
		"go": "ready",
	}

	// Check Python AI Service if gRPC client is available.
	if h.GRPCClient != nil {
		aiHealth := h.GRPCClient.HealthCheck(c.Request.Context())
		status["aiService"] = gin.H{
			"available": aiHealth.Available,
			"mode":      aiHealth.Mode,
		}
		if !aiHealth.Available {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "degraded",
				"details": status,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"details": status,
	})
}
