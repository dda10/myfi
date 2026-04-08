package knowledge

// ---------------------------------------------------------------------------
// Knowledge_Base — persistent store for pattern observations and outcomes
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 13.1 Persist observations with symbol, pattern type, detection date,
//          confidence score, supporting data, and price at detection
//   - 13.2 Track outcomes at 1d, 7d, 14d, 30d intervals after detection
//   - 13.3 Query historical observations for pattern success rate computation
//   - 13.4 Query interface with filters (symbol, pattern type, date range, min confidence)
//   - 13.6 Retain observations for minimum 2 years
//   - 13.7 Compute aggregate accuracy metrics per pattern type

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/infra"
)

// ObservationFilters defines query filters for pattern observations.
type ObservationFilters struct {
	Symbol        string
	PatternType   string
	StartDate     time.Time
	EndDate       time.Time
	MinConfidence int
	Limit         int
	Offset        int
}

// KnowledgeBase provides persistent storage and querying of pattern observations.
type KnowledgeBase struct {
	db     *sql.DB
	router *infra.DataSourceRouter
}

// NewKnowledgeBase creates a new Knowledge_Base service.
func NewKnowledgeBase(db *sql.DB, router *infra.DataSourceRouter) *KnowledgeBase {
	return &KnowledgeBase{
		db:     db,
		router: router,
	}
}

// ---------------------------------------------------------------------------
// StoreObservation (Requirement 13.1)
// ---------------------------------------------------------------------------

