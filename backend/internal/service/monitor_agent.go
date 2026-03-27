package service

// ---------------------------------------------------------------------------
// Monitor_Agent — autonomous market pattern scanner
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 12.1 Run autonomously on schedule (30min trading hours 9:00-15:00 ICT, 2hr off-hours)
//   - 12.2 Scan OHLCV data for watchlist symbols via Data_Source_Router
//   - 12.6 Generate structured observations with confidence scores
//   - 12.7 Trigger Alert_Service for confidence ≥60
//   - 12.8 Store observations in Knowledge_Base
//   - 12.9 5-minute timeout with graceful failure handling

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	"github.com/dda10/vnstock-go"
	"github.com/tmc/langchaingo/llms"
)

// MonitorAgentConfig holds configuration for the Monitor_Agent.
type MonitorAgentConfig struct {
	// TradingHoursInterval is the scan interval during trading hours (default 30min)
	TradingHoursInterval time.Duration
	// OffHoursInterval is the scan interval outside trading hours (default 2hr)
	OffHoursInterval time.Duration
	// ScanTimeout is the maximum duration for a complete scan cycle (default 5min)
	ScanTimeout time.Duration
	// AlertConfidenceThreshold is the minimum confidence to trigger alerts (default 60)
	AlertConfidenceThreshold int
	// OHLCVLookbackDays is the number of days of historical data to fetch (default 60)
	OHLCVLookbackDays int
	// TradingHoursStart is the start of trading hours in ICT (default 9:00)
	TradingHoursStart int
	// TradingHoursEnd is the end of trading hours in ICT (default 15:00)
	TradingHoursEnd int
}

// DefaultMonitorAgentConfig returns the default configuration.
func DefaultMonitorAgentConfig() MonitorAgentConfig {
	return MonitorAgentConfig{
		TradingHoursInterval:     30 * time.Minute,
		OffHoursInterval:         2 * time.Hour,
		ScanTimeout:              5 * time.Minute,
		AlertConfidenceThreshold: 60,
		OHLCVLookbackDays:        60,
		TradingHoursStart:        9,
		TradingHoursEnd:          15,
	}
}

// MonitorAgent is an autonomous agent that scans market data for patterns.
type MonitorAgent struct {
	router           *infra.DataSourceRouter
	patternDetector  *PatternDetector
	watchlistService *WatchlistService
	db               *sql.DB
	config           MonitorAgentConfig

	// Scheduling
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex

	// ICT timezone (UTC+7)
	ictLocation *time.Location
}

// NewMonitorAgent creates a new Monitor_Agent with the given dependencies.
func NewMonitorAgent(
	router *infra.DataSourceRouter,
	patternDetector *PatternDetector,
	watchlistService *WatchlistService,
	db *sql.DB,
) *MonitorAgent {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		// Fallback to UTC+7 offset
		ict = time.FixedZone("ICT", 7*60*60)
	}

	return &MonitorAgent{
		router:           router,
		patternDetector:  patternDetector,
		watchlistService: watchlistService,
		db:               db,
		config:           DefaultMonitorAgentConfig(),
		stopChan:         make(chan struct{}),
		ictLocation:      ict,
	}
}

// SetConfig updates the agent configuration.
func (m *MonitorAgent) SetConfig(cfg MonitorAgentConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

// ---------------------------------------------------------------------------
// SubAgent interface implementation
// ---------------------------------------------------------------------------

// Name returns the agent identifier.
func (m *MonitorAgent) Name() string {
	return "Monitor_Agent"
}

// Execute runs a single scan cycle when invoked by the MultiAgentSystem.
// This allows the Monitor_Agent to be triggered on-demand in addition to
// its autonomous scheduled operation.
func (m *MonitorAgent) Execute(ctx context.Context, intent model.QueryIntent, llm llms.Model) (*model.AgentMessage, error) {
	// Run a scan cycle for the symbols in the intent, or all watchlist symbols
	symbols := intent.Symbols
	if len(symbols) == 0 {
		// Get all watched symbols from all users (for autonomous scanning)
		// In a real implementation, this would be scoped to a specific user
		symbols = m.getAllWatchedSymbols(ctx)
	}

	if len(symbols) == 0 {
		return &model.AgentMessage{
			AgentName:   m.Name(),
			PayloadType: "scan_result",
			Payload: map[string]interface{}{
				"status":       "no_symbols",
				"message":      "No symbols to scan",
				"observations": []model.PatternObservation{},
			},
			Timestamp: time.Now(),
		}, nil
	}

	observations, err := m.scanSymbols(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	return &model.AgentMessage{
		AgentName:   m.Name(),
		PayloadType: "scan_result",
		Payload: map[string]interface{}{
			"status":          "completed",
			"symbols_scanned": len(symbols),
			"patterns_found":  len(observations),
			"observations":    observations,
			"high_confidence": m.countHighConfidence(observations),
		},
		Timestamp: time.Now(),
	}, nil
}

// ---------------------------------------------------------------------------
// Autonomous scheduling (Requirement 12.1)
// ---------------------------------------------------------------------------

// Start begins the autonomous scanning schedule.
func (m *MonitorAgent) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	go m.runScheduler()
	log.Printf("[Monitor_Agent] Started autonomous scanning")
}

// Stop halts the autonomous scanning schedule.
func (m *MonitorAgent) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	close(m.stopChan)
	m.running = false
	log.Printf("[Monitor_Agent] Stopped autonomous scanning")
}

