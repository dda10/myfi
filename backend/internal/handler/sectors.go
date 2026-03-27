package handler

import (
	"net/http"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandleAllSectorPerformances returns performance for all ICB sectors.
// GET /api/sectors/performance
func (h *Handlers) HandleAllSectorPerformances(c *gin.Context) {
	result, err := h.SectorService.GetAllSectorPerformances(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleSectorPerformance returns performance for a specific sector.
// GET /api/sectors/:sector/performance
func (h *Handlers) HandleSectorPerformance(c *gin.Context) {
	sector := model.ICBSector(c.Param("sector"))
	result, err := h.SectorService.GetSectorPerformance(c.Request.Context(), sector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleSymbolSector returns the ICB sector for a given symbol.
// GET /api/sectors/symbol/:symbol
func (h *Handlers) HandleSymbolSector(c *gin.Context) {
	symbol := c.Param("symbol")
	sector, err := h.SectorService.GetStockSector(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"symbol": symbol, "sector": sector})
}

// HandleSectorAverages returns fundamental averages for a sector.
// GET /api/sectors/:sector/averages
func (h *Handlers) HandleSectorAverages(c *gin.Context) {
	sector := model.ICBSector(c.Param("sector"))
	result, err := h.SectorService.GetSectorAverages(c.Request.Context(), sector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleSectorStocks returns all stocks in a given sector.
// GET /api/sectors/:sector/stocks
func (h *Handlers) HandleSectorStocks(c *gin.Context) {
	sector := model.ICBSector(c.Param("sector"))
	result, err := h.ComparisonEngine.GetSectorStocks(c.Request.Context(), sector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
