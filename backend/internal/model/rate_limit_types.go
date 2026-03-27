package model

// RateLimitMetrics contains metrics for a source
type RateLimitMetrics struct {
	Source         string `json:"source"`
	CurrentCount   int    `json:"currentCount"`
	MaxRequests    int    `json:"maxRequests"`
	QueueDepth     int    `json:"queueDepth"`
	WindowDuration string `json:"windowDuration"`
}
