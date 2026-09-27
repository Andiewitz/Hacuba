package reports

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresReportPersistence(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset — skipping PostgreSQL support integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	triageMigration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "002_triage.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(triageMigration)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE support_report_actions, support_reports"); err != nil {
		t.Fatal(err)
	}
	store, err := NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	created, err := store.Create(ctx, Report{ID: uuid.New(), Category: CategorySafety, Description: "The seller requested a payment before I could visit the property in person.", Status: StatusOpen})
	if err != nil {
		t.Fatal(err)
	}
	count, err := store.OpenCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("open reports = %d, %v", count, err)
	}
	updated, err := store.Transition(ctx, created.ID, uuid.New(), StatusTriaged)
	if err != nil || updated.Status != StatusTriaged {
		t.Fatalf("triage report = %#v, %v", updated, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM support_report_actions WHERE report_id=$1`, created.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("audit actions = %d, %v", count, err)
	}
	items, err := store.List(ctx, 10)
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("staff list = %#v, %v", items, err)
	}
	count, err = store.OpenCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("unresolved reports after triage = %d, %v", count, err)
	}
}
