package service

import (
	"myfi-backend/internal/domain/ranking"
)

// Type aliases bridging domain/ranking backtest services into the service package
// for backward compatibility during migration. These will be removed once
// all services are migrated to domain packages.

type BacktestEngine = ranking.BacktestEngine

var NewBacktestEngine = ranking.NewBacktestEngine

// Unexported function wrappers for backward compatibility with existing tests.
var computeWinRate = ranking.ComputeWinRate
var computeMaxDrawdownFromEquity = ranking.ComputeMaxDrawdownFromEquity
var computeSharpeFromEquity = ranking.ComputeSharpeFromEquity
