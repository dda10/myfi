package handler

import (
	"net/http"
	"strconv"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandleRecommendationSummary returns overall AI recommendation accuracy metrics.
// GET /api/recommendations/summary
func (h *Handlers) HandleRecommendationSummary(c *gin.Context) {
	if h.RecommendationTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracking not enabled"})
		return
	}

	summary := h.RecommendationTracker.GetSummary()
	c.JSON(http.StatusOK, summary)
}

// HandleRecommendationAccuracy returns accuracy metrics for a specific action type.
// GET /api/recommendations/accuracy?action=buy
func (h *Handlers) HandleRecommendationAccuracy(c *gin.Context) {
	if h.RecommendationTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracking not enabled"})
		return
	}

	actionStr := c.Query("action")
	if actionStr == "" {
		actionStr = "buy"
	}

	action := model.RecommendationAction(actionStr)
	if action != model.ActionBuy && action != model.ActionSell && action != model.ActionHold {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be buy/sell/hold"})
		return
	}

	accuracy := h.RecommendationTracker.GetAccuracyByAction(action)
	c.JSON(http.StatusOK, accuracy)
}

// HandleRecommendationList returns filtered list of recommendations.
// GET /api/recommendations?symbol=FPT&action=buy&minConfidence=60&limit=50
func (h *Handlers) HandleRecommendationList(c *gin.Context) {
	if h.RecommendationTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracking not enabled"})
		return
	}

	filter := model.RecommendationFilter{
		Symbol: c.Query("symbol"),
	}

	if action := c.Query("action"); action != "" {
		filter.Action = model.RecommendationAction(action)
	}

	if minConf := c.Query("minConfidence"); minConf != "" {
		if v, err := strconv.Atoi(minConf); err == nil {
			filter.MinConfidence = v
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			filter.Limit = v
		}
	}

	if filter.Limit == 0 {
		filter.Limit = 100
	}

	records := h.RecommendationTracker.GetRecommendations(filter)
	c.JSON(http.StatusOK, gin.H{
		"count":           len(records),
		"recommendations": records,
	})
}

// HandleRecommendationByID returns a specific recommendation by ID.
// GET /api/recommendations/:id
func (h *Handlers) HandleRecommendationByID(c *gin.Context) {
	if h.RecommendationTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracking not enabled"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	record := h.RecommendationTracker.GetRecordByID(id)
	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

// HandleUpdateOutcomes triggers an update of recommendation outcomes.
// POST /api/recommendations/update-outcomes
func (h *Handlers) HandleUpdateOutcomes(c *gin.Context) {
	if h.RecommendationTracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracking not enabled"})
		return
	}

	if err := h.RecommendationTracker.UpdateOutcomes(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "outcomes updated"})
}
