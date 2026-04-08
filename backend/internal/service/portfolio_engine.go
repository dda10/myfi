package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"myfi-backend/internal/model"
)

// HoldingStore abstracts asset holding persistence (replaces the deleted AssetRegistry).
type HoldingStore interface {
	GetAssetsByUser(ctx context.Context, userID string) ([]model.Asset, error)
	AddAsset(ctx context.Context, asset model.Asset) (int64, error)
	UpdateAsset(ctx context.Context, asset model.Asset) error
	DeleteAsset(ctx context.Context, assetID int64, userID string) error
}

// dbHoldingStore implements HoldingStore using *sql.DB directly.
type dbHoldingStore struct {
	db *sql.DB
}

func newDBHoldingStore(db *sql.DB) *dbHoldingStore {
	return &dbHoldingStore{db: db}
}

func (s *dbHoldingStore) GetAssetsByUser(ctx context.Context, userID string) ([]model.Asset, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, average_cost, acquisition_date, COALESCE(account,''), created_at, updated_at
		 FROM assets WHERE user_id = $1 ORDER BY symbol`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []model.Asset
	for rows.Next() {
		var a model.Asset
		var at string
		if err := rows.Scan(&a.ID, &a.UserID, &at, &a.Symbol, &a.Quantity,
			&a.AverageCost, &a.AcquisitionDate, &a.Account, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		a.AssetType = model.AssetType(at)
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (s *dbHoldingStore) AddAsset(ctx context.Context, asset model.Asset) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO assets (user_id, asset_type, symbol, quantity, average_cost, acquisition_date, account)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		asset.UserID, string(asset.AssetType), asset.Symbol, asset.Quantity,
		asset.AverageCost, asset.AcquisitionDate, asset.Account,
	).Scan(&id)
	return id, err
}

func (s *dbHoldingStore) UpdateAsset(ctx context.Context, asset model.Asset) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE assets SET quantity=$1, average_cost=$2, updated_at=NOW() WHERE id=$3 AND user_id=$4`,
		asset.Quantity, asset.AverageCost, asset.ID, asset.UserID)
	return err
}

