package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the database connection and runs migrations
func InitDB() error {
	dbType := getEnv("DB_TYPE", "sqlite")
	var err error

	if dbType == "sqlite" {
		dbPath := getEnv("DB_PATH", "./myfi.db")
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return fmt.Errorf("failed to open SQLite database: %w", err)
		}
		log.Printf("Connected to SQLite database at %s", dbPath)
	} else if dbType == "postgres" {
		connStr := getEnv("DATABASE_URL", "")
		if connStr == "" {
			return fmt.Errorf("DATABASE_URL environment variable is required for PostgreSQL")
		}
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to open PostgreSQL database: %w", err)
		}
		log.Println("Connected to PostgreSQL database")
	} else {
		return fmt.Errorf("unsupported DB_TYPE: %s (use 'sqlite' or 'postgres')", dbType)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err = runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// runMigrations executes all database migrations
func runMigrations() error {
	migrations := []string{
		createUsersTable,
		createAssetsTable,
		createTransactionsTable,
		createSavingsAccountsTable,
		createNavSnapshotsTable,
		createPatternObservationsTable,
		createAlertsTable,
		createWatchlistsTable,
		createWatchlistSymbolsTable,
		createFilterPresetsTable,
		createRecommendationAuditLogTable,
		createFinancialGoalsTable,
		createStockSectorMappingTable,
		createCacheEntriesTable,
	}

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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until DATETIME,
    theme_preference TEXT DEFAULT 'light',
    language_preference TEXT DEFAULT 'vi-VN'
);
`

const createAssetsTable = `
CREATE TABLE IF NOT EXISTS assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('vn_stock', 'crypto', 'gold', 'savings', 'bond', 'cash')),
    symbol TEXT NOT NULL,
    quantity REAL NOT NULL,
    average_cost REAL NOT NULL,
    acquisition_date DATETIME NOT NULL,
    account TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_symbol ON assets(symbol);
`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    asset_type TEXT NOT NULL,
    symbol TEXT NOT NULL,
    quantity REAL NOT NULL,
    unit_price REAL NOT NULL,
    total_value REAL NOT NULL,
    transaction_date DATETIME NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('buy', 'sell', 'deposit', 'withdrawal', 'interest', 'dividend')),
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_symbol ON transactions(symbol);
`

const createSavingsAccountsTable = `
CREATE TABLE IF NOT EXISTS savings_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    account_name TEXT NOT NULL,
    principal REAL NOT NULL,
    annual_rate REAL NOT NULL,
    compounding_frequency TEXT NOT NULL CHECK (compounding_frequency IN ('monthly', 'quarterly', 'yearly')),
    start_date DATE NOT NULL,
    maturity_date DATE,
    is_matured INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_savings_user_id ON savings_accounts(user_id);
`

const createNavSnapshotsTable = `
CREATE TABLE IF NOT EXISTS nav_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    nav REAL NOT NULL,
    snapshot_date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, snapshot_date)
);

CREATE INDEX IF NOT EXISTS idx_nav_user_date ON nav_snapshots(user_id, snapshot_date);
`

const createPatternObservationsTable = `
CREATE TABLE IF NOT EXISTS pattern_observations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    pattern_type TEXT NOT NULL CHECK (pattern_type IN ('accumulation', 'distribution', 'breakout')),
    detection_date DATETIME NOT NULL,
    confidence_score INTEGER NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 100),
    price_at_detection REAL NOT NULL,
    supporting_data TEXT,
    outcome_1day REAL,
    outcome_7day REAL,
    outcome_14day REAL,
    outcome_30day REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_observations_symbol ON pattern_observations(symbol);
CREATE INDEX IF NOT EXISTS idx_observations_pattern ON pattern_observations(pattern_type);
CREATE INDEX IF NOT EXISTS idx_observations_date ON pattern_observations(detection_date);
`

const createAlertsTable = `
CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    pattern_type TEXT NOT NULL,
    confidence_score INTEGER NOT NULL,
    explanation TEXT,
    detection_timestamp DATETIME NOT NULL,
    viewed INTEGER DEFAULT 0,
    expired INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_viewed ON alerts(viewed);
`

const createWatchlistsTable = `
CREATE TABLE IF NOT EXISTS watchlists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_watchlists_user_id ON watchlists(user_id);
`

const createWatchlistSymbolsTable = `
CREATE TABLE IF NOT EXISTS watchlist_symbols (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    watchlist_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    position INTEGER NOT NULL,
    price_alert_above REAL,
    price_alert_below REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (watchlist_id) REFERENCES watchlists(id) ON DELETE CASCADE,
    UNIQUE(watchlist_id, symbol)
);

CREATE INDEX IF NOT EXISTS idx_watchlist_symbols_watchlist_id ON watchlist_symbols(watchlist_id);
`

const createFilterPresetsTable = `
CREATE TABLE IF NOT EXISTS filter_presets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    filters TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_presets_user_id ON filter_presets(user_id);
`

const createRecommendationAuditLogTable = `
CREATE TABLE IF NOT EXISTS recommendation_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    timestamp DATETIME NOT NULL,
    user_query TEXT NOT NULL,
    sub_agent_inputs TEXT NOT NULL,
    recommendation_output TEXT NOT NULL,
    symbols_involved TEXT,
    recommended_actions TEXT,
    outcome_1day TEXT,
    outcome_7day TEXT,
    outcome_14day TEXT,
    outcome_30day TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_user_id ON recommendation_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON recommendation_audit_log(timestamp);
`

const createFinancialGoalsTable = `
CREATE TABLE IF NOT EXISTS financial_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    target_amount REAL NOT NULL,
    target_date DATE NOT NULL,
    associated_asset_types TEXT,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_goals_user_id ON financial_goals(user_id);
`

const createStockSectorMappingTable = `
CREATE TABLE IF NOT EXISTS stock_sector_mapping (
    symbol TEXT PRIMARY KEY,
    sector TEXT NOT NULL,
    sector_name TEXT,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sector_mapping_sector ON stock_sector_mapping(sector);
`

const createCacheEntriesTable = `
CREATE TABLE IF NOT EXISTS cache_entries (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at);
`
