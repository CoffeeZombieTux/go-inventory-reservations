package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// OpenTestDB opens a Postgres connection for integration tests.
func OpenTestDB(t testing.TB) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5433"),
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", "RooT!@123"),
		getEnv("DB_NAME", "inventory_reservations_test"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("cannot connect to test db: %v", err)
	}

	return db
}

// ResetAndMigrateTestDB drops and reapplies schema migrations.
func ResetAndMigrateTestDB(t testing.TB, ctx context.Context, db *sql.DB) {
	t.Helper()

	migrationsDir := MigrationsDir(t)
	downSQL := readMigrationFile(t, filepath.Join(migrationsDir, "0001_initial_schema.down.sql"))
	upSQL := readMigrationFile(t, filepath.Join(migrationsDir, "0001_initial_schema.up.sql"))

	if _, err := db.ExecContext(ctx, downSQL); err != nil {
		t.Fatalf("failed to run down migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, upSQL); err != nil {
		t.Fatalf("failed to run up migration: %v", err)
	}
}

// MigrationsDir returns project migrations directory path.
func MigrationsDir(t testing.TB) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve current test file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "migrations"))
}

func readMigrationFile(t testing.TB, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file %s: %v", path, err)
	}

	return string(content)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
