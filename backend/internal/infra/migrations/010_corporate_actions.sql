-- 010_corporate_actions.sql: Corporate action cache
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS corporate_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    ex_date DATE,
    record_date DATE,
    payment_date DATE,
    ratio NUMERIC(10, 4),
    amount NUMERIC(18, 4),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_corp_actions_symbol ON corporate_actions(symbol, ex_date DESC);
CREATE INDEX IF NOT EXISTS idx_corp_actions_ex_date ON corporate_actions(ex_date);