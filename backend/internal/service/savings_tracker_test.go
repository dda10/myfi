package service

import (
	"context"
	"math"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"
)

func newTestSavingsTracker(t *testing.T) *SavingsTracker {
	t.Helper()
	return NewSavingsTracker(testutil.SetupPostgresTestDB(t))
}

func TestAddSavingsAccount_TermDeposit(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	maturity := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	id, err := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID:               1,
		AccountName:          "TPBank 12M",
		Principal:            100_000_000, // 100M VND
		AnnualRate:           0.055,       // 5.5%
		CompoundingFrequency: model.Monthly,
		StartDate:            time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		MaturityDate:         &maturity,
	})
	if err != nil {
		t.Fatalf("AddSavingsAccount failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// Verify persistence
	account, err := tracker.GetSavingsAccount(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetSavingsAccount failed: %v", err)
	}
	if account.AccountName != "TPBank 12M" {
		t.Errorf("expected account name 'TPBank 12M', got %q", account.AccountName)
	}
	if account.Principal != 100_000_000 {
		t.Errorf("expected principal 100000000, got %f", account.Principal)
	}
	if account.AnnualRate != 0.055 {
		t.Errorf("expected annual rate 0.055, got %f", account.AnnualRate)
	}
	if account.CompoundingFrequency != model.Monthly {
		t.Errorf("expected monthly compounding, got %s", account.CompoundingFrequency)
	}
	if account.MaturityDate == nil {
		t.Error("expected maturity date to be set")
	}
	if account.IsMatured {
		t.Error("expected is_matured to be false")
	}
}

func TestAddSavingsAccount_CurrentAccount(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	id, err := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID:               1,
		AccountName:          "VCB Current",
		Principal:            50_000_000,
		AnnualRate:           0,
		CompoundingFrequency: model.Yearly,
		StartDate:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MaturityDate:         nil,
	})
	if err != nil {
		t.Fatalf("AddSavingsAccount failed: %v", err)
	}

	account, err := tracker.GetSavingsAccount(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetSavingsAccount failed: %v", err)
	}
	if account.AnnualRate != 0 {
		t.Errorf("expected zero rate, got %f", account.AnnualRate)
	}
	if account.MaturityDate != nil {
		t.Error("expected nil maturity date for current account")
	}
}

func TestAddSavingsAccount_Validation(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		account model.SavingsAccount
	}{
		{"invalid user ID", model.SavingsAccount{UserID: 0, AccountName: "test", Principal: 100, CompoundingFrequency: model.Monthly}},
		{"empty account name", model.SavingsAccount{UserID: 1, AccountName: "", Principal: 100, CompoundingFrequency: model.Monthly}},
		{"negative principal", model.SavingsAccount{UserID: 1, AccountName: "test", Principal: -100, CompoundingFrequency: model.Monthly}},
		{"negative rate", model.SavingsAccount{UserID: 1, AccountName: "test", Principal: 100, AnnualRate: -0.05, CompoundingFrequency: model.Monthly}},
		{"invalid frequency", model.SavingsAccount{UserID: 1, AccountName: "test", Principal: 100, CompoundingFrequency: "daily"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tracker.AddSavingsAccount(ctx, tc.account)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestCalculateAccruedValue_CompoundInterest(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(100_000_000, 0.06, model.Monthly, start, asOf)
	expected := 100_000_000 * math.Pow(1+0.06/12, 12*1)

	if math.Abs(value-expected) > 1 {
		t.Errorf("expected ~%.2f, got %.2f", expected, value)
	}
	t.Logf("100M VND at 6%% monthly compounding for 1 year = %.0f VND (interest: %.0f)", value, value-100_000_000)
}

func TestCalculateAccruedValue_QuarterlyCompounding(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(100_000_000, 0.06, model.Quarterly, start, asOf)
	expected := 100_000_000 * math.Pow(1+0.06/4, 4*1)

	if math.Abs(value-expected) > 1 {
		t.Errorf("expected ~%.2f, got %.2f", expected, value)
	}
}

func TestCalculateAccruedValue_YearlyCompounding(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(100_000_000, 0.06, model.Yearly, start, asOf)
	expected := 100_000_000 * math.Pow(1+0.06/1, 1*1)

	if math.Abs(value-expected) > 1 {
		t.Errorf("expected ~%.2f, got %.2f", expected, value)
	}
}

func TestCalculateAccruedValue_ZeroRate(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(50_000_000, 0, model.Monthly, start, asOf)
	if value != 50_000_000 {
		t.Errorf("expected principal unchanged at 50000000, got %f", value)
	}
}

func TestCalculateAccruedValue_ZeroPrincipal(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(0, 0.06, model.Monthly, start, asOf)
	if value != 0 {
		t.Errorf("expected 0 for zero principal, got %f", value)
	}
}

func TestCalculateAccruedValue_FutureStartDate(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	value := CalculateAccruedValue(100_000_000, 0.06, model.Monthly, start, asOf)
	if value != 100_000_000 {
		t.Errorf("expected principal unchanged for future start, got %f", value)
	}
}

func TestCalculateAccruedInterest(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	interest := CalculateAccruedInterest(100_000_000, 0.06, model.Monthly, start, asOf)
	expectedInterest := 100_000_000*math.Pow(1+0.06/12, 12*1) - 100_000_000

	if math.Abs(interest-expectedInterest) > 1 {
		t.Errorf("expected interest ~%.2f, got %.2f", expectedInterest, interest)
	}
}

func TestCheckMaturity(t *testing.T) {
	maturity := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	if CheckMaturity(&maturity, time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected not matured before maturity date")
	}
	if !CheckMaturity(&maturity, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected matured on maturity date")
	}
	if !CheckMaturity(&maturity, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected matured after maturity date")
	}
	if CheckMaturity(nil, time.Now()) {
		t.Error("expected not matured for nil maturity date")
	}
}

func TestUpdateMaturityStatus(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	pastMaturity := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	id, err := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID:               1,
		AccountName:          "Matured Deposit",
		Principal:            100_000_000,
		AnnualRate:           0.06,
		CompoundingFrequency: model.Monthly,
		StartDate:            time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		MaturityDate:         &pastMaturity,
	})
	if err != nil {
		t.Fatalf("AddSavingsAccount failed: %v", err)
	}

	matured, err := tracker.UpdateMaturityStatus(ctx, id, 1)
	if err != nil {
		t.Fatalf("UpdateMaturityStatus failed: %v", err)
	}
	if !matured {
		t.Error("expected deposit to be flagged as matured")
	}

	account, _ := tracker.GetSavingsAccount(ctx, id, 1)
	if !account.IsMatured {
		t.Error("expected is_matured to be true in database")
	}
}

func TestComputeSavingsNAV(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	maturity := time.Now().AddDate(0, 6, 0)
	tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID:               1,
		AccountName:          "Term Deposit",
		Principal:            100_000_000,
		AnnualRate:           0.06,
		CompoundingFrequency: model.Monthly,
		StartDate:            oneYearAgo,
		MaturityDate:         &maturity,
	})

	tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID:               1,
		AccountName:          "Current Account",
		Principal:            50_000_000,
		AnnualRate:           0,
		CompoundingFrequency: model.Yearly,
		StartDate:            oneYearAgo,
	})

	nav, err := tracker.ComputeSavingsNAV(ctx, 1)
	if err != nil {
		t.Fatalf("ComputeSavingsNAV failed: %v", err)
	}

	if nav <= 150_000_000 {
		t.Errorf("expected NAV > 150M, got %.0f", nav)
	}

	expectedTermValue := 100_000_000 * math.Pow(1+0.06/12, 12*1)
	expectedNAV := expectedTermValue + 50_000_000

	if math.Abs(nav-expectedNAV) > 100_000 {
		t.Errorf("expected NAV ~%.0f, got %.0f", expectedNAV, nav)
	}
	t.Logf("Savings NAV: %.0f VND", nav)
}

