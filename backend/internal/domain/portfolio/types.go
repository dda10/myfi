package portfolio

import "time"

// Holding represents a user's stock holding with cost basis tracking.
type Holding struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"userId"`
	Symbol          string    `json:"symbol"`
	Quantity        float64   `json:"quantity"`
	AverageCost     float64   `json:"averageCost"`
	TotalDividends  float64   `json:"totalDividends"`
	AcquisitionDate time.Time `json:"acquisitionDate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// HoldingDetail contains per-holding valuation and P&L data.
type HoldingDetail struct {
	Holding         Holding `json:"holding"`
	CurrentPrice    float64 `json:"currentPrice"`
	MarketValue     float64 `json:"marketValue"`
	UnrealizedPL    float64 `json:"unrealizedPL"`
	UnrealizedPLPct float64 `json:"unrealizedPLPct"`
}

// NAVResult contains the computed net asset value for a portfolio.
type NAVResult struct {
	TotalNAV         float64         `json:"totalNav"`
	NAVChange24h     float64         `json:"navChange24h"`
	NAVChangePercent float64         `json:"navChangePercent"`
	Holdings         []HoldingDetail `json:"holdings"`
	ComputedAt       time.Time       `json:"computedAt"`
}

// NAVSnapshot represents a daily NAV data point for the equity curve.
type NAVSnapshot struct {
	Date time.Time `json:"date"`
	NAV  float64   `json:"nav"`
}

// SectorAllocation represents the portfolio allocation to a single sector.
type SectorAllocation struct {
	Sector     string  `json:"sector"`
	SectorName string  `json:"sectorName"`
	Value      float64 `json:"value"`
	Weight     float64 `json:"weight"`
	StockCount int     `json:"stockCount"`
}
