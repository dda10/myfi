package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleGetSettings serves GET /api/settings — get user settings.
func (h *Handlers) HandleGetSettings(c *gin.Context) {
	userID := getUserID(c)
	user, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"theme":    user.ThemePreference,
		"language": user.LanguagePreference,
	})
}

// HandleUpdateSettings serves PUT /api/settings — update user settings.
func (h *Handlers) HandleUpdateSettings(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		Theme    *string `json:"theme"`
		Language *string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate theme
	if req.Theme != nil {
		if *req.Theme != "light" && *req.Theme != "dark" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "theme must be 'light' or 'dark'"})
			return
		}
	}

	// Validate language
	if req.Language != nil {
		if *req.Language != "vi-VN" && *req.Language != "en-US" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "language must be 'vi-VN' or 'en-US'"})
			return
		}
	}

	if err := h.AuthService.UpdateSettings(userID, req.Theme, req.Language); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
