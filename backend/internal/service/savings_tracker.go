package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"myfi-backend/internal/model"
)

// SavingsTracker manages savings accounts, term deposits, and interest calculations.
type SavingsTracker struct {
	db *sql.DB
}

// NewSavingsTracker creates a new SavingsTracker instance.
func NewSavingsTracker(db *sql.DB) *SavingsTracker {
	return &SavingsTracker{db: db}
}

// AddSavingsAccount persists a new savings account to the database.
// Requirement 6.1: record principal, annual rate, compounding frequency, start date, maturity date.
// Requirement 6.5: bank current accounts use zero or specified interest rate (maturity_date = nil).
func (s *SavingsTracker) AddSavingsAccount(ctx context.Context, account model.SavingsAccount) (int64, error) {
	if account.UserID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", account.UserID)
	}
	if account.AccountName == "" {
		return 0, fmt.Errorf("account name is required")
	}
	if account.Principal < 0 {
		return 0, fmt.Errorf("principal must be non-negative")
	}
	if account.AnnualRate < 0 {
		return 0, fmt.Errorf("annual rate must be non-negative")
	}
	if err := model.ValidateCompoundingFrequency(account.CompoundingFrequency); err != nil {
		return 0, err
	}

	var id int64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO savings_accounts (user_id, account_name, principal, annual_rate, compounding_frequency, start_date, maturity_date, is_matured)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		account.UserID, account.AccountName, account.Principal, account.AnnualRate,
		string(account.CompoundingFrequency), account.StartDate,
		account.MaturityDate, false,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert savings account: %w", err)
	}
	return id, nil
}

// GetSavingsAccount retrieves a single savings account by ID and user.
func (s *SavingsTracker) GetSavingsAccount(ctx context.Context, accountID, userID int64) (model.SavingsAccount, error) {
	var sa model.SavingsAccount
	var freq string
	var startDate, createdAt time.Time
	var maturityDate sql.NullTime
	var isMatured bool

	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_name, principal, annual_rate, compounding_frequency, start_date, maturity_date, is_matured, created_at
		 FROM savings_accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID,
	).Scan(&sa.ID, &sa.UserID, &sa.AccountName, &sa.Principal, &sa.AnnualRate,
		&freq, &startDate, &maturityDate, &isMatured, &createdAt)
	if err != nil {
		return model.SavingsAccount{}, fmt.Errorf("savings account not found: %w", err)
	}

	sa.CompoundingFrequency = model.CompoundingFrequency(freq)
	sa.StartDate = startDate
	sa.IsMatured = isMatured
	sa.CreatedAt = createdAt
	if maturityDate.Valid {
		sa.MaturityDate = &maturityDate.Time
	}
	return sa, nil
}

// GetSavingsAccountsByUser retrieves all savings accounts for a given user.
func (s *SavingsTracker) GetSavingsAccountsByUser(ctx context.Context, userID int64) ([]model.SavingsAccount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, account_name, principal, annual_rate, compounding_frequency, start_date, maturity_date, is_matured, created_at
		 FROM savings_accounts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query savings accounts: %w", err)
	}
	defer rows.Close()

	var accounts []model.SavingsAccount
	for rows.Next() {
		var sa model.SavingsAccount
		var freq string
		var startDate, createdAt time.Time
		var maturityDate sql.NullTime
		var isMatured bool

		if err := rows.Scan(&sa.ID, &sa.UserID, &sa.AccountName, &sa.Principal, &sa.AnnualRate,
			&freq, &startDate, &maturityDate, &isMatured, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan savings account row: %w", err)
		}

		sa.CompoundingFrequency = model.CompoundingFrequency(freq)
		sa.StartDate = startDate
		sa.IsMatured = isMatured
		sa.CreatedAt = createdAt
		if maturityDate.Valid {
			sa.MaturityDate = &maturityDate.Time
		}
		accounts = append(accounts, sa)
	}
	return accounts, rows.Err()
}

// CalculateAccruedValue computes the total value (principal + accrued interest) using
// the compound interest formula: A = P × (1 + r/n)^(n×t)
// Requirement 6.2: P = principal, r = annual rate, n = compounding frequency, t = elapsed time in years.
func CalculateAccruedValue(principal, annualRate float64, freq model.CompoundingFrequency, startDate, asOf time.Time) float64 {
	if principal <= 0 || annualRate <= 0 {
		return principal
	}

	n := float64(model.CompoundingPeriodsPerYear(freq))
	t := elapsedYears(startDate, asOf)
	if t <= 0 {
		return principal
	}

	// A = P × (1 + r/n)^(n×t)
	return principal * math.Pow(1+annualRate/n, n*t)
}

