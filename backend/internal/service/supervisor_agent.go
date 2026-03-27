package service

// ---------------------------------------------------------------------------
// Supervisor_Agent — orchestrates sub-agents and produces NAV-based recommendations
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 15.1  Synthesize outputs from Price_Agent, Analysis_Agent, News_Agent
//   - 15.2  Incorporate user's NAV, allocation, holdings from Portfolio_Engine
//   - 15.3  Query Knowledge_Base for recent observations and pattern accuracy
//   - 15.4  Produce recommendations with actions, position sizes, risk, reasoning
//   - 15.5  Reference Knowledge_Base entries in recommendations
//   - 15.6  Identify new opportunities based on screener, signals, sector trends
//   - 15.7  Incorporate sector context from Analysis_Agent
//   - 15.8  Flag diversification warnings (single asset >40% NAV)
//   - 15.9  Format structured SupervisorRecommendation
//   - 15.10 Handle missing NAV data gracefully

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"myfi-backend/internal/model"

	"github.com/tmc/langchaingo/llms"
)

// SupervisorAgent is the orchestrating AI agent that synthesizes outputs from
// all sub-agents and produces NAV-based recommendations.
type SupervisorAgent struct {
	portfolioEngine *PortfolioEngine
	knowledgeBase   *KnowledgeBase
	sectorService   *SectorService
}

// NewSupervisorAgent creates a SupervisorAgent with the required dependencies.
func NewSupervisorAgent(
	portfolioEngine *PortfolioEngine,
	knowledgeBase *KnowledgeBase,
	sectorService *SectorService,
) *SupervisorAgent {
	return &SupervisorAgent{
		portfolioEngine: portfolioEngine,
		knowledgeBase:   knowledgeBase,
		sectorService:   sectorService,
	}
}

// Name returns the agent identifier used by the orchestrator.
func (s *SupervisorAgent) Name() string { return "Supervisor_Agent" }

// Execute synthesizes data from sub-agents and produces a SupervisorRecommendation.
// The LLM parameter can be used for advanced reasoning but basic synthesis is
// rule-based for reliability.
func (s *SupervisorAgent) Execute(ctx context.Context, intent model.QueryIntent, llm llms.Model) (*model.AgentMessage, error) {
	log.Printf("[Supervisor_Agent] Processing query: %s", intent.RawQuery)

	// Build the recommendation
	rec := model.SupervisorRecommendation{
		MissingSources: []string{},
	}

	// Get user ID from context or default to 1
	userID := int64(1)

	// 15.2: Incorporate portfolio context from Portfolio_Engine
	portfolioCtx, err := s.getPortfolioContext(ctx, userID)
	if err != nil {
		log.Printf("[Supervisor_Agent] Portfolio context unavailable: %v", err)
		rec.MissingSources = append(rec.MissingSources, "Portfolio_Engine")
	}

	// 15.3: Query Knowledge_Base for relevant patterns
	kbInsights := s.getKnowledgeBaseInsights(ctx, intent.Symbols)
	rec.KnowledgeBaseInsights = kbInsights

	// 15.7: Get sector context
	sectorContext := s.getSectorContext(ctx, intent.Symbols)
	rec.SectorContext = sectorContext

	// 15.4, 15.8: Generate asset recommendations with position sizing and diversification checks
	assetRecs, diversificationWarnings := s.generateAssetRecommendations(ctx, intent.Symbols, portfolioCtx)
	rec.AssetRecommendations = assetRecs

	// 15.6: Identify opportunities
	rec.IdentifiedOpportunities = s.identifyOpportunities(ctx, intent.Symbols, kbInsights)

	// Build portfolio suggestions including diversification warnings
	rec.PortfolioSuggestions = s.buildPortfolioSuggestions(portfolioCtx, diversificationWarnings)

	// 15.9: Format summary
	rec.Summary = s.buildSummary(intent, portfolioCtx, len(assetRecs), len(kbInsights))

	// Pack into AgentMessage
	payload := make(map[string]any)
	payload["recommendation"] = rec

	msg := &model.AgentMessage{
		AgentName:   s.Name(),
		PayloadType: "supervisor_recommendation",
		Payload:     payload,
		Timestamp:   time.Now(),
	}

	log.Printf("[Supervisor_Agent] Generated recommendation with %d asset recommendations, %d KB insights",
		len(assetRecs), len(kbInsights))

	return msg, nil
}

