package market

import (
	"net/http"

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
	sector := ICBSector(c.Param("sector"))
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

// HandleSectorAverages returns median fundamental metrics for a sector.
// GET /api/sectors/:sector/averages
func (h *Handlers) HandleSectorAverages(c *gin.Context) {
	sectorCode := c.Param("sector")
	result, err := h.SectorService.GetSectorMedianFundamentals(c.Request.Context(), sectorCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleSectorTrend returns the trend direction for a sector.
// GET /api/sectors/:sector/trend
func (h *Handlers) HandleSectorTrend(c *gin.Context) {
	sectorCode := c.Param("sector")
	trend, err := h.SectorService.GetSectorTrend(c.Request.Context(), sectorCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sector": sectorCode, "trend": trend})
}

// HandleSectorStocks returns all stocks in a given sector with key metrics.
// GET /api/sectors/:sector/stocks
func (h *Handlers) HandleSectorStocks(c *gin.Context) {
	sectorCode := c.Param("sector")
	stocks, err := h.SectorService.GetSectorStocks(c.Request.Context(), sectorCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stocks)
}

// HandleStockSectorMapping returns the full stock-to-sector mapping.
// GET /api/sectors/mapping
func (h *Handlers) HandleStockSectorMapping(c *gin.Context) {
	mapping, err := h.SectorService.GetStockSectorMapping(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapping)
}
