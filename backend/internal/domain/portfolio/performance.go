package portfolio

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"myfi-backend/internal/infra"

	"github.com/dda10/vnstock-go"
)

// PerformanceEngine computes portfolio performance metrics:
// TWR, XIRR, NAV equity curve, and benchmark comparison (VN-Index, VN30).
type PerformanceEngine struct {
	db     *sql.DB
	router *infra.DataSourceRouter
}

// NewPerformanceEngine creates a new PerformanceEngine instance.
func NewPerformanceEngine(db *sql.DB, router *infra.DataSourceRouter) *PerformanceEngine {
	return &PerformanceEngine{
		db:     db,
		router: router,
	}
}

// ComputeTWR computes the time-weighted return using chain-linking of sub-period returns.
func (e *PerformanceEngine) ComputeTWR(ctx context.Context, userID string, startDate, endDate time.Time) (float64, error) {
	snapshots, err := e.GetEquityCurve(ctx, userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get equity curve: %w", err)
	}
	if len(snapshots) < 2 {
		return 0, nil
	}

	cashFlows, err := e.getCashFlowEvents(ctx, userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get cash flows: %w", err)
	}

	cfByDate := make(map[string]float64)
	for _, cf := range cashFlows {
		key := cf.Date.Format("2006-01-02")
		cfByDate[key] += cf.Amount
	}

	product := 1.0
	for i := 1; i < len(snapshots); i++ {
		navStart := snapshots[i-1].NAV
		navEnd := snapshots[i].NAV
		if navStart <= 0 {
			continue
		}
		dateKey := snapshots[i].Date.Format("2006-01-02")
		cf := cfByDate[dateKey]
		subReturn := (navEnd - navStart - cf) / navStart
		product *= (1 + subReturn)
	}

	return product - 1, nil
}

// ComputeXIRR computes the money-weighted rate of return (XIRR) using Newton-Raphson.
func (e *PerformanceEngine) ComputeXIRR(ctx context.Context, userID string) (float64, error) {
	// Use trailing 1 year as default range.
	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0)

	startNAV, err := e.getNAVAtDate(ctx, userID, startDate)
	if err != nil || startNAV <= 0 {
		return 0, nil
	}

	endNAV, err := e.getNAVAtDate(ctx, userID, endDate)
	if err != nil || endNAV <= 0 {
		return 0, nil
	}

	cashFlows, err := e.getCashFlowEvents(ctx, userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get cash flows: %w", err)
	}

	totalDays := endDate.Sub(startDate).Hours() / 24
	if totalDays <= 0 {
		return 0, nil
	}

	type irrCF struct {
		days   float64
		amount float64
	}

	var flows []irrCF
	flows = append(flows, irrCF{days: 0, amount: -startNAV})
	for _, cf := range cashFlows {
		days := cf.Date.Sub(startDate).Hours() / 24
		if days <= 0 || days >= totalDays {
			continue
		}
		flows = append(flows, irrCF{days: days, amount: -cf.Amount})
	}
	flows = append(flows, irrCF{days: totalDays, amount: endNAV})

	// Newton-Raphson iteration.
	rate := 0.0001
	for iter := 0; iter < 200; iter++ {
		f := 0.0
		fPrime := 0.0
		for _, cf := range flows {
			denom := math.Pow(1+rate, cf.days)
			if denom == 0 {
				continue
			}
			f += cf.amount / denom
			fPrime -= cf.days * cf.amount / math.Pow(1+rate, cf.days+1)
		}
		if math.Abs(fPrime) < 1e-15 {
			break
		}
		newRate := rate - f/fPrime
		if math.Abs(newRate-rate) < 1e-10 {
			rate = newRate
			break
		}
		rate = newRate
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < -0.99 {
			return 0, nil
		}
	}

	return math.Pow(1+rate, 365) - 1, nil
}