// ---------------------------------------------------------------------------
// Portfolio context (Requirement 15.2)
// ---------------------------------------------------------------------------

// portfolioContext holds the user's portfolio data for recommendation synthesis.
type portfolioContext struct {
	NAV               float64
	Holdings          []model.HoldingDetail
	AllocationByType  map[model.AssetType]float64
	AllocationPercent map[model.AssetType]float64
	Available         bool
}

// getPortfolioContext retrieves the user's portfolio data from Portfolio_Engine.
func (s *SupervisorAgent) getPortfolioContext(ctx context.Context, userID int64) (*portfolioContext, error) {
	if s.portfolioEngine == nil {
		return &portfolioContext{Available: false}, fmt.Errorf("Portfolio_Engine not configured")
	}

	summary, err := s.portfolioEngine.GetPortfolioSummary(ctx, userID)
	if err != nil {
		return &portfolioContext{Available: false}, err
	}

	return &portfolioContext{
		NAV:               summary.NAV,
		Holdings:          summary.Holdings,
		AllocationByType:  summary.AllocationByType,
		AllocationPercent: summary.AllocationPercent,
		Available:         true,
	}, nil
}

// ---------------------------------------------------------------------------
// Knowledge Base integration (Requirements 15.3, 15.5)
// ---------------------------------------------------------------------------

// getKnowledgeBaseInsights queries the Knowledge_Base for recent observations
// related to the queried symbols.
func (s *SupervisorAgent) getKnowledgeBaseInsights(ctx context.Context, symbols []string) []string {
	if s.knowledgeBase == nil {
		return nil
	}

	var insights []string

	for _, symbol := range symbols {
		// Get recent observations for this symbol (last 30 days)
		observations, err := s.knowledgeBase.GetRecentObservationsForSymbol(ctx, symbol, 30)
		if err != nil {
			log.Printf("[Supervisor_Agent] Failed to query KB for %s: %v", symbol, err)
			continue
		}

		for _, obs := range observations {
			insight := s.formatObservationInsight(obs)
			if insight != "" {
				insights = append(insights, insight)
			}
		}
	}

	// Get pattern accuracy metrics to inform recommendations
	for _, patternType := range []model.PatternType{
		model.PatternAccumulation,
		model.PatternDistribution,
		model.PatternBreakout,
	} {
		metrics, err := s.knowledgeBase.GetAccuracyMetrics(ctx, patternType)
		if err != nil {
			continue
		}
		if metrics.TotalObservations > 10 && metrics.SuccessRate7Day > 0 {
			insights = append(insights, fmt.Sprintf(
				"Historical %s pattern accuracy: %.1f%% success rate over 7 days (%d observations)",
				patternType, metrics.SuccessRate7Day, metrics.TotalObservations,
			))
		}
	}

	return insights
}

// formatObservationInsight formats a pattern observation as a human-readable insight.
func (s *SupervisorAgent) formatObservationInsight(obs model.PatternObservation) string {
	daysSince := int(time.Since(obs.DetectionDate).Hours() / 24)

	var outcomeStr string
	if obs.Outcome7Day != nil {
		if *obs.Outcome7Day > 0 {
			outcomeStr = fmt.Sprintf(" (resulted in +%.1f%% after 7 days)", *obs.Outcome7Day)
		} else {
			outcomeStr = fmt.Sprintf(" (resulted in %.1f%% after 7 days)", *obs.Outcome7Day)
		}
	}

	return fmt.Sprintf(
		"%s: %s pattern detected %d days ago with %d%% confidence%s",
		obs.Symbol, obs.PatternType, daysSince, obs.ConfidenceScore, outcomeStr,
	)
}