// StoreObservation persists a pattern observation with all required fields.
func (k *KnowledgeBase) StoreObservation(ctx context.Context, obs *PatternObservation) error {
	if k.db == nil {
		return fmt.Errorf("database not configured")
	}

	query := `
		INSERT INTO pattern_observations 
		(symbol, pattern_type, detection_date, confidence_score, price_at_detection, supporting_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := k.db.QueryRowContext(ctx, query,
		obs.Symbol,
		string(obs.PatternType),
		obs.DetectionDate,
		obs.ConfidenceScore,
		obs.PriceAtDetection,
		obs.SupportingData,
	).Scan(&obs.ID)

	if err != nil {
		return fmt.Errorf("failed to insert observation: %w", err)
	}

	log.Printf("[Knowledge_Base] Stored observation: %s %s confidence=%d id=%d",
		obs.Symbol, obs.PatternType, obs.ConfidenceScore, obs.ID)

	return nil
}

// ---------------------------------------------------------------------------
// UpdateOutcomes (Requirement 13.2)
// ---------------------------------------------------------------------------

// UpdateOutcomes tracks price changes at 1d, 7d, 14d, 30d intervals.
// This should be called periodically (e.g., daily) to update outcome data.
func (k *KnowledgeBase) UpdateOutcomes(ctx context.Context) error {
	if k.db == nil {
		return fmt.Errorf("database not configured")
	}

	// Get observations that need outcome updates
	observations, err := k.getObservationsNeedingOutcomeUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get observations for update: %w", err)
	}

	if len(observations) == 0 {
		log.Printf("[Knowledge_Base] No observations need outcome updates")
		return nil
	}

	log.Printf("[Knowledge_Base] Updating outcomes for %d observations", len(observations))

	updated := 0
	for _, obs := range observations {
		if err := k.updateObservationOutcome(ctx, &obs); err != nil {
			log.Printf("[Knowledge_Base] Failed to update outcome for observation %d: %v", obs.ID, err)
			continue
		}
		updated++
	}

	log.Printf("[Knowledge_Base] Updated outcomes for %d/%d observations", updated, len(observations))
	return nil
}

// getObservationsNeedingOutcomeUpdate returns observations that have missing outcome data
// and are old enough to have that outcome measured.
func (k *KnowledgeBase) getObservationsNeedingOutcomeUpdate(ctx context.Context) ([]PatternObservation, error) {
	now := time.Now()

	query := `
		SELECT id, symbol, pattern_type, detection_date, confidence_score, 
		       price_at_detection, supporting_data,
		       outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM pattern_observations
		WHERE 
			(outcome_1day IS NULL AND detection_date <= $1)
			OR (outcome_7day IS NULL AND detection_date <= $2)
			OR (outcome_14day IS NULL AND detection_date <= $3)
			OR (outcome_30day IS NULL AND detection_date <= $4)
		ORDER BY detection_date ASC
		LIMIT 100
	`

	rows, err := k.db.QueryContext(ctx, query,
		now.AddDate(0, 0, -1),
		now.AddDate(0, 0, -7),
		now.AddDate(0, 0, -14),
		now.AddDate(0, 0, -30),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []PatternObservation
	for rows.Next() {
		var obs PatternObservation
		var patternType string
		err := rows.Scan(
			&obs.ID, &obs.Symbol, &patternType, &obs.DetectionDate,
			&obs.ConfidenceScore, &obs.PriceAtDetection, &obs.SupportingData,
			&obs.Outcome1Day, &obs.Outcome7Day, &obs.Outcome14Day, &obs.Outcome30Day,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan observation: %w", err)
		}
		obs.PatternType = PatternType(patternType)
		observations = append(observations, obs)
	}

	return observations, rows.Err()
}

// updateObservationOutcome fetches current price and updates the appropriate outcome fields.
func (k *KnowledgeBase) updateObservationOutcome(ctx context.Context, obs *PatternObservation) error {
	now := time.Now()
	daysSinceDetection := int(now.Sub(obs.DetectionDate).Hours() / 24)

	// Fetch current price for the symbol
	currentPrice, err := k.fetchCurrentPrice(ctx, obs.Symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch current price for %s: %w", obs.Symbol, err)
	}

	// Calculate price change percentage
	priceChange := ((currentPrice - obs.PriceAtDetection) / obs.PriceAtDetection) * 100

	// Determine which outcome fields to update
	var updates []string
	var args []any
	argIdx := 1

	if obs.Outcome1Day == nil && daysSinceDetection >= 1 {
		updates = append(updates, fmt.Sprintf("outcome_1day = $%d", argIdx))
		args = append(args, priceChange)
		argIdx++
	}

	if obs.Outcome7Day == nil && daysSinceDetection >= 7 {
		updates = append(updates, fmt.Sprintf("outcome_7day = $%d", argIdx))
		args = append(args, priceChange)
		argIdx++
	}

	if obs.Outcome14Day == nil && daysSinceDetection >= 14 {
		updates = append(updates, fmt.Sprintf("outcome_14day = $%d", argIdx))
		args = append(args, priceChange)
		argIdx++
	}

	if obs.Outcome30Day == nil && daysSinceDetection >= 30 {
		updates = append(updates, fmt.Sprintf("outcome_30day = $%d", argIdx))
		args = append(args, priceChange)
		argIdx++
	}

	if len(updates) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE pattern_observations SET %s WHERE id = $%d",
		joinStrings(updates, ", "), argIdx)
	args = append(args, obs.ID)

	_, err = k.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update observation %d: %w", obs.ID, err)
	}

	log.Printf("[Knowledge_Base] Updated outcomes for observation %d (%s %s): price change %.2f%%",
		obs.ID, obs.Symbol, obs.PatternType, priceChange)

	return nil
}

// fetchCurrentPrice fetches the current price for a symbol via the Data_Source_Router.
func (k *KnowledgeBase) fetchCurrentPrice(ctx context.Context, symbol string) (float64, error) {
	if k.router == nil {
		return 0, fmt.Errorf("data source router not configured")
	}

	quotes, _, err := k.router.FetchRealTimeQuotes(ctx, []string{symbol})
	if err != nil {
		return 0, err
	}

	if len(quotes) == 0 {
		return 0, fmt.Errorf("no price data for symbol %s", symbol)
	}

	return quotes[0].Close, nil
}

// ---------------------------------------------------------------------------
// QueryObservations (Requirement 13.4)
// ---------------------------------------------------------------------------

// QueryObservations retrieves observations with filters.
func (k *KnowledgeBase) QueryObservations(ctx context.Context, filters ObservationFilters) ([]PatternObservation, error) {
	if k.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, symbol, pattern_type, detection_date, confidence_score, 
		       price_at_detection, supporting_data, 
		       outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM pattern_observations
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

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
		query += fmt.Sprintf(" AND detection_date >= $%d", argIdx)
		args = append(args, filters.StartDate)
		argIdx++
	}

	if !filters.EndDate.IsZero() {
		query += fmt.Sprintf(" AND detection_date <= $%d", argIdx)
		args = append(args, filters.EndDate)
		argIdx++
	}

	if filters.MinConfidence > 0 {
		query += fmt.Sprintf(" AND confidence_score >= $%d", argIdx)
		args = append(args, filters.MinConfidence)
		argIdx++
	}

	query += " ORDER BY detection_date DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filters.Limit)
		argIdx++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filters.Offset)
	}

	rows, err := k.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query observations: %w", err)
	}
	defer rows.Close()

	var observations []PatternObservation
	for rows.Next() {
		var obs PatternObservation
		var patternType string
		err := rows.Scan(
			&obs.ID, &obs.Symbol, &patternType, &obs.DetectionDate,
			&obs.ConfidenceScore, &obs.PriceAtDetection, &obs.SupportingData,
			&obs.Outcome1Day, &obs.Outcome7Day, &obs.Outcome14Day, &obs.Outcome30Day,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan observation: %w", err)
		}
		obs.PatternType = PatternType(patternType)
		observations = append(observations, obs)
	}

	return observations, rows.Err()
}

// ---------------------------------------------------------------------------
// GetAccuracyMetrics (Requirement 13.7)
// ---------------------------------------------------------------------------

// AccuracyMetrics holds aggregate accuracy metrics for a pattern type.
type AccuracyMetrics struct {
	PatternType       PatternType `json:"patternType"`
	TotalObservations int         `json:"totalObservations"`
	SuccessCount1Day  int         `json:"successCount1Day"`
	FailureCount1Day  int         `json:"failureCount1Day"`
	SuccessCount7Day  int         `json:"successCount7Day"`
	FailureCount7Day  int         `json:"failureCount7Day"`
	SuccessCount14Day int         `json:"successCount14Day"`
	FailureCount14Day int         `json:"failureCount14Day"`
	SuccessCount30Day int         `json:"successCount30Day"`
	FailureCount30Day int         `json:"failureCount30Day"`
	AvgPriceChange1D  float64     `json:"avgPriceChange1Day"`
	AvgPriceChange7D  float64     `json:"avgPriceChange7Day"`
	AvgPriceChange14D float64     `json:"avgPriceChange14Day"`
	AvgPriceChange30D float64     `json:"avgPriceChange30Day"`
	AvgConfidence     float64     `json:"avgConfidence"`
	SuccessRate7Day   float64     `json:"successRate7Day"`
	SuccessRate30Day  float64     `json:"successRate30Day"`
}

// GetAccuracyMetrics computes aggregate accuracy metrics for a pattern type.
func (k *KnowledgeBase) GetAccuracyMetrics(ctx context.Context, patternType PatternType) (*AccuracyMetrics, error) {
	if k.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN outcome_1day > 0 THEN 1 END) as success_1d,
			COUNT(CASE WHEN outcome_1day <= 0 THEN 1 END) as failure_1d,
			COUNT(CASE WHEN outcome_7day > 0 THEN 1 END) as success_7d,
			COUNT(CASE WHEN outcome_7day <= 0 THEN 1 END) as failure_7d,
			COUNT(CASE WHEN outcome_14day > 0 THEN 1 END) as success_14d,
			COUNT(CASE WHEN outcome_14day <= 0 THEN 1 END) as failure_14d,
			COUNT(CASE WHEN outcome_30day > 0 THEN 1 END) as success_30d,
			COUNT(CASE WHEN outcome_30day <= 0 THEN 1 END) as failure_30d,
			COALESCE(AVG(outcome_1day), 0) as avg_change_1d,
			COALESCE(AVG(outcome_7day), 0) as avg_change_7d,
			COALESCE(AVG(outcome_14day), 0) as avg_change_14d,
			COALESCE(AVG(outcome_30day), 0) as avg_change_30d,
			COALESCE(AVG(confidence_score), 0) as avg_confidence
		FROM pattern_observations
		WHERE pattern_type = $1
	`

	metrics := &AccuracyMetrics{
		PatternType: patternType,
	}

	err := k.db.QueryRowContext(ctx, query, string(patternType)).Scan(
		&metrics.TotalObservations,
		&metrics.SuccessCount1Day,
		&metrics.FailureCount1Day,
		&metrics.SuccessCount7Day,
		&metrics.FailureCount7Day,
		&metrics.SuccessCount14Day,
		&metrics.FailureCount14Day,
		&metrics.SuccessCount30Day,
		&metrics.FailureCount30Day,
		&metrics.AvgPriceChange1D,
		&metrics.AvgPriceChange7D,
		&metrics.AvgPriceChange14D,
		&metrics.AvgPriceChange30D,
		&metrics.AvgConfidence,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to compute accuracy metrics: %w", err)
	}

	total7Day := metrics.SuccessCount7Day + metrics.FailureCount7Day
	if total7Day > 0 {
		metrics.SuccessRate7Day = float64(metrics.SuccessCount7Day) / float64(total7Day) * 100
	}

	total30Day := metrics.SuccessCount30Day + metrics.FailureCount30Day
	if total30Day > 0 {
		metrics.SuccessRate30Day = float64(metrics.SuccessCount30Day) / float64(total30Day) * 100
	}

	return metrics, nil
}

