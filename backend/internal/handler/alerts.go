package handler

import (
	"net/http"
	"strconv"

	"myfi-backend/internal/model"
	"myfi-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HandleGetAlerts returns alerts for the authenticated user.
// GET /api/alerts?includeViewed=false&includeExpired=false&symbol=SSI&limit=50
func (h *Handlers) HandleGetAlerts(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)

	filters := service.AlertFilters{
		Symbol:         c.Query("symbol"),
		PatternType:    c.Query("patternType"),
		IncludeViewed:  c.Query("includeViewed") == "true",
		IncludeExpired: c.Query("includeExpired") == "true",
	}
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			filters.Limit = n
		}
	}

	alerts, err := h.AlertService.GetAlerts(c.Request.Context(), userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

// HandleMarkAlertViewed marks an alert as viewed.
// PUT /api/alerts/:id/viewed
func (h *Handlers) HandleMarkAlertViewed(c *gin.Context) {
	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
		return
	}
	if err := h.AlertService.MarkAlertViewed(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleUpdateAlertPreferences saves alert preferences for the authenticated user.
// PUT /api/alerts/preferences
func (h *Handlers) HandleUpdateAlertPreferences(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	var prefs model.AlertPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prefs.UserID = userID
	if err := h.AlertService.SaveUserPreferences(c.Request.Context(), &prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