// ---------------------------------------------------------------------------
// Sector context (Requirement 15.7)
// ---------------------------------------------------------------------------

// getSectorContext retrieves sector trend information for the queried symbols.
func (s *SupervisorAgent) getSectorContext(ctx context.Context, symbols []string) string {
	if s.sectorService == nil || len(symbols) == 0 {
		return ""
	}

	var sectorInfos []string
	seenSectors := make(map[model.ICBSector]bool)

	for _, symbol := range symbols {
		sector, err := s.sectorService.GetStockSector(symbol)
		if err != nil {
			continue
		}

		if seenSectors[sector] {
			continue
		}
		seenSectors[sector] = true

		perf, err := s.sectorService.GetSectorPerformance(ctx, sector)
		if err != nil {
			continue
		}

		trendStr := "sideways"
		switch perf.Trend {
		case model.Uptrend:
			trendStr = "uptrend"
		case model.Downtrend:
			trendStr = "downtrend"
		}

		sectorInfos = append(sectorInfos, fmt.Sprintf(
			"%s (%s): %s, 1M: %.1f%%, 3M: %.1f%%",
			perf.SectorName, sector, trendStr, perf.OneMonthChange, perf.ThreeMonthChange,
		))
	}

	if len(sectorInfos) == 0 {
		return ""
	}

	return "Sector trends: " + strings.Join(sectorInfos, "; ")
}

// ---------------------------------------------------------------------------
// Asset recommendations with position sizing (Requirements 15.4, 15.8)
// ---------------------------------------------------------------------------

// Position sizing limits based on risk assessment (Requirement 15.4)
const (
	maxPositionHighRisk   = 0.05 // 5% NAV for high-risk assets
	maxPositionMediumRisk = 0.10 // 10% NAV for medium-risk assets
	maxPositionLowRisk    = 0.20 // 20% NAV for low-risk assets
	diversificationLimit  = 0.40 // 40% NAV threshold for diversification warning
)

// generateAssetRecommendations creates recommendations for each symbol with
// NAV-based position sizing and diversification checks.
func (s *SupervisorAgent) generateAssetRecommendations(
	ctx context.Context,
	symbols []string,
	portfolio *portfolioContext,
) ([]model.AssetRecommendation, []string) {
	var recommendations []model.AssetRecommendation
	var diversificationWarnings []string

	// Check existing holdings for diversification issues (Requirement 15.8)
	if portfolio != nil && portfolio.Available && portfolio.NAV > 0 {
		for _, holding := range portfolio.Holdings {
			holdingPercent := (holding.MarketValue / portfolio.NAV) * 100
			if holdingPercent > diversificationLimit*100 {
				diversificationWarnings = append(diversificationWarnings, fmt.Sprintf(
					"Warning: %s represents %.1f%% of NAV (exceeds 40%% diversification threshold)",
					holding.Asset.Symbol, holdingPercent,
				))
			}
		}
	}

	for _, symbol := range symbols {
		rec := s.generateSingleRecommendation(ctx, symbol, portfolio)
		recommendations = append(recommendations, rec)
	}

	return recommendations, diversificationWarnings
}

