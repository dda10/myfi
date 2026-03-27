package handler

import (
	"net/http"
	"strings"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandleComparisonValuation compares valuation metrics across symbols.
// GET /api/comparison/valuation?symbols=SSI,FPT&period=1Y
func (h *Handlers) HandleComparisonValuation(c *gin.Context) {
	symbols, period, err := parseComparisonParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.ComparisonEngine.CompareValuation(c.Request.Context(), symbols, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleComparisonPerformance compares price performance across symbols.
// GET /api/comparison/performance?symbols=SSI,FPT&period=1Y
func (h *Handlers) HandleComparisonPerformance(c *gin.Context) {
	symbols, period, err := parseComparisonParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.ComparisonEngine.ComparePerformance(c.Request.Context(), symbols, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// HandleComparisonCorrelation computes return correlation across symbols.
// GET /api/comparison/correlation?symbols=SSI,FPT&period=1Y
func (h *Handlers) HandleComparisonCorrelation(c *gin.Context) {
	symbols, period, err := parseComparisonParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.ComparisonEngine.ComputeCorrelation(c.Request.Context(), symbols, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// parseComparisonParams extracts and validates symbols and period from query params.
func parseComparisonParams(c *gin.Context) ([]string, model.TimePeriod, error) {
	raw := c.Query("symbols")
	if raw == "" {
		return nil, "", &paramError{"symbols parameter is required"}
	}
	symbols := strings.Split(raw, ",")
	if len(symbols) > model.MaxComparisonStocks {
		return nil, "", &paramError{"too many symbols: max " + string(rune('0'+model.MaxComparisonStocks))}
	}

	period := model.TimePeriod(c.DefaultQuery("period", string(model.Period1Y)))
	return symbols, period, nil
}

type paramError struct{ msg string }

func (e *paramError) Error() string { return e.msg }
