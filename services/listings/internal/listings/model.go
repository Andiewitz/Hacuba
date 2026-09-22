package listings

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ModeSale = "for_sale"
	ModeRent = "for_rent"

	TypeHouse      = "house"
	TypeApartment  = "apartment"
	TypeCondo      = "condo"
	TypeLot        = "lot"
	TypeCommercial = "commercial"

	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusClosed    = "closed"
	StatusArchived  = "archived"
)

// Listing holds the domain model. Nullable fields stay pointers so a draft
// can be incomplete without manufacturing misleading zero values.
type Listing struct {
	ID            uuid.UUID       `json:"id"`
	OwnerID       uuid.UUID       `json:"-"`
	ListingMode   string          `json:"listing_mode,omitempty"`
	PropertyType  string          `json:"property_type,omitempty"`
	Title         string          `json:"title,omitempty"`
	Description   string          `json:"description,omitempty"`
	PriceCentavos *int64          `json:"price_centavos,omitempty"`
	Currency      string          `json:"currency"`
	PricePeriod   *string         `json:"price_period,omitempty"`
	City          string          `json:"city,omitempty"`
	Barangay      *string         `json:"barangay,omitempty"`
	AddressLine   *string         `json:"-"`
	Lat           *float64        `json:"lat,omitempty"`
	Lng           *float64        `json:"lng,omitempty"`
	Bedrooms      *int16          `json:"bedrooms,omitempty"`
	Bathrooms     *int16          `json:"bathrooms,omitempty"`
	FloorAreaSQM  *float64        `json:"floor_area_sqm,omitempty"`
	LotAreaSQM    *float64        `json:"lot_area_sqm,omitempty"`
	Details       json.RawMessage `json:"details,omitempty"`
	Status        string          `json:"status"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Image struct {
	ID          uuid.UUID `json:"id"`
	ListingID   uuid.UUID `json:"listing_id"`
	ObjectKey   string    `json:"object_key"`
	Position    int16     `json:"position"`
	Alt         *string   `json:"alt,omitempty"`
	ContentType string    `json:"content_type"`
	ByteSize    int64     `json:"byte_size"`
	CreatedAt   time.Time `json:"created_at"`
}

// Input uses pointers so PATCH can distinguish an omitted field from an
// explicit JSON null, which clears nullable draft fields.
type Input struct {
	ListingMode   *string          `json:"listing_mode"`
	PropertyType  *string          `json:"property_type"`
	Title         *string          `json:"title"`
	Description   *string          `json:"description"`
	PriceCentavos *int64           `json:"price_centavos"`
	Currency      *string          `json:"currency"`
	PricePeriod   *string          `json:"price_period"`
	City          *string          `json:"city"`
	Barangay      *string          `json:"barangay"`
	AddressLine   *string          `json:"address_line"`
	Lat           *float64         `json:"lat"`
	Lng           *float64         `json:"lng"`
	Bedrooms      *int16           `json:"bedrooms"`
	Bathrooms     *int16           `json:"bathrooms"`
	FloorAreaSQM  *float64         `json:"floor_area_sqm"`
	LotAreaSQM    *float64         `json:"lot_area_sqm"`
	Details       *json.RawMessage `json:"details"`
}

type ListFilter struct {
	Mode     string
	Type     string
	City     string
	Barangay string
	MinPrice *int64
	MaxPrice *int64
	Beds     *int16
	Query    string
	Sort     string
	Cursor   string
	Limit    int
}
