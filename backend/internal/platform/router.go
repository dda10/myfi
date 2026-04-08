package platform

import (
	"myfi-backend/internal/domain/agent"
	"myfi-backend/internal/domain/analyst"
	"myfi-backend/internal/domain/auth"
	"myfi-backend/internal/domain/consensus"
	"myfi-backend/internal/domain/fund"
	"myfi-backend/internal/domain/knowledge"
	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/domain/mission"
	"myfi-backend/internal/domain/notification"
	"myfi-backend/internal/domain/portfolio"
	"myfi-backend/internal/domain/ranking"
	"myfi-backend/internal/domain/research"
	"myfi-backend/internal/domain/screener"
	"myfi-backend/internal/domain/sentiment"
	"myfi-backend/internal/domain/watchlist"
	"myfi-backend/internal/infra"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DomainHandlers aggregates all domain handler structs for route registration.
// Constructed in cmd/server/main.go and passed to NewRouter.
type DomainHandlers struct {
	Market       *market.Handlers
	Portfolio    *portfolio.Handlers
	Screener     *screener.Handlers
	Watchlist    *watchlist.Handlers
	Auth         *auth.Handlers
	Agent        *agent.Handlers
	Ranking      *ranking.Handlers
	Mission      *mission.Handlers
	Notification *notification.Handlers
	Knowledge    *knowledge.Handlers
	Analyst      *analyst.Handlers
	Research     *research.Handlers
	Consensus    *consensus.Handlers
	Sentiment    *sentiment.Handlers
	Fund         *fund.Handlers
}

// NewRouter creates a Gin engine with middleware and all domain route registrations.
//
// Dependency direction: cmd → platform → domain → infra (acyclic)
func NewRouter(cfg Config, dh *DomainHandlers) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// --- Global middleware ---
	r.Use(infra.JSONLoggerMiddleware()) // Structured JSON request logging (Req 50.9)
	r.Use(RecoveryMiddleware())
	r.Use(GlobalErrorHandler())
	r.Use(infra.MetricsMiddleware())

	// HTTPS redirect in production
	if cfg.IsProduction() {
		r.Use(HTTPSRedirectMiddleware())
	}

	// Security headers: X-Content-Type-Options, X-Frame-Options, HSTS
	r.Use(securityHeadersMiddleware(cfg))

	// Input validation middleware (reject oversized bodies, sanitize)
	r.Use(inputValidationMiddleware())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"},
		AllowCredentials: true,
	}))

	// Per-IP rate limiting
	r.Use(PerIPRateLimitMiddleware(cfg.IPRateLimit))

	// --- Public routes (no JWT required) ---
	r.GET("/api/health", dh.Market.HandleHealth)
	r.GET("/api/healthz", dh.Market.HandleHealthz)
	r.GET("/api/readyz", dh.Market.HandleReadyz)
	r.POST("/api/auth/register", dh.Auth.HandleRegister)
	r.POST("/api/auth/login", dh.Auth.HandleLogin)

	// --- API Documentation (no JWT required) ---
	RegisterDocsRoutes(r)

	// --- Prometheus metrics (no JWT required) ---
	r.GET("/metrics", infra.PrometheusHandler())

	// --- Protected routes (JWT required) ---
	protected := r.Group("/api")
	protected.Use(dh.Auth.JWTMiddleware())
	protected.Use(PerUserRateLimitMiddleware(cfg.UserRateLimit))
	{
		// Auth
		protected.POST("/auth/logout", dh.Auth.HandleLogout)
		protected.POST("/auth/change-password", dh.Auth.HandleChangePassword)
		protected.GET("/auth/me", dh.Auth.HandleMe)

		// User settings
		protected.GET("/settings", dh.Auth.HandleGetSettings)
		protected.PUT("/settings", dh.Auth.HandleUpdateSettings)

		// Market data
		protected.GET("/market/quote", dh.Market.HandleMarketQuote)
		protected.GET("/market/chart", dh.Market.HandleMarketChart)
		protected.GET("/market/listing", dh.Market.HandleUnifiedListing)
		protected.GET("/market/screener", dh.Market.HandleMarketScreener)
		protected.GET("/market/company/:symbol", dh.Market.HandleCompanyData)
		protected.GET("/market/finance/:symbol", dh.Market.HandleFinancialReports)
		protected.GET("/market/trading/:symbol", dh.Market.HandleTradingStatistics)
		protected.GET("/market/statistics", dh.Market.HandleMarketStatistics)
		protected.GET("/market/valuation", dh.Market.HandleValuationMetrics)
		protected.GET("/market/macro", dh.Market.HandleMacro)
		protected.POST("/market/trading/batch", dh.Market.HandleBatchTradingQuotes)
		protected.GET("/market/search", dh.Market.HandleSearch)
		protected.GET("/market/index", dh.Market.HandleMarketIndex)
		protected.GET("/market/hot-topics", dh.Market.HandleHotTopics)
		protected.GET("/market/ratios/:symbol", dh.Market.HandleMarketRatios)
		protected.GET("/market/technical/:symbol", dh.Market.HandleMarketTechnical)

		// Macro indicators
		protected.GET("/market/macro/indicators", dh.Market.HandleMacroIndicators)
		protected.GET("/market/macro/interbank", dh.Market.HandleInterbankRates)
		protected.GET("/market/macro/bonds", dh.Market.HandleBondYields)
		protected.GET("/market/macro/fx", dh.Market.HandleFXRates)

		// World markets, gold, exchange rates, trading hours (vnstock-go v2)
		protected.GET("/market/world-indices", dh.Market.HandleWorldIndices)
		protected.GET("/market/gold", dh.Market.HandleGoldPrices)
		protected.GET("/market/exchange-rates", dh.Market.HandleExchangeRates)
		protected.GET("/market/status", dh.Market.HandleMarketStatus)

		// Price board, price depth, shareholders, subsidiaries (vnstock-go v2)
		protected.GET("/market/price-board", dh.Market.HandlePriceBoard)
		protected.GET("/market/price-depth", dh.Market.HandlePriceDepth)
		protected.GET("/market/shareholders", dh.Market.HandleShareholders)
		protected.GET("/market/subsidiaries", dh.Market.HandleSubsidiaries)

		// Liquidity whitelist
		protected.GET("/market/whitelist", dh.Market.HandleWhitelist)
		protected.GET("/market/whitelist/check", dh.Market.HandleWhitelistCheck)
		protected.POST("/market/whitelist/refresh", dh.Market.HandleWhitelistRefresh)

		// Rate limit metrics
		protected.GET("/metrics/rate-limits", dh.Market.HandleRateLimitMetrics)

		// Sectors
		protected.GET("/sectors/performance", dh.Market.HandleAllSectorPerformances)
		protected.GET("/sectors/symbol/:symbol", dh.Market.HandleSymbolSector)
		protected.GET("/sectors/:sector/performance", dh.Market.HandleSectorPerformance)
		protected.GET("/sectors/:sector/averages", dh.Market.HandleSectorAverages)
		protected.GET("/sectors/:sector/trend", dh.Market.HandleSectorTrend)
		protected.GET("/sectors/:sector/stocks", dh.Market.HandleSectorStocks)
		protected.GET("/sectors/mapping", dh.Market.HandleStockSectorMapping)

		// Portfolio
		protected.GET("/portfolio/holdings", dh.Portfolio.HandleGetHoldings)
		protected.GET("/portfolio/nav", dh.Portfolio.HandleGetNAV)
		protected.POST("/portfolio/buy", dh.Portfolio.HandleRecordBuy)
		protected.POST("/portfolio/sell", dh.Portfolio.HandleRecordSell)
		protected.GET("/portfolio/transactions", dh.Portfolio.HandleGetTransactions)
		protected.GET("/portfolio/performance", dh.Portfolio.HandleGetPerformance)
		protected.GET("/portfolio/risk", dh.Portfolio.HandleGetRisk)
		protected.GET("/portfolio/allocation", dh.Portfolio.HandleGetAllocation)
		protected.GET("/portfolio/pnl", dh.Portfolio.HandleGetPnL)

		// Portfolio corporate actions
		protected.GET("/portfolio/corporate/upcoming", dh.Portfolio.HandleUpcomingCorporateEvents)
		protected.GET("/portfolio/corporate/:symbol", dh.Portfolio.HandleSymbolCorporateEvents)

		// Export
		protected.GET("/export/transactions", dh.Portfolio.HandleExportTransactions)
		protected.GET("/export/portfolio", dh.Portfolio.HandleExportPortfolio)
		protected.GET("/export/pnl", dh.Portfolio.HandleExportPnL)
		protected.GET("/export/tax", dh.Portfolio.HandleExportTax)

		// Screener
		protected.POST("/screener", dh.Screener.HandleScreener)
		protected.GET("/screener/presets", dh.Screener.HandleGetPresets)
		protected.POST("/screener/presets", dh.Screener.HandleSavePreset)
		protected.DELETE("/screener/presets/:id", dh.Screener.HandleDeletePreset)

		// Watchlists
		protected.GET("/watchlists", dh.Watchlist.HandleGetWatchlists)
		protected.POST("/watchlists", dh.Watchlist.HandleCreateWatchlist)
		protected.PUT("/watchlists/:id", dh.Watchlist.HandleRenameWatchlist)
		protected.DELETE("/watchlists/:id", dh.Watchlist.HandleDeleteWatchlist)
		protected.POST("/watchlists/:id/symbols", dh.Watchlist.HandleAddWatchlistSymbol)
		protected.DELETE("/watchlists/:id/symbols/:symbol", dh.Watchlist.HandleRemoveWatchlistSymbol)
		protected.PUT("/watchlists/:id/reorder", dh.Watchlist.HandleReorderWatchlist)
		protected.PUT("/watchlists/:id/symbols/:symbol/alert", dh.Watchlist.HandleSetPriceAlert)

		// AI chat & analysis
		protected.POST("/chat", dh.Agent.HandleChat)
		protected.POST("/analyze/:symbol", dh.Agent.HandleAnalyzeStock)
		protected.GET("/analyze/:symbol", dh.Agent.HandleAnalyzeStock)

		// Ranking & backtest
		protected.POST("/ranking", dh.Ranking.HandleRanking)
		protected.GET("/ranking", dh.Ranking.HandleGetRankingList)
		protected.POST("/ranking/backtest", dh.Ranking.HandleBacktest)
		protected.GET("/ranking/factors", dh.Ranking.HandleGetFactors)
		protected.GET("/ranking/recommendations", dh.Ranking.HandleGetRecommendations)
		protected.GET("/ranking/accuracy", dh.Ranking.HandleGetAccuracy)
		protected.GET("/ideas", dh.Ranking.HandleGetIdeas)

		// Missions
		protected.POST("/missions", dh.Mission.HandleCreateMission)
		protected.GET("/missions", dh.Mission.HandleListMissions)
		protected.GET("/missions/:id", dh.Mission.HandleGetMission)
		protected.PUT("/missions/:id", dh.Mission.HandleUpdateMission)
		protected.DELETE("/missions/:id", dh.Mission.HandleDeleteMission)
		protected.POST("/missions/:id/pause", dh.Mission.HandlePauseMission)
		protected.POST("/missions/:id/resume", dh.Mission.HandleResumeMission)

		// Notifications
		protected.GET("/notifications", dh.Notification.HandleListNotifications)
		protected.PUT("/notifications/:id/read", dh.Notification.HandleMarkRead)
		protected.PUT("/notifications/read-all", dh.Notification.HandleMarkAllRead)
		protected.GET("/notifications/unread", dh.Notification.HandleUnreadCount)

		// Knowledge base
		protected.GET("/knowledge/observations", dh.Knowledge.HandleGetObservations)
		protected.GET("/knowledge/accuracy/:patternType", dh.Knowledge.HandleGetAccuracy)

		// Analyst IQ
		protected.GET("/analyst/consensus/:symbol", dh.Analyst.HandleGetConsensus)
		protected.GET("/analyst/reports/:symbol", dh.Analyst.HandleGetReports)
		protected.GET("/analyst/accuracy/:analyst", dh.Analyst.HandleGetAnalystAccuracy)

		// Research reports
		protected.GET("/research/reports", dh.Research.HandleListReports)
		protected.GET("/research/reports/:id", dh.Research.HandleGetReport)
		protected.GET("/research/reports/:id/pdf", dh.Research.HandleDownloadReportPDF)

		// Sentiment analysis
		if dh.Sentiment != nil {
			protected.POST("/sentiment/analyze", dh.Sentiment.HandleAnalyze)
			protected.GET("/sentiment/trend", dh.Sentiment.HandleGetTrend)
			protected.GET("/sentiment/timeseries", dh.Sentiment.HandleGetTimeSeries)
			protected.GET("/sentiment/articles", dh.Sentiment.HandleGetArticles)
		}

		// Market consensus
		if dh.Consensus != nil {
			protected.GET("/consensus/score", dh.Consensus.HandleGetConsensus)
			protected.GET("/consensus/divergences", dh.Consensus.HandleGetDivergences)
			protected.GET("/consensus/mood", dh.Consensus.HandleGetMarketMood)
		}

		// Mutual funds (FMarket connector)
		if dh.Fund != nil {
			protected.GET("/funds", dh.Fund.HandleListFunds)
			protected.GET("/funds/:code", dh.Fund.HandleGetFund)
			protected.GET("/funds/:code/holdings", dh.Fund.HandleGetFundHoldings)
			protected.GET("/funds/:code/nav", dh.Fund.HandleGetFundNAV)
		}
	}

	return r
}
