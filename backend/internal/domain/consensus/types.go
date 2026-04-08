package consensus

import "time"

// SourceType identifies the origin of a sentiment signal.
type SourceType string

const (
	SourceNews        SourceType = "news"         // financial news articles
	SourceAnalyst     SourceType = "analyst"      // analyst reports & recommendations
	SourceSocialMedia SourceType = "social_media" // social media posts (Facebook groups, X/Twitter)
	SourceForum       SourceType = "forum"        // investor forums (f319, stockbiz, etc.)
)

// ConsensusScore is the composite market opinion for a symbol.
type ConsensusScore struct {
	Symbol          string            `json:"symbol"`
	CompositeScore  float64           `json:"compositeScore"` // -1.0 (very bearish) to +1.0 (very bullish)
	Confidence      float64           `json:"confidence"`     // 0.0–1.0 based on data volume & agreement
	SignalStrength  string            `json:"signalStrength"` // "strong_buy", "buy", "neutral", "sell", "strong_sell"
	SourceBreakdown []SourceSentiment `json:"sourceBreakdown"`
	TotalSignals    int               `json:"totalSignals"`
	Period          string            `json:"period"` // "1d", "7d", "30d"
	ComputedAt      time.Time         `json:"computedAt"`
	PreviousScore   *float64          `json:"previousScore,omitempty"` // for trend comparison
	ScoreChange     *float64          `json:"scoreChange,omitempty"`   // delta from previous period
}

// SourceSentiment is the sentiment breakdown from a single source type.
type SourceSentiment struct {
	Source      SourceType `json:"source"`
	Score       float64    `json:"score"` // -1.0 to +1.0
	SignalCount int        `json:"signalCount"`
	Weight      float64    `json:"weight"`      // contribution weight to composite
	TopPositive []string   `json:"topPositive"` // top positive themes
	TopNegative []string   `json:"topNegative"` // top negative themes
}

// Divergence detects when different source types disagree significantly.
type Divergence struct {
	Symbol       string     `json:"symbol"`
	SourceA      SourceType `json:"sourceA"`
	SourceB      SourceType `json:"sourceB"`
	ScoreA       float64    `json:"scoreA"`
	ScoreB       float64    `json:"scoreB"`
	Gap          float64    `json:"gap"`          // absolute difference
	Significance string     `json:"significance"` // "low", "medium", "high"
	DetectedAt   time.Time  `json:"detectedAt"`
	Description  string     `json:"description"` // human-readable explanation
}

// MarketMood is the overall market-wide sentiment (not symbol-specific).
type MarketMood struct {
	OverallScore float64           `json:"overallScore"` // -1.0 to +1.0
	Label        string            `json:"label"`        // "fear", "cautious", "neutral", "optimistic", "greed"
	TopBullish   []SymbolSentiment `json:"topBullish"`   // top 5 most bullish symbols
	TopBearish   []SymbolSentiment `json:"topBearish"`   // top 5 most bearish symbols
	SectorMood   []SectorSentiment `json:"sectorMood"`
	TotalSignals int               `json:"totalSignals"`
	ComputedAt   time.Time         `json:"computedAt"`
}

// SymbolSentiment is a lightweight sentiment entry for a single symbol.
type SymbolSentiment struct {
	Symbol  string  `json:"symbol"`
	Score   float64 `json:"score"`
	Signals int     `json:"signals"`
}

// SectorSentiment is sentiment aggregated at the sector level.
type SectorSentiment struct {
	Sector  string  `json:"sector"`
	Score   float64 `json:"score"`
	Signals int     `json:"signals"`
}

// SignalRecord is a single sentiment signal from any source, stored in DB.
type SignalRecord struct {
	ID         int64      `json:"id"`
	Symbol     string     `json:"symbol"`
	Source     SourceType `json:"source"`
	Score      float64    `json:"score"`      // -1.0 to +1.0
	Confidence float64    `json:"confidence"` // 0.0–1.0
	Text       string     `json:"text"`       // source text snippet
	URL        string     `json:"url"`
	Topics     []string   `json:"topics"`
	CreatedAt  time.Time  `json:"createdAt"`
}
