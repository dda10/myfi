-- 004_watchlists.sql: Watchlists and watchlist entries with alert thresholds
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_watchlists_user_id ON watchlists(user_id);
CREATE TABLE IF NOT EXISTS watchlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    watchlist_id UUID REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    sort_order INT DEFAULT 0,
    price_alert_above NUMERIC(18, 4),
    price_alert_below NUMERIC(18, 4),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(watchlist_id, symbol)
);
CREATE INDEX IF NOT EXISTS idx_watchlist_entries_watchlist_id ON watchlist_entries(watchlist_id);
CREATE INDEX IF NOT EXISTS idx_watchlist_entries_symbol ON watchlist_entries(symbol);