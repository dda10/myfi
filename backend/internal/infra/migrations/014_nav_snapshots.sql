-- 014_nav_snapshots.sql: Daily NAV snapshots
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS nav_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    nav NUMERIC(18, 4) NOT NULL,
    holdings_snapshot JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, snapshot_date)
);
CREATE INDEX IF NOT EXISTS idx_nav_user_date ON nav_snapshots(user_id, snapshot_date);