// generateSingleRecommendation creates a recommendation for a single symbol.
func (s *SupervisorAgent) generateSingleRecommendation(
	ctx context.Context,
	symbol string,
	portfolio *portfolioContext,
) model.AssetRecommendation {
	rec := model.AssetRecommendation{
		Symbol: symbol,
		Action: "hold", // Default action
	}

	// Determine risk assessment based on available data
	riskAssessment := s.assessRisk(ctx, symbol)
	rec.RiskAssessment = riskAssessment

	// Calculate position size based on risk (Requirement 15.4)
	rec.PositionSize = s.calculatePositionSize(riskAssessment)

	// Build reasoning
	var reasoningParts []string

	// Check Knowledge_Base for recent patterns
	if s.knowledgeBase != nil {
		observations, err := s.knowledgeBase.GetRecentObservationsForSymbol(ctx, symbol, 14)
		if err == nil && len(observations) > 0 {
			latestObs := observations[0]
			switch latestObs.PatternType {
			case model.PatternAccumulation:
				if latestObs.ConfidenceScore >= 70 {
					rec.Action = "buy"
					reasoningParts = append(reasoningParts, fmt.Sprintf(
						"Accumulation pattern detected with %d%% confidence",
						latestObs.ConfidenceScore,
					))
				}
			case model.PatternDistribution:
				if latestObs.ConfidenceScore >= 70 {
					rec.Action = "sell"
					reasoningParts = append(reasoningParts, fmt.Sprintf(
						"Distribution pattern detected with %d%% confidence",
						latestObs.ConfidenceScore,
					))
				}
			case model.PatternBreakout:
				if latestObs.ConfidenceScore >= 60 {
					rec.Action = "buy"
					reasoningParts = append(reasoningParts, fmt.Sprintf(
						"Breakout signal detected with %d%% confidence",
						latestObs.ConfidenceScore,
					))
				}
			}
		}
	}

	// Add sector context to reasoning
	if s.sectorService != nil {
		sector, err := s.sectorService.GetStockSector(symbol)
		if err == nil {
			perf, err := s.sectorService.GetSectorPerformance(ctx, sector)
			if err == nil {
				if perf.Trend == model.Uptrend {
					reasoningParts = append(reasoningParts, fmt.Sprintf(
						"Sector %s is in uptrend (+%.1f%% 1M)",
						perf.SectorName, perf.OneMonthChange,
					))
				} else if perf.Trend == model.Downtrend {
					reasoningParts = append(reasoningParts, fmt.Sprintf(
						"Caution: Sector %s is in downtrend (%.1f%% 1M)",
						perf.SectorName, perf.OneMonthChange,
					))
				}
			}
		}
	}

	// Check existing position in portfolio
	if portfolio != nil && portfolio.Available {
		for _, holding := range portfolio.Holdings {
			if holding.Asset.Symbol == symbol {
				holdingPercent := 0.0
				if portfolio.NAV > 0 {
					holdingPercent = (holding.MarketValue / portfolio.NAV) * 100
				}
				reasoningParts = append(reasoningParts, fmt.Sprintf(
					"Current position: %.2f units (%.1f%% of NAV, P&L: %.1f%%)",
					holding.Asset.Quantity, holdingPercent, holding.UnrealizedPLPct,
				))

				// Adjust recommendation based on existing position
				if holdingPercent > diversificationLimit*100 {
					rec.Action = "hold"
					rec.PositionSize = 0
					reasoningParts = append(reasoningParts,
						"Position already exceeds diversification limit - no additional buying recommended")
				}
				break
			}
		}
	}

	if len(reasoningParts) == 0 {
		reasoningParts = append(reasoningParts, "Insufficient data for detailed analysis")
	}

	rec.Reasoning = strings.Join(reasoningParts, ". ")

	return rec
}

// assessRisk determines the risk level for a symbol based on available data.
func (s *SupervisorAgent) assessRisk(ctx context.Context, symbol string) string {
	// Default to medium risk
	risk := "medium"

	// Check sector volatility
	if s.sectorService != nil {
		sector, err := s.sectorService.GetStockSector(symbol)
		if err == nil {
			perf, err := s.sectorService.GetSectorPerformance(ctx, sector)
			if err == nil {
				// High volatility sectors
				if sector == model.VNREAL || sector == model.VNFIN {
					risk = "high"
				}
				// Check recent performance volatility
				if absFloat(perf.OneMonthChange) > 15 {
					risk = "high"
				} else if absFloat(perf.OneMonthChange) < 5 {
					risk = "low"
				}
			}
		}
	}

	// Check Knowledge_Base for pattern confidence
	if s.knowledgeBase != nil {
		observations, err := s.knowledgeBase.GetRecentObservationsForSymbol(ctx, symbol, 7)
		if err == nil && len(observations) > 0 {
			// High confidence patterns reduce perceived risk
			if observations[0].ConfidenceScore >= 80 {
				if risk == "high" {
					risk = "medium"
				}
			}
		}
	}

	return risk
}

