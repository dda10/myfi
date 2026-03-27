package handler

import (
	"myfi-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleSignalScan triggers a full market scan and returns ranked signals.
// GET /api/signals/scan
func (h *Handlers) HandleSignalScan(c *gin.Context) {
	if h.SignalEngine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "signal engine not enabled"})
		return
	}

	result, err := h.SignalEngine.ScanMarket(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleSignalConfig returns the current signal engine configuration.
// GET /api/signals/config
func (h *Handlers) HandleSignalConfig(c *gin.Context) {
	if h.SignalEngine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "signal engine not enabled"})
		return
	}

	// Return default config for now (could be made configurable per user)
	c.JSON(http.StatusOK, gin.H{
		"config": h.SignalEngine,
	})
}

// BacktestSignals runs a historical backtest of the signal engine strategy.
// POST /api/signals/backtest
func (h *Handlers) BacktestSignals(c *gin.Context) {
	var req service.SignalBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate are required"})
		return
	}
	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be after startDate"})
		return
	}

	result, err := h.SignalBacktester.RunBacktest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// OptimizeSignalWeights runs a grid search to find optimal factor weights.
// POST /api/signals/optimize
func (h *Handlers) OptimizeSignalWeights(c *gin.Context) {
	var req service.SignalBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate are required"})
		return
	}

	results, err := h.SignalBacktester.OptimizeWeights(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topConfigurations": results,
		"count":             len(results),
	})
}
