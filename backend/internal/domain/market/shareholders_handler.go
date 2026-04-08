package market

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleShareholders serves GET /api/market/shareholders?symbol=VNM
func (h *Handlers) HandleShareholders(c *gin.Context) {
	symbol := strings.ToUpper(c.Query("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol query parameter is required"})
		return
	}

	client := h.DataSourceRouter.VCIClient()
	if client == nil {
		client = h.DataSourceRouter.KBSClient()
	}
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}

	shareholders, err := client.Shareholders(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shareholders, "symbol": symbol})
}

// HandleSubsidiaries serves GET /api/market/subsidiaries?symbol=VNM
func (h *Handlers) HandleSubsidiaries(c *gin.Context) {
	symbol := strings.ToUpper(c.Query("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol query parameter is required"})
		return
	}

	client := h.DataSourceRouter.VCIClient()
	if client == nil {
		client = h.DataSourceRouter.KBSClient()
	}
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data source available"})
		return
	}

	subsidiaries, err := client.Subsidiaries(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subsidiaries, "symbol": symbol})
}
