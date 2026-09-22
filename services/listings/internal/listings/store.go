package listings

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

// Store owns listings-db only. Every owner operation takes ownerID from a
// verified JWT context, never from an HTTP body or path value.
type Store interface {
	Create(context.Context, Listing) (*Listing, error)
	GetPublished(context.Context, uuid.UUID) (*Listing, error)
	GetOwner(context.Context, uuid.UUID, uuid.UUID) (*Listing, error)
	SaveOwner(context.Context, Listing, uuid.UUID) (*Listing, error)
	ListPublished(context.Context, ListFilter) ([]Listing, error)
	ListOwner(context.Context, uuid.UUID) ([]Listing, error)
	ListImages(context.Context, uuid.UUID) ([]Image, error)
	AddImage(context.Context, Image, uuid.UUID) (*Image, error)
	DeleteImage(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}
