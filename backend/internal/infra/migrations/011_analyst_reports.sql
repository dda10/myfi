-- 011_analyst_reports.sql: Analyst reports with accuracy tracking
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS analyst_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    analyst_name VARCHAR(100),
    brokerage VARCHAR(100),
    recommendation VARCHAR(20),
    -- 'strong_buy', 'buy', 'hold', 'sell', 'strong_sell'
    target_price NUMERIC(18, 4),
    report_date DATE,
    accuracy_1m NUMERIC(5, 2),
    accuracy_3m NUMERIC(5, 2),
    accuracy_6m NUMERIC(5, 2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_analyst_reports_symbol ON analyst_reports(symbol, report_date DESC);
CREATE INDEX IF NOT EXISTS idx_analyst_reports_brokerage ON analyst_reports(brokerage);