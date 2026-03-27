package handler

import (
	"database/sql"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/service"

	"github.com/dda10/vnstock-go"
)

// Handlers holds all service dependencies for HTTP handler methods.
// Constructed in cmd/server/main.go and passed to route registration.
// Replaces the 11 package-level var declarations and init() from the flat layout.
type Handlers struct {
	VnstockClient     *vnstock.Client
	DataSourceRouter  *infra.DataSourceRouter
	FXService         *service.FXService
	SharedCache       *infra.Cache
	GoldService       *service.GoldService
	PriceService      *service.PriceService
	SectorService     *service.SectorService
	MarketDataService *service.MarketDataService
	FundService       *service.FundService
	CommodityService  *service.CommodityService
	MacroService      *service.MacroService
	CryptoService     *service.CryptoService
	DB                *sql.DB

	// Portfolio and tracking services
	WatchlistService  *service.WatchlistService
	ScreenerService   *service.ScreenerService
	PortfolioEngine   *service.PortfolioEngine
	PerformanceEngine *service.PerformanceEngine
	ComparisonEngine  *service.ComparisonEngine
	TransactionLedger *service.TransactionLedger
	AssetRegistry     *service.AssetRegistry
	SavingsTracker    *service.SavingsTracker

	// Market quality filter
	LiquidityFilter *service.LiquidityFilter

	// AI recommendation tracking
	RecommendationTracker *service.RecommendationTracker

	// Signal engine for systematic recommendations
	SignalEngine *service.SignalEngine

	// Signal backtester for historical simulation
	SignalBacktester *service.SignalBacktester

	// Authentication service
	AuthService *service.AuthService

	// Risk, goals, backtest, export, alerts, knowledge
	RiskService    *service.RiskService
	GoalPlanner    *service.GoalPlanner
	BacktestEngine *service.BacktestEngine
	ExportService  *service.ExportService
	AlertService   *service.AlertService
	KnowledgeBase  *service.KnowledgeBase
}