// GetAllAccuracyMetrics returns accuracy metrics for all pattern types.
func (k *KnowledgeBase) GetAllAccuracyMetrics(ctx context.Context) ([]AccuracyMetrics, error) {
	patternTypes := []PatternType{
		PatternAccumulation,
		PatternDistribution,
		PatternBreakout,
	}

	var allMetrics []AccuracyMetrics
	for _, pt := range patternTypes {
		metrics, err := k.GetAccuracyMetrics(ctx, pt)
		if err != nil {
			log.Printf("[Knowledge_Base] Failed to get metrics for %s: %v", pt, err)
			continue
		}
		allMetrics = append(allMetrics, *metrics)
	}

	return allMetrics, nil
}

// ---------------------------------------------------------------------------
// GetHistoricalSuccessRate (Requirement 13.3)
// ---------------------------------------------------------------------------

// GetHistoricalSuccessRate computes the historical success rate for a pattern type.
func (k *KnowledgeBase) GetHistoricalSuccessRate(ctx context.Context, patternType PatternType) (float64, int, error) {
	metrics, err := k.GetAccuracyMetrics(ctx, patternType)
	if err != nil {
		return 0, 0, err
	}

	total := metrics.SuccessCount7Day + metrics.FailureCount7Day
	if total == 0 {
		return 0, 0, nil
	}

	return metrics.SuccessRate7Day, total, nil
}

