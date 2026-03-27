package model

import "fmt"

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

// ValidateTransactionType checks if the given transaction type is supported.
func ValidateTransactionType(tt TransactionType) error {
	if ValidTransactionTypes[tt] {
		return nil
	}
	return fmt.Errorf("invalid transaction type %q; supported types: buy, sell, deposit, withdrawal, interest, dividend", tt)
}
