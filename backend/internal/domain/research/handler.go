package research

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers holds research domain dependencies for HTTP handler methods.
type Handlers struct {
	// DB and S3 storage will be wired in Task 14.3 (DI container update).
}

// HandleListReports serves GET /api/research/reports — list published research reports.
func (h *Handlers) HandleListReports(c *gin.Context) {
	reportType := c.Query("type") // factor_snapshot, sector_deep_dive, market_outlook

	// Placeholder: query research_reports table.
	_ = reportType
	c.JSON(http.StatusOK, gin.H{"data": []ResearchReport{}, "total": 0})
}

// HandleGetReport serves GET /api/research/reports/:id — get a single research report.
func (h *Handlers) HandleGetReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report id is required"})
		return
	}

	// Placeholder: query research_reports table by ID.
	c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
}

// HandleDownloadReportPDF serves GET /api/research/reports/:id/pdf — download report PDF.
func (h *Handlers) HandleDownloadReportPDF(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report id is required"})
		return
	}

	// Placeholder: fetch PDF from S3 using the report's pdfS3Key and return as download.
	c.JSON(http.StatusNotFound, gin.H{"error": "PDF not available"})
}
