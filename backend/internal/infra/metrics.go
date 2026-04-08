// Package infra provides infrastructure adapters.
//
// metrics.go implements Prometheus-compatible metrics collection for
// request latency, error rates, cache hit ratios, agent response times,
// and data source availability.
// Requirements: 50.10
package infra

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics collects Prometheus-compatible application metrics.
type Metrics struct {
	// HTTP request metrics
	requestCount    sync.Map // key: "method:path:status" → *int64
	requestDuration sync.Map // key: "method:path" → *durationBucket

	// Cache metrics
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64

	// Agent (Python AI Service) metrics
	agentRequestCount    sync.Map // key: "method" → *int64
	agentRequestDuration sync.Map // key: "method" → *durationBucket
	agentErrors          atomic.Int64

	// Data source metrics
	dataSourceRequests sync.Map // key: "source:status" → *int64
	circuitBreakerOpen sync.Map // key: "source" → *int64 (1=open, 0=closed)
}

type durationBucket struct {
	mu    sync.Mutex
	sum   float64
	count int64
}

func (d *durationBucket) observe(seconds float64) {
	d.mu.Lock()
	d.sum += seconds
	d.count++
	d.mu.Unlock()
}

func (d *durationBucket) get() (float64, int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.sum, d.count
}

// Global metrics instance.
var globalMetrics = &Metrics{}

// GetMetrics returns the global metrics instance.
func GetMetrics() *Metrics { return globalMetrics }

// RecordRequest records an HTTP request metric.
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	// Normalize path to reduce cardinality (strip IDs)
	normalized := normalizePath(path)

	countKey := fmt.Sprintf("%s:%s:%d", method, normalized, status)
	counter := m.getOrCreateCounter(&m.requestCount, countKey)
	atomic.AddInt64(counter, 1)

	durKey := fmt.Sprintf("%s:%s", method, normalized)
	bucket := m.getOrCreateBucket(&m.requestDuration, durKey)
	bucket.observe(duration.Seconds())
}

// RecordCacheHit increments the cache hit counter.
func (m *Metrics) RecordCacheHit() { m.cacheHits.Add(1) }

// RecordCacheMiss increments the cache miss counter.
func (m *Metrics) RecordCacheMiss() { m.cacheMisses.Add(1) }

// RecordAgentRequest records a Python AI Service call metric.
func (m *Metrics) RecordAgentRequest(method string, duration time.Duration, err error) {
	statusKey := method + ":ok"
	if err != nil {
		statusKey = method + ":error"
		m.agentErrors.Add(1)
	}
	counter := m.getOrCreateCounter(&m.agentRequestCount, statusKey)
	atomic.AddInt64(counter, 1)

	bucket := m.getOrCreateBucket(&m.agentRequestDuration, method)
	bucket.observe(duration.Seconds())
}

// RecordDataSourceRequest records a data source request metric.
func (m *Metrics) RecordDataSourceRequest(source string, success bool) {
	status := "ok"
	if !success {
		status = "error"
	}
	key := source + ":" + status
	counter := m.getOrCreateCounter(&m.dataSourceRequests, key)
	atomic.AddInt64(counter, 1)
}

// SetCircuitBreakerState records whether a circuit breaker is open.
func (m *Metrics) SetCircuitBreakerState(source string, open bool) {
	val := int64(0)
	if open {
		val = 1
	}
	counter := m.getOrCreateCounter(&m.circuitBreakerOpen, source)
	atomic.StoreInt64(counter, val)
}

func (m *Metrics) getOrCreateCounter(store *sync.Map, key string) *int64 {
	if v, ok := store.Load(key); ok {
		return v.(*int64)
	}
	counter := new(int64)
	actual, _ := store.LoadOrStore(key, counter)
	return actual.(*int64)
}

func (m *Metrics) getOrCreateBucket(store *sync.Map, key string) *durationBucket {
	if v, ok := store.Load(key); ok {
		return v.(*durationBucket)
	}
	bucket := &durationBucket{}
	actual, _ := store.LoadOrStore(key, bucket)
	return actual.(*durationBucket)
}

