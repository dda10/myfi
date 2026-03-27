package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"
)

// setupTestDB delegates to the shared PostgreSQL test helper.
func setupTestDB(t testing.TB) *sql.DB {
	t.Helper()
	return testutil.SetupPostgresTestDB(t)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestValidateAssetType(t *testing.T) {
	// Valid types
	for _, at := range []model.AssetType{model.VNStock, model.Crypto, model.Gold, model.Savings, model.Bond, model.Cash} {
		if err := model.ValidateAssetType(at); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", at, err)
		}
	}
	// Invalid type
	err := model.ValidateAssetType(model.AssetType("forex"))
	if err == nil {
		t.Fatal("expected error for invalid asset type, got nil")
	}
	if !contains(err.Error(), "unrecognized asset type") {
		t.Errorf("error should mention unrecognized type, got: %v", err)
	}
	// Error should list supported types
	for _, at := range []model.AssetType{model.VNStock, model.Crypto, model.Gold, model.Savings, model.Bond, model.Cash} {
		if !contains(err.Error(), string(at)) {
			t.Errorf("error should list supported type %q, got: %v", at, err)
		}
	}
}

func TestAddAsset(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	asset := model.Asset{
		UserID:          1,
		AssetType:       model.VNStock,
		Symbol:          "FPT",
		Quantity:        100,
		AverageCost:     85000, // VND
		AcquisitionDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Account:         "SSI",
	}

	id, err := registry.AddAsset(ctx, asset)
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// Verify persistence
	got, err := registry.GetAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if got.Symbol != "FPT" {
		t.Errorf("expected symbol FPT, got %s", got.Symbol)
	}
	if got.AssetType != model.VNStock {
		t.Errorf("expected asset type %s, got %s", model.VNStock, got.AssetType)
	}
	if got.Quantity != 100 {
		t.Errorf("expected quantity 100, got %f", got.Quantity)
	}
	if got.AverageCost != 85000 {
		t.Errorf("expected average cost 85000, got %f", got.AverageCost)
	}
	if got.Account != "SSI" {
		t.Errorf("expected account SSI, got %s", got.Account)
	}
}

func TestAddAsset_AllTypes(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	types := []struct {
		assetType model.AssetType
		symbol    string
	}{
		{model.VNStock, "VNM"},
		{model.Crypto, "BTC"},
		{model.Gold, "SJC"},
		{model.Savings, "TPBANK_12M"},
		{model.Bond, "GOV_5Y"},
		{model.Cash, "VCB_CURRENT"},
	}

	for _, tc := range types {
		t.Run(string(tc.assetType), func(t *testing.T) {
			asset := model.Asset{
				UserID:          1,
				AssetType:       tc.assetType,
				Symbol:          tc.symbol,
				Quantity:        10,
				AverageCost:     1000000,
				AcquisitionDate: time.Now(),
			}
			id, err := registry.AddAsset(ctx, asset)
			if err != nil {
				t.Fatalf("AddAsset(%s) failed: %v", tc.assetType, err)
			}
			if id <= 0 {
				t.Fatalf("expected positive ID for %s", tc.assetType)
			}
		})
	}
}

func TestAddAsset_InvalidType(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	asset := model.Asset{
		UserID:          1,
		AssetType:       model.AssetType("forex"),
		Symbol:          "USDVND",
		Quantity:        1,
		AverageCost:     25000,
		AcquisitionDate: time.Now(),
	}

	_, err := registry.AddAsset(ctx, asset)
	if err == nil {
		t.Fatal("expected error for invalid asset type, got nil")
	}
	if !contains(err.Error(), "unrecognized asset type") {
		t.Errorf("expected unrecognized type error, got: %v", err)
	}
}

func TestAddAsset_ValidationErrors(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		asset       model.Asset
		errContains string
	}{
		{
			name:        "zero user ID",
			asset:       model.Asset{UserID: 0, AssetType: model.VNStock, Symbol: "FPT", Quantity: 1, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "invalid user ID",
		},
		{
			name:        "empty symbol",
			asset:       model.Asset{UserID: 1, AssetType: model.VNStock, Symbol: "", Quantity: 1, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "symbol is required",
		},
		{
			name:        "zero quantity",
			asset:       model.Asset{UserID: 1, AssetType: model.VNStock, Symbol: "FPT", Quantity: 0, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "quantity must be positive",
		},
		{
			name:        "negative cost",
			asset:       model.Asset{UserID: 1, AssetType: model.VNStock, Symbol: "FPT", Quantity: 1, AverageCost: -100, AcquisitionDate: time.Now()},
			errContains: "average cost must be non-negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := registry.AddAsset(ctx, tc.asset)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tc.errContains) {
				t.Errorf("expected error containing %q, got: %v", tc.errContains, err)
			}
		})
	}
}