func (s *dbHoldingStore) DeleteAsset(ctx context.Context, assetID int64, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM assets WHERE id=$1 AND user_id=$2`, assetID, userID)
	return err
}

// PortfolioEngine orchestrates asset holdings, transactions, and P&L calculations.
type PortfolioEngine struct {
	holdings HoldingStore
	ledger   *TransactionLedger
	prices   *PriceService
}

// NewPortfolioEngine creates a new PortfolioEngine instance.
func NewPortfolioEngine(holdings HoldingStore, ledger *TransactionLedger, prices *PriceService) *PortfolioEngine {
	return &PortfolioEngine{
		holdings: holdings,
		ledger:   ledger,
		prices:   prices,
	}
}

// NewPortfolioEngineFromDB creates a PortfolioEngine using a raw *sql.DB for holding storage.
func NewPortfolioEngineFromDB(db *sql.DB, ledger *TransactionLedger, prices *PriceService) *PortfolioEngine {
	return &PortfolioEngine{
		holdings: newDBHoldingStore(db),
		ledger:   ledger,
		prices:   prices,
	}
}

// ProcessBuy handles a buy transaction with double-entry accounting:
// debits cash (reduces cash holding) and credits the asset holding.
// Requirement 4.1: debit cash account, credit asset holding with quantity and cost basis.
func (e *PortfolioEngine) ProcessBuy(ctx context.Context, userID string, assetType model.AssetType, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (int64, error) {
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return 0, fmt.Errorf("unit price must be non-negative")
	}

	totalValue := quantity * unitPrice

	// Record the buy transaction in the ledger
	tx := model.Transaction{
		UserID:          userID,
		AssetType:       assetType,
		Symbol:          symbol,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		TotalValue:      totalValue,
		TransactionDate: txDate,
		TransactionType: model.Buy,
		Notes:           notes,
	}
	txID, err := e.ledger.RecordTransaction(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("failed to record buy transaction: %w", err)
	}

	// Update or create the asset holding
	assets, err := e.holdings.GetAssetsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user assets: %w", err)
	}

	var existing *model.Asset
	for i := range assets {
		if assets[i].Symbol == symbol && assets[i].AssetType == assetType {
			existing = &assets[i]
			break
		}
	}

	if existing != nil {
		// Weighted average cost: (oldQty*oldCost + newQty*newCost) / (oldQty + newQty)
		newQty := existing.Quantity + quantity
		newAvgCost := (existing.Quantity*existing.AverageCost + quantity*unitPrice) / newQty
		existing.Quantity = newQty
		existing.AverageCost = newAvgCost
		if err := e.holdings.UpdateAsset(ctx, *existing); err != nil {
			return 0, fmt.Errorf("failed to update asset holding: %w", err)
		}
	} else {
		// Create new holding
		_, err := e.holdings.AddAsset(ctx, model.Asset{
			UserID:          userID,
			AssetType:       assetType,
			Symbol:          symbol,
			Quantity:        quantity,
			AverageCost:     unitPrice,
			AcquisitionDate: txDate,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to create asset holding: %w", err)
		}
	}

	return txID, nil
}

// ProcessSell handles a sell transaction with weighted average cost P&L.
// Requirement 4.2: credit cash, debit asset, compute realized P&L using weighted average cost.
// Requirement 4.7: reject if sell quantity exceeds current holding.
func (e *PortfolioEngine) ProcessSell(ctx context.Context, userID string, assetType model.AssetType, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (*model.SellResult, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("unit price must be non-negative")
	}

	// Find the existing holding
	assets, err := e.holdings.GetAssetsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assets: %w", err)
	}

	var existing *model.Asset
	for i := range assets {
		if assets[i].Symbol == symbol && assets[i].AssetType == assetType {
			existing = &assets[i]
			break
		}
	}

	if existing == nil {
		return nil, fmt.Errorf("insufficient holdings: no holding found for %s (%s)", symbol, assetType)
	}

	if quantity > existing.Quantity {
		return nil, fmt.Errorf("insufficient holdings: requested sell %.4f but only hold %.4f of %s", quantity, existing.Quantity, symbol)
	}

	// Compute realized P&L using weighted average cost
	costBasis := existing.AverageCost * quantity
	saleProceeds := unitPrice * quantity
	realizedPL := saleProceeds - costBasis

	// Record the sell transaction
	tx := model.Transaction{
		UserID:          userID,
		AssetType:       assetType,
		Symbol:          symbol,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		TotalValue:      saleProceeds,
		TransactionDate: txDate,
		TransactionType: model.Sell,
		Notes:           notes,
	}
	txID, err := e.ledger.RecordTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to record sell transaction: %w", err)
	}

	// Update the holding (reduce quantity, average cost stays the same)
	remaining := existing.Quantity - quantity
	if remaining <= 0 {
		if err := e.holdings.DeleteAsset(ctx, existing.ID, userID); err != nil {
			return nil, fmt.Errorf("failed to delete depleted holding: %w", err)
		}
	} else {
		existing.Quantity = remaining
		if err := e.holdings.UpdateAsset(ctx, *existing); err != nil {
			return nil, fmt.Errorf("failed to update holding after sell: %w", err)
		}
	}

	return &model.SellResult{
		TransactionID: txID,
		RealizedPL:    realizedPL,
	}, nil
}

// ComputeUnrealizedPL calculates unrealized P&L for a single holding.
// Requirement 4.3: compare current market price against weighted average cost basis.
func (e *PortfolioEngine) ComputeUnrealizedPL(holding model.Asset, currentPrice float64) (unrealizedPL float64, unrealizedPLPct float64) {
	costBasis := holding.Quantity * holding.AverageCost
	marketValue := holding.Quantity * currentPrice
	unrealizedPL = marketValue - costBasis
	if costBasis > 0 {
		unrealizedPLPct = (unrealizedPL / costBasis) * 100
	}
	return unrealizedPL, unrealizedPLPct
}

// ComputeNAV calculates the total NAV for a user by summing market values of all holdings.
// Requirement 4.4: sum current market value of all holdings across all asset types.
// Uses PriceService when available, falls back to average cost.
func (e *PortfolioEngine) ComputeNAV(ctx context.Context, userID string) (float64, error) {
	assets, err := e.holdings.GetAssetsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user assets: %w", err)
	}

	var nav float64
	for _, a := range assets {
		price := a.AverageCost // fallback
		if e.prices != nil {
			quotes, err := e.prices.GetQuotes(ctx, []string{a.Symbol})
			if err == nil && len(quotes) > 0 && quotes[0].Price > 0 {
				price = quotes[0].Price
			}
		}
		nav += a.Quantity * price
	}
	return nav, nil
}

// ComputeAllocation computes allocation breakdown by asset type.
// Requirement 4.5: both absolute VND values and percentage of total NAV.
func (e *PortfolioEngine) ComputeAllocation(ctx context.Context, userID string) (byType map[model.AssetType]float64, byPercent map[model.AssetType]float64, totalNAV float64, err error) {
	assets, err := e.holdings.GetAssetsByUser(ctx, userID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get user assets: %w", err)
	}

	byType = make(map[model.AssetType]float64)
	byPercent = make(map[model.AssetType]float64)

	for _, a := range assets {
		price := a.AverageCost
		if e.prices != nil {
			quotes, qErr := e.prices.GetQuotes(ctx, []string{a.Symbol})
			if qErr == nil && len(quotes) > 0 && quotes[0].Price > 0 {
				price = quotes[0].Price
			}
		}
		value := a.Quantity * price
		byType[a.AssetType] += value
		totalNAV += value
	}

	if totalNAV > 0 {
		for at, val := range byType {
			byPercent[at] = (val / totalNAV) * 100
		}
	}

	return byType, byPercent, totalNAV, nil
}

// GetPortfolioSummary returns a full portfolio overview for a user.
func (e *PortfolioEngine) GetPortfolioSummary(ctx context.Context, userID string) (model.PortfolioSummary, error) {
	assets, err := e.holdings.GetAssetsByUser(ctx, userID)
	if err != nil {
		return model.PortfolioSummary{}, fmt.Errorf("failed to get user assets: %w", err)
	}

	summary := model.PortfolioSummary{}

	for _, a := range assets {
		price := a.AverageCost
		if e.prices != nil {
			quotes, qErr := e.prices.GetQuotes(ctx, []string{a.Symbol})
			if qErr == nil && len(quotes) > 0 && quotes[0].Price > 0 {
				price = quotes[0].Price
			}
		}

		marketValue := a.Quantity * price
		uPL, uPLPct := e.ComputeUnrealizedPL(a, price)

		summary.Holdings = append(summary.Holdings, model.HoldingDetail{
			Holding: model.Holding{
				ID:              a.ID,
				UserID:          a.UserID,
				Symbol:          a.Symbol,
				Quantity:        a.Quantity,
				AverageCost:     a.AverageCost,
				AcquisitionDate: a.AcquisitionDate,
				Account:         a.Account,
				CreatedAt:       a.CreatedAt,
				UpdatedAt:       a.UpdatedAt,
			},
			CurrentPrice:    price,
			MarketValue:     marketValue,
			UnrealizedPL:    uPL,
			UnrealizedPLPct: uPLPct,
		})

		summary.NAV += marketValue
	}

	return summary, nil
}
