package service

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"

	"pgregory.net/rapid"
)

// --- Generators ---

func genAssetType() *rapid.Generator[model.AssetType] {
	return rapid.SampledFrom([]model.AssetType{model.VNStock, model.Crypto, model.Gold, model.Savings, model.Bond, model.Cash})
}

func genSymbol() *rapid.Generator[string] {
	return rapid.StringMatching(`[A-Z]{3,4}`)
}

func genQuantity() *rapid.Generator[float64] {
	return rapid.Float64Range(0.001, 100000)
}

func genPrice() *rapid.Generator[float64] {
	return rapid.Float64Range(1, 100_000_000)
}

func newPropEngine(db *sql.DB) (*PortfolioEngine, *AssetRegistry) {
	// Clean data tables so each rapid iteration starts fresh.
	// Order matters due to foreign key constraints.
	for _, tbl := range []string{"transactions", "assets", "savings_accounts"} {
		db.Exec("DELETE FROM " + tbl)
	}
	reg := NewAssetRegistry(db, nil)
	ledger := NewTransactionLedger(db)
	engine := NewPortfolioEngine(reg, ledger, nil)
	return engine, reg
}

// --- Property 14: Buy Transaction Double-Entry ---
// For any sequence of buys on the same symbol, the holding quantity equals the
// sum of buy quantities, and average cost equals the weighted average.

func TestProperty14_BuyTransactionDoubleEntry(t *testing.T) {
	db := testutil.SetupPostgresTestDB(t)
	rapid.Check(t, func(t *rapid.T) {
		engine, registry := newPropEngine(db)
		ctx := context.Background()

		assetType := genAssetType().Draw(t, "assetType")
		symbol := genSymbol().Draw(t, "symbol")
		numBuys := rapid.IntRange(1, 10).Draw(t, "numBuys")

		var totalQty, totalCost float64
		for i := 0; i < numBuys; i++ {
			qty := genQuantity().Draw(t, "qty")
			price := genPrice().Draw(t, "price")
			_, err := engine.ProcessBuy(ctx, 1, assetType, symbol, qty, price, time.Now(), "")
			if err != nil {
				t.Fatalf("ProcessBuy failed: %v", err)
			}
			totalQty += qty
			totalCost += qty * price
		}

		assets, err := registry.GetAssetsByUser(ctx, 1)
		if err != nil {
			t.Fatalf("GetAssetsByUser: %v", err)
		}
		var found *model.Asset
		for i := range assets {
			if assets[i].Symbol == symbol && assets[i].AssetType == assetType {
				found = &assets[i]
				break
			}
		}
		if found == nil {
			t.Fatal("holding not found after buys")
		}

		// Property: quantity == sum of all buy quantities
		if math.Abs(found.Quantity-totalQty) > 1e-6 {
			t.Errorf("quantity: got %f, want %f", found.Quantity, totalQty)
		}
		// Property: average cost == weighted average
		expectedAvg := totalCost / totalQty
		if math.Abs(found.AverageCost-expectedAvg)/expectedAvg > 1e-9 {
			t.Errorf("avg cost: got %f, want %f", found.AverageCost, expectedAvg)
		}
	})
}

// --- Property 15: Sell Transaction P&L Computation ---
// Realized P&L = (sellPrice - weightedAvgCost) * sellQty.
// Remaining holding preserves the original average cost.

func TestProperty15_SellTransactionPLComputation(t *testing.T) {
	db := testutil.SetupPostgresTestDB(t)
	rapid.Check(t, func(t *rapid.T) {
		engine, registry := newPropEngine(db)
		ctx := context.Background()

		assetType := genAssetType().Draw(t, "assetType")
		symbol := genSymbol().Draw(t, "symbol")

		numBuys := rapid.IntRange(1, 5).Draw(t, "numBuys")
		var totalQty, totalCost float64
		for i := 0; i < numBuys; i++ {
			qty := genQuantity().Draw(t, "buyQty")
			price := genPrice().Draw(t, "buyPrice")
			_, err := engine.ProcessBuy(ctx, 1, assetType, symbol, qty, price, time.Now(), "")
			if err != nil {
				t.Fatalf("ProcessBuy failed: %v", err)
			}
			totalQty += qty
			totalCost += qty * price
		}
		weightedAvgCost := totalCost / totalQty

		sellFraction := rapid.Float64Range(0.01, 0.99).Draw(t, "sellFraction")
		sellQty := totalQty * sellFraction
		sellPrice := genPrice().Draw(t, "sellPrice")

		result, err := engine.ProcessSell(ctx, 1, assetType, symbol, sellQty, sellPrice, time.Now(), "")
		if err != nil {
			t.Fatalf("ProcessSell failed: %v", err)
		}

		// Property: realized P&L = (sellPrice - avgCost) * sellQty
		expectedPL := (sellPrice - weightedAvgCost) * sellQty
		relErr := math.Abs(result.RealizedPL-expectedPL) / math.Max(1, math.Abs(expectedPL))
		if relErr > 1e-6 {
			t.Errorf("P&L: got %f, want %f", result.RealizedPL, expectedPL)
		}

		// Property: avg cost unchanged after sell
		assets, _ := registry.GetAssetsByUser(ctx, 1)
		var found *model.Asset
		for i := range assets {
			if assets[i].Symbol == symbol && assets[i].AssetType == assetType {
				found = &assets[i]
			}
		}
		if found == nil {
			t.Fatal("holding should exist after partial sell")
		}
		if math.Abs(found.AverageCost-weightedAvgCost)/weightedAvgCost > 1e-9 {
			t.Errorf("avg cost changed: got %f, want %f", found.AverageCost, weightedAvgCost)
		}
		remainingQty := totalQty - sellQty
		if math.Abs(found.Quantity-remainingQty)/remainingQty > 1e-6 {
			t.Errorf("remaining qty: got %f, want %f", found.Quantity, remainingQty)
		}
	})
}

