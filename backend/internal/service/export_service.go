package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"myfi-backend/internal/model"
)

// ExportService generates CSV and PDF exports of transaction history,
// portfolio snapshots, P&L reports, and tax-friendly capital gains summaries.
// Requirement 28: Export and Reporting
// Requirement 29.7: Include both VND and USD value columns in all exports
type ExportService struct {
	fxRate float64 // USD/VND rate for dual-currency columns
}

// NewExportService creates a new ExportService with the given USD/VND rate.
func NewExportService(usdVndRate float64) *ExportService {
	if usdVndRate <= 0 {
		usdVndRate = model.FallbackUSDVND
	}
	return &ExportService{fxRate: usdVndRate}
}

// SetFXRate updates the USD/VND exchange rate used for conversions.
func (s *ExportService) SetFXRate(rate float64) {
	if rate > 0 {
		s.fxRate = rate
	}
}

// vndToUSD converts a VND amount to USD using the current rate.
func (s *ExportService) vndToUSD(vnd float64) float64 {
	if s.fxRate <= 0 {
		return 0
	}
	return vnd / s.fxRate
}

// ExportTransactionsCSV writes transaction history as CSV to the given writer.
// Supports date range filtering via from/to parameters (zero time means no bound).
// Requirement 28.1, 28.5, 29.7
func (s *ExportService) ExportTransactionsCSV(w io.Writer, transactions []model.Transaction, from, to time.Time) error {
	filtered := filterTransactions(transactions, from, to)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"Date", "Type", "Asset Type", "Symbol",
		"Quantity", "Unit Price (VND)", "Total Value (VND)",
		"Unit Price (USD)", "Total Value (USD)", "Notes",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, tx := range filtered {
		row := []string{
			tx.TransactionDate.Format("2006-01-02"),
			string(tx.TransactionType),
			string(tx.AssetType),
			tx.Symbol,
			fmt.Sprintf("%.4f", tx.Quantity),
			fmt.Sprintf("%.2f", tx.UnitPrice),
			fmt.Sprintf("%.2f", tx.TotalValue),
			fmt.Sprintf("%.4f", s.vndToUSD(tx.UnitPrice)),
			fmt.Sprintf("%.4f", s.vndToUSD(tx.TotalValue)),
			tx.Notes,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	return cw.Error()
}

// SnapshotRow represents a single holding in a portfolio snapshot export.
type SnapshotRow struct {
	AssetType    model.AssetType
	Symbol       string
	Quantity     float64
	AverageCost  float64
	CurrentPrice float64
	MarketValue  float64
	UnrealizedPL float64
}

// ExportSnapshotCSV writes a portfolio snapshot as CSV to the given writer.
// Requirement 28.2, 29.7
func (s *ExportService) ExportSnapshotCSV(w io.Writer, holdings []SnapshotRow) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"Asset Type", "Symbol", "Quantity",
		"Avg Cost (VND)", "Current Price (VND)", "Market Value (VND)", "Unrealized P&L (VND)",
		"Avg Cost (USD)", "Current Price (USD)", "Market Value (USD)", "Unrealized P&L (USD)",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, h := range holdings {
		row := []string{
			string(h.AssetType),
			h.Symbol,
			fmt.Sprintf("%.4f", h.Quantity),
			fmt.Sprintf("%.2f", h.AverageCost),
			fmt.Sprintf("%.2f", h.CurrentPrice),
			fmt.Sprintf("%.2f", h.MarketValue),
			fmt.Sprintf("%.2f", h.UnrealizedPL),
			fmt.Sprintf("%.4f", s.vndToUSD(h.AverageCost)),
			fmt.Sprintf("%.4f", s.vndToUSD(h.CurrentPrice)),
			fmt.Sprintf("%.4f", s.vndToUSD(h.MarketValue)),
			fmt.Sprintf("%.4f", s.vndToUSD(h.UnrealizedPL)),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	return cw.Error()
}

// TaxGainEntry represents a single realized gain/loss for tax reporting.
type TaxGainEntry struct {
	AssetType    model.AssetType
	Symbol       string
	SellDate     time.Time
	Quantity     float64
	CostBasis    float64 // total cost in VND
	SaleProceeds float64 // total proceeds in VND
	RealizedPL   float64 // gain or loss in VND
}

// TaxReport groups capital gains by asset type and year.
type TaxReport struct {
	Year      int
	AssetType model.AssetType
	Entries   []TaxGainEntry
	TotalGain float64
	TotalLoss float64
	NetPL     float64
}

// ComputeTaxEntries derives realized gain/loss entries from sell transactions
// paired with average cost data. Each sell transaction produces one TaxGainEntry.
func ComputeTaxEntries(transactions []model.Transaction, avgCosts map[string]float64) []TaxGainEntry {
	var entries []TaxGainEntry
	for _, tx := range transactions {
		if tx.TransactionType != model.Sell {
			continue
		}
		avgCost, ok := avgCosts[tx.Symbol]
		if !ok {
			avgCost = 0
		}
		costBasis := avgCost * tx.Quantity
		saleProceeds := tx.TotalValue
		if saleProceeds == 0 {
			saleProceeds = tx.UnitPrice * tx.Quantity
		}
		entries = append(entries, TaxGainEntry{
			AssetType:    tx.AssetType,
			Symbol:       tx.Symbol,
			SellDate:     tx.TransactionDate,
			Quantity:     tx.Quantity,
			CostBasis:    costBasis,
			SaleProceeds: saleProceeds,
			RealizedPL:   saleProceeds - costBasis,
		})
	}
	return entries
}

// GroupTaxEntriesByYearAndType groups tax entries into TaxReport summaries
// by year and asset type.
func GroupTaxEntriesByYearAndType(entries []TaxGainEntry) []TaxReport {
	type key struct {
		Year      int
		AssetType model.AssetType
	}
	groups := make(map[key]*TaxReport)

	for _, e := range entries {
		k := key{Year: e.SellDate.Year(), AssetType: e.AssetType}
		rpt, ok := groups[k]
		if !ok {
			rpt = &TaxReport{Year: k.Year, AssetType: k.AssetType}
			groups[k] = rpt
		}
		rpt.Entries = append(rpt.Entries, e)
		if e.RealizedPL >= 0 {
			rpt.TotalGain += e.RealizedPL
		} else {
			rpt.TotalLoss += e.RealizedPL
		}
		rpt.NetPL += e.RealizedPL
	}

	result := make([]TaxReport, 0, len(groups))
	for _, rpt := range groups {
		result = append(result, *rpt)
	}
	return result
}

// ExportTaxReportCSV writes a tax report (capital gains by asset type and year) as CSV.
// Requirement 28.4, 28.5, 29.7
func (s *ExportService) ExportTaxReportCSV(w io.Writer, transactions []model.Transaction, avgCosts map[string]float64, from, to time.Time) error {
	filtered := filterTransactions(transactions, from, to)
	entries := ComputeTaxEntries(filtered, avgCosts)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"Year", "Asset Type", "Symbol", "Sell Date",
		"Quantity", "Cost Basis (VND)", "Sale Proceeds (VND)", "Realized P&L (VND)",
		"Cost Basis (USD)", "Sale Proceeds (USD)", "Realized P&L (USD)",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, e := range entries {
		row := []string{
			fmt.Sprintf("%d", e.SellDate.Year()),
			string(e.AssetType),
			e.Symbol,
			e.SellDate.Format("2006-01-02"),
			fmt.Sprintf("%.4f", e.Quantity),
			fmt.Sprintf("%.2f", e.CostBasis),
			fmt.Sprintf("%.2f", e.SaleProceeds),
			fmt.Sprintf("%.2f", e.RealizedPL),
			fmt.Sprintf("%.4f", s.vndToUSD(e.CostBasis)),
			fmt.Sprintf("%.4f", s.vndToUSD(e.SaleProceeds)),
			fmt.Sprintf("%.4f", s.vndToUSD(e.RealizedPL)),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	return cw.Error()
}

// ExportPortfolioReportPDF writes a structured text-based portfolio report
// that can be rendered as PDF. Contains NAV summary, allocation breakdown,
// and P&L by holding with both VND and USD columns.
// Requirement 28.3, 29.7
func (s *ExportService) ExportPortfolioReportPDF(w io.Writer, nav float64, allocation map[model.AssetType]float64, holdings []SnapshotRow) error {
	buf := &bytes.Buffer{}

	// Title
	fmt.Fprintf(buf, "PORTFOLIO REPORT\n")
	fmt.Fprintf(buf, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(buf, "USD/VND Rate: %.0f\n", s.fxRate)
	fmt.Fprintf(buf, "=====================================\n\n")

	// NAV Summary
	fmt.Fprintf(buf, "NAV SUMMARY\n")
	fmt.Fprintf(buf, "-------------------------------------\n")
	fmt.Fprintf(buf, "Total NAV (VND): %20.2f\n", nav)
	fmt.Fprintf(buf, "Total NAV (USD): %20.4f\n", s.vndToUSD(nav))
	fmt.Fprintf(buf, "\n")

	// Allocation
	fmt.Fprintf(buf, "ASSET ALLOCATION\n")
	fmt.Fprintf(buf, "-------------------------------------\n")
	if nav > 0 {
		for at, val := range allocation {
			pct := (val / nav) * 100
			fmt.Fprintf(buf, "%-12s %16.2f VND (%5.1f%%)\n", at, val, pct)
		}
	}
	fmt.Fprintf(buf, "\n")

	// Holdings P&L
	fmt.Fprintf(buf, "HOLDINGS P&L BREAKDOWN\n")
	fmt.Fprintf(buf, "-------------------------------------\n")
	fmt.Fprintf(buf, "%-8s %-10s %12s %12s %14s %14s\n",
		"Type", "Symbol", "Mkt Val VND", "Mkt Val USD", "Unreal PL VND", "Unreal PL USD")
	for _, h := range holdings {
		fmt.Fprintf(buf, "%-8s %-10s %12.2f %12.4f %14.2f %14.4f\n",
			h.AssetType, h.Symbol,
			h.MarketValue, s.vndToUSD(h.MarketValue),
			h.UnrealizedPL, s.vndToUSD(h.UnrealizedPL))
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// ExportTransactionsCSVBytes is a convenience method that returns CSV bytes.
func (s *ExportService) ExportTransactionsCSVBytes(transactions []model.Transaction, from, to time.Time) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.ExportTransactionsCSV(&buf, transactions, from, to); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportSnapshotCSVBytes is a convenience method that returns CSV bytes.
func (s *ExportService) ExportSnapshotCSVBytes(holdings []SnapshotRow) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.ExportSnapshotCSV(&buf, holdings); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportTaxReportCSVBytes is a convenience method that returns CSV bytes.
func (s *ExportService) ExportTaxReportCSVBytes(transactions []model.Transaction, avgCosts map[string]float64, from, to time.Time) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.ExportTaxReportCSV(&buf, transactions, avgCosts, from, to); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportPortfolioReportPDFBytes is a convenience method that returns report bytes.
func (s *ExportService) ExportPortfolioReportPDFBytes(nav float64, allocation map[model.AssetType]float64, holdings []SnapshotRow) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.ExportPortfolioReportPDF(&buf, nav, allocation, holdings); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// filterTransactions returns transactions within the [from, to] date range.
// Zero-value from/to means no bound on that side.
func filterTransactions(txns []model.Transaction, from, to time.Time) []model.Transaction {
	if from.IsZero() && to.IsZero() {
		return txns
	}
	var result []model.Transaction
	for _, tx := range txns {
		if !from.IsZero() && tx.TransactionDate.Before(from) {
			continue
		}
		if !to.IsZero() && tx.TransactionDate.After(to) {
			continue
		}
		result = append(result, tx)
	}
	return result
}
