package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TransactionType represents the type of a financial transaction.
type TransactionType string

const (
	Buy        TransactionType = "buy"
	Sell       TransactionType = "sell"
	Deposit    TransactionType = "deposit"
	Withdrawal TransactionType = "withdrawal"
	Interest   TransactionType = "interest"
	Dividend   TransactionType = "dividend"
)

// ValidTransactionTypes contains all supported transaction types for validation.
var ValidTransactionTypes = map[TransactionType]bool{
	Buy:        true,
	Sell:       true,
	Deposit:    true,
	Withdrawal: true,
	Interest:   true,
	Dividend:   true,
}

// Transaction represents a single financial transaction in the ledger.
type Transaction struct {
	ID              int64           `json:"id"`
	UserID          int64           `json:"userId"`
	AssetType       AssetType       `json:"assetType"`
	Symbol          string          `json:"symbol"`
	Quantity        float64         `json:"quantity"`
	UnitPrice       float64         `json:"unitPrice"`
	TotalValue      float64         `json:"totalValue"`
	TransactionDate time.Time       `json:"transactionDate"`
	TransactionType TransactionType `json:"transactionType"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// TransactionLedger manages transaction recording and retrieval.
type TransactionLedger struct {
	db *sql.DB
}

// NewTransactionLedger creates a new TransactionLedger instance.
func NewTransactionLedger(db *sql.DB) *TransactionLedger {
	return &TransactionLedger{db: db}
}

// ValidateTransactionType checks if the given transaction type is supported.
func ValidateTransactionType(tt TransactionType) error {
	if ValidTransactionTypes[tt] {
		return nil
	}
	return fmt.Errorf("invalid transaction type %q; supported types: buy, sell, deposit, withdrawal, interest, dividend", tt)
}

// RecordTransaction persists a new transaction to the database.
// All monetary values (UnitPrice, TotalValue) must be in VND.
func (l *TransactionLedger) RecordTransaction(ctx context.Context, tx Transaction) (int64, error) {
	if err := ValidateTransactionType(tx.TransactionType); err != nil {
		return 0, err
	}
	if err := ValidateAssetType(tx.AssetType); err != nil {
		return 0, err
	}
	if tx.UserID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", tx.UserID)
	}
	if tx.Symbol == "" {
		return 0, fmt.Errorf("symbol is required")
	}
	if tx.Quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	if tx.UnitPrice < 0 {
		return 0, fmt.Errorf("unit price must be non-negative")
	}

	// Compute total value if not provided
	if tx.TotalValue == 0 {
		tx.TotalValue = tx.Quantity * tx.UnitPrice
	}

	result, err := l.db.ExecContext(ctx,
		`INSERT INTO transactions (user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tx.UserID, string(tx.AssetType), tx.Symbol, tx.Quantity,
		tx.UnitPrice, tx.TotalValue,
		tx.TransactionDate.Format(time.RFC3339),
		string(tx.TransactionType), tx.Notes,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted transaction ID: %w", err)
	}
	return id, nil
}

// GetTransactionsByUser retrieves all transactions for a given user, ordered by date descending.
func (l *TransactionLedger) GetTransactionsByUser(ctx context.Context, userID int64) ([]Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes, created_at
		 FROM transactions WHERE user_id = ? ORDER BY transaction_date DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// GetTransactionsBySymbol retrieves transactions for a specific user and symbol.
func (l *TransactionLedger) GetTransactionsBySymbol(ctx context.Context, userID int64, symbol string) ([]Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes, created_at
		 FROM transactions WHERE user_id = ? AND symbol = ? ORDER BY transaction_date DESC`,
		userID, symbol,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// scanTransactions scans rows into a slice of Transaction.
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var txns []Transaction
	for rows.Next() {
		var t Transaction
		var assetType, txType, txDate, createdAt string
		var notes sql.NullString

		if err := rows.Scan(&t.ID, &t.UserID, &assetType, &t.Symbol, &t.Quantity,
			&t.UnitPrice, &t.TotalValue, &txDate, &txType, &notes, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		t.AssetType = AssetType(assetType)
		t.TransactionType = TransactionType(txType)
		t.TransactionDate, _ = time.Parse(time.RFC3339, txDate)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if notes.Valid {
			t.Notes = notes.String
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}
