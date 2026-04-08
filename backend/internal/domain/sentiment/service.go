package sentiment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"myfi-backend/internal/infra"
)

// LLMAnalyzer abstracts the LLM call for sentiment analysis.
// Implemented by the GRPCClient wrapper in the service layer.
type LLMAnalyzer interface {
	AnalyzeSentiment(ctx context.Context, text string, symbol string) (*ArticleAnalysis, error)
}

// SentimentService orchestrates sentiment analysis, caching, and persistence.
type SentimentService struct {
	db       *sql.DB
	cache    *infra.Cache
	analyzer LLMAnalyzer
}

// NewSentimentService creates a new sentiment service.
func NewSentimentService(db *sql.DB, cache *infra.Cache, analyzer LLMAnalyzer) *SentimentService {
	return &SentimentService{
		db:       db,
		cache:    cache,
		analyzer: analyzer,
	}
}

// AnalyzeArticle runs LLM sentiment analysis on a single article and persists the result.
func (s *SentimentService) AnalyzeArticle(ctx context.Context, req AnalyzeRequest) (*ArticleAnalysis, error) {
	text := req.Text
	if text == "" && req.URL != "" {
		// In production, fetch article content from URL via an HTTP client.
		// For now, return an error if no text is provided.
		return nil, fmt.Errorf("article text is required (URL fetching not yet implemented)")
	}
	if text == "" {
		return nil, fmt.Errorf("either text or url must be provided")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("sentiment:article:%s:%s", req.Symbol, hashText(text))
	if cached, found := s.cache.Get(cacheKey); found {
		if analysis, ok := cached.(*ArticleAnalysis); ok {
			return analysis, nil
		}
	}

	// Call LLM
	analysis, err := s.analyzer.AnalyzeSentiment(ctx, text, req.Symbol)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	analysis.Symbol = req.Symbol
	analysis.URL = req.URL
	analysis.AnalyzedAt = time.Now()
	analysis.RawText = text

	// Persist to DB
	if err := s.storeAnalysis(ctx, analysis); err != nil {
		log.Printf("[SentimentService] Failed to persist analysis: %v", err)
		// Non-fatal: still return the analysis
	}

	// Cache for 1 hour
	s.cache.Set(cacheKey, analysis, 1*time.Hour)

	return analysis, nil
}

// GetTrend returns aggregated sentiment trend for a symbol over a period.
func (s *SentimentService) GetTrend(ctx context.Context, symbol string, period string) (*SentimentTrend, error) {
	if period == "" {
		period = "7d"
	}

	days, err := parsePeriodDays(period)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("sentiment:trend:%s:%s", symbol, period)
	if cached, found := s.cache.Get(cacheKey); found {
		if trend, ok := cached.(*SentimentTrend); ok {
			return trend, nil
		}
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	trend, err := s.queryTrend(ctx, symbol, startDate, endDate)
	if err != nil {
		return nil, err
	}

	trend.Symbol = symbol
	trend.Period = period
	trend.StartDate = startDate
	trend.EndDate = endDate

	// Determine trend direction by comparing first half vs second half
	trend.Trend = s.computeTrendDirection(ctx, symbol, startDate, endDate)

	// Cache for 15 minutes
	s.cache.Set(cacheKey, trend, 15*time.Minute)

	return trend, nil
}

// GetTimeSeries returns daily sentiment snapshots for charting.
func (s *SentimentService) GetTimeSeries(ctx context.Context, symbol string, days int) ([]SentimentSnapshot, error) {
	if days <= 0 {
		days = 30
	}

	cacheKey := fmt.Sprintf("sentiment:ts:%s:%d", symbol, days)
	if cached, found := s.cache.Get(cacheKey); found {
		if snapshots, ok := cached.([]SentimentSnapshot); ok {
			return snapshots, nil
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT DATE(analyzed_at) as date,
		       AVG(CASE sentiment 
		           WHEN 'positive' THEN 1.0 
		           WHEN 'negative' THEN -1.0 
		           ELSE 0.0 END) as sentiment_score,
		       COUNT(*) as article_count
		FROM article_sentiments
		WHERE symbol = $1 AND analyzed_at >= $2
		GROUP BY DATE(analyzed_at)
		ORDER BY date ASC
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query time series: %w", err)
	}
	defer rows.Close()

	var snapshots []SentimentSnapshot
	for rows.Next() {
		var snap SentimentSnapshot
		if err := rows.Scan(&snap.Date, &snap.SentimentScore, &snap.ArticleCount); err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snap)
	}

	s.cache.Set(cacheKey, snapshots, 15*time.Minute)
	return snapshots, rows.Err()
}

// GetRecentArticles returns the most recently analyzed articles for a symbol.
func (s *SentimentService) GetRecentArticles(ctx context.Context, symbol string, limit int) ([]ArticleAnalysis, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, symbol, title, url, source, published_at, analyzed_at,
		       sentiment, confidence_score, summary, key_topics, impact_score
		FROM article_sentiments
		WHERE symbol = $1
		ORDER BY analyzed_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []ArticleAnalysis
	for rows.Next() {
		var a ArticleAnalysis
		var topicsJSON string
		if err := rows.Scan(
			&a.ID, &a.Symbol, &a.Title, &a.URL, &a.Source,
			&a.PublishedAt, &a.AnalyzedAt, &a.Sentiment,
			&a.ConfidenceScore, &a.Summary, &topicsJSON, &a.ImpactScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		_ = json.Unmarshal([]byte(topicsJSON), &a.KeyTopics)
		articles = append(articles, a)
	}

	return articles, rows.Err()
}

// ---------------------------------------------------------------------------
// Persistence helpers
// ---------------------------------------------------------------------------

func (s *SentimentService) storeAnalysis(ctx context.Context, a *ArticleAnalysis) error {
	if s.db == nil {
		return fmt.Errorf("database not configured")
	}

	topicsJSON, _ := json.Marshal(a.KeyTopics)

	query := `
		INSERT INTO article_sentiments
		(symbol, title, url, source, published_at, analyzed_at,
		 sentiment, confidence_score, summary, key_topics, impact_score, raw_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	return s.db.QueryRowContext(ctx, query,
		a.Symbol, a.Title, a.URL, a.Source, a.PublishedAt, a.AnalyzedAt,
		string(a.Sentiment), a.ConfidenceScore, a.Summary,
		string(topicsJSON), a.ImpactScore, a.RawText,
	).Scan(&a.ID)
}

func (s *SentimentService) queryTrend(ctx context.Context, symbol string, start, end time.Time) (*SentimentTrend, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN sentiment = 'positive' THEN 1 END) as positive,
			COUNT(CASE WHEN sentiment = 'negative' THEN 1 END) as negative,
			COUNT(CASE WHEN sentiment = 'neutral' THEN 1 END) as neutral_count,
			COALESCE(AVG(confidence_score), 0) as avg_confidence,
			COALESCE(AVG(impact_score), 0) as avg_impact,
			COALESCE(AVG(CASE sentiment 
				WHEN 'positive' THEN 1.0 
				WHEN 'negative' THEN -1.0 
				ELSE 0.0 END), 0) as sentiment_score
		FROM article_sentiments
		WHERE symbol = $1 AND analyzed_at BETWEEN $2 AND $3
	`

	trend := &SentimentTrend{}
	err := s.db.QueryRowContext(ctx, query, symbol, start, end).Scan(
		&trend.TotalArticles, &trend.PositiveCount, &trend.NegativeCount,
		&trend.NeutralCount, &trend.AvgConfidence, &trend.AvgImpact,
		&trend.SentimentScore,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend: %w", err)
	}

	return trend, nil
}

func (s *SentimentService) computeTrendDirection(ctx context.Context, symbol string, start, end time.Time) string {
	mid := start.Add(end.Sub(start) / 2)

	firstHalf, err := s.queryTrend(ctx, symbol, start, mid)
	if err != nil || firstHalf.TotalArticles == 0 {
		return "stable"
	}

	secondHalf, err := s.queryTrend(ctx, symbol, mid, end)
	if err != nil || secondHalf.TotalArticles == 0 {
		return "stable"
	}

	diff := secondHalf.SentimentScore - firstHalf.SentimentScore
	switch {
	case diff > 0.15:
		return "improving"
	case diff < -0.15:
		return "deteriorating"
	default:
		return "stable"
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parsePeriodDays(period string) (int, error) {
	period = strings.TrimSpace(period)
	switch period {
	case "1d":
		return 1, nil
	case "7d":
		return 7, nil
	case "30d":
		return 30, nil
	case "90d":
		return 90, nil
	default:
		return 0, fmt.Errorf("invalid period %q, use 1d/7d/30d/90d", period)
	}
}

// hashText produces a simple hash for cache keys. Not cryptographic.
func hashText(text string) string {
	if len(text) > 100 {
		text = text[:100]
	}
	h := uint32(0)
	for _, c := range text {
		h = h*31 + uint32(c)
	}
	return fmt.Sprintf("%x", h)
}
