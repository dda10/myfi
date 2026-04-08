-- 007_recommendations.sql: AI recommendation audit trail with outcome tracking
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    action VARCHAR(10) NOT NULL,
    -- 'buy', 'sell', 'hold'
    position_size NUMERIC(5, 2),
    risk_assessment VARCHAR(20),
    -- 'low', 'medium', 'high'
    confidence_score INT,
    reasoning TEXT,
    input_snapshot JSONB,
    price_at_signal NUMERIC(18, 4),
    stop_loss NUMERIC(18, 4),
    take_profit NUMERIC(18, 4),
    price_1d NUMERIC(18, 4),
    price_7d NUMERIC(18, 4),
    price_14d NUMERIC(18, 4),
    price_30d NUMERIC(18, 4),
    return_1d NUMERIC(10, 4),
    return_7d NUMERIC(10, 4),
    return_14d NUMERIC(10, 4),
    return_30d NUMERIC(10, 4),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_recommendations_symbol ON recommendations(symbol, created_at DESC);