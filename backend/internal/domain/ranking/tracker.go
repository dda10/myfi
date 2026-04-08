package ranking

// ---------------------------------------------------------------------------
// Recommendation_Audit_Log — persistent store for AI recommendations
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 35.1 Persist every recommendation with full inputs/outputs
//   - 35.2 Track outcomes at 1d, 7d, 14d, 30d intervals
//   - 35.4 Compute recommendation accuracy metrics
//   - 35.5 Feed outcomes back to Knowledge_Base
//   - 35.6 Retain audit records for 2 years

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"myfi-backend/internal/domain/knowledge"
	"myfi-backend/internal/domain/market"
)

// RecommendationTracker logs AI recommendations and tracks their outcomes.
// Implements the Recommendation_Audit_Log from the design spec.
type RecommendationTracker struct {
	mu            sync.RWMutex
	records       []RecommendationRecord // In-memory cache for recent records
	nextID        int64
	priceService  *market.PriceService
	db            *sql.DB
	knowledgeBase *knowledge.KnowledgeBase
	cacheMaxSize  int // Maximum number of records to keep in memory
}

// NewRecommendationTracker creates a new tracker instance.
func NewRecommendationTracker(priceService *market.PriceService) *RecommendationTracker {
	return &RecommendationTracker{
		records:      make([]RecommendationRecord, 0),
		nextID:       1,
		priceService: priceService,
		cacheMaxSize: 1000,
	}
}

// NewRecommendationTrackerWithDB creates a tracker with database persistence.
func NewRecommendationTrackerWithDB(priceService *market.PriceService, db *sql.DB, kb *knowledge.KnowledgeBase) *RecommendationTracker {
	tracker := &RecommendationTracker{
		records:       make([]RecommendationRecord, 0),
		nextID:        1,
		priceService:  priceService,
		db:            db,
		knowledgeBase: kb,
		cacheMaxSize:  1000,
	}

	// Initialize nextID from database if available
	if db != nil {
		var maxID sql.NullInt64
		err := db.QueryRow("SELECT MAX(id) FROM recommendation_audit_log").Scan(&maxID)
		if err == nil && maxID.Valid {
			tracker.nextID = maxID.Int64 + 1
		}
	}

	return tracker
}

// SetDB sets the database connection for persistence.
func (t *RecommendationTracker) SetDB(db *sql.DB) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.db = db
}

// SetKnowledgeBase sets the knowledge base for outcome feedback.
func (t *RecommendationTracker) SetKnowledgeBase(kb *knowledge.KnowledgeBase) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.knowledgeBase = kb
}

