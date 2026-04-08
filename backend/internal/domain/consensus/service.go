package consensus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"myfi-backend/internal/infra"
)

// sourceWeights defines how much each source type contributes to the composite score.
// Analyst reports carry the most weight, followed by news, then social/forum.
var sourceWeights = map[SourceType]float64{
	SourceAnalyst:     0.35,
	SourceNews:        0.30,
	SourceSocialMedia: 0.20,
	SourceForum:       0.15,
}

// ConsensusService aggregates sentiment signals from multiple sources
// into a unified market consensus view.
type ConsensusService struct {
	db    *sql.DB
	cache *infra.Cache
}

// NewConsensusService creates a new consensus service.
func NewConsensusService(db *sql.DB, cache *infra.Cache) *ConsensusService {
	return &ConsensusService{
		db:    db,
		cache: cache,
	}
}

// GetConsensus computes the composite consensus score for a symbol.
func (s *ConsensusService) GetConsensus(ctx context.Context, symbol string, period string) (*ConsensusScore, error) {
	if period == "" {
		period = "7d"
	}

	cacheKey := fmt.Sprintf("consensus:%s:%s", symbol, period)
	if cached, found := s.cache.Get(cacheKey); found {
		if score, ok := cached.(*ConsensusScore); ok {
			return score, nil
		}
	}

	days, err := parsePeriodDays(period)
	if err != nil {
		return nil, err
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Query signals grouped by source
	breakdown, totalSignals, err := s.querySourceBreakdown(ctx, symbol, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query source breakdown: %w", err)
	}

	// Compute weighted composite score
	composite := computeComposite(breakdown)

	// Get previous period score for trend comparison
	prevStart := startDate.AddDate(0, 0, -days)
	prevBreakdown, _, _ := s.querySourceBreakdown(ctx, symbol, prevStart, startDate)
	var prevScore *float64
	var scoreChange *float64
	if len(prevBreakdown) > 0 {
		ps := computeComposite(prevBreakdown)
		prevScore = &ps
		change := composite - ps
		scoreChange = &change
	}

	result := &ConsensusScore{
		Symbol:          symbol,
		CompositeScore:  composite,
		Confidence:      computeConfidence(breakdown, totalSignals),
		SignalStrength:  classifySignal(composite),
		SourceBreakdown: breakdown,
		TotalSignals:    totalSignals,
		Period:          period,
		ComputedAt:      time.Now(),
		PreviousScore:   prevScore,
		ScoreChange:     scoreChange,
	}

	s.cache.Set(cacheKey, result, 15*time.Minute)
	return result, nil
}

// GetDivergences detects disagreements between source types for a symbol.
func (s *ConsensusService) GetDivergences(ctx context.Context, symbol string) ([]Divergence, error) {
	consensus, err := s.GetConsensus(ctx, symbol, "7d")
	if err != nil {
		return nil, err
	}

	var divergences []Divergence
	sources := consensus.SourceBreakdown

	for i := 0; i < len(sources); i++ {
		for j := i + 1; j < len(sources); j++ {
			gap := math.Abs(sources[i].Score - sources[j].Score)
			if gap < 0.4 {
				continue // not significant enough
			}

			significance := "low"
			if gap > 0.8 {
				significance = "high"
			} else if gap > 0.6 {
				significance = "medium"
			}

			divergences = append(divergences, Divergence{
				Symbol:       symbol,
				SourceA:      sources[i].Source,
				SourceB:      sources[j].Source,
				ScoreA:       sources[i].Score,
				ScoreB:       sources[j].Score,
				Gap:          gap,
				Significance: significance,
				DetectedAt:   time.Now(),
				Description: fmt.Sprintf("%s sentiment (%.2f) diverges from %s (%.2f)",
					sources[i].Source, sources[i].Score, sources[j].Source, sources[j].Score),
			})
		}
	}

	return divergences, nil
}

// GetMarketMood computes the overall market-wide sentiment.
func (s *ConsensusService) GetMarketMood(ctx context.Context) (*MarketMood, error) {
	cacheKey := "consensus:market_mood"
	if cached, found := s.cache.Get(cacheKey); found {
		if mood, ok := cached.(*MarketMood); ok {
			return mood, nil
		}
	}

	startDate := time.Now().AddDate(0, 0, -7)

	// Get overall market score
	query := `
		SELECT 
			COALESCE(AVG(score), 0) as overall_score,
			COUNT(*) as total_signals
		FROM consensus_signals
		WHERE created_at >= $1
	`

	mood := &MarketMood{ComputedAt: time.Now()}
	if err := s.db.QueryRowContext(ctx, query, startDate).Scan(
		&mood.OverallScore, &mood.TotalSignals,
	); err != nil {
		return nil, fmt.Errorf("failed to query market mood: %w", err)
	}

	mood.Label = classifyMood(mood.OverallScore)

	// Top bullish symbols
	mood.TopBullish, _ = s.queryTopSymbols(ctx, startDate, "DESC", 5)
	// Top bearish symbols
	mood.TopBearish, _ = s.queryTopSymbols(ctx, startDate, "ASC", 5)
	// Sector mood
	mood.SectorMood, _ = s.querySectorMood(ctx, startDate)

	s.cache.Set(cacheKey, mood, 15*time.Minute)
	return mood, nil
}

// StoreSignal persists a sentiment signal from any source.
func (s *ConsensusService) StoreSignal(ctx context.Context, signal *SignalRecord) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}

	topicsJSON, _ := json.Marshal(signal.Topics)

	query := `
		INSERT INTO consensus_signals
		(symbol, source, score, confidence, text_snippet, url, topics, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	return s.db.QueryRowContext(ctx, query,
		signal.Symbol, string(signal.Source), signal.Score, signal.Confidence,
		signal.Text, signal.URL, string(topicsJSON), signal.CreatedAt,
	).Scan(&signal.ID)
}

// ---------------------------------------------------------------------------
// Query helpers
// ---------------------------------------------------------------------------

func (s *ConsensusService) querySourceBreakdown(ctx context.Context, symbol string, start, end time.Time) ([]SourceSentiment, int, error) {
	query := `
		SELECT source,
		       AVG(score) as avg_score,
		       COUNT(*) as signal_count
		FROM consensus_signals
		WHERE symbol = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY source
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, start, end)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var breakdown []SourceSentiment
	totalSignals := 0

	for rows.Next() {
		var ss SourceSentiment
		var source string
		if err := rows.Scan(&source, &ss.Score, &ss.SignalCount); err != nil {
			return nil, 0, err
		}
		ss.Source = SourceType(source)
		ss.Weight = sourceWeights[ss.Source]
		totalSignals += ss.SignalCount

		// Get top themes for this source
		ss.TopPositive, ss.TopNegative = s.queryTopThemes(ctx, symbol, source, start, end)

		breakdown = append(breakdown, ss)
	}

	return breakdown, totalSignals, rows.Err()
}

