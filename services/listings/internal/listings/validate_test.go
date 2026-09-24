package listings

import "testing"

func TestPublishRequiresPublicSellerContact(t *testing.T) {
	price := int64(750_000_000)
	bedrooms, bathrooms := int16(3), int16(2)
	floorArea, lotArea := 142.0, 250.0
	listing := Listing{
		ListingMode:   ModeSale,
		PropertyType:  TypeHouse,
		Title:         "Family home near Cebu IT Park",
		Description:   "A complete development listing for contact validation.",
		PriceCentavos: &price,
		Currency:      "PHP",
		City:          "Cebu City",
		Bedrooms:      &bedrooms,
		Bathrooms:     &bathrooms,
		FloorAreaSQM:  &floorArea,
		LotAreaSQM:    &lotArea,
	}

	fields := MissingPublishFields(listing, 1)
	if fields["seller_name"] != "missing" || fields["contact"] == "" {
		t.Fatalf("missing seller contact fields = %#v", fields)
	}

	listing.SellerName = "Mika Santos"
	listing.ContactPhone = "+63 917 555 0142"
	if fields := MissingPublishFields(listing, 1); len(fields) != 0 {
		t.Fatalf("complete contact fields = %#v", fields)
	}
}

func TestDraftRejectsMalformedPublicContact(t *testing.T) {
	fields := ValidateDraft(Listing{SellerName: "M", ContactPhone: "call me", ContactEmail: "not-an-email"})
	for _, field := range []string{"seller_name", "contact_phone", "contact_email"} {
		if fields[field] == "" {
			t.Fatalf("expected %s validation error: %#v", field, fields)
		}
	}
}
