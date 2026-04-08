-- 002_portfolios.sql: Holdings table for stock portfolio
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    quantity NUMERIC(18, 4) NOT NULL,
    avg_cost NUMERIC(18, 4) NOT NULL,
    total_dividends NUMERIC(18, 4) DEFAULT 0,
    acquisition_date TIMESTAMPTZ,
    account VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, symbol)
);
CREATE INDEX IF NOT EXISTS idx_holdings_user_id ON holdings(user_id);
CREATE INDEX IF NOT EXISTS idx_holdings_symbol ON holdings(symbol);