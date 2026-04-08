package sentiment

import "time"

// Sentiment represents the sentiment classification of a text.
type Sentiment string

const (
	Positive Sentiment = "positive"
	Negative Sentiment = "negative"
	Neutral  Sentiment = "neutral"
)

// ArticleAnalysis is the result of LLM sentiment analysis on a single article.
type ArticleAnalysis struct {
	ID              int64     `json:"id"`
	Symbol          string    `json:"symbol"` // stock symbol mentioned (empty if market-wide)
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	Source          string    `json:"source"` // e.g. "cafef", "vietstock", "vnexpress"
	PublishedAt     time.Time `json:"publishedAt"`
	AnalyzedAt      time.Time `json:"analyzedAt"`
	Sentiment       Sentiment `json:"sentiment"`
	ConfidenceScore float64   `json:"confidenceScore"` // 0.0–1.0
	Summary         string    `json:"summary"`         // LLM-generated 1-2 sentence summary
	KeyTopics       []string  `json:"keyTopics"`       // extracted topics/themes
	ImpactScore     float64   `json:"impactScore"`     // estimated market impact 0.0–1.0
	RawText         string    `json:"-"`               // original article text (not exposed in API)
}

// SentimentTrend aggregates sentiment over a time window for a symbol.
type SentimentTrend struct {
	Symbol         string    `json:"symbol"`
	Period         string    `json:"period"` // "1d", "7d", "30d"
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	TotalArticles  int       `json:"totalArticles"`
	PositiveCount  int       `json:"positiveCount"`
	NegativeCount  int       `json:"negativeCount"`
	NeutralCount   int       `json:"neutralCount"`
	AvgConfidence  float64   `json:"avgConfidence"`
	AvgImpact      float64   `json:"avgImpact"`
	SentimentScore float64   `json:"sentimentScore"` // -1.0 (bearish) to +1.0 (bullish)
	Trend          string    `json:"trend"`          // "improving", "deteriorating", "stable"
}

// SentimentSnapshot is a point-in-time sentiment reading for a symbol,
// used for time-series charting on the frontend.
type SentimentSnapshot struct {
	Date           time.Time `json:"date"`
	SentimentScore float64   `json:"sentimentScore"`
	ArticleCount   int       `json:"articleCount"`
}

// AnalyzeRequest is the input for on-demand article analysis.
type AnalyzeRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	URL    string `json:"url,omitempty"`  // analyze a specific article
	Text   string `json:"text,omitempty"` // or provide raw text directly
}

// AnalyzeResponse wraps the analysis result.
type AnalyzeResponse struct {
	Analysis *ArticleAnalysis `json:"analysis"`
}

// TrendRequest is the query for sentiment trends.
type TrendRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Period string `json:"period"` // "1d", "7d", "30d" (default "7d")
}

// TimeSeriesRequest is the query for sentiment time-series data.
type TimeSeriesRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Days   int    `json:"days"` // number of days (default 30)
}
