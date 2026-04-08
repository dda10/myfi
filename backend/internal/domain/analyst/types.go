package analyst

import "time"

// AnalystReport represents a single analyst report from a brokerage.
type AnalystReport struct {
	ID             int64     `json:"id"`
	Symbol         string    `json:"symbol"`
	Analyst        string    `json:"analyst"`
	Brokerage      string    `json:"brokerage"`
	Recommendation string    `json:"recommendation"` // buy, sell, hold, outperform, underperform
	TargetPrice    float64   `json:"targetPrice"`
	PublishedAt    time.Time `json:"publishedAt"`
	Accuracy1M     *float64  `json:"accuracy1m,omitempty"`
	Accuracy3M     *float64  `json:"accuracy3m,omitempty"`
	Accuracy6M     *float64  `json:"accuracy6m,omitempty"`
}

// ConsensusRecommendation aggregates analyst reports for a symbol.
type ConsensusRecommendation struct {
	Symbol             string  `json:"symbol"`
	ConsensusAction    string  `json:"consensusAction"` // buy, sell, hold
	AverageTargetPrice float64 `json:"averageTargetPrice"`
	HighTargetPrice    float64 `json:"highTargetPrice"`
	LowTargetPrice     float64 `json:"lowTargetPrice"`
	TotalAnalysts      int     `json:"totalAnalysts"`
	BuyCount           int     `json:"buyCount"`
	HoldCount          int     `json:"holdCount"`
	SellCount          int     `json:"sellCount"`
	AverageAccuracy    float64 `json:"averageAccuracy"`
}
