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
	id, mode, propertyType, title, city, barangay, sellerName, phone, email string
	imageKeys                                                               []string
	price                                                                   int64
	bedrooms, bathrooms                                                     *int16
	floorArea, lotArea                                                      *float64
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
		price := item.price
		barangay := item.barangay
		now := time.Now().UTC()
		listing := listings.Listing{ID: id, OwnerID: ownerID, ListingMode: item.mode, PropertyType: item.propertyType, Title: item.title, Description: "Development seed listing for the full browse and property-card flow.", PriceCentavos: &price, Currency: "PHP", City: item.city, Barangay: &barangay, Bedrooms: item.bedrooms, Bathrooms: item.bathrooms, FloorAreaSQM: item.floorArea, LotAreaSQM: item.lotArea, SellerName: item.sellerName, ContactPhone: item.phone, ContactEmail: item.email, Status: listings.StatusPublished, PublishedAt: &now}
		if existing, err := store.GetOwner(context.Background(), id, ownerID); err == nil {
			listing.CreatedAt, listing.UpdatedAt = existing.CreatedAt, existing.UpdatedAt
			if _, err := store.SaveOwner(context.Background(), listing, ownerID); err != nil {
				log.Fatal(err)
			}
		} else if _, err := store.Create(context.Background(), listing); err != nil {
			log.Fatal(err)
		}
		existingImages, err := store.ListImages(context.Background(), id)
		if err != nil {
			log.Fatal(err)
		}
		existingKeys := map[string]bool{}
		for _, image := range existingImages {
			existingKeys[image.ObjectKey] = true
		}
		for position, key := range item.imageKeys {
			if existingKeys[key] {
				continue
			}
			if _, err := store.AddImage(context.Background(), listings.Image{ID: uuid.New(), ListingID: id, ObjectKey: key, Position: int16(position), ContentType: "image/jpeg", ByteSize: 128_000}, ownerID); err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Print("devseed: published development listings are ready")
}

func seeds() []seed {
	bedrooms, bathrooms := int16(3), int16(2)
	floorArea, lotArea := 142.0, 250.0
	return []seed{
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550001", mode: listings.ModeSale, propertyType: listings.TypeHouse, title: "Family home near Cebu IT Park", city: "Cebu City", barangay: "Lahug", sellerName: "Mika Santos", phone: "+63 917 555 0142", email: "mika.santos@example.test", imageKeys: []string{"development/lahug-house.jpg", "development/lahug-house-living.jpg", "development/lahug-house-garden.jpg"}, price: 750_000_000, bedrooms: &bedrooms, bathrooms: &bathrooms, floorArea: &floorArea, lotArea: &lotArea},
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550002", mode: listings.ModeRent, propertyType: listings.TypeCondo, title: "Two-bedroom Lahug condo", city: "Cebu City", barangay: "Lahug", sellerName: "Paolo Reyes", phone: "+63 917 555 0188", email: "paolo.reyes@example.test", imageKeys: []string{"development/lahug-condo.jpg", "development/lahug-condo-bedroom.jpg", "development/lahug-condo-view.jpg"}, price: 3_800_000, bedrooms: &bedrooms, bathrooms: &bathrooms, floorArea: &floorArea},
		{id: "3fca4f5b-55ec-44e4-a9a7-762d8d550003", mode: listings.ModeSale, propertyType: listings.TypeLot, title: "Hillside lot in Busay", city: "Cebu City", barangay: "Busay", sellerName: "Elena Cruz", phone: "+63 917 555 0164", email: "elena.cruz@example.test", imageKeys: []string{"development/busay-lot.jpg", "development/busay-lot-view.jpg", "development/busay-lot-road.jpg"}, price: 125_000_000, lotArea: &lotArea},
	}
}