// --- Property 16: Unrealized P&L Calculation ---
// unrealized P&L = (currentPrice - avgCost) * qty; sign matches price direction.

func TestProperty16_UnrealizedPLCalculation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		qty := genQuantity().Draw(t, "qty")
		avgCost := genPrice().Draw(t, "avgCost")
		currentPrice := genPrice().Draw(t, "currentPrice")

		engine := &PortfolioEngine{} // no DB needed for pure computation
		holding := model.Asset{Quantity: qty, AverageCost: avgCost}
		uPL, uPLPct := engine.ComputeUnrealizedPL(holding, currentPrice)

		expectedPL := (currentPrice - avgCost) * qty
		if math.Abs(uPL-expectedPL)/math.Max(1, math.Abs(expectedPL)) > 1e-9 {
			t.Errorf("P&L: got %f, want %f", uPL, expectedPL)
		}

		costBasis := qty * avgCost
		expectedPct := (expectedPL / costBasis) * 100
		if math.Abs(uPLPct-expectedPct)/math.Max(1, math.Abs(expectedPct)) > 1e-9 {
			t.Errorf("P&L pct: got %f, want %f", uPLPct, expectedPct)
		}

		if currentPrice > avgCost && uPL <= 0 {
			t.Errorf("positive P&L expected: price %f > cost %f, got %f", currentPrice, avgCost, uPL)
		}
		if currentPrice < avgCost && uPL >= 0 {
			t.Errorf("negative P&L expected: price %f < cost %f, got %f", currentPrice, avgCost, uPL)
		}
	})
}

// --- Property 17: NAV Aggregation ---
// NAV = sum(qty * price) for all holdings. Non-negative when all inputs positive.

func TestProperty17_NAVAggregation(t *testing.T) {
	db := testutil.SetupPostgresTestDB(t)
	rapid.Check(t, func(t *rapid.T) {
		engine, _ := newPropEngine(db)
		ctx := context.Background()

		numHoldings := rapid.IntRange(1, 6).Draw(t, "numHoldings")
		type hs struct {
			at  model.AssetType
			sym string
			qty float64
			px  float64
		}
		seen := make(map[string]bool)
		var holdings []hs
		for len(holdings) < numHoldings {
			at := genAssetType().Draw(t, "at")
			sym := genSymbol().Draw(t, "sym")
			key := string(at) + ":" + sym
			if seen[key] {
				continue
			}
			seen[key] = true
			holdings = append(holdings, hs{at, sym, genQuantity().Draw(t, "q"), genPrice().Draw(t, "p")})
		}

		var expectedNAV float64
		for _, h := range holdings {
			_, err := engine.ProcessBuy(ctx, 1, h.at, h.sym, h.qty, h.px, time.Now(), "")
			if err != nil {
				t.Fatalf("ProcessBuy: %v", err)
			}
			expectedNAV += h.qty * h.px
		}

		nav, err := engine.ComputeNAV(ctx, 1)
		if err != nil {
			t.Fatalf("ComputeNAV: %v", err)
		}
		if math.Abs(nav-expectedNAV)/math.Max(1, expectedNAV) > 1e-6 {
			t.Errorf("NAV: got %f, want %f", nav, expectedNAV)
		}
		if nav < 0 {
			t.Errorf("NAV must be non-negative, got %f", nav)
		}
	})
}

// --- Property 18: Insufficient Holdings Rejection ---
// Selling non-existent or over-quantity holdings always returns "insufficient holdings".

func TestProperty18_InsufficientHoldingsRejection(t *testing.T) {
	db := testutil.SetupPostgresTestDB(t)
	rapid.Check(t, func(t *rapid.T) {
		engine, _ := newPropEngine(db)
		ctx := context.Background()

		assetType := genAssetType().Draw(t, "assetType")
		symbol := genSymbol().Draw(t, "symbol")
		sellPrice := genPrice().Draw(t, "sellPrice")

		// Case 1: no holding exists
		_, err := engine.ProcessSell(ctx, 1, assetType, symbol, 1, sellPrice, time.Now(), "")
		if err == nil {
			t.Fatal("expected error selling non-existent holding")
		}
		if !contains(err.Error(), "insufficient holdings") {
			t.Errorf("want 'insufficient holdings', got: %v", err)
		}

		// Case 2: buy some, sell more than held
		buyQty := genQuantity().Draw(t, "buyQty")
		buyPrice := genPrice().Draw(t, "buyPrice")
		_, err = engine.ProcessBuy(ctx, 1, assetType, symbol, buyQty, buyPrice, time.Now(), "")
		if err != nil {
			t.Fatalf("ProcessBuy: %v", err)
		}

		excess := rapid.Float64Range(1.001, 10.0).Draw(t, "excess")
		oversellQty := buyQty * excess

		_, err = engine.ProcessSell(ctx, 1, assetType, symbol, oversellQty, sellPrice, time.Now(), "")
		if err == nil {
			t.Fatalf("expected error selling %f, hold %f", oversellQty, buyQty)
		}
		if !contains(err.Error(), "insufficient holdings") {
			t.Errorf("want 'insufficient holdings', got: %v", err)
		}
	})
}
