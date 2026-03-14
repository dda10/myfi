package main

import (
	"context"
	"math"
	"testing"
	"time"
)

func newTestPortfolioEngine(t *testing.T) (*PortfolioEngine, *AssetRegistry, *TransactionLedger) {
	t.Helper()
	db := setupTestDB(t)
	registry := NewAssetRegistry(db, nil)
	ledger := NewTransactionLedger(db)
	engine := NewPortfolioEngine(registry, ledger, nil) // nil PriceService = use avg cost fallback
	return engine, registry, ledger
}

func TestBuyTransaction_CreatesHolding(t *testing.T) {
	engine, registry, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	txID, err := engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "first buy")
	if err != nil {
		t.Fatalf("ProcessBuy failed: %v", err)
	}
	if txID <= 0 {
		t.Fatalf("expected positive transaction ID, got %d", txID)
	}

	// Verify holding was created
	assets, err := registry.GetAssetsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetAssetsByUser failed: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	if assets[0].Symbol != "FPT" {
		t.Errorf("expected symbol FPT, got %s", assets[0].Symbol)
	}
	if assets[0].Quantity != 100 {
		t.Errorf("expected quantity 100, got %f", assets[0].Quantity)
	}
	if assets[0].AverageCost != 85000 {
		t.Errorf("expected avg cost 85000, got %f", assets[0].AverageCost)
	}
}

func TestBuyTransaction_DoubleEntry_WeightedAvgCost(t *testing.T) {
	engine, registry, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	// First buy: 100 shares at 85000
	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")
	// Second buy: 50 shares at 90000
	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 50, 90000, time.Now(), "")

	assets, _ := registry.GetAssetsByUser(ctx, 1)
	if len(assets) != 1 {
		t.Fatalf("expected 1 consolidated asset, got %d", len(assets))
	}

	expectedQty := 150.0
	// Weighted avg: (100*85000 + 50*90000) / 150 = (8500000 + 4500000) / 150 = 86666.666...
	expectedAvgCost := (100*85000.0 + 50*90000.0) / 150.0

	if assets[0].Quantity != expectedQty {
		t.Errorf("expected quantity %f, got %f", expectedQty, assets[0].Quantity)
	}
	if math.Abs(assets[0].AverageCost-expectedAvgCost) > 0.01 {
		t.Errorf("expected avg cost ~%f, got %f", expectedAvgCost, assets[0].AverageCost)
	}
}

func TestSellTransaction_RealizedPL(t *testing.T) {
	engine, registry, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	// Buy 100 shares at 85000
	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")

	// Sell 50 shares at 95000
	result, err := engine.ProcessSell(ctx, 1, VNStock, "FPT", 50, 95000, time.Now(), "take profit")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	// Realized P&L = (95000 - 85000) * 50 = 500000
	expectedPL := 500000.0
	if math.Abs(result.RealizedPL-expectedPL) > 0.01 {
		t.Errorf("expected realized P&L %f, got %f", expectedPL, result.RealizedPL)
	}

	// Verify remaining holding
	assets, _ := registry.GetAssetsByUser(ctx, 1)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset remaining, got %d", len(assets))
	}
	if assets[0].Quantity != 50 {
		t.Errorf("expected remaining quantity 50, got %f", assets[0].Quantity)
	}
	// Average cost should remain unchanged after sell
	if assets[0].AverageCost != 85000 {
		t.Errorf("expected avg cost 85000 unchanged, got %f", assets[0].AverageCost)
	}
}

func TestSellTransaction_FullLiquidation(t *testing.T) {
	engine, registry, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, VNStock, "SSI", 200, 30000, time.Now(), "")

	// Sell all 200 shares
	result, err := engine.ProcessSell(ctx, 1, VNStock, "SSI", 200, 35000, time.Now(), "")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	// P&L = (35000 - 30000) * 200 = 1000000
	if math.Abs(result.RealizedPL-1000000) > 0.01 {
		t.Errorf("expected realized P&L 1000000, got %f", result.RealizedPL)
	}

	// Holding should be deleted
	assets, _ := registry.GetAssetsByUser(ctx, 1)
	if len(assets) != 0 {
		t.Errorf("expected 0 assets after full liquidation, got %d", len(assets))
	}
}

func TestInsufficientHoldings_NoHolding(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	_, err := engine.ProcessSell(ctx, 1, VNStock, "FPT", 50, 85000, time.Now(), "")
	if err == nil {
		t.Fatal("expected error for selling without holding, got nil")
	}
	if !contains(err.Error(), "insufficient holdings") {
		t.Errorf("expected insufficient holdings error, got: %v", err)
	}
}

func TestInsufficientHoldings_ExceedsQuantity(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")

	_, err := engine.ProcessSell(ctx, 1, VNStock, "FPT", 150, 90000, time.Now(), "")
	if err == nil {
		t.Fatal("expected error for overselling, got nil")
	}
	if !contains(err.Error(), "insufficient holdings") {
		t.Errorf("expected insufficient holdings error, got: %v", err)
	}
}

