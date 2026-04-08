package service

import (
	"context"
	"database/sql"
	"fmt"

	"myfi-backend/internal/model"
)

// TransactionLedger manages transaction recording and retrieval.
type TransactionLedger struct {
	db *sql.DB
}

// NewTransactionLedger creates a new TransactionLedger instance.
func NewTransactionLedger(db *sql.DB) *TransactionLedger {
	return &TransactionLedger{db: db}
}

// RecordTransaction persists a new transaction to the database.
// All monetary values (UnitPrice, TotalValue) must be in VND.
func (l *TransactionLedger) RecordTransaction(ctx context.Context, tx model.Transaction) (int64, error) {
	if err := model.ValidateTransactionType(tx.TransactionType); err != nil {
		return 0, err
	}
	if err := model.ValidateAssetType(tx.AssetType); err != nil {
		return 0, err
	}
	if tx.UserID == "" {
		return 0, fmt.Errorf("invalid user ID: %s", tx.UserID)
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

	var id int64
	err := l.db.QueryRowContext(ctx,
		`INSERT INTO transactions (user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		tx.UserID, string(tx.AssetType), tx.Symbol, tx.Quantity,
		tx.UnitPrice, tx.TotalValue,
		tx.TransactionDate,
		string(tx.TransactionType), tx.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %w", err)
	}

	return id, nil
}

// GetTransactionsByUser retrieves all transactions for a given user, ordered by date descending.
func (l *TransactionLedger) GetTransactionsByUser(ctx context.Context, userID string) ([]model.Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY transaction_date DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// GetTransactionsBySymbol retrieves transactions for a specific user and symbol.
func (l *TransactionLedger) GetTransactionsBySymbol(ctx context.Context, userID string, symbol string) ([]model.Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, unit_price, total_value, transaction_date, transaction_type, notes, created_at
		 FROM transactions WHERE user_id = $1 AND symbol = $2 ORDER BY transaction_date DESC`,
		userID, symbol,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// scanTransactions scans rows into a slice of Transaction.
func scanTransactions(rows *sql.Rows) ([]model.Transaction, error) {
	var txns []model.Transaction
	for rows.Next() {
		var t model.Transaction
		var assetType, txType string
		var notes sql.NullString

		if err := rows.Scan(&t.ID, &t.UserID, &assetType, &t.Symbol, &t.Quantity,
			&t.UnitPrice, &t.TotalValue, &t.TransactionDate, &txType, &notes, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		t.AssetType = model.AssetType(assetType)
		t.TransactionType = model.TransactionType(txType)
		if notes.Valid {
			t.Notes = notes.String
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}
