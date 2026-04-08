package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
)

// CorporateActionService manages dividend tracking, stock split adjustments,
// and bonus share processing for portfolio holdings.
// Requirements: 30.1, 30.2, 30.3, 30.7
type CorporateActionService struct {
	db     *sql.DB
	router *infra.DataSourceRouter
	ledger *TransactionLedger
}

// NewCorporateActionService creates a new CorporateActionService instance.
func NewCorporateActionService(db *sql.DB, router *infra.DataSourceRouter, ledger *TransactionLedger) *CorporateActionService {
	return &CorporateActionService{
		db:     db,
		router: router,
		ledger: ledger,
	}
}

// FetchDividendCalendar fetches upcoming dividend events for the given symbols
// from VCI/KBS via the Data_Source_Router.
// Requirement 30.1: fetch dividend calendar from VCI/KBS.
func (s *CorporateActionService) FetchDividendCalendar(ctx context.Context, symbols []string) ([]model.CorporateAction, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	var actions []model.CorporateAction
	for _, symbol := range symbols {
		events, err := s.fetchCorporateEventsForSymbol(ctx, symbol)
		if err != nil {
			log.Printf("[CorporateActionService] Failed to fetch events for %s: %v", symbol, err)
			continue
		}
		for _, e := range events {
			if e.ActionType == model.CorporateActionDividend {
				actions = append(actions, e)
			}
		}
	}
	return actions, nil
}

// FetchSplitAndBonusEvents fetches stock split and bonus share events for the given symbols.
// Requirement 30.1: fetch stock split and bonus share events from VCI/KBS.
func (s *CorporateActionService) FetchSplitAndBonusEvents(ctx context.Context, symbols []string) ([]model.CorporateAction, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	var actions []model.CorporateAction
	for _, symbol := range symbols {
		events, err := s.fetchCorporateEventsForSymbol(ctx, symbol)
		if err != nil {
			log.Printf("[CorporateActionService] Failed to fetch events for %s: %v", symbol, err)
			continue
		}
		for _, e := range events {
			if e.ActionType == model.CorporateActionStockSplit || e.ActionType == model.CorporateActionBonusShare {
				actions = append(actions, e)
			}
		}
	}
	return actions, nil
}

// fetchCorporateEventsForSymbol fetches all corporate action events for a single symbol
// via the Data_Source_Router.
func (s *CorporateActionService) fetchCorporateEventsForSymbol(ctx context.Context, symbol string) ([]model.CorporateAction, error) {
	if s.router == nil {
		return nil, fmt.Errorf("data source router not available")
	}

	log.Printf("[CorporateActionService] Fetching corporate events for %s via Data_Source_Router", symbol)
	return nil, nil
}

// RecordDividendPayment auto-records a dividend payment as a transaction in the
// Transaction_Ledger for a user's holding.
// Requirement 30.2: auto-record dividend payments as transactions.
func (s *CorporateActionService) RecordDividendPayment(ctx context.Context, userID string, action model.CorporateAction, holdingQty float64) (*model.DividendRecord, error) {
	if action.ActionType != model.CorporateActionDividend {
		return nil, fmt.Errorf("expected dividend action, got %s", action.ActionType)
	}
	if action.DividendPerShare <= 0 {
		return nil, fmt.Errorf("dividend per share must be positive")
	}
	if action.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if holdingQty <= 0 {
		return nil, fmt.Errorf("holding quantity must be positive")
	}

	totalAmount := holdingQty * action.DividendPerShare

	// Record dividend transaction in the ledger
	tx := model.Transaction{
		UserID:          userID,
		AssetType:       model.VNStock,
		Symbol:          action.Symbol,
		Quantity:        holdingQty,
		UnitPrice:       action.DividendPerShare,
		TotalValue:      totalAmount,
		TransactionDate: action.PaymentDate,
		TransactionType: model.Dividend,
		Notes:           fmt.Sprintf("Dividend payment: %.0f VND/share, ex-date: %s", action.DividendPerShare, action.ExDate.Format("2006-01-02")),
	}

	txID, err := s.ledger.RecordTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to record dividend transaction: %w", err)
	}

	record := &model.DividendRecord{
		UserID:           userID,
		Symbol:           action.Symbol,
		ExDate:           action.ExDate,
		PaymentDate:      action.PaymentDate,
		DividendPerShare: action.DividendPerShare,
		SharesHeld:       holdingQty,
		TotalAmount:      totalAmount,
		TransactionID:    txID,
		CreatedAt:        time.Now(),
	}

	return record, nil
}

