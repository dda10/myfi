package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"myfi-backend/internal/model"
)

// GoalPlanner manages user-defined financial goals with progress tracking.
// Requirement 31: Goal-based financial planning with CRUD, progress, and monthly contribution.
type GoalPlanner struct {
	db *sql.DB
}

// NewGoalPlanner creates a new GoalPlanner instance.
func NewGoalPlanner(db *sql.DB) *GoalPlanner {
	return &GoalPlanner{db: db}
}

// CreateGoal persists a new financial goal to the database.
// Requirement 31.1: create goals with name, target amount, target date, associated asset types.
// Requirement 31.6: support goal categories.
// Requirement 31.7: persist goals in database.
func (g *GoalPlanner) CreateGoal(ctx context.Context, goal model.FinancialGoal) (int64, error) {
	if goal.UserID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", goal.UserID)
	}
	if goal.Name == "" {
		return 0, fmt.Errorf("goal name is required")
	}
	if goal.TargetAmount <= 0 {
		return 0, fmt.Errorf("target amount must be positive")
	}
	if goal.TargetDate.IsZero() {
		return 0, fmt.Errorf("target date is required")
	}
	if err := model.ValidateGoalCategory(goal.Category); err != nil {
		return 0, err
	}

	assetTypesStr := joinAssetTypes(goal.AssociatedAssetTypes)

	var id int64
	err := g.db.QueryRowContext(ctx,
		`INSERT INTO financial_goals (user_id, name, target_amount, target_date, associated_asset_types, category)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		goal.UserID, goal.Name, goal.TargetAmount, goal.TargetDate, assetTypesStr, string(goal.Category),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert financial goal: %w", err)
	}
	return id, nil
}

// UpdateGoal updates an existing financial goal.
// Requirement 31.1: edit goals.
func (g *GoalPlanner) UpdateGoal(ctx context.Context, goal model.FinancialGoal) error {
	if goal.ID <= 0 {
		return fmt.Errorf("invalid goal ID: %d", goal.ID)
	}
	if goal.UserID <= 0 {
		return fmt.Errorf("invalid user ID: %d", goal.UserID)
	}
	if goal.Name == "" {
		return fmt.Errorf("goal name is required")
	}
	if goal.TargetAmount <= 0 {
		return fmt.Errorf("target amount must be positive")
	}
	if goal.TargetDate.IsZero() {
		return fmt.Errorf("target date is required")
	}
	if err := model.ValidateGoalCategory(goal.Category); err != nil {
		return err
	}

	assetTypesStr := joinAssetTypes(goal.AssociatedAssetTypes)

	result, err := g.db.ExecContext(ctx,
		`UPDATE financial_goals SET name = $1, target_amount = $2, target_date = $3,
		 associated_asset_types = $4, category = $5, updated_at = NOW()
		 WHERE id = $6 AND user_id = $7`,
		goal.Name, goal.TargetAmount, goal.TargetDate, assetTypesStr,
		string(goal.Category), goal.ID, goal.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update financial goal: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("goal %d not found for user %d", goal.ID, goal.UserID)
	}
	return nil
}

// DeleteGoal removes a financial goal.
// Requirement 31.1: delete goals.
func (g *GoalPlanner) DeleteGoal(ctx context.Context, goalID, userID int64) error {
	result, err := g.db.ExecContext(ctx,
		`DELETE FROM financial_goals WHERE id = $1 AND user_id = $2`,
		goalID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete financial goal: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("goal %d not found for user %d", goalID, userID)
	}
	return nil
}

// GetGoal retrieves a single financial goal by ID and user.
func (g *GoalPlanner) GetGoal(ctx context.Context, goalID, userID int64) (model.FinancialGoal, error) {
	var goal model.FinancialGoal
	var assetTypesStr sql.NullString
	var category sql.NullString

	err := g.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, target_amount, target_date, associated_asset_types, category, created_at, updated_at
		 FROM financial_goals WHERE id = $1 AND user_id = $2`,
		goalID, userID,
	).Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.TargetDate,
		&assetTypesStr, &category, &goal.CreatedAt, &goal.UpdatedAt)
	if err != nil {
		return model.FinancialGoal{}, fmt.Errorf("goal not found: %w", err)
	}

	if assetTypesStr.Valid && assetTypesStr.String != "" {
		goal.AssociatedAssetTypes = parseAssetTypes(assetTypesStr.String)
	}
	if category.Valid {
		goal.Category = model.GoalCategory(category.String)
	}
	return goal, nil
}

// GetGoalsByUser retrieves all financial goals for a given user.
func (g *GoalPlanner) GetGoalsByUser(ctx context.Context, userID int64) ([]model.FinancialGoal, error) {
	rows, err := g.db.QueryContext(ctx,
		`SELECT id, user_id, name, target_amount, target_date, associated_asset_types, category, created_at, updated_at
		 FROM financial_goals WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query financial goals: %w", err)
	}
	defer rows.Close()

	var goals []model.FinancialGoal
	for rows.Next() {
		var goal model.FinancialGoal
		var assetTypesStr sql.NullString
		var category sql.NullString

		if err := rows.Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.TargetDate,
			&assetTypesStr, &category, &goal.CreatedAt, &goal.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan goal row: %w", err)
		}

		if assetTypesStr.Valid && assetTypesStr.String != "" {
			goal.AssociatedAssetTypes = parseAssetTypes(assetTypesStr.String)
		}
		if category.Valid {
			goal.Category = model.GoalCategory(category.String)
		}
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}

// ComputeProgress computes the progress percentage for a goal.
// Requirement 31.2: progress = (currentNAV / targetAmount) * 100.
// Requirement 31.3: required monthly contribution = (targetAmount - currentNAV) / monthsRemaining.
func ComputeProgress(currentNAV, targetAmount float64, targetDate, now time.Time) model.GoalProgress {
	progress := model.GoalProgress{
		CurrentValue: currentNAV,
		TargetAmount: targetAmount,
	}

	if targetAmount > 0 {
		progress.ProgressPercent = (currentNAV / targetAmount) * 100
		// Cap at 100 for display, but allow >100 to indicate over-achievement
	}

	monthsRemaining := monthsBetween(now, targetDate)
	progress.MonthsRemaining = monthsRemaining

	shortfall := targetAmount - currentNAV
	if shortfall > 0 && monthsRemaining > 0 {
		progress.RequiredMonthlyContribution = math.Ceil(shortfall / float64(monthsRemaining))
	}
	// If shortfall <= 0, goal is already met — no contribution needed.
	// If monthsRemaining <= 0, target date has passed — contribution is 0 (cannot be computed).

	return progress
}

// monthsBetween computes the number of whole months between two dates.
// Returns 0 if target is in the past or same month.
func monthsBetween(from, to time.Time) int {
	if !to.After(from) {
		return 0
	}
	months := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month())
	// If the day hasn't been reached yet in the target month, don't count that partial month
	if to.Day() < from.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

// joinAssetTypes serializes a slice of AssetType to a comma-separated string for DB storage.
func joinAssetTypes(types []model.AssetType) string {
	if len(types) == 0 {
		return ""
	}
	parts := make([]string, len(types))
	for i, t := range types {
		parts[i] = string(t)
	}
	return strings.Join(parts, ",")
}

// parseAssetTypes deserializes a comma-separated string back to a slice of AssetType.
func parseAssetTypes(s string) []model.AssetType {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	types := make([]model.AssetType, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			types = append(types, model.AssetType(trimmed))
		}
	}
	return types
}
