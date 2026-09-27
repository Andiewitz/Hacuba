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
		contact_email TEXT, description TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL
	)`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize support SQLite: %w", err)
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
	_, err := s.db.ExecContext(ctx, `INSERT INTO dev_reports (id,reporter_id,category,listing_reference,contact_email,description,status,created_at) VALUES (?,?,?,?,?,?,?,?)`, report.ID.String(), reporter, report.Category, nullString(report.ListingReference), nullString(report.ContactEmail), report.Description, report.Status, report.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *SQLiteStore) OpenCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM dev_reports WHERE status=?`, StatusOpen).Scan(&count)
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
