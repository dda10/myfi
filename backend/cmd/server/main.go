package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"myfi-backend/internal/domain/agent"
	"myfi-backend/internal/domain/analyst"
	domainAuth "myfi-backend/internal/domain/auth"
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
	"myfi-backend/internal/platform"

	_ "github.com/dda10/vnstock-go/all"           // Register all connectors
	_ "github.com/dda10/vnstock-go/connector/kbs" // Register KBS connector (missing from all package)
	"github.com/joho/godotenv"
)

func main() {
	// Load .env from project root (best-effort, not fatal if missing)
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Initialize structured JSON logging (configurable via LOG_LEVEL env var)
	infra.InitLogger()

	// Load platform configuration from environment
	cfg := platform.LoadConfig()

	// Initialize database
	db, err := infra.InitDB()
	if err != nil {
		slog.Error("database initialization failed", "err", err)
		os.Exit(1)
	}
	defer infra.CloseDB(db)

	// Initialize infrastructure
	cache := infra.NewCache()
	dsRouter, err := infra.NewDataSourceRouter()
	if err != nil {
		slog.Error("data source router initialization failed", "err", err)
		os.Exit(1)
	}

	// Initialize gRPC client to Python AI Service
	grpcClient, err := infra.NewGRPCClient(slog.Default())
	if err != nil {
		slog.Warn("gRPC client initialization failed, AI features degraded", "err", err)
		// Non-fatal: AI features will be unavailable but the rest of the app works.
	}

	// --- Market domain services ---
	priceSvc := market.NewPriceService(dsRouter, cache)
	sectorSvc := market.NewSectorService(dsRouter, cache)
	marketDataSvc := market.NewMarketDataService(dsRouter, priceSvc, sectorSvc, cache)
	macroSvc := market.NewMacroService(cache, dsRouter)
	searchSvc := market.NewSearchService(marketDataSvc, cache)
	worldMarketSvc := market.NewWorldMarketService(dsRouter, cache)
	goldPriceSvc := market.NewGoldPriceService(dsRouter, cache)
	exchangeRateSvc := market.NewExchangeRateService(cache)
	tradingHoursSvc := market.NewTradingHoursService()

	// --- Auth ---
	authSvc := domainAuth.NewAuthService(db, domainAuth.DefaultAuthConfig())

	// --- Watchlist ---
	watchlistSvc := watchlist.NewWatchlistService(db)

	// --- Screener ---
	screenerSvc := screener.NewScreenerService(dsRouter, sectorSvc, cache, db)

	// --- Liquidity filter ---
	liquidityFilter := screener.NewLiquidityFilter(dsRouter, cache)
	liquidityFilter.Start(context.Background())
	screenerSvc.SetLiquidityFilter(liquidityFilter)

	// --- Portfolio ---
	ledger := portfolio.NewTransactionLedger(db)
	portfolioEngine := portfolio.NewPortfolioEngine(db, ledger, priceSvc)
	performanceEngine := portfolio.NewPerformanceEngine(db, dsRouter)
	riskSvc := portfolio.NewRiskService(db, performanceEngine, dsRouter)
	exportSvc := portfolio.NewExportService(db)

	// --- Knowledge base ---
	knowledgeBase := knowledge.NewKnowledgeBase(db, dsRouter)

	// --- Ranking ---
	recTracker := ranking.NewRecommendationTracker(priceSvc)
	backtestEngine := ranking.NewBacktestEngine()

	// --- Mission ---
	missionSvc := mission.NewMissionService(db)

	// --- Notification ---
	notificationSvc := notification.NewNotificationService(db, nil) // email sender wired later if needed

	// --- Analyst IQ ---
	analystSvc := analyst.NewAnalystIQService(db)

	// --- Sentiment & Consensus ---
	sentimentSvc := sentiment.NewSentimentService(db, cache, nil) // LLM analyzer wired after AI init
	consensusSvc := consensus.NewConsensusService(db, cache)

	// --- Fund (FMarket connector) ---
	fundSvc := fund.NewFundService(dsRouter, cache)

	// --- Build domain handler structs ---
	dh := &platform.DomainHandlers{
		Market: &market.Handlers{
			DataSourceRouter:    dsRouter,
			PriceService:        priceSvc,
			SectorService:       sectorSvc,
			MarketDataService:   marketDataSvc,
			MacroService:        macroSvc,
			SearchService:       searchSvc,
			WorldMarketService:  worldMarketSvc,
			GoldPriceService:    goldPriceSvc,
			ExchangeRateService: exchangeRateSvc,
			TradingHoursService: tradingHoursSvc,
			LiquidityFilter:     liquidityFilter,
			GRPCClient:          grpcClient,
		},
		Portfolio: &portfolio.Handlers{
			Engine:      portfolioEngine,
			Ledger:      ledger,
			Performance: performanceEngine,
			Risk:        riskSvc,
			Export:      exportSvc,
		},
		Screener:     &screener.Handlers{ScreenerService: screenerSvc},
		Watchlist:    &watchlist.Handlers{WatchlistService: watchlistSvc},
		Auth:         &domainAuth.Handlers{AuthService: authSvc},
		Agent:        &agent.Handlers{GRPCClient: grpcClient},
		Ranking:      &ranking.Handlers{Tracker: recTracker, BacktestEngine: backtestEngine, GRPCClient: grpcClient},
		Mission:      &mission.Handlers{MissionService: missionSvc},
		Notification: &notification.Handlers{NotificationService: notificationSvc},
		Knowledge:    &knowledge.Handlers{KnowledgeBase: knowledgeBase},
		Analyst:      &analyst.Handlers{AnalystService: analystSvc},
		Research:     &research.Handlers{},
		Consensus:    &consensus.Handlers{ConsensusService: consensusSvc},
		Sentiment:    &sentiment.Handlers{SentimentService: sentimentSvc},
		Fund:         &fund.Handlers{FundService: fundSvc},
	}

	// --- Build search index at startup ---
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := searchSvc.BuildIndex(ctx); err != nil {
			slog.Error("failed to build search index", "err", err)
		} else {
			slog.Info("search index built successfully")
		}
	}()

	// --- Start scheduler ---
	scheduler := infra.NewScheduler(slog.Default())

	// Register mission evaluation job
	if missionSvc != nil {
		scheduler.RegisterEvalJob("mission-evaluation", func(ctx context.Context) error {
			_, err := missionSvc.EvaluateTriggers(ctx, func(ctx context.Context, symbols []string) (map[string]float64, error) {
				quotes, _, err := dsRouter.FetchRealTimeQuotes(ctx, symbols)
				if err != nil {
					return nil, err
				}
				prices := make(map[string]float64, len(quotes))
				for _, q := range quotes {
					prices[q.Symbol] = q.Close
				}
				return prices, nil
			})
			return err
		})
	}

	// Register recommendation outcome tracking (daily)
	scheduler.RegisterArchivalJob("recommendation-outcomes", func(ctx context.Context) error {
		return recTracker.UpdateOutcomes(ctx)
	})

	// Register knowledge base archival (nightly)
	// Note: S3 storage will be wired when the full archival pipeline is configured.
	// For now, the archival job is a no-op placeholder.
	scheduler.RegisterArchivalJob("knowledge-archival", func(ctx context.Context) error {
		slog.Info("knowledge archival job triggered (storage not yet wired)")
		return nil
	})

	scheduler.Start()
	defer scheduler.Stop()

	// Create router and server through the platform layer
	engine := platform.NewRouter(cfg, dh)
	srv := platform.NewServer(cfg, engine)

	if err := srv.Run(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
