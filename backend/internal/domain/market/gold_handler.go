package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleGoldPrices serves GET /api/market/gold — SJC + BTMC gold prices.
func (h *Handlers) HandleGoldPrices(c *gin.Context) {
	if h.GoldPriceService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gold price service not configured"})
		return
	}

	prices, err := h.GoldPriceService.GetGoldPrices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prices})
}
