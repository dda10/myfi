package market

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandlePriceDepth serves GET /api/market/price-depth?symbol=VNM
func (h *Handlers) HandlePriceDepth(c *gin.Context) {
	symbol := strings.ToUpper(c.Query("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol query parameter is required"})
		return
	}

	depth, err := h.PriceService.GetPriceDepth(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": depth})
}
