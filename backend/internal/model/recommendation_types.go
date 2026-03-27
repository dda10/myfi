package model

import "time"

// --- Recommendation Audit Log Types ---
// Tracks AI recommendations and their outcomes for accuracy measurement.

// RecommendationAction represents the action type in a recommendation.
type RecommendationAction string

const (
	ActionBuy  RecommendationAction = "buy"
	ActionSell RecommendationAction = "sell"
	ActionHold RecommendationAction = "hold"
)

// RecommendationRecord stores a single AI recommendation with its context and outcomes.
type RecommendationRecord struct {
	ID              int64                `json:"id"`
	Symbol          string               `json:"symbol"`
	Action          RecommendationAction `json:"action"`
	PositionSize    float64              `json:"positionSize"`    // % of NAV suggested
	RiskAssessment  string               `json:"riskAssessment"`  // low/medium/high
	ConfidenceScore int                  `json:"confidenceScore"` // 0-100 from analysis
	Reasoning       string               `json:"reasoning"`
	PriceAtSignal   float64              `json:"priceAtSignal"`
	CreatedAt       time.Time            `json:"createdAt"`

	// Outcomes - filled in by periodic updates
	Price1Day   *float64 `json:"price1Day,omitempty"`
	Price7Day   *float64 `json:"price7Day,omitempty"`
	Price14Day  *float64 `json:"price14Day,omitempty"`
	Price30Day  *float64 `json:"price30Day,omitempty"`
	Return1Day  *float64 `json:"return1Day,omitempty"` // % change
	Return7Day  *float64 `json:"return7Day,omitempty"`
	Return14Day *float64 `json:"return14Day,omitempty"`
	Return30Day *float64 `json:"return30Day,omitempty"`
}

// RecommendationAccuracy holds aggregated accuracy metrics for recommendations.
type RecommendationAccuracy struct {
	Action          RecommendationAction `json:"action"`
	TotalCount      int                  `json:"totalCount"`
	WinCount1Day    int                  `json:"winCount1Day"`
	WinCount7Day    int                  `json:"winCount7Day"`
	WinCount14Day   int                  `json:"winCount14Day"`
	WinCount30Day   int                  `json:"winCount30Day"`
	WinRate1Day     float64              `json:"winRate1Day"`
	WinRate7Day     float64              `json:"winRate7Day"`
	WinRate14Day    float64              `json:"winRate14Day"`
	WinRate30Day    float64              `json:"winRate30Day"`
	AvgReturn1Day   float64              `json:"avgReturn1Day"`
	AvgReturn7Day   float64              `json:"avgReturn7Day"`
	AvgReturn14Day  float64              `json:"avgReturn14Day"`
	AvgReturn30Day  float64              `json:"avgReturn30Day"`
	AvgConfidence   float64              `json:"avgConfidence"`
	HighConfWinRate float64              `json:"highConfWinRate"` // win rate for confidence >= 70
	MedConfWinRate  float64              `json:"medConfWinRate"`  // win rate for 40 <= confidence < 70
	LowConfWinRate  float64              `json:"lowConfWinRate"`  // win rate for confidence < 40
}

// RecommendationFilter specifies criteria for querying recommendations.
type RecommendationFilter struct {
	Symbol        string               `json:"symbol,omitempty"`
	Action        RecommendationAction `json:"action,omitempty"`
	MinConfidence int                  `json:"minConfidence,omitempty"`
	StartDate     *time.Time           `json:"startDate,omitempty"`
	EndDate       *time.Time           `json:"endDate,omitempty"`
	Limit         int                  `json:"limit,omitempty"`
}

// RecommendationSummary provides a high-level view of recommendation performance.
type RecommendationSummary struct {
	TotalRecommendations  int                      `json:"totalRecommendations"`
	ByAction              []RecommendationAccuracy `json:"byAction"`
	BestPerformingSymbol  string                   `json:"bestPerformingSymbol"`
	WorstPerformingSymbol string                   `json:"worstPerformingSymbol"`
	OverallWinRate7Day    float64                  `json:"overallWinRate7Day"`
	OverallAvgReturn7Day  float64                  `json:"overallAvgReturn7Day"`
}
