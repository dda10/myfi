package analyst

// ---------------------------------------------------------------------------
// AnalystIQService — aggregate analyst reports and track accuracy
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 24.1 Aggregate analyst reports from Vietnamese brokerages
//   - 24.2 Track accuracy at 1m, 3m, 6m intervals
//   - 24.3 Compute consensus recommendation + target price

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// AnalystIQService aggregates analyst reports, tracks accuracy, and computes consensus.
type AnalystIQService struct {
	db *sql.DB
}

// NewAnalystIQService creates a new AnalystIQService.
func NewAnalystIQService(db *sql.DB) *AnalystIQService {
	return &AnalystIQService{db: db}
}

// AddReport persists a new analyst report.
func (s *AnalystIQService) AddReport(ctx context.Context, report *AnalystReport) error {
	if report.Symbol == "" || report.Analyst == "" {
		return fmt.Errorf("symbol and analyst are required")
	}

	query := `INSERT INTO analyst_reports (symbol, analyst, brokerage, recommendation,
		target_price, published_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := s.db.QueryRowContext(ctx, query,
		report.Symbol, report.Analyst, report.Brokerage,
		report.Recommendation, report.TargetPrice, report.PublishedAt,
	).Scan(&report.ID)
	if err != nil {
		return fmt.Errorf("failed to add analyst report: %w", err)
	}

	log.Printf("[AnalystIQ] Added report: %s by %s (%s) target=%.0f",
		report.Symbol, report.Analyst, report.Recommendation, report.TargetPrice)
	return nil
}

// GetReports returns analyst reports for a symbol, ordered by most recent.
func (s *AnalystIQService) GetReports(ctx context.Context, symbol string, limit int) ([]AnalystReport, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `SELECT id, symbol, analyst, brokerage, recommendation, target_price,
		published_at, accuracy_1m, accuracy_3m, accuracy_6m
		FROM analyst_reports WHERE symbol = $1
		ORDER BY published_at DESC LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query analyst reports: %w", err)
	}
	defer rows.Close()

	var reports []AnalystReport
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// GetConsensus computes the consensus recommendation for a symbol.
func (s *AnalystIQService) GetConsensus(ctx context.Context, symbol string) (*ConsensusRecommendation, error) {
	query := `SELECT recommendation, target_price, accuracy_3m
		FROM analyst_reports
		WHERE symbol = $1 AND published_at > $2
		ORDER BY published_at DESC`

	// Only consider reports from the last 6 months
	cutoff := time.Now().AddDate(0, -6, 0)
	rows, err := s.db.QueryContext(ctx, query, symbol, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports for consensus: %w", err)
	}
	defer rows.Close()

	consensus := &ConsensusRecommendation{Symbol: symbol}
	var totalTarget, highTarget, lowTarget float64
	var totalAccuracy float64
	var accCount int
	first := true

	for rows.Next() {
		var rec string
		var target float64
		var acc3m sql.NullFloat64

		if err := rows.Scan(&rec, &target, &acc3m); err != nil {
			continue
		}

		consensus.TotalAnalysts++
		totalTarget += target

		if first || target > highTarget {
			highTarget = target
		}
		if first || target < lowTarget {
			lowTarget = target
		}
		first = false

		switch normalizeRecommendation(rec) {
		case "buy":
			consensus.BuyCount++
		case "hold":
			consensus.HoldCount++
		case "sell":
			consensus.SellCount++
		}

		if acc3m.Valid {
			totalAccuracy += acc3m.Float64
			accCount++
		}
	}

	if consensus.TotalAnalysts == 0 {
		return consensus, nil
	}

	consensus.AverageTargetPrice = totalTarget / float64(consensus.TotalAnalysts)
	consensus.HighTargetPrice = highTarget
	consensus.LowTargetPrice = lowTarget

	if accCount > 0 {
		consensus.AverageAccuracy = totalAccuracy / float64(accCount)
	}

	// Determine consensus action
	if consensus.BuyCount > consensus.HoldCount && consensus.BuyCount > consensus.SellCount {
		consensus.ConsensusAction = "buy"
	} else if consensus.SellCount > consensus.BuyCount && consensus.SellCount > consensus.HoldCount {
		consensus.ConsensusAction = "sell"
	} else {
		consensus.ConsensusAction = "hold"
	}

	return consensus, nil
}

