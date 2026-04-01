package model

import "time"

// PatternType represents the type of market pattern detected.
type PatternType string

const (
	PatternAccumulation PatternType = "accumulation"
	PatternDistribution PatternType = "distribution"
	PatternBreakout     PatternType = "breakout"
)

// PatternObservation represents a detected market pattern with outcome tracking.
type PatternObservation struct {
	ID               int64       `json:"id"`
	Symbol           string      `json:"symbol"`
	PatternType      PatternType `json:"patternType"`
	DetectionDate    time.Time   `json:"detectionDate"`
	ConfidenceScore  int         `json:"confidenceScore"`
	PriceAtDetection float64     `json:"priceAtDetection"`
	SupportingData   string      `json:"supportingData"`
	Outcome1Day      *float64    `json:"outcome1Day,omitempty"`
	Outcome7Day      *float64    `json:"outcome7Day,omitempty"`
	Outcome14Day     *float64    `json:"outcome14Day,omitempty"`
	Outcome30Day     *float64    `json:"outcome30Day,omitempty"`
}