// CalculateAccruedInterest returns only the interest portion (A - P).
func CalculateAccruedInterest(principal, annualRate float64, freq model.CompoundingFrequency, startDate, asOf time.Time) float64 {
	return CalculateAccruedValue(principal, annualRate, freq, startDate, asOf) - principal
}

// CheckMaturity determines if a term deposit has reached its maturity date.
// Requirement 6.3: flag deposit as matured when current date >= maturity date.
func CheckMaturity(maturityDate *time.Time, asOf time.Time) bool {
	if maturityDate == nil {
		return false // No maturity date (e.g., current account) — never matures
	}
	return !asOf.Before(*maturityDate)
}

// UpdateMaturityStatus checks and updates the maturity flag for a savings account.
func (s *SavingsTracker) UpdateMaturityStatus(ctx context.Context, accountID, userID int64) (bool, error) {
	account, err := s.GetSavingsAccount(ctx, accountID, userID)
	if err != nil {
		return false, err
	}

	matured := CheckMaturity(account.MaturityDate, time.Now())
	if matured && !account.IsMatured {
		_, err := s.db.ExecContext(ctx,
			`UPDATE savings_accounts SET is_matured = $1 WHERE id = $2 AND user_id = $3`,
			true, accountID, userID,
		)
		if err != nil {
			return false, fmt.Errorf("failed to update maturity status: %w", err)
		}
	}
	return matured, nil
}

// RefreshAllMaturityStatuses checks and updates maturity for all accounts of a user.
func (s *SavingsTracker) RefreshAllMaturityStatuses(ctx context.Context, userID int64) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx,
		`UPDATE savings_accounts SET is_matured = TRUE
		 WHERE user_id = $1 AND maturity_date IS NOT NULL AND maturity_date <= $2 AND is_matured = FALSE`,
		userID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to refresh maturity statuses: %w", err)
	}
	return nil
}

// ComputeSavingsNAV calculates the total NAV contribution from all savings accounts.
// Requirement 6.4: include accrued interest in NAV calculation.
func (s *SavingsTracker) ComputeSavingsNAV(ctx context.Context, userID int64) (float64, error) {
	accounts, err := s.GetSavingsAccountsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var totalNAV float64
	for _, acc := range accounts {
		totalNAV += CalculateAccruedValue(acc.Principal, acc.AnnualRate, acc.CompoundingFrequency, acc.StartDate, now)
	}
	return totalNAV, nil
}

// DeleteSavingsAccount removes a savings account.
func (s *SavingsTracker) DeleteSavingsAccount(ctx context.Context, accountID, userID int64) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM savings_accounts WHERE id = $1 AND user_id = $2`,
		accountID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete savings account: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("savings account %d not found for user %d", accountID, userID)
	}
	return nil
}

// elapsedYears computes the fractional number of years between two dates
// using exact calendar-based calculation for precision.
func elapsedYears(start, end time.Time) float64 {
	if start.IsZero() || end.IsZero() {
		return 0
	}

	// Normalize both to UTC to avoid timezone issues
	start = start.UTC()
	end = end.UTC()

	// Calculate whole years
	years := end.Year() - start.Year()

	// Create anniversary date in the end year
	anniversary := time.Date(end.Year(), start.Month(), start.Day(),
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), time.UTC)

	if end.Before(anniversary) {
		years--
		// Calculate fractional part: days since last anniversary / days in that year
		prevAnniversary := time.Date(end.Year()-1, start.Month(), start.Day(),
			start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), time.UTC)
		daysSinceAnniversary := end.Sub(prevAnniversary).Hours() / 24
		daysInYear := anniversary.Sub(prevAnniversary).Hours() / 24
		if daysInYear == 0 {
			return float64(years)
		}
		return float64(years) + daysSinceAnniversary/daysInYear
	}

	// Calculate fractional part: days since anniversary / days until next anniversary
	nextAnniversary := time.Date(end.Year()+1, start.Month(), start.Day(),
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond(), time.UTC)
	daysSinceAnniversary := end.Sub(anniversary).Hours() / 24
	daysInYear := nextAnniversary.Sub(anniversary).Hours() / 24
	if daysInYear == 0 {
		return float64(years)
	}
	return float64(years) + daysSinceAnniversary/daysInYear
}
