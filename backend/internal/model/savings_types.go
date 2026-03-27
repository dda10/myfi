package model

import (
	"fmt"
	"time"
)

// CompoundingFrequency represents how often interest is compounded.
type CompoundingFrequency string

const (
	Monthly   CompoundingFrequency = "monthly"
	Quarterly CompoundingFrequency = "quarterly"
	Yearly    CompoundingFrequency = "yearly"
)

// ValidCompoundingFrequencies contains all supported compounding frequencies.
var ValidCompoundingFrequencies = map[CompoundingFrequency]bool{
	Monthly:   true,
	Quarterly: true,
	Yearly:    true,
}

// CompoundingPeriodsPerYear returns the number of compounding periods per year.
func CompoundingPeriodsPerYear(freq CompoundingFrequency) int {
	switch freq {
	case Monthly:
		return 12
	case Quarterly:
		return 4
	case Yearly:
		return 1
	default:
		return 1
	}
}

// ValidateCompoundingFrequency checks if the given frequency is supported.
func ValidateCompoundingFrequency(freq CompoundingFrequency) error {
	if ValidCompoundingFrequencies[freq] {
		return nil
	}
	return fmt.Errorf("invalid compounding frequency %q; supported: monthly, quarterly, yearly", freq)
}

// SavingsAccount represents a savings account or term deposit.
type SavingsAccount struct {
	ID                   int64                `json:"id"`
	UserID               int64                `json:"userId"`
	AccountName          string               `json:"accountName"`
	Principal            float64              `json:"principal"`
	AnnualRate           float64              `json:"annualRate"`
	CompoundingFrequency CompoundingFrequency `json:"compoundingFrequency"`
	StartDate            time.Time            `json:"startDate"`
	MaturityDate         *time.Time           `json:"maturityDate,omitempty"`
	IsMatured            bool                 `json:"isMatured"`
	CreatedAt            time.Time            `json:"createdAt"`
}
