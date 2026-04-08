package consensus

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers holds consensus domain dependencies for HTTP handler methods.
type Handlers struct {
	ConsensusService *ConsensusService
}

// HandleGetConsensus returns the composite consensus score for a symbol.
// GET /api/consensus/score?symbol=SSI&period=7d
func (h *Handlers) HandleGetConsensus(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	period := c.DefaultQuery("period", "7d")

	score, err := h.ConsensusService.GetConsensus(c.Request.Context(), symbol, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}

// HandleGetDivergences returns sentiment divergences between sources.
// GET /api/consensus/divergences?symbol=SSI
func (h *Handlers) HandleGetDivergences(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	divergences, err := h.ConsensusService.GetDivergences(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, divergences)
}

// HandleGetMarketMood returns the overall market-wide sentiment.
// GET /api/consensus/mood
func (h *Handlers) HandleGetMarketMood(c *gin.Context) {
	mood, err := h.ConsensusService.GetMarketMood(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mood)
}
