package service

import (
	"context"
	"testing"
	"time"

	"myfi-backend/internal/model"
)

func TestRecordTransaction_Buy(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	tx := model.Transaction{
		UserID:          1,
		AssetType:       model.VNStock,
		Symbol:          "FPT",
		Quantity:        100,
		UnitPrice:       85000,
		TransactionDate: time.Now(),
		TransactionType: model.Buy,
		Notes:           "Initial purchase",
	}

	id, err := ledger.RecordTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("RecordTransaction failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// Verify persistence
	txns, err := ledger.GetTransactionsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetTransactionsByUser failed: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if txns[0].Symbol != "FPT" {
		t.Errorf("expected symbol FPT, got %s", txns[0].Symbol)
	}
	if txns[0].TransactionType != model.Buy {
		t.Errorf("expected type buy, got %s", txns[0].TransactionType)
	}
	if txns[0].TotalValue != 8500000 {
		t.Errorf("expected total value 8500000, got %f", txns[0].TotalValue)
	}
}

func TestRecordTransaction_AllTypes(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	types := []model.TransactionType{model.Buy, model.Sell, model.Deposit, model.Withdrawal, model.Interest, model.Dividend}
	for _, tt := range types {
		t.Run(string(tt), func(t *testing.T) {
			tx := model.Transaction{
				UserID:          1,
				AssetType:       model.VNStock,
				Symbol:          "SSI",
				Quantity:        10,
				UnitPrice:       30000,
				TransactionDate: time.Now(),
				TransactionType: tt,
			}
			id, err := ledger.RecordTransaction(ctx, tx)
			if err != nil {
				t.Fatalf("RecordTransaction(%s) failed: %v", tt, err)
			}
			if id <= 0 {
				t.Fatalf("expected positive ID for %s", tt)
			}
		})
	}
}

func TestTransactionTypeValidation(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	tx := model.Transaction{
		UserID:          1,
		AssetType:       model.VNStock,
		Symbol:          "FPT",
		Quantity:        10,
		UnitPrice:       85000,
		TransactionDate: time.Now(),
		TransactionType: model.TransactionType("swap"),
	}

	_, err := ledger.RecordTransaction(ctx, tx)
	if err == nil {
		t.Fatal("expected error for invalid transaction type, got nil")
	}
	if !contains(err.Error(), "invalid transaction type") {
		t.Errorf("expected invalid transaction type error, got: %v", err)
	}
}

func TestRecordTransaction_ValidationErrors(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		tx          model.Transaction
		errContains string
	}{
		{
			name: "zero user ID",
			tx: model.Transaction{
				UserID: 0, AssetType: model.VNStock, Symbol: "FPT",
				Quantity: 10, UnitPrice: 85000, TransactionDate: time.Now(), TransactionType: model.Buy,
			},
			errContains: "invalid user ID",
		},
		{
			name: "empty symbol",
			tx: model.Transaction{
				UserID: 1, AssetType: model.VNStock, Symbol: "",
				Quantity: 10, UnitPrice: 85000, TransactionDate: time.Now(), TransactionType: model.Buy,
			},
			errContains: "symbol is required",
		},
		{
			name: "zero quantity",
			tx: model.Transaction{
				UserID: 1, AssetType: model.VNStock, Symbol: "FPT",
				Quantity: 0, UnitPrice: 85000, TransactionDate: time.Now(), TransactionType: model.Buy,
			},
			errContains: "quantity must be positive",
		},
		{
			name: "negative unit price",
			tx: model.Transaction{
				UserID: 1, AssetType: model.VNStock, Symbol: "FPT",
				Quantity: 10, UnitPrice: -100, TransactionDate: time.Now(), TransactionType: model.Buy,
			},
			errContains: "unit price must be non-negative",
		},
		{
			name: "invalid asset type",
			tx: model.Transaction{
				UserID: 1, AssetType: model.AssetType("forex"), Symbol: "USDVND",
				Quantity: 10, UnitPrice: 25000, TransactionDate: time.Now(), TransactionType: model.Buy,
			},
			errContains: "unrecognized asset type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ledger.RecordTransaction(ctx, tc.tx)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tc.errContains) {
				t.Errorf("expected error containing %q, got: %v", tc.errContains, err)
			}
		})
	}
}

func TestRecordTransaction_AutoComputeTotalValue(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	tx := model.Transaction{
		UserID:          1,
		AssetType:       model.Gold,
		Symbol:          "SJC",
		Quantity:        2,
		UnitPrice:       74000000,
		TotalValue:      0, // should be auto-computed
		TransactionDate: time.Now(),
		TransactionType: model.Buy,
	}

	id, err := ledger.RecordTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("RecordTransaction failed: %v", err)
	}

	txns, err := ledger.GetTransactionsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetTransactionsByUser failed: %v", err)
	}

	for _, txn := range txns {
		if txn.ID == id {
			expected := 2.0 * 74000000.0
			if txn.TotalValue != expected {
				t.Errorf("expected auto-computed total value %f, got %f", expected, txn.TotalValue)
			}
			return
		}
	}
	t.Fatal("recorded transaction not found")
}

func TestGetTransactionsBySymbol(t *testing.T) {
	db := setupTestDB(t)
	ledger := NewTransactionLedger(db)
	ctx := context.Background()

	// Record transactions for different symbols
	for _, sym := range []string{"FPT", "FPT", "SSI"} {
		ledger.RecordTransaction(ctx, model.Transaction{
			UserID: 1, AssetType: model.VNStock, Symbol: sym,
			Quantity: 10, UnitPrice: 85000, TransactionDate: time.Now(), TransactionType: model.Buy,
		})
	}

	txns, err := ledger.GetTransactionsBySymbol(ctx, 1, "FPT")
	if err != nil {
		t.Fatalf("GetTransactionsBySymbol failed: %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("expected 2 FPT transactions, got %d", len(txns))
	}
}
