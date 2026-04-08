package handler

import (
	"database/sql"

	"myfi-backend/internal/domain/consensus"
	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/domain/sentiment"
	"myfi-backend/internal/infra"
	"myfi-backend/internal/service"

	"github.com/dda10/vnstock-go"
)

// Handlers holds all service dependencies for HTTP handler methods.
// Constructed in cmd/server/main.go and passed to route registration.
type Handlers struct {
	// Market domain handlers (embedded for method promotion)
	*market.Handlers

	VnstockClient    *vnstock.Client
	DataSourceRouter *infra.DataSourceRouter
	SharedCache      *infra.Cache
	DB               *sql.DB

	// Portfolio and tracking services
	WatchlistService  *service.WatchlistService
	ScreenerService   *service.ScreenerService
	PortfolioEngine   *service.PortfolioEngine
	PerformanceEngine *service.PerformanceEngine
	TransactionLedger *service.TransactionLedger

	// Market quality filter
	LiquidityFilter *service.LiquidityFilter

	// AI recommendation tracking
	RecommendationTracker *service.RecommendationTracker

	// Authentication service
	AuthService *service.AuthService

	// Risk, backtest, export, knowledge
	RiskService    *service.RiskService
	BacktestEngine *service.BacktestEngine
	ExportService  *service.ExportService
	KnowledgeBase  *service.KnowledgeBase

	// Sentiment & consensus (LLM-powered qualitative analysis)
	SentimentHandlers *sentiment.Handlers
	ConsensusHandlers *consensus.Handlers
}