func (s *ConsensusService) queryTopThemes(ctx context.Context, symbol, source string, start, end time.Time) ([]string, []string) {
	// Query topics from positive signals
	posQuery := `
		SELECT topics FROM consensus_signals
		WHERE symbol = $1 AND source = $2 AND score > 0.3
		AND created_at BETWEEN $3 AND $4
		ORDER BY score DESC LIMIT 10
	`

	negQuery := `
		SELECT topics FROM consensus_signals
		WHERE symbol = $1 AND source = $2 AND score < -0.3
		AND created_at BETWEEN $3 AND $4
		ORDER BY score ASC LIMIT 10
	`

	positive := extractTopics(s.db, ctx, posQuery, symbol, source, start, end)
	negative := extractTopics(s.db, ctx, negQuery, symbol, source, start, end)

	return dedup(positive, 3), dedup(negative, 3)
}

func extractTopics(db *sql.DB, ctx context.Context, query, symbol, source string, start, end time.Time) []string {
	rows, err := db.QueryContext(ctx, query, symbol, source, start, end)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var all []string
	for rows.Next() {
		var topicsJSON string
		if err := rows.Scan(&topicsJSON); err != nil {
			continue
		}
		var topics []string
		if err := json.Unmarshal([]byte(topicsJSON), &topics); err == nil {
			all = append(all, topics...)
		}
	}
	return all
}

