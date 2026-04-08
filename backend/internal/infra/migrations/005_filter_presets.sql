-- 005_filter_presets.sql: Screener filter presets (max 10 per user enforced via trigger)
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS filter_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    filters JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_filter_presets_user_id ON filter_presets(user_id);
-- Enforce max 10 filter presets per user
CREATE OR REPLACE FUNCTION check_filter_preset_limit() RETURNS TRIGGER AS $$ BEGIN IF (
        SELECT COUNT(*)
        FROM filter_presets
        WHERE user_id = NEW.user_id
    ) >= 10 THEN RAISE EXCEPTION 'Maximum of 10 filter presets per user reached';
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_filter_preset_limit ON filter_presets;
CREATE TRIGGER trg_filter_preset_limit BEFORE
INSERT ON filter_presets FOR EACH ROW EXECUTE FUNCTION check_filter_preset_limit();