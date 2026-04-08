-- 008_missions.sql: Missions with trigger_type, target_symbols, action_type, status
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL,
    -- 'price_threshold', 'schedule', 'event', 'news'
    trigger_params JSONB NOT NULL,
    target_symbols TEXT [] NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    -- 'alert', 'report', 'agent_analysis'
    action_params JSONB,
    status VARCHAR(20) DEFAULT 'active',
    -- 'active', 'paused', 'triggered'
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_missions_user ON missions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_missions_status ON missions(status);