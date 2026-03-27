package handler

import (
	"net/http"
	"strconv"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandleGetWatchlists returns all watchlists for the authenticated user.
// GET /api/watchlists
func (h *Handlers) HandleGetWatchlists(c *gin.Context) {
	userID := int(c.MustGet("claims").(*model.JWTClaims).UserID)
	lists, err := h.WatchlistService.GetWatchlists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lists)
}

// HandleCreateWatchlist creates a new watchlist.
// POST /api/watchlists
func (h *Handlers) HandleCreateWatchlist(c *gin.Context) {
	userID := int(c.MustGet("claims").(*model.JWTClaims).UserID)
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	wl, err := h.WatchlistService.CreateWatchlist(userID, body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wl)
}

// HandleRenameWatchlist renames an existing watchlist.
// PUT /api/watchlists/:id
func (h *Handlers) HandleRenameWatchlist(c *gin.Context) {
	userID := int(c.MustGet("claims").(*model.JWTClaims).UserID)
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if err := h.WatchlistService.RenameWatchlist(userID, wlID, body.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleDeleteWatchlist deletes a watchlist.
// DELETE /api/watchlists/:id
func (h *Handlers) HandleDeleteWatchlist(c *gin.Context) {
	userID := int(c.MustGet("claims").(*model.JWTClaims).UserID)
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	if err := h.WatchlistService.DeleteWatchlist(userID, wlID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleAddWatchlistSymbol adds a symbol to a watchlist.
// POST /api/watchlists/:id/symbols
func (h *Handlers) HandleAddWatchlistSymbol(c *gin.Context) {
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	var body struct {
		Symbol string `json:"symbol"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	if err := h.WatchlistService.AddSymbol(wlID, body.Symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// HandleRemoveWatchlistSymbol removes a symbol from a watchlist.
// DELETE /api/watchlists/:id/symbols/:symbol
func (h *Handlers) HandleRemoveWatchlistSymbol(c *gin.Context) {
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	symbol := c.Param("symbol")
	if err := h.WatchlistService.RemoveSymbol(wlID, symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleSetPriceAlert sets price alert thresholds for a symbol in a watchlist.
// PUT /api/watchlists/:id/symbols/:symbol/alert
func (h *Handlers) HandleSetPriceAlert(c *gin.Context) {
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	symbol := c.Param("symbol")
	var body struct {
		AlertAbove *float64 `json:"alertAbove"`
		AlertBelow *float64 `json:"alertBelow"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.WatchlistService.SetPriceAlert(wlID, symbol, body.AlertAbove, body.AlertBelow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// HandleReorderWatchlist reorders symbols in a watchlist.
// PUT /api/watchlists/:id/reorder
func (h *Handlers) HandleReorderWatchlist(c *gin.Context) {
	wlID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid watchlist id"})
		return
	}
	var body struct {
		Symbols []string `json:"symbols"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Symbols) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbols array is required"})
		return
	}
	if err := h.WatchlistService.ReorderSymbols(wlID, body.Symbols); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