// IsRunning returns whether the agent is currently running.
func (m *MonitorAgent) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// runScheduler is the main scheduling loop.
func (m *MonitorAgent) runScheduler() {
	for {
		interval := m.getNextInterval()
		log.Printf("[Monitor_Agent] Next scan in %v", interval)

		select {
		case <-time.After(interval):
			m.runScheduledScan()
		case <-m.stopChan:
			return
		}
	}
}

// getNextInterval returns the appropriate interval based on trading hours.
func (m *MonitorAgent) getNextInterval() time.Duration {
	now := time.Now().In(m.ictLocation)
	hour := now.Hour()

	if hour >= m.config.TradingHoursStart && hour < m.config.TradingHoursEnd {
		return m.config.TradingHoursInterval
	}
	return m.config.OffHoursInterval
}

// IsTradingHours returns whether the current time is within trading hours.
func (m *MonitorAgent) IsTradingHours() bool {
	now := time.Now().In(m.ictLocation)
	hour := now.Hour()
	return hour >= m.config.TradingHoursStart && hour < m.config.TradingHoursEnd
}

// runScheduledScan executes a scheduled scan cycle.
func (m *MonitorAgent) runScheduledScan() {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.ScanTimeout)
	defer cancel()

	log.Printf("[Monitor_Agent] Starting scheduled scan cycle")
	startTime := time.Now()

	symbols := m.getAllWatchedSymbols(ctx)
	if len(symbols) == 0 {
		log.Printf("[Monitor_Agent] No symbols to scan")
		return
	}

	observations, err := m.scanSymbols(ctx, symbols)
	if err != nil {
		log.Printf("[Monitor_Agent] Scan cycle failed: %v", err)
		return
	}

	elapsed := time.Since(startTime)
	highConfidence := m.countHighConfidence(observations)
	log.Printf("[Monitor_Agent] Scan cycle completed: %d symbols, %d patterns found, %d high-confidence, took %v",
		len(symbols), len(observations), highConfidence, elapsed)
}

// ---------------------------------------------------------------------------
// Core scanning logic (Requirements 12.2, 12.6, 12.7, 12.8, 12.9)
// ---------------------------------------------------------------------------

// RunScanCycle executes a complete scan cycle for all watchlist symbols.
// This is the main entry point for manual or scheduled scans.
func (m *MonitorAgent) RunScanCycle(ctx context.Context) error {
	// Apply 5-minute timeout (Requirement 12.9)
	ctx, cancel := context.WithTimeout(ctx, m.config.ScanTimeout)
	defer cancel()

	symbols := m.getAllWatchedSymbols(ctx)
	if len(symbols) == 0 {
		log.Printf("[Monitor_Agent] No symbols to scan")
		return nil
	}

	_, err := m.scanSymbols(ctx, symbols)
	return err
}

