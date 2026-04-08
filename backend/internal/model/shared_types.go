package model

import "myfi-backend/internal/domain/market"

// --- Asset Type (canonical definition in domain/market) ---

// AssetType represents the type of a financial asset.
type AssetType = market.AssetType

const (
	VNStock = market.VNStock
)

// Deprecated asset type constants — retained for backward compatibility during migration.
// These will be removed when portfolio/transaction services are refactored to stock-only.
const (
	Crypto  AssetType = "crypto"
	Gold    AssetType = "gold"
	Savings AssetType = "savings"
	Bond    AssetType = "bond"
	Cash    AssetType = "cash"
)

// FallbackUSDVND is the default USD/VND exchange rate.
// Deprecated: will be removed when export service is refactored.
const FallbackUSDVND float64 = 25000.0

// ValidAssetTypes contains all supported asset types for validation.
var ValidAssetTypes = market.ValidAssetTypes

// ValidateAssetType checks if the given asset type is supported.
var ValidateAssetType = market.ValidateAssetType

// --- ICB Sector (canonical definition in domain/market) ---

// ICBSector represents an ICB sector index code.
type ICBSector = market.ICBSector

const (
	VNIT   = market.VNIT
	VNIND  = market.VNIND
	VNCONS = market.VNCONS
	VNCOND = market.VNCOND
	VNHEAL = market.VNHEAL
	VNENE  = market.VNENE
	VNUTI  = market.VNUTI
	VNREAL = market.VNREAL
	VNFIN  = market.VNFIN
	VNMAT  = market.VNMAT
)

// AllICBSectors contains all 10 ICB sector indices.
var AllICBSectors = market.AllICBSectors

// SectorNameMap maps ICB sector codes to Vietnamese names.
var SectorNameMap = market.SectorNameMap

// --- Sector Trend (canonical definition in domain/market) ---

// SectorTrend represents the trend direction of a sector.
type SectorTrend = market.SectorTrend

const (
	Uptrend   = market.Uptrend
	Downtrend = market.Downtrend
	Sideways  = market.Sideways
)