// GetEquityCurve retrieves the NAV equity curve for a user over a date range.
func (e *PerformanceEngine) GetEquityCurve(ctx context.Context, userID string, startDate, endDate time.Time) ([]NAVSnapshot, error) {
	rows, err := e.db.QueryContext(ctx,
		`SELECT nav, snapshot_date FROM nav_snapshots
		 WHERE user_id = $1 AND snapshot_date >= $2 AND snapshot_date <= $3
		 ORDER BY snapshot_date ASC`,
		userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query NAV snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []NAVSnapshot
	for rows.Next() {
		var s NAVSnapshot
		if err := rows.Scan(&s.NAV, &s.Date); err != nil {
			return nil, fmt.Errorf("failed to scan NAV snapshot: %w", err)
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// StoreNAVSnapshot stores a daily NAV snapshot for a user.
func (e *PerformanceEngine) StoreNAVSnapshot(ctx context.Context, userID string, nav float64) error {
	loc := time.FixedZone("ICT", 7*3600)
	snapshotDate := time.Now().In(loc).Truncate(24 * time.Hour)

	_, err := e.db.ExecContext(ctx,
		`INSERT INTO nav_snapshots (user_id, nav, snapshot_date, created_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id, snapshot_date) DO UPDATE SET nav = EXCLUDED.nav`,
		userID, nav, snapshotDate,
	)
	if err != nil {
		return fmt.Errorf("failed to store NAV snapshot: %w", err)
	}
	return nil
}

// ComputeBenchmarkComparison computes portfolio return vs VN-Index and VN30.
func (e *PerformanceEngine) ComputeBenchmarkComparison(ctx context.Context, userID string, benchmark string, startDate, endDate time.Time) (BenchmarkComparison, error) {
	result := BenchmarkComparison{}

	portfolioReturn, err := e.ComputeTWR(ctx, userID, startDate, endDate)
	if err != nil {
		return result, fmt.Errorf("failed to compute portfolio TWR: %w", err)
	}
	result.PortfolioReturn = portfolioReturn

	vnIndexReturn, err := e.fetchIndexReturn(ctx, "VNINDEX", startDate, endDate)
	if err != nil {
		log.Printf("[PerformanceEngine] Failed to fetch VN-Index data: %v", err)
	} else {
		result.VNIndexReturn = vnIndexReturn
	}

	vn30Return, err := e.fetchIndexReturn(ctx, "VN30", startDate, endDate)
	if err != nil {
		log.Printf("[PerformanceEngine] Failed to fetch VN30 data: %v", err)
	} else {
		result.VN30Return = vn30Return
	}

	result.Alpha = result.PortfolioReturn - result.VNIndexReturn

	return result, nil
}

// fetchIndexReturn fetches historical index data and computes the return over the period.
func (e *PerformanceEngine) fetchIndexReturn(ctx context.Context, indexName string, startDate, endDate time.Time) (float64, error) {
	req := vnstock.IndexHistoryRequest{
		Name:     indexName,
		Start:    startDate,
		End:      endDate,
		Interval: "1D",
	}

	records, _, err := e.router.FetchIndexHistory(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch %s history: %w", indexName, err)
	}
	if len(records) < 2 {
		return 0, fmt.Errorf("insufficient %s data points: got %d", indexName, len(records))
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})

	startClose := records[0].Close
	endClose := records[len(records)-1].Close
	if startClose <= 0 {
		return 0, fmt.Errorf("%s start close price is zero", indexName)
	}

	return (endClose - startClose) / startClose, nil
}

// getCashFlowEvents retrieves external cash flow events for TWR and XIRR calculations.
func (e *PerformanceEngine) getCashFlowEvents(ctx context.Context, userID string, startDate, endDate time.Time) ([]CashFlowEvent, error) {
	rows, err := e.db.QueryContext(ctx,
		`SELECT transaction_date, tx_type, total_value
		 FROM transactions
		 WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date <= $3
		 ORDER BY transaction_date ASC`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query cash flows: %w", err)
	}
	defer rows.Close()

	var events []CashFlowEvent
	for rows.Next() {
		var txDate time.Time
		var txType string
		var totalValue float64
		if err := rows.Scan(&txDate, &txType, &totalValue); err != nil {
			return nil, fmt.Errorf("failed to scan cash flow: %w", err)
		}

		var amount float64
		switch TransactionType(txType) {
		case TxBuy:
			amount = totalValue // inflow
		case TxSell:
			amount = -totalValue // outflow
		default:
			continue
		}

		events = append(events, CashFlowEvent{Date: txDate, Amount: amount})
	}
	return events, rows.Err()
}

// getNAVAtDate retrieves the NAV snapshot closest to (on or before) the given date.
func (e *PerformanceEngine) getNAVAtDate(ctx context.Context, userID string, date time.Time) (float64, error) {
	var nav float64
	err := e.db.QueryRowContext(ctx,
		`SELECT nav FROM nav_snapshots
		 WHERE user_id = $1 AND snapshot_date <= $2
		 ORDER BY snapshot_date DESC LIMIT 1`,
		userID, date.Format("2006-01-02"),
	).Scan(&nav)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get NAV at date %s: %w", date.Format("2006-01-02"), err)
	}
	return nav, nil
}
