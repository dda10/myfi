package portfolio

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handlers holds portfolio domain dependencies for HTTP handler methods.
type Handlers struct {
	Engine      *PortfolioEngine
	Ledger      *TransactionLedger
	Performance *PerformanceEngine
	Risk        *RiskService
	Export      *ExportService
}

// HandleGetHoldings returns all holdings for the authenticated user.
// GET /api/portfolio/holdings
func (h *Handlers) HandleGetHoldings(c *gin.Context) {
	userID := getUserID(c)
	holdings, err := h.Engine.GetHoldings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": holdings})
}

// HandleGetNAV returns the current NAV for the authenticated user.
// GET /api/portfolio/nav
func (h *Handlers) HandleGetNAV(c *gin.Context) {
	userID := getUserID(c)
	nav, err := h.Engine.ComputeNAV(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nav)
}

// HandleRecordBuy records a stock buy transaction.
// POST /api/portfolio/buy
func (h *Handlers) HandleRecordBuy(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		Symbol   string  `json:"symbol" binding:"required"`
		Quantity float64 `json:"quantity" binding:"required"`
		Price    float64 `json:"price" binding:"required"`
		Date     string  `json:"date"`
		Notes    string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txDate := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			txDate = t
		}
	}

	txID, err := h.Engine.RecordBuy(c.Request.Context(), userID, req.Symbol, req.Quantity, req.Price, txDate, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"transactionId": txID})
}

// HandleRecordSell records a stock sell transaction.
// POST /api/portfolio/sell
func (h *Handlers) HandleRecordSell(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		Symbol   string  `json:"symbol" binding:"required"`
		Quantity float64 `json:"quantity" binding:"required"`
		Price    float64 `json:"price" binding:"required"`
		Date     string  `json:"date"`
		Notes    string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txDate := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			txDate = t
		}
	}

	result, err := h.Engine.RecordSell(c.Request.Context(), userID, req.Symbol, req.Quantity, req.Price, txDate, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// HandleGetTransactions returns transaction history for the authenticated user.
// GET /api/portfolio/transactions
func (h *Handlers) HandleGetTransactions(c *gin.Context) {
	userID := getUserID(c)
	txns, err := h.Ledger.GetTransactionsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": txns})
}

// HandleGetPerformance returns performance metrics (TWR, XIRR, benchmark comparison).
// GET /api/portfolio/performance?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleGetPerformance(c *gin.Context) {
	userID := getUserID(c)
	start, end := parseDateRange(c)

	twr, err := h.Performance.ComputeTWR(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	xirr, _ := h.Performance.ComputeXIRR(c.Request.Context(), userID)
	equityCurve, _ := h.Performance.GetEquityCurve(c.Request.Context(), userID, start, end)
	benchmark, _ := h.Performance.ComputeBenchmarkComparison(c.Request.Context(), userID, "VNINDEX", start, end)

	c.JSON(http.StatusOK, gin.H{
		"twr":         twr,
		"xirr":        xirr,
		"equityCurve": equityCurve,
		"benchmark":   benchmark,
	})
}

// HandleGetRisk returns risk metrics (Sharpe, max drawdown, beta, volatility, VaR).
// GET /api/portfolio/risk
func (h *Handlers) HandleGetRisk(c *gin.Context) {
	userID := getUserID(c)

	sharpe, _ := h.Risk.ComputeSharpe(c.Request.Context(), userID, DefaultVNRiskFreeRate)
	maxDD, _ := h.Risk.ComputeMaxDrawdown(c.Request.Context(), userID)
	beta, _ := h.Risk.ComputeBeta(c.Request.Context(), userID, "VNINDEX")
	vol, _ := h.Risk.ComputeVolatility(c.Request.Context(), userID)
	var95, _ := h.Risk.ComputeVaR(c.Request.Context(), userID, 0.95)

	c.JSON(http.StatusOK, RiskMetrics{
		SharpeRatio: sharpe,
		MaxDrawdown: maxDD,
		Beta:        beta,
		Volatility:  vol,
		VaR95:       var95,
	})
}

// HandleGetAllocation returns sector allocation for the portfolio.
// GET /api/portfolio/allocation
func (h *Handlers) HandleGetAllocation(c *gin.Context) {
	userID := getUserID(c)
	alloc, err := h.Engine.GetSectorAllocation(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alloc})
}

// HandleGetPnL returns P&L summary with VN tax calculation.
// GET /api/portfolio/pnl?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleGetPnL(c *gin.Context) {
	userID := getUserID(c)
	start, end := parseDateRange(c)

	pnl, err := h.Export.ComputePnL(c.Request.Context(), userID, &start, &end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pnl})
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

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return v, true
}
