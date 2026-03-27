package service

import (
	"context"
	"testing"
	"time"

	"myfi-backend/internal/model"
	"myfi-backend/internal/testutil"
)

func newTestGoalPlanner(t *testing.T) *GoalPlanner {
	t.Helper()
	return NewGoalPlanner(testutil.SetupPostgresTestDB(t))
}

// --- Pure unit tests for ComputeProgress ---

func TestComputeProgress_Basic(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	target := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // 12 months away

	p := ComputeProgress(500_000_000, 1_000_000_000, target, now)

	if p.ProgressPercent != 50.0 {
		t.Errorf("expected progress 50%%, got %.2f%%", p.ProgressPercent)
	}
	if p.MonthsRemaining != 12 {
		t.Errorf("expected 12 months remaining, got %d", p.MonthsRemaining)
	}
	// shortfall = 500M, months = 12 → ceil(500M/12) = 41_666_667
	expected := float64(41_666_667)
	if p.RequiredMonthlyContribution != expected {
		t.Errorf("expected monthly contribution %.0f, got %.0f", expected, p.RequiredMonthlyContribution)
	}
}

func TestComputeProgress_GoalAlreadyMet(t *testing.T) {
	now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	target := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	p := ComputeProgress(1_200_000_000, 1_000_000_000, target, now)

	if p.ProgressPercent != 120.0 {
		t.Errorf("expected progress 120%%, got %.2f%%", p.ProgressPercent)
	}
	if p.RequiredMonthlyContribution != 0 {
		t.Errorf("expected 0 contribution when goal met, got %.0f", p.RequiredMonthlyContribution)
	}
}

func TestComputeProgress_TargetDatePassed(t *testing.T) {
	now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	target := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // in the past

	p := ComputeProgress(300_000_000, 1_000_000_000, target, now)

	if p.MonthsRemaining != 0 {
		t.Errorf("expected 0 months remaining for past target, got %d", p.MonthsRemaining)
	}
	if p.RequiredMonthlyContribution != 0 {
		t.Errorf("expected 0 contribution for past target, got %.0f", p.RequiredMonthlyContribution)
	}
	if p.ProgressPercent != 30.0 {
		t.Errorf("expected progress 30%%, got %.2f%%", p.ProgressPercent)
	}
}

func TestComputeProgress_ZeroNAV(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	target := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC) // 6 months

	p := ComputeProgress(0, 600_000_000, target, now)

	if p.ProgressPercent != 0 {
		t.Errorf("expected 0%% progress, got %.2f%%", p.ProgressPercent)
	}
	// shortfall = 600M, months = 6 → 100M
	if p.RequiredMonthlyContribution != 100_000_000 {
		t.Errorf("expected 100M monthly, got %.0f", p.RequiredMonthlyContribution)
	}
}

func TestComputeProgress_ZeroTargetAmount(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	target := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	p := ComputeProgress(100_000, 0, target, now)

	if p.ProgressPercent != 0 {
		t.Errorf("expected 0%% for zero target, got %.2f%%", p.ProgressPercent)
	}
}

// --- Unit tests for monthsBetween ---