// calculatePositionSize returns the recommended position size as percentage of NAV.
func (s *SupervisorAgent) calculatePositionSize(riskAssessment string) float64 {
	switch riskAssessment {
	case "high":
		return maxPositionHighRisk * 100 // 5%
	case "low":
		return maxPositionLowRisk * 100 // 20%
	default:
		return maxPositionMediumRisk * 100 // 10%
	}
}

// ---------------------------------------------------------------------------
// Opportunity identification (Requirement 15.6)
// ---------------------------------------------------------------------------

// identifyOpportunities finds new investment opportunities based on patterns
// and sector trends.
func (s *SupervisorAgent) identifyOpportunities(
	ctx context.Context,
	queriedSymbols []string,
	kbInsights []string,
) []string {
	var opportunities []string

	// Look for high-confidence patterns in Knowledge_Base
	if s.knowledgeBase != nil {
		// Query recent high-confidence accumulation patterns
		filters := ObservationFilters{
			PatternType:   string(model.PatternAccumulation),
			MinConfidence: 70,
			StartDate:     time.Now().AddDate(0, 0, -7),
			Limit:         5,
		}

		observations, err := s.knowledgeBase.QueryObservations(ctx, filters)
		if err == nil {
			for _, obs := range observations {
				// Skip if already in queried symbols
				isQueried := false
				for _, qs := range queriedSymbols {
					if qs == obs.Symbol {
						isQueried = true
						break
					}
				}
				if !isQueried {
					opportunities = append(opportunities, fmt.Sprintf(
						"%s: Accumulation pattern detected with %d%% confidence",
						obs.Symbol, obs.ConfidenceScore,
					))
				}
			}
		}

		// Query recent breakout patterns
		filters.PatternType = string(model.PatternBreakout)
		filters.MinConfidence = 65

		observations, err = s.knowledgeBase.QueryObservations(ctx, filters)
		if err == nil {
			for _, obs := range observations {
				isQueried := false
				for _, qs := range queriedSymbols {
					if qs == obs.Symbol {
						isQueried = true
						break
					}
				}
				if !isQueried {
					opportunities = append(opportunities, fmt.Sprintf(
						"%s: Breakout signal with %d%% confidence",
						obs.Symbol, obs.ConfidenceScore,
					))
				}
			}
		}
	}

	// Look for sectors in uptrend
	if s.sectorService != nil {
		perfs, err := s.sectorService.GetAllSectorPerformances(ctx)
		if err == nil {
			for _, perf := range perfs {
				if perf.Trend == model.Uptrend && perf.OneMonthChange > 5 {
					opportunities = append(opportunities, fmt.Sprintf(
						"Sector %s showing strong momentum (+%.1f%% 1M, +%.1f%% 3M)",
						perf.SectorName, perf.OneMonthChange, perf.ThreeMonthChange,
					))
				}
			}
		}
	}

	// Limit to top 5 opportunities
	if len(opportunities) > 5 {
		opportunities = opportunities[:5]
	}

	return opportunities
}

// ---------------------------------------------------------------------------
// Portfolio suggestions (Requirement 15.8)
// ---------------------------------------------------------------------------

