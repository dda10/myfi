package portfolio

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleExportTransactions exports transaction history as CSV.
// GET /api/export/transactions?format=csv&start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleExportTransactions(c *gin.Context) {
	userID := getUserID(c)
	format := ExportFormat(c.DefaultQuery("format", "csv"))
	start, end := parseDateRange(c)

	result, err := h.Export.Export(c.Request.Context(), ExportRequest{
		UserID:    userID,
		Type:      ExportTransactions,
		Format:    format,
		StartDate: &start,
		EndDate:   &end,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

// HandleExportPortfolio exports current portfolio snapshot as CSV.
// GET /api/export/portfolio?format=csv
func (h *Handlers) HandleExportPortfolio(c *gin.Context) {
	userID := getUserID(c)
	format := ExportFormat(c.DefaultQuery("format", "csv"))

	result, err := h.Export.Export(c.Request.Context(), ExportRequest{
		UserID: userID,
		Type:   ExportPortfolio,
		Format: format,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

// HandleExportPnL exports P&L report with VN tax as CSV/PDF.
// GET /api/export/pnl?format=csv&start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleExportPnL(c *gin.Context) {
	userID := getUserID(c)
	format := ExportFormat(c.DefaultQuery("format", "csv"))
	start, end := parseDateRange(c)

	result, err := h.Export.Export(c.Request.Context(), ExportRequest{
		UserID:    userID,
		Type:      ExportPnL,
		Format:    format,
		StartDate: &start,
		EndDate:   &end,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

// HandleExportTax exports a VN tax report (0.1% sell tax) as CSV.
// GET /api/export/tax?start=2024-01-01&end=2024-12-31
func (h *Handlers) HandleExportTax(c *gin.Context) {
	userID := getUserID(c)
	start, end := parseDateRange(c)

	pnl, err := h.Export.ComputePnL(c.Request.Context(), userID, &start, &end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "ezistock_tax_" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.JSON(http.StatusOK, gin.H{"data": pnl})
}
