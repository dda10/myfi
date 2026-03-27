package handler

import (
	"net/http"
	"strconv"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HandleGetObservations returns pattern observations with optional filters.
// GET /api/knowledge/observations?symbol=SSI&patternType=breakout&minConfidence=70&limit=50
func (h *Handlers) HandleGetObservations(c *gin.Context) {
	filters := service.ObservationFilters{
		Symbol:      c.Query("symbol"),
		PatternType: c.Query("patternType"),
	}
	if mc := c.Query("minConfidence"); mc != "" {
		if n, err := strconv.Atoi(mc); err == nil {
			filters.MinConfidence = n
		}
	}
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			filters.Limit = n
		}
	}
	if s := c.Query("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.StartDate = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			filters.EndDate = t
		}
	}

	observations, err := h.KnowledgeBase.QueryObservations(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, observations)
}

// HandleGetAccuracy returns accuracy metrics for a pattern type.
// GET /api/knowledge/accuracy/:patternType
func (h *Handlers) HandleGetAccuracy(c *gin.Context) {
	patternType := model.PatternType(c.Param("patternType"))
	metrics, err := h.KnowledgeBase.GetAccuracyMetrics(c.Request.Context(), patternType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}