func TestMonthsBetween(t *testing.T) {
	tests := []struct {
		name     string
		from, to time.Time
		want     int
	}{
		{"same date", time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), 0},
		{"one month", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), 1},
		{"12 months", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 12},
		{"partial month not counted", time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC), 0},
		{"past date", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monthsBetween(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("monthsBetween(%v, %v) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

// --- Unit tests for asset type serialization ---

func TestJoinAndParseAssetTypes(t *testing.T) {
	types := []model.AssetType{model.VNStock, model.Gold, model.Crypto}
	joined := joinAssetTypes(types)
	parsed := parseAssetTypes(joined)

	if len(parsed) != 3 {
		t.Fatalf("expected 3 asset types, got %d", len(parsed))
	}
	for i, expected := range types {
		if parsed[i] != expected {
			t.Errorf("index %d: expected %s, got %s", i, expected, parsed[i])
		}
	}
}

func TestJoinAssetTypes_Empty(t *testing.T) {
	if joinAssetTypes(nil) != "" {
		t.Error("expected empty string for nil slice")
	}
}

func TestParseAssetTypes_Empty(t *testing.T) {
	if parseAssetTypes("") != nil {
		t.Error("expected nil for empty string")
	}
}

// --- Unit tests for ValidateGoalCategory ---

func TestValidateGoalCategory_Valid(t *testing.T) {
	categories := []model.GoalCategory{
		model.GoalRetirement, model.GoalEmergencyFund,
		model.GoalProperty, model.GoalEducation, model.GoalCustom,
	}
	for _, c := range categories {
		if err := model.ValidateGoalCategory(c); err != nil {
			t.Errorf("expected %s to be valid, got error: %v", c, err)
		}
	}
}

func TestValidateGoalCategory_Invalid(t *testing.T) {
	if err := model.ValidateGoalCategory("invalid_category"); err == nil {
		t.Error("expected error for invalid category")
	}
}

// --- Integration tests (testcontainers) ---

func TestCreateGoal(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	id, err := planner.CreateGoal(ctx, model.FinancialGoal{
		UserID:               1,
		Name:                 "Retirement Fund",
		TargetAmount:         5_000_000_000,
		TargetDate:           time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC),
		AssociatedAssetTypes: []model.AssetType{model.VNStock, model.Gold},
		Category:             model.GoalRetirement,
	})
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// Verify retrieval
	goal, err := planner.GetGoal(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetGoal failed: %v", err)
	}
	if goal.Name != "Retirement Fund" {
		t.Errorf("expected name 'Retirement Fund', got %q", goal.Name)
	}
	if goal.TargetAmount != 5_000_000_000 {
		t.Errorf("expected target 5B, got %.0f", goal.TargetAmount)
	}
	if goal.Category != model.GoalRetirement {
		t.Errorf("expected category retirement, got %s", goal.Category)
	}
	if len(goal.AssociatedAssetTypes) != 2 {
		t.Errorf("expected 2 asset types, got %d", len(goal.AssociatedAssetTypes))
	}
}

func TestCreateGoal_Validation(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	tests := []struct {
		name string
		goal model.FinancialGoal
	}{
		{"empty name", model.FinancialGoal{UserID: 1, TargetAmount: 100, TargetDate: time.Now().Add(24 * time.Hour), Category: model.GoalCustom}},
		{"zero target", model.FinancialGoal{UserID: 1, Name: "Test", TargetAmount: 0, TargetDate: time.Now().Add(24 * time.Hour), Category: model.GoalCustom}},
		{"negative target", model.FinancialGoal{UserID: 1, Name: "Test", TargetAmount: -100, TargetDate: time.Now().Add(24 * time.Hour), Category: model.GoalCustom}},
		{"zero user", model.FinancialGoal{UserID: 0, Name: "Test", TargetAmount: 100, TargetDate: time.Now().Add(24 * time.Hour), Category: model.GoalCustom}},
		{"zero date", model.FinancialGoal{UserID: 1, Name: "Test", TargetAmount: 100, Category: model.GoalCustom}},
		{"invalid category", model.FinancialGoal{UserID: 1, Name: "Test", TargetAmount: 100, TargetDate: time.Now().Add(24 * time.Hour), Category: "bogus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planner.CreateGoal(ctx, tt.goal)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestUpdateGoal(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	id, err := planner.CreateGoal(ctx, model.FinancialGoal{
		UserID:       1,
		Name:         "Emergency Fund",
		TargetAmount: 100_000_000,
		TargetDate:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Category:     model.GoalEmergencyFund,
	})
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	err = planner.UpdateGoal(ctx, model.FinancialGoal{
		ID:           id,
		UserID:       1,
		Name:         "Emergency Fund v2",
		TargetAmount: 200_000_000,
		TargetDate:   time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		Category:     model.GoalEmergencyFund,
	})
	if err != nil {
		t.Fatalf("UpdateGoal failed: %v", err)
	}

	goal, err := planner.GetGoal(ctx, id, 1)
	if err != nil {
		t.Fatalf("GetGoal failed: %v", err)
	}
	if goal.Name != "Emergency Fund v2" {
		t.Errorf("expected updated name, got %q", goal.Name)
	}
	if goal.TargetAmount != 200_000_000 {
		t.Errorf("expected updated target 200M, got %.0f", goal.TargetAmount)
	}
}

func TestUpdateGoal_NotFound(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	err := planner.UpdateGoal(ctx, model.FinancialGoal{
		ID: 9999, UserID: 1, Name: "X", TargetAmount: 100,
		TargetDate: time.Now().Add(24 * time.Hour), Category: model.GoalCustom,
	})
	if err == nil {
		t.Error("expected error for non-existent goal")
	}
}

func TestDeleteGoal(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	id, err := planner.CreateGoal(ctx, model.FinancialGoal{
		UserID: 1, Name: "To Delete", TargetAmount: 50_000_000,
		TargetDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Category: model.GoalCustom,
	})
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	if err := planner.DeleteGoal(ctx, id, 1); err != nil {
		t.Fatalf("DeleteGoal failed: %v", err)
	}

	_, err = planner.GetGoal(ctx, id, 1)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestDeleteGoal_NotFound(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	err := planner.DeleteGoal(ctx, 9999, 1)
	if err == nil {
		t.Error("expected error for non-existent goal")
	}
}

func TestGetGoalsByUser(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	for _, name := range []string{"Goal A", "Goal B", "Goal C"} {
		_, err := planner.CreateGoal(ctx, model.FinancialGoal{
			UserID: 1, Name: name, TargetAmount: 100_000_000,
			TargetDate: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), Category: model.GoalCustom,
		})
		if err != nil {
			t.Fatalf("CreateGoal %s failed: %v", name, err)
		}
	}

	goals, err := planner.GetGoalsByUser(ctx, 1)
	if err != nil {
		t.Fatalf("GetGoalsByUser failed: %v", err)
	}
	if len(goals) != 3 {
		t.Errorf("expected 3 goals, got %d", len(goals))
	}
}

func TestCreateGoal_AllCategories(t *testing.T) {
	planner := newTestGoalPlanner(t)
	ctx := context.Background()

	categories := []model.GoalCategory{
		model.GoalRetirement, model.GoalEmergencyFund,
		model.GoalProperty, model.GoalEducation, model.GoalCustom,
	}
	for _, cat := range categories {
		_, err := planner.CreateGoal(ctx, model.FinancialGoal{
			UserID: 1, Name: "Goal " + string(cat), TargetAmount: 100_000,
			TargetDate: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), Category: cat,
		})
		if err != nil {
			t.Errorf("CreateGoal with category %s failed: %v", cat, err)
		}
	}
}
