package portfolio

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"myfi-backend/internal/domain/market"
)

// PortfolioEngine orchestrates stock holdings, transactions, and P&L calculations.
// Stock-only: no multi-asset types, no AssetType parameter.
type PortfolioEngine struct {
	db     *sql.DB
	ledger *TransactionLedger
	prices *market.PriceService
}

// NewPortfolioEngine creates a new PortfolioEngine instance.
func NewPortfolioEngine(db *sql.DB, ledger *TransactionLedger, prices *market.PriceService) *PortfolioEngine {
	return &PortfolioEngine{
		db:     db,
		ledger: ledger,
		prices: prices,
	}
}

// RecordBuy handles a stock buy transaction: updates or creates a holding
// with weighted-average cost basis, and records the transaction in the ledger.
func (e *PortfolioEngine) RecordBuy(ctx context.Context, userID string, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (int64, error) {
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return 0, fmt.Errorf("unit price must be non-negative")
	}

	totalValue := quantity * unitPrice

	// Record the buy transaction in the ledger.
	tx := Transaction{
		UserID:          userID,
		Symbol:          symbol,
		TransactionType: TxBuy,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		TotalValue:      totalValue,
		TransactionDate: txDate,
		Notes:           notes,
	}
	txID, err := e.ledger.RecordTransaction(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("failed to record buy transaction: %w", err)
	}

	// Upsert the holding with weighted-average cost.
	holding, err := e.getHolding(ctx, userID, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get holding: %w", err)
	}

	if holding != nil {
		newQty := holding.Quantity + quantity
		newAvgCost := (holding.Quantity*holding.AverageCost + quantity*unitPrice) / newQty
		_, err = e.db.ExecContext(ctx,
			`UPDATE holdings SET quantity=$1, avg_cost=$2, updated_at=NOW() WHERE id=$3 AND user_id=$4`,
			newQty, newAvgCost, holding.ID, userID)
		if err != nil {
			return 0, fmt.Errorf("failed to update holding: %w", err)
		}
	} else {
		_, err = e.db.ExecContext(ctx,
			`INSERT INTO holdings (user_id, symbol, quantity, avg_cost, total_dividends, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 0, NOW(), NOW())`,
			userID, symbol, quantity, unitPrice)
		if err != nil {
			return 0, fmt.Errorf("failed to create holding: %w", err)
		}
	}

	return txID, nil
}

