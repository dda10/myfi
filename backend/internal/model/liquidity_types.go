package model

import "myfi-backend/internal/domain/screener"

// --- Type aliases bridging domain/screener liquidity types into model for backward compatibility ---

type LiquidityConfig = screener.LiquidityConfig
type LiquidityTier = screener.LiquidityTier
type LiquidityScore = screener.LiquidityScore
type LiquiditySnapshot = screener.LiquiditySnapshot

var DefaultLiquidityConfig = screener.DefaultLiquidityConfig

type WhitelistEntry = screener.WhitelistEntry
type WhitelistSnapshot = screener.WhitelistSnapshot
