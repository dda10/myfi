package handler

import (
	"net/http"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// HandleExportTransactions exports transaction history as CSV.
// GET /api/export/transactions?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleExportTransactions(c *gin.Context) {
	userID := c.MustGet("claims").(*model.JWTClaims).UserID
	from, to := parseDateRange(c)

	txns, err := h.TransactionLedger.GetTransactionsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := h.ExportService.ExportTransactionsCSVBytes(txns, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "transactions_" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", data)
}

// HandleExportSnapshot exports the current portfolio snapshot as CSV.
// GET /api/export/snapshot
func (h *Handlers) HandleExportSnapshot(c *gin.Context) {
	userID := c.MustGet("claims").(*model.JWTClaims).UserID

	summary, err := h.PortfolioEngine.GetPortfolioSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows := holdingsToSnapshotRows(summary.Holdings)
	data, err := h.ExportService.ExportSnapshotCSVBytes(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "snapshot_" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", data)
}

// HandleExportReport exports a portfolio report as PDF (text-based).
// GET /api/export/report
func (h *Handlers) HandleExportReport(c *gin.Context) {
	userID := c.MustGet("claims").(*model.JWTClaims).UserID

	summary, err := h.PortfolioEngine.GetPortfolioSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows := holdingsToSnapshotRows(summary.Holdings)

	// Build allocation map from holdings for the report
	allocation := make(map[model.AssetType]float64)
	for _, row := range rows {
		allocation[row.AssetType] += row.MarketValue
	}

	data, err := h.ExportService.ExportPortfolioReportPDFBytes(summary.NAV, allocation, rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "portfolio_report_" + time.Now().Format("20060102") + ".txt"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// HandleExportTax exports a tax report (capital gains) as CSV.
// GET /api/export/tax?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleExportTax(c *gin.Context) {
	userID := c.MustGet("claims").(*model.JWTClaims).UserID
	from, to := parseDateRange(c)

	txns, err := h.TransactionLedger.GetTransactionsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build average cost map from current holdings
	summary, err := h.PortfolioEngine.GetPortfolioSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	avgCosts := make(map[string]float64, len(summary.Holdings))
	for _, h := range summary.Holdings {
		avgCosts[h.Holding.Symbol] = h.Holding.AverageCost
	}

	data, err := h.ExportService.ExportTaxReportCSVBytes(txns, avgCosts, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "tax_report_" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv", data)
}

// holdingsToSnapshotRows converts HoldingDetail slice to SnapshotRow slice.
func holdingsToSnapshotRows(holdings []model.HoldingDetail) []service.SnapshotRow {
	rows := make([]service.SnapshotRow, 0, len(holdings))
	for _, hd := range holdings {
		rows = append(rows, service.SnapshotRow{
			AssetType:    model.AssetType(model.VNStock), // default to VNStock for EziStock
			Symbol:       hd.Holding.Symbol,
			Quantity:     hd.Holding.Quantity,
			AverageCost:  hd.Holding.AverageCost,
			CurrentPrice: hd.CurrentPrice,
			MarketValue:  hd.MarketValue,
			UnrealizedPL: hd.UnrealizedPL,
		})
	}
	return rows
}
