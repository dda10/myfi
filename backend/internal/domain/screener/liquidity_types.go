package screener

import "time"

// LiquidityTier classifies a stock's tradability level.
type LiquidityTier int

const (
	Tier1 LiquidityTier = 1 // Highly liquid (score >= 70)
	Tier2 LiquidityTier = 2 // Moderate liquidity (score 40-69)
	Tier3 LiquidityTier = 3 // Illiquid (score < 40)
)

// String returns a human-readable tier label.
func (t LiquidityTier) String() string {
	switch t {
	case Tier1:
		return "Tier 1 (Highly Liquid)"
	case Tier2:
		return "Tier 2 (Moderate)"
	case Tier3:
		return "Tier 3 (Illiquid)"
	default:
		return "Unknown"
	}
}

// LiquidityConfig holds the filter thresholds and scoring weights.
type LiquidityConfig struct {
	// MinAvgDailyVolume is the minimum 20-day average daily volume in shares.
	// Default: 50,000 shares (Req 39.3).
	MinAvgDailyVolume float64 `json:"minAvgDailyVolume"`

	// MinAvgDailyValue is the minimum 20-day average daily trading value in VND.
	// Default: 500,000,000 VND (500 million) (Req 39.3).
	MinAvgDailyValue float64 `json:"minAvgDailyValue"`

	// MaxZeroVolumeDays is the maximum allowed zero-volume days out of 20 sessions.
	// Default: 3 (Req 39.3).
	MaxZeroVolumeDays int `json:"maxZeroVolumeDays"`

	// LookbackDays is the number of trading days for score calculation.
	// Default: 20.
	LookbackDays int `json:"lookbackDays"`

	// Tier1Threshold is the minimum score for Tier 1 classification.
	// Default: 70 (Req 39.2).
	Tier1Threshold int `json:"tier1Threshold"`

	// Tier2Threshold is the minimum score for Tier 2 classification.
	// Default: 40 (Req 39.2).
	Tier2Threshold int `json:"tier2Threshold"`

	// Weights for each scoring component (must sum to 100).
	WeightADTV              int `json:"weightAdtv"`
	WeightVolumeConsistency int `json:"weightVolumeConsistency"`
	WeightSpreadProxy       int `json:"weightSpreadProxy"`
	WeightZeroVolumeDays    int `json:"weightZeroVolumeDays"`
	WeightFreeFloat         int `json:"weightFreeFloat"`
}

// DefaultLiquidityConfig returns production defaults for the VN market.
func DefaultLiquidityConfig() LiquidityConfig {
	return LiquidityConfig{
		MinAvgDailyVolume:       50_000,      // 50k shares
		MinAvgDailyValue:        500_000_000, // 500 million VND
		MaxZeroVolumeDays:       3,           // max 3 out of 20
		LookbackDays:            20,          // 20 trading days
		Tier1Threshold:          70,          // score >= 70 → Tier 1
		Tier2Threshold:          40,          // score 40-69 → Tier 2
		WeightADTV:              35,          // ADTV is the strongest signal
		WeightVolumeConsistency: 20,          // volume consistency
		WeightSpreadProxy:       15,          // bid-ask spread proxy
		WeightZeroVolumeDays:    15,          // zero-volume penalty
		WeightFreeFloat:         15,          // free-float ratio
	}
}

// LiquidityScore represents the computed tradability score for a stock.
type LiquidityScore struct {
	Symbol    string        `json:"symbol"`
	Exchange  string        `json:"exchange"`
	Score     int           `json:"score"` // 0-100
	Tier      LiquidityTier `json:"tier"`
	ADTV      float64       `json:"adtv"`                // 20-day avg daily trading value (VND)
	AvgVolume float64       `json:"avgVolume"`           // 20-day avg daily volume (shares)
	ZeroDays  int           `json:"zeroDays"`            // zero-volume days in lookback
	FreeFloat float64       `json:"freeFloat,omitempty"` // free-float ratio (0-1)
	UpdatedAt time.Time     `json:"updatedAt"`

	// Sub-scores for transparency (each 0-100 before weighting).
	ADTVScore              int `json:"adtvScore"`
	VolumeConsistencyScore int `json:"volumeConsistencyScore"`
	SpreadProxyScore       int `json:"spreadProxyScore"`
	ZeroVolumeDaysScore    int `json:"zeroVolumeDaysScore"`
	FreeFloatScore         int `json:"freeFloatScore"`
}

// LiquiditySnapshot is the full result of a filter refresh run.
type LiquiditySnapshot struct {
	Scores    []LiquidityScore `json:"scores"`
	Total     int              `json:"total"`
	Config    LiquidityConfig  `json:"config"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// --- Legacy aliases for backward compatibility ---

// WhitelistEntry is kept for backward compatibility with existing code.
type WhitelistEntry struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	ADTV     float64 `json:"adtv"`
	Price    float64 `json:"price"`
	IsVN100  bool    `json:"isVn100"`
}

// WhitelistSnapshot is kept for backward compatibility.
type WhitelistSnapshot struct {
	Entries   []WhitelistEntry `json:"entries"`
	Total     int              `json:"total"`
	Config    LiquidityConfig  `json:"config"`
	UpdatedAt time.Time        `json:"updatedAt"`
}
