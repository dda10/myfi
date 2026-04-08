// Package sentiment provides LLM-powered sentiment analysis for Vietnamese
// financial news articles. It scores articles as positive/negative/neutral,
// extracts key topics, and tracks sentiment trends per symbol over time.
//
// Architecture:
//   - SentimentService: orchestrates LLM calls, caching, and persistence
//   - Handlers: HTTP handlers for sentiment API endpoints
//   - Types: domain types for sentiment scores, article analysis, trends
//
// The service delegates actual LLM inference to the Python AI Service via
// the GRPCClient (gRPC primary, REST fallback), keeping the Go layer
// focused on orchestration, caching, and storage.
package sentiment
