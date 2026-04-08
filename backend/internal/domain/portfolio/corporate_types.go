package portfolio

import "time"

// CorporateActionType represents the type of corporate action event.
type CorporateActionType string

const (
	CorporateActionDividend   CorporateActionType = "dividend"
	CorporateActionStockSplit CorporateActionType = "stock_split"
	CorporateActionBonusShare CorporateActionType = "bonus_share"
)

// CorporateAction represents a corporate action event (dividend, split, bonus).
type CorporateAction struct {
	ID               int64               `json:"id"`
	Symbol           string              `json:"symbol"`
	ActionType       CorporateActionType `json:"actionType"`
	ExDate           time.Time           `json:"exDate"`
	RecordDate       time.Time           `json:"recordDate"`
	PaymentDate      time.Time           `json:"paymentDate"`
	DividendPerShare float64             `json:"dividendPerShare,omitempty"`
	SplitRatioFrom   float64             `json:"splitRatioFrom,omitempty"`
	SplitRatioTo     float64             `json:"splitRatioTo,omitempty"`
	Description      string              `json:"description,omitempty"`
}

// DividendRecord tracks a dividend payment for a specific user holding.
type DividendRecord struct {
	ID               int64     `json:"id"`
	UserID           string    `json:"userId"`
	Symbol           string    `json:"symbol"`
	ExDate           time.Time `json:"exDate"`
	PaymentDate      time.Time `json:"paymentDate"`
	DividendPerShare float64   `json:"dividendPerShare"`
	SharesHeld       float64   `json:"sharesHeld"`
	TotalAmount      float64   `json:"totalAmount"`
	TransactionID    int64     `json:"transactionId"`
	CreatedAt        time.Time `json:"createdAt"`
}

// SplitAdjustment records a cost basis / quantity adjustment from a split or bonus.
type SplitAdjustment struct {
	Symbol       string  `json:"symbol"`
	RatioFrom    float64 `json:"ratioFrom"`
	RatioTo      float64 `json:"ratioTo"`
	OldQuantity  float64 `json:"oldQuantity"`
	NewQuantity  float64 `json:"newQuantity"`
	OldCostBasis float64 `json:"oldCostBasis"`
	NewCostBasis float64 `json:"newCostBasis"`
}
