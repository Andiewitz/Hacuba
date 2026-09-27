package reports

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		RETURNING created_at, updated_at`, report.ID, report.ReporterID, report.Category, nullString(report.ListingReference), nullString(report.ContactEmail), report.Description, report.Status).Scan(&report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *PostgresStore) List(ctx context.Context, limit int) ([]Report, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,category,listing_reference,contact_email,description,status,created_at,updated_at FROM support_reports ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Report{}
	for rows.Next() {
		var r Report
		var listingReference, contactEmail *string
		if err := rows.Scan(&r.ID, &r.Category, &listingReference, &contactEmail, &r.Description, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		if listingReference != nil {
			r.ListingReference = *listingReference
		}
		if contactEmail != nil {
			r.ContactEmail = *contactEmail
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *PostgresStore) Transition(ctx context.Context, id, actorID uuid.UUID, status string) (*Report, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var current string
	if err := tx.QueryRow(ctx, `SELECT status FROM support_reports WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !ValidTransition(current, status) {
		return nil, ErrNotFound
	}
	var report Report
	if err := tx.QueryRow(ctx, `UPDATE support_reports SET status=$1,updated_at=now() WHERE id=$2 RETURNING id,status,updated_at`, status, id).Scan(&report.ID, &report.Status, &report.UpdatedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO support_report_actions (id,report_id,actor_id,from_status,to_status) VALUES ($1,$2,$3,$4,$5)`, uuid.New(), id, actorID, current, status); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *PostgresStore) OpenCount(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM support_reports WHERE status<>'resolved'`).Scan(&count)
	return count, err
}
