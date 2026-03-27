package handler

import (
	"net/http"
	"strconv"
	"time"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandlePortfolioSummary returns the full portfolio overview for the authenticated user.
// GET /api/portfolio/summary
func (h *Handlers) HandlePortfolioSummary(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	summary, err := h.PortfolioEngine.GetPortfolioSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// HandleAddAsset adds a new asset to the user's portfolio.
// POST /api/portfolio/assets
func (h *Handlers) HandleAddAsset(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.UserID = userID
	id, err := h.AssetRegistry.AddAsset(c.Request.Context(), asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// HandleUpdateAsset updates an existing asset.
// PUT /api/portfolio/assets/:id
func (h *Handlers) HandleUpdateAsset(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	assetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset id"})
		return
	}
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.ID = assetID
	asset.UserID = userID
	if err := h.AssetRegistry.UpdateAsset(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleDeleteAsset removes an asset from the user's portfolio.
// DELETE /api/portfolio/assets/:id
func (h *Handlers) HandleDeleteAsset(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	assetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset id"})
		return
	}
	if err := h.AssetRegistry.DeleteAsset(c.Request.Context(), assetID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleGetTransactions returns all transactions for the authenticated user.
// GET /api/portfolio/transactions
func (h *Handlers) HandleGetTransactions(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	txns, err := h.TransactionLedger.GetTransactionsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txns)
}

// HandleRecordTransaction records a new transaction.
// POST /api/portfolio/transactions
func (h *Handlers) HandleRecordTransaction(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	var tx model.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx.UserID = userID
	id, err := h.TransactionLedger.RecordTransaction(c.Request.Context(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// HandlePortfolioPerformance returns performance metrics for the authenticated user.
// GET /api/portfolio/performance?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandlePortfolioPerformance(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)

	start, end := parseDateRange(c)
	metrics, err := h.PerformanceEngine.GetPerformanceMetrics(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// HandlePortfolioRisk returns risk metrics for the authenticated user.
// GET /api/portfolio/risk
func (h *Handlers) HandlePortfolioRisk(c *gin.Context) {
	userID := int64(c.MustGet("claims").(*model.JWTClaims).UserID)
	metrics, err := h.RiskService.ComputeRiskMetrics(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// parseDateRange parses optional start/end query params (YYYY-MM-DD).
// Defaults to 1 year ago → now if not provided.
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	end := now
	start := now.AddDate(-1, 0, 0)

	if s := c.Query("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}
	return start, end
}
