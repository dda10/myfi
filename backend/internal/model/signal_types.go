package model

import "time"

// --- Signal Engine Types ---
// Systematic stock signal generation with composite scoring.

// SignalDirection indicates the trade direction.
type SignalDirection string

const (
	SignalLong  SignalDirection = "long"
	SignalShort SignalDirection = "short"
)

// SignalStrength categorizes signal confidence.
type SignalStrength string

const (
	StrengthStrong   SignalStrength = "strong"
	StrengthModerate SignalStrength = "moderate"
	StrengthWeak     SignalStrength = "weak"
)

// FactorScores holds individual factor scores (0-100 scale).
type FactorScores struct {
	Momentum    float64 `json:"momentum"`    // RSI, MACD, price vs MA
	Trend       float64 `json:"trend"`       // ADX, Supertrend direction
	Volume      float64 `json:"volume"`      // OBV trend, volume spike
	Fundamental float64 `json:"fundamental"` // P/E vs sector, ROE, growth
	Sector      float64 `json:"sector"`      // Sector trend, rotation
}

// TradingSignal represents a systematic buy/sell signal with entry/exit levels.
type TradingSignal struct {
	Symbol         string          `json:"symbol"`
	Direction      SignalDirection `json:"direction"`
	Strength       SignalStrength  `json:"strength"`
	CompositeScore float64         `json:"compositeScore"` // 0-100
	Factors        FactorScores    `json:"factors"`

	// Price levels
	CurrentPrice float64 `json:"currentPrice"`
	EntryPrice   float64 `json:"entryPrice"`
	StopLoss     float64 `json:"stopLoss"`
	TakeProfit   float64 `json:"takeProfit"`
	RiskReward   float64 `json:"riskReward"` // TP distance / SL distance

	// Context
	Exchange   string    `json:"exchange"`
	Sector     ICBSector `json:"sector"`
	SectorName string    `json:"sectorName"`
	Reasoning  []string  `json:"reasoning"` // Key factors that triggered signal

	GeneratedAt time.Time `json:"generatedAt"`
}

// SignalConfig holds configurable parameters for signal generation.
type SignalConfig struct {
	// Factor weights (must sum to 1.0)
	MomentumWeight    float64 `json:"momentumWeight"`
	TrendWeight       float64 `json:"trendWeight"`
	VolumeWeight      float64 `json:"volumeWeight"`
	FundamentalWeight float64 `json:"fundamentalWeight"`
	SectorWeight      float64 `json:"sectorWeight"`

	// Thresholds
	MinCompositeScore float64 `json:"minCompositeScore"` // Minimum score to generate signal
	TopN              int     `json:"topN"`              // Max signals to return

	// Risk parameters
	ATRMultiplierSL float64 `json:"atrMultiplierSL"` // ATR multiplier for stop-loss
	RiskRewardRatio float64 `json:"riskRewardRatio"` // Target risk-reward ratio

	// Indicator periods
	RSIPeriod      int `json:"rsiPeriod"`
	MACDFast       int `json:"macdFast"`
	MACDSlow       int `json:"macdSlow"`
	MACDSignal     int `json:"macdSignal"`
	ADXPeriod      int `json:"adxPeriod"`
	ATRPeriod      int `json:"atrPeriod"`
	OBVSMAPeriod   int `json:"obvSmaPeriod"`
	PriceSMAPeriod int `json:"priceSMAPeriod"`
	VolumeLookback int `json:"volumeLookback"`
}

// DefaultSignalConfig returns sensible defaults.
func DefaultSignalConfig() SignalConfig {
	return SignalConfig{
		MomentumWeight:    0.30,
		TrendWeight:       0.20,
		VolumeWeight:      0.20,
		FundamentalWeight: 0.20,
		SectorWeight:      0.10,

		MinCompositeScore: 60,
		TopN:              10,

		ATRMultiplierSL: 2.0,
		RiskRewardRatio: 2.0,

		RSIPeriod:      14,
		MACDFast:       12,
		MACDSlow:       26,
		MACDSignal:     9,
		ADXPeriod:      14,
		ATRPeriod:      14,
		OBVSMAPeriod:   20,
		PriceSMAPeriod: 20,
		VolumeLookback: 20,
	}
}

// SignalScanResult holds the output of a full market scan.
type SignalScanResult struct {
	Signals      []TradingSignal `json:"signals"`
	TotalScanned int             `json:"totalScanned"`
	PassedFilter int             `json:"passedFilter"`
	GeneratedAt  time.Time       `json:"generatedAt"`
	Config       SignalConfig    `json:"config"`
}