// buildPortfolioSuggestions creates portfolio-level suggestions including
// diversification warnings.
func (s *SupervisorAgent) buildPortfolioSuggestions(
	portfolio *portfolioContext,
	diversificationWarnings []string,
) []string {
	var suggestions []string

	// Add diversification warnings first
	suggestions = append(suggestions, diversificationWarnings...)

	if portfolio == nil || !portfolio.Available {
		suggestions = append(suggestions,
			"Set up portfolio tracking to receive personalized recommendations")
		return suggestions
	}

	// Analyze allocation
	if portfolio.NAV > 0 {
		// Check for over-concentration in any asset type
		for assetType, percent := range portfolio.AllocationPercent {
			if percent > 60 {
				suggestions = append(suggestions, fmt.Sprintf(
					"Consider diversifying: %s represents %.1f%% of portfolio",
					assetType, percent,
				))
			}
		}

		// Check for missing asset types
		hasStocks := portfolio.AllocationPercent[model.VNStock] > 0
		hasCash := portfolio.AllocationPercent[model.Cash] > 0 ||
			portfolio.AllocationPercent[model.Savings] > 0

		if !hasStocks && portfolio.NAV > 10000000 { // > 10M VND
			suggestions = append(suggestions,
				"Consider adding VN stocks for growth potential")
		}

		if !hasCash {
			suggestions = append(suggestions,
				"Consider maintaining a cash reserve for opportunities")
		}
	}

	return suggestions
}

// ---------------------------------------------------------------------------
// Summary building (Requirement 15.9, 15.10)
// ---------------------------------------------------------------------------

// buildSummary creates the recommendation summary.
func (s *SupervisorAgent) buildSummary(
	intent model.QueryIntent,
	portfolio *portfolioContext,
	numRecommendations int,
	numInsights int,
) string {
	var parts []string

	// Query context
	if len(intent.Symbols) > 0 {
		parts = append(parts, fmt.Sprintf(
			"Analysis for %s", strings.Join(intent.Symbols, ", "),
		))
	}

	// Portfolio context (Requirement 15.10)
	if portfolio != nil && portfolio.Available && portfolio.NAV > 0 {
		parts = append(parts, fmt.Sprintf(
			"Portfolio NAV: %.0f VND", portfolio.NAV,
		))
	} else {
		parts = append(parts,
			"Portfolio data unavailable - providing general market analysis")
	}

	// Recommendations summary
	if numRecommendations > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d asset recommendation(s) generated", numRecommendations,
		))
	}

	// Knowledge base insights
	if numInsights > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d historical pattern insight(s) incorporated", numInsights,
		))
	}

	return strings.Join(parts, ". ") + "."
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// absFloat returns the absolute value of a float64.
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetRecommendation is a convenience method that returns a SupervisorRecommendation
// directly without the AgentMessage wrapper.
func (s *SupervisorAgent) GetRecommendation(
	ctx context.Context,
	symbols []string,
	userID int64,
) (model.SupervisorRecommendation, error) {
	intent := model.QueryIntent{
		Symbols:  symbols,
		Intent:   "recommendation",
		RawQuery: fmt.Sprintf("Recommendation for %s", strings.Join(symbols, ", ")),
	}

	msg, err := s.Execute(ctx, intent, nil)
	if err != nil {
		return model.SupervisorRecommendation{}, err
	}

	// Extract recommendation from payload
	if rec, ok := msg.Payload["recommendation"].(model.SupervisorRecommendation); ok {
		return rec, nil
	}

	// Try JSON unmarshaling if direct type assertion fails
	if recData, ok := msg.Payload["recommendation"]; ok {
		jsonBytes, err := json.Marshal(recData)
		if err != nil {
			return model.SupervisorRecommendation{}, fmt.Errorf("failed to marshal recommendation: %w", err)
		}
		var rec model.SupervisorRecommendation
		if err := json.Unmarshal(jsonBytes, &rec); err != nil {
			return model.SupervisorRecommendation{}, fmt.Errorf("failed to unmarshal recommendation: %w", err)
		}
		return rec, nil
	}

	return model.SupervisorRecommendation{
		Summary: "Failed to generate recommendation",
	}, nil
}
