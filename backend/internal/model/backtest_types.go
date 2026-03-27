package model

import "time"

// --- Backtest Engine Types ---
// Requirement 32: Backtesting and Strategy Simulation

// IndicatorType represents a supported technical indicator.
type IndicatorType string

const (
	IndicatorSMA            IndicatorType = "SMA"
	IndicatorEMA            IndicatorType = "EMA"
	IndicatorRSI            IndicatorType = "RSI"
	IndicatorMACD           IndicatorType = "MACD"
	IndicatorBollingerBands IndicatorType = "BOLLINGER"
	IndicatorStochastic     IndicatorType = "STOCHASTIC"
	IndicatorADX            IndicatorType = "ADX"
	IndicatorAroon          IndicatorType = "AROON"
	IndicatorParabolicSAR   IndicatorType = "PARABOLIC_SAR"
	IndicatorSupertrend     IndicatorType = "SUPERTREND"
	IndicatorVWAP           IndicatorType = "VWAP"
	IndicatorVWMA           IndicatorType = "VWMA"
	IndicatorWilliamsR      IndicatorType = "WILLIAMS_R"
	IndicatorCMO            IndicatorType = "CMO"
	IndicatorROC            IndicatorType = "ROC"
	IndicatorMomentum       IndicatorType = "MOMENTUM"
	IndicatorKeltnerChannel IndicatorType = "KELTNER"
	IndicatorATR            IndicatorType = "ATR"
	IndicatorStdDev         IndicatorType = "STDDEV"
	IndicatorOBV            IndicatorType = "OBV"
	IndicatorLinearReg      IndicatorType = "LINEAR_REG"
)

// ConditionOperator represents a comparison operator in a strategy condition.
type ConditionOperator string

const (
	OpLessThan     ConditionOperator = "LT"
	OpGreaterThan  ConditionOperator = "GT"
	OpCrossesAbove ConditionOperator = "CROSSES_ABOVE"
	OpCrossesBelow ConditionOperator = "CROSSES_BELOW"
	OpLessEqual    ConditionOperator = "LTE"
	OpGreaterEqual ConditionOperator = "GTE"
)

// ConditionOperand represents one side of a condition (indicator value, price, or constant).
type ConditionOperand struct {
	Type      string        `json:"type"` // "indicator", "price", "constant"
	Indicator IndicatorType `json:"indicator,omitempty"`
	Field     string        `json:"field,omitempty"`  // e.g. "value", "signal", "upper", "lower", "bandwidth"
	Period    int           `json:"period,omitempty"` // indicator period
	Param2    int           `json:"param2,omitempty"` // secondary param (e.g. MACD signal period)
	Param3    int           `json:"param3,omitempty"` // tertiary param
	ParamF    float64       `json:"paramF,omitempty"` // float param (e.g. Bollinger multiplier)
	Constant  float64       `json:"constant,omitempty"`
}

// StrategyCondition represents a single condition in a strategy rule.
type StrategyCondition struct {
	Left     ConditionOperand  `json:"left"`
	Operator ConditionOperator `json:"operator"`
	Right    ConditionOperand  `json:"right"`
}

// StrategyRule defines a complete backtest strategy.
type StrategyRule struct {
	Name            string              `json:"name"`
	EntryConditions []StrategyCondition `json:"entryConditions"` // AND-combined
	ExitConditions  []StrategyCondition `json:"exitConditions"`  // AND-combined
	StopLossPct     float64             `json:"stopLossPct"`     // e.g. 0.05 = 5%
	TakeProfitPct   float64             `json:"takeProfitPct"`   // e.g. 0.10 = 10%
}

// BacktestRequest is the input for running a backtest.
type BacktestRequest struct {
	Symbol    string       `json:"symbol"`
	StartDate time.Time    `json:"startDate"`
	EndDate   time.Time    `json:"endDate"`
	Strategy  StrategyRule `json:"strategy"`
}

// BacktestTrade represents a single trade executed during the backtest.
type BacktestTrade struct {
	EntryDate   time.Time `json:"entryDate"`
	ExitDate    time.Time `json:"exitDate"`
	EntryPrice  float64   `json:"entryPrice"`
	ExitPrice   float64   `json:"exitPrice"`
	ReturnPct   float64   `json:"returnPct"`
	ExitReason  string    `json:"exitReason"` // "signal", "stop_loss", "take_profit", "end_of_data"
	HoldingDays int       `json:"holdingDays"`
}

// BacktestResult contains the full output of a backtest simulation.
type BacktestResult struct {
	TotalReturn    float64         `json:"totalReturn"`
	WinRate        float64         `json:"winRate"`
	MaxDrawdown    float64         `json:"maxDrawdown"`
	SharpeRatio    float64         `json:"sharpeRatio"`
	NumTrades      int             `json:"trades"`
	AvgHoldingDays float64         `json:"avgHoldingPeriod"`
	EquityCurve    []EquityPoint   `json:"equityCurve"`
	Trades         []BacktestTrade `json:"tradeList"`
}

// EquityPoint represents a point on the equity curve.
type EquityPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}
