package model

import "time"

// AgentMessage is the structured message schema passed between agents in the
// Multi_Agent_System. Each message carries the originating agent name, a
// payload type discriminator, and the payload itself.
// Requirement 8.4: structured data between agents.
type AgentMessage struct {
	AgentName   string                 `json:"agentName"`
	PayloadType string                 `json:"payloadType"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
}

// QueryIntent represents the parsed intent from a user query.
type QueryIntent struct {
	Symbols    []string `json:"symbols"`
	AssetTypes []string `json:"assetTypes"`
	Intent     string   `json:"intent"` // "analysis", "news", "recommendation", "price", "general"
	RawQuery   string   `json:"rawQuery"`
}

// PriceAgentResponse is the structured output of the Price_Agent.
type PriceAgentResponse struct {
	Symbol         string     `json:"symbol"`
	CurrentPrice   float64    `json:"currentPrice"`
	Change         float64    `json:"change"`
	ChangePercent  float64    `json:"changePercent"`
	Volume         int64      `json:"volume"`
	Source         string     `json:"source"`
	HistoricalData []OHLCVBar `json:"historicalData,omitempty"`
}

// SectorContextData holds sector-relative context for analysis.
type SectorContextData struct {
	SectorName          string      `json:"sectorName"`
	SectorTrend         SectorTrend `json:"sectorTrend"`
	StockVsSectorPerf   string      `json:"stockVsSectorPerf"`   // outperforming / underperforming
	SectorRotationPhase string      `json:"sectorRotationPhase"` // capital flowing in / out / stable
}

// AnalysisAgentResponse is the structured output of the Analysis_Agent.
type AnalysisAgentResponse struct {
	Symbol           string             `json:"symbol"`
	TrendAssessment  string             `json:"trendAssessment"`
	IndicatorSignals map[string]string  `json:"indicatorSignals"` // indicator -> bullish/bearish/neutral
	KeyPriceLevels   map[string]float64 `json:"keyPriceLevels"`   // support/resistance
	SectorContext    SectorContextData  `json:"sectorContext"`
	CompositeSignal  string             `json:"compositeSignal"` // strongly bullish / bullish / neutral / bearish / strongly bearish
	BullishCount     int                `json:"bullishCount"`
	BearishCount     int                `json:"bearishCount"`
	ConfidenceScore  int                `json:"confidenceScore"` // 0-100
}

// NewsArticle represents a single news article returned by the News_Agent.
type NewsArticle struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	Snippet     string    `json:"snippet"`
}

// NewsAgentResponse is the structured output of the News_Agent.
type NewsAgentResponse struct {
	Articles []NewsArticle `json:"articles"`
	Summary  string        `json:"summary"`
}

// AssetRecommendation is a single buy/sell/hold recommendation for an asset.
type AssetRecommendation struct {
	Symbol         string  `json:"symbol"`
	Action         string  `json:"action"`         // buy / sell / hold
	PositionSize   float64 `json:"positionSize"`   // percentage of NAV
	RiskAssessment string  `json:"riskAssessment"` // low / medium / high
	Reasoning      string  `json:"reasoning"`
}

// SupervisorRecommendation is the final aggregated output of the Supervisor_Agent.
type SupervisorRecommendation struct {
	Summary                 string                `json:"summary"`
	AssetRecommendations    []AssetRecommendation `json:"assetRecommendations"`
	PortfolioSuggestions    []string              `json:"portfolioSuggestions"`
	IdentifiedOpportunities []string              `json:"identifiedOpportunities"`
	SectorContext           string                `json:"sectorContext"`
	KnowledgeBaseInsights   []string              `json:"knowledgeBaseInsights"`
	MissingSources          []string              `json:"missingSources,omitempty"`
}

// AgentResult wraps the output of a single sub-agent execution, including
// any error that occurred. Used by the orchestrator to collect parallel results.
type AgentResult struct {
	AgentName string
	Message   *AgentMessage
	Err       error
}

// LLMConfig holds the configuration needed to instantiate an LLM provider.
type LLMConfig struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	ApiKey       string `json:"apiKey"`
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
}

// PatternType represents the type of market pattern detected by the Pattern_Detector.
// Requirements: 12.3, 12.4, 12.5
type PatternType string

const (
	// Accumulation pattern: price consolidation within 5% for 10+ days,
	// volume >1.5x 20-day avg, institutional buying.
	PatternAccumulation PatternType = "accumulation"

	// Distribution pattern: price near highs, volume on down days,
	// institutional selling.
	PatternDistribution PatternType = "distribution"

	// Breakout pattern: price above resistance, volume >2x 20-day avg.
	PatternBreakout PatternType = "breakout"
)

// PatternObservation represents a detected market pattern with supporting data.
// Requirements: 12.6, 13.1
type PatternObservation struct {
	ID               int64       `json:"id"`
	Symbol           string      `json:"symbol"`
	PatternType      PatternType `json:"patternType"`
	DetectionDate    time.Time   `json:"detectionDate"`
	ConfidenceScore  int         `json:"confidenceScore"` // 0-100
	PriceAtDetection float64     `json:"priceAtDetection"`
	SupportingData   string      `json:"supportingData"` // JSON blob with pattern details
	Outcome1Day      *float64    `json:"outcome1Day,omitempty"`
	Outcome7Day      *float64    `json:"outcome7Day,omitempty"`
	Outcome14Day     *float64    `json:"outcome14Day,omitempty"`
	Outcome30Day     *float64    `json:"outcome30Day,omitempty"`
}

// AccumulationData holds supporting data for accumulation pattern detection.
type AccumulationData struct {
	ConsolidationDays   int     `json:"consolidationDays"`
	PriceRangePercent   float64 `json:"priceRangePercent"`
	VolumeRatio         float64 `json:"volumeRatio"` // Current volume / 20-day avg
	AvgVolume20Day      float64 `json:"avgVolume20Day"`
	CurrentVolume       float64 `json:"currentVolume"`
	InstitutionalBuying bool    `json:"institutionalBuying"`
	PriceHigh           float64 `json:"priceHigh"`
	PriceLow            float64 `json:"priceLow"`
}

// DistributionData holds supporting data for distribution pattern detection.
type DistributionData struct {
	PriceNearHighPercent float64 `json:"priceNearHighPercent"` // How close to recent high (%)
	RecentHigh           float64 `json:"recentHigh"`
	CurrentPrice         float64 `json:"currentPrice"`
	DownDayVolumeRatio   float64 `json:"downDayVolumeRatio"` // Volume on down days vs avg
	DownDaysCount        int     `json:"downDaysCount"`
	InstitutionalSelling bool    `json:"institutionalSelling"`
}

// BreakoutData holds supporting data for breakout pattern detection.
type BreakoutData struct {
	ResistanceLevel   float64 `json:"resistanceLevel"`
	BreakoutPrice     float64 `json:"breakoutPrice"`
	BreakoutPercent   float64 `json:"breakoutPercent"` // % above resistance
	VolumeRatio       float64 `json:"volumeRatio"`     // Current volume / 20-day avg
	AvgVolume20Day    float64 `json:"avgVolume20Day"`
	CurrentVolume     float64 `json:"currentVolume"`
	PriorConsolidDays int     `json:"priorConsolidDays"`
}

// ---------------------------------------------------------------------------
// Alert types for Alert_Service
// Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7
// ---------------------------------------------------------------------------

// Alert represents a proactive notification for a detected market pattern.
// Requirements: 14.2, 14.3, 14.4
type Alert struct {
	ID                 int64       `json:"id"`
	UserID             int64       `json:"userId"`
	Symbol             string      `json:"symbol"`
	PatternType        PatternType `json:"patternType"`
	ConfidenceScore    int         `json:"confidenceScore"` // 0-100
	Explanation        string      `json:"explanation"`
	DetectionTimestamp time.Time   `json:"detectionTimestamp"`
	ChartLink          string      `json:"chartLink"` // Link to view the stock's chart
	Viewed             bool        `json:"viewed"`
	Expired            bool        `json:"expired"`
	CreatedAt          time.Time   `json:"createdAt"`
}

// AlertPreferences holds user-configurable alert settings.
// Requirement 14.5: Support user alert preferences
type AlertPreferences struct {
	ID             int64         `json:"id"`
	UserID         int64         `json:"userId"`
	MinConfidence  int           `json:"minConfidence"`  // Minimum confidence threshold (default 60)
	PatternTypes   []PatternType `json:"patternTypes"`   // Specific pattern types to monitor (empty = all)
	IncludeSymbols []string      `json:"includeSymbols"` // Symbols to include (empty = all)
	ExcludeSymbols []string      `json:"excludeSymbols"` // Symbols to exclude
	UpdatedAt      time.Time     `json:"updatedAt"`
}
