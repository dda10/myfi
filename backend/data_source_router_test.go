package main

import (
	"context"
	"testing"
	"time"

	"github.com/dda10/vnstock-go"
)

func TestDataSourceRouter_Creation(t *testing.T) {
	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	if router == nil {
		t.Fatal("Router should not be nil")
	}

	if router.vciClient == nil {
		t.Error("VCI client should not be nil")
	}

	if router.kbsClient == nil {
		t.Error("KBS client should not be nil")
	}
}

func TestDataSourceRouter_SourcePreferences(t *testing.T) {
	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	prefs := router.GetSourcePreferences()

	// Verify all 12 data categories have preferences
	expectedCategories := []DataCategory{
		PriceQuotes,
		OHLCVHistory,
		IntradayData,
		OrderBook,
		CompanyOverview,
		Shareholders,
		Officers,
		News,
		IncomeStatement,
		BalanceSheet,
		CashFlow,
		FinancialRatios,
	}

	if len(prefs) != len(expectedCategories) {
		t.Errorf("Expected %d preferences, got %d", len(expectedCategories), len(prefs))
	}

	for _, cat := range expectedCategories {
		pref, exists := prefs[cat]
		if !exists {
			t.Errorf("Missing preference for category: %s", cat)
			continue
		}

		if pref.Primary != "VCI" && pref.Primary != "KBS" {
			t.Errorf("Invalid primary source for %s: %s", cat, pref.Primary)
		}

		if pref.Fallback != "VCI" && pref.Fallback != "KBS" {
			t.Errorf("Invalid fallback source for %s: %s", cat, pref.Fallback)
		}

		if pref.Primary == pref.Fallback {
			t.Errorf("Primary and fallback should be different for %s", cat)
		}
	}
}

func TestDataSourceRouter_SelectSource(t *testing.T) {
	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	// Test selecting primary source for PriceQuotes
	client, source := router.selectSource(PriceQuotes)
	if client == nil {
		t.Error("Selected client should not be nil")
	}
	if source != "VCI" {
		t.Errorf("Expected VCI as primary source for PriceQuotes, got %s", source)
	}

	// Test fallback source
	fallbackClient, fallbackSource := router.getFallbackSource(PriceQuotes)
	if fallbackClient == nil {
		t.Error("Fallback client should not be nil")
	}
	if fallbackSource != "KBS" {
		t.Errorf("Expected KBS as fallback source for PriceQuotes, got %s", fallbackSource)
	}
}

func TestDataSourceRouter_IsEmptyData(t *testing.T) {
	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	tests := []struct {
		name     string
		quotes   []vnstock.Quote
		expected bool
	}{
		{
			name:     "Empty slice",
			quotes:   []vnstock.Quote{},
			expected: true,
		},
		{
			name: "All zero prices",
			quotes: []vnstock.Quote{
				{Symbol: "SSI", Close: 0},
				{Symbol: "FPT", Close: 0},
			},
			expected: true,
		},
		{
			name: "Valid data",
			quotes: []vnstock.Quote{
				{Symbol: "SSI", Close: 45000},
				{Symbol: "FPT", Close: 120000},
			},
			expected: false,
		},
		{
			name: "Mixed data",
			quotes: []vnstock.Quote{
				{Symbol: "SSI", Close: 0},
				{Symbol: "FPT", Close: 120000},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.isEmptyData(tt.quotes)
			if result != tt.expected {
				t.Errorf("isEmptyData() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDataSourceRouter_FetchRealTimeQuotes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	symbols := []string{"SSI", "FPT"}
	quotes, source, err := router.FetchRealTimeQuotes(ctx, symbols)

	if err != nil {
		t.Logf("FetchRealTimeQuotes returned error (may be expected if API is unavailable): %v", err)
		return
	}

	if len(quotes) == 0 {
		t.Log("No quotes returned (may be expected if market is closed)")
		return
	}

	t.Logf("Successfully fetched %d quotes from source: %s", len(quotes), source)

	for _, q := range quotes {
		t.Logf("Quote: %s - Close: %.2f", q.Symbol, q.Close)
	}
}

func TestDataSourceRouter_FetchQuoteHistory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router, err := NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create DataSourceRouter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	end := time.Now()
	start := end.AddDate(0, 0, -30) // Last 30 days

	req := vnstock.QuoteHistoryRequest{
		Symbol:   "SSI",
		Start:    start,
		End:      end,
		Interval: "1d",
	}

	history, source, err := router.FetchQuoteHistory(ctx, req)

	if err != nil {
		t.Logf("FetchQuoteHistory returned error (may be expected if API is unavailable): %v", err)
		return
	}

	if len(history) == 0 {
		t.Log("No history returned (may be expected)")
		return
	}

	t.Logf("Successfully fetched %d historical records from source: %s", len(history), source)
	t.Logf("First record: Close: %.2f", history[0].Close)
	t.Logf("Last record: Close: %.2f", history[len(history)-1].Close)
}