// RecommendationAuditRecord represents a database record for the audit log.
type RecommendationAuditRecord struct {
	ID                   int64     `json:"id"`
	UserID               int64     `json:"userId"`
	Timestamp            time.Time `json:"timestamp"`
	UserQuery            string    `json:"userQuery"`
	SubAgentInputs       string    `json:"subAgentInputs"`
	RecommendationOutput string    `json:"recommendationOutput"`
	SymbolsInvolved      string    `json:"symbolsInvolved"`
	RecommendedActions   string    `json:"recommendedActions"`
	Outcome1Day          *string   `json:"outcome1Day,omitempty"`
	Outcome7Day          *string   `json:"outcome7Day,omitempty"`
	Outcome14Day         *string   `json:"outcome14Day,omitempty"`
	Outcome30Day         *string   `json:"outcome30Day,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
}

// LogRecommendation records a new AI recommendation with current price.
// Persists to database if configured (Requirement 35.1).
func (t *RecommendationTracker) LogRecommendation(ctx context.Context, rec AssetRecommendation, confidence int) (int64, error) {
	// Fetch current price for the symbol
	var priceAtSignal float64
	if t.priceService != nil {
		quotes, err := t.priceService.GetQuotes(ctx, []string{rec.Symbol})
		if err == nil && len(quotes) > 0 {
			priceAtSignal = quotes[0].Price
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	record := RecommendationRecord{
		ID:              t.nextID,
		Symbol:          rec.Symbol,
		Action:          RecommendationAction(rec.Action),
		PositionSize:    rec.PositionSize,
		RiskAssessment:  rec.RiskAssessment,
		ConfidenceScore: confidence,
		Reasoning:       rec.Reasoning,
		PriceAtSignal:   priceAtSignal,
		CreatedAt:       time.Now(),
	}

	// Persist to database if available
	if t.db != nil {
		id, err := t.persistRecommendation(ctx, &record)
		if err != nil {
			log.Printf("[RecommendationTracker] Failed to persist to DB: %v", err)
		} else {
			record.ID = id
			t.nextID = id + 1
		}
	}

	// Add to in-memory cache
	t.records = append(t.records, record)
	if t.nextID <= record.ID {
		t.nextID = record.ID + 1
	}

	// Trim cache if too large
	if len(t.records) > t.cacheMaxSize {
		t.records = t.records[len(t.records)-t.cacheMaxSize:]
	}

	log.Printf("[RecommendationTracker] Logged recommendation: %s %s confidence=%d id=%d",
		rec.Symbol, rec.Action, confidence, record.ID)

	return record.ID, nil
}

// persistRecommendation saves a recommendation to the database.
func (t *RecommendationTracker) persistRecommendation(ctx context.Context, rec *RecommendationRecord) (int64, error) {
	subAgentInputs := map[string]interface{}{
		"priceAtSignal": rec.PriceAtSignal,
		"confidence":    rec.ConfidenceScore,
	}
	subAgentInputsJSON, _ := json.Marshal(subAgentInputs)

	recOutput := map[string]interface{}{
		"action":         rec.Action,
		"positionSize":   rec.PositionSize,
		"riskAssessment": rec.RiskAssessment,
		"reasoning":      rec.Reasoning,
	}
	recOutputJSON, _ := json.Marshal(recOutput)

	query := `
		INSERT INTO recommendation_audit_log 
		(user_id, timestamp, user_query, sub_agent_inputs, recommendation_output, 
		 symbols_involved, recommended_actions)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := t.db.QueryRowContext(ctx, query,
		1,
		rec.CreatedAt,
		rec.Reasoning,
		string(subAgentInputsJSON),
		string(recOutputJSON),
		rec.Symbol,
		string(rec.Action),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert recommendation: %w", err)
	}

	return id, nil
}

// UpdateOutcomes fetches current prices and updates outcome fields for records
// that have reached their measurement intervals (Requirement 35.2).
func (t *RecommendationTracker) UpdateOutcomes(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	updatedCount := 0

	for i := range t.records {
		rec := &t.records[i]
		if rec.PriceAtSignal == 0 {
			continue
		}

		age := now.Sub(rec.CreatedAt)
		updated := false

		if age >= 24*time.Hour && rec.Price1Day == nil {
			if price := t.fetchPrice(ctx, rec.Symbol); price > 0 {
				rec.Price1Day = &price
				ret := (price - rec.PriceAtSignal) / rec.PriceAtSignal
				rec.Return1Day = &ret
				updated = true
			}
		}

		if age >= 7*24*time.Hour && rec.Price7Day == nil {
			if price := t.fetchPrice(ctx, rec.Symbol); price > 0 {
				rec.Price7Day = &price
				ret := (price - rec.PriceAtSignal) / rec.PriceAtSignal
				rec.Return7Day = &ret
				updated = true
			}
		}

		if age >= 14*24*time.Hour && rec.Price14Day == nil {
			if price := t.fetchPrice(ctx, rec.Symbol); price > 0 {
				rec.Price14Day = &price
				ret := (price - rec.PriceAtSignal) / rec.PriceAtSignal
				rec.Return14Day = &ret
				updated = true
			}
		}

		if age >= 30*24*time.Hour && rec.Price30Day == nil {
			if price := t.fetchPrice(ctx, rec.Symbol); price > 0 {
				rec.Price30Day = &price
				ret := (price - rec.PriceAtSignal) / rec.PriceAtSignal
				rec.Return30Day = &ret
				updated = true
			}
		}

		if updated {
			updatedCount++
			if t.db != nil {
				if err := t.updateOutcomeInDB(ctx, rec); err != nil {
					log.Printf("[RecommendationTracker] Failed to update outcome in DB for id=%d: %v", rec.ID, err)
				}
			}
			t.feedOutcomeToKnowledgeBase(ctx, rec)
		}
	}

	if updatedCount > 0 {
		log.Printf("[RecommendationTracker] Updated outcomes for %d recommendations", updatedCount)
	}

	if t.db != nil {
		if err := t.updateOutcomesFromDB(ctx); err != nil {
			log.Printf("[RecommendationTracker] Failed to update DB outcomes: %v", err)
		}
	}

	return nil
}

