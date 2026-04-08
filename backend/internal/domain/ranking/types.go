package ranking

// FactorGroupName identifies a factor group for ranking.
type FactorGroupName string

const (
	FactorQuality    FactorGroupName = "quality"
	FactorValue      FactorGroupName = "value"
	FactorGrowth     FactorGroupName = "growth"
	FactorMomentum   FactorGroupName = "momentum"
	FactorVolatility FactorGroupName = "volatility"
)

// FactorGroup defines a group of factors with their weights.
type FactorGroup struct {
	Name    FactorGroupName `json:"name"`
	Weight  float64         `json:"weight"`
	Factors []Factor        `json:"factors"`
}

// Factor represents a single ranking factor within a group.
type Factor struct {
	Code   string  `json:"code"` // e.g. "roe", "pe", "momentum_3m"
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

// RankingConfig defines the configuration for a stock ranking run.
type RankingConfig struct {
	Universe         string        `json:"universe"` // VN30, VN100, HOSE, HNX, UPCOM, custom
	CustomSymbols    []string      `json:"customSymbols,omitempty"`
	FactorGroups     []FactorGroup `json:"factorGroups"`
	RebalanceFreq    string        `json:"rebalanceFreq"` // weekly, monthly, quarterly
	TopN             int           `json:"topN"`
	MinLiquidityTier int           `json:"minLiquidityTier,omitempty"` // 1, 2, or 3
}

// RankingResult contains the output of a ranking computation.
type RankingResult struct {
	Config      RankingConfig `json:"config"`
	Rankings    []RankedStock `json:"rankings"`
	TotalStocks int           `json:"totalStocks"`
	ComputedAt  string        `json:"computedAt"`
}

// RankedStock represents a single stock in the ranking output.
type RankedStock struct {
	Rank           int                `json:"rank"`
	Symbol         string             `json:"symbol"`
	CompositeScore float64            `json:"compositeScore"`
	FactorScores   map[string]float64 `json:"factorScores"`
}
