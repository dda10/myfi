-- 015_liquidity_scores.sql: Daily liquidity score cache
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS liquidity_scores (
    symbol VARCHAR(20) PRIMARY KEY,
    score INT NOT NULL,
    -- 0-100
    tier VARCHAR(10) NOT NULL,
    -- 'tier1', 'tier2', 'tier3'
    avg_volume NUMERIC(18, 2),
    avg_value NUMERIC(18, 2),
    zero_volume_days INT,
    computed_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_liquidity_scores_tier ON liquidity_scores(tier);