package model

import (
	"myfi-backend/internal/domain/ranking"
)

// --- Type aliases bridging domain/ranking backtest types into model for backward compatibility ---
// These will be removed once all services are migrated to domain packages.

type IndicatorType = ranking.IndicatorType

const (
	IndicatorSMA            = ranking.IndicatorSMA
	IndicatorEMA            = ranking.IndicatorEMA
	IndicatorRSI            = ranking.IndicatorRSI
	IndicatorMACD           = ranking.IndicatorMACD
	IndicatorBollingerBands = ranking.IndicatorBollingerBands
	IndicatorStochastic     = ranking.IndicatorStochastic
	IndicatorADX            = ranking.IndicatorADX
	IndicatorAroon          = ranking.IndicatorAroon
	IndicatorParabolicSAR   = ranking.IndicatorParabolicSAR
	IndicatorSupertrend     = ranking.IndicatorSupertrend
	IndicatorVWAP           = ranking.IndicatorVWAP
	IndicatorVWMA           = ranking.IndicatorVWMA
	IndicatorWilliamsR      = ranking.IndicatorWilliamsR
	IndicatorCMO            = ranking.IndicatorCMO
	IndicatorROC            = ranking.IndicatorROC
	IndicatorMomentum       = ranking.IndicatorMomentum
	IndicatorKeltnerChannel = ranking.IndicatorKeltnerChannel
	IndicatorATR            = ranking.IndicatorATR
	IndicatorStdDev         = ranking.IndicatorStdDev
	IndicatorOBV            = ranking.IndicatorOBV
	IndicatorLinearReg      = ranking.IndicatorLinearReg
)

type ConditionOperator = ranking.ConditionOperator

const (
	OpLessThan     = ranking.OpLessThan
	OpGreaterThan  = ranking.OpGreaterThan
	OpCrossesAbove = ranking.OpCrossesAbove
	OpCrossesBelow = ranking.OpCrossesBelow
	OpLessEqual    = ranking.OpLessEqual
	OpGreaterEqual = ranking.OpGreaterEqual
)

type ConditionOperand = ranking.ConditionOperand
type StrategyCondition = ranking.StrategyCondition
type StrategyRule = ranking.StrategyRule
type BacktestRequest = ranking.BacktestRequest
type BacktestTrade = ranking.BacktestTrade
type BacktestResult = ranking.BacktestResult
type EquityPoint = ranking.EquityPoint
