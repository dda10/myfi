package agent

import (
	"net/http"
	"strings"

	"myfi-backend/internal/generated/proto/ezistockpb"
	"myfi-backend/internal/infra"

	"github.com/gin-gonic/gin"
)

// Handlers holds agent domain dependencies for HTTP handler methods.
type Handlers struct {
	GRPCClient *infra.GRPCClient
}

// aiUnavailableResponse returns a 503 JSON response with the ai_service_unavailable flag.
// Requirements: 1.6, 45.3
func aiUnavailableResponse(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error":                  "ai_service_unavailable",
		"message":                message,
		"ai_service_unavailable": true,
	})
}

// HandleChat serves POST /api/chat — proxy to Python AI Service via gRPC.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleChat(c *gin.Context) {
	var req struct {
		Message   string `json:"message" binding:"required"`
		Symbol    string `json:"symbol"`
		SessionID string `json:"sessionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}

	// Build protobuf request.
	pbReq := &ezistockpb.ChatRequest{
		UserId:  req.SessionID,
		Message: req.Message,
	}

	resp, err := h.GRPCClient.Chat(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}

	// Convert citations to JSON-friendly format.
	citations := make([]gin.H, 0, len(resp.Citations))
	for _, cit := range resp.Citations {
		citations = append(citations, gin.H{
			"source":     cit.Source,
			"url":        cit.Url,
			"claim":      cit.Claim,
			"data_point": cit.DataPoint,
		})
	}

	// Convert suggestions to JSON-friendly format.
	suggestions := make([]gin.H, 0, len(resp.Suggestions))
	for _, s := range resp.Suggestions {
		suggestions = append(suggestions, gin.H{
			"text":   s.Text,
			"type":   s.Type,
			"symbol": s.Symbol,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":       resp.Response,
		"citations":   citations,
		"suggestions": suggestions,
		"disclaimer":  resp.Disclaimer,
	})
}

// HandleAnalyzeStock serves POST/GET /api/analyze/:symbol — proxy to Python AI Service via gRPC.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleAnalyzeStock(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}

	pbReq := &ezistockpb.AnalyzeStockRequest{
		Symbol: symbol,
	}

	resp, err := h.GRPCClient.AnalyzeStock(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, buildAnalyzeResponse(resp))
}

// HandleGetHotTopics serves GET /api/market/hot-topics — proxy to Python AI Service via gRPC.
// This is called from the market handler but can also be invoked directly.
// Requirements: 1.2, 1.6
func (h *Handlers) HandleGetHotTopics(c *gin.Context) {
	if h.GRPCClient == nil {
		aiUnavailableResponse(c, "AI service is not configured")
		return
	}

	pbReq := &ezistockpb.HotTopicsRequest{
		Limit:  20,
		Market: c.Query("market"),
	}

	resp, err := h.GRPCClient.GetHotTopics(c.Request.Context(), pbReq)
	if err != nil {
		aiUnavailableResponse(c, err.Error())
		return
	}

	topics := make([]gin.H, 0, len(resp.Topics))
	for _, t := range resp.Topics {
		topics = append(topics, gin.H{
			"symbol":          t.Symbol,
			"topic":           t.Topic,
			"category":        t.Category,
			"relevance_score": t.RelevanceScore,
			"summary":         t.Summary,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         topics,
		"generated_at": resp.GeneratedAt,
	})
}

// buildAnalyzeResponse converts the protobuf AnalyzeStockResponse to a JSON-friendly map.
func buildAnalyzeResponse(resp *ezistockpb.AnalyzeStockResponse) gin.H {
	result := gin.H{}

	if tech := resp.Technical; tech != nil {
		result["technical"] = gin.H{
			"symbol":            tech.Symbol,
			"composite_signal":  tech.CompositeSignal,
			"indicators":        tech.Indicators,
			"support_levels":    tech.SupportLevels,
			"resistance_levels": tech.ResistanceLevels,
			"patterns":          tech.Patterns,
			"smart_money_flow":  tech.SmartMoneyFlow,
			"ma_crossovers":     tech.MaCrossovers,
		}
	}

	if news := resp.News; news != nil {
		articles := make([]gin.H, 0, len(news.Articles))
		for _, a := range news.Articles {
			articles = append(articles, gin.H{
				"title":        a.Title,
				"url":          a.Url,
				"source":       a.Source,
				"published_at": a.PublishedAt,
				"summary":      a.Summary,
			})
		}
		result["news"] = gin.H{
			"symbol":       news.Symbol,
			"sentiment":    news.Sentiment,
			"confidence":   news.Confidence,
			"catalysts":    news.Catalysts,
			"risk_factors": news.RiskFactors,
			"articles":     articles,
		}
	}

	if rec := resp.Recommendation; rec != nil {
		result["recommendation"] = gin.H{
			"symbol":            rec.Symbol,
			"action":            rec.Action,
			"target_price":      rec.TargetPrice,
			"upside_percent":    rec.UpsidePercent,
			"confidence_score":  rec.ConfidenceScore,
			"risk_level":        rec.RiskLevel,
			"reasoning":         rec.Reasoning,
			"technical_factors": rec.TechnicalFactors,
			"news_factors":      rec.NewsFactors,
		}
	}

	if strat := resp.Strategy; strat != nil {
		result["strategy"] = gin.H{
			"symbol":                strat.Symbol,
			"signal_direction":      strat.SignalDirection,
			"entry_price":           strat.EntryPrice,
			"stop_loss":             strat.StopLoss,
			"take_profit":           strat.TakeProfit,
			"risk_reward_ratio":     strat.RiskRewardRatio,
			"confidence_score":      strat.ConfidenceScore,
			"position_size_percent": strat.PositionSizePercent,
			"reasoning":             strat.Reasoning,
		}
	}

	citations := make([]gin.H, 0, len(resp.Citations))
	for _, cit := range resp.Citations {
		citations = append(citations, gin.H{
			"source":     cit.Source,
			"url":        cit.Url,
			"claim":      cit.Claim,
			"data_point": cit.DataPoint,
		})
	}
	result["citations"] = citations
	result["disclaimer"] = resp.Disclaimer

	return result
}
