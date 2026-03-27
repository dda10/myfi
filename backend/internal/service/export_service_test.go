package service

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"myfi-backend/internal/model"
)

func TestExportTransactionsCSV_BasicOutput(t *testing.T) {
	svc := NewExportService(25000)
	txns := []model.Transaction{
		{
			TransactionDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			TransactionType: model.Buy,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        100,
			UnitPrice:       85000,
			TotalValue:      8500000,
			Notes:           "initial buy",
		},
		{
			TransactionDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			TransactionType: model.Sell,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        50,
			UnitPrice:       90000,
			TotalValue:      4500000,
			Notes:           "",
		},
	}

	var buf bytes.Buffer
	err := svc.ExportTransactionsCSV(&buf, txns, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("ExportTransactionsCSV failed: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// header + 2 data rows
	if len(records) != 3 {
		t.Fatalf("expected 3 rows (header + 2 data), got %d", len(records))
	}

	// Verify header has VND and USD columns
	header := records[0]
	if header[5] != "Unit Price (VND)" {
		t.Errorf("expected 'Unit Price (VND)', got %q", header[5])
	}
	if header[7] != "Unit Price (USD)" {
		t.Errorf("expected 'Unit Price (USD)', got %q", header[7])
	}

	// Verify first data row
	row := records[1]
	if row[0] != "2024-03-15" {
		t.Errorf("expected date '2024-03-15', got %q", row[0])
	}
	if row[1] != "buy" {
		t.Errorf("expected type 'buy', got %q", row[1])
	}
	if row[3] != "FPT" {
		t.Errorf("expected symbol 'FPT', got %q", row[3])
	}
	// USD unit price: 85000 / 25000 = 3.4
	if row[7] != "3.4000" {
		t.Errorf("expected USD unit price '3.4000', got %q", row[7])
	}
}

func TestExportTransactionsCSV_DateRangeFilter(t *testing.T) {
	svc := NewExportService(25000)
	txns := []model.Transaction{
		{
			TransactionDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			TransactionType: model.Buy,
			AssetType:       model.VNStock,
			Symbol:          "VNM",
			Quantity:        10,
			UnitPrice:       70000,
			TotalValue:      700000,
		},
		{
			TransactionDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			TransactionType: model.Sell,
			AssetType:       model.VNStock,
			Symbol:          "VNM",
			Quantity:        5,
			UnitPrice:       75000,
			TotalValue:      375000,
		},
		{
			TransactionDate: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			TransactionType: model.Buy,
			AssetType:       model.Crypto,
			Symbol:          "BTC",
			Quantity:        0.01,
			UnitPrice:       1500000000,
			TotalValue:      15000000,
		},
	}

	from := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	err := svc.ExportTransactionsCSV(&buf, txns, from, to)
	if err != nil {
		t.Fatalf("ExportTransactionsCSV failed: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// header + 1 data row (only the June transaction is in range)
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1 filtered), got %d", len(records))
	}
	if records[1][3] != "VNM" {
		t.Errorf("expected symbol 'VNM', got %q", records[1][3])
	}
}

func TestExportSnapshotCSV(t *testing.T) {
	svc := NewExportService(25400)
	holdings := []SnapshotRow{
		{
			AssetType:    model.VNStock,
			Symbol:       "SSI",
			Quantity:     200,
			AverageCost:  30000,
			CurrentPrice: 35000,
			MarketValue:  7000000,
			UnrealizedPL: 1000000,
		},
		{
			AssetType:    model.Gold,
			Symbol:       "SJC",
			Quantity:     1,
			AverageCost:  70000000,
			CurrentPrice: 72000000,
			MarketValue:  72000000,
			UnrealizedPL: 2000000,
		},
	}

	var buf bytes.Buffer
	err := svc.ExportSnapshotCSV(&buf, holdings)
	if err != nil {
		t.Fatalf("ExportSnapshotCSV failed: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 rows (header + 2), got %d", len(records))
	}

	// Verify header includes both VND and USD columns
	header := records[0]
	hasVND := false
	hasUSD := false
	for _, h := range header {
		if strings.Contains(h, "VND") {
			hasVND = true
		}
		if strings.Contains(h, "USD") {
			hasUSD = true
		}
	}
	if !hasVND || !hasUSD {
		t.Error("header should contain both VND and USD columns")
	}

	// Verify first row symbol
	if records[1][1] != "SSI" {
		t.Errorf("expected symbol 'SSI', got %q", records[1][1])
	}
}

func TestComputeTaxEntries(t *testing.T) {
	txns := []model.Transaction{
		{
			TransactionType: model.Buy,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        100,
			UnitPrice:       80000,
			TotalValue:      8000000,
			TransactionDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			TransactionType: model.Sell,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        50,
			UnitPrice:       90000,
			TotalValue:      4500000,
			TransactionDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			TransactionType: model.Sell,
			AssetType:       model.Crypto,
			Symbol:          "BTC",
			Quantity:        0.5,
			UnitPrice:       1600000000,
			TotalValue:      800000000,
			TransactionDate: time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	avgCosts := map[string]float64{
		"FPT": 80000,
		"BTC": 1500000000,
	}

	entries := ComputeTaxEntries(txns, avgCosts)

	if len(entries) != 2 {
		t.Fatalf("expected 2 tax entries (sells only), got %d", len(entries))
	}

	// FPT: sold 50 @ 90000, cost 50*80000=4000000, proceeds=4500000, PL=500000
	fpt := entries[0]
	if fpt.Symbol != "FPT" {
		t.Errorf("expected FPT, got %s", fpt.Symbol)
	}
	if fpt.CostBasis != 4000000 {
		t.Errorf("expected cost basis 4000000, got %.2f", fpt.CostBasis)
	}
	if fpt.RealizedPL != 500000 {
		t.Errorf("expected realized PL 500000, got %.2f", fpt.RealizedPL)
	}

	// BTC: sold 0.5 @ 1.6B, cost 0.5*1.5B=750M, proceeds=800M, PL=50M
	btc := entries[1]
	if btc.Symbol != "BTC" {
		t.Errorf("expected BTC, got %s", btc.Symbol)
	}
	expectedCost := 0.5 * 1500000000.0
	if btc.CostBasis != expectedCost {
		t.Errorf("expected cost basis %.2f, got %.2f", expectedCost, btc.CostBasis)
	}
	expectedPL := 800000000.0 - expectedCost
	if btc.RealizedPL != expectedPL {
		t.Errorf("expected realized PL %.2f, got %.2f", expectedPL, btc.RealizedPL)
	}
}

func TestGroupTaxEntriesByYearAndType(t *testing.T) {
	entries := []TaxGainEntry{
		{
			AssetType:    model.VNStock,
			Symbol:       "FPT",
			SellDate:     time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			Quantity:     50,
			CostBasis:    4000000,
			SaleProceeds: 4500000,
			RealizedPL:   500000,
		},
		{
			AssetType:    model.VNStock,
			Symbol:       "VNM",
			SellDate:     time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
			Quantity:     30,
			CostBasis:    2100000,
			SaleProceeds: 1800000,
			RealizedPL:   -300000,
		},
		{
			AssetType:    model.Crypto,
			Symbol:       "BTC",
			SellDate:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			Quantity:     0.5,
			CostBasis:    750000000,
			SaleProceeds: 800000000,
			RealizedPL:   50000000,
		},
		{
			AssetType:    model.VNStock,
			Symbol:       "SSI",
			SellDate:     time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			Quantity:     100,
			CostBasis:    3000000,
			SaleProceeds: 3500000,
			RealizedPL:   500000,
		},
	}

	reports := GroupTaxEntriesByYearAndType(entries)

	// Should have 3 groups: 2024/vn_stock, 2024/crypto, 2023/vn_stock
	if len(reports) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(reports))
	}

	// Find the 2024 VN stock group
	var found bool
	for _, rpt := range reports {
		if rpt.Year == 2024 && rpt.AssetType == model.VNStock {
			found = true
			if len(rpt.Entries) != 2 {
				t.Errorf("expected 2 entries in 2024/vn_stock, got %d", len(rpt.Entries))
			}
			if rpt.TotalGain != 500000 {
				t.Errorf("expected total gain 500000, got %.2f", rpt.TotalGain)
			}
			if rpt.TotalLoss != -300000 {
				t.Errorf("expected total loss -300000, got %.2f", rpt.TotalLoss)
			}
			if rpt.NetPL != 200000 {
				t.Errorf("expected net PL 200000, got %.2f", rpt.NetPL)
			}
		}
	}
	if !found {
		t.Error("2024/vn_stock group not found")
	}
}

func TestExportTaxReportCSV(t *testing.T) {
	svc := NewExportService(25000)
	txns := []model.Transaction{
		{
			TransactionType: model.Sell,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        50,
			UnitPrice:       90000,
			TotalValue:      4500000,
			TransactionDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	avgCosts := map[string]float64{"FPT": 80000}

	var buf bytes.Buffer
	err := svc.ExportTaxReportCSV(&buf, txns, avgCosts, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("ExportTaxReportCSV failed: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1), got %d", len(records))
	}

	// Verify year column
	if records[1][0] != "2024" {
		t.Errorf("expected year '2024', got %q", records[1][0])
	}

	// Verify USD columns exist in header
	header := records[0]
	hasUSD := false
	for _, h := range header {
		if strings.Contains(h, "USD") {
			hasUSD = true
			break
		}
	}
	if !hasUSD {
		t.Error("tax report header should contain USD columns")
	}
}

func TestExportPortfolioReportPDF(t *testing.T) {
	svc := NewExportService(25000)
	allocation := map[model.AssetType]float64{
		model.VNStock: 7000000,
		model.Gold:    72000000,
	}
	holdings := []SnapshotRow{
		{
			AssetType:    model.VNStock,
			Symbol:       "SSI",
			Quantity:     200,
			AverageCost:  30000,
			CurrentPrice: 35000,
			MarketValue:  7000000,
			UnrealizedPL: 1000000,
		},
		{
			AssetType:    model.Gold,
			Symbol:       "SJC",
			Quantity:     1,
			AverageCost:  70000000,
			CurrentPrice: 72000000,
			MarketValue:  72000000,
			UnrealizedPL: 2000000,
		},
	}

	nav := 79000000.0
	var buf bytes.Buffer
	err := svc.ExportPortfolioReportPDF(&buf, nav, allocation, holdings)
	if err != nil {
		t.Fatalf("ExportPortfolioReportPDF failed: %v", err)
	}

	content := buf.String()

	// Verify key sections exist
	if !strings.Contains(content, "PORTFOLIO REPORT") {
		t.Error("report should contain 'PORTFOLIO REPORT' title")
	}
	if !strings.Contains(content, "NAV SUMMARY") {
		t.Error("report should contain 'NAV SUMMARY' section")
	}
	if !strings.Contains(content, "ASSET ALLOCATION") {
		t.Error("report should contain 'ASSET ALLOCATION' section")
	}
	if !strings.Contains(content, "HOLDINGS P&L BREAKDOWN") {
		t.Error("report should contain 'HOLDINGS P&L BREAKDOWN' section")
	}
	if !strings.Contains(content, "USD") {
		t.Error("report should contain USD values")
	}
	if !strings.Contains(content, "SSI") {
		t.Error("report should contain holding symbol SSI")
	}
	if !strings.Contains(content, "SJC") {
		t.Error("report should contain holding symbol SJC")
	}
}

func TestNewExportService_FallbackRate(t *testing.T) {
	// Zero rate should use fallback
	svc := NewExportService(0)
	if svc.fxRate != model.FallbackUSDVND {
		t.Errorf("expected fallback rate %.0f, got %.0f", model.FallbackUSDVND, svc.fxRate)
	}

	// Negative rate should use fallback
	svc = NewExportService(-100)
	if svc.fxRate != model.FallbackUSDVND {
		t.Errorf("expected fallback rate %.0f, got %.0f", model.FallbackUSDVND, svc.fxRate)
	}

	// Valid rate should be used
	svc = NewExportService(25000)
	if svc.fxRate != 25000 {
		t.Errorf("expected rate 25000, got %.0f", svc.fxRate)
	}
}

func TestSetFXRate(t *testing.T) {
	svc := NewExportService(25000)
	svc.SetFXRate(26000)
	if svc.fxRate != 26000 {
		t.Errorf("expected rate 26000, got %.0f", svc.fxRate)
	}

	// Zero should not change rate
	svc.SetFXRate(0)
	if svc.fxRate != 26000 {
		t.Errorf("expected rate to remain 26000, got %.0f", svc.fxRate)
	}
}

func TestFilterTransactions(t *testing.T) {
	txns := []model.Transaction{
		{TransactionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Symbol: "A"},
		{TransactionDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Symbol: "B"},
		{TransactionDate: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), Symbol: "C"},
	}

	// No filter
	result := filterTransactions(txns, time.Time{}, time.Time{})
	if len(result) != 3 {
		t.Errorf("no filter: expected 3, got %d", len(result))
	}

	// From only
	from := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	result = filterTransactions(txns, from, time.Time{})
	if len(result) != 2 {
		t.Errorf("from filter: expected 2, got %d", len(result))
	}

	// To only
	to := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	result = filterTransactions(txns, time.Time{}, to)
	if len(result) != 2 {
		t.Errorf("to filter: expected 2, got %d", len(result))
	}

	// Both
	result = filterTransactions(txns, from, to)
	if len(result) != 1 {
		t.Errorf("both filters: expected 1, got %d", len(result))
	}
	if result[0].Symbol != "B" {
		t.Errorf("expected symbol 'B', got %q", result[0].Symbol)
	}
}

func TestExportTransactionsCSV_EmptyList(t *testing.T) {
	svc := NewExportService(25000)
	var buf bytes.Buffer
	err := svc.ExportTransactionsCSV(&buf, nil, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("ExportTransactionsCSV failed on empty: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}
	// Should have header only
	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

func TestComputeTaxEntries_NoSells(t *testing.T) {
	txns := []model.Transaction{
		{
			TransactionType: model.Buy,
			AssetType:       model.VNStock,
			Symbol:          "FPT",
			Quantity:        100,
			UnitPrice:       80000,
			TotalValue:      8000000,
			TransactionDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	entries := ComputeTaxEntries(txns, map[string]float64{"FPT": 80000})
	if len(entries) != 0 {
		t.Errorf("expected 0 tax entries for buy-only, got %d", len(entries))
	}
}

func TestComputeTaxEntries_MissingAvgCost(t *testing.T) {
	txns := []model.Transaction{
		{
			TransactionType: model.Sell,
			AssetType:       model.VNStock,
			Symbol:          "UNKNOWN",
			Quantity:        10,
			UnitPrice:       50000,
			TotalValue:      500000,
			TransactionDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	// No avg cost for UNKNOWN — should default to 0 cost basis
	entries := ComputeTaxEntries(txns, map[string]float64{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].CostBasis != 0 {
		t.Errorf("expected cost basis 0 for missing avg cost, got %.2f", entries[0].CostBasis)
	}
	if entries[0].RealizedPL != 500000 {
		t.Errorf("expected realized PL 500000, got %.2f", entries[0].RealizedPL)
	}
}
