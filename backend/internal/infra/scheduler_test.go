package infra

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsTradingHours(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{
			name: "Monday 10:00 ICT is trading",
			time: time.Date(2025, 7, 7, 3, 0, 0, 0, time.UTC), // 10:00 ICT
			want: true,
		},
		{
			name: "Monday 9:00 ICT is trading",
			time: time.Date(2025, 7, 7, 2, 0, 0, 0, time.UTC), // 9:00 ICT
			want: true,
		},
		{
			name: "Monday 14:59 ICT is trading",
			time: time.Date(2025, 7, 7, 7, 59, 0, 0, time.UTC), // 14:59 ICT
			want: true,
		},
		{
			name: "Monday 15:00 ICT is not trading",
			time: time.Date(2025, 7, 7, 8, 0, 0, 0, time.UTC), // 15:00 ICT
			want: false,
		},
		{
			name: "Monday 8:59 ICT is not trading",
			time: time.Date(2025, 7, 7, 1, 59, 0, 0, time.UTC), // 8:59 ICT
			want: false,
		},
		{
			name: "Saturday 10:00 ICT is not trading",
			time: time.Date(2025, 7, 12, 3, 0, 0, 0, time.UTC), // Saturday
			want: false,
		},
		{
			name: "Sunday 10:00 ICT is not trading",
			time: time.Date(2025, 7, 13, 3, 0, 0, 0, time.UTC), // Sunday
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTradingHours(tt.time)
			if got != tt.want {
				t.Errorf("IsTradingHours(%v) = %v, want %v", tt.time, got, tt.want)
			}
		})
	}
}

func TestScheduler_EvalInterval(t *testing.T) {
	s := NewScheduler(nil)

	// During trading hours: 5 min.
	s.nowFunc = func() time.Time {
		return time.Date(2025, 7, 7, 3, 0, 0, 0, time.UTC) // Mon 10:00 ICT
	}
	if got := s.currentEvalInterval(); got != 5*time.Minute {
		t.Errorf("trading hours interval = %v, want 5m", got)
	}

	// Outside trading hours: 30 min.
	s.nowFunc = func() time.Time {
		return time.Date(2025, 7, 7, 10, 0, 0, 0, time.UTC) // Mon 17:00 ICT
	}
	if got := s.currentEvalInterval(); got != 30*time.Minute {
		t.Errorf("off-hours interval = %v, want 30m", got)
	}
}

func TestScheduler_TimeUntilArchival(t *testing.T) {
	s := NewScheduler(nil)

	// At 20:00 ICT → 3 hours until 23:00.
	s.nowFunc = func() time.Time {
		return time.Date(2025, 7, 7, 13, 0, 0, 0, time.UTC) // 20:00 ICT
	}
	got := s.timeUntilArchival()
	if got != 3*time.Hour {
		t.Errorf("timeUntilArchival at 20:00 ICT = %v, want 3h", got)
	}

	// At 23:30 ICT → should schedule for next day (23.5 hours).
	s.nowFunc = func() time.Time {
		return time.Date(2025, 7, 7, 16, 30, 0, 0, time.UTC) // 23:30 ICT
	}
	got = s.timeUntilArchival()
	expected := 23*time.Hour + 30*time.Minute
	if got != expected {
		t.Errorf("timeUntilArchival at 23:30 ICT = %v, want %v", got, expected)
	}
}

func TestScheduler_RegisterAndRunJobs(t *testing.T) {
	s := NewScheduler(nil)

	var evalCount atomic.Int32
	var archCount atomic.Int32

	s.RegisterEvalJob("test-eval", func(_ context.Context) error {
		evalCount.Add(1)
		return nil
	})
	s.RegisterArchivalJob("test-arch", func(_ context.Context) error {
		archCount.Add(1)
		return nil
	})

	// Directly invoke runJobs to verify registration works.
	ctx := context.Background()
	s.runJobs(ctx, s.evalJobs, "eval")
	s.runJobs(ctx, s.archJobs, "archival")

	if evalCount.Load() != 1 {
		t.Errorf("eval job ran %d times, want 1", evalCount.Load())
	}
	if archCount.Load() != 1 {
		t.Errorf("archival job ran %d times, want 1", archCount.Load())
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s := NewScheduler(nil)

	var called atomic.Int32
	s.RegisterEvalJob("quick-job", func(_ context.Context) error {
		called.Add(1)
		return nil
	})

	s.Start()
	// Double-start should be a no-op.
	s.Start()

	// Give the eval loop a moment to tick (it sleeps for the interval first,
	// so we just verify start/stop doesn't panic or deadlock).
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	// Double-stop should be a no-op.
	s.Stop()
}
