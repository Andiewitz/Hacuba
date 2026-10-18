// Package dbmigrate applies immutable SQL migrations for one service-owned
// database. It intentionally has no HTTP integration: deployment invokes the
// migration binary before an application process starts serving requests.
package dbmigrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

const advisoryLockID int64 = 8_416_201_001

// Apply opens a short-lived pool, applies pending files, then closes it.
func Apply(ctx context.Context, databaseURL, directory string) (int, error) {
	if databaseURL == "" {
		return 0, fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return 0, fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	return ApplyWithPool(ctx, pool, directory)
}

// ApplyWithPool makes migration behavior testable against a real Postgres
// service. Filenames are the durable migration identity, so edited migrations
// never silently rerun.
func ApplyWithPool(ctx context.Context, pool *pgxpool.Pool, directory string) (int, error) {
	files, err := migrationFiles(directory)
	if err != nil {
		return 0, err
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockID); err != nil {
		return 0, fmt.Errorf("acquire migration lock: %w", err)
	}
	defer conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", advisoryLockID) //nolint:errcheck

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return 0, fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := 0
	for _, path := range files {
		filename := filepath.Base(path)
		var alreadyApplied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)`, filename).Scan(&alreadyApplied); err != nil {
			return applied, fmt.Errorf("check migration %s: %w", filename, err)
		}
		if alreadyApplied {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return applied, fmt.Errorf("read migration %s: %w", filename, err)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return applied, fmt.Errorf("begin migration %s: %w", filename, err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("apply migration %s: %w", filename, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, filename); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("record migration %s: %w", filename, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, fmt.Errorf("commit migration %s: %w", filename, err)
		}
		applied++
	}
	return applied, nil
}

func migrationFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}
