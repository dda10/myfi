package handler

import (
	"net/http"
	"strings"
	"time"

	"myfi-backend/internal/model"

	"github.com/gin-gonic/gin"
)

// HandlePriceQuotes returns real-time quotes for the given symbols and asset type.
// GET /api/prices/quotes?symbols=SSI,FPT&assetType=VNStock
func (h *Handlers) HandlePriceQuotes(c *gin.Context) {
	raw := c.Query("symbols")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbols parameter is required"})
		return
	}
	symbols := strings.Split(raw, ",")
	assetType := model.AssetType(c.DefaultQuery("assetType", string(model.VNStock)))

	quotes, err := h.PriceService.GetQuotes(c.Request.Context(), symbols, assetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

// HandlePriceHistory returns historical OHLCV data for a symbol.
// GET /api/prices/history?symbol=SSI&start=2024-01-01&end=2024-12-31&interval=1D
func (h *Handlers) HandlePriceHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}

	interval := c.DefaultQuery("interval", "1D")
	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	end := now

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

	bars, err := h.PriceService.GetHistoricalData(c.Request.Context(), symbol, start, end, interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bars)
}

// HandleGoldPrices returns current gold prices.
// GET /api/prices/gold
func (h *Handlers) HandleGoldPrices(c *gin.Context) {
	prices, err := h.GoldService.GetGoldPrices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prices)
}

// HandleCryptoPrices returns crypto quotes via CryptoService.
// GET /api/prices/crypto?ids=bitcoin,ethereum
func (h *Handlers) HandleCryptoPrices(c *gin.Context) {
	ids := c.DefaultQuery("ids", "bitcoin,ethereum,binancecoin")
	symbols := strings.Split(ids, ",")

	quotes, err := h.CryptoService.GetCryptoPrices(symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}