// scanSymbols scans a list of symbols for patterns.
func (m *MonitorAgent) scanSymbols(ctx context.Context, symbols []string) ([]model.PatternObservation, error) {
	var allObservations []model.PatternObservation
	var failedSymbols []string
	var mu sync.Mutex

	// Use a semaphore to limit concurrent requests
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, symbol := range symbols {
		// Check for context cancellation (timeout)
		select {
		case <-ctx.Done():
			log.Printf("[Monitor_Agent] Scan timeout reached, skipping remaining symbols")
			// Log failure and skip incomplete symbols (Requirement 12.9)
			mu.Lock()
			failedSymbols = append(failedSymbols, symbol)
			mu.Unlock()
			continue
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(sym string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			observations, err := m.scanSymbol(ctx, sym)
			if err != nil {
				log.Printf("[Monitor_Agent] Failed to scan %s: %v", sym, err)
				mu.Lock()
				failedSymbols = append(failedSymbols, sym)
				mu.Unlock()
				return
			}

			mu.Lock()
			allObservations = append(allObservations, observations...)
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()

	if len(failedSymbols) > 0 {
		log.Printf("[Monitor_Agent] Failed to scan %d symbols: %v", len(failedSymbols), failedSymbols)
	}

	return allObservations, nil
}

// scanSymbol scans a single symbol for patterns.
func (m *MonitorAgent) scanSymbol(ctx context.Context, symbol string) ([]model.PatternObservation, error) {
	// Fetch OHLCV data via Data_Source_Router (Requirement 12.2)
	bars, err := m.fetchOHLCVData(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV data: %w", err)
	}

	if len(bars) < 20 {
		log.Printf("[Monitor_Agent] Insufficient data for %s: %d bars", symbol, len(bars))
		return nil, nil
	}

	// Use Pattern_Detector to identify patterns (Requirements 12.3, 12.4, 12.5)
	observations := m.patternDetector.DetectPatterns(symbol, bars)

	// Process each observation
	for i := range observations {
		obs := &observations[i]

		// Store observation in Knowledge_Base (Requirement 12.8)
		if err := m.storeObservation(ctx, obs); err != nil {
			log.Printf("[Monitor_Agent] Failed to store observation for %s: %v", symbol, err)
		}

		// Trigger Alert_Service for high-confidence patterns (Requirement 12.7)
		if obs.ConfidenceScore >= m.config.AlertConfidenceThreshold {
			if err := m.triggerAlert(ctx, obs); err != nil {
				log.Printf("[Monitor_Agent] Failed to trigger alert for %s: %v", symbol, err)
			}
		}
	}

	return observations, nil
}

// fetchOHLCVData fetches historical OHLCV data for a symbol.
func (m *MonitorAgent) fetchOHLCVData(ctx context.Context, symbol string) ([]model.OHLCVBar, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -m.config.OHLCVLookbackDays)

	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    startDate,
		End:      endDate,
		Interval: "1D",
	}

	quotes, _, err := m.router.FetchQuoteHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert vnstock.Quote to model.OHLCVBar
	bars := make([]model.OHLCVBar, len(quotes))
	for i, q := range quotes {
		bars[i] = model.OHLCVBar{
			Time:   q.Timestamp,
			Open:   q.Open,
			High:   q.High,
			Low:    q.Low,
			Close:  q.Close,
			Volume: q.Volume,
		}
	}

	return bars, nil
}

// ---------------------------------------------------------------------------
// Knowledge_Base integration (Requirement 12.8)
// ---------------------------------------------------------------------------

// storeObservation persists a pattern observation to the database.
func (m *MonitorAgent) storeObservation(ctx context.Context, obs *model.PatternObservation) error {
	if m.db == nil {
		log.Printf("[Monitor_Agent] Database not configured, skipping observation storage")
		return nil
	}

	query := `
		INSERT INTO pattern_observations 
		(symbol, pattern_type, detection_date, confidence_score, price_at_detection, supporting_data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := m.db.QueryRowContext(ctx, query,
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

	log.Printf("[Monitor_Agent] Stored observation: %s %s confidence=%d id=%d",
		obs.Symbol, obs.PatternType, obs.ConfidenceScore, obs.ID)

	return nil
}

// QueryObservations retrieves observations from the Knowledge_Base with filters.
func (m *MonitorAgent) QueryObservations(ctx context.Context, filters ObservationFilters) ([]model.PatternObservation, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT id, symbol, pattern_type, detection_date, confidence_score, 
		       price_at_detection, supporting_data, 
		       outcome_1day, outcome_7day, outcome_14day, outcome_30day
		FROM pattern_observations
		WHERE 1=1
	`
	args := []interface{}{}
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
	}

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query observations: %w", err)
	}
	defer rows.Close()

	var observations []model.PatternObservation
	for rows.Next() {
		var obs model.PatternObservation
		var patternType string
		err := rows.Scan(
			&obs.ID, &obs.Symbol, &patternType, &obs.DetectionDate,
			&obs.ConfidenceScore, &obs.PriceAtDetection, &obs.SupportingData,
			&obs.Outcome1Day, &obs.Outcome7Day, &obs.Outcome14Day, &obs.Outcome30Day,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan observation: %w", err)
		}
		obs.PatternType = model.PatternType(patternType)
		observations = append(observations, obs)
	}

	return observations, rows.Err()
}

// ObservationFilters defines filters for querying observations.
type ObservationFilters struct {
	Symbol        string
	PatternType   string
	StartDate     time.Time
	EndDate       time.Time
	MinConfidence int
	Limit         int
	Offset        int
}

// ---------------------------------------------------------------------------
// Alert_Service integration (Requirement 12.7)
// ---------------------------------------------------------------------------