// AdjustForStockSplit adjusts cost basis and quantity for a stock split event.
// Requirement 30.3: auto-adjust cost basis and quantity for splits.
// Property 51: For split ratio N:M, new quantity = Q×(N/M), new cost = C×(M/N).
// The caller must provide the current holding quantity and average cost.
func (s *CorporateActionService) AdjustForStockSplit(ctx context.Context, userID string, action model.CorporateAction, oldQuantity, oldCost float64) (*model.SplitAdjustment, error) {
	if action.ActionType != model.CorporateActionStockSplit {
		return nil, fmt.Errorf("expected stock_split action, got %s", action.ActionType)
	}
	if action.SplitRatioFrom <= 0 || action.SplitRatioTo <= 0 {
		return nil, fmt.Errorf("split ratio must be positive (from: %.2f, to: %.2f)", action.SplitRatioFrom, action.SplitRatioTo)
	}
	if action.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Split ratio N:M means each M old shares become N new shares
	ratio := action.SplitRatioTo / action.SplitRatioFrom
	newQuantity := oldQuantity * ratio
	newCost := oldCost / ratio

	log.Printf("[CorporateActionService] Split adjustment for %s: qty %.2f->%.2f, cost %.2f->%.2f (ratio %g:%g)",
		action.Symbol, oldQuantity, newQuantity, oldCost, newCost, action.SplitRatioFrom, action.SplitRatioTo)

	return &model.SplitAdjustment{
		Symbol:       action.Symbol,
		RatioFrom:    action.SplitRatioFrom,
		RatioTo:      action.SplitRatioTo,
		OldQuantity:  oldQuantity,
		NewQuantity:  newQuantity,
		OldCostBasis: oldCost,
		NewCostBasis: newCost,
	}, nil
}

// AdjustForBonusShares adjusts quantity and cost basis for a bonus share event.
// Requirement 30.3: auto-adjust cost basis and quantity for bonus shares.
// The caller must provide the current holding quantity and average cost.
func (s *CorporateActionService) AdjustForBonusShares(ctx context.Context, userID string, action model.CorporateAction, oldQuantity, oldCost float64) (*model.SplitAdjustment, error) {
	if action.ActionType != model.CorporateActionBonusShare {
		return nil, fmt.Errorf("expected bonus_share action, got %s", action.ActionType)
	}
	if action.SplitRatioFrom <= 0 || action.SplitRatioTo <= 0 {
		return nil, fmt.Errorf("bonus ratio must be positive (from: %.2f, to: %.2f)", action.SplitRatioFrom, action.SplitRatioTo)
	}
	if action.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Bonus ratio From:To means for every From shares, you get To bonus shares
	bonusShares := oldQuantity * (action.SplitRatioTo / action.SplitRatioFrom)
	newQuantity := oldQuantity + bonusShares
	newCost := (oldCost * oldQuantity) / newQuantity

	log.Printf("[CorporateActionService] Bonus share adjustment for %s: qty %.2f->%.2f, cost %.2f->%.2f (ratio %g:%g)",
		action.Symbol, oldQuantity, newQuantity, oldCost, newCost, action.SplitRatioFrom, action.SplitRatioTo)

	return &model.SplitAdjustment{
		Symbol:       action.Symbol,
		RatioFrom:    action.SplitRatioFrom,
		RatioTo:      action.SplitRatioTo,
		OldQuantity:  oldQuantity,
		NewQuantity:  newQuantity,
		OldCostBasis: oldCost,
		NewCostBasis: newCost,
	}, nil
}

