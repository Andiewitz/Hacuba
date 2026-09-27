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
