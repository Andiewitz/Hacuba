package reports

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect support postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping support postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) Create(ctx context.Context, report Report) (*Report, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO support_reports (id,reporter_id,category,listing_reference,contact_email,description,status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at`, report.ID, report.ReporterID, report.Category, nullString(report.ListingReference), nullString(report.ContactEmail), report.Description, report.Status).Scan(&report.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *PostgresStore) OpenCount(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM support_reports WHERE status='open'`).Scan(&count)
	return count, err
}