// ---------------------------------------------------------------------------
// CleanupOldObservations (Requirement 13.6)
// ---------------------------------------------------------------------------

// CleanupOldObservations removes observations older than 2 years.
func (k *KnowledgeBase) CleanupOldObservations(ctx context.Context) (int64, error) {
	if k.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	cutoffDate := time.Now().AddDate(-2, 0, 0)

	query := `
		DELETE FROM pattern_observations
		WHERE detection_date < $1
	`

	result, err := k.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old observations: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		log.Printf("[Knowledge_Base] Cleaned up %d observations older than %s",
			rowsAffected, cutoffDate.Format("2006-01-02"))
	}

	return rowsAffected, nil
}

// ---------------------------------------------------------------------------
// ArchiveOldObservations (Requirement 26.5) — archive >90 days to S3
// ---------------------------------------------------------------------------

// ArchiveOldObservations archives observations older than 90 days to S3 storage
// and removes them from the database. Returns the number of archived records.
func (k *KnowledgeBase) ArchiveOldObservations(ctx context.Context, storage interface {
	PutParquet(ctx context.Context, key string, data []byte) error
}) (int64, error) {
	if k.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	cutoffDate := time.Now().AddDate(0, 0, -90)

	// Fetch observations to archive
	query := `
		SELECT id, symbol, pattern_type, detection_date, confidence_score,
		       price_at_detection, supporting_data,
		       outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM pattern_observations
		WHERE detection_date < $1
		ORDER BY detection_date ASC
	`

	rows, err := k.db.QueryContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to query observations for archival: %w", err)
	}
	defer rows.Close()

	var observations []PatternObservation
	for rows.Next() {
		var obs PatternObservation
		var patternType string
		err := rows.Scan(
			&obs.ID, &obs.Symbol, &patternType, &obs.DetectionDate,
			&obs.ConfidenceScore, &obs.PriceAtDetection, &obs.SupportingData,
			&obs.Outcome1Day, &obs.Outcome7Day, &obs.Outcome14Day, &obs.Outcome30Day,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to scan observation for archival: %w", err)
		}
		obs.PatternType = PatternType(patternType)
		observations = append(observations, obs)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating observations for archival: %w", err)
	}

	if len(observations) == 0 {
		return 0, nil
	}

	// Serialize to simple CSV-like format for S3 storage
	archiveKey := fmt.Sprintf("knowledge/archive/%s.csv", time.Now().Format("20060102_150405"))
	var archiveData []byte
	header := "id,symbol,pattern_type,detection_date,confidence_score,price_at_detection\n"
	archiveData = append(archiveData, []byte(header)...)
	for _, obs := range observations {
		line := fmt.Sprintf("%d,%s,%s,%s,%d,%.2f\n",
			obs.ID, obs.Symbol, obs.PatternType,
			obs.DetectionDate.Format("2006-01-02"), obs.ConfidenceScore, obs.PriceAtDetection)
		archiveData = append(archiveData, []byte(line)...)
	}

	if storage != nil {
		if err := storage.PutParquet(ctx, archiveKey, archiveData); err != nil {
			return 0, fmt.Errorf("failed to upload archive to S3: %w", err)
		}
	}

	// Delete archived observations from DB
	deleteQuery := `DELETE FROM pattern_observations WHERE detection_date < $1`
	result, err := k.db.ExecContext(ctx, deleteQuery, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete archived observations: %w", err)
	}

	archived, _ := result.RowsAffected()
	log.Printf("[Knowledge_Base] Archived %d observations older than 90 days to %s", archived, archiveKey)
	return archived, nil
}

