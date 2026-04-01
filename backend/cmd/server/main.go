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
	priceSvc := service.NewPriceService(router, cache, nil)
	sectorSvc := service.NewSectorService(router, cache)
	marketDataSvc := service.NewMarketDataService(router, priceSvc, sectorSvc, cache)
	macroSvc := service.NewMacroService(cache)

	// Initialize DB-backed services
	authSvc := service.NewAuthService(db, model.DefaultAuthConfig())
	watchlistSvc := service.NewWatchlistService(db)
	screenerSvc := service.NewScreenerService(router, sectorSvc, cache, db)
	ledger := service.NewTransactionLedger(db)
	portfolioEngine := service.NewPortfolioEngine(nil, ledger, priceSvc)
	performanceEngine := service.NewPerformanceEngine(db, router)

	// Initialize new services
	riskSvc := service.NewRiskService(db, performanceEngine, router)
	backtestEngine := service.NewBacktestEngine()
	exportSvc := service.NewExportService(25400.0) // default USD/VND rate
	knowledgeBase := service.NewKnowledgeBase(db, router)

	// Initialize liquidity filter (dynamic whitelist)
	liquidityFilter := service.NewLiquidityFilter(router, cache)
	liquidityFilter.Start(context.Background())

	// Attach liquidity filter to screener for pre-filtering
	screenerSvc.SetLiquidityFilter(liquidityFilter)

	// Initialize recommendation tracker for AI signal accuracy
	recTracker := service.NewRecommendationTracker(priceSvc)

	// Construct handlers with all dependencies
	h := &handler.Handlers{
		VnstockClient:     vnstockClient,
		DataSourceRouter:  router,
		SharedCache:       cache,
		PriceService:      priceSvc,
		SectorService:     sectorSvc,
		MarketDataService: marketDataSvc,
		MacroService:      macroSvc,
		DB:                db,
		AuthService:       authSvc,
		WatchlistService:  watchlistSvc,
		ScreenerService:   screenerSvc,
		PortfolioEngine:   portfolioEngine,
		PerformanceEngine: performanceEngine,
		TransactionLedger: ledger,
		LiquidityFilter:   liquidityFilter,

		// AI recommendation tracking
		RecommendationTracker: recTracker,

		// Services
		RiskService:    riskSvc,
		BacktestEngine: backtestEngine,
		ExportService:  exportSvc,
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