// updateOutcomeInDB persists outcome updates to the database.
func (t *RecommendationTracker) updateOutcomeInDB(ctx context.Context, rec *RecommendationRecord) error {
	var outcome1Day, outcome7Day, outcome14Day, outcome30Day *string

	if rec.Return1Day != nil {
		s := fmt.Sprintf(`{"price":%.2f,"return":%.4f}`, *rec.Price1Day, *rec.Return1Day)
		outcome1Day = &s
	}
	if rec.Return7Day != nil {
		s := fmt.Sprintf(`{"price":%.2f,"return":%.4f}`, *rec.Price7Day, *rec.Return7Day)
		outcome7Day = &s
	}
	if rec.Return14Day != nil {
		s := fmt.Sprintf(`{"price":%.2f,"return":%.4f}`, *rec.Price14Day, *rec.Return14Day)
		outcome14Day = &s
	}
	if rec.Return30Day != nil {
		s := fmt.Sprintf(`{"price":%.2f,"return":%.4f}`, *rec.Price30Day, *rec.Return30Day)
		outcome30Day = &s
	}

	query := `
		UPDATE recommendation_audit_log 
		SET outcome_1day = COALESCE($1, outcome_1day),
		    outcome_7day = COALESCE($2, outcome_7day),
		    outcome_14day = COALESCE($3, outcome_14day),
		    outcome_30day = COALESCE($4, outcome_30day)
		WHERE id = $5
	`

	_, err := t.db.ExecContext(ctx, query, outcome1Day, outcome7Day, outcome14Day, outcome30Day, rec.ID)
	return err
}

// updateOutcomesFromDB updates outcomes for records in the database that need updating.
func (t *RecommendationTracker) updateOutcomesFromDB(ctx context.Context) error {
	now := time.Now()

	query := `
		SELECT id, symbols_involved, timestamp, sub_agent_inputs
		FROM recommendation_audit_log
		WHERE 
			(outcome_1day IS NULL AND timestamp <= $1)
			OR (outcome_7day IS NULL AND timestamp <= $2)
			OR (outcome_14day IS NULL AND timestamp <= $3)
			OR (outcome_30day IS NULL AND timestamp <= $4)
		ORDER BY timestamp ASC
		LIMIT 100
	`

	rows, err := t.db.QueryContext(ctx, query,
		now.Add(-24*time.Hour),
		now.Add(-7*24*time.Hour),
		now.Add(-14*24*time.Hour),
		now.Add(-30*24*time.Hour),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var symbol string
		var timestamp time.Time
		var subAgentInputsJSON string

		if err := rows.Scan(&id, &symbol, &timestamp, &subAgentInputsJSON); err != nil {
			continue
		}

		var inputs map[string]interface{}
		if err := json.Unmarshal([]byte(subAgentInputsJSON), &inputs); err != nil {
			continue
		}

		priceAtSignal, ok := inputs["priceAtSignal"].(float64)
		if !ok || priceAtSignal == 0 {
			continue
		}

		currentPrice := t.fetchPrice(ctx, symbol)
		if currentPrice == 0 {
			continue
		}

		age := now.Sub(timestamp)
		ret := (currentPrice - priceAtSignal) / priceAtSignal

		t.updateDBOutcomeField(ctx, id, age, currentPrice, ret)
	}

	return rows.Err()
}

// updateDBOutcomeField updates a specific outcome field in the database.
func (t *RecommendationTracker) updateDBOutcomeField(ctx context.Context, id int64, age time.Duration, price, ret float64) {
	outcomeJSON := fmt.Sprintf(`{"price":%.2f,"return":%.4f}`, price, ret)

	var query string
	if age >= 30*24*time.Hour {
		query = `UPDATE recommendation_audit_log SET outcome_30day = $1 WHERE id = $2 AND outcome_30day IS NULL`
	} else if age >= 14*24*time.Hour {
		query = `UPDATE recommendation_audit_log SET outcome_14day = $1 WHERE id = $2 AND outcome_14day IS NULL`
	} else if age >= 7*24*time.Hour {
		query = `UPDATE recommendation_audit_log SET outcome_7day = $1 WHERE id = $2 AND outcome_7day IS NULL`
	} else if age >= 24*time.Hour {
		query = `UPDATE recommendation_audit_log SET outcome_1day = $1 WHERE id = $2 AND outcome_1day IS NULL`
	} else {
		return
	}

	if _, err := t.db.ExecContext(ctx, query, outcomeJSON, id); err != nil {
		log.Printf("[RecommendationTracker] Failed to update outcome field for id=%d: %v", id, err)
	}
}

