package portfolio

import (
	"fmt"
	"time"
)

// TransactionType represents the type of a stock transaction.
type TransactionType string

const (
	TxBuy      TransactionType = "buy"
	TxSell     TransactionType = "sell"
	TxDividend TransactionType = "dividend"
	TxSplit    TransactionType = "split"
	TxBonus    TransactionType = "bonus"
)

// ValidTransactionTypes contains all supported transaction types.
var ValidTransactionTypes = map[TransactionType]bool{
	TxBuy:      true,
	TxSell:     true,
	TxDividend: true,
	TxSplit:    true,
	TxBonus:    true,
}

// ValidateTransactionType checks if the given transaction type is supported.
func ValidateTransactionType(tt TransactionType) error {
	if ValidTransactionTypes[tt] {
		return nil
	}
	return fmt.Errorf("invalid transaction type %q", tt)
}

// Transaction represents a single stock transaction in the ledger.
type Transaction struct {
	ID              int64           `json:"id"`
	UserID          string          `json:"userId"`
	Symbol          string          `json:"symbol"`
	TransactionType TransactionType `json:"transactionType"`
	Quantity        float64         `json:"quantity"`
	UnitPrice       float64         `json:"unitPrice"`
	TotalValue      float64         `json:"totalValue"`
	RealizedPnL     float64         `json:"realizedPnl,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	TransactionDate time.Time       `json:"transactionDate"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// SellResult contains the outcome of a sell transaction.
type SellResult struct {
	TransactionID int64   `json:"transactionId"`
	RealizedPL    float64 `json:"realizedPL"`
}

// CashFlowEvent represents an external cash flow for MWRR/XIRR calculation.
type CashFlowEvent struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}

// BenchmarkComparison contains benchmark comparison results.
type BenchmarkComparison struct {
	VNIndexReturn   float64 `json:"vnIndexReturn"`
	VN30Return      float64 `json:"vn30Return"`
	PortfolioReturn float64 `json:"portfolioReturn"`
	Alpha           float64 `json:"alpha"`
}

// RiskMetrics contains portfolio-level risk analytics.
type RiskMetrics struct {
	SharpeRatio float64 `json:"sharpeRatio"`
	MaxDrawdown float64 `json:"maxDrawdown"`
	Beta        float64 `json:"beta"`
	Volatility  float64 `json:"volatility"`
	VaR95       float64 `json:"var95"`
}
