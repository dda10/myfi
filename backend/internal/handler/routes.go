package handler

import (
	"myfi-backend/internal/infra"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter creates a Gin engine with CORS middleware and all route registrations.
// CORS configuration is identical to the original main.go.
func SetupRouter(h *Handlers) *gin.Engine {
	r := gin.Default()

	// Global panic recovery middleware (Task 38.1)
	r.Use(infra.RecoveryMiddleware())

	// HTTPS redirect in production (Requirement 36.9)
	if os.Getenv("ENV") == "production" {
		r.Use(infra.HTTPSRedirectMiddleware())
	}

	// CORS — restrict to frontend origin in production (Requirement 36.9)
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:3000" // dev default
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"},
		AllowCredentials: true,
	}))

	// Per-IP rate limiting for all endpoints (Requirement 36.9)
	r.Use(infra.PerIPRateLimitMiddleware(200))

	// Health check — public, no auth required
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public auth endpoints (no JWT required)
	r.POST("/api/auth/register", h.HandleRegister)
	r.POST("/api/auth/login", h.HandleLogin)

	// All other endpoints require JWT authentication (Requirement 36.3, 36.4)
	// Per-user rate limiting applied after JWT validation (Requirement 36.9)
	protected := r.Group("/api")
	protected.Use(h.JWTMiddleware())
	protected.Use(infra.PerUserRateLimitMiddleware(100))
	{
		// Auth endpoints
		protected.POST("/auth/logout", h.HandleLogout)
		protected.POST("/auth/change-password", h.HandleChangePassword)
		protected.GET("/auth/me", h.HandleMe)

		// Metrics
		protected.GET("/metrics/rate-limits", h.HandleRateLimitMetrics)

		// Market data endpoints
		protected.GET("/market/quote", h.HandleMarketQuote)
		protected.GET("/market/chart", h.HandleMarketChart)
		protected.GET("/market/listing", h.HandleUnifiedListing)
		protected.GET("/market/screener", h.HandleMarketScreener)
		protected.GET("/market/company/:symbol", h.HandleCompanyData)
		protected.GET("/market/finance/:symbol", h.HandleFinancialReports)
		protected.GET("/market/trading/:symbol", h.HandleTradingStatistics)
		protected.GET("/market/statistics", h.HandleMarketStatistics)
		protected.GET("/market/valuation", h.HandleValuationMetrics)
		protected.GET("/market/funds", h.HandleFunds)
		protected.GET("/market/commodities", h.HandleCommodities)
		protected.GET("/market/macro", h.HandleMacro)
		protected.POST("/market/trading/batch", h.HandleBatchTradingQuotes)

		// Liquidity whitelist endpoints
		protected.GET("/market/whitelist", h.HandleWhitelist)
		protected.GET("/market/whitelist/check", h.HandleWhitelistCheck)
		protected.POST("/market/whitelist/refresh", h.HandleWhitelistRefresh)

		// Price and crypto endpoints
		protected.GET("/crypto/quote", h.HandleCryptoQuote)
		protected.GET("/prices/fx", h.HandleFXRate)

		// News and AI chat
		protected.GET("/news", h.HandleNews)
		protected.POST("/chat", h.HandleChat)
		protected.POST("/models", h.HandleModels)

		// AI recommendation tracking endpoints
		protected.GET("/recommendations/summary", h.HandleRecommendationSummary)
		protected.GET("/recommendations/accuracy", h.HandleRecommendationAccuracy)
		protected.GET("/recommendations", h.HandleRecommendationList)
		protected.GET("/recommendations/:id", h.HandleRecommendationByID)
		protected.POST("/recommendations/update-outcomes", h.HandleUpdateOutcomes)

		// Signal engine endpoints
		protected.GET("/signals/scan", h.HandleSignalScan)
		protected.POST("/signals/backtest", h.BacktestSignals)
		protected.POST("/signals/optimize", h.OptimizeSignalWeights)

		// Price endpoints
		protected.GET("/prices/quotes", h.HandlePriceQuotes)
		protected.GET("/prices/history", h.HandlePriceHistory)
		protected.GET("/prices/gold", h.HandleGoldPrices)
		protected.GET("/prices/crypto", h.HandleCryptoPrices)

		// Portfolio endpoints
		protected.GET("/portfolio/summary", h.HandlePortfolioSummary)
		protected.POST("/portfolio/assets", h.HandleAddAsset)
		protected.PUT("/portfolio/assets/:id", h.HandleUpdateAsset)
		protected.DELETE("/portfolio/assets/:id", h.HandleDeleteAsset)
		protected.GET("/portfolio/transactions", h.HandleGetTransactions)
		protected.POST("/portfolio/transactions", h.HandleRecordTransaction)
		protected.GET("/portfolio/performance", h.HandlePortfolioPerformance)
		protected.GET("/portfolio/risk", h.HandlePortfolioRisk)

		// Sector endpoints
		protected.GET("/sectors/performance", h.HandleAllSectorPerformances)
		protected.GET("/sectors/symbol/:symbol", h.HandleSymbolSector)
		protected.GET("/sectors/:sector/performance", h.HandleSectorPerformance)
		protected.GET("/sectors/:sector/averages", h.HandleSectorAverages)
		protected.GET("/sectors/:sector/stocks", h.HandleSectorStocks)

		// Screener endpoints
		protected.POST("/screener", h.HandleScreener)
		protected.GET("/screener/presets", h.HandleGetPresets)
		protected.POST("/screener/presets", h.HandleSavePreset)
		protected.DELETE("/screener/presets/:id", h.HandleDeletePreset)

		// Comparison endpoints
		protected.GET("/comparison/valuation", h.HandleComparisonValuation)
		protected.GET("/comparison/performance", h.HandleComparisonPerformance)
		protected.GET("/comparison/correlation", h.HandleComparisonCorrelation)

		// Watchlist endpoints
		protected.GET("/watchlists", h.HandleGetWatchlists)
		protected.POST("/watchlists", h.HandleCreateWatchlist)
		protected.PUT("/watchlists/:id", h.HandleRenameWatchlist)
		protected.DELETE("/watchlists/:id", h.HandleDeleteWatchlist)
		protected.POST("/watchlists/:id/symbols", h.HandleAddWatchlistSymbol)
		protected.DELETE("/watchlists/:id/symbols/:symbol", h.HandleRemoveWatchlistSymbol)
		protected.PUT("/watchlists/:id/reorder", h.HandleReorderWatchlist)
		protected.PUT("/watchlists/:id/symbols/:symbol/alert", h.HandleSetPriceAlert)

		// Alert endpoints
		protected.GET("/alerts", h.HandleGetAlerts)
		protected.PUT("/alerts/preferences", h.HandleUpdateAlertPreferences)
		protected.PUT("/alerts/:id/viewed", h.HandleMarkAlertViewed)

		// Knowledge base endpoints
		protected.GET("/knowledge/observations", h.HandleGetObservations)
		protected.GET("/knowledge/accuracy/:patternType", h.HandleGetAccuracy)

		// Goal endpoints
		protected.GET("/goals", h.HandleGetGoals)
		protected.POST("/goals", h.HandleCreateGoal)
		protected.PUT("/goals/:id", h.HandleUpdateGoal)
		protected.DELETE("/goals/:id", h.HandleDeleteGoal)
		protected.GET("/goals/:id/progress", h.HandleGetGoalProgress)

		// Backtest endpoint
		protected.POST("/backtest", h.HandleBacktest)

		// Export endpoints
		protected.GET("/export/transactions", h.HandleExportTransactions)
		protected.GET("/export/snapshot", h.HandleExportSnapshot)
		protected.GET("/export/report", h.HandleExportReport)
		protected.GET("/export/tax", h.HandleExportTax)
	}

	return r
}
