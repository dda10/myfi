package service

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"
)

func newTestPortfolioEngine(t *testing.T) (*PortfolioEngine, HoldingStore, *TransactionLedger) {
	t.Helper()
	db := testutil.SetupPostgresTestDB(t)
	store := newDBHoldingStore(db)
	ledger := NewTransactionLedger(db)
	engine := NewPortfolioEngine(store, ledger, nil) // nil PriceService = use avg cost fallback
	return engine, store, ledger
}

func TestBuyTransaction_CreatesHolding(t *testing.T) {
	engine, store, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	txID, err := engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "first buy")
	if err != nil {
		t.Fatalf("ProcessBuy failed: %v", err)
	}
	if txID <= 0 {
		t.Fatalf("expected positive transaction ID, got %d", txID)
	}

	// Verify holding was created
	assets, err := store.GetAssetsByUser(ctx, 1)
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
	engine, store, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	// First buy: 100 shares at 85000
	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")
	// Second buy: 50 shares at 90000
	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 50, 90000, time.Now(), "")

	assets, _ := store.GetAssetsByUser(ctx, 1)
	if len(assets) != 1 {
		t.Fatalf("expected 1 consolidated asset, got %d", len(assets))
	}

	expectedQty := 150.0
	expectedAvgCost := (100*85000.0 + 50*90000.0) / 150.0

	if assets[0].Quantity != expectedQty {
		t.Errorf("expected quantity %f, got %f", expectedQty, assets[0].Quantity)
	}
	if math.Abs(assets[0].AverageCost-expectedAvgCost) > 0.01 {
		t.Errorf("expected avg cost ~%f, got %f", expectedAvgCost, assets[0].AverageCost)
	}
}

func TestSellTransaction_RealizedPL(t *testing.T) {
	engine, store, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")

	result, err := engine.ProcessSell(ctx, 1, model.VNStock, "FPT", 50, 95000, time.Now(), "take profit")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	expectedPL := 500000.0
	if math.Abs(result.RealizedPL-expectedPL) > 0.01 {
		t.Errorf("expected realized P&L %f, got %f", expectedPL, result.RealizedPL)
	}

	assets, _ := store.GetAssetsByUser(ctx, 1)
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset remaining, got %d", len(assets))
	}
	if assets[0].Quantity != 50 {
		t.Errorf("expected remaining quantity 50, got %f", assets[0].Quantity)
	}
	if assets[0].AverageCost != 85000 {
		t.Errorf("expected avg cost 85000 unchanged, got %f", assets[0].AverageCost)
	}
}

func TestSellTransaction_FullLiquidation(t *testing.T) {
	engine, store, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, model.VNStock, "SSI", 200, 30000, time.Now(), "")

	result, err := engine.ProcessSell(ctx, 1, model.VNStock, "SSI", 200, 35000, time.Now(), "")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	if math.Abs(result.RealizedPL-1000000) > 0.01 {
		t.Errorf("expected realized P&L 1000000, got %f", result.RealizedPL)
	}

	assets, _ := store.GetAssetsByUser(ctx, 1)
	if len(assets) != 0 {
		t.Errorf("expected 0 assets after full liquidation, got %d", len(assets))
	}
}

