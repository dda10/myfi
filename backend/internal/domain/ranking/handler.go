package ranking

import (
	"net/http"
	"strconv"

	pb "myfi-backend/internal/generated/proto/ezistockpb"
	"myfi-backend/internal/infra"

	"github.com/gin-gonic/gin"
)

// Handlers holds ranking domain dependencies for HTTP handler methods.
type Handlers struct {
	Tracker        *RecommendationTracker
	BacktestEngine *BacktestEngine
	GRPCClient     *infra.GRPCClient
}

// defaultFactorGroups is the standard set of factor groups used across ranking and backtest requests.
var defaultFactorGroups = []string{"quality", "value", "growth", "momentum"}

// aiUnavailableResponse returns a 503 JSON response when the Python AI Service is down.
// Requirements: 1.6, 45.3
func aiUnavailableResponse(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error":                  "ai_service_unavailable",
		"message":                message,
		"ai_service_unavailable": true,
	})
}

// HandleRanking serves POST /api/ranking — compute stock ranking via Python AI Service.
// Requirements: 1.2, 1.6, 45.3
func (h *Handlers) HandleRanking(c *gin.Context) {
	var config RankingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if config.Universe == "" {
		config.Universe = "VN30"
	}
	if config.TopN <= 0 {
		config.TopN = 20
	}
	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}

	factorGroupNames := make([]string, 0, len(config.FactorGroups))
	for _, fg := range config.FactorGroups {
		factorGroupNames = append(factorGroupNames, string(fg.Name))
	}
	if len(factorGroupNames) == 0 {
		factorGroupNames = defaultFactorGroups
	}
	pbReq := &pb.RankingRequest{
		Universe:     []string{config.Universe},
		FactorGroups: factorGroupNames,
		TopN:         int32(config.TopN),
	}
	resp, err := h.GRPCClient.GetRanking(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}
	rankings := make([]RankedStock, 0, len(resp.Rankings))
	for _, r := range resp.Rankings {
		rankings = append(rankings, RankedStock{
			Symbol:         r.Symbol,
			Rank:           int(r.Rank),
			CompositeScore: r.CompositeScore,
		})
	}
	c.JSON(http.StatusOK, RankingResult{
		Config:      config,
		Rankings:    rankings,
		TotalStocks: len(rankings),
	})
}

// HandleBacktest serves POST /api/ranking/backtest — run strategy backtest via Python AI Service.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleBacktest(c *gin.Context) {
	var req struct {
		Symbol   string       `json:"symbol" binding:"required"`
		Strategy StrategyRule `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}
	pbReq := &pb.BacktestRequest{
		Universe:     []string{req.Symbol},
		FactorGroups: defaultFactorGroups,
	}
	resp, err := h.GRPCClient.RunBacktest(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}
	result := gin.H{"symbol": req.Symbol}
	if m := resp.Metrics; m != nil {
		result["cumulative_return"] = m.CumulativeReturn
		result["annualized_return"] = m.AnnualizedReturn
		result["sharpe_ratio"] = m.SharpeRatio
		result["max_drawdown"] = m.MaxDrawdown
		result["win_rate"] = m.WinRate
		result["profit_factor"] = m.ProfitFactor
	}
	c.JSON(http.StatusOK, result)
}

// HandleGetFactors serves GET /api/ranking/factors — list available factor groups.
func (h *Handlers) HandleGetFactors(c *gin.Context) {
	factors := []FactorGroup{
		{Name: FactorQuality, Weight: 0.2, Factors: []Factor{
			{Code: "roe", Name: "Return on Equity", Weight: 0.5},
			{Code: "roa", Name: "Return on Assets", Weight: 0.3},
			{Code: "debt_equity", Name: "Debt to Equity", Weight: 0.2},
		}},
		{Name: FactorValue, Weight: 0.2, Factors: []Factor{
			{Code: "pe", Name: "Price to Earnings", Weight: 0.4},
			{Code: "pb", Name: "Price to Book", Weight: 0.3},
			{Code: "ev_ebitda", Name: "EV/EBITDA", Weight: 0.3},
		}},
		{Name: FactorGrowth, Weight: 0.2, Factors: []Factor{
			{Code: "revenue_growth", Name: "Revenue Growth", Weight: 0.5},
			{Code: "profit_growth", Name: "Profit Growth", Weight: 0.5},
		}},
		{Name: FactorMomentum, Weight: 0.2, Factors: []Factor{
			{Code: "momentum_3m", Name: "3-Month Momentum", Weight: 0.5},
			{Code: "momentum_6m", Name: "6-Month Momentum", Weight: 0.5},
		}},
		{Name: FactorVolatility, Weight: 0.2, Factors: []Factor{
			{Code: "volatility_30d", Name: "30-Day Volatility", Weight: 1.0},
		}},
	}
	c.JSON(http.StatusOK, gin.H{"data": factors})
}

// HandleGetIdeas serves GET /api/ideas — proactive investment ideas via Python AI Service.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleGetIdeas(c *gin.Context) {
	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}
	maxIdeas := int32(10)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdeas = int32(n)
		}
	}
	pbReq := &pb.IdeaRequest{MaxIdeas: maxIdeas}
	resp, err := h.GRPCClient.GenerateInvestmentIdeas(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}
	ideas := make([]gin.H, 0, len(resp.Ideas))
	for _, idea := range resp.Ideas {
		ideas = append(ideas, gin.H{
			"symbol":              idea.Symbol,
			"direction":           idea.SignalDirection,
			"entry_price":         idea.EntryPrice,
			"target_price":        idea.TargetPrice,
			"confidence":          idea.ConfidenceScore,
			"reasoning":           idea.Reasoning,
			"historical_accuracy": idea.HistoricalAccuracy,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": ideas, "total": len(ideas)})
}

// HandleGetRankingList serves GET /api/ranking — return top ranked stocks via Python AI Service.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleGetRankingList(c *gin.Context) {
	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}
	universe := c.DefaultQuery("universe", "VN30")
	topN := int32(20)
	if v := c.Query("top_n"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topN = int32(n)
		}
	}
	pbReq := &pb.RankingRequest{
		Universe:     []string{universe},
		FactorGroups: defaultFactorGroups,
		TopN:         topN,
	}
	resp, err := h.GRPCClient.GetRanking(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}
	rankings := make([]gin.H, 0, len(resp.Rankings))
	for _, r := range resp.Rankings {
		rankings = append(rankings, gin.H{
			"symbol":          r.Symbol,
			"rank":            r.Rank,
			"composite_score": r.CompositeScore,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": rankings, "total": len(rankings)})
}

// HandleGetRecommendations serves GET /api/ranking/recommendations — recommendation history.
func (h *Handlers) HandleGetRecommendations(c *gin.Context) {
	if h.Tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracker not initialized"})
		return
	}
	filter := RecommendationFilter{
		Symbol: c.Query("symbol"),
		Limit:  50,
	}
	if action := c.Query("action"); action != "" {
		filter.Action = RecommendationAction(action)
	}
	records := h.Tracker.GetRecommendations(filter)
	c.JSON(http.StatusOK, gin.H{"data": records, "total": len(records)})
}

// HandleGetAccuracy serves GET /api/ranking/accuracy — recommendation accuracy metrics.
func (h *Handlers) HandleGetAccuracy(c *gin.Context) {
	if h.Tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommendation tracker not initialized"})
		return
	}
	summary := h.Tracker.GetSummary()
	c.JSON(http.StatusOK, summary)
}
