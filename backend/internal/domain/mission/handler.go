package mission

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handlers holds mission domain dependencies for HTTP handler methods.
type Handlers struct {
	MissionService *MissionService
}

// HandleCreateMission serves POST /api/missions — create a new mission.
func (h *Handlers) HandleCreateMission(c *gin.Context) {
	userID := getUserID(c)
	var m Mission
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.UserID = userID

	if err := h.MissionService.Create(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

// HandleListMissions serves GET /api/missions — list all missions for the user.
func (h *Handlers) HandleListMissions(c *gin.Context) {
	userID := getUserID(c)
	missions, err := h.MissionService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": missions})
}

// HandleGetMission serves GET /api/missions/:id — get a single mission.
func (h *Handlers) HandleGetMission(c *gin.Context) {
	userID := getUserID(c)
	missionID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	m, err := h.MissionService.Get(c.Request.Context(), userID, missionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// HandleUpdateMission serves PUT /api/missions/:id — update a mission.
func (h *Handlers) HandleUpdateMission(c *gin.Context) {
	userID := getUserID(c)
	missionID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	var m Mission
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.ID = missionID
	m.UserID = userID

	if err := h.MissionService.Update(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleDeleteMission serves DELETE /api/missions/:id — delete a mission.
func (h *Handlers) HandleDeleteMission(c *gin.Context) {
	userID := getUserID(c)
	missionID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	if err := h.MissionService.Delete(c.Request.Context(), userID, missionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandlePauseMission serves POST /api/missions/:id/pause — pause an active mission.
func (h *Handlers) HandlePauseMission(c *gin.Context) {
	userID := getUserID(c)
	missionID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	if err := h.MissionService.Pause(c.Request.Context(), userID, missionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": "paused"})
}

// HandleResumeMission serves POST /api/missions/:id/resume — resume a paused mission.
func (h *Handlers) HandleResumeMission(c *gin.Context) {
	userID := getUserID(c)
	missionID, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}

	if err := h.MissionService.Resume(c.Request.Context(), userID, missionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "status": "active"})
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

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return v, true
}
