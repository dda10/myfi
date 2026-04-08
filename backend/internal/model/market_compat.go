package model

import (
	"myfi-backend/internal/domain/market"
)

// --- Type aliases bridging domain/market types into model for backward compatibility ---
// These allow existing service/ code to reference model.X while types live in domain/market.
// They will be removed once all services are migrated to domain packages.

// OHLCVBar represents a single OHLCV data point.
type OHLCVBar = market.OHLCVBar

// PriceQuote represents a price quote for a Vietnamese stock.
type PriceQuote = market.PriceQuote

// StockQuote is an alias for PriceQuote.
type StockQuote = market.StockQuote

// MacroIndicator represents a single macroeconomic indicator.
type MacroIndicator = market.MacroIndicator

// MacroData contains all macroeconomic indicators.
type MacroData = market.MacroData

// ListingData contains all listing-related information.
type ListingData = market.ListingData

// CompanyData contains all company-related information.
type CompanyData = market.CompanyData

// FinancialReportData contains all financial report data for a symbol.
type FinancialReportData = market.FinancialReportData

// TradingStatistics contains trading data for a symbol.
type TradingStatistics = market.TradingStatistics

// MarketStatistics contains market-level statistics.
type MarketStatistics = market.MarketStatistics

// ValuationMetrics contains valuation data at various levels.
type ValuationMetrics = market.ValuationMetrics

// SectorPerformance holds performance metrics for a single ICB sector.
type SectorPerformance = market.SectorPerformance

// SectorAverages holds median fundamental metrics for stocks in a sector.
type SectorAverages = market.SectorAverages

// ForeignTradingData represents foreign investor trading data.
type ForeignTradingData = market.ForeignTradingData
