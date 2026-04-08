package infra

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

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

// AllMigrations returns all DDL migration statements for use by tests and runMigrations.
// Migrations are loaded from embedded SQL files in sorted order (001_, 002_, ...).
func AllMigrations() []string {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		log.Fatalf("failed to read embedded migrations: %v", err)
	}

	// Sort entries by name to ensure deterministic ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var migrations []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", entry.Name(), err)
		}
		migrations = append(migrations, string(data))
	}

	return migrations
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
