package ranking

import (
	"context"
	"testing"
	"time"
)

func TestRecommendationTracker_LogAndRetrieve(t *testing.T) {
	tracker := NewRecommendationTracker(nil)

	rec := AssetRecommendation{
		Symbol:         "FPT",
		Action:         "buy",
		PositionSize:   5.0,
		RiskAssessment: "medium",
		Reasoning:      "Strong technical setup",
	}

	id, err := tracker.LogRecommendation(context.Background(), rec, 75)
	if err != nil {
		t.Fatalf("LogRecommendation failed: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id=1, got %d", id)
	}

	record := tracker.GetRecordByID(id)
	if record == nil {
		t.Fatal("GetRecordByID returned nil")
	}
	if record.Symbol != "FPT" {
		t.Errorf("expected symbol=FPT, got %s", record.Symbol)
	}
	if record.Action != ActionBuy {
		t.Errorf("expected action=buy, got %s", record.Action)
	}
	if record.ConfidenceScore != 75 {
		t.Errorf("expected confidence=75, got %d", record.ConfidenceScore)
	}
}

func TestRecommendationTracker_FilterBySymbol(t *testing.T) {
	tracker := NewRecommendationTracker(nil)
	ctx := context.Background()

	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "FPT", Action: "buy"}, 70)
	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "VNM", Action: "sell"}, 60)
	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "FPT", Action: "hold"}, 50)

	results := tracker.GetRecommendations(RecommendationFilter{Symbol: "FPT"})
	if len(results) != 2 {
		t.Errorf("expected 2 FPT records, got %d", len(results))
	}
}

func TestRecommendationTracker_FilterByAction(t *testing.T) {
	tracker := NewRecommendationTracker(nil)
	ctx := context.Background()

	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "FPT", Action: "buy"}, 70)
	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "VNM", Action: "buy"}, 60)
	tracker.LogRecommendation(ctx, AssetRecommendation{Symbol: "SSI", Action: "sell"}, 80)

	results := tracker.GetRecommendations(RecommendationFilter{Action: ActionBuy})
	if len(results) != 2 {
		t.Errorf("expected 2 buy records, got %d", len(results))
	}
}

func TestRecommendationTracker_AccuracyCalculation(t *testing.T) {
	tracker := NewRecommendationTracker(nil)

	tracker.mu.Lock()
	now := time.Now()
	ret1 := 0.05
	ret2 := -0.03
	ret3 := 0.08

	tracker.records = []RecommendationRecord{
		{ID: 1, Symbol: "FPT", Action: ActionBuy, ConfidenceScore: 75, PriceAtSignal: 100, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret1},
		{ID: 2, Symbol: "VNM", Action: ActionBuy, ConfidenceScore: 45, PriceAtSignal: 80, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret2},
		{ID: 3, Symbol: "SSI", Action: ActionBuy, ConfidenceScore: 80, PriceAtSignal: 50, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret3},
	}
	tracker.nextID = 4
	tracker.mu.Unlock()

	acc := tracker.GetAccuracyByAction(ActionBuy)

	if acc.TotalCount != 3 {
		t.Errorf("expected TotalCount=3, got %d", acc.TotalCount)
	}
	if acc.WinCount7Day != 2 {
		t.Errorf("expected WinCount7Day=2, got %d", acc.WinCount7Day)
	}

	expectedWinRate := 2.0 / 3.0
	if abs(acc.WinRate7Day-expectedWinRate) > 0.001 {
		t.Errorf("expected WinRate7Day=%.3f, got %.3f", expectedWinRate, acc.WinRate7Day)
	}

	expectedAvgReturn := (0.05 - 0.03 + 0.08) / 3.0
	if abs(acc.AvgReturn7Day-expectedAvgReturn) > 0.001 {
		t.Errorf("expected AvgReturn7Day=%.4f, got %.4f", expectedAvgReturn, acc.AvgReturn7Day)
	}
}

func TestRecommendationTracker_SellWinLogic(t *testing.T) {
	tracker := NewRecommendationTracker(nil)

	tracker.mu.Lock()
	now := time.Now()
	ret1 := -0.05
	ret2 := 0.03

	tracker.records = []RecommendationRecord{
		{ID: 1, Symbol: "FPT", Action: ActionSell, ConfidenceScore: 70, PriceAtSignal: 100, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret1},
		{ID: 2, Symbol: "VNM", Action: ActionSell, ConfidenceScore: 60, PriceAtSignal: 80, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret2},
	}
	tracker.nextID = 3
	tracker.mu.Unlock()

	acc := tracker.GetAccuracyByAction(ActionSell)

	if acc.WinCount7Day != 1 {
		t.Errorf("expected WinCount7Day=1 for sell, got %d", acc.WinCount7Day)
	}
	if acc.WinRate7Day != 0.5 {
		t.Errorf("expected WinRate7Day=0.5, got %.2f", acc.WinRate7Day)
	}
}

func TestRecommendationTracker_Summary(t *testing.T) {
	tracker := NewRecommendationTracker(nil)

	tracker.mu.Lock()
	now := time.Now()
	ret1 := 0.10
	ret2 := -0.02
	ret3 := 0.05

	tracker.records = []RecommendationRecord{
		{ID: 1, Symbol: "FPT", Action: ActionBuy, ConfidenceScore: 80, PriceAtSignal: 100, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret1},
		{ID: 2, Symbol: "VNM", Action: ActionBuy, ConfidenceScore: 60, PriceAtSignal: 80, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret2},
		{ID: 3, Symbol: "FPT", Action: ActionBuy, ConfidenceScore: 70, PriceAtSignal: 105, CreatedAt: now.Add(-8 * 24 * time.Hour), Return7Day: &ret3},
	}
	tracker.nextID = 4
	tracker.mu.Unlock()

	summary := tracker.GetSummary()

	if summary.TotalRecommendations != 3 {
		t.Errorf("expected TotalRecommendations=3, got %d", summary.TotalRecommendations)
	}
	if summary.BestPerformingSymbol != "FPT" {
		t.Errorf("expected BestPerformingSymbol=FPT, got %s", summary.BestPerformingSymbol)
	}
	if summary.WorstPerformingSymbol != "VNM" {
		t.Errorf("expected WorstPerformingSymbol=VNM, got %s", summary.WorstPerformingSymbol)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
