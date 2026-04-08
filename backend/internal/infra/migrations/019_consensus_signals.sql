-- Multi-source sentiment signals for market consensus aggregation.
CREATE TABLE IF NOT EXISTS consensus_signals (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    source VARCHAR(30) NOT NULL CHECK (
        source IN ('news', 'analyst', 'social_media', 'forum')
    ),
    score DOUBLE PRECISION NOT NULL,
    -- -1.0 to +1.0
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    text_snippet TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    topics JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_consensus_signals_symbol ON consensus_signals (symbol);
CREATE INDEX IF NOT EXISTS idx_consensus_signals_created_at ON consensus_signals (created_at);
CREATE INDEX IF NOT EXISTS idx_consensus_signals_symbol_source ON consensus_signals (symbol, source);
CREATE INDEX IF NOT EXISTS idx_consensus_signals_symbol_created ON consensus_signals (symbol, created_at);