package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleWorldIndices serves GET /api/market/world-indices — world market indices.
func (h *Handlers) HandleWorldIndices(c *gin.Context) {
	if h.WorldMarketService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "world market service not configured"})
		return
	}

	indices, err := h.WorldMarketService.GetWorldIndices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": indices})
}
