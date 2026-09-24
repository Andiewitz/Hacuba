package listings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore is the production listings-db boundary. The service receives
// no auth-db DSN and never queries users directly.
type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect listings postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping listings postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}
func (s *PostgresStore) Close() { s.pool.Close() }

const listingColumns = `id, owner_id, listing_mode, property_type, title, description, price_centavos, currency, price_period, city, barangay, address_line, lat, lng, bedrooms, bathrooms, floor_area_sqm, lot_area_sqm, details, seller_name, contact_phone, contact_email, status, published_at, created_at, updated_at`
const listingJSON = `owner_id, row_to_json(listings)`

func scanListing(row pgx.Row) (*Listing, error) {
	var owner uuid.UUID
	var raw []byte
	var l Listing
	if err := row.Scan(&owner, &raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, fmt.Errorf("decode listing: %w", err)
	}
	l.OwnerID = owner
	return &l, nil
}

func (s *PostgresStore) Create(ctx context.Context, l Listing) (*Listing, error) {
	if l.Details == nil {
		l.Details = json.RawMessage(`{}`)
	}
	return scanListing(s.pool.QueryRow(ctx, `INSERT INTO listings (`+listingColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,now(),now()) RETURNING `+listingJSON,
		l.ID, l.OwnerID, nullString(l.ListingMode), nullString(l.PropertyType), nullString(l.Title), nullString(l.Description), l.PriceCentavos, l.Currency, l.PricePeriod, nullString(l.City), l.Barangay, l.AddressLine, l.Lat, l.Lng, l.Bedrooms, l.Bathrooms, l.FloorAreaSQM, l.LotAreaSQM, l.Details, nullString(l.SellerName), nullString(l.ContactPhone), nullString(l.ContactEmail), l.Status, l.PublishedAt))
}

func (s *PostgresStore) GetPublished(ctx context.Context, id uuid.UUID) (*Listing, error) {
	return scanListing(s.pool.QueryRow(ctx, `SELECT `+listingJSON+` FROM listings WHERE id=$1 AND status='published'`, id))
}
func (s *PostgresStore) GetOwner(ctx context.Context, id, owner uuid.UUID) (*Listing, error) {
	return scanListing(s.pool.QueryRow(ctx, `SELECT `+listingJSON+` FROM listings WHERE id=$1 AND owner_id=$2`, id, owner))
}

func (s *PostgresStore) SaveOwner(ctx context.Context, l Listing, owner uuid.UUID) (*Listing, error) {
	if l.Details == nil {
		l.Details = json.RawMessage(`{}`)
	}
	return scanListing(s.pool.QueryRow(ctx, `UPDATE listings SET listing_mode=$1, property_type=$2, title=$3, description=$4, price_centavos=$5, currency=$6, price_period=$7, city=$8, barangay=$9, address_line=$10, lat=$11, lng=$12, bedrooms=$13, bathrooms=$14, floor_area_sqm=$15, lot_area_sqm=$16, details=$17, seller_name=$18, contact_phone=$19, contact_email=$20, status=$21, published_at=$22, updated_at=now() WHERE id=$23 AND owner_id=$24 RETURNING `+listingJSON,
		nullString(l.ListingMode), nullString(l.PropertyType), nullString(l.Title), nullString(l.Description), l.PriceCentavos, l.Currency, l.PricePeriod, nullString(l.City), l.Barangay, l.AddressLine, l.Lat, l.Lng, l.Bedrooms, l.Bathrooms, l.FloorAreaSQM, l.LotAreaSQM, l.Details, nullString(l.SellerName), nullString(l.ContactPhone), nullString(l.ContactEmail), l.Status, l.PublishedAt, l.ID, owner))
}

func (s *PostgresStore) ListPublished(ctx context.Context, f ListFilter) ([]Listing, error) {
	where, args := []string{"status = 'published'"}, []any{}
	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}
	if f.Mode != "" {
		add("listing_mode=$%d", f.Mode)
	}
	if f.Type != "" {
		add("property_type=$%d", f.Type)
	}
	if f.City != "" {
		add("city ILIKE $%d", f.City)
	}
	if f.Barangay != "" {
		add("barangay ILIKE $%d", f.Barangay)
	}
	if f.MinPrice != nil {
		add("price_centavos >= $%d", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		add("price_centavos <= $%d", *f.MaxPrice)
	}
	if f.Beds != nil {
		add("bedrooms >= $%d", *f.Beds)
	}
	if f.Query != "" {
		add("(title ILIKE '%%' || $%d || '%%' OR city ILIKE '%%' || $%d || '%%')", f.Query)
	}
	order := "published_at DESC, id DESC"
	if f.Sort == "price_asc" {
		order = "price_centavos ASC, id ASC"
	}
	if f.Sort == "price_desc" {
		order = "price_centavos DESC, id DESC"
	}
	args = append(args, f.Limit)
	rows, err := s.pool.Query(ctx, `SELECT `+listingJSON+` FROM listings WHERE `+strings.Join(where, " AND ")+` ORDER BY `+order+fmt.Sprintf(" LIMIT $%d", len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ListOwner(ctx context.Context, owner uuid.UUID) ([]Listing, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+listingJSON+` FROM listings WHERE owner_id=$1 ORDER BY created_at DESC,id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}
func (s *PostgresStore) ListImages(ctx context.Context, id uuid.UUID) ([]Image, error) {
	rows, err := s.pool.Query(ctx, `SELECT row_to_json(listing_images) FROM listing_images WHERE listing_id=$1 ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Image
	for rows.Next() {
		var raw []byte
		var image Image
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &image); err != nil {
			return nil, err
		}
		out = append(out, image)
	}
	return out, rows.Err()
}
func (s *PostgresStore) AddImage(ctx context.Context, image Image, owner uuid.UUID) (*Image, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `INSERT INTO listing_images (id,listing_id,object_key,position,alt,content_type,byte_size) SELECT $1,$2,$3,$4,$5,$6,$7 WHERE EXISTS (SELECT 1 FROM listings WHERE id=$2 AND owner_id=$8) RETURNING row_to_json(listing_images)`, image.ID, image.ListingID, image.ObjectKey, image.Position, image.Alt, image.ContentType, image.ByteSize, owner).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &image); err != nil {
		return nil, err
	}
	return &image, nil
}
func (s *PostgresStore) DeleteImage(ctx context.Context, listingID, imageID, owner uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM listing_images WHERE id=$1 AND listing_id=$2 AND EXISTS (SELECT 1 FROM listings WHERE id=$2 AND owner_id=$3)`, imageID, listingID, owner)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
