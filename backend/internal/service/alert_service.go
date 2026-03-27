package service

// ---------------------------------------------------------------------------
// Alert_Service — proactive notification delivery for pattern alerts
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 14.1 Create alerts for confidence ≥60
//   - 14.2 Deliver in-app notifications
//   - 14.3 Include symbol, pattern type, confidence, explanation, timestamp, chart link
//   - 14.4 Persist alerts in database
//   - 14.5 Support user alert preferences (min confidence, pattern types, symbols)
//   - 14.6 Deduplication (48-hour window, unless confidence +10)
//   - 14.7 Mark alerts as expired after 24 hours if not viewed

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/model"
)

// AlertServiceConfig holds configuration for the Alert_Service.
type AlertServiceConfig struct {
	// DefaultConfidenceThreshold is the minimum confidence to create alerts (default 60)
	DefaultConfidenceThreshold int
	// DeduplicationWindow is the time window for deduplication (default 48 hours)
	DeduplicationWindow time.Duration
	// ConfidenceIncreaseThreshold allows new alert if confidence increases by this amount (default 10)
	ConfidenceIncreaseThreshold int
	// ExpirationWindow is the time after which unviewed alerts are marked expired (default 24 hours)
	ExpirationWindow time.Duration
}

// DefaultAlertServiceConfig returns the default configuration.
func DefaultAlertServiceConfig() AlertServiceConfig {
	return AlertServiceConfig{
		DefaultConfidenceThreshold:  60,
		DeduplicationWindow:         48 * time.Hour,
		ConfidenceIncreaseThreshold: 10,
		ExpirationWindow:            24 * time.Hour,
	}
}

// AlertService manages alert creation, retrieval, and lifecycle.
type AlertService struct {
	db     *sql.DB
	config AlertServiceConfig
}

// NewAlertService creates a new Alert_Service with the given database connection.
func NewAlertService(db *sql.DB) *AlertService {
	return &AlertService{
		db:     db,
		config: DefaultAlertServiceConfig(),
	}
}

// SetConfig updates the service configuration.
func (s *AlertService) SetConfig(cfg AlertServiceConfig) {
	s.config = cfg
}

// ---------------------------------------------------------------------------
// Alert Creation (Requirements 14.1, 14.3, 14.4, 14.6)
// ---------------------------------------------------------------------------

// CreateAlertFromObservation creates an alert from a pattern observation.
// Returns the created alert or nil if the alert was deduplicated or filtered by preferences.
func (s *AlertService) CreateAlertFromObservation(ctx context.Context, userID int64, obs *model.PatternObservation) (*model.Alert, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	// Check confidence threshold (Requirement 14.1)
	if obs.ConfidenceScore < s.config.DefaultConfidenceThreshold {
		log.Printf("[Alert_Service] Skipping alert for %s: confidence %d < threshold %d",
			obs.Symbol, obs.ConfidenceScore, s.config.DefaultConfidenceThreshold)
		return nil, nil
	}

	// Check user preferences (Requirement 14.5)
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		log.Printf("[Alert_Service] Failed to get user preferences: %v", err)
		// Continue with default behavior if preferences unavailable
	} else if prefs != nil {
		if !s.matchesPreferences(obs, prefs) {
			log.Printf("[Alert_Service] Alert filtered by user preferences for %s %s",
				obs.Symbol, obs.PatternType)
			return nil, nil
		}
	}

	// Check for duplicate alerts (Requirement 14.6)
	isDuplicate, err := s.isDuplicateAlert(ctx, userID, obs)
	if err != nil {
		log.Printf("[Alert_Service] Failed to check for duplicate alert: %v", err)
	}
	if isDuplicate {
		log.Printf("[Alert_Service] Skipping duplicate alert for %s %s", obs.Symbol, obs.PatternType)
		return nil, nil
	}

	// Generate explanation (Requirement 14.3)
	explanation := s.generateExplanation(obs)

	// Generate chart link (Requirement 14.3)
	chartLink := fmt.Sprintf("/markets?symbol=%s&tab=chart", obs.Symbol)

	// Create and persist alert (Requirement 14.4)
	alert := &model.Alert{
		UserID:             userID,
		Symbol:             obs.Symbol,
		PatternType:        obs.PatternType,
		ConfidenceScore:    obs.ConfidenceScore,
		Explanation:        explanation,
		DetectionTimestamp: obs.DetectionDate,
		ChartLink:          chartLink,
		Viewed:             false,
		Expired:            false,
		CreatedAt:          time.Now(),
	}

	err = s.persistAlert(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to persist alert: %w", err)
	}

	log.Printf("[Alert_Service] Created alert: %s %s confidence=%d id=%d",
		alert.Symbol, alert.PatternType, alert.ConfidenceScore, alert.ID)

	return alert, nil
}

