-- 006_observations.sql: Knowledge base observations (hot data, last 90 days)
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_type VARCHAR(100) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    observation TEXT,
    confidence NUMERIC(5, 2) NOT NULL,
    detection_date TIMESTAMPTZ NOT NULL,
    data_snapshot JSONB,
    price_at_detection NUMERIC(18, 4),
    outcome_1d NUMERIC(10, 4),
    outcome_7d NUMERIC(10, 4),
    outcome_14d NUMERIC(10, 4),
    outcome_30d NUMERIC(10, 4),
    agent_name VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- Index on created_at for 90-day hot data window queries
CREATE INDEX IF NOT EXISTS idx_observations_created_at ON observations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_observations_pattern ON observations(pattern_type, detection_date DESC);
CREATE INDEX IF NOT EXISTS idx_observations_symbol ON observations(symbol, detection_date DESC);