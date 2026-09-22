package listings

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSQLiteStorePersistsOwnerListings(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/listings.db"
	store, err := NewSQLiteStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	owner, id := uuid.New(), uuid.New()
	price := int64(450_000_000)
	created, err := store.Create(ctx, Listing{ID: id, OwnerID: owner, ListingMode: ModeSale, PropertyType: TypeLot, Title: "Development lot", PriceCentavos: &price, Currency: "PHP", City: "Cebu City", Status: StatusDraft})
	if err != nil {
		t.Fatal(err)
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected create timestamp")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := NewSQLiteStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.GetOwner(ctx, id, owner)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != "Development lot" || loaded.PriceCentavos == nil || *loaded.PriceCentavos != price {
		t.Fatalf("unexpected persisted listing: %#v", loaded)
	}
}
