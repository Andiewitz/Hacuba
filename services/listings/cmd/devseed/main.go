// Command devseed creates deterministic published listings in the local
// SQLite development database. It is never included in production images.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/listings/internal/listings"
)

var ownerID = uuid.MustParse("6fb13f6e-0c8d-4af7-9c5c-1a34510e9eac")

type seed struct {
	id, mode, propertyType, title, city, barangay, imageKey string
	price                                                   int64
	bedrooms, bathrooms                                     *int16
	floorArea, lotArea                                      *float64
}

func main() {
	path := os.Getenv("DEV_SQLITE_PATH")
	if path == "" {
		log.Fatal("devseed: DEV_SQLITE_PATH is required")
	}
	store, err := listings.NewSQLiteStore(context.Background(), path)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	for _, item := range seeds() {
		id := uuid.MustParse(item.id)
		if _, err := store.GetOwner(context.Background(), id, ownerID); err == nil {
			continue
		}
		price := item.price
		barangay := item.barangay
		now := time.Now().UTC()
		listing := listings.Listing{ID: id, OwnerID: ownerID, ListingMode: item.mode, PropertyType: item.propertyType, Title: item.title, Description: "Development seed listing for the full browse and property-card flow.", PriceCentavos: &price, Currency: "PHP", City: item.city, Barangay: &barangay, Bedrooms: item.bedrooms, Bathrooms: item.bathrooms, FloorAreaSQM: item.floorArea, LotAreaSQM: item.lotArea, Status: listings.StatusPublished, PublishedAt: &now}
		if _, err := store.Create(context.Background(), listing); err != nil {
			log.Fatal(err)
		}
		if _, err := store.AddImage(context.Background(), listings.Image{ID: uuid.New(), ListingID: id, ObjectKey: item.imageKey, Position: 0, ContentType: "image/jpeg", ByteSize: 128_000}, ownerID); err != nil {
			log.Fatal(err)
		}
	}
	log.Print("devseed: published development listings are ready")
}

func seeds() []seed {
	bedrooms, bathrooms := int16(3), int16(2)
	floorArea, lotArea := 142.0, 250.0
	return []seed{
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550001", mode: listings.ModeSale, propertyType: listings.TypeHouse, title: "Family home near Cebu IT Park", city: "Cebu City", barangay: "Lahug", imageKey: "development/lahug-house.jpg", price: 750_000_000, bedrooms: &bedrooms, bathrooms: &bathrooms, floorArea: &floorArea, lotArea: &lotArea},
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550002", mode: listings.ModeRent, propertyType: listings.TypeCondo, title: "Two-bedroom Lahug condo", city: "Cebu City", barangay: "Lahug", imageKey: "development/lahug-condo.jpg", price: 3_800_000, bedrooms: &bedrooms, bathrooms: &bathrooms, floorArea: &floorArea},
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550003", mode: listings.ModeSale, propertyType: listings.TypeLot, title: "Hillside lot in Busay", city: "Cebu City", barangay: "Busay", imageKey: "development/busay-lot.jpg", price: 125_000_000, lotArea: &lotArea},
	}
}