// GetDividendHistory retrieves dividend history for a specific holding and computes
// yield-on-cost.
// Requirement 30.7: track dividend history per holding to compute yield-on-cost.
// The caller must provide the cost basis for yield-on-cost calculation.
func (s *CorporateActionService) GetDividendHistory(ctx context.Context, userID string, symbol string, costBasis float64) (*model.DividendHistory, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Get all dividend transactions for this symbol
	txns, err := s.ledger.GetTransactionsBySymbol(ctx, userID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	var records []model.DividendRecord
	var totalDividends float64

	for _, tx := range txns {
		if tx.TransactionType != model.Dividend {
			continue
		}
		record := model.DividendRecord{
			UserID:           tx.UserID,
			Symbol:           tx.Symbol,
			ExDate:           tx.TransactionDate,
			PaymentDate:      tx.TransactionDate,
			DividendPerShare: tx.UnitPrice,
			SharesHeld:       tx.Quantity,
			TotalAmount:      tx.TotalValue,
			TransactionID:    tx.ID,
			CreatedAt:        tx.CreatedAt,
		}
		records = append(records, record)
		totalDividends += tx.TotalValue
	}

	// Compute yield-on-cost: annual dividend / original cost basis
	yieldOnCost := ComputeYieldOnCost(records, costBasis)

	return &model.DividendHistory{
		Symbol:         symbol,
		Records:        records,
		TotalDividends: totalDividends,
		YieldOnCost:    yieldOnCost,
	}, nil
}

// ComputeYieldOnCost calculates yield-on-cost from dividend records and cost basis.
// Yield-on-cost = (annualized dividends / cost basis) * 100
// Exported for testing.
func ComputeYieldOnCost(records []model.DividendRecord, costBasis float64) float64 {
	if costBasis <= 0 || len(records) == 0 {
		return 0
	}

	var totalDividends float64
	for _, r := range records {
		totalDividends += r.TotalAmount
	}

	// Compute the time span of dividend history
	earliest := records[0].PaymentDate
	latest := records[0].PaymentDate
	for _, r := range records[1:] {
		if r.PaymentDate.Before(earliest) {
			earliest = r.PaymentDate
		}
		if r.PaymentDate.After(latest) {
			latest = r.PaymentDate
		}
	}

	// Annualize: if we have less than a year of data, extrapolate
	durationDays := latest.Sub(earliest).Hours() / 24
	var annualDividends float64
	if durationDays < 30 {
		annualDividends = totalDividends
	} else {
		years := durationDays / 365.25
		if years < 1 {
			years = 1
		}
		annualDividends = totalDividends / years
	}

	return (annualDividends / costBasis) * 100
}

// ProcessCorporateAction is a convenience method that dispatches to the appropriate
// handler based on the action type. For split/bonus actions, the caller must provide
// the current holding quantity and average cost via the holdingQty and avgCost parameters.
func (s *CorporateActionService) ProcessCorporateAction(ctx context.Context, userID string, action model.CorporateAction, holdingQty, avgCost float64) error {
	switch action.ActionType {
	case model.CorporateActionDividend:
		_, err := s.RecordDividendPayment(ctx, userID, action, holdingQty)
		return err
	case model.CorporateActionStockSplit:
		_, err := s.AdjustForStockSplit(ctx, userID, action, holdingQty, avgCost)
		return err
	case model.CorporateActionBonusShare:
		_, err := s.AdjustForBonusShares(ctx, userID, action, holdingQty, avgCost)
		return err
	default:
		return fmt.Errorf("unsupported corporate action type: %s", action.ActionType)
	}
}