func TestUnrealizedPL_Calculation(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)

	holding := Asset{
		Quantity:    100,
		AverageCost: 85000,
	}

	// Price went up
	uPL, uPLPct := engine.ComputeUnrealizedPL(holding, 95000)
	expectedPL := (95000 - 85000) * 100.0 // 1,000,000
	if math.Abs(uPL-expectedPL) > 0.01 {
		t.Errorf("expected unrealized P&L %f, got %f", expectedPL, uPL)
	}
	expectedPct := (expectedPL / (85000 * 100)) * 100 // ~11.76%
	if math.Abs(uPLPct-expectedPct) > 0.01 {
		t.Errorf("expected unrealized P&L pct %f, got %f", expectedPct, uPLPct)
	}

	// Price went down
	uPL2, uPLPct2 := engine.ComputeUnrealizedPL(holding, 80000)
	expectedPL2 := (80000 - 85000) * 100.0 // -500,000
	if math.Abs(uPL2-expectedPL2) > 0.01 {
		t.Errorf("expected unrealized P&L %f, got %f", expectedPL2, uPL2)
	}
	if uPLPct2 >= 0 {
		t.Errorf("expected negative unrealized P&L pct, got %f", uPLPct2)
	}
}

func TestNAV_Computation(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	// Buy multiple asset types
	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, Gold, "SJC", 2, 74000000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, Cash, "VCB", 1, 10000000, time.Now(), "")

	nav, err := engine.ComputeNAV(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeNAV failed: %v", err)
	}

	// Expected: 100*85000 + 2*74000000 + 1*10000000 = 8500000 + 148000000 + 10000000 = 166500000
	expected := 166500000.0
	if math.Abs(nav-expected) > 0.01 {
		t.Errorf("expected NAV %f, got %f", expected, nav)
	}
}

func TestNAV_EmptyPortfolio(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	nav, err := engine.ComputeNAV(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeNAV failed: %v", err)
	}
	if nav != 0 {
		t.Errorf("expected NAV 0 for empty portfolio, got %f", nav)
	}
}

func TestAllocation_ByAssetType(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, Gold, "SJC", 1, 74000000, time.Now(), "")

	byType, byPercent, totalNAV, err := engine.ComputeAllocation(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeAllocation failed: %v", err)
	}

	stockValue := 100 * 85000.0           // 8,500,000
	goldValue := 1 * 74000000.0           // 74,000,000
	expectedNAV := stockValue + goldValue // 82,500,000

	if math.Abs(totalNAV-expectedNAV) > 0.01 {
		t.Errorf("expected total NAV %f, got %f", expectedNAV, totalNAV)
	}

	if math.Abs(byType[VNStock]-stockValue) > 0.01 {
		t.Errorf("expected stock allocation %f, got %f", stockValue, byType[VNStock])
	}
	if math.Abs(byType[Gold]-goldValue) > 0.01 {
		t.Errorf("expected gold allocation %f, got %f", goldValue, byType[Gold])
	}

	// Percentages
	expectedStockPct := (stockValue / expectedNAV) * 100
	expectedGoldPct := (goldValue / expectedNAV) * 100
	if math.Abs(byPercent[VNStock]-expectedStockPct) > 0.01 {
		t.Errorf("expected stock pct %f, got %f", expectedStockPct, byPercent[VNStock])
	}
	if math.Abs(byPercent[Gold]-expectedGoldPct) > 0.01 {
		t.Errorf("expected gold pct %f, got %f", expectedGoldPct, byPercent[Gold])
	}

	// Percentages should sum to 100
	totalPct := byPercent[VNStock] + byPercent[Gold]
	if math.Abs(totalPct-100) > 0.01 {
		t.Errorf("expected percentages to sum to 100, got %f", totalPct)
	}
}

func TestPortfolioSummary(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, Crypto, "BTC", 0.01, 2500000000, time.Now(), "")

	summary, err := engine.GetPortfolioSummary(ctx, 1)
	if err != nil {
		t.Fatalf("GetPortfolioSummary failed: %v", err)
	}

	if len(summary.Holdings) != 2 {
		t.Fatalf("expected 2 holdings, got %d", len(summary.Holdings))
	}

	expectedNAV := 100*85000.0 + 0.01*2500000000.0 // 8500000 + 25000000 = 33500000
	if math.Abs(summary.NAV-expectedNAV) > 0.01 {
		t.Errorf("expected NAV %f, got %f", expectedNAV, summary.NAV)
	}

	if summary.AllocationByType[VNStock] == 0 {
		t.Error("expected non-zero stock allocation")
	}
	if summary.AllocationByType[Crypto] == 0 {
		t.Error("expected non-zero crypto allocation")
	}
	if summary.AllocationPercent[VNStock] == 0 {
		t.Error("expected non-zero stock allocation percent")
	}
}

func TestSellTransaction_WeightedAvgCostPL(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	// Buy 100 at 80000, then 100 at 90000 → avg cost = 85000
	engine.ProcessBuy(ctx, 1, VNStock, "VNM", 100, 80000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, VNStock, "VNM", 100, 90000, time.Now(), "")

	// Sell 50 at 100000
	result, err := engine.ProcessSell(ctx, 1, VNStock, "VNM", 50, 100000, time.Now(), "")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	// Weighted avg cost = (100*80000 + 100*90000) / 200 = 85000
	// Realized P&L = (100000 - 85000) * 50 = 750000
	expectedPL := 750000.0
	if math.Abs(result.RealizedPL-expectedPL) > 0.01 {
		t.Errorf("expected realized P&L %f, got %f", expectedPL, result.RealizedPL)
	}
}
