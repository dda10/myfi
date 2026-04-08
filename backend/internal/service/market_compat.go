package service

import (
	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/infra"
)

// Type aliases bridging domain/market services into the service package
// for backward compatibility during migration. These will be removed once
// all services are migrated to domain packages.

type PriceService = market.PriceService
type SectorService = market.SectorService
type MarketDataService = market.MarketDataService
type MacroService = market.MacroService

// Constructor wrappers for backward compatibility.

var NewPriceService = market.NewPriceService
var NewSectorService = market.NewSectorService
var NewMacroService = market.NewMacroService

func NewMarketDataService(router *infra.DataSourceRouter, ps *PriceService, ss *SectorService, cache *infra.Cache) *MarketDataService {
	return market.NewMarketDataService(router, ps, ss, cache)
}
