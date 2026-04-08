package market

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandlePriceBoard serves GET /api/market/price-board?symbols=VNM,FPT
func (h *Handlers) HandlePriceBoard(c *gin.Context) {
	raw := c.Query("symbols")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbols query parameter is required"})
		return
	}

	symbols := strings.Split(strings.ToUpper(raw), ",")
	boards, err := h.PriceService.GetPriceBoard(c.Request.Context(), symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": boards})
}
