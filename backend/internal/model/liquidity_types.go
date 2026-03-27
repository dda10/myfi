package model

import "time"

// LiquidityConfig holds the filter thresholds for the dynamic whitelist.
type LiquidityConfig struct {
	// MinADTV is the minimum 20-day Average Daily Trading Value in VND
	// for HOSE and HNX stocks. Default: 10 billion VND (10_000_000_000).
	MinADTV float64 `json:"minAdtv"`

	// MinADTVUpcom is the stricter ADTV threshold for UPCoM stocks.
	// Default: 20 billion VND (20_000_000_000).
	MinADTVUpcom float64 `json:"minAdtvUpcom"`

	// MinPrice is the minimum closing price in VND to exclude penny stocks.
	// Default: 10,000 VND.
	MinPrice float64 `json:"minPrice"`

	// LookbackDays is the number of trading days for ADTV calculation.
	// Default: 20.
	LookbackDays int `json:"lookbackDays"`
}

// DefaultLiquidityConfig returns production defaults for the VN market.
func DefaultLiquidityConfig() LiquidityConfig {
	return LiquidityConfig{
		MinADTV:      10_000_000_000, // 10 billion VND
		MinADTVUpcom: 20_000_000_000, // 20 billion VND (stricter for UPCoM)
		MinPrice:     10_000,         // 10,000 VND
		LookbackDays: 20,
	}
}

// WhitelistEntry represents a single ticker that passed the liquidity filter.
type WhitelistEntry struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	ADTV     float64 `json:"adtv"`    // 20-day average daily trading value (VND)
	Price    float64 `json:"price"`   // Latest closing price (VND)
	IsVN100  bool    `json:"isVn100"` // true if part of VN100 always-on list
}

// WhitelistSnapshot is the full result of a filter run.
type WhitelistSnapshot struct {
	Entries   []WhitelistEntry `json:"entries"`
	Total     int              `json:"total"`
	Config    LiquidityConfig  `json:"config"`
	UpdatedAt time.Time        `json:"updatedAt"`
}
