-- 017_notification_preferences.sql: Per-user notification channel preferences
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    in_app_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT FALSE,
    email_enabled BOOLEAN DEFAULT FALSE,
    quiet_start TIME,
    -- e.g., 22:00
    quiet_end TIME,
    -- e.g., 07:00
    price_alerts BOOLEAN DEFAULT TRUE,
    mission_alerts BOOLEAN DEFAULT TRUE,
    idea_alerts BOOLEAN DEFAULT TRUE,
    research_alerts BOOLEAN DEFAULT TRUE
);