package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database with the required schema.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	// Create users table (assets has FK to users)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
	// Create assets table
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		asset_type TEXT NOT NULL CHECK (asset_type IN ('vn_stock', 'crypto', 'gold', 'savings', 'bond', 'cash')),
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL,
		average_cost REAL NOT NULL,
		acquisition_date DATETIME NOT NULL,
		account TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`); err != nil {
		t.Fatalf("failed to create assets table: %v", err)
	}
	// Create transactions table
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		asset_type TEXT NOT NULL,
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL,
		unit_price REAL NOT NULL,
		total_value REAL NOT NULL,
		transaction_date DATETIME NOT NULL,
		transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell', 'deposit', 'withdrawal', 'interest', 'dividend')),
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`); err != nil {
		t.Fatalf("failed to create transactions table: %v", err)
	}
	// Insert a test user
	if _, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ('testuser', 'hash')`); err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestValidateAssetType(t *testing.T) {
	// Valid types
	for _, at := range []AssetType{VNStock, Crypto, Gold, Savings, Bond, Cash} {
		if err := ValidateAssetType(at); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", at, err)
		}
	}
	// Invalid type
	err := ValidateAssetType(AssetType("forex"))
	if err == nil {
		t.Fatal("expected error for invalid asset type, got nil")
	}
	if !contains(err.Error(), "unrecognized asset type") {
		t.Errorf("error should mention unrecognized type, got: %v", err)
	}
	// Error should list supported types
	for _, at := range []AssetType{VNStock, Crypto, Gold, Savings, Bond, Cash} {
		if !contains(err.Error(), string(at)) {
			t.Errorf("error should list supported type %q, got: %v", at, err)
		}
	}
}

func TestAddAsset(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	asset := Asset{
		UserID:          1,
		AssetType:       VNStock,
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
	if got.AssetType != VNStock {
		t.Errorf("expected asset type %s, got %s", VNStock, got.AssetType)
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
		assetType AssetType
		symbol    string
	}{
		{VNStock, "VNM"},
		{Crypto, "BTC"},
		{Gold, "SJC"},
		{Savings, "TPBANK_12M"},
		{Bond, "GOV_5Y"},
		{Cash, "VCB_CURRENT"},
	}

	for _, tc := range types {
		t.Run(string(tc.assetType), func(t *testing.T) {
			asset := Asset{
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

	asset := Asset{
		UserID:          1,
		AssetType:       AssetType("forex"),
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
		asset       Asset
		errContains string
	}{
		{
			name:        "zero user ID",
			asset:       Asset{UserID: 0, AssetType: VNStock, Symbol: "FPT", Quantity: 1, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "invalid user ID",
		},
		{
			name:        "empty symbol",
			asset:       Asset{UserID: 1, AssetType: VNStock, Symbol: "", Quantity: 1, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "symbol is required",
		},
		{
			name:        "zero quantity",
			asset:       Asset{UserID: 1, AssetType: VNStock, Symbol: "FPT", Quantity: 0, AverageCost: 1000, AcquisitionDate: time.Now()},
			errContains: "quantity must be positive",
		},
		{
			name:        "negative cost",
			asset:       Asset{UserID: 1, AssetType: VNStock, Symbol: "FPT", Quantity: 1, AverageCost: -100, AcquisitionDate: time.Now()},
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

	// Add an asset first
	asset := Asset{
		UserID:          1,
		AssetType:       VNStock,
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

	// Update it
	asset.ID = id
	asset.Quantity = 100
	asset.AverageCost = 28000
	asset.Account = "SSI"

	err = registry.UpdateAsset(ctx, asset)
	if err != nil {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	// Verify update
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

	err := registry.UpdateAsset(ctx, Asset{
		ID: 999, UserID: 1, AssetType: VNStock, Symbol: "X",
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

	// Add an asset
	id, err := registry.AddAsset(ctx, Asset{
		UserID: 1, AssetType: VNStock, Symbol: "VNM",
		Quantity: 200, AverageCost: 75000,
		AcquisitionDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}

	// Insert associated transactions
	for i := 0; i < 3; i++ {
		_, err := testDB.ExecContext(ctx,
			`INSERT INTO transactions (user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			1, "vn_stock", "VNM", 50, 75000, 3750000, time.Now().Format(time.RFC3339), "buy",
		)
		if err != nil {
			t.Fatalf("failed to insert transaction: %v", err)
		}
	}

	// Verify transactions exist
	var txCount int
	testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE user_id = 1 AND symbol = 'VNM'`).Scan(&txCount)
	if txCount != 3 {
		t.Fatalf("expected 3 transactions, got %d", txCount)
	}

	// Delete the asset
	err = registry.DeleteAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("DeleteAsset failed: %v", err)
	}

	// Verify asset is gone
	_, err = registry.GetAsset(ctx, id, 1)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}

	// Verify transactions are cascade-deleted
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

	// Add multiple assets
	assets := []Asset{
		{UserID: 1, AssetType: VNStock, Symbol: "FPT", Quantity: 100, AverageCost: 85000, AcquisitionDate: time.Now()},
		{UserID: 1, AssetType: Gold, Symbol: "SJC", Quantity: 2, AverageCost: 74000000, AcquisitionDate: time.Now()},
		{UserID: 1, AssetType: Cash, Symbol: "VCB", Quantity: 1, AverageCost: 50000000, AcquisitionDate: time.Now()},
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

	// All monetary values should be stored in VND (Requirement 1.6)
	id, err := registry.AddAsset(ctx, Asset{
		UserID: 1, AssetType: Gold, Symbol: "SJC_1L",
		Quantity: 1, AverageCost: 74500000, // VND
		AcquisitionDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}

	got, err := registry.GetAsset(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	// Verify the cost is stored as-is in VND
	if got.AverageCost != 74500000 {
		t.Errorf("expected VND cost 74500000, got %f", got.AverageCost)
	}
}

func TestComputeNAV(t *testing.T) {
	testDB := setupTestDB(t)
	registry := NewAssetRegistry(testDB, nil)
	ctx := context.Background()

	// Add assets with known values
	registry.AddAsset(ctx, Asset{
		UserID: 1, AssetType: VNStock, Symbol: "FPT",
		Quantity: 100, AverageCost: 85000, AcquisitionDate: time.Now(),
	})
	registry.AddAsset(ctx, Asset{
		UserID: 1, AssetType: Cash, Symbol: "VCB",
		Quantity: 1, AverageCost: 10000000, AcquisitionDate: time.Now(),
	})

	nav, err := registry.computeNAV(ctx, 1)
	if err != nil {
		t.Fatalf("computeNAV failed: %v", err)
	}
	// Expected: 100*85000 + 1*10000000 = 8500000 + 10000000 = 18500000
	expected := 18500000.0
	if nav != expected {
		t.Errorf("expected NAV %f, got %f", expected, nav)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