// feedOutcomeToKnowledgeBase sends recommendation outcomes to the Knowledge_Base
// to improve future pattern detection (Requirement 35.5).
func (t *RecommendationTracker) feedOutcomeToKnowledgeBase(ctx context.Context, rec *RecommendationRecord) {
	if t.knowledgeBase == nil {
		return
	}

	if rec.Return7Day == nil {
		return
	}

	obs := &knowledge.PatternObservation{
		Symbol:           rec.Symbol,
		PatternType:      knowledge.PatternType("recommendation_" + string(rec.Action)),
		DetectionDate:    rec.CreatedAt,
		ConfidenceScore:  rec.ConfidenceScore,
		PriceAtDetection: rec.PriceAtSignal,
		SupportingData:   fmt.Sprintf(`{"action":"%s","reasoning":"%s"}`, rec.Action, rec.Reasoning),
	}

	if rec.Return1Day != nil {
		outcome := *rec.Return1Day * 100
		obs.Outcome1Day = &outcome
	}
	if rec.Return7Day != nil {
		outcome := *rec.Return7Day * 100
		obs.Outcome7Day = &outcome
	}
	if rec.Return14Day != nil {
		outcome := *rec.Return14Day * 100
		obs.Outcome14Day = &outcome
	}
	if rec.Return30Day != nil {
		outcome := *rec.Return30Day * 100
		obs.Outcome30Day = &outcome
	}

	if err := t.knowledgeBase.StoreObservation(ctx, obs); err != nil {
		log.Printf("[RecommendationTracker] Failed to feed outcome to Knowledge_Base: %v", err)
	} else {
		log.Printf("[RecommendationTracker] Fed outcome to Knowledge_Base: %s %s return7d=%.2f%%",
			rec.Symbol, rec.Action, *rec.Return7Day*100)
	}
}

func (t *RecommendationTracker) fetchPrice(ctx context.Context, symbol string) float64 {
	if t.priceService == nil {
		return 0
	}
	quotes, err := t.priceService.GetQuotes(ctx, []string{symbol})
	if err != nil || len(quotes) == 0 {
		return 0
	}
	return quotes[0].Price
}

// CleanupOldRecords removes recommendation records older than 2 years (Requirement 35.6).
func (t *RecommendationTracker) CleanupOldRecords(ctx context.Context) (int64, error) {
	cutoffDate := time.Now().AddDate(-2, 0, 0)

	t.mu.Lock()
	newRecords := make([]RecommendationRecord, 0, len(t.records))
	for _, rec := range t.records {
		if rec.CreatedAt.After(cutoffDate) {
			newRecords = append(newRecords, rec)
		}
	}
	removedFromCache := len(t.records) - len(newRecords)
	t.records = newRecords
	t.mu.Unlock()

	var removedFromDB int64
	if t.db != nil {
		query := `DELETE FROM recommendation_audit_log WHERE timestamp < $1`
		result, err := t.db.ExecContext(ctx, query, cutoffDate)
		if err != nil {
			return int64(removedFromCache), fmt.Errorf("failed to cleanup old records from DB: %w", err)
		}
		removedFromDB, _ = result.RowsAffected()
	}

	totalRemoved := int64(removedFromCache) + removedFromDB
	if totalRemoved > 0 {
		log.Printf("[RecommendationTracker] Cleaned up %d records older than %s (cache: %d, db: %d)",
			totalRemoved, cutoffDate.Format("2006-01-02"), removedFromCache, removedFromDB)
	}

	return totalRemoved, nil
}

// GetRecommendations returns recommendations matching the filter criteria.
func (t *RecommendationTracker) GetRecommendations(filter RecommendationFilter) []RecommendationRecord {
	if t.db != nil {
		records, err := t.getRecommendationsFromDB(context.Background(), filter)
		if err == nil {
			return records
		}
		log.Printf("[RecommendationTracker] DB query failed, falling back to cache: %v", err)
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	var results []RecommendationRecord

	for _, rec := range t.records {
		if filter.Symbol != "" && rec.Symbol != filter.Symbol {
			continue
		}
		if filter.Action != "" && rec.Action != filter.Action {
			continue
		}
		if filter.MinConfidence > 0 && rec.ConfidenceScore < filter.MinConfidence {
			continue
		}
		if filter.StartDate != nil && rec.CreatedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && rec.CreatedAt.After(*filter.EndDate) {
			continue
		}
		results = append(results, rec)
	}

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[len(results)-filter.Limit:]
	}

	return results
}

