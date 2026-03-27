package service

import (
	"context"
	"fmt"
	"time"

	"myfi-backend/internal/model"
)

// PortfolioEngine orchestrates asset holdings, transactions, and P&L calculations.
type PortfolioEngine struct {
	registry *AssetRegistry
	ledger   *TransactionLedger
	prices   *PriceService
}

// NewPortfolioEngine creates a new PortfolioEngine instance.
func NewPortfolioEngine(registry *AssetRegistry, ledger *TransactionLedger, prices *PriceService) *PortfolioEngine {
	return &PortfolioEngine{
		registry: registry,
		ledger:   ledger,
		prices:   prices,
	}
}

// ProcessBuy handles a buy transaction with double-entry accounting:
// debits cash (reduces cash holding) and credits the asset holding.
// Requirement 4.1: debit cash account, credit asset holding with quantity and cost basis.
func (e *PortfolioEngine) ProcessBuy(ctx context.Context, userID int64, assetType model.AssetType, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (int64, error) {
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
	assets, err := e.registry.GetAssetsByUser(ctx, userID)
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
		if err := e.registry.UpdateAsset(ctx, *existing); err != nil {
			return 0, fmt.Errorf("failed to update asset holding: %w", err)
		}
	} else {
		// Create new holding
		_, err := e.registry.AddAsset(ctx, model.Asset{
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
func (e *PortfolioEngine) ProcessSell(ctx context.Context, userID int64, assetType model.AssetType, symbol string, quantity, unitPrice float64, txDate time.Time, notes string) (*model.SellResult, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("unit price must be non-negative")
	}

	// Find the existing holding
	assets, err := e.registry.GetAssetsByUser(ctx, userID)
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
		// Delete the holding entirely
		if err := e.registry.DeleteAsset(ctx, existing.ID, userID); err != nil {
			return nil, fmt.Errorf("failed to delete depleted holding: %w", err)
		}
	} else {
		existing.Quantity = remaining
		if err := e.registry.UpdateAsset(ctx, *existing); err != nil {
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
func (e *PortfolioEngine) ComputeNAV(ctx context.Context, userID int64) (float64, error) {
	assets, err := e.registry.GetAssetsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user assets: %w", err)
	}

	var nav float64
	for _, a := range assets {
		price := a.AverageCost // fallback
		if e.prices != nil {
			quotes, err := e.prices.GetQuotes(ctx, []string{a.Symbol}, a.AssetType)
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
func (e *PortfolioEngine) ComputeAllocation(ctx context.Context, userID int64) (byType map[model.AssetType]float64, byPercent map[model.AssetType]float64, totalNAV float64, err error) {
	assets, err := e.registry.GetAssetsByUser(ctx, userID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get user assets: %w", err)
	}

	byType = make(map[model.AssetType]float64)
	byPercent = make(map[model.AssetType]float64)

	for _, a := range assets {
		price := a.AverageCost
		if e.prices != nil {
			quotes, qErr := e.prices.GetQuotes(ctx, []string{a.Symbol}, a.AssetType)
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
func (e *PortfolioEngine) GetPortfolioSummary(ctx context.Context, userID int64) (model.PortfolioSummary, error) {
	assets, err := e.registry.GetAssetsByUser(ctx, userID)
	if err != nil {
		return model.PortfolioSummary{}, fmt.Errorf("failed to get user assets: %w", err)
	}

	summary := model.PortfolioSummary{
		AllocationByType:  make(map[model.AssetType]float64),
		AllocationPercent: make(map[model.AssetType]float64),
	}

	for _, a := range assets {
		price := a.AverageCost
		if e.prices != nil {
			quotes, qErr := e.prices.GetQuotes(ctx, []string{a.Symbol}, a.AssetType)
			if qErr == nil && len(quotes) > 0 && quotes[0].Price > 0 {
				price = quotes[0].Price
			}
		}

		marketValue := a.Quantity * price
		uPL, uPLPct := e.ComputeUnrealizedPL(a, price)

		summary.Holdings = append(summary.Holdings, model.HoldingDetail{
			Asset:           a,
			CurrentPrice:    price,
			MarketValue:     marketValue,
			UnrealizedPL:    uPL,
			UnrealizedPLPct: uPLPct,
		})

		summary.NAV += marketValue
		summary.AllocationByType[a.AssetType] += marketValue
	}

	if summary.NAV > 0 {
		for at, val := range summary.AllocationByType {
			summary.AllocationPercent[at] = (val / summary.NAV) * 100
		}
	}

	return summary, nil
}
