package platform

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterDocsRoutes adds Swagger UI and OpenAPI spec endpoints.
// GET /api/docs         → Swagger UI HTML page
// GET /api/docs/swagger.json → OpenAPI 3.0 spec
// Requirements: 43.1, 43.2, 43.3, 43.7
func RegisterDocsRoutes(r *gin.Engine) {
	r.GET("/api/docs/swagger.json", handleSwaggerJSON)
	r.GET("/api/docs", handleSwaggerUI)
}

func handleSwaggerJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, openAPISpec)
}

func handleSwaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	specURL := "/api/docs/swagger.json"
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>EziStock API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "` + specURL + `",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`
	c.String(http.StatusOK, html)
}

// openAPISpec is the OpenAPI 3.0 specification for the EziStock API.
// Grouped by feature tag per Requirements 43.7.
var openAPISpec string

func init() {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "EziStock API",
			"description": "Vietnamese stock quantitative analysis and AI-powered investment platform API",
			"version":     "1.0.0",
		},
		"servers": []map[string]any{
			{"url": "http://localhost:8080", "description": "Local development"},
		},
		"tags": []map[string]any{
			{"name": "Health", "description": "Service health checks"},
			{"name": "Auth", "description": "Authentication and user management"},
			{"name": "Market", "description": "Market data, quotes, charts, indices"},
			{"name": "Sectors", "description": "ICB sector performance and trends"},
			{"name": "Portfolio", "description": "Portfolio holdings, transactions, performance, risk"},
			{"name": "Screener", "description": "Stock screener with filters and presets"},
			{"name": "Watchlist", "description": "Named watchlists with price alerts"},
			{"name": "AI", "description": "AI chat, stock analysis, hot topics"},
			{"name": "Ranking", "description": "AI stock ranking and backtesting"},
			{"name": "Ideas", "description": "Proactive investment ideas"},
			{"name": "Missions", "description": "User-defined monitoring tasks"},
			{"name": "Notifications", "description": "User notifications"},
			{"name": "Knowledge", "description": "Knowledge base observations"},
			{"name": "Analyst", "description": "Analyst consensus and reports"},
			{"name": "Research", "description": "Research reports and PDFs"},
			{"name": "Export", "description": "CSV/PDF export"},
			{"name": "Macro", "description": "Macroeconomic indicators"},
			{"name": "Sentiment", "description": "Sentiment analysis"},
			{"name": "Consensus", "description": "Market consensus signals"},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
		"paths": buildPaths(),
	}
	b, _ := json.MarshalIndent(spec, "", "  ")
	openAPISpec = string(b)
}

func buildPaths() map[string]any {
	paths := map[string]any{}

	// Helper to create a path entry
	ep := func(tag, summary, desc string) map[string]any {
		m := map[string]any{
			"tags":        []string{tag},
			"summary":     summary,
			"description": desc,
			"responses": map[string]any{
				"200": map[string]any{"description": "Success"},
			},
		}
		if tag != "Health" && tag != "Auth" {
			m["security"] = []map[string]any{{"BearerAuth": []string{}}}
		}
		return m
	}

	// Health
	paths["/api/health"] = map[string]any{"get": ep("Health", "Health check", "Returns service health status")}
	paths["/api/healthz"] = map[string]any{"get": ep("Health", "Liveness probe", "Kubernetes liveness probe")}
	paths["/api/readyz"] = map[string]any{"get": ep("Health", "Readiness probe", "Kubernetes readiness probe")}

	// Auth
	paths["/api/auth/login"] = map[string]any{"post": ep("Auth", "Login", "Authenticate and receive JWT token")}
	paths["/api/auth/register"] = map[string]any{"post": ep("Auth", "Register", "Create a new user account")}
	paths["/api/auth/logout"] = map[string]any{"post": ep("Auth", "Logout", "Invalidate current session")}
	paths["/api/auth/me"] = map[string]any{"get": ep("Auth", "Current user", "Get authenticated user profile")}

	// Market
	paths["/api/market/quote"] = map[string]any{"get": ep("Market", "Stock quotes", "Real-time quotes for symbols")}
	paths["/api/market/chart"] = map[string]any{"get": ep("Market", "OHLCV chart data", "Historical OHLCV data for charting")}
	paths["/api/market/listing"] = map[string]any{"get": ep("Market", "Stock listing", "All listed stocks on VN exchanges")}
	paths["/api/market/search"] = map[string]any{"get": ep("Market", "Global search", "Fuzzy search across symbols and company names")}
	paths["/api/market/index"] = map[string]any{"get": ep("Market", "Market index", "VN-Index, HNX, UPCOM index data")}
	paths["/api/market/statistics"] = map[string]any{"get": ep("Market", "Market statistics", "Market-wide statistics")}
	paths["/api/market/company/{symbol}"] = map[string]any{"get": ep("Market", "Company profile", "Company profile and details")}
	paths["/api/market/finance/{symbol}"] = map[string]any{"get": ep("Market", "Financial reports", "Financial statements for a symbol")}
	paths["/api/market/ratios/{symbol}"] = map[string]any{"get": ep("Market", "Financial ratios", "Financial ratios for a symbol")}
	paths["/api/market/hot-topics"] = map[string]any{"get": ep("Market", "Hot topics", "Trending market topics from AI")}

	// Macro
	paths["/api/market/macro"] = map[string]any{"get": ep("Macro", "Macro overview", "Macroeconomic indicators overview")}
	paths["/api/market/macro/interbank"] = map[string]any{"get": ep("Macro", "Interbank rates", "Vietnamese interbank lending rates")}
	paths["/api/market/macro/bonds"] = map[string]any{"get": ep("Macro", "Bond yields", "Government bond yields")}
	paths["/api/market/macro/fx"] = map[string]any{"get": ep("Macro", "FX rates", "Vietcombank official exchange rates")}

	// Sectors
	paths["/api/sectors/performance"] = map[string]any{"get": ep("Sectors", "All sectors", "ICB sector performance overview")}
	paths["/api/sectors/{sector}/trend"] = map[string]any{"get": ep("Sectors", "Sector trend", "Trend analysis for a sector")}
	paths["/api/sectors/{sector}/stocks"] = map[string]any{"get": ep("Sectors", "Sector stocks", "Stocks in a sector")}

	// Portfolio
	paths["/api/portfolio/holdings"] = map[string]any{"get": ep("Portfolio", "Holdings", "Current portfolio holdings")}
	paths["/api/portfolio/nav"] = map[string]any{"get": ep("Portfolio", "NAV", "Portfolio net asset value")}
	paths["/api/portfolio/buy"] = map[string]any{"post": ep("Portfolio", "Record buy", "Record a stock purchase")}
	paths["/api/portfolio/sell"] = map[string]any{"post": ep("Portfolio", "Record sell", "Record a stock sale")}
	paths["/api/portfolio/transactions"] = map[string]any{"get": ep("Portfolio", "Transactions", "Transaction history")}
	paths["/api/portfolio/performance"] = map[string]any{"get": ep("Portfolio", "Performance", "TWR, XIRR, equity curve")}
	paths["/api/portfolio/risk"] = map[string]any{"get": ep("Portfolio", "Risk metrics", "Sharpe, VaR, beta, drawdown")}

	// Screener
	paths["/api/screener"] = map[string]any{"post": ep("Screener", "Screen stocks", "Filter stocks by criteria")}
	paths["/api/screener/presets"] = map[string]any{
		"get":  ep("Screener", "List presets", "Saved screener filter presets"),
		"post": ep("Screener", "Save preset", "Save a new filter preset"),
	}

	// Watchlists
	paths["/api/watchlists"] = map[string]any{
		"get":  ep("Watchlist", "List watchlists", "All user watchlists"),
		"post": ep("Watchlist", "Create watchlist", "Create a named watchlist"),
	}

	// AI
	paths["/api/chat"] = map[string]any{"post": ep("AI", "AI chat", "Conversational AI with citation support")}
	paths["/api/analyze/{symbol}"] = map[string]any{
		"post": ep("AI", "Analyze stock", "Full multi-agent stock analysis"),
		"get":  ep("AI", "Analyze stock (GET)", "Full multi-agent stock analysis"),
	}

	// Ranking
	paths["/api/ranking"] = map[string]any{
		"post": ep("Ranking", "Compute ranking", "AI stock ranking with custom factors"),
		"get":  ep("Ranking", "Top ranked", "Get top ranked stocks"),
	}
	paths["/api/ranking/backtest"] = map[string]any{"post": ep("Ranking", "Backtest", "Run strategy backtest")}
	paths["/api/ranking/factors"] = map[string]any{"get": ep("Ranking", "Factor groups", "Available factor groups")}

	// Ideas
	paths["/api/ideas"] = map[string]any{"get": ep("Ideas", "Investment ideas", "Proactive buy/sell recommendations")}

	// Missions
	paths["/api/missions"] = map[string]any{
		"get":  ep("Missions", "List missions", "User monitoring tasks"),
		"post": ep("Missions", "Create mission", "Create a monitoring task"),
	}

	// Notifications
	paths["/api/notifications"] = map[string]any{"get": ep("Notifications", "List notifications", "User notifications")}
	paths["/api/notifications/unread"] = map[string]any{"get": ep("Notifications", "Unread count", "Unread notification count")}

	// Knowledge
	paths["/api/knowledge/observations"] = map[string]any{"get": ep("Knowledge", "Observations", "Knowledge base observations")}

	// Analyst
	paths["/api/analyst/consensus/{symbol}"] = map[string]any{"get": ep("Analyst", "Consensus", "Analyst consensus for a symbol")}
	paths["/api/analyst/reports/{symbol}"] = map[string]any{"get": ep("Analyst", "Reports", "Analyst reports for a symbol")}

	// Research
	paths["/api/research/reports"] = map[string]any{"get": ep("Research", "List reports", "Research reports")}
	paths["/api/research/reports/{id}"] = map[string]any{"get": ep("Research", "Report detail", "Single research report")}
	paths["/api/research/reports/{id}/pdf"] = map[string]any{"get": ep("Research", "Download PDF", "Download report as PDF")}

	// Export
	paths["/api/export/transactions"] = map[string]any{"get": ep("Export", "Export transactions", "CSV/PDF transaction export")}
	paths["/api/export/portfolio"] = map[string]any{"get": ep("Export", "Export portfolio", "CSV/PDF portfolio export")}
	paths["/api/export/pnl"] = map[string]any{"get": ep("Export", "Export P&L", "CSV/PDF P&L export")}
	paths["/api/export/tax"] = map[string]any{"get": ep("Export", "Export tax", "VN tax report export")}

	return paths
}
