package market

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleExchangeRates serves GET /api/market/exchange-rates — VCB official rates.
func (h *Handlers) HandleExchangeRates(c *gin.Context) {
	if h.ExchangeRateService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "exchange rate service not configured"})
		return
	}

	rates, err := h.ExchangeRateService.GetExchangeRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rates})
}