// getRecommendationsFromDB queries recommendations from the database.
func (t *RecommendationTracker) getRecommendationsFromDB(ctx context.Context, filter RecommendationFilter) ([]RecommendationRecord, error) {
	query := `
		SELECT id, symbols_involved, recommended_actions, timestamp, sub_agent_inputs, 
		       recommendation_output, outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM recommendation_audit_log
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if filter.Symbol != "" {
		query += fmt.Sprintf(" AND symbols_involved = $%d", argIdx)
		args = append(args, filter.Symbol)
		argIdx++
	}

	if filter.Action != "" {
		query += fmt.Sprintf(" AND recommended_actions = $%d", argIdx)
		args = append(args, string(filter.Action))
		argIdx++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, *filter.EndDate)
		argIdx++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
	}

	rows, err := t.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RecommendationRecord
	for rows.Next() {
		var rec RecommendationRecord
		var subAgentInputsJSON, recOutputJSON string
		var outcome1Day, outcome7Day, outcome14Day, outcome30Day sql.NullString
		var action string

		err := rows.Scan(
			&rec.ID, &rec.Symbol, &action, &rec.CreatedAt,
			&subAgentInputsJSON, &recOutputJSON,
			&outcome1Day, &outcome7Day, &outcome14Day, &outcome30Day,
		)
		if err != nil {
			continue
		}

		rec.Action = RecommendationAction(action)

		var inputs map[string]interface{}
		if err := json.Unmarshal([]byte(subAgentInputsJSON), &inputs); err == nil {
			if price, ok := inputs["priceAtSignal"].(float64); ok {
				rec.PriceAtSignal = price
			}
			if conf, ok := inputs["confidence"].(float64); ok {
				rec.ConfidenceScore = int(conf)
			}
		}

		var output map[string]interface{}
		if err := json.Unmarshal([]byte(recOutputJSON), &output); err == nil {
			if ps, ok := output["positionSize"].(float64); ok {
				rec.PositionSize = ps
			}
			if ra, ok := output["riskAssessment"].(string); ok {
				rec.RiskAssessment = ra
			}
			if r, ok := output["reasoning"].(string); ok {
				rec.Reasoning = r
			}
		}

		parseOutcome := func(s sql.NullString) (*float64, *float64) {
			if !s.Valid {
				return nil, nil
			}
			var o map[string]float64
			if err := json.Unmarshal([]byte(s.String), &o); err == nil {
				price := o["price"]
				ret := o["return"]
				return &price, &ret
			}
			return nil, nil
		}

		rec.Price1Day, rec.Return1Day = parseOutcome(outcome1Day)
		rec.Price7Day, rec.Return7Day = parseOutcome(outcome7Day)
		rec.Price14Day, rec.Return14Day = parseOutcome(outcome14Day)
		rec.Price30Day, rec.Return30Day = parseOutcome(outcome30Day)

		records = append(records, rec)
	}

	return records, rows.Err()
}

// GetAccuracyByAction computes accuracy metrics for a specific action type (Requirement 35.4).
func (t *RecommendationTracker) GetAccuracyByAction(action RecommendationAction) RecommendationAccuracy {
	t.mu.RLock()
	defer t.mu.RUnlock()

	acc := RecommendationAccuracy{Action: action}

	var (
		sum1Day, sum7Day, sum14Day, sum30Day         float64
		count1Day, count7Day, count14Day, count30Day int
		totalConf                                    int
		highConfWins, highConfTotal                  int
		medConfWins, medConfTotal                    int
		lowConfWins, lowConfTotal                    int
	)

	for _, rec := range t.records {
		if rec.Action != action {
			continue
		}
		acc.TotalCount++
		totalConf += rec.ConfidenceScore

		isWinFunc := func(ret float64) bool {
			switch action {
			case ActionBuy:
				return ret > 0
			case ActionSell:
				return ret < 0
			default:
				return false
			}
		}

		if rec.Return1Day != nil {
			count1Day++
			sum1Day += *rec.Return1Day
			if isWinFunc(*rec.Return1Day) {
				acc.WinCount1Day++
			}
		}
		if rec.Return7Day != nil {
			count7Day++
			sum7Day += *rec.Return7Day
			if isWinFunc(*rec.Return7Day) {
				acc.WinCount7Day++
				if rec.ConfidenceScore >= 70 {
					highConfWins++
				} else if rec.ConfidenceScore >= 40 {
					medConfWins++
				} else {
					lowConfWins++
				}
			}
			if rec.ConfidenceScore >= 70 {
				highConfTotal++
			} else if rec.ConfidenceScore >= 40 {
				medConfTotal++
			} else {
				lowConfTotal++
			}
		}
		if rec.Return14Day != nil {
			count14Day++
			sum14Day += *rec.Return14Day
			if isWinFunc(*rec.Return14Day) {
				acc.WinCount14Day++
			}
		}
		if rec.Return30Day != nil {
			count30Day++
			sum30Day += *rec.Return30Day
			if isWinFunc(*rec.Return30Day) {
				acc.WinCount30Day++
			}
		}
	}

	if count1Day > 0 {
		acc.WinRate1Day = float64(acc.WinCount1Day) / float64(count1Day)
		acc.AvgReturn1Day = sum1Day / float64(count1Day)
	}
	if count7Day > 0 {
		acc.WinRate7Day = float64(acc.WinCount7Day) / float64(count7Day)
		acc.AvgReturn7Day = sum7Day / float64(count7Day)
	}
	if count14Day > 0 {
		acc.WinRate14Day = float64(acc.WinCount14Day) / float64(count14Day)
		acc.AvgReturn14Day = sum14Day / float64(count14Day)
	}
	if count30Day > 0 {
		acc.WinRate30Day = float64(acc.WinCount30Day) / float64(count30Day)
		acc.AvgReturn30Day = sum30Day / float64(count30Day)
	}
	if acc.TotalCount > 0 {
		acc.AvgConfidence = float64(totalConf) / float64(acc.TotalCount)
	}
	if highConfTotal > 0 {
		acc.HighConfWinRate = float64(highConfWins) / float64(highConfTotal)
	}
	if medConfTotal > 0 {
		acc.MedConfWinRate = float64(medConfWins) / float64(medConfTotal)
	}
	if lowConfTotal > 0 {
		acc.LowConfWinRate = float64(lowConfWins) / float64(lowConfTotal)
	}

	return acc
}

// GetSummary returns an overall summary of recommendation performance.
func (t *RecommendationTracker) GetSummary() RecommendationSummary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	summary := RecommendationSummary{
		TotalRecommendations: len(t.records),
	}

	for _, action := range []RecommendationAction{ActionBuy, ActionSell, ActionHold} {
		acc := t.GetAccuracyByAction(action)
		if acc.TotalCount > 0 {
			summary.ByAction = append(summary.ByAction, acc)
		}
	}

	symbolReturns := make(map[string][]float64)
	var totalReturn7Day float64
	var count7Day int
	var wins7Day int

	for _, rec := range t.records {
		if rec.Return7Day != nil {
			totalReturn7Day += *rec.Return7Day
			count7Day++
			if rec.Action == ActionBuy && *rec.Return7Day > 0 {
				wins7Day++
			} else if rec.Action == ActionSell && *rec.Return7Day < 0 {
				wins7Day++
			}
			symbolReturns[rec.Symbol] = append(symbolReturns[rec.Symbol], *rec.Return7Day)
		}
	}

	if count7Day > 0 {
		summary.OverallWinRate7Day = float64(wins7Day) / float64(count7Day)
		summary.OverallAvgReturn7Day = totalReturn7Day / float64(count7Day)
	}

	var bestSymbol, worstSymbol string
	var bestAvg, worstAvg float64
	first := true

	for symbol, returns := range symbolReturns {
		if len(returns) == 0 {
			continue
		}
		var sum float64
		for _, r := range returns {
			sum += r
		}
		avg := sum / float64(len(returns))

		if first {
			bestSymbol, worstSymbol = symbol, symbol
			bestAvg, worstAvg = avg, avg
			first = false
		} else {
			if avg > bestAvg {
				bestAvg = avg
				bestSymbol = symbol
			}
			if avg < worstAvg {
				worstAvg = avg
				worstSymbol = symbol
			}
		}
	}

	summary.BestPerformingSymbol = bestSymbol
	summary.WorstPerformingSymbol = worstSymbol

	return summary
}

// GetRecordByID returns a specific recommendation by ID.
func (t *RecommendationTracker) GetRecordByID(id int64) *RecommendationRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for i := range t.records {
		if t.records[i].ID == id {
			return &t.records[i]
		}
	}
	return nil
}

// GetRecordCount returns the total number of records (in-memory + database).
func (t *RecommendationTracker) GetRecordCount(ctx context.Context) int {
	t.mu.RLock()
	cacheCount := len(t.records)
	t.mu.RUnlock()

	if t.db == nil {
		return cacheCount
	}

	var dbCount int
	err := t.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recommendation_audit_log").Scan(&dbCount)
	if err != nil {
		return cacheCount
	}

	return dbCount
}

// SaveToDatabase persists all in-memory records to the database (Requirement 35.1).
func (t *RecommendationTracker) SaveToDatabase(ctx context.Context) error {
	if t.db == nil {
		return fmt.Errorf("database not configured")
	}

	t.mu.RLock()
	records := make([]RecommendationRecord, len(t.records))
	copy(records, t.records)
	t.mu.RUnlock()

	savedCount := 0
	for i := range records {
		rec := &records[i]

		var exists bool
		err := t.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM recommendation_audit_log WHERE id = $1)",
			rec.ID,
		).Scan(&exists)

		if err != nil {
			log.Printf("[RecommendationTracker] Failed to check existence for id=%d: %v", rec.ID, err)
			continue
		}

		if exists {
			if err := t.updateOutcomeInDB(ctx, rec); err != nil {
				log.Printf("[RecommendationTracker] Failed to update record id=%d: %v", rec.ID, err)
			}
			continue
		}

		_, err = t.persistRecommendation(ctx, rec)
		if err != nil {
			log.Printf("[RecommendationTracker] Failed to save record id=%d: %v", rec.ID, err)
			continue
		}
		savedCount++
	}

	log.Printf("[RecommendationTracker] SaveToDatabase: saved %d new records, total %d in cache",
		savedCount, len(records))

	return nil
}

// LoadFromDatabase restores recommendation records from the database into memory.
func (t *RecommendationTracker) LoadFromDatabase(ctx context.Context) error {
	if t.db == nil {
		return fmt.Errorf("database not configured")
	}

	filter := RecommendationFilter{
		Limit: t.cacheMaxSize,
	}

	records, err := t.getRecommendationsFromDB(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to load recommendations from database: %w", err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	existingIDs := make(map[int64]bool)
	for _, rec := range t.records {
		existingIDs[rec.ID] = true
	}

	loadedCount := 0
	for _, rec := range records {
		if !existingIDs[rec.ID] {
			t.records = append(t.records, rec)
			existingIDs[rec.ID] = true
			loadedCount++
		}
	}

	for _, rec := range t.records {
		if rec.ID >= t.nextID {
			t.nextID = rec.ID + 1
		}
	}

	if len(t.records) > t.cacheMaxSize {
		t.records = t.records[len(t.records)-t.cacheMaxSize:]
	}

	log.Printf("[RecommendationTracker] LoadFromDatabase: loaded %d records, total %d in cache, nextID=%d",
		loadedCount, len(t.records), t.nextID)

	return nil
}

// FeedbackToKnowledgeBase explicitly feeds all recommendation outcomes to the Knowledge_Base.
func (t *RecommendationTracker) FeedbackToKnowledgeBase(ctx context.Context) (int, error) {
	if t.knowledgeBase == nil {
		return 0, fmt.Errorf("knowledge base not configured")
	}

	t.mu.RLock()
	records := make([]RecommendationRecord, len(t.records))
	copy(records, t.records)
	t.mu.RUnlock()

	fedCount := 0
	for i := range records {
		rec := &records[i]

		if rec.Return7Day == nil {
			continue
		}

		t.feedOutcomeToKnowledgeBase(ctx, rec)
		fedCount++
	}

	log.Printf("[RecommendationTracker] FeedbackToKnowledgeBase: fed %d records to Knowledge_Base", fedCount)

	return fedCount, nil
}

// GetAccuracyFromDB computes accuracy metrics from database records (Requirement 35.4).
func (t *RecommendationTracker) GetAccuracyFromDB(ctx context.Context, action RecommendationAction) (RecommendationAccuracy, error) {
	if t.db == nil {
		return t.GetAccuracyByAction(action), nil
	}

	acc := RecommendationAccuracy{Action: action}

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(outcome_1day) as count_1day,
			COUNT(outcome_7day) as count_7day,
			COUNT(outcome_14day) as count_14day,
			COUNT(outcome_30day) as count_30day
		FROM recommendation_audit_log
		WHERE recommended_actions = $1
	`

	var total, count1Day, count7Day, count14Day, count30Day int
	err := t.db.QueryRowContext(ctx, query, string(action)).Scan(
		&total, &count1Day, &count7Day, &count14Day, &count30Day,
	)
	if err != nil {
		return acc, fmt.Errorf("failed to query accuracy metrics: %w", err)
	}

	acc.TotalCount = total

	filter := RecommendationFilter{
		Action: action,
		Limit:  10000,
	}

	records, err := t.getRecommendationsFromDB(ctx, filter)
	if err != nil {
		return acc, fmt.Errorf("failed to load records for accuracy: %w", err)
	}

	var (
		sum1Day, sum7Day, sum14Day, sum30Day     float64
		wins1Day, wins7Day, wins14Day, wins30Day int
		totalConf                                int
		highConfWins, highConfTotal              int
		medConfWins, medConfTotal                int
		lowConfWins, lowConfTotal                int
	)

	isWinFunc := func(ret float64) bool {
		switch action {
		case ActionBuy:
			return ret > 0
		case ActionSell:
			return ret < 0
		default:
			return false
		}
	}

	for _, rec := range records {
		totalConf += rec.ConfidenceScore

		if rec.Return1Day != nil {
			sum1Day += *rec.Return1Day
			if isWinFunc(*rec.Return1Day) {
				wins1Day++
			}
		}
		if rec.Return7Day != nil {
			sum7Day += *rec.Return7Day
			if isWinFunc(*rec.Return7Day) {
				wins7Day++
				if rec.ConfidenceScore >= 70 {
					highConfWins++
				} else if rec.ConfidenceScore >= 40 {
					medConfWins++
				} else {
					lowConfWins++
				}
			}
			if rec.ConfidenceScore >= 70 {
				highConfTotal++
			} else if rec.ConfidenceScore >= 40 {
				medConfTotal++
			} else {
				lowConfTotal++
			}
		}
		if rec.Return14Day != nil {
			sum14Day += *rec.Return14Day
			if isWinFunc(*rec.Return14Day) {
				wins14Day++
			}
		}
		if rec.Return30Day != nil {
			sum30Day += *rec.Return30Day
			if isWinFunc(*rec.Return30Day) {
				wins30Day++
			}
		}
	}

	acc.WinCount1Day = wins1Day
	acc.WinCount7Day = wins7Day
	acc.WinCount14Day = wins14Day
	acc.WinCount30Day = wins30Day

	if count1Day > 0 {
		acc.WinRate1Day = float64(wins1Day) / float64(count1Day)
		acc.AvgReturn1Day = sum1Day / float64(count1Day)
	}
	if count7Day > 0 {
		acc.WinRate7Day = float64(wins7Day) / float64(count7Day)
		acc.AvgReturn7Day = sum7Day / float64(count7Day)
	}
	if count14Day > 0 {
		acc.WinRate14Day = float64(wins14Day) / float64(count14Day)
		acc.AvgReturn14Day = sum14Day / float64(count14Day)
	}
	if count30Day > 0 {
		acc.WinRate30Day = float64(wins30Day) / float64(count30Day)
		acc.AvgReturn30Day = sum30Day / float64(count30Day)
	}
	if total > 0 {
		acc.AvgConfidence = float64(totalConf) / float64(total)
	}
	if highConfTotal > 0 {
		acc.HighConfWinRate = float64(highConfWins) / float64(highConfTotal)
	}
	if medConfTotal > 0 {
		acc.MedConfWinRate = float64(medConfWins) / float64(medConfTotal)
	}
	if lowConfTotal > 0 {
		acc.LowConfWinRate = float64(lowConfWins) / float64(lowConfTotal)
	}

	return acc, nil
}

// ApplyRetentionPolicy removes records older than 2 years from both cache and database.
func (t *RecommendationTracker) ApplyRetentionPolicy(ctx context.Context) (int64, error) {
	return t.CleanupOldRecords(ctx)
}