func (s *ConsensusService) queryTopSymbols(ctx context.Context, since time.Time, order string, limit int) ([]SymbolSentiment, error) {
	query := fmt.Sprintf(`
		SELECT symbol, AVG(score) as avg_score, COUNT(*) as signals
		FROM consensus_signals
		WHERE created_at >= $1
		GROUP BY symbol
		HAVING COUNT(*) >= 3
		ORDER BY avg_score %s
		LIMIT $2
	`, order)

	rows, err := s.db.QueryContext(ctx, query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SymbolSentiment
	for rows.Next() {
		var ss SymbolSentiment
		if err := rows.Scan(&ss.Symbol, &ss.Score, &ss.Signals); err != nil {
			continue
		}
		result = append(result, ss)
	}
	return result, nil
}

func (s *ConsensusService) querySectorMood(ctx context.Context, since time.Time) ([]SectorSentiment, error) {
	// This joins with a sector mapping. For now, return signals grouped by
	// the first 3 chars of symbol as a rough sector proxy.
	// In production, join with the sector table.
	query := `
		SELECT symbol, AVG(score) as avg_score, COUNT(*) as signals
		FROM consensus_signals
		WHERE created_at >= $1
		GROUP BY symbol
		ORDER BY avg_score DESC
	`

	rows, err := s.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sectorMap := make(map[string]*SectorSentiment)
	for rows.Next() {
		var symbol string
		var score float64
		var signals int
		if err := rows.Scan(&symbol, &score, &signals); err != nil {
			continue
		}
		// Group by sector would happen here with a proper sector lookup
		sector := "market" // placeholder
		if ss, ok := sectorMap[sector]; ok {
			ss.Score = (ss.Score*float64(ss.Signals) + score*float64(signals)) / float64(ss.Signals+signals)
			ss.Signals += signals
		} else {
			sectorMap[sector] = &SectorSentiment{Sector: sector, Score: score, Signals: signals}
		}
	}

	var result []SectorSentiment
	for _, ss := range sectorMap {
		result = append(result, *ss)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Computation helpers
// ---------------------------------------------------------------------------

func computeComposite(breakdown []SourceSentiment) float64 {
	if len(breakdown) == 0 {
		return 0
	}

	totalWeight := 0.0
	weightedSum := 0.0

	for _, ss := range breakdown {
		w := ss.Weight
		if w == 0 {
			w = 0.1 // default weight for unknown sources
		}
		weightedSum += ss.Score * w
		totalWeight += w
	}

	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}

func computeConfidence(breakdown []SourceSentiment, totalSignals int) float64 {
	if totalSignals == 0 {
		return 0
	}

	// Confidence based on: data volume + source agreement
	volumeScore := math.Min(float64(totalSignals)/50.0, 1.0) // max at 50 signals
	sourceCount := float64(len(breakdown))
	diversityScore := math.Min(sourceCount/4.0, 1.0) // max at 4 source types

	// Agreement: low variance = high agreement
	if len(breakdown) < 2 {
		return volumeScore * 0.5
	}

	mean := computeComposite(breakdown)
	variance := 0.0
	for _, ss := range breakdown {
		diff := ss.Score - mean
		variance += diff * diff
	}
	variance /= float64(len(breakdown))
	agreementScore := math.Max(0, 1.0-variance)

	return (volumeScore*0.3 + diversityScore*0.3 + agreementScore*0.4)
}

func classifySignal(score float64) string {
	switch {
	case score > 0.6:
		return "strong_buy"
	case score > 0.2:
		return "buy"
	case score > -0.2:
		return "neutral"
	case score > -0.6:
		return "sell"
	default:
		return "strong_sell"
	}
}

func classifyMood(score float64) string {
	switch {
	case score > 0.5:
		return "greed"
	case score > 0.2:
		return "optimistic"
	case score > -0.2:
		return "neutral"
	case score > -0.5:
		return "cautious"
	default:
		return "fear"
	}
}

func parsePeriodDays(period string) (int, error) {
	switch period {
	case "1d":
		return 1, nil
	case "7d":
		return 7, nil
	case "30d":
		return 30, nil
	default:
		return 0, fmt.Errorf("invalid period %q", period)
	}
}

func dedup(items []string, max int) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if seen[item] || item == "" {
			continue
		}
		seen[item] = true
		result = append(result, item)
		if len(result) >= max {
			break
		}
	}
	return result
}
