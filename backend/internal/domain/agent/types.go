package agent

import "time"

// AgentResponse is the structured response from the multi-agent AI system.
type AgentResponse struct {
	TechnicalAnalysis        *TechnicalAnalysis        `json:"technicalAnalysis,omitempty"`
	NewsAnalysis             *NewsAnalysis             `json:"newsAnalysis,omitempty"`
	InvestmentRecommendation *InvestmentRecommendation `json:"investmentRecommendation,omitempty"`
	TradingStrategy          *TradingStrategy          `json:"tradingStrategy,omitempty"`
	Citations                []Citation                `json:"citations,omitempty"`
	MissingAgents            []string                  `json:"missingAgents,omitempty"`
	ProcessingTimeMs         int64                     `json:"processingTimeMs"`
}

// TechnicalAnalysis contains the output from the Technical Analyst agent.
type TechnicalAnalysis struct {
	Symbol           string             `json:"symbol"`
	CompositeSignal  string             `json:"compositeSignal"` // strongly_bullish, bullish, neutral, bearish, strongly_bearish
	Indicators       map[string]float64 `json:"indicators"`
	SupportLevels    []float64          `json:"supportLevels"`
	ResistanceLevels []float64          `json:"resistanceLevels"`
	Patterns         []string           `json:"patterns"`
	SmartMoneyFlow   string             `json:"smartMoneyFlow"` // strong_inflow, moderate_inflow, neutral, moderate_outflow, strong_outflow
	Summary          string             `json:"summary"`
}

// NewsAnalysis contains the output from the News Analyst agent.
type NewsAnalysis struct {
	Symbol      string        `json:"symbol"`
	Sentiment   string        `json:"sentiment"` // positive, negative, neutral
	Confidence  float64       `json:"confidence"`
	Catalysts   []string      `json:"catalysts"`
	RiskFactors []string      `json:"riskFactors"`
	Articles    []NewsArticle `json:"articles"`
	Summary     string        `json:"summary"`
}

// NewsArticle represents a single news article reference.
type NewsArticle struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Source    string    `json:"source"`
	Sentiment string    `json:"sentiment"`
	Published time.Time `json:"published"`
}

// InvestmentRecommendation contains the output from the Investment Advisor agent.
type InvestmentRecommendation struct {
	Symbol          string  `json:"symbol"`
	Action          string  `json:"action"` // buy, sell, hold
	TargetPrice     float64 `json:"targetPrice"`
	UpsidePercent   float64 `json:"upsidePercent"`
	ConfidenceScore int     `json:"confidenceScore"` // 0-100
	RiskLevel       string  `json:"riskLevel"`       // low, medium, high
	Reasoning       string  `json:"reasoning"`
}

// TradingStrategy contains the output from the Strategy Builder agent.
type TradingStrategy struct {
	Symbol       string  `json:"symbol"`
	Direction    string  `json:"direction"` // long, short
	EntryPrice   float64 `json:"entryPrice"`
	StopLoss     float64 `json:"stopLoss"`
	TakeProfit   float64 `json:"takeProfit"`
	RiskReward   float64 `json:"riskReward"`
	PositionSize float64 `json:"positionSize"` // percentage of NAV
	Confidence   int     `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

// Citation links a claim in the agent response to its source data point.
type Citation struct {
	ClaimText  string `json:"claimText"`
	SourceType string `json:"sourceType"` // indicator, news, financial, price
	SourceRef  string `json:"sourceRef"`
	Value      string `json:"value,omitempty"`
}

// ChatContext holds the context for an AI chat session.
type ChatContext struct {
	UserID    int64         `json:"userId"`
	Symbol    string        `json:"symbol,omitempty"`
	SessionID string        `json:"sessionId"`
	History   []ChatMessage `json:"history,omitempty"`
}

// ChatMessage represents a single message in a chat session.
type ChatMessage struct {
	Role      string    `json:"role"` // user, assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatResponse is the response from the AI chat endpoint.
type ChatResponse struct {
	Message   string         `json:"message"`
	Agent     *AgentResponse `json:"agent,omitempty"`
	SessionID string         `json:"sessionId"`
}
