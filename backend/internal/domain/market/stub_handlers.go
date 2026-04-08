package market

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleMarketIndex serves GET /api/market/index?name=VNINDEX
// Returns index data (value, change, changePercent) for a given market index.
func (h *Handlers) HandleMarketIndex(c *gin.Context) {
	name := strings.ToUpper(c.Query("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name query parameter is required"})
		return
	}

	// Try to get real index data from MarketDataService
	indices, err := h.MarketDataService.GetMarketIndices(c.Request.Context())
	if err == nil && len(indices) > 0 {
		for _, idx := range indices {
			if strings.EqualFold(idx.Name, name) {
				c.JSON(http.StatusOK, gin.H{
					"name":          idx.Name,
					"value":         idx.Close,
					"change":        idx.Change,
					"changePercent": idx.ChangePct,
				})
				return
			}
		}
	}

	// Return empty placeholder if not found
	c.JSON(http.StatusOK, gin.H{
		"name":          name,
		"value":         0,
		"change":        0,
		"changePercent": 0,
	})
}

// HandleHotTopics serves GET /api/market/hot-topics
// Returns trending market topics/themes.
func (h *Handlers) HandleHotTopics(c *gin.Context) {
	// Placeholder — will be populated by AI service later
	c.JSON(http.StatusOK, gin.H{"data": []any{}})
}

// HandleMarketRatios serves GET /api/market/ratios/:symbol
// Returns financial ratios for a stock.
func (h *Handlers) HandleMarketRatios(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}
	// Placeholder — will be populated when financial data service is ready
	c.JSON(http.StatusOK, gin.H{"symbol": symbol, "data": nil})
}

// HandleMarketTechnical serves GET /api/market/technical/:symbol
// Returns technical analysis indicators for a stock.
func (h *Handlers) HandleMarketTechnical(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}
	// Placeholder — will be populated when technical analysis service is ready
	c.JSON(http.StatusOK, gin.H{"symbol": symbol, "data": nil})
}
