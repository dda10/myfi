package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"myfi-backend/internal/infra"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupPostgresTestDB starts a PostgreSQL 16 container, runs all production DDL
// migrations, seeds a test user, and returns a *sql.DB. The container and
// connection are cleaned up automatically when the test finishes.
func SetupPostgresTestDB(t testing.TB) *sql.DB {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("myfi_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Run all production DDL migrations
	for i, m := range infra.AllMigrations() {
		if _, err := db.Exec(m); err != nil {
			t.Fatalf("migration %d failed: %v", i+1, err)
		}
	}

	// Seed test user
	if _, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ('testuser', 'hash')`); err != nil {
		t.Fatalf("failed to seed test user: %v", err)
	}

	return db
}