// PrometheusHandler returns a Gin handler that serves metrics in Prometheus text format.
// Served at GET /metrics.
func PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := globalMetrics
		var b strings.Builder

		// HTTP request count
		b.WriteString("# HELP http_requests_total Total HTTP requests.\n")
		b.WriteString("# TYPE http_requests_total counter\n")
		m.requestCount.Range(func(key, value any) bool {
			parts := strings.SplitN(key.(string), ":", 3)
			if len(parts) == 3 {
				fmt.Fprintf(&b, "http_requests_total{method=%q,path=%q,status=%q} %d\n",
					parts[0], parts[1], parts[2], atomic.LoadInt64(value.(*int64)))
			}
			return true
		})

		// HTTP request duration
		b.WriteString("# HELP http_request_duration_seconds HTTP request duration.\n")
		b.WriteString("# TYPE http_request_duration_seconds summary\n")
		m.requestDuration.Range(func(key, value any) bool {
			parts := strings.SplitN(key.(string), ":", 2)
			if len(parts) == 2 {
				sum, count := value.(*durationBucket).get()
				fmt.Fprintf(&b, "http_request_duration_seconds_sum{method=%q,path=%q} %f\n", parts[0], parts[1], sum)
				fmt.Fprintf(&b, "http_request_duration_seconds_count{method=%q,path=%q} %d\n", parts[0], parts[1], count)
			}
			return true
		})

		// Cache metrics
		hits := m.cacheHits.Load()
		misses := m.cacheMisses.Load()
		b.WriteString("# HELP cache_hits_total Cache hit count.\n")
		b.WriteString("# TYPE cache_hits_total counter\n")
		fmt.Fprintf(&b, "cache_hits_total %d\n", hits)
		b.WriteString("# HELP cache_misses_total Cache miss count.\n")
		b.WriteString("# TYPE cache_misses_total counter\n")
		fmt.Fprintf(&b, "cache_misses_total %d\n", misses)
		if total := hits + misses; total > 0 {
			b.WriteString("# HELP cache_hit_ratio Cache hit ratio.\n")
			b.WriteString("# TYPE cache_hit_ratio gauge\n")
			fmt.Fprintf(&b, "cache_hit_ratio %f\n", float64(hits)/float64(total))
		}

		// Agent metrics
		b.WriteString("# HELP agent_requests_total AI agent request count.\n")
		b.WriteString("# TYPE agent_requests_total counter\n")
		m.agentRequestCount.Range(func(key, value any) bool {
			parts := strings.SplitN(key.(string), ":", 2)
			if len(parts) == 2 {
				fmt.Fprintf(&b, "agent_requests_total{method=%q,status=%q} %d\n",
					parts[0], parts[1], atomic.LoadInt64(value.(*int64)))
			}
			return true
		})

		b.WriteString("# HELP agent_request_duration_seconds AI agent response time.\n")
		b.WriteString("# TYPE agent_request_duration_seconds summary\n")
		m.agentRequestDuration.Range(func(key, value any) bool {
			sum, count := value.(*durationBucket).get()
			fmt.Fprintf(&b, "agent_request_duration_seconds_sum{method=%q} %f\n", key.(string), sum)
			fmt.Fprintf(&b, "agent_request_duration_seconds_count{method=%q} %d\n", key.(string), count)
			return true
		})

		b.WriteString("# HELP agent_errors_total AI agent error count.\n")
		b.WriteString("# TYPE agent_errors_total counter\n")
		fmt.Fprintf(&b, "agent_errors_total %d\n", m.agentErrors.Load())

		// Data source metrics
		b.WriteString("# HELP datasource_requests_total Data source request count.\n")
		b.WriteString("# TYPE datasource_requests_total counter\n")
		m.dataSourceRequests.Range(func(key, value any) bool {
			parts := strings.SplitN(key.(string), ":", 2)
			if len(parts) == 2 {
				fmt.Fprintf(&b, "datasource_requests_total{source=%q,status=%q} %d\n",
					parts[0], parts[1], atomic.LoadInt64(value.(*int64)))
			}
			return true
		})

		b.WriteString("# HELP circuit_breaker_open Circuit breaker state (1=open, 0=closed).\n")
		b.WriteString("# TYPE circuit_breaker_open gauge\n")
		m.circuitBreakerOpen.Range(func(key, value any) bool {
			fmt.Fprintf(&b, "circuit_breaker_open{source=%q} %d\n",
				key.(string), atomic.LoadInt64(value.(*int64)))
			return true
		})

		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(b.String()))
	}
}

// MetricsMiddleware records request latency and status for every HTTP request.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		globalMetrics.RecordRequest(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// normalizePath reduces path cardinality by replacing dynamic segments with placeholders.
func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if looksLikeID(p) {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}

func looksLikeID(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Pure numeric
	allDigits := true
	for _, c := range s {
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if allDigits && len(s) > 0 {
		return true
	}
	// UUID-like
	if len(s) == 36 && strings.Count(s, "-") == 4 {
		return true
	}
	return false
}

// SortedKeys returns sorted keys from a sync.Map for deterministic output.
func SortedKeys(m *sync.Map) []string {
	var keys []string
	m.Range(func(key, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	return keys
}