func TestGetSavingsAccountsByUser(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID: 1, AccountName: "Account 1", Principal: 100_000_000,
		AnnualRate: 0.05, CompoundingFrequency: model.Monthly,
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID: 1, AccountName: "Account 2", Principal: 50_000_000,
		AnnualRate: 0, CompoundingFrequency: model.Yearly,
		StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	})

	accounts, err := tracker.GetSavingsAccountsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetSavingsAccountsByUser failed: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestDeleteSavingsAccount(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	id, _ := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID: 1, AccountName: "To Delete", Principal: 100_000_000,
		AnnualRate: 0.05, CompoundingFrequency: model.Monthly,
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	err := tracker.DeleteSavingsAccount(ctx, id, 1)
	if err != nil {
		t.Fatalf("DeleteSavingsAccount failed: %v", err)
	}

	_, err = tracker.GetSavingsAccount(ctx, id, 1)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestDeleteSavingsAccount_NotFound(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	err := tracker.DeleteSavingsAccount(ctx, 999, 1)
	if err == nil {
		t.Error("expected error for non-existent account")
	}
}

func TestRefreshAllMaturityStatuses(t *testing.T) {
	tracker := newTestSavingsTracker(t)
	ctx := context.Background()

	pastMaturity := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	id1, _ := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID: 1, AccountName: "Matured", Principal: 100_000_000,
		AnnualRate: 0.06, CompoundingFrequency: model.Monthly,
		StartDate:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		MaturityDate: &pastMaturity,
	})

	futureMaturity := time.Now().AddDate(1, 0, 0)
	id2, _ := tracker.AddSavingsAccount(ctx, model.SavingsAccount{
		UserID: 1, AccountName: "Not Matured", Principal: 50_000_000,
		AnnualRate: 0.05, CompoundingFrequency: model.Quarterly,
		StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MaturityDate: &futureMaturity,
	})

	err := tracker.RefreshAllMaturityStatuses(ctx, 1)
	if err != nil {
		t.Fatalf("RefreshAllMaturityStatuses failed: %v", err)
	}

	acc1, _ := tracker.GetSavingsAccount(ctx, id1, 1)
	if !acc1.IsMatured {
		t.Error("expected matured deposit to be flagged")
	}

	acc2, _ := tracker.GetSavingsAccount(ctx, id2, 1)
	if acc2.IsMatured {
		t.Error("expected future deposit to NOT be flagged")
	}
}

func TestCompoundingPeriodsPerYear(t *testing.T) {
	if model.CompoundingPeriodsPerYear(model.Monthly) != 12 {
		t.Error("expected 12 for monthly")
	}
	if model.CompoundingPeriodsPerYear(model.Quarterly) != 4 {
		t.Error("expected 4 for quarterly")
	}
	if model.CompoundingPeriodsPerYear(model.Yearly) != 1 {
		t.Error("expected 1 for yearly")
	}
}

func TestValidateCompoundingFrequency(t *testing.T) {
	for _, freq := range []model.CompoundingFrequency{model.Monthly, model.Quarterly, model.Yearly} {
		if err := model.ValidateCompoundingFrequency(freq); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", freq, err)
		}
	}
	if err := model.ValidateCompoundingFrequency("daily"); err == nil {
		t.Error("expected error for invalid frequency 'daily'")
	}
}
