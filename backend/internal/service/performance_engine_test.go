package service

import (
	"context"
	"math"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"
)

// newTestPerformanceEngine creates a PerformanceEngine backed by a real PostgreSQL
// test container with all migrations applied and a seeded test user (id=1).
func newTestPerformanceEngine(t *testing.T) *PerformanceEngine {
	t.Helper()
	db := testutil.SetupPostgresTestDB(t)
	return NewPerformanceEngine(db, nil) // nil router — benchmark tests are separate
}

// --- helpers ---

// insertNAVSnapshot inserts a NAV snapshot directly for test setup.
func insertNAVSnapshot(t *testing.T, engine *PerformanceEngine, userID int64, date time.Time, nav float64) {
	t.Helper()
	_, err := engine.db.Exec(
		`INSERT INTO nav_snapshots (user_id, nav, snapshot_date, created_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id, snapshot_date) DO UPDATE SET nav = EXCLUDED.nav`,
		userID, nav, date.Format("2006-01-02"),
	)
	if err != nil {
		t.Fatalf("insertNAVSnapshot failed: %v", err)
	}
}

// insertTransaction inserts a transaction directly for test setup.
func insertTransaction(t *testing.T, engine *PerformanceEngine, userID int64, assetType model.AssetType, symbol string, qty, unitPrice, totalValue float64, txDate time.Time, txType model.TransactionType) {
	t.Helper()
	_, err := engine.db.Exec(
		`INSERT INTO transactions (user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		userID, string(assetType), symbol, qty, unitPrice, totalValue, txDate, string(txType),
	)
	if err != nil {
		t.Fatalf("insertTransaction failed: %v", err)
	}
}

// --- StoreNAVSnapshot tests (Req 26.3) ---

func TestStoreNAVSnapshot_PersistsAndRetrieves(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	err := engine.StoreNAVSnapshot(ctx, 1, 100_000_000)
	if err != nil {
		t.Fatalf("StoreNAVSnapshot failed: %v", err)
	}

	// Retrieve via GetEquityCurve over a wide range
	now := time.Now()
	start := now.AddDate(0, 0, -1)
	end := now.AddDate(0, 0, 1)
	snapshots, err := engine.GetEquityCurve(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("GetEquityCurve failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if math.Abs(snapshots[0].NAV-100_000_000) > 0.01 {
		t.Errorf("expected NAV 100000000, got %f", snapshots[0].NAV)
	}
}

func TestStoreNAVSnapshot_UpsertSameDay(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	// Store twice on the same day — should upsert
	engine.StoreNAVSnapshot(ctx, 1, 100_000_000)
	engine.StoreNAVSnapshot(ctx, 1, 120_000_000)

	now := time.Now()
	snapshots, err := engine.GetEquityCurve(ctx, 1, now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("GetEquityCurve failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot after upsert, got %d", len(snapshots))
	}
	if math.Abs(snapshots[0].NAV-120_000_000) > 0.01 {
		t.Errorf("expected updated NAV 120000000, got %f", snapshots[0].NAV)
	}
}

// --- GetEquityCurve tests (Req 26.3) ---

func TestGetEquityCurve_OrderedByDate(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 1), 102_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 2), 101_000_000)

	snapshots, err := engine.GetEquityCurve(ctx, 1, base.AddDate(0, 0, -1), base.AddDate(0, 0, 5))
	if err != nil {
		t.Fatalf("GetEquityCurve failed: %v", err)
	}
	if len(snapshots) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snapshots))
	}
	// Verify ascending order
	for i := 1; i < len(snapshots); i++ {
		if !snapshots[i].Date.After(snapshots[i-1].Date) {
			t.Errorf("snapshots not in ascending order at index %d", i)
		}
	}
}

func TestGetEquityCurve_EmptyRange(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	far := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshots, err := engine.GetEquityCurve(ctx, 1, far, far.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("GetEquityCurve failed: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots for empty range, got %d", len(snapshots))
	}
}

// --- ComputeTWR tests (Req 26.1) ---

func TestComputeTWR_SimpleGrowth(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// NAV grows from 100M to 110M over 3 days with no cash flows
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 1), 105_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 2), 110_000_000)

	twr, err := engine.ComputeTWR(ctx, 1, base.AddDate(0, 0, -1), base.AddDate(0, 0, 5))
	if err != nil {
		t.Fatalf("ComputeTWR failed: %v", err)
	}

	// Expected: (105/100) * (110/105) - 1 = 1.05 * 1.04762 - 1 ≈ 0.10
	expected := 0.10
	if math.Abs(twr-expected) > 0.001 {
		t.Errorf("expected TWR ~%f, got %f", expected, twr)
	}
}

func TestComputeTWR_WithCashFlow(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Day 0: NAV = 100M
	// Day 1: deposit 10M, NAV = 115M (so growth = (115 - 100 - 10) / 100 = 0.05)
	// Day 2: NAV = 120M (growth = (120 - 115) / 115 ≈ 0.04348)
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 1), 115_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 2), 120_000_000)

	// Insert a deposit on day 1
	insertTransaction(t, engine, 1, model.Cash, "VND", 1, 10_000_000, 10_000_000,
		base.AddDate(0, 0, 1), model.Deposit)

	twr, err := engine.ComputeTWR(ctx, 1, base.AddDate(0, 0, -1), base.AddDate(0, 0, 5))
	if err != nil {
		t.Fatalf("ComputeTWR failed: %v", err)
	}

	// Expected: (1 + 0.05) * (1 + 120/115 - 1) - 1 = 1.05 * 1.04348 - 1 ≈ 0.09565
	expected := (1.05)*(120_000_000.0/115_000_000.0) - 1
	if math.Abs(twr-expected) > 0.001 {
		t.Errorf("expected TWR ~%f, got %f", expected, twr)
	}
}

func TestComputeTWR_InsufficientData(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Only 1 snapshot — not enough for TWR
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)

	twr, err := engine.ComputeTWR(ctx, 1, base.AddDate(0, 0, -1), base.AddDate(0, 0, 5))
	if err != nil {
		t.Fatalf("ComputeTWR failed: %v", err)
	}
	if twr != 0 {
		t.Errorf("expected TWR 0 for insufficient data, got %f", twr)
	}
}

// --- ComputeMWRR tests (Req 26.2) ---

func TestComputeMWRR_SimpleGrowth(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := base.AddDate(1, 0, 0) // 1 year later

	// Start NAV = 100M, End NAV = 110M, no intermediate cash flows
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, end, 110_000_000)

	mwrr, err := engine.ComputeMWRR(ctx, 1, base, end)
	if err != nil {
		t.Fatalf("ComputeMWRR failed: %v", err)
	}

	// Simple case: 10% growth over 1 year → annualized ≈ 10%
	// The Newton-Raphson should converge to ~10% annualized
	if mwrr < 0.05 || mwrr > 0.15 {
		t.Errorf("expected MWRR around 0.10, got %f", mwrr)
	}
}

func TestComputeMWRR_NoData(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mwrr, err := engine.ComputeMWRR(ctx, 1, base, base.AddDate(1, 0, 0))
	if err != nil {
		t.Fatalf("ComputeMWRR failed: %v", err)
	}
	if mwrr != 0 {
		t.Errorf("expected MWRR 0 for no data, got %f", mwrr)
	}
}

// --- ComputePerformanceByAssetType tests (Req 26.6) ---

func TestComputePerformanceByAssetType_MultipleTypes(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := base.AddDate(0, 6, 0)

	// Stock: buy 100 at 85000 = 8.5M, sell 100 at 95000 = 9.5M → return = (9.5-8.5)/8.5
	insertTransaction(t, engine, 1, model.VNStock, "FPT", 100, 85000, 8_500_000, base, model.Buy)
	insertTransaction(t, engine, 1, model.VNStock, "FPT", 100, 95000, 9_500_000, base.AddDate(0, 3, 0), model.Sell)

	// Gold: buy 1 at 74M, sell 1 at 78M → return = (78-74)/74
	insertTransaction(t, engine, 1, model.Gold, "SJC", 1, 74_000_000, 74_000_000, base, model.Buy)
	insertTransaction(t, engine, 1, model.Gold, "SJC", 1, 78_000_000, 78_000_000, base.AddDate(0, 4, 0), model.Sell)

	byType, err := engine.ComputePerformanceByAssetType(ctx, 1, base.AddDate(0, 0, -1), end)
	if err != nil {
		t.Fatalf("ComputePerformanceByAssetType failed: %v", err)
	}

	// Stock return: (9.5M - 8.5M) / 8.5M ≈ 0.1176
	stockReturn := byType[model.VNStock]
	expectedStock := (9_500_000.0 - 8_500_000.0) / 8_500_000.0
	if math.Abs(stockReturn-expectedStock) > 0.01 {
		t.Errorf("expected stock return ~%f, got %f", expectedStock, stockReturn)
	}

	// Gold return: (78M - 74M) / 74M ≈ 0.054
	goldReturn := byType[model.Gold]
	expectedGold := (78_000_000.0 - 74_000_000.0) / 74_000_000.0
	if math.Abs(goldReturn-expectedGold) > 0.01 {
		t.Errorf("expected gold return ~%f, got %f", expectedGold, goldReturn)
	}
}

func TestComputePerformanceByAssetType_WithDividends(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := base.AddDate(1, 0, 0)

	// Buy stock, receive dividend
	insertTransaction(t, engine, 1, model.VNStock, "VNM", 200, 100_000, 20_000_000, base, model.Buy)
	insertTransaction(t, engine, 1, model.VNStock, "VNM", 0, 0, 1_000_000, base.AddDate(0, 6, 0), model.Dividend)

	byType, err := engine.ComputePerformanceByAssetType(ctx, 1, base.AddDate(0, 0, -1), end)
	if err != nil {
		t.Fatalf("ComputePerformanceByAssetType failed: %v", err)
	}

	// Return = (dividends - buys) / buys = (1M - 20M) / 20M = -0.95
	// But since no sell, totalReturned = 0 + 1M = 1M, totalInvested = 20M
	stockReturn := byType[model.VNStock]
	expected := (1_000_000.0 - 20_000_000.0) / 20_000_000.0
	if math.Abs(stockReturn-expected) > 0.01 {
		t.Errorf("expected stock return ~%f, got %f", expected, stockReturn)
	}
}

// --- GetPerformanceMetrics integration test (Req 26.1-26.6) ---

func TestGetPerformanceMetrics_ReturnsAllFields(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := base.AddDate(0, 0, 5)

	// Set up NAV snapshots
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 1), 102_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 2), 105_000_000)

	// Set up a transaction
	insertTransaction(t, engine, 1, model.VNStock, "FPT", 100, 85000, 8_500_000, base, model.Buy)

	metrics, err := engine.GetPerformanceMetrics(ctx, 1, base.AddDate(0, 0, -1), end)
	if err != nil {
		t.Fatalf("GetPerformanceMetrics failed: %v", err)
	}

	// TWR should be computed (non-zero since we have 3 snapshots)
	if metrics.TWR == 0 {
		t.Error("expected non-zero TWR")
	}

	// Equity curve should have 3 points
	if len(metrics.EquityCurve) != 3 {
		t.Errorf("expected 3 equity curve points, got %d", len(metrics.EquityCurve))
	}

	// PerformanceByType should have at least the stock entry
	if metrics.PerformanceByType == nil {
		t.Error("expected non-nil PerformanceByType map")
	}
}

// --- getNAVAtDate tests ---

func TestGetNAVAtDate_ReturnsClosest(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	insertNAVSnapshot(t, engine, 1, base, 100_000_000)
	insertNAVSnapshot(t, engine, 1, base.AddDate(0, 0, 3), 110_000_000)

	// Query for day 2 — should return day 0's NAV (closest on or before)
	nav, err := engine.getNAVAtDate(ctx, 1, base.AddDate(0, 0, 2))
	if err != nil {
		t.Fatalf("getNAVAtDate failed: %v", err)
	}
	if math.Abs(nav-100_000_000) > 0.01 {
		t.Errorf("expected NAV 100000000, got %f", nav)
	}

	// Query for day 3 — should return day 3's NAV
	nav, err = engine.getNAVAtDate(ctx, 1, base.AddDate(0, 0, 3))
	if err != nil {
		t.Fatalf("getNAVAtDate failed: %v", err)
	}
	if math.Abs(nav-110_000_000) > 0.01 {
		t.Errorf("expected NAV 110000000, got %f", nav)
	}
}

func TestGetNAVAtDate_NoData(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	nav, err := engine.getNAVAtDate(ctx, 1, time.Now())
	if err != nil {
		t.Fatalf("getNAVAtDate failed: %v", err)
	}
	if nav != 0 {
		t.Errorf("expected NAV 0 for no data, got %f", nav)
	}
}

// --- getCashFlowEvents tests ---

func TestGetCashFlowEvents_ClassifiesCorrectly(t *testing.T) {
	engine := newTestPerformanceEngine(t)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := base.AddDate(0, 6, 0)

	// Deposit = positive inflow
	insertTransaction(t, engine, 1, model.Cash, "VND", 1, 10_000_000, 10_000_000, base.AddDate(0, 1, 0), model.Deposit)
	// Buy = positive inflow (money entering portfolio)
	insertTransaction(t, engine, 1, model.VNStock, "FPT", 100, 85000, 8_500_000, base.AddDate(0, 2, 0), model.Buy)
	// Sell = negative outflow
	insertTransaction(t, engine, 1, model.VNStock, "FPT", 50, 90000, 4_500_000, base.AddDate(0, 3, 0), model.Sell)
	// Withdrawal = negative outflow
	insertTransaction(t, engine, 1, model.Cash, "VND", 1, 5_000_000, 5_000_000, base.AddDate(0, 4, 0), model.Withdrawal)

	events, err := engine.getCashFlowEvents(ctx, 1, base, end)
	if err != nil {
		t.Fatalf("getCashFlowEvents failed: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("expected 4 cash flow events, got %d", len(events))
	}

	// Deposit → positive
	if events[0].Amount <= 0 {
		t.Errorf("expected positive amount for deposit, got %f", events[0].Amount)
	}
	// Buy → positive
	if events[1].Amount <= 0 {
		t.Errorf("expected positive amount for buy, got %f", events[1].Amount)
	}
	// Sell → negative
	if events[2].Amount >= 0 {
		t.Errorf("expected negative amount for sell, got %f", events[2].Amount)
	}
	// Withdrawal → negative
	if events[3].Amount >= 0 {
		t.Errorf("expected negative amount for withdrawal, got %f", events[3].Amount)
	}
}