// persistAlert saves an alert to the database.
func (s *AlertService) persistAlert(ctx context.Context, alert *model.Alert) error {
	query := `
		INSERT INTO alerts 
		(user_id, symbol, pattern_type, confidence_score, explanation, detection_timestamp, chart_link, viewed, expired)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err := s.db.QueryRowContext(ctx, query,
		alert.UserID,
		alert.Symbol,
		string(alert.PatternType),
		alert.ConfidenceScore,
		alert.Explanation,
		alert.DetectionTimestamp,
		alert.ChartLink,
		alert.Viewed,
		alert.Expired,
	).Scan(&alert.ID, &alert.CreatedAt)

	return err
}

// isDuplicateAlert checks if a similar alert was created within the deduplication window.
// Requirement 14.6: 48-hour window, unless confidence increased by 10+
func (s *AlertService) isDuplicateAlert(ctx context.Context, userID int64, obs *model.PatternObservation) (bool, error) {
	query := `
		SELECT confidence_score FROM alerts
		WHERE user_id = $1 AND symbol = $2 AND pattern_type = $3 
		AND detection_timestamp > $4
		ORDER BY detection_timestamp DESC
		LIMIT 1
	`

	deduplicationCutoff := obs.DetectionDate.Add(-s.config.DeduplicationWindow)

	var prevConfidence int
	err := s.db.QueryRowContext(ctx, query, userID, obs.Symbol, string(obs.PatternType), deduplicationCutoff).Scan(&prevConfidence)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Allow new alert if confidence increased by threshold (default 10+ points)
	if obs.ConfidenceScore >= prevConfidence+s.config.ConfidenceIncreaseThreshold {
		return false, nil
	}

	return true, nil
}

// matchesPreferences checks if an observation matches user alert preferences.
func (s *AlertService) matchesPreferences(obs *model.PatternObservation, prefs *model.AlertPreferences) bool {
	// Check minimum confidence threshold
	if obs.ConfidenceScore < prefs.MinConfidence {
		return false
	}

	// Check pattern types filter (empty means all patterns)
	if len(prefs.PatternTypes) > 0 {
		found := false
		for _, pt := range prefs.PatternTypes {
			if pt == obs.PatternType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check symbol inclusion list (empty means all symbols)
	if len(prefs.IncludeSymbols) > 0 {
		found := false
		for _, sym := range prefs.IncludeSymbols {
			if sym == obs.Symbol {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check symbol exclusion list
	for _, sym := range prefs.ExcludeSymbols {
		if sym == obs.Symbol {
			return false
		}
	}

	return true
}

// generateExplanation creates a human-readable explanation for the alert.
func (s *AlertService) generateExplanation(obs *model.PatternObservation) string {
	var explanation string

	switch obs.PatternType {
	case model.PatternAccumulation:
		explanation = fmt.Sprintf(
			"Accumulation pattern detected for %s with %d%% confidence. "+
				"Price is consolidating with elevated volume, suggesting institutional buying.",
			obs.Symbol, obs.ConfidenceScore,
		)
	case model.PatternDistribution:
		explanation = fmt.Sprintf(
			"Distribution pattern detected for %s with %d%% confidence. "+
				"Price near highs with increased selling pressure on down days.",
			obs.Symbol, obs.ConfidenceScore,
		)
	case model.PatternBreakout:
		explanation = fmt.Sprintf(
			"Breakout signal detected for %s with %d%% confidence. "+
				"Price breaking above resistance with strong volume confirmation.",
			obs.Symbol, obs.ConfidenceScore,
		)
	default:
		explanation = fmt.Sprintf(
			"Pattern %s detected for %s with %d%% confidence.",
			obs.PatternType, obs.Symbol, obs.ConfidenceScore,
		)
	}

	// Add supporting data details if available
	if obs.SupportingData != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(obs.SupportingData), &data); err == nil {
			if volumeRatio, ok := data["volumeRatio"].(float64); ok {
				explanation += fmt.Sprintf(" Volume is %.1fx the 20-day average.", volumeRatio)
			}
		}
	}

	return explanation
}

// ---------------------------------------------------------------------------
// Alert Retrieval (Requirement 14.2, 14.4)
// ---------------------------------------------------------------------------

// GetAlerts retrieves alerts for a user with optional filters.
func (s *AlertService) GetAlerts(ctx context.Context, userID int64, filters AlertFilters) ([]model.Alert, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, user_id, symbol, pattern_type, confidence_score, explanation, 
		       detection_timestamp, COALESCE(chart_link, ''), viewed, expired, created_at
		FROM alerts
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if !filters.IncludeExpired {
		query += fmt.Sprintf(" AND expired = $%d", argIdx)
		args = append(args, false)
		argIdx++
	}

	if !filters.IncludeViewed {
		query += fmt.Sprintf(" AND viewed = $%d", argIdx)
		args = append(args, false)
		argIdx++
	}

	if filters.Symbol != "" {
		query += fmt.Sprintf(" AND symbol = $%d", argIdx)
		args = append(args, filters.Symbol)
		argIdx++
	}

	if filters.PatternType != "" {
		query += fmt.Sprintf(" AND pattern_type = $%d", argIdx)
		args = append(args, filters.PatternType)
		argIdx++
	}

	if !filters.StartDate.IsZero() {
		query += fmt.Sprintf(" AND detection_timestamp >= $%d", argIdx)
		args = append(args, filters.StartDate)
		argIdx++
	}

	if !filters.EndDate.IsZero() {
		query += fmt.Sprintf(" AND detection_timestamp <= $%d", argIdx)
		args = append(args, filters.EndDate)
		argIdx++
	}

	query += " ORDER BY detection_timestamp DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filters.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var alert model.Alert
		var patternType string
		err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.Symbol, &patternType,
			&alert.ConfidenceScore, &alert.Explanation, &alert.DetectionTimestamp,
			&alert.ChartLink, &alert.Viewed, &alert.Expired, &alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alert.PatternType = model.PatternType(patternType)
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// GetAlertByID retrieves a single alert by ID.
func (s *AlertService) GetAlertByID(ctx context.Context, alertID int64) (*model.Alert, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, user_id, symbol, pattern_type, confidence_score, explanation, 
		       detection_timestamp, COALESCE(chart_link, ''), viewed, expired, created_at
		FROM alerts
		WHERE id = $1
	`

	var alert model.Alert
	var patternType string
	err := s.db.QueryRowContext(ctx, query, alertID).Scan(
		&alert.ID, &alert.UserID, &alert.Symbol, &patternType,
		&alert.ConfidenceScore, &alert.Explanation, &alert.DetectionTimestamp,
		&alert.ChartLink, &alert.Viewed, &alert.Expired, &alert.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	alert.PatternType = model.PatternType(patternType)

	return &alert, nil
}

// GetUnviewedAlertCount returns the count of unviewed, non-expired alerts for a user.
func (s *AlertService) GetUnviewedAlertCount(ctx context.Context, userID int64) (int, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	query := `SELECT COUNT(*) FROM alerts WHERE user_id = $1 AND viewed = false AND expired = false`

	var count int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	return count, nil
}

// AlertFilters defines filters for querying alerts.
type AlertFilters struct {
	Symbol         string
	PatternType    string
	StartDate      time.Time
	EndDate        time.Time
	IncludeViewed  bool
	IncludeExpired bool
	Limit          int
	Offset         int
}

// ---------------------------------------------------------------------------
// Alert Lifecycle Management (Requirement 14.7)
// ---------------------------------------------------------------------------

// MarkAlertViewed marks an alert as viewed.
func (s *AlertService) MarkAlertViewed(ctx context.Context, alertID int64) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}

	query := `UPDATE alerts SET viewed = true WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, alertID)
	if err != nil {
		return fmt.Errorf("failed to mark alert viewed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert not found: %d", alertID)
	}

	return nil
}

// MarkAlertsExpired marks all unviewed alerts older than 24 hours as expired.
// Requirement 14.7: Mark alerts as expired after 24 hours if not viewed.
func (s *AlertService) MarkAlertsExpired(ctx context.Context) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	expirationCutoff := time.Now().Add(-s.config.ExpirationWindow)

	query := `
		UPDATE alerts 
		SET expired = true 
		WHERE viewed = false AND expired = false AND created_at < $1
	`

	result, err := s.db.ExecContext(ctx, query, expirationCutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to mark alerts expired: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Alert_Service] Marked %d alerts as expired", rowsAffected)
	}

	return rowsAffected, nil
}

// ---------------------------------------------------------------------------
// User Alert Preferences (Requirement 14.5)
// ---------------------------------------------------------------------------

// GetUserPreferences retrieves alert preferences for a user.
func (s *AlertService) GetUserPreferences(ctx context.Context, userID int64) (*model.AlertPreferences, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, user_id, min_confidence, pattern_types, include_symbols, exclude_symbols, updated_at
		FROM alert_preferences
		WHERE user_id = $1
	`

	var prefs model.AlertPreferences
	var patternTypesJSON, includeSymbolsJSON, excludeSymbolsJSON sql.NullString

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.MinConfidence,
		&patternTypesJSON, &includeSymbolsJSON, &excludeSymbolsJSON,
		&prefs.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Return default preferences if none exist
		return &model.AlertPreferences{
			UserID:        userID,
			MinConfidence: s.config.DefaultConfidenceThreshold,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	// Parse JSON arrays
	if patternTypesJSON.Valid && patternTypesJSON.String != "" {
		var patternStrings []string
		if err := json.Unmarshal([]byte(patternTypesJSON.String), &patternStrings); err == nil {
			for _, pt := range patternStrings {
				prefs.PatternTypes = append(prefs.PatternTypes, model.PatternType(pt))
			}
		}
	}
	if includeSymbolsJSON.Valid && includeSymbolsJSON.String != "" {
		json.Unmarshal([]byte(includeSymbolsJSON.String), &prefs.IncludeSymbols)
	}
	if excludeSymbolsJSON.Valid && excludeSymbolsJSON.String != "" {
		json.Unmarshal([]byte(excludeSymbolsJSON.String), &prefs.ExcludeSymbols)
	}

	return &prefs, nil
}

