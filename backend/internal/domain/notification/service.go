package notification

// ---------------------------------------------------------------------------
// NotificationService — user notifications with rate limiting
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 42.1 Send notifications
//   - 42.2 List notifications
//   - 42.3 MarkRead
//   - 42.4 MarkAllRead
//   - 42.5 GetUnreadCount
//   - 42.6 Rate limiting: 1/symbol/channel/hour, 10 emails/user/day
//   - 42.8 Rate limiting enforcement

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// rateLimitEntry tracks notification sends for rate limiting.
type rateLimitEntry struct {
	Count    int
	WindowAt time.Time
}

// NotificationService manages user notifications with rate limiting.
type NotificationService struct {
	db    *sql.DB
	email EmailSender

	// Rate limiting state (in-memory)
	mu             sync.Mutex
	symbolLimits   map[string]*rateLimitEntry // key: "userID:symbol:channel"
	emailDayLimits map[string]*rateLimitEntry // key: "userID:date"
}

// EmailSender is the interface for sending email notifications.
type EmailSender interface {
	Send(to, subject, htmlBody string) error
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(db *sql.DB, email EmailSender) *NotificationService {
	return &NotificationService{
		db:             db,
		email:          email,
		symbolLimits:   make(map[string]*rateLimitEntry),
		emailDayLimits: make(map[string]*rateLimitEntry),
	}
}

// Send creates and persists a notification, respecting rate limits.
// Rate limits: 1 per symbol per channel per hour, 10 emails per user per day.
func (s *NotificationService) Send(ctx context.Context, n *Notification) error {
	if n.Title == "" {
		return fmt.Errorf("notification title is required")
	}

	// Check rate limits
	if err := s.checkRateLimit(n); err != nil {
		log.Printf("[NotificationService] Rate limited: %v", err)
		return err
	}

	now := time.Now()
	n.IsRead = false
	n.CreatedAt = now

	query := `INSERT INTO notifications (user_id, type, title, message, link, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := s.db.QueryRowContext(ctx, query,
		n.UserID, string(n.Type), n.Title, n.Message, n.Link, n.IsRead, n.CreatedAt,
	).Scan(&n.ID)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Record rate limit usage
	s.recordRateLimitUsage(n)

	log.Printf("[NotificationService] Sent notification %d to user %s: %s", n.ID, n.UserID, n.Title)
	return nil
}

// SendMissionAlert is a convenience method for mission-triggered notifications.
func (s *NotificationService) SendMissionAlert(ctx context.Context, userID string, missionName, message string) error {
	return s.Send(ctx, &Notification{
		UserID:  userID,
		Type:    NotifMission,
		Title:   fmt.Sprintf("Mission: %s", missionName),
		Message: message,
	})
}

// List returns notifications for a user, ordered by most recent first.
func (s *NotificationService) List(ctx context.Context, userID string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, user_id, type, title, message, link, is_read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var nType string
		if err := rows.Scan(&n.ID, &n.UserID, &nType, &n.Title, &n.Message, &n.Link, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		n.Type = NotificationType(nType)
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, userID string, notifID int64) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`,
		notifID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}
	return nil
}

// MarkAllRead marks all notifications as read for a user.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`,
		userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetUnreadCount returns the number of unread notifications for a user.
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`,
		userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// --- Rate Limiting ---

// checkRateLimit enforces:
//   - 1 notification per symbol per channel per hour
//   - 10 emails per user per day
func (s *NotificationService) checkRateLimit(n *Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Per-symbol per-channel hourly limit (extract symbol from link or title)
	symbolKey := fmt.Sprintf("%s:%s:%s", n.UserID, n.Title, n.Type)
	if entry, ok := s.symbolLimits[symbolKey]; ok {
		if now.Sub(entry.WindowAt) < time.Hour && entry.Count >= 1 {
			return fmt.Errorf("rate limit: max 1 notification per symbol per channel per hour")
		}
		if now.Sub(entry.WindowAt) >= time.Hour {
			entry.Count = 0
			entry.WindowAt = now
		}
	}

	// Daily email limit
	dayKey := fmt.Sprintf("%s:%s", n.UserID, now.Format("2006-01-02"))
	if n.Type == NotifPriceAlert || n.Type == NotifMission {
		if entry, ok := s.emailDayLimits[dayKey]; ok {
			if entry.Count >= 10 {
				return fmt.Errorf("rate limit: max 10 email notifications per user per day")
			}
		}
	}

	return nil
}

// recordRateLimitUsage records that a notification was sent for rate limiting.
func (s *NotificationService) recordRateLimitUsage(n *Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Per-symbol per-channel
	symbolKey := fmt.Sprintf("%s:%s:%s", n.UserID, n.Title, n.Type)
	entry, ok := s.symbolLimits[symbolKey]
	if !ok {
		entry = &rateLimitEntry{WindowAt: now}
		s.symbolLimits[symbolKey] = entry
	}
	entry.Count++

	// Daily email count
	dayKey := fmt.Sprintf("%s:%s", n.UserID, now.Format("2006-01-02"))
	dayEntry, ok := s.emailDayLimits[dayKey]
	if !ok {
		dayEntry = &rateLimitEntry{WindowAt: now}
		s.emailDayLimits[dayKey] = dayEntry
	}
	dayEntry.Count++
}

// CleanupRateLimits removes stale rate limit entries (call periodically).
func (s *NotificationService) CleanupRateLimits() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, v := range s.symbolLimits {
		if now.Sub(v.WindowAt) > 2*time.Hour {
			delete(s.symbolLimits, k)
		}
	}
	for k, v := range s.emailDayLimits {
		if now.Sub(v.WindowAt) > 48*time.Hour {
			delete(s.emailDayLimits, k)
		}
	}
}
