package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"myfi-backend/internal/handler"
	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
	"myfi-backend/internal/service"

	"github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all" // Register all connectors
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Initialize database
	db, err := infra.InitDB()
	if err != nil {
		slog.Error("database initialization failed", "err", err)
		os.Exit(1)
	}
	defer infra.CloseDB(db)

	// Initialize infrastructure
	cache := infra.NewCache()
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		slog.Error("data source router initialization failed", "err", err)
		os.Exit(1)
	}
	cb := infra.NewCircuitBreaker(3, 60*time.Second)

	// Initialize vnstock client for direct API calls (used by agent handler)
	vnstockClient, err := vnstock.New(vnstock.Config{
		Connector: "VCI",
		Timeout:   30 * time.Second,
	})
	if err != nil {
		slog.Error("failed to initialize vnstock client", "err", err)
		os.Exit(1)
	}

	// Initialize services
	fxSvc := service.NewFXService(cache, router.RateLimiter(), cb)
	goldSvc, err := service.NewGoldService(cache, router.RateLimiter())
	if err != nil {
		slog.Error("gold service initialization failed", "err", err)
		os.Exit(1)
	}
	cryptoSvc := service.NewCryptoService(cache, router.RateLimiter(), infra.NewCircuitBreaker(3, 60*time.Second))
	priceSvc := service.NewPriceService(router, cache, goldSvc)
	sectorSvc := service.NewSectorService(router, cache)
	marketDataSvc := service.NewMarketDataService(router, priceSvc, sectorSvc, cache)
	fundSvc := service.NewFundService(cache)
	commoditySvc := service.NewCommodityService(goldSvc, cache)
	macroSvc := service.NewMacroService(cache)

	// Initialize DB-backed services
	authSvc := service.NewAuthService(db, model.DefaultAuthConfig())
	watchlistSvc := service.NewWatchlistService(db)
	screenerSvc := service.NewScreenerService(router, sectorSvc, cache, db)
	savingsTracker := service.NewSavingsTracker(db)
	ledger := service.NewTransactionLedger(db)
	registry := service.NewAssetRegistry(db, priceSvc)
	portfolioEngine := service.NewPortfolioEngine(registry, ledger, priceSvc)
	performanceEngine := service.NewPerformanceEngine(db, router)
	comparisonEngine := service.NewComparisonEngine(router, sectorSvc, cache)

	// Initialize new services
	riskSvc := service.NewRiskService(db, performanceEngine, router)
	goalPlanner := service.NewGoalPlanner(db)
	backtestEngine := service.NewBacktestEngine()
	fxRateData, _ := fxSvc.GetUSDVNDRate()
	fxRate := 25400.0
	if fxRateData != nil && fxRateData.Rate > 0 {
		fxRate = fxRateData.Rate
	}
	exportSvc := service.NewExportService(fxRate)
	alertSvc := service.NewAlertService(db)
	knowledgeBase := service.NewKnowledgeBase(db, router)

	// Initialize liquidity filter (dynamic whitelist)
	liquidityFilter := service.NewLiquidityFilter(router, cache)
	liquidityFilter.Start(context.Background())

	// Attach liquidity filter to screener for pre-filtering
	screenerSvc.SetLiquidityFilter(liquidityFilter)

	// Initialize recommendation tracker for AI signal accuracy
	recTracker := service.NewRecommendationTracker(priceSvc)

	// Initialize signal engine for systematic recommendations
	signalEngine := service.NewSignalEngine(router, liquidityFilter, sectorSvc, screenerSvc)
	signalEngine.SetRecommendationTracker(recTracker)

	// Initialize signal backtester for historical simulation
	signalBacktester := service.NewSignalBacktester(router, liquidityFilter, sectorSvc, screenerSvc)

	// Construct handlers with all dependencies
	h := &handler.Handlers{
		VnstockClient:     vnstockClient,
		DataSourceRouter:  router,
		FXService:         fxSvc,
		SharedCache:       cache,
		GoldService:       goldSvc,
		PriceService:      priceSvc,
		SectorService:     sectorSvc,
		MarketDataService: marketDataSvc,
		FundService:       fundSvc,
		CommodityService:  commoditySvc,
		MacroService:      macroSvc,
		CryptoService:     cryptoSvc,
		DB:                db,
		AuthService:       authSvc,
		WatchlistService:  watchlistSvc,
		ScreenerService:   screenerSvc,
		PortfolioEngine:   portfolioEngine,
		PerformanceEngine: performanceEngine,
		ComparisonEngine:  comparisonEngine,
		TransactionLedger: ledger,
		AssetRegistry:     registry,
		SavingsTracker:    savingsTracker,
		LiquidityFilter:   liquidityFilter,

		// AI recommendation tracking
		RecommendationTracker: recTracker,
		SignalEngine:          signalEngine,
		SignalBacktester:      signalBacktester,

		// New services
		RiskService:    riskSvc,
		GoalPlanner:    goalPlanner,
		BacktestEngine: backtestEngine,
		ExportService:  exportSvc,
		AlertService:   alertSvc,
		KnowledgeBase:  knowledgeBase,
	}

	// Initialize AI chat
	h.InitLLM()

	// Start background task to update recommendation outcomes daily
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := recTracker.UpdateOutcomes(ctx); err != nil {
				slog.Error("failed to update recommendation outcomes", "err", err)
			}
			cancel()
		}
	}()

	// Set up router and start server
	r := handler.SetupRouter(h)
	slog.Info("starting server", "port", 8080)
	if err := r.Run(":8080"); err != nil {
		slog.Error("server failed to start", "err", err)
		os.Exit(1)
	}
}