// triggerAlert creates an alert for a high-confidence pattern observation.
func (m *MonitorAgent) triggerAlert(ctx context.Context, obs *model.PatternObservation) error {
	if m.db == nil {
		log.Printf("[Monitor_Agent] Database not configured, skipping alert creation")
		return nil
	}

	// Check for duplicate alerts (48-hour deduplication window)
	isDuplicate, err := m.isDuplicateAlert(ctx, obs)
	if err != nil {
		log.Printf("[Monitor_Agent] Failed to check for duplicate alert: %v", err)
	}
	if isDuplicate {
		log.Printf("[Monitor_Agent] Skipping duplicate alert for %s %s", obs.Symbol, obs.PatternType)
		return nil
	}

	// Generate explanation
	explanation := m.generateAlertExplanation(obs)

	// Insert alert (user_id=1 for now, will be expanded for multi-user support)
	query := `
		INSERT INTO alerts 
		(user_id, symbol, pattern_type, confidence_score, explanation, detection_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = m.db.ExecContext(ctx, query,
		1, // Default user_id
		obs.Symbol,
		string(obs.PatternType),
		obs.ConfidenceScore,
		explanation,
		obs.DetectionDate,
	)

	if err != nil {
		return fmt.Errorf("failed to insert alert: %w", err)
	}

	log.Printf("[Monitor_Agent] Created alert: %s %s confidence=%d",
		obs.Symbol, obs.PatternType, obs.ConfidenceScore)

	return nil
}

// isDuplicateAlert checks if a similar alert was created within the deduplication window.
func (m *MonitorAgent) isDuplicateAlert(ctx context.Context, obs *model.PatternObservation) (bool, error) {
	// 48-hour deduplication window, unless confidence increased by 10+
	query := `
		SELECT confidence_score FROM alerts
		WHERE symbol = $1 AND pattern_type = $2 
		AND detection_timestamp > $3
		ORDER BY detection_timestamp DESC
		LIMIT 1
	`

	deduplicationWindow := obs.DetectionDate.Add(-48 * time.Hour)

	var prevConfidence int
	err := m.db.QueryRowContext(ctx, query, obs.Symbol, string(obs.PatternType), deduplicationWindow).Scan(&prevConfidence)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Allow new alert if confidence increased by 10+ points
	if obs.ConfidenceScore >= prevConfidence+10 {
		return false, nil
	}

	return true, nil
}

// generateAlertExplanation creates a human-readable explanation for the alert.
func (m *MonitorAgent) generateAlertExplanation(obs *model.PatternObservation) string {
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
// Helper methods
// ---------------------------------------------------------------------------

// getAllWatchedSymbols retrieves all unique symbols from all user watchlists.
func (m *MonitorAgent) getAllWatchedSymbols(ctx context.Context) []string {
	if m.db == nil {
		return nil
	}

	query := `
		SELECT DISTINCT ws.symbol
		FROM watchlist_symbols ws
		JOIN watchlists w ON ws.watchlist_id = w.id
		ORDER BY ws.symbol
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[Monitor_Agent] Failed to query watched symbols: %v", err)
		return nil
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var sym string
		if err := rows.Scan(&sym); err != nil {
			log.Printf("[Monitor_Agent] Failed to scan symbol: %v", err)
			continue
		}
		symbols = append(symbols, sym)
	}

	return symbols
}

// countHighConfidence counts observations with confidence >= threshold.
func (m *MonitorAgent) countHighConfidence(observations []model.PatternObservation) int {
	count := 0
	for _, obs := range observations {
		if obs.ConfidenceScore >= m.config.AlertConfidenceThreshold {
			count++
		}
	}
	return count
}

// GetPatternAccuracy computes accuracy metrics for a pattern type.
func (m *MonitorAgent) GetPatternAccuracy(ctx context.Context, patternType model.PatternType) (*PatternAccuracy, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN outcome_7day > 0 THEN 1 END) as success_count,
			COUNT(CASE WHEN outcome_7day <= 0 THEN 1 END) as failure_count,
			COALESCE(AVG(outcome_7day), 0) as avg_price_change,
			COALESCE(AVG(confidence_score), 0) as avg_confidence
		FROM pattern_observations
		WHERE pattern_type = $1 AND outcome_7day IS NOT NULL
	`

	var accuracy PatternAccuracy
	accuracy.PatternType = patternType

	err := m.db.QueryRowContext(ctx, query, string(patternType)).Scan(
		&accuracy.TotalObservations,
		&accuracy.SuccessCount,
		&accuracy.FailureCount,
		&accuracy.AvgPriceChange,
		&accuracy.AvgConfidence,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to compute accuracy: %w", err)
	}

	return &accuracy, nil
}

// PatternAccuracy holds accuracy metrics for a pattern type.
type PatternAccuracy struct {
	PatternType       model.PatternType `json:"patternType"`
	TotalObservations int               `json:"totalObservations"`
	SuccessCount      int               `json:"successCount"`
	FailureCount      int               `json:"failureCount"`
	AvgPriceChange    float64           `json:"avgPriceChange"`
	AvgConfidence     float64           `json:"avgConfidence"`
}
