package analyst

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handlers holds analyst domain dependencies for HTTP handler methods.
type Handlers struct {
	AnalystService *AnalystIQService
}

// HandleGetConsensus serves GET /api/analyst/consensus/:symbol — consensus recommendation.
func (h *Handlers) HandleGetConsensus(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	consensus, err := h.AnalystService.GetConsensus(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consensus)
}

// HandleGetReports serves GET /api/analyst/reports/:symbol — analyst reports for a symbol.
func (h *Handlers) HandleGetReports(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	reports, err := h.AnalystService.GetReports(c.Request.Context(), symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reports, "total": len(reports)})
}

// HandleGetAnalystAccuracy serves GET /api/analyst/accuracy/:analyst — analyst accuracy.
func (h *Handlers) HandleGetAnalystAccuracy(c *gin.Context) {
	analystName := c.Param("analyst")
	if analystName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "analyst name is required"})
		return
	}

	accuracy, count, err := h.AnalystService.GetAnalystAccuracy(c.Request.Context(), analystName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"analyst":  analystName,
		"accuracy": accuracy,
		"reports":  count,
	})
}
