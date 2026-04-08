package research

import "time"

// ReportType represents the type of research report.
type ReportType string

const (
	ReportFactorSnapshot ReportType = "factor_snapshot"
	ReportSectorDeepDive ReportType = "sector_deep_dive"
	ReportMarketOutlook  ReportType = "market_outlook"
)

// ResearchReport represents a generated research report.
type ResearchReport struct {
	ID          int64      `json:"id"`
	Type        ReportType `json:"type"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	PDFS3Key    string     `json:"pdfS3Key,omitempty"`
	GeneratedAt time.Time  `json:"generatedAt"`
}