func TestUpdateAsset(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	asset := model.Asset{
		UserID:          1,
		AssetType:       model.VNStock,
		Symbol:          "SSI",
		Quantity:        50,
		AverageCost:     30000,
		AcquisitionDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		Account:         "VPS",
	}
	id, err := registry.AddAsset(ctx, asset)
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}

	asset.ID = id
	asset.Quantity = 100
	asset.AverageCost = 28000
	asset.Account = "SSI"

	err = registry.UpdateAsset(ctx, asset)
	if err != nil {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	got, err := registry.GetAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if got.Quantity != 100 {
		t.Errorf("expected quantity 100, got %f", got.Quantity)
	}
	if got.AverageCost != 28000 {
		t.Errorf("expected average cost 28000, got %f", got.AverageCost)
	}
	if got.Account != "SSI" {
		t.Errorf("expected account SSI, got %s", got.Account)
	}
}

func TestUpdateAsset_NotFound(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	err := registry.UpdateAsset(ctx, model.Asset{
		ID: 999, UserID: 1, AssetType: model.VNStock, Symbol: "X",
		Quantity: 1, AverageCost: 1000, AcquisitionDate: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent asset, got nil")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestDeleteAsset_CascadeTransactions(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	id, err := registry.AddAsset(ctx, model.Asset{
		UserID: 1, AssetType: model.VNStock, Symbol: "VNM",
		Quantity: 200, AverageCost: 75000,
		AcquisitionDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}

	// Insert associated transactions using PostgreSQL placeholders and time.Time directly
	for i := 0; i < 3; i++ {
		_, err := testDB.ExecContext(ctx,
			`INSERT INTO transactions (user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			1, "vn_stock", "VNM", 50, 75000, 3750000, time.Now(), "buy",
		)
		if err != nil {
			t.Fatalf("failed to insert transaction: %v", err)
		}
	}

	var txCount int
	testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE user_id = 1 AND symbol = 'VNM'`).Scan(&txCount)
	if txCount != 3 {
		t.Fatalf("expected 3 transactions, got %d", txCount)
	}

	err = registry.DeleteAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("DeleteAsset failed: %v", err)
	}

	_, err = registry.GetAsset(ctx, id, 1)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}

	testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE user_id = 1 AND symbol = 'VNM'`).Scan(&txCount)
	if txCount != 0 {
		t.Errorf("expected 0 transactions after cascade delete, got %d", txCount)
	}
}

func TestDeleteAsset_NotFound(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	err := registry.DeleteAsset(ctx, 999, 1)
	if err == nil {
		t.Fatal("expected error for non-existent asset, got nil")
	}
}

func TestGetAssetsByUser(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	assets := []model.Asset{
		{UserID: 1, AssetType: model.VNStock, Symbol: "FPT", Quantity: 100, AverageCost: 85000, AcquisitionDate: time.Now()},
		{UserID: 1, AssetType: model.Gold, Symbol: "SJC", Quantity: 2, AverageCost: 74000000, AcquisitionDate: time.Now()},
		{UserID: 1, AssetType: model.Cash, Symbol: "VCB", Quantity: 1, AverageCost: 50000000, AcquisitionDate: time.Now()},
	}
	for _, a := range assets {
		if _, err := registry.AddAsset(ctx, a); err != nil {
			t.Fatalf("AddAsset failed: %v", err)
		}
	}

	got, err := registry.GetAssetsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetAssetsByUser failed: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(got))
	}
}

func TestVNDCurrencyInvariant(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	id, err := registry.AddAsset(ctx, model.Asset{
		UserID: 1, AssetType: model.Gold, Symbol: "SJC_1L",
		Quantity: 1, AverageCost: 74500000,
		AcquisitionDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}

	got, err := registry.GetAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if got.AverageCost != 74500000 {
		t.Errorf("expected VND cost 74500000, got %f", got.AverageCost)
	}
}

func TestComputeNAV(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	registry.AddAsset(ctx, model.Asset{
		UserID: 1, AssetType: model.VNStock, Symbol: "FPT",
		Quantity: 100, AverageCost: 85000, AcquisitionDate: time.Now(),
	})
	registry.AddAsset(ctx, model.Asset{
		UserID: 1, AssetType: model.Cash, Symbol: "VCB",
		Quantity: 1, AverageCost: 10000000, AcquisitionDate: time.Now(),
	})

	nav, err := registry.computeNAV(ctx, 1)
	if err != nil {
		t.Fatalf("computeNAV failed: %v", err)
	}
	expected := 18500000.0
	if nav != expected {
		t.Errorf("expected NAV %f, got %f", expected, nav)
	}
}

// contains and searchSubstring helpers are provided by sector_service.go in the same package
