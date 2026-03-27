package handler

import (
	"net/http"
	"strconv"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HandleGetGoals returns all financial goals for the authenticated user.
// GET /api/goals
func (h *Handlers) HandleGetGoals(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	goals, err := h.GoalPlanner.GetGoalsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goals)
}

// HandleCreateGoal creates a new financial goal.
// POST /api/goals
func (h *Handlers) HandleCreateGoal(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	var goal model.FinancialGoal
	if err := c.ShouldBindJSON(&goal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	goal.UserID = userID
	id, err := h.GoalPlanner.CreateGoal(c.Request.Context(), goal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// HandleUpdateGoal updates an existing financial goal.
// PUT /api/goals/:id
func (h *Handlers) HandleUpdateGoal(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}
	var goal model.FinancialGoal
	if err := c.ShouldBindJSON(&goal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	goal.ID = goalID
	goal.UserID = userID
	if err := h.GoalPlanner.UpdateGoal(c.Request.Context(), goal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleDeleteGoal deletes a financial goal.
// DELETE /api/goals/:id
func (h *Handlers) HandleDeleteGoal(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}
	if err := h.GoalPlanner.DeleteGoal(c.Request.Context(), goalID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleGetGoalProgress returns progress metrics for a specific goal.
// GET /api/goals/:id/progress
func (h *Handlers) HandleGetGoalProgress(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	goal, err := h.GoalPlanner.GetGoal(c.Request.Context(), goalID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get current NAV from portfolio
	summary, err := h.PortfolioEngine.GetPortfolioSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	progress := service.ComputeProgress(summary.NAV, goal.TargetAmount, goal.TargetDate, time.Now())
	progress.GoalID = goalID
	c.JSON(http.StatusOK, progress)
}
