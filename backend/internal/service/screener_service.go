package service

import (
	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/domain/screener"
	"myfi-backend/internal/infra"

	"database/sql"
)

// Type aliases bridging domain/screener services into the service package
// for backward compatibility during migration.

type ScreenerService = screener.ScreenerService
type LiquidityFilter = screener.LiquidityFilter

// Constructor wrappers for backward compatibility.

var NewLiquidityFilter = screener.NewLiquidityFilter
var NewLiquidityFilterWithConfig = screener.NewLiquidityFilterWithConfig

func NewScreenerService(router *infra.DataSourceRouter, sectorService *market.SectorService, cache *infra.Cache, database *sql.DB) *ScreenerService {
	return screener.NewScreenerService(router, sectorService, cache, database)
}
