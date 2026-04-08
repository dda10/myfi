package fund

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleListFunds serves GET /api/funds — list all mutual funds or search by query.
func (h *Handlers) HandleListFunds(c *gin.Context) {
	if h.FundService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fund service not configured"})
		return
	}

	query := c.Query("q")
	var funds []FundRecord
	var err error

	if query != "" {
		funds, err = h.FundService.SearchFunds(c.Request.Context(), query)
	} else {
		funds, err = h.FundService.ListFunds(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": funds})
}

// HandleGetFund serves GET /api/funds/:code — get fund details (holdings + allocation).
func (h *Handlers) HandleGetFund(c *gin.Context) {
	if h.FundService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fund service not configured"})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fund code is required"})
		return
	}

	holdings, err := h.FundService.GetTopHoldings(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	allocation, err := h.FundService.GetAllocation(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"fundCode":   code,
			"holdings":   holdings,
			"allocation": allocation,
		},
	})
}

// HandleGetFundHoldings serves GET /api/funds/:code/holdings — top holdings for a fund.
func (h *Handlers) HandleGetFundHoldings(c *gin.Context) {
	if h.FundService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fund service not configured"})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fund code is required"})
		return
	}

	holdings, err := h.FundService.GetTopHoldings(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": holdings})
}

// HandleGetFundNAV serves GET /api/funds/:code/nav — NAV history for a fund.
func (h *Handlers) HandleGetFundNAV(c *gin.Context) {
	if h.FundService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fund service not configured"})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fund code is required"})
		return
	}

	// Default: last 1 year
	end := time.Now()
	start := end.AddDate(-1, 0, 0)

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

	navs, err := h.FundService.GetNAVHistory(c.Request.Context(), code, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": navs})
}
