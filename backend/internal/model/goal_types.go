package model

import (
	"fmt"
	"strings"
	"time"
)

// GoalCategory represents the category of a financial goal.
type GoalCategory string

const (
	GoalRetirement    GoalCategory = "retirement"
	GoalEmergencyFund GoalCategory = "emergency_fund"
	GoalProperty      GoalCategory = "property"
	GoalEducation     GoalCategory = "education"
	GoalCustom        GoalCategory = "custom"
)

// ValidGoalCategories contains all supported goal categories.
var ValidGoalCategories = map[GoalCategory]bool{
	GoalRetirement:    true,
	GoalEmergencyFund: true,
	GoalProperty:      true,
	GoalEducation:     true,
	GoalCustom:        true,
}

// ValidateGoalCategory checks if the given goal category is supported.
func ValidateGoalCategory(c GoalCategory) error {
	if ValidGoalCategories[c] {
		return nil
	}
	var supported []string
	for k := range ValidGoalCategories {
		supported = append(supported, string(k))
	}
	return fmt.Errorf("unrecognized goal category %q; supported categories: %s", c, strings.Join(supported, ", "))
}

// FinancialGoal represents a user-defined financial goal.
type FinancialGoal struct {
	ID                   int64        `json:"id"`
	UserID               int64        `json:"userId"`
	Name                 string       `json:"name"`
	TargetAmount         float64      `json:"targetAmount"`
	TargetDate           time.Time    `json:"targetDate"`
	AssociatedAssetTypes []AssetType  `json:"associatedAssetTypes,omitempty"`
	Category             GoalCategory `json:"category"`
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
}

// GoalProgress contains computed progress metrics for a financial goal.
type GoalProgress struct {
	GoalID                      int64   `json:"goalId"`
	CurrentValue                float64 `json:"currentValue"`
	TargetAmount                float64 `json:"targetAmount"`
	ProgressPercent             float64 `json:"progressPercent"`
	RequiredMonthlyContribution float64 `json:"requiredMonthlyContribution"`
	MonthsRemaining             int     `json:"monthsRemaining"`
}
