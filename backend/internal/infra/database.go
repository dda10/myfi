package infra

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB initializes the database connection and runs migrations
func InitDB() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// CloseDB closes the database connection
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// AllMigrations returns all DDL migration statements for use by tests and runMigrations
func AllMigrations() []string {
	return []string{
		createUsersTable,
		createAssetsTable,
		createTransactionsTable,
		createSavingsAccountsTable,
		createNavSnapshotsTable,
		createPatternObservationsTable,
		createAlertsTable,
		createAlertPreferencesTable,
		createWatchlistsTable,
		createWatchlistSymbolsTable,
		createFilterPresetsTable,
		createRecommendationAuditLogTable,
		createFinancialGoalsTable,
		createStockSectorMappingTable,
		createCacheEntriesTable,
		alterAlertsAddChartLink,
		createPerformanceIndexes,
	}
}

// runMigrations executes all database migrations
func runMigrations(db *sql.DB) error {
	migrations := AllMigrations()

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Printf("Successfully ran %d migrations", len(migrations))
	return nil
}

// Migration SQL statements
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMP WITH TIME ZONE,
    theme_preference TEXT DEFAULT 'light',
    language_preference TEXT DEFAULT 'vi-VN'
);
`

const createAssetsTable = `
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('vn_stock', 'crypto', 'gold', 'savings', 'bond', 'cash')),
    symbol TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    average_cost DOUBLE PRECISION NOT NULL,
    acquisition_date TIMESTAMP WITH TIME ZONE NOT NULL,
    account TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_symbol ON assets(symbol);
`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    asset_type TEXT NOT NULL,
    symbol TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    unit_price DOUBLE PRECISION NOT NULL,
    total_value DOUBLE PRECISION NOT NULL,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell', 'deposit', 'withdrawal', 'interest', 'dividend')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(symbol);
`

const createSavingsAccountsTable = `
CREATE TABLE IF NOT EXISTS savings_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    account_name TEXT NOT NULL,
    principal DOUBLE PRECISION NOT NULL,
    annual_rate DOUBLE PRECISION NOT NULL,
    compounding_frequency TEXT NOT NULL CHECK (compounding_frequency IN ('monthly', 'quarterly', 'yearly')),
    start_date DATE NOT NULL,
    maturity_date DATE,
    is_matured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_savings_user_id ON savings_accounts(user_id);
`

const createNavSnapshotsTable = `
CREATE TABLE IF NOT EXISTS nav_snapshots (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    nav DOUBLE PRECISION NOT NULL,
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_nav_user_date ON nav_snapshots(user_id, snapshot_date);
`

const createPatternObservationsTable = `
CREATE TABLE IF NOT EXISTS pattern_observations (
    id SERIAL PRIMARY KEY,
    symbol TEXT NOT NULL,
    pattern_type TEXT NOT NULL CHECK (pattern_type IN ('accumulation', 'distribution', 'breakout')),
    detection_date TIMESTAMP WITH TIME ZONE NOT NULL,
    confidence_score INTEGER NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 100),
    price_at_detection DOUBLE PRECISION NOT NULL,
    supporting_data TEXT,
    outcome_1day DOUBLE PRECISION,
    outcome_7day DOUBLE PRECISION,
    outcome_14day DOUBLE PRECISION,
    outcome_30day DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_observations_symbol ON pattern_observations(symbol);
CREATE INDEX IF NOT EXISTS idx_observations_pattern ON pattern_observations(pattern_type);
CREATE INDEX IF NOT EXISTS idx_observations_date ON pattern_observations(detection_date);
`

const createAlertsTable = `
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    pattern_type TEXT NOT NULL,
    confidence_score INTEGER NOT NULL,
    explanation TEXT,
    detection_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    viewed BOOLEAN DEFAULT FALSE,
    expired BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_viewed ON alerts(viewed);
`

const createWatchlistsTable = `
CREATE TABLE IF NOT EXISTS watchlists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_watchlists_user_id ON watchlists(user_id);
`

const createWatchlistSymbolsTable = `
CREATE TABLE IF NOT EXISTS watchlist_symbols (
    id SERIAL PRIMARY KEY,
    watchlist_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    position INTEGER NOT NULL,
    price_alert_above DOUBLE PRECISION,
    price_alert_below DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (watchlist_id) REFERENCES watchlists(id) ON DELETE CASCADE,
    UNIQUE(watchlist_id, symbol)
);

CREATE INDEX IF NOT EXISTS idx_watchlist_symbols_watchlist_id ON watchlist_symbols(watchlist_id);
`

const createFilterPresetsTable = `
CREATE TABLE IF NOT EXISTS filter_presets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    filters TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_presets_user_id ON filter_presets(user_id);
`

const createRecommendationAuditLogTable = `
CREATE TABLE IF NOT EXISTS recommendation_audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    user_query TEXT NOT NULL,
    sub_agent_inputs TEXT NOT NULL,
    recommendation_output TEXT NOT NULL,
    symbols_involved TEXT,
    recommended_actions TEXT,
    outcome_1day TEXT,
    outcome_7day TEXT,
    outcome_14day TEXT,
    outcome_30day TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_user_id ON recommendation_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON recommendation_audit_log(timestamp);
`

const createFinancialGoalsTable = `
CREATE TABLE IF NOT EXISTS financial_goals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    target_amount DOUBLE PRECISION NOT NULL,
    target_date DATE NOT NULL,
    associated_asset_types TEXT,
    category TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_goals_user_id ON financial_goals(user_id);
`

const createStockSectorMappingTable = `
CREATE TABLE IF NOT EXISTS stock_sector_mapping (
    symbol TEXT PRIMARY KEY,
    sector TEXT NOT NULL,
    sector_name TEXT,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sector_mapping_sector ON stock_sector_mapping(sector);
`

const createCacheEntriesTable = `
CREATE TABLE IF NOT EXISTS cache_entries (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at);
`

// Alert preferences table for user-configurable alert settings
// Requirement 14.5: Support user alert preferences
const createAlertPreferencesTable = `
CREATE TABLE IF NOT EXISTS alert_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    min_confidence INTEGER DEFAULT 60,
    pattern_types TEXT,
    include_symbols TEXT,
    exclude_symbols TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alert_prefs_user_id ON alert_preferences(user_id);
`

// Add chart_link column to alerts table if it doesn't exist
// Requirement 14.3: Include chart link in alerts
const alterAlertsAddChartLink = `
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'alerts' AND column_name = 'chart_link'
    ) THEN
        ALTER TABLE alerts ADD COLUMN chart_link TEXT;
    END IF;
END $$;
`

// Performance indexes for frequently queried fields (Task 38.1)
const createPerformanceIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_assets_user_type ON assets(user_id, asset_type);
CREATE INDEX IF NOT EXISTS idx_transactions_user_symbol ON transactions(user_id, symbol);
CREATE INDEX IF NOT EXISTS idx_alerts_user_symbol ON alerts(user_id, symbol);
CREATE INDEX IF NOT EXISTS idx_alerts_expired ON alerts(expired);
CREATE INDEX IF NOT EXISTS idx_watchlist_symbols_symbol ON watchlist_symbols(symbol);
CREATE INDEX IF NOT EXISTS idx_nav_snapshots_user ON nav_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_symbols ON recommendation_audit_log(symbols_involved);
CREATE INDEX IF NOT EXISTS idx_goals_target_date ON financial_goals(target_date);
`
