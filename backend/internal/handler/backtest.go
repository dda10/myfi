package handler

import (
	"net/http"

	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
	"github.com/gin-gonic/gin"
)

// HandleBacktest runs a strategy backtest over historical OHLCV data.
// POST /api/backtest
func (h *Handlers) HandleBacktest(c *gin.Context) {
	var req model.BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	// Fetch OHLCV history via DataSourceRouter
	quotes, _, err := h.DataSourceRouter.FetchQuoteHistory(c.Request.Context(), vnstock.QuoteHistoryRequest{
		Symbol:   req.Symbol,
		Start:    req.StartDate,
		End:      req.EndDate,
		Interval: "1D",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch price history: " + err.Error()})
		return
	}

	// Convert vnstock.Quote to model.OHLCVBar
	bars := make([]model.OHLCVBar, 0, len(quotes))
	for _, q := range quotes {
		bars = append(bars, model.OHLCVBar{
			Time:   q.Timestamp,
			Open:   q.Open,
			High:   q.High,
			Low:    q.Low,
			Close:  q.Close,
			Volume: q.Volume,
		})
	}

	result, err := h.BacktestEngine.RunBacktest(bars, req.Strategy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