// SaveUserPreferences saves or updates alert preferences for a user.
func (s *AlertService) SaveUserPreferences(ctx context.Context, prefs *model.AlertPreferences) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}

	// Convert slices to JSON
	patternTypesJSON, _ := json.Marshal(prefs.PatternTypes)
	includeSymbolsJSON, _ := json.Marshal(prefs.IncludeSymbols)
	excludeSymbolsJSON, _ := json.Marshal(prefs.ExcludeSymbols)

	query := `
		INSERT INTO alert_preferences (user_id, min_confidence, pattern_types, include_symbols, exclude_symbols, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			min_confidence = EXCLUDED.min_confidence,
			pattern_types = EXCLUDED.pattern_types,
			include_symbols = EXCLUDED.include_symbols,
			exclude_symbols = EXCLUDED.exclude_symbols,
			updated_at = NOW()
		RETURNING id, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		prefs.UserID,
		prefs.MinConfidence,
		string(patternTypesJSON),
		string(includeSymbolsJSON),
		string(excludeSymbolsJSON),
	).Scan(&prefs.ID, &prefs.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save preferences: %w", err)
	}

	log.Printf("[Alert_Service] Saved preferences for user %d: minConfidence=%d",
		prefs.UserID, prefs.MinConfidence)

	return nil
}

// ---------------------------------------------------------------------------
// Batch Operations
// ---------------------------------------------------------------------------

// CreateAlertsForAllUsers creates alerts for all users who have the given symbol in their watchlist.
// This is used by the Monitor_Agent to broadcast alerts to relevant users.
func (s *AlertService) CreateAlertsForAllUsers(ctx context.Context, obs *model.PatternObservation) (int, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	// Find all users who have this symbol in their watchlist
	query := `
		SELECT DISTINCT w.user_id
		FROM watchlists w
		JOIN watchlist_symbols ws ON w.id = ws.watchlist_id
		WHERE ws.symbol = $1
	`

	rows, err := s.db.QueryContext(ctx, query, obs.Symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		log.Printf("[Alert_Service] No users watching symbol %s", obs.Symbol)
		return 0, nil
	}

	// Create alerts for each user
	alertsCreated := 0
	for _, userID := range userIDs {
		alert, err := s.CreateAlertFromObservation(ctx, userID, obs)
		if err != nil {
			log.Printf("[Alert_Service] Failed to create alert for user %d: %v", userID, err)
			continue
		}
		if alert != nil {
			alertsCreated++
		}
	}

	log.Printf("[Alert_Service] Created %d alerts for %d users watching %s",
		alertsCreated, len(userIDs), obs.Symbol)

	return alertsCreated, nil
}

// DeleteOldAlerts removes alerts older than the specified retention period.
func (s *AlertService) DeleteOldAlerts(ctx context.Context, retentionDays int) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM alerts WHERE created_at < $1`
	result, err := s.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old alerts: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[Alert_Service] Deleted %d alerts older than %d days", rowsAffected, retentionDays)
	}

	return rowsAffected, nil
}