func TestInsufficientHoldings_NoHolding(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	_, err := engine.ProcessSell(ctx, 1, model.VNStock, "FPT", 50, 85000, time.Now(), "")
	if err == nil {
		t.Fatal("expected error for selling without holding, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient holdings") {
		t.Errorf("expected insufficient holdings error, got: %v", err)
	}
}

func TestInsufficientHoldings_ExceedsQuantity(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")

	_, err := engine.ProcessSell(ctx, 1, model.VNStock, "FPT", 150, 90000, time.Now(), "")
	if err == nil {
		t.Fatal("expected error for overselling, got nil")
	}
	if !strings.Contains(err.Error(), "insufficient holdings") {
		t.Errorf("expected insufficient holdings error, got: %v", err)
	}
}

func TestUnrealizedPL_Calculation(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)

	holding := model.Asset{
		Quantity:    100,
		AverageCost: 85000,
	}

	uPL, uPLPct := engine.ComputeUnrealizedPL(holding, 95000)
	expectedPL := (95000 - 85000) * 100.0
	if math.Abs(uPL-expectedPL) > 0.01 {
		t.Errorf("expected unrealized P&L %f, got %f", expectedPL, uPL)
	}
	expectedPct := (expectedPL / (85000 * 100)) * 100
	if math.Abs(uPLPct-expectedPct) > 0.01 {
		t.Errorf("expected unrealized P&L pct %f, got %f", expectedPct, uPLPct)
	}

	uPL2, uPLPct2 := engine.ComputeUnrealizedPL(holding, 80000)
	expectedPL2 := (80000 - 85000) * 100.0
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

	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, model.Gold, "SJC", 2, 74000000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, model.Cash, "VCB", 1, 10000000, time.Now(), "")

	nav, err := engine.ComputeNAV(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeNAV failed: %v", err)
	}

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

	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, model.Gold, "SJC", 1, 74000000, time.Now(), "")

	byType, byPercent, totalNAV, err := engine.ComputeAllocation(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeAllocation failed: %v", err)
	}

	stockValue := 100 * 85000.0
	goldValue := 1 * 74000000.0
	expectedNAV := stockValue + goldValue

	if math.Abs(totalNAV-expectedNAV) > 0.01 {
		t.Errorf("expected total NAV %f, got %f", expectedNAV, totalNAV)
	}

	if math.Abs(byType[model.VNStock]-stockValue) > 0.01 {
		t.Errorf("expected stock allocation %f, got %f", stockValue, byType[model.VNStock])
	}
	if math.Abs(byType[model.Gold]-goldValue) > 0.01 {
		t.Errorf("expected gold allocation %f, got %f", goldValue, byType[model.Gold])
	}

	expectedStockPct := (stockValue / expectedNAV) * 100
	expectedGoldPct := (goldValue / expectedNAV) * 100
	if math.Abs(byPercent[model.VNStock]-expectedStockPct) > 0.01 {
		t.Errorf("expected stock pct %f, got %f", expectedStockPct, byPercent[model.VNStock])
	}
	if math.Abs(byPercent[model.Gold]-expectedGoldPct) > 0.01 {
		t.Errorf("expected gold pct %f, got %f", expectedGoldPct, byPercent[model.Gold])
	}

	totalPct := byPercent[model.VNStock] + byPercent[model.Gold]
	if math.Abs(totalPct-100) > 0.01 {
		t.Errorf("expected percentages to sum to 100, got %f", totalPct)
	}
}

func TestPortfolioSummary(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, model.VNStock, "FPT", 100, 85000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, model.Crypto, "BTC", 0.01, 2500000000, time.Now(), "")

	summary, err := engine.GetPortfolioSummary(ctx, 1)
	if err != nil {
		t.Fatalf("GetPortfolioSummary failed: %v", err)
	}

	if len(summary.Holdings) != 2 {
		t.Fatalf("expected 2 holdings, got %d", len(summary.Holdings))
	}

	expectedNAV := 100*85000.0 + 0.01*2500000000.0
	if math.Abs(summary.NAV-expectedNAV) > 0.01 {
		t.Errorf("expected NAV %f, got %f", expectedNAV, summary.NAV)
	}
}

func TestSellTransaction_WeightedAvgCostPL(t *testing.T) {
	engine, _, _ := newTestPortfolioEngine(t)
	ctx := context.Background()

	engine.ProcessBuy(ctx, 1, model.VNStock, "VNM", 100, 80000, time.Now(), "")
	engine.ProcessBuy(ctx, 1, model.VNStock, "VNM", 100, 90000, time.Now(), "")

	result, err := engine.ProcessSell(ctx, 1, model.VNStock, "VNM", 50, 100000, time.Now(), "")
	if err != nil {
		t.Fatalf("ProcessSell failed: %v", err)
	}

	expectedPL := 750000.0
	if math.Abs(result.RealizedPL-expectedPL) > 0.01 {
		t.Errorf("expected realized P&L %f, got %f", expectedPL, result.RealizedPL)
	}
}