// UpdateAccuracy tracks accuracy at 1m, 3m, 6m intervals by comparing
// target prices against actual prices.
func (s *AnalystIQService) UpdateAccuracy(ctx context.Context, fetchPrice func(ctx context.Context, symbol string) (float64, error)) error {
	now := time.Now()

	// Find reports needing accuracy updates
	query := `SELECT id, symbol, target_price, published_at, accuracy_1m, accuracy_3m, accuracy_6m
		FROM analyst_reports
		WHERE (accuracy_1m IS NULL AND published_at <= $1)
		   OR (accuracy_3m IS NULL AND published_at <= $2)
		   OR (accuracy_6m IS NULL AND published_at <= $3)
		ORDER BY published_at ASC LIMIT 100`

	rows, err := s.db.QueryContext(ctx, query,
		now.AddDate(0, -1, 0),
		now.AddDate(0, -3, 0),
		now.AddDate(0, -6, 0),
	)
	if err != nil {
		return fmt.Errorf("failed to query reports for accuracy update: %w", err)
	}
	defer rows.Close()

	type reportToUpdate struct {
		ID          int64
		Symbol      string
		TargetPrice float64
		PublishedAt time.Time
		Acc1M       sql.NullFloat64
		Acc3M       sql.NullFloat64
		Acc6M       sql.NullFloat64
	}

	var reports []reportToUpdate
	for rows.Next() {
		var r reportToUpdate
		if err := rows.Scan(&r.ID, &r.Symbol, &r.TargetPrice, &r.PublishedAt, &r.Acc1M, &r.Acc3M, &r.Acc6M); err != nil {
			continue
		}
		reports = append(reports, r)
	}

	updated := 0
	for _, r := range reports {
		if fetchPrice == nil {
			continue
		}
		currentPrice, err := fetchPrice(ctx, r.Symbol)
		if err != nil || currentPrice == 0 {
			continue
		}

		// Accuracy = 1 - |actual - target| / target (clamped to 0-1)
		accuracy := 1.0 - abs(currentPrice-r.TargetPrice)/r.TargetPrice
		if accuracy < 0 {
			accuracy = 0
		}

		age := now.Sub(r.PublishedAt)

		if !r.Acc1M.Valid && age >= 30*24*time.Hour {
			s.updateAccuracyField(ctx, r.ID, "accuracy_1m", accuracy)
			updated++
		}
		if !r.Acc3M.Valid && age >= 90*24*time.Hour {
			s.updateAccuracyField(ctx, r.ID, "accuracy_3m", accuracy)
			updated++
		}
		if !r.Acc6M.Valid && age >= 180*24*time.Hour {
			s.updateAccuracyField(ctx, r.ID, "accuracy_6m", accuracy)
			updated++
		}
	}

	if updated > 0 {
		log.Printf("[AnalystIQ] Updated accuracy for %d report fields", updated)
	}
	return nil
}

// GetAnalystAccuracy returns the average accuracy for a specific analyst.
func (s *AnalystIQService) GetAnalystAccuracy(ctx context.Context, analyst string) (float64, int, error) {
	query := `SELECT AVG(accuracy_3m), COUNT(*)
		FROM analyst_reports
		WHERE analyst = $1 AND accuracy_3m IS NOT NULL`

	var avgAcc sql.NullFloat64
	var count int
	err := s.db.QueryRowContext(ctx, query, analyst).Scan(&avgAcc, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get analyst accuracy: %w", err)
	}
	if !avgAcc.Valid {
		return 0, count, nil
	}
	return avgAcc.Float64, count, nil
}

// --- Helpers ---

func (s *AnalystIQService) updateAccuracyField(ctx context.Context, id int64, field string, value float64) {
	query := fmt.Sprintf("UPDATE analyst_reports SET %s = $1 WHERE id = $2", field)
	if _, err := s.db.ExecContext(ctx, query, value, id); err != nil {
		log.Printf("[AnalystIQ] Failed to update %s for report %d: %v", field, id, err)
	}
}

func normalizeRecommendation(rec string) string {
	switch rec {
	case "buy", "outperform", "overweight", "strong_buy":
		return "buy"
	case "sell", "underperform", "underweight", "strong_sell":
		return "sell"
	default:
		return "hold"
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

type reportScannable interface {
	Scan(dest ...any) error
}

func scanReport(rows reportScannable) (AnalystReport, error) {
	var r AnalystReport
	err := rows.Scan(
		&r.ID, &r.Symbol, &r.Analyst, &r.Brokerage,
		&r.Recommendation, &r.TargetPrice, &r.PublishedAt,
		&r.Accuracy1M, &r.Accuracy3M, &r.Accuracy6M,
	)
	if err != nil {
		return r, fmt.Errorf("failed to scan analyst report: %w", err)
	}
	return r, nil
}
