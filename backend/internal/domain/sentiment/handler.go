package sentiment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers holds sentiment domain dependencies for HTTP handler methods.
type Handlers struct {
	SentimentService *SentimentService
}

// HandleAnalyze runs on-demand sentiment analysis on an article.
// POST /api/sentiment/analyze
func (h *Handlers) HandleAnalyze(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analysis, err := h.SentimentService.AnalyzeArticle(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AnalyzeResponse{Analysis: analysis})
}

// HandleGetTrend returns aggregated sentiment trend for a symbol.
// GET /api/sentiment/trend?symbol=SSI&period=7d
func (h *Handlers) HandleGetTrend(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	period := c.DefaultQuery("period", "7d")

	trend, err := h.SentimentService.GetTrend(c.Request.Context(), symbol, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trend)
}

// HandleGetTimeSeries returns daily sentiment snapshots for charting.
// GET /api/sentiment/timeseries?symbol=SSI&days=30
func (h *Handlers) HandleGetTimeSeries(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}

	snapshots, err := h.SentimentService.GetTimeSeries(c.Request.Context(), symbol, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, snapshots)
}

// HandleGetArticles returns recently analyzed articles for a symbol.
// GET /api/sentiment/articles?symbol=SSI&limit=20
func (h *Handlers) HandleGetArticles(c *gin.Context) {
	symbol := c.Query("symbol")
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

	articles, err := h.SentimentService.GetRecentArticles(c.Request.Context(), symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, articles)
}
