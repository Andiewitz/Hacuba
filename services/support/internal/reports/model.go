package reports

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CategoryListing = "listing"
	CategorySafety  = "safety"
	CategoryBug     = "bug"
	CategoryAccount = "account"
	StatusOpen      = "open"
	StatusTriaged   = "triaged"
	StatusResolved  = "resolved"
)

var ErrNotFound = errors.New("not found")

type Report struct {
	ID               uuid.UUID  `json:"id"`
	ReporterID       *uuid.UUID `json:"-"`
	Category         string     `json:"category"`
	ListingReference string     `json:"listing_reference,omitempty"`
	ContactEmail     string     `json:"contact_email,omitempty"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Action struct {
	ID         uuid.UUID `json:"id"`
	ReportID   uuid.UUID `json:"report_id"`
	ActorID    uuid.UUID `json:"actor_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store interface {
	Create(context.Context, Report) (*Report, error)
	OpenCount(context.Context) (int, error)
	List(context.Context, int) ([]Report, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, string) (*Report, error)
}

func ValidTransition(from, to string) bool {
	return (from == StatusOpen && (to == StatusTriaged || to == StatusResolved)) || (from == StatusTriaged && to == StatusResolved)
}

func Validate(input Report) map[string]string {
	fields := map[string]string{}
	switch input.Category {
	case CategoryListing, CategorySafety, CategoryBug, CategoryAccount:
	default:
		fields["category"] = "choose listing, safety, bug, or account"
	}
	if len(strings.TrimSpace(input.Description)) < 20 || len(input.Description) > 2_000 {
		fields["description"] = "must be between 20 and 2,000 characters"
	}
	if len(input.ListingReference) > 500 {
		fields["listing_reference"] = "must be 500 characters or fewer"
	}
	if len(input.ContactEmail) > 254 || (input.ContactEmail != "" && !strings.Contains(input.ContactEmail, "@")) {
		fields["contact_email"] = "enter a valid email address"
	}
	return fields
}
