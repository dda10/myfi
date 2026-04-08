package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers holds notification domain dependencies for HTTP handler methods.
type Handlers struct {
	NotificationService *NotificationService
}

// HandleListNotifications serves GET /api/notifications — list notifications for the user.
func (h *Handlers) HandleListNotifications(c *gin.Context) {
	userID := getUserID(c)
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	notifications, err := h.NotificationService.List(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

// HandleMarkRead serves PUT /api/notifications/:id/read — mark a notification as read.
func (h *Handlers) HandleMarkRead(c *gin.Context) {
	userID := getUserID(c)
	notifID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := h.NotificationService.MarkRead(c.Request.Context(), userID, notifID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleMarkAllRead serves PUT /api/notifications/read-all — mark all notifications as read.
func (h *Handlers) HandleMarkAllRead(c *gin.Context) {
	userID := getUserID(c)
	if err := h.NotificationService.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleUnreadCount serves GET /api/notifications/unread — get unread notification count.
func (h *Handlers) HandleUnreadCount(c *gin.Context) {
	userID := getUserID(c)
	count, err := h.NotificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
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