// ---------------------------------------------------------------------------
// QuerySimilarPatterns (Requirement 26.3)
// ---------------------------------------------------------------------------

// QuerySimilarPatterns finds observations with the same pattern type and
// minimum confidence score, useful for pattern matching.
func (k *KnowledgeBase) QuerySimilarPatterns(ctx context.Context, patternType string, minConfidence float64) ([]PatternObservation, error) {
	return k.QueryObservations(ctx, ObservationFilters{
		PatternType:   patternType,
		MinConfidence: int(minConfidence),
		Limit:         50,
	})
}

// ---------------------------------------------------------------------------
// GetRecentObservationsForSymbol
// ---------------------------------------------------------------------------

// GetRecentObservationsForSymbol returns recent observations for a symbol.
func (k *KnowledgeBase) GetRecentObservationsForSymbol(ctx context.Context, symbol string, days int) ([]PatternObservation, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	return k.QueryObservations(ctx, ObservationFilters{
		Symbol:    symbol,
		StartDate: startDate,
		Limit:     10,
	})
}

// ---------------------------------------------------------------------------
// GetObservationByID
// ---------------------------------------------------------------------------

// GetObservationByID retrieves a single observation by ID.
func (k *KnowledgeBase) GetObservationByID(ctx context.Context, id int64) (*PatternObservation, error) {
	if k.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, symbol, pattern_type, detection_date, confidence_score, 
		       price_at_detection, supporting_data, 
		       outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM pattern_observations
		WHERE id = $1
	`

	var obs PatternObservation
	var patternType string
	err := k.db.QueryRowContext(ctx, query, id).Scan(
		&obs.ID, &obs.Symbol, &patternType, &obs.DetectionDate,
		&obs.ConfidenceScore, &obs.PriceAtDetection, &obs.SupportingData,
		&obs.Outcome1Day, &obs.Outcome7Day, &obs.Outcome14Day, &obs.Outcome30Day,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	obs.PatternType = PatternType(patternType)
	return &obs, nil
}

// ---------------------------------------------------------------------------
// GetObservationCount
// ---------------------------------------------------------------------------

// GetObservationCount returns the total count of observations matching filters.
func (k *KnowledgeBase) GetObservationCount(ctx context.Context, filters ObservationFilters) (int, error) {
	if k.db == nil {
		return 0, fmt.Errorf("database not configured")
	}

	query := `SELECT COUNT(*) FROM pattern_observations WHERE 1=1`
	args := []any{}
	argIdx := 1

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
		query += fmt.Sprintf(" AND detection_date >= $%d", argIdx)
		args = append(args, filters.StartDate)
		argIdx++
	}

	if !filters.EndDate.IsZero() {
		query += fmt.Sprintf(" AND detection_date <= $%d", argIdx)
		args = append(args, filters.EndDate)
		argIdx++
	}

	if filters.MinConfidence > 0 {
		query += fmt.Sprintf(" AND confidence_score >= $%d", argIdx)
		args = append(args, filters.MinConfidence)
	}

	var count int
	err := k.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count observations: %w", err)
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
