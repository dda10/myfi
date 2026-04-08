package mission

import "time"

// MissionStatus represents the current state of a mission.
type MissionStatus string

const (
	MissionActive  MissionStatus = "active"
	MissionPaused  MissionStatus = "paused"
	MissionExpired MissionStatus = "expired"
)

// TriggerType represents the type of trigger condition for a mission.
type TriggerType string

const (
	TriggerPriceAbove TriggerType = "price_above"
	TriggerPriceBelow TriggerType = "price_below"
	TriggerSchedule   TriggerType = "schedule"
	TriggerEvent      TriggerType = "event"
	TriggerNews       TriggerType = "news"
)

// ActionType represents the action to take when a mission triggers.
type ActionType string

const (
	ActionAlert    ActionType = "alert"
	ActionReport   ActionType = "report"
	ActionAnalysis ActionType = "analysis"
)

// Mission represents a user-defined scheduled monitoring task.
type Mission struct {
	ID            int64          `json:"id"`
	UserID        string         `json:"userId"`
	Name          string         `json:"name"`
	Status        MissionStatus  `json:"status"`
	Trigger       MissionTrigger `json:"trigger"`
	Action        MissionAction  `json:"action"`
	TargetSymbols []string       `json:"targetSymbols"`
	LastTriggered *time.Time     `json:"lastTriggered,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// MissionTrigger defines the condition that activates a mission.
type MissionTrigger struct {
	Type           TriggerType `json:"type"`
	PriceThreshold float64     `json:"priceThreshold,omitempty"`
	CronExpression string      `json:"cronExpression,omitempty"`
	EventType      string      `json:"eventType,omitempty"`
}

// MissionAction defines what happens when a mission triggers.
type MissionAction struct {
	Type    ActionType `json:"type"`
	Channel string     `json:"channel,omitempty"` // in_app, email
	Message string     `json:"message,omitempty"`
}
