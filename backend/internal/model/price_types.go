package model

import "time"

const (
	// FallbackUSDVND is the hardcoded fallback rate when CoinGecko is unavailable.
	FallbackUSDVND = 25400.0
)

// PriceQuote represents a price quote for an asset
type PriceQuote struct {
	Symbol        string    `json:"symbol"`
	AssetType     AssetType `json:"assetType"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
	Volume        int64     `json:"volume"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	IsStale       bool      `json:"isStale"`
}

// OHLCVBar represents a single OHLCV data point
type OHLCVBar struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// FXRate represents a USD/VND exchange rate with metadata.
type FXRate struct {
	Rate      float64   `json:"rate"`
	Source    string    `json:"source"`
	IsStale   bool      `json:"is_stale"`
	Timestamp time.Time `json:"timestamp"`
}

// GoldPriceResponse represents the API response for gold prices.
type GoldPriceResponse struct {
	TypeName  string    `json:"type_name"`
	Branch    string    `json:"branch,omitempty"`
	BuyPrice  float64   `json:"buy_price"`
	SellPrice float64   `json:"sell_price"`
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
}

// CryptoPriceResponse represents a cryptocurrency price with metadata
type CryptoPriceResponse struct {
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	PriceVND         float64   `json:"price_vnd"`
	PriceUSD         float64   `json:"price_usd"`
	Change24h        float64   `json:"change_24h"`
	PercentChange24h float64   `json:"percent_change_24h"`
	Volume24h        float64   `json:"volume_24h"`
	MarketCapVND     float64   `json:"market_cap_vnd"`
	Source           string    `json:"source"`
	Timestamp        time.Time `json:"timestamp"`
}

// CoinGeckoResponse represents the API response from CoinGecko
type CoinGeckoResponse struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	TotalVolume              float64 `json:"total_volume"`
	PriceChange24h           float64 `json:"price_change_24h"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
}
