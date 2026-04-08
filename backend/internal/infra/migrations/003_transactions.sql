-- 003_transactions.sql: Transaction ledger
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    quantity NUMERIC(18, 4) NOT NULL,
    unit_price NUMERIC(18, 4) NOT NULL,
    total_value NUMERIC(18, 4) NOT NULL,
    tx_type VARCHAR(20) NOT NULL,
    -- 'buy', 'sell', 'dividend', 'split', 'bonus'
    realized_pnl NUMERIC(18, 4),
    transaction_date TIMESTAMPTZ NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(symbol);