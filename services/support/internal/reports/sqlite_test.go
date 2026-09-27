package reports

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSQLiteStorePersistsOpenReport(t *testing.T) {
	store, err := NewSQLiteStore(context.Background(), t.TempDir()+"/support.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	_, err = store.Create(context.Background(), Report{ID: uuid.New(), Category: CategoryBug, Description: "The help page did not accept my report after I completed the form.", Status: StatusOpen})
	if err != nil {
		t.Fatal(err)
	}
	count, err := store.OpenCount(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("open reports = %d, %v", count, err)
	}
}

func TestSQLiteStoreAuditsStaffTransition(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteStore(ctx, t.TempDir()+"/support.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	created, err := store.Create(ctx, Report{ID: uuid.New(), Category: CategorySafety, Description: "The property contact requested payment before an in-person viewing.", Status: StatusOpen})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.Transition(ctx, created.ID, uuid.New(), StatusTriaged)
	if err != nil || updated.Status != StatusTriaged {
		t.Fatalf("triage = %#v, %v", updated, err)
	}
	var actions int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM dev_report_actions WHERE report_id=?`, created.ID.String()).Scan(&actions); err != nil || actions != 1 {
		t.Fatalf("audit actions = %d, %v", actions, err)
	}
	count, err := store.OpenCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("unresolved reports = %d, %v", count, err)
	}
}
