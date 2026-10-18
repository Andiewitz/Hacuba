package dbmigrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestApplyRecordsAndSkipsAppliedMigrations(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	directory := t.TempDir()
	filename := "900_migration_runner_" + uuid.NewString()[0:8] + ".sql"
	table := "migration_runner_" + uuid.NewString()[0:8]
	if err := os.WriteFile(filepath.Join(directory, filename), []byte(fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY);", table)), 0o600); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, "DROP TABLE IF EXISTS "+table) //nolint:errcheck

	applied, err := ApplyWithPool(ctx, pool, directory)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("first ApplyWithPool applied %d migrations, want 1", applied)
	}
	applied, err = ApplyWithPool(ctx, pool, directory)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Fatalf("second ApplyWithPool applied %d migrations, want 0", applied)
	}
	var recorded int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations WHERE filename=$1`, filename).Scan(&recorded); err != nil {
		t.Fatal(err)
	}
	if recorded != 1 {
		t.Fatalf("recorded migrations = %d, want 1", recorded)
	}
}

func TestMigrationFilesUsesLexicalSQLOrder(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{"010_last.sql", "ignore.txt", "002_first.sql"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := migrationFiles(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || filepath.Base(files[0]) != "002_first.sql" || filepath.Base(files[1]) != "010_last.sql" {
		t.Fatalf("migration files = %v", files)
	}
}
