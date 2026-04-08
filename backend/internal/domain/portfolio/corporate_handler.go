package portfolio

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleUpcomingCorporateEvents returns upcoming corporate events for portfolio + watchlist stocks.
// GET /api/portfolio/corporate/upcoming
func (h *Handlers) HandleUpcomingCorporateEvents(c *gin.Context) {
	// Placeholder: in a full implementation, this would query the corporate_actions table
	// for events with ex_date in the future, filtered to the user's holdings and watchlist.
	c.JSON(http.StatusOK, gin.H{"data": []CorporateAction{}, "message": "no upcoming events"})
}

// HandleSymbolCorporateEvents returns corporate events for a specific symbol.
// GET /api/portfolio/corporate/:symbol
func (h *Handlers) HandleSymbolCorporateEvents(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	// Placeholder: query corporate_actions table for this symbol.
	c.JSON(http.StatusOK, gin.H{"data": []CorporateAction{}, "symbol": symbol})
}
