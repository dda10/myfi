package screener

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers holds screener domain dependencies for HTTP handler methods.
type Handlers struct {
	ScreenerService *ScreenerService
}

// HandleScreener runs the stock screener with the provided filters.
// POST /api/screener
func (h *Handlers) HandleScreener(c *gin.Context) {
	var filters ScreenerFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	results, total, err := h.ScreenerService.Screen(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, ScreenerResponse{
		Data:       results,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// HandleGetPresets returns saved screener presets for the authenticated user.
// GET /api/screener/presets
func (h *Handlers) HandleGetPresets(c *gin.Context) {
	userID := getUserID(c)
	presets, err := h.ScreenerService.GetPresets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, presets)
}

// HandleSavePreset saves a new screener preset for the authenticated user.
// POST /api/screener/presets
func (h *Handlers) HandleSavePreset(c *gin.Context) {
	userID := getUserID(c)
	var preset FilterPreset
	if err := c.ShouldBindJSON(&preset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	preset.UserID = userID
	id, err := h.ScreenerService.SavePreset(c.Request.Context(), preset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// HandleDeletePreset deletes a screener preset owned by the authenticated user.
// DELETE /api/screener/presets/:id
func (h *Handlers) HandleDeletePreset(c *gin.Context) {
	userID := getUserID(c)
	presetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preset id"})
		return
	}
	if err := h.ScreenerService.DeletePreset(c.Request.Context(), presetID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- Helpers ---

func getUserID(c *gin.Context) string {
	if id, exists := c.Get("userID"); exists {
		if v, ok := id.(string); ok {
			return v
		}
	}
	return ""
}
