package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleMarketStatus serves GET /api/market/status — trading session status.
func (h *Handlers) HandleMarketStatus(c *gin.Context) {
	if h.TradingHoursService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "trading hours service not configured"})
		return
	}

	market := c.DefaultQuery("market", "")
	if market != "" {
		status := h.TradingHoursService.GetStatus(c.Request.Context(), market)
		c.JSON(http.StatusOK, gin.H{"data": status})
		return
	}

	statuses := h.TradingHoursService.GetAllStatuses(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}
