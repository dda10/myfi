package portfolio

import (
	"context"
	"database/sql"
	"fmt"
)

// TransactionLedger manages stock-only transaction recording and retrieval.
type TransactionLedger struct {
	db *sql.DB
}

// NewTransactionLedger creates a new TransactionLedger instance.
func NewTransactionLedger(db *sql.DB) *TransactionLedger {
	return &TransactionLedger{db: db}
}

// RecordTransaction persists a new stock transaction to the database.
// Supports buy, sell, dividend, split, and bonus transaction types.
func (l *TransactionLedger) RecordTransaction(ctx context.Context, tx Transaction) (int64, error) {
	if err := ValidateTransactionType(tx.TransactionType); err != nil {
		return 0, err
	}
	if tx.UserID == "" {
		return 0, fmt.Errorf("invalid user ID: %s", tx.UserID)
	}
	if tx.Symbol == "" {
		return 0, fmt.Errorf("symbol is required")
	}
	if tx.Quantity <= 0 && tx.TransactionType != TxDividend {
		return 0, fmt.Errorf("quantity must be positive")
	}
	if tx.UnitPrice < 0 {
		return 0, fmt.Errorf("unit price must be non-negative")
	}

	// Auto-compute total value if not provided.
	if tx.TotalValue == 0 {
		tx.TotalValue = tx.Quantity * tx.UnitPrice
	}

	var id int64
	err := l.db.QueryRowContext(ctx,
		`INSERT INTO transactions (user_id, symbol, tx_type, quantity, unit_price, total_value, realized_pnl, transaction_date, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()) RETURNING id`,
		tx.UserID, tx.Symbol, string(tx.TransactionType), tx.Quantity,
		tx.UnitPrice, tx.TotalValue, tx.RealizedPnL,
		tx.TransactionDate, tx.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %w", err)
	}

	return id, nil
}

// GetTransactionsByUser retrieves all transactions for a user, ordered by date descending.
func (l *TransactionLedger) GetTransactionsByUser(ctx context.Context, userID string) ([]Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, symbol, tx_type, quantity, unit_price, total_value, COALESCE(realized_pnl, 0), transaction_date, COALESCE(notes, ''), created_at
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
func (l *TransactionLedger) GetTransactionsBySymbol(ctx context.Context, userID string, symbol string) ([]Transaction, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT id, user_id, symbol, tx_type, quantity, unit_price, total_value, COALESCE(realized_pnl, 0), transaction_date, COALESCE(notes, ''), created_at
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
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var txns []Transaction
	for rows.Next() {
		var t Transaction
		var txType string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Symbol, &txType, &t.Quantity,
			&t.UnitPrice, &t.TotalValue, &t.RealizedPnL, &t.TransactionDate, &t.Notes, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		t.TransactionType = TransactionType(txType)
		txns = append(txns, t)
	}
	return txns, rows.Err()
}
