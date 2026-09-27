package reports

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct{ db *sql.DB }

func NewSQLiteStore(ctx context.Context, path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS dev_reports (
		id TEXT PRIMARY KEY, reporter_id TEXT, category TEXT NOT NULL, listing_reference TEXT,
		listing_id TEXT, contact_email TEXT, description TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	)`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize support SQLite: %w", err)
	}
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(dev_reports)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	hasUpdated, hasListingID := false, false
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			db.Close()
			return nil, err
		}
		if name == "updated_at" {
			hasUpdated = true
		}
		if name == "listing_id" {
			hasListingID = true
		}
	}
	rows.Close()
	if !hasUpdated {
		if _, err := db.ExecContext(ctx, `ALTER TABLE dev_reports ADD COLUMN updated_at TEXT`); err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.ExecContext(ctx, `UPDATE dev_reports SET updated_at=created_at WHERE updated_at IS NULL`); err != nil {
			db.Close()
			return nil, err
		}
	}
	if !hasListingID {
		if _, err := db.ExecContext(ctx, `ALTER TABLE dev_reports ADD COLUMN listing_id TEXT`); err != nil {
			db.Close()
			return nil, err
		}
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS dev_report_actions (id TEXT PRIMARY KEY, report_id TEXT NOT NULL, actor_id TEXT NOT NULL, from_status TEXT NOT NULL, to_status TEXT NOT NULL, created_at TEXT NOT NULL)`); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) Create(ctx context.Context, report Report) (*Report, error) {
	report.CreatedAt = time.Now().UTC()
	var reporter any
	if report.ReporterID != nil {
		reporter = report.ReporterID.String()
	}
	report.UpdatedAt = report.CreatedAt
	var listingID any
	if report.ListingID != nil {
		listingID = report.ListingID.String()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO dev_reports (id,reporter_id,category,listing_reference,listing_id,contact_email,description,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`, report.ID.String(), reporter, report.Category, nullString(report.ListingReference), listingID, nullString(report.ContactEmail), report.Description, report.Status, report.CreatedAt.Format(time.RFC3339Nano), report.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *SQLiteStore) List(ctx context.Context, limit int) ([]Report, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,category,listing_reference,listing_id,contact_email,description,status,created_at,updated_at FROM dev_reports ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Report{}
	for rows.Next() {
		var id, created, updated string
		var listing, linkedID, email sql.NullString
		var report Report
		if err := rows.Scan(&id, &report.Category, &listing, &linkedID, &email, &report.Description, &report.Status, &created, &updated); err != nil {
			return nil, err
		}
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		report.ID = parsed
		report.ListingReference = listing.String
		if linkedID.Valid {
			parsedListingID, err := uuid.Parse(linkedID.String)
			if err != nil {
				return nil, err
			}
			report.ListingID = &parsedListingID
		}
		report.ContactEmail = email.String
		report.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, err
		}
		report.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, err
		}
		out = append(out, report)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) Transition(ctx context.Context, id, actorID uuid.UUID, status string) (*Report, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var current string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM dev_reports WHERE id=?`, id.String()).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !ValidTransition(current, status) {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE dev_reports SET status=?,updated_at=? WHERE id=?`, status, now.Format(time.RFC3339Nano), id.String()); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO dev_report_actions (id,report_id,actor_id,from_status,to_status,created_at) VALUES (?,?,?,?,?,?)`, uuid.New().String(), id.String(), actorID.String(), current, status, now.Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Report{ID: id, Status: status, UpdatedAt: now}, nil
}

func (s *SQLiteStore) OpenCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM dev_reports WHERE status<>?`, StatusResolved).Scan(&count)
	return count, err
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func parseReporter(raw sql.NullString) (*uuid.UUID, error) {
	if !raw.Valid {
		return nil, nil
	}
	id, err := uuid.Parse(raw.String)
	return &id, err
}
