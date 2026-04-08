package fund

import (
	"time"

	"myfi-backend/internal/infra"
)

// FundRecord represents a mutual fund listing.
type FundRecord struct {
	FundCode          string    `json:"fundCode"`
	FundName          string    `json:"fundName"`
	ManagementCompany string    `json:"managementCompany"`
	FundType          string    `json:"fundType"`
	NAV               float64   `json:"nav"`
	InceptionDate     time.Time `json:"inceptionDate"`
}

// FundHolding represents a stock holding within a fund.
type FundHolding struct {
	StockSymbol string  `json:"stockSymbol"`
	StockName   string  `json:"stockName"`
	Percentage  float64 `json:"percentage"`
	MarketValue float64 `json:"marketValue"`
}

// FundNAV represents daily NAV data for a fund.
type FundNAV struct {
	Date       time.Time `json:"date"`
	NAVPerUnit float64   `json:"navPerUnit"`
	TotalNAV   float64   `json:"totalNav"`
}

// FundIndustryAlloc represents industry allocation within a fund.
type FundIndustryAlloc struct {
	IndustryName string  `json:"industryName"`
	Percentage   float64 `json:"percentage"`
}

// FundAssetAlloc represents asset class allocation within a fund.
type FundAssetAlloc struct {
	AssetClass string  `json:"assetClass"`
	Percentage float64 `json:"percentage"`
}

// FundAllocation combines industry and asset allocation for a fund.
type FundAllocation struct {
	Industry []FundIndustryAlloc `json:"industry"`
	Asset    []FundAssetAlloc    `json:"asset"`
}

// Handlers holds fund-domain service dependencies for HTTP handler methods.
type Handlers struct {
	FundService *FundService
}

// FundService provides mutual fund data via the FMarket connector.
type FundService struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
}
