package listings

import (
	"fmt"
	"strings"
)

// Apply merges a create or patch body. Call ValidateDraft before persisting.
func (l *Listing) Apply(in Input) {
	if in.ListingMode != nil {
		l.ListingMode = *in.ListingMode
	}
	if in.PropertyType != nil {
		l.PropertyType = *in.PropertyType
	}
	if in.Title != nil {
		l.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		l.Description = strings.TrimSpace(*in.Description)
	}
	if in.PriceCentavos != nil {
		l.PriceCentavos = in.PriceCentavos
	}
	if in.Currency != nil {
		l.Currency = *in.Currency
	}
	if in.PricePeriod != nil {
		l.PricePeriod = in.PricePeriod
	}
	if in.City != nil {
		l.City = strings.TrimSpace(*in.City)
	}
	if in.Barangay != nil {
		l.Barangay = in.Barangay
	}
	if in.AddressLine != nil {
		l.AddressLine = in.AddressLine
	}
	if in.Lat != nil {
		l.Lat = in.Lat
	}
	if in.Lng != nil {
		l.Lng = in.Lng
	}
	if in.Bedrooms != nil {
		l.Bedrooms = in.Bedrooms
	}
	if in.Bathrooms != nil {
		l.Bathrooms = in.Bathrooms
	}
	if in.FloorAreaSQM != nil {
		l.FloorAreaSQM = in.FloorAreaSQM
	}
	if in.LotAreaSQM != nil {
		l.LotAreaSQM = in.LotAreaSQM
	}
	if in.Details != nil {
		l.Details = *in.Details
	}
}

func validMode(v string) bool { return v == ModeSale || v == ModeRent }
func validType(v string) bool {
	return v == TypeHouse || v == TypeApartment || v == TypeCondo || v == TypeLot || v == TypeCommercial
}

// ValidateDraft rejects malformed values while permitting missing fields.
func ValidateDraft(l Listing) map[string]string {
	fields := map[string]string{}
	if l.ListingMode != "" && !validMode(l.ListingMode) {
		fields["listing_mode"] = "must be for_sale or for_rent"
	}
	if l.PropertyType != "" && !validType(l.PropertyType) {
		fields["property_type"] = "invalid property type"
	}
	if l.Title != "" && (len(l.Title) < 5 || len(l.Title) > 120) {
		fields["title"] = "must be 5 to 120 characters"
	}
	if len(l.Description) > 5000 {
		fields["description"] = "must be at most 5000 characters"
	}
	if l.PriceCentavos != nil && *l.PriceCentavos <= 0 {
		fields["price_centavos"] = "must be greater than zero"
	}
	if l.Currency != "" && l.Currency != "PHP" {
		fields["currency"] = "must be PHP"
	}
	if l.PricePeriod != nil && *l.PricePeriod != "monthly" {
		fields["price_period"] = "must be monthly"
	}
	if (l.Lat == nil) != (l.Lng == nil) {
		fields["coordinates"] = "lat and lng must be set together"
	}
	if l.Lat != nil && (*l.Lat < -90 || *l.Lat > 90) {
		fields["lat"] = "must be between -90 and 90"
	}
	if l.Lng != nil && (*l.Lng < -180 || *l.Lng > 180) {
		fields["lng"] = "must be between -180 and 180"
	}
	for key, value := range map[string]*int16{"bedrooms": l.Bedrooms, "bathrooms": l.Bathrooms} {
		if value != nil && *value < 0 {
			fields[key] = "must not be negative"
		}
	}
	for key, value := range map[string]*float64{"floor_area_sqm": l.FloorAreaSQM, "lot_area_sqm": l.LotAreaSQM} {
		if value != nil && *value <= 0 {
			fields[key] = "must be greater than zero"
		}
	}
	return fields
}

// MissingPublishFields lists every missing or incompatible field at once.
func MissingPublishFields(l Listing, imageCount int) map[string]string {
	fields := ValidateDraft(l)
	if !validMode(l.ListingMode) {
		fields["listing_mode"] = "missing"
	}
	if !validType(l.PropertyType) {
		fields["property_type"] = "missing"
	}
	if len(l.Title) < 5 {
		fields["title"] = "missing"
	}
	if strings.TrimSpace(l.Description) == "" {
		fields["description"] = "missing"
	}
	if l.PriceCentavos == nil || *l.PriceCentavos <= 0 {
		fields["price_centavos"] = "missing"
	}
	if strings.TrimSpace(l.City) == "" {
		fields["city"] = "missing"
	}
	if imageCount < 1 {
		fields["images"] = "at least one image is required"
	}
	if l.Currency == "" {
		l.Currency = "PHP"
	}
	if l.Currency != "PHP" {
		fields["currency"] = "must be PHP"
	}
	if l.ListingMode == ModeRent && (l.PricePeriod == nil || *l.PricePeriod != "monthly") {
		fields["price_period"] = "monthly is required for rent"
	}
	if l.ListingMode == ModeSale && l.PricePeriod != nil {
		fields["price_period"] = "must be empty for sale"
	}
	require := func(name string, present bool) {
		if !present {
			fields[name] = "missing"
		}
	}
	switch l.PropertyType {
	case TypeHouse:
		require("bedrooms", l.Bedrooms != nil)
		require("bathrooms", l.Bathrooms != nil)
		require("floor_area_sqm", l.FloorAreaSQM != nil)
		require("lot_area_sqm", l.LotAreaSQM != nil)
	case TypeApartment, TypeCondo:
		require("bedrooms", l.Bedrooms != nil)
		require("bathrooms", l.Bathrooms != nil)
		require("floor_area_sqm", l.FloorAreaSQM != nil)
	case TypeLot:
		require("lot_area_sqm", l.LotAreaSQM != nil)
	case TypeCommercial:
		require("floor_area_sqm", l.FloorAreaSQM != nil)
	}
	return fields
}

func (l Listing) String() string { return fmt.Sprintf("listing %s", l.ID) }
