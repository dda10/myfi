package handler

import (
	"myfi-backend/internal/infra"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter creates a Gin engine with CORS middleware and all route registrations.
func SetupRouter(h *Handlers) *gin.Engine {
	r := gin.Default()

	// Global panic recovery middleware
	r.Use(infra.RecoveryMiddleware())

	// HTTPS redirect in production
	if os.Getenv("ENV") == "production" {
		r.Use(infra.HTTPSRedirectMiddleware())
	}

	// CORS — restrict to frontend origin in production
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

	// Per-IP rate limiting for all endpoints
	r.Use(infra.PerIPRateLimitMiddleware(200))

	// Health check — public, no auth required
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public auth endpoints (no JWT required)
	r.POST("/api/auth/register", h.HandleRegister)
	r.POST("/api/auth/login", h.HandleLogin)

	// All other endpoints require JWT authentication
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
		protected.GET("/market/macro", h.HandleMacro)
		protected.POST("/market/trading/batch", h.HandleBatchTradingQuotes)

		// Liquidity whitelist endpoints
		protected.GET("/market/whitelist", h.HandleWhitelist)
		protected.GET("/market/whitelist/check", h.HandleWhitelistCheck)
		protected.POST("/market/whitelist/refresh", h.HandleWhitelistRefresh)

		// AI chat
		protected.POST("/chat", h.HandleChat)
		protected.POST("/models", h.HandleModels)

		// AI recommendation tracking endpoints — TODO: replaced by ranking handler in Task 12

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

		// Watchlist endpoints
		protected.GET("/watchlists", h.HandleGetWatchlists)
		protected.POST("/watchlists", h.HandleCreateWatchlist)
		protected.PUT("/watchlists/:id", h.HandleRenameWatchlist)
		protected.DELETE("/watchlists/:id", h.HandleDeleteWatchlist)
		protected.POST("/watchlists/:id/symbols", h.HandleAddWatchlistSymbol)
		protected.DELETE("/watchlists/:id/symbols/:symbol", h.HandleRemoveWatchlistSymbol)
		protected.PUT("/watchlists/:id/reorder", h.HandleReorderWatchlist)
		protected.PUT("/watchlists/:id/symbols/:symbol/alert", h.HandleSetPriceAlert)

		// Knowledge base endpoints
		protected.GET("/knowledge/observations", h.HandleGetObservations)
		protected.GET("/knowledge/accuracy/:patternType", h.HandleGetAccuracy)

		// Export endpoints
		protected.GET("/export/transactions", h.HandleExportTransactions)
		protected.GET("/export/snapshot", h.HandleExportSnapshot)
		protected.GET("/export/report", h.HandleExportReport)
		protected.GET("/export/tax", h.HandleExportTax)

		// Sentiment analysis endpoints (LLM-powered Vietnamese news sentiment)
		if h.SentimentHandlers != nil {
			protected.POST("/sentiment/analyze", h.SentimentHandlers.HandleAnalyze)
			protected.GET("/sentiment/trend", h.SentimentHandlers.HandleGetTrend)
			protected.GET("/sentiment/timeseries", h.SentimentHandlers.HandleGetTimeSeries)
			protected.GET("/sentiment/articles", h.SentimentHandlers.HandleGetArticles)
		}

		// Market consensus endpoints (multi-source sentiment aggregation)
		if h.ConsensusHandlers != nil {
			protected.GET("/consensus/score", h.ConsensusHandlers.HandleGetConsensus)
			protected.GET("/consensus/divergences", h.ConsensusHandlers.HandleGetDivergences)
			protected.GET("/consensus/mood", h.ConsensusHandlers.HandleGetMarketMood)
		}
	}

	return r
}
