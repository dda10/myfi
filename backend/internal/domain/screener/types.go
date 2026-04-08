package screener

import (
	"myfi-backend/internal/domain/market"
	"time"
)

// ICBSector and SectorTrend are re-exported from market domain for use in screener types.
type ICBSector = market.ICBSector
type SectorTrend = market.SectorTrend

// ScreenerFilters defines all filter criteria for the stock screener.
type ScreenerFilters struct {
	MinPE            *float64      `json:"minPE,omitempty"`
	MaxPE            *float64      `json:"maxPE,omitempty"`
	MinPB            *float64      `json:"minPB,omitempty"`
	MaxPB            *float64      `json:"maxPB,omitempty"`
	MinMarketCap     *float64      `json:"minMarketCap,omitempty"`
	MinEVEBITDA      *float64      `json:"minEVEBITDA,omitempty"`
	MaxEVEBITDA      *float64      `json:"maxEVEBITDA,omitempty"`
	MinROE           *float64      `json:"minROE,omitempty"`
	MaxROE           *float64      `json:"maxROE,omitempty"`
	MinROA           *float64      `json:"minROA,omitempty"`
	MaxROA           *float64      `json:"maxROA,omitempty"`
	MinRevenueGrowth *float64      `json:"minRevenueGrowth,omitempty"`
	MaxRevenueGrowth *float64      `json:"maxRevenueGrowth,omitempty"`
	MinProfitGrowth  *float64      `json:"minProfitGrowth,omitempty"`
	MaxProfitGrowth  *float64      `json:"maxProfitGrowth,omitempty"`
	MinDivYield      *float64      `json:"minDivYield,omitempty"`
	MaxDivYield      *float64      `json:"maxDivYield,omitempty"`
	MinDebtToEquity  *float64      `json:"minDebtToEquity,omitempty"`
	MaxDebtToEquity  *float64      `json:"maxDebtToEquity,omitempty"`
	Sectors          []ICBSector   `json:"sectors,omitempty"`
	Exchanges        []string      `json:"exchanges,omitempty"`
	SectorTrends     []SectorTrend `json:"sectorTrends,omitempty"`
	SortBy           string        `json:"sortBy"`
	SortOrder        string        `json:"sortOrder"`
	Page             int           `json:"page"`
	PageSize         int           `json:"pageSize"`
}

// ScreenerResult represents a single stock in the screener output.
type ScreenerResult struct {
	Symbol        string      `json:"symbol"`
	Exchange      string      `json:"exchange"`
	Sector        ICBSector   `json:"sector"`
	SectorName    string      `json:"sectorName"`
	MarketCap     float64     `json:"marketCap"`
	PE            float64     `json:"pe"`
	PB            float64     `json:"pb"`
	EVEBITDA      float64     `json:"evEbitda"`
	ROE           float64     `json:"roe"`
	ROA           float64     `json:"roa"`
	RevenueGrowth float64     `json:"revenueGrowth"`
	ProfitGrowth  float64     `json:"profitGrowth"`
	DivYield      float64     `json:"divYield"`
	DebtToEquity  float64     `json:"debtToEquity"`
	SectorTrend   SectorTrend `json:"sectorTrend"`
}

// ScreenerResponse wraps paginated screener results.
type ScreenerResponse struct {
	Data       []ScreenerResult `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// FilterPreset represents a saved set of screener filter criteria.
type FilterPreset struct {
	ID        int64           `json:"id"`
	UserID    string          `json:"userId"`
	Name      string          `json:"name"`
	Filters   ScreenerFilters `json:"filters"`
	CreatedAt time.Time       `json:"createdAt"`
}