// RecordSell handles a stock sell transaction: validates sufficient holdings,
// computes realized P&L using weighted-average cost, and updates the holding.
func (e *PortfolioEngine) RecordSell(ctx context.Context, userID string, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (*SellResult, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("unit price must be non-negative")
	}

	holding, err := e.getHolding(ctx, userID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get holding: %w", err)
	}
	if holding == nil {
		return nil, fmt.Errorf("insufficient holdings: no holding found for %s", symbol)
	}
	if quantity > holding.Quantity {
		return nil, fmt.Errorf("insufficient holdings: requested sell %.4f but only hold %.4f of %s", quantity, holding.Quantity, symbol)
	}

	costBasis := holding.AverageCost * quantity
	saleProceeds := unitPrice * quantity
	realizedPL := saleProceeds - costBasis

	tx := Transaction{
		UserID:          userID,
		Symbol:          symbol,
		TransactionType: TxSell,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		TotalValue:      saleProceeds,
		RealizedPnL:     realizedPL,
		TransactionDate: txDate,
		Notes:           notes,
	}
	txID, err := e.ledger.RecordTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to record sell transaction: %w", err)
	}

	remaining := holding.Quantity - quantity
	if remaining <= 0 {
		_, err = e.db.ExecContext(ctx, `DELETE FROM holdings WHERE id=$1 AND user_id=$2`, holding.ID, userID)
	} else {
		_, err = e.db.ExecContext(ctx,
			`UPDATE holdings SET quantity=$1, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
			remaining, holding.ID, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update holding after sell: %w", err)
	}

	return &SellResult{TransactionID: txID, RealizedPL: realizedPL}, nil
}

// GetHoldings returns all stock holdings for a user with current valuation.
func (e *PortfolioEngine) GetHoldings(ctx context.Context, userID string) ([]HoldingDetail, error) {
	rows, err := e.db.QueryContext(ctx,
		`SELECT id, user_id, symbol, quantity, avg_cost, total_dividends, created_at, updated_at
		 FROM holdings WHERE user_id = $1 ORDER BY symbol`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query holdings: %w", err)
	}
	defer rows.Close()

	var details []HoldingDetail
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.ID, &h.UserID, &h.Symbol, &h.Quantity,
			&h.AverageCost, &h.TotalDividends, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}

		price := h.AverageCost // fallback
		if e.prices != nil {
			quotes, qErr := e.prices.GetQuotes(ctx, []string{h.Symbol})
			if qErr == nil && len(quotes) > 0 && quotes[0].Price > 0 {
				price = quotes[0].Price
			}
		}

		marketValue := h.Quantity * price
		costBasis := h.Quantity * h.AverageCost
		unrealizedPL := marketValue - costBasis
		var unrealizedPLPct float64
		if costBasis > 0 {
			unrealizedPLPct = (unrealizedPL / costBasis) * 100
		}

		details = append(details, HoldingDetail{
			Holding:         h,
			CurrentPrice:    price,
			MarketValue:     marketValue,
			UnrealizedPL:    unrealizedPL,
			UnrealizedPLPct: unrealizedPLPct,
		})
	}
	return details, rows.Err()
}

// ComputeNAV calculates the total NAV for a user by summing market values of all holdings.
func (e *PortfolioEngine) ComputeNAV(ctx context.Context, userID string) (NAVResult, error) {
	details, err := e.GetHoldings(ctx, userID)
	if err != nil {
		return NAVResult{}, fmt.Errorf("failed to get holdings: %w", err)
	}

	var totalNAV float64
	for _, d := range details {
		totalNAV += d.MarketValue
	}

	return NAVResult{
		TotalNAV:   totalNAV,
		Holdings:   details,
		ComputedAt: time.Now(),
	}, nil
}

// GetSectorAllocation computes portfolio allocation by ICB sector.
func (e *PortfolioEngine) GetSectorAllocation(ctx context.Context, userID string) ([]SectorAllocation, error) {
	details, err := e.GetHoldings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	var totalValue float64
	sectorMap := make(map[string]*SectorAllocation)

	for _, d := range details {
		totalValue += d.MarketValue

		// Look up sector from DB (populated by sector service).
		sector := e.lookupSector(ctx, d.Holding.Symbol)
		if sector == "" {
			sector = "UNKNOWN"
		}

		alloc, ok := sectorMap[sector]
		if !ok {
			alloc = &SectorAllocation{
				Sector:     sector,
				SectorName: market.SectorNameMap[market.ICBSector(sector)],
			}
			sectorMap[sector] = alloc
		}
		alloc.Value += d.MarketValue
		alloc.StockCount++
	}

	var allocations []SectorAllocation
	for _, alloc := range sectorMap {
		if totalValue > 0 {
			alloc.Weight = (alloc.Value / totalValue) * 100
		}
		allocations = append(allocations, *alloc)
	}

	return allocations, nil
}

// getHolding retrieves a single holding for a user and symbol.
func (e *PortfolioEngine) getHolding(ctx context.Context, userID string, symbol string) (*Holding, error) {
	var h Holding
	err := e.db.QueryRowContext(ctx,
		`SELECT id, user_id, symbol, quantity, avg_cost, total_dividends, created_at, updated_at
		 FROM holdings WHERE user_id = $1 AND symbol = $2`,
		userID, symbol,
	).Scan(&h.ID, &h.UserID, &h.Symbol, &h.Quantity, &h.AverageCost, &h.TotalDividends, &h.CreatedAt, &h.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// lookupSector returns the ICB sector code for a symbol from the stock-sector mapping cache.
func (e *PortfolioEngine) lookupSector(ctx context.Context, symbol string) string {
	var sector string
	_ = e.db.QueryRowContext(ctx,
		`SELECT sector FROM stock_sector_mapping WHERE symbol = $1`, symbol,
	).Scan(&sector)
	return sector
}
