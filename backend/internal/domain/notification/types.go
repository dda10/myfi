package notification

import "time"

// NotificationType represents the category of a notification.
type NotificationType string

const (
	NotifPriceAlert NotificationType = "price_alert"
	NotifMission    NotificationType = "mission"
	NotifCorporate  NotificationType = "corporate_action"
	NotifRecommend  NotificationType = "recommendation"
	NotifSystem     NotificationType = "system"
)

// Notification represents a single notification sent to a user.
type Notification struct {
	ID        int64            `json:"id"`
	UserID    string           `json:"userId"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Link      string           `json:"link,omitempty"`
	IsRead    bool             `json:"isRead"`
	CreatedAt time.Time        `json:"createdAt"`
}

// NotificationPreference defines per-user notification channel settings.
type NotificationPreference struct {
	ID        int64            `json:"id"`
	UserID    string           `json:"userId"`
	Type      NotificationType `json:"type"`
	InApp     bool             `json:"inApp"`
	Email     bool             `json:"email"`
	UpdatedAt time.Time        `json:"updatedAt"`
}
