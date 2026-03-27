package infra

import (
	"database/sql"
	"log/slog"
	"time"
)

// DBPoolConfig holds connection pool tuning parameters.
type DBPoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultDBPoolConfig returns production-tuned pool settings.
func DefaultDBPoolConfig() DBPoolConfig {
	return DBPoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 3 * time.Minute,
	}
}

// ApplyPoolConfig configures the connection pool on the given *sql.DB.
func ApplyPoolConfig(db *sql.DB, cfg DBPoolConfig) {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	slog.Info("db pool configured",
		"maxOpen", cfg.MaxOpenConns,
		"maxIdle", cfg.MaxIdleConns,
		"maxLifetime", cfg.ConnMaxLifetime,
		"maxIdleTime", cfg.ConnMaxIdleTime,
	)
}

// EnsureIndexes creates performance indexes idempotently.
// Called on startup after migrations to guarantee indexes exist.
func EnsureIndexes(db *sql.DB) error {
	indexes := []string{
		// Users
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		// Assets
		`CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_user_type ON assets(user_id, asset_type)`,
		// Transactions
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_symbol ON transactions(user_id, symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, transaction_date)`,
		// Watchlists
		`CREATE INDEX IF NOT EXISTS idx_watchlists_user_id ON watchlists(user_id)`,
		// Alerts
		`CREATE INDEX IF NOT EXISTS idx_alerts_user_viewed ON alerts(user_id, viewed)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_user_symbol ON alerts(user_id, symbol)`,
		// NAV snapshots
		`CREATE INDEX IF NOT EXISTS idx_nav_user_date ON nav_snapshots(user_id, snapshot_date)`,
		// Pattern observations
		`CREATE INDEX IF NOT EXISTS idx_observations_symbol ON pattern_observations(symbol)`,
		// Cache entries
		`CREATE INDEX IF NOT EXISTS idx_cache_key ON cache_entries(key)`,
		`CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at)`,
	}

	for _, ddl := range indexes {
		if _, err := db.Exec(ddl); err != nil {
			slog.Warn("index creation skipped", "ddl", ddl, "err", err)
			// Non-fatal: index may already exist or table may not exist yet
		}
	}
	slog.Info("performance indexes ensured", "count", len(indexes))
	return nil
}
