package infra

import (
	"context"
	"log/slog"
	"sync"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// Vietnam timezone for archival scheduling.
var ictLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.FixedZone("ICT", 7*60*60)
	}
	return loc
}()

// Job intervals.
const (
	tradingInterval  = 5 * time.Minute
	offHoursInterval = 30 * time.Minute
)

// Archival hour (ICT).
const archivalHour = 23

// JobFunc is a function executed by the scheduler. It receives a context
// that is cancelled on shutdown.
type JobFunc func(ctx context.Context) error

// Scheduler manages periodic jobs with trading-hours-aware intervals
// and a nightly archival window.
//
// Two job categories:
//   - Evaluation jobs: run every 5 min during trading hours (9:00–15:00 ICT, Mon–Fri),
//     every 30 min outside trading hours.
//   - Archival jobs: run once nightly at ~23:00 ICT.
//
// Requirements: 36.6 (mission trigger evaluation intervals), 40.5 (nightly archival job).
type Scheduler struct {
	evalJobs []namedJob
	archJobs []namedJob
	logger   *slog.Logger
	mu       sync.Mutex
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  bool
	nowFunc  func() time.Time // for testing
}

type namedJob struct {
	name string
	fn   JobFunc
}

// NewScheduler creates a scheduler. Call Start to begin execution.
func NewScheduler(logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		logger:  logger,
		nowFunc: time.Now,
	}
}

// RegisterEvalJob adds a mission/trigger evaluation job.
// Must be called before Start.
func (s *Scheduler) RegisterEvalJob(name string, fn JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evalJobs = append(s.evalJobs, namedJob{name: name, fn: fn})
	s.logger.Info("registered evaluation job", "job", name)
}

// RegisterArchivalJob adds a nightly archival job.
// Must be called before Start.
func (s *Scheduler) RegisterArchivalJob(name string, fn JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.archJobs = append(s.archJobs, namedJob{name: name, fn: fn})
	s.logger.Info("registered archival job", "job", name)
}

// Start launches the scheduler goroutines. It is safe to call only once.
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(2)
	go s.evalLoop(ctx)
	go s.archivalLoop(ctx)

	s.logger.Info("scheduler started",
		"evalJobs", len(s.evalJobs),
		"archivalJobs", len(s.archJobs),
	)
}

// Stop gracefully shuts down the scheduler and waits for in-flight jobs.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

// IsTradingHours reports whether the given time falls within Vietnam market
// trading hours by querying vnstock-go's TradingHours API for HOSE.
// Falls back to static 9:00–15:00 ICT Mon–Fri if the API call fails.
func IsTradingHours(t time.Time) bool {
	status, err := vnstock.TradingHours("HOSE")
	if err == nil {
		return status.IsTradingHour
	}
	// Fallback: static check for 9:00–15:00 ICT, Mon–Fri.
	ict := t.In(ictLocation)
	wd := ict.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	hour := ict.Hour()
	return hour >= 9 && hour < 15
}

// evalLoop runs evaluation jobs at the appropriate interval.
func (s *Scheduler) evalLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		interval := s.currentEvalInterval()
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			s.runJobs(ctx, s.evalJobs, "eval")
		}
	}
}

// archivalLoop runs archival jobs once per night around 23:00 ICT.
func (s *Scheduler) archivalLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		delay := s.timeUntilArchival()
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			s.runJobs(ctx, s.archJobs, "archival")
		}
	}
}

// currentEvalInterval returns 5 min during trading hours, 30 min otherwise.
func (s *Scheduler) currentEvalInterval() time.Duration {
	if IsTradingHours(s.nowFunc()) {
		return tradingInterval
	}
	return offHoursInterval
}

// timeUntilArchival computes the duration until the next 23:00 ICT.
func (s *Scheduler) timeUntilArchival() time.Duration {
	now := s.nowFunc().In(ictLocation)
	next := time.Date(now.Year(), now.Month(), now.Day(), archivalHour, 0, 0, 0, ictLocation)
	if !now.Before(next) {
		// Already past 23:00 today — schedule for tomorrow.
		next = next.AddDate(0, 0, 1)
	}
	return next.Sub(now)
}

// runJobs executes all jobs in a category sequentially, logging errors.
func (s *Scheduler) runJobs(ctx context.Context, jobs []namedJob, category string) {
	for _, j := range jobs {
		if ctx.Err() != nil {
			return
		}
		start := s.nowFunc()
		err := j.fn(ctx)
		elapsed := s.nowFunc().Sub(start)
		if err != nil {
			s.logger.Error("scheduler job failed",
				"category", category,
				"job", j.name,
				"elapsed", elapsed,
				"error", err,
			)
		} else {
			s.logger.Info("scheduler job completed",
				"category", category,
				"job", j.name,
				"elapsed", elapsed,
			)
		}
	}
}